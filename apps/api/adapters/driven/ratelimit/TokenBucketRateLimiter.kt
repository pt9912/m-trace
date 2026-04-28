// In-memory token-bucket. Per Spec §6.9 the spike is single-process;
// distributed rate limiting is out of scope.
package dev.mtrace.api.adapters.driven.ratelimit

import dev.mtrace.api.hexagon.port.driven.RateLimiter
import io.micronaut.context.annotation.Factory
import jakarta.inject.Singleton
import java.time.Clock
import java.time.Instant
import java.util.concurrent.locks.ReentrantLock
import kotlin.concurrent.withLock

const val DEFAULT_RATE_LIMIT_CAPACITY: Int = 100
const val DEFAULT_RATE_LIMIT_REFILL_PER_SEC: Double = 100.0

/**
 * Plain class — instantiated either by tests (with custom clock) or
 * by [TokenBucketFactory] for production wiring. No @Singleton on
 * the class itself because Micronaut's KSP processor does not
 * resolve Kotlin default constructor values; adding @Singleton would
 * make the DI container try to inject Int/Double beans and fail.
 */
class TokenBucketRateLimiter(
    private val capacity: Int = DEFAULT_RATE_LIMIT_CAPACITY,
    private val refillPerSec: Double = DEFAULT_RATE_LIMIT_REFILL_PER_SEC,
    private val clock: Clock = Clock.systemUTC(),
) : RateLimiter {

    private val lock = ReentrantLock()
    private val buckets = mutableMapOf<String, Bucket>()

    override fun allow(projectId: String, n: Int): Boolean {
        if (n <= 0) return true
        return lock.withLock {
            val now = Instant.now(clock)
            val bucket = buckets.getOrPut(projectId) {
                Bucket(tokens = capacity.toDouble(), lastRefill = now)
            }
            val elapsed = (now.toEpochMilli() - bucket.lastRefill.toEpochMilli()) / MILLIS_PER_SECOND
            if (elapsed > 0) {
                bucket.tokens = (bucket.tokens + elapsed * refillPerSec).coerceAtMost(capacity.toDouble())
                bucket.lastRefill = now
            }
            if (bucket.tokens >= n.toDouble()) {
                bucket.tokens -= n.toDouble()
                true
            } else {
                false
            }
        }
    }

    private class Bucket(var tokens: Double, var lastRefill: Instant)

    private companion object {
        private const val MILLIS_PER_SECOND: Double = 1000.0
    }
}

/** Production wiring for [TokenBucketRateLimiter]. */
@Factory
class TokenBucketFactory {
    @Singleton
    fun rateLimiter(): RateLimiter = TokenBucketRateLimiter()
}
