package application

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// URL-Redaction-Matrix aus spec/telemetry-model.md §1.4 und plan-0.4.0
// §4.4 DoD-Item 5. Vor Persistenz und Dashboard-Anzeige werden alle
// URL-verdächtigen Meta-Keys redigiert; nichts roh persistiert.
//
// Bekannte URL-Keys aus der Matrix (case-insensitiv): `url`, `uri`,
// `manifest_url`, `segment_url`, `media_url`, `network.url`,
// `request.url`, `response.url`. `network.redacted_url` ist
// reserviert und wird in validateReservedEventMeta strikt geprüft —
// die Redaction lässt diesen Key unangetastet.
//
// Signed-Query-Parameter aus der Matrix (informell, weil Queries
// bereits vollständig entfernt werden): `token`, `signature`, `sig`,
// `expires`, `key`, `policy` (case-insensitiv).
const redactedTokenSegment = ":redacted"

// hexSegmentPattern erfasst Hex-Strings mit gerader Länge ≥ 32
// (Token-Heuristik aus spec/telemetry-model.md §1.4).
// `regexp.MustCompile`-Globals werden vom gochecknoglobals-Default
// ignoriert (kompilierte Regex ohne Mutationsfläche).
var hexSegmentPattern = regexp.MustCompile(`^(?:[0-9A-Fa-f]{2}){16,}$`)

// jwtSegmentPattern erfasst JWT-/SAS-artige Pfadsegmente (drei mit
// Punkten getrennte Base64URL-Blöcke).
var jwtSegmentPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+$`)

// isKnownURLMetaKey deckt die explizit benannte Liste der URL-
// verdächtigen Meta-Keys aus spec/telemetry-model.md §1.4 case-
// insensitiv ab. Die switch-Form ersetzt eine globale Map, damit
// gochecknoglobals nicht ausschlägt.
func isKnownURLMetaKey(key string) bool {
	switch strings.ToLower(key) {
	case "url",
		"uri",
		"manifest_url",
		"segment_url",
		"media_url",
		"network.url",
		"request.url",
		"response.url":
		return true
	default:
		return false
	}
}

// redactEventMetaURLs mutiert die übergebene Meta-Map und ersetzt alle
// URL-verdächtigen String-Werte durch redigierte Repräsentanten gemäß
// spec/telemetry-model.md §1.4. `network.redacted_url` wird nicht
// erneut redigiert — der Key ist bereits durch validateReservedEventMeta
// strikt geprüft. Reserved-Key-Validation passiert immer **vor** der
// Redaction; Aufrufer müssen die Reihenfolge einhalten, damit ein
// invalider `network.redacted_url`-Wert weiterhin 422 auslöst.
func redactEventMetaURLs(meta domain.EventMeta) {
	if len(meta) == 0 {
		return
	}
	for k, v := range meta {
		if k == metaKeyNetworkRedactedURL {
			continue
		}
		s, ok := v.(string)
		if !ok || s == "" {
			continue
		}
		if !shouldRedactMetaValue(k, s) {
			continue
		}
		meta[k] = redactURLString(s)
	}
}

// shouldRedactMetaValue entscheidet, ob ein String-Wert als URL
// behandelt und redigiert werden muss. Bekannte URL-Keys (case-
// insensitiv) werden immer redigiert; bei unbekannten Keys gilt der
// Heuristik-Check (`://` oder absolute URL parsebar).
func shouldRedactMetaValue(key, value string) bool {
	if isKnownURLMetaKey(key) {
		return true
	}
	if strings.Contains(value, "://") {
		return true
	}
	if u, err := url.Parse(value); err == nil && u.IsAbs() {
		return true
	}
	return false
}

// redactURLString redigiert einen einzelnen URL-Wert gemäß
// spec/telemetry-model.md §1.4: Scheme/Host bleiben erhalten,
// userinfo/Query/Fragment werden entfernt, tokenartige Pfadsegmente
// werden durch ":redacted" ersetzt. Unparsebare Werte werden komplett
// als ":redacted" persistiert (kein roher Fallback).
func redactURLString(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return redactedTokenSegment
	}
	out := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   redactPathSegments(u.EscapedPath()),
	}
	return out.String()
}

// redactPathSegments ersetzt tokenartige Segmente durch ":redacted".
// Erhält die führenden/trailing Slashes des Pfades.
func redactPathSegments(escapedPath string) string {
	if escapedPath == "" {
		return ""
	}
	hadLeading := strings.HasPrefix(escapedPath, "/")
	hadTrailing := strings.HasSuffix(escapedPath, "/") && len(escapedPath) > 1
	trimmed := strings.Trim(escapedPath, "/")
	if trimmed == "" {
		return escapedPath
	}
	parts := strings.Split(trimmed, "/")
	for i, seg := range parts {
		if isTokenLikePathSegment(seg) {
			parts[i] = redactedTokenSegment
		}
	}
	rebuilt := strings.Join(parts, "/")
	if hadLeading {
		rebuilt = "/" + rebuilt
	}
	if hadTrailing {
		rebuilt += "/"
	}
	return rebuilt
}

// isTokenLikePathSegment implementiert die Token-Heuristik aus
// spec/telemetry-model.md §1.4:
//
//   - mindestens 24 Zeichen UND mindestens 80 % aus [A-Za-z0-9_-];
//   - ODER Hex-String mit gerader Länge mindestens 32;
//   - ODER bekanntes JWT-/SAS-/Signed-URL-Muster (drei Base64URL-Blöcke
//     getrennt durch Punkte).
func isTokenLikePathSegment(seg string) bool {
	if len(seg) == 0 {
		return false
	}
	if hexSegmentPattern.MatchString(seg) {
		return true
	}
	if jwtSegmentPattern.MatchString(seg) {
		return true
	}
	if len(seg) < 24 {
		return false
	}
	allowed := 0
	for _, r := range seg {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') ||
			(r >= '0' && r <= '9') || r == '_' || r == '-' {
			allowed++
		}
	}
	// "mindestens 80 % aus [A-Za-z0-9_-]"
	return allowed*100 >= len(seg)*80
}

// isAlreadyRedactedURL prüft, ob ein URL-Wert die Redaction-Matrix
// einhält: parsebar, Scheme + Host vorhanden, kein userinfo, keine
// Query, kein Fragment, keine tokenartigen Pfadsegmente. Wird in
// validateRedactedURLValue für `network.redacted_url` verwendet.
func isAlreadyRedactedURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	if u.Scheme == "" || u.Host == "" {
		return false
	}
	if u.User != nil {
		return false
	}
	if u.RawQuery != "" || u.Fragment != "" {
		return false
	}
	for _, seg := range strings.Split(strings.Trim(u.EscapedPath(), "/"), "/") {
		if isTokenLikePathSegment(seg) {
			return false
		}
	}
	return true
}
