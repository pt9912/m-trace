// Package application — Sessions-Read-Use-Cases.
//
// SessionsService implementiert driving.SessionsInbound und liefert
// die zwei Read-Operationen ListSessions / GetSession aus
//  Sub-Item 4. Wire-Cursor-Codec und HTTP-Mapping
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
// : Default 100 Sessions, hartes Maximum 1000. Der
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
// Cursor-Versionierung und Legacy-Detection liegen im
// HTTP-Adapter (ADR-0004); der Application-Layer arbeitet nur
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

	//  / R-7: Bulk-Read der
	// `network_signal_absent[]`-Blöcke pro Page in einer einzigen
	// IN-Clause-Query — ersetzt den N+1-Pfad aus D3.
	// Reihenfolge bleibt parallel zu page.Sessions; Default für eine
	// Session ohne Boundaries ist ein leerer Slice (Map-Miss).
	sessionIDs := make([]string, len(page.Sessions))
	for i, sess := range page.Sessions {
		sessionIDs[i] = sess.ID
	}
	boundaryMap, err := s.sessions.ListBoundariesForSessions(ctx, in.ProjectID, sessionIDs)
	if err != nil {
		return driving.ListSessionsResult{}, err
	}
	boundaries := make([][]domain.SessionBoundary, len(page.Sessions))
	for i, sess := range page.Sessions {
		boundaries[i] = boundaryMap[sess.ID]
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
			Watermark:        in.EventsAfter.Watermark,
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

	//  D3: persistierten `network_signal_absent[]`-Block
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
			Watermark:        page.NextAfter.Watermark,
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
