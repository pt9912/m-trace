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

// VaultAuthMethod selektiert die Login-Methode für den Vault-Adapter
// (`0.12.5` Tranche 3 + `0.12.6` Tranche 8 / R-20). Drei Werte:
// `token` (statischer X-Vault-Token), `approle` (zwei-Phasen-Login
// mit role_id+secret_id) und `kubernetes` (Pod-ServiceAccount-Token
// + Vault-K8s-Login). Wire-Vertrag und Konfigurations-ENV siehe
// Doku in `docs/user/auth.md` §5.5.
type VaultAuthMethod string

const (
	// VaultAuthMethodToken nutzt einen statischen Vault-Token aus
	// `MTRACE_AUTH_VAULT_TOKEN` direkt als `X-Vault-Token`-Header.
	VaultAuthMethodToken VaultAuthMethod = "token"
	// VaultAuthMethodAppRole macht einen AppRole-Login-Roundtrip
	// (`/v1/auth/approle/login`) mit `role_id` + `secret_id`.
	VaultAuthMethodAppRole VaultAuthMethod = "approle"
	// VaultAuthMethodKubernetes liest den Pod-ServiceAccount-Token
	// und ruft `/v1/auth/kubernetes/login` mit Role+JWT auf.
	VaultAuthMethodKubernetes VaultAuthMethod = "kubernetes"
)

// VaultSecretBackend ist der konfigurierte Vault-KV-v2-Adapter mit
// drei Auth-Methoden. Eigener minimaler `net/http`-Client gegen die
// Vault-REST-API (`/v1/<mount>/data/<path>` für KV-Read;
// `/v1/auth/<method>/login` für AppRole/K8s), bewusst ohne
// `hashicorp/vault/api`-Dependency. Boot-Time-Load, fail-closed bei
// Backend-Outage. Wire-Format für das Vault-Secret entspricht dem
// `MTRACE_AUTH_SIGNING_KEYS` / `_ACTIVE_KID`-Schema; Operator-Doku
// siehe `docs/user/auth.md` §5.5.
type VaultSecretBackend struct {
	HTTPClient *http.Client
	Now        func() time.Time

	address    string
	mount      string
	path       string
	keysKey    string
	activeKey  string
	authMethod VaultAuthMethod

	// Auth-Method-spezifische Felder. Genau eines davon ist gesetzt.
	token          string // VaultAuthMethodToken
	approleRoleID  string // VaultAuthMethodAppRole
	approleSecret  string // VaultAuthMethodAppRole
	k8sRole        string // VaultAuthMethodKubernetes
	k8sJWTPath     string // VaultAuthMethodKubernetes
	approleMount   string // VaultAuthMethodAppRole (default "approle")
	k8sMount       string // VaultAuthMethodKubernetes (default "kubernetes")
}

// NewVaultSecretBackend baut den Adapter aus den `MTRACE_AUTH_VAULT_*`-
// ENV-Variablen. Die Funktion validiert nur die Pflichtfelder; ob
// Vault erreichbar ist und der Pfad existiert, prüft `LoadSigningKeys`
// beim ersten Aufruf.
//
// Gemeinsame Pflichtfelder:
//   - `MTRACE_AUTH_VAULT_ADDR` (z. B. `http://127.0.0.1:8200`)
//   - `MTRACE_AUTH_VAULT_PATH` (KV-v2-Pfad, z. B. `secret/data/m-trace/signing`)
//
// Auth-Method-Selektor `MTRACE_AUTH_VAULT_AUTH_METHOD`:
//   - `token` (Default): Pflicht-ENV `MTRACE_AUTH_VAULT_TOKEN`.
//   - `approle`: Pflicht-ENV `MTRACE_AUTH_VAULT_APPROLE_ROLE_ID`
//     und `MTRACE_AUTH_VAULT_APPROLE_SECRET_ID`; optional
//     `MTRACE_AUTH_VAULT_APPROLE_MOUNT` (Default `approle`).
//   - `kubernetes`: Pflicht-ENV `MTRACE_AUTH_VAULT_K8S_ROLE`;
//     optional `MTRACE_AUTH_VAULT_K8S_JWT_PATH` (Default
//     `/var/run/secrets/kubernetes.io/serviceaccount/token`),
//     `MTRACE_AUTH_VAULT_K8S_MOUNT` (Default `kubernetes`).
//
// Optionale gemeinsame Felder:
//   - `MTRACE_AUTH_VAULT_KEYS_FIELD` (Default `keys`)
//   - `MTRACE_AUTH_VAULT_ACTIVE_KID_FIELD` (Default `active_kid`)
func NewVaultSecretBackend(lookup func(key string) string) (*VaultSecretBackend, error) {
	if lookup == nil {
		lookup = os.Getenv
	}
	addr := strings.TrimSpace(lookup("MTRACE_AUTH_VAULT_ADDR"))
	path := strings.TrimSpace(lookup("MTRACE_AUTH_VAULT_PATH"))
	if addr == "" || path == "" {
		return nil, errors.New("auth vault backend: MTRACE_AUTH_VAULT_ADDR and MTRACE_AUTH_VAULT_PATH are required")
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

	method := VaultAuthMethod(strings.ToLower(strings.TrimSpace(lookup("MTRACE_AUTH_VAULT_AUTH_METHOD"))))
	if method == "" {
		method = VaultAuthMethodToken
	}

	b := &VaultSecretBackend{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		Now:        time.Now,
		address:    strings.TrimRight(addr, "/"),
		mount:      mount,
		path:       secretPath,
		keysKey:    keysField,
		activeKey:  activeField,
		authMethod: method,
	}

	if err := applyVaultAuthMethod(b, method, lookup); err != nil {
		return nil, err
	}
	return b, nil
}

// applyVaultAuthMethod liest die methoden-spezifischen ENV-Variablen
// und befüllt das passende Feldset auf `b`. Trennt den Boot-Validator
// vom Constructor, damit `NewVaultSecretBackend` unter dem
// gocognit-Limit bleibt.
func applyVaultAuthMethod(b *VaultSecretBackend, method VaultAuthMethod, lookup func(string) string) error {
	switch method {
	case VaultAuthMethodToken:
		token := strings.TrimSpace(lookup("MTRACE_AUTH_VAULT_TOKEN"))
		if token == "" {
			return errors.New("auth vault backend: MTRACE_AUTH_VAULT_TOKEN required for auth_method=token")
		}
		b.token = token
		return nil
	case VaultAuthMethodAppRole:
		roleID := strings.TrimSpace(lookup("MTRACE_AUTH_VAULT_APPROLE_ROLE_ID"))
		secretID := strings.TrimSpace(lookup("MTRACE_AUTH_VAULT_APPROLE_SECRET_ID"))
		if roleID == "" || secretID == "" {
			return errors.New("auth vault backend: MTRACE_AUTH_VAULT_APPROLE_ROLE_ID and _SECRET_ID required for auth_method=approle")
		}
		b.approleRoleID = roleID
		b.approleSecret = secretID
		b.approleMount = firstNonEmpty(lookup("MTRACE_AUTH_VAULT_APPROLE_MOUNT"), "approle")
		return nil
	case VaultAuthMethodKubernetes:
		role := strings.TrimSpace(lookup("MTRACE_AUTH_VAULT_K8S_ROLE"))
		if role == "" {
			return errors.New("auth vault backend: MTRACE_AUTH_VAULT_K8S_ROLE required for auth_method=kubernetes")
		}
		b.k8sRole = role
		b.k8sJWTPath = firstNonEmpty(lookup("MTRACE_AUTH_VAULT_K8S_JWT_PATH"), "/var/run/secrets/kubernetes.io/serviceaccount/token")
		b.k8sMount = firstNonEmpty(lookup("MTRACE_AUTH_VAULT_K8S_MOUNT"), "kubernetes")
		return nil
	default:
		return fmt.Errorf("auth vault backend: MTRACE_AUTH_VAULT_AUTH_METHOD=%q not supported (valid: token|approle|kubernetes)", method)
	}
}

// firstNonEmpty trimmt `raw` und fällt auf `fallback` zurück, wenn es
// danach leer ist.
func firstNonEmpty(raw, fallback string) string {
	if v := strings.TrimSpace(raw); v != "" {
		return v
	}
	return fallback
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
// gemeinsamen `ParseSigningKeysEnv`-Pfad weiter. Vor dem KV-Read
// führt der Adapter ggf. einen Login-Roundtrip aus (AppRole / K8s).
func (b *VaultSecretBackend) LoadSigningKeys(ctx context.Context) ([]domain.SessionSigningKey, domain.SigningKeyID, error) {
	if b == nil {
		return nil, "", errors.New("auth vault backend: nil receiver")
	}
	now := time.Now
	if b.Now != nil {
		now = b.Now
	}
	clientToken, err := b.resolveClientToken(ctx)
	if err != nil {
		return nil, "", err
	}
	endpoint := fmt.Sprintf("%s/v1/%s/data/%s", b.address, b.mount, b.path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, "", fmt.Errorf("auth vault backend: build request: %w", err)
	}
	req.Header.Set("X-Vault-Token", clientToken)
	req.Header.Set("Accept", "application/json")

	resp, err := b.HTTPClient.Do(req)
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

// resolveClientToken liefert den `X-Vault-Token`-Wert für den KV-Read,
// abhängig von der konfigurierten Auth-Method:
//   - token: direkter Wert aus `MTRACE_AUTH_VAULT_TOKEN`.
//   - approle: Login-Roundtrip an `/v1/auth/<mount>/login`.
//   - kubernetes: JWT aus Pod-File lesen, dann Login an
//     `/v1/auth/<mount>/login`.
func (b *VaultSecretBackend) resolveClientToken(ctx context.Context) (string, error) {
	// Constructor garantiert eine der drei Methoden — keine Default-
	// Branch nötig.
	switch b.authMethod {
	case VaultAuthMethodAppRole:
		return b.loginViaAppRole(ctx)
	case VaultAuthMethodKubernetes:
		return b.loginViaKubernetes(ctx)
	default:
		// VaultAuthMethodToken (Constructor-validiert).
		return b.token, nil
	}
}

// loginViaAppRole macht den AppRole-Login-Roundtrip.
//
// Wire-Format Request:
//
//	{"role_id":"<role_id>","secret_id":"<secret_id>"}
//
// Response:
//
//	{"auth":{"client_token":"<token>","lease_duration":<seconds>,...}}
func (b *VaultSecretBackend) loginViaAppRole(ctx context.Context) (string, error) {
	endpoint := fmt.Sprintf("%s/v1/auth/%s/login", b.address, b.approleMount)
	// json.Marshal über map[string]string mit primitiven Strings kann
	// nicht fehlschlagen — kein Error-Check nötig.
	body, _ := json.Marshal(map[string]string{
		"role_id":   b.approleRoleID,
		"secret_id": b.approleSecret,
	})
	return b.loginRoundtrip(ctx, endpoint, body, "approle")
}

// loginViaKubernetes liest das ServiceAccount-Token aus dem
// konfigurierten Pfad und ruft den K8s-Login-Endpoint auf.
//
// Wire-Format Request:
//
//	{"role":"<role>","jwt":"<pod-sa-token>"}
//
// Response: identisch zu AppRole (`auth.client_token`).
func (b *VaultSecretBackend) loginViaKubernetes(ctx context.Context) (string, error) {
	jwt, err := os.ReadFile(b.k8sJWTPath)
	if err != nil {
		return "", fmt.Errorf("auth vault backend: read kubernetes JWT %q: %w", b.k8sJWTPath, err)
	}
	jwtTrim := strings.TrimSpace(string(jwt))
	if jwtTrim == "" {
		return "", fmt.Errorf("auth vault backend: kubernetes JWT %q is empty", b.k8sJWTPath)
	}
	endpoint := fmt.Sprintf("%s/v1/auth/%s/login", b.address, b.k8sMount)
	body, _ := json.Marshal(map[string]string{
		"role": b.k8sRole,
		"jwt":  jwtTrim,
	})
	return b.loginRoundtrip(ctx, endpoint, body, "kubernetes")
}

// loginRoundtrip POST'd `body` an `endpoint`, parsed
// `auth.client_token` aus der Response.
func (b *VaultSecretBackend) loginRoundtrip(ctx context.Context, endpoint string, body []byte, method string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(body)))
	if err != nil {
		return "", fmt.Errorf("auth vault backend: build %s login request: %w", method, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := b.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("auth vault backend: %s login: %w", method, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("auth vault backend: %s login returned %d: %s",
			method, resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	var payload struct {
		Auth struct {
			ClientToken string `json:"client_token"`
		} `json:"auth"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&payload); err != nil {
		return "", fmt.Errorf("auth vault backend: decode %s login: %w", method, err)
	}
	if strings.TrimSpace(payload.Auth.ClientToken) == "" {
		return "", fmt.Errorf("auth vault backend: %s login response missing auth.client_token", method)
	}
	return payload.Auth.ClientToken, nil
}
