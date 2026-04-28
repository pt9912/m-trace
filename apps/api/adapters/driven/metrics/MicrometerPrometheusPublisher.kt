// Prometheus exposition for the four mandatory mtrace_* counters
// from docs/spike/backend-api-contract.md §7. The Prometheus scrape
// endpoint is provided by Micronaut's built-in /api/metrics
// management endpoint (see resources/application.yml); this adapter
// just owns the counters and registers them with Micrometer.
package dev.mtrace.api.adapters.driven.metrics

import dev.mtrace.api.hexagon.port.driven.MetricsPublisher
import io.micrometer.core.instrument.Counter
import io.micrometer.core.instrument.MeterRegistry
import io.micronaut.context.annotation.Context

// @Context (eager singleton) ensures the four mtrace_* counters are
// registered with the MeterRegistry at application startup, so they
// appear at GET /api/metrics even before the first POST event lands.
@Context
class MicrometerPrometheusPublisher(registry: MeterRegistry) : MetricsPublisher {

    private val acceptedCounter: Counter = Counter
        .builder("mtrace_playback_events_total")
        .description("Total number of accepted playback events.")
        .register(registry)

    private val invalidCounter: Counter = Counter
        .builder("mtrace_invalid_events_total")
        .description("Total number of events rejected for schema or validation reasons (HTTP 400 / 422).")
        .register(registry)

    private val rateLimitedCounter: Counter = Counter
        .builder("mtrace_rate_limited_events_total")
        .description("Total number of events rejected by the rate limiter (HTTP 429).")
        .register(registry)

    private val droppedCounter: Counter = Counter
        .builder("mtrace_dropped_events_total")
        .description("Total number of events dropped internally (e.g. on persistence backpressure).")
        .register(registry)

    override fun eventsAccepted(n: Int) {
        if (n > 0) acceptedCounter.increment(n.toDouble())
    }

    override fun invalidEvents(n: Int) {
        if (n > 0) invalidCounter.increment(n.toDouble())
    }

    override fun rateLimitedEvents(n: Int) {
        if (n > 0) rateLimitedCounter.increment(n.toDouble())
    }

    override fun droppedEvents(n: Int) {
        if (n > 0) droppedCounter.increment(n.toDouble())
    }
}
