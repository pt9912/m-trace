package auth

import (
	"context"
	"sync"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// InMemoryIssuanceRateLimiter implementiert
// `driven.IssuanceRateLimiter` für `POST /api/auth/session-tokens`
// (`0.12.0`, RAK-72). Token-Bucket pro `(global, projectID)` mit
// `Capacity` (max Burst) und `RefillPerSecond` (steady state). Beide
// Buckets werden bei jedem `Allow` in Reihenfolge `global → project`
// geprüft; ein `false` auf einer Stufe verbraucht **keine** Tokens
// auf der anderen.
//
// Sicherheitsprofil:
//   - keine Daten leaken in Logs/Metriken; der Application-Service
//     mappt `Allow=false` auf `domain.ErrAuthIssuanceRateLimited`,
//     den der HTTP-Adapter zu `429 auth_issuance_rate_limited` macht.
//   - in-process state; ein Multi-Instance-Setup braucht einen
//     gemeinsamen Backend-Limiter (Folge-Scope).
type InMemoryIssuanceRateLimiter struct {
	now            func() time.Time
	mu             sync.Mutex
	globalBucket   tokenBucket
	projectBuckets map[string]tokenBucket
	projectCfg     bucketConfig
}

// bucketConfig kapselt Capacity + RefillPerSecond.
type bucketConfig struct {
	Capacity        int
	RefillPerSecond float64
}

// tokenBucket ist die in-memory-Form: aktuelle Token-Anzahl + letzte
// Aktualisierung.
type tokenBucket struct {
	cfg    bucketConfig
	tokens float64
	lastAt time.Time
}

// NewInMemoryIssuanceRateLimiter konstruiert den Limiter mit globalen
// und Project-Default-Buckets. Tests können `now` injecten, Produktion
// nutzt `time.Now()`. Ein nicht konfiguriertes Bucket
// (`Capacity == 0` und `RefillPerSecond == 0`) wird als „kein Limit"
// behandelt — das ist ausschließlich für Tests und einen späteren
// Soft-Mode sinnvoll.
func NewInMemoryIssuanceRateLimiter(globalCap int, globalRefill float64, projectCap int, projectRefill float64) *InMemoryIssuanceRateLimiter {
	cfg := bucketConfig{Capacity: globalCap, RefillPerSecond: globalRefill}
	return &InMemoryIssuanceRateLimiter{
		now:            time.Now,
		globalBucket:   tokenBucket{cfg: cfg, tokens: float64(cfg.Capacity), lastAt: time.Now()},
		projectBuckets: make(map[string]tokenBucket),
		projectCfg:     bucketConfig{Capacity: projectCap, RefillPerSecond: projectRefill},
	}
}

// Compile-time check.
var _ driven.IssuanceRateLimiter = (*InMemoryIssuanceRateLimiter)(nil)

// Allow prüft beide Buckets. Reihenfolge ist global → project; ein
// Verstoß auf einer Stufe verbraucht **keine** Tokens auf der anderen
// (`all-or-nothing`-Commit, analog zum bestehenden `ratelimit`-
// Adapter).
func (l *InMemoryIssuanceRateLimiter) Allow(ctx context.Context, projectID string) (bool, error) {
	if l == nil {
		return true, nil
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	if !consume(&l.globalBucket, now) {
		return false, nil
	}
	pBucket, ok := l.projectBuckets[projectID]
	if !ok {
		pBucket = tokenBucket{cfg: l.projectCfg, tokens: float64(l.projectCfg.Capacity), lastAt: now}
	}
	if !consume(&pBucket, now) {
		// Refund das gerade verbrauchte globale Token, damit ein
		// Project-Limit-Verstoß nicht das globale Bucket leerlaufen
		// lässt.
		l.globalBucket.tokens = clampMax(l.globalBucket.tokens+1.0, float64(l.globalBucket.cfg.Capacity))
		l.projectBuckets[projectID] = pBucket
		return false, nil
	}
	l.projectBuckets[projectID] = pBucket
	return true, nil
}

// consume aktualisiert das Bucket gegen `now`, prüft und verbraucht
// einen Token. Liefert `true`, wenn ein Token zur Verfügung stand.
func consume(b *tokenBucket, now time.Time) bool {
	if b.cfg.Capacity == 0 && b.cfg.RefillPerSecond == 0 {
		return true // Bucket deaktiviert — kein Limit.
	}
	if b.lastAt.IsZero() {
		b.lastAt = now
		b.tokens = float64(b.cfg.Capacity)
	}
	elapsed := now.Sub(b.lastAt).Seconds()
	if elapsed > 0 {
		b.tokens = clampMax(b.tokens+elapsed*b.cfg.RefillPerSecond, float64(b.cfg.Capacity))
		b.lastAt = now
	}
	if b.tokens < 1.0 {
		return false
	}
	b.tokens -= 1.0
	return true
}

// clampMax pinnt einen `float64`-Token-Stand an die Bucket-Capacity.
func clampMax(v, ceiling float64) float64 {
	if v > ceiling {
		return ceiling
	}
	return v
}
