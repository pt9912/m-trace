package inmemory_test

import (
	"context"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/persistence/inmemory"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// TestSessionRepository_SetSessionSampleRatePPMIfDefault_FirstSet
// (R-10): erstes Setzen auf einer Session
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
// : zweites Setzen lässt den ersten Wert
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
// : ppm == SampleRateFull ist No-Op und
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
// : Aufruf auf nicht-existenter Session liefert
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

// TestSessionRepository_ListBoundariesForSessions (
// / R-7): Bulk-Variante liefert pro SessionID die
// sortierten Boundaries; SessionIDs ohne Boundaries fehlen in der
// Map; Cross-Project-Scope wird respektiert; leere Input-Liste ist
// ein No-Op.
func TestSessionRepository_ListBoundariesForSessions(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := inmemory.NewSessionRepository()
	t0 := time.Date(2026, 5, 4, 12, 0, 0, 0, time.UTC)

	mkSession := func(project, id string) domain.PlaybackEvent {
		return domain.PlaybackEvent{
			ProjectID: project, SessionID: id, EventName: "playback_started",
			ServerReceivedAt: t0,
		}
	}
	if _, err := repo.UpsertFromEvents(ctx, []domain.PlaybackEvent{
		mkSession("demo", "s1"),
		mkSession("demo", "s2"),
		mkSession("demo", "s3"),
		mkSession("other", "s1"),
	}); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if err := repo.AppendBoundaries(ctx, []domain.SessionBoundary{
		{Kind: domain.BoundaryKindNetworkSignalAbsent, ProjectID: "demo", SessionID: "s1",
			NetworkKind: "segment", Adapter: "native_hls", Reason: "native_hls_unavailable",
			ClientTimestamp: t0, ServerReceivedAt: t0},
		{Kind: domain.BoundaryKindNetworkSignalAbsent, ProjectID: "demo", SessionID: "s1",
			NetworkKind: "manifest", Adapter: "hls.js", Reason: "cors_timing_blocked",
			ClientTimestamp: t0, ServerReceivedAt: t0},
		{Kind: domain.BoundaryKindNetworkSignalAbsent, ProjectID: "demo", SessionID: "s3",
			NetworkKind: "manifest", Adapter: "hls.js", Reason: "cors_timing_blocked",
			ClientTimestamp: t0, ServerReceivedAt: t0},
		// Cross-Project-Probe: gleiche session_id in „other"-Project.
		{Kind: domain.BoundaryKindNetworkSignalAbsent, ProjectID: "other", SessionID: "s1",
			NetworkKind: "manifest", Adapter: "hls.js", Reason: "cors_timing_blocked",
			ClientTimestamp: t0, ServerReceivedAt: t0},
	}); err != nil {
		t.Fatalf("append: %v", err)
	}

	got, err := repo.ListBoundariesForSessions(ctx, "demo", []string{"s1", "s2", "s3"})
	if err != nil {
		t.Fatalf("bulk read: %v", err)
	}
	if len(got["s1"]) != 2 {
		t.Fatalf("s1 = %d, want 2", len(got["s1"]))
	}
	// Sort: kind asc → adapter asc → reason asc.
	if got["s1"][0].Adapter != "hls.js" || got["s1"][1].Adapter != "native_hls" {
		t.Errorf("s1 sort: got %+v", got["s1"])
	}
	if _, ok := got["s2"]; ok {
		t.Errorf("s2 should be missing from map, got %+v", got["s2"])
	}
	if len(got["s3"]) != 1 {
		t.Errorf("s3 = %d, want 1", len(got["s3"]))
	}
	// Cross-Project-Isolation.
	for sid, bs := range got {
		for _, b := range bs {
			if b.ProjectID != "demo" {
				t.Errorf("session %q: cross-project boundary leaked: %+v", sid, b)
			}
		}
	}

	// Empty input → empty map, no allocation.
	emptyGot, err := repo.ListBoundariesForSessions(ctx, "demo", nil)
	if err != nil {
		t.Fatalf("empty input: %v", err)
	}
	if len(emptyGot) != 0 {
		t.Errorf("expected empty map, got %+v", emptyGot)
	}
}
