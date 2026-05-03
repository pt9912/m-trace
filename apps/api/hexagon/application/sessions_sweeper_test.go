package application_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/application"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// recordingRepo speichert die zuletzt gesehene Sweep-Eingabe und kann
// Fehler simulieren — reicht für SessionsSweeper-Tests, ohne den
// vollen In-Memory-Adapter zu fahren.
type recordingRepo struct {
	mu          sync.Mutex
	sweepCalls  int
	lastNow     time.Time
	lastStalled time.Duration
	lastEnded   time.Duration
	failNext    bool
}

func (r *recordingRepo) UpsertFromEvents(_ context.Context, _ []domain.PlaybackEvent) (map[string]string, error) {
	return map[string]string{}, nil
}
func (r *recordingRepo) List(_ context.Context, _ driven.SessionListQuery) (driven.SessionPage, error) {
	return driven.SessionPage{}, nil
}
func (r *recordingRepo) Get(_ context.Context, _ string, _ string) (domain.StreamSession, error) {
	return domain.StreamSession{}, domain.ErrSessionNotFound
}
func (r *recordingRepo) GetByCorrelationID(_ context.Context, _ string, _ string) (domain.StreamSession, error) {
	return domain.StreamSession{}, domain.ErrSessionNotFound
}
func (r *recordingRepo) CountByState(_ context.Context, _ domain.SessionState) (int64, error) {
	return 0, nil
}
func (r *recordingRepo) Sweep(_ context.Context, now time.Time, stalled, ended time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sweepCalls++
	r.lastNow = now
	r.lastStalled = stalled
	r.lastEnded = ended
	if r.failNext {
		r.failNext = false
		return errors.New("simulated repo error")
	}
	return nil
}

// TestSessionsSweeper_SweepOnce_PassesDefaults verifiziert, dass
// SweepOnce die Default-Schwellwerte und das now() weitergibt.
func TestSessionsSweeper_SweepOnce_PassesDefaults(t *testing.T) {
	t.Parallel()
	repo := &recordingRepo{}
	now := time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC)
	sw := application.NewSessionsSweeper(repo, func() time.Time { return now }, slog.New(slog.NewJSONHandler(io.Discard, nil)))

	if err := sw.SweepOnce(context.Background()); err != nil {
		t.Fatalf("SweepOnce: %v", err)
	}

	repo.mu.Lock()
	defer repo.mu.Unlock()
	if repo.sweepCalls != 1 {
		t.Errorf("sweepCalls=%d want 1", repo.sweepCalls)
	}
	if !repo.lastNow.Equal(now) {
		t.Errorf("lastNow=%v want %v", repo.lastNow, now)
	}
	if repo.lastStalled != application.DefaultSessionStalledAfter {
		t.Errorf("stalledAfter=%v want %v", repo.lastStalled, application.DefaultSessionStalledAfter)
	}
	if repo.lastEnded != application.DefaultSessionEndedAfter {
		t.Errorf("endedAfter=%v want %v", repo.lastEnded, application.DefaultSessionEndedAfter)
	}
}

// TestSessionsSweeper_SweepOnce_PropagatesError prüft, dass ein
// Adapter-Fehler aus SweepOnce zurück an den Caller fließt — wichtig,
// damit Run() ihn loggen kann, ohne den Loop einzufrieren.
func TestSessionsSweeper_SweepOnce_PropagatesError(t *testing.T) {
	t.Parallel()
	repo := &recordingRepo{failNext: true}
	sw := application.NewSessionsSweeper(repo, time.Now, slog.New(slog.NewJSONHandler(io.Discard, nil)))
	if err := sw.SweepOnce(context.Background()); err == nil {
		t.Errorf("expected error to propagate, got nil")
	}
}

// TestSessionsSweeper_Run_StopsOnContextCancel verifiziert den
// graceful-shutdown-Pfad: Run kehrt zurück, sobald ctx geschlossen
// wird. Wir cancellen den Context unmittelbar nach Run-Start; der
// Ticker hat dann noch nicht gefeuert, und die Test-Goroutine räumt
// in unter einer Sekunde ab.
func TestSessionsSweeper_Run_StopsOnContextCancel(t *testing.T) {
	t.Parallel()
	repo := &recordingRepo{}
	sw := application.NewSessionsSweeper(repo, time.Now, slog.New(slog.NewJSONHandler(io.Discard, nil)))

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		sw.Run(ctx)
		close(done)
	}()
	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not exit after ctx cancellation")
	}
}
