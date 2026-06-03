package domain

import (
	"strings"
)

// Project-Policy-Domäne (RAK-74).
//
// liefert das Modell und die deterministischen Validierungs-
// /Effective-TTL-Funktionen ohne HTTP-, Storage- oder Rate-Limit-
// Adapterabhängigkeit. Die HTTP-Enforcement-Schicht wird in Tranche 4
// gegen diese Typen verdrahtet.
//
// Wire-Vertrag in `spec/backend-api-contract.md` §3.9 (Auth-Matrix,
// CORS-Preflight, Project-Policies-Block); Out-of-Scope-Klammern in
// `docs/planning/in-progress/` §0.1 (Origin-/IP-Buckets
// optionaler Zusatz).

// HTTPMethod modelliert die im Pflichtpfad zulässigen
// Methoden für tokenpflichtige Konsum-Endpoints. Plan §0.1 begrenzt
// die Allowlist auf `POST`/`OPTIONS`; die Konstanten sind Wire-
// stabil.
type HTTPMethod string

// HTTPMethod-Werte aus dem Muss-Scope.
const (
	HTTPMethodPOST    HTTPMethod = "POST"
	HTTPMethodOPTIONS HTTPMethod = "OPTIONS"
)

// IsKnown prüft, ob eine Methode in der Allowlist steht.
// Andere Methoden liefern `auth_policy_denied`.
func (m HTTPMethod) IsKnown() bool {
	switch m {
	case HTTPMethodPOST, HTTPMethodOPTIONS:
		return true
	default:
		return false
	}
}

// AllowedRequestHeader modelliert die im Pflichtpfad
// zulässigen Request-Header (CORS-Preflight-Allowlist plus den
// `traceparent`-Header für Trace-Korrelation). Plan §0.5/§0.6 und
// 9 pinnen die Liste; weitere Header bleiben Folge-Scope.
type AllowedRequestHeader string

// AllowedRequestHeader-Werte aus dem Muss-Scope. Strings
// folgen RFC-7230 case-insensitive Konvention; Vergleich passiert
// case-insensitive in `IsHeaderAllowed`.
const (
	HeaderContentType        AllowedRequestHeader = "Content-Type"
	HeaderAuthorization      AllowedRequestHeader = "Authorization"
	HeaderXMTraceToken       AllowedRequestHeader = "X-MTrace-Token"
	HeaderXMTraceSessionToken AllowedRequestHeader = "X-MTrace-Session-Token"
	HeaderTraceparent        AllowedRequestHeader = "traceparent"
)

// GlobalPreflightHeaderAllowlist liefert die globale Header-Liste,
// die der CORS-Preflight-Handler für `Access-Control-Allow-Headers`
// ausliefert (RAK-74, §3.9). Reihenfolge ist stabil, damit Contract-
// Fixtures und Smoke-Tests die ausgegebene Header-Liste byte-genau
// pinnen können.
func GlobalPreflightHeaderAllowlist() []AllowedRequestHeader {
	return []AllowedRequestHeader{
		HeaderContentType,
		HeaderAuthorization,
		HeaderXMTraceToken,
		HeaderXMTraceSessionToken,
		HeaderTraceparent,
	}
}

// GlobalPreflightMethodAllowlist liefert die globale Method-Liste
// (`POST, OPTIONS`) für den Preflight-Handler.
func GlobalPreflightMethodAllowlist() []HTTPMethod {
	return []HTTPMethod{HTTPMethodPOST, HTTPMethodOPTIONS}
}

// RateLimitBucket beschreibt die Parameter eines einzelnen Token-
// Bucket-Limits (Capacity = max Burst, RefillPerSecond = Steady-State-
// Rate). Werte 0 deaktivieren das Limit; das ist nur für Tests
// sinnvoll, in Produktion müssen Project- und Issuance-Bucket immer
// gesetzt sein.
type RateLimitBucket struct {
	Capacity        int
	RefillPerSecond float64
}

// IsZero gibt true zurück, wenn der Bucket nicht konfiguriert ist
// (Capacity == 0 und RefillPerSecond == 0). Adapter nutzen das, um
// einen Folge-Scope-Bucket (Origin/IP) zu erkennen, der in der
// aktuellen Konfiguration deaktiviert bleibt.
func (b RateLimitBucket) IsZero() bool {
	return b.Capacity == 0 && b.RefillPerSecond == 0
}

// RateLimitPolicy bündelt die Project-gebundenen Rate-Limit-Buckets
// aus RAK-74. `ProjectBucket` und `IssuanceBucket` sind Muss-Pfad;
// `OriginBucket` und `IPBucket` sind optionaler Zusatz oder Folge-
// Scope und dürfen `IsZero` sein.
//
// `IssuanceBucket` ist exklusiv für `POST /api/auth/session-tokens`;
// Überschreitung liefert `429 auth_issuance_rate_limited`. Andere
// Endpoints nutzen `ProjectBucket` und ihren endpoint-spezifischen
// Rate-Limit-Code (z. B. `429 Too Many Requests` für
// `/api/playback-events`).
type RateLimitPolicy struct {
	ProjectBucket  RateLimitBucket
	IssuanceBucket RateLimitBucket
	OriginBucket   RateLimitBucket
	IPBucket       RateLimitBucket
}

// OriginPolicy bündelt origin-spezifische Overrides innerhalb einer
// Project Policy. Im Muss-Scope ist nur die einfache
// Origin-Allowlist Pflicht; `RestrictedAudiences` ist ein optionales
// Feld, mit dem ein einzelner Origin nur für ein eingeschränktes
// Audience-Set Session Tokens minten darf. Leeres Set bedeutet „alle
// Audiences der Project Policy erlaubt".
type OriginPolicy struct {
	Origin              string
	RestrictedAudiences []SessionTokenAudience
}

// AllowsAudience prüft, ob die Origin-Policy eine bestimmte Audience
// für Session-Token-Issuance erlaubt. Leeres `RestrictedAudiences`-
// Feld ist permissiv (alles erlaubt, was die Project-Policy erlaubt).
func (o OriginPolicy) AllowsAudience(aud SessionTokenAudience) bool {
	if len(o.RestrictedAudiences) == 0 {
		return true
	}
	for _, a := range o.RestrictedAudiences {
		if a == aud {
			return true
		}
	}
	return false
}

// ProjectPolicy ist die normative Tenant-Policy aus RAK-74 plus
// Plan §0.5 (TTL, Audience-Allowlist). Alle Felder zusammen ergeben
// den deterministischen Enforcement-Vertrag — einzelne Felder dürfen
// nicht in Adaptern dupliziert werden, damit Drift zwischen Storage,
// HTTP und Doku ausgeschlossen ist.
//
// `ProjectMaxTTLSeconds` ≤ 0 bedeutet „nimm den globalen Default
// 900"; das wird in `EffectiveMaxTTLSeconds` aufgelöst.
//
// `BrowserIngest` (`0.12.5`/RAK-80) hebt den RAK-74-Scope-Cut für
// `/api/ingest/*` kontrolliert auf, sobald ein Project sie aktiv
// schaltet. Default `IsZero` → RAK-74-Scope-Cut bleibt strikt:
// `/api/ingest/*` ist operator-only, kein Browser-Zugriff.
type ProjectPolicy struct {
	ProjectID             string
	AllowedOrigins        []string
	AllowedMethods        []HTTPMethod
	AllowedRequestHeaders []AllowedRequestHeader
	AllowedAudiences      []SessionTokenAudience
	OriginOverrides       []OriginPolicy
	ProjectMaxTTLSeconds  int
	RateLimit             RateLimitPolicy
	BrowserIngest         BrowserIngestPolicy
}

// BrowserIngestPolicy ist die Project-spezifische Browser-Ingest-
// Konfiguration aus (RAK-80). Hebt den
// RAK-74-Scope-Cut auf `/api/ingest/*` kontrolliert auf:
//
//  - `Enabled=false` (Default): `/api/ingest/*` bleibt für dieses
//  Project operator-/CLI-only. Browser-Preflights für Origins
//  dieses Projects laufen über die globale, konservative
//  Allowlist (RAK-74-Scope-Cut).
//  - `Enabled=true`: Browser-Origins aus `CORSAllowlist` dürfen
//  Preflight und POST-Pfade nutzen. CSRF und Origin-Pin sind
//  optionale Defense-in-Depth-Felder; sobald gesetzt, gelten sie
//  **strikt** und ein POST ohne entsprechende Header bzw. mit
//  mismatchendem Origin wird mit `403 ingest_browser_*` abgelehnt.
//
// Sicherheitsprofil: ein leeres `CORSAllowlist` bei `Enabled=true`
// ergibt einen aktivierten, aber leeren Allowlist-Pfad — kein Origin
// passt, jeder Preflight läuft als RAK-74-fallback. Das ist erlaubt,
// damit ein Operator die Policy stufenweise aktivieren kann
// (Schema aktiv, Allowlist noch leer).
type BrowserIngestPolicy struct {
	Enabled       bool
	CORSAllowlist []string
	CSRFRequired  bool
	OriginPin     string
}

// IsZero gibt true zurück, wenn keine Browser-Ingest-Konfiguration
// gesetzt wurde — der Default-Pfad (RAK-74-Scope-Cut bleibt strikt).
func (b BrowserIngestPolicy) IsZero() bool {
	return !b.Enabled && len(b.CORSAllowlist) == 0 && !b.CSRFRequired && b.OriginPin == ""
}

// AllowsBrowserOrigin prüft, ob ein konkreter Browser-Origin für
// `/api/ingest/*` zugelassen ist. Setzt `Enabled=true` voraus; ein
// leerer Origin ist im Browser-Kontext nicht erlaubt (im Gegensatz
// zu `ProjectPolicy.AllowsOrigin`, das den CLI-Pfad offen hält).
func (b BrowserIngestPolicy) AllowsBrowserOrigin(origin string) bool {
	if !b.Enabled || origin == "" {
		return false
	}
	for _, o := range b.CORSAllowlist {
		if o == origin {
			return true
		}
	}
	return false
}

// MatchesOriginPin prüft den `OriginPin`-Constraint (falls gesetzt):
// `origin` muss exakt dem konfigurierten Pin entsprechen. Ein leerer
// Pin deaktiviert den Check (Defense-in-Depth ist opt-in).
func (b BrowserIngestPolicy) MatchesOriginPin(origin string) bool {
	if b.OriginPin == "" {
		return true
	}
	return origin == b.OriginPin
}

// EffectiveMaxTTLSeconds liefert die wirksame Project-TTL-Grenze,
// gepinnt gegen die globale Pflichtgrenze `MaxSessionTokenTTLSeconds`.
// Werte ≤ 0 fallen auf den Default zurück (§3.9: „fehlendes
// `project_max_ttl_seconds` defaultet auf exakt 900"). Werte > Default
// werden auf den Default geclampt — die Project-Policy darf niedriger
// begrenzen, aber niemals höher.
func (p ProjectPolicy) EffectiveMaxTTLSeconds() int {
	if p.ProjectMaxTTLSeconds <= 0 {
		return MaxSessionTokenTTLSeconds
	}
	if p.ProjectMaxTTLSeconds > MaxSessionTokenTTLSeconds {
		return MaxSessionTokenTTLSeconds
	}
	return p.ProjectMaxTTLSeconds
}

// AllowsOrigin prüft, ob ein konkreter Origin in der Allowlist steht.
// Leerer Origin gilt als erlaubt, damit der CLI/curl-Pfad (z. B. für
// Issuance vom Server-Skript aus) offen bleibt. Browser-Pfade haben
// im Adapter immer einen `Origin`-Header — fehlende Origins im
// Browser-Kontext werden vor dieser Funktion in abgefangen.
func (p ProjectPolicy) AllowsOrigin(origin string) bool {
	if origin == "" {
		return true
	}
	for _, o := range p.AllowedOrigins {
		if o == origin {
			return true
		}
	}
	return false
}

// AllowsMethod prüft, ob eine HTTP-Methode in der Project-Allowlist
// steht. Wenn die Project-Policy keine Methoden gesetzt hat, gilt
// die globale Allowlist (`POST, OPTIONS`).
func (p ProjectPolicy) AllowsMethod(method HTTPMethod) bool {
	if len(p.AllowedMethods) == 0 {
		return method.IsKnown()
	}
	for _, m := range p.AllowedMethods {
		if m == method {
			return true
		}
	}
	return false
}

// AllowsHeader prüft, ob ein Request-Header in der Project-Allowlist
// steht (case-insensitive). Wenn die Project-Policy keine Header
// gesetzt hat, gilt die globale Allowlist
// (`GlobalPreflightHeaderAllowlist`).
func (p ProjectPolicy) AllowsHeader(header string) bool {
	if header == "" {
		return false
	}
	if len(p.AllowedRequestHeaders) == 0 {
		for _, h := range GlobalPreflightHeaderAllowlist() {
			if strings.EqualFold(string(h), header) {
				return true
			}
		}
		return false
	}
	for _, h := range p.AllowedRequestHeaders {
		if strings.EqualFold(string(h), header) {
			return true
		}
	}
	return false
}

// AllowsAudience prüft, ob eine Session-Token-Audience in der Project-
// Allowlist steht. Wenn die Project-Policy keine Audiences gesetzt
// hat, gilt der globale Default (`playback-events` als einzige Muss-
// Audience im Pflichtpfad). Eine in der globalen Allowlist
// nicht bekannte Audience wird hier auch dann abgelehnt, wenn sie in
// der Project-Allowlist steht — globale Allowlist und Audience-Wert-
// Allowlist müssen synchron sein.
func (p ProjectPolicy) AllowsAudience(aud SessionTokenAudience) bool {
	if !aud.IsKnown() {
		return false
	}
	if len(p.AllowedAudiences) == 0 {
		return aud == SessionTokenAudiencePlaybackEvents
	}
	for _, a := range p.AllowedAudiences {
		if a == aud {
			return true
		}
	}
	return false
}

// AllowsAudienceForOrigin kombiniert Project- und Origin-Override-
// Audience-Checks. Eine Origin-Override mit `RestrictedAudiences` darf
// die Project-Allowlist nur weiter einschränken, niemals erweitern.
func (p ProjectPolicy) AllowsAudienceForOrigin(aud SessionTokenAudience, origin string) bool {
	if !p.AllowsAudience(aud) {
		return false
	}
	if origin == "" {
		return true
	}
	for _, ov := range p.OriginOverrides {
		if ov.Origin == origin {
			return ov.AllowsAudience(aud)
		}
	}
	return true
}

// ValidateAudience prüft eine Session-Token-Audience gegen die
// Project-Policy und liefert den Domain-Fehler aus `errors.go`.
// Adapter mappen das auf `403 auth_session_scope_denied`.
func ValidateAudience(p ProjectPolicy, aud SessionTokenAudience, origin string) error {
	if !p.AllowsAudienceForOrigin(aud, origin) {
		return ErrAuthSessionScopeDenied
	}
	return nil
}

// ValidateOriginAgainstPolicy prüft den Browser-`Origin`-Header gegen
// die Project-Allowlist. Adapter mappen den Fehler auf
// `403 auth_policy_denied`. Leerer Origin gilt als erlaubt — der
// Adapter entscheidet vor dieser Funktion, ob das aktuelle Endpoint
// einen leeren Origin überhaupt zulässt (CLI vs. Browser).
func ValidateOriginAgainstPolicy(p ProjectPolicy, origin string) error {
	if !p.AllowsOrigin(origin) {
		return ErrAuthPolicyDenied
	}
	return nil
}

// ValidateMethodAgainstPolicy prüft die HTTP-Methode gegen die
// Project-Allowlist. Adapter mappen den Fehler auf
// `403 auth_policy_denied`.
func ValidateMethodAgainstPolicy(p ProjectPolicy, method HTTPMethod) error {
	if !p.AllowsMethod(method) {
		return ErrAuthPolicyDenied
	}
	return nil
}

// ValidateHeaderAgainstPolicy prüft einen einzelnen Request-Header
// gegen die Project-Allowlist. Adapter rufen die Funktion in einer
// Schleife pro angefragtem Header (z. B. aus
// `Access-Control-Request-Headers`).
func ValidateHeaderAgainstPolicy(p ProjectPolicy, header string) error {
	if !p.AllowsHeader(header) {
		return ErrAuthPolicyDenied
	}
	return nil
}

// ResolveTTLSeconds berechnet die effektiv anzuwendenden
// `ttl_seconds` aus der Issuance-Anfrage. §3.9-Vertrag:
//
//  - `requested == 0`: nutzt `min(project_max_ttl_seconds, 900)`.
//  - `requested > 0`: muss `<= EffectiveMaxTTLSeconds(p)` sein,
//  sonst `ErrAuthTokenTTLTooLarge` ohne stillen Clamp.
//  - `requested < 0`: ebenfalls `ErrAuthTokenTTLTooLarge`.
//
// Es gibt keinen stillen Clamp, weil Plan §0.5 das explizit verbietet:
// die API muss `422 auth_token_ttl_too_large` melden, damit Clients
// nicht stillschweigend kürzere Tokens bekommen, als sie angefordert
// haben.
func ResolveTTLSeconds(p ProjectPolicy, requested int) (int, error) {
	effectiveMax := p.EffectiveMaxTTLSeconds()
	if requested == 0 {
		return effectiveMax, nil
	}
	if requested < 0 || requested > effectiveMax {
		return 0, ErrAuthTokenTTLTooLarge
	}
	return requested, nil
}

// IssuanceQuota beschreibt die wirksamen Issuance-Limits aus der
// Project-Policy plus globalem Default. Der Application-Service nutzt
// das, um den Issuance-Adapter (`429 auth_issuance_rate_limited`)
// deterministisch zu konfigurieren.
type IssuanceQuota struct {
	Bucket RateLimitBucket
}

// EffectiveIssuanceQuota gibt das wirksame Issuance-Bucket zurück. Ein
// nicht konfiguriertes Bucket (`IsZero`) ist im Pflichtpfad
// kein erlaubter Zustand — der Adapter-Setup muss sicherstellen, dass
// jede Project-Policy ein Issuance-Bucket trägt; diese Funktion
// liefert das Bucket unverändert weiter, der Adapter prüft `IsZero`
// vor dem Use-Case-Aufruf und lehnt fehlkonfigurierte Projects mit
// einem deutlichen Operator-Fehler ab (kein stilles „kein Limit").
func (p ProjectPolicy) EffectiveIssuanceQuota() IssuanceQuota {
	return IssuanceQuota{Bucket: p.RateLimit.IssuanceBucket}
}

// ProjectPolicyFromBaseProject baut eine konservative Default-Policy
// aus einem bestehenden `domain.Project` (§4 Auth) — sinnvoll für die
// Übergangsphase, in der der Static-Resolver noch keine
// Project-Policies kennt. Defaults entsprechen §3.9: globale Methoden-
// /Header-Allowlist, Default-Audience `playback-events`,
// `ProjectMaxTTLSeconds` = Default 900. Rate-Limit-Buckets bleiben
// `IsZero` und müssen vom Adapter-Setup gesetzt werden.
func ProjectPolicyFromBaseProject(p Project) ProjectPolicy {
	return ProjectPolicy{
		ProjectID:             p.ID,
		AllowedOrigins:        append([]string(nil), p.AllowedOrigins...),
		AllowedMethods:        GlobalPreflightMethodAllowlist(),
		AllowedRequestHeaders: GlobalPreflightHeaderAllowlist(),
		AllowedAudiences:      []SessionTokenAudience{SessionTokenAudiencePlaybackEvents},
	}
}

