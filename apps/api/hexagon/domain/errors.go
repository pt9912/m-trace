package domain

import "errors"

// Validation and authorization errors returned by the application
// layer. The HTTP adapter maps these to status codes per
// spec/backend-api-contract.md
var (
	ErrSchemaVersionMismatch = errors.New("schema_version not supported")
	ErrUnauthorized          = errors.New("authentication failed")
	ErrBatchEmpty            = errors.New("events array is empty or missing")
	ErrBatchTooLarge         = errors.New("batch exceeds 100 events")
	ErrInvalidEvent          = errors.New("event is missing required fields or has invalid values")
	ErrRateLimited           = errors.New("rate limit exceeded")
	// ErrSessionNotFound wird vom Detail-Use-Case zurückgegeben, wenn
	// keine Session zur angefragten ID existiert. HTTP-Mapping: 404.
	ErrSessionNotFound = errors.New("session not found")
	// ErrOriginNotAllowed wird vom Use Case zurückgegeben, wenn der
	// `Origin`-Header eines POST-Requests gegen die Allowed-Origins-
	// Liste des aufgelösten Projects mismatcht (CORS Variante B,
	// ). HTTP-Mapping: 403 Forbidden — vor Step 4
	// (Rate-Limit), damit weder Tokens noch Counter inkrementiert
	// werden.
	ErrOriginNotAllowed = errors.New("origin not allowed for project")
)

// Auth-/Token-Lifecycle-Domainfehler (`0.12.0`, RAK-71..RAK-76). Alle
// Konstanten sind stabil und werden vom HTTP-Adapter über `errors.Is`
// auf die in `spec/backend-api-contract.md` §3.9 gepinnten Codes
// gemappt:
//
//  ErrAuthTokenMissing → 401 auth_token_missing
//  ErrAuthTokenInvalid → 401 auth_token_invalid
//  ErrAuthTokenRevoked → 401 auth_token_revoked
//  ErrAuthTokenExpired → 401 auth_token_expired
//  ErrAuthTokenNotYetValid → 401 auth_token_not_yet_valid
//  ErrAuthProjectMismatch → 401 auth_project_mismatch
//  ErrAuthSessionScopeDenied → 403 auth_session_scope_denied
//  ErrAuthPolicyDenied → 403 auth_policy_denied
//  ErrAuthTokenTTLTooLarge → 422 auth_token_ttl_too_large
//  ErrAuthIssuanceRateLimited → 429 auth_issuance_rate_limited
//
// Die Reihenfolge spiegelt die neunstufige Fehlerpräzedenz aus
// 9 wider; Domain-Code liefert die Fehler einzeln zurück, der HTTP-
// Adapter ist für die Reihenfolge zuständig.
var (
	ErrAuthTokenMissing        = errors.New("auth token missing")
	ErrAuthTokenInvalid        = errors.New("auth token invalid")
	ErrAuthTokenRevoked        = errors.New("auth token revoked")
	ErrAuthTokenExpired        = errors.New("auth token expired")
	ErrAuthTokenNotYetValid    = errors.New("auth token not yet valid")
	ErrAuthProjectMismatch     = errors.New("auth token project mismatch")
	ErrAuthSessionScopeDenied  = errors.New("auth session scope denied")
	ErrAuthPolicyDenied        = errors.New("auth project policy denied")
	ErrAuthTokenTTLTooLarge    = errors.New("auth token ttl too large")
	ErrAuthIssuanceRateLimited = errors.New("auth issuance rate limited")
)

var (
	// ErrAnalyzeManifestEmpty wird zurückgegeben, wenn weder ManifestText
	// noch ManifestURL gesetzt ist; HTTP-Mapping: 400.
	ErrAnalyzeManifestEmpty = errors.New("analyze manifest: weder ManifestText noch ManifestURL gesetzt")
	// ErrSrtHealthStreamUnknown wird vom Detail-Read-Pfad zurückgegeben,
	// wenn die stream_id keinen Sample im Repository hat; HTTP-Mapping: 404.
	ErrSrtHealthStreamUnknown = errors.New("srt health: stream unknown")
)
