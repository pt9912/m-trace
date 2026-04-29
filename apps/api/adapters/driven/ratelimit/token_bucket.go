// Package ratelimit holds in-memory rate limiters for the spike. Per
// Spec §6.9 plus plan-0.1.0.md §5.1 (F-110): drei Dimensionen
// (project_id, client_ip, origin) als unabhängige Token-Buckets;
// distributed rate limiting bleibt out of scope.
package ratelimit

import (
	"context"
	"sync"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// dimension benennt die drei Bucket-Klassen, damit gleiche Werte aus
// verschiedenen Dimensionen sich nicht im Map-Key kollidieren.
type dimension string

const (
	dimProject  dimension = "project_id"
	dimClientIP dimension = "client_ip"
	dimOrigin   dimension = "origin"
)

type bucketKey struct {
	dim dimension
	val string
}

type bucket struct {
	tokens     float64
	lastRefill time.Time
}

// TokenBucketRateLimiter erzwingt ein Event-Budget pro Dimension
// (project_id / client_ip / origin). Jede Dimension teilt sich
// Capacity und Refill-Rate; eine fail-fast „all-or-nothing"-Semantik
// stellt sicher, dass ein 429 in einer Dimension keine Tokens in den
// anderen Dimensionen verbraucht (plan-0.1.0.md §5.1).
type TokenBucketRateLimiter struct {
	capacity   int
	refillRate float64
	now        func() time.Time

	mu      sync.Mutex
	buckets map[bucketKey]*bucket
}

// NewTokenBucketRateLimiter builds a limiter mit capacity Tokens und
// refillRate Tokens/sec pro Dimension. If now is nil, time.Now is used.
func NewTokenBucketRateLimiter(capacity int, refillRate float64, now func() time.Time) *TokenBucketRateLimiter {
	if now == nil {
		now = time.Now
	}
	return &TokenBucketRateLimiter{
		capacity:   capacity,
		refillRate: refillRate,
		now:        now,
		buckets:    make(map[bucketKey]*bucket),
	}
}

// Allow konsumiert n Tokens aus jeder gesetzten Dimension von key.
// Leere Dimensions-Werte werden übersprungen — ein CLI/curl-Pfad ohne
// Origin verbraucht z. B. nur project- und client-IP-Tokens.
//
// All-or-nothing: schlägt eine Dimension fehl, werden keine Tokens in
// den anderen verbraucht — der Test §5.1 zur 403-Side-Effect-Variante
// hängt von dieser Garantie ab.
func (l *TokenBucketRateLimiter) Allow(_ context.Context, key driven.RateLimitKey, n int) error {
	if n <= 0 {
		return nil
	}

	keys := buildKeys(key)
	if len(keys) == 0 {
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	for _, k := range keys {
		b := l.refillLocked(k, now)
		if b.tokens < float64(n) {
			return domain.ErrRateLimited
		}
	}
	for _, k := range keys {
		l.buckets[k].tokens -= float64(n)
	}
	return nil
}

// refillLocked refresht den Bucket für k zum Zeitpunkt now und legt
// ihn an, falls er noch nicht existiert. Caller hält l.mu.
func (l *TokenBucketRateLimiter) refillLocked(k bucketKey, now time.Time) *bucket {
	b, ok := l.buckets[k]
	if !ok {
		b = &bucket{
			tokens:     float64(l.capacity),
			lastRefill: now,
		}
		l.buckets[k] = b
		return b
	}
	elapsed := now.Sub(b.lastRefill).Seconds()
	if elapsed > 0 {
		b.tokens += elapsed * l.refillRate
		if b.tokens > float64(l.capacity) {
			b.tokens = float64(l.capacity)
		}
		b.lastRefill = now
	}
	return b
}

// buildKeys filtert die nicht-leeren Dimensionen von key heraus.
func buildKeys(key driven.RateLimitKey) []bucketKey {
	out := make([]bucketKey, 0, 3)
	if key.ProjectID != "" {
		out = append(out, bucketKey{dim: dimProject, val: key.ProjectID})
	}
	if key.ClientIP != "" {
		out = append(out, bucketKey{dim: dimClientIP, val: key.ClientIP})
	}
	if key.Origin != "" {
		out = append(out, bucketKey{dim: dimOrigin, val: key.Origin})
	}
	return out
}

var _ driven.RateLimiter = (*TokenBucketRateLimiter)(nil)
