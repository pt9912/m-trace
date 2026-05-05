package sqlite_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/persistence/sqlite"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
	"github.com/pt9912/m-trace/apps/api/internal/storage"
)

// TestRestartPreservesData verifiziert die Restart-Stabilität (DoD
// `plan-0.4.0.md` §2.3): nach Close + Re-Open derselben SQLite-Datei
// sind Sessions und Events erhalten, und der Sequencer resumiert beim
// höchsten persistierten ingest_sequence + 1.
func TestRestartPreservesData(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "m-trace.db")
	t0 := time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)

	// Pass 1: zwei Events schreiben.
	db1, err := storage.Open(ctx, path)
	if err != nil {
		t.Fatalf("open #1: %v", err)
	}
	seq1, err := sqlite.NewIngestSequencer(ctx, db1)
	if err != nil {
		t.Fatalf("seq #1: %v", err)
	}
	sess1 := sqlite.NewSessionRepository(db1)
	evt1 := sqlite.NewEventRepository(db1)

	events := []domain.PlaybackEvent{
		mkRestartEvent(seq1, "demo", "s1", t0),
		mkRestartEvent(seq1, "demo", "s1", t0.Add(1*time.Second)),
	}
	if _, err := sess1.UpsertFromEvents(ctx, events); err != nil {
		t.Fatalf("upsert #1: %v", err)
	}
	if err := evt1.Append(ctx, events); err != nil {
		t.Fatalf("append #1: %v", err)
	}
	if err := db1.Close(); err != nil {
		t.Fatalf("close #1: %v", err)
	}

	// Pass 2: re-open, prüfen.
	db2, err := storage.Open(ctx, path)
	if err != nil {
		t.Fatalf("open #2: %v", err)
	}
	t.Cleanup(func() { _ = db2.Close() })
	seq2, err := sqlite.NewIngestSequencer(ctx, db2)
	if err != nil {
		t.Fatalf("seq #2: %v", err)
	}

	if next := seq2.Next(); next != 3 {
		t.Errorf("after restart Next() = %d, want 3 (max persisted +1)", next)
	}

	sess2 := sqlite.NewSessionRepository(db2)
	got, err := sess2.Get(ctx, "demo", "s1")
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if got.EventCount != 2 {
		t.Errorf("event_count = %d, want 2 (preserved across restart)", got.EventCount)
	}

	evt2 := sqlite.NewEventRepository(db2)
	page, err := evt2.ListBySession(ctx, driven.EventListQuery{ProjectID: "demo",
		SessionID: "s1", Limit: 10,
	})
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	if len(page.Events) != 2 {
		t.Fatalf("len(events) = %d, want 2", len(page.Events))
	}
}

// TestRestartCursorStability verifiziert, dass ein Cursor, der vor
// dem Restart erzeugt wurde, nach dem Restart weiter funktioniert
// (kanonische Sortierung restart-stabil — ADR-0002 §8.1, ADR-0004 §5).
func TestRestartCursorStability(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "m-trace.db")
	t0 := time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)

	// Pass 1: 4 Events schreiben, page 1 (limit=2) lesen, NextAfter
	// merken.
	db1, err := storage.Open(ctx, path)
	if err != nil {
		t.Fatalf("open #1: %v", err)
	}
	seq1, err := sqlite.NewIngestSequencer(ctx, db1)
	if err != nil {
		t.Fatalf("seq #1: %v", err)
	}
	sess1 := sqlite.NewSessionRepository(db1)
	evt1 := sqlite.NewEventRepository(db1)

	events := make([]domain.PlaybackEvent, 0, 4)
	for i := 0; i < 4; i++ {
		events = append(events, mkRestartEvent(seq1, "demo", "s1",
			t0.Add(time.Duration(i)*time.Second)))
	}
	if _, err := sess1.UpsertFromEvents(ctx, events); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if err := evt1.Append(ctx, events); err != nil {
		t.Fatalf("append: %v", err)
	}

	page1, err := evt1.ListBySession(ctx, driven.EventListQuery{ProjectID: "demo",
		SessionID: "s1", Limit: 2,
	})
	if err != nil {
		t.Fatalf("list page1: %v", err)
	}
	cursor := page1.NextAfter
	if cursor == nil {
		t.Fatalf("page1.NextAfter = nil, want set")
	}
	if err := db1.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	// Pass 2: re-open, mit dem alten Cursor weiterpaginieren.
	db2, err := storage.Open(ctx, path)
	if err != nil {
		t.Fatalf("open #2: %v", err)
	}
	t.Cleanup(func() { _ = db2.Close() })
	evt2 := sqlite.NewEventRepository(db2)

	page2, err := evt2.ListBySession(ctx, driven.EventListQuery{ProjectID: "demo",
		SessionID: "s1", Limit: 10, After: cursor,
	})
	if err != nil {
		t.Fatalf("list page2: %v", err)
	}
	if len(page2.Events) != 2 {
		t.Errorf("page2 = %d events, want 2 (remainder after pre-restart cursor)",
			len(page2.Events))
	}
	for i, e := range page2.Events {
		want := t0.Add(time.Duration(i+2) * time.Second)
		if !e.ServerReceivedAt.Equal(want) {
			t.Errorf("page2[%d] rcv = %v, want %v", i, e.ServerReceivedAt, want)
		}
	}
}

// TestRestartPreservesEndSource verifiziert plan-0.4.0 §5 H1
// (Restart-Stabilität für `end_source`): nach Close + Re-Open ist
// das Feld identisch lesbar — sowohl `client` (explizites
// session_ended) als auch `sweeper` (zeitbasiertes Ende).
func TestRestartPreservesEndSource(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "m-trace-endsource.db")
	t0 := time.Date(2026, 5, 4, 14, 0, 0, 0, time.UTC)

	// Pass 1: zwei Sessions — eine via client-event, eine via sweeper.
	db1, err := storage.Open(ctx, path)
	if err != nil {
		t.Fatalf("open #1: %v", err)
	}
	seq1, err := sqlite.NewIngestSequencer(ctx, db1)
	if err != nil {
		t.Fatalf("seq #1: %v", err)
	}
	sess1 := sqlite.NewSessionRepository(db1)

	// Session "client-end" mit explizitem session_ended.
	startedClient := mkRestartEvent(seq1, "demo", "client-end", t0)
	endedClient := mkRestartEvent(seq1, "demo", "client-end", t0.Add(time.Second))
	endedClient.EventName = "session_ended"
	if _, err := sess1.UpsertFromEvents(ctx, []domain.PlaybackEvent{startedClient, endedClient}); err != nil {
		t.Fatalf("upsert client-end: %v", err)
	}

	// Session "sweep-end" — wird durch Sweep nach `endedAfter` beendet.
	startedSweep := mkRestartEvent(seq1, "demo", "sweep-end", t0)
	if _, err := sess1.UpsertFromEvents(ctx, []domain.PlaybackEvent{startedSweep}); err != nil {
		t.Fatalf("upsert sweep-end: %v", err)
	}
	// Sweep mit großem now-Offset → stalled → ended.
	if err := sess1.Sweep(ctx, t0.Add(20*time.Minute), 60*time.Second, 600*time.Second); err != nil {
		t.Fatalf("sweep: %v", err)
	}
	if err := db1.Close(); err != nil {
		t.Fatalf("close #1: %v", err)
	}

	// Pass 2: re-open und end_source prüfen.
	db2, err := storage.Open(ctx, path)
	if err != nil {
		t.Fatalf("open #2: %v", err)
	}
	t.Cleanup(func() { _ = db2.Close() })
	sess2 := sqlite.NewSessionRepository(db2)

	gotClient, err := sess2.Get(ctx, "demo", "client-end")
	if err != nil {
		t.Fatalf("get client-end: %v", err)
	}
	if gotClient.EndSource != domain.SessionEndSourceClient {
		t.Errorf("client-end end_source = %q, want %q", gotClient.EndSource, domain.SessionEndSourceClient)
	}

	gotSweep, err := sess2.Get(ctx, "demo", "sweep-end")
	if err != nil {
		t.Fatalf("get sweep-end: %v", err)
	}
	if gotSweep.EndSource != domain.SessionEndSourceSweeper {
		t.Errorf("sweep-end end_source = %q, want %q", gotSweep.EndSource, domain.SessionEndSourceSweeper)
	}
}

// TestRestartPreservesSessionBoundaries verifiziert plan-0.4.0 §4.4
// DoD-Item 3 (Restart-Stabilität): persistierte
// `session_boundaries[]`-Records sind nach Close + Re-Open derselben
// SQLite-Datei identisch lesbar — kein In-Memory-only-Pfad.
func TestRestartPreservesSessionBoundaries(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "m-trace-boundaries.db")
	t0 := time.Date(2026, 5, 4, 12, 0, 0, 0, time.UTC)

	// Pass 1: Session anlegen + zwei distinct Boundary-Tripel +
	// Duplikat des ersten Tripels mit späterem Timestamp (Idempotenz).
	db1, err := storage.Open(ctx, path)
	if err != nil {
		t.Fatalf("open #1: %v", err)
	}
	seq1, err := sqlite.NewIngestSequencer(ctx, db1)
	if err != nil {
		t.Fatalf("seq #1: %v", err)
	}
	sess1 := sqlite.NewSessionRepository(db1)
	if _, err := sess1.UpsertFromEvents(ctx, []domain.PlaybackEvent{
		mkRestartEvent(seq1, "demo", "s1", t0),
	}); err != nil {
		t.Fatalf("seed session: %v", err)
	}
	first := domain.SessionBoundary{
		Kind: domain.BoundaryKindNetworkSignalAbsent, ProjectID: "demo", SessionID: "s1",
		NetworkKind: "segment", Adapter: "native_hls", Reason: "native_hls_unavailable",
		ClientTimestamp: t0, ServerReceivedAt: t0,
	}
	second := domain.SessionBoundary{
		Kind: domain.BoundaryKindNetworkSignalAbsent, ProjectID: "demo", SessionID: "s1",
		NetworkKind: "manifest", Adapter: "hls.js", Reason: "cors_timing_blocked",
		ClientTimestamp: t0, ServerReceivedAt: t0,
	}
	dup := first
	dup.ServerReceivedAt = t0.Add(time.Hour)
	if err := sess1.AppendBoundaries(ctx, []domain.SessionBoundary{first, second, dup}); err != nil {
		t.Fatalf("AppendBoundaries: %v", err)
	}
	if err := db1.Close(); err != nil {
		t.Fatalf("close #1: %v", err)
	}

	// Pass 2: re-open, Boundaries müssen identisch lesbar sein.
	db2, err := storage.Open(ctx, path)
	if err != nil {
		t.Fatalf("open #2: %v", err)
	}
	t.Cleanup(func() { _ = db2.Close() })
	sess2 := sqlite.NewSessionRepository(db2)
	got, err := sess2.ListBoundariesForSession(ctx, "demo", "s1")
	if err != nil {
		t.Fatalf("ListBoundariesForSession #2: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 deduped boundaries after restart, got %d (%+v)", len(got), got)
	}
	// Sortierung kind asc → adapter asc → reason asc.
	if got[0].Adapter != "hls.js" || got[0].Reason != "cors_timing_blocked" {
		t.Errorf("got[0] = %+v want adapter=hls.js reason=cors_timing_blocked", got[0])
	}
	if got[1].Adapter != "native_hls" || got[1].Reason != "native_hls_unavailable" {
		t.Errorf("got[1] = %+v want adapter=native_hls reason=native_hls_unavailable", got[1])
	}
	// Duplikat hat den späteren Timestamp persistiert.
	if !got[1].ServerReceivedAt.Equal(t0.Add(time.Hour)) {
		t.Errorf("duplicate triple should refresh server_received_at across restart, got %v", got[1].ServerReceivedAt)
	}
}

func mkRestartEvent(seq driven.IngestSequencer, project, session string, recv time.Time) domain.PlaybackEvent {
	return domain.PlaybackEvent{
		EventName:        "playback_started",
		ProjectID:        project,
		SessionID:        session,
		ClientTimestamp:  recv,
		ServerReceivedAt: recv,
		IngestSequence:   seq.Next(),
		SDK:              domain.SDKInfo{Name: "@npm9912/player-sdk", Version: "0.5.0"},
	}
}
