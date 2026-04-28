// In-memory storage for the spike. Per
// docs/spike/0001-backend-stack.md §6.10 there is no on-disk
// persistence; data does not survive a restart, on purpose.
package dev.mtrace.api.adapters.driven.persistence

import dev.mtrace.api.hexagon.domain.PlaybackEvent
import dev.mtrace.api.hexagon.port.driven.EventRepository
import jakarta.inject.Singleton
import java.util.concurrent.locks.ReentrantLock
import kotlin.concurrent.withLock

@Singleton
class InMemoryEventRepository : EventRepository {

    private val lock = ReentrantLock()
    private val store = mutableListOf<PlaybackEvent>()

    override fun append(events: List<PlaybackEvent>) {
        lock.withLock { store += events }
    }

    /** Test helper — returns a defensive copy of the current store. */
    fun snapshot(): List<PlaybackEvent> = lock.withLock { store.toList() }
}
