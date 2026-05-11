package driven

import (
	"context"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// AuthSecretBackend abstrahiert die Quelle des Signing-Key-Materials
// (`0.12.5`/RAK-79, R-20). Mit diesem Port wird die Bootstrap-Logik
// vom konkreten Konfigurations-Pfad entkoppelt:
//
//   - `EnvSecretBackend` liest weiterhin aus den ENV-Variablen
//     `MTRACE_AUTH_SIGNING_KEYS`/`_ACTIVE_KID` (mit Single-Key-
//     Backwards-Compat zu `0.12.0`).
//   - `VaultSecretBackend` ist das erste externe Backend-Skelett —
//     holt die gleiche Felderstruktur aus dem Vault KV-v2-Pfad.
//   - KMS-Adapter bleibt additive Folge-Option nach `0.12.5`.
//
// `LoadSigningKeys` ist eine Boot-Time-Operation; ein periodischer
// Refresh ist im `0.12.5`-Scope nicht eingebaut (Operator-Restart
// für Schlüsselwechsel). Bei nicht erreichbarem externen Backend
// schlägt der Adapter mit einem konkreten Fehler fehl — der
// Application-Boot failt damit fail-closed, was sicherer ist als
// ein stiller Fallback auf veraltetes Material.
type AuthSecretBackend interface {
	LoadSigningKeys(ctx context.Context) (keys []domain.SessionSigningKey, activeKID domain.SigningKeyID, err error)
}
