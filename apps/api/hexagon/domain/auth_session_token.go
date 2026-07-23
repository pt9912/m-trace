package domain

import (
	"crypto/subtle"
	"strings"
	"time"
)

// Auth-Session-Token-Domäne (RAK-72).
//
// Sicherheitsprofil und Wire-Vertrag in
// `spec/backend-api-contract.md` §3.9 und
// `docs/plan/planning/in-progress/` §0.5/§0.6.
//
// Domain-Typen (Claims, Signing-Key-Identifier,
// zeitbasierte Validierung, Audience-/Origin-/Session-Bindung) ohne
// HTTP-, JSON-, SQLite- oder Crypto-Library-Abhängigkeit. Die
// tatsächliche Signatur-Erzeugung und -Prüfung sitzt in im
// Application-Service-/Adapter-Pfad.

// SessionTokenPrefix ist der projektweite Wire-Prefix für signierte
// Session Tokens. Tokens haben das Format
// `mtr_st_<base64url(header).base64url(claims).base64url(signature)>`;
// der Prefix ist Teil des Klartexts und damit auch in der Signatur.
const SessionTokenPrefix = "mtr_st_"

// MaxSessionTokenTTLSeconds ist die harte globale Obergrenze für
// `ttl_seconds` aus `POST /api/auth/session-tokens` (Wire-Vertrag
// RAK-72). Project Policies dürfen niedriger begrenzen, aber
// nicht höher; das wird im Issuance-Pfad und in `ValidateTTLSeconds`
// erzwungen.
const MaxSessionTokenTTLSeconds = 900

// DefaultSessionTokenIssuer ist der Default-`iss`-Claim für Session
// Tokens. Der Wert ist stabil über Releases; Adapter-Konfiguration
// kann ihn überschreiben, aber Tests pinnen den Wert für die
// Default-Konfiguration.
const DefaultSessionTokenIssuer = "m-trace"

// SessionTokenAudience ist die Allowlist-Form der `aud`-Claims. Im
// Pflichtpfad ist `playback-events` die einzige Muss-
// Audience (RAK-72). Folge-Releases erweitern additiv; bestehende
// Werte bleiben stabil.
type SessionTokenAudience string

// SessionTokenAudience-Werte aus dem Muss-Scope.
const (
	SessionTokenAudiencePlaybackEvents SessionTokenAudience = "playback-events"
)

// IsKnown prüft, ob ein Audience-Wert in der Allowlist steht.
// Tests und Adapter mappen unbekannte Werte auf
// `403 auth_session_scope_denied`.
func (a SessionTokenAudience) IsKnown() bool {
	switch a {
	case SessionTokenAudiencePlaybackEvents:
		return true
	default:
		return false
	}
}

// SigningKeyID identifiziert einen Signaturschlüssel im server-
// seitigen Key-Ring. Der Wert landet als `kid` im Token-Header und
// erlaubt parallele Signing-Keys plus restart-stabile Verifikation
// alter Tokens nach einem Key-Rollover.
type SigningKeyID string

// SigningKeyAlgorithm benennt das Signaturverfahren. `0.12.0`
// implementiert ausschließlich HMAC-SHA-256 (`HS256`); ECDSA/EdDSA
// bleiben Folge-Scope. Die Konstante ist als String-Enum modelliert,
// damit Folge-Releases additiv erweitern können.
type SigningKeyAlgorithm string

// SigningKeyAlgorithm-Werte aus dem Muss-Scope.
const (
	SigningKeyAlgorithmHS256 SigningKeyAlgorithm = "HS256"
)

// IsKnown prüft, ob das Signaturverfahren in der Allowlist
// steht. Tokens mit unbekanntem `alg` liefern stabil
// `401 auth_token_invalid`.
func (a SigningKeyAlgorithm) IsKnown() bool {
	switch a {
	case SigningKeyAlgorithmHS256:
		return true
	default:
		return false
	}
}

// SessionSigningKey beschreibt einen Eintrag im server-seitigen
// Key-Ring (`spec/backend-api-contract.md` §3.9, RAK-72). `Secret`
// trägt das Schlüsselmaterial — dieser Wert darf weder geloggt noch
// in Persistenz, Fixtures oder Lifecycle-Events erscheinen. Der
// Domain-Layer benutzt ihn ausschließlich für die Signatur-Primitive
// in; das Modell pinnt nur die Felder, die für
// Verify-Lookup, Rotation und Restart-Stabilität nötig sind.
//
// `NotBefore` und `RetiresAt` umrahmen das Signing-Fenster: ein Key
// darf erst ab `NotBefore` neue Tokens ausstellen und wird nach
// `RetiresAt` aus dem Active-Set entfernt; alte Verify-Keys bleiben
// dennoch geladen, bis alle damit signierten Tokens abgelaufen sind.
type SessionSigningKey struct {
	KID        SigningKeyID
	Algorithm  SigningKeyAlgorithm
	Secret     []byte
	NotBefore  time.Time
	RetiresAt  time.Time
}

// SessionTokenClaims spiegelt den signierten Claim-Set eines Session
// Tokens. Pflichtfelder sind `iss`, `sub` (`project_id`), `aud`,
// `iat`, `nbf`, `exp` und `jti`; `session_id` und `origin` sind
// optional und binden den Token zusätzlich (RAK-72,
// `spec/backend-api-contract.md` §3.9).
//
// `JTI` ist der signierte Identifier; in Wire-Responses, Logs und
// Fixtures wird derselbe Wert als `token_id` exposed. Tests pinnen,
// dass beide Werte identisch sind und dass `jti` nur innerhalb des
// signierten Claim-Sets verwendet wird.
//
// Optionale Felder sind als String-Pointer modelliert, damit das
// Domain-Modell eindeutig zwischen „nicht gesetzt" und „explizit
// leerer String" unterscheidet — der HTTP-Adapter validiert leere
// Strings als `400 invalid_request`, bevor sie in den Claim-Set
// kommen.
type SessionTokenClaims struct {
	Iss       string
	Sub       string
	Aud       SessionTokenAudience
	Iat       time.Time
	Nbf       time.Time
	Exp       time.Time
	JTI       string
	SessionID *string
	Origin    *string
}

// TokenID ist der öffentliche Wire-/Log-Name des `jti`-Claims. Beide
// Werte sind per Konstruktion identisch; diese Funktion existiert
// nur, damit Aufrufer im Code dokumentieren können, ob sie den
// signierten oder den Wire-/Log-Wert lesen.
func (c SessionTokenClaims) TokenID() string {
	return c.JTI
}

// SessionTokenIssuanceInput bündelt die Eingaben aus dem Issuance-
// Endpoint (`POST /api/auth/session-tokens`). Der Application-Layer
// baut daraus einen `SessionTokenClaims`; das Domain-Modell pinnt nur
// die Felder, die für die Konsistenzprüfung mit der Project-Policy
// nötig sind.
type SessionTokenIssuanceInput struct {
	ProjectID    string
	Audience     SessionTokenAudience
	TTLSeconds   int
	SessionID    *string
	Origin       *string
}

// BuildSessionTokenClaims erzeugt einen Claim-Set aus einem
// validierten Issuance-Input und der zum Issuance-Zeitpunkt
// gültigen Clock. `tokenID` ist der vom Aufrufer vergebene Wert
// (typischerweise ein ULID); `issuer` kommt aus der Adapter-
// Konfiguration und defaultet auf `DefaultSessionTokenIssuer`.
//
// Die Funktion validiert den Input nicht — Audience, Project-ID-
// Konsistenz und TTL müssen vorher über `ValidateAudience`,
// `ValidateProjectIDConsistency` und `ValidateTTLSeconds` geprüft
// sein. So bleibt der Claim-Builder deterministisch und testbar.
func BuildSessionTokenClaims(input SessionTokenIssuanceInput, tokenID string, issuer string, now time.Time) SessionTokenClaims {
	if issuer == "" {
		issuer = DefaultSessionTokenIssuer
	}
	exp := now.Add(time.Duration(input.TTLSeconds) * time.Second)
	return SessionTokenClaims{
		Iss:       issuer,
		Sub:       input.ProjectID,
		Aud:       input.Audience,
		Iat:       now,
		Nbf:       now,
		Exp:       exp,
		JTI:       tokenID,
		SessionID: cloneStringPointer(input.SessionID),
		Origin:    cloneStringPointer(input.Origin),
	}
}

// ValidateClaimsTime prüft `nbf`/`exp` gegen die übergebene Clock.
// Der Aufrufer (Application-Service) wendet diese Prüfung **nach**
// der Signatur- und `kid`-Validierung an, damit die Reihenfolge der
// Fehlerpräzedenz aus §3.9 stabil bleibt: malformed/invalid → revoked
// → expired → not-yet-valid → project-mismatch → scope-denied.
//
// Tokens, deren `nbf` oder `exp` exakt der aktuellen Clock
// entsprechen, werden konservativ behandelt: `now == exp` zählt als
// abgelaufen, `now == nbf` zählt als gültig. So vermeidet der Token-
// Pfad eine Edge-Case-Lücke an der Sekundengrenze.
func ValidateClaimsTime(claims SessionTokenClaims, now time.Time) error {
	if !now.Before(claims.Exp) {
		return ErrAuthTokenExpired
	}
	if now.Before(claims.Nbf) {
		return ErrAuthTokenNotYetValid
	}
	return nil
}

// ValidateClaimsAudience prüft die `aud`-Bindung und gleichzeitig,
// dass die Audience in der globalen Allowlist steht. Eine
// nicht in der Allowlist enthaltene Audience liefert immer
// `ErrAuthSessionScopeDenied`, auch wenn das Token die richtige
// Audience trägt — so kann ein Audience-Allowlist-Eintrag im
// Compatibility-Fenster zurückgezogen werden, ohne dass alte Tokens
// plötzlich gültig wären.
func ValidateClaimsAudience(claims SessionTokenClaims, expected SessionTokenAudience) error {
	if !expected.IsKnown() || !claims.Aud.IsKnown() {
		return ErrAuthSessionScopeDenied
	}
	if claims.Aud != expected {
		return ErrAuthSessionScopeDenied
	}
	return nil
}

// ValidateClaimsProject prüft die `sub`-Bindung gegen das aus dem
// Token-Pfad aufgelöste Project. Ein Mismatch liefert
// `ErrAuthProjectMismatch`; der HTTP-Adapter mappt das auf `401`.
//
// Die Funktion ist auch dann relevant, wenn die Auth-Pipeline einen
// Session Token plus einen Legacy-`X-MTrace-Token` erlaubt — beide
// müssen denselben `project_id` binden. Der Aufrufer ruft die
// Funktion einmal mit dem Session-Claim und einmal mit dem
// resolvierten Project-Token-Generation-Wert auf.
func ValidateClaimsProject(claims SessionTokenClaims, expectedProjectID string) error {
	if claims.Sub == "" || claims.Sub != expectedProjectID {
		return ErrAuthProjectMismatch
	}
	return nil
}

// ValidateClaimsSession prüft, dass der Token-Pfad genau die Session
// bindet, die im konsumierenden Request steht — sofern der Token
// überhaupt eine `session_id` trägt. Wenn der Claim-Wert leer ist,
// lässt die Funktion alle Sessions zu (der Token wurde ohne
// Session-Bindung ausgestellt).
//
// Ein im Claim gesetzter aber im Request fehlender Session-Wert
// liefert `ErrAuthSessionScopeDenied`; ebenso ein Mismatch.
func ValidateClaimsSession(claims SessionTokenClaims, requestSessionID string) error {
	if claims.SessionID == nil {
		return nil
	}
	if requestSessionID == "" || requestSessionID != *claims.SessionID {
		return ErrAuthSessionScopeDenied
	}
	return nil
}

// ValidateClaimsOrigin prüft, dass der Token-Pfad genau den Origin
// bindet, der im konsumierenden Request kommt — sofern der Token
// überhaupt einen `origin` trägt. Ein im Claim gesetzter aber im
// Request fehlender Origin-Wert liefert `ErrAuthSessionScopeDenied`;
// ebenso ein Mismatch. Origin-Vergleich ist case-sensitive — RFC 6454
// definiert Origins als case-sensitive Tupel.
func ValidateClaimsOrigin(claims SessionTokenClaims, requestOrigin string) error {
	if claims.Origin == nil {
		return nil
	}
	if requestOrigin == "" || requestOrigin != *claims.Origin {
		return ErrAuthSessionScopeDenied
	}
	return nil
}

// LookupSigningKey findet einen Verify-Key im Key-Ring anhand seiner
// `kid`. Unbekannte `kid` liefern `ErrAuthTokenInvalid` — der HTTP-
// Adapter mappt das auf `401 auth_token_invalid`, ohne den `kid`-Wert
// im Klartext zu echo'en.
//
// Die Liste enthält **alle** geladenen Keys (aktive Signing-Keys plus
// alte Verify-Keys, die noch nicht aus allen aktiven Tokens
// herausgealtert sind). Aktive vs. retired wird hier nicht
// unterschieden; entscheidet beim Sign-Pfad, welcher Key
// für neue Tokens benutzt wird.
func LookupSigningKey(keys []SessionSigningKey, kid SigningKeyID) (SessionSigningKey, error) {
	if kid == "" {
		return SessionSigningKey{}, ErrAuthTokenInvalid
	}
	for _, k := range keys {
		if k.KID == kid {
			return k, nil
		}
	}
	return SessionSigningKey{}, ErrAuthTokenInvalid
}

// ConstantTimeEqualSignature vergleicht zwei Signature-Bytes-Slices
// in konstanter Zeit. Wrapper über `crypto/subtle.ConstantTimeCompare`
// mit klarer Domain-Semantik: `false` bei jeder Form von Mismatch
// (auch unterschiedliche Längen). nutzt diese Funktion in
// allen Verify-Pfaden, damit ein Side-Channel über Vergleichszeit
// strukturell ausgeschlossen ist.
func ConstantTimeEqualSignature(a, b []byte) bool {
	return subtle.ConstantTimeCompare(a, b) == 1
}

// HasSessionTokenPrefix gibt true zurück, wenn `token` mit dem
// `mtr_st_`-Wire-Prefix beginnt. Adapter nutzen das, um zwischen
// Project-Token-Pfad (`X-MTrace-Token`) und Session-Token-Pfad
// (`Authorization: Bearer mtr_st_*` / `X-MTrace-Session-Token`)
// strukturell zu unterscheiden, bevor die teure Signatur-Verifikation
// läuft.
func HasSessionTokenPrefix(token string) bool {
	return strings.HasPrefix(token, SessionTokenPrefix)
}

// cloneStringPointer kopiert einen optionalen String-Pointer, damit
// der Claim-Set keine Aliase auf den Issuance-Input behält. Wichtig
// für die Clean-Hexagon-Trennung: ein nachträgliches Mutieren des
// Eingangs darf den ausgestellten Token nicht ändern.
func cloneStringPointer(s *string) *string {
	if s == nil {
		return nil
	}
	v := *s
	return &v
}
