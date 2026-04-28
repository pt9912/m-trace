// Inbound (driving) port shapes — the wire-format-neutral input and
// result types that adapters such as HTTP, gRPC, or future MCP
// implementations exchange with the application layer. Per
// docs/plan-spike.md §5.2 nothing here imports any driven adapter or
// wire-format concern.
//
// The port itself is the function signature
// `dev.mtrace.api.hexagon.application.RegisterPlaybackEventBatch
// .execute(...)` — Kotlin idiom prefers an `object` with a stateless
// function over a class + interface for a single use case (cleaner
// than @Singleton, leaves the inner hexagon free of jakarta.inject).
package dev.mtrace.api.hexagon.port.driving

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
