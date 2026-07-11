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
// Ab sind die Read-Pfade projekt-skopiert: ein Event-
// Cursor aus Project A darf weder Treffer in Project B liefern noch
// Cross-Project-Vermischung produzieren.
type EventRepository interface {
	Append(ctx context.Context, events []domain.PlaybackEvent) error
	// ListBySession liefert Events einer (projectID, sessionID)-Session
	// in stabiler Sortierung (server_received_at asc, sequence_number
	// asc, ingest_sequence asc). Limit und optionaler After-Cursor
	// steuern die Pagination — siehe driving.GetSessionInput. Der
	// Adapter ist für die Sortierung verantwortlich; der Use Case
	// clampt nur Limit und prüft Cursor-Validität.
	ListBySession(ctx context.Context, q EventListQuery) (EventPage, error)
	// ListAfterIngestSequence liefert Events eines Projects mit
	// `ingest_sequence > afterSeq`, sortiert aufsteigend nach
	// `ingest_sequence`. Maximal `limit` Treffer. Nutzung: SSE-
	// `Last-Event-ID`-Backfill ( H4; spec §10a). Der
	// Aufrufer entscheidet, ob ein Truncation-Marker an den
	// Konsumenten geht, wenn `len(out) == limit` ist.
	ListAfterIngestSequence(ctx context.Context, projectID string, afterSeq int64, limit int) ([]domain.PlaybackEvent, error)
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
	// Watermark ist das R-27-Commit-Zeit-Wasserzeichen (ADR-0006): der
	// Postgres-Adapter erfasst es bei Paginierungs-Session-Start und
	// trägt es über die Pages, damit spät-committende Früh-Rows nicht
	// übersprungen werden (Filter über pg_xact_commit_timestamp(xmin)).
	// Für Adapter ohne Commit-Order-Tracking (SQLite/InMemory) nil —
	// dort ist der Store single-writer bzw. snapshot-konsistent, das
	// Skip-Risiko existiert nicht.
	Watermark *time.Time
}

// EventPage bündelt eine Page Events plus optional die nächste
// Cursor-Position.
type EventPage struct {
	Events    []domain.PlaybackEvent
	NextAfter *EventCursorPosition
}
