// Outbound (driven) ports — interfaces the application layer needs
// from the outside world. Implementations live in adapters/driven/*.
package dev.mtrace.api.hexagon.port.driven

import dev.mtrace.api.hexagon.domain.PlaybackEvent

/**
 * Persists accepted events. The spike uses an in-memory implementation;
 * production will likely move to an event store. Implementations must
 * be safe for concurrent use.
 */
interface EventRepository {
    fun append(events: List<PlaybackEvent>)
}
