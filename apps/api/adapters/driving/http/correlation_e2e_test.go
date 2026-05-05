package http_test

import (
	"net/http"
	"strings"
	"testing"
)

// plan-0.4.0 §4.7 — End-to-End-Tests für die gemischte Korrelation
// und die Degradationsmatrix. Spec-Anker:
// `spec/telemetry-model.md` §2.5 (correlation_id vs trace_id),
// §1.4 (network.*-Meta + session_boundaries[]); API-Kontrakt §3.6.

// TestE2E_MixedEventTypes_ShareCorrelationID pinnt §4.7 DoD-Item 1:
// Innerhalb einer Session tragen alle Event-Typen — Player-,
// Manifest- und Segment-Events — dieselbe `correlation_id`.
func TestE2E_MixedEventTypes_ShareCorrelationID(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	body := `{
	  "schema_version": "1.0",
	  "events": [
	    {"event_name":"playback_started","project_id":"demo","session_id":"sess-mix","client_timestamp":"2026-04-28T12:00:00.000Z","sequence_number":1,"sdk":{"name":"@npm9912/player-sdk","version":"0.5.0"}},
	    {"event_name":"manifest_loaded","project_id":"demo","session_id":"sess-mix","client_timestamp":"2026-04-28T12:00:00.100Z","sequence_number":2,"sdk":{"name":"@npm9912/player-sdk","version":"0.5.0"},"meta":{"network.kind":"manifest","network.detail_status":"available"}},
	    {"event_name":"segment_loaded","project_id":"demo","session_id":"sess-mix","client_timestamp":"2026-04-28T12:00:00.200Z","sequence_number":3,"sdk":{"name":"@npm9912/player-sdk","version":"0.5.0"},"meta":{"network.kind":"segment","network.detail_status":"available"}},
	    {"event_name":"rebuffer_started","project_id":"demo","session_id":"sess-mix","client_timestamp":"2026-04-28T12:00:00.300Z","sequence_number":4,"sdk":{"name":"@npm9912/player-sdk","version":"0.5.0"}},
	    {"event_name":"playback_error","project_id":"demo","session_id":"sess-mix","client_timestamp":"2026-04-28T12:00:00.400Z","sequence_number":5,"sdk":{"name":"@npm9912/player-sdk","version":"0.5.0"}}
	  ]
	}`
	if r := postEvents(t, srv, "demo-token", body); r.StatusCode != http.StatusAccepted {
		t.Fatalf("post: %d", r.StatusCode)
	}

	resp, payload := getJSON(t, srv.URL, "/api/stream-sessions/sess-mix")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("detail: %d", resp.StatusCode)
	}
	session, _ := payload["session"].(map[string]any)
	sessionCID, _ := session["correlation_id"].(string)
	if sessionCID == "" {
		t.Fatalf("session correlation_id is empty: %+v", session)
	}
	events, _ := payload["events"].([]any)
	if len(events) != 5 {
		t.Fatalf("expected 5 events, got %d", len(events))
	}
	for i, raw := range events {
		ev, _ := raw.(map[string]any)
		eventCID, _ := ev["correlation_id"].(string)
		if eventCID != sessionCID {
			t.Errorf("event[%d].correlation_id = %q, want %q (session)", i, eventCID, sessionCID)
		}
	}
}

// TestE2E_CrossBatch_SameCorrelationID pinnt §4.7 DoD-Item 1
// (correlation_id-Hälfte): zwei separate Batches derselben Session
// teilen die `correlation_id`. Die `trace_id`-batchbezogene
// Semantik wird durch
// `TestE2E_TraceparentPropagation_SameTraceID` abgedeckt — dort
// zeigt der Propagation-Pfad explizit, dass `trace_id` aus dem
// Header kommt und nicht intrinsisch session-gebunden ist.
func TestE2E_CrossBatch_SameCorrelationID(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)

	first := `{
	  "schema_version": "1.0",
	  "events": [{"event_name":"manifest_loaded","project_id":"demo","session_id":"sess-2bat","client_timestamp":"2026-04-28T12:00:00.000Z","sequence_number":1,"sdk":{"name":"@npm9912/player-sdk","version":"0.5.0"}}]
	}`
	second := `{
	  "schema_version": "1.0",
	  "events": [{"event_name":"segment_loaded","project_id":"demo","session_id":"sess-2bat","client_timestamp":"2026-04-28T12:00:01.000Z","sequence_number":2,"sdk":{"name":"@npm9912/player-sdk","version":"0.5.0"}}]
	}`
	for _, body := range []string{first, second} {
		if r := postEvents(t, srv, "demo-token", body); r.StatusCode != http.StatusAccepted {
			t.Fatalf("post: %d", r.StatusCode)
		}
	}

	resp, payload := getJSON(t, srv.URL, "/api/stream-sessions/sess-2bat")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("detail: %d", resp.StatusCode)
	}
	events, _ := payload["events"].([]any)
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	cid := func(i int) string {
		ev, _ := events[i].(map[string]any)
		v, _ := ev["correlation_id"].(string)
		return v
	}
	if cid(0) == "" || cid(0) != cid(1) {
		t.Errorf("correlation_id must be equal across batches (cid0=%q, cid1=%q)", cid(0), cid(1))
	}
}

// TestE2E_Degradation_NetworkDetailUnavailable pinnt §4.7 DoD-Item 2:
// CORS-/Resource-Timing-Lücken werden als
// `network_detail_unavailable` mit `unavailable_reason` akzeptiert,
// das Event bleibt in der Timeline mit voller `correlation_id`.
func TestE2E_Degradation_NetworkDetailUnavailable(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	body := `{
	  "schema_version": "1.0",
	  "events": [
	    {"event_name":"manifest_loaded","project_id":"demo","session_id":"sess-deg","client_timestamp":"2026-04-28T12:00:00.000Z","sequence_number":1,"sdk":{"name":"@npm9912/player-sdk","version":"0.5.0"},"meta":{"network.kind":"manifest","network.detail_status":"network_detail_unavailable","network.unavailable_reason":"cors_timing_blocked"}}
	  ]
	}`
	if r := postEvents(t, srv, "demo-token", body); r.StatusCode != http.StatusAccepted {
		t.Fatalf("post: %d", r.StatusCode)
	}
	resp, payload := getJSON(t, srv.URL, "/api/stream-sessions/sess-deg")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("detail: %d", resp.StatusCode)
	}
	events, _ := payload["events"].([]any)
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	ev, _ := events[0].(map[string]any)
	meta, _ := ev["meta"].(map[string]any)
	if meta["network.detail_status"] != "network_detail_unavailable" {
		t.Errorf("network.detail_status preserved: got %v", meta["network.detail_status"])
	}
	if meta["network.unavailable_reason"] != "cors_timing_blocked" {
		t.Errorf("network.unavailable_reason preserved: got %v", meta["network.unavailable_reason"])
	}
	if cid, _ := ev["correlation_id"].(string); cid == "" {
		t.Errorf("event with degraded network detail still must carry correlation_id")
	}
}

// TestE2E_Degradation_CapabilityOnlyBoundary pinnt §4.7 DoD-Item 2
// Schreibpfad: SDK liefert kein Manifest-/Segment-Event, sondern
// nur den Capability-Marker als `session_boundaries[]`. Der
// Boundary wird mit dem Event-Batch geschickt (Partition-Match auf
// ein generisches `playback_started`-Event), persistiert und im
// Read-Shape als `network_signal_absent[]` exposed.
func TestE2E_Degradation_CapabilityOnlyBoundary(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	body := `{
	  "schema_version": "1.0",
	  "events": [
	    {"event_name":"playback_started","project_id":"demo","session_id":"sess-cap","client_timestamp":"2026-04-28T12:00:00.000Z","sequence_number":1,"sdk":{"name":"@npm9912/player-sdk","version":"0.5.0"}}
	  ],
	  "session_boundaries": [
	    {"kind":"network_signal_absent","project_id":"demo","session_id":"sess-cap","network_kind":"segment","adapter":"native_hls","reason":"native_hls_unavailable","client_timestamp":"2026-04-28T12:00:00.000Z"}
	  ]
	}`
	if r := postEvents(t, srv, "demo-token", body); r.StatusCode != http.StatusAccepted {
		t.Fatalf("post: %d", r.StatusCode)
	}
	resp, payload := getJSON(t, srv.URL, "/api/stream-sessions/sess-cap")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("detail: %d", resp.StatusCode)
	}
	session, _ := payload["session"].(map[string]any)
	arr, _ := session["network_signal_absent"].([]any)
	if len(arr) != 1 {
		t.Fatalf("expected 1 boundary in read shape, got %v", session["network_signal_absent"])
	}
	entry, _ := arr[0].(map[string]any)
	if entry["kind"] != "segment" || entry["adapter"] != "native_hls" || entry["reason"] != "native_hls_unavailable" {
		t.Errorf("unexpected boundary read shape: %v", entry)
	}
	// DoD-Anker: native HLS-Limitierung → keine Manifest-/Segment-
	// Events nötig; nur das Capability-Signal ist persistiert.
	events, _ := payload["events"].([]any)
	if len(events) != 1 {
		t.Errorf("expected 1 event (the seed), got %d", len(events))
	}
}

// TestE2E_TraceparentPropagation_SameTraceID pinnt §4.7 DoD-Item 1
// (alternativer Pfad): wenn zwei Batches mit demselben
// `traceparent`-Header geschickt werden, übernimmt der Server die
// `trace_id` — das ist die Tempo-Cross-Trace-Suche aus §2.5.
func TestE2E_TraceparentPropagation_SameTraceID(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)

	traceparent := "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
	body := `{
	  "schema_version": "1.0",
	  "events": [{"event_name":"manifest_loaded","project_id":"demo","session_id":"sess-tp","client_timestamp":"2026-04-28T12:00:00.000Z","sequence_number":1,"sdk":{"name":"@npm9912/player-sdk","version":"0.5.0"}}]
	}`
	for i := 0; i < 2; i++ {
		req, err := http.NewRequest(http.MethodPost, srv.URL+"/api/playback-events", strings.NewReader(body))
		if err != nil {
			t.Fatalf("new request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-MTrace-Token", "demo-token")
		req.Header.Set("traceparent", traceparent)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("post: %v", err)
		}
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusAccepted {
			t.Fatalf("post[%d] expected 202, got %d", i, resp.StatusCode)
		}
	}

	resp, payload := getJSON(t, srv.URL, "/api/stream-sessions/sess-tp")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("detail: %d", resp.StatusCode)
	}
	events, _ := payload["events"].([]any)
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	tid := func(i int) string {
		ev, _ := events[i].(map[string]any)
		v, _ := ev["trace_id"].(string)
		return v
	}
	const wantTraceID = "4bf92f3577b34da6a3ce929d0e0e4736"
	if tid(0) != wantTraceID || tid(1) != wantTraceID {
		t.Errorf("trace_id should be propagated from traceparent: got tid0=%q, tid1=%q, want %q", tid(0), tid(1), wantTraceID)
	}
}
