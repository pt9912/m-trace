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

// corsHeaders sind die statischen CORS-Antwort-Header pro Pfad-Klasse
// (plan-0.1.0.md §5.1, CORS Variante B). Beide Klassen lassen
// `Access-Control-Allow-Credentials` bewusst weg — das SDK nutzt
// `credentials: "omit"` (NF-31/NF-32).
var (
	playerSDKAllowedMethods = "POST, OPTIONS"
	dashboardAllowedMethods = "GET, OPTIONS"
	allowedHeaders          = "Content-Type, X-MTrace-Project, X-MTrace-Token"
	preflightMaxAge         = "600"
	varyHeader              = "Origin, Access-Control-Request-Method, Access-Control-Request-Headers"
)

// corsMiddleware setzt den `Vary`-Header auf jede ausgehende Antwort.
// Das verhindert, dass CDNs/Proxies eine Origin-spezifische Antwort
// für eine andere Origin ausliefern (plan-0.1.0.md §5.1).
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		appendVary(w)
		next.ServeHTTP(w, r)
	})
}

// playerSDKPreflightHandler bedient `OPTIONS /api/playback-events`.
// Origin in der globalen Union → `204 No Content` mit konkretem
// `Access-Control-Allow-Origin` (niemals `*`); sonst `403 Forbidden`.
func playerSDKPreflightHandler(allowlist OriginAllowlist) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appendVary(w)
		origin := r.Header.Get("Origin")
		if !allowlist.IsOriginInGlobalUnion(origin) {
			writeStatus(w, http.StatusForbidden)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", playerSDKAllowedMethods)
		w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
		w.Header().Set("Access-Control-Max-Age", preflightMaxAge)
		w.WriteHeader(http.StatusNoContent)
	}
}

// dashboardPreflightHandler bedient `OPTIONS` für die Lese-Pfade
// (`/api/stream-sessions`, `/api/stream-sessions/{id}`, `/api/health`).
// Methods-Header verbeißt sich auf `GET, OPTIONS` (NF-35 gilt nur
// für den SDK-Telemetrie-Pfad, plan-0.1.0.md §5.1).
func dashboardPreflightHandler(allowlist OriginAllowlist) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appendVary(w)
		origin := r.Header.Get("Origin")
		if !allowlist.IsOriginInGlobalUnion(origin) {
			writeStatus(w, http.StatusForbidden)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", dashboardAllowedMethods)
		w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
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
