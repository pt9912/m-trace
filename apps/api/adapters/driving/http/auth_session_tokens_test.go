package http_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/auth"
	"github.com/pt9912/m-trace/apps/api/adapters/driven/metrics"
	apihttp "github.com/pt9912/m-trace/apps/api/adapters/driving/http"
	"github.com/pt9912/m-trace/apps/api/hexagon/application"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// `0.12.0` Tranche 2 / RAK-72: HTTP-Vertrag von
// `POST /api/auth/session-tokens`. Tests fahren den vollen Stack:
// HMAC-Signer + In-Memory-Limiter + In-Memory-Policy-Resolver +
// Random-ID-Generator (alles `apps/api/adapters/driven/auth`), gegen
// die Static-Project-Resolver-Konfiguration mit `demo`-Token. Damit
// sind sowohl Issuance-Happy-Path als auch alle Fehlerpräzedenz-
// Stufen aus §3.9 verifiziert.

const authIssuanceSigningSecret = "test-signing-secret-please-do-not-use-in-production"

func newAuthSessionTestServer(t *testing.T, opts ...authServerOption) *httptest.Server {
	t.Helper()
	cfg := authServerConfig{
		globalCapacity: 50,
		globalRefill:   50,
		projectCap:     20,
		projectRefill:  20,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	resolver := auth.NewStaticProjectResolver(map[string]auth.ProjectConfig{
		"demo":  {Token: "demo-token", AllowedOrigins: []string{"http://localhost:5173"}},
		"other": {Token: "other-token", AllowedOrigins: []string{"http://other.example"}},
	})
	baseProjects := map[string]domain.Project{
		"demo":  {ID: "demo", Token: "demo-token", AllowedOrigins: []string{"http://localhost:5173"}},
		"other": {ID: "other", Token: "other-token", AllowedOrigins: []string{"http://other.example"}},
	}
	signingKey := domain.SessionSigningKey{
		KID:       "test-kid",
		Algorithm: domain.SigningKeyAlgorithmHS256,
		Secret:    []byte(authIssuanceSigningSecret),
		NotBefore: time.Now().Add(-time.Hour).UTC(),
		RetiresAt: time.Now().Add(time.Hour).UTC(),
	}
	keyResolver, err := auth.NewMultiKeySigningResolver("test-kid", signingKey)
	if err != nil {
		t.Fatalf("static signing key resolver: %v", err)
	}
	signer := auth.NewHMACSessionTokenSigner(keyResolver)
	limiter := auth.NewInMemoryIssuanceRateLimiter(cfg.globalCapacity, cfg.globalRefill, cfg.projectCap, cfg.projectRefill)
	policies, err := auth.NewInMemoryProjectPolicyResolver(nil, baseProjects)
	if err != nil {
		t.Fatalf("policies: %v", err)
	}
	ids := auth.NewRandomTokenIDGenerator()
	svc := application.NewIssueSessionTokenService(policies, limiter, signer, ids)
	publisher := metrics.NewPrometheusPublisher()
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	router := apihttp.NewRouter(nil, nil, nil, resolver, resolver, publisher.Handler(), nil, nil, nil, nil, nil, svc, nil, nil, nil, false, nil, logger)
	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)
	return srv
}

type authServerConfig struct {
	globalCapacity int
	globalRefill   float64
	projectCap     int
	projectRefill  float64
}

type authServerOption func(*authServerConfig)

func withProjectIssuanceQuota(capacity int, refill float64) authServerOption {
	return func(c *authServerConfig) {
		c.projectCap = capacity
		c.projectRefill = refill
	}
}

func postAuthSessionToken(t *testing.T, srv *httptest.Server, body string, headers map[string]string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, srv.URL+"/api/auth/session-tokens", bytes.NewReader([]byte(body)))
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	return resp
}

func decodeAuthBody(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal: %v\nraw=%s", err, string(raw))
	}
	return out
}

func TestAuthSessionTokens_HappyPath(t *testing.T) {
	t.Parallel()
	srv := newAuthSessionTestServer(t)
	body := `{"audience":"playback-events","ttl_seconds":120,"session_id":"sess-a","origin":"http://localhost:5173"}`
	resp := postAuthSessionToken(t, srv, body, map[string]string{"X-MTrace-Token": "demo-token"})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status: want 201, got %d", resp.StatusCode)
	}
	out := decodeAuthBody(t, resp)
	tok, ok := out["session_token"].(map[string]any)
	if !ok {
		t.Fatalf("session_token missing or wrong shape: %v", out)
	}
	value, _ := tok["value"].(string)
	if !strings.HasPrefix(value, "mtr_st_") {
		t.Errorf("value missing prefix: %q", value)
	}
	if parts := strings.Split(strings.TrimPrefix(value, "mtr_st_"), "."); len(parts) != 3 {
		t.Errorf("value wire format: want 3 dot-separated segments, got %d", len(parts))
	}
	if tokID, _ := tok["token_id"].(string); !strings.HasPrefix(tokID, "st_") {
		t.Errorf("token_id format: want st_*, got %q", tokID)
	}
	if pid, _ := tok["project_id"].(string); pid != "demo" {
		t.Errorf("project_id: want demo, got %q", pid)
	}
	if aud, _ := tok["audience"].(string); aud != "playback-events" {
		t.Errorf("audience: want playback-events, got %q", aud)
	}
	if sess, _ := tok["session_id"].(string); sess != "sess-a" {
		t.Errorf("session_id: want sess-a, got %q", sess)
	}
}

func TestAuthSessionTokens_MissingMTraceToken(t *testing.T) {
	t.Parallel()
	srv := newAuthSessionTestServer(t)
	resp := postAuthSessionToken(t, srv, `{"audience":"playback-events"}`, nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status: want 401, got %d", resp.StatusCode)
	}
	out := decodeAuthBody(t, resp)
	if code, _ := out["code"].(string); code != "auth_token_missing" {
		t.Errorf("code: want auth_token_missing, got %q", code)
	}
}

func TestAuthSessionTokens_InvalidMTraceToken(t *testing.T) {
	t.Parallel()
	srv := newAuthSessionTestServer(t)
	resp := postAuthSessionToken(t, srv, `{"audience":"playback-events"}`, map[string]string{"X-MTrace-Token": "bogus"})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status: want 401, got %d", resp.StatusCode)
	}
	out := decodeAuthBody(t, resp)
	if code, _ := out["code"].(string); code != "auth_token_invalid" {
		t.Errorf("code: want auth_token_invalid, got %q", code)
	}
}

func TestAuthSessionTokens_ProjectMismatch(t *testing.T) {
	t.Parallel()
	srv := newAuthSessionTestServer(t)
	body := `{"audience":"playback-events","project_id":"other"}`
	resp := postAuthSessionToken(t, srv, body, map[string]string{"X-MTrace-Token": "demo-token"})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status: want 401, got %d", resp.StatusCode)
	}
	out := decodeAuthBody(t, resp)
	if code, _ := out["code"].(string); code != "auth_project_mismatch" {
		t.Errorf("code: want auth_project_mismatch, got %q", code)
	}
}

func TestAuthSessionTokens_AudienceDenied(t *testing.T) {
	t.Parallel()
	srv := newAuthSessionTestServer(t)
	body := `{"audience":"admin"}`
	resp := postAuthSessionToken(t, srv, body, map[string]string{"X-MTrace-Token": "demo-token"})
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status: want 403, got %d", resp.StatusCode)
	}
	out := decodeAuthBody(t, resp)
	if code, _ := out["code"].(string); code != "auth_session_scope_denied" {
		t.Errorf("code: want auth_session_scope_denied, got %q", code)
	}
}

func TestAuthSessionTokens_TTLTooLarge(t *testing.T) {
	t.Parallel()
	srv := newAuthSessionTestServer(t)
	body := `{"audience":"playback-events","ttl_seconds":10000}`
	resp := postAuthSessionToken(t, srv, body, map[string]string{"X-MTrace-Token": "demo-token"})
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status: want 422, got %d", resp.StatusCode)
	}
	out := decodeAuthBody(t, resp)
	if code, _ := out["code"].(string); code != "auth_token_ttl_too_large" {
		t.Errorf("code: want auth_token_ttl_too_large, got %q", code)
	}
}

func TestAuthSessionTokens_TTLNegative(t *testing.T) {
	t.Parallel()
	srv := newAuthSessionTestServer(t)
	body := `{"audience":"playback-events","ttl_seconds":-1}`
	resp := postAuthSessionToken(t, srv, body, map[string]string{"X-MTrace-Token": "demo-token"})
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status: want 422, got %d", resp.StatusCode)
	}
}

func TestAuthSessionTokens_RateLimited(t *testing.T) {
	t.Parallel()
	srv := newAuthSessionTestServer(t, withProjectIssuanceQuota(2, 0))
	body := `{"audience":"playback-events","ttl_seconds":60}`
	for i := 0; i < 2; i++ {
		resp := postAuthSessionToken(t, srv, body, map[string]string{"X-MTrace-Token": "demo-token"})
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("issuance %d: want 201, got %d", i, resp.StatusCode)
		}
		_ = resp.Body.Close()
	}
	resp := postAuthSessionToken(t, srv, body, map[string]string{"X-MTrace-Token": "demo-token"})
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("status: want 429, got %d", resp.StatusCode)
	}
	out := decodeAuthBody(t, resp)
	if code, _ := out["code"].(string); code != "auth_issuance_rate_limited" {
		t.Errorf("code: want auth_issuance_rate_limited, got %q", code)
	}
}

func TestAuthSessionTokens_InvalidContentType(t *testing.T) {
	t.Parallel()
	srv := newAuthSessionTestServer(t)
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/auth/session-tokens", strings.NewReader(`x`))
	req.Header.Set("X-MTrace-Token", "demo-token")
	req.Header.Set("Content-Type", "text/plain")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if resp.StatusCode != http.StatusUnsupportedMediaType {
		t.Fatalf("status: want 415, got %d", resp.StatusCode)
	}
}

func TestAuthSessionTokens_BodyTooLarge(t *testing.T) {
	t.Parallel()
	srv := newAuthSessionTestServer(t)
	huge := strings.Repeat("x", 8*1024)
	body := `{"audience":"playback-events","ttl_seconds":60,"origin":"` + huge + `"}`
	resp := postAuthSessionToken(t, srv, body, map[string]string{"X-MTrace-Token": "demo-token"})
	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("status: want 413, got %d", resp.StatusCode)
	}
}

func TestAuthSessionTokens_InvalidJSON(t *testing.T) {
	t.Parallel()
	srv := newAuthSessionTestServer(t)
	resp := postAuthSessionToken(t, srv, `{not-json`, map[string]string{"X-MTrace-Token": "demo-token"})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: want 400, got %d", resp.StatusCode)
	}
	out := decodeAuthBody(t, resp)
	if code, _ := out["code"].(string); code != "invalid_json" {
		t.Errorf("code: want invalid_json, got %q", code)
	}
}

func TestAuthSessionTokens_TTLDefaultsTo900OnZero(t *testing.T) {
	t.Parallel()
	srv := newAuthSessionTestServer(t)
	body := `{"audience":"playback-events"}`
	resp := postAuthSessionToken(t, srv, body, map[string]string{"X-MTrace-Token": "demo-token"})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status: want 201, got %d", resp.StatusCode)
	}
	out := decodeAuthBody(t, resp)
	tok := out["session_token"].(map[string]any)
	expRaw, _ := tok["expires_at"].(string)
	exp, err := time.Parse(time.RFC3339, expRaw)
	if err != nil {
		t.Fatalf("parse expires_at %q: %v", expRaw, err)
	}
	delta := time.Until(exp).Round(time.Second)
	if delta < 880*time.Second || delta > 920*time.Second {
		t.Errorf("expires_at delta: want ~900s, got %v", delta)
	}
}

func TestAuthSessionTokens_HandlerRejectsGET(t *testing.T) {
	t.Parallel()
	srv := newAuthSessionTestServer(t)
	resp, err := http.Get(srv.URL + "/api/auth/session-tokens")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	// Router selektiert nach Methode → unmatched Method liefert 405
	// vom Mux. Kein expliziter Body-Check — wir verifizieren nur, dass
	// der GET-Pfad nicht versehentlich Issuance triggert.
	if resp.StatusCode == http.StatusCreated {
		t.Errorf("GET must not trigger issuance, got 201")
	}
}

func TestAuthSessionTokens_InvalidJSONBeforeAuth(t *testing.T) {
	t.Parallel()
	// §3.9-Validierungsreihenfolge: JSON-Parse läuft VOR Auth-Resolve.
	// Ohne Token + kaputtes JSON → 400 invalid_json (nicht 401).
	srv := newAuthSessionTestServer(t)
	resp := postAuthSessionToken(t, srv, `{not-json`, nil) // kein X-MTrace-Token
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: want 400 (invalid_json before auth), got %d", resp.StatusCode)
	}
	out := decodeAuthBody(t, resp)
	if code, _ := out["code"].(string); code != "invalid_json" {
		t.Errorf("code: want invalid_json, got %q", code)
	}
}

func TestAuthSessionTokens_RoundTripVerify(t *testing.T) {
	t.Parallel()
	srv := newAuthSessionTestServer(t)
	body := `{"audience":"playback-events","ttl_seconds":60}`
	resp := postAuthSessionToken(t, srv, body, map[string]string{"X-MTrace-Token": "demo-token"})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status: want 201, got %d", resp.StatusCode)
	}
	out := decodeAuthBody(t, resp)
	tok := out["session_token"].(map[string]any)
	value, _ := tok["value"].(string)
	expectedTokenID, _ := tok["token_id"].(string)

	// Dieselbe Konfiguration wie der Test-Server: HS256-Signer mit
	// "test-kid"/`authIssuanceSigningSecret`. Verify außerhalb des
	// Servers, damit der Round-Trip explizit gepinnt ist.
	signingKey := domain.SessionSigningKey{
		KID:       "test-kid",
		Algorithm: domain.SigningKeyAlgorithmHS256,
		Secret:    []byte(authIssuanceSigningSecret),
		NotBefore: time.Now().Add(-time.Hour).UTC(),
		RetiresAt: time.Now().Add(time.Hour).UTC(),
	}
	resolver, err := auth.NewMultiKeySigningResolver("test-kid", signingKey)
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	signer := auth.NewHMACSessionTokenSigner(resolver)
	claims, err := signer.Verify(value)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if claims.JTI != expectedTokenID {
		t.Errorf("jti: want %q, got %q", expectedTokenID, claims.JTI)
	}
	if claims.Sub != "demo" {
		t.Errorf("sub: want demo, got %q", claims.Sub)
	}
	if claims.Aud != domain.SessionTokenAudiencePlaybackEvents {
		t.Errorf("aud: want playback-events, got %q", claims.Aud)
	}
}
