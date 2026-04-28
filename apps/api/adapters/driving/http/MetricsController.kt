package dev.mtrace.api.adapters.driving.http

import io.micrometer.prometheusmetrics.PrometheusMeterRegistry
import io.micronaut.http.MediaType
import io.micronaut.http.annotation.Controller
import io.micronaut.http.annotation.Get
import io.micronaut.http.annotation.Produces

/**
 * Exposes Prometheus-format metrics at `GET /api/metrics`. Per
 * docs/spike/backend-api-contract.md §7 the four mandatory mtrace_*
 * counters must appear here.
 *
 * Implementing the endpoint as a regular Micronaut controller (rather
 * than via the management prometheus endpoint) keeps full control
 * over the URL and avoids the management auto-config plumbing for
 * the spike scope.
 */
@Controller("/api")
class MetricsController(
    private val registry: PrometheusMeterRegistry,
) {

    @Get("/metrics")
    @Produces(MediaType.TEXT_PLAIN)
    fun metrics(): String = registry.scrape()
}
