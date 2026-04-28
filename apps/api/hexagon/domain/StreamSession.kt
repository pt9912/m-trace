package dev.mtrace.api.hexagon.domain

import java.time.Instant

/**
 * Coarse session lifecycle. The spike auto-creates sessions on first
 * event and keeps them ACTIVE; explicit ENDED transition is bonus
 * scope per Spec §7.
 */
enum class SessionState { ACTIVE, ENDED }

data class StreamSession(
    val id: String,
    val projectId: String,
    val state: SessionState,
    val startedAt: Instant,
    val endedAt: Instant?,
)
