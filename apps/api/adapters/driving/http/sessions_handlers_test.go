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
	resp, err := http.Get(srv + path)
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
// bei syntaktisch defektem Cursor.
func TestHTTP_StreamSessions_MalformedCursor(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	resp, body := getJSON(t, srv.URL, "/api/stream-sessions?cursor=not-a-base64")
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	if body["error"] != "cursor_invalid" {
		t.Errorf("error=%v want cursor_invalid", body["error"])
	}
}

// TestHTTP_StreamSessions_StaleCursor verifiziert den
// Storage-Restart-Pfad: ein Cursor mit fremder process_instance_id
// gibt 400 cursor_invalid mit reason=storage_restart.
func TestHTTP_StreamSessions_StaleCursor(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	// Wohlgeformter Cursor — andere PID als der Test-Server.
	staleCursor := encodeCursorForTest(t, `{"pid":"other-process","sa":"2026-04-28T12:00:00Z","sid":"s1"}`)
	resp, body := getJSON(t, srv.URL, "/api/stream-sessions?cursor="+url.QueryEscape(staleCursor))
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	if body["error"] != "cursor_invalid" || body["reason"] != "storage_restart" {
		t.Errorf("body=%v want cursor_invalid/storage_restart", body)
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

// encodeCursorForTest base64-url-encoded eine raw-JSON-Payload (ohne
// Padding) — gleiche Codec-Form wie der Handler, aber bewusst gegen
// die Wire-Form gekoppelt statt gegen die interne Codec-API.
func encodeCursorForTest(t *testing.T, raw string) string {
	t.Helper()
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}
