package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// VaultSecretBackend ist das externe Backend-Adapter-**Skelett** aus
// `0.12.5` Tranche 3 (RAK-79, R-20). Holt die Signing-Key-Felder
// `keys` und `active_kid` aus einem Vault KV-v2-Pfad und reicht sie
// an `ParseSigningKeysEnv` weiter — damit nutzen ENV- und Vault-
// Backend dieselbe Validierungs-Logik (Duplikate, leere KIDs,
// Base64, ACTIVE_KID-in-list).
//
// Skelett-Charakter (siehe Plan-0.12.5 §0.3 / §13.15 RAK-79):
//
//   - Eigener minimaler HTTP-Client gegen die Vault-REST-API
//     (`/v1/<mount>/data/<path>`). Bewusst ohne `hashicorp/vault/api`-
//     Dependency, damit die go.mod schlank bleibt; eine produktive
//     Anbindung kann den Skelett-Adapter durch einen
//     `hashicorp/vault/api`-Adapter ersetzen, ohne den Port zu
//     ändern.
//   - Authentication ausschließlich über Token (`X-Vault-Token`).
//     AppRole, AWS-IAM-Auth und Kubernetes-Service-Account-Auth
//     bleiben Folge-Item für die produktive Anbindung.
//   - Boot-Time-Load; kein periodischer Refresh und kein TTL-
//     Caching. Operator-Restart ist der einzige Schlüsselwechsel-
//     Pfad.
//   - Fail-closed bei Backend-Outage oder ungültiger Konfiguration:
//     ein API-Boot, der `vault` ausgewählt aber den Vault-Server
//     nicht erreichen kann, hard-failt — kein stiller Fallback.
//
// Lab-Pfad: ein `vault dev`-Server (Bind 127.0.0.1:8200, Default-
// Token aus stdout) reicht. Wire-Format für das Vault-Secret:
//
//	{
//	  "data": {
//	    "data": {
//	      "keys": "kid_a:<base64>,kid_b:<base64>",
//	      "active_kid": "kid_a"
//	    }
//	  }
//	}
//
// Geliefert wird also dasselbe Schema wie bei `MTRACE_AUTH_SIGNING_KEYS`,
// nur über Vault zugestellt.
type VaultSecretBackend struct {
	HTTPClient *http.Client
	Now        func() time.Time

	address   string
	token     string
	mount     string
	path      string
	keysKey   string
	activeKey string
}

// NewVaultSecretBackend baut den Adapter aus den `MTRACE_AUTH_VAULT_*`-
// ENV-Variablen. Die Funktion validiert nur die Pflichtfelder; ob
// Vault erreichbar ist und der Pfad existiert, prüft `LoadSigningKeys`
// beim ersten Aufruf.
//
// Pflichtfelder:
//   - `MTRACE_AUTH_VAULT_ADDR` (z. B. `http://127.0.0.1:8200`)
//   - `MTRACE_AUTH_VAULT_TOKEN` (z. B. der `vault dev`-Token)
//   - `MTRACE_AUTH_VAULT_PATH` (KV-v2-Pfad, z. B. `secret/data/m-trace/signing`)
//
// Optional (Defaults siehe Code):
//   - `MTRACE_AUTH_VAULT_KEYS_FIELD` (Default `keys`)
//   - `MTRACE_AUTH_VAULT_ACTIVE_KID_FIELD` (Default `active_kid`)
func NewVaultSecretBackend(lookup func(key string) string) (*VaultSecretBackend, error) {
	if lookup == nil {
		lookup = os.Getenv
	}
	addr := strings.TrimSpace(lookup("MTRACE_AUTH_VAULT_ADDR"))
	token := strings.TrimSpace(lookup("MTRACE_AUTH_VAULT_TOKEN"))
	path := strings.TrimSpace(lookup("MTRACE_AUTH_VAULT_PATH"))
	if addr == "" || token == "" || path == "" {
		return nil, errors.New("auth vault backend: MTRACE_AUTH_VAULT_ADDR, MTRACE_AUTH_VAULT_TOKEN and MTRACE_AUTH_VAULT_PATH are all required")
	}
	if _, err := url.Parse(addr); err != nil {
		return nil, fmt.Errorf("auth vault backend: MTRACE_AUTH_VAULT_ADDR invalid: %w", err)
	}
	mount, secretPath, err := splitVaultKVv2Path(path)
	if err != nil {
		return nil, err
	}
	keysField := strings.TrimSpace(lookup("MTRACE_AUTH_VAULT_KEYS_FIELD"))
	if keysField == "" {
		keysField = "keys"
	}
	activeField := strings.TrimSpace(lookup("MTRACE_AUTH_VAULT_ACTIVE_KID_FIELD"))
	if activeField == "" {
		activeField = "active_kid"
	}
	return &VaultSecretBackend{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		Now:        time.Now,
		address:    strings.TrimRight(addr, "/"),
		token:      token,
		mount:      mount,
		path:       secretPath,
		keysKey:    keysField,
		activeKey:  activeField,
	}, nil
}

// splitVaultKVv2Path zerlegt einen `mount/data/<path>`-Vault-Pfad in
// `mount` und `path`. Der KV-v2-Read-Endpoint lautet
// `/v1/<mount>/data/<path>`, daher erwarten wir den `data/`-Marker
// im Operator-Wert (analog zur Vault-CLI-Konvention).
func splitVaultKVv2Path(p string) (mount, secret string, err error) {
	trimmed := strings.Trim(p, "/")
	const marker = "/data/"
	idx := strings.Index(trimmed, marker)
	if idx <= 0 || idx+len(marker) >= len(trimmed) {
		return "", "", fmt.Errorf("auth vault backend: MTRACE_AUTH_VAULT_PATH %q must be a KV-v2 path of form mount/data/secret-path", p)
	}
	return trimmed[:idx], trimmed[idx+len(marker):], nil
}

// Compile-time check.
var _ driven.AuthSecretBackend = (*VaultSecretBackend)(nil)

// LoadSigningKeys holt das Vault-KV-v2-Secret, extrahiert die zwei
// Pflichtfelder (`keys`, `active_kid`) und reicht sie an den
// gemeinsamen `ParseSigningKeysEnv`-Pfad weiter.
func (b *VaultSecretBackend) LoadSigningKeys(ctx context.Context) ([]domain.SessionSigningKey, domain.SigningKeyID, error) {
	if b == nil {
		return nil, "", errors.New("auth vault backend: nil receiver")
	}
	now := time.Now
	if b.Now != nil {
		now = b.Now
	}
	endpoint := fmt.Sprintf("%s/v1/%s/data/%s", b.address, b.mount, b.path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, "", fmt.Errorf("auth vault backend: build request: %w", err)
	}
	req.Header.Set("X-Vault-Token", b.token)
	req.Header.Set("Accept", "application/json")

	resp, err := b.client().Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("auth vault backend: GET %s: %w", endpoint, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, "", fmt.Errorf("auth vault backend: GET %s returned %d: %s", endpoint, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var payload struct {
		Data struct {
			Data map[string]string `json:"data"`
		} `json:"data"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&payload); err != nil {
		return nil, "", fmt.Errorf("auth vault backend: decode response: %w", err)
	}

	keysRaw := strings.TrimSpace(payload.Data.Data[b.keysKey])
	activeRaw := strings.TrimSpace(payload.Data.Data[b.activeKey])
	if keysRaw == "" {
		return nil, "", fmt.Errorf("auth vault backend: %s missing field %q at %s", endpoint, b.keysKey, b.path)
	}
	if activeRaw == "" {
		return nil, "", fmt.Errorf("auth vault backend: %s missing field %q at %s", endpoint, b.activeKey, b.path)
	}

	keys, activeKID, noKeys, parseErr := ParseSigningKeysEnv(keysRaw, activeRaw, "", "", now().UTC())
	if parseErr != nil {
		return nil, "", fmt.Errorf("auth vault backend: parse %s: %w", b.path, parseErr)
	}
	if noKeys {
		// Defensive: ParseSigningKeysEnv liefert noKeys nur, wenn weder
		// Multi- noch Single-Key gesetzt sind — bei der Vault-Aufruf-Form
		// haben wir keys+active gerade validiert, sollte also nicht
		// passieren.
		return nil, "", fmt.Errorf("auth vault backend: %s yielded no signing keys after parse", b.path)
	}
	return keys, activeKID, nil
}

func (b *VaultSecretBackend) client() *http.Client {
	if b.HTTPClient != nil {
		return b.HTTPClient
	}
	return &http.Client{Timeout: 10 * time.Second}
}
