package driven

import (
	"context"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// SessionRepository hält den aggregierten Sessions-Zustand (plan-0.1.0
// §5.1). Der Use Case ruft UpsertFromEvents nach jedem akzeptierten
// Batch auf; List und Get bedienen die Read-Endpoints (plan-0.1.0 §5.1
// Sub-Item 4); Sweep wird vom Lifecycle-Sweeper aufgerufen
// (plan-0.1.0 §5.1 Sub-Item 8).
//
// Implementierungen müssen für nebenläufige Aufrufe sicher sein.
type SessionRepository interface {
	// UpsertFromEvents legt für jede unbekannte session_id eine neue
	// StreamSession (State=Active) an und aktualisiert für bekannte
	// session_id LastEventAt und EventCount. Trifft ein Event mit
	// event_name=session_ended ein, wird die Session direkt auf
	// State=Ended gesetzt und EndedAt=event.ServerReceivedAt.
	UpsertFromEvents(ctx context.Context, events []domain.PlaybackEvent) error
	// List gibt Sessions in stabiler Sortierung (started_at desc,
	// session_id asc) zurück. Der Adapter ist für die Sortierung
	// verantwortlich; der Use Case clampt nur Limit und prüft Cursor-
	// Validität.
	List(ctx context.Context, q SessionListQuery) (SessionPage, error)
	// Get liefert eine einzelne Session über ihre ID. ErrSessionNotFound
	// wenn keine Session existiert.
	Get(ctx context.Context, id string) (domain.StreamSession, error)
	// Sweep wertet die zeitbasierten Lifecycle-Übergänge aus:
	//   Active  + (now - LastEventAt > stalledAfter) → Stalled
	//   Stalled + (now - LastEventAt > endedAfter)   → Ended (EndedAt=now)
	// Bereits beendete Sessions werden nicht erneut angefasst. Idempotent.
	Sweep(ctx context.Context, now time.Time, stalledAfter, endedAfter time.Duration) error
}

// SessionListQuery ist die Eingabe für SessionRepository.List.
type SessionListQuery struct {
	Limit int
	After *SessionCursorPosition
}

// SessionCursorPosition ist die Repository-Sicht auf den Cursor —
// die Sortier-Felder ohne Wire-Format.
type SessionCursorPosition struct {
	StartedAt time.Time
	SessionID string
}

// SessionPage bündelt eine Page Sessions plus optional die nächste
// Cursor-Position.
type SessionPage struct {
	Sessions  []domain.StreamSession
	NextAfter *SessionCursorPosition
}
