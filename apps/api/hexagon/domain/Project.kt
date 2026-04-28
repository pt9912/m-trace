package dev.mtrace.api.hexagon.domain

/**
 * Opaque secret a project's SDK presents in the X-MTrace-Token
 * header. The spike compares it plaintext against a hardcoded map
 * (Spec §6.4); production deployments will use proper hashing or
 * external token validation.
 */
@JvmInline
value class ProjectToken(val value: String)

/** Tenant-like principal that owns events. */
data class Project(
    val id: String,
    val token: ProjectToken,
)
