package application_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/pt9912/m-trace/apps/api/hexagon/application"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driving"
)

// plan-0.4.0 §4.4 D2 — End-to-End-Tests für den
// `session_boundaries[]`-Wrapper. Use-Case validiert den Block atomar
// vor jedem Persist; ein invalider Block persistiert weder Events noch
// Boundaries (API-Kontrakt §3.4).

func validBoundary() driving.BoundaryInput {
	return driving.BoundaryInput{
		Kind:            domain.BoundaryKindNetworkSignalAbsent,
		ProjectID:       "demo",
		SessionID:       "01J7K9X4Z2QHB6V3WS5R8Y4D1F",
		NetworkKind:     "segment",
		Adapter:         "native_hls",
		Reason:          "native_hls_unavailable",
		ClientTimestamp: "2026-04-28T12:00:00.000Z",
	}
}

func TestRegisterBatch_AcceptsBoundariesAlongsideEvents(t *testing.T) {
	t.Parallel()
	uc, _, repo, sessions, _, _, _, _ := newUseCase()
	in := validBatch()
	in.Boundaries = []driving.BoundaryInput{validBoundary()}
	res, err := uc.RegisterPlaybackEventBatch(context.Background(), in)
	if err != nil {
		t.Fatalf("expected accept, got %v", err)
	}
	if res.Accepted != 1 {
		t.Fatalf("Accepted=%d want 1 (boundary darf nicht in accepted zählen)", res.Accepted)
	}
	if len(repo.appended) != 1 {
		t.Fatalf("expected 1 event persisted, got %d", len(repo.appended))
	}
	if len(sessions.boundaries) != 1 || len(sessions.boundaries[0]) != 1 {
		t.Fatalf("expected 1 boundary persisted, got %v", sessions.boundaries)
	}
	got := sessions.boundaries[0][0]
	if got.Kind != domain.BoundaryKindNetworkSignalAbsent ||
		got.NetworkKind != "segment" || got.Adapter != "native_hls" ||
		got.Reason != "native_hls_unavailable" {
		t.Errorf("boundary persisted with wrong fields: %+v", got)
	}
	if got.ServerReceivedAt.IsZero() {
		t.Error("ServerReceivedAt must be set by use case")
	}
}

func TestRegisterBatch_RejectsInvalidBoundary_NoPersist(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		mutate func(*driving.BoundaryInput)
	}{
		{"unknown kind", func(b *driving.BoundaryInput) { b.Kind = "totally_made_up" }},
		{"unknown network_kind", func(b *driving.BoundaryInput) { b.NetworkKind = "audio" }},
		{"unknown adapter", func(b *driving.BoundaryInput) { b.Adapter = "shaka" }},
		{"unknown reason enum", func(b *driving.BoundaryInput) { b.Reason = "totally_made_up" }},
		{"reason violates pattern", func(b *driving.BoundaryInput) { b.Reason = "BAD-VALUE" }},
		{"missing required field", func(b *driving.BoundaryInput) { b.SessionID = "" }},
		{"client timestamp not RFC3339", func(b *driving.BoundaryInput) { b.ClientTimestamp = "not-a-time" }},
		{"boundary project mismatch", func(b *driving.BoundaryInput) { b.ProjectID = "other" }},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			uc, _, repo, sessions, metrics, _, _, _ := newUseCase()
			in := validBatch()
			b := validBoundary()
			tc.mutate(&b)
			in.Boundaries = []driving.BoundaryInput{b}
			_, err := uc.RegisterPlaybackEventBatch(context.Background(), in)
			if !errors.Is(err, domain.ErrInvalidEvent) {
				t.Fatalf("expected ErrInvalidEvent, got %v", err)
			}
			if len(repo.appended) != 0 {
				t.Errorf("invalid boundary must not persist events, got %d", len(repo.appended))
			}
			if len(sessions.boundaries) != 0 {
				t.Errorf("invalid boundary must not persist boundaries, got %v", sessions.boundaries)
			}
			if metrics.invalid != len(in.Events) {
				t.Errorf("invalid_events should increment by batch size, got %d", metrics.invalid)
			}
		})
	}
}

func TestRegisterBatch_RejectsBoundaryForUnknownSession(t *testing.T) {
	t.Parallel()
	// Boundary referenziert eine session_id, für die im selben Batch
	// kein Event vorhanden ist — API-Kontrakt §3.4 verlangt 422.
	uc, _, repo, sessions, _, _, _, _ := newUseCase()
	in := validBatch()
	b := validBoundary()
	b.SessionID = "different-session-id"
	in.Boundaries = []driving.BoundaryInput{b}
	_, err := uc.RegisterPlaybackEventBatch(context.Background(), in)
	if !errors.Is(err, domain.ErrInvalidEvent) {
		t.Fatalf("expected ErrInvalidEvent, got %v", err)
	}
	if len(repo.appended) != 0 || len(sessions.boundaries) != 0 {
		t.Fatalf("orphan-session boundary must not persist anything")
	}
}

func TestRegisterBatch_RejectsTooManyBoundaries(t *testing.T) {
	t.Parallel()
	uc, _, repo, sessions, _, _, _, _ := newUseCase()
	in := validBatch()
	in.Boundaries = make([]driving.BoundaryInput, application.MaxSessionBoundaries+1)
	for i := range in.Boundaries {
		in.Boundaries[i] = validBoundary()
	}
	_, err := uc.RegisterPlaybackEventBatch(context.Background(), in)
	if !errors.Is(err, domain.ErrInvalidEvent) {
		t.Fatalf("expected ErrInvalidEvent, got %v", err)
	}
	if !strings.Contains(err.Error(), "session_boundaries") {
		t.Errorf("expected boundary-overrun hint, got %v", err)
	}
	if len(repo.appended) != 0 || len(sessions.boundaries) != 0 {
		t.Fatalf("over-limit boundaries must not persist anything")
	}
}

func TestRegisterBatch_BoundaryOnlyEmptyEventsStillFails(t *testing.T) {
	t.Parallel()
	// Boundary-only-Batches ohne Events bleiben außerhalb des Vertrags
	// (API-Kontrakt §3.4 — `events` muss min. 1 Eintrag haben).
	uc, _, repo, sessions, _, _, _, _ := newUseCase()
	in := validBatch()
	in.Events = nil
	in.Boundaries = []driving.BoundaryInput{validBoundary()}
	_, err := uc.RegisterPlaybackEventBatch(context.Background(), in)
	if !errors.Is(err, domain.ErrBatchEmpty) {
		t.Fatalf("expected ErrBatchEmpty, got %v", err)
	}
	if len(repo.appended) != 0 || len(sessions.boundaries) != 0 {
		t.Fatalf("boundary-only batch must not persist anything")
	}
}

func TestRegisterBatch_EmptyBoundariesIsNoop(t *testing.T) {
	t.Parallel()
	uc, _, _, sessions, _, _, _, _ := newUseCase()
	in := validBatch()
	in.Boundaries = nil
	if _, err := uc.RegisterPlaybackEventBatch(context.Background(), in); err != nil {
		t.Fatalf("expected accept, got %v", err)
	}
	// AppendBoundaries darf trotzdem aufgerufen worden sein, aber mit
	// leerem Slice — der Adapter behandelt das als no-op. Wir prüfen
	// hier nur, dass der Aufruf nicht mit nicht-leeren Boundaries
	// landete.
	for _, batch := range sessions.boundaries {
		if len(batch) != 0 {
			t.Errorf("expected no boundary persists, got %v", batch)
		}
	}
}
