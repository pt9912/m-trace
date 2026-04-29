// Package driven holds the outbound (driven) ports — the interfaces
// the application layer needs from the outside world (persistence,
// metrics, rate limiting, project lookup). Implementations live in
// adapters/driven/*.
package driven

import (
	"context"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// EventRepository persists accepted events. The spike uses an
// in-memory implementation; production will likely move to an event
// store. Implementations must be safe for concurrent use.
type EventRepository interface {
	Append(ctx context.Context, events []domain.PlaybackEvent) error
}
