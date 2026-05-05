package driven

import "github.com/pt9912/m-trace/apps/api/hexagon/domain"

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

	// SRT-Health-Aggregate (plan-0.6.0 §4 Sub-3.6,
	// spec/telemetry-model.md §7.7). Wertebereiche der Labels
	// kommen aus den Domain-Enums und sind in §3.2 als bounded
	// Aggregat-Labels freigegeben.

	// SrtHealthSampleAccepted incrementiert
	// `mtrace_srt_health_samples_total{health_state}`.
	SrtHealthSampleAccepted(state domain.HealthState)

	// SrtCollectorRun incrementiert
	// `mtrace_srt_health_collector_runs_total{source_status}`.
	// Aufruf am Ende jedes Collect-Runs (auch bei Source-Fehlern).
	SrtCollectorRun(status domain.SourceStatus)

	// SrtCollectorError incrementiert
	// `mtrace_srt_health_collector_errors_total{source_error_code}`.
	// Aufruf nur bei nicht-`none` Fehlerklassen.
	SrtCollectorError(code domain.SourceErrorCode)
}
