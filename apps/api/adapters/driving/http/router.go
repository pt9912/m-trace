package http

import (
	"log/slog"
	"net/http"

	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"

	"github.com/pt9912/m-trace/apps/api/hexagon/application"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driving"
)

// RequestMetrics is the small metrics surface the HTTP adapter needs
// for aggregate request counting.
type RequestMetrics interface {
	APIRequests(n int)
}

// SseStreamConfig bündelt die Driven-Ports, die der SSE-Handler aus
// dem Hexagon braucht. `nil` deaktiviert die SSE-Route — der Router
// registriert dann weder `GET /api/stream-sessions/stream` noch den
// CORS-Preflight (plan-0.4.0 §5 H4).
type SseStreamConfig struct {
	Broker *application.EventBroker
	Events driven.EventRepository
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
//
// sseConfig aktiviert die SSE-Route (plan-0.4.0 §5 H4); `nil`
// deaktiviert sie für Tests, die den Stream nicht brauchen.
func NewRouter(
	useCase driving.PlaybackEventInbound,
	sessions driving.SessionsInbound,
	analysis driving.StreamAnalysisInbound,
	resolver driven.ProjectResolver,
	allowlist OriginAllowlist,
	metricsHandler http.Handler,
	analyzeMetrics AnalyzeMetrics,
	sseConfig *SseStreamConfig,
	srtHealth SrtHealthInbound,
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
		UseCase:  sessions,
		Resolver: resolver,
		Tracer:   tracer,
		Logger:   logger,
	}
	sessionsGet := &SessionsGetHandler{
		UseCase:  sessions,
		Resolver: resolver,
		Tracer:   tracer,
		Logger:   logger,
	}

	mux.Handle("POST /api/playback-events", playback)
	mux.HandleFunc("GET /api/health", HealthHandler)
	mux.Handle("GET /api/metrics", metricsHandler)
	mux.Handle("GET /api/stream-sessions", sessionsList)
	// SSE-Route VOR der Catch-all-`{id}`-Pattern registrieren, damit
	// `/api/stream-sessions/stream` nicht als `id="stream"` an den
	// Detail-Handler routet (Go 1.22 method+path-Patterns sind
	// präfix-spezifisch).
	if sseConfig != nil && sseConfig.Broker != nil && sseConfig.Events != nil {
		sseHandler := &SseStreamHandler{
			Resolver: resolver,
			Events:   sseConfig.Events,
			Broker:   sseConfig.Broker,
			Logger:   logger,
		}
		mux.Handle("GET /api/stream-sessions/stream", sseHandler)
		mux.HandleFunc("OPTIONS /api/stream-sessions/stream", ssePreflightHandler(allowlist))
	}
	mux.Handle("GET /api/stream-sessions/{id}", sessionsGet)

	if analysis != nil {
		analyzeHandler := &AnalyzeHandler{
			UseCase:  analysis,
			Resolver: resolver,
			Logger:   logger,
			Metrics:  analyzeMetrics,
		}
		mux.Handle("POST /api/analyze", analyzeHandler)
		mux.HandleFunc("OPTIONS /api/analyze", analyzePreflightHandler(allowlist))
	}

	registerSrtHealthRoutes(mux, srtHealth, resolver, allowlist, tracer, logger)

	// CORS-Preflight-Handler — Player-SDK-Pfad (POST + OPTIONS) und
	// Dashboard-Lese-Pfad (GET + OPTIONS). plan-0.1.0.md §5.1.
	mux.HandleFunc("OPTIONS /api/playback-events", playerSDKPreflightHandler(allowlist))
	mux.HandleFunc("OPTIONS /api/stream-sessions", dashboardPreflightHandler(allowlist))
	mux.HandleFunc("OPTIONS /api/stream-sessions/{id}", dashboardPreflightHandler(allowlist))
	mux.HandleFunc("OPTIONS /api/health", dashboardPreflightHandler(allowlist))

	return corsMiddleware(mux, allowlist)
}

// registerSrtHealthRoutes verdrahtet die SRT-Health-Read-Pfade
// (plan-0.6.0 §5 Tranche 4 — RAK-43, spec/backend-api-contract.md
// §7a). Wenn `srtHealth` nil ist (Sub-3.5 verdrahtet das opt-in über
// `MTRACE_SRT_SOURCE_URL`), bleibt die Funktion no-op.
func registerSrtHealthRoutes(
	mux *http.ServeMux,
	srtHealth SrtHealthInbound,
	resolver driven.ProjectResolver,
	allowlist OriginAllowlist,
	tracer trace.Tracer,
	logger *slog.Logger,
) {
	if srtHealth == nil {
		return
	}
	listHandler := &SrtHealthListHandler{
		UseCase:  srtHealth,
		Resolver: resolver,
		Tracer:   tracer,
		Logger:   logger,
	}
	getHandler := &SrtHealthGetHandler{
		UseCase:  srtHealth,
		Resolver: resolver,
		Tracer:   tracer,
		Logger:   logger,
	}
	mux.Handle("GET /api/srt/health", listHandler)
	mux.Handle("GET /api/srt/health/{stream_id}", getHandler)
	mux.HandleFunc("OPTIONS /api/srt/health", dashboardPreflightHandler(allowlist))
	mux.HandleFunc("OPTIONS /api/srt/health/{stream_id}", dashboardPreflightHandler(allowlist))
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
