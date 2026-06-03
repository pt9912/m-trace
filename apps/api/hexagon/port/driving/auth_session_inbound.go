package driving

import (
	"context"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// AuthSessionInbound ist der Driving-Port für den Session-Token-
// Issuance-Pfad (RAK-72). Der HTTP-Adapter dekodiert
// `POST /api/auth/session-tokens`, leitet die Eingabe an den
// Application-Service weiter und mappt das Ergebnis auf das
// JSON-Schema aus `spec/backend-api-contract.md` §3.9.
//
// Sicherheitsprofil:
//  - `IssueSessionTokenResult.Value` ist der Klartext-Session-Token
//  und darf nur in der Issuance-Antwort erscheinen. Jeder andere
//  Pfad (Logs, Metriken, Traces, Persistenz, Fixtures) bekommt
//  ausschließlich `TokenID` oder Fingerprints.
//  - Der Application-Service akzeptiert ausschließlich vorher
//  resolvierte `ResolvedProjectID`-Werte aus dem Project-Token-
//  Pfad; ein Session Token darf keinen weiteren Session Token
//  minten.
type AuthSessionInbound interface {
	IssueSessionToken(ctx context.Context, req IssueSessionTokenRequest) (IssueSessionTokenResult, error)
}

// IssueSessionTokenRequest ist die Driving-Port-Eingabe für
// `POST /api/auth/session-tokens`. `ResolvedProjectID` ist das aus
// dem Project Token aufgelöste Project; `RequestProjectID` ist der
// optionale Wire-Vertrag-Wert (Konsistenzcheck zum Token, §3.9).
//
// `RequestedTTLSeconds == 0` triggert den Default-Pfad
// (`min(project_max_ttl_seconds, 900)`); explizite Werte werden
// gegen die wirksame Project-Grenze ohne stillen Clamp validiert.
type IssueSessionTokenRequest struct {
	ResolvedProjectID   string
	RequestProjectID    string
	Audience            string
	RequestedTTLSeconds int
	SessionID           string
	Origin              string
	IssuanceClientID    string
}

// IssueSessionTokenResult bündelt das Ergebnis der Issuance. Der
// HTTP-Adapter reicht `Value` genau einmal in der `201`-Antwort an
// den Aufrufer durch; `TokenID`, `ProjectID`, `Audience`, `SessionID`
// und `ExpiresAt` sind log-/audit-tauglich.
type IssueSessionTokenResult struct {
	Value     string
	TokenID   string
	ProjectID string
	Audience  domain.SessionTokenAudience
	SessionID string
	ExpiresAt time.Time
}
