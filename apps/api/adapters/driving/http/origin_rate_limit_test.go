package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/auth"
	"github.com/pt9912/m-trace/apps/api/adapters/driven/metrics"
	apihttp "github.com/pt9912/m-trace/apps/api/adapters/driving/http"
	"github.com/pt9912/m-trace/apps/api/hexagon/application"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// newOriginLimitedAuthServer baut den Auth-Issuance-Endpoint mit
// einem Origin-Limiter davor (R-22). Tests
// kontrollieren die Limiter-Bucket-Größe und können den XFF-Trust
// per Flag aktivieren.
func newOriginLimitedAuthServer(t *testing.T, originLimiter driven.OriginRateLimiter, trustXFF bool) *httptest.Server {
	t.Helper()
	resolver := auth.NewStaticProjectResolver(map[string]auth.ProjectConfig{
		"demo": {Token: "demo-token", AllowedOrigins: []string{"http://localhost:5173"}},
	})
	baseProjects := map[string]domain.Project{
		"demo": {ID: "demo", Token: "demo-token", AllowedOrigins: []string{"http://localhost:5173"}},
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
		t.Fatalf("key resolver: %v", err)
	}
	signer := auth.NewHMACSessionTokenSigner(keyResolver)
	issuance := auth.NewInMemoryIssuanceRateLimiter(50, 50, 20, 20)
	policies, err := auth.NewInMemoryProjectPolicyResolver(nil, baseProjects)
	if err != nil {
		t.Fatalf("policies: %v", err)
	}
	svc := application.NewIssueSessionTokenService(policies, issuance, signer, auth.NewRandomTokenIDGenerator())
	publisher := metrics.NewPrometheusPublisher()
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	router := apihttp.NewRouter(nil, nil, nil, resolver, resolver, publisher.Handler(), nil, nil, nil, nil, nil, svc, nil, nil, originLimiter, trustXFF, nil, logger)
	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)
	return srv
}

// postSessionToken ist ein minimaler Wrapper für den Issuance-Call.
func postSessionToken(t *testing.T, srv *httptest.Server) *http.Response {
	t.Helper()
	body := `{"audience":"playback-events"}`
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, srv.URL+"/api/auth/session-tokens", bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-MTrace-Token", "demo-token")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	return resp
}

// TestOriginRateLimit_Burst (R-22): bei
// capacity=2 + refill=0 sind zwei Aufrufe in Folge erlaubt; der
// dritte aus derselben Quelle bekommt `429 origin_rate_limited`.
func TestOriginRateLimit_Burst(t *testing.T) {
	t.Parallel()
	// refill=0 → Bucket füllt sich nie wieder; deterministic 429 nach 2 calls.
	limiter := auth.NewInMemoryOriginRateLimiter(2, 0)
	srv := newOriginLimitedAuthServer(t, limiter, false)

	for i := 1; i <= 2; i++ {
		resp := postSessionToken(t, srv)
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("call %d: status=%d, want 201", i, resp.StatusCode)
		}
	}
	resp := postSessionToken(t, srv)
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusTooManyRequests {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("3rd call: status=%d, want 429: body=%s", resp.StatusCode, body)
	}
	var bodyMap map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&bodyMap); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if bodyMap["error"] != "origin_rate_limited" {
		t.Errorf("error = %q, want origin_rate_limited", bodyMap["error"])
	}
}

// TestOriginRateLimit_NoOpWhenDisabled (R-22):
// `nil`-Limiter (Disabled-Pfad) lässt alle Aufrufe durch.
func TestOriginRateLimit_NoOpWhenDisabled(t *testing.T) {
	t.Parallel()
	srv := newOriginLimitedAuthServer(t, nil, false)
	for i := 0; i < 5; i++ {
		resp := postSessionToken(t, srv)
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("call %d: status=%d, want 201 (limiter disabled)", i, resp.StatusCode)
		}
	}
}

// TestOriginRateLimit_XFFTrust (R-22): mit
// `trustXFF=true` nutzt der Limiter das letzte XFF-Element als Key;
// zwei verschiedene XFF-Header haben getrennte Buckets.
func TestOriginRateLimit_XFFTrust(t *testing.T) {
	t.Parallel()
	// capacity=1, refill=0 → genau 1 erfolgreicher Call pro Key.
	limiter := auth.NewInMemoryOriginRateLimiter(1, 0)
	srv := newOriginLimitedAuthServer(t, limiter, true)

	mkReq := func(xff string) *http.Response {
		t.Helper()
		body := `{"audience":"playback-events"}`
		req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, srv.URL+"/api/auth/session-tokens", bytes.NewBufferString(body))
		if err != nil {
			t.Fatalf("new request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-MTrace-Token", "demo-token")
		req.Header.Set("X-Forwarded-For", xff)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("do: %v", err)
		}
		return resp
	}

	// Erster Call mit XFF=1.1.1.1 → ok.
	r1 := mkReq("1.1.1.1")
	_ = r1.Body.Close()
	if r1.StatusCode != http.StatusCreated {
		t.Fatalf("1.1.1.1 #1: status=%d, want 201", r1.StatusCode)
	}
	// Zweiter Call mit gleichem XFF → 429 (Bucket leer).
	r2 := mkReq("1.1.1.1")
	_ = r2.Body.Close()
	if r2.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("1.1.1.1 #2: status=%d, want 429", r2.StatusCode)
	}
	// Dritter Call mit anderem XFF → ok (anderer Bucket).
	r3 := mkReq("2.2.2.2")
	_ = r3.Body.Close()
	if r3.StatusCode != http.StatusCreated {
		t.Fatalf("2.2.2.2: status=%d, want 201 (different bucket)", r3.StatusCode)
	}
}
