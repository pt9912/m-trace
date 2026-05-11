package auth_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/auth"
)

func TestNewEnvSecretBackend_DefaultsToOSGetenv(t *testing.T) {
	t.Parallel()
	b := auth.NewEnvSecretBackend()
	if b == nil {
		t.Fatal("NewEnvSecretBackend returned nil")
	}
	if b.LookupFn == nil {
		t.Error("LookupFn must default to os.Getenv")
	}
	if b.Now == nil {
		t.Error("Now must default to time.Now")
	}
	// Ohne ENV-Variablen → ErrNoSecretConfigured.
	_, _, err := b.LoadSigningKeys(context.Background())
	if !errors.Is(err, auth.ErrNoSecretConfigured) {
		t.Errorf("default lookup with no env should return ErrNoSecretConfigured, got %v", err)
	}
}

func TestEnvSecretBackend_MultiKeyHappyPath(t *testing.T) {
	t.Parallel()
	secretA := base64.RawURLEncoding.EncodeToString([]byte("secret-a-32-bytes-1234567890abcd"))
	secretB := base64.RawURLEncoding.EncodeToString([]byte("secret-b-32-bytes-1234567890abcd"))
	env := map[string]string{
		"MTRACE_AUTH_SIGNING_KEYS":       "kid_a:" + secretA + ",kid_b:" + secretB,
		"MTRACE_AUTH_SIGNING_ACTIVE_KID": "kid_a",
	}
	b := &auth.EnvSecretBackend{
		LookupFn: func(k string) string { return env[k] },
		Now:      func() time.Time { return time.Date(2026, 5, 11, 12, 0, 0, 0, time.UTC) },
	}
	keys, active, err := b.LoadSigningKeys(context.Background())
	if err != nil {
		t.Fatalf("LoadSigningKeys err: %v", err)
	}
	if active != "kid_a" {
		t.Errorf("active: want kid_a, got %s", active)
	}
	if len(keys) != 2 {
		t.Fatalf("want 2 keys, got %d", len(keys))
	}
}

func TestEnvSecretBackend_SingleKeyBackwardsCompat(t *testing.T) {
	t.Parallel()
	secret := base64.RawURLEncoding.EncodeToString([]byte("single-key-secret-32-bytes-1234"))
	env := map[string]string{
		"MTRACE_AUTH_SIGNING_KEY": secret,
		"MTRACE_AUTH_SIGNING_KID": "legacy",
	}
	b := &auth.EnvSecretBackend{LookupFn: func(k string) string { return env[k] }}
	keys, active, err := b.LoadSigningKeys(context.Background())
	if err != nil {
		t.Fatalf("LoadSigningKeys err: %v", err)
	}
	if active != "legacy" || len(keys) != 1 {
		t.Errorf("want legacy/1, got %s/%d", active, len(keys))
	}
}

func TestEnvSecretBackend_NoSecretConfigured(t *testing.T) {
	t.Parallel()
	b := &auth.EnvSecretBackend{LookupFn: func(string) string { return "" }}
	_, _, err := b.LoadSigningKeys(context.Background())
	if !errors.Is(err, auth.ErrNoSecretConfigured) {
		t.Errorf("expected ErrNoSecretConfigured, got %v", err)
	}
}

func TestEnvSecretBackend_PropagatesParseError(t *testing.T) {
	t.Parallel()
	env := map[string]string{
		"MTRACE_AUTH_SIGNING_KEYS":       "kid_a:!not!valid!base64!",
		"MTRACE_AUTH_SIGNING_ACTIVE_KID": "kid_a",
	}
	b := &auth.EnvSecretBackend{LookupFn: func(k string) string { return env[k] }}
	_, _, err := b.LoadSigningKeys(context.Background())
	if err == nil || errors.Is(err, auth.ErrNoSecretConfigured) {
		t.Errorf("expected parse error, got %v", err)
	}
}

// vaultMock baut einen httptest.Server, der die KV-v2-Antwort für
// einen festen Pfad liefert. Erlaubt das Spiegeln verschiedener
// Vault-Antworten (401, 404, 500, malformed) ohne echtes Vault.
func vaultMock(t *testing.T, expectedPath, expectedToken string, status int, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != expectedPath {
			http.Error(w, "wrong path: "+r.URL.Path, http.StatusNotFound)
			return
		}
		if r.Header.Get("X-Vault-Token") != expectedToken {
			http.Error(w, "wrong token", http.StatusUnauthorized)
			return
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestVaultSecretBackend_HappyPath(t *testing.T) {
	t.Parallel()
	secretA := base64.RawURLEncoding.EncodeToString([]byte("vault-secret-a-32-bytes-1234567"))
	secretB := base64.RawURLEncoding.EncodeToString([]byte("vault-secret-b-32-bytes-1234567"))
	bodyMap := map[string]any{
		"data": map[string]any{
			"data": map[string]string{
				"keys":       "kid_v1:" + secretA + ",kid_v2:" + secretB,
				"active_kid": "kid_v2",
			},
		},
	}
	bodyBytes, _ := json.Marshal(bodyMap)
	srv := vaultMock(t, "/v1/secret/data/m-trace/signing", "tok-dev", http.StatusOK, string(bodyBytes))

	env := map[string]string{
		"MTRACE_AUTH_VAULT_ADDR":  srv.URL,
		"MTRACE_AUTH_VAULT_TOKEN": "tok-dev",
		"MTRACE_AUTH_VAULT_PATH":  "secret/data/m-trace/signing",
	}
	b, err := auth.NewVaultSecretBackend(func(k string) string { return env[k] })
	if err != nil {
		t.Fatalf("NewVaultSecretBackend: %v", err)
	}
	keys, active, err := b.LoadSigningKeys(context.Background())
	if err != nil {
		t.Fatalf("LoadSigningKeys: %v", err)
	}
	if active != "kid_v2" || len(keys) != 2 {
		t.Errorf("want kid_v2/2, got %s/%d", active, len(keys))
	}
}

func TestVaultSecretBackend_MissingConfig(t *testing.T) {
	t.Parallel()
	cases := []map[string]string{
		{}, // alles leer
		{"MTRACE_AUTH_VAULT_ADDR": "http://x"},
		{"MTRACE_AUTH_VAULT_ADDR": "http://x", "MTRACE_AUTH_VAULT_TOKEN": "t"},
	}
	for i, env := range cases {
		_, err := auth.NewVaultSecretBackend(func(k string) string { return env[k] })
		if err == nil {
			t.Errorf("case %d: expected error on missing config, got nil", i)
		}
	}
}

func TestVaultSecretBackend_PathFormatRequiresDataMarker(t *testing.T) {
	t.Parallel()
	env := map[string]string{
		"MTRACE_AUTH_VAULT_ADDR":  "http://localhost:8200",
		"MTRACE_AUTH_VAULT_TOKEN": "t",
		"MTRACE_AUTH_VAULT_PATH":  "secret/m-trace/signing", // missing data/
	}
	_, err := auth.NewVaultSecretBackend(func(k string) string { return env[k] })
	if err == nil || !strings.Contains(err.Error(), "KV-v2") {
		t.Errorf("expected KV-v2 path error, got %v", err)
	}
}

func TestVaultSecretBackend_Unauthorized(t *testing.T) {
	t.Parallel()
	srv := vaultMock(t, "/v1/secret/data/x", "right-token", http.StatusOK, "{}")
	env := map[string]string{
		"MTRACE_AUTH_VAULT_ADDR":  srv.URL,
		"MTRACE_AUTH_VAULT_TOKEN": "wrong-token",
		"MTRACE_AUTH_VAULT_PATH":  "secret/data/x",
	}
	b, _ := auth.NewVaultSecretBackend(func(k string) string { return env[k] })
	_, _, err := b.LoadSigningKeys(context.Background())
	if err == nil || !strings.Contains(err.Error(), "401") {
		t.Errorf("expected 401 error, got %v", err)
	}
}

func TestVaultSecretBackend_NotFound(t *testing.T) {
	t.Parallel()
	srv := vaultMock(t, "/v1/secret/data/right", "t", http.StatusOK, "{}")
	env := map[string]string{
		"MTRACE_AUTH_VAULT_ADDR":  srv.URL,
		"MTRACE_AUTH_VAULT_TOKEN": "t",
		"MTRACE_AUTH_VAULT_PATH":  "secret/data/missing",
	}
	b, _ := auth.NewVaultSecretBackend(func(k string) string { return env[k] })
	_, _, err := b.LoadSigningKeys(context.Background())
	if err == nil || !strings.Contains(err.Error(), "404") {
		t.Errorf("expected 404 error, got %v", err)
	}
}

func TestVaultSecretBackend_MalformedJSON(t *testing.T) {
	t.Parallel()
	srv := vaultMock(t, "/v1/secret/data/x", "t", http.StatusOK, "not-json-{{")
	env := map[string]string{
		"MTRACE_AUTH_VAULT_ADDR":  srv.URL,
		"MTRACE_AUTH_VAULT_TOKEN": "t",
		"MTRACE_AUTH_VAULT_PATH":  "secret/data/x",
	}
	b, _ := auth.NewVaultSecretBackend(func(k string) string { return env[k] })
	_, _, err := b.LoadSigningKeys(context.Background())
	if err == nil || !strings.Contains(err.Error(), "decode response") {
		t.Errorf("expected decode error, got %v", err)
	}
}

func TestVaultSecretBackend_MissingKeyField(t *testing.T) {
	t.Parallel()
	body := `{"data":{"data":{"active_kid":"kid_a"}}}` // no "keys"
	srv := vaultMock(t, "/v1/secret/data/x", "t", http.StatusOK, body)
	env := map[string]string{
		"MTRACE_AUTH_VAULT_ADDR":  srv.URL,
		"MTRACE_AUTH_VAULT_TOKEN": "t",
		"MTRACE_AUTH_VAULT_PATH":  "secret/data/x",
	}
	b, _ := auth.NewVaultSecretBackend(func(k string) string { return env[k] })
	_, _, err := b.LoadSigningKeys(context.Background())
	if err == nil || !strings.Contains(err.Error(), `"keys"`) {
		t.Errorf("expected missing-keys-field error, got %v", err)
	}
}

func TestVaultSecretBackend_CustomFieldNames(t *testing.T) {
	t.Parallel()
	secret := base64.RawURLEncoding.EncodeToString([]byte("custom-secret-32-bytes-12345678"))
	body := `{"data":{"data":{"signing_ring":"kid_x:` + secret + `","kid_marker":"kid_x"}}}`
	srv := vaultMock(t, "/v1/secret/data/x", "t", http.StatusOK, body)
	env := map[string]string{
		"MTRACE_AUTH_VAULT_ADDR":             srv.URL,
		"MTRACE_AUTH_VAULT_TOKEN":            "t",
		"MTRACE_AUTH_VAULT_PATH":             "secret/data/x",
		"MTRACE_AUTH_VAULT_KEYS_FIELD":       "signing_ring",
		"MTRACE_AUTH_VAULT_ACTIVE_KID_FIELD": "kid_marker",
	}
	b, err := auth.NewVaultSecretBackend(func(k string) string { return env[k] })
	if err != nil {
		t.Fatalf("NewVaultSecretBackend: %v", err)
	}
	keys, active, err := b.LoadSigningKeys(context.Background())
	if err != nil {
		t.Fatalf("LoadSigningKeys: %v", err)
	}
	if active != "kid_x" || len(keys) != 1 {
		t.Errorf("want kid_x/1, got %s/%d", active, len(keys))
	}
}
