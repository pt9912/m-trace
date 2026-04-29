// Package ratelimit holds in-memory rate limiters for the spike. Per
// Spec §6.9 the spike uses a single-process token bucket; distributed
// rate limiting is out of scope.
package ratelimit

import (
	"context"
	"sync"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// TokenBucketRateLimiter enforces an event-rate budget per project.
// Each project gets its own bucket lazily on first use. Buckets refill
// continuously at refillRate tokens per second, capped at capacity.
type TokenBucketRateLimiter struct {
	capacity   int
	refillRate float64
	now        func() time.Time

	mu      sync.Mutex
	buckets map[string]*bucket
}

type bucket struct {
	tokens     float64
	lastRefill time.Time
}

// NewTokenBucketRateLimiter builds a limiter with capacity tokens and
// refillRate tokens/sec. If now is nil, time.Now is used.
func NewTokenBucketRateLimiter(capacity int, refillRate float64, now func() time.Time) *TokenBucketRateLimiter {
	if now == nil {
		now = time.Now
	}
	return &TokenBucketRateLimiter{
		capacity:   capacity,
		refillRate: refillRate,
		now:        now,
		buckets:    make(map[string]*bucket),
	}
}

// Allow consumes n tokens for the given project. Returns
// domain.ErrRateLimited if the bucket is exhausted.
func (l *TokenBucketRateLimiter) Allow(_ context.Context, projectID string, n int) error {
	if n <= 0 {
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	b, ok := l.buckets[projectID]
	if !ok {
		b = &bucket{
			tokens:     float64(l.capacity),
			lastRefill: now,
		}
		l.buckets[projectID] = b
	} else {
		elapsed := now.Sub(b.lastRefill).Seconds()
		if elapsed > 0 {
			b.tokens += elapsed * l.refillRate
			if b.tokens > float64(l.capacity) {
				b.tokens = float64(l.capacity)
			}
			b.lastRefill = now
		}
	}

	if b.tokens < float64(n) {
		return domain.ErrRateLimited
	}
	b.tokens -= float64(n)
	return nil
}

var _ driven.RateLimiter = (*TokenBucketRateLimiter)(nil)
