// Package driving holds the inbound (driving) ports — the use-case
// entry points that adapters such as HTTP, gRPC, or future MCP
// implementations call into.
//
// Per docs/plan-spike.md §5.2 nothing in this package may import any
// driven adapter (persistence, metrics, telemetry) or any wire-format
// concern (JSON, Prometheus, OTel). The HTTP adapter is responsible
// for parsing JSON into BatchInput.
package driving

import "context"

// PlaybackEventInbound is the single use-case entry point for the
// spike: accept a batch of player events.
type PlaybackEventInbound interface {
	RegisterPlaybackEventBatch(ctx context.Context, in BatchInput) (BatchResult, error)
}

// BatchInput is the wire-format-neutral representation of a request to
// the API. It carries the raw header value (AuthToken), den optionalen
// Origin-Header (CORS Variante B, plan-0.1.0.md §5.1), die ermittelte
// ClientIP (für die Rate-Limit-Dimension F-110) und die parsed
// payload. Per docs/spike/backend-api-contract.md §5 the use case is
// responsible for the full validation order from step 2 onwards.
// Origin="" → CLI/curl-Pfad: keine Project-Bindung. ClientIP="" →
// keine IP-basierte Rate-Limit-Dimension (Tests/headless flows).
type BatchInput struct {
	SchemaVersion string
	AuthToken     string
	Origin        string
	ClientIP      string
	Events        []EventInput
}

// EventInput carries raw fields straight from the wire. The use case
// parses ClientTimestamp, normalizes identifiers, and rejects
// malformed data with domain.ErrInvalidEvent.
type EventInput struct {
	EventName       string
	ProjectID       string
	SessionID       string
	ClientTimestamp string
	SequenceNumber  *int64
	SDK             SDKInput
}

// SDKInput is the wire counterpart of domain.SDKInfo.
type SDKInput struct {
	Name    string
	Version string
}

// BatchResult is what the use case returns on success.
type BatchResult struct {
	Accepted int
}
