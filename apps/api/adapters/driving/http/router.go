package http

import (
	"log/slog"
	"net/http"

	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"

	"github.com/pt9912/m-trace/apps/api/hexagon/application"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
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
// CORS-Preflight ( H4).
type SseStreamConfig struct {
	Broker *application.EventBroker
	Events driven.EventRepository
}

// NewRouter wires the pflicht-Endpoints onto a single mux. Method
// routing uses Go 1.22 method-aware patterns ("POST /path"), so
// non-matching methods fall through to a 404 from the mux.
//
// Tracer wraps POST /api/playback-events in a request-span
// (spec/architecture.md — der HTTP-Adapter ist neben
// adapters/driven/telemetry der einzige Ort mit OTel-Imports). A nil
// tracer falls back to a no-op tracer so tests can wire the router
// without an OTel SDK setup.
//
// allowlist liefert die globale Union der Allowed-Origins für die
// CORS-Preflight-Handler (Variante B). nil
// deaktiviert den CORS-Pfad — alle Preflights werden dann mit `403`
// abgelehnt; der `Vary`-Header bleibt trotzdem auf jeder Antwort.
//
// sseConfig aktiviert die SSE-Route ( H4); `nil`
// deaktiviert sie für Tests, die den Stream nicht brauchen.
func NewRouter(
	useCase driving.PlaybackEventInbound,
	sessions driving.SessionsInbound,
	analysis driving.StreamAnalysisInbound,
	resolver driven.ProjectResolver,
	allowlist OriginAllowlist,
	metricsHandler http.Handler,
	analyzeMetrics AnalyzeMetrics,
	preflightMetrics PreflightMetrics,
	sseConfig *SseStreamConfig,
	srtHealth SrtHealthInbound,
	ingestControl driving.IngestControlInbound,
	authSession driving.AuthSessionInbound,
	playbackAuthHeaders *AuthHeaderParser,
	browserIngestPolicies BrowserIngestPolicies,
	originLimiter driven.OriginRateLimiter,
	trustForwardedFor bool,
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
		UseCase:     useCase,
		AuthHeaders: playbackAuthHeaders,
		Tracer:      tracer,
		Logger:      logger,
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

	//  / R-22: Origin-/IP-Rate-Limiter sitzt vor
	// `POST /api/playback-events` und `POST /api/auth/session-tokens`.
	// `nil`-Limiter (Disabled-Pfad) → Middleware ist No-Op und kein
	// Wrap.
	mux.Handle("POST /api/playback-events",
		originRateLimitMiddleware(playback, originLimiter, trustForwardedFor, logger))
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
		mux.HandleFunc("OPTIONS /api/stream-sessions/stream", ssePreflightHandler(allowlist, preflightMetrics))
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
		mux.HandleFunc("OPTIONS /api/analyze", analyzePreflightHandler(allowlist, preflightMetrics))
	}

	registerSrtHealthRoutes(mux, srtHealth, resolver, allowlist, preflightMetrics, tracer, logger)
	registerIngestControlRoutes(mux, ingestControl, resolver, allowlist, browserIngestPolicies, preflightMetrics, logger)
	registerAuthSessionRoutes(mux, authSession, resolver, allowlist, preflightMetrics, originLimiter, trustForwardedFor, logger)

	// CORS-Preflight-Handler — Player-SDK-Pfad (POST + OPTIONS) und
	// Dashboard-Lese-Pfad (GET + OPTIONS).
	mux.HandleFunc("OPTIONS /api/playback-events", playerSDKPreflightHandler(allowlist, preflightMetrics))
	mux.HandleFunc("OPTIONS /api/stream-sessions", dashboardPreflightHandler(allowlist, preflightMetrics))
	mux.HandleFunc("OPTIONS /api/stream-sessions/{id}", dashboardPreflightHandler(allowlist, preflightMetrics))
	mux.HandleFunc("OPTIONS /api/health", dashboardPreflightHandler(allowlist, preflightMetrics))

	return corsMiddleware(mux, allowlist)
}

// registerSrtHealthRoutes verdrahtet die SRT-Health-Read-Pfade
// (RAK-43, spec/backend-api-contract.md
// §7a). Wenn `srtHealth` nil ist (Sub-3.5 verdrahtet das opt-in über
// `MTRACE_SRT_SOURCE_URL`), bleibt die Funktion no-op.
func registerSrtHealthRoutes(
	mux *http.ServeMux,
	srtHealth SrtHealthInbound,
	resolver driven.ProjectResolver,
	allowlist OriginAllowlist,
	preflightMetrics PreflightMetrics,
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
	mux.HandleFunc("OPTIONS /api/srt/health", dashboardPreflightHandler(allowlist, preflightMetrics))
	mux.HandleFunc("OPTIONS /api/srt/health/{stream_id}", dashboardPreflightHandler(allowlist, preflightMetrics))
}

// registerIngestControlRoutes verdrahtet die Ingest-Control-Pfade
// (NF-13 / RAK-65..RAK-70). `nil`-Use-Case
// deaktiviert den Pfad — der Router registriert dann keine
// `/api/ingest/*`-Routen, was Tests (z. B. ohne SQLite-Volume) und
// alte Compose-Stände abdeckt.
func registerIngestControlRoutes(
	mux *http.ServeMux,
	ingest driving.IngestControlInbound,
	resolver driven.ProjectResolver,
	allowlist OriginAllowlist,
	browserPolicies BrowserIngestPolicies,
	preflightMetrics PreflightMetrics,
	logger *slog.Logger,
) {
	if ingest == nil {
		return
	}
	// `0.12.5`/RAK-80: wenn mindestens ein Project eine aktivierte
	// `BrowserIngestPolicy` hat, läuft der Preflight gegen die
	// Browser-Ingest-Allowlist (Project-Policy) und POST-Pfade
	// werden durch `browserIngestEnforcement` Middleware-gefiltert.
	// Ohne aktivierte Policy bleibt der RAK-74-Scope-Cut-Pfad
	// (dashboardPreflight, globale konservative Allowlist).
	preflight := dashboardPreflightHandler(allowlist, preflightMetrics)
	wrap := func(h http.Handler) http.Handler { return h }
	if browserPolicies != nil {
		preflight = browserIngestPreflightHandler(browserPolicies, preflightMetrics)
		wrap = browserIngestEnforcement(BrowserIngestEnforcementConfig{
			Projects: resolver,
			Policies: browserPolicies,
			Logger:   logger,
		})
	}
	collection := &IngestStreamHandler{UseCase: ingest, Resolver: resolver, Logger: logger}
	detail := &IngestStreamDetailHandler{UseCase: ingest, Resolver: resolver, Logger: logger}
	rotate := &IngestStreamRotateHandler{UseCase: ingest, Resolver: resolver, Logger: logger}
	validate := &IngestStreamValidateHandler{UseCase: ingest, Resolver: resolver, Logger: logger}
	mediaConfig := &IngestMediaServerConfigHandler{UseCase: ingest, Resolver: resolver, Logger: logger}
	hookStarted := &IngestLifecycleHookHandler{
		UseCase: ingest, Resolver: resolver, Logger: logger,
		Kind: domain.StreamLifecycleEventStarted,
	}
	hookEnded := &IngestLifecycleHookHandler{
		UseCase: ingest, Resolver: resolver, Logger: logger,
		Kind: domain.StreamLifecycleEventEnded,
	}
	// `0.12.5`/RAK-81 (R-14): MediaMTX-Auth-Bridge. Endpoint nutzt
	// MediaMTX-`externalAuth`-Form-Format, **kein** Browser-Ingest-
	// Pfad — kein wrap, kein OPTIONS-Preflight. Netzwerk-Isolation
	// ist Operator-Verantwortung (siehe `auth.md` §5.7).
	authHook := &MediaMTXAuthHookHandler{UseCase: ingest, Logger: logger}
	mux.Handle("POST /api/ingest/streams", wrap(collection))
	mux.Handle("GET /api/ingest/streams", collection)
	mux.Handle("GET /api/ingest/streams/{id}", detail)
	mux.Handle("POST /api/ingest/streams/{id}/rotate-key", wrap(rotate))
	mux.Handle("POST /api/ingest/streams/{id}/validate-key", wrap(validate))
	mux.Handle("GET /api/ingest/media-server-config", mediaConfig)
	mux.Handle("POST /api/ingest/hooks/stream-started", wrap(hookStarted))
	mux.Handle("POST /api/ingest/hooks/stream-ended", wrap(hookEnded))
	mux.Handle("POST /api/ingest/auth-hook", authHook)
	mux.HandleFunc("OPTIONS /api/ingest/streams", preflight)
	mux.HandleFunc("OPTIONS /api/ingest/streams/{id}", preflight)
	mux.HandleFunc("OPTIONS /api/ingest/streams/{id}/rotate-key", preflight)
	mux.HandleFunc("OPTIONS /api/ingest/streams/{id}/validate-key", preflight)
	mux.HandleFunc("OPTIONS /api/ingest/media-server-config", preflight)
	mux.HandleFunc("OPTIONS /api/ingest/hooks/stream-started", preflight)
	mux.HandleFunc("OPTIONS /api/ingest/hooks/stream-ended", preflight)
}

// registerAuthSessionRoutes verdrahtet den Session-Token-Issuance-
// Pfad (RAK-72). `nil`-Use-Case deaktiviert den Pfad —
// `POST /api/auth/session-tokens` antwortet dann mit `404` und alte
// Compose-Stände bzw. Tests ohne Auth-Setup bleiben unverändert.
func registerAuthSessionRoutes(
	mux *http.ServeMux,
	authSession driving.AuthSessionInbound,
	resolver driven.ProjectResolver,
	allowlist OriginAllowlist,
	preflightMetrics PreflightMetrics,
	originLimiter driven.OriginRateLimiter,
	trustForwardedFor bool,
	logger *slog.Logger,
) {
	if authSession == nil {
		return
	}
	handler := &AuthSessionTokensHandler{
		UseCase:  authSession,
		Resolver: resolver,
		Logger:   logger,
	}
	//  / R-22: Origin-Limiter vor dem
	// Issuance-Pfad. `nil`-Limiter ist No-Op.
	mux.Handle("POST /api/auth/session-tokens",
		originRateLimitMiddleware(handler, originLimiter, trustForwardedFor, logger))
	mux.HandleFunc("OPTIONS /api/auth/session-tokens", playerSDKPreflightHandler(allowlist, preflightMetrics))
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
