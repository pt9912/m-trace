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

// sessionKey ist das projekt-skopierte Composite-Key-Tupel der
// Session-Map ab plan-0.4.0 §4.2. Dieselbe session_id in zwei
// Projekten landet als zwei separate Einträge mit unterschiedlichen
// Korrelations-IDs.
type sessionKey struct {
	ProjectID string
	SessionID string
}

// SessionRepository hält die bekannten Sessions in einer Map
// (project_id, session_id) → StreamSession. Pro Composite-Key wird
// die StartedAt vom ersten gesehenen Event gesetzt
// (ServerReceivedAt); LastEventAt und EventCount werden bei jedem
// Folge-Event aktualisiert. Lifecycle-Übergänge (Stalled/Ended)
// übernimmt §5.1 Sub-Item 8.
//
// Safe für nebenläufige Aufrufe.
type SessionRepository struct {
	mu       sync.Mutex
	sessions map[sessionKey]domain.StreamSession
}

// NewSessionRepository konstruiert ein leeres Repository.
func NewSessionRepository() *SessionRepository {
	return &SessionRepository{
		sessions: make(map[sessionKey]domain.StreamSession),
	}
}

// UpsertFromEvents legt für unbekannte (project_id, session_id) eine
// neue StreamSession an und aktualisiert für bekannte LastEventAt und
// EventCount. Reihenfolge folgt der Slice-Reihenfolge. Ein Event mit
// event_name=session_ended schaltet die Session sofort auf Ended
// (plan-0.1.0.md §5.1 Sub-Item 8).
//
// Rückgabe (R-6-Fix, plan-0.4.0 §4.2 C2): map[sessionID]canonicalCID
// — die nach Upsert in der Map sichtbare CorrelationID. Im InMemory-
// Backend ist das immer der Wert aus dem ersten Insert (alle weiteren
// Events derselben Session lesen ihn aus der Map). Der Use-Case nutzt
// die Map, um Events vor dem EventRepository.Append-Aufruf mit dem
// DB-finalen Wert zu enrichen.
func (r *SessionRepository) UpsertFromEvents(_ context.Context, events []domain.PlaybackEvent) (map[string]string, error) {
	if len(events) == 0 {
		return map[string]string{}, nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	canonical := make(map[string]string, len(events))
	for _, e := range events {
		k := sessionKey{ProjectID: e.ProjectID, SessionID: e.SessionID}
		s, ok := r.sessions[k]
		if !ok {
			s = domain.StreamSession{
				ID:            e.SessionID,
				ProjectID:     e.ProjectID,
				State:         domain.SessionStateActive,
				StartedAt:     e.ServerReceivedAt,
				LastEventAt:   e.ServerReceivedAt,
				EventCount:    1,
				CorrelationID: e.CorrelationID,
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
		r.sessions[k] = s
		canonical[e.SessionID] = s.CorrelationID
	}
	return canonical, nil
}

// Sweep wertet zeitbasierte Lifecycle-Übergänge aus (plan-0.1.0.md
// §5.1 Sub-Item 8). Idempotent: bereits Ended-Sessions werden nicht
// erneut angefasst. Sweep ist global — kein Project-Filter, weil der
// Lifecycle-Sweeper kein Project-Fan-out macht.
func (r *SessionRepository) Sweep(_ context.Context, now time.Time, stalledAfter, endedAfter time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for k, s := range r.sessions {
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
		r.sessions[k] = s
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

// CountByState zählt Sessions im gegebenen Lifecycle-State über alle
// Projekte hinweg (Prometheus-Gauge ist project-agnostisch, siehe
// telemetry-model §3 Cardinality-Regel).
func (r *SessionRepository) CountByState(_ context.Context, state domain.SessionState) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var n int64
	for _, s := range r.sessions {
		if s.State == state {
			n++
		}
	}
	return n, nil
}

// List gibt Sessions in stabiler Sortierung (started_at desc,
// session_id asc) zurück, gefiltert nach q.ProjectID. After=nil →
// erste Seite. Wenn nach dem Limit weitere Sessions vorhanden sind,
// ist NextAfter gesetzt.
func (r *SessionRepository) List(_ context.Context, q driven.SessionListQuery) (driven.SessionPage, error) {
	r.mu.Lock()
	all := make([]domain.StreamSession, 0, len(r.sessions))
	for k, s := range r.sessions {
		if k.ProjectID != q.ProjectID {
			continue
		}
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

// Get liefert eine einzelne Session per (projectID, sessionID).
// ErrSessionNotFound wenn die Session in diesem Project nicht
// existiert; ein Treffer in einem anderen Project gilt als nicht
// gefunden.
func (r *SessionRepository) Get(_ context.Context, projectID, sessionID string) (domain.StreamSession, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.sessions[sessionKey{ProjectID: projectID, SessionID: sessionID}]
	if !ok {
		return domain.StreamSession{}, domain.ErrSessionNotFound
	}
	return s, nil
}

// GetByCorrelationID liefert die Session, deren CorrelationID im
// gegebenen Project gesetzt ist. Legacy-Sessions ohne CorrelationID
// (= Leerwert) zählen nicht als Treffer.
func (r *SessionRepository) GetByCorrelationID(_ context.Context, projectID, correlationID string) (domain.StreamSession, error) {
	if correlationID == "" {
		return domain.StreamSession{}, domain.ErrSessionNotFound
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for k, s := range r.sessions {
		if k.ProjectID != projectID {
			continue
		}
		if s.CorrelationID == correlationID {
			return s, nil
		}
	}
	return domain.StreamSession{}, domain.ErrSessionNotFound
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
