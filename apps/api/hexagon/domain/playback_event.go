// Package domain holds the framework-free fact-types of the spike.
//
// Per docs/planning/plan-spike.md §5.2/§14.6 nothing in this package may import
// HTTP, JSON, OpenTelemetry, Prometheus, or any other adapter concern.
package domain

import "time"

// PlaybackEvent is a normalized player-side event accepted by the API.
// The wire-format counterpart lives in
// hexagon/port/driving (BatchInput / EventInput).
//
// IngestSequence ist ein serverseitig gesetzter Pflicht-Counter
// (plan-0.1.0.md §5.1): monoton steigend pro apps/api-Prozess, vor
// Append im Use Case gesetzt. Begründet die Eindeutigkeit der
// Pagination-Sortierung auch bei identischen Client-Feldern und
// gleichen ServerReceivedAt-Werten.
type PlaybackEvent struct {
	EventName        string
	ProjectID        string
	SessionID        string
	ClientTimestamp  time.Time
	ServerReceivedAt time.Time
	IngestSequence   int64
	SequenceNumber   *int64
	SDK              SDKInfo
	Meta             EventMeta
	// TraceID ist die W3C-Trace-ID (32 Hex-Zeichen) des Batches, in
	// dem das Event registriert wurde — entweder vom SDK propagiert
	// (`traceparent`-Header) oder server-generiert. Empty-String =
	// nicht gesetzt (Edge-Case in Tests/Fallbacks); Read-Pfad mappt
	// das auf JSON `null`. Siehe spec/telemetry-model.md §2.5.
	TraceID string
	// SpanID ist die ID des Server-Spans, der diesen Event verarbeitet
	// hat (16 Hex-Zeichen). Empty-String = nicht gesetzt.
	SpanID string
	// CorrelationID ist die server-generierte, durable Source-of-Truth
	// für die Tempo-unabhängige Dashboard-Korrelation. Beim ersten
	// Event einer Session erzeugt (UUIDv4), für alle Folge-Events
	// derselben Session konstant. Niemals leer in 0.4.0+-Read-Pfaden.
	CorrelationID string
}

// SDKInfo identifies the producing player SDK.
type SDKInfo struct {
	Name    string
	Version string
}

// EventMeta carries event-specific scalar attributes from the player
// SDK wire format. Values stay generic inside the domain because only
// selected, bounded aggregate metrics are interpreted by the use case.
type EventMeta map[string]any
