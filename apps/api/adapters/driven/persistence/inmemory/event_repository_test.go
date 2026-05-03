package inmemory_test

import (
	"context"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/persistence/inmemory"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// TestEventRepository_ListBySession_SortAndCursor verifiziert
// die Sort-Order (server_received_at asc, sequence_number asc,
// ingest_sequence asc) und dass die Pagination Events strikt hinter dem
// After-Cursor liefert — ohne Duplikate, ohne Lücken.
func TestEventRepository_ListBySession_SortAndCursor(t *testing.T) {
	t.Parallel()
	repo := inmemory.NewEventRepository()
	t0 := time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC)

	// Mix mit absichtlich unsortierter Insertion.
	mk := func(sess string, recv time.Time, seq *int64, ing int64) domain.PlaybackEvent {
		return domain.PlaybackEvent{
			ProjectID:        "demo",
			SessionID:        sess,
			ServerReceivedAt: recv,
			SequenceNumber:   seq,
			IngestSequence:   ing,
		}
	}
	intp := func(v int64) *int64 { return &v }

	other := mk("other", t0, intp(1), 1)
	if err := repo.Append(context.Background(), []domain.PlaybackEvent{
		mk("s1", t0.Add(2*time.Second), intp(2), 4),
		other,
		mk("s1", t0, intp(1), 1),
		mk("s1", t0.Add(time.Second), intp(2), 3),
		mk("s1", t0.Add(time.Second), intp(1), 2),
	}); err != nil {
		t.Fatalf("append: %v", err)
	}

	first, err := repo.ListBySession(context.Background(), driven.EventListQuery{ProjectID: "demo",
		SessionID: "s1",
		Limit:     2,
	})
	if err != nil {
		t.Fatalf("list page 1: %v", err)
	}
	if len(first.Events) != 2 {
		t.Fatalf("page 1: expected 2 events, got %d", len(first.Events))
	}
	// Erwartete Reihenfolge: ingest 1 (t0,seq1), ingest 2 (t0+1,seq1).
	if first.Events[0].IngestSequence != 1 || first.Events[1].IngestSequence != 2 {
		t.Errorf("page 1 ingest order: %d %d want 1 2",
			first.Events[0].IngestSequence, first.Events[1].IngestSequence)
	}
	if first.NextAfter == nil {
		t.Fatalf("page 1 expected NextAfter")
	}

	second, err := repo.ListBySession(context.Background(), driven.EventListQuery{ProjectID: "demo",
		SessionID: "s1",
		Limit:     2,
		After:     first.NextAfter,
	})
	if err != nil {
		t.Fatalf("list page 2: %v", err)
	}
	if len(second.Events) != 2 {
		t.Fatalf("page 2: expected 2 events, got %d", len(second.Events))
	}
	if second.Events[0].IngestSequence != 3 || second.Events[1].IngestSequence != 4 {
		t.Errorf("page 2 ingest order: %d %d want 3 4",
			second.Events[0].IngestSequence, second.Events[1].IngestSequence)
	}
	if second.NextAfter != nil {
		t.Errorf("page 2 should be last (NextAfter=nil), got %v", second.NextAfter)
	}
}

// TestEventRepository_Snapshot deckt die for-Tests-gedachte
// Hilfsmethode ab.
func TestEventRepository_Snapshot(t *testing.T) {
	t.Parallel()
	repo := inmemory.NewEventRepository()
	t0 := time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC)
	in := []domain.PlaybackEvent{
		{SessionID: "s1", ServerReceivedAt: t0, IngestSequence: 1},
		{SessionID: "s2", ServerReceivedAt: t0, IngestSequence: 2},
	}
	if err := repo.Append(context.Background(), in); err != nil {
		t.Fatalf("append: %v", err)
	}
	got := repo.Snapshot()
	if len(got) != 2 {
		t.Errorf("Snapshot len=%d want 2", len(got))
	}
	// Snapshot ist eine Kopie — Mutationen am Snapshot dürfen den
	// internen State nicht ändern.
	got[0].SessionID = "mutated"
	again := repo.Snapshot()
	if again[0].SessionID == "mutated" {
		t.Errorf("Snapshot returned a shared slice (mutation leaked)")
	}
}

// TestEventRepository_ListBySession_NilSequenceNumberSortsFirst
// deckt nullableSeqValue's nil-Pfad ab und verifiziert die
// Sort-Order: bei gleichem ServerReceivedAt sortiert ein Event mit
// nil SequenceNumber vor einem Event mit gesetzter Nummer
// (plan-0.1.0.md §5.1).
func TestEventRepository_ListBySession_NilSequenceNumberSortsFirst(t *testing.T) {
	t.Parallel()
	repo := inmemory.NewEventRepository()
	t0 := time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC)
	intp := func(v int64) *int64 { return &v }
	if err := repo.Append(context.Background(), []domain.PlaybackEvent{
		{ProjectID: "demo", SessionID: "s1", ServerReceivedAt: t0, SequenceNumber: intp(5), IngestSequence: 2},
		{ProjectID: "demo", SessionID: "s1", ServerReceivedAt: t0, SequenceNumber: nil, IngestSequence: 1},
	}); err != nil {
		t.Fatalf("append: %v", err)
	}
	page, err := repo.ListBySession(context.Background(), driven.EventListQuery{ProjectID: "demo",
		SessionID: "s1", Limit: 10,
	})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(page.Events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(page.Events))
	}
	if page.Events[0].SequenceNumber != nil {
		t.Errorf("nil-SequenceNumber must sort first; got SeqNum=%v as event[0]", page.Events[0].SequenceNumber)
	}
}

// TestEventRepository_ListBySession_FiltersBySessionID
// verifiziert, dass Events anderer Sessions nicht durchsickern.
func TestEventRepository_ListBySession_FiltersBySessionID(t *testing.T) {
	t.Parallel()
	repo := inmemory.NewEventRepository()
	t0 := time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC)
	if err := repo.Append(context.Background(), []domain.PlaybackEvent{
		{ProjectID: "demo", SessionID: "s1", ServerReceivedAt: t0, IngestSequence: 1},
		{ProjectID: "demo", SessionID: "s2", ServerReceivedAt: t0, IngestSequence: 2},
		{ProjectID: "demo", SessionID: "s1", ServerReceivedAt: t0.Add(time.Second), IngestSequence: 3},
	}); err != nil {
		t.Fatalf("append: %v", err)
	}
	page, err := repo.ListBySession(context.Background(), driven.EventListQuery{ProjectID: "demo",
		SessionID: "s1",
		Limit:     10,
	})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(page.Events) != 2 {
		t.Errorf("expected 2 events for s1, got %d", len(page.Events))
	}
	for _, e := range page.Events {
		if e.SessionID != "s1" {
			t.Errorf("foreign session leaked: %q", e.SessionID)
		}
	}
}
