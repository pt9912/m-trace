// Inbound (driving) ports — entry points called by adapters such as
// HTTP. Per docs/plan-spike.md §5.2 nothing here imports any driven
// adapter or wire-format concern; the HTTP adapter parses JSON into
// BatchInput before invoking the use case.
package dev.mtrace.api.hexagon.port.driving

/** Single use-case entry point for the spike: accept a batch. */
interface PlaybackEventInbound {
    fun registerPlaybackEventBatch(input: BatchInput): RegisterBatchResult
}

/** Wire-format-neutral request representation. */
data class BatchInput(
    val schemaVersion: String,
    val authToken: String,
    val events: List<EventInput>,
)

/**
 * Raw fields straight from the wire. The use case parses
 * clientTimestamp, normalizes identifiers, and rejects malformed
 * data with [RegisterBatchResult.InvalidEvent].
 */
data class EventInput(
    val eventName: String,
    val projectId: String,
    val sessionId: String,
    val clientTimestamp: String,
    val sequenceNumber: Long?,
    val sdk: SDKInput,
)

data class SDKInput(
    val name: String,
    val version: String,
)

/**
 * Sealed result of a register call. The HTTP adapter maps each
 * subtype to the status code defined in
 * docs/spike/backend-api-contract.md §5.
 */
sealed class RegisterBatchResult {
    /** Successful path: events persisted and counted. */
    data class Accepted(val count: Int) : RegisterBatchResult()

    /** schema_version != "1.0" -> 400. */
    data object SchemaVersionMismatch : RegisterBatchResult()

    /** Token invalid, missing project, or project_id mismatch -> 401. */
    data object Unauthorized : RegisterBatchResult()

    /** Rate-limit budget for the project exhausted -> 429. */
    data object RateLimited : RegisterBatchResult()

    /** events array missing or empty -> 422. */
    data object BatchEmpty : RegisterBatchResult()

    /** events.size > 100 -> 422. */
    data object BatchTooLarge : RegisterBatchResult()

    /** Per-event field check or timestamp parse failed -> 422. */
    data object InvalidEvent : RegisterBatchResult()

    /** Persistence failure or unexpected error -> 500. */
    data class InternalFailure(val cause: Throwable) : RegisterBatchResult()
}
