package dev.mtrace.api.hexagon.port.driven

/**
 * Consumes [n] event-tokens for the given [projectId]. Returns true
 * when allowed, false when the budget is exhausted. Spike
 * implementation is an in-memory token bucket; distributed rate
 * limiting is out of scope (Spec §6.9).
 */
interface RateLimiter {
    fun allow(projectId: String, n: Int): Boolean
}
