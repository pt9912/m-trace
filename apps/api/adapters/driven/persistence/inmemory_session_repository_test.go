package persistence_test

import (
	"context"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/persistence"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
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
