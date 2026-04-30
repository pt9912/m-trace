package driven

// MetricsPublisher exposes aggregate API and playback metrics.
// Implementations live in adapters/driven/metrics. Calls must be safe
// for concurrent use and must not add high-cardinality labels such as
// session_id, user_agent, segment_url, or client_ip.
type MetricsPublisher interface {
	EventsAccepted(n int)
	InvalidEvents(n int)
	RateLimitedEvents(n int)
	DroppedEvents(n int)
	PlaybackErrors(n int)
	RebufferEvents(n int)
	StartupTimeMS(ms float64)
}
