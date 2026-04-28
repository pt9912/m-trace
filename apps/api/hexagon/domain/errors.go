package domain

import "errors"

// Validation and authorization errors returned by the application
// layer. The HTTP adapter maps these to status codes per
// docs/spike/backend-api-contract.md §5.
var (
	ErrSchemaVersionMismatch = errors.New("schema_version not supported")
	ErrUnauthorized          = errors.New("authentication failed")
	ErrBatchEmpty            = errors.New("events array is empty or missing")
	ErrBatchTooLarge         = errors.New("batch exceeds 100 events")
	ErrInvalidEvent          = errors.New("event is missing required fields or has invalid values")
	ErrRateLimited           = errors.New("rate limit exceeded")
)
