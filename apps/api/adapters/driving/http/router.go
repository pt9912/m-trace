package http

import (
	"log/slog"
	"net/http"

	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"

	"github.com/pt9912/m-trace/apps/api/hexagon/port/driving"
)

// RequestMetrics is the small metrics surface the HTTP adapter needs
// for aggregate request counting.
type RequestMetrics interface {
	APIRequests(n int)
}

// NewRouter wires the pflicht-Endpoints onto a single mux. Method
// routing uses Go 1.22 method-aware patterns ("POST /path"), so
// non-matching methods fall through to a 404 from the mux.
//
// Tracer wraps POST /api/playback-events in a request-span
// (spec/architecture.md §5.3 — der HTTP-Adapter ist neben
// adapters/driven/telemetry der einzige Ort mit OTel-Imports). A nil
// tracer falls back to a no-op tracer so tests can wire the router
// without an OTel SDK setup.
//
// allowlist liefert die globale Union der Allowed-Origins für die
// CORS-Preflight-Handler (plan-0.1.0.md §5.1, Variante B). nil
// deaktiviert den CORS-Pfad — alle Preflights werden dann mit `403`
// abgelehnt; der `Vary`-Header bleibt trotzdem auf jeder Antwort.
func NewRouter(
	useCase driving.PlaybackEventInbound,
	sessions driving.SessionsInbound,
	analysis driving.StreamAnalysisInbound,
	allowlist OriginAllowlist,
	metricsHandler http.Handler,
	tracer trace.Tracer,
	logger *slog.Logger,
) http.Handler {
	if tracer == nil {
		tracer = tracenoop.NewTracerProvider().Tracer("noop")
	}
	if allowlist == nil {
		allowlist = noopAllowlist{}
	}

	mux := http.NewServeMux()

	playback := &PlaybackEventsHandler{
		UseCase: useCase,
		Tracer:  tracer,
		Logger:  logger,
	}
	sessionsList := &SessionsListHandler{
		UseCase: sessions,
		Tracer:  tracer,
		Logger:  logger,
	}
	sessionsGet := &SessionsGetHandler{
		UseCase: sessions,
		Tracer:  tracer,
		Logger:  logger,
	}

	mux.Handle("POST /api/playback-events", playback)
	mux.HandleFunc("GET /api/health", HealthHandler)
	mux.Handle("GET /api/metrics", metricsHandler)
	mux.Handle("GET /api/stream-sessions", sessionsList)
	mux.Handle("GET /api/stream-sessions/{id}", sessionsGet)

	if analysis != nil {
		analyzeHandler := &AnalyzeHandler{UseCase: analysis, Logger: logger}
		mux.Handle("POST /api/analyze", analyzeHandler)
		mux.HandleFunc("OPTIONS /api/analyze", dashboardPreflightHandler(allowlist))
	}

	// CORS-Preflight-Handler — Player-SDK-Pfad (POST + OPTIONS) und
	// Dashboard-Lese-Pfad (GET + OPTIONS). plan-0.1.0.md §5.1.
	mux.HandleFunc("OPTIONS /api/playback-events", playerSDKPreflightHandler(allowlist))
	mux.HandleFunc("OPTIONS /api/stream-sessions", dashboardPreflightHandler(allowlist))
	mux.HandleFunc("OPTIONS /api/stream-sessions/{id}", dashboardPreflightHandler(allowlist))
	mux.HandleFunc("OPTIONS /api/health", dashboardPreflightHandler(allowlist))

	return corsMiddleware(mux, allowlist)
}

// RequestMetricsMiddleware counts every HTTP request that enters the
// API router. It intentionally emits no labels to keep Prometheus
// cardinality bounded.
func RequestMetricsMiddleware(next http.Handler, metrics RequestMetrics) http.Handler {
	if metrics == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metrics.APIRequests(1)
		next.ServeHTTP(w, r)
	})
}

// noopAllowlist lehnt jeden Origin ab — Fallback für nil-Allowlist
// (Test-Server ohne CORS-Setup).
type noopAllowlist struct{}

func (noopAllowlist) IsOriginInGlobalUnion(_ string) bool { return false }
