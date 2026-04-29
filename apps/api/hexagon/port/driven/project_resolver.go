package driven

import (
	"context"

	"github.com/example/m-trace/apps/api/hexagon/domain"
)

// ProjectResolver looks up a project by its presented X-MTrace-Token.
// Returns domain.ErrUnauthorized if no project owns the given token.
// The spike implementation reads from a static map (Spec §6.4).
type ProjectResolver interface {
	ResolveByToken(ctx context.Context, token string) (domain.Project, error)
}
