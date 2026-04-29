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
	sessions   driven.SessionRepository
	events     driven.EventRepository
	processID  domain.ProcessInstanceID
}

// NewSessionsService konstruiert den Service mit den Driven Ports und
// der ProcessInstanceID, gegen die eingehende Cursor validiert werden.
func NewSessionsService(
	sessions driven.SessionRepository,
	events driven.EventRepository,
	processID domain.ProcessInstanceID,
) *SessionsService {
	return &SessionsService{
		sessions:  sessions,
		events:    events,
		processID: processID,
	}
}

// ListSessions liefert eine geblätterte Liste der Sessions. Mismatcht
// der Cursor die aktuelle ProcessInstanceID, gibt die Methode
// ErrCursorInvalid zurück (Storage-Restart-Pfad, plan-0.1.0.md §5.1).
func (s *SessionsService) ListSessions(ctx context.Context, in driving.ListSessionsInput) (driving.ListSessionsResult, error) {
	limit := clampLimit(in.Limit, DefaultSessionListLimit, MaxSessionListLimit)

	var after *driven.SessionCursorPosition
	if in.After != nil {
		if in.After.ProcessInstanceID != s.processID {
			return driving.ListSessionsResult{}, domain.ErrCursorInvalid
		}
		after = &driven.SessionCursorPosition{
			StartedAt: in.After.StartedAt,
			SessionID: in.After.SessionID,
		}
	}

	page, err := s.sessions.List(ctx, driven.SessionListQuery{
		Limit: limit,
		After: after,
	})
	if err != nil {
		return driving.ListSessionsResult{}, err
	}

	out := driving.ListSessionsResult{Sessions: page.Sessions}
	if page.NextAfter != nil {
		out.NextCursor = &driving.ListSessionsCursor{
			ProcessInstanceID: s.processID,
			StartedAt:         page.NextAfter.StartedAt,
			SessionID:         page.NextAfter.SessionID,
		}
	}
	return out, nil
}

// GetSession liefert Header + geblätterte Event-Liste einer Session.
// ErrSessionNotFound wenn die ID unbekannt; ErrCursorInvalid wenn der
// Event-Cursor von einer fremden ProcessInstanceID kommt.
func (s *SessionsService) GetSession(ctx context.Context, in driving.GetSessionInput) (driving.GetSessionResult, error) {
	session, err := s.sessions.Get(ctx, in.SessionID)
	if err != nil {
		return driving.GetSessionResult{}, err
	}

	limit := clampLimit(in.EventsLimit, DefaultSessionEventsLimit, MaxSessionEventsLimit)

	var after *driven.EventCursorPosition
	if in.EventsAfter != nil {
		if in.EventsAfter.ProcessInstanceID != s.processID {
			return driving.GetSessionResult{}, domain.ErrCursorInvalid
		}
		after = &driven.EventCursorPosition{
			ServerReceivedAt: in.EventsAfter.ServerReceivedAt,
			SequenceNumber:   in.EventsAfter.SequenceNumber,
			IngestSequence:   in.EventsAfter.IngestSequence,
		}
	}

	page, err := s.events.ListBySession(ctx, driven.EventListQuery{
		SessionID: in.SessionID,
		Limit:     limit,
		After:     after,
	})
	if err != nil {
		return driving.GetSessionResult{}, err
	}

	out := driving.GetSessionResult{Session: session, Events: page.Events}
	if page.NextAfter != nil {
		out.NextCursor = &driving.SessionEventsCursor{
			ProcessInstanceID: s.processID,
			ServerReceivedAt:  page.NextAfter.ServerReceivedAt,
			SequenceNumber:    page.NextAfter.SequenceNumber,
			IngestSequence:    page.NextAfter.IngestSequence,
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
