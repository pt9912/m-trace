package driven

import (
	"context"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// ProjectPolicyResolver liefert die Project-Policy zu einer
// `project_id` (RAK-74). Die Implementierung lebt im Adapter-Layer
// und wird vom Application-Service über diesen Port konsumiert.
//
// Liefert `domain.ErrAuthPolicyDenied`, wenn das Project keine
// Policy konfiguriert hat (kein implizites „alles erlaubt") — der
// HTTP-Adapter mappt das auf `403 auth_policy_denied`.
type ProjectPolicyResolver interface {
	ResolvePolicy(ctx context.Context, projectID string) (domain.ProjectPolicy, error)
}
