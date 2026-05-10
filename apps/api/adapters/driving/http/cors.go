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

// CORS-Preflight-Vertrag (`0.12.0`, RAK-74; siehe
// `spec/backend-api-contract.md` §3.9):
//
//   - `OPTIONS` ohne Project-/Session-Token kann kein
//     deterministisches Project-Enforcement; deshalb läuft Preflight
//     gegen eine globale, konservative Origin-Allowlist plus
//     pfadspezifische Methods-Allowlist.
//   - Bekannte Origins: `204` mit gespiegeltem Allow-Origin (niemals
//     `*`), Allow-Methods, Allow-Headers, `Access-Control-Max-Age:
//     600`, `Vary: Origin` und `Cache-Control: no-store`.
//   - Unbekannte Origins: `204` mit leerem Body, **ohne**
//     Allow-Origin/Methods/Headers, aber mit `Vary: Origin` und
//     `Cache-Control: no-store`. Keine Project- oder Origin-
//     Enumeration.
//   - Project-spezifisches Origin-Enforcement passiert beim
//     tatsächlichen `POST` über `domain.Project.IsOriginAllowed`.
//
// `Access-Control-Allow-Credentials` bleibt grundsätzlich aus — das
// SDK nutzt `credentials: "omit"` (NF-31/NF-32; Plan §0.1 schließt
// Cookies für Player-Telemetrie aus).
const (
	playerSDKAllowedMethods = "POST, OPTIONS"
	dashboardAllowedMethods = "GET, OPTIONS"
	analyzeAllowedMethods   = "POST, OPTIONS"
	preflightMaxAge         = "600"
	preflightCacheControl   = "no-store"
	varyHeader              = "Origin, Access-Control-Request-Method, Access-Control-Request-Headers"
)

// preflightAllowedHeaders ist die §3.9-globale Header-Allowlist für
// tokenpflichtige Telemetrie-Endpoints. Reihenfolge ist Wire-stabil
// für Contract-Fixtures.
const preflightAllowedHeaders = "Content-Type, Authorization, X-MTrace-Token, X-MTrace-Session-Token, traceparent"

// dashboardAllowedHeaders bedient die Lese-Pfade (Dashboard,
// SRT-Health). Diese Endpoints unterstützen `0.12.0`-Session Tokens
// nicht — der Auth-Pfad bleibt `X-MTrace-Token`-only — und brauchen
// daher Authorization/X-MTrace-Session-Token nicht in der Allowlist.
const dashboardAllowedHeaders = "Content-Type, X-MTrace-Project, X-MTrace-Token"

// sseAllowedHeaders ergänzt `Last-Event-ID` für den SSE-Reconnect-
// Backfill (Spec §10a). SSE läuft ebenfalls token-/session-frei.
const sseAllowedHeaders = "Content-Type, X-MTrace-Project, X-MTrace-Token, Last-Event-ID"

// corsMiddleware setzt `Vary` plus — sofern Origin in der Union steht
// — den gespiegelten Allow-Origin auf jede ausgehende Antwort.
// Verhindert, dass CDNs/Proxies eine Origin-spezifische Antwort für
// eine andere Origin ausliefern (plan-0.1.0.md §5.1).
func corsMiddleware(next http.Handler, allowlist OriginAllowlist) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		appendVary(w)
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
func playerSDKPreflightHandler(allowlist OriginAllowlist) http.HandlerFunc {
	return preflightHandler(allowlist, playerSDKAllowedMethods, preflightAllowedHeaders)
}

// dashboardPreflightHandler bedient die Read-Endpoints (Dashboard,
// SRT-Health). `GET, OPTIONS`-Methods und Legacy-Header-Allowlist —
// diese Endpoints konsumieren keine Session Tokens.
func dashboardPreflightHandler(allowlist OriginAllowlist) http.HandlerFunc {
	return preflightHandler(allowlist, dashboardAllowedMethods, dashboardAllowedHeaders)
}

// ssePreflightHandler bedient `OPTIONS /api/stream-sessions/stream`
// (plan-0.4.0 §5 H4). Methods sind `GET, OPTIONS`; Allow-Headers
// ergänzen `Last-Event-ID` für den fetch-basierten SSE-Reconnect-
// Backfill (Spec §10a).
func ssePreflightHandler(allowlist OriginAllowlist) http.HandlerFunc {
	return preflightHandler(allowlist, dashboardAllowedMethods, sseAllowedHeaders)
}

// analyzePreflightHandler bedient `OPTIONS /api/analyze`. Das
// Analyze-Endpoint nutzt Session Tokens analog `playback-events`
// (Plan §3.9 Auth-Matrix), deshalb die erweiterte Header-Allowlist.
func analyzePreflightHandler(allowlist OriginAllowlist) http.HandlerFunc {
	return preflightHandler(allowlist, analyzeAllowedMethods, preflightAllowedHeaders)
}

// preflightHandler ist der zentrale §3.9-konforme Preflight-
// Generator. Bekannte Origin → `204` mit Allow-Origin/Methods/
// Headers/Max-Age/Vary/Cache-Control; unbekannte Origin → `204` mit
// **nur** Vary und Cache-Control. Beide Antworten haben einen leeren
// Body und sind wire-byte-stabil.
func preflightHandler(allowlist OriginAllowlist, methods, headers string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appendVary(w)
		w.Header().Set("Cache-Control", preflightCacheControl)
		origin := r.Header.Get("Origin")
		if !allowlist.IsOriginInGlobalUnion(origin) {
			// §3.9 minimale Ablehnung: `204` ohne Allow-* Header,
			// damit kein Origin-/Project-Enumeration-Signal leakt.
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

// appendVary fügt die Pflicht-`Vary`-Tokens zum Header an, ohne
// existierende Werte zu überschreiben.
func appendVary(w http.ResponseWriter) {
	existing := w.Header().Get("Vary")
	if existing == "" {
		w.Header().Set("Vary", varyHeader)
		return
	}
	if !strings.Contains(existing, "Origin") {
		w.Header().Set("Vary", existing+", "+varyHeader)
	}
}
