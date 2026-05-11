// Package main is the entry point of the m-trace API spike (Go variant).
//
// Wires the driven adapters (auth, persistence, ratelimit, metrics,
// telemetry) into the application use case and exposes the three
// pflicht endpoints (POST /api/playback-events, GET /api/health,
// GET /api/metrics) over HTTP. See docs/spike/0001-backend-stack.md
// and docs/planning/done/plan-spike.md for scope.
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/auth"
	"github.com/pt9912/m-trace/apps/api/adapters/driven/metrics"
	"github.com/pt9912/m-trace/apps/api/adapters/driven/persistence/inmemory"
	persistencesqlite "github.com/pt9912/m-trace/apps/api/adapters/driven/persistence/sqlite"
	"github.com/pt9912/m-trace/apps/api/adapters/driven/ratelimit"
	"github.com/pt9912/m-trace/apps/api/adapters/driven/srt/mediamtxclient"
	"github.com/pt9912/m-trace/apps/api/adapters/driven/streamanalyzer"
	"github.com/pt9912/m-trace/apps/api/adapters/driven/telemetry"
	"github.com/pt9912/m-trace/apps/api/adapters/driven/webhooks"
	apihttp "github.com/pt9912/m-trace/apps/api/adapters/driving/http"
	"github.com/pt9912/m-trace/apps/api/hexagon/application"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driving"
	"github.com/pt9912/m-trace/apps/api/internal/storage"
)

const (
	envSrtSourceURL          = "MTRACE_SRT_SOURCE_URL"
	envSrtSourceUser         = "MTRACE_SRT_SOURCE_USER"
	envSrtSourcePass         = "MTRACE_SRT_SOURCE_PASS"
	envSrtPollInterval       = "MTRACE_SRT_POLL_INTERVAL_SECONDS"
	envSrtProjectID          = "MTRACE_SRT_PROJECT_ID"
	envSrtRequiredBandwidth  = "MTRACE_SRT_REQUIRED_BANDWIDTH_BPS"
	envAuthSigningKID        = "MTRACE_AUTH_SIGNING_KID"
	envAuthSigningKey        = "MTRACE_AUTH_SIGNING_KEY"
	envAuthSigningKeys       = "MTRACE_AUTH_SIGNING_KEYS"
	envAuthSigningActiveKID  = "MTRACE_AUTH_SIGNING_ACTIVE_KID"
	envAuthLabDefault        = "MTRACE_AUTH_LAB_DEFAULT"
	envAuthIssuanceLimiter   = "MTRACE_AUTH_ISSUANCE_LIMITER"
	envAuthSecretBackend     = "MTRACE_AUTH_SECRET_BACKEND"
	envOutboundWebhookURL    = "MTRACE_OUTBOUND_WEBHOOK_URL"
	envOutboundWebhookSecret = "MTRACE_OUTBOUND_WEBHOOK_SECRET"
)

// Auth-/Token-Lifecycle Default-Limits (`0.12.0`, RAK-72). Ein
// produktives Setup soll diese Werte über künftige Env-Vars
// überschreiben können — der Spike pinnt sichere Lab-Defaults.
const (
	authIssuanceGlobalCapacity      = 100
	authIssuanceGlobalRefillPerSec  = 10.0
	authIssuanceProjectCapacity     = 30
	authIssuanceProjectRefillPerSec = 5.0
	authDefaultLabSigningKeySecret  = "mtrace-lab-only-do-not-use-in-production-replace-via-env"
	authDefaultLabSigningKID        = "lab-default"
)

const (
	serviceName       = "m-trace-api"
	serviceVersion    = "0.12.1"
	defaultListenAddr = ":8080"

	// Spike Spec §6.9: 100 events/sec/project.
	rateLimitCapacity = 100
	rateLimitRefill   = 100.0

	// Persistenz-Konfiguration (ADR-0002 §8.1, plan-0.4.0 §2.4):
	// Default ab 0.4.0 ist SQLite; In-Memory bleibt opt-in für Tests
	// oder expliziten Dev-Fallback.
	persistenceModeSQLite   = "sqlite"
	persistenceModeInMemory = "inmemory"
	defaultPersistenceMode  = persistenceModeSQLite
	defaultSQLitePath       = "/var/lib/mtrace/m-trace.db"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	if err := run(logger); err != nil {
		logger.Error("api shutdown with error", "error", err)
		os.Exit(1)
	}
}

// run hält die gesamte Bootstrap- und Shutdown-Logik zusammen, damit
// `defer`-Cleanups (insbesondere SQLite-DB-Close) auch im Fehler-Pfad
// garantiert ausgeführt werden — ein direktes `os.Exit` aus dem Body
// würde sie überspringen.
func run(logger *slog.Logger) error {
	otelProviders, err := telemetry.Setup(context.Background(), serviceName, serviceVersion)
	if err != nil {
		return fmt.Errorf("otel setup: %w", err)
	}

	persist, err := newPersistence(context.Background(), logger)
	if err != nil {
		return fmt.Errorf("persistence init: %w", err)
	}
	defer persist.Close()

	handler, sweeper, publisher, otelTelemetry, err := buildHandler(persist, otelProviders, logger)
	if err != nil {
		return err
	}

	srtCollector := buildSrtHealthCollector(persist, publisher, otelTelemetry, logger)

	srv := newHTTPServer(handler, listenAddr())
	return serve(srv, sweeper, srtCollector, otelProviders, logger)
}

// buildSrtHealthCollector verdrahtet den SRT-Health-Pfad
// (plan-0.6.0 §4 Sub-3.5/3.6). Wenn `MTRACE_SRT_SOURCE_URL` leer
// ist, bleibt der Collector deaktiviert (nil) — der Default-Lab-Pfad
// wird damit nicht durch fehlende ENV-Variablen blockiert.
//
// Sub-3.6 verdrahtet zusätzlich den geteilten PrometheusPublisher
// (für `mtrace_srt_health_*`-Aggregate) und den OTel-Telemetry-Adapter
// (`mtrace.srt.health.collect`-Spans).
func buildSrtHealthCollector(
	persist *persistenceBundle,
	publisher driven.MetricsPublisher,
	otelTelemetry driven.Telemetry,
	logger *slog.Logger,
) *application.SrtHealthCollector {
	baseURL := strings.TrimSpace(os.Getenv(envSrtSourceURL))
	if baseURL == "" {
		logger.Info("srt-health collector disabled (MTRACE_SRT_SOURCE_URL not set)")
		return nil
	}
	if persist.db == nil {
		logger.Warn("srt-health collector disabled (persistence is in-memory; SQLite required)")
		return nil
	}
	projectID := strings.TrimSpace(os.Getenv(envSrtProjectID))
	if projectID == "" {
		projectID = "demo"
	}
	user := os.Getenv(envSrtSourceUser)
	pass := os.Getenv(envSrtSourcePass)
	requiredBandwidth := parseSrtRequiredBandwidth(logger)

	sourceOpts := []mediamtxclient.Option{
		mediamtxclient.WithBasicAuth(user, pass),
	}
	if requiredBandwidth > 0 {
		sourceOpts = append(sourceOpts, mediamtxclient.WithRequiredBandwidthBPS(requiredBandwidth))
	}
	source := mediamtxclient.New(baseURL, sourceOpts...)
	repo := persistencesqlite.NewSrtHealthRepository(persist.db)

	collector, err := application.NewSrtHealthCollector(
		source, repo, projectID, time.Now, application.DefaultThresholds(),
	)
	if err != nil {
		logger.Error("srt-health collector init failed", "error", err)
		return nil
	}
	collector.
		WithLogger(logger).
		WithPollInterval(parseSrtPollInterval(logger)).
		WithMetrics(publisher).
		WithTelemetry(otelTelemetry)
	logger.Info(
		"srt-health collector enabled",
		"source_url", baseURL,
		"project_id", projectID,
		"auth", user != "" || pass != "",
		"required_bandwidth_bps", requiredBandwidth,
	)
	return collector
}

// parseSrtRequiredBandwidth liest `MTRACE_SRT_REQUIRED_BANDWIDTH_BPS`.
// Ohne ENV oder bei ungültigem/non-positivem Wert wird 0 zurück-
// gegeben — der Adapter setzt `RequiredBandwidthBPS` dann nicht und
// die Health-Bewertung wertet die Bandbreite gemäß spec/telemetry-
// model.md §7.4 nur an, ohne sie zu bewerten.
func parseSrtRequiredBandwidth(logger *slog.Logger) int64 {
	raw := strings.TrimSpace(os.Getenv(envSrtRequiredBandwidth))
	if raw == "" {
		return 0
	}
	bps, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || bps <= 0 {
		logger.Warn(
			"srt-health required bandwidth ignored (invalid)",
			"raw", raw,
			"hint", "set MTRACE_SRT_REQUIRED_BANDWIDTH_BPS to a positive bit/s value",
		)
		return 0
	}
	return bps
}

// parseSrtPollInterval liest `MTRACE_SRT_POLL_INTERVAL_SECONDS`. Bei
// fehlendem oder ungültigem Wert bleibt der Default aus
// application.DefaultSrtHealthPollInterval gültig.
func parseSrtPollInterval(logger *slog.Logger) time.Duration {
	raw := strings.TrimSpace(os.Getenv(envSrtPollInterval))
	if raw == "" {
		return application.DefaultSrtHealthPollInterval
	}
	secs, err := time.ParseDuration(raw + "s")
	if err != nil || secs <= 0 {
		logger.Warn(
			"srt-health poll interval ignored (invalid)",
			"raw", raw,
			"default_seconds", int(application.DefaultSrtHealthPollInterval.Seconds()),
		)
		return application.DefaultSrtHealthPollInterval
	}
	return secs
}

// buildHandler wirt die driven Adapter (Auth, Rate-Limit, Metrics,
// Telemetry, Analyzer) mit den persistierten Repos zusammen, baut
// die drei Use Cases und liefert den fertig konfigurierten HTTP-
// Handler plus den Sessions-Sweeper, dessen Lifecycle der Caller
// (run → serve) gegen den Signal-Kontext bindet.
func buildHandler(
	persist *persistenceBundle,
	otelProviders *telemetry.Providers,
	logger *slog.Logger,
) (http.Handler, *application.SessionsSweeper, *metrics.PrometheusPublisher, *telemetry.OTelTelemetry, error) {
	otelTelemetry, err := telemetry.NewOTelTelemetry(otelProviders.Meter.Meter(telemetry.MeterName))
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("otel telemetry adapter init: %w", err)
	}

	projectConfigs := map[string]auth.ProjectConfig{
		"demo": {
			Token: "demo-token",
			AllowedOrigins: []string{
				"http://localhost:5173",
				"http://localhost:3000",
				"http://dashboard-e2e:5173",
			},
		},
	}
	staticResolver := auth.NewStaticProjectResolver(projectConfigs)
	baseProjects := make(map[string]domain.Project, len(projectConfigs))
	for projectID, cfg := range projectConfigs {
		baseProjects[projectID] = domain.Project{
			ID:             projectID,
			Token:          domain.ProjectToken(cfg.Token),
			AllowedOrigins: append([]string(nil), cfg.AllowedOrigins...),
		}
	}
	// plan-0.12.0 Tranche 3 (RAK-73): Wenn die Persistenz SQLite hält,
	// wickeln wir den Static-Resolver in einen RotatingProjectResolver
	// ein, der `mtr_pt_*`-Tokens über `project_token_generations`
	// auflöst und sonst auf den Static-Pfad fällt. InMemory-Modus
	// behält den reinen Static-Resolver.
	var (
		resolver           driven.ProjectResolver = staticResolver
		projectTokenRepo   driven.ProjectTokenRepository
	)
	if persist.db != nil {
		projectTokenRepo = persistencesqlite.NewProjectTokenRepository(persist.db)
		resolver = auth.NewRotatingProjectResolver(projectTokenRepo, staticResolver, staticResolver)
	}
	limiter := ratelimit.NewTokenBucketRateLimiter(rateLimitCapacity, rateLimitRefill, time.Now)
	publisher := metrics.NewPrometheusPublisher(metrics.WithActiveSessionsFunc(activeSessionsGauge(persist.sessions, logger)))
	analyzer := newAnalyzer(logger)

	broker := application.NewEventBroker()
	useCase := application.NewRegisterPlaybackEventBatchUseCase(
		resolver, limiter, persist.events, persist.sessions, publisher, otelTelemetry, analyzer, persist.sequencer, time.Now,
	).WithBroker(broker)
	sessionsService := application.NewSessionsService(persist.sessions, persist.events)
	sessionsSweeper := application.NewSessionsSweeper(persist.sessions, time.Now, logger)
	analysisService := application.NewAnalyzeManifestUseCase(analyzer, persist.sessions)

	tracer := otelProviders.Tracer.Tracer(telemetry.TracerName)
	sseConfig := &apihttp.SseStreamConfig{Broker: broker, Events: persist.events}
	srtHealthService, err := application.NewSrtHealthQueryService(persistencesqlite.NewSrtHealthRepository(persist.db), time.Now, application.DefaultThresholds())
	if err != nil {
		// Persist ist InMemory → kein durable SRT-Health-Storage; Read-
		// Pfad bleibt deaktiviert. Logger-Notice in Sub-3.5 hat das schon
		// geloggt; hier nur Service als nil weiterreichen.
		srtHealthService = nil
	}
	var srtHealthInbound apihttp.SrtHealthInbound
	if srtHealthService != nil {
		srtHealthInbound = srtHealthService
	}

	// plan-0.11.0 Tranche 2: Ingest-Control-Pfad nur dann verdrahten,
	// wenn die Persistenz SQLite hält (durable SQLite-Repo). InMemory-
	// Lab-Modus liefert `nil` → der Router lässt `/api/ingest/*`
	// deaktiviert (404), was für Spike-/CLI-Smoke-Aufrufe okay ist.
	var ingestControlService *application.IngestControlService
	if persist.db != nil {
		ingestRepo := persistencesqlite.NewIngestStreamRepository(persist.db)
		ingestControlService = application.NewIngestControlService(ingestRepo, time.Now)
		if wh := buildOutboundWebhookDispatcher(logger); wh != nil {
			ingestControlService = ingestControlService.WithOutboundWebhookDispatcher(wh)
		}
	}
	var ingestControlInbound driving.IngestControlInbound
	if ingestControlService != nil {
		ingestControlInbound = ingestControlService
	}

	// plan-0.12.0 Tranche 2: Session-Token-Issuance verdrahten. Der
	// Spike nutzt einen Default-Signing-Key aus
	// `MTRACE_AUTH_SIGNING_KEY` (Base64-URL); ohne Env-Var wird ein
	// deterministischer Lab-Key benutzt und der Logger warnt einmal,
	// damit Production-Setups nicht mit dem Lab-Key in Betrieb gehen.
	authBundle, authErr := buildAuthSessionService(baseProjects, persist.db, logger)
	if authErr != nil {
		logger.Warn("auth session service disabled", "error", authErr.Error())
	}
	var (
		authSessionInbound  driving.AuthSessionInbound
		playbackAuthHeaders *apihttp.AuthHeaderParser
	)
	if authBundle != nil {
		authSessionInbound = authBundle.Inbound
		playbackAuthHeaders = &apihttp.AuthHeaderParser{
			Resolver: resolver,
			Verifier: authBundle.Signer,
			Projects: staticResolver,
			Audience: domain.SessionTokenAudiencePlaybackEvents,
		}
	}

	var browserIngestPolicies apihttp.BrowserIngestPolicies
	if authBundle != nil && authBundle.PolicyResolver != nil {
		browserIngestPolicies = authBundle.PolicyResolver
	}
	router := apihttp.NewRouter(useCase, sessionsService, analysisService, resolver, staticResolver, publisher.Handler(), publisher, publisher, sseConfig, srtHealthInbound, ingestControlInbound, authSessionInbound, playbackAuthHeaders, browserIngestPolicies, tracer, logger)
	return apihttp.RequestMetricsMiddleware(router, publisher), sessionsSweeper, publisher, otelTelemetry, nil
}

// activeSessionsGauge liefert den Gauge-Provider für
// `mtrace_active_sessions`. On-demand-Lookup im Prometheus-Scrape-
// Pfad, mit 2-Sekunden-Timeout und Error-to-0-Mapping, damit der
// Gauge auch bei Backend-Aussetzern lesbar bleibt.
func activeSessionsGauge(sessions driven.SessionRepository, logger *slog.Logger) func() float64 {
	return func() float64 {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		n, err := sessions.CountByState(ctx, domain.SessionStateActive)
		if err != nil {
			logger.Warn("active sessions count failed", "error", err)
			return 0
		}
		return float64(n)
	}
}

// newHTTPServer setzt die Pflicht-Timeouts (ReadHeader/Read/Write/Idle)
// für den API-HTTP-Server.
func newHTTPServer(handler http.Handler, addr string) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

// serve startet den Sessions-Sweeper plus optional den SRT-Health-
// Collector und den HTTP-Server und führt den Graceful-Shutdown aus.
// Beendet entweder bei SIGINT/SIGTERM oder wenn ListenAndServe einen
// non-ErrServerClosed-Fehler liefert.
func serve(
	srv *http.Server,
	sweeper *application.SessionsSweeper,
	srtCollector *application.SrtHealthCollector,
	otelProviders *telemetry.Providers,
	logger *slog.Logger,
) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go sweeper.Run(ctx)
	if srtCollector != nil {
		go srtCollector.Run(ctx)
	}

	listenErr := make(chan error, 1)
	go func() {
		logger.Info("server starting",
			"addr", srv.Addr,
			"service", serviceName,
			"version", serviceVersion,
		)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			listenErr <- err
			return
		}
		listenErr <- nil
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case err := <-listenErr:
		if err != nil {
			return fmt.Errorf("server: %w", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	if err := otelProviders.Shutdown(shutdownCtx); err != nil {
		logger.Error("otel shutdown failed", "error", err)
	}
	logger.Info("server stopped")
	return nil
}

func listenAddr() string {
	if addr := strings.TrimSpace(os.Getenv("MTRACE_API_LISTEN_ADDR")); addr != "" {
		return addr
	}
	return defaultListenAddr
}

// newAnalyzer wählt zwischen dem Noop-Slot und dem HTTP-Adapter
// gegen den analyzer-service (plan-0.3.0 §7 Tranche 6). Setzt der
// Operator `ANALYZER_BASE_URL`, wird der HTTP-Adapter aktiv; sonst
// bleibt es beim Noop, damit lokale Smokes ohne Begleitservice
// laufen können.
func newAnalyzer(logger *slog.Logger) driven.StreamAnalyzer {
	baseURL := strings.TrimSpace(os.Getenv("ANALYZER_BASE_URL"))
	if baseURL == "" {
		logger.Info("analyzer adapter: noop (ANALYZER_BASE_URL nicht gesetzt)")
		return streamanalyzer.NewNoopStreamAnalyzer()
	}
	logger.Info("analyzer adapter: http", "base_url", baseURL)
	return streamanalyzer.NewHTTPStreamAnalyzer(baseURL)
}

// persistenceBundle bündelt die drei Driven-Adapter, die der Use Case
// braucht, plus einen optionalen Close-Handle für das Backend (SQLite-
// DB schließt im Shutdown-Pfad).
type persistenceBundle struct {
	events    driven.EventRepository
	sessions  driven.SessionRepository
	sequencer driven.IngestSequencer
	db        *sql.DB // nil im InMemory-Modus
}

// Close schließt die zugrundeliegende DB, falls vorhanden. No-op für
// InMemory.
func (p *persistenceBundle) Close() {
	if p.db != nil {
		_ = p.db.Close()
	}
}

// newPersistence wählt den Persistenz-Adapter über die env vars
// `MTRACE_PERSISTENCE` (Default: `sqlite`) und `MTRACE_SQLITE_PATH`
// (Default: `/var/lib/mtrace/m-trace.db`). Im SQLite-Modus öffnet die
// Funktion die DB via internal/storage (führt Migrationen aus) und
// initialisiert den Sequencer aus `MAX(ingest_sequence)`.
func newPersistence(ctx context.Context, logger *slog.Logger) (*persistenceBundle, error) {
	mode := strings.TrimSpace(strings.ToLower(os.Getenv("MTRACE_PERSISTENCE")))
	if mode == "" {
		mode = defaultPersistenceMode
	}
	switch mode {
	case persistenceModeInMemory:
		logger.Info("persistence: in-memory (data does not survive restart)")
		return &persistenceBundle{
			events:    inmemory.NewEventRepository(),
			sessions:  inmemory.NewSessionRepository(),
			sequencer: inmemory.NewIngestSequencer(),
		}, nil
	case persistenceModeSQLite:
		path := strings.TrimSpace(os.Getenv("MTRACE_SQLITE_PATH"))
		if path == "" {
			path = defaultSQLitePath
		}
		logger.Info("persistence: sqlite", "path", path)
		db, err := storage.Open(ctx, path)
		if err != nil {
			return nil, fmt.Errorf("storage.Open: %w", err)
		}
		seq, err := persistencesqlite.NewIngestSequencer(ctx, db)
		if err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("ingest sequencer: %w", err)
		}
		return &persistenceBundle{
			events:    persistencesqlite.NewEventRepository(db),
			sessions:  persistencesqlite.NewSessionRepository(db),
			sequencer: seq,
			db:        db,
		}, nil
	default:
		return nil, fmt.Errorf("unknown MTRACE_PERSISTENCE=%q (expected 'sqlite' or 'inmemory')", mode)
	}
}

// authBundle bündelt das, was main.go für `0.12.0` Tranche 2 baut:
// Issuance-Service (Driving-Port) plus den Signer für den
// Konsum-Pfad (PlaybackEventsHandler verifiziert damit Bearer-/
// X-MTrace-Session-Token-Header).
type authBundle struct {
	Inbound        driving.AuthSessionInbound
	Signer         *auth.HMACSessionTokenSigner
	PolicyResolver *auth.InMemoryProjectPolicyResolver
}

// buildAuthSessionService verdrahtet den Auth-Pfad
// (`0.12.0` RAK-72/RAK-75 + `0.12.5` RAK-78): Signing-Key-Ring,
// In-Memory-Issuance-Limiter (global + Project) und
// In-Memory-Project-Policy-Resolver (Fallback aus Static-Project-
// Origins).
//
// Signing-Key-Ring kommt aus zwei alternativen ENV-Pfaden — Parser-
// Logik in `auth.ParseSigningKeysEnv`:
//   - **Multi-Key (`0.12.5`)**: `MTRACE_AUTH_SIGNING_KEYS=
//     kid_a:b64[,kid_b:b64,…]` plus `MTRACE_AUTH_SIGNING_ACTIVE_KID`.
//     Mehrere Keys verifizieren parallel; nur der aktive `kid`
//     signiert (RAK-78). Operator-Workflow siehe `auth.md` §5.3.1.
//   - **Single-Key (Backwards-Compat zu `0.12.0`)**:
//     `MTRACE_AUTH_SIGNING_KEY` plus optional `MTRACE_AUTH_SIGNING_KID`.
//     Degenerierter `len(keys)==1`-Resolver.
//
// Backend-Auswahl per `MTRACE_AUTH_SECRET_BACKEND` (`0.12.5` RAK-79):
//   - `env` (Default): liest aus den ENV-Variablen wie oben.
//   - `vault`: Vault KV-v2-Pfad über `MTRACE_AUTH_VAULT_*`.
//
// Ist beim `env`-Backend keiner der ENV-Pfade gesetzt, fällt das
// Setup auf den markierten Lab-Default zurück
// (`MTRACE_AUTH_LAB_DEFAULT=1` als Opt-in, sonst hard-fail). Bei
// `vault` (oder anderen externen Backends) gibt es **kein** Lab-
// Default — ein nicht erreichbares Backend ist immer ein Fehler
// (fail-closed). Klartext-Token-Material wird im Resolver defensiv
// kopiert.
func buildAuthSessionService(baseProjects map[string]domain.Project, db *sql.DB, logger *slog.Logger) (*authBundle, error) {
	now := time.Now().UTC()
	backend, backendName, err := buildAuthSecretBackend(logger)
	if err != nil {
		return nil, err
	}
	keys, activeKID, err := backend.LoadSigningKeys(context.Background())
	switch {
	case errors.Is(err, auth.ErrNoSecretConfigured):
		// Lab-Default-Fallback ist ausschließlich für das ENV-Backend
		// sinnvoll — externe Backends, die explizit ausgewählt wurden,
		// dürfen nicht stillschweigend einen lokalen Lab-Key benutzen.
		if backendName != "env" {
			return nil, fmt.Errorf(
				"auth secret backend %q reported no signing keys configured; check %s/%s/%s",
				backendName, "MTRACE_AUTH_VAULT_ADDR", "MTRACE_AUTH_VAULT_TOKEN", "MTRACE_AUTH_VAULT_PATH")
		}
		if !labDefaultOptIn() {
			return nil, fmt.Errorf(
				"%s or %s is required (set %s=1 to opt into the lab default key, NOT for production)",
				envAuthSigningKeys, envAuthSigningKey, envAuthLabDefault)
		}
		labKID := activeKID
		if labKID == "" {
			labKID = authDefaultLabSigningKID
		}
		keys = []domain.SessionSigningKey{
			{
				KID:       labKID,
				Algorithm: domain.SigningKeyAlgorithmHS256,
				Secret:    []byte(authDefaultLabSigningKeySecret),
				NotBefore: now.Add(-time.Hour),
				RetiresAt: now.Add(365 * 24 * time.Hour),
			},
		}
		activeKID = labKID
		logger.Warn("auth signing key falls back to lab default — set MTRACE_AUTH_SIGNING_KEYS (or MTRACE_AUTH_SIGNING_KEY) for production",
			"kid", string(activeKID))
	case err != nil:
		return nil, fmt.Errorf("auth secret backend %q load: %w", backendName, err)
	}
	keyResolver, err := auth.NewMultiKeySigningResolver(activeKID, keys...)
	if err != nil {
		return nil, fmt.Errorf("multi-key signing resolver: %w", err)
	}
	if len(keys) > 1 {
		// Operator-sichtbarer Log-Hinweis, dass der Multi-Key-Pfad
		// aktiv ist — hilft beim Rotation-Smoke (RAK-78).
		verifyKIDs := make([]string, 0, len(keys))
		for _, k := range keys {
			verifyKIDs = append(verifyKIDs, string(k.KID))
		}
		logger.Info("auth multi-key signing resolver active",
			"active_kid", string(activeKID),
			"verify_kids", verifyKIDs)
	}
	signer := auth.NewHMACSessionTokenSigner(keyResolver)
	limiter, err := buildIssuanceRateLimiter(db, logger)
	if err != nil {
		return nil, err
	}
	policies, err := auth.NewInMemoryProjectPolicyResolver(nil, baseProjects)
	if err != nil {
		return nil, fmt.Errorf("project policy resolver: %w", err)
	}
	ids := auth.NewRandomTokenIDGenerator()
	return &authBundle{
		Inbound:        application.NewIssueSessionTokenService(policies, limiter, signer, ids),
		Signer:         signer,
		PolicyResolver: policies,
	}, nil
}

// buildOutboundWebhookDispatcher liest die Outbound-Webhook-
// Konfiguration aus den ENV-Variablen `MTRACE_OUTBOUND_WEBHOOK_URL`
// und `MTRACE_OUTBOUND_WEBHOOK_SECRET` (`0.12.5`/RAK-82, R-16).
// Ist keine URL gesetzt → `nil` (Adapter deaktiviert, identisch
// zum `0.11.0`-Verhalten ohne Outbound-Webhook). Sonst:
// `webhooks.NewHTTPDispatcher` mit Default-Retry/-Timeout-Werten.
func buildOutboundWebhookDispatcher(logger *slog.Logger) driven.OutboundWebhookDispatcher {
	url := strings.TrimSpace(os.Getenv(envOutboundWebhookURL))
	if url == "" {
		return nil
	}
	secret := []byte(os.Getenv(envOutboundWebhookSecret))
	logger.Info("outbound webhook dispatcher active",
		"endpoint", url,
		"signed", len(secret) > 0,
	)
	return webhooks.NewHTTPDispatcher(url, secret, logger)
}

// buildAuthSecretBackend wählt das Signing-Key-Backend
// (`0.12.5` RAK-79 / R-20). ENV-Selektor `MTRACE_AUTH_SECRET_BACKEND`:
//   - leer / `env`: In-Process-Default — liest `MTRACE_AUTH_SIGNING_KEYS`/
//     `_KEY` / `_ACTIVE_KID` / `_KID` aus dem Prozess-ENV
//     (Backwards-Compat zum `0.12.0`-Pfad inkl. Lab-Default-Opt-in).
//   - `vault`: opt-in externer Adapter über Vault KV-v2
//     (`MTRACE_AUTH_VAULT_ADDR/_TOKEN/_PATH`). Fail-closed bei
//     Outage; **kein** Lab-Default-Fallback.
//   - andere Werte (insb. `kms`): explizit nicht unterstützt — der
//     Boot-Validator failt mit klarer Fehlermeldung. KMS-Adapter
//     bleibt additive Folge-Option nach `0.12.5`.
//
// Rückgabe `backendName` wird vom Caller fürs Fehler-Wording und
// das Lab-Default-Fallback-Gate genutzt — externe Backends dürfen
// nicht stillschweigend auf einen Lab-Key fallen.
func buildAuthSecretBackend(logger *slog.Logger) (driven.AuthSecretBackend, string, error) {
	backend := strings.ToLower(strings.TrimSpace(os.Getenv(envAuthSecretBackend)))
	switch backend {
	case "", "env":
		logger.Info("auth secret backend active", "backend", "env")
		return auth.NewEnvSecretBackend(), "env", nil
	case "vault":
		vb, err := auth.NewVaultSecretBackend(os.Getenv)
		if err != nil {
			return nil, "", fmt.Errorf("%s=vault: %w", envAuthSecretBackend, err)
		}
		logger.Info("auth secret backend active", "backend", "vault",
			"note", "boot-time load, no refresh; restart for key rotation; fail-closed on outage")
		return vb, "vault", nil
	default:
		return nil, "", fmt.Errorf(
			"%s=%q is not supported (valid: env|vault; kms is a follow-up item after 0.12.5)",
			envAuthSecretBackend, backend,
		)
	}
}

// buildIssuanceRateLimiter wählt zwischen In-Process- und
// SQLite-basiertem Token-Bucket-Limiter (`0.12.5` RAK-77 / R-17).
// ENV-Selektor `MTRACE_AUTH_ISSUANCE_LIMITER`:
//   - leer / `memory`: In-Process-Default (Backwards-Compat zu
//     `0.12.0`). Misst pro Replica — passt nur für Single-Instance-
//     Setups.
//   - `sqlite`: opt-in Shared-State-Pfad. Braucht aktive SQLite-
//     Persistenz (`MTRACE_PERSISTENCE=sqlite`); fehlt sie, hard-fail.
//     Single-Host-Multi-Replica-Setups (Compose-`volumes:`,
//     K8s-`hostPath`) teilen sich den Counter. Multi-Host bleibt
//     Folge-Item.
//   - andere Werte (`redis`, `memcached`, …): explizit nicht
//     unterstützt — der Boot-Validator failt mit klarer
//     Fehlermeldung, statt still auf `memory` zu fallen.
func buildIssuanceRateLimiter(db *sql.DB, logger *slog.Logger) (driven.IssuanceRateLimiter, error) {
	backend := strings.ToLower(strings.TrimSpace(os.Getenv(envAuthIssuanceLimiter)))
	switch backend {
	case "", "memory":
		logger.Info("auth issuance limiter active", "backend", "memory")
		return auth.NewInMemoryIssuanceRateLimiter(
			authIssuanceGlobalCapacity, authIssuanceGlobalRefillPerSec,
			authIssuanceProjectCapacity, authIssuanceProjectRefillPerSec,
		), nil
	case "sqlite":
		if db == nil {
			return nil, fmt.Errorf(
				"%s=sqlite requires MTRACE_PERSISTENCE=sqlite (got nil DB handle — InMemory persistence is incompatible)",
				envAuthIssuanceLimiter,
			)
		}
		logger.Info("auth issuance limiter active", "backend", "sqlite",
			"topology_constraint", "single-host with shared volume (Compose volumes / K8s hostPath)",
		)
		return auth.NewSqliteIssuanceRateLimiter(
			db,
			authIssuanceGlobalCapacity, authIssuanceGlobalRefillPerSec,
			authIssuanceProjectCapacity, authIssuanceProjectRefillPerSec,
		), nil
	default:
		return nil, fmt.Errorf(
			"%s=%q is not supported (valid: memory|sqlite; redis/memcached are follow-up items, see plan-0.12.5 §6 / R-17 follow-ups)",
			envAuthIssuanceLimiter, backend,
		)
	}
}

// labDefaultOptIn liest `MTRACE_AUTH_LAB_DEFAULT` und akzeptiert nur
// die explizit truthy Werte `1`/`true`/`yes`. Alles andere (inklusive
// fehlend) gilt als „nicht opt-in" — der Aufrufer hard-failt dann.
func labDefaultOptIn() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(envAuthLabDefault))) {
	case "1", "true", "yes":
		return true
	default:
		return false
	}
}
