package driven

import (
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// SessionTokenSigner signiert und verifiziert kurzlebige Session
// Tokens (RAK-72). Der Application-Service kennt keinen
// Algorithmus und keine Schlüssel direkt — er ruft den Signer-/
// Verifier-Pfad über diesen Port.
//
// Sicherheitsprofil:
//  - `Sign` darf den Klartext-Token-String ausschließlich an den
//  Aufrufer zurückgeben; weder Logs noch Metriken noch Traces
//  dürfen den Wert sehen.
//  - `Verify` validiert Signatur, `kid`, Header-Format und
//  Claim-Set-Decoding. Zeit-, Audience-, Project- und Origin-
//  Bindung prüft der Application-Service über die Domain-
//  Funktionen aus `auth_session_token.go` — der Verifier liefert
//  nur den decodierten Claim-Set zurück.
type SessionTokenSigner interface {
	Sign(claims domain.SessionTokenClaims) (string, error)
	Verify(token string) (domain.SessionTokenClaims, error)
}
