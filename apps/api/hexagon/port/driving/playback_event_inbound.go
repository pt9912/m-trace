// Package driving holds the inbound (driving) ports — the use-case
// entry points that adapters such as HTTP, gRPC, or future MCP
// implementations call into.
//
// Per docs/planning/plan-spike.md §5.2 nothing in this package may import any
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
// payload. Per spec/backend-api-contract.md §5 the use case is
// responsible for the full validation order from step 2 onwards.
// Origin="" → CLI/curl-Pfad: keine Project-Bindung. ClientIP="" →
// keine IP-basierte Rate-Limit-Dimension (Tests/headless flows).
type BatchInput struct {
	SchemaVersion string
	AuthToken     string
	Origin        string
	ClientIP      string
	Events        []EventInput
	// Trace ist der vom HTTP-Adapter aufgelöste Trace-Kontext für
	// diesen Batch (siehe spec/telemetry-model.md §2.5). Adapter füllt
	// `TraceID` und `SpanID` mit den IDs des Server-Spans (entweder als
	// Child eines validen `traceparent`-Headers oder als neuer Root);
	// `ParseError` ist true, wenn ein eingehender `traceparent` formal
	// kaputt war. Use Case kennt OTel nicht und liest nur diese drei
	// String-Werte.
	Trace BatchTraceContext
}

// BatchTraceContext ist die frameworkneutrale Sicht des HTTP-Adapters
// auf den Server-Span. Hex-Strings, kein OTel-Import.
type BatchTraceContext struct {
	TraceID    string
	SpanID     string
	ParseError bool
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
	Meta            map[string]any
}

// SDKInput is the wire counterpart of domain.SDKInfo.
type SDKInput struct {
	Name    string
	Version string
}

// BatchResult is what the use case returns on success. Trace-relevante
// Output-Felder reicht der HTTP-Adapter als Span-Attribute weiter
// (siehe spec/telemetry-model.md §2.5).
type BatchResult struct {
	Accepted int
	// SessionCount ist die Anzahl distinkter `session_id` im Batch —
	// für `mtrace.batch.session_count`.
	SessionCount int
	// SessionCorrelationID ist nur gesetzt, wenn alle Events im Batch
	// dieselbe `session_id` teilen (Single-Session-Batch); sonst leer.
	// Adapter setzt das Span-Attribut `mtrace.session.correlation_id`
	// nur, wenn der Wert nicht leer ist.
	SessionCorrelationID string
	// TimeSkewWarning ist true, wenn mindestens ein Event im Batch
	// `|client_timestamp - server_received_at| > 60s` hat. Adapter
	// setzt dann `mtrace.time.skew_warning=true` (siehe §5.3).
	TimeSkewWarning bool
}
