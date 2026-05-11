package auth_test

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/auth"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// startMiniredis startet einen In-Process-Mock-Server und konstruiert
// einen go-redis-Client, der dagegen läuft. Cleanup räumt Server +
// Client auf.
func startMiniredis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()
	s := miniredis.RunT(t)
	c := redis.NewClient(&redis.Options{Addr: s.Addr()})
	t.Cleanup(func() { _ = c.Close() })
	return s, c
}

// TestRedisIssuance_HappyPath_CrossInstanceSharing
// (plan-0.12.6 Tranche 7 / R-17): zwei Limiter-Instanzen (analog zwei
// API-Replicas) teilen sich den Bucket-Counter über den geteilten
// Redis-Mock. Replica A verbraucht 2 Tokens; Replica B sieht nur noch
// 1 Token verfügbar.
func TestRedisIssuance_HappyPath_CrossInstanceSharing(t *testing.T) {
	t.Parallel()
	_, client := startMiniredis(t)
	ctx := context.Background()

	mkLimiter := func() *auth.RedisIssuanceRateLimiter {
		l, err := auth.NewRedisIssuanceRateLimiter(client, auth.RedisIssuanceLimiterConfig{
			GlobalCapacity:      10,
			GlobalRefillPerSec:  0, // no refill — deterministic for the test
			ProjectCapacity:     3,
			ProjectRefillPerSec: 0,
			TTLSeconds:          60,
			KeyPrefix:           "test:issuance",
		}, nil)
		if err != nil {
			t.Fatalf("limiter: %v", err)
		}
		return l
	}
	a := mkLimiter()
	b := mkLimiter()

	// Replica A verbraucht 2/3 Project-Tokens.
	for i := 1; i <= 2; i++ {
		ok, err := a.Allow(ctx, "demo", domain.RateLimitBucket{})
		if err != nil || !ok {
			t.Fatalf("replica A call %d: ok=%v err=%v", i, ok, err)
		}
	}
	// Replica B verbraucht das letzte Token.
	ok, err := b.Allow(ctx, "demo", domain.RateLimitBucket{})
	if err != nil || !ok {
		t.Fatalf("replica B call 3: ok=%v err=%v (cross-instance sharing broken)", ok, err)
	}
	// 4. Aufruf von beiden Seiten muss abgelehnt werden.
	for label, l := range map[string]*auth.RedisIssuanceRateLimiter{"A": a, "B": b} {
		ok, err := l.Allow(ctx, "demo", domain.RateLimitBucket{})
		if err != nil {
			t.Errorf("replica %s call 4 err: %v", label, err)
		}
		if ok {
			t.Errorf("replica %s call 4 allowed; want deny (cross-instance bucket should be empty)", label)
		}
	}
}

// TestRedisIssuance_ProjectIsolation: verschiedene project_ids haben
// unabhängige Buckets.
func TestRedisIssuance_ProjectIsolation(t *testing.T) {
	t.Parallel()
	_, client := startMiniredis(t)
	ctx := context.Background()
	l, err := auth.NewRedisIssuanceRateLimiter(client, auth.RedisIssuanceLimiterConfig{
		GlobalCapacity:      100,
		GlobalRefillPerSec:  0,
		ProjectCapacity:     1,
		ProjectRefillPerSec: 0,
		TTLSeconds:          60,
		KeyPrefix:           "test:isolation",
	}, nil)
	if err != nil {
		t.Fatalf("limiter: %v", err)
	}
	// demo-Bucket leeren.
	if ok, _ := l.Allow(ctx, "demo", domain.RateLimitBucket{}); !ok {
		t.Fatalf("demo #1 should be allowed")
	}
	if ok, _ := l.Allow(ctx, "demo", domain.RateLimitBucket{}); ok {
		t.Fatalf("demo #2 should be denied (project bucket empty)")
	}
	// other-Bucket ist unbenutzt.
	if ok, _ := l.Allow(ctx, "other", domain.RateLimitBucket{}); !ok {
		t.Errorf("other #1 should be allowed (project isolation broken)")
	}
}

// TestRedisIssuance_GlobalRefundOnProjectDeny: Project-deny darf das
// globale Bucket nicht dekrementieren (Refund-Pfad im Lua-Script).
func TestRedisIssuance_GlobalRefundOnProjectDeny(t *testing.T) {
	t.Parallel()
	_, client := startMiniredis(t)
	ctx := context.Background()
	l, err := auth.NewRedisIssuanceRateLimiter(client, auth.RedisIssuanceLimiterConfig{
		GlobalCapacity:      2, // <-- bewusst klein, damit Refund-Wirkung sichtbar wird
		GlobalRefillPerSec:  0,
		ProjectCapacity:     1, // first project call empties the project bucket
		ProjectRefillPerSec: 0,
		TTLSeconds:          60,
		KeyPrefix:           "test:refund",
	}, nil)
	if err != nil {
		t.Fatalf("limiter: %v", err)
	}
	// 1. Aufruf: global -> 1 token left; project -> 0 left.
	if ok, _ := l.Allow(ctx, "demo", domain.RateLimitBucket{}); !ok {
		t.Fatalf("call 1 should be allowed")
	}
	// 2. Aufruf: global consume → 0 left; project consume → -1 → deny;
	//     global refund → 1 left.
	if ok, _ := l.Allow(ctx, "demo", domain.RateLimitBucket{}); ok {
		t.Fatalf("call 2 should be denied (project bucket empty)")
	}
	// 3. Aufruf auf einer ANDEREN Project-ID (frisches project bucket).
	//     Wenn das globale Bucket korrekt refunded wurde, ist noch 1
	//     Token global vorhanden und der Aufruf erlaubt.
	if ok, _ := l.Allow(ctx, "other", domain.RateLimitBucket{}); !ok {
		t.Errorf("call 3 on 'other' should be allowed (global token refunded after project deny)")
	}
}

// TestRedisIssuance_FailClosedOnOutage: nach Server-Stop muss der
// Limiter `(false, nil)` liefern (fail-closed Default).
func TestRedisIssuance_FailClosedOnOutage(t *testing.T) {
	t.Parallel()
	s, client := startMiniredis(t)
	ctx := context.Background()
	l, err := auth.NewRedisIssuanceRateLimiter(client, auth.RedisIssuanceLimiterConfig{
		GlobalCapacity:      10,
		GlobalRefillPerSec:  0,
		ProjectCapacity:     5,
		ProjectRefillPerSec: 0,
		TTLSeconds:          60,
		KeyPrefix:           "test:closed",
	}, nil)
	if err != nil {
		t.Fatalf("limiter: %v", err)
	}
	s.Close() // Simulate outage AFTER limiter creation.
	ok, err := l.Allow(ctx, "demo", domain.RateLimitBucket{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if ok {
		t.Errorf("fail-closed: outage Allow returned true; want false")
	}
}

// TestRedisIssuance_FailOpenOnOutage: mit `FailOpen=true` greift der
// In-Memory-Fallback und lässt die Aufrufe durch (bis das lokale
// Bucket leer ist).
func TestRedisIssuance_FailOpenOnOutage(t *testing.T) {
	t.Parallel()
	s, client := startMiniredis(t)
	ctx := context.Background()
	l, err := auth.NewRedisIssuanceRateLimiter(client, auth.RedisIssuanceLimiterConfig{
		GlobalCapacity:      10,
		GlobalRefillPerSec:  0,
		ProjectCapacity:     5,
		ProjectRefillPerSec: 0,
		TTLSeconds:          60,
		FailOpen:            true,
		KeyPrefix:           "test:open",
	}, nil)
	if err != nil {
		t.Fatalf("limiter: %v", err)
	}
	s.Close()
	ok, err := l.Allow(ctx, "demo", domain.RateLimitBucket{})
	if err != nil {
		t.Errorf("fail-open: unexpected error: %v", err)
	}
	if !ok {
		t.Errorf("fail-open: outage Allow returned false; want true (memory fallback)")
	}
}

// TestRedisIssuance_NilClient: defensive check.
func TestRedisIssuance_NilClient(t *testing.T) {
	t.Parallel()
	_, err := auth.NewRedisIssuanceRateLimiter(nil, auth.RedisIssuanceLimiterConfig{}, nil)
	if err == nil {
		t.Errorf("expected error for nil client")
	}
}

// TestRedisIssuance_ContextCanceled: ein bereits abgebrochener Context
// propagiert sauber.
func TestRedisIssuance_ContextCanceled(t *testing.T) {
	t.Parallel()
	_, client := startMiniredis(t)
	l, err := auth.NewRedisIssuanceRateLimiter(client, auth.RedisIssuanceLimiterConfig{
		GlobalCapacity: 5, ProjectCapacity: 5, TTLSeconds: 60,
		KeyPrefix: "test:cancel",
	}, nil)
	if err != nil {
		t.Fatalf("limiter: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = l.Allow(ctx, "demo", domain.RateLimitBucket{})
	if err == nil {
		t.Errorf("expected canceled context to propagate")
	}
}

// TestRedisIssuance_ProjectBucketOverride (plan-0.12.6 Tranche 7):
// ein nicht-leeres `RateLimitBucket`-Override überschreibt die
// Konstruktor-Default-Konfig und wird im Lua-Script benutzt.
func TestRedisIssuance_ProjectBucketOverride(t *testing.T) {
	t.Parallel()
	_, client := startMiniredis(t)
	l, err := auth.NewRedisIssuanceRateLimiter(client, auth.RedisIssuanceLimiterConfig{
		GlobalCapacity: 100, GlobalRefillPerSec: 0,
		ProjectCapacity: 5, ProjectRefillPerSec: 0,
		TTLSeconds: 60, KeyPrefix: "test:override",
	}, nil)
	if err != nil {
		t.Fatalf("limiter: %v", err)
	}
	// Override auf cap=1 → exakt 1 erlaubt, dann deny.
	override := domain.RateLimitBucket{Capacity: 1, RefillPerSecond: 0}
	if ok, _ := l.Allow(context.Background(), "demo", override); !ok {
		t.Fatalf("call 1 should be allowed (override cap=1)")
	}
	if ok, _ := l.Allow(context.Background(), "demo", override); ok {
		t.Errorf("call 2 should be denied (override cap=1 depleted)")
	}
}

// TestRedisIssuance_ConstructorDefaults (plan-0.12.6 Tranche 7):
// leere `KeyPrefix`/`TTLSeconds`/`Now`-Felder fallen auf Defaults
// zurück (Constructor-Branch-Coverage).
func TestRedisIssuance_ConstructorDefaults(t *testing.T) {
	t.Parallel()
	_, client := startMiniredis(t)
	l, err := auth.NewRedisIssuanceRateLimiter(client, auth.RedisIssuanceLimiterConfig{
		GlobalCapacity: 5, ProjectCapacity: 5,
		// KeyPrefix, TTLSeconds, Now bleiben Zero — Defaults greifen.
	}, nil)
	if err != nil {
		t.Fatalf("limiter: %v", err)
	}
	if ok, _ := l.Allow(context.Background(), "demo", domain.RateLimitBucket{}); !ok {
		t.Errorf("default-config limiter denies first call; want allow")
	}
}

// TestRedisIssuance_FailOpenLogger (plan-0.12.6 Tranche 7):
// fail-open mit Logger deckt den `failModeName(failOpen=true)`-Pfad
// im Outage-Log ab.
func TestRedisIssuance_FailOpenLogger(t *testing.T) {
	t.Parallel()
	s, client := startMiniredis(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	l, err := auth.NewRedisIssuanceRateLimiter(client, auth.RedisIssuanceLimiterConfig{
		GlobalCapacity: 5, ProjectCapacity: 5, TTLSeconds: 60,
		FailOpen: true, KeyPrefix: "test:fail-open-logger",
	}, logger)
	if err != nil {
		t.Fatalf("limiter: %v", err)
	}
	s.Close()
	if ok, _ := l.Allow(context.Background(), "demo", domain.RateLimitBucket{}); !ok {
		t.Errorf("fail-open with logger should fall back to memory; got deny")
	}
}

// TestRedisIssuance_FailClosedLogger (plan-0.12.6 Tranche 7):
// non-nil Logger fängt die Outage-Warnung; deckt `handleRedisError`-
// Logger-Pfad ab.
func TestRedisIssuance_FailClosedLogger(t *testing.T) {
	t.Parallel()
	s, client := startMiniredis(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	l, err := auth.NewRedisIssuanceRateLimiter(client, auth.RedisIssuanceLimiterConfig{
		GlobalCapacity: 5, ProjectCapacity: 5, TTLSeconds: 60,
		KeyPrefix: "test:fail-logger",
		Now:       func() time.Time { return time.Now() },
	}, logger)
	if err != nil {
		t.Fatalf("limiter: %v", err)
	}
	s.Close()
	if ok, _ := l.Allow(context.Background(), "demo", domain.RateLimitBucket{}); ok {
		t.Errorf("expected deny on outage with logger; got allow")
	}
}
