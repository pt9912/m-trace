package http_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/auth"
	"github.com/pt9912/m-trace/apps/api/adapters/driven/metrics"
	"github.com/pt9912/m-trace/apps/api/adapters/driven/persistence/inmemory"
	"github.com/pt9912/m-trace/apps/api/adapters/driven/ratelimit"
	"github.com/pt9912/m-trace/apps/api/adapters/driven/streamanalyzer"
	apihttp "github.com/pt9912/m-trace/apps/api/adapters/driving/http"
	"github.com/pt9912/m-trace/apps/api/hexagon/application"
)

// Tests in this file cover plan-0.4.0 §3.4a "Backend-Tests Trace-
// Konsistenz". The DoD items map as follows:
//
//   §3.4a #1 Multi-Batch trace consistency           -> below: TestHTTP_Trace_MultiBatchSameSessionConsistency
//   §3.4a #2 Missing client context (no traceparent) -> below: TestHTTP_Trace_MissingTraceparent_ServerGeneratesTrace
//   §3.4a #3 Invalid client context (parse_error)    -> already covered: traceparent_span_test.go::TestHTTP_Span_TraceParent_InvalidSetsParseError
//   §3.4a #4 session_ended preserves correlation_id  -> below: TestHTTP_Trace_SessionEnded_PreservesCorrelationID
//   §3.4a #5 Time-skew span attribute                -> already covered: traceparent_span_test.go::TestHTTP_Span_TimeSkew_SetsWarning
//   §3.4a #6 NoOp-Tracer correlation_id persistence  -> below: TestHTTP_Trace_NoopTracer_CorrelationStillPersisted
//
// The bodies below intentionally avoid the shared validBody constant
// from handler_test.go so each test owns its session_id and event
// timeline.

// TestHTTP_Trace_MultiBatchSameSessionConsistency deckt §3.4a #1:
// drei aufeinanderfolgende Batches mit gleicher session_id erzeugen
// drei verschiedene trace_ids (jeder Batch eigene Trace), aber genau
// eine correlation_id über alle Events und über die Session selbst.
func TestHTTP_Trace_MultiBatchSameSessionConsistency(t *testing.T) {
	t.Parallel()

	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	defer func() { _ = tp.Shutdown(context.Background()) }()

	srv, eventRepo, sessionRepo := newTestServerWithTracerAndRepos(t, tp.Tracer("test"))

	const sessionID = "01J7K9X4Z2QHB6V3WS5R8Y4MULT"
	timestamps := []string{
		"2026-04-28T12:00:00.000Z",
		"2026-04-28T12:00:01.000Z",
		"2026-04-28T12:00:02.000Z",
	}
	for i, ts := range timestamps {
		body := singleEventBody(t, sessionID, "rebuffer_started", ts)
		resp := postEvents(t, srv, "demo-token", body)
		if resp.StatusCode != http.StatusAccepted {
			t.Fatalf("batch %d: expected 202, got %d", i, resp.StatusCode)
		}
	}

	spans := recorder.Ended()
	if len(spans) != 3 {
		t.Fatalf("expected 3 spans (one per batch), got %d", len(spans))
	}

	traceIDs := map[string]struct{}{}
	correlationIDs := map[string]struct{}{}
	spanCorrelationIDs := make([]string, len(spans))
	for i, span := range spans {
		traceIDs[span.SpanContext().TraceID().String()] = struct{}{}
		attrs := attrMap(span.Attributes())
		corr, ok := attrs["mtrace.session.correlation_id"].(string)
		if !ok || corr == "" {
			t.Fatalf("span[%d] mtrace.session.correlation_id missing or empty: got %v", i, attrs["mtrace.session.correlation_id"])
		}
		correlationIDs[corr] = struct{}{}
		spanCorrelationIDs[i] = corr
	}
	if len(traceIDs) != 3 {
		t.Errorf("expected 3 distinct trace_ids, got %d (set=%v)", len(traceIDs), traceIDs)
	}
	if len(correlationIDs) != 1 {
		t.Errorf("expected 1 shared correlation_id, got %d (set=%v)", len(correlationIDs), correlationIDs)
	}

	persisted := eventRepo.Snapshot()
	if len(persisted) != 3 {
		t.Fatalf("expected 3 persisted events, got %d", len(persisted))
	}
	for i, e := range persisted {
		if e.CorrelationID == "" {
			t.Errorf("event[%d] CorrelationID empty", i)
		}
		if e.CorrelationID != persisted[0].CorrelationID {
			t.Errorf("event[%d] CorrelationID = %q, want %q (shared across batches)",
				i, e.CorrelationID, persisted[0].CorrelationID)
		}
		// SF#2: jeder Batch persistiert die trace_id seines Spans —
		// stellt sicher, dass „3 distinkte trace_ids in Spans" sich
		// auch in der Zeile widerspiegeln.
		spanTraceID := spans[i].SpanContext().TraceID().String()
		if e.TraceID != spanTraceID {
			t.Errorf("event[%d] TraceID = %q, want span[%d] TraceID %q",
				i, e.TraceID, i, spanTraceID)
		}
		// Anm#3: Span-Attribut mtrace.session.correlation_id matcht
		// die persistierte CorrelationID — schließt die Lücke
		// Wire-Attribut ↔ DB-Spalte.
		if spanCorrelationIDs[i] != e.CorrelationID {
			t.Errorf("span[%d] correlation_id %q != event[%d] CorrelationID %q",
				i, spanCorrelationIDs[i], i, e.CorrelationID)
		}
	}

	sessions := sessionRepo.Snapshot()
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
	if sessions[0].CorrelationID != persisted[0].CorrelationID {
		t.Errorf("session.CorrelationID = %q, want %q",
			sessions[0].CorrelationID, persisted[0].CorrelationID)
	}
}

// TestHTTP_Trace_MissingTraceparent_ServerGeneratesTrace deckt
// §3.4a #2: ein Batch ohne traceparent-Header → der Server erzeugt
// eine eigene trace_id (Span-Root), mtrace.trace.parse_error ist nicht
// gesetzt, und die server-generierte trace_id landet im persistierten
// Event (Voraussetzung für die Tempo-Korrelation aus
// telemetry-model.md §2.5).
func TestHTTP_Trace_MissingTraceparent_ServerGeneratesTrace(t *testing.T) {
	t.Parallel()

	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	defer func() { _ = tp.Shutdown(context.Background()) }()

	srv, eventRepo, _ := newTestServerWithTracerAndRepos(t, tp.Tracer("test"))

	resp := postWithHeaders(t, srv, "demo-token", validBody, nil)
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", resp.StatusCode)
	}

	spans := recorder.Ended()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	span := spans[0]
	traceID := span.SpanContext().TraceID()
	if !traceID.IsValid() {
		t.Errorf("server-generated trace_id is invalid (all-zero) — server must mint a TraceID when no client header is present")
	}
	if span.Parent().IsValid() {
		t.Errorf("expected root span (no remote parent) for missing traceparent, got parent %s", span.Parent().SpanID())
	}

	attrs := attrMap(span.Attributes())
	if _, present := attrs["mtrace.trace.parse_error"]; present {
		t.Error("mtrace.trace.parse_error must NOT be set when no traceparent was sent")
	}

	persisted := eventRepo.Snapshot()
	if len(persisted) != 1 {
		t.Fatalf("expected 1 persisted event, got %d", len(persisted))
	}
	if got, want := persisted[0].TraceID, traceID.String(); got != want {
		t.Errorf("persisted event.TraceID = %q, want server-span trace_id %q", got, want)
	}
}

// TestHTTP_Trace_SessionEnded_PreservesCorrelationID deckt §3.4a #4:
// ein session_ended-Event innerhalb eines Batches behält die
// correlation_id der Session bei; ein nachfolgender Batch mit derselben
// session_id (auch nach session_ended) erhält dieselbe correlation_id —
// die Session bleibt im Repository auffindbar, der Use-Case liest die
// existing CorrelationID via Get + reuse-Pfad in resolveCorrelationIDs.
//
// "Reihenfolge ist Tranche-1-Verhalten" (Plan §3.4a #4): wir prüfen
// hier nur die correlation_id-Konsistenz, nicht das Session-State-
// Modell — letzteres ist durch §2.x abgedeckt.
func TestHTTP_Trace_SessionEnded_PreservesCorrelationID(t *testing.T) {
	t.Parallel()

	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	defer func() { _ = tp.Shutdown(context.Background()) }()

	srv, eventRepo, sessionRepo := newTestServerWithTracerAndRepos(t, tp.Tracer("test"))

	const sessionID = "01J7K9X4Z2QHB6V3WS5R8Y4ENDD"

	// Batch 1: erstes Event begründet die Session und ihre correlation_id.
	resp1 := postEvents(t, srv, "demo-token",
		singleEventBody(t, sessionID, "rebuffer_started", "2026-04-28T12:00:00.000Z"))
	if resp1.StatusCode != http.StatusAccepted {
		t.Fatalf("batch 1: expected 202, got %d", resp1.StatusCode)
	}

	// Batch 2: session_ended schließt die Session; muss die existing
	// correlation_id übernehmen, nicht eine neue erzeugen.
	resp2 := postEvents(t, srv, "demo-token",
		singleEventBody(t, sessionID, "session_ended", "2026-04-28T12:00:01.000Z"))
	if resp2.StatusCode != http.StatusAccepted {
		t.Fatalf("batch 2 (session_ended): expected 202, got %d", resp2.StatusCode)
	}

	// Batch 3: Nachzügler mit derselben session_id bekommt weiter die
	// gleiche correlation_id — das Repository hat die Session noch.
	resp3 := postEvents(t, srv, "demo-token",
		singleEventBody(t, sessionID, "rebuffer_ended", "2026-04-28T12:00:02.000Z"))
	if resp3.StatusCode != http.StatusAccepted {
		t.Fatalf("batch 3 (post-end): expected 202, got %d", resp3.StatusCode)
	}

	persisted := eventRepo.Snapshot()
	if len(persisted) != 3 {
		t.Fatalf("expected 3 persisted events, got %d", len(persisted))
	}

	// SF#4 Spot-Check: das mittlere Event ist tatsächlich session_ended,
	// nicht still gedroppt — sonst wäre der „session_ended preserves
	// correlation_id"-Vertrag des Tests gar nicht ausgeübt.
	if got, want := persisted[1].EventName, "session_ended"; got != want {
		t.Errorf("persisted[1].EventName = %q, want %q (Test-Setup-Sanity)", got, want)
	}

	first := persisted[0].CorrelationID
	if first == "" {
		t.Fatal("first event CorrelationID is empty")
	}
	for i, e := range persisted {
		if e.CorrelationID != first {
			t.Errorf("event[%d] (%s) CorrelationID = %q, want %q",
				i, e.EventName, e.CorrelationID, first)
		}
	}

	sessions := sessionRepo.Snapshot()
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
	if sessions[0].CorrelationID != first {
		t.Errorf("session.CorrelationID = %q, want %q", sessions[0].CorrelationID, first)
	}
}

// TestHTTP_Trace_NoopTracer_CorrelationStillPersisted deckt §3.4a #6:
// mit einem NoOp-Tracer (hier via `nil` an `NewRouter`, der den
// `tracenoop.NewTracerProvider().Tracer("noop")`-Fallback in
// router.go aktiviert) bleiben correlation_id-Resolver und -Persistenz
// funktionsfähig. Der Test belegt: das Dashboard kann eine Timeline
// rein aus der lokalen Persistenz aufbauen, auch wenn kein
// OTel-Exporter läuft. Spans sind dabei nicht beobachtbar (NoOp),
// daher prüfen wir ausschließlich die persistierten Events und
// Sessions. (Der frühere Name "TempoDeactivated" suggerierte einen
// Test der Config-Resolution; tatsächlich wird hier der NoOp-Pfad
// unter Beweis gestellt.)
func TestHTTP_Trace_NoopTracer_CorrelationStillPersisted(t *testing.T) {
	t.Parallel()

	srv, eventRepo, sessionRepo := newTestServerWithTracerAndRepos(t, nil)

	resp := postEvents(t, srv, "demo-token", validBody)
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", resp.StatusCode)
	}

	persisted := eventRepo.Snapshot()
	if len(persisted) != 1 {
		t.Fatalf("expected 1 persisted event, got %d", len(persisted))
	}
	if persisted[0].CorrelationID == "" {
		t.Errorf("persisted event has empty CorrelationID — Use-Case-Resolver muss unabhängig vom Tracer arbeiten")
	}

	sessions := sessionRepo.Snapshot()
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
	if sessions[0].CorrelationID != persisted[0].CorrelationID {
		t.Errorf("session.CorrelationID = %q, want %q",
			sessions[0].CorrelationID, persisted[0].CorrelationID)
	}
}

// newTestServerWithTracerAndRepos verdrahtet denselben Router wie
// newTestServerWithTracerProvider, gibt aber Event- und
// Session-Repositories mit zurück, damit §3.4a-Tests den persistierten
// Zustand nach einem Batch inspizieren können. Ein nil-Tracer wird im
// Router zu einem NoOp-Tracer expandiert (siehe router.go).
func newTestServerWithTracerAndRepos(
	t *testing.T,
	tracer trace.Tracer,
) (*httptest.Server, *inmemory.EventRepository, *inmemory.SessionRepository) {
	t.Helper()
	repo := inmemory.NewEventRepository()
	resolver := auth.NewStaticProjectResolver(map[string]auth.ProjectConfig{
		"demo": {Token: "demo-token", AllowedOrigins: []string{"http://localhost:5173"}},
	})
	limiter := ratelimit.NewTokenBucketRateLimiter(100, 100, time.Now)
	publisher := metrics.NewPrometheusPublisher()
	sessionRepo := inmemory.NewSessionRepository()
	uc := application.NewRegisterPlaybackEventBatchUseCase(
		resolver, limiter, repo, sessionRepo, publisher, noopTelemetry{},
		streamanalyzer.NewNoopStreamAnalyzer(), inmemory.NewIngestSequencer(), time.Now,
	)
	sessionsService := application.NewSessionsService(sessionRepo, repo)
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	router := apihttp.NewRouter(uc, sessionsService, nil, resolver,
		publisher.Handler(), nil, tracer, logger)
	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)
	return srv, repo, sessionRepo
}

// singleEventBody baut ein Wire-Payload mit genau einem Event für die
// Multi-Batch- und Session-Ende-Tests. project_id und sdk-Block
// kopieren das Format aus validBody (handler_test.go), damit der
// Wire-Validator zufrieden ist. Marshal-Fehler eskalieren über
// `t.Fatalf`, nicht über `panic`, damit ein Regression im Marshaller
// als sauberes Test-Failure auftaucht.
func singleEventBody(t *testing.T, sessionID, eventName, clientTimestamp string) string {
	t.Helper()
	type evt struct {
		EventName       string `json:"event_name"`
		ProjectID       string `json:"project_id"`
		SessionID       string `json:"session_id"`
		ClientTimestamp string `json:"client_timestamp"`
		SDK             struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"sdk"`
	}
	type batch struct {
		SchemaVersion string `json:"schema_version"`
		Events        []evt  `json:"events"`
	}
	e := evt{
		EventName:       eventName,
		ProjectID:       "demo",
		SessionID:       sessionID,
		ClientTimestamp: clientTimestamp,
	}
	e.SDK.Name = "@npm9912/player-sdk"
	e.SDK.Version = "0.2.0"
	b, err := json.Marshal(batch{SchemaVersion: "1.0", Events: []evt{e}})
	if err != nil {
		t.Fatalf("marshal singleEventBody: %v", err)
	}
	return string(b)
}
