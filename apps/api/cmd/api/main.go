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

	"github.com/redis/go-redis/v9"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/auth"
	"github.com/pt9912/m-trace/apps/api/adapters/driven/mediaserver"
	"github.com/pt9912/m-trace/apps/api/adapters/driven/metrics"
	"github.com/pt9912/m-trace/apps/api/adapters/driven/persistence/inmemory"
	persistencepostgres "github.com/pt9912/m-trace/apps/api/adapters/driven/persistence/postgres"
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
	envMediaServerProvURL    = "MTRACE_MEDIASERVER_PROVISION_URL"
	envMediaServerProvToken  = "MTRACE_MEDIASERVER_PROVISION_TOKEN"
	envOriginRateLimiter     = "MTRACE_ORIGIN_RATE_LIMITER"
	envTrustForwardedFor     = "MTRACE_TRUST_FORWARDED_FOR"
	envRedisAddr             = "MTRACE_REDIS_ADDR"
	envRedisAuth             = "MTRACE_REDIS_AUTH"
	envRedisDB               = "MTRACE_REDIS_DB"
	envAuthIssuanceFailOpen  = "MTRACE_AUTH_ISSUANCE_FAIL_OPEN"
	envRateLimitCapacity     = "MTRACE_RATE_LIMIT_CAPACITY"
	envRateLimitRefill       = "MTRACE_RATE_LIMIT_REFILL"
	envRateLimitBackend      = "MTRACE_RATE_LIMIT_BACKEND"
	envRateLimitFailClosed   = "MTRACE_RATE_LIMIT_FAIL_CLOSED"
	envLabProjects           = "MTRACE_LAB_PROJECTS"
)

// Auth-/Token-Lifecycle Default-Limits (RAK-72). Ein
// produktives Setup soll diese Werte über künftige Env-Vars
// überschreiben können — der Spike pinnt sichere Lab-Defaults.
const (
	authIssuanceGlobalCapacity      = 100
	authIssuanceGlobalRefillPerSec  = 10.0
	authIssuanceProjectCapacity     = 30
	authIssuanceProjectRefillPerSec = 5.0
	authDefaultLabSigningKeySecret  = "mtrace-lab-only-do-not-use-in-production-replace-via-env"
	authDefaultLabSigningKID        = "lab-default"

	// Origin-Rate-Limiter Default-Bucket R-22). Lab-konservativ: 20 Requests Burst, 5 Refill/s ≈ 5 RPS
	// steady state pro Client-IP. Wirksam nur wenn
	// `MTRACE_ORIGIN_RATE_LIMITER=memory`.
	originRateLimitCapacity     = 20
	originRateLimitRefillPerSec = 5.0
)

const (
	serviceName       = "m-trace-api"
	serviceVersion    = "0.25.0"
	defaultListenAddr = ":8080"

	// Ingest-Rate-Limit pro Project. Default „Spike Spec: 100 events/
	// sec/project"; per MTRACE_RATE_LIMIT_CAPACITY / -REFILL
	// überschreibbar (z. B. Kapazitäts-Modus eines Last-Smoke, damit
	// der Lasttest die echte Ingest-Kapazität statt der Limiter-Decke
	// misst).
	defaultRateLimitCapacity = 100
	defaultRateLimitRefill   = 100.0

	// Persistenz-Konfiguration (ADR-0002):
	// Default ist SQLite; In-Memory bleibt opt-in für Tests
	// oder expliziten Dev-Fallback.
	persistenceModeSQLite   = "sqlite"
	persistenceModeInMemory = "inmemory"
	persistenceModePostgres = "postgres"
	defaultPersistenceMode  = persistenceModeSQLite
	defaultSQLitePath       = "/var/lib/mtrace/m-trace.db"
	envPostgresDSN          = "MTRACE_POSTGRES_DSN"
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
// ( Sub-3.5/3.6). Wenn `MTRACE_SRT_SOURCE_URL` leer
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
		logger.Warn("srt-health collector disabled (persistence is in-memory; a durable store — sqlite or postgres — is required)")
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
	repo := persist.srtHealth

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

// parseIngestRateLimit liest die Ingest-Rate-Limit-Konfiguration aus
// MTRACE_RATE_LIMIT_CAPACITY (Token-Bucket-Kapazität, int) und
// MTRACE_RATE_LIMIT_REFILL (Refill-Rate in Token/s, float). Ohne ENV
// oder bei ungültigem/non-positivem Wert bleibt der Default
// (defaultRateLimitCapacity / defaultRateLimitRefill = 100/100) gültig.
// Der Override existiert für den Kapazitäts-Modus eines Last-Smoke,
// damit der Lasttest die echte Ingest-/Persistenz-Kapazität misst statt
// nur der Limiter-Decke.
func parseIngestRateLimit(logger *slog.Logger) (int, float64) {
	capacity := defaultRateLimitCapacity
	refill := defaultRateLimitRefill
	if raw := strings.TrimSpace(os.Getenv(envRateLimitCapacity)); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n <= 0 {
			logger.Warn(
				"ingest rate limit capacity ignored (invalid)",
				"raw", raw,
				"default", defaultRateLimitCapacity,
				"hint", "set "+envRateLimitCapacity+" to a positive integer",
			)
		} else {
			capacity = n
		}
	}
	if raw := strings.TrimSpace(os.Getenv(envRateLimitRefill)); raw != "" {
		f, err := strconv.ParseFloat(raw, 64)
		if err != nil || f <= 0 {
			logger.Warn(
				"ingest rate limit refill ignored (invalid)",
				"raw", raw,
				"default", defaultRateLimitRefill,
				"hint", "set "+envRateLimitRefill+" to a positive number (tokens/s)",
			)
		} else {
			refill = f
		}
	}
	if capacity != defaultRateLimitCapacity || refill != defaultRateLimitRefill {
		logger.Info(
			"ingest rate limit overridden via env",
			"capacity", capacity,
			"refill_per_sec", refill,
		)
	}
	return capacity, refill
}

// labProjectCount (R-26 b, Multi-Tenant-Lab) liest `MTRACE_LAB_PROJECTS`.
// N > 0 seedet ZUSÄTZLICH zum `demo`-Projekt N deterministische
// Lab-Projekte `lab-1`..`lab-N` (Token `lab-token-<i>`, gleiche
// Allowed-Origins wie `demo`) — die Grundlage für den Multi-Tenant-
// Last-Smoke (Token-Fan-out über N Projekte). Ohne ENV bleibt die
// Projekt-Menge byte-identisch (nur `demo`). Nur für Lab-/Lasttest-
// Setups: die Tokens sind vorhersagbar (gleiche Klasse wie das
// hartkodierte `demo-token`), Produktions-Projekte kommen aus der
// Projekt-/Token-Verwaltung. Ungültige Werte werden mit WARN ignoriert;
// Obergrenze 256 gegen Tippfehler-Explosion.
func labProjectCount(logger *slog.Logger) int {
	raw := strings.TrimSpace(os.Getenv(envLabProjects))
	if raw == "" {
		return 0
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 || n > 256 {
		logger.Warn(
			"lab projects ignored (invalid)",
			"raw", raw,
			"hint", "set "+envLabProjects+" to an integer in 1..256",
		)
		return 0
	}
	return n
}

// labProjectSetup baut die statische Lab-Projekt-Konfiguration (`demo`
// + optionale `MTRACE_LAB_PROJECTS`-Zusatzprojekte, R-26 b) samt der
// Domain-Projektion für den Rotating-Resolver-Basisbestand.
func labProjectSetup(logger *slog.Logger) (map[string]auth.ProjectConfig, map[string]domain.Project) {
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
	seedLabProjects(projectConfigs, logger)
	baseProjects := make(map[string]domain.Project, len(projectConfigs))
	for projectID, cfg := range projectConfigs {
		baseProjects[projectID] = domain.Project{
			ID:             projectID,
			Token:          domain.ProjectToken(cfg.Token),
			AllowedOrigins: append([]string(nil), cfg.AllowedOrigins...),
		}
	}
	return projectConfigs, baseProjects
}

// seedLabProjects (R-26 b) ergänzt projectConfigs um die env-getriebenen
// Lab-Projekte `lab-1..N` — additiv zu `demo` (dessen Allowed-Origins
// übernommen werden); No-op ohne/bei ungültiger ENV.
func seedLabProjects(projectConfigs map[string]auth.ProjectConfig, logger *slog.Logger) {
	n := labProjectCount(logger)
	if n <= 0 {
		return
	}
	demoOrigins := projectConfigs["demo"].AllowedOrigins
	for i := 1; i <= n; i++ {
		projectConfigs[fmt.Sprintf("lab-%d", i)] = auth.ProjectConfig{
			Token:          fmt.Sprintf("lab-token-%d", i),
			AllowedOrigins: append([]string(nil), demoOrigins...),
		}
	}
	// Laut warnen (Parität zu MTRACE_AUTH_LAB_DEFAULT): die Tokens sind
	// vorhersagbar und gelten über baseProjects auch auf dem
	// Auth-Session-/Policy-Pfad — NICHT für Produktion.
	logger.Warn("multi-tenant lab projects seeded — predictable lab tokens, NOT for production",
		"count", n, "ids", fmt.Sprintf("lab-1..lab-%d", n),
		"env", envLabProjects)
}

// envTruthyOptIn ist DIE Truthy-Auswertung für Opt-in-ENV-Flags:
// nur explizit "1"/"true"/"yes" (case-insensitiv) schalten ein. Vor
// diesem Helper existierte der Switch als vier identische Kopien —
// eine Erweiterung der akzeptierten Tokens hätte die Flags
// inkonsistent gemacht (z. B. "on" wirkt bei einem, nicht beim anderen).
func envTruthyOptIn(name string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(name))) {
	case "1", "true", "yes":
		return true
	default:
		return false
	}
}

// rateLimitFailClosedOptIn liest `MTRACE_RATE_LIMIT_FAIL_CLOSED` und
// akzeptiert nur explizit-truthy Werte. Default ist fail-open auf den
// lokalen In-Memory-Fallback (s. buildIngestRateLimiter).
func rateLimitFailClosedOptIn() bool { return envTruthyOptIn(envRateLimitFailClosed) }

// buildIngestRateLimiter (R-26 b) wählt das Backend des Ingest-Rate-
// Limiters per ENV-Selektor `MTRACE_RATE_LIMIT_BACKEND`:
//   - leer / `memory` (Default): In-Process-Token-Bucket pro Replica —
//     unverändertes Verhalten. Über N Replicas ist die effektive
//     Per-Projekt-Decke damit N × Capacity (gemessen: budgets.md §8).
//   - `redis`: shared Token-Bucket auf dem gemeinsamen Redis-Server
//     (derselbe `MTRACE_REDIS_*`-ENV-Block wie Issuance-/Origin-Limiter,
//     eigener Key-Prefix `mtrace:ingest`) — EIN Per-Projekt-Budget über
//     alle Replicas. Fail-Mode default **fail-open** auf den lokalen
//     Memory-Fallback: BEWUSST anders als der geteilte fail-closed-
//     Schalter der Auth-Limiter (s. buildOriginRateLimiter) — Schutzgut
//     ist hier Telemetrie-Verfügbarkeit, nicht Auth-Flutung; die
//     Degradation entspricht exakt dem Verhalten vor R-26 b. Striktes
//     Verhalten opt-in via `MTRACE_RATE_LIMIT_FAIL_CLOSED=1` (Outage →
//     429). Die daraus möglichen gemischten Fail-Modi auf demselben
//     Redis sind Owner-entschieden und operator-dokumentiert.
//   - `sqlite`: NICHT unterstützt — ein Hot-Path-Bucket über Hosts
//     hinweg braucht ein Network-Backend; SQLite via Shared-Volume ist
//     nicht Multi-Host-tauglich (siehe Backend-Strategie).
//   - `memcached`: Folge-Item gemeinsam mit Issuance-/Origin-Limiter,
//     falls Operator-Bedarf nach Memcached entsteht.
func buildIngestRateLimiter(logger *slog.Logger) (driven.RateLimiter, error) {
	capacity, refill := parseIngestRateLimit(logger)
	backend := strings.ToLower(strings.TrimSpace(os.Getenv(envRateLimitBackend)))
	switch backend {
	case "", "memory":
		return ratelimit.NewTokenBucketRateLimiter(capacity, refill, time.Now), nil
	case "redis":
		client, err := buildRedisClient()
		if err != nil {
			return nil, fmt.Errorf("%s=redis: %w", envRateLimitBackend, err)
		}
		failClosed := rateLimitFailClosedOptIn()
		logger.Info("ingest rate limiter active", "backend", "redis",
			"capacity", capacity,
			"refill_per_second", refill,
			"fail_mode", failModeLabel(!failClosed),
		)
		return ratelimit.NewRedisTokenBucketRateLimiter(client, ratelimit.RedisTokenBucketConfig{
			Capacity:        capacity,
			RefillPerSecond: refill,
			FailClosed:      failClosed,
		}, logger)
	case "sqlite":
		return nil, fmt.Errorf(
			"%s=sqlite is not supported (a shared ingest bucket is not Multi-Host-safe on shared SQLite volumes)",
			envRateLimitBackend,
		)
	case "memcached":
		return nil, fmt.Errorf(
			"%s=memcached is a follow-up item — gets delivered jointly with the issuance-/origin-limiter in a future tranche to avoid backend fragmentation",
			envRateLimitBackend,
		)
	default:
		return nil, fmt.Errorf(
			"%s=%q is not supported (valid: memory|redis; memcached is a follow-up item)",
			envRateLimitBackend, backend,
		)
	}
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

	projectConfigs, baseProjects := labProjectSetup(logger)
	staticResolver := auth.NewStaticProjectResolver(projectConfigs)
	//  (RAK-73): Wenn die Persistenz SQLite hält,
	// wickeln wir den Static-Resolver in einen RotatingProjectResolver
	// ein, der `mtr_pt_*`-Tokens über `project_token_generations`
	// auflöst und sonst auf den Static-Pfad fällt. InMemory-Modus
	// behält den reinen Static-Resolver.
	var (
		resolver           driven.ProjectResolver = staticResolver
		projectTokenRepo   driven.ProjectTokenRepository
	)
	if persist.db != nil {
		projectTokenRepo = persist.projectToken
		resolver = auth.NewRotatingProjectResolver(projectTokenRepo, staticResolver, staticResolver)
	}
	limiter, err := buildIngestRateLimiter(logger)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("ingest rate limiter init: %w", err)
	}
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
	srtHealthService, err := application.NewSrtHealthQueryService(persist.srtHealth, time.Now, application.DefaultThresholds())
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

	// : Ingest-Control-Pfad nur dann verdrahten,
	// wenn die Persistenz SQLite hält (durable SQLite-Repo). InMemory-
	// Lab-Modus liefert `nil` → der Router lässt `/api/ingest/*`
	// deaktiviert (404), was für Spike-/CLI-Smoke-Aufrufe okay ist.
	var ingestControlService *application.IngestControlService
	if persist.db != nil {
		ingestRepo := persist.ingestStream
		ingestControlService = wireIngestControlService(ingestRepo, logger)
	}
	var ingestControlInbound driving.IngestControlInbound
	if ingestControlService != nil {
		ingestControlInbound = ingestControlService
	}

	// : Session-Token-Issuance verdrahten. Der
	// Spike nutzt einen Default-Signing-Key aus
	// `MTRACE_AUTH_SIGNING_KEY` (Base64-URL); ohne Env-Var wird ein
	// deterministischer Lab-Key benutzt und der Logger warnt einmal,
	// damit Production-Setups nicht mit dem Lab-Key in Betrieb gehen.
	// Nur die SQLite-DB an die Auth-Verdrahtung reichen: der einzige
	// db-Konsument dort ist der optionale SQLite-Issuance-Limiter, der
	// zwingend eine SQLite-DB braucht. In Postgres-/InMemory-Modus ist das
	// nil → der Limiter-Guard lehnt `MTRACE_AUTH_ISSUANCE_LIMITER=sqlite`
	// mit klarer Meldung ab; wie jeder Auth-Fehler wird Auth dann mit
	// Warnung deaktiviert (nicht fatal — bestehendes lab-freundliches
	// Verhalten). Postgres hat keinen Issuance-Adapter — redis/memory nutzen.
	authBundle, authErr := buildAuthSessionService(baseProjects, persist.sqliteDB(), logger)
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
	originLimiter, trustXFF, err := setupOriginRateLimiter(logger)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	router := apihttp.NewRouter(useCase, sessionsService, analysisService, resolver, staticResolver, publisher.Handler(), publisher, publisher, sseConfig, srtHealthInbound, ingestControlInbound, authSessionInbound, playbackAuthHeaders, browserIngestPolicies, originLimiter, trustXFF, tracer, logger)
	return apihttp.RequestMetricsMiddleware(router, publisher), sessionsSweeper, publisher, otelTelemetry, nil
}

// setupOriginRateLimiter konsolidiert Build + XFF-Trust-Setup in
// einen Helper, damit `buildHandler` unter dem funlen-Limit bleibt.
func setupOriginRateLimiter(logger *slog.Logger) (driven.OriginRateLimiter, bool, error) {
	limiter, err := buildOriginRateLimiter(logger)
	if err != nil {
		return nil, false, fmt.Errorf("origin rate limiter: %w", err)
	}
	trustXFF := trustForwardedForOptIn()
	if trustXFF {
		logger.Info("origin rate limiter trusts X-Forwarded-For",
			"env", envTrustForwardedFor,
			"reminder", "operator MUST ensure reverse proxy strips client-supplied XFF headers; otherwise the limiter buckets are spoofable",
		)
	}
	return limiter, trustXFF, nil
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
// gegen den analyzer-service. Setzt der
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
	// Die folgenden drei sind nur in den DB-gestützten Modi (sqlite/
	// postgres) gesetzt, im InMemory-Modus nil (die zugehörigen Features
	// sind dann deaktiviert). Adapter-selektiv, damit der Postgres-Modus
	// nicht die SQLite-Query-Strings gegen PG laufen lässt.
	srtHealth    driven.SrtHealthRepository
	projectToken driven.ProjectTokenRepository
	ingestStream driven.IngestStreamRepository
	// mode trägt den gewählten Persistenz-Modus (sqlite/postgres/inmemory)
	// für Kompatibilitäts-Guards (z. B. SQLite-Issuance-Limiter braucht
	// MTRACE_PERSISTENCE=sqlite).
	mode string
	db   *sql.DB // nil im InMemory-Modus
}

// Close schließt die zugrundeliegende DB, falls vorhanden. No-op für
// InMemory.
func (p *persistenceBundle) Close() {
	if p.db != nil {
		_ = p.db.Close()
	}
}

// sqliteDB liefert die DB nur im SQLite-Modus zurück, sonst nil. Der
// SQLite-Issuance-Limiter (`auth_issuance_counters`) braucht zwingend eine
// SQLite-DB; in Postgres-/InMemory-Modus verhindert das nil, dass
// SQLite-Query-Strings gegen einen fremden Store laufen (der Limiter-Guard
// lehnt den Modus dann mit klarer Meldung ab → Auth deaktiviert).
func (p *persistenceBundle) sqliteDB() *sql.DB {
	if p.mode == persistenceModeSQLite {
		return p.db
	}
	return nil
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
			mode:      persistenceModeInMemory,
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
			mode:         persistenceModeSQLite,
			events:       persistencesqlite.NewEventRepository(db),
			sessions:     persistencesqlite.NewSessionRepository(db),
			sequencer:    seq,
			srtHealth:    persistencesqlite.NewSrtHealthRepository(db),
			projectToken: persistencesqlite.NewProjectTokenRepository(db),
			ingestStream: persistencesqlite.NewIngestStreamRepository(db),
			db:           db,
		}, nil
	case persistenceModePostgres:
		dsn := strings.TrimSpace(os.Getenv(envPostgresDSN))
		if dsn == "" {
			return nil, fmt.Errorf("MTRACE_PERSISTENCE=postgres requires %s (postgres:// DSN)", envPostgresDSN)
		}
		logger.Info("persistence: postgres")
		db, err := storage.OpenPostgres(ctx, dsn)
		if err != nil {
			return nil, fmt.Errorf("storage.OpenPostgres: %w", err)
		}
		// blockSize 0 → defaultBlockSize (512): DB-autoritativer Sequencer
		// via nextval + Block-Allokation (R-28).
		seq, err := persistencepostgres.NewIngestSequencer(ctx, db, 0)
		if err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("ingest sequencer: %w", err)
		}
		return &persistenceBundle{
			mode:         persistenceModePostgres,
			events:       persistencepostgres.NewEventRepository(db),
			sessions:     persistencepostgres.NewSessionRepository(db),
			sequencer:    seq,
			srtHealth:    persistencepostgres.NewSrtHealthRepository(db),
			projectToken: persistencepostgres.NewProjectTokenRepository(db),
			ingestStream: persistencepostgres.NewIngestStreamRepository(db),
			db:           db,
		}, nil
	default:
		return nil, fmt.Errorf("unknown MTRACE_PERSISTENCE=%q (expected 'sqlite', 'postgres', or 'inmemory')", mode)
	}
}

// authBundle bündelt das, was main.go für baut:
// Issuance-Service (Driving-Port) plus den Signer für den
// Konsum-Pfad (PlaybackEventsHandler verifiziert damit Bearer-/
// X-MTrace-Session-Token-Header).
type authBundle struct {
	Inbound        driving.AuthSessionInbound
	Signer         *auth.HMACSessionTokenSigner
	PolicyResolver *auth.InMemoryProjectPolicyResolver
}

// buildAuthSessionService verdrahtet den Auth-Pfad
// (`0.12.0` RAK-72/RAK-75 + `0.12.6` RAK-78): Signing-Key-Ring,
// In-Memory-Issuance-Limiter (global + Project) und
// In-Memory-Project-Policy-Resolver (Fallback aus Static-Project-
// Origins).
//
// Signing-Key-Ring kommt aus zwei alternativen ENV-Pfaden — Parser-
// Logik in `auth.ParseSigningKeysEnv`:
//  - **Multi-Key **: `MTRACE_AUTH_SIGNING_KEYS=
//  kid_a:b64[,kid_b:b64,…]` plus `MTRACE_AUTH_SIGNING_ACTIVE_KID`.
//  Mehrere Keys verifizieren parallel; nur der aktive `kid`
//  signiert (RAK-78). Operator-Workflow siehe `auth.md` §5.3.1.
//  - **Single-Key (Backwards-Compat zu `0.12.0`)**:
//  `MTRACE_AUTH_SIGNING_KEY` plus optional `MTRACE_AUTH_SIGNING_KID`.
//  Degenerierter `len(keys)==1`-Resolver.
//
// Backend-Auswahl per `MTRACE_AUTH_SECRET_BACKEND` (`0.12.6` RAK-79):
//  - `env` (Default): liest aus den ENV-Variablen wie oben.
//  - `vault`: Vault KV-v2-Pfad über `MTRACE_AUTH_VAULT_*`.
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

// wireIngestControlService verdrahtet den IngestControlService mit
// allen optionalen Adapter-Hooks (Outbound-Webhook,
// Media-Server-Provisioner). Eigener Helper, damit `buildHandler`
// unter dem funlen-Limit bleibt.
func wireIngestControlService(repo driven.IngestStreamRepository, logger *slog.Logger) *application.IngestControlService {
	svc := application.NewIngestControlService(repo, time.Now)
	if wh := buildOutboundWebhookDispatcher(logger); wh != nil {
		svc = svc.WithOutboundWebhookDispatcher(wh)
	}
	if prov := buildMediaServerProvisioner(logger); prov != nil {
		svc = svc.WithMediaServerProvisioner(prov)
	}
	return svc
}

// buildMediaServerProvisioner (R-15) liest
// `MTRACE_MEDIASERVER_PROVISION_URL` und `_TOKEN`. Ohne URL ist der
// Adapter deaktiviert — `provision=true` antwortet dann mit
// `media_server_state="disabled"`. Sonst wird ein MediaMTX-
// HTTP-Adapter konstruiert. Aktuell nur MediaMTX; SRS bleibt
// Folge-Item nach `0.12.6`.
func buildMediaServerProvisioner(logger *slog.Logger) driven.MediaServerProvisioner {
	url := strings.TrimSpace(os.Getenv(envMediaServerProvURL))
	if url == "" {
		return nil
	}
	prov, err := mediaserver.New(mediaserver.Config{
		Endpoint:  url,
		AuthToken: os.Getenv(envMediaServerProvToken),
	}, logger)
	if err != nil {
		logger.Error("media server provisioner build failed; disabling provision path",
			"error", err)
		return nil
	}
	logger.Info("media server provisioner active",
		"backend", "mediamtx",
		"endpoint", url,
		"auth_token_present", os.Getenv(envMediaServerProvToken) != "",
	)
	return prov
}

// buildOutboundWebhookDispatcher liest die Outbound-Webhook-
// Konfiguration aus den ENV-Variablen `MTRACE_OUTBOUND_WEBHOOK_URL`
// und `MTRACE_OUTBOUND_WEBHOOK_SECRET` (`0.12.6`/RAK-82, R-16).
// Ist keine URL gesetzt → `nil` (Adapter deaktiviert, identisch
// zum Verhalten ohne Outbound-Webhook). Sonst:
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
// (`0.12.6` RAK-79 + R-20). ENV-Selektor
// `MTRACE_AUTH_SECRET_BACKEND`:
//  - leer / `env`: In-Process-Default — liest `MTRACE_AUTH_SIGNING_KEYS`/
//  `_KEY` / `_ACTIVE_KID` / `_KID` aus dem Prozess-ENV
//  (Backwards-Compat zum Pfad inkl. Lab-Default-Opt-in).
//  - `vault`: externer Adapter über Vault KV-v2; T8
//  mit drei Auth-Methoden (token, approle, kubernetes) über
//  `MTRACE_AUTH_VAULT_AUTH_METHOD`. Fail-closed bei Outage; kein
//  Lab-Default-Fallback.
//  - `kms` (`0.12.6` T8): externer Adapter über
//  `auth.KMSSecretBackend` mit einem injizierten
//  `KMSDecrypter`. Production-Wiring (AWS-SDK-v2 Adapter)
//  ist Folge-Item nach `0.12.6`. Für Lab-Smokes existiert ein
//  `MTRACE_AUTH_KMS_LAB_MODE=1`-Opt-in, der einen
//  Pass-Through-Decrypter aktiviert (Ciphertext = Plaintext).
//  Ohne diesen Opt-in fails der Boot mit klarer Fehlermeldung,
//  damit kein produktiver Boot still auf Lab-Decryption fällt.
//
// Refresh-TTL (`MTRACE_AUTH_SECRET_BACKEND_REFRESH_SECONDS`):
// gelesen + im Boot-Log als Status-Hinweis ausgegeben. Default 0 =
// keine Refresh (Boot-time-only, wie heute). Werte > 0 sind heute
// no-op (`0.12.6` T8 markiert den Refresh-Loop als Folge-Item) —
// der Boot-Log nennt das explizit, damit der Operator weiß, dass
// ein konfigurierter Wert noch nicht greift.
//
// Rückgabe `backendName` wird vom Caller fürs Fehler-Wording und
// das Lab-Default-Fallback-Gate genutzt.
func buildAuthSecretBackend(logger *slog.Logger) (driven.AuthSecretBackend, string, error) {
	backend := strings.ToLower(strings.TrimSpace(os.Getenv(envAuthSecretBackend)))
	logRefreshTTL(logger)
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
			"auth_method", strings.ToLower(strings.TrimSpace(os.Getenv("MTRACE_AUTH_VAULT_AUTH_METHOD"))),
			"note", "boot-time load; refresh-loop is follow-up item; fail-closed on outage")
		return vb, "vault", nil
	case "kms":
		decrypter, err := buildKMSDecrypter()
		if err != nil {
			return nil, "", fmt.Errorf("%s=kms: %w", envAuthSecretBackend, err)
		}
		kb, err := auth.NewKMSSecretBackend(os.Getenv, decrypter)
		if err != nil {
			return nil, "", fmt.Errorf("%s=kms: %w", envAuthSecretBackend, err)
		}
		logger.Info("auth secret backend active", "backend", "kms",
			"decrypter", kmsDecrypterLabel(),
			"note", "boot-time decrypt; refresh-loop is follow-up item; fail-closed on outage")
		return kb, "kms", nil
	default:
		return nil, "", fmt.Errorf(
			"%s=%q is not supported (valid: env|vault|kms)",
			envAuthSecretBackend, backend,
		)
	}
}

// logRefreshTTL berichtet den Operator-konfigurierten Refresh-TTL
// im Boot-Log. Default `0` = boot-only. Werte > 0 sind heute no-op
// (Refresh-Loop ist Folge-Item nach `0.12.6` T8).
func logRefreshTTL(logger *slog.Logger) {
	raw := strings.TrimSpace(os.Getenv("MTRACE_AUTH_SECRET_BACKEND_REFRESH_SECONDS"))
	if raw == "" || raw == "0" {
		return
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		logger.Warn("auth secret backend refresh TTL invalid; ignoring",
			"value", raw, "fallback", "boot-only")
		return
	}
	logger.Warn("auth secret backend refresh TTL is configured but not yet active",
		"requested_seconds", n,
		"current_behavior", "boot-only load",
		"follow_up", "refresh-loop is a follow-up item after 0.12.6 Tranche 8")
}

// buildKMSDecrypter baut den `KMSDecrypter` für den `kms`-Pfad.
// Heute unterstützt nur der Lab-Mode-Decrypter (Pass-Through, opt-in
// via `MTRACE_AUTH_KMS_LAB_MODE=1`) — produktive AWS-SDK-Anbindung
// ist Folge-Item. Ohne den Lab-Opt-in fails die Funktion, damit kein
// produktiver Boot still auf Lab-Decryption fällt.
//
//nolint:ireturn // KMSDecrypter ist der vendor-neutrale Port; Boot-Wiring darf bewusst Interface zurückgeben (Operator-Injection in Folge-Item).
func buildKMSDecrypter() (auth.KMSDecrypter, error) {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("MTRACE_AUTH_KMS_LAB_MODE")), "1") ||
		strings.EqualFold(strings.TrimSpace(os.Getenv("MTRACE_AUTH_KMS_LAB_MODE")), "true") {
		return auth.LabPassThroughKMSDecrypter{}, nil
	}
	return nil, fmt.Errorf(
		"MTRACE_AUTH_KMS_LAB_MODE=1 is required (production AWS-SDK-v2 decrypter is a follow-up item after 0.12.6 Tranche 8; operators with KMS in production must inject their own KMSDecrypter)",
	)
}

func kmsDecrypterLabel() string {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("MTRACE_AUTH_KMS_LAB_MODE")), "1") ||
		strings.EqualFold(strings.TrimSpace(os.Getenv("MTRACE_AUTH_KMS_LAB_MODE")), "true") {
		return "lab-pass-through (NOT FOR PRODUCTION)"
	}
	return "operator-injected"
}

// buildIssuanceRateLimiter wählt zwischen In-Process-, SQLite- und
// Redis-basiertem Token-Bucket-Limiter (`0.12.6` RAK-77 / R-17 +
// R-17-Resttrigger). ENV-Selektor
// `MTRACE_AUTH_ISSUANCE_LIMITER`:
//  - leer / `memory`: In-Process-Default (Backwards-Compat zu
//  `0.12.0`). Misst pro Replica — passt nur für Single-Instance-
//  Setups.
//  - `sqlite`: opt-in Shared-State-Pfad für Single-Host-Multi-
//  Replica-Setups. Multi-Host bleibt nicht-Multi-Host-safe.
//  - `redis` (`0.12.6` T7): Network-Backend für echte Multi-Host-
//  Setups. Atomare Lua-Token-Bucket-Operation; teilt sich den
//  Redis-Server mit dem Origin-Limiter (`R-22`) für Backend-
//  Konsistenz. Pflicht-ENV `MTRACE_REDIS_ADDR`; optional
//  `MTRACE_REDIS_AUTH`/`MTRACE_REDIS_DB`. Fail-mode default
//  fail-closed (Outage → 429); `MTRACE_AUTH_ISSUANCE_FAIL_OPEN=1`
//  aktiviert lokalen In-Memory-Fallback pro Replica.
//  - `memcached` (Folge-Item nach `0.12.6`): explizit nicht
//  unterstützt — bleibt gemeinsames Folge-Item mit R-22, falls
//  ein Operator Memcached vorzieht.
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
				"%s=sqlite requires MTRACE_PERSISTENCE=sqlite (got nil SQLite handle — postgres/inmemory have no SQLite issuance store; use redis or memory)",
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
	case "redis":
		client, err := buildRedisClient()
		if err != nil {
			return nil, fmt.Errorf("%s=redis: %w", envAuthIssuanceLimiter, err)
		}
		failOpen := issuanceFailOpenOptIn()
		logger.Info("auth issuance limiter active", "backend", "redis",
			"fail_mode", failModeLabel(failOpen),
		)
		return auth.NewRedisIssuanceRateLimiter(client, auth.RedisIssuanceLimiterConfig{
			GlobalCapacity:      authIssuanceGlobalCapacity,
			GlobalRefillPerSec:  authIssuanceGlobalRefillPerSec,
			ProjectCapacity:     authIssuanceProjectCapacity,
			ProjectRefillPerSec: authIssuanceProjectRefillPerSec,
			TTLSeconds:          24 * 3600,
			FailOpen:            failOpen,
		}, logger)
	case "memcached":
		return nil, fmt.Errorf(
			"%s=memcached is a follow-up item — gets delivered jointly with R-22 in a future tranche to avoid backend fragmentation",
			envAuthIssuanceLimiter,
		)
	default:
		return nil, fmt.Errorf(
			"%s=%q is not supported (valid: memory|sqlite|redis; memcached is a follow-up item)",
			envAuthIssuanceLimiter, backend,
		)
	}
}

// buildRedisClient liest `MTRACE_REDIS_ADDR`/`_AUTH`/`_DB` und
// liefert einen go-redis-Client. `Addr` ist Pflicht; Auth und DB
// sind optional.
func buildRedisClient() (*redis.Client, error) {
	addr := strings.TrimSpace(os.Getenv(envRedisAddr))
	if addr == "" {
		return nil, fmt.Errorf("%s is required for redis backend (host:port form)", envRedisAddr)
	}
	dbIndex := 0
	if raw := strings.TrimSpace(os.Getenv(envRedisDB)); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 0 {
			return nil, fmt.Errorf("%s=%q must be a non-negative integer", envRedisDB, raw)
		}
		dbIndex = n
	}
	return redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: os.Getenv(envRedisAuth),
		DB:       dbIndex,
	}), nil
}

// issuanceFailOpenOptIn liest `MTRACE_AUTH_ISSUANCE_FAIL_OPEN` und
// akzeptiert nur explizit-truthy Werte. Default ist fail-closed.
func issuanceFailOpenOptIn() bool { return envTruthyOptIn(envAuthIssuanceFailOpen) }

func failModeLabel(failOpen bool) string {
	if failOpen {
		return "fail-open (local memory fallback on outage)"
	}
	return "fail-closed (deny on outage)"
}

// buildOriginRateLimiter (R-22) wählt den
// Origin-/IP-Rate-Limiter-Backend per ENV-Selektor
// `MTRACE_ORIGIN_RATE_LIMITER`:
//  - leer / `disabled`: kein Limiter (Backwards-Compat-Default für
//  bestehende Operatoren). Memory-Backend bleibt opt-in.
//  - `memory`: In-Process-Token-Bucket pro Key (`r.RemoteAddr` oder
//  `X-Forwarded-For`-Client-IP). Single-Replica-Pfad oder
//  Defense-in-Depth-Ergänzung zum Edge-Layer-Limit (Reverse-Proxy/
//  CDN).
//  - `sqlite`: NICHT unterstützt — Origin-Limits über Hosts hinweg
//  brauchen ein Network-Backend; SQLite via Shared-Volume produziert
//  false-negative-Limits, sobald Replicas auf verschiedenen Hosts
//  laufen (siehe Backend-Strategie).
//  - `redis` (`0.12.6` T7): Network-Backend für Multi-Host-Setups.
//  Atomare Lua-Token-Bucket-Operation; teilt sich den Redis-
//  Server mit dem Issuance-Limiter (`R-17`) — derselbe
//  `MTRACE_REDIS_*`-ENV-Block, aber eigener Key-Prefix
//  `mtrace:origin`. Fail-mode default fail-closed; opt-in
//  fail-open via `MTRACE_AUTH_ISSUANCE_FAIL_OPEN=1` (gemeinsam
//  mit dem Issuance-Limiter — beide Limiter teilen denselben
//  Fail-Mode-Schalter, damit ein Operator nicht versehentlich
//  einen halb-fail-closed Pfad konstruiert).
//  - `memcached`: Folge-Item gemeinsam mit dem Issuance-Limiter,
//  falls Operator-Bedarf nach Memcached entsteht.
//
// Liefert `nil, nil` für den Disabled-Pfad — der HTTP-Adapter prüft
// auf nil und überspringt den Middleware-Aufruf.
func buildOriginRateLimiter(logger *slog.Logger) (driven.OriginRateLimiter, error) {
	backend := strings.ToLower(strings.TrimSpace(os.Getenv(envOriginRateLimiter)))
	switch backend {
	case "", "disabled":
		logger.Info("origin rate limiter disabled (set MTRACE_ORIGIN_RATE_LIMITER=memory|redis to enable)")
		return nil, nil
	case "memory":
		logger.Info("origin rate limiter active", "backend", "memory",
			"capacity", originRateLimitCapacity,
			"refill_per_second", originRateLimitRefillPerSec,
		)
		return auth.NewInMemoryOriginRateLimiter(
			originRateLimitCapacity, originRateLimitRefillPerSec,
		), nil
	case "redis":
		client, err := buildRedisClient()
		if err != nil {
			return nil, fmt.Errorf("%s=redis: %w", envOriginRateLimiter, err)
		}
		failOpen := issuanceFailOpenOptIn()
		logger.Info("origin rate limiter active", "backend", "redis",
			"capacity", originRateLimitCapacity,
			"refill_per_second", originRateLimitRefillPerSec,
			"fail_mode", failModeLabel(failOpen),
		)
		return auth.NewRedisOriginRateLimiter(client, auth.RedisOriginLimiterConfig{
			Capacity:        originRateLimitCapacity,
			RefillPerSecond: originRateLimitRefillPerSec,
			TTLSeconds:      600,
			FailOpen:        failOpen,
		}, logger)
	case "sqlite":
		return nil, fmt.Errorf(
			"%s=sqlite is not supported (Origin-Limits are not Multi-Host-safe on shared SQLite volumes)",
			envOriginRateLimiter,
		)
	case "memcached":
		return nil, fmt.Errorf(
			"%s=memcached is a follow-up item — gets delivered jointly with R-17 in a future tranche to avoid backend fragmentation between issuance- and origin-limiter",
			envOriginRateLimiter,
		)
	default:
		return nil, fmt.Errorf(
			"%s=%q is not supported (valid: disabled|memory|redis; memcached is a follow-up item)",
			envOriginRateLimiter, backend,
		)
	}
}

// trustForwardedForOptIn liest `MTRACE_TRUST_FORWARDED_FOR` und
// akzeptiert nur `1`/`true`/`yes`. Alles andere (inkl. fehlend) →
// HTTP-Adapter nutzt `r.RemoteAddr` als client_ip-Quelle (Default).
// Operator muss XFF explizit aktivieren, sonst trifft der Origin-
// Limiter den Reverse-Proxy statt den Client.
func trustForwardedForOptIn() bool { return envTruthyOptIn(envTrustForwardedFor) }

// labDefaultOptIn liest `MTRACE_AUTH_LAB_DEFAULT` und akzeptiert nur
// die explizit truthy Werte `1`/`true`/`yes`. Alles andere (inklusive
// fehlend) gilt als „nicht opt-in" — der Aufrufer hard-failt dann.
func labDefaultOptIn() bool { return envTruthyOptIn(envAuthLabDefault) }
