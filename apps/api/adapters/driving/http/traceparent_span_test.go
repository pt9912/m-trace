package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// uuidv4Pattern matcht eine kanonische UUIDv4-Form (8-4-4-4-12 mit
// Versions-/Variant-Bits) — das Format, das newCorrelationID erzeugt.
var uuidv4Pattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

// TestHTTP_Span_TraceParent_ValidPropagates verifiziert: ein gültiger
// `traceparent`-Header überträgt die Trace-ID; der Server-Span ist
// Child der vom Client gestarteten Trace; `mtrace.trace.parse_error`
// ist nicht gesetzt.
func TestHTTP_Span_TraceParent_ValidPropagates(t *testing.T) {
	t.Parallel()

	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	defer func() { _ = tp.Shutdown(context.Background()) }()

	srv := newTestServerWithTracerProvider(t, tp)

	const incomingTraceID = "0af7651916cd43dd8448eb211c80319c"
	const incomingParentID = "b7ad6b7169203331"
	tp00 := "00-" + incomingTraceID + "-" + incomingParentID + "-01"

	resp := postWithHeaders(t, srv, "demo-token", validBody, map[string]string{
		"traceparent": tp00,
	})
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", resp.StatusCode)
	}

	spans := recorder.Ended()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	span := spans[0]
	if got := span.SpanContext().TraceID().String(); got != incomingTraceID {
		t.Errorf("trace_id = %q, want %q (Header propagation)", got, incomingTraceID)
	}
	if got := span.Parent().SpanID().String(); got != incomingParentID {
		t.Errorf("parent span_id = %q, want %q", got, incomingParentID)
	}
	attrs := attrMap(span.Attributes())
	if _, present := attrs["mtrace.trace.parse_error"]; present {
		t.Error("mtrace.trace.parse_error should not be set for valid header")
	}
}

// TestHTTP_Span_TraceParent_InvalidSetsParseError verifiziert: ein
// kaputter `traceparent`-Header → Span hat
// `mtrace.trace.parse_error=true`, Server-Root-Span (kein Parent), und
// die Antwort bleibt 202 (kein 4xx, telemetry-model.md §2.5).
func TestHTTP_Span_TraceParent_InvalidSetsParseError(t *testing.T) {
	t.Parallel()

	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	defer func() { _ = tp.Shutdown(context.Background()) }()

	srv := newTestServerWithTracerProvider(t, tp)

	resp := postWithHeaders(t, srv, "demo-token", validBody, map[string]string{
		"traceparent": "this-is-not-a-w3c-traceparent",
	})
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("expected 202 (parse error must not 4xx), got %d", resp.StatusCode)
	}

	spans := recorder.Ended()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	attrs := attrMap(spans[0].Attributes())
	if got, want := attrs["mtrace.trace.parse_error"], true; got != want {
		t.Errorf("mtrace.trace.parse_error = %v, want %v", got, want)
	}
	if spans[0].Parent().IsValid() {
		t.Error("expected root span (no parent) for invalid traceparent")
	}
}

// TestHTTP_Span_SingleSessionBatch_SetsCorrelationID verifiziert,
// dass für einen Single-Session-Batch die Span-Attribute
// `mtrace.session.correlation_id` (UUIDv4-geformt),
// `mtrace.batch.session_count = 1` und `mtrace.project.id` gesetzt
// werden.
func TestHTTP_Span_SingleSessionBatch_SetsCorrelationID(t *testing.T) {
	t.Parallel()

	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	defer func() { _ = tp.Shutdown(context.Background()) }()

	srv := newTestServerWithTracerProvider(t, tp)

	resp := postWithHeaders(t, srv, "demo-token", validBody, nil)
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", resp.StatusCode)
	}

	spans := recorder.Ended()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	attrs := attrMap(spans[0].Attributes())
	if got, want := attrs["mtrace.batch.session_count"], int64(1); got != want {
		t.Errorf("mtrace.batch.session_count = %v, want %v", got, want)
	}
	corr, ok := attrs["mtrace.session.correlation_id"].(string)
	if !ok || corr == "" {
		t.Fatalf("mtrace.session.correlation_id missing or wrong type, got %v",
			attrs["mtrace.session.correlation_id"])
	}
	if !uuidv4Pattern.MatchString(corr) {
		t.Errorf("correlation_id %q is not UUIDv4-shape", corr)
	}
	if got, want := attrs["mtrace.project.id"], "demo"; got != want {
		t.Errorf("mtrace.project.id = %v, want %v", got, want)
	}
}

// TestHTTP_Span_TimeSkew_SetsWarning verifiziert, dass ein Skew
// > 60 s zwischen `client_timestamp` und Server-Empfangszeit das
// Span-Attribut `mtrace.time.skew_warning=true` setzt.
func TestHTTP_Span_TimeSkew_SetsWarning(t *testing.T) {
	t.Parallel()

	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	defer func() { _ = tp.Shutdown(context.Background()) }()

	srv := newTestServerWithTracerProvider(t, tp)

	// Skew via client_timestamp aus 2020 (Server nutzt time.Now).
	skewedBody := strings.Replace(validBody,
		`"client_timestamp": "2026-04-28T12:00:00.000Z"`,
		`"client_timestamp": "2020-01-01T00:00:00.000Z"`, 1)

	resp := postWithHeaders(t, srv, "demo-token", skewedBody, nil)
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", resp.StatusCode)
	}

	spans := recorder.Ended()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	attrs := attrMap(spans[0].Attributes())
	if got, want := attrs["mtrace.time.skew_warning"], true; got != want {
		t.Errorf("mtrace.time.skew_warning = %v, want %v", got, want)
	}
}

// postWithHeaders ist ein lokales Helper-Pendant zu postEvents
// (handler_test.go), das zusätzliche Custom-Header (z. B. traceparent)
// setzen kann.
func postWithHeaders(t *testing.T, srv *httptest.Server, token, body string, extraHeaders map[string]string) *http.Response {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(),
		http.MethodPost, srv.URL+"/api/playback-events", strings.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if token != "" {
		req.Header.Set("X-MTrace-Token", token)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	return resp
}
