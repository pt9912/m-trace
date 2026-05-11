package auth_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/auth"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/internal/storage"
)

// openTestDB hängt eine frische SQLite-Datei mit allen Migrationen
// (inkl. V5__auth_issuance_counters.sql) ans Test-Lifecycle. Pattern
// analog `sqlite/race_test.go`.
func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "m-trace.db")
	db, err := storage.Open(ctx, path)
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// openSharedTestDB öffnet zwei *sql.DB-Verbindungen auf dieselbe
// SQLite-Datei. Damit lassen sich Cross-Instance-Sharing-Szenarien
// abbilden, ohne zwei Go-Prozesse zu starten — semantisch deckt das
// den Single-Host-Compose-Replica-Pfad ab (zwei Replicas auf dem
// gleichen Bind-Mount). Anti-Pattern: nicht zwei `Open` auf
// derselben Datei in Production — der Adapter ist auf Concurrent
// Access ausgelegt, aber zwei Open-Calls im selben Prozess
// duplizieren das Connection-Pool-Setup unnötig.
func openSharedTestDB(t *testing.T) (*sql.DB, *sql.DB) {
	t.Helper()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "m-trace.db")
	db1, err := storage.Open(ctx, path)
	if err != nil {
		t.Fatalf("storage.Open #1: %v", err)
	}
	t.Cleanup(func() { _ = db1.Close() })
	db2, err := storage.Open(ctx, path)
	if err != nil {
		t.Fatalf("storage.Open #2: %v", err)
	}
	t.Cleanup(func() { _ = db2.Close() })
	return db1, db2
}

func TestSqliteIssuanceRateLimiter_AllowsUpToCapacity(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := openTestDB(t)
	now := time.Date(2026, 5, 11, 12, 0, 0, 0, time.UTC)
	l := auth.NewSqliteIssuanceRateLimiter(db, 3, 0, 2, 0,
		auth.WithSqliteIssuanceLimiterNow(func() time.Time { return now }))

	for i := 0; i < 2; i++ {
		allowed, err := l.Allow(ctx, "p1", domain.RateLimitBucket{})
		if err != nil {
			t.Fatalf("Allow #%d err: %v", i, err)
		}
		if !allowed {
			t.Fatalf("Allow #%d must be true (within capacity)", i)
		}
	}
	// Project-Capacity erschöpft (cap=2).
	allowed, err := l.Allow(ctx, "p1", domain.RateLimitBucket{})
	if err != nil {
		t.Fatalf("Allow #3 err: %v", err)
	}
	if allowed {
		t.Errorf("Allow #3 must be false (project cap exhausted)")
	}
}

func TestSqliteIssuanceRateLimiter_RefillsOverTime(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := openTestDB(t)
	now := time.Date(2026, 5, 11, 12, 0, 0, 0, time.UTC)
	cur := now
	l := auth.NewSqliteIssuanceRateLimiter(db, 5, 1.0, 1, 1.0,
		auth.WithSqliteIssuanceLimiterNow(func() time.Time { return cur }))

	// Erste Allow ok.
	if ok, err := l.Allow(ctx, "p1", domain.RateLimitBucket{}); err != nil || !ok {
		t.Fatalf("Allow #1: ok=%v err=%v", ok, err)
	}
	// Zweite ohne Zeitfortschritt: project cap 1 ist erschöpft.
	if ok, _ := l.Allow(ctx, "p1", domain.RateLimitBucket{}); ok {
		t.Errorf("Allow #2 must be false before refill")
	}
	// 2 Sekunden vorrücken → bei refill=1/s reicht 1.0 Token.
	cur = cur.Add(2 * time.Second)
	if ok, err := l.Allow(ctx, "p1", domain.RateLimitBucket{}); err != nil || !ok {
		t.Fatalf("Allow after refill: ok=%v err=%v", ok, err)
	}
}

func TestSqliteIssuanceRateLimiter_ProjectBucketOverride(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := openTestDB(t)
	now := time.Date(2026, 5, 11, 12, 0, 0, 0, time.UTC)
	// Default-project-cap 100 — wird durch override auf 1 reduziert.
	l := auth.NewSqliteIssuanceRateLimiter(db, 100, 0, 100, 0,
		auth.WithSqliteIssuanceLimiterNow(func() time.Time { return now }))
	override := domain.RateLimitBucket{Capacity: 1, RefillPerSecond: 0}

	if ok, err := l.Allow(ctx, "p1", override); err != nil || !ok {
		t.Fatalf("Allow #1 with override: ok=%v err=%v", ok, err)
	}
	if ok, _ := l.Allow(ctx, "p1", override); ok {
		t.Errorf("Allow #2 with override-cap=1 must be false")
	}
}

func TestSqliteIssuanceRateLimiter_RefundsGlobalOnProjectDeny(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := openTestDB(t)
	now := time.Date(2026, 5, 11, 12, 0, 0, 0, time.UTC)
	// Global=2, Project=1 — zweiter Allow muss vom Project rejected
	// werden, ohne den globalen Counter zu verbrauchen (Refund).
	l := auth.NewSqliteIssuanceRateLimiter(db, 2, 0, 1, 0,
		auth.WithSqliteIssuanceLimiterNow(func() time.Time { return now }))

	if ok, _ := l.Allow(ctx, "p1", domain.RateLimitBucket{}); !ok {
		t.Fatalf("Allow #1 must succeed")
	}
	// Zweiter Allow: global wäre noch frei (1 übrig), project ist
	// erschöpft. Adapter muss project deny + global refund liefern.
	if ok, _ := l.Allow(ctx, "p1", domain.RateLimitBucket{}); ok {
		t.Fatalf("Allow #2 must be denied (project cap=1)")
	}
	// Wenn der Refund-Pfad wirkt, hat das globale Bucket immer noch
	// 1 freien Token. Ein anderer Project muss daher noch durchgehen.
	if ok, err := l.Allow(ctx, "p2", domain.RateLimitBucket{}); err != nil || !ok {
		t.Errorf("Allow on different project must succeed (proves global refund): ok=%v err=%v", ok, err)
	}
}

func TestSqliteIssuanceRateLimiter_DisabledBucketsAlwaysAllow(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := openTestDB(t)
	l := auth.NewSqliteIssuanceRateLimiter(db, 0, 0, 0, 0)
	for i := 0; i < 100; i++ {
		if ok, err := l.Allow(ctx, "p1", domain.RateLimitBucket{}); err != nil || !ok {
			t.Fatalf("disabled bucket must always allow: ok=%v err=%v", ok, err)
		}
	}
}

func TestSqliteIssuanceRateLimiter_NilReceiver(t *testing.T) {
	t.Parallel()
	var l *auth.SqliteIssuanceRateLimiter
	if ok, err := l.Allow(context.Background(), "p1", domain.RateLimitBucket{}); err != nil || !ok {
		t.Errorf("nil receiver must allow (no-op): ok=%v err=%v", ok, err)
	}
}

func TestSqliteIssuanceRateLimiter_ContextCancelled(t *testing.T) {
	t.Parallel()
	db := openTestDB(t)
	l := auth.NewSqliteIssuanceRateLimiter(db, 10, 1, 5, 1)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := l.Allow(ctx, "p1", domain.RateLimitBucket{}); err == nil {
		t.Errorf("expected ctx err, got nil")
	}
}

// TestSqliteIssuanceRateLimiter_SharedAcrossInstances ist der
// semantische Kern von RAK-77: zwei Adapter-Instances, die auf
// dieselbe SQLite-Datei zeigen, teilen sich den Bucket-Counter.
// Wenn Instance A das gesamte Project-Limit verbraucht, muss
// Instance B den nächsten Allow als „denied" sehen — das ist genau
// das Verhalten, das R-17 für den Multi-Replica-Pfad fordert.
func TestSqliteIssuanceRateLimiter_SharedAcrossInstances(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	dbA, dbB := openSharedTestDB(t)
	now := time.Date(2026, 5, 11, 12, 0, 0, 0, time.UTC)
	limiterA := auth.NewSqliteIssuanceRateLimiter(dbA, 100, 0, 2, 0,
		auth.WithSqliteIssuanceLimiterNow(func() time.Time { return now }))
	limiterB := auth.NewSqliteIssuanceRateLimiter(dbB, 100, 0, 2, 0,
		auth.WithSqliteIssuanceLimiterNow(func() time.Time { return now }))

	// Instance A verbraucht beide Project-Tokens.
	for i := 0; i < 2; i++ {
		if ok, err := limiterA.Allow(ctx, "p1", domain.RateLimitBucket{}); err != nil || !ok {
			t.Fatalf("limiterA Allow #%d failed: ok=%v err=%v", i, ok, err)
		}
	}
	// Instance B sieht das Project-Bucket als erschöpft an — der
	// Shared-State greift.
	if ok, err := limiterB.Allow(ctx, "p1", domain.RateLimitBucket{}); err != nil {
		t.Fatalf("limiterB Allow err: %v", err)
	} else if ok {
		t.Errorf("limiterB must observe exhausted project bucket from shared SQLite — got Allow=true")
	}
	// Ein anderer Project muss in Instance B immer noch durchgehen.
	if ok, err := limiterB.Allow(ctx, "p2", domain.RateLimitBucket{}); err != nil || !ok {
		t.Errorf("limiterB allow for different project must succeed: ok=%v err=%v", ok, err)
	}
	// Und Instance A sieht nun ebenfalls p2 als verbraucht (1/2).
	if ok, err := limiterA.Allow(ctx, "p2", domain.RateLimitBucket{}); err != nil || !ok {
		t.Errorf("limiterA second-p2 must succeed (1 of 2 used by B): ok=%v err=%v", ok, err)
	}
	if ok, _ := limiterA.Allow(ctx, "p2", domain.RateLimitBucket{}); ok {
		t.Errorf("limiterA third-p2 must be denied (shared cap=2 exhausted by A+B)")
	}
}
