// Package inmemory liefert die In-Memory-Variante der Driven-
// Persistence-Ports (Sessions, Events, Ingest-Sequencer). Sie wird
// für Tests, lokale Entwicklung ohne SQLite-Volume und für Adapter-
// Vergleichs-Contract-Tests genutzt; Daten überleben keinen Restart.
package inmemory

import (
	"context"
	"sort"
	"sync"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// EventRepository keeps accepted events in a slice. Safe for
// concurrent use; the spike scope does not require performance tuning.
type EventRepository struct {
	mu     sync.Mutex
	events []domain.PlaybackEvent
}

// NewEventRepository constructs an empty repository.
func NewEventRepository() *EventRepository {
	return &EventRepository{}
}

// Append stores all events atomically.
func (r *EventRepository) Append(_ context.Context, events []domain.PlaybackEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, events...)
	return nil
}

// Snapshot returns a copy of the stored events. Useful for tests.
func (r *EventRepository) Snapshot() []domain.PlaybackEvent {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]domain.PlaybackEvent, len(r.events))
	copy(out, r.events)
	return out
}

// ListBySession liefert Events einer Session in stabiler Sortierung
// (server_received_at asc, sequence_number asc, ingest_sequence asc).
// After=nil → erste Seite. Wenn nach Limit weitere Events vorhanden,
// ist NextAfter gesetzt.
func (r *EventRepository) ListBySession(_ context.Context, q driven.EventListQuery) (driven.EventPage, error) {
	r.mu.Lock()
	matching := make([]domain.PlaybackEvent, 0)
	for _, e := range r.events {
		if e.ProjectID == q.ProjectID && e.SessionID == q.SessionID {
			matching = append(matching, e)
		}
	}
	r.mu.Unlock()

	sort.Slice(matching, func(i, j int) bool {
		return eventLess(matching[i], matching[j])
	})

	if q.After != nil {
		idx := sort.Search(len(matching), func(i int) bool {
			return eventCursorPasses(matching[i], *q.After)
		})
		matching = matching[idx:]
	}

	limit := q.Limit
	if limit <= 0 {
		return driven.EventPage{Events: []domain.PlaybackEvent{}}, nil
	}

	page := driven.EventPage{}
	if len(matching) > limit {
		page.Events = append(page.Events, matching[:limit]...)
		last := page.Events[limit-1]
		page.NextAfter = &driven.EventCursorPosition{
			ServerReceivedAt: last.ServerReceivedAt,
			SequenceNumber:   last.SequenceNumber,
			IngestSequence:   last.IngestSequence,
		}
		return page, nil
	}
	page.Events = append(page.Events, matching...)
	return page, nil
}

// eventLess implementiert die Sort-Order (server_received_at asc,
// sequence_number asc, ingest_sequence asc). nil-SequenceNumber wird
// als kleiner als jede gesetzte Nummer behandelt; ingest_sequence ist
// der finale Tie-Breaker (plan-0.1.0.md §5.1).
func eventLess(a, b domain.PlaybackEvent) bool {
	if !a.ServerReceivedAt.Equal(b.ServerReceivedAt) {
		return a.ServerReceivedAt.Before(b.ServerReceivedAt)
	}
	an, bn := nullableSeqValue(a.SequenceNumber), nullableSeqValue(b.SequenceNumber)
	if an != bn {
		return an < bn
	}
	return a.IngestSequence < b.IngestSequence
}

// eventCursorPasses gibt true zurück, sobald e strikt hinter dem
// After-Cursor in der Sort-Order liegt.
func eventCursorPasses(e domain.PlaybackEvent, after driven.EventCursorPosition) bool {
	if !e.ServerReceivedAt.Equal(after.ServerReceivedAt) {
		return e.ServerReceivedAt.After(after.ServerReceivedAt)
	}
	en, an := nullableSeqValue(e.SequenceNumber), nullableSeqValue(after.SequenceNumber)
	if en != an {
		return en > an
	}
	return e.IngestSequence > after.IngestSequence
}

// nullableSeqValue normalisiert die optionale SequenceNumber für die
// Sort-Order: nil → Sentinel-Wert vor allen Zahlen.
func nullableSeqValue(p *int64) int64 {
	if p == nil {
		// Math-Min int64 — sorts first.
		const minInt64 = -1 << 63
		return minInt64
	}
	return *p
}

// ListAfterIngestSequence liefert Events eines Projects mit
// `ingest_sequence > afterSeq`, sortiert aufsteigend, max `limit`
// Treffer. Backfill-Quelle für SSE-`Last-Event-ID`-Reconnect
// (plan-0.4.0 §5 H4).
func (r *EventRepository) ListAfterIngestSequence(_ context.Context, projectID string, afterSeq int64, limit int) ([]domain.PlaybackEvent, error) {
	if limit <= 0 {
		return nil, nil
	}
	r.mu.Lock()
	matching := make([]domain.PlaybackEvent, 0)
	for _, e := range r.events {
		if e.ProjectID == projectID && e.IngestSequence > afterSeq {
			matching = append(matching, e)
		}
	}
	r.mu.Unlock()

	sort.Slice(matching, func(i, j int) bool {
		return matching[i].IngestSequence < matching[j].IngestSequence
	})
	if len(matching) > limit {
		matching = matching[:limit]
	}
	return matching, nil
}

var _ driven.EventRepository = (*EventRepository)(nil)
