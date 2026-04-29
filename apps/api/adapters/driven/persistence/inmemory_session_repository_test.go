package persistence_test

import (
	"context"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/persistence"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// TestInMemorySessionRepository_UpsertFromEvents_CreateAndUpdate
// verifiziert die zwei Code-Pfade des Adapters: (1) erste Session-
// Beobachtung legt an, (2) Folge-Events derselben session_id zählen
// EventCount hoch und schieben LastEventAt vor.
func TestInMemorySessionRepository_UpsertFromEvents_CreateAndUpdate(t *testing.T) {
	t.Parallel()
	repo := persistence.NewInMemorySessionRepository()

	t0 := time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC)
	t1 := t0.Add(1 * time.Second)
	t2 := t0.Add(3 * time.Second)

	first := []domain.PlaybackEvent{{
		SessionID:        "sess-A",
		ProjectID:        "demo",
		ServerReceivedAt: t0,
	}}
	if err := repo.UpsertFromEvents(context.Background(), first); err != nil {
		t.Fatalf("first upsert: %v", err)
	}

	second := []domain.PlaybackEvent{
		{SessionID: "sess-A", ProjectID: "demo", ServerReceivedAt: t1},
		{SessionID: "sess-B", ProjectID: "demo", ServerReceivedAt: t2},
	}
	if err := repo.UpsertFromEvents(context.Background(), second); err != nil {
		t.Fatalf("second upsert: %v", err)
	}

	got := indexByID(repo.Snapshot())
	if len(got) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(got))
	}

	a := got["sess-A"]
	if a.State != domain.SessionStateActive {
		t.Errorf("sess-A.State=%q want %q", a.State, domain.SessionStateActive)
	}
	if !a.StartedAt.Equal(t0) {
		t.Errorf("sess-A.StartedAt=%v want %v", a.StartedAt, t0)
	}
	if !a.LastEventAt.Equal(t1) {
		t.Errorf("sess-A.LastEventAt=%v want %v", a.LastEventAt, t1)
	}
	if a.EventCount != 2 {
		t.Errorf("sess-A.EventCount=%d want 2", a.EventCount)
	}

	b := got["sess-B"]
	if !b.StartedAt.Equal(t2) || !b.LastEventAt.Equal(t2) {
		t.Errorf("sess-B times: started=%v last=%v want both %v", b.StartedAt, b.LastEventAt, t2)
	}
	if b.EventCount != 1 {
		t.Errorf("sess-B.EventCount=%d want 1", b.EventCount)
	}
}

// TestInMemorySessionRepository_EmptyEventsIsNoop verifiziert, dass
// ein leerer Slice keinen Fehler wirft und nichts ändert.
func TestInMemorySessionRepository_EmptyEventsIsNoop(t *testing.T) {
	t.Parallel()
	repo := persistence.NewInMemorySessionRepository()
	if err := repo.UpsertFromEvents(context.Background(), nil); err != nil {
		t.Errorf("nil upsert: %v", err)
	}
	if got := repo.Snapshot(); len(got) != 0 {
		t.Errorf("expected 0 sessions after no-op, got %d", len(got))
	}
}

func indexByID(in []domain.StreamSession) map[string]domain.StreamSession {
	out := make(map[string]domain.StreamSession, len(in))
	for _, s := range in {
		out[s.ID] = s
	}
	return out
}

// TestInMemorySessionRepository_List_SortAndCursor verifiziert, dass
// List in (started_at desc, session_id asc) sortiert und dass der
// After-Cursor die nächste Page strikt hinter dem letzten Eintrag der
// vorherigen aufnimmt — Pagination ohne Duplikate, ohne Lücken.
func TestInMemorySessionRepository_List_SortAndCursor(t *testing.T) {
	t.Parallel()
	repo := persistence.NewInMemorySessionRepository()
	t0 := time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC)

	// 4 sessions: s4@t0+3, s3@t0+2, s2@t0+1, s1@t0 → sort desc → s4,s3,s2,s1.
	for i, id := range []string{"s1", "s2", "s3", "s4"} {
		err := repo.UpsertFromEvents(context.Background(), []domain.PlaybackEvent{{
			SessionID:        id,
			ProjectID:        "demo",
			ServerReceivedAt: t0.Add(time.Duration(i) * time.Second),
		}})
		if err != nil {
			t.Fatalf("upsert %s: %v", id, err)
		}
	}

	first, err := repo.List(context.Background(), driven.SessionListQuery{Limit: 2})
	if err != nil {
		t.Fatalf("list page 1: %v", err)
	}
	if len(first.Sessions) != 2 {
		t.Fatalf("page 1: expected 2 sessions, got %d", len(first.Sessions))
	}
	if first.Sessions[0].ID != "s4" || first.Sessions[1].ID != "s3" {
		t.Errorf("page 1 order: got %v want [s4 s3]", []string{first.Sessions[0].ID, first.Sessions[1].ID})
	}
	if first.NextAfter == nil {
		t.Fatalf("page 1 expected NextAfter")
	}

	second, err := repo.List(context.Background(), driven.SessionListQuery{Limit: 2, After: first.NextAfter})
	if err != nil {
		t.Fatalf("list page 2: %v", err)
	}
	if len(second.Sessions) != 2 {
		t.Fatalf("page 2: expected 2 sessions, got %d", len(second.Sessions))
	}
	if second.Sessions[0].ID != "s2" || second.Sessions[1].ID != "s1" {
		t.Errorf("page 2 order: got %v want [s2 s1]", []string{second.Sessions[0].ID, second.Sessions[1].ID})
	}
	if second.NextAfter != nil {
		t.Errorf("page 2 should be last (NextAfter=nil), got %v", second.NextAfter)
	}
}

// TestInMemorySessionRepository_Get_NotFound deckt den Pflicht-Pfad
// für 404-Mapping in plan-0.1.0.md §5.1 ab.
func TestInMemorySessionRepository_Get_NotFound(t *testing.T) {
	t.Parallel()
	repo := persistence.NewInMemorySessionRepository()
	_, err := repo.Get(context.Background(), "nope")
	if err != domain.ErrSessionNotFound {
		t.Errorf("expected ErrSessionNotFound, got %v", err)
	}
}
