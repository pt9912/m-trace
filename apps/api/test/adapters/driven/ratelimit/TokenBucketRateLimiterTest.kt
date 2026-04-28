package dev.mtrace.api.adapters.driven.ratelimit

import io.kotest.core.spec.style.StringSpec
import io.kotest.matchers.shouldBe
import java.time.Clock
import java.time.Instant
import java.time.ZoneOffset

private val START: Instant = Instant.parse("2026-04-28T12:00:00.000Z")

private class MutableClock(var now: Instant) : Clock() {
    override fun getZone() = ZoneOffset.UTC
    override fun withZone(zone: java.time.ZoneId?): Clock = this
    override fun instant(): Instant = now
}

class TokenBucketRateLimiterTest : StringSpec({

    "allows up to capacity per project" {
        val clock = MutableClock(START)
        val rl = TokenBucketRateLimiter(
            capacity = 100,
            refillPerSec = 100.0,
            clock = clock,
        )
        repeat(100) {
            rl.allow("demo", 1) shouldBe true
        }
        rl.allow("demo", 1) shouldBe false
    }

    "refills over time" {
        val clock = MutableClock(START)
        val rl = TokenBucketRateLimiter(
            capacity = 100,
            refillPerSec = 100.0,
            clock = clock,
        )
        rl.allow("demo", 100) shouldBe true
        rl.allow("demo", 1) shouldBe false

        // Advance the clock by one second; the bucket should refill
        // back to capacity.
        clock.now = START.plusSeconds(1)
        rl.allow("demo", 100) shouldBe true
    }

    "isolates buckets per project" {
        val clock = MutableClock(START)
        val rl = TokenBucketRateLimiter(
            capacity = 100,
            refillPerSec = 100.0,
            clock = clock,
        )
        rl.allow("demo", 100) shouldBe true
        rl.allow("other", 100) shouldBe true
    }

    "rejects single batch larger than capacity" {
        val rl = TokenBucketRateLimiter(
            capacity = 100,
            refillPerSec = 100.0,
            clock = MutableClock(START),
        )
        rl.allow("demo", 101) shouldBe false
    }
})
