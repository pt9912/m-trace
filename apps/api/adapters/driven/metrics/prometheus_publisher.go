// Package metrics holds the Prometheus exposition adapter for the
// four mandatory mtrace_* counters from
// docs/spike/backend-api-contract.md §7.
package metrics

import (
	"net/http"

	"github.com/example/m-trace/apps/api/hexagon/port/driven"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// PrometheusPublisher exposes the four mandatory counters
// (mtrace_playback_events_total, mtrace_invalid_events_total,
// mtrace_rate_limited_events_total, mtrace_dropped_events_total) and
// returns the HTTP handler for GET /api/metrics.
//
// No high-cardinality labels are emitted (Spec §7).
type PrometheusPublisher struct {
	registry *prometheus.Registry

	eventsAccepted     prometheus.Counter
	invalidEvents      prometheus.Counter
	rateLimitedEvents  prometheus.Counter
	droppedEvents      prometheus.Counter
}

// NewPrometheusPublisher creates and registers the four counters.
func NewPrometheusPublisher() *PrometheusPublisher {
	registry := prometheus.NewRegistry()

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
	}
	registry.MustRegister(
		p.eventsAccepted,
		p.invalidEvents,
		p.rateLimitedEvents,
		p.droppedEvents,
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

// Handler returns the HTTP handler for GET /api/metrics.
func (p *PrometheusPublisher) Handler() http.Handler {
	return promhttp.HandlerFor(p.registry, promhttp.HandlerOpts{})
}

var _ driven.MetricsPublisher = (*PrometheusPublisher)(nil)
