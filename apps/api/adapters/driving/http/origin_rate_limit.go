package http

import (
	"log/slog"
	"net"
	"net/http"
	"strings"

	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// originRateLimitMiddleware (R-22) sitzt vor
// `POST /api/auth/session-tokens` und `POST /api/playback-events` und
// rejected Anfragen über `429 origin_rate_limited`, wenn der per-Key-
// Token-Bucket leer ist. Reihenfolge gemäß §6 §0.1 (Plan-DoD):
// Origin-Limit zuerst, Project-Limit (Issuance- oder Event-Counter)
// erst danach.
//
// Key-Wahl:
//  - `trustXFF=false` (Default): `clientIPFromRemoteAddr(r)` —
//  entkoppelt den Port aus `host:port` und nutzt den Host als Key.
//  - `trustXFF=true`: nutzt **das letzte** (rechteste) Element der
//  `X-Forwarded-For`-Liste als Client-IP. Das ist nur korrekt,
//  wenn der Operator dem Reverse-Proxy traut, der XFF setzt
//  (sonst kann ein Client den Header beliebig vorgeben und so
//  die Limit-Buckets stündlich switchen). Operator-Doku in
//  `docs/user/auth.md` §5.8.
//
// `limiter=nil` → Middleware ist No-Op (Disabled-Pfad aus dem
// Boot-Validator).
func originRateLimitMiddleware(next http.Handler, limiter driven.OriginRateLimiter, trustXFF bool, logger *slog.Logger) http.Handler {
	if limiter == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := originLimiterKey(r, trustXFF)
		allowed, err := limiter.Allow(r.Context(), key)
		if err != nil {
			if logger != nil {
				logger.Error("origin rate limiter error", "error", err)
			}
			writeStatus(w, http.StatusInternalServerError)
			return
		}
		if !allowed {
			writeJSON(w, http.StatusTooManyRequests, map[string]string{
				"error": "origin_rate_limited",
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}

// originLimiterKey extrahiert den Bucket-Key für eine Anfrage. Bei
// fehlender RemoteAddr oder leerem XFF-Header liefert die Funktion
// einen Empty-String — der Limiter mappt das auf `(true, nil)` (No-Op,
// damit lokale Lab-Pfade ohne RemoteAddr nicht blockiert werden).
func originLimiterKey(r *http.Request, trustXFF bool) string {
	if trustXFF {
		if ip := xffClientIP(r); ip != "" {
			return "ip:" + ip
		}
	}
	return clientIPFromRemoteAddr(r)
}

// xffClientIP liefert das **letzte** Element der `X-Forwarded-For`-
// Liste — die vom Reverse-Proxy zuletzt hinzugefügte IP (also der Hop
// direkt vor dem Server; bei genau einem Proxy die Client-IP) — oder
// "" ohne/bei leerem Header. Nur hinter der Trust-Boundary
// (MTRACE_TRUST_FORWARDED_FOR) verwenden: ohne vertrauten Proxy ist
// der Header client-kontrolliert. Geteilt zwischen Origin-Limiter-Key
// und der client_ip-Dimension des Ingest-Limiters (R-26 b: hinter
// LB/Proxy ist RemoteAddr die Proxy-IP — ohne XFF teilen sich dort
// alle Clients einen client_ip-Bucket).
func xffClientIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff == "" {
		return ""
	}
	parts := strings.Split(xff, ",")
	return strings.TrimSpace(parts[len(parts)-1])
}

// clientIPFromRemoteAddr entkoppelt `host:port` aus `r.RemoteAddr`.
// Bei `host`-only (Test-Server) wird der ganze Wert genommen. Bei
// nicht-parsbarem RemoteAddr liefert die Funktion Empty-String.
func clientIPFromRemoteAddr(r *http.Request) string {
	raw := r.RemoteAddr
	if raw == "" {
		return ""
	}
	host, _, err := net.SplitHostPort(raw)
	if err != nil {
		// `r.RemoteAddr` kann in Tests auch ohne Port kommen.
		host = raw
	}
	if host == "" {
		return ""
	}
	return "ip:" + host
}
