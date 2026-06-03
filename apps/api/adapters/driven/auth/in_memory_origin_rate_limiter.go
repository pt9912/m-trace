package auth

import (
	"context"
	"sync"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// InMemoryOriginRateLimiter implementiert
// `driven.OriginRateLimiter` für den Defense-in-Depth-Pfad vor den
// Auth- und Ingest-Hot-Path-Handlern (R-22).
// Token-Bucket pro Key mit `Capacity` (max Burst) und
// `RefillPerSecond` (steady state); wiederverwendet die `tokenBucket`-
// und `consume`-Helper aus `in_memory_issuance_rate_limiter.go`.
//
// Sicherheitsprofil:
//  - Single-Process State; ein Multi-Host-Setup misst pro Replica
//  (gleicher Trade-off wie der Issuance-Limiter; Multi-
//  Host-Lösung ist Folge-Item für die Redis-Variante in
//  ).
//  - Bucket-Map wächst pro neuem Key. Opportunistisches Eviction
//  räumt idle Buckets (siehe `sweepInterval`).
//  - `key == ""` ist No-Op `true` — fehlende RemoteAddr-/XFF-Info
//  darf den Hot-Path nicht blockieren.
type InMemoryOriginRateLimiter struct {
	now           func() time.Time
	mu            sync.Mutex
	buckets       map[string]tokenBucket
	cfg           bucketConfig
	lastSweep     time.Time
	sweepInterval time.Duration
	idleTTL       time.Duration
}

// NewInMemoryOriginRateLimiter konstruiert den Limiter. `capacity=0`
// und `refill=0` bedeutet „kein Limit" (Default-disabled-Pfad);
// sinnvolle Lab-Werte sind capacity=20, refill=5 pro Sekunde
// (Operator-Doku in `docs/user/auth.md`).
func NewInMemoryOriginRateLimiter(capacity int, refillPerSecond float64, opts ...InMemoryOriginRateLimiterOption) *InMemoryOriginRateLimiter {
	l := &InMemoryOriginRateLimiter{
		now:           time.Now,
		buckets:       make(map[string]tokenBucket),
		cfg:           bucketConfig{Capacity: capacity, RefillPerSecond: refillPerSecond},
		sweepInterval: 5 * time.Minute,
		idleTTL:       10 * time.Minute,
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

// InMemoryOriginRateLimiterOption konfiguriert optionale Felder
// (Now-Stub) für Tests.
type InMemoryOriginRateLimiterOption func(*InMemoryOriginRateLimiter)

// WithInMemoryOriginRateLimiterNow injiziert einen deterministischen
// Zeit-Provider für Tests (analog `WithSqliteIssuanceLimiterNow`).
func WithInMemoryOriginRateLimiterNow(now func() time.Time) InMemoryOriginRateLimiterOption {
	return func(l *InMemoryOriginRateLimiter) {
		if now != nil {
			l.now = now
		}
	}
}

// Compile-time check.
var _ driven.OriginRateLimiter = (*InMemoryOriginRateLimiter)(nil)

// Allow prüft das Bucket für `key` und verbraucht bei Erfolg einen
// Token. `key == ""` → `(true, nil)` (No-Op). Disabled Bucket
// (`Capacity == 0 && Refill == 0`) → `(true, nil)`.
func (l *InMemoryOriginRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	if l == nil {
		return true, nil
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if key == "" {
		return true, nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	l.maybeSweep(now)
	b, ok := l.buckets[key]
	if !ok {
		b = tokenBucket{cfg: l.cfg, tokens: float64(l.cfg.Capacity), lastAt: now}
	}
	allowed := consume(&b, now)
	l.buckets[key] = b
	return allowed, nil
}

// maybeSweep entfernt idle Buckets (kein Hit innerhalb `idleTTL`)
// alle `sweepInterval`. Eviction ist opportunistisch und läuft auf
// einem Allow-Aufruf — keine Background-Goroutine.
func (l *InMemoryOriginRateLimiter) maybeSweep(now time.Time) {
	if now.Sub(l.lastSweep) < l.sweepInterval {
		return
	}
	l.lastSweep = now
	if l.idleTTL <= 0 {
		return
	}
	for k, b := range l.buckets {
		if now.Sub(b.lastAt) > l.idleTTL {
			delete(l.buckets, k)
		}
	}
}
