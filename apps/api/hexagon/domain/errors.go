package domain

import "errors"

// Validation and authorization errors returned by the application
// layer. The HTTP adapter maps these to status codes per
// spec/backend-api-contract.md §5.
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
	// plan-0.1.0.md §5.1). HTTP-Mapping: 403 Forbidden — vor Step 4
	// (Rate-Limit), damit weder Tokens noch Counter inkrementiert
	// werden.
	ErrOriginNotAllowed = errors.New("origin not allowed for project")
)
