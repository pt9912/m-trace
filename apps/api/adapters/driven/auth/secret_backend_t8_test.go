package auth_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/auth"
)

//  (R-20) — Tests für die neuen Vault-Auth-Methoden
// (AppRole, Kubernetes) und den KMS-Skelett-Adapter.

// multiPathVaultMock bedient zwei Pfade auf demselben httptest-Server:
// einen Login-Endpoint (POST mit JSON-Body) und den KV-Read-Endpoint
// (GET mit `X-Vault-Token`-Header). Test-Adapter macht erst Login,
// dann KV-Read mit dem zurückgegebenen client_token.
func multiPathVaultMock(t *testing.T, loginPath string, expectedLoginBody map[string]string, clientToken string, kvPath string, kvBody string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case loginPath:
			if r.Method != http.MethodPost {
				http.Error(w, "login: wrong method", http.StatusMethodNotAllowed)
				return
			}
			var body map[string]string
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "login: bad json", http.StatusBadRequest)
				return
			}
			for k, want := range expectedLoginBody {
				if body[k] != want {
					http.Error(w, "login: field "+k+" mismatch", http.StatusUnauthorized)
					return
				}
			}
			resp := map[string]any{
				"auth": map[string]any{
					"client_token":   clientToken,
					"lease_duration": 3600,
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
		case kvPath:
			if r.Method != http.MethodGet {
				http.Error(w, "kv: wrong method", http.StatusMethodNotAllowed)
				return
			}
			if r.Header.Get("X-Vault-Token") != clientToken {
				http.Error(w, "kv: wrong token", http.StatusUnauthorized)
				return
			}
			_, _ = w.Write([]byte(kvBody))
		default:
			http.Error(w, "unknown path "+r.URL.Path, http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func kvBody(t *testing.T) string {
	t.Helper()
	secret := base64.RawURLEncoding.EncodeToString([]byte("vault-secret-32-bytes-1234567890ab"))
	payload := map[string]any{
		"data": map[string]any{
			"data": map[string]string{
				"keys":       "kid_v1:" + secret,
				"active_kid": "kid_v1",
			},
		},
	}
	raw, _ := json.Marshal(payload)
	return string(raw)
}

// TestVault_AppRoleLogin_HappyPath (`0.12.6` T8): Adapter macht
// AppRole-Login, bekommt einen `client_token` und nutzt ihn für den
// KV-Read.
func TestVault_AppRoleLogin_HappyPath(t *testing.T) {
	t.Parallel()
	srv := multiPathVaultMock(t,
		"/v1/auth/approle/login",
		map[string]string{"role_id": "role-abc", "secret_id": "sec-xyz"},
		"client-tok-1",
		"/v1/secret/data/m-trace/signing",
		kvBody(t),
	)
	env := map[string]string{
		"MTRACE_AUTH_VAULT_ADDR":              srv.URL,
		"MTRACE_AUTH_VAULT_PATH":              "secret/data/m-trace/signing",
		"MTRACE_AUTH_VAULT_AUTH_METHOD":       "approle",
		"MTRACE_AUTH_VAULT_APPROLE_ROLE_ID":   "role-abc",
		"MTRACE_AUTH_VAULT_APPROLE_SECRET_ID": "sec-xyz",
	}
	b, err := auth.NewVaultSecretBackend(func(k string) string { return env[k] })
	if err != nil {
		t.Fatalf("NewVaultSecretBackend: %v", err)
	}
	keys, active, err := b.LoadSigningKeys(context.Background())
	if err != nil {
		t.Fatalf("LoadSigningKeys: %v", err)
	}
	if active != "kid_v1" || len(keys) != 1 {
		t.Errorf("got active=%s len=%d, want kid_v1/1", active, len(keys))
	}
}

// TestVault_AppRoleLogin_MissingSecret: Constructor lehnt
// AppRole-Pfad ohne Pflicht-ENV ab.
func TestVault_AppRoleLogin_MissingSecret(t *testing.T) {
	t.Parallel()
	env := map[string]string{
		"MTRACE_AUTH_VAULT_ADDR":            "http://x",
		"MTRACE_AUTH_VAULT_PATH":            "secret/data/x",
		"MTRACE_AUTH_VAULT_AUTH_METHOD":     "approle",
		"MTRACE_AUTH_VAULT_APPROLE_ROLE_ID": "r",
		// SECRET_ID fehlt
	}
	if _, err := auth.NewVaultSecretBackend(func(k string) string { return env[k] }); err == nil {
		t.Errorf("expected error on missing APPROLE_SECRET_ID")
	}
}

// TestVault_AppRoleLogin_WrongSecret_PropagatesError: Server lehnt
// Login mit 401 ab; Adapter propagiert den Error fail-closed.
func TestVault_AppRoleLogin_WrongSecret_PropagatesError(t *testing.T) {
	t.Parallel()
	srv := multiPathVaultMock(t,
		"/v1/auth/approle/login",
		map[string]string{"role_id": "role-abc", "secret_id": "sec-correct"},
		"client-tok-1",
		"/v1/secret/data/m-trace/signing",
		kvBody(t),
	)
	env := map[string]string{
		"MTRACE_AUTH_VAULT_ADDR":              srv.URL,
		"MTRACE_AUTH_VAULT_PATH":              "secret/data/m-trace/signing",
		"MTRACE_AUTH_VAULT_AUTH_METHOD":       "approle",
		"MTRACE_AUTH_VAULT_APPROLE_ROLE_ID":   "role-abc",
		"MTRACE_AUTH_VAULT_APPROLE_SECRET_ID": "sec-WRONG",
	}
	b, err := auth.NewVaultSecretBackend(func(k string) string { return env[k] })
	if err != nil {
		t.Fatalf("NewVaultSecretBackend: %v", err)
	}
	if _, _, err := b.LoadSigningKeys(context.Background()); err == nil {
		t.Errorf("expected login error, got nil")
	}
}

// TestVault_KubernetesLogin_HappyPath (`0.12.6` T8): Adapter liest
// JWT aus konfiguriertem File, macht K8s-Login, bekommt
// client_token und nutzt ihn für KV-Read.
func TestVault_KubernetesLogin_HappyPath(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	jwtPath := filepath.Join(dir, "sa-token")
	if err := os.WriteFile(jwtPath, []byte("ey.jwt.payload"), 0o600); err != nil {
		t.Fatalf("write jwt: %v", err)
	}
	srv := multiPathVaultMock(t,
		"/v1/auth/kubernetes/login",
		map[string]string{"role": "m-trace", "jwt": "ey.jwt.payload"},
		"client-tok-k8s",
		"/v1/secret/data/m-trace/signing",
		kvBody(t),
	)
	env := map[string]string{
		"MTRACE_AUTH_VAULT_ADDR":          srv.URL,
		"MTRACE_AUTH_VAULT_PATH":          "secret/data/m-trace/signing",
		"MTRACE_AUTH_VAULT_AUTH_METHOD":   "kubernetes",
		"MTRACE_AUTH_VAULT_K8S_ROLE":      "m-trace",
		"MTRACE_AUTH_VAULT_K8S_JWT_PATH":  jwtPath,
	}
	b, err := auth.NewVaultSecretBackend(func(k string) string { return env[k] })
	if err != nil {
		t.Fatalf("NewVaultSecretBackend: %v", err)
	}
	if _, _, err := b.LoadSigningKeys(context.Background()); err != nil {
		t.Errorf("LoadSigningKeys: %v", err)
	}
}

// TestVault_KubernetesLogin_JWTFileMissing: nicht-existenter Pfad
// liefert fail-closed-Error.
func TestVault_KubernetesLogin_JWTFileMissing(t *testing.T) {
	t.Parallel()
	env := map[string]string{
		"MTRACE_AUTH_VAULT_ADDR":         "http://x",
		"MTRACE_AUTH_VAULT_PATH":         "secret/data/x",
		"MTRACE_AUTH_VAULT_AUTH_METHOD":  "kubernetes",
		"MTRACE_AUTH_VAULT_K8S_ROLE":     "r",
		"MTRACE_AUTH_VAULT_K8S_JWT_PATH": "/tmp/nope-does-not-exist-mtrace-test",
	}
	b, err := auth.NewVaultSecretBackend(func(k string) string { return env[k] })
	if err != nil {
		t.Fatalf("NewVaultSecretBackend: %v", err)
	}
	if _, _, err := b.LoadSigningKeys(context.Background()); err == nil {
		t.Errorf("expected missing-JWT-file error")
	}
}

// TestVault_KubernetesLogin_EmptyJWT: leeres JWT-File wird abgelehnt.
func TestVault_KubernetesLogin_EmptyJWT(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	jwtPath := filepath.Join(dir, "sa-token")
	if err := os.WriteFile(jwtPath, []byte("   \n"), 0o600); err != nil {
		t.Fatalf("write jwt: %v", err)
	}
	env := map[string]string{
		"MTRACE_AUTH_VAULT_ADDR":         "http://x",
		"MTRACE_AUTH_VAULT_PATH":         "secret/data/x",
		"MTRACE_AUTH_VAULT_AUTH_METHOD":  "kubernetes",
		"MTRACE_AUTH_VAULT_K8S_ROLE":     "r",
		"MTRACE_AUTH_VAULT_K8S_JWT_PATH": jwtPath,
	}
	b, _ := auth.NewVaultSecretBackend(func(k string) string { return env[k] })
	if _, _, err := b.LoadSigningKeys(context.Background()); err == nil {
		t.Errorf("expected empty-JWT error")
	}
}

// TestVault_AppRoleLogin_MissingClientToken (`0.12.6` T8):
// Login-Endpoint antwortet 200 OK aber mit leerem `auth.client_token`
// — Adapter propagiert fail-closed-Error.
func TestVault_AppRoleLogin_MissingClientToken(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/auth/approle/login" {
			_, _ = w.Write([]byte(`{"auth":{"client_token":""}}`))
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(srv.Close)
	env := map[string]string{
		"MTRACE_AUTH_VAULT_ADDR":              srv.URL,
		"MTRACE_AUTH_VAULT_PATH":              "secret/data/x",
		"MTRACE_AUTH_VAULT_AUTH_METHOD":       "approle",
		"MTRACE_AUTH_VAULT_APPROLE_ROLE_ID":   "r",
		"MTRACE_AUTH_VAULT_APPROLE_SECRET_ID": "s",
	}
	b, err := auth.NewVaultSecretBackend(func(k string) string { return env[k] })
	if err != nil {
		t.Fatalf("NewVaultSecretBackend: %v", err)
	}
	if _, _, err := b.LoadSigningKeys(context.Background()); err == nil {
		t.Errorf("expected missing-client_token error")
	}
}

// TestVault_AppRoleLogin_MalformedJSON: Login-Endpoint liefert
// malformed JSON; Decode-Fehler propagiert.
func TestVault_AppRoleLogin_MalformedJSON(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/auth/approle/login" {
			_, _ = w.Write([]byte(`{"auth":not-json`))
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(srv.Close)
	env := map[string]string{
		"MTRACE_AUTH_VAULT_ADDR":              srv.URL,
		"MTRACE_AUTH_VAULT_PATH":              "secret/data/x",
		"MTRACE_AUTH_VAULT_AUTH_METHOD":       "approle",
		"MTRACE_AUTH_VAULT_APPROLE_ROLE_ID":   "r",
		"MTRACE_AUTH_VAULT_APPROLE_SECRET_ID": "s",
	}
	b, _ := auth.NewVaultSecretBackend(func(k string) string { return env[k] })
	if _, _, err := b.LoadSigningKeys(context.Background()); err == nil {
		t.Errorf("expected malformed-json error")
	}
}

// TestVault_UnsupportedAuthMethod: Constructor lehnt unbekannte
// Methoden ab.
func TestVault_UnsupportedAuthMethod(t *testing.T) {
	t.Parallel()
	env := map[string]string{
		"MTRACE_AUTH_VAULT_ADDR":        "http://x",
		"MTRACE_AUTH_VAULT_PATH":        "secret/data/x",
		"MTRACE_AUTH_VAULT_AUTH_METHOD": "aws-iam",
	}
	_, err := auth.NewVaultSecretBackend(func(k string) string { return env[k] })
	if err == nil {
		t.Errorf("expected error for unsupported auth method")
	}
}

// stubKMSDecrypter ist ein Mock, der einen vor-konfigurierten
// Plaintext zurückgibt.
type stubKMSDecrypter struct {
	plaintext []byte
	err       error
}

func (s stubKMSDecrypter) Decrypt(_ context.Context, _ []byte) ([]byte, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.plaintext, nil
}

// TestKMSSecretBackend_HappyPath (`0.12.6` T8): Stub-Decrypter
// liefert den Plaintext im `ParseSigningKeysEnv`-Format; Adapter
// reicht ihn an die gemeinsame Validierungs-Logik weiter.
func TestKMSSecretBackend_HappyPath(t *testing.T) {
	t.Parallel()
	secret := base64.RawURLEncoding.EncodeToString([]byte("kms-secret-32-bytes-1234567890abcd"))
	plaintext := []byte("kid_a:" + secret)
	env := map[string]string{
		"MTRACE_AUTH_KMS_ENCRYPTED_KEYS": base64.StdEncoding.EncodeToString([]byte("opaque-ciphertext")),
		"MTRACE_AUTH_KMS_ACTIVE_KID":     "kid_a",
	}
	b, err := auth.NewKMSSecretBackend(func(k string) string { return env[k] }, stubKMSDecrypter{plaintext: plaintext})
	if err != nil {
		t.Fatalf("NewKMSSecretBackend: %v", err)
	}
	keys, active, err := b.LoadSigningKeys(context.Background())
	if err != nil {
		t.Fatalf("LoadSigningKeys: %v", err)
	}
	if active != "kid_a" || len(keys) != 1 {
		t.Errorf("got %s/%d, want kid_a/1", active, len(keys))
	}
}

// TestKMSSecretBackend_DecryptError: Decrypter-Fehler propagiert.
func TestKMSSecretBackend_DecryptError(t *testing.T) {
	t.Parallel()
	env := map[string]string{
		"MTRACE_AUTH_KMS_ENCRYPTED_KEYS": base64.StdEncoding.EncodeToString([]byte("opaque-ciphertext")),
		"MTRACE_AUTH_KMS_ACTIVE_KID":     "kid_a",
	}
	b, err := auth.NewKMSSecretBackend(func(k string) string { return env[k] },
		stubKMSDecrypter{err: errors.New("kms unreachable")})
	if err != nil {
		t.Fatalf("NewKMSSecretBackend: %v", err)
	}
	if _, _, err := b.LoadSigningKeys(context.Background()); err == nil {
		t.Errorf("expected decrypt error to propagate")
	}
}

// TestKMSSecretBackend_LabPassThrough (`0.12.6` T8): der opt-in
// `LabPassThroughKMSDecrypter` reicht den Ciphertext als Plaintext
// durch — der Test verifiziert den End-to-End-Pfad fürs Lab-Smoke.
func TestKMSSecretBackend_LabPassThrough(t *testing.T) {
	t.Parallel()
	secret := base64.RawURLEncoding.EncodeToString([]byte("kms-lab-secret-32-bytes-1234567890ab"))
	// "Ciphertext" == Plaintext, weil LabPassThroughKMSDecrypter
	// nichts decryptet.
	plaintext := "kid_lab:" + secret
	env := map[string]string{
		"MTRACE_AUTH_KMS_ENCRYPTED_KEYS": base64.StdEncoding.EncodeToString([]byte(plaintext)),
		"MTRACE_AUTH_KMS_ACTIVE_KID":     "kid_lab",
	}
	b, err := auth.NewKMSSecretBackend(func(k string) string { return env[k] }, auth.LabPassThroughKMSDecrypter{})
	if err != nil {
		t.Fatalf("NewKMSSecretBackend: %v", err)
	}
	keys, active, err := b.LoadSigningKeys(context.Background())
	if err != nil {
		t.Fatalf("LoadSigningKeys: %v", err)
	}
	if active != "kid_lab" || len(keys) != 1 {
		t.Errorf("got %s/%d, want kid_lab/1", active, len(keys))
	}
}

// TestKMSSecretBackend_MissingConfig: Constructor lehnt fehlende
// Pflicht-ENV ab.
func TestKMSSecretBackend_MissingConfig(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		env  map[string]string
	}{
		{"no decrypter, no env", map[string]string{}},
		{"missing active_kid", map[string]string{
			"MTRACE_AUTH_KMS_ENCRYPTED_KEYS": "Zm9v",
		}},
		{"missing ciphertext", map[string]string{
			"MTRACE_AUTH_KMS_ACTIVE_KID": "kid_a",
		}},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := auth.NewKMSSecretBackend(func(k string) string { return tc.env[k] }, stubKMSDecrypter{plaintext: []byte("kid_a:zzzz")})
			if err == nil {
				t.Errorf("expected error for missing config")
			}
		})
	}
}

// TestKMSSecretBackend_NilDecrypter: nil-Decrypter wird abgelehnt.
func TestKMSSecretBackend_NilDecrypter(t *testing.T) {
	t.Parallel()
	env := map[string]string{
		"MTRACE_AUTH_KMS_ENCRYPTED_KEYS": "Zm9v",
		"MTRACE_AUTH_KMS_ACTIVE_KID":     "kid_a",
	}
	if _, err := auth.NewKMSSecretBackend(func(k string) string { return env[k] }, nil); err == nil {
		t.Errorf("expected nil-decrypter rejection")
	}
}

// TestKMSSecretBackend_EncryptedKeysFromPath: Path-Variante
// (MTRACE_AUTH_KMS_ENCRYPTED_KEYS_PATH) anstelle der base64-ENV.
func TestKMSSecretBackend_EncryptedKeysFromPath(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cipherPath := filepath.Join(dir, "kms.bin")
	if err := os.WriteFile(cipherPath, []byte("opaque-bytes"), 0o600); err != nil {
		t.Fatalf("write cipher: %v", err)
	}
	secret := base64.RawURLEncoding.EncodeToString([]byte("from-path-32-bytes-secret-1234567"))
	env := map[string]string{
		"MTRACE_AUTH_KMS_ENCRYPTED_KEYS_PATH": cipherPath,
		"MTRACE_AUTH_KMS_ACTIVE_KID":          "kid_path",
	}
	b, err := auth.NewKMSSecretBackend(func(k string) string { return env[k] },
		stubKMSDecrypter{plaintext: []byte("kid_path:" + secret)})
	if err != nil {
		t.Fatalf("NewKMSSecretBackend: %v", err)
	}
	if _, _, err := b.LoadSigningKeys(context.Background()); err != nil {
		t.Errorf("LoadSigningKeys: %v", err)
	}
}

// _ marker für strings-Import (manche Test-Pfade brauchen ihn nicht
// explizit, aber `multiPathVaultMock` o. ä. kann ihn nutzen).
var _ = strings.TrimSpace
