package driven

import (
	"context"
	"errors"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// ProjectTokenRepository persistiert rotierbare Project-Token-
// Generationen (`0.12.0`, RAK-73). Implementierungen liegen im Adapter-
// Layer (InMemory + SQLite); Application- und Domain-Layer rufen den
// Repository-Pfad über diesen Port.
//
// Sicherheitsprofil:
//   - Klartext-Tokens werden weder Eingang noch Ausgang dieses Ports
//     sehen — Persistenz speichert ausschließlich `KeyHash`,
//     `Fingerprint` und Lifecycle-Metadaten aus
//     `domain.ProjectTokenGeneration`.
//   - `Create` erlaubt mehrere Generationen pro Project; gleichzeitig
//     aktive plus eine Vorgänger-Generation in Grace.
//   - Cross-Project-Lookups gibt es nicht — `FindByHash` ist
//     repository-weit, weil ein Hash über alle Projects unique sein
//     muss (wir prüfen das via Constraint).
//   - `Revoke` und `SetGraceUntil` sind idempotent und liefern
//     `ErrProjectTokenNotFound`, wenn die Generation nicht existiert
//     oder nicht zum angegebenen Project gehört.
type ProjectTokenRepository interface {
	Create(ctx context.Context, gen domain.ProjectTokenGeneration) error
	ListByProject(ctx context.Context, projectID string) ([]domain.ProjectTokenGeneration, error)
	FindByHash(ctx context.Context, keyHash string) (domain.ProjectTokenGeneration, error)
	SetGraceUntil(ctx context.Context, projectID, tokenID string, graceUntil time.Time) error
	Revoke(ctx context.Context, projectID, tokenID string, revokedAt time.Time) error
}

// ErrProjectTokenNotFound signalisiert, dass eine Token-Generation
// nicht im Repository steht oder nicht zum angegebenen Project
// gehört. Adapter mappen das auf `domain.ErrAuthTokenInvalid`, damit
// kein Cross-Project-Existenz-Hinweis leakt.
var ErrProjectTokenNotFound = errors.New("project token generation not found")
