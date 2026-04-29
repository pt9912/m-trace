package persistence

import (
	"context"
	"sync"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// InMemorySessionRepository hält die bekannten Sessions in einer Map
// session_id → StreamSession. Pro session_id wird die StartedAt vom
// ersten gesehenen Event gesetzt (ServerReceivedAt); LastEventAt und
// EventCount werden bei jedem Folge-Event aktualisiert. Lifecycle-
// Übergänge (Stalled/Ended) übernimmt §5.1 Sub-Item 8 — bis dahin
// bleiben Sessions im State Active.
//
// Safe für nebenläufige Aufrufe.
type InMemorySessionRepository struct {
	mu       sync.Mutex
	sessions map[string]domain.StreamSession
}

// NewInMemorySessionRepository konstruiert ein leeres Repository.
func NewInMemorySessionRepository() *InMemorySessionRepository {
	return &InMemorySessionRepository{
		sessions: make(map[string]domain.StreamSession),
	}
}

// UpsertFromEvents legt für unbekannte session_id eine neue
// StreamSession an und aktualisiert für bekannte LastEventAt und
// EventCount. Reihenfolge folgt der Slice-Reihenfolge.
func (r *InMemorySessionRepository) UpsertFromEvents(_ context.Context, events []domain.PlaybackEvent) error {
	if len(events) == 0 {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, e := range events {
		s, ok := r.sessions[e.SessionID]
		if !ok {
			r.sessions[e.SessionID] = domain.StreamSession{
				ID:          e.SessionID,
				ProjectID:   e.ProjectID,
				State:       domain.SessionStateActive,
				StartedAt:   e.ServerReceivedAt,
				LastEventAt: e.ServerReceivedAt,
				EventCount:  1,
			}
			continue
		}
		s.LastEventAt = e.ServerReceivedAt
		s.EventCount++
		r.sessions[e.SessionID] = s
	}
	return nil
}

// Snapshot gibt eine Kopie aller bekannten Sessions zurück. Reihenfolge
// nicht garantiert. Für Tests gedacht — Read-Pfade für die
// Sessions-Endpoints folgen in plan-0.1.0 §5.1 Sub-Item 4.
func (r *InMemorySessionRepository) Snapshot() []domain.StreamSession {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]domain.StreamSession, 0, len(r.sessions))
	for _, s := range r.sessions {
		out = append(out, s)
	}
	return out
}

var _ driven.SessionRepository = (*InMemorySessionRepository)(nil)
