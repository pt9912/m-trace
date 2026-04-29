package application

import (
	"context"
	"log/slog"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// Default-Schwellwerte für den Sessions-Lifecycle (plan-0.1.0.md §5.1
// Sub-Item 8). Im 0.1.0-Spike als Konstante; ENV-/Konfig-Tunable folgt,
// sobald ein Lab-Szenario es braucht.
const (
	DefaultSessionStalledAfter = 60 * time.Second
	DefaultSessionEndedAfter   = 5 * time.Minute
	DefaultSessionSweepEvery   = 10 * time.Second
)

// SessionsSweeper ruft SessionRepository.Sweep periodisch auf und
// realisiert damit die Active→Stalled→Ended-Übergänge. Eine
// Implementierung ohne externe Scheduler-Abhängigkeit reicht für
// 0.1.0; eine SQLite-Migration kann den Sweep bei Bedarf in einen
// On-Read-Pfad verschieben (Plan Roadmap §4).
type SessionsSweeper struct {
	repo          driven.SessionRepository
	now           func() time.Time
	stalledAfter  time.Duration
	endedAfter    time.Duration
	sweepInterval time.Duration
	logger        *slog.Logger
}

// NewSessionsSweeper konstruiert einen Sweeper mit den
// Default-Schwellwerten. now=nil → time.Now.
func NewSessionsSweeper(
	repo driven.SessionRepository,
	now func() time.Time,
	logger *slog.Logger,
) *SessionsSweeper {
	if now == nil {
		now = time.Now
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &SessionsSweeper{
		repo:          repo,
		now:           now,
		stalledAfter:  DefaultSessionStalledAfter,
		endedAfter:    DefaultSessionEndedAfter,
		sweepInterval: DefaultSessionSweepEvery,
		logger:        logger,
	}
}

// SweepOnce führt einen einzelnen Sweep-Pass aus. Für deterministische
// Tests verwendbar; in main.go wird Run aus einer Goroutine gerufen.
func (s *SessionsSweeper) SweepOnce(ctx context.Context) error {
	return s.repo.Sweep(ctx, s.now(), s.stalledAfter, s.endedAfter)
}

// Run startet einen Ticker-Loop und ruft SweepOnce, bis ctx geschlossen
// wird. Errors werden geloggt, der Loop bricht nicht ab — die Sicht
// auf Sessions soll bei vorübergehenden Adapter-Fehlern nicht
// einfrieren.
func (s *SessionsSweeper) Run(ctx context.Context) {
	ticker := time.NewTicker(s.sweepInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.SweepOnce(ctx); err != nil {
				s.logger.Error("sessions sweep failed", "error", err)
			}
		}
	}
}
