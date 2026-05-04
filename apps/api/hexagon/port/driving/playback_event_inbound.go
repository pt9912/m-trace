// Package driving holds the inbound (driving) ports — the use-case
// entry points that adapters such as HTTP, gRPC, or future MCP
// implementations call into.
//
// Per docs/planning/done/plan-spike.md §5.2 nothing in this package may import any
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
	// Boundaries ist der optionale `session_boundaries[]`-Wrapper aus
	// API-Kontrakt §3.4 (plan-0.4.0 §4.4). Maximal 20 pro Batch, jede
	// Boundary muss eine `(project_id, session_id)`-Partition
	// referenzieren, für die mindestens ein Event im selben Batch
	// existiert. Use-Case validiert atomar; Verstöße liefern 422 und
	// persistieren weder Events noch Boundaries.
	Boundaries []BoundaryInput
	// Trace ist der vom HTTP-Adapter aufgelöste Trace-Kontext für
	// diesen Batch (siehe spec/telemetry-model.md §2.5). Adapter füllt
	// `TraceID` und `SpanID` mit den IDs des Server-Spans (entweder als
	// Child eines validen `traceparent`-Headers oder als neuer Root).
	// Use Case kennt OTel nicht und liest nur die zwei Hex-Strings.
	// Parse-Errors aus dem `traceparent`-Header markiert der Adapter
	// direkt am Span (`mtrace.trace.parse_error=true`); der Use-Case
	// braucht den Wert nicht zu kennen.
	Trace BatchTraceContext
}

// BatchTraceContext ist die frameworkneutrale Sicht des HTTP-Adapters
// auf den Server-Span. Hex-Strings, kein OTel-Import.
type BatchTraceContext struct {
	TraceID string
	SpanID  string
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

// BoundaryInput is the wire-format-neutral representation eines
// `session_boundaries[]`-Eintrags (API-Kontrakt §3.4). Use-Case
// validiert Pflichtfelder, Reason-Enum/Pattern und Partition-Match
// gegen die Events des Batches.
type BoundaryInput struct {
	Kind            string
	ProjectID       string
	SessionID       string
	NetworkKind     string
	Adapter         string
	Reason          string
	ClientTimestamp string
}

// BatchResult is what the use case returns on success. Trace-relevante
// Output-Felder reicht der HTTP-Adapter als Span-Attribute weiter
// (siehe spec/telemetry-model.md §2.5).
type BatchResult struct {
	Accepted int
	// ProjectID ist das aufgelöste Project (Allowlist). Setzt das
	// Span-Attribut `mtrace.project.id` (Pflicht laut §2.5). Bei
	// Use-Case-Errors vor der Auth-Resolution leer.
	ProjectID string
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
