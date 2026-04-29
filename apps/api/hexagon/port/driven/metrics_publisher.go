package driven

// MetricsPublisher exposes the four mandatory counters from
// docs/spike/backend-api-contract.md §7. Implementations live in
// adapters/driven/metrics. Calls must be safe for concurrent use.
type MetricsPublisher interface {
	EventsAccepted(n int)
	InvalidEvents(n int)
	RateLimitedEvents(n int)
	DroppedEvents(n int)
}
