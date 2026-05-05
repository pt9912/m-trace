package http_test

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/auth"
	"github.com/pt9912/m-trace/apps/api/adapters/driven/persistence/inmemory"
	apihttp "github.com/pt9912/m-trace/apps/api/adapters/driving/http"
	"github.com/pt9912/m-trace/apps/api/hexagon/application"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// plan-0.4.0 §5 H4 — SSE-Backend-Endpunkt-Tests. Spec-Anker:
// `spec/backend-api-contract.md` §10a.

func newSseHandler(broker *application.EventBroker, events *inmemory.EventRepository, hb time.Duration) *apihttp.SseStreamHandler {
	resolver := auth.NewStaticProjectResolver(map[string]auth.ProjectConfig{
		"demo": {Token: "demo-token", AllowedOrigins: []string{"http://localhost:5173"}},
	})
	return &apihttp.SseStreamHandler{
		Resolver:  resolver,
		Events:    events,
		Broker:    broker,
		Logger:    slog.New(slog.NewJSONHandler(io.Discard, nil)),
		Heartbeat: hb,
	}
}

// TestSse_Auth_RejectsMissingToken pinnt Spec §10a: fehlender
// X-MTrace-Token → 401, ohne Stream zu öffnen.
func TestSse_Auth_RejectsMissingToken(t *testing.T) {
	t.Parallel()
	broker := application.NewEventBroker()
	events := inmemory.NewEventRepository()
	srv := httptest.NewServer(newSseHandler(broker, events, time.Hour))
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/api/stream-sessions/stream")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d want 401", resp.StatusCode)
	}
}

// TestSse_Auth_RejectsInvalidToken pinnt Spec §10a: ungültiger Token
// → 401.
func TestSse_Auth_RejectsInvalidToken(t *testing.T) {
	t.Parallel()
	broker := application.NewEventBroker()
	events := inmemory.NewEventRepository()
	srv := httptest.NewServer(newSseHandler(broker, events, time.Hour))
	t.Cleanup(srv.Close)

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/api/stream-sessions/stream", nil)
	req.Header.Set("X-MTrace-Token", "totally-bogus")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d want 401", resp.StatusCode)
	}
}

// TestSse_StreamHeaders pinnt Spec §10a: Content-Type, Cache-Control.
func TestSse_StreamHeaders(t *testing.T) {
	t.Parallel()
	broker := application.NewEventBroker()
	events := inmemory.NewEventRepository()
	srv := httptest.NewServer(newSseHandler(broker, events, time.Hour))
	t.Cleanup(srv.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/api/stream-sessions/stream", nil)
	req.Header.Set("X-MTrace-Token", "demo-token")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d want 200", resp.StatusCode)
	}
	if !strings.HasPrefix(resp.Header.Get("Content-Type"), "text/event-stream") {
		t.Errorf("Content-Type = %q want text/event-stream", resp.Header.Get("Content-Type"))
	}
	if resp.Header.Get("Cache-Control") != "no-store" {
		t.Errorf("Cache-Control = %q want no-store", resp.Header.Get("Cache-Control"))
	}
	// Drain bis Timeout, damit der Server-Goroutine sauber endet.
	_, _ = io.ReadAll(resp.Body)
}

// TestSse_LiveFrame pinnt Spec §10a: Live-Push eines neuen Events
// erzeugt einen `event: event_appended`-Frame mit `id`-Header
// (ingest_sequence) und JSON-data.
func TestSse_LiveFrame(t *testing.T) {
	t.Parallel()
	broker := application.NewEventBroker()
	events := inmemory.NewEventRepository()
	srv := httptest.NewServer(newSseHandler(broker, events, time.Hour))
	t.Cleanup(srv.Close)

	// Publish nach 50ms — gibt dem Server-Goroutine Zeit, sich beim
	// Broker zu registrieren.
	go func() {
		time.Sleep(50 * time.Millisecond)
		broker.Publish([]domain.PlaybackEvent{{
			EventName:      "manifest_loaded",
			ProjectID:      "demo",
			SessionID:      "sess-1",
			IngestSequence: 42,
		}})
	}()

	got := streamBytesUntilTimeout(t, srv.URL, http.Header{
		"X-MTrace-Token": []string{"demo-token"},
	}, 250*time.Millisecond)
	if !strings.Contains(got, "id: 42") {
		t.Errorf("frame missing id: 42, got %q", got)
	}
	if !strings.Contains(got, `"session_id":"sess-1"`) {
		t.Errorf("frame missing session_id, got %q", got)
	}
	if !strings.Contains(got, `"event_name":"manifest_loaded"`) {
		t.Errorf("frame missing event_name, got %q", got)
	}
	if !strings.Contains(got, "event: event_appended") {
		t.Errorf("frame missing event-name, got %q", got)
	}
}

// TestSse_Heartbeat pinnt Spec §10a: bei Idle-Timeout schickt der
// Server einen `: heartbeat`-Comment-Frame.
func TestSse_Heartbeat(t *testing.T) {
	t.Parallel()
	broker := application.NewEventBroker()
	events := inmemory.NewEventRepository()
	srv := httptest.NewServer(newSseHandler(broker, events, 30*time.Millisecond))
	t.Cleanup(srv.Close)

	got := streamBytesUntilTimeout(t, srv.URL, http.Header{
		"X-MTrace-Token": []string{"demo-token"},
	}, 200*time.Millisecond)
	if !strings.Contains(got, ": heartbeat") {
		t.Errorf("did not see heartbeat comment in stream output: %q", got)
	}
}

// TestSse_StreamSurvivesShortWriteTimeout pinnt: ein produktiver
// `http.Server` mit `WriteTimeout` kürzer als der SSE-Heartbeat darf
// den Stream NICHT vor dem ersten Heartbeat abbrechen. `cmd/api/main.go`
// setzt WriteTimeout=10s während der Heartbeat erst nach 15s kommt;
// ohne `SetWriteDeadline(time.Time{})` im SSE-Handler killt der Server
// die Connection mid-stream. `httptest.NewServer` setzt keinen
// WriteTimeout, deckt dieses Risiko also nicht ab — dieser Test nutzt
// einen echten `http.Server` mit aggressivem WriteTimeout, der ohne
// den SetWriteDeadline-Fix garantiert vor dem ersten Heartbeat
// zuschlägt. Spec §10a + plan-0.4.0 §9.4 (Findings post-§9.4-Closeout).
func TestSse_StreamSurvivesShortWriteTimeout(t *testing.T) {
	t.Parallel()
	broker := application.NewEventBroker()
	events := inmemory.NewEventRepository()
	handler := newSseHandler(broker, events, 80*time.Millisecond)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	server := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		// Aggressiver WriteTimeout: deutlich unter dem 80ms-Heartbeat.
		// Ohne SetWriteDeadline-Fix killt das den Stream, bevor der
		// erste Heartbeat-Write durchkommt.
		WriteTimeout: 50 * time.Millisecond,
	}
	t.Cleanup(func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	})
	go func() { _ = server.Serve(listener) }()

	baseURL := "http://" + listener.Addr().String()
	got := streamBytesUntilTimeout(t, baseURL, http.Header{
		"X-MTrace-Token": []string{"demo-token"},
	}, 300*time.Millisecond)
	// 300ms gibt mindestens drei Heartbeat-Ticks Spielraum (alle 80ms).
	if !strings.Contains(got, ": heartbeat") {
		t.Errorf("expected heartbeat through short WriteTimeout, got %q", got)
	}
}

// TestSse_Backfill_LastEventID pinnt Spec §10a: mit `Last-Event-ID`-
// Header bekommt der Konsument vor dem ersten Live-Frame alle
// persistierten Events mit `ingest_sequence > Last-Event-ID`.
func TestSse_Backfill_LastEventID(t *testing.T) {
	t.Parallel()
	broker := application.NewEventBroker()
	events := inmemory.NewEventRepository()
	// Persistierte Historie: drei Events, ingest_sequence 1..3.
	for i := int64(1); i <= 3; i++ {
		_ = events.Append(context.Background(), []domain.PlaybackEvent{{
			EventName:      "manifest_loaded",
			ProjectID:      "demo",
			SessionID:      "sess-old",
			IngestSequence: i,
		}})
	}
	srv := httptest.NewServer(newSseHandler(broker, events, time.Hour))
	t.Cleanup(srv.Close)

	got := streamBytesUntilTimeout(t, srv.URL, http.Header{
		"X-MTrace-Token": []string{"demo-token"},
		"Last-Event-ID":  []string{"1"},
	}, 200*time.Millisecond)
	if !strings.Contains(got, "id: 2") || !strings.Contains(got, "id: 3") {
		t.Errorf("backfill missing id 2 or 3, got %q", got)
	}
	if strings.Contains(got, "id: 1") {
		t.Errorf("backfill should skip Last-Event-ID itself (id 1), got %q", got)
	}
}

// TestSse_Backfill_CrossProjectScoped pinnt Spec §10a: Backfill ist
// project-skopiert. Ein `Last-Event-ID`-Wert eines fremden Events
// liefert keine fremden Events.
func TestSse_Backfill_CrossProjectScoped(t *testing.T) {
	t.Parallel()
	broker := application.NewEventBroker()
	events := inmemory.NewEventRepository()
	_ = events.Append(context.Background(), []domain.PlaybackEvent{
		{EventName: "manifest_loaded", ProjectID: "other", SessionID: "s", IngestSequence: 5},
		{EventName: "manifest_loaded", ProjectID: "demo", SessionID: "s", IngestSequence: 7},
	})
	srv := httptest.NewServer(newSseHandler(broker, events, time.Hour))
	t.Cleanup(srv.Close)

	got := streamBytesUntilTimeout(t, srv.URL, http.Header{
		"X-MTrace-Token": []string{"demo-token"},
		"Last-Event-ID":  []string{"0"},
	}, 200*time.Millisecond)
	if !strings.Contains(got, "id: 7") {
		t.Errorf("expected demo project event id=7 in backfill, got %q", got)
	}
	if strings.Contains(got, "id: 5") {
		t.Errorf("cross-project event id=5 must not appear, got %q", got)
	}
}

// TestSse_Backfill_TruncatedMarker pinnt Spec §10a: bei Backfill ≥
// Limit öffnet der Stream mit einem `event: backfill_truncated`-
// Frame, gefolgt von den `limit` neuesten Events. Konsumenten laden
// danach den Detail-Snapshot neu.
func TestSse_Backfill_TruncatedMarker(t *testing.T) {
	t.Parallel()
	broker := application.NewEventBroker()
	events := inmemory.NewEventRepository()
	// Drei persistierte Events, Limit für den Test auf 2 → der Stream
	// liefert Frame 1 + 2 plus den `backfill_truncated`-Header. Frame
	// 3 fällt raus (Konsument lädt Detail neu).
	for i := int64(1); i <= 3; i++ {
		_ = events.Append(context.Background(), []domain.PlaybackEvent{{
			EventName: "manifest_loaded", ProjectID: "demo",
			SessionID: "s", IngestSequence: i,
		}})
	}
	resolver := auth.NewStaticProjectResolver(map[string]auth.ProjectConfig{
		"demo": {Token: "demo-token", AllowedOrigins: []string{"http://localhost:5173"}},
	})
	handler := &apihttp.SseStreamHandler{
		Resolver:      resolver,
		Events:        events,
		Broker:        broker,
		Logger:        slog.New(slog.NewJSONHandler(io.Discard, nil)),
		Heartbeat:     time.Hour,
		BackfillLimit: 2, // <-- Test-only override
	}
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	got := streamBytesUntilTimeout(t, srv.URL, http.Header{
		"X-MTrace-Token": []string{"demo-token"},
		"Last-Event-ID":  []string{"0"},
	}, 200*time.Millisecond)
	if !strings.Contains(got, "event: backfill_truncated") {
		t.Errorf("expected backfill_truncated marker, got %q", got)
	}
	if !strings.Contains(got, `"oldest_ingest_sequence":1`) {
		t.Errorf("expected oldest_ingest_sequence=1 (smallest delivered), got %q", got)
	}
}

// TestSse_Backfill_ExactLimitNoTruncation pinnt B1-Fix: bei genau
// `BackfillLimit` Events darf der Server keinen `backfill_truncated`-
// Frame senden — nur > limit zählt als echte Lücke (Spec §10a "bei
// größerer Lücke").
func TestSse_Backfill_ExactLimitNoTruncation(t *testing.T) {
	t.Parallel()
	broker := application.NewEventBroker()
	events := inmemory.NewEventRepository()
	for i := int64(1); i <= 2; i++ {
		_ = events.Append(context.Background(), []domain.PlaybackEvent{{
			EventName: "manifest_loaded", ProjectID: "demo",
			SessionID: "s", IngestSequence: i,
		}})
	}
	resolver := auth.NewStaticProjectResolver(map[string]auth.ProjectConfig{
		"demo": {Token: "demo-token", AllowedOrigins: []string{"http://localhost:5173"}},
	})
	handler := &apihttp.SseStreamHandler{
		Resolver:      resolver,
		Events:        events,
		Broker:        broker,
		Logger:        slog.New(slog.NewJSONHandler(io.Discard, nil)),
		Heartbeat:     time.Hour,
		BackfillLimit: 2, // exakt = #events in DB → KEIN truncation-frame
	}
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	got := streamBytesUntilTimeout(t, srv.URL, http.Header{
		"X-MTrace-Token": []string{"demo-token"},
		"Last-Event-ID":  []string{"0"},
	}, 200*time.Millisecond)
	if strings.Contains(got, "event: backfill_truncated") {
		t.Errorf("exact-limit must not emit backfill_truncated, got %q", got)
	}
	if !strings.Contains(got, "id: 1") || !strings.Contains(got, "id: 2") {
		t.Errorf("expected both events in backfill, got %q", got)
	}
}

// TestSse_Backfill_IgnoresInvalidLastEventID pinnt: ein kaputter
// Last-Event-ID-Wert (z. B. "abc") darf den Stream nicht 4xx-en —
// EventSource-Browser senden den Header automatisch.
func TestSse_Backfill_IgnoresInvalidLastEventID(t *testing.T) {
	t.Parallel()
	broker := application.NewEventBroker()
	events := inmemory.NewEventRepository()
	srv := httptest.NewServer(newSseHandler(broker, events, time.Hour))
	t.Cleanup(srv.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/api/stream-sessions/stream", nil)
	req.Header.Set("X-MTrace-Token", "demo-token")
	req.Header.Set("Last-Event-ID", "not-a-number")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("invalid Last-Event-ID must not 4xx, got %d", resp.StatusCode)
	}
	_, _ = io.ReadAll(resp.Body)
}

// TestSse_DisconnectCleanup pinnt Spec §10a Pflichttest "Client-
// Disconnect (Server stoppt Loop und gibt Ressourcen frei)": nach
// Client-Schließen muss der Broker keinen Subscriber mehr halten,
// die Server-Goroutine darf nicht leaken.
func TestSse_DisconnectCleanup(t *testing.T) {
	broker := application.NewEventBroker()
	events := inmemory.NewEventRepository()
	srv := httptest.NewServer(newSseHandler(broker, events, time.Hour))
	t.Cleanup(srv.Close)

	ctx, cancel := context.WithCancel(context.Background())
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/api/stream-sessions/stream", nil)
	req.Header.Set("X-MTrace-Token", "demo-token")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		cancel()
		t.Fatalf("get: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		cancel()
		_ = resp.Body.Close()
		t.Fatalf("status = %d want 200", resp.StatusCode)
	}

	// Wait until the broker registered the subscriber (Subscribe is
	// async w.r.t. the HTTP handler returning headers).
	if !waitFor(100*time.Millisecond, func() bool { return broker.SubscriberCount() == 1 }) {
		cancel()
		_ = resp.Body.Close()
		t.Fatalf("subscriber not registered, count=%d", broker.SubscriberCount())
	}
	if !waitFor(100*time.Millisecond, func() bool { return sseHandlerGoroutineCount() > 0 }) {
		cancel()
		_ = resp.Body.Close()
		t.Fatalf("SSE handler goroutine not observed before disconnect")
	}

	// Client-side disconnect: cancel context, close body.
	cancel()
	_ = resp.Body.Close()

	// Broker-Cleanup und Server-Loop-Ende passieren asynchron nach
	// dem Client-Disconnect. Beide Invarianten sind relevant: der
	// Broker darf keinen Slot behalten, und der Handler darf trotz
	// Hour-Heartbeat nicht im select hängen bleiben.
	if !waitFor(500*time.Millisecond, func() bool { return broker.SubscriberCount() == 0 }) {
		t.Errorf("subscriber not cleaned up after client disconnect, count=%d", broker.SubscriberCount())
	}
	if !waitFor(500*time.Millisecond, func() bool { return sseHandlerGoroutineCount() == 0 }) {
		t.Errorf("SSE handler goroutine still running after client disconnect")
	}
}

func waitFor(timeout time.Duration, pred func() bool) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if pred() {
			return true
		}
		time.Sleep(5 * time.Millisecond)
	}
	return pred()
}

func sseHandlerGoroutineCount() int {
	const maxStackSnapshot = 1 << 20
	for size := 1 << 16; size <= maxStackSnapshot; size *= 2 {
		buf := make([]byte, size)
		n := runtime.Stack(buf, true)
		if n < len(buf) {
			return strings.Count(string(buf[:n]), "SseStreamHandler).ServeHTTP")
		}
	}
	buf := make([]byte, maxStackSnapshot)
	n := runtime.Stack(buf, true)
	return strings.Count(string(buf[:n]), "SseStreamHandler).ServeHTTP")
}

// streamBytesUntilTimeout öffnet den SSE-Endpunkt mit den
// übergebenen Headern, liest bis der Request-Context ausläuft und
// gibt die akkumulierten Bytes zurück. Timeout-driven, damit der
// Test nicht blockt, wenn der Server nach dem Backfill in den
// Live-Wartezustand geht.
func streamBytesUntilTimeout(t *testing.T, baseURL string, headers http.Header, timeout time.Duration) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/stream-sessions/stream", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	for k, vv := range headers {
		for _, v := range vv {
			req.Header.Add(k, v)
		}
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body) // returns when ctx-deadline closes the connection
	return string(body)
}
