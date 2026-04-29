package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/application"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driving"
)

// fakeSessionRepo implementiert driven.SessionRepository für Tests.
// Speichert Sessions in einer Map; List sortiert deterministisch nach
// (started_at desc, session_id asc) und nutzt den After-Cursor.
type fakeSessionRepo struct {
	store map[string]domain.StreamSession
}

func newFakeSessionRepo() *fakeSessionRepo {
	return &fakeSessionRepo{store: make(map[string]domain.StreamSession)}
}

func (r *fakeSessionRepo) UpsertFromEvents(_ context.Context, events []domain.PlaybackEvent) error {
	for _, e := range events {
		s, ok := r.store[e.SessionID]
		if !ok {
			r.store[e.SessionID] = domain.StreamSession{
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
		r.store[e.SessionID] = s
	}
	return nil
}

func (r *fakeSessionRepo) List(_ context.Context, q driven.SessionListQuery) (driven.SessionPage, error) {
	all := make([]domain.StreamSession, 0, len(r.store))
	for _, s := range r.store {
		all = append(all, s)
	}
	// In-place insertion sort: simple, stable, no extra deps.
	for i := 1; i < len(all); i++ {
		j := i
		for j > 0 && sessionLess(all[j], all[j-1]) {
			all[j], all[j-1] = all[j-1], all[j]
			j--
		}
	}
	if q.After != nil {
		// drop everything <= cursor
		out := all[:0]
		passed := false
		for _, s := range all {
			if !passed {
				if s.StartedAt.Equal(q.After.StartedAt) && s.ID == q.After.SessionID {
					passed = true
					continue
				}
				if s.StartedAt.Before(q.After.StartedAt) {
					passed = true
				} else if s.StartedAt.Equal(q.After.StartedAt) && s.ID > q.After.SessionID {
					passed = true
				}
			}
			if passed {
				out = append(out, s)
			}
		}
		all = out
	}
	page := driven.SessionPage{}
	if q.Limit > 0 && len(all) > q.Limit {
		page.Sessions = append(page.Sessions, all[:q.Limit]...)
		last := page.Sessions[q.Limit-1]
		page.NextAfter = &driven.SessionCursorPosition{
			StartedAt: last.StartedAt,
			SessionID: last.ID,
		}
		return page, nil
	}
	page.Sessions = append(page.Sessions, all...)
	return page, nil
}

func sessionLess(a, b domain.StreamSession) bool {
	if !a.StartedAt.Equal(b.StartedAt) {
		return a.StartedAt.After(b.StartedAt)
	}
	return a.ID < b.ID
}

func (r *fakeSessionRepo) Get(_ context.Context, id string) (domain.StreamSession, error) {
	s, ok := r.store[id]
	if !ok {
		return domain.StreamSession{}, domain.ErrSessionNotFound
	}
	return s, nil
}

// fakeEventRepo implementiert driven.EventRepository für Tests.
type fakeEventRepo struct {
	events []domain.PlaybackEvent
}

func (r *fakeEventRepo) Append(_ context.Context, events []domain.PlaybackEvent) error {
	r.events = append(r.events, events...)
	return nil
}

func (r *fakeEventRepo) ListBySession(_ context.Context, q driven.EventListQuery) (driven.EventPage, error) {
	matching := make([]domain.PlaybackEvent, 0)
	for _, e := range r.events {
		if e.SessionID == q.SessionID {
			matching = append(matching, e)
		}
	}
	// sort by ingest_sequence asc (test fakes use it as primary).
	for i := 1; i < len(matching); i++ {
		j := i
		for j > 0 && matching[j].IngestSequence < matching[j-1].IngestSequence {
			matching[j], matching[j-1] = matching[j-1], matching[j]
			j--
		}
	}
	if q.After != nil {
		out := matching[:0]
		for _, e := range matching {
			if e.IngestSequence > q.After.IngestSequence {
				out = append(out, e)
			}
		}
		matching = out
	}
	page := driven.EventPage{}
	if q.Limit > 0 && len(matching) > q.Limit {
		page.Events = append(page.Events, matching[:q.Limit]...)
		last := page.Events[q.Limit-1]
		page.NextAfter = &driven.EventCursorPosition{
			ServerReceivedAt: last.ServerReceivedAt,
			SequenceNumber:   last.SequenceNumber,
			IngestSequence:   last.IngestSequence,
		}
		return page, nil
	}
	page.Events = append(page.Events, matching...)
	return page, nil
}

const testProcessID domain.ProcessInstanceID = "test-process-123"

func TestSessionsService_ListSessions_LimitClampedToDefault(t *testing.T) {
	t.Parallel()
	repo := newFakeSessionRepo()
	repo.store["s1"] = domain.StreamSession{ID: "s1", StartedAt: time.Unix(100, 0)}
	repo.store["s2"] = domain.StreamSession{ID: "s2", StartedAt: time.Unix(200, 0)}

	svc := application.NewSessionsService(repo, &fakeEventRepo{}, testProcessID)
	res, err := svc.ListSessions(context.Background(), driving.ListSessionsInput{})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(res.Sessions) != 2 {
		t.Errorf("expected 2 sessions, got %d", len(res.Sessions))
	}
	// started_at desc → s2, s1
	if res.Sessions[0].ID != "s2" || res.Sessions[1].ID != "s1" {
		t.Errorf("sort order wrong: %v", []string{res.Sessions[0].ID, res.Sessions[1].ID})
	}
}

func TestSessionsService_ListSessions_LimitClampedToMax(t *testing.T) {
	t.Parallel()
	repo := newFakeSessionRepo()
	for i := 0; i < 5; i++ {
		id := []string{"a", "b", "c", "d", "e"}[i]
		repo.store[id] = domain.StreamSession{ID: id, StartedAt: time.Unix(int64(100+i), 0)}
	}

	svc := application.NewSessionsService(repo, &fakeEventRepo{}, testProcessID)
	res, err := svc.ListSessions(context.Background(), driving.ListSessionsInput{Limit: 1_000_000})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	// 5 sessions exist, cap is 1000, request was 1_000_000 → all 5 returned.
	if len(res.Sessions) != 5 {
		t.Errorf("expected 5 sessions (cap above store size), got %d", len(res.Sessions))
	}
	// Limit honored against the cap: pass cap+1 worth via small store; cursor advanced when over limit.
	res2, err := svc.ListSessions(context.Background(), driving.ListSessionsInput{Limit: 2})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(res2.Sessions) != 2 || res2.NextCursor == nil {
		t.Errorf("expected 2 sessions + cursor, got %d / cursor=%v", len(res2.Sessions), res2.NextCursor)
	}
}

func TestSessionsService_ListSessions_CursorMismatchInvalidatesPagination(t *testing.T) {
	t.Parallel()
	repo := newFakeSessionRepo()
	repo.store["s1"] = domain.StreamSession{ID: "s1", StartedAt: time.Unix(100, 0)}
	svc := application.NewSessionsService(repo, &fakeEventRepo{}, testProcessID)

	_, err := svc.ListSessions(context.Background(), driving.ListSessionsInput{
		After: &driving.ListSessionsCursor{
			ProcessInstanceID: "stale-process-from-before-restart",
			StartedAt:         time.Unix(200, 0),
			SessionID:         "s99",
		},
	})
	if !errors.Is(err, domain.ErrCursorInvalid) {
		t.Errorf("expected ErrCursorInvalid, got %v", err)
	}
}

func TestSessionsService_GetSession_NotFound(t *testing.T) {
	t.Parallel()
	svc := application.NewSessionsService(newFakeSessionRepo(), &fakeEventRepo{}, testProcessID)
	_, err := svc.GetSession(context.Background(), driving.GetSessionInput{SessionID: "missing"})
	if !errors.Is(err, domain.ErrSessionNotFound) {
		t.Errorf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestSessionsService_GetSession_PaginatesEvents(t *testing.T) {
	t.Parallel()
	repo := newFakeSessionRepo()
	repo.store["s1"] = domain.StreamSession{ID: "s1", StartedAt: time.Unix(100, 0), EventCount: 3}

	events := &fakeEventRepo{}
	t0 := time.Unix(100, 0)
	for i := 1; i <= 3; i++ {
		events.events = append(events.events, domain.PlaybackEvent{
			SessionID:        "s1",
			ServerReceivedAt: t0,
			IngestSequence:   int64(i),
		})
	}

	svc := application.NewSessionsService(repo, events, testProcessID)
	res, err := svc.GetSession(context.Background(), driving.GetSessionInput{
		SessionID:   "s1",
		EventsLimit: 2,
	})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(res.Events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(res.Events))
	}
	if res.NextCursor == nil {
		t.Fatalf("expected NextCursor on partial page")
	}
	if res.NextCursor.IngestSequence != 2 {
		t.Errorf("expected NextCursor.IngestSequence=2, got %d", res.NextCursor.IngestSequence)
	}
	if res.NextCursor.ProcessInstanceID != testProcessID {
		t.Errorf("expected NextCursor.ProcessInstanceID=%q, got %q", testProcessID, res.NextCursor.ProcessInstanceID)
	}
}

func TestSessionsService_GetSession_CursorMismatch(t *testing.T) {
	t.Parallel()
	repo := newFakeSessionRepo()
	repo.store["s1"] = domain.StreamSession{ID: "s1"}
	svc := application.NewSessionsService(repo, &fakeEventRepo{}, testProcessID)

	_, err := svc.GetSession(context.Background(), driving.GetSessionInput{
		SessionID: "s1",
		EventsAfter: &driving.SessionEventsCursor{
			ProcessInstanceID: "stale",
			IngestSequence:    1,
		},
	})
	if !errors.Is(err, domain.ErrCursorInvalid) {
		t.Errorf("expected ErrCursorInvalid, got %v", err)
	}
}
