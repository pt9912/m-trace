package dev.mtrace.api.adapters.driving.http

import dev.mtrace.api.hexagon.application.RegisterPlaybackEventBatch
import dev.mtrace.api.hexagon.port.driven.EventRepository
import dev.mtrace.api.hexagon.port.driven.MetricsPublisher
import dev.mtrace.api.hexagon.port.driven.ProjectResolver
import dev.mtrace.api.hexagon.port.driven.RateLimiter
import dev.mtrace.api.hexagon.port.driving.RegisterBatchResult
import io.micronaut.http.HttpResponse
import io.micronaut.http.HttpStatus
import io.micronaut.http.MediaType
import io.micronaut.http.annotation.Body
import io.micronaut.http.annotation.Consumes
import io.micronaut.http.annotation.Controller
import io.micronaut.http.annotation.Header
import io.micronaut.http.annotation.Post
import io.micronaut.http.annotation.Produces

/**
 * POST /api/playback-events.
 *
 * The HTTP adapter walks docs/spike/backend-api-contract.md §5
 * step 1 (body size) via Micronaut's micronaut.server.max-request-size
 * config (returns 413 automatically) and step 2's bare presence check
 * for X-MTrace-Token. Steps 3-7 happen inside [RegisterPlaybackEventBatch];
 * the sealed [RegisterBatchResult] is mapped 1:1 to HTTP status codes here.
 */
@Controller("/api")
class PlaybackEventsController(
    private val projects: ProjectResolver,
    private val limiter: RateLimiter,
    private val events: EventRepository,
    private val metrics: MetricsPublisher,
) {

    @Post("/playback-events")
    @Consumes(MediaType.APPLICATION_JSON)
    @Produces(MediaType.APPLICATION_JSON)
    fun register(
        @Header("X-MTrace-Token") token: String?,
        @Body body: WireBatch,
    ): HttpResponse<Map<String, Any>> {
        if (token.isNullOrBlank()) {
            return HttpResponse.unauthorized()
        }

        val result = RegisterPlaybackEventBatch.execute(
            input = body.toBatchInput(token),
            projects = projects,
            limiter = limiter,
            events = events,
            metrics = metrics,
        )

        return when (result) {
            is RegisterBatchResult.Accepted ->
                HttpResponse.accepted<Map<String, Any>>().body(mapOf("accepted" to result.count))
            RegisterBatchResult.SchemaVersionMismatch ->
                HttpResponse.badRequest()
            RegisterBatchResult.Unauthorized ->
                HttpResponse.unauthorized()
            RegisterBatchResult.BatchEmpty,
            RegisterBatchResult.BatchTooLarge,
            RegisterBatchResult.InvalidEvent ->
                HttpResponse.unprocessableEntity()
            RegisterBatchResult.RateLimited ->
                HttpResponse
                    .status<Map<String, Any>>(HttpStatus.TOO_MANY_REQUESTS)
                    .header("Retry-After", "1")
            is RegisterBatchResult.InternalFailure ->
                HttpResponse.serverError()
        }
    }
}
