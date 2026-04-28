package dev.mtrace.api.hexagon.port.driven

import dev.mtrace.api.hexagon.domain.Project

/**
 * Looks up a project by its presented X-MTrace-Token. Returns null
 * when no project owns the token. Spike implementation reads from a
 * static map (Spec §6.4).
 */
interface ProjectResolver {
    fun resolveByToken(token: String): Project?
}
