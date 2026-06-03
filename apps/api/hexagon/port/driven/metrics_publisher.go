package driven

import "github.com/pt9912/m-trace/apps/api/hexagon/domain"

// MetricsPublisher exposes aggregate API and playback metrics.
// Implementations live in adapters/driven/metrics. Calls must be safe
// for concurrent use and must not add high-cardinality labels such as
// session_id, user_agent, segment_url, or client_ip.
//
//nolint:interfacebloat // Aggregat-Metrik-Vertrag aus spec/lastenheft.md: vier Pflichtcounter + drei SRT-Health- + ein WebRTC-Sample-Aufruf bündeln den gesamten driven-Port.
type MetricsPublisher interface {
	EventsAccepted(n int)
	InvalidEvents(n int)
	RateLimitedEvents(n int)
	DroppedEvents(n int)
	PlaybackErrors(n int)
	RebufferEvents(n int)
	StartupTimeMS(ms float64)

	// SRT-Health-Aggregate ( Sub-3.6,
	// spec/telemetry-model.md). Wertebereiche der Labels
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

	// WebRTC-Aggregate (
	// spec/telemetry-model.md). Wertebereiche der Labels kommen
	// aus den W3C-Enums und sind in §3.2 als bounded Aggregat-Labels
	// freigegeben. State-Counter zählen Samples (nicht Gauges);
	// Counter-Felder werden serverseitig deltadiffenziert über den
	// (project_id, session_id, peer_connection_run_id, metric)-Schlüssel
	// (siehe §3.5.1). Aufruf nach erfolgreicher Validation des
	// `metrics_sampled`-Events mit reservierten webrtc.*-Keys.
	WebRTCSample(sample WebRTCSampleSnapshot)

	// SampleRateDrift incrementiert
	// `mtrace_sample_rate_drift_total{project_id}` (
	// / R-10, spec/telemetry-model.md). Aufruf nur,
	// wenn ein `meta.session_sample_rate`-Wert eingegangen ist, der
	// vom bereits persistierten Wert um mehr als die Toleranzschwelle
	// (`SampleRateDriftToleranceP_PM` = 100 ppm) abweicht.
	// Cardinality-Profil: `project_id` ist explizit als bounded-Label
	// freigegeben (Operator-konfigurierte Allowlist; siehe §3 / §7.6).
	SampleRateDrift(projectID string)
}

// WebRTCSampleSnapshot transportiert die §3.5.1-Sample-Daten zum
// Metriken-Adapter. Alle Felder kommen aus den validierten
// webrtc.*-Meta-Keys; Domain-Validation ist in der Application-Layer
// abgeschlossen.
type WebRTCSampleSnapshot struct {
	ProjectID       string
	SessionID       string
	RunID           string
	SampleID        int64
	ConnectionState string
	IceState        string
	DtlsState       string
	PacketsLost     int64
	BytesReceived   int64
	BytesSent       int64
}
