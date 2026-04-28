// Spike-only auth. Per Spec §6.4 the project_id -> token map is
// hardcoded; production deployments will replace this with a real
// token validator.
package dev.mtrace.api.adapters.driven.auth

import dev.mtrace.api.hexagon.domain.Project
import dev.mtrace.api.hexagon.domain.ProjectToken
import dev.mtrace.api.hexagon.port.driven.ProjectResolver
import jakarta.inject.Singleton

@Singleton
class StaticProjectResolver(
    byProjectId: Map<String, String> = mapOf("demo" to "demo-token"),
) : ProjectResolver {

    private val byToken: Map<String, Project> = byProjectId.entries
        .associate { (projectId, token) ->
            token to Project(id = projectId, token = ProjectToken(token))
        }

    override fun resolveByToken(token: String): Project? = byToken[token]
}
