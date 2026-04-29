package driven

import (
	"context"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// SessionRepository hält den aggregierten Sessions-Zustand (plan-0.1.0
// §5.1). Der Use Case ruft UpsertFromEvents nach jedem akzeptierten
// Batch auf; Read-Pfade (List/Get) folgen mit den Sessions-Endpoints
// in plan-0.1.0 §5.1 Sub-Item 4.
//
// Implementierungen müssen für nebenläufige Aufrufe sicher sein.
type SessionRepository interface {
	// UpsertFromEvents legt für jede unbekannte session_id eine neue
	// StreamSession (State=Active) an und aktualisiert für bekannte
	// session_id LastEventAt und EventCount.
	UpsertFromEvents(ctx context.Context, events []domain.PlaybackEvent) error
}
