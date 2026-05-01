package http_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
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

// TestNewRouter_NilAllowlistRejectsAllPreflights deckt den Fallback-
// Pfad in NewRouter ab: allowlist=nil → noopAllowlist; Preflights
// werden für jeden Origin abgelehnt (`403`). Das verifiziert die
// dokumentierte „kein CORS"-Semantik (router.go), ohne zusätzliche
// Test-Server-Konfiguration zu brauchen.
func TestNewRouter_NilAllowlistRejectsAllPreflights(t *testing.T) {
	t.Parallel()
	repo := persistence.NewInMemoryEventRepository()
	resolver := auth.NewStaticProjectResolver(map[string]auth.ProjectConfig{
		"demo": {Token: "demo-token", AllowedOrigins: []string{"http://localhost:5173"}},
	})
	limiter := ratelimit.NewTokenBucketRateLimiter(100, 100, time.Now)
	publisher := metrics.NewPrometheusPublisher()
	sessionRepo := persistence.NewInMemorySessionRepository()
	uc := application.NewRegisterPlaybackEventBatchUseCase(
		resolver, limiter, repo, sessionRepo, publisher,
		noopTelemetry{}, streamanalyzer.NewNoopStreamAnalyzer(),
		persistence.NewInMemoryIngestSequencer(), time.Now,
	)
	sessionsService := application.NewSessionsService(sessionRepo, repo, "test-process")
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	// allowlist=nil → noopAllowlist greift.
	router := apihttp.NewRouter(uc, sessionsService, nil, nil, publisher.Handler(), nil, logger)
	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodOptions, srv.URL+"/api/playback-events", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("noopAllowlist must reject every preflight; got %d want 403", resp.StatusCode)
	}
}

type countingRequestMetrics struct {
	requests int
}

func (m *countingRequestMetrics) APIRequests(n int) {
	m.requests += n
}

func TestRequestMetricsMiddlewareCountsEveryRequest(t *testing.T) {
	t.Parallel()
	counter := &countingRequestMetrics{}
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	handler := apihttp.RequestMetricsMiddleware(next, counter)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status=%d want 204", rec.Code)
	}
	if counter.requests != 1 {
		t.Errorf("expected APIRequests=1, got %d", counter.requests)
	}
}
