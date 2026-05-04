// Package application — Sessions-Read-Use-Cases.
//
// SessionsService implementiert driving.SessionsInbound und liefert
// die zwei Read-Operationen ListSessions / GetSession aus
// plan-0.1.0.md §5.1 Sub-Item 4. Wire-Cursor-Codec und HTTP-Mapping
// liegen im HTTP-Adapter; die typisierten Cursor in driving.* sind
// die Schnittstelle.
package application

import (
	"context"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driving"
)

// DefaultSessionListLimit / MaxSessionListLimit sind die Limits aus
// plan-0.1.0.md §5.1: Default 100 Sessions, hartes Maximum 1000. Der
// Use Case clampt eingehende Limit-Werte gegen [1, MaxSessionListLimit];
// fehlt oder ist <=0, wird das Default angewandt.
const (
	DefaultSessionListLimit = 100
	MaxSessionListLimit     = 1000

	DefaultSessionEventsLimit = 100
	MaxSessionEventsLimit     = 1000
)

// SessionsService bündelt die Read-Use-Cases für Stream-Sessions.
type SessionsService struct {
	sessions driven.SessionRepository
	events   driven.EventRepository
}

// NewSessionsService konstruiert den Service mit den Driven Ports.
// Cursor-Versionierung und Legacy-Detection liegen ab 0.4.0 im
// HTTP-Adapter (ADR-0004 §7); der Application-Layer arbeitet nur
// noch mit durable Sortier-Werten.
func NewSessionsService(
	sessions driven.SessionRepository,
	events driven.EventRepository,
) *SessionsService {
	return &SessionsService{
		sessions: sessions,
		events:   events,
	}
}

// ListSessions liefert eine geblätterte Liste der Sessions.
func (s *SessionsService) ListSessions(ctx context.Context, in driving.ListSessionsInput) (driving.ListSessionsResult, error) {
	limit := clampLimit(in.Limit, DefaultSessionListLimit, MaxSessionListLimit)

	var after *driven.SessionCursorPosition
	if in.After != nil {
		after = &driven.SessionCursorPosition{
			StartedAt: in.After.StartedAt,
			SessionID: in.After.SessionID,
		}
	}

	page, err := s.sessions.List(ctx, driven.SessionListQuery{
		ProjectID: in.ProjectID,
		Limit:     limit,
		After:     after,
	})
	if err != nil {
		return driving.ListSessionsResult{}, err
	}

	// plan-0.4.0 §4.4 D3: pro Session den persistierten
	// `network_signal_absent[]`-Block laden (spec §3.7.1). Reihenfolge
	// ist parallel zu page.Sessions; Default für eine Session ohne
	// Boundaries ist ein leerer Slice. N+1 ist akzeptiert für 0.4.0
	// (Hard-Max 1000 Sessions pro Page); eine Bulk-Read-Methode kann
	// später ohne Vertragsbruch nachgereicht werden.
	boundaries := make([][]domain.SessionBoundary, len(page.Sessions))
	for i, sess := range page.Sessions {
		bs, err := s.sessions.ListBoundariesForSession(ctx, in.ProjectID, sess.ID)
		if err != nil {
			return driving.ListSessionsResult{}, err
		}
		boundaries[i] = bs
	}

	out := driving.ListSessionsResult{
		Sessions:   page.Sessions,
		Boundaries: boundaries,
	}
	if page.NextAfter != nil {
		out.NextCursor = &driving.ListSessionsCursor{
			StartedAt: page.NextAfter.StartedAt,
			SessionID: page.NextAfter.SessionID,
		}
	}
	return out, nil
}

// GetSession liefert Header + geblätterte Event-Liste einer Session.
// ErrSessionNotFound wenn die (ProjectID, SessionID) unbekannt; ein
// Treffer in einem anderen Project gilt als nicht gefunden.
func (s *SessionsService) GetSession(ctx context.Context, in driving.GetSessionInput) (driving.GetSessionResult, error) {
	session, err := s.sessions.Get(ctx, in.ProjectID, in.SessionID)
	if err != nil {
		return driving.GetSessionResult{}, err
	}

	limit := clampLimit(in.EventsLimit, DefaultSessionEventsLimit, MaxSessionEventsLimit)

	var after *driven.EventCursorPosition
	if in.EventsAfter != nil {
		after = &driven.EventCursorPosition{
			ServerReceivedAt: in.EventsAfter.ServerReceivedAt,
			SequenceNumber:   in.EventsAfter.SequenceNumber,
			IngestSequence:   in.EventsAfter.IngestSequence,
		}
	}

	page, err := s.events.ListBySession(ctx, driven.EventListQuery{
		ProjectID: in.ProjectID,
		SessionID: in.SessionID,
		Limit:     limit,
		After:     after,
	})
	if err != nil {
		return driving.GetSessionResult{}, err
	}

	// plan-0.4.0 §4.4 D3: persistierten `network_signal_absent[]`-Block
	// laden (spec §3.7.1). Default leerer Slice — der HTTP-Adapter
	// rendert das als JSON-Array `[]`, kein `null`.
	boundaries, err := s.sessions.ListBoundariesForSession(ctx, in.ProjectID, in.SessionID)
	if err != nil {
		return driving.GetSessionResult{}, err
	}

	out := driving.GetSessionResult{Session: session, Events: page.Events, Boundaries: boundaries}
	if page.NextAfter != nil {
		out.NextCursor = &driving.SessionEventsCursor{
			ServerReceivedAt: page.NextAfter.ServerReceivedAt,
			SequenceNumber:   page.NextAfter.SequenceNumber,
			IngestSequence:   page.NextAfter.IngestSequence,
		}
	}
	return out, nil
}

// clampLimit gibt das Default zurück wenn limit<=0, das Max wenn
// limit>max, sonst limit selbst.
func clampLimit(limit, defaultLimit, maxLimit int) int {
	if limit <= 0 {
		return defaultLimit
	}
	if limit > maxLimit {
		return maxLimit
	}
	return limit
}

// Compile-time check.
var _ driving.SessionsInbound = (*SessionsService)(nil)
