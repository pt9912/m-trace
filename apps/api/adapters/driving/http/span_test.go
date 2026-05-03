package http_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/auth"
	"github.com/pt9912/m-trace/apps/api/adapters/driven/metrics"
	"github.com/pt9912/m-trace/apps/api/adapters/driven/persistence/inmemory"
	"github.com/pt9912/m-trace/apps/api/adapters/driven/ratelimit"
	"github.com/pt9912/m-trace/apps/api/adapters/driven/streamanalyzer"
	apihttp "github.com/pt9912/m-trace/apps/api/adapters/driving/http"
	"github.com/pt9912/m-trace/apps/api/hexagon/application"
)

// TestHTTP_Span_HappyPathAttributes verifies that POST /api/playback-events
// emits exactly one span with the attributes documented in
// spec/telemetry-model.md §2.1: http.method, http.route,
// http.status_code, batch.size and batch.outcome=accepted on the 202
// path. SpanRecorder is used in lieu of an OTel-SDK exporter.
func TestHTTP_Span_HappyPathAttributes(t *testing.T) {
	t.Parallel()

	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	defer func() { _ = tp.Shutdown(context.Background()) }()

	srv := newTestServerWithTracerProvider(t, tp)

	resp := postEvents(t, srv, "demo-token", validBody)
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", resp.StatusCode)
	}

	spans := recorder.Ended()
	if len(spans) != 1 {
		t.Fatalf("expected exactly 1 span, got %d", len(spans))
	}
	span := spans[0]

	if got, want := span.Name(), "http.handler POST /api/playback-events"; got != want {
		t.Errorf("span name: got %q want %q", got, want)
	}
	if got := span.Status().Code; got != codes.Ok {
		t.Errorf("span status: got %v want Ok", got)
	}

	attrs := attrMap(span.Attributes())
	if got := attrs["http.method"]; got != "POST" {
		t.Errorf("http.method=%v want POST", got)
	}
	if got := attrs["http.route"]; got != "/api/playback-events" {
		t.Errorf("http.route=%v want /api/playback-events", got)
	}
	if got := attrs["http.status_code"]; got != int64(http.StatusAccepted) {
		t.Errorf("http.status_code=%v want 202", got)
	}
	if got := attrs["batch.size"]; got != int64(1) {
		t.Errorf("batch.size=%v want 1", got)
	}
	if got := attrs["batch.outcome"]; got != "accepted" {
		t.Errorf("batch.outcome=%v want accepted", got)
	}
}

// TestHTTP_Span_AuthRejectMissingTokenAttributes verifies that the
// fast-reject path (401, no body parsed) still emits one span with
// http.status_code=401 and batch.outcome=unauthorized. batch.size
// must be absent because the JSON was never parsed.
func TestHTTP_Span_AuthRejectMissingTokenAttributes(t *testing.T) {
	t.Parallel()

	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	defer func() { _ = tp.Shutdown(context.Background()) }()

	srv := newTestServerWithTracerProvider(t, tp)

	resp := postEvents(t, srv, "", validBody)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}

	spans := recorder.Ended()
	if len(spans) != 1 {
		t.Fatalf("expected exactly 1 span, got %d", len(spans))
	}
	span := spans[0]

	attrs := attrMap(span.Attributes())
	if got := attrs["http.status_code"]; got != int64(http.StatusUnauthorized) {
		t.Errorf("http.status_code=%v want 401", got)
	}
	if got := attrs["batch.outcome"]; got != "unauthorized" {
		t.Errorf("batch.outcome=%v want unauthorized", got)
	}
	if _, ok := attrs["batch.size"]; ok {
		t.Errorf("batch.size unexpectedly present on pre-parse reject")
	}
}

// newTestServerWithTracerProvider wires the same router as
// newTestServerWithClock but with an explicit tracer derived from the
// caller's TracerProvider so the span attributes can be inspected.
func newTestServerWithTracerProvider(t *testing.T, tp *sdktrace.TracerProvider) *httptest.Server {
	t.Helper()
	repo := inmemory.NewEventRepository()
	resolver := auth.NewStaticProjectResolver(map[string]auth.ProjectConfig{
		"demo": {Token: "demo-token", AllowedOrigins: []string{"http://localhost:5173"}},
	})
	limiter := ratelimit.NewTokenBucketRateLimiter(100, 100, time.Now)
	publisher := metrics.NewPrometheusPublisher()
	sessionRepo := inmemory.NewSessionRepository()
	uc := application.NewRegisterPlaybackEventBatchUseCase(
		resolver, limiter, repo, sessionRepo, publisher, noopTelemetry{}, streamanalyzer.NewNoopStreamAnalyzer(), inmemory.NewIngestSequencer(), time.Now,
	)
	sessionsService := application.NewSessionsService(sessionRepo, repo)
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	tracer := tp.Tracer("test")
	router := apihttp.NewRouter(uc, sessionsService, nil, resolver, resolver, publisher.Handler(), nil, tracer, logger)
	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)
	return srv
}

// attrMap turns an OTel attribute slice into a value-typed map for
// terse `got != want` assertions in the span tests.
func attrMap(kvs []attribute.KeyValue) map[string]any {
	out := make(map[string]any, len(kvs))
	for _, kv := range kvs {
		switch kv.Value.Type() {
		case attribute.STRING:
			out[string(kv.Key)] = kv.Value.AsString()
		case attribute.INT64:
			out[string(kv.Key)] = kv.Value.AsInt64()
		case attribute.BOOL:
			out[string(kv.Key)] = kv.Value.AsBool()
		case attribute.FLOAT64:
			out[string(kv.Key)] = kv.Value.AsFloat64()
		default:
			out[string(kv.Key)] = kv.Value.Emit()
		}
	}
	return out
}

