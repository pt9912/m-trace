// Spike-only auth. Per Spec §6.4 the project_id -> token map is
// hardcoded; production deployments will replace this with a real
// token validator.
package dev.mtrace.api.adapters.driven.auth

import dev.mtrace.api.hexagon.domain.Project
import dev.mtrace.api.hexagon.domain.ProjectToken
import dev.mtrace.api.hexagon.port.driven.ProjectResolver
import io.micronaut.context.annotation.Factory
import jakarta.inject.Singleton

/**
 * Plain class — instantiated by [StaticProjectResolverFactory] for
 * production wiring or by tests with a custom map. No @Singleton on
 * the class itself; Micronaut's KSP processor does not respect
 * Kotlin default constructor values for Map parameters.
 */
class StaticProjectResolver(
    byProjectId: Map<String, String>,
) : ProjectResolver {

    private val byToken: Map<String, Project> = byProjectId.entries
        .associate { (projectId, token) ->
            token to Project(id = projectId, token = ProjectToken(token))
        }

    override fun resolveByToken(token: String): Project? = byToken[token]
}

/** Production wiring with the demo project map (Spec §6.4). */
@Factory
class StaticProjectResolverFactory {
    @Singleton
    fun projectResolver(): ProjectResolver = StaticProjectResolver(
        byProjectId = mapOf("demo" to "demo-token"),
    )
}
