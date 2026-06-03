package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/auth"
)

// TestInMemoryOriginRateLimiter_BucketDepletesAndRefills
// (R-22): bei Capacity=3 + Refill=1/s sind
// drei Aufrufe in Folge erlaubt; der vierte wird abgewiesen;
// nach 1.5s Wartezeit ist wieder ein Token verfügbar.
func TestInMemoryOriginRateLimiter_BucketDepletesAndRefills(t *testing.T) {
	t.Parallel()
	clock := time.Date(2026, 5, 11, 12, 0, 0, 0, time.UTC)
	l := auth.NewInMemoryOriginRateLimiter(3, 1.0,
		auth.WithInMemoryOriginRateLimiterNow(func() time.Time { return clock }))

	ctx := context.Background()
	for i := 0; i < 3; i++ {
		ok, err := l.Allow(ctx, "ip:1.2.3.4")
		if err != nil || !ok {
			t.Fatalf("call %d: allow=%v err=%v, want true/nil", i, ok, err)
		}
	}
	// 4. Aufruf: Bucket leer.
	ok, err := l.Allow(ctx, "ip:1.2.3.4")
	if err != nil {
		t.Fatalf("4th call err = %v", err)
	}
	if ok {
		t.Errorf("4th call allowed; want denial (bucket empty)")
	}

	// Nach 1.5s Wartezeit (Refill=1/s) ist wieder ≥1 Token verfügbar.
	clock = clock.Add(1500 * time.Millisecond)
	ok, err = l.Allow(ctx, "ip:1.2.3.4")
	if err != nil || !ok {
		t.Errorf("post-refill: allow=%v err=%v, want true/nil", ok, err)
	}
}

// TestInMemoryOriginRateLimiter_KeyIsolation:
// zwei verschiedene Keys (Client-IPs) haben unabhängige Buckets;
// das Auffüllen oder Leeren des einen beeinflusst den anderen nicht.
func TestInMemoryOriginRateLimiter_KeyIsolation(t *testing.T) {
	t.Parallel()
	l := auth.NewInMemoryOriginRateLimiter(2, 1.0)
	ctx := context.Background()
	// Bucket für ip:A leeren.
	for i := 0; i < 2; i++ {
		_, _ = l.Allow(ctx, "ip:A")
	}
	if ok, _ := l.Allow(ctx, "ip:A"); ok {
		t.Fatalf("ip:A bucket should be empty after 2 calls")
	}
	// ip:B ist noch unbenutzt.
	if ok, _ := l.Allow(ctx, "ip:B"); !ok {
		t.Errorf("ip:B should still have tokens; isolation broken")
	}
}

// TestInMemoryOriginRateLimiter_EmptyKeyIsNoOp ( T6):
// fehlende RemoteAddr/XFF → key="" → Allow=true (No-Op-Pfad). Sonst
// würde der Lab-Test-Pfad ohne RemoteAddr-Setup blockiert.
func TestInMemoryOriginRateLimiter_EmptyKeyIsNoOp(t *testing.T) {
	t.Parallel()
	l := auth.NewInMemoryOriginRateLimiter(1, 0.1)
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		ok, err := l.Allow(ctx, "")
		if err != nil || !ok {
			t.Fatalf("empty-key call %d: allow=%v err=%v, want true/nil", i, ok, err)
		}
	}
}

// TestInMemoryOriginRateLimiter_DisabledBucket ( T6):
// Capacity=0 und Refill=0 → kein Limit (Disabled-Bucket-Pfad).
func TestInMemoryOriginRateLimiter_DisabledBucket(t *testing.T) {
	t.Parallel()
	l := auth.NewInMemoryOriginRateLimiter(0, 0)
	ctx := context.Background()
	for i := 0; i < 100; i++ {
		ok, err := l.Allow(ctx, "ip:1.2.3.4")
		if err != nil || !ok {
			t.Fatalf("disabled-bucket call %d: allow=%v err=%v, want true/nil", i, ok, err)
		}
	}
}

// TestInMemoryOriginRateLimiter_NilReceiver ( T6):
// `nil`-Limiter (Disabled-Pfad aus dem Boot-Validator) ist Allow-No-Op.
func TestInMemoryOriginRateLimiter_NilReceiver(t *testing.T) {
	t.Parallel()
	var l *auth.InMemoryOriginRateLimiter // nil
	ok, err := l.Allow(context.Background(), "ip:1.2.3.4")
	if err != nil || !ok {
		t.Errorf("nil receiver: allow=%v err=%v, want true/nil", ok, err)
	}
}

// TestInMemoryOriginRateLimiter_EvictsIdleBuckets (
// / R-22): nach `idleTTL` (10 min Default) räumt der
// nächste Allow-Call die ungenutzten Buckets per opportunistic
// sweep. Test treibt den Clock manuell weit über die Schwelle.
func TestInMemoryOriginRateLimiter_EvictsIdleBuckets(t *testing.T) {
	t.Parallel()
	clock := time.Date(2026, 5, 11, 12, 0, 0, 0, time.UTC)
	l := auth.NewInMemoryOriginRateLimiter(5, 1.0,
		auth.WithInMemoryOriginRateLimiterNow(func() time.Time { return clock }))
	ctx := context.Background()
	// Bucket für "ip:idle" anlegen.
	if _, err := l.Allow(ctx, "ip:idle"); err != nil {
		t.Fatalf("seed: %v", err)
	}
	// 20 Minuten in die Zukunft → über sweepInterval (5 min) und
	// idleTTL (10 min) hinaus.
	clock = clock.Add(20 * time.Minute)
	// Allow für anderen Key triggert den Sweep; idle bucket wird
	// entfernt.
	if _, err := l.Allow(ctx, "ip:fresh"); err != nil {
		t.Fatalf("sweep trigger: %v", err)
	}
	// Smoke: nach Sweep+frisch ist das Verhalten kontinuierlich
	// — kein Crash, kein Side-Effekt. Innenstand-Inspection ist
	// adapter-intern, daher nur Funktional-Check.
	if _, err := l.Allow(ctx, "ip:idle"); err != nil {
		t.Errorf("post-sweep: %v", err)
	}
}

// TestInMemoryOriginRateLimiter_ContextCanceled: ein bereits
// abgebrochener Context wird durchgereicht.
func TestInMemoryOriginRateLimiter_ContextCanceled(t *testing.T) {
	t.Parallel()
	l := auth.NewInMemoryOriginRateLimiter(5, 1.0)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := l.Allow(ctx, "ip:1.2.3.4")
	if err == nil {
		t.Errorf("expected canceled context to propagate")
	}
}
