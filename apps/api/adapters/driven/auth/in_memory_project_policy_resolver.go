package auth

import (
	"context"
	"fmt"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// InMemoryProjectPolicyResolver implementiert
// `driven.ProjectPolicyResolver` gegen eine vom Operator
// konfigurierte Map (`0.12.0`, RAK-74). Restart-Stabilität wird
// durch die Operator-Konfiguration garantiert.
//
// Falls für ein Project keine Policy konfiguriert ist, fällt der
// Resolver auf `domain.ProjectPolicyFromBaseProject` zurück, sofern
// ein `BaseProject` bekannt ist. Andernfalls liefert er
// `domain.ErrAuthPolicyDenied`, was der HTTP-Adapter auf `403`
// mappt.
type InMemoryProjectPolicyResolver struct {
	policies     map[string]domain.ProjectPolicy
	baseProjects map[string]domain.Project
}

// NewInMemoryProjectPolicyResolver baut den Resolver aus expliziten
// Policies und optionalen Base-Projects (für den Fallback).
//
// Validiert die Policies beim Aufbau — `ProjectMaxTTLSeconds`
// > `domain.MaxSessionTokenTTLSeconds` (900) wird **nicht** stillschweigend
// geclampt, sondern als Operator-Konfigurationsfehler signalisiert.
// `EffectiveMaxTTLSeconds` clampt zur Defense-in-Depth weiterhin am
// Request-Pfad.
func NewInMemoryProjectPolicyResolver(
	policies map[string]domain.ProjectPolicy,
	baseProjects map[string]domain.Project,
) (*InMemoryProjectPolicyResolver, error) {
	for projectID, p := range policies {
		if p.ProjectMaxTTLSeconds > domain.MaxSessionTokenTTLSeconds {
			return nil, fmt.Errorf(
				"auth: project %q ProjectMaxTTLSeconds=%d exceeds global ceiling %d",
				projectID, p.ProjectMaxTTLSeconds, domain.MaxSessionTokenTTLSeconds)
		}
	}
	out := &InMemoryProjectPolicyResolver{
		policies:     make(map[string]domain.ProjectPolicy, len(policies)),
		baseProjects: make(map[string]domain.Project, len(baseProjects)),
	}
	for k, v := range policies {
		out.policies[k] = v
	}
	for k, v := range baseProjects {
		out.baseProjects[k] = v
	}
	return out, nil
}

// Compile-time check.
var _ driven.ProjectPolicyResolver = (*InMemoryProjectPolicyResolver)(nil)

// ResolvePolicy liefert die konfigurierte Policy oder einen sicheren
// Default aus dem Base-Project. Unbekannte Projects ohne Base-Project
// liefern `domain.ErrAuthPolicyDenied`.
func (r *InMemoryProjectPolicyResolver) ResolvePolicy(_ context.Context, projectID string) (domain.ProjectPolicy, error) {
	if r == nil {
		return domain.ProjectPolicy{}, domain.ErrAuthPolicyDenied
	}
	if p, ok := r.policies[projectID]; ok {
		return p, nil
	}
	if base, ok := r.baseProjects[projectID]; ok {
		return domain.ProjectPolicyFromBaseProject(base), nil
	}
	return domain.ProjectPolicy{}, domain.ErrAuthPolicyDenied
}
