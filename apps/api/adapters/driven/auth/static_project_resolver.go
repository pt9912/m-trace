// Package auth holds spike-only authentication adapters. The spike
// uses a static map (Spec §6.4); production deployments will replace
// this with a real token validator.
package auth

import (
	"context"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// StaticProjectResolver resolves a presented X-MTrace-Token to a
// project using a hardcoded reverse map (token -> project_id).
type StaticProjectResolver struct {
	byToken map[string]domain.Project
}

// NewStaticProjectResolver builds a resolver from a project_id ->
// token map (the canonical map defined by Spec §6.4 and
// docs/spike/backend-api-contract.md §4).
func NewStaticProjectResolver(byProjectID map[string]string) *StaticProjectResolver {
	byToken := make(map[string]domain.Project, len(byProjectID))
	for projectID, token := range byProjectID {
		byToken[token] = domain.Project{
			ID:    projectID,
			Token: domain.ProjectToken(token),
		}
	}
	return &StaticProjectResolver{byToken: byToken}
}

// ResolveByToken returns the project owning the given token, or
// domain.ErrUnauthorized if no such project exists.
func (r *StaticProjectResolver) ResolveByToken(_ context.Context, token string) (domain.Project, error) {
	project, ok := r.byToken[token]
	if !ok {
		return domain.Project{}, domain.ErrUnauthorized
	}
	return project, nil
}

var _ driven.ProjectResolver = (*StaticProjectResolver)(nil)
