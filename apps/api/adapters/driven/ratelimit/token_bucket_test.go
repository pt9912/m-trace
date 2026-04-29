package ratelimit_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/ratelimit"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

func keyFor(projectID string) driven.RateLimitKey {
	return driven.RateLimitKey{ProjectID: projectID}
}

func TestTokenBucket_AllowsUpToCapacity(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC)
	rl := ratelimit.NewTokenBucketRateLimiter(100, 100, func() time.Time { return now })

	ctx := context.Background()
	for i := 0; i < 100; i++ {
		if err := rl.Allow(ctx, keyFor("demo"), 1); err != nil {
			t.Fatalf("event %d denied: %v", i, err)
		}
	}
	if err := rl.Allow(ctx, keyFor("demo"), 1); !errors.Is(err, domain.ErrRateLimited) {
		t.Errorf("expected ErrRateLimited on 101st event, got %v", err)
	}
}

func TestTokenBucket_RefillsOverTime(t *testing.T) {
	t.Parallel()
	clock := time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC)
	rl := ratelimit.NewTokenBucketRateLimiter(100, 100, func() time.Time { return clock })

	ctx := context.Background()
	if err := rl.Allow(ctx, keyFor("demo"), 100); err != nil {
		t.Fatalf("first 100 denied: %v", err)
	}
	if err := rl.Allow(ctx, keyFor("demo"), 1); !errors.Is(err, domain.ErrRateLimited) {
		t.Errorf("expected ErrRateLimited immediately after exhaustion, got %v", err)
	}

	// Advance the clock by 1 second; bucket should refill to capacity.
	clock = clock.Add(1 * time.Second)
	if err := rl.Allow(ctx, keyFor("demo"), 100); err != nil {
		t.Errorf("expected refilled bucket to allow 100 events, got %v", err)
	}
}

func TestTokenBucket_PerProjectIsolation(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC)
	rl := ratelimit.NewTokenBucketRateLimiter(100, 100, func() time.Time { return now })

	ctx := context.Background()
	if err := rl.Allow(ctx, keyFor("demo"), 100); err != nil {
		t.Fatalf("demo denied: %v", err)
	}
	if err := rl.Allow(ctx, keyFor("other"), 100); err != nil {
		t.Errorf("other should have its own bucket; got %v", err)
	}
}

func TestTokenBucket_RejectsLargeBatch(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC)
	rl := ratelimit.NewTokenBucketRateLimiter(100, 100, func() time.Time { return now })

	if err := rl.Allow(context.Background(), keyFor("demo"), 101); !errors.Is(err, domain.ErrRateLimited) {
		t.Errorf("expected ErrRateLimited for batch larger than capacity, got %v", err)
	}
}

// TestTokenBucket_PerClientIPIsolation verifiziert, dass Project- und
// Client-IP-Buckets unabhängig voneinander geführt werden — zwei
// Clients dürfen jeweils ihr volles Project-Budget aufzehren.
func TestTokenBucket_PerClientIPIsolation(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC)
	rl := ratelimit.NewTokenBucketRateLimiter(100, 100, func() time.Time { return now })

	ctx := context.Background()
	keyA := driven.RateLimitKey{ProjectID: "demo", ClientIP: "10.0.0.1"}
	keyB := driven.RateLimitKey{ProjectID: "demo", ClientIP: "10.0.0.2"}

	// Erste 100 für Client A leeren den Project- und den IP-A-Bucket.
	if err := rl.Allow(ctx, keyA, 100); err != nil {
		t.Fatalf("A first 100: %v", err)
	}
	// Project-Bucket leer → 429, Client-IP-Bucket B unberührt.
	if err := rl.Allow(ctx, keyB, 1); !errors.Is(err, domain.ErrRateLimited) {
		t.Errorf("expected ErrRateLimited on Project-shared dimension, got %v", err)
	}
}

// TestTokenBucket_AllOrNothing_CommitOnly verifiziert, dass ein 429
// in einer Dimension keine Tokens in den anderen Dimensionen
// verbraucht. Setup: zwei Requests mit unterschiedlichen Projects, aber
// derselben Client-IP; nach 100 Events von Project A gegen Client-X ist
// die Client-IP-Dimension erschöpft. Ein Folge-Request für Project B
// (gleiche IP) muss fehlschlagen, ohne den Project-B-Bucket zu
// dezimieren — Folge-Requests gegen Project B von einer anderen IP
// dürfen weiterhin durchgehen.
func TestTokenBucket_AllOrNothing_CommitOnly(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC)
	rl := ratelimit.NewTokenBucketRateLimiter(100, 100, func() time.Time { return now })

	ctx := context.Background()

	// Project A + IP-X: erste 100 leeren beide Buckets.
	keyAX := driven.RateLimitKey{ProjectID: "A", ClientIP: "X"}
	if err := rl.Allow(ctx, keyAX, 100); err != nil {
		t.Fatalf("AX 100: %v", err)
	}

	// Project B + IP-X: IP-X ist erschöpft → 429. Aber Project-B-Bucket
	// darf kein Token verloren haben.
	keyBX := driven.RateLimitKey{ProjectID: "B", ClientIP: "X"}
	if err := rl.Allow(ctx, keyBX, 1); !errors.Is(err, domain.ErrRateLimited) {
		t.Fatalf("BX expected 429, got %v", err)
	}

	// Project B + IP-Y: muss durchgehen — wenn der Project-B-Bucket
	// fälschlich beim BX-Versuch Tokens verloren hätte, würde das hier
	// brechen.
	keyBY := driven.RateLimitKey{ProjectID: "B", ClientIP: "Y"}
	if err := rl.Allow(ctx, keyBY, 100); err != nil {
		t.Errorf("BY 100 should pass (B-Bucket untouched), got %v", err)
	}
}

// TestTokenBucket_EmptyKeyIsNoop deckt den Fall ab, dass alle drei
// Dimensionen leer sind — der Limiter macht nichts und gibt nil
// zurück. Verteidigt gegen Misskonfiguration.
func TestTokenBucket_EmptyKeyIsNoop(t *testing.T) {
	t.Parallel()
	rl := ratelimit.NewTokenBucketRateLimiter(100, 100, time.Now)
	if err := rl.Allow(context.Background(), driven.RateLimitKey{}, 1000); err != nil {
		t.Errorf("empty key with n=1000 should be noop, got %v", err)
	}
}
