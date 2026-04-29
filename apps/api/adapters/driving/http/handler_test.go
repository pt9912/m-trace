package http_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/auth"
	"github.com/pt9912/m-trace/apps/api/adapters/driven/metrics"
	"github.com/pt9912/m-trace/apps/api/adapters/driven/persistence"
	"github.com/pt9912/m-trace/apps/api/adapters/driven/ratelimit"
	"github.com/pt9912/m-trace/apps/api/adapters/driven/streamanalyzer"
	apihttp "github.com/pt9912/m-trace/apps/api/adapters/driving/http"
	"github.com/pt9912/m-trace/apps/api/hexagon/application"
)

// noopTelemetry satisfies driven.Telemetry without recording anything;
// the driven-port Telemetry is exercised in the use case tests
// (hexagon/application) and the OTel adapter tests
// (adapters/driven/telemetry).
type noopTelemetry struct{}

func (noopTelemetry) BatchReceived(_ context.Context, _ int) {}

const validBody = `{
  "schema_version": "1.0",
  "events": [
    {
      "event_name": "rebuffer_started",
      "project_id": "demo",
      "session_id": "sess-1",
      "client_timestamp": "2026-04-28T12:00:00.000Z",
      "sdk": { "name": "@m-trace/player-sdk", "version": "0.1.0" }
    }
  ]
}`

// newTestServerWithClock builds a fully wired router around an
// injectable clock. All eight Pflichttests use the wall clock; the
// rate-limit test uses a stuck clock so refill cannot interfere.
func newTestServerWithClock(t *testing.T, clock func() time.Time) *httptest.Server {
	t.Helper()
	repo := persistence.NewInMemoryEventRepository()
	resolver := auth.NewStaticProjectResolver(map[string]auth.ProjectConfig{
		"demo":  {Token: "demo-token", AllowedOrigins: []string{"http://localhost:5173"}},
		"other": {Token: "other-token", AllowedOrigins: []string{"http://other.example"}},
	})
	limiter := ratelimit.NewTokenBucketRateLimiter(100, 100, clock)
	publisher := metrics.NewPrometheusPublisher()
	sessionRepo := persistence.NewInMemorySessionRepository()
	uc := application.NewRegisterPlaybackEventBatchUseCase(
		resolver, limiter, repo, sessionRepo, publisher, noopTelemetry{}, streamanalyzer.NewNoopStreamAnalyzer(), persistence.NewInMemoryIngestSequencer(), clock,
	)
	sessionsService := application.NewSessionsService(sessionRepo, repo, "test-process")
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	router := apihttp.NewRouter(uc, sessionsService, resolver, publisher.Handler(), nil, logger)
	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)
	return srv
}

func newTestServer(t *testing.T) *httptest.Server {
	return newTestServerWithClock(t, time.Now)
}

// unlimitedLimiter always returns nil. It is the test-only adapter
// used by the TooManyEvents test, which would otherwise be masked by
// the production rate limiter (a 101-event batch can't fit in the
// 100-token bucket and would return 429 before the batch-size check
// runs). Verifying the 422 path requires bypassing that earlier gate.
type unlimitedLimiter struct{}

func (unlimitedLimiter) Allow(_ context.Context, _ string, _ int) error { return nil }

// newServerWithUnlimitedRate wires a router whose rate limiter never
// rejects, so tests can reach validation rules downstream of the
// limiter (specifically the §5 step 5 batch-size check).
func newServerWithUnlimitedRate(t *testing.T) *httptest.Server {
	t.Helper()
	repo := persistence.NewInMemoryEventRepository()
	resolver := auth.NewStaticProjectResolver(map[string]auth.ProjectConfig{
		"demo":  {Token: "demo-token", AllowedOrigins: []string{"http://localhost:5173"}},
		"other": {Token: "other-token", AllowedOrigins: []string{"http://other.example"}},
	})
	publisher := metrics.NewPrometheusPublisher()
	sessionRepo := persistence.NewInMemorySessionRepository()
	uc := application.NewRegisterPlaybackEventBatchUseCase(
		resolver, unlimitedLimiter{}, repo, sessionRepo, publisher, noopTelemetry{}, streamanalyzer.NewNoopStreamAnalyzer(), persistence.NewInMemoryIngestSequencer(), time.Now,
	)
	sessionsService := application.NewSessionsService(sessionRepo, repo, "test-process")
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	router := apihttp.NewRouter(uc, sessionsService, resolver, publisher.Handler(), nil, logger)
	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)
	return srv
}

func postEvents(t *testing.T, srv *httptest.Server, token, body string) *http.Response {
	t.Helper()
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		srv.URL+"/api/playback-events",
		strings.NewReader(body),
	)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if token != "" {
		req.Header.Set("X-MTrace-Token", token)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	return resp
}

func batchOf(n int) string {
	var sb strings.Builder
	sb.WriteString(`{"schema_version":"1.0","events":[`)
	for i := 0; i < n; i++ {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(`{"event_name":"e","project_id":"demo","session_id":"s",`)
		sb.WriteString(`"client_timestamp":"2026-04-28T12:00:00.000Z",`)
		sb.WriteString(`"sdk":{"name":"x","version":"1"}}`)
	}
	sb.WriteString(`]}`)
	return sb.String()
}

func TestHTTP_HappyPath(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	resp := postEvents(t, srv, "demo-token", validBody)
	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("expected 202, got %d", resp.StatusCode)
	}
}

func TestHTTP_400_SchemaVersion(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	body := strings.Replace(validBody, `"schema_version": "1.0"`, `"schema_version": "2.0"`, 1)
	resp := postEvents(t, srv, "demo-token", body)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHTTP_401_MissingToken(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	resp := postEvents(t, srv, "", validBody)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestHTTP_401_WrongToken(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	resp := postEvents(t, srv, "wrong-token", validBody)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestHTTP_401_ProjectMismatch(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	body := strings.Replace(validBody, `"project_id": "demo"`, `"project_id": "other"`, 1)
	resp := postEvents(t, srv, "demo-token", body)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestHTTP_413_BodyTooLarge(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	body := strings.Repeat("X", apihttp.MaxBodyBytes+1)
	resp := postEvents(t, srv, "demo-token", body)
	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("expected 413, got %d", resp.StatusCode)
	}
}

// Verifies §5 Auth-vor-Body-Reihenfolge: ein Body > 256 KB ohne
// X-MTrace-Token muss 401 liefern (Header-Check feuert zuerst), nicht
// 413. Der Pflichttest stand früher nicht in §11 — siehe
// docs/spike/backend-api-contract.md §5/§11.
func TestHTTP_401_BodyTooLarge_NoToken(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	body := strings.Repeat("X", apihttp.MaxBodyBytes+1)
	resp := postEvents(t, srv, "", body)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 (Auth-Header vor Body-Read), got %d", resp.StatusCode)
	}
}

func TestHTTP_422_MissingField(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	// drop event_name
	body := strings.Replace(validBody, `"event_name": "rebuffer_started",`, ``, 1)
	resp := postEvents(t, srv, "demo-token", body)
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", resp.StatusCode)
	}
}

func TestHTTP_422_EmptyEvents(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	resp := postEvents(t, srv, "demo-token", `{"schema_version":"1.0","events":[]}`)
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", resp.StatusCode)
	}
}

func TestHTTP_422_TooManyEvents(t *testing.T) {
	t.Parallel()
	// A 101-event batch exceeds the §5 step-5 batch-size limit (100).
	// The default rate limiter (capacity 100) would also reject it at
	// step 3 because 101 > capacity, so we use an unlimited limiter
	// here to surface the §5 step-5 path explicitly.
	srv := newServerWithUnlimitedRate(t)
	resp := postEvents(t, srv, "demo-token", batchOf(101))
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", resp.StatusCode)
	}
}

func TestHTTP_429_RateLimit(t *testing.T) {
	t.Parallel()
	// Stuck clock: bucket starts at 100 tokens, never refills. After
	// exhausting all 100 tokens with one batch the next event must hit
	// the limiter deterministically — no timing race.
	fixed := time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC)
	srv := newTestServerWithClock(t, func() time.Time { return fixed })

	// First batch consumes 100 tokens.
	resp := postEvents(t, srv, "demo-token", batchOf(100))
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("expected 202 for first batch, got %d", resp.StatusCode)
	}

	// Second request must be rate-limited with Retry-After.
	resp = postEvents(t, srv, "demo-token", validBody)
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Retry-After"); got == "" {
		t.Errorf("expected Retry-After header, got empty")
	}
}
