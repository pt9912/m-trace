package auth

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// ParseSigningKeysEnv baut den Signing-Key-Ring aus den
// `0.12.5`-ENV-Variablen `MTRACE_AUTH_SIGNING_KEYS` und
// `MTRACE_AUTH_SIGNING_ACTIVE_KID` (RAK-78). Wenn beide leer sind,
// fällt die Funktion auf den `0.12.0`-Single-Key-Pfad zurück
// (`fallbackKey`/`fallbackKID`); ein leerer Fallback-Key bedeutet
// `noKeysConfigured == true` und der Caller muss entscheiden, ob
// er den Lab-Default-Pfad aktiviert oder hart fehlschlägt.
//
// Multi-Key-Format: `kid_a:<base64url-secret>[,kid_b:<base64url-secret>,…]`.
// Whitespace um Kommas und Doppelpunkte wird getrimmt; leere Einträge
// werden ignoriert. Duplikate, leere KIDs und ungültige Base64-Werte
// liefern einen deterministischen Fehler — kein stiller Fallback.
//
// Restart-Stabilität (RAK-72/RAK-78): dieselbe ENV-Konfiguration
// produziert deterministisch denselben Resolver, weil der Parser
// die Eingangsreihenfolge erhält und Secret-Material defensiv kopiert.
//
// Backwards-Compat: bei `keysEnv == ""` darf `activeKIDEnv` auch leer
// sein; der Fallback-Pfad nutzt dann `fallbackKID` (ggf. ergänzt durch
// einen Caller-Default). Bei `keysEnv != ""` ist `activeKIDEnv`
// erforderlich.
func ParseSigningKeysEnv(
	keysEnv, activeKIDEnv string,
	fallbackKey, fallbackKID string,
	now time.Time,
) (keys []domain.SessionSigningKey, activeKID domain.SigningKeyID, noKeysConfigured bool, err error) {
	trimmedKeys := strings.TrimSpace(keysEnv)
	if trimmedKeys != "" {
		parsed, parseErr := parseSigningKeysList(trimmedKeys, now)
		if parseErr != nil {
			return nil, "", false, parseErr
		}
		active := domain.SigningKeyID(strings.TrimSpace(activeKIDEnv))
		if active == "" {
			return nil, "", false, errors.New("auth: MTRACE_AUTH_SIGNING_ACTIVE_KID is required when MTRACE_AUTH_SIGNING_KEYS is set")
		}
		// Sicherstellen, dass activeKID in der Liste vorkommt — der
		// Resolver-Konstruktor prüft das auch, aber ein expliziter
		// Frontload-Fehler ist operator-freundlicher.
		found := false
		for _, k := range parsed {
			if k.KID == active {
				found = true
				break
			}
		}
		if !found {
			return nil, "", false, fmt.Errorf("auth: MTRACE_AUTH_SIGNING_ACTIVE_KID %q is not present in MTRACE_AUTH_SIGNING_KEYS", string(active))
		}
		return parsed, active, false, nil
	}

	// Backwards-Compat-Pfad: einzelner Key über alten ENV-Style.
	trimmedFallback := strings.TrimSpace(fallbackKey)
	if trimmedFallback == "" {
		// Kein Multi-Key, kein Single-Key — Caller entscheidet (Lab-Default
		// oder hard-fail). `activeKID` wird auf den Fallback-KID gesetzt,
		// damit der Caller bei aktiviertem Lab-Default-Pfad direkt damit
		// weiterarbeiten kann.
		return nil, domain.SigningKeyID(strings.TrimSpace(fallbackKID)), true, nil
	}
	decoded, decodeErr := base64DecodeURLSafe(trimmedFallback)
	if decodeErr != nil {
		return nil, "", false, fmt.Errorf("auth: MTRACE_AUTH_SIGNING_KEY base64 decode: %w", decodeErr)
	}
	kid := domain.SigningKeyID(strings.TrimSpace(fallbackKID))
	if kid == "" {
		return nil, "", false, errors.New("auth: MTRACE_AUTH_SIGNING_KID is required when MTRACE_AUTH_SIGNING_KEY is set")
	}
	return []domain.SessionSigningKey{
		newDomainSigningKey(kid, decoded, now),
	}, kid, false, nil
}

// parseSigningKeysList parst `kid_a:b64,kid_b:b64`-Listen. Whitespace
// um Kommas und Doppelpunkte wird getrimmt; leere Tokens (z. B. durch
// doppelte Kommas) werden übersprungen.
func parseSigningKeysList(raw string, now time.Time) ([]domain.SessionSigningKey, error) {
	parts := strings.Split(raw, ",")
	out := make([]domain.SessionSigningKey, 0, len(parts))
	seen := make(map[domain.SigningKeyID]struct{}, len(parts))
	for i, p := range parts {
		token := strings.TrimSpace(p)
		if token == "" {
			continue
		}
		colon := strings.IndexByte(token, ':')
		if colon <= 0 || colon == len(token)-1 {
			return nil, fmt.Errorf("auth: MTRACE_AUTH_SIGNING_KEYS entry %d malformed (expected kid:base64), got %q", i, token)
		}
		kid := domain.SigningKeyID(strings.TrimSpace(token[:colon]))
		secretRaw := strings.TrimSpace(token[colon+1:])
		if kid == "" {
			return nil, fmt.Errorf("auth: MTRACE_AUTH_SIGNING_KEYS entry %d has empty kid", i)
		}
		if secretRaw == "" {
			return nil, fmt.Errorf("auth: MTRACE_AUTH_SIGNING_KEYS entry %d (kid %q) has empty secret", i, string(kid))
		}
		if _, dup := seen[kid]; dup {
			return nil, fmt.Errorf("auth: MTRACE_AUTH_SIGNING_KEYS contains duplicate kid %q", string(kid))
		}
		decoded, err := base64DecodeURLSafe(secretRaw)
		if err != nil {
			return nil, fmt.Errorf("auth: MTRACE_AUTH_SIGNING_KEYS entry %d (kid %q) base64 decode: %w", i, string(kid), err)
		}
		out = append(out, newDomainSigningKey(kid, decoded, now))
		seen[kid] = struct{}{}
	}
	if len(out) == 0 {
		return nil, errors.New("auth: MTRACE_AUTH_SIGNING_KEYS is empty after trim")
	}
	return out, nil
}

// newDomainSigningKey baut den Domain-`SessionSigningKey` für einen
// frisch geladenen ENV-Key. NotBefore/RetiresAt folgen dem
// `0.12.0`-Pattern aus `main.go` (NotBefore = -1h, RetiresAt = +365d);
// die konkreten Werte sind nicht teil von RAK-78 und können später
// per ENV überschrieben werden.
func newDomainSigningKey(kid domain.SigningKeyID, secret []byte, now time.Time) domain.SessionSigningKey {
	return domain.SessionSigningKey{
		KID:       kid,
		Algorithm: domain.SigningKeyAlgorithmHS256,
		Secret:    secret,
		NotBefore: now.Add(-time.Hour),
		RetiresAt: now.Add(365 * 24 * time.Hour),
	}
}

// base64DecodeURLSafe akzeptiert sowohl `RawURLEncoding` (ohne
// Padding) als auch `URLEncoding` (mit `=`-Padding). Identisches
// Verhalten wie in `cmd/api/main.go#base64DecodeURLSafe`, hier
// dupliziert, damit der Parser ohne Caller-Hilfsfunktion auskommt.
func base64DecodeURLSafe(s string) ([]byte, error) {
	if decoded, err := base64.RawURLEncoding.DecodeString(s); err == nil {
		return decoded, nil
	}
	return base64.URLEncoding.DecodeString(s)
}
