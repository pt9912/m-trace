package driven

import "context"

// RateLimiter consumes n event-tokens for the given project_id.
// Returns domain.ErrRateLimited if the budget is exhausted; nil
// otherwise. The spike implementation is an in-memory token bucket;
// distributed rate limiting is out of scope (Spec §6.9).
type RateLimiter interface {
	Allow(ctx context.Context, projectID string, n int) error
}
