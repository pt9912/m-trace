package http_test

import (
	"context"
	"io"
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
	// Read-Endpunkte sind ab plan-0.4.0 §4.2 tokenpflichtig; CORS-Tests
	// senden den Default-Test-Token, damit der Auth-Layer den Request
	// nicht 401t bevor der CORS-Layer dazu kommt.
	req.Header.Set("X-MTrace-Token", "demo-token")
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

// CORS-Test 2 (`0.12.0` §3.9): Preflight OPTIONS /api/playback-events
// mit unbekanntem Origin → `204` mit leerem Body und ohne
// Allow-Origin/Methods/Headers (minimale Ablehnung, keine Project-/
// Origin-Enumeration). `Vary: Origin` und `Cache-Control: no-store`
// bleiben gesetzt — das ist die deterministische Form aus
// `spec/backend-api-contract.md` §3.9.
func TestCORS_Preflight_PlaybackEvents_UnknownOrigin(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	resp := optionsRequest(t, srv.URL, "/api/playback-events", "http://attacker.example", http.MethodPost)
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected 204, got %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Allow-Origin must not leak on unknown origin, got %q", got)
	}
	if got := resp.Header.Get("Access-Control-Allow-Methods"); got != "" {
		t.Errorf("Allow-Methods must not leak on unknown origin, got %q", got)
	}
	if got := resp.Header.Get("Access-Control-Allow-Headers"); got != "" {
		t.Errorf("Allow-Headers must not leak on unknown origin, got %q", got)
	}
	if got := resp.Header.Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control: want no-store, got %q", got)
	}
	if got := resp.Header.Get("Vary"); !strings.Contains(got, "Origin") {
		t.Errorf("Vary: must include Origin, got %q", got)
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

// TestCORS_Preflight_Dashboard_UnknownOrigin deckt den `0.12.0`-§3.9-
// Pfad des Dashboard-Preflights ab — `204` ohne Allow-Header für
// unbekannte Origin (analog zum SDK-Pfad).
func TestCORS_Preflight_Dashboard_UnknownOrigin(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	resp := optionsRequest(t, srv.URL, "/api/stream-sessions", "http://attacker.example", http.MethodGet)
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected 204, got %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Allow-Origin must not leak: %q", got)
	}
	if got := resp.Header.Get("Access-Control-Allow-Methods"); got != "" {
		t.Errorf("Allow-Methods must not leak: %q", got)
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

// TestCORS_Preflight_SseStream_Allowed pinnt Spec §10a: SSE-Preflight
// liefert `Allow-Methods: GET, OPTIONS` und exakte
// `Allow-Headers`-Liste inklusive `Last-Event-ID` für den
// fetch-basierten Reconnect-Backfill (plan-0.4.0 §5 H4 F2).
func TestCORS_Preflight_SseStream_Allowed(t *testing.T) {
	t.Parallel()
	srv := newTestServerWithSse(t)
	resp := optionsRequest(t, srv.URL, "/api/stream-sessions/stream", "http://localhost:5173", http.MethodGet)
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Access-Control-Allow-Methods"); got != "GET, OPTIONS" {
		t.Errorf("Access-Control-Allow-Methods=%q want %q", got, "GET, OPTIONS")
	}
	wantHeaders := "Content-Type, X-MTrace-Project, X-MTrace-Token, Last-Event-ID"
	if got := resp.Header.Get("Access-Control-Allow-Headers"); got != wantHeaders {
		t.Errorf("Access-Control-Allow-Headers=%q want %q", got, wantHeaders)
	}
	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Errorf("Access-Control-Allow-Origin=%q want concrete origin", got)
	}
}

// TestCORS_Preflight_SseStream_UnknownOrigin pinnt §3.9 / Spec §10a:
// unbekannter Origin → `204` mit leerem Body und ohne Allow-Header.
func TestCORS_Preflight_SseStream_UnknownOrigin(t *testing.T) {
	t.Parallel()
	srv := newTestServerWithSse(t)
	resp := optionsRequest(t, srv.URL, "/api/stream-sessions/stream", "http://attacker.example", http.MethodGet)
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected 204, got %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Allow-Origin must not leak on unknown origin, got %q", got)
	}
}

// TestCORS_Preflight_PlaybackEvents_HeaderSetExact pinnt die exakte
// `0.12.0`-Header-Allowlist (§3.9): Content-Type, Authorization,
// X-MTrace-Token, X-MTrace-Session-Token, traceparent — in dieser
// Reihenfolge, damit Contract-Fixtures byte-stabil bleiben.
func TestCORS_Preflight_PlaybackEvents_HeaderSetExact(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	resp := optionsRequest(t, srv.URL, "/api/playback-events", "http://localhost:5173", http.MethodPost)
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
	wantHeaders := "Content-Type, Authorization, X-MTrace-Token, X-MTrace-Session-Token, traceparent"
	if got := resp.Header.Get("Access-Control-Allow-Headers"); got != wantHeaders {
		t.Errorf("Allow-Headers: want %q, got %q", wantHeaders, got)
	}
	if got := resp.Header.Get("Access-Control-Max-Age"); got != "600" {
		t.Errorf("Max-Age: want 600, got %q", got)
	}
	if got := resp.Header.Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control: want no-store, got %q", got)
	}
}

// TestCORS_Preflight_PlaybackEvents_BodyEmpty pinnt §3.9: bekannte
// und unbekannte Origins liefern beide einen leeren Body. Verhindert
// Project-Enumeration über JSON-Bodies.
func TestCORS_Preflight_PlaybackEvents_BodyEmpty(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	cases := []struct {
		name, origin string
	}{
		{"known", "http://localhost:5173"},
		{"unknown", "http://attacker.example"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := optionsRequest(t, srv.URL, "/api/playback-events", tc.origin, http.MethodPost)
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("read: %v", err)
			}
			if len(body) != 0 {
				t.Errorf("body must be empty for %s origin, got %d bytes: %q", tc.name, len(body), body)
			}
		})
	}
}

// TestCORS_Preflight_AuthSessionTokens_Allowed pinnt §3.9 für den
// neuen Issuance-Endpoint — gleiche Header-Allowlist wie Playback,
// damit der Browser-Issuance-Pfad konsistent ist.
func TestCORS_Preflight_AuthSessionTokens_Allowed(t *testing.T) {
	t.Parallel()
	srv := newAuthSessionTestServer(t)
	resp := optionsRequest(t, srv.URL, "/api/auth/session-tokens", "http://localhost:5173", http.MethodPost)
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Errorf("Allow-Origin: want concrete origin, got %q", got)
	}
	if got := resp.Header.Get("Access-Control-Allow-Methods"); got != "POST, OPTIONS" {
		t.Errorf("Allow-Methods: want %q, got %q", "POST, OPTIONS", got)
	}
}

// TestCORS_Preflight_RequestMethod_Ignored: §3.9 bindet die globale
// Methods-Allowlist — der Server gibt unabhängig vom
// `Access-Control-Request-Method`-Header dieselbe `Allow-Methods`
// aus. Der Browser entscheidet selbst, ob die requested Method
// erlaubt ist.
func TestCORS_Preflight_RequestMethod_Ignored(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	// Wrong method (GET on a POST-only path) → still 204 with the
	// configured POST, OPTIONS allowlist. Der Browser wird den
	// Request-Method-Mismatch auf seiner Seite erkennen.
	resp := optionsRequest(t, srv.URL, "/api/playback-events", "http://localhost:5173", http.MethodGet)
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("status: want 204, got %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Access-Control-Allow-Methods"); got != "POST, OPTIONS" {
		t.Errorf("Allow-Methods: want POST, OPTIONS, got %q", got)
	}
}

// TestCORS_ValidateKeyNotRegisteredWithoutIngest pinnt §0.1
// Out-of-Scope deterministisch: `/api/ingest/streams/{id}/validate-key`
// ist **kein** produktiver Media-Server-Auth-Pfad und wird nur
// registriert, wenn die Ingest-Use-Case (SQLite-Persistenz) gewired
// ist. Der Standard-Test-Server hat kein Ingest → `404`. Falls sich
// das je ändert (z. B. routenfehlbedingte Registrierung), schlägt
// der Test sofort an (Review-Finding G2: deterministisch statt
// `204 OR 404`).
func TestCORS_ValidateKeyNotRegisteredWithoutIngest(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	resp := optionsRequest(t, srv.URL, "/api/ingest/streams/abc/validate-key", "http://localhost:5173", http.MethodPost)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("validate-key preflight without ingest setup: want 404, got %d", resp.StatusCode)
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
