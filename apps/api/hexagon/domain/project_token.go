package domain

// ProjectToken is the opaque secret a project's SDK presents in the
// X-MTrace-Token header. The spike compares it plaintext against a
// hardcoded map (Spec §6.4); production deployments will use proper
// hashing or external token validation.
type ProjectToken string
