// Package driven holds the outbound (driven) ports — the interfaces
// the application layer needs from the outside world (persistence,
// metrics, rate limiting, project lookup). Implementations live in
// adapters/driven/*.
package driven

import (
	"context"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// EventRepository persists accepted events. The spike uses an
// in-memory implementation; production will likely move to an event
// store. Implementations must be safe for concurrent use.
//
// Ab plan-0.4.0 §4.2 sind die Read-Pfade projekt-skopiert: ein Event-
// Cursor aus Project A darf weder Treffer in Project B liefern noch
// Cross-Project-Vermischung produzieren.
type EventRepository interface {
	Append(ctx context.Context, events []domain.PlaybackEvent) error
	// ListBySession liefert Events einer (projectID, sessionID)-Session
	// in stabiler Sortierung (server_received_at asc, sequence_number
	// asc, ingest_sequence asc). Limit und optionaler After-Cursor
	// steuern die Pagination — siehe driving.GetSessionInput. Der
	// Adapter ist für die Sortierung verantwortlich; der Use Case
	// clampt nur Limit und prüft Cursor-Validität (plan-0.1.0.md §5.1).
	ListBySession(ctx context.Context, q EventListQuery) (EventPage, error)
}

// EventListQuery ist die Eingabe für EventRepository.ListBySession.
// ProjectID und SessionID sind beide Pflicht; ein Leerwert ist ein
// Programmierfehler. After ist nil für die erste Seite; danach hält
// der Adapter den nächsten After-Wert in EventPage.NextAfter.
type EventListQuery struct {
	ProjectID string
	SessionID string
	Limit     int
	After     *EventCursorPosition
}

// EventCursorPosition ist die Repository-Sicht auf den Cursor. Der
// Wire-Codec lebt im HTTP-Adapter; hier sind die Sortier-Felder roh.
type EventCursorPosition struct {
	ServerReceivedAt time.Time
	SequenceNumber   *int64
	IngestSequence   int64
}

// EventPage bündelt eine Page Events plus optional die nächste
// Cursor-Position.
type EventPage struct {
	Events    []domain.PlaybackEvent
	NextAfter *EventCursorPosition
}
