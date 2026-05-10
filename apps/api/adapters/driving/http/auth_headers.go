package http

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// Auth-Header-Parser für tokenpflichtige Konsum-Endpoints (`0.12.0`,
// RAK-72/RAK-75). Spiegelt die §3.9-Regeln aus
// `spec/backend-api-contract.md`:
//
//  1. `Authorization: Bearer mtr_st_*` ist der bevorzugte Session-
//     Token-Pfad. Andere `Authorization`-Werte sind kein m-trace
//     Auth-Versuch und werden ignoriert, solange ein gültiger
//     m-trace Header (`X-MTrace-Token` oder `X-MTrace-Session-Token`)
//     vorhanden ist.
//  2. `X-MTrace-Session-Token: mtr_st_*` ist der alternative
//     Session-Token-Pfad.
//  3. `X-MTrace-Token` ist der Legacy-/Project-Token-Pfad.
//  4. Werden mehrere Tokens präsentiert, müssen alle dasselbe Project
//     binden. Widersprüche → `auth_project_mismatch`. Ein zusätzlich
//     präsentiertes ungültiges Token → `auth_token_invalid`. Kein
//     stiller Fallback von einem ungültigen höher priorisierten Token
//     auf ein gültiges niedriger priorisiertes Token.
//  5. Wenn `Authorization: Bearer mtr_st_*` und
//     `X-MTrace-Session-Token` beide gesetzt sind und unterschiedliche
//     Tokens enthalten → `auth_token_invalid`.
//
// Der Parser kennt keine HTTP-Antwort und kein Logging — er liefert
// ein `AuthDecision` mit dem aufgelösten Project (oder `nil` für den
// Legacy-Pfad) plus dem rohen `X-MTrace-Token`-Wert. HTTP-Handler
// mappen das auf `4xx`-Codes über `errors.Is`.

// AuthDecision ist das Ergebnis der Auth-Header-Auswertung.
//
//   - LegacyToken ist der `X-MTrace-Token`-Wert, falls präsentiert.
//   - ResolvedProject ist das aus einem Session Token aufgelöste
//     Project (Bearer- oder X-MTrace-Session-Token-Pfad). `nil` für
//     den reinen Legacy-Pfad.
//
// Wenn beide Werte gesetzt sind, hat der Parser bereits geprüft, dass
// sie zum selben Project gehören.
type AuthDecision struct {
	LegacyToken     string
	ResolvedProject *domain.Project
}

// AuthHeaderParser bündelt die Driven-Ports, die der Parser zum
// Auflösen braucht. `Verifier` darf `nil` sein, wenn die Konfiguration
// den Session-Token-Pfad nicht aktiviert hat — der Parser akzeptiert
// dann nur Legacy-`X-MTrace-Token`-Header, und ein präsentierter
// `Bearer mtr_st_*` oder `X-MTrace-Session-Token` liefert
// `domain.ErrAuthTokenInvalid`.
type AuthHeaderParser struct {
	Resolver  driven.ProjectResolver
	Verifier  driven.SessionTokenSigner
	Projects  ProjectByIDLookup
	Now       func() time.Time
	Audience  domain.SessionTokenAudience
}

// ProjectByIDLookup ist ein optionaler Backref vom `sub`-Claim auf
// das `domain.Project` (Allowed-Origins, ID). `StaticProjectResolver`
// implementiert das implementiert über `ResolveByID`.
type ProjectByIDLookup interface {
	ResolveByID(projectID string) (domain.Project, error)
}

// Parse wertet die drei Auth-Header aus und liefert eine
// `AuthDecision`. Die Reihenfolge entspricht §3.9: Bearer →
// X-MTrace-Session-Token → X-MTrace-Token. Multi-Token-Konflikte
// werden gegen `domain.ErrAuthProjectMismatch` bzw.
// `domain.ErrAuthTokenInvalid` validiert.
func (p AuthHeaderParser) Parse(ctx context.Context, headerGetter func(string) string, requestOrigin string) (AuthDecision, error) {
	raw := readRawAuthHeaders(headerGetter)
	sessionToken, err := pickSessionToken(raw)
	if err != nil {
		return AuthDecision{}, err
	}
	if sessionToken == "" && raw.legacy == "" {
		return AuthDecision{}, domain.ErrAuthTokenMissing
	}
	resolved, err := p.resolveSessionPath(sessionToken, requestOrigin)
	if err != nil {
		return AuthDecision{}, err
	}
	return p.applyLegacyHeader(ctx, raw.legacy, resolved)
}

// rawAuthHeaders bündelt die für die Auswertung benötigten Header-
// Werte (bereits getrimmt). `bearer` ist nur befüllt, wenn der
// Authorization-Header das `Bearer mtr_st_*`-Schema trägt — fremde
// Bearer- oder Basic-Header werden hier auf "" gemappt.
type rawAuthHeaders struct {
	authorization string
	bearer        string
	sessionHeader string
	legacy        string
}

func readRawAuthHeaders(headerGetter func(string) string) rawAuthHeaders {
	authHeader := strings.TrimSpace(headerGetter("Authorization"))
	return rawAuthHeaders{
		authorization: authHeader,
		bearer:        extractMTraceBearer(authHeader),
		sessionHeader: strings.TrimSpace(headerGetter("X-MTrace-Session-Token")),
		legacy:        strings.TrimSpace(headerGetter("X-MTrace-Token")),
	}
}

// pickSessionToken wählt den effektiven Session-Token-Wert. Wenn
// Bearer und X-MTrace-Session-Token beide gesetzt sind und sich
// unterscheiden, liefert die Funktion `ErrAuthTokenInvalid` — auch
// wenn einer der beiden Tokens für sich gültig wäre (§3.9 Punkt 5).
func pickSessionToken(raw rawAuthHeaders) (string, error) {
	if raw.bearer != "" && raw.sessionHeader != "" && raw.bearer != raw.sessionHeader {
		return "", domain.ErrAuthTokenInvalid
	}
	if raw.bearer != "" {
		return raw.bearer, nil
	}
	return raw.sessionHeader, nil
}

// resolveSessionPath verifiziert das Session Token (sofern
// präsentiert) und liefert das aufgelöste Project. Ohne Session
// Token bleibt das Ergebnis `nil` — der Aufrufer fällt auf den
// Legacy-Pfad zurück.
func (p AuthHeaderParser) resolveSessionPath(sessionToken, requestOrigin string) (*domain.Project, error) {
	if sessionToken == "" {
		return nil, nil
	}
	if p.Verifier == nil || p.Projects == nil {
		return nil, domain.ErrAuthTokenInvalid
	}
	project, err := p.resolveSessionToken(sessionToken, requestOrigin)
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// applyLegacyHeader prüft den optionalen Legacy-`X-MTrace-Token`-
// Header. Wenn präsentiert, muss er — neben einem etwaigen Session-
// Token — dasselbe Project binden. Ein ungültiger Legacy-Token blockt
// den Fallback (§3.9 Punkt 4).
//
// Lifecycle-Fehler aus dem `RotatingProjectResolver` (`mtr_pt_*`-
// Generationen: revoked/expired/not_yet_valid) werden 1:1 propagiert,
// damit der HTTP-Adapter sie auf die distinkten §3.9-Codes mappen
// kann. Andere Resolver-Fehler (Static-Resolver-Miss,
// Unauthorized-Sentinel) werden konservativ auf
// `ErrAuthTokenInvalid` gemappt — kein Hinweis auf Existenz.
func (p AuthHeaderParser) applyLegacyHeader(ctx context.Context, legacy string, resolved *domain.Project) (AuthDecision, error) {
	if legacy == "" {
		return AuthDecision{ResolvedProject: resolved}, nil
	}
	legacyProject, err := p.Resolver.ResolveByToken(ctx, legacy)
	if err != nil {
		return AuthDecision{}, mapLegacyResolverError(err)
	}
	if resolved != nil && resolved.ID != legacyProject.ID {
		return AuthDecision{}, domain.ErrAuthProjectMismatch
	}
	return AuthDecision{
		LegacyToken:     legacy,
		ResolvedProject: resolved,
	}, nil
}

// mapLegacyResolverError leitet Auth-Lifecycle-Fehler aus dem
// `RotatingProjectResolver` weiter und mapt alle anderen Resolver-
// Fehler (insb. `domain.ErrUnauthorized` aus dem Static-Resolver) auf
// `domain.ErrAuthTokenInvalid`. Damit bleibt die §3.9-Präzedenz
// (revoked > expired > not_yet_valid > invalid) erhalten — aus dem
// Pre-`0.12.0`-Static-Pfad gibt es nur invalid, was korrekt bleibt.
func mapLegacyResolverError(err error) error {
	switch {
	case errors.Is(err, domain.ErrAuthTokenRevoked),
		errors.Is(err, domain.ErrAuthTokenExpired),
		errors.Is(err, domain.ErrAuthTokenNotYetValid),
		errors.Is(err, domain.ErrAuthSessionScopeDenied),
		errors.Is(err, domain.ErrAuthPolicyDenied):
		return err
	default:
		return domain.ErrAuthTokenInvalid
	}
}

// extractMTraceBearer extrahiert ein `Bearer mtr_st_*`-Token aus
// dem Authorization-Header. Das `Bearer`-Schema wird gemäß RFC 7235
// case-insensitive geprüft — `Bearer`, `bearer`, `BEARER` etc.
// gelten alle, weil Reverse-Proxies und Client-Libraries (z. B.
// OkHttp, requests) unterschiedliche Normalisierungen anwenden.
//
// Andere Schemas (OAuth/Basic/„Bearer abc") werden als „kein m-trace
// Auth-Versuch" behandelt → leerer String.
func extractMTraceBearer(authHeader string) string {
	const scheme = "Bearer "
	if len(authHeader) < len(scheme) {
		return ""
	}
	if !strings.EqualFold(authHeader[:len(scheme)], scheme) {
		return ""
	}
	candidate := strings.TrimSpace(authHeader[len(scheme):])
	if !strings.HasPrefix(candidate, domain.SessionTokenPrefix) {
		// Bearer mit fremdem Token (z. B. OAuth-Token-Format) — wir
		// ignorieren ihn als nicht-m-trace Auth.
		return ""
	}
	return candidate
}

// resolveSessionToken validiert Signatur, Zeit, Audience und Origin
// eines präsentierten Session Tokens und liefert das aufgelöste
// Project zurück.
func (p AuthHeaderParser) resolveSessionToken(token, origin string) (domain.Project, error) {
	claims, err := p.Verifier.Verify(token)
	if err != nil {
		return domain.Project{}, err
	}
	now := p.now()
	if err := domain.ValidateClaimsTime(claims, now); err != nil {
		return domain.Project{}, err
	}
	expectedAud := p.Audience
	if expectedAud == "" {
		expectedAud = domain.SessionTokenAudiencePlaybackEvents
	}
	if err := domain.ValidateClaimsAudience(claims, expectedAud); err != nil {
		return domain.Project{}, err
	}
	if err := domain.ValidateClaimsOrigin(claims, origin); err != nil {
		return domain.Project{}, err
	}
	project, err := p.Projects.ResolveByID(claims.Sub)
	if err != nil {
		// Project ist serverseitig nicht (mehr) bekannt — `auth_token_invalid`.
		return domain.Project{}, domain.ErrAuthTokenInvalid
	}
	if err := domain.ValidateClaimsProject(claims, project.ID); err != nil {
		return domain.Project{}, err
	}
	return project, nil
}

// now liefert die aktuelle Clock — Tests injizieren eine
// deterministische Funktion.
func (p AuthHeaderParser) now() time.Time {
	if p.Now == nil {
		return time.Now().UTC()
	}
	return p.Now()
}
