// Package metrics holds the Prometheus exposition adapter for the
// aggregate mtrace_* metrics from spec/lastenheft.md §7.9 and
// spec/backend-api-contract.md §7.
package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// ActiveSessionsFunc returns the current number of active sessions.
// It must not inspect or emit per-session labels.
type ActiveSessionsFunc func() float64

type publisherConfig struct {
	activeSessions ActiveSessionsFunc
}

// PublisherOption configures optional metric sources.
type PublisherOption func(*publisherConfig)

// WithActiveSessionsFunc wires the active-session gauge to the current
// session repository state. nil leaves the gauge at 0.
func WithActiveSessionsFunc(fn ActiveSessionsFunc) PublisherOption {
	return func(cfg *publisherConfig) {
		cfg.activeSessions = fn
	}
}

// PrometheusPublisher exposes the aggregate mtrace_* counters/gauges
// and returns the HTTP handler for GET /api/metrics.
//
// No high-cardinality labels are emitted (Spec §7).
type PrometheusPublisher struct {
	registry *prometheus.Registry

	eventsAccepted    prometheus.Counter
	invalidEvents     prometheus.Counter
	rateLimitedEvents prometheus.Counter
	droppedEvents     prometheus.Counter
	playbackErrors    prometheus.Counter
	rebufferEvents    prometheus.Counter
	apiRequests       prometheus.Counter
	activeSessions    prometheus.GaugeFunc
	startupTimeMS     prometheus.Gauge
	analyzeRequests   *prometheus.CounterVec

	// SRT-Health-Aggregate (plan-0.6.0 §4 Sub-3.6,
	// spec/telemetry-model.md §7.7). Bounded Labels; Wertebereiche
	// kommen aus den Domain-Enums.
	srtHealthSamples         *prometheus.CounterVec
	srtCollectorRuns         *prometheus.CounterVec
	srtCollectorErrors       *prometheus.CounterVec
}

// NewPrometheusPublisher creates and registers the aggregate metrics.
func NewPrometheusPublisher(opts ...PublisherOption) *PrometheusPublisher {
	registry := prometheus.NewRegistry()
	cfg := publisherConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}
	activeSessions := cfg.activeSessions
	if activeSessions == nil {
		activeSessions = func() float64 { return 0 }
	}

	p := &PrometheusPublisher{
		registry: registry,
		eventsAccepted: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "mtrace_playback_events_total",
			Help: "Total number of accepted playback events.",
		}),
		invalidEvents: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "mtrace_invalid_events_total",
			Help: "Total number of events rejected for schema or validation reasons (HTTP 400 / 422).",
		}),
		rateLimitedEvents: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "mtrace_rate_limited_events_total",
			Help: "Total number of events rejected by the rate limiter (HTTP 429).",
		}),
		droppedEvents: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "mtrace_dropped_events_total",
			Help: "Total number of events dropped internally (e.g. on persistence backpressure).",
		}),
		playbackErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "mtrace_playback_errors_total",
			Help: "Total number of accepted playback error events.",
		}),
		rebufferEvents: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "mtrace_rebuffer_events_total",
			Help: "Total number of accepted rebuffer start events.",
		}),
		apiRequests: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "mtrace_api_requests_total",
			Help: "Total number of HTTP requests handled by the API.",
		}),
		activeSessions: prometheus.NewGaugeFunc(prometheus.GaugeOpts{
			Name: "mtrace_active_sessions",
			Help: "Current number of active stream sessions.",
		}, activeSessions),
		startupTimeMS: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "mtrace_startup_time_ms",
			Help: "Most recently observed player startup time in milliseconds.",
		}),
		// `outcome` ist {ok, error}, `code` aus der abgeschlossenen
		// Domäne (`invalid_request`, `analyzer_unavailable`, plus die
		// AnalysisErrorCode-Werte aus dem stream-analyzer-Vertrag).
		// Cardinality bleibt damit beschränkt (plan-0.3.0 §9 Tranche
		// 7.5; Spec §7 — keine Cardinality-Explosion).
		analyzeRequests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "mtrace_analyze_requests_total",
				Help: "Total number of POST /api/analyze invocations grouped by outcome and result/error code.",
			},
			[]string{"outcome", "code"},
		),
	}
	srtSamples, srtRuns, srtErrors := newSrtHealthCounters()
	p.srtHealthSamples = srtSamples
	p.srtCollectorRuns = srtRuns
	p.srtCollectorErrors = srtErrors
	registry.MustRegister(
		p.eventsAccepted,
		p.invalidEvents,
		p.rateLimitedEvents,
		p.droppedEvents,
		p.playbackErrors,
		p.rebufferEvents,
		p.apiRequests,
		p.activeSessions,
		p.startupTimeMS,
		p.analyzeRequests,
		p.srtHealthSamples,
		p.srtCollectorRuns,
		p.srtCollectorErrors,
	)
	return p
}

// EventsAccepted increments the accepted counter by n.
func (p *PrometheusPublisher) EventsAccepted(n int) {
	if n > 0 {
		p.eventsAccepted.Add(float64(n))
	}
}

// InvalidEvents increments the invalid counter by n. Allows n == 0
// for empty-batch rejections so call sites can stay uniform.
func (p *PrometheusPublisher) InvalidEvents(n int) {
	if n > 0 {
		p.invalidEvents.Add(float64(n))
	}
}

// RateLimitedEvents increments the rate-limited counter by n.
func (p *PrometheusPublisher) RateLimitedEvents(n int) {
	if n > 0 {
		p.rateLimitedEvents.Add(float64(n))
	}
}

// DroppedEvents increments the dropped counter by n.
func (p *PrometheusPublisher) DroppedEvents(n int) {
	if n > 0 {
		p.droppedEvents.Add(float64(n))
	}
}

// PlaybackErrors increments the accepted playback-error counter by n.
func (p *PrometheusPublisher) PlaybackErrors(n int) {
	if n > 0 {
		p.playbackErrors.Add(float64(n))
	}
}

// RebufferEvents increments the accepted rebuffer-start counter by n.
func (p *PrometheusPublisher) RebufferEvents(n int) {
	if n > 0 {
		p.rebufferEvents.Add(float64(n))
	}
}

// StartupTimeMS records the latest accepted startup duration.
func (p *PrometheusPublisher) StartupTimeMS(ms float64) {
	if ms >= 0 {
		p.startupTimeMS.Set(ms)
	}
}

// APIRequests increments the API request counter by n.
func (p *PrometheusPublisher) APIRequests(n int) {
	if n > 0 {
		p.apiRequests.Add(float64(n))
	}
}

// AnalyzeRequest erhöht den Counter um 1 für eine abgeschlossene
// Analyse-Anfrage (POST /api/analyze). `outcome` ist "ok" oder
// "error"; `code` ist entweder "ok" (bei outcome="ok") oder der
// fachliche Fehler-Code aus der Domäne (siehe knownAnalyzeCodes).
// Unbekannte Werte werden auf "_unknown" gemappt — Cardinality-
// Defense-in-Depth, falls je ein Aufrufer einen unklassifizierten
// Code übergibt (plan-0.3.0 §9 Tranche 7.5/1; Spec §7).
func (p *PrometheusPublisher) AnalyzeRequest(outcome, code string) {
	p.analyzeRequests.WithLabelValues(normalizeOutcome(outcome), normalizeAnalyzeCode(code)).Inc()
}

// normalizeOutcome bildet einen Outcome-Wert auf die abgeschlossene
// Domäne {"ok","error"} ab; alles andere fällt auf "_unknown" (Cardinality-
// Defense-in-Depth, falls je ein Aufrufer einen unklassifizierten Wert
// übergibt — Spec §7).
func normalizeOutcome(value string) string {
	switch value {
	case "ok", "error":
		return value
	default:
		return "_unknown"
	}
}

// normalizeAnalyzeCode bildet einen Code-Wert auf die abgeschlossene
// Code-Domäne von `mtrace_analyze_requests_total` ab (API-Eingabe-Codes
// + alle `domain.StreamAnalysisErrorCode`-Werte + `analyzer_unavailable`
// als Transport-Fall); alles andere fällt auf "_unknown".
func normalizeAnalyzeCode(value string) string {
	switch value {
	case "ok",
		"invalid_request",
		"invalid_json",
		"unsupported_media_type",
		"payload_too_large",
		"invalid_input",
		"manifest_not_hls",
		"fetch_blocked",
		"fetch_failed",
		"manifest_too_large",
		"internal_error",
		"analyzer_unavailable":
		return value
	default:
		return "_unknown"
	}
}

// newSrtHealthCounters konstruiert die drei SRT-Health-Aggregate
// (plan-0.6.0 §4 Sub-3.6 — spec/telemetry-model.md §7.7). Werte-
// bereiche je Label kommen aus den Domain-Enums; §3.2 hat sie als
// bounded Aggregat-Labels freigegeben.
func newSrtHealthCounters() (samples, runs, errs *prometheus.CounterVec) {
	samples = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mtrace_srt_health_samples_total",
			Help: "Total number of SRT health samples persisted, by health_state.",
		},
		[]string{"health_state"},
	)
	runs = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mtrace_srt_health_collector_runs_total",
			Help: "Total number of SRT health collector runs, by source_status.",
		},
		[]string{"source_status"},
	)
	errs = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mtrace_srt_health_collector_errors_total",
			Help: "Total number of SRT health collector errors, by source_error_code.",
		},
		[]string{"source_error_code"},
	)
	return samples, runs, errs
}

// Handler returns the HTTP handler for GET /api/metrics.
func (p *PrometheusPublisher) Handler() http.Handler {
	return promhttp.HandlerFor(p.registry, promhttp.HandlerOpts{})
}

// SrtHealthSampleAccepted incrementiert den Sample-Counter für das
// gegebene HealthState-Label. Unbekannte Domain-Werte werden via
// Allowlist abgefangen (Cardinality-Defense-in-Depth, Spec §7.7).
func (p *PrometheusPublisher) SrtHealthSampleAccepted(state domain.HealthState) {
	p.srtHealthSamples.WithLabelValues(normalizeHealthState(state)).Inc()
}

// SrtCollectorRun incrementiert den Run-Counter für das gegebene
// SourceStatus-Label.
func (p *PrometheusPublisher) SrtCollectorRun(status domain.SourceStatus) {
	p.srtCollectorRuns.WithLabelValues(normalizeSourceStatus(status)).Inc()
}

// SrtCollectorError incrementiert den Error-Counter für das gegebene
// SourceErrorCode-Label. `none` ist explizit kein Fehler — der Aufrufer
// (Collector) ruft die Methode nur bei nicht-`none` Werten.
func (p *PrometheusPublisher) SrtCollectorError(code domain.SourceErrorCode) {
	p.srtCollectorErrors.WithLabelValues(normalizeSourceErrorCode(code)).Inc()
}

func normalizeHealthState(state domain.HealthState) string {
	switch state {
	case domain.HealthStateHealthy, domain.HealthStateDegraded,
		domain.HealthStateCritical, domain.HealthStateUnknown:
		return string(state)
	default:
		return "_unknown"
	}
}

func normalizeSourceStatus(status domain.SourceStatus) string {
	switch status {
	case domain.SourceStatusOK, domain.SourceStatusUnavailable,
		domain.SourceStatusPartial, domain.SourceStatusStale,
		domain.SourceStatusNoActiveConnection:
		return string(status)
	default:
		return "_unknown"
	}
}

func normalizeSourceErrorCode(code domain.SourceErrorCode) string {
	switch code {
	case domain.SourceErrorCodeNone, domain.SourceErrorCodeSourceUnavailable,
		domain.SourceErrorCodeNoActiveConnection, domain.SourceErrorCodePartialSample,
		domain.SourceErrorCodeStaleSample, domain.SourceErrorCodeParseError:
		return string(code)
	default:
		return "_unknown"
	}
}

var _ driven.MetricsPublisher = (*PrometheusPublisher)(nil)
