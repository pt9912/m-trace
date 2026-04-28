// Inner-hexagon domain types. Per docs/plan-spike.md §5.2/§14.6
// nothing in this package may import HTTP, JSON, OpenTelemetry,
// Prometheus, or any other adapter concern.
package dev.mtrace.api.hexagon.domain

import java.time.Instant

/**
 * Normalized player-side event accepted by the API. The wire-format
 * counterpart lives in hexagon.port.driving (BatchInput / EventInput).
 */
data class PlaybackEvent(
    val eventName: String,
    val projectId: String,
    val sessionId: String,
    val clientTimestamp: Instant,
    val serverReceivedAt: Instant,
    val sequenceNumber: Long?,
    val sdk: SDKInfo,
)

data class SDKInfo(
    val name: String,
    val version: String,
)
