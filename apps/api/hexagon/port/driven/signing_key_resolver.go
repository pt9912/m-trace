package driven

import (
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// SigningKeyResolver liefert den Signatur-Key-Ring (`0.12.0`,
// RAK-72). Implementierungen leben im Adapter-Layer und kapseln das
// Schlüsselmaterial — Domain- und Application-Layer rufen den
// Resolver ausschließlich über diese Abstraktion auf.
//
// `ActiveSigningKey` liefert den Key, den `Sign` für neue Tokens
// nutzt; `AllVerifyKeys` liefert alle geladenen Keys (aktiv plus
// retired), gegen die `Verify` prüfen darf. Restart-Stabilität
// (RAK-72): der Resolver muss nach Reinitialisierung dieselben Keys
// liefern, damit ein vor dem Restart ausgestellter Token weiterhin
// validiert.
type SigningKeyResolver interface {
	ActiveSigningKey() (domain.SessionSigningKey, error)
	AllVerifyKeys() ([]domain.SessionSigningKey, error)
}
