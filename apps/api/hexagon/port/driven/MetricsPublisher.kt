package dev.mtrace.api.hexagon.port.driven

/**
 * Exposes the four mandatory counters from
 * docs/spike/backend-api-contract.md §7. Implementation lives in
 * adapters/driven/metrics. Calls must be safe for concurrent use.
 */
interface MetricsPublisher {
    fun eventsAccepted(n: Int)
    fun invalidEvents(n: Int)
    fun rateLimitedEvents(n: Int)
    fun droppedEvents(n: Int)
}
