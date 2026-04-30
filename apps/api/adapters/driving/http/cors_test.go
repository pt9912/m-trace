package http_test

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

// optionsRequest sendet eine Preflight-OPTIONS-Anfrage mit dem
// passenden Access-Control-Request-Method-Header.
func optionsRequest(t *testing.T, srvURL, path, origin, method string) *http.Response {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodOptions, srvURL+path, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	if method != "" {
		req.Header.Set("Access-Control-Request-Method", method)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	return resp
}

// postEventsWithOrigin ist eine Variante von postEvents mit
// zusätzlichem Origin-Header.
func postEventsWithOrigin(t *testing.T, srvURL, token, origin, body string) *http.Response {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, srvURL+"/api/playback-events", strings.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if token != "" {
		req.Header.Set("X-MTrace-Token", token)
	}
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	return resp
}

func getWithOrigin(t *testing.T, srvURL, path, origin string) *http.Response {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srvURL+path, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	return resp
}

// CORS-Test 1: Preflight OPTIONS /api/playback-events mit registriertem
// Origin → 204 mit konkretem Allow-Origin (kein `*`).
func TestCORS_Preflight_PlaybackEvents_Allowed(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	resp := optionsRequest(t, srv.URL, "/api/playback-events", "http://localhost:5173", http.MethodPost)
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Errorf("Access-Control-Allow-Origin=%q want concrete origin", got)
	}
	if got := resp.Header.Get("Access-Control-Allow-Methods"); !strings.Contains(got, "POST") {
		t.Errorf("Access-Control-Allow-Methods=%q want to contain POST", got)
	}
	if resp.Header.Get("Access-Control-Allow-Credentials") != "" {
		t.Errorf("Access-Control-Allow-Credentials present — must be omitted (NF-31/NF-32)")
	}
}

// CORS-Test 2: Preflight OPTIONS /api/playback-events mit unbekanntem
// Origin → 403.
func TestCORS_Preflight_PlaybackEvents_UnknownOrigin(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	resp := optionsRequest(t, srv.URL, "/api/playback-events", "http://attacker.example", http.MethodPost)
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403, got %d", resp.StatusCode)
	}
}

// CORS-Test 3+3a+6: POST mit gültigem demo-Token aber Origin aus other-
// Allowlist → 403; keine Side-Effects (kein Event persistiert, kein
// Rate-Limit-Token verbraucht); keine `*`-Response.
func TestCORS_Post_ProjectOriginMismatch_403(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	resp := postEventsWithOrigin(t, srv.URL, "demo-token", "http://other.example", validBody)
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Access-Control-Allow-Origin"); got == "*" {
		t.Errorf("Access-Control-Allow-Origin=%q must never be * with project token", got)
	}

	// Side-effect-test: ein erneuter, valider POST danach muss
	// problemlos durchgehen. Falls der Origin-Mismatch fälschlich Rate-
	// Limit-Tokens verbraucht hätte, schlüge der Folge-Request fehl.
	clean := postEvents(t, srv, "demo-token", validBody)
	if clean.StatusCode != http.StatusAccepted {
		t.Errorf("subsequent valid post should be accepted (no token-budget consumed by 403), got %d", clean.StatusCode)
	}
}

// CORS-Test 4: POST ohne Origin (CLI/curl) → 202 wie zuvor.
func TestCORS_Post_NoOrigin_StillAccepted(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	resp := postEventsWithOrigin(t, srv.URL, "demo-token", "", validBody)
	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("expected 202 for CLI/curl path, got %d", resp.StatusCode)
	}
}

// CORS-Test 5: Antworten tragen `Vary` mit Origin und den beiden
// Access-Control-Request-Headern.
func TestCORS_Vary_HeaderOnEveryResponse(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	// Preflight ✓.
	preflight := optionsRequest(t, srv.URL, "/api/playback-events", "http://localhost:5173", http.MethodPost)
	if got := preflight.Header.Get("Vary"); !strings.Contains(got, "Origin") {
		t.Errorf("preflight Vary=%q missing Origin", got)
	}
	// Echter POST ✓.
	post := postEventsWithOrigin(t, srv.URL, "demo-token", "http://localhost:5173", validBody)
	if got := post.Header.Get("Vary"); !strings.Contains(got, "Origin") {
		t.Errorf("POST Vary=%q missing Origin", got)
	}
}

// TestCORS_Preflight_Dashboard_UnknownOrigin deckt den 403-Pfad des
// Dashboard-Preflights ab — analog zum SDK-Pfad, aber an einer anderen
// Route.
func TestCORS_Preflight_Dashboard_UnknownOrigin(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	resp := optionsRequest(t, srv.URL, "/api/stream-sessions", "http://attacker.example", http.MethodGet)
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403, got %d", resp.StatusCode)
	}
}

// CORS-Test 7: Preflight OPTIONS /api/stream-sessions mit registriertem
// Origin → 204 mit `Access-Control-Allow-Methods: GET, OPTIONS`.
func TestCORS_Preflight_Dashboard_Allowed(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	resp := optionsRequest(t, srv.URL, "/api/stream-sessions", "http://localhost:5173", http.MethodGet)
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Access-Control-Allow-Methods"); !strings.Contains(got, "GET") || !strings.Contains(got, "OPTIONS") {
		t.Errorf("Access-Control-Allow-Methods=%q want GET, OPTIONS", got)
	}
	if strings.Contains(resp.Header.Get("Access-Control-Allow-Methods"), "POST") {
		t.Errorf("dashboard preflight must not advertise POST")
	}
}

func TestCORS_DashboardGet_AllowedOriginHeader(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)

	post := postEventsWithOrigin(t, srv.URL, "demo-token", "http://localhost:5173", validBody)
	if post.StatusCode != http.StatusAccepted {
		t.Fatalf("seed post: expected 202, got %d", post.StatusCode)
	}

	resp := getWithOrigin(t, srv.URL, "/api/stream-sessions", "http://localhost:5173")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Errorf("Access-Control-Allow-Origin=%q want concrete origin", got)
	}
}
