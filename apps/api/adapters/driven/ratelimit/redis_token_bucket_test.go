package ratelimit_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/ratelimit"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// startMiniredis startet einen In-Process-Mock-Server und konstruiert
// einen go-redis-Client, der dagegen läuft (Muster aus den Auth-
// Limiter-Tests). Cleanup räumt Server + Client auf.
func startMiniredis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()
	s := miniredis.RunT(t)
	c := redis.NewClient(&redis.Options{Addr: s.Addr()})
	t.Cleanup(func() { _ = c.Close() })
	return s, c
}

// fakeClock ist eine injizierbare, manuell vorrückbare Uhr — die
// Zeitquelle des Adapters ist Client-seitig (ARGV), daher sind Refill-
// und Skew-Verhalten damit deterministisch testbar.
type fakeClock struct {
	mu sync.Mutex
	t  time.Time
}

func newFakeClock(t0 time.Time) *fakeClock { return &fakeClock{t: t0} }

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

func (c *fakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	c.t = c.t.Add(d)
	c.mu.Unlock()
}

func testBase() time.Time { return time.Unix(1_700_000_000, 0) }

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func mkRedisLimiter(t *testing.T, client redis.UniversalClient, cfg ratelimit.RedisTokenBucketConfig) *ratelimit.RedisTokenBucketRateLimiter {
	t.Helper()
	l, err := ratelimit.NewRedisTokenBucketRateLimiter(client, cfg, discardLogger())
	if err != nil {
		t.Fatalf("limiter: %v", err)
	}
	return l
}

// TestRedisTokenBucket_CrossInstanceSharing ist der R-26-b-Kern als
// Unit-Beleg: zwei Adapter-Instanzen (analog zwei API-Replicas) teilen
// EIN Per-Projekt-Budget über den geteilten Redis — statt N × Capacity
// beim In-Process-Bucket.
func TestRedisTokenBucket_CrossInstanceSharing(t *testing.T) {
	t.Parallel()
	_, client := startMiniredis(t)
	ctx := context.Background()
	clock := newFakeClock(testBase())
	cfg := ratelimit.RedisTokenBucketConfig{
		Capacity: 10, RefillPerSecond: 0, Now: clock.Now,
	}
	a := mkRedisLimiter(t, client, cfg)
	b := mkRedisLimiter(t, client, cfg)
	key := driven.RateLimitKey{ProjectID: "demo"}

	if err := a.Allow(ctx, key, 6); err != nil {
		t.Fatalf("replica A n=6: %v", err)
	}
	if err := b.Allow(ctx, key, 4); err != nil {
		t.Fatalf("replica B n=4 (cross-instance sharing broken): %v", err)
	}
	for label, l := range map[string]*ratelimit.RedisTokenBucketRateLimiter{"A": a, "B": b} {
		if err := l.Allow(ctx, key, 1); !errors.Is(err, domain.ErrRateLimited) {
			t.Errorf("replica %s after budget exhausted: err=%v, want ErrRateLimited (shared bucket should be empty)", label, err)
		}
	}
}

// TestRedisTokenBucket_AllOrNothingAcrossDimensions: ein Deny in einer
// Dimension darf in keiner anderen Tokens verbrauchen (Parität zum
// In-Memory-Adapter).
func TestRedisTokenBucket_AllOrNothingAcrossDimensions(t *testing.T) {
	t.Parallel()
	_, client := startMiniredis(t)
	ctx := context.Background()
	clock := newFakeClock(testBase())
	l := mkRedisLimiter(t, client, ratelimit.RedisTokenBucketConfig{
		Capacity: 10, RefillPerSecond: 0, Now: clock.Now,
	})
	full := driven.RateLimitKey{ProjectID: "p1", ClientIP: "10.0.0.1", Origin: "https://app.example"}

	// Projekt-Bucket auf 2 Rest-Tokens bringen (nur Projekt-Dimension).
	if err := l.Allow(ctx, driven.RateLimitKey{ProjectID: "p1"}, 8); err != nil {
		t.Fatalf("prime project bucket: %v", err)
	}
	// Voller Key mit n=3: Projekt hat nur noch 2 → Deny …
	if err := l.Allow(ctx, full, 3); !errors.Is(err, domain.ErrRateLimited) {
		t.Fatalf("full key n=3: err=%v, want ErrRateLimited", err)
	}
	// … und ip/origin sind unangetastet (volle Kapazität abrufbar).
	if err := l.Allow(ctx, driven.RateLimitKey{ClientIP: "10.0.0.1"}, 10); err != nil {
		t.Errorf("client_ip bucket must be untouched after deny: %v", err)
	}
	if err := l.Allow(ctx, driven.RateLimitKey{Origin: "https://app.example"}, 10); err != nil {
		t.Errorf("origin bucket must be untouched after deny: %v", err)
	}
	// Projekt-Rest (2) ist ebenfalls noch da.
	if err := l.Allow(ctx, driven.RateLimitKey{ProjectID: "p1"}, 2); err != nil {
		t.Errorf("project remainder must survive the deny: %v", err)
	}
}

// TestRedisTokenBucket_RefillOverTime: Refill in Token/s, gedeckelt auf
// Capacity — wie der In-Memory-Adapter, nur mit ARGV-Zeit.
func TestRedisTokenBucket_RefillOverTime(t *testing.T) {
	t.Parallel()
	_, client := startMiniredis(t)
	ctx := context.Background()
	clock := newFakeClock(testBase())
	l := mkRedisLimiter(t, client, ratelimit.RedisTokenBucketConfig{
		Capacity: 10, RefillPerSecond: 5, Now: clock.Now,
	})
	key := driven.RateLimitKey{ProjectID: "refill"}

	if err := l.Allow(ctx, key, 10); err != nil {
		t.Fatalf("drain: %v", err)
	}
	if err := l.Allow(ctx, key, 1); !errors.Is(err, domain.ErrRateLimited) {
		t.Fatalf("drained bucket: err=%v, want ErrRateLimited", err)
	}
	clock.Advance(time.Second)
	if err := l.Allow(ctx, key, 5); err != nil {
		t.Fatalf("after 1s at 5/s want 5 tokens: %v", err)
	}
	if err := l.Allow(ctx, key, 1); !errors.Is(err, domain.ErrRateLimited) {
		t.Fatalf("refill must not exceed elapsed*rate: err=%v", err)
	}
	clock.Advance(10 * time.Second) // 50 Tokens rechnerisch → Cap bei 10.
	if err := l.Allow(ctx, key, 10); err != nil {
		t.Fatalf("cap at capacity: %v", err)
	}
	if err := l.Allow(ctx, key, 1); !errors.Is(err, domain.ErrRateLimited) {
		t.Fatalf("nothing beyond capacity: err=%v", err)
	}
}

// TestRedisTokenBucket_ClockSkewNoInflation ist der Plan-§4.1-Beleg
// (nur hier testbar — ein Compose-Lab teilt eine Host-Uhr): zwei
// Replicas mit gegeneinander versetzten Uhren dürfen sich über das
// unconditional `last_at = now`-Muster KEINE Extra-Tokens erschleichen.
// Ohne monotones last_at würde die nachgehende Replica B last_at
// regressieren und die vorgehende Replica A bekäme die Skew-Differenz
// bei jedem Wechsel erneut als Refill (hier 50 ms × 1000/s = 50 Tokens
// pro Wechsel — der Test würde massiv über Budget erlauben).
func TestRedisTokenBucket_ClockSkewNoInflation(t *testing.T) {
	t.Parallel()
	_, client := startMiniredis(t)
	ctx := context.Background()
	const capacity = 10
	fast := newFakeClock(testBase())                        // Replica A
	slow := newFakeClock(testBase().Add(-50 * time.Millisecond)) // Replica B, nachgehend
	cfg := func(now func() time.Time) ratelimit.RedisTokenBucketConfig {
		return ratelimit.RedisTokenBucketConfig{Capacity: capacity, RefillPerSecond: 1000, Now: now}
	}
	a := mkRedisLimiter(t, client, cfg(fast.Now))
	b := mkRedisLimiter(t, client, cfg(slow.Now))
	key := driven.RateLimitKey{ProjectID: "skew"}

	allowed := 0
	for i := 0; i < 4*capacity; i++ {
		l := a
		if i%2 == 1 {
			l = b
		}
		switch err := l.Allow(ctx, key, 1); {
		case err == nil:
			allowed++
		case !errors.Is(err, domain.ErrRateLimited):
			t.Fatalf("call %d: unexpected error %v", i, err)
		}
	}
	// Die Uhren stehen still → es darf exakt die Kapazität erlaubt werden,
	// kein einziges Skew-Inflations-Token.
	if allowed != capacity {
		t.Fatalf("alternating skewed replicas allowed %d tokens, want exactly %d (refill inflation!)", allowed, capacity)
	}
}

// TestRedisTokenBucket_OutageFailOpen (Default): Redis weg → Delegation
// an den lokalen In-Memory-Fallback; die Limitierung selbst bleibt
// (Fallback-Budget wird durchgesetzt), nur die repliken-übergreifende
// Fairness pausiert.
func TestRedisTokenBucket_OutageFailOpen(t *testing.T) {
	t.Parallel()
	s, client := startMiniredis(t)
	ctx := context.Background()
	clock := newFakeClock(testBase())
	l := mkRedisLimiter(t, client, ratelimit.RedisTokenBucketConfig{
		Capacity: 3, RefillPerSecond: 0, Now: clock.Now,
	})
	key := driven.RateLimitKey{ProjectID: "outage"}

	s.Close() // Outage simulieren.
	for i := 1; i <= 3; i++ {
		if err := l.Allow(ctx, key, 1); err != nil {
			t.Fatalf("fail-open call %d must be served by fallback: %v", i, err)
		}
	}
	if err := l.Allow(ctx, key, 1); !errors.Is(err, domain.ErrRateLimited) {
		t.Fatalf("fallback must still enforce the budget: err=%v", err)
	}
}

// TestRedisTokenBucket_OutageRecovery: nach dem Redis-Comeback wird
// wieder der geteilte Bucket bedient — inkl. NOSCRIPT-Pfad (der
// Neustart leert den Script-Cache, EVALSHA fällt auf EVAL zurück).
func TestRedisTokenBucket_OutageRecovery(t *testing.T) {
	t.Parallel()
	s, client := startMiniredis(t)
	ctx := context.Background()
	clock := newFakeClock(testBase())
	l := mkRedisLimiter(t, client, ratelimit.RedisTokenBucketConfig{
		Capacity: 5, RefillPerSecond: 0, Now: clock.Now,
	})
	key := driven.RateLimitKey{ProjectID: "recovery"}

	s.Close()
	if err := l.Allow(ctx, key, 1); err != nil {
		t.Fatalf("during outage (fail-open): %v", err)
	}
	if err := s.Restart(); err != nil {
		t.Fatalf("miniredis restart: %v", err)
	}
	// Frischer Server = frischer Bucket; entscheidend: kein Fehler trotz
	// geleertem Script-Cache (EVALSHA→NOSCRIPT→EVAL) und Budget wirkt.
	if err := l.Allow(ctx, key, 5); err != nil {
		t.Fatalf("after recovery: %v", err)
	}
	if err := l.Allow(ctx, key, 1); !errors.Is(err, domain.ErrRateLimited) {
		t.Fatalf("after recovery budget must be shared again: err=%v", err)
	}
}

// TestRedisTokenBucket_OutageFailClosed (Opt-in): Redis weg → Deny
// (429-Pfad), nie ein anderer Fehler durch den Port.
func TestRedisTokenBucket_OutageFailClosed(t *testing.T) {
	t.Parallel()
	s, client := startMiniredis(t)
	ctx := context.Background()
	clock := newFakeClock(testBase())
	l := mkRedisLimiter(t, client, ratelimit.RedisTokenBucketConfig{
		Capacity: 3, RefillPerSecond: 0, FailClosed: true, Now: clock.Now,
	})
	s.Close()
	err := l.Allow(ctx, driven.RateLimitKey{ProjectID: "strict"}, 1)
	if !errors.Is(err, domain.ErrRateLimited) {
		t.Fatalf("fail-closed outage: err=%v, want ErrRateLimited", err)
	}
}

// TestRedisTokenBucket_ContextCanceled: ein toter Context wird wie eine
// Outage behandelt (Fallback bzw. Deny) — der Port kennt nur error, und
// ctx.Err() durch den Port würde an der Call-Site als rate-limited
// gezählt und als 500 enden.
func TestRedisTokenBucket_ContextCanceled(t *testing.T) {
	t.Parallel()
	_, client := startMiniredis(t)
	clock := newFakeClock(testBase())
	canceled, cancel := context.WithCancel(context.Background())
	cancel()

	open := mkRedisLimiter(t, client, ratelimit.RedisTokenBucketConfig{
		Capacity: 2, RefillPerSecond: 0, Now: clock.Now,
	})
	if err := open.Allow(canceled, driven.RateLimitKey{ProjectID: "ctx"}, 1); err != nil {
		t.Fatalf("fail-open + canceled ctx: fallback must decide, got %v", err)
	}

	strict := mkRedisLimiter(t, client, ratelimit.RedisTokenBucketConfig{
		Capacity: 2, RefillPerSecond: 0, FailClosed: true, Now: clock.Now,
	})
	if err := strict.Allow(canceled, driven.RateLimitKey{ProjectID: "ctx"}, 1); !errors.Is(err, domain.ErrRateLimited) {
		t.Fatalf("fail-closed + canceled ctx: err=%v, want ErrRateLimited", err)
	}
}

// TestRedisTokenBucket_EmptyKeyAndNonPositiveN: No-op-Verträge des
// In-Memory-Adapters gelten unverändert — und erzeugen keine Redis-Keys.
func TestRedisTokenBucket_EmptyKeyAndNonPositiveN(t *testing.T) {
	t.Parallel()
	s, client := startMiniredis(t)
	ctx := context.Background()
	clock := newFakeClock(testBase())
	l := mkRedisLimiter(t, client, ratelimit.RedisTokenBucketConfig{
		Capacity: 1, RefillPerSecond: 0, Now: clock.Now,
	})
	if err := l.Allow(ctx, driven.RateLimitKey{}, 5); err != nil {
		t.Fatalf("empty key must be a no-op: %v", err)
	}
	if err := l.Allow(ctx, driven.RateLimitKey{ProjectID: "p"}, 0); err != nil {
		t.Fatalf("n=0 must be a no-op: %v", err)
	}
	if err := l.Allow(ctx, driven.RateLimitKey{ProjectID: "p"}, -3); err != nil {
		t.Fatalf("n<0 must be a no-op: %v", err)
	}
	if keys := s.Keys(); len(keys) != 0 {
		t.Fatalf("no-ops must not create redis keys, got %v", keys)
	}
}

// TestRedisTokenBucket_OriginKeyHashed: der client-kontrollierte,
// unbegrenzt lange Origin-Header landet nur gehasht im Redis-Key
// (Key-Längen-Bounding, Plan §4.1); project/ip gehen raw.
func TestRedisTokenBucket_OriginKeyHashed(t *testing.T) {
	t.Parallel()
	s, client := startMiniredis(t)
	ctx := context.Background()
	clock := newFakeClock(testBase())
	l := mkRedisLimiter(t, client, ratelimit.RedisTokenBucketConfig{
		Capacity: 5, RefillPerSecond: 0, Now: clock.Now,
	})
	longOrigin := "https://" + strings.Repeat("very-long-subdomain.", 50) + "example.com"
	if err := l.Allow(ctx, driven.RateLimitKey{ProjectID: "p", Origin: longOrigin}, 1); err != nil {
		t.Fatalf("allow: %v", err)
	}
	var originKey string
	for _, k := range s.Keys() {
		if strings.Contains(k, longOrigin) {
			t.Fatalf("raw origin must not appear in redis keys: %s", k)
		}
		if strings.HasPrefix(k, "mtrace:ingest:origin:") {
			originKey = k
		}
	}
	if originKey == "" {
		t.Fatal("origin bucket key missing")
	}
	if suffix := strings.TrimPrefix(originKey, "mtrace:ingest:origin:"); len(suffix) != 32 {
		t.Fatalf("origin key suffix must be a 32-hex-char hash, got %q", suffix)
	}
}

// warnCountHandler zählt WARN-Records — Beleg für den „kein
// False-Outage"-Vertrag des Degraded-Signals.
type warnCountHandler struct{ warns *int32 }

func (h warnCountHandler) Enabled(context.Context, slog.Level) bool { return true }
func (h warnCountHandler) Handle(_ context.Context, r slog.Record) error {
	if r.Level == slog.LevelWarn {
		atomic.AddInt32(h.warns, 1)
	}
	return nil
}
func (h warnCountHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h warnCountHandler) WithGroup(string) slog.Handler      { return h }

// TestRedisTokenBucket_ContextCancelIsNotAnOutage: Client-Abbrüche bei
// GESUNDEM Redis dürfen weder Outage- noch Recovery-WARNs erzeugen —
// sonst verbraucht Cancellation-Rauschen die „log once"-Kante eines
// echten Ausfalls und das Degraded-Signal wird wertlos.
func TestRedisTokenBucket_ContextCancelIsNotAnOutage(t *testing.T) {
	t.Parallel()
	_, client := startMiniredis(t)
	clock := newFakeClock(testBase())
	var warns int32
	logger := slog.New(warnCountHandler{warns: &warns})
	l, err := ratelimit.NewRedisTokenBucketRateLimiter(client, ratelimit.RedisTokenBucketConfig{
		Capacity: 10, RefillPerSecond: 0, Now: clock.Now,
	}, logger)
	if err != nil {
		t.Fatalf("limiter: %v", err)
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	key := driven.RateLimitKey{ProjectID: "cancel-noise"}

	for i := 0; i < 3; i++ {
		if err := l.Allow(canceled, key, 1); err != nil {
			t.Fatalf("canceled ctx call %d: fallback must decide, got %v", i, err)
		}
	}
	if err := l.Allow(context.Background(), key, 1); err != nil {
		t.Fatalf("healthy call: %v", err)
	}
	if got := atomic.LoadInt32(&warns); got != 0 {
		t.Fatalf("context cancellations produced %d WARN(s); want 0 (no false outage/recovery)", got)
	}
}
