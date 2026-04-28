package ratelimit_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/m-trace/apps/api/adapters/driven/ratelimit"
	"github.com/example/m-trace/apps/api/hexagon/domain"
)

func TestTokenBucket_AllowsUpToCapacity(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC)
	rl := ratelimit.NewTokenBucketRateLimiter(100, 100, func() time.Time { return now })

	ctx := context.Background()
	for i := 0; i < 100; i++ {
		if err := rl.Allow(ctx, "demo", 1); err != nil {
			t.Fatalf("event %d denied: %v", i, err)
		}
	}
	if err := rl.Allow(ctx, "demo", 1); !errors.Is(err, domain.ErrRateLimited) {
		t.Errorf("expected ErrRateLimited on 101st event, got %v", err)
	}
}

func TestTokenBucket_RefillsOverTime(t *testing.T) {
	t.Parallel()
	clock := time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC)
	rl := ratelimit.NewTokenBucketRateLimiter(100, 100, func() time.Time { return clock })

	ctx := context.Background()
	if err := rl.Allow(ctx, "demo", 100); err != nil {
		t.Fatalf("first 100 denied: %v", err)
	}
	if err := rl.Allow(ctx, "demo", 1); !errors.Is(err, domain.ErrRateLimited) {
		t.Errorf("expected ErrRateLimited immediately after exhaustion, got %v", err)
	}

	// Advance the clock by 1 second; bucket should refill to capacity.
	clock = clock.Add(1 * time.Second)
	if err := rl.Allow(ctx, "demo", 100); err != nil {
		t.Errorf("expected refilled bucket to allow 100 events, got %v", err)
	}
}

func TestTokenBucket_PerProjectIsolation(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC)
	rl := ratelimit.NewTokenBucketRateLimiter(100, 100, func() time.Time { return now })

	ctx := context.Background()
	if err := rl.Allow(ctx, "demo", 100); err != nil {
		t.Fatalf("demo denied: %v", err)
	}
	if err := rl.Allow(ctx, "other", 100); err != nil {
		t.Errorf("other should have its own bucket; got %v", err)
	}
}

func TestTokenBucket_RejectsLargeBatch(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC)
	rl := ratelimit.NewTokenBucketRateLimiter(100, 100, func() time.Time { return now })

	if err := rl.Allow(context.Background(), "demo", 101); !errors.Is(err, domain.ErrRateLimited) {
		t.Errorf("expected ErrRateLimited for batch larger than capacity, got %v", err)
	}
}
