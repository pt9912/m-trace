// Package auth holds spike-only authentication adapters. The spike
// uses a static map (Spec §6.4); production deployments will replace
// this with a real token validator.
package auth

import (
	"context"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// ProjectConfig beschreibt die statische Konfiguration eines Projects:
// Token plus optional konkrete Allowed-Origins für die CORS-Variante-B-
// Origin-Validierung (plan-0.1.0.md §5.1). Wildcards (`*`) werden
// abgelehnt — der Pflicht-Pfad fordert konkrete Origins.
type ProjectConfig struct {
	Token          string
	AllowedOrigins []string
}

// StaticProjectResolver resolves a presented X-MTrace-Token to a
// project using a hardcoded reverse map (token -> project_id) und
// hält die globale Union der Allowed-Origins über alle Projects für
// den CORS-Preflight-Pfad bereit.
type StaticProjectResolver struct {
	byToken      map[string]domain.Project
	originsUnion map[string]struct{}
}

// NewStaticProjectResolver builds a resolver from a project_id ->
// ProjectConfig map. Tokens müssen über alle Projects unique sein —
// kollidierende Tokens würden im Reverse-Mapping einander
// überschreiben.
func NewStaticProjectResolver(byProjectID map[string]ProjectConfig) *StaticProjectResolver {
	byToken := make(map[string]domain.Project, len(byProjectID))
	originsUnion := make(map[string]struct{})
	for projectID, cfg := range byProjectID {
		project := domain.Project{
			ID:             projectID,
			Token:          domain.ProjectToken(cfg.Token),
			AllowedOrigins: append([]string(nil), cfg.AllowedOrigins...),
		}
		byToken[cfg.Token] = project
		for _, o := range cfg.AllowedOrigins {
			originsUnion[o] = struct{}{}
		}
	}
	return &StaticProjectResolver{
		byToken:      byToken,
		originsUnion: originsUnion,
	}
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

// IsOriginInGlobalUnion gibt true zurück, wenn `origin` exakt in der
// globalen Union aller Allowed-Origins über alle bekannten Projects
// steht. Wird vom CORS-Preflight-Handler verwendet (plan-0.1.0.md
// §5.1, CORS Variante B). Leerer Origin → false (Preflight ohne
// Origin-Header existiert nicht im Browser-Pfad).
func (r *StaticProjectResolver) IsOriginInGlobalUnion(origin string) bool {
	if origin == "" {
		return false
	}
	_, ok := r.originsUnion[origin]
	return ok
}

var _ driven.ProjectResolver = (*StaticProjectResolver)(nil)
