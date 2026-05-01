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
