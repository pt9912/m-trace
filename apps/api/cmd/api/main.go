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
	"encoding/base64"
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
	envAuthLabDefault        = "MTRACE_AUTH_LAB_DEFAULT"
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
	authBundle, authErr := buildAuthSessionService(baseProjects, logger)
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

	router := apihttp.NewRouter(useCase, sessionsService, analysisService, resolver, staticResolver, publisher.Handler(), publisher, publisher, sseConfig, srtHealthInbound, ingestControlInbound, authSessionInbound, playbackAuthHeaders, tracer, logger)
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
	Inbound driving.AuthSessionInbound
	Signer  *auth.HMACSessionTokenSigner
}

// buildAuthSessionService verdrahtet den `0.12.0`-Auth-Pfad
// (RAK-72/RAK-75): Signing-Key-Ring (`MTRACE_AUTH_SIGNING_KID` /
// `MTRACE_AUTH_SIGNING_KEY` als Base64-URL-encoded Bytes — fehlend
// fällt auf einen markierten Lab-Default zurück), In-Memory-Issuance-
// Limiter (global + Project) und In-Memory-Project-Policy-Resolver
// (Fallback aus den Static-Project-Origins). Klartext-Token-Material
// wird im Resolver defensiv kopiert. Der zurückgegebene Bundle
// enthält den Issuance-Service und den Signer (für den
// PlaybackEventsHandler-Verify-Pfad).
func buildAuthSessionService(baseProjects map[string]domain.Project, logger *slog.Logger) (*authBundle, error) {
	kid := domain.SigningKeyID(strings.TrimSpace(os.Getenv(envAuthSigningKID)))
	if kid == "" {
		kid = authDefaultLabSigningKID
	}
	rawSecret := strings.TrimSpace(os.Getenv(envAuthSigningKey))
	var secret []byte
	if rawSecret == "" {
		// `0.12.0`-Hardening: ohne explizit gesetzten Signing-Key
		// muss der Operator den Lab-Default per
		// `MTRACE_AUTH_LAB_DEFAULT=1` aktivieren. Sonst hard-fail —
		// ein deterministischer, world-known HMAC-Key wäre eine
		// Production-Falle (Plan-0.12.0 §0.6 Threat Model).
		if !labDefaultOptIn() {
			return nil, fmt.Errorf(
				"%s is required (set %s=1 to opt into the lab default key, NOT for production)",
				envAuthSigningKey, envAuthLabDefault)
		}
		secret = []byte(authDefaultLabSigningKeySecret)
		logger.Warn("auth signing key falls back to lab default — set MTRACE_AUTH_SIGNING_KEY for production",
			"kid", string(kid))
	} else {
		decoded, err := base64DecodeURLSafe(rawSecret)
		if err != nil {
			return nil, fmt.Errorf("MTRACE_AUTH_SIGNING_KEY base64 decode: %w", err)
		}
		secret = decoded
	}
	now := time.Now().UTC()
	signingKey := domain.SessionSigningKey{
		KID:       kid,
		Algorithm: domain.SigningKeyAlgorithmHS256,
		Secret:    secret,
		NotBefore: now.Add(-time.Hour),
		RetiresAt: now.Add(365 * 24 * time.Hour),
	}
	keyResolver, err := auth.NewStaticSigningKeyResolver(kid, signingKey)
	if err != nil {
		return nil, fmt.Errorf("static signing key resolver: %w", err)
	}
	signer := auth.NewHMACSessionTokenSigner(keyResolver)
	limiter := auth.NewInMemoryIssuanceRateLimiter(
		authIssuanceGlobalCapacity, authIssuanceGlobalRefillPerSec,
		authIssuanceProjectCapacity, authIssuanceProjectRefillPerSec,
	)
	policies, err := auth.NewInMemoryProjectPolicyResolver(nil, baseProjects)
	if err != nil {
		return nil, fmt.Errorf("project policy resolver: %w", err)
	}
	ids := auth.NewRandomTokenIDGenerator()
	return &authBundle{
		Inbound: application.NewIssueSessionTokenService(policies, limiter, signer, ids),
		Signer:  signer,
	}, nil
}

// base64DecodeURLSafe akzeptiert sowohl `RawURLEncoding` als auch
// `URLEncoding` mit Padding — robust gegenüber den beiden in der
// Praxis vorkommenden Varianten.
func base64DecodeURLSafe(s string) ([]byte, error) {
	if decoded, err := base64.RawURLEncoding.DecodeString(s); err == nil {
		return decoded, nil
	}
	return base64.URLEncoding.DecodeString(s)
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
