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
	if ip := requestClientIP(r, trustXFF); ip != "" {
		return "ip:" + ip
	}
	return ""
}

// requestClientIP ist DIE Client-IP-Auflösung für Rate-Limit-Zwecke —
// geteilt zwischen dem Origin-/IP-Limiter-Key und der client_ip-
// Dimension des Ingest-Limiters (R-26 b: hinter LB/Proxy ist RemoteAddr
// die Proxy-IP — ohne XFF teilen sich dort alle Clients einen Bucket).
// Mit trustXFF zählt das letzte XFF-Element, sonst (und als Fallback
// bei fehlendem/ungültigem XFF) r.RemoteAddr. Liefert die nackte IP;
// Aufrufer setzen ggf. ihren eigenen Key-Prefix.
func requestClientIP(r *http.Request, trustXFF bool) string {
	if trustXFF {
		if ip := xffClientIP(r); ip != "" {
			return ip
		}
	}
	return clientIPFromRequest(r)
}

// xffClientIP liefert das **letzte** Element der `X-Forwarded-For`-
// Liste — die vom Reverse-Proxy zuletzt hinzugefügte IP (also der Hop
// direkt vor dem Server; bei genau einem Proxy die Client-IP) — oder
// "" ohne/bei leerem Header. Nur hinter der Trust-Boundary
// (MTRACE_TRUST_FORWARDED_FOR) verwenden: ohne vertrauten Proxy ist
// der Header client-kontrolliert.
//
// Der Wert wird als IP VALIDIERT (net.ParseIP, kanonisiert): die
// XFF-abgeleitete client_ip landet raw in Redis-Bucket-Keys
// (mtrace:ingest:ip:<val>) — ohne Validierung wäre der Key hinter
// einem nicht-sanitisierenden Proxy ein unbounded client-kontrollierter
// String (Key-Länge/-Kardinalität auf dem geteilten Redis). Nicht-IP-
// Werte fallen auf RemoteAddr zurück.
func xffClientIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff == "" {
		return ""
	}
	parts := strings.Split(xff, ",")
	ip := net.ParseIP(strings.TrimSpace(parts[len(parts)-1]))
	if ip == nil {
		return ""
	}
	return ip.String()
}

// clientIPFromRemoteAddr ist der "ip:"-präfixte Origin-Limiter-Wrapper
// um den geteilten RemoteAddr-Parser (clientIPFromRequest, handler.go).
func clientIPFromRemoteAddr(r *http.Request) string {
	host := clientIPFromRequest(r)
	if host == "" {
		return ""
	}
	return "ip:" + host
}
