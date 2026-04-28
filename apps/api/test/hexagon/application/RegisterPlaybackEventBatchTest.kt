package dev.mtrace.api.hexagon.application

import dev.mtrace.api.hexagon.domain.PlaybackEvent
import dev.mtrace.api.hexagon.domain.Project
import dev.mtrace.api.hexagon.domain.ProjectToken
import dev.mtrace.api.hexagon.port.driven.EventRepository
import dev.mtrace.api.hexagon.port.driven.MetricsPublisher
import dev.mtrace.api.hexagon.port.driven.ProjectResolver
import dev.mtrace.api.hexagon.port.driven.RateLimiter
import dev.mtrace.api.hexagon.port.driving.BatchInput
import dev.mtrace.api.hexagon.port.driving.EventInput
import dev.mtrace.api.hexagon.port.driving.RegisterBatchResult
import dev.mtrace.api.hexagon.port.driving.SDKInput
import io.kotest.core.spec.style.StringSpec
import io.kotest.matchers.shouldBe
import io.kotest.matchers.types.shouldBeInstanceOf
import java.time.Clock
import java.time.Instant
import java.time.ZoneOffset

private val FIXED_NOW: Instant = Instant.parse("2026-04-28T12:00:00.000Z")
private val FIXED_CLOCK: Clock = Clock.fixed(FIXED_NOW, ZoneOffset.UTC)
private val DEMO_PROJECT = Project(id = "demo", token = ProjectToken("demo-token"))

private class StubResolver : ProjectResolver {
    override fun resolveByToken(token: String): Project? =
        if (token == "demo-token") DEMO_PROJECT else null
}

private class StubLimiter(var deny: Boolean = false) : RateLimiter {
    override fun allow(projectId: String, n: Int): Boolean = !deny
}

private class StubRepo(var failNext: Boolean = false) : EventRepository {
    val appended: MutableList<PlaybackEvent> = mutableListOf()

    override fun append(events: List<PlaybackEvent>) {
        if (failNext) {
            failNext = false
            error("repo failure")
        }
        appended += events
    }
}

private class SpyMetrics : MetricsPublisher {
    var accepted = 0
    var invalid = 0
    var rateLimited = 0
    var dropped = 0

    override fun eventsAccepted(n: Int) { accepted += n }
    override fun invalidEvents(n: Int) { invalid += n }
    override fun rateLimitedEvents(n: Int) { rateLimited += n }
    override fun droppedEvents(n: Int) { dropped += n }
}

private fun validBatch(): BatchInput = BatchInput(
    schemaVersion = RegisterPlaybackEventBatch.SUPPORTED_SCHEMA_VERSION,
    authToken = "demo-token",
    events = listOf(
        EventInput(
            eventName = "rebuffer_started",
            projectId = "demo",
            sessionId = "01J7K9X4Z2QHB6V3WS5R8Y4D1F",
            clientTimestamp = "2026-04-28T12:00:00.000Z",
            sequenceNumber = null,
            sdk = SDKInput(name = "@m-trace/player-sdk", version = "0.1.0"),
        ),
    ),
)

private fun run(
    input: BatchInput,
    limiter: StubLimiter = StubLimiter(),
    repo: StubRepo = StubRepo(),
    metrics: SpyMetrics = SpyMetrics(),
): Triple<RegisterBatchResult, StubRepo, SpyMetrics> {
    val result = RegisterPlaybackEventBatch.execute(
        input = input,
        projects = StubResolver(),
        limiter = limiter,
        events = repo,
        metrics = metrics,
        clock = FIXED_CLOCK,
    )
    return Triple(result, repo, metrics)
}

class RegisterPlaybackEventBatchTest : StringSpec({

    "happy path returns Accepted with count and persists events" {
        val (result, repo, metrics) = run(validBatch())
        result shouldBe RegisterBatchResult.Accepted(1)
        repo.appended.size shouldBe 1
        metrics.accepted shouldBe 1
    }

    "unknown token returns Unauthorized" {
        val (result, _, _) = run(validBatch().copy(authToken = "wrong-token"))
        result shouldBe RegisterBatchResult.Unauthorized
    }

    "wrong schema version returns SchemaVersionMismatch and counts as invalid" {
        val (result, _, metrics) = run(validBatch().copy(schemaVersion = "2.0"))
        result shouldBe RegisterBatchResult.SchemaVersionMismatch
        metrics.invalid shouldBe 1
    }

    "empty events returns BatchEmpty" {
        val (result, _, _) = run(validBatch().copy(events = emptyList()))
        result shouldBe RegisterBatchResult.BatchEmpty
    }

    "more than 100 events returns BatchTooLarge with full count invalid" {
        val template = validBatch().events.first()
        val input = validBatch().copy(
            events = List(RegisterPlaybackEventBatch.MAX_BATCH_SIZE + 1) { template },
        )
        val (result, _, metrics) = run(input)
        result shouldBe RegisterBatchResult.BatchTooLarge
        metrics.invalid shouldBe RegisterPlaybackEventBatch.MAX_BATCH_SIZE + 1
    }

    "missing required field returns InvalidEvent" {
        val input = validBatch().let { b ->
            b.copy(events = b.events.map { it.copy(eventName = "") })
        }
        val (result, _, _) = run(input)
        result shouldBe RegisterBatchResult.InvalidEvent
    }

    "bad client_timestamp returns InvalidEvent" {
        val input = validBatch().let { b ->
            b.copy(events = b.events.map { it.copy(clientTimestamp = "not-a-timestamp") })
        }
        val (result, _, _) = run(input)
        result shouldBe RegisterBatchResult.InvalidEvent
    }

    "project_id mismatch returns Unauthorized" {
        val input = validBatch().let { b ->
            b.copy(events = b.events.map { it.copy(projectId = "other") })
        }
        val (result, _, _) = run(input)
        result shouldBe RegisterBatchResult.Unauthorized
    }

    "rate limited returns RateLimited and counts toward rate-limit metric" {
        val (result, _, metrics) = run(validBatch(), limiter = StubLimiter(deny = true))
        result shouldBe RegisterBatchResult.RateLimited
        metrics.rateLimited shouldBe 1
    }

    "repo failure returns InternalFailure and counts toward dropped metric" {
        val (result, _, metrics) = run(validBatch(), repo = StubRepo(failNext = true))
        result.shouldBeInstanceOf<RegisterBatchResult.InternalFailure>()
        metrics.dropped shouldBe 1
    }
})
