package auth_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/auth"
)

// TestRedisOrigin_HappyPath_CrossInstanceSharing (plan-0.12.6
// Tranche 7 / R-22-Resttrigger): zwei Limiter-Instanzen teilen das
// Bucket über Redis.
func TestRedisOrigin_HappyPath_CrossInstanceSharing(t *testing.T) {
	t.Parallel()
	_, client := startMiniredis(t)
	ctx := context.Background()

	mk := func() *auth.RedisOriginRateLimiter {
		l, err := auth.NewRedisOriginRateLimiter(client, auth.RedisOriginLimiterConfig{
			Capacity:        2,
			RefillPerSecond: 0,
			TTLSeconds:      60,
			KeyPrefix:       "test:origin",
		}, nil)
		if err != nil {
			t.Fatalf("limiter: %v", err)
		}
		return l
	}
	a := mk()
	b := mk()

	if ok, _ := a.Allow(ctx, "ip:1.2.3.4"); !ok {
		t.Fatalf("A #1 should be allowed")
	}
	if ok, _ := b.Allow(ctx, "ip:1.2.3.4"); !ok {
		t.Fatalf("B #2 should be allowed (shared bucket has cap=2)")
	}
	// 3. Call von beiden Seiten: leerer Bucket.
	if ok, _ := a.Allow(ctx, "ip:1.2.3.4"); ok {
		t.Errorf("A #3 should be denied (shared bucket empty)")
	}
	if ok, _ := b.Allow(ctx, "ip:1.2.3.4"); ok {
		t.Errorf("B #3 should be denied (shared bucket empty)")
	}
}

// TestRedisOrigin_EmptyKey: leerer Key ist No-Op.
func TestRedisOrigin_EmptyKey(t *testing.T) {
	t.Parallel()
	_, client := startMiniredis(t)
	l, err := auth.NewRedisOriginRateLimiter(client, auth.RedisOriginLimiterConfig{
		Capacity: 1, RefillPerSecond: 0, TTLSeconds: 60,
		KeyPrefix: "test:origin-empty",
	}, nil)
	if err != nil {
		t.Fatalf("limiter: %v", err)
	}
	for i := 0; i < 5; i++ {
		if ok, err := l.Allow(context.Background(), ""); err != nil || !ok {
			t.Errorf("call %d: ok=%v err=%v, want true/nil", i, ok, err)
		}
	}
}

// TestRedisOrigin_FailClosedOnOutage: ohne FailOpen denies bei Outage.
func TestRedisOrigin_FailClosedOnOutage(t *testing.T) {
	t.Parallel()
	s, client := startMiniredis(t)
	l, err := auth.NewRedisOriginRateLimiter(client, auth.RedisOriginLimiterConfig{
		Capacity: 5, RefillPerSecond: 0, TTLSeconds: 60,
		KeyPrefix: "test:origin-closed",
	}, nil)
	if err != nil {
		t.Fatalf("limiter: %v", err)
	}
	s.Close()
	ok, err := l.Allow(context.Background(), "ip:1.2.3.4")
	if err != nil {
		t.Errorf("unexpected err: %v", err)
	}
	if ok {
		t.Errorf("fail-closed: outage allow returned true; want false")
	}
}

// TestRedisOrigin_FailOpenOnOutage: mit FailOpen fällt der Adapter
// auf das InMemory-Fallback zurück.
func TestRedisOrigin_FailOpenOnOutage(t *testing.T) {
	t.Parallel()
	s, client := startMiniredis(t)
	l, err := auth.NewRedisOriginRateLimiter(client, auth.RedisOriginLimiterConfig{
		Capacity: 5, RefillPerSecond: 0, TTLSeconds: 60,
		FailOpen:  true,
		KeyPrefix: "test:origin-open",
	}, nil)
	if err != nil {
		t.Fatalf("limiter: %v", err)
	}
	s.Close()
	ok, err := l.Allow(context.Background(), "ip:1.2.3.4")
	if err != nil {
		t.Errorf("fail-open: unexpected err: %v", err)
	}
	if !ok {
		t.Errorf("fail-open: outage allow returned false; want true (memory fallback)")
	}
}

// TestRedisOrigin_NilClient: defensive check.
func TestRedisOrigin_NilClient(t *testing.T) {
	t.Parallel()
	_, err := auth.NewRedisOriginRateLimiter(nil, auth.RedisOriginLimiterConfig{}, nil)
	if err == nil {
		t.Errorf("expected error for nil client")
	}
}

// TestRedisOrigin_ConstructorDefaults (plan-0.12.6 Tranche 7): leere
// `KeyPrefix`/`TTLSeconds`/`Now`-Felder fallen auf Defaults zurück.
func TestRedisOrigin_ConstructorDefaults(t *testing.T) {
	t.Parallel()
	_, client := startMiniredis(t)
	l, err := auth.NewRedisOriginRateLimiter(client, auth.RedisOriginLimiterConfig{
		Capacity: 3, // restliche Felder default
	}, nil)
	if err != nil {
		t.Fatalf("limiter: %v", err)
	}
	if ok, _ := l.Allow(context.Background(), "ip:1.2.3.4"); !ok {
		t.Errorf("default-config limiter denies first call; want allow")
	}
}

// TestRedisOrigin_FailOpenLogger (plan-0.12.6 Tranche 7): fail-open
// mit Logger deckt den `failModeName(failOpen=true)`-Pfad ab.
func TestRedisOrigin_FailOpenLogger(t *testing.T) {
	t.Parallel()
	s, client := startMiniredis(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	l, err := auth.NewRedisOriginRateLimiter(client, auth.RedisOriginLimiterConfig{
		Capacity: 5, TTLSeconds: 60, FailOpen: true,
		KeyPrefix: "test:origin-fail-open-logger",
	}, logger)
	if err != nil {
		t.Fatalf("limiter: %v", err)
	}
	s.Close()
	if ok, _ := l.Allow(context.Background(), "ip:1.2.3.4"); !ok {
		t.Errorf("fail-open with logger should fall back to memory; got deny")
	}
}

// TestRedisOrigin_FailClosedLogger (plan-0.12.6 Tranche 7): non-nil
// Logger deckt den `handleRedisError`-Log-Pfad ab.
func TestRedisOrigin_FailClosedLogger(t *testing.T) {
	t.Parallel()
	s, client := startMiniredis(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	l, err := auth.NewRedisOriginRateLimiter(client, auth.RedisOriginLimiterConfig{
		Capacity: 5, TTLSeconds: 60, KeyPrefix: "test:origin-fail-logger",
	}, logger)
	if err != nil {
		t.Fatalf("limiter: %v", err)
	}
	s.Close()
	if ok, _ := l.Allow(context.Background(), "ip:1.2.3.4"); ok {
		t.Errorf("expected deny on outage with logger; got allow")
	}
}

// TestRedisOrigin_ContextCanceled.
func TestRedisOrigin_ContextCanceled(t *testing.T) {
	t.Parallel()
	_, client := startMiniredis(t)
	l, err := auth.NewRedisOriginRateLimiter(client, auth.RedisOriginLimiterConfig{
		Capacity: 5, TTLSeconds: 60, KeyPrefix: "test:origin-cancel",
	}, nil)
	if err != nil {
		t.Fatalf("limiter: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = l.Allow(ctx, "ip:1.2.3.4")
	if err == nil {
		t.Errorf("expected canceled context to propagate")
	}
}
