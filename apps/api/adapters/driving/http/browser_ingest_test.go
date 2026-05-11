package http_test

import (
	"bytes"
	"context"
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
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driving"
)

// drivingCreateStreamResultOK liefert ein gültiges Stub-Resultat
// für den `CreateStream`-Pfad. Mindestfelder, damit der HTTP-Handler
// `201 Created` antwortet.
func drivingCreateStreamResultOK() driving.CreateStreamResult {
	return driving.CreateStreamResult{
		Stream: domain.IngestStream{
			ID:        "ing_browser_test",
			ProjectID: browserIngestProject,
			Protocol:  domain.IngestProtocolSRT,
			CreatedAt: time.Now().UTC(),
		},
		Material: domain.StreamKeyMaterial{
			Value:       "plaintext-once",
			Hash:        "h",
			Fingerprint: "fp",
			CreatedAt:   time.Now().UTC(),
		},
	}
}

const (
	browserIngestProject = "ing-tenant"
	browserIngestToken   = "demo-token"
	browserIngestOrigin  = "https://app.example.com"
)

func browserIngestResolver(t *testing.T) *auth.StaticProjectResolver {
	t.Helper()
	return auth.NewStaticProjectResolver(map[string]auth.ProjectConfig{
		browserIngestProject: {
			Token:          browserIngestToken,
			AllowedOrigins: []string{browserIngestOrigin, "http://localhost:5173"},
		},
	})
}

func browserIngestPolicies(t *testing.T, policy domain.BrowserIngestPolicy) *auth.InMemoryProjectPolicyResolver {
	t.Helper()
	p, err := auth.NewInMemoryProjectPolicyResolver(
		map[string]domain.ProjectPolicy{
			browserIngestProject: {
				ProjectID:      browserIngestProject,
				AllowedOrigins: []string{browserIngestOrigin},
				BrowserIngest:  policy,
			},
		},
		nil,
	)
	if err != nil {
		t.Fatalf("NewInMemoryProjectPolicyResolver: %v", err)
	}
	return p
}

func newBrowserIngestRouter(t *testing.T, policies apihttp.BrowserIngestPolicies, stub stubIngestControl) *httptest.Server {
	t.Helper()
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	publisher := metrics.NewPrometheusPublisher()
	resolver := browserIngestResolver(t)
	router := apihttp.NewRouter(
		nil, nil, nil, resolver, resolver,
		publisher.Handler(), nil, nil, nil, nil, &stub, nil, nil, policies, nil, false, nil, logger,
	)
	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)
	return srv
}

func newRequest(t *testing.T, method, url string, headers map[string]string, body string) *http.Request {
	t.Helper()
	var rdr io.Reader
	if body != "" {
		rdr = bytes.NewReader([]byte(body))
	}
	req, _ := http.NewRequestWithContext(context.Background(), method, url, rdr)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return req
}

// TestBrowserIngest_Preflight_PolicyDisabled_ReturnsMinimal204
// verifiziert, dass ohne aktivierte Policy der RAK-74-Scope-Cut weiter
// gilt: Preflight gibt 204 ohne Allow-Origin-Header zurück.
func TestBrowserIngest_Preflight_PolicyDisabled_ReturnsMinimal204(t *testing.T) {
	t.Parallel()
	policies := browserIngestPolicies(t, domain.BrowserIngestPolicy{Enabled: false})
	srv := newBrowserIngestRouter(t, policies, stubIngestControl{})

	req := newRequest(t, http.MethodOptions, srv.URL+"/api/ingest/streams", map[string]string{
		"Origin":                        browserIngestOrigin,
		"Access-Control-Request-Method": http.MethodPost,
	}, "")
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("status: want 204, got %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("policy disabled must NOT echo Allow-Origin, got %q", got)
	}
}

// TestBrowserIngest_Preflight_PolicyEnabled_OriginAllowed_ReturnsCORS
// hebt den RAK-74-Scope-Cut auf: bei aktivierter Policy und Match
// liefert der Preflight die echten CORS-Header.
func TestBrowserIngest_Preflight_PolicyEnabled_OriginAllowed_ReturnsCORS(t *testing.T) {
	t.Parallel()
	policies := browserIngestPolicies(t, domain.BrowserIngestPolicy{
		Enabled:       true,
		CORSAllowlist: []string{browserIngestOrigin},
	})
	srv := newBrowserIngestRouter(t, policies, stubIngestControl{})

	req := newRequest(t, http.MethodOptions, srv.URL+"/api/ingest/streams", map[string]string{
		"Origin":                        browserIngestOrigin,
		"Access-Control-Request-Method": http.MethodPost,
	}, "")
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("status: want 204, got %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != browserIngestOrigin {
		t.Errorf("Allow-Origin: want %q, got %q", browserIngestOrigin, got)
	}
	if got := resp.Header.Get("Access-Control-Allow-Methods"); !strings.Contains(got, "POST") {
		t.Errorf("Allow-Methods must include POST, got %q", got)
	}
	if got := resp.Header.Get("Access-Control-Allow-Headers"); !strings.Contains(got, "X-MTrace-CSRF") {
		t.Errorf("Allow-Headers must include X-MTrace-CSRF, got %q", got)
	}
}

// TestBrowserIngest_Preflight_PolicyEnabled_OriginNotInAllowlist
// prüft, dass ein nicht-allowed Origin bei aktivierter Policy
// trotzdem 204 ohne Allow-Origin liefert (kein Enum-Leak).
func TestBrowserIngest_Preflight_PolicyEnabled_OriginNotInAllowlist(t *testing.T) {
	t.Parallel()
	policies := browserIngestPolicies(t, domain.BrowserIngestPolicy{
		Enabled:       true,
		CORSAllowlist: []string{browserIngestOrigin},
	})
	srv := newBrowserIngestRouter(t, policies, stubIngestControl{})

	req := newRequest(t, http.MethodOptions, srv.URL+"/api/ingest/streams", map[string]string{
		"Origin":                        "https://evil.example.com",
		"Access-Control-Request-Method": http.MethodPost,
	}, "")
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("status: want 204, got %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("non-allowlist origin must NOT echo Allow-Origin, got %q", got)
	}
}

// TestBrowserIngest_POST_PolicyDisabled_PassesThrough beweist, dass
// die POST-Middleware den Pfad unverändert lässt, wenn keine Policy
// aktiv ist (Backwards-Compat zum 0.11.0-Verhalten).
func TestBrowserIngest_POST_PolicyDisabled_PassesThrough(t *testing.T) {
	t.Parallel()
	policies := browserIngestPolicies(t, domain.BrowserIngestPolicy{Enabled: false})
	stub := stubIngestControl{
		createResult: drivingCreateStreamResultOK(),
	}
	srv := newBrowserIngestRouter(t, policies, stub)

	req := newRequest(t, http.MethodPost, srv.URL+"/api/ingest/streams", map[string]string{
		"X-MTrace-Token": browserIngestToken,
		"Content-Type":   "application/json",
		// Kein Origin gesetzt → Operator-Pfad.
	}, `{"display_name":"x","protocol":"srt","endpoint_id":"ep","target_id":"tg"}`)
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("policy disabled must reach handler (got status %d)", resp.StatusCode)
	}
}

// TestBrowserIngest_POST_PolicyEnabled_NoOriginNoPin_PassesThrough
// deckt den Edge-Case ab: aktivierte Policy ohne Origin-Pin, kein
// Origin-Header (Operator-/CLI-Pfad) → Request läuft durch, ohne
// 403. Origin-Allowlist greift nur für gesetzte Origins.
func TestBrowserIngest_POST_PolicyEnabled_NoOriginNoPin_PassesThrough(t *testing.T) {
	t.Parallel()
	policies := browserIngestPolicies(t, domain.BrowserIngestPolicy{
		Enabled:       true,
		CORSAllowlist: []string{browserIngestOrigin},
		// OriginPin leer, CSRFRequired false
	})
	stub := stubIngestControl{createResult: drivingCreateStreamResultOK()}
	srv := newBrowserIngestRouter(t, policies, stub)

	req := newRequest(t, http.MethodPost, srv.URL+"/api/ingest/streams", map[string]string{
		"X-MTrace-Token": browserIngestToken,
		"Content-Type":   "application/json",
		// Bewusst kein Origin
	}, `{"display_name":"x","protocol":"srt","endpoint_id":"ep","target_id":"tg"}`)
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("policy enabled, no origin, no pin: want 201 passthrough, got %d", resp.StatusCode)
	}
}

// TestBrowserIngest_POST_PolicyEnabled_OriginPinMismatch_403
// zeigt, dass der Origin-Pin als Defense-in-Depth greift.
func TestBrowserIngest_POST_PolicyEnabled_OriginPinMismatch_403(t *testing.T) {
	t.Parallel()
	policies := browserIngestPolicies(t, domain.BrowserIngestPolicy{
		Enabled:       true,
		CORSAllowlist: []string{browserIngestOrigin, "https://other.example.com"},
		OriginPin:     browserIngestOrigin,
	})
	srv := newBrowserIngestRouter(t, policies, stubIngestControl{})

	req := newRequest(t, http.MethodPost, srv.URL+"/api/ingest/streams", map[string]string{
		"X-MTrace-Token": browserIngestToken,
		"Content-Type":   "application/json",
		// In Allowlist enthalten, aber nicht der Pin.
		"Origin": "https://other.example.com",
	}, `{"display_name":"x","protocol":"srt","endpoint_id":"ep","target_id":"tg"}`)
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("origin-pin mismatch: want 403, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "ingest_browser_origin_pin_mismatch") {
		t.Errorf("body must mention origin_pin_mismatch, got %s", body)
	}
}

// TestBrowserIngest_POST_PolicyEnabled_CSRFRequiredMissing_403
// verifiziert die CSRF-Pflicht.
func TestBrowserIngest_POST_PolicyEnabled_CSRFRequiredMissing_403(t *testing.T) {
	t.Parallel()
	policies := browserIngestPolicies(t, domain.BrowserIngestPolicy{
		Enabled:       true,
		CORSAllowlist: []string{browserIngestOrigin},
		CSRFRequired:  true,
	})
	srv := newBrowserIngestRouter(t, policies, stubIngestControl{})

	req := newRequest(t, http.MethodPost, srv.URL+"/api/ingest/streams", map[string]string{
		"X-MTrace-Token": browserIngestToken,
		"Content-Type":   "application/json",
		"Origin":         browserIngestOrigin,
		// Keine X-MTrace-CSRF.
	}, `{"display_name":"x","protocol":"srt","endpoint_id":"ep","target_id":"tg"}`)
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("csrf missing: want 403, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "ingest_browser_csrf_missing") {
		t.Errorf("body must mention csrf_missing, got %s", body)
	}
}

// TestBrowserIngest_POST_PolicyEnabled_AllChecksPass durchläuft den
// kompletten Happy-Path: Origin in Allowlist + matched Pin + CSRF
// gesetzt → Handler wird aufgerufen.
func TestBrowserIngest_POST_PolicyEnabled_AllChecksPass(t *testing.T) {
	t.Parallel()
	policies := browserIngestPolicies(t, domain.BrowserIngestPolicy{
		Enabled:       true,
		CORSAllowlist: []string{browserIngestOrigin},
		OriginPin:     browserIngestOrigin,
		CSRFRequired:  true,
	})
	stub := stubIngestControl{createResult: drivingCreateStreamResultOK()}
	srv := newBrowserIngestRouter(t, policies, stub)

	req := newRequest(t, http.MethodPost, srv.URL+"/api/ingest/streams", map[string]string{
		"X-MTrace-Token": browserIngestToken,
		"Content-Type":   "application/json",
		"Origin":         browserIngestOrigin,
		"X-MTrace-CSRF":  "csrf-token-value",
	}, `{"display_name":"x","protocol":"srt","endpoint_id":"ep","target_id":"tg"}`)
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("all checks pass: want 201, got %d", resp.StatusCode)
	}
}
