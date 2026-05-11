package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// SqliteIssuanceRateLimiter implementiert `driven.IssuanceRateLimiter`
// gegen die SQLite-Tabelle `auth_issuance_counters` (Migration V5,
// `0.12.5`/RAK-77). Adressiert R-17 (Multi-Replica-Issuance-Limiter)
// für Single-Host-Deployments mit Shared-Persistent-Volume — alle
// API-Replicas teilen sich die SQLite-Datei und damit den Token-
// Bucket-Counter.
//
// Atomarität: jede `Allow`-Operation läuft in einer `BeginTx`-
// Transaktion, deren `BEGIN IMMEDIATE`-Semantik durch die DSN aus
// `internal/storage` erzwungen wird (siehe ADR-0002 §8.3). Damit
// serialisiert SQLite die konkurrenten Allow-Calls über alle
// Replicas hinweg — der globale und der pro-Project-Bucket bleiben
// konsistent.
//
// Topologie-Constraint (Operator-Doku `auth.md` §5.4):
//   - Nur Single-Host (Compose-`volumes:` oder K8s-`hostPath` auf
//     demselben Host). Echte Multi-Host-Topologie braucht einen
//     Network-Backend-Adapter (Redis/Memcached) — Folge-Item nach
//     `0.12.5`.
//
// RAK-74-Scope-Cut bleibt aktiv: dieser Limiter darf nicht vor
// `/api/ingest/*` hängen. Der Application-Service ruft `Allow` nur
// im Browser-/Issuance-Pfad (`POST /api/auth/session-tokens` plus
// `POST /api/playback-events`).
type SqliteIssuanceRateLimiter struct {
	db          *sql.DB
	now         func() time.Time
	globalCfg   bucketConfig
	projectCfg  bucketConfig
	bucketTTL   time.Duration
	cleanupProb int // 1..100; 0 disabled
}

// SqliteIssuanceLimiterOption konfiguriert optionale Felder
// (Now-Stub, TTL, Cleanup-Probability) ohne Konstruktor-Aufblähung.
type SqliteIssuanceLimiterOption func(*SqliteIssuanceRateLimiter)

// WithSqliteIssuanceLimiterNow injiziert einen deterministischen
// Zeit-Provider für Tests.
func WithSqliteIssuanceLimiterNow(now func() time.Time) SqliteIssuanceLimiterOption {
	return func(l *SqliteIssuanceRateLimiter) { l.now = now }
}

// WithSqliteIssuanceLimiterBucketTTL setzt die TTL, nach der ein
// stehender Bucket per opportunistischem Cleanup entfernt werden
// darf. Default 24h.
func WithSqliteIssuanceLimiterBucketTTL(ttl time.Duration) SqliteIssuanceLimiterOption {
	return func(l *SqliteIssuanceRateLimiter) {
		if ttl > 0 {
			l.bucketTTL = ttl
		}
	}
}

// NewSqliteIssuanceRateLimiter konstruiert den Adapter. Default-TTL
// 24h, Cleanup-Probability 5 (= alle ~20 Allow-Calls räumt der
// Adapter ein Batch abgelaufener Buckets auf).
func NewSqliteIssuanceRateLimiter(
	db *sql.DB,
	globalCap int, globalRefill float64,
	projectCap int, projectRefill float64,
	opts ...SqliteIssuanceLimiterOption,
) *SqliteIssuanceRateLimiter {
	l := &SqliteIssuanceRateLimiter{
		db:          db,
		now:         time.Now,
		globalCfg:   bucketConfig{Capacity: globalCap, RefillPerSecond: globalRefill},
		projectCfg:  bucketConfig{Capacity: projectCap, RefillPerSecond: projectRefill},
		bucketTTL:   24 * time.Hour,
		cleanupProb: 5,
	}
	for _, o := range opts {
		o(l)
	}
	return l
}

// Compile-time check.
var _ driven.IssuanceRateLimiter = (*SqliteIssuanceRateLimiter)(nil)

// Allow prüft globalen und Project-Bucket atomar in einer Transaktion.
// Disabled-Buckets (Capacity 0 + Refill 0) werden vor dem DB-Roundtrip
// kurzgeschlossen — sie sind „kein Limit" und produzieren keinen
// Counter-State. Bei Project-Deny nach Global-Allow erfolgt ein
// expliziter Refund auf dem globalen Bucket innerhalb derselben Tx.
func (l *SqliteIssuanceRateLimiter) Allow(ctx context.Context, projectID string, projectBucket domain.RateLimitBucket) (bool, error) {
	if l == nil {
		return true, nil
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}

	projectCfg := l.resolveProjectConfig(projectBucket)
	globalDisabledLocal := bucketDisabled(l.globalCfg)
	projectDisabledLocal := bucketDisabled(projectCfg)

	if globalDisabledLocal && projectDisabledLocal {
		return true, nil
	}

	now := l.now().UTC()
	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("auth issuance limiter: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := l.maybeCleanupExpired(ctx, tx, now); err != nil {
		return false, err
	}

	if !globalDisabledLocal {
		ok, err := l.tryConsumeGlobal(ctx, tx, now)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, commitWithContext(tx, "global deny")
		}
	}

	if !projectDisabledLocal {
		ok, err := l.tryConsumeProject(ctx, tx, projectID, projectCfg, now, globalDisabledLocal)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, commitWithContext(tx, "project deny")
		}
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return false, fmt.Errorf("auth issuance limiter: commit (allow): %w", commitErr)
	}
	return true, nil
}

// maybeCleanupExpired räumt opportunistisch abgelaufene
// Bucket-Einträge auf. Wahrscheinlichkeit über `cleanupProb`
// gesteuert, damit der Hot-Path nicht jeden Allow scannt.
func (l *SqliteIssuanceRateLimiter) maybeCleanupExpired(ctx context.Context, tx *sql.Tx, now time.Time) error {
	if l.cleanupProb <= 0 || (int(now.UnixNano())%100) >= l.cleanupProb {
		return nil
	}
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM auth_issuance_counters WHERE expires_at < ?`,
		now.Format(time.RFC3339Nano),
	); err != nil {
		return fmt.Errorf("auth issuance limiter: cleanup: %w", err)
	}
	return nil
}

// tryConsumeGlobal versucht einen Token aus dem globalen Bucket zu
// verbrauchen. Liefert `true` bei erfolgreichem Consume, `false`
// bei deny — die offene Transaktion wird vom Caller committed.
func (l *SqliteIssuanceRateLimiter) tryConsumeGlobal(ctx context.Context, tx *sql.Tx, now time.Time) (bool, error) {
	return l.refillAndConsume(ctx, tx, "global", l.globalCfg, now)
}

// tryConsumeProject versucht einen Token aus dem Project-Bucket zu
// verbrauchen. Bei deny refundet die Methode (sofern globaler
// Bucket aktiv) den im selben Call dekrementierten globalen Token —
// das ist der asymmetrische Refund-Pfad analog zur In-Memory-
// Implementierung.
func (l *SqliteIssuanceRateLimiter) tryConsumeProject(ctx context.Context, tx *sql.Tx, projectID string, projectCfg bucketConfig, now time.Time, globalDisabled bool) (bool, error) {
	ok, err := l.refillAndConsume(ctx, tx, "project:"+projectID, projectCfg, now)
	if err != nil {
		return false, err
	}
	if ok {
		return true, nil
	}
	if !globalDisabled {
		if err := l.refundOne(ctx, tx, "global", l.globalCfg, now); err != nil {
			return false, err
		}
	}
	return false, nil
}

// commitWithContext committet die laufende Tx und liefert bei
// Commit-Fehler eine kontextualisierte Fehlermeldung. Wird in den
// deny-Pfaden genutzt, damit der `Allow`-Body kompakt bleibt.
func commitWithContext(tx *sql.Tx, label string) error {
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("auth issuance limiter: commit (%s): %w", label, err)
	}
	return nil
}

// refillAndConsume liest den Bucket, baut Refill-Tokens auf, prüft
// und persistiert den neuen Zustand. Existiert der Bucket noch nicht,
// wird er mit voller Kapazität initialisiert.
func (l *SqliteIssuanceRateLimiter) refillAndConsume(ctx context.Context, tx *sql.Tx, key string, cfg bucketConfig, now time.Time) (bool, error) {
	current, found, err := loadBucket(ctx, tx, key)
	if err != nil {
		return false, err
	}
	expiresAt := now.Add(l.bucketTTL).Format(time.RFC3339Nano)
	if !found || current.capacity != cfg.Capacity || current.refill != cfg.RefillPerSecond {
		// Erstaufruf oder Konfig-Wechsel: voller Bucket.
		tokens := float64(cfg.Capacity)
		if tokens < 1.0 {
			// Capacity 0 sollte hier nicht ankommen (disabled-Pfad),
			// aber zur Sicherheit: kein Token konsumierbar.
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO auth_issuance_counters(bucket_key, capacity, refill_per_second, tokens, last_at, expires_at)
                 VALUES (?, ?, ?, ?, ?, ?)
                 ON CONFLICT(bucket_key) DO UPDATE SET
                     capacity = excluded.capacity,
                     refill_per_second = excluded.refill_per_second,
                     tokens = excluded.tokens,
                     last_at = excluded.last_at,
                     expires_at = excluded.expires_at`,
				key, cfg.Capacity, cfg.RefillPerSecond, tokens, now.Format(time.RFC3339Nano), expiresAt,
			); err != nil {
				return false, fmt.Errorf("auth issuance limiter: upsert empty bucket %s: %w", key, err)
			}
			return false, nil
		}
		tokens -= 1.0
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO auth_issuance_counters(bucket_key, capacity, refill_per_second, tokens, last_at, expires_at)
             VALUES (?, ?, ?, ?, ?, ?)
             ON CONFLICT(bucket_key) DO UPDATE SET
                 capacity = excluded.capacity,
                 refill_per_second = excluded.refill_per_second,
                 tokens = excluded.tokens,
                 last_at = excluded.last_at,
                 expires_at = excluded.expires_at`,
			key, cfg.Capacity, cfg.RefillPerSecond, tokens, now.Format(time.RFC3339Nano), expiresAt,
		); err != nil {
			return false, fmt.Errorf("auth issuance limiter: upsert init bucket %s: %w", key, err)
		}
		return true, nil
	}

	// Bekannter Bucket mit gleicher Cfg — refill + consume.
	elapsed := now.Sub(current.lastAt).Seconds()
	tokens := current.tokens
	if elapsed > 0 {
		tokens = clampMax(tokens+elapsed*cfg.RefillPerSecond, float64(cfg.Capacity))
	}
	if tokens < 1.0 {
		if _, err := tx.ExecContext(ctx,
			`UPDATE auth_issuance_counters
             SET tokens = ?, last_at = ?, expires_at = ?
             WHERE bucket_key = ?`,
			tokens, now.Format(time.RFC3339Nano), expiresAt, key,
		); err != nil {
			return false, fmt.Errorf("auth issuance limiter: update bucket %s: %w", key, err)
		}
		return false, nil
	}
	tokens -= 1.0
	if _, err := tx.ExecContext(ctx,
		`UPDATE auth_issuance_counters
         SET tokens = ?, last_at = ?, expires_at = ?
         WHERE bucket_key = ?`,
		tokens, now.Format(time.RFC3339Nano), expiresAt, key,
	); err != nil {
		return false, fmt.Errorf("auth issuance limiter: update bucket %s: %w", key, err)
	}
	return true, nil
}

// refundOne erhöht den Token-Stand eines Buckets um 1, geclamped auf
// Capacity. Wird vom asymmetrischen Refund-Pfad genutzt
// (global wurde im selben Call dekrementiert, project hat denied —
// global bekommt sein Token zurück).
func (l *SqliteIssuanceRateLimiter) refundOne(ctx context.Context, tx *sql.Tx, key string, cfg bucketConfig, now time.Time) error {
	current, found, err := loadBucket(ctx, tx, key)
	if err != nil {
		return err
	}
	if !found {
		// Sollte nicht passieren — refund kommt direkt nach einem
		// erfolgreichen consume in derselben Tx; defensive.
		return nil
	}
	tokens := clampMax(current.tokens+1.0, float64(cfg.Capacity))
	expiresAt := now.Add(l.bucketTTL).Format(time.RFC3339Nano)
	if _, err := tx.ExecContext(ctx,
		`UPDATE auth_issuance_counters
         SET tokens = ?, last_at = ?, expires_at = ?
         WHERE bucket_key = ?`,
		tokens, now.Format(time.RFC3339Nano), expiresAt, key,
	); err != nil {
		return fmt.Errorf("auth issuance limiter: refund %s: %w", key, err)
	}
	return nil
}

func (l *SqliteIssuanceRateLimiter) resolveProjectConfig(override domain.RateLimitBucket) bucketConfig {
	if override.IsZero() {
		return l.projectCfg
	}
	return bucketConfig{Capacity: override.Capacity, RefillPerSecond: override.RefillPerSecond}
}

// bucketRow ist der DB-Read-Snapshot eines Counter-Eintrags.
type bucketRow struct {
	capacity int
	refill   float64
	tokens   float64
	lastAt   time.Time
}

// loadBucket lädt einen Counter-Eintrag in der laufenden Tx; liefert
// `found=false`, wenn der Eintrag fehlt.
func loadBucket(ctx context.Context, tx *sql.Tx, key string) (bucketRow, bool, error) {
	row := tx.QueryRowContext(ctx,
		`SELECT capacity, refill_per_second, tokens, last_at
         FROM auth_issuance_counters
         WHERE bucket_key = ?`, key)
	var (
		out      bucketRow
		lastAtTS string
	)
	if err := row.Scan(&out.capacity, &out.refill, &out.tokens, &lastAtTS); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return bucketRow{}, false, nil
		}
		return bucketRow{}, false, fmt.Errorf("auth issuance limiter: load bucket %s: %w", key, err)
	}
	parsed, parseErr := time.Parse(time.RFC3339Nano, lastAtTS)
	if parseErr != nil {
		return bucketRow{}, false, fmt.Errorf("auth issuance limiter: parse last_at for bucket %s: %w", key, parseErr)
	}
	out.lastAt = parsed.UTC()
	return out, true, nil
}

// bucketDisabled gibt true zurück, wenn Capacity und Refill beide 0
// sind. Disabled-Buckets erzeugen keinen DB-State (siehe Allow).
func bucketDisabled(cfg bucketConfig) bool {
	return cfg.Capacity == 0 && cfg.RefillPerSecond == 0
}
