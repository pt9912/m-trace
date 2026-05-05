package http_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/auth"
	"github.com/pt9912/m-trace/apps/api/adapters/driven/metrics"
	apihttp "github.com/pt9912/m-trace/apps/api/adapters/driving/http"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// plan-0.4.0 §4.5 E2 — End-to-End-Tests für die `{analysis,
// session_link}`-Hülle, die endpoint-spezifische Auth (drei Klassen)
// und den OPTIONS-Preflight für /api/analyze.

// linkAwareAnalysisInbound sieht den vollen Use-Case-Vertrag: er
// reicht den `req.ProjectID`/`CorrelationID`/`SessionID` durch und
// liefert deterministisch je nach Eingabe einen passenden
// SessionLink-Status.
type linkAwareAnalysisInbound struct {
	gotReq      domain.StreamAnalysisRequest
	resultLink  domain.SessionLink
	analysis    domain.StreamAnalysisResult
	called      int
}

func (s *linkAwareAnalysisInbound) AnalyzeManifest(_ context.Context, req domain.StreamAnalysisRequest) (domain.AnalyzeManifestResult, error) {
	s.called++
	s.gotReq = req
	link := s.resultLink
	if link.Status == "" {
		// Default: detached, falls der Test keinen Status setzt.
		link.Status = domain.SessionLinkStatusDetached
	}
	return domain.AnalyzeManifestResult{Analysis: s.analysis, SessionLink: link}, nil
}

func newAnalyzeServer(t *testing.T, stub *linkAwareAnalysisInbound) *httptest.Server {
	t.Helper()
	resolver := auth.NewStaticProjectResolver(map[string]auth.ProjectConfig{
		"demo": {Token: "demo-token", AllowedOrigins: []string{"http://localhost:5173"}},
	})
	handler := &apihttp.AnalyzeHandler{
		UseCase:  stub,
		Resolver: resolver,
		Logger:   slog.New(slog.NewJSONHandler(io.Discard, nil)),
	}
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

func postAnalyze(t *testing.T, srv *httptest.Server, token, body string) (*http.Response, map[string]any) {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost,
		srv.URL+"/api/analyze", strings.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("X-MTrace-Token", token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if len(raw) == 0 {
		return resp, nil
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("decode body %q: %v", raw, err)
	}
	return resp, out
}

func TestAnalyze_WrapperShape_DetachedDefault(t *testing.T) {
	t.Parallel()
	stub := &linkAwareAnalysisInbound{
		analysis: domain.StreamAnalysisResult{
			AnalyzerVersion: "0.4.0", PlaylistType: domain.PlaylistTypeMaster,
		},
	}
	srv := newAnalyzeServer(t, stub)
	body := `{"kind":"url","url":"https://cdn.example.test/m.m3u8"}`
	resp, payload := postAnalyze(t, srv, "", body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d want 200", resp.StatusCode)
	}
	analysis, ok := payload["analysis"].(map[string]any)
	if !ok {
		t.Fatalf("response missing top-level `analysis`: %v", payload)
	}
	if analysis["playlistType"] != "master" {
		t.Errorf("analysis.playlistType = %v want master", analysis["playlistType"])
	}
	link, ok := payload["session_link"].(map[string]any)
	if !ok {
		t.Fatalf("response missing top-level `session_link`: %v", payload)
	}
	if link["status"] != "detached" {
		t.Errorf("session_link.status = %v want detached", link["status"])
	}
	// Optional-Felder dürfen bei detached nicht ausgegeben werden.
	if _, present := link["session_id"]; present {
		t.Errorf("session_link.session_id must be omitted for detached: %v", link)
	}
	if stub.gotReq.ProjectID != "" {
		t.Errorf("ProjectID must stay empty without token+link fields, got %q", stub.gotReq.ProjectID)
	}
}

func TestAnalyze_WrapperShape_LinkedFields(t *testing.T) {
	t.Parallel()
	stub := &linkAwareAnalysisInbound{
		analysis: domain.StreamAnalysisResult{
			AnalyzerVersion: "0.4.0", PlaylistType: domain.PlaylistTypeMedia,
		},
		resultLink: domain.SessionLink{
			Status:        domain.SessionLinkStatusLinked,
			ProjectID:     "demo",
			SessionID:     "sess-1",
			CorrelationID: "2f6f1a3c-9fb9-4c0b-a78f-2f41d8f6e1e7",
		},
	}
	srv := newAnalyzeServer(t, stub)
	body := `{"kind":"url","url":"https://cdn.example.test/m.m3u8","correlation_id":"2f6f1a3c-9fb9-4c0b-a78f-2f41d8f6e1e7"}`
	resp, payload := postAnalyze(t, srv, "demo-token", body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d want 200", resp.StatusCode)
	}
	link, _ := payload["session_link"].(map[string]any)
	if link["status"] != "linked" || link["session_id"] != "sess-1" ||
		link["correlation_id"] != "2f6f1a3c-9fb9-4c0b-a78f-2f41d8f6e1e7" ||
		link["project_id"] != "demo" {
		t.Errorf("link payload mismatch: %v", link)
	}
	// Project wurde aus Token aufgelöst und in den Use-Case-Request
	// gereicht.
	if stub.gotReq.ProjectID != "demo" || stub.gotReq.CorrelationID == "" {
		t.Errorf("use-case request missing project/cid: %+v", stub.gotReq)
	}
}

func TestAnalyze_AuthMatrix_NoTokenNoLinkFields_200Detached(t *testing.T) {
	t.Parallel()
	stub := &linkAwareAnalysisInbound{
		analysis: domain.StreamAnalysisResult{AnalyzerVersion: "0.4.0"},
	}
	srv := newAnalyzeServer(t, stub)
	body := `{"kind":"url","url":"https://cdn.example.test/m.m3u8"}`
	resp, payload := postAnalyze(t, srv, "", body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d want 200 (no token, no link fields)", resp.StatusCode)
	}
	link, _ := payload["session_link"].(map[string]any)
	if link["status"] != "detached" {
		t.Errorf("link.status = %v want detached", link["status"])
	}
	if stub.called != 1 {
		t.Errorf("use case must be called for unbound analyze, got %d calls", stub.called)
	}
}

func TestAnalyze_AuthMatrix_LinkFieldsWithoutToken_401(t *testing.T) {
	t.Parallel()
	stub := &linkAwareAnalysisInbound{}
	srv := newAnalyzeServer(t, stub)
	body := `{"kind":"url","url":"https://cdn.example.test/m.m3u8","correlation_id":"some-cid"}`
	resp, payload := postAnalyze(t, srv, "", body)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d want 401 (link fields without token)", resp.StatusCode)
	}
	if payload["code"] != "unauthorized" {
		t.Errorf("error code = %v want unauthorized", payload["code"])
	}
	if stub.called != 0 {
		t.Errorf("use case must not be called on auth failure, got %d calls", stub.called)
	}
}

func TestAnalyze_AuthMatrix_LinkFieldsWithInvalidToken_401(t *testing.T) {
	t.Parallel()
	stub := &linkAwareAnalysisInbound{}
	srv := newAnalyzeServer(t, stub)
	body := `{"kind":"url","url":"https://cdn.example.test/m.m3u8","session_id":"sess-1"}`
	resp, _ := postAnalyze(t, srv, "totally-bogus-token", body)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d want 401 (invalid token + link fields)", resp.StatusCode)
	}
	if stub.called != 0 {
		t.Errorf("use case must not be called on auth failure, got %d calls", stub.called)
	}
}

func TestAnalyze_AuthMatrix_OnlySessionIDTriggersAuth(t *testing.T) {
	t.Parallel()
	// Pin: `session_id` allein reicht, um die Auth-Pflicht zu
	// triggern (API-Kontrakt §4: `correlation_id` ODER `session_id`).
	stub := &linkAwareAnalysisInbound{}
	srv := newAnalyzeServer(t, stub)
	body := `{"kind":"text","text":"#EXTM3U\n","session_id":"sess-1"}`
	resp, _ := postAnalyze(t, srv, "", body)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("session_id alone must trigger 401, got %d", resp.StatusCode)
	}
}

// TestAnalyze_OptionsPreflight pinnt §4.5 DoD-Item 4: OPTIONS
// /api/analyze liefert `Access-Control-Allow-Methods: POST, OPTIONS`
// und `Access-Control-Allow-Headers: Content-Type, X-MTrace-Token,
// X-MTrace-Project` für einen erlaubten Origin.
func TestAnalyze_OptionsPreflight(t *testing.T) {
	t.Parallel()
	// Voller Router, damit der OPTIONS-Pfad inkl. CORS gewired ist.
	srv := newRouterTestServer(t)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodOptions,
		srv.URL+"/api/analyze", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "POST")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d want 204", resp.StatusCode)
	}
	methods := resp.Header.Get("Access-Control-Allow-Methods")
	if !strings.Contains(methods, "POST") || !strings.Contains(methods, "OPTIONS") {
		t.Errorf("Allow-Methods = %q want POST + OPTIONS", methods)
	}
	headers := resp.Header.Get("Access-Control-Allow-Headers")
	for _, want := range []string{"Content-Type", "X-MTrace-Token", "X-MTrace-Project"} {
		if !strings.Contains(headers, want) {
			t.Errorf("Allow-Headers missing %q (got %q)", want, headers)
		}
	}
	if resp.Header.Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Errorf("Allow-Origin echo missing: %q", resp.Header.Get("Access-Control-Allow-Origin"))
	}
}

// newRouterTestServer wired den vollen API-Router mit Analyze-Handler
// und CORS-Allowlist, damit OPTIONS-Preflight-Tests gegen die echte
// Routing-Konfiguration laufen. Stub-AnalyzeInbound liefert ein
// minimales detached-Result für POST-Pfade.
func newRouterTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	resolver := auth.NewStaticProjectResolver(map[string]auth.ProjectConfig{
		"demo": {Token: "demo-token", AllowedOrigins: []string{"http://localhost:5173"}},
	})
	publisher := metrics.NewPrometheusPublisher()
	stubAnalyze := &linkAwareAnalysisInbound{
		analysis: domain.StreamAnalysisResult{AnalyzerVersion: "0.4.0"},
	}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	router := apihttp.NewRouter(nil, nil, stubAnalyze, resolver, resolver, publisher.Handler(), nil, nil, nil, logger)
	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)
	return srv
}

func TestAnalyze_OptionsPreflight_RejectsForeignOrigin(t *testing.T) {
	t.Parallel()
	srv := newRouterTestServer(t)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodOptions,
		srv.URL+"/api/analyze", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Origin", "http://attacker.example")
	req.Header.Set("Access-Control-Request-Method", "POST")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("status = %d want 403 (origin not allowed)", resp.StatusCode)
	}
}
