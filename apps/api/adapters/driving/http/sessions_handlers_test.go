package http_test

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func getJSON(t *testing.T, srv string, path string) (*http.Response, map[string]any) {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, srv+path, nil)
	if err != nil {
		t.Fatalf("new request %s: %v", path, err)
	}
	// Read-Endpunkte sind ab plan-0.4.0 §4.2 tokenpflichtig; Tests
	// verwenden den im Test-Setup wired Static-Resolver (Token
	// "demo-token", Project "demo").
	req.Header.Set("X-MTrace-Token", "demo-token")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("get %s: %v", path, err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if len(body) == 0 {
		return resp, nil
	}
	var out map[string]any
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("decode body %q: %v", body, err)
	}
	return resp, out
}

// TestHTTP_StreamSessions_HappyPath verifiziert den 0.1.0-Smoke-Pfad:
// nach einem POST /api/playback-events liefert
// GET /api/stream-sessions die so erzeugte Session.
func TestHTTP_StreamSessions_HappyPath(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	resp := postEvents(t, srv, "demo-token", validBody)
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("post: expected 202, got %d", resp.StatusCode)
	}

	resp, body := getJSON(t, srv.URL, "/api/stream-sessions")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list: expected 200, got %d", resp.StatusCode)
	}
	sessions, ok := body["sessions"].([]any)
	if !ok || len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %v", body["sessions"])
	}
	first, _ := sessions[0].(map[string]any)
	if first["session_id"] != "sess-1" {
		t.Errorf("session_id=%v want sess-1", first["session_id"])
	}
	if first["state"] != "active" {
		t.Errorf("state=%v want active", first["state"])
	}
	if first["event_count"].(float64) != 1 {
		t.Errorf("event_count=%v want 1", first["event_count"])
	}
}

// TestHTTP_StreamSessions_InvalidLimit prüft den Validierungs-Pfad
// für eine kaputte Limit-Query.
func TestHTTP_StreamSessions_InvalidLimit(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	resp, body := getJSON(t, srv.URL, "/api/stream-sessions?limit=abc")
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	if body["error"] != "limit_invalid" {
		t.Errorf("error=%v want limit_invalid", body["error"])
	}
}

// TestHTTP_StreamSessions_MalformedCursor verifiziert den 400-Pfad
// bei syntaktisch defektem Cursor — Wire-Format-Klasse aus
// API-Kontrakt §10.3 / ADR-0004 §6.
func TestHTTP_StreamSessions_MalformedCursor(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	resp, body := getJSON(t, srv.URL, "/api/stream-sessions?cursor=not-a-base64")
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	if body["error"] != "cursor_invalid_malformed" {
		t.Errorf("error=%v want cursor_invalid_malformed", body["error"])
	}
}

// TestHTTP_StreamSessions_LegacyCursor verifiziert die dauerhafte
// Reject-Klasse: ein 0.1.x/0.2.x/0.3.x-Cursor mit `pid`-Feld liefert
// 400 cursor_invalid_legacy (ADR-0004 §6).
func TestHTTP_StreamSessions_LegacyCursor(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	legacyCursor := encodeCursorForTest(t, `{"pid":"other-process","sa":"2026-04-28T12:00:00Z","sid":"s1"}`)
	resp, body := getJSON(t, srv.URL, "/api/stream-sessions?cursor="+url.QueryEscape(legacyCursor))
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	if body["error"] != "cursor_invalid_legacy" {
		t.Errorf("body=%v want cursor_invalid_legacy", body)
	}
}

// TestHTTP_StreamSessions_Pagination verifiziert mehrere Pages mit
// next_cursor — keine Duplikate, keine Lücken.
func TestHTTP_StreamSessions_Pagination(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	// 3 Sessions per separaten POSTs, damit unterschiedliche
	// session_ids existieren.
	for _, sid := range []string{"sa", "sb", "sc"} {
		body := strings.Replace(validBody, `"session_id": "sess-1"`, `"session_id": "`+sid+`"`, 1)
		if r := postEvents(t, srv, "demo-token", body); r.StatusCode != http.StatusAccepted {
			t.Fatalf("post %s: %d", sid, r.StatusCode)
		}
	}

	resp, body := getJSON(t, srv.URL, "/api/stream-sessions?limit=2")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("page 1: %d", resp.StatusCode)
	}
	if got := len(body["sessions"].([]any)); got != 2 {
		t.Fatalf("page 1 size %d want 2", got)
	}
	cursor, ok := body["next_cursor"].(string)
	if !ok || cursor == "" {
		t.Fatalf("page 1 missing next_cursor")
	}

	resp, body = getJSON(t, srv.URL, "/api/stream-sessions?limit=2&cursor="+url.QueryEscape(cursor))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("page 2: %d", resp.StatusCode)
	}
	if got := len(body["sessions"].([]any)); got != 1 {
		t.Fatalf("page 2 size %d want 1", got)
	}
	if _, has := body["next_cursor"]; has {
		t.Errorf("page 2 should not carry next_cursor (last page)")
	}
}

// TestHTTP_StreamSessionsByID_NotFound deckt 404 ab.
func TestHTTP_StreamSessionsByID_NotFound(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	resp, _ := getJSON(t, srv.URL, "/api/stream-sessions/missing")
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

// TestHTTP_StreamSessionsByID_InvalidEventsLimit deckt den 400-Pfad
// für eine kaputte events_limit-Query.
func TestHTTP_StreamSessionsByID_InvalidEventsLimit(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	resp, body := getJSON(t, srv.URL, "/api/stream-sessions/sess-1?events_limit=abc")
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	if body["error"] != "events_limit_invalid" {
		t.Errorf("error=%v want events_limit_invalid", body["error"])
	}
}

// TestHTTP_StreamSessionsByID_MalformedCursor deckt den
// cursor_invalid_malformed-Pfad für ein defektes events_cursor.
func TestHTTP_StreamSessionsByID_MalformedCursor(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	resp, body := getJSON(t, srv.URL, "/api/stream-sessions/sess-1?events_cursor=AAAA")
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	if body["error"] != "cursor_invalid_malformed" {
		t.Errorf("body=%v want cursor_invalid_malformed", body)
	}
}

// TestHTTP_StreamSessionsByID_LegacyCursor verifiziert die dauerhafte
// Legacy-Reject-Klasse für den Event-Cursor.
func TestHTTP_StreamSessionsByID_LegacyCursor(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	if r := postEvents(t, srv, "demo-token", validBody); r.StatusCode != http.StatusAccepted {
		t.Fatalf("post: %d", r.StatusCode)
	}
	legacy := encodeCursorForTest(t, `{"pid":"other","rcv":"2026-04-28T12:00:00Z","ing":1}`)
	resp, body := getJSON(t, srv.URL, "/api/stream-sessions/sess-1?events_cursor="+url.QueryEscape(legacy))
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	if body["error"] != "cursor_invalid_legacy" {
		t.Errorf("body=%v want cursor_invalid_legacy", body)
	}
}

// TestHTTP_StreamSessionsByID_EmptyID prüft, dass die Trailing-Slash-
// Route ohne ID vom mux auf 404 abgebildet wird (keine eigene Route
// für `/api/stream-sessions/`).
func TestHTTP_StreamSessionsByID_EmptyID(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	req, err := http.NewRequest(http.MethodGet, srv.URL+"/api/stream-sessions/", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("X-MTrace-Token", "demo-token")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

// TestHTTP_StreamSessionsByID_HappyPath erzeugt eine Session via POST
// und holt sie inkl. Events.
func TestHTTP_StreamSessionsByID_HappyPath(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	resp := postEvents(t, srv, "demo-token", validBody)
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("post: %d", resp.StatusCode)
	}

	resp, body := getJSON(t, srv.URL, "/api/stream-sessions/sess-1")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get: %d", resp.StatusCode)
	}
	session, _ := body["session"].(map[string]any)
	if session["session_id"] != "sess-1" {
		t.Errorf("session_id=%v want sess-1", session["session_id"])
	}
	events, _ := body["events"].([]any)
	if len(events) != 1 {
		t.Errorf("events len=%d want 1", len(events))
	}
	first, _ := events[0].(map[string]any)
	if first["ingest_sequence"].(float64) != 1 {
		t.Errorf("ingest_sequence=%v want 1", first["ingest_sequence"])
	}
}

// TestHTTP_StreamSessions_NetworkSignalAbsentDefaultEmpty pinnt den
// §4.4-D3-Default: jede Session-Read-Antwort trägt
// `network_signal_absent` als JSON-Array, auch wenn keine Boundaries
// persistiert sind (Spec §3.7.1: Default `[]`, kein `null`).
func TestHTTP_StreamSessions_NetworkSignalAbsentDefaultEmpty(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	if r := postEvents(t, srv, "demo-token", validBody); r.StatusCode != http.StatusAccepted {
		t.Fatalf("post: %d", r.StatusCode)
	}

	resp, body := getJSON(t, srv.URL, "/api/stream-sessions")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list: %d", resp.StatusCode)
	}
	sessions, _ := body["sessions"].([]any)
	first, _ := sessions[0].(map[string]any)
	raw, present := first["network_signal_absent"]
	if !present {
		t.Fatalf("session is missing network_signal_absent (default `[]`)")
	}
	arr, ok := raw.([]any)
	if !ok {
		t.Fatalf("network_signal_absent is not a JSON array: %T = %v", raw, raw)
	}
	if len(arr) != 0 {
		t.Errorf("expected empty default, got %v", arr)
	}

	// Detail-Read trägt das Feld ebenfalls.
	resp, body = getJSON(t, srv.URL, "/api/stream-sessions/sess-1")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("detail: %d", resp.StatusCode)
	}
	session, _ := body["session"].(map[string]any)
	rawDetail, present := session["network_signal_absent"]
	if !present {
		t.Fatalf("detail session is missing network_signal_absent")
	}
	if arr, ok := rawDetail.([]any); !ok || len(arr) != 0 {
		t.Errorf("detail expected empty default, got %v", rawDetail)
	}
}

// TestHTTP_StreamSessions_NetworkSignalAbsentRoundTrip pinnt den
// vollständigen §4.4-Pfad: ein POST /api/playback-events mit
// `session_boundaries[]` persistiert die Tripel, der Detail-Read
// liefert das `network_signal_absent`-Read-Shape (kind=NetworkKind,
// adapter, reason) und sortiert nach kind/adapter/reason.
func TestHTTP_StreamSessions_NetworkSignalAbsentRoundTrip(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	body := `{
	  "schema_version": "1.0",
	  "events": [
	    {
	      "event_name": "rebuffer_started",
	      "project_id": "demo",
	      "session_id": "sess-bnd",
	      "client_timestamp": "2026-04-28T12:00:00.000Z",
	      "sdk": { "name": "@npm9912/player-sdk", "version": "0.4.0" }
	    }
	  ],
	  "session_boundaries": [
	    {
	      "kind": "network_signal_absent",
	      "project_id": "demo",
	      "session_id": "sess-bnd",
	      "network_kind": "segment",
	      "adapter": "native_hls",
	      "reason": "native_hls_unavailable",
	      "client_timestamp": "2026-04-28T12:00:00.000Z"
	    },
	    {
	      "kind": "network_signal_absent",
	      "project_id": "demo",
	      "session_id": "sess-bnd",
	      "network_kind": "manifest",
	      "adapter": "hls.js",
	      "reason": "cors_timing_blocked",
	      "client_timestamp": "2026-04-28T12:00:00.000Z"
	    }
	  ]
	}`
	if r := postEvents(t, srv, "demo-token", body); r.StatusCode != http.StatusAccepted {
		t.Fatalf("post: %d", r.StatusCode)
	}

	resp, payload := getJSON(t, srv.URL, "/api/stream-sessions/sess-bnd")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("detail: %d", resp.StatusCode)
	}
	session, _ := payload["session"].(map[string]any)
	arr, ok := session["network_signal_absent"].([]any)
	if !ok || len(arr) != 2 {
		t.Fatalf("expected 2 boundaries in detail, got %v", session["network_signal_absent"])
	}
	// Sortierung: kind asc → adapter asc → reason asc.
	first, _ := arr[0].(map[string]any)
	if first["kind"] != "manifest" || first["adapter"] != "hls.js" || first["reason"] != "cors_timing_blocked" {
		t.Errorf("first boundary = %v", first)
	}
	second, _ := arr[1].(map[string]any)
	if second["kind"] != "segment" || second["adapter"] != "native_hls" || second["reason"] != "native_hls_unavailable" {
		t.Errorf("second boundary = %v", second)
	}
}

// TestHTTP_StreamSessions_BoundaryRejectedDoesNotPersist pinnt den
// atomaren 422-Pfad aus §4.4 D2: ein invalider Boundary-Block
// persistiert weder Events noch Boundaries und liefert 422.
func TestHTTP_StreamSessions_BoundaryRejectedDoesNotPersist(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	body := `{
	  "schema_version": "1.0",
	  "events": [
	    {
	      "event_name": "rebuffer_started",
	      "project_id": "demo",
	      "session_id": "sess-rej",
	      "client_timestamp": "2026-04-28T12:00:00.000Z",
	      "sdk": { "name": "@npm9912/player-sdk", "version": "0.4.0" }
	    }
	  ],
	  "session_boundaries": [
	    {
	      "kind": "totally_made_up",
	      "project_id": "demo",
	      "session_id": "sess-rej",
	      "network_kind": "segment",
	      "adapter": "native_hls",
	      "reason": "native_hls_unavailable",
	      "client_timestamp": "2026-04-28T12:00:00.000Z"
	    }
	  ]
	}`
	resp := postEvents(t, srv, "demo-token", body)
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", resp.StatusCode)
	}

	resp, payload := getJSON(t, srv.URL, "/api/stream-sessions/sess-rej")
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("rejected batch must not create session, got status %d body=%v", resp.StatusCode, payload)
	}
}

// TestHTTP_StreamSessions_EndSource_NullForActiveSession pinnt
// plan-0.4.0 §5 H1: aktive Sessions liefern `end_source: null`
// (Pflichtfeld, kein omitempty).
func TestHTTP_StreamSessions_EndSource_NullForActiveSession(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	if r := postEvents(t, srv, "demo-token", validBody); r.StatusCode != http.StatusAccepted {
		t.Fatalf("post: %d", r.StatusCode)
	}
	resp, payload := getJSON(t, srv.URL, "/api/stream-sessions/sess-1")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("detail: %d", resp.StatusCode)
	}
	session, _ := payload["session"].(map[string]any)
	raw, present := session["end_source"]
	if !present {
		t.Fatalf("end_source must be present (Pflichtfeld), got missing")
	}
	if raw != nil {
		t.Errorf("end_source for active session must be JSON null, got %T = %v", raw, raw)
	}
}

// TestHTTP_StreamSessions_EndSource_ClientForSessionEnded pinnt §5
// H1: explizites `session_ended`-Event setzt
// `end_source="client"` im Read-Shape.
func TestHTTP_StreamSessions_EndSource_ClientForSessionEnded(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	body := `{
	  "schema_version": "1.0",
	  "events": [
	    {"event_name":"playback_started","project_id":"demo","session_id":"sess-end","client_timestamp":"2026-04-28T12:00:00.000Z","sequence_number":1,"sdk":{"name":"@npm9912/player-sdk","version":"0.4.0"}},
	    {"event_name":"session_ended","project_id":"demo","session_id":"sess-end","client_timestamp":"2026-04-28T12:00:01.000Z","sequence_number":2,"sdk":{"name":"@npm9912/player-sdk","version":"0.4.0"}}
	  ]
	}`
	if r := postEvents(t, srv, "demo-token", body); r.StatusCode != http.StatusAccepted {
		t.Fatalf("post: %d", r.StatusCode)
	}
	resp, payload := getJSON(t, srv.URL, "/api/stream-sessions/sess-end")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("detail: %d", resp.StatusCode)
	}
	session, _ := payload["session"].(map[string]any)
	if session["end_source"] != "client" {
		t.Errorf("end_source = %v, want \"client\"", session["end_source"])
	}
	if session["state"] != "ended" {
		t.Errorf("state = %v, want \"ended\"", session["state"])
	}
}

// encodeCursorForTest base64-url-encoded eine raw-JSON-Payload (ohne
// Padding) — gleiche Codec-Form wie der Handler, aber bewusst gegen
// die Wire-Form gekoppelt statt gegen die interne Codec-API.
func encodeCursorForTest(t *testing.T, raw string) string {
	t.Helper()
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}
