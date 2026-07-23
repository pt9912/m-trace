package driving

import (
	"context"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// EventStreamInbound ist der Driving-Port für den SSE-Live-Update-Read:
// der HTTP-Adapter abonniert einen project-skopierten Frame-Strom für die
// Lebensdauer einer Connection. Implementiert vom in-process EventBroker
// (application). Streaming-Schnittstelle → Rückgabe eines Channels.
type EventStreamInbound interface {
	Subscribe(ctx context.Context, projectID string) <-chan domain.EventAppendedFrame
}
