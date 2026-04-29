package driving

import (
	"context"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// SessionsInbound ist der Read-Pfad zu Stream-Sessions (plan-0.1.0.md
// §5.1, Sub-Item 4): zwei Operationen für Liste und Detail. Der HTTP-
// Adapter encoded und decoded den opaken Wire-Cursor; hier fließen
// nur typisierte Cursor.
type SessionsInbound interface {
	ListSessions(ctx context.Context, in ListSessionsInput) (ListSessionsResult, error)
	GetSession(ctx context.Context, in GetSessionInput) (GetSessionResult, error)
}

// ListSessionsInput trägt Limit und optionalen After-Cursor. Limit wird
// im Use Case auf das Default- bzw. Hard-Maximum-Fenster geclampt
// (plan-0.1.0.md §5.1: Default 100, Hard-Max 1000).
type ListSessionsInput struct {
	Limit int
	After *ListSessionsCursor
}

// ListSessionsCursor kapselt die Sortier-Position der Sessions-Liste:
// (started_at desc, session_id asc). ProcessInstanceID dient zur
// Cursor-Invalidierung nach Storage-Restart.
type ListSessionsCursor struct {
	ProcessInstanceID domain.ProcessInstanceID
	StartedAt         time.Time
	SessionID         string
}

// ListSessionsResult bündelt die geblätterte Sessions-Page. NextCursor
// ist nil, wenn die letzte Seite erreicht ist.
type ListSessionsResult struct {
	Sessions   []domain.StreamSession
	NextCursor *ListSessionsCursor
}

// GetSessionInput identifiziert eine Session und steuert die
// Event-Pagination innerhalb der Detail-Antwort.
type GetSessionInput struct {
	SessionID   string
	EventsLimit int
	EventsAfter *SessionEventsCursor
}

// SessionEventsCursor kapselt die Sortier-Position der Event-Liste
// einer Session: (server_received_at asc, sequence_number asc,
// ingest_sequence asc). ingest_sequence ist serverseitig gesetzt und
// damit der finale Tie-Breaker (plan-0.1.0.md §5.1).
type SessionEventsCursor struct {
	ProcessInstanceID domain.ProcessInstanceID
	ServerReceivedAt  time.Time
	SequenceNumber    *int64
	IngestSequence    int64
}

// GetSessionResult bündelt Sessions-Header und die geblätterte
// Event-Page. NextCursor ist nil, wenn die letzte Seite erreicht ist.
type GetSessionResult struct {
	Session    domain.StreamSession
	Events     []domain.PlaybackEvent
	NextCursor *SessionEventsCursor
}
