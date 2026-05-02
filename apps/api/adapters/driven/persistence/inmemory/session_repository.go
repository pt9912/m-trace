package inmemory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/persistence"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// SessionRepository hält die bekannten Sessions in einer Map
// session_id → StreamSession. Pro session_id wird die StartedAt vom
// ersten gesehenen Event gesetzt (ServerReceivedAt); LastEventAt und
// EventCount werden bei jedem Folge-Event aktualisiert. Lifecycle-
// Übergänge (Stalled/Ended) übernimmt §5.1 Sub-Item 8 — bis dahin
// bleiben Sessions im State Active.
//
// Safe für nebenläufige Aufrufe.
type SessionRepository struct {
	mu       sync.Mutex
	sessions map[string]domain.StreamSession
}

// NewSessionRepository konstruiert ein leeres Repository.
func NewSessionRepository() *SessionRepository {
	return &SessionRepository{
		sessions: make(map[string]domain.StreamSession),
	}
}

// UpsertFromEvents legt für unbekannte session_id eine neue
// StreamSession an und aktualisiert für bekannte LastEventAt und
// EventCount. Reihenfolge folgt der Slice-Reihenfolge. Ein Event mit
// event_name=session_ended schaltet die Session sofort auf Ended
// (plan-0.1.0.md §5.1 Sub-Item 8).
func (r *SessionRepository) UpsertFromEvents(_ context.Context, events []domain.PlaybackEvent) error {
	if len(events) == 0 {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, e := range events {
		s, ok := r.sessions[e.SessionID]
		if !ok {
			s = domain.StreamSession{
				ID:          e.SessionID,
				ProjectID:   e.ProjectID,
				State:       domain.SessionStateActive,
				StartedAt:   e.ServerReceivedAt,
				LastEventAt: e.ServerReceivedAt,
				EventCount:  1,
			}
		} else {
			s.LastEventAt = e.ServerReceivedAt
			s.EventCount++
		}
		if e.EventName == persistence.SessionEndedEventName && s.State != domain.SessionStateEnded {
			s.State = domain.SessionStateEnded
			endedAt := e.ServerReceivedAt
			s.EndedAt = &endedAt
		}
		r.sessions[e.SessionID] = s
	}
	return nil
}

// Sweep wertet zeitbasierte Lifecycle-Übergänge aus (plan-0.1.0.md
// §5.1 Sub-Item 8). Idempotent: bereits Ended-Sessions werden nicht
// erneut angefasst.
func (r *SessionRepository) Sweep(_ context.Context, now time.Time, stalledAfter, endedAfter time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for id, s := range r.sessions {
		if s.State == domain.SessionStateEnded {
			continue
		}
		idle := now.Sub(s.LastEventAt)
		if s.State == domain.SessionStateActive && idle > stalledAfter {
			s.State = domain.SessionStateStalled
		}
		if s.State == domain.SessionStateStalled && idle > endedAfter {
			s.State = domain.SessionStateEnded
			endedAt := now
			s.EndedAt = &endedAt
		}
		r.sessions[id] = s
	}
	return nil
}

// Snapshot gibt eine Kopie aller bekannten Sessions zurück. Reihenfolge
// nicht garantiert; für Tests gedacht.
func (r *SessionRepository) Snapshot() []domain.StreamSession {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]domain.StreamSession, 0, len(r.sessions))
	for _, s := range r.sessions {
		out = append(out, s)
	}
	return out
}

// List gibt Sessions in stabiler Sortierung (started_at desc,
// session_id asc) zurück. After=nil → erste Seite. Wenn nach dem
// Limit weitere Sessions vorhanden sind, ist NextAfter gesetzt.
func (r *SessionRepository) List(_ context.Context, q driven.SessionListQuery) (driven.SessionPage, error) {
	r.mu.Lock()
	all := make([]domain.StreamSession, 0, len(r.sessions))
	for _, s := range r.sessions {
		all = append(all, s)
	}
	r.mu.Unlock()

	sort.Slice(all, func(i, j int) bool {
		if !all[i].StartedAt.Equal(all[j].StartedAt) {
			return all[i].StartedAt.After(all[j].StartedAt)
		}
		return all[i].ID < all[j].ID
	})

	if q.After != nil {
		idx := sort.Search(len(all), func(i int) bool {
			return sessionPageCursorPasses(all[i], *q.After)
		})
		all = all[idx:]
	}

	limit := q.Limit
	if limit <= 0 {
		return driven.SessionPage{Sessions: []domain.StreamSession{}}, nil
	}

	page := driven.SessionPage{}
	if len(all) > limit {
		page.Sessions = append(page.Sessions, all[:limit]...)
		last := page.Sessions[limit-1]
		page.NextAfter = &driven.SessionCursorPosition{
			StartedAt: last.StartedAt,
			SessionID: last.ID,
		}
		return page, nil
	}
	page.Sessions = append(page.Sessions, all...)
	return page, nil
}

// Get liefert eine einzelne Session per ID. ErrSessionNotFound wenn
// keine Session existiert (plan-0.1.0.md §5.1).
func (r *SessionRepository) Get(_ context.Context, id string) (domain.StreamSession, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.sessions[id]
	if !ok {
		return domain.StreamSession{}, domain.ErrSessionNotFound
	}
	return s, nil
}

// sessionPageCursorPasses gibt true zurück, sobald ein Session-Eintrag
// strikt hinter dem After-Cursor in der Sort-Order (started_at desc,
// session_id asc) liegt. Verwendet als Predicate für sort.Search auf
// einer bereits sortierten Slice.
func sessionPageCursorPasses(s domain.StreamSession, after driven.SessionCursorPosition) bool {
	if !s.StartedAt.Equal(after.StartedAt) {
		// desc order: passes wenn strikt früher gestartet.
		return s.StartedAt.Before(after.StartedAt)
	}
	return s.ID > after.SessionID
}

var _ driven.SessionRepository = (*SessionRepository)(nil)
