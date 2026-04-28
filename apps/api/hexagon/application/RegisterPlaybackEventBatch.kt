// Use-case service. Per docs/plan-spike.md §5.2 application code may
// import domain and ports but no adapter (HTTP, JSON, Prometheus,
// OTel) — and crucially no DI framework either. The Kotlin prototype
// expresses the use case as an object with a stateless execute()
// function rather than a @Singleton class, so the inner hexagon
// stays free of jakarta.inject.
package dev.mtrace.api.hexagon.application

import dev.mtrace.api.hexagon.domain.PlaybackEvent
import dev.mtrace.api.hexagon.domain.SDKInfo
import dev.mtrace.api.hexagon.port.driven.EventRepository
import dev.mtrace.api.hexagon.port.driven.MetricsPublisher
import dev.mtrace.api.hexagon.port.driven.ProjectResolver
import dev.mtrace.api.hexagon.port.driven.RateLimiter
import dev.mtrace.api.hexagon.port.driving.BatchInput
import dev.mtrace.api.hexagon.port.driving.EventInput
import dev.mtrace.api.hexagon.port.driving.RegisterBatchResult
import java.time.Clock
import java.time.Instant
import java.time.format.DateTimeParseException

/**
 * Validates and persists a batch of player events. Walks the
 * docs/spike/backend-api-contract.md §5 validation order from step 2
 * (auth) onwards; steps 1 (body size) and the bare presence of
 * X-MTrace-Token are the HTTP adapter's responsibility.
 *
 * Stateless — driving adapters call [execute] with their injected
 * dependencies. No instance is needed; no DI annotation appears in
 * the inner hexagon.
 */
object RegisterPlaybackEventBatch {

    const val SUPPORTED_SCHEMA_VERSION = "1.0"
    const val MAX_BATCH_SIZE = 100

    @Suppress("LongParameterList")
    fun execute(
        input: BatchInput,
        projects: ProjectResolver,
        limiter: RateLimiter,
        events: EventRepository,
        metrics: MetricsPublisher,
        clock: Clock = Clock.systemUTC(),
    ): RegisterBatchResult {
        // Step 2 — auth: resolve token to project.
        val project = projects.resolveByToken(input.authToken)
            ?: return RegisterBatchResult.Unauthorized

        // Step 3 — rate limit: charged for the requested batch size,
        // even if validation later rejects the batch — so a caller
        // can't probe validation responses without paying the
        // per-project budget.
        if (!limiter.allow(project.id, input.events.size)) {
            metrics.rateLimitedEvents(input.events.size)
            return RegisterBatchResult.RateLimited
        }

        // Step 4 — schema version.
        if (input.schemaVersion != SUPPORTED_SCHEMA_VERSION) {
            metrics.invalidEvents(input.events.size)
            return RegisterBatchResult.SchemaVersionMismatch
        }

        // Step 5 — batch shape.
        if (input.events.isEmpty()) {
            metrics.invalidEvents(0)
            return RegisterBatchResult.BatchEmpty
        }
        if (input.events.size > MAX_BATCH_SIZE) {
            metrics.invalidEvents(input.events.size)
            return RegisterBatchResult.BatchTooLarge
        }

        val now = Instant.now(clock)
        val parsed = mutableListOf<PlaybackEvent>()
        for (e in input.events) {
            // Step 6 — per-event fields.
            if (!hasRequiredFields(e)) {
                metrics.invalidEvents(input.events.size)
                return RegisterBatchResult.InvalidEvent
            }
            // Step 7 — token/project binding.
            if (e.projectId != project.id) {
                metrics.invalidEvents(input.events.size)
                return RegisterBatchResult.Unauthorized
            }
            val ts = try {
                Instant.parse(e.clientTimestamp)
            } catch (_: DateTimeParseException) {
                metrics.invalidEvents(input.events.size)
                return RegisterBatchResult.InvalidEvent
            }
            parsed += PlaybackEvent(
                eventName = e.eventName,
                projectId = e.projectId,
                sessionId = e.sessionId,
                clientTimestamp = ts,
                serverReceivedAt = now,
                sequenceNumber = e.sequenceNumber,
                sdk = SDKInfo(name = e.sdk.name, version = e.sdk.version),
            )
        }

        // Step 8 — persist + accept.
        return try {
            events.append(parsed)
            metrics.eventsAccepted(parsed.size)
            RegisterBatchResult.Accepted(parsed.size)
        } catch (@Suppress("TooGenericExceptionCaught") t: Throwable) {
            // Any persistence failure must surface as a "dropped"
            // counter increment per
            // docs/spike/backend-api-contract.md §7. The repo
            // contract does not enumerate exception types.
            metrics.droppedEvents(parsed.size)
            RegisterBatchResult.InternalFailure(t)
        }
    }

    private fun hasRequiredFields(e: EventInput): Boolean =
        e.eventName.isNotBlank() &&
            e.projectId.isNotBlank() &&
            e.sessionId.isNotBlank() &&
            e.clientTimestamp.isNotBlank() &&
            e.sdk.name.isNotBlank() &&
            e.sdk.version.isNotBlank()
}
