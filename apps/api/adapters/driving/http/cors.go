package http

import (
	"net/http"
	"strings"
)

// OriginAllowlist abstrahiert die globale Union der Allowed-Origins
// über alle bekannten Projects. Der Auth-Adapter (StaticProjectResolver)
// erfüllt das Interface; alternative Resolver (z. B. DB-gestützt) auch.
type OriginAllowlist interface {
	IsOriginInGlobalUnion(origin string) bool
}

// BrowserIngestOriginAllowlist abstrahiert die Union der Browser-
// Origins, die in einer aktivierten
// `domain.BrowserIngestPolicy.CORSAllowlist` stehen (`0.12.5`/RAK-80).
// Wird vom `/api/ingest/*`-Preflight-Handler genutzt, um den
// RAK-74-Scope-Cut kontrolliert aufzuheben. `nil`-Implementierung
// (oder eine Implementierung, die für jede Origin `false` liefert)
// hält den RAK-74-Scope-Cut strikt.
type BrowserIngestOriginAllowlist interface {
	IsBrowserIngestOriginAllowed(origin string) bool
}

// BrowserIngestPolicies bündelt Allowlist- und Policy-Lookup-API für
// den Browser-Ingest-Pfad (`0.12.5`/RAK-80). Der
// `auth.InMemoryProjectPolicyResolver` erfüllt das Interface — er
// bedient beide Methoden in einer Implementierung. `nil` deaktiviert
// den Browser-Ingest-Pfad komplett (RAK-74-Scope-Cut bleibt strikt).
type BrowserIngestPolicies interface {
	BrowserIngestOriginAllowlist
	BrowserIngestPolicyLookup
}

// PreflightMetrics bekommt einen Hook für jede `OPTIONS`-Antwort,
// die wegen unbekannter Origin minimal abgewiesen wurde (`0.12.0`
// / Review-Finding Y3). `path` ist die registrierte
// Preflight-Route (z. B. `/api/playback-events`). `nil` deaktiviert
// das Metrik-Reporting.
type PreflightMetrics interface {
	CORSPreflightRefused(path string)
}

// CORS-Preflight-Vertrag (`0.12.0`, RAK-74; siehe
// `spec/backend-api-contract.md` §3.9):
//
//  - `OPTIONS` ohne Project-/Session-Token kann kein
//  deterministisches Project-Enforcement; deshalb läuft Preflight
//  gegen eine globale, konservative Origin-Allowlist plus
//  pfadspezifische Methods-Allowlist.
//  - Bekannte Origins: `204` mit gespiegeltem Allow-Origin (niemals
//  `*`), Allow-Methods, Allow-Headers, `Access-Control-Max-Age:
//  600`, `Vary: Origin` und `Cache-Control: no-store`.
//  - Unbekannte Origins: `204` mit leerem Body, **ohne**
//  Allow-Origin/Methods/Headers, aber mit `Vary: Origin` und
//  `Cache-Control: no-store`. Keine Project- oder Origin-
//  Enumeration.
//  - Project-spezifisches Origin-Enforcement passiert beim
//  tatsächlichen `POST` über `domain.Project.IsOriginAllowed`.
//
// `Access-Control-Allow-Credentials` bleibt grundsätzlich aus — das
// SDK nutzt `credentials: "omit"` (NF-31/NF-32; Plan §0.1 schließt
// Cookies für Player-Telemetrie aus).
const (
	playerSDKAllowedMethods = "POST, OPTIONS"
	dashboardAllowedMethods = "GET, OPTIONS"
	analyzeAllowedMethods   = "POST, OPTIONS"
	// browserIngestAllowedMethods sind die Methoden, die der Browser-
	// Ingest-Pfad nach RAK-80 erlaubt: POST für Stream-Lifecycle-Hooks
	// und Stream-Anlage; OPTIONS für den Preflight selbst.
	browserIngestAllowedMethods = "POST, OPTIONS"
	preflightMaxAge         = "600"
	// preflightCacheControl: §3.9 fordert `no-store` für jede
	// Preflight-Antwort. Trade-off: shared Caches (CDNs/Proxies)
	// sehen `no-store` und cachen den Preflight nicht — das pinnt
	// sicheres Verhalten nach Operator-Token-/Origin-Rotation.
	// Browser-Per-Origin-Caches behandeln `Cache-Control` und
	// `Access-Control-Max-Age` per Fetch-Spec unabhängig, deshalb
	// bleibt `Max-Age: 600` für Browser-Preflight-Caching wirksam.
	preflightCacheControl = "no-store"
)

// pflichtVaryTokens listet die `Vary`-Tokens für jede ausgehende
// Antwort. `Origin` ist §3.9-Vorgabe; `Access-Control-Request-*`
// sind eine additive Härtung, damit Proxies Preflight-Antworten nicht
// nach Methoden-/Header-Mix mischen — die spec erlaubt zusätzliche
// `Vary`-Tokens. Funktion statt Global, damit der `gochecknoglobals`-
// Linter keinen veränderlichen Package-State sieht.
func pflichtVaryTokens() []string {
	return []string{
		"Origin",
		"Access-Control-Request-Method",
		"Access-Control-Request-Headers",
	}
}

// preflightAllowedHeaders ist die §3.9-globale Header-Allowlist für
// tokenpflichtige Telemetrie-Endpoints. Reihenfolge ist Wire-stabil
// für Contract-Fixtures.
const preflightAllowedHeaders = "Content-Type, Authorization, X-MTrace-Token, X-MTrace-Session-Token, traceparent"

// dashboardAllowedHeaders bedient die Lese-Pfade (Dashboard,
// SRT-Health). Diese Endpoints unterstützen Session Tokens
// nicht — der Auth-Pfad bleibt `X-MTrace-Token`-only — und brauchen
// daher Authorization/X-MTrace-Session-Token nicht in der Allowlist.
//
// TODO(`0.13.0` o. später, F-111..F-113): Wenn Dashboard auf Session
// Tokens migriert wird, muss diese Allowlist auf
// `preflightAllowedHeaders` umgestellt werden. Eine stille Beibehaltung
// der Legacy-Liste produziert Browser-seitige Preflight-Failures, die
// schwer zu diagnostizieren sind. Review-Finding G3 / R-N-Eintrag im
// `risks-backlog.md`.
const dashboardAllowedHeaders = "Content-Type, X-MTrace-Project, X-MTrace-Token"

// sseAllowedHeaders ergänzt `Last-Event-ID` für den SSE-Reconnect-
// Backfill (Spec §10a). SSE läuft ebenfalls token-/session-frei.
const sseAllowedHeaders = "Content-Type, X-MTrace-Project, X-MTrace-Token, Last-Event-ID"

// browserIngestAllowedHeaders deckt den Browser-Ingest-Pfad ab
// (`0.12.5`/RAK-80). `X-MTrace-CSRF` ist der Defense-in-Depth-Header,
// wenn `BrowserIngestPolicy.CSRFRequired` aktiv ist. `X-MTrace-Token`
// ist die heutige Operator-/Browser-Auth.
const browserIngestAllowedHeaders = "Content-Type, X-MTrace-Token, X-MTrace-CSRF"

// corsMiddleware setzt `Vary` plus — sofern Origin in der Union steht
// — den gespiegelten Allow-Origin auf jede ausgehende Antwort.
// Verhindert, dass CDNs/Proxies eine Origin-spezifische Antwort für
// eine andere Origin ausliefern.
//
// `OPTIONS`-Preflights laufen über den dedizierten `preflightHandler`
// und setzen ihren Allow-Origin selbst — die Middleware skipt den
// Allow-Origin-Set in dem Fall, damit kein doppelter Header
// geschrieben wird (Code-Reading-Cost; Review G4).
func corsMiddleware(next http.Handler, allowlist OriginAllowlist) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		appendVary(w)
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		origin := r.Header.Get("Origin")
		if allowlist.IsOriginInGlobalUnion(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		next.ServeHTTP(w, r)
	})
}

// playerSDKPreflightHandler bedient `OPTIONS /api/playback-events`
// und `OPTIONS /api/auth/session-tokens` — der Player-SDK-Pfad hat
// in `0.12.0` Bearer-/Session-Token-Support, deshalb die
// erweiterte `preflightAllowedHeaders`-Allowlist.
func playerSDKPreflightHandler(allowlist OriginAllowlist, metrics PreflightMetrics) http.HandlerFunc {
	return preflightHandler(allowlist, metrics, playerSDKAllowedMethods, preflightAllowedHeaders)
}

// dashboardPreflightHandler bedient die Read-Endpoints (Dashboard,
// SRT-Health). `GET, OPTIONS`-Methods und Legacy-Header-Allowlist —
// diese Endpoints konsumieren keine Session Tokens.
func dashboardPreflightHandler(allowlist OriginAllowlist, metrics PreflightMetrics) http.HandlerFunc {
	return preflightHandler(allowlist, metrics, dashboardAllowedMethods, dashboardAllowedHeaders)
}

// browserIngestPreflightHandler bedient `OPTIONS /api/ingest/*`,
// wenn mindestens ein Project eine aktivierte
// `domain.BrowserIngestPolicy` hat (`0.12.5`/RAK-80). Allowlist
// kommt aus dem Project-Policy-Resolver, nicht aus der globalen
// Project-Origin-Union — damit Origins, die nur in einer
// BrowserIngest-Policy stehen, durchkommen.
//
// Antwort-Wire (analog `dashboardPreflightHandler`):
//  - Match → `204` + Allow-Origin/Methods/Headers/Max-Age, plus
//  Vary und `Cache-Control: no-store`.
//  - kein Match → `204` ohne Allow-* (RAK-74-Scope-Cut-Verhalten);
//  `preflightMetrics.CORSPreflightRefused(path)` wird inkrementiert.
//
// `allowlist == nil` wirft auf jeden Origin „kein Match" und ist
// gleichwertig zum dashboard-Preflight ohne Browser-Ingest-Policy.
func browserIngestPreflightHandler(allowlist BrowserIngestOriginAllowlist, metrics PreflightMetrics) http.HandlerFunc {
	wrapper := browserIngestAllowlistAdapter{inner: allowlist}
	return preflightHandler(wrapper, metrics, browserIngestAllowedMethods, browserIngestAllowedHeaders)
}

// browserIngestAllowlistAdapter mappt eine
// `BrowserIngestOriginAllowlist` auf die `OriginAllowlist`-Form, die
// `preflightHandler` konsumiert.
type browserIngestAllowlistAdapter struct {
	inner BrowserIngestOriginAllowlist
}

func (a browserIngestAllowlistAdapter) IsOriginInGlobalUnion(origin string) bool {
	if a.inner == nil {
		return false
	}
	return a.inner.IsBrowserIngestOriginAllowed(origin)
}

// ssePreflightHandler bedient `OPTIONS /api/stream-sessions/stream`
//  Methods sind `GET, OPTIONS`; Allow-Headers
// ergänzen `Last-Event-ID` für den fetch-basierten SSE-Reconnect-
// Backfill (Spec §10a).
func ssePreflightHandler(allowlist OriginAllowlist, metrics PreflightMetrics) http.HandlerFunc {
	return preflightHandler(allowlist, metrics, dashboardAllowedMethods, sseAllowedHeaders)
}

// analyzePreflightHandler bedient `OPTIONS /api/analyze`. Das
// Analyze-Endpoint nutzt Session Tokens analog `playback-events`
// (Plan §3.9 Auth-Matrix), deshalb die erweiterte Header-Allowlist.
func analyzePreflightHandler(allowlist OriginAllowlist, metrics PreflightMetrics) http.HandlerFunc {
	return preflightHandler(allowlist, metrics, analyzeAllowedMethods, preflightAllowedHeaders)
}

// preflightHandler ist der zentrale §3.9-konforme Preflight-
// Generator. Bekannte Origin → `204` mit Allow-Origin/Methods/
// Headers/Max-Age/Vary/Cache-Control; unbekannte Origin → `204` mit
// **nur** Vary und Cache-Control. Beide Antworten haben einen leeren
// Body und sind wire-byte-stabil.
//
// `metrics` ist optional — wenn nicht-`nil`, wird jede minimale
// Ablehnung als `mtrace_cors_preflight_refused_total{path=…}`
// inkrementiert (Review-Finding Y3). `path` ist `r.URL.Path` und
// damit auf die registrierten Preflight-Routen begrenzt — die
// Cardinality bleibt klein.
func preflightHandler(allowlist OriginAllowlist, metrics PreflightMetrics, methods, headers string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appendVary(w)
		w.Header().Set("Cache-Control", preflightCacheControl)
		origin := r.Header.Get("Origin")
		if !allowlist.IsOriginInGlobalUnion(origin) {
			// 9 minimale Ablehnung: `204` ohne Allow-* Header,
			// damit kein Origin-/Project-Enumeration-Signal leakt.
			if metrics != nil {
				metrics.CORSPreflightRefused(r.URL.Path)
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", methods)
		w.Header().Set("Access-Control-Allow-Headers", headers)
		w.Header().Set("Access-Control-Max-Age", preflightMaxAge)
		w.WriteHeader(http.StatusNoContent)
	}
}

// appendVary unioniert die Pflicht-`Vary`-Tokens (siehe `varyTokens`)
// mit etwaigen vorhandenen Vary-Werten. Tokens werden einzeln
// geprüft, damit ein vorgelagerter Middleware-Setter (z. B. `Vary:
// Origin` allein) nicht die Request-Method-/Header-Tokens
// verschluckt — Review-Finding Y2.
func appendVary(w http.ResponseWriter) {
	existing := strings.TrimSpace(w.Header().Get("Vary"))
	for _, token := range pflichtVaryTokens() {
		if existing == "" {
			existing = token
			continue
		}
		if !varyContainsToken(existing, token) {
			existing = existing + ", " + token
		}
	}
	w.Header().Set("Vary", existing)
}

// varyContainsToken prüft, ob `token` als ganzes Komma-getrenntes
// Element in `header` steht (case-insensitive). Verhindert
// Substring-False-Positives (z. B. `Origin` matched nicht
// fälschlich `OriginCustom`).
func varyContainsToken(header, token string) bool {
	for _, part := range strings.Split(header, ",") {
		if strings.EqualFold(strings.TrimSpace(part), token) {
			return true
		}
	}
	return false
}
