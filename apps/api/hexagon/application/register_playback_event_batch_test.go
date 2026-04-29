package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/application"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driving"
)

// stubProjectResolver returns a single demo project for "demo-token",
// nothing else (matches the contract's hardcoded map).
type stubProjectResolver struct{}

func (stubProjectResolver) ResolveByToken(_ context.Context, token string) (domain.Project, error) {
	if token == "demo-token" {
		return domain.Project{ID: "demo", Token: domain.ProjectToken("demo-token")}, nil
	}
	return domain.Project{}, domain.ErrUnauthorized
}

type stubLimiter struct {
	deny bool
}

func (s *stubLimiter) Allow(_ context.Context, _ string, _ int) error {
	if s.deny {
		return domain.ErrRateLimited
	}
	return nil
}

type stubRepo struct {
	appended []domain.PlaybackEvent
	failNext bool
}

func (s *stubRepo) Append(_ context.Context, events []domain.PlaybackEvent) error {
	if s.failNext {
		s.failNext = false
		return errors.New("repo failure")
	}
	s.appended = append(s.appended, events...)
	return nil
}

type spyMetrics struct {
	accepted, invalid, rateLimited, dropped int
}

func (s *spyMetrics) EventsAccepted(n int)    { s.accepted += n }
func (s *spyMetrics) InvalidEvents(n int)     { s.invalid += n }
func (s *spyMetrics) RateLimitedEvents(n int) { s.rateLimited += n }
func (s *spyMetrics) DroppedEvents(n int)     { s.dropped += n }

// stubTelemetry zählt BatchReceived-Aufrufe. Pro Aufruf wird die
// gemeldete Batch-Größe addiert; calls misst die reine Aufrufzahl.
type stubTelemetry struct {
	calls       int
	totalSize   int
	lastSize    int
}

func (s *stubTelemetry) BatchReceived(_ context.Context, size int) {
	s.calls++
	s.totalSize += size
	s.lastSize = size
}

func validBatch() driving.BatchInput {
	return driving.BatchInput{
		SchemaVersion: application.SupportedSchemaVersion,
		AuthToken:     "demo-token",
		Events: []driving.EventInput{
			{
				EventName:       "rebuffer_started",
				ProjectID:       "demo",
				SessionID:       "01J7K9X4Z2QHB6V3WS5R8Y4D1F",
				ClientTimestamp: "2026-04-28T12:00:00.000Z",
				SDK:             driving.SDKInput{Name: "@m-trace/player-sdk", Version: "0.1.0"},
			},
		},
	}
}

func newUseCase() (*application.RegisterPlaybackEventBatchUseCase, *stubLimiter, *stubRepo, *spyMetrics, *stubTelemetry) {
	limiter := &stubLimiter{}
	repo := &stubRepo{}
	metrics := &spyMetrics{}
	telemetry := &stubTelemetry{}
	uc := application.NewRegisterPlaybackEventBatchUseCase(
		stubProjectResolver{}, limiter, repo, metrics, telemetry,
		func() time.Time { return time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC) },
	)
	return uc, limiter, repo, metrics, telemetry
}

func TestHappyPath(t *testing.T) {
	t.Parallel()
	uc, _, repo, metrics, telemetry := newUseCase()
	res, err := uc.RegisterPlaybackEventBatch(context.Background(), validBatch())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Accepted != 1 {
		t.Errorf("expected 1 accepted, got %d", res.Accepted)
	}
	if len(repo.appended) != 1 {
		t.Errorf("expected 1 appended event, got %d", len(repo.appended))
	}
	if metrics.accepted != 1 {
		t.Errorf("expected EventsAccepted=1, got %d", metrics.accepted)
	}
	if telemetry.calls != 1 {
		t.Errorf("expected Telemetry.BatchReceived calls=1, got %d", telemetry.calls)
	}
	if telemetry.lastSize != 1 {
		t.Errorf("expected Telemetry.lastSize=1, got %d", telemetry.lastSize)
	}
}

// TestTelemetryReceivedBeforeAuth verifiziert, dass BatchReceived auch
// bei fehlgeschlagener Auth gerufen wird (Counter misst received,
// nicht validated — siehe Telemetry-Port-Doc).
func TestTelemetryReceivedBeforeAuth(t *testing.T) {
	t.Parallel()
	uc, _, _, _, telemetry := newUseCase()
	in := validBatch()
	in.AuthToken = "wrong-token"
	_, err := uc.RegisterPlaybackEventBatch(context.Background(), in)
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
	if telemetry.calls != 1 {
		t.Errorf("expected Telemetry.BatchReceived calls=1 (received zählt vor Auth), got %d", telemetry.calls)
	}
}

func TestUnauthorizedToken(t *testing.T) {
	t.Parallel()
	uc, _, _, _, _ := newUseCase()
	in := validBatch()
	in.AuthToken = "wrong-token"
	_, err := uc.RegisterPlaybackEventBatch(context.Background(), in)
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestSchemaVersionMismatch(t *testing.T) {
	t.Parallel()
	uc, _, _, metrics, _ := newUseCase()
	in := validBatch()
	in.SchemaVersion = "2.0"
	_, err := uc.RegisterPlaybackEventBatch(context.Background(), in)
	if !errors.Is(err, domain.ErrSchemaVersionMismatch) {
		t.Errorf("expected ErrSchemaVersionMismatch, got %v", err)
	}
	if metrics.invalid != 1 {
		t.Errorf("expected InvalidEvents=1, got %d", metrics.invalid)
	}
}

func TestEmptyBatch(t *testing.T) {
	t.Parallel()
	uc, _, _, metrics, _ := newUseCase()
	in := validBatch()
	in.Events = nil
	_, err := uc.RegisterPlaybackEventBatch(context.Background(), in)
	if !errors.Is(err, domain.ErrBatchEmpty) {
		t.Errorf("expected ErrBatchEmpty, got %v", err)
	}
	// Counter zählt Events, nicht Batches — bei n=0 kein Increment
	// (Lastenheft 1.1.2 §7.9).
	if metrics.invalid != 0 {
		t.Errorf("expected InvalidEvents=0 (empty batch counts no events), got %d", metrics.invalid)
	}
}

func TestBatchTooLarge(t *testing.T) {
	t.Parallel()
	uc, _, _, metrics, _ := newUseCase()
	in := validBatch()
	template := in.Events[0]
	in.Events = make([]driving.EventInput, application.MaxBatchSize+1)
	for i := range in.Events {
		in.Events[i] = template
	}
	_, err := uc.RegisterPlaybackEventBatch(context.Background(), in)
	if !errors.Is(err, domain.ErrBatchTooLarge) {
		t.Errorf("expected ErrBatchTooLarge, got %v", err)
	}
	if metrics.invalid != application.MaxBatchSize+1 {
		t.Errorf("expected InvalidEvents=%d, got %d", application.MaxBatchSize+1, metrics.invalid)
	}
}

func TestInvalidEventMissingField(t *testing.T) {
	t.Parallel()
	uc, _, _, _, _ := newUseCase()
	in := validBatch()
	in.Events[0].EventName = ""
	_, err := uc.RegisterPlaybackEventBatch(context.Background(), in)
	if !errors.Is(err, domain.ErrInvalidEvent) {
		t.Errorf("expected ErrInvalidEvent, got %v", err)
	}
}

func TestInvalidEventBadTimestamp(t *testing.T) {
	t.Parallel()
	uc, _, _, _, _ := newUseCase()
	in := validBatch()
	in.Events[0].ClientTimestamp = "not-a-timestamp"
	_, err := uc.RegisterPlaybackEventBatch(context.Background(), in)
	if !errors.Is(err, domain.ErrInvalidEvent) {
		t.Errorf("expected ErrInvalidEvent, got %v", err)
	}
}

func TestProjectIDTokenMismatch(t *testing.T) {
	t.Parallel()
	uc, _, _, metrics, _ := newUseCase()
	in := validBatch()
	in.Events[0].ProjectID = "other-project"
	_, err := uc.RegisterPlaybackEventBatch(context.Background(), in)
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
	// Auth-Fehler (401) zählen nicht in invalid_events (API-Kontrakt §7).
	if metrics.invalid != 0 {
		t.Errorf("expected InvalidEvents=0 (auth-Fehler zählen nicht), got %d", metrics.invalid)
	}
}

func TestRateLimited(t *testing.T) {
	t.Parallel()
	uc, limiter, _, metrics, _ := newUseCase()
	limiter.deny = true
	_, err := uc.RegisterPlaybackEventBatch(context.Background(), validBatch())
	if !errors.Is(err, domain.ErrRateLimited) {
		t.Errorf("expected ErrRateLimited, got %v", err)
	}
	if metrics.rateLimited != 1 {
		t.Errorf("expected RateLimitedEvents=1, got %d", metrics.rateLimited)
	}
}

func TestRepoFailureDoesNotCountAsDropped(t *testing.T) {
	t.Parallel()
	uc, _, repo, metrics, _ := newUseCase()
	repo.failNext = true
	_, err := uc.RegisterPlaybackEventBatch(context.Background(), validBatch())
	if err == nil {
		t.Fatal("expected an error")
	}
	// Synchron fehlgeschlagenes Append ist kein Backpressure-Drop;
	// dropped_events bleibt unverändert (API-Kontrakt §7,
	// Lastenheft 1.1.2 §7.9 nach Plan §4.2).
	if metrics.dropped != 0 {
		t.Errorf("expected DroppedEvents=0 (synchron fehlgeschlagenes Append ist kein Backpressure-Drop), got %d", metrics.dropped)
	}
	if metrics.accepted != 0 {
		t.Errorf("expected EventsAccepted=0 on repo failure, got %d", metrics.accepted)
	}
}
