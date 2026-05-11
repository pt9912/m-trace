package inmemory_test

import (
	"context"
	"testing"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/persistence/inmemory"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// TestSessionRepository_SetSessionSampleRatePPMIfDefault_FirstSet
// (plan-0.12.6 Tranche 4 / R-10): erstes Setzen auf einer Session
// mit Default-Wert (SampleRateFull) erfolgreich; applied=true.
func TestSessionRepository_SetSessionSampleRatePPMIfDefault_FirstSet(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := inmemory.NewSessionRepository()
	// Session via UpsertFromEvents anlegen (Default SampleRateFull).
	if _, err := repo.UpsertFromEvents(ctx, []domain.PlaybackEvent{
		{ProjectID: "demo", SessionID: "s1", EventName: "playback_started"},
	}); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	got, applied, err := repo.SetSessionSampleRatePPMIfDefault(ctx, "demo", "s1", 500_000)
	if err != nil {
		t.Fatalf("set: %v", err)
	}
	if !applied {
		t.Errorf("applied = false, want true (first set on default)")
	}
	if got != 500_000 {
		t.Errorf("got = %d, want 500000", got)
	}
}

// TestSessionRepository_SetSessionSampleRatePPMIfDefault_AlreadySet
// (plan-0.12.6 Tranche 4): zweites Setzen lässt den ersten Wert
// unverändert; applied=false; got = existing.
func TestSessionRepository_SetSessionSampleRatePPMIfDefault_AlreadySet(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := inmemory.NewSessionRepository()
	if _, err := repo.UpsertFromEvents(ctx, []domain.PlaybackEvent{
		{ProjectID: "demo", SessionID: "s1", EventName: "playback_started"},
	}); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if _, _, err := repo.SetSessionSampleRatePPMIfDefault(ctx, "demo", "s1", 500_000); err != nil {
		t.Fatalf("first set: %v", err)
	}

	got, applied, err := repo.SetSessionSampleRatePPMIfDefault(ctx, "demo", "s1", 250_000)
	if err != nil {
		t.Fatalf("second set: %v", err)
	}
	if applied {
		t.Errorf("applied = true, want false (already set)")
	}
	if got != 500_000 {
		t.Errorf("got = %d, want 500000 (immutable)", got)
	}
}

// TestSessionRepository_SetSessionSampleRatePPMIfDefault_NoOpFull
// (plan-0.12.6 Tranche 4): ppm == SampleRateFull ist No-Op und
// returnt (SampleRateFull, false, nil).
func TestSessionRepository_SetSessionSampleRatePPMIfDefault_NoOpFull(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := inmemory.NewSessionRepository()
	got, applied, err := repo.SetSessionSampleRatePPMIfDefault(ctx, "demo", "s1", domain.SampleRateFull)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if applied {
		t.Errorf("applied = true, want false (no-op for SampleRateFull)")
	}
	if got != domain.SampleRateFull {
		t.Errorf("got = %d, want SampleRateFull", got)
	}
}

// TestSessionRepository_SetSessionSampleRatePPMIfDefault_UnknownSession
// (plan-0.12.6 Tranche 4): Aufruf auf nicht-existenter Session liefert
// (0, false, nil) — defensiv; Aufrufer ruft die Methode nach
// UpsertFromEvents, also sollte das nicht vorkommen.
func TestSessionRepository_SetSessionSampleRatePPMIfDefault_UnknownSession(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := inmemory.NewSessionRepository()
	got, applied, err := repo.SetSessionSampleRatePPMIfDefault(ctx, "demo", "missing", 500_000)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if applied {
		t.Errorf("applied = true, want false (no session)")
	}
	if got != 0 {
		t.Errorf("got = %d, want 0", got)
	}
}
