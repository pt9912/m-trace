// Package main is the entry point of the m-trace API spike (Go variant).
//
// Wires the driven adapters (auth, persistence, ratelimit, metrics,
// telemetry) into the application use case and exposes the three
// pflicht endpoints (POST /api/playback-events, GET /api/health,
// GET /api/metrics) over HTTP. See docs/spike/0001-backend-stack.md
// and docs/planning/plan-spike.md for scope.
package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/auth"
	"github.com/pt9912/m-trace/apps/api/adapters/driven/metrics"
	"github.com/pt9912/m-trace/apps/api/adapters/driven/persistence/inmemory"
	persistencesqlite "github.com/pt9912/m-trace/apps/api/adapters/driven/persistence/sqlite"
	"github.com/pt9912/m-trace/apps/api/adapters/driven/ratelimit"
	"github.com/pt9912/m-trace/apps/api/adapters/driven/streamanalyzer"
	"github.com/pt9912/m-trace/apps/api/adapters/driven/telemetry"
	apihttp "github.com/pt9912/m-trace/apps/api/adapters/driving/http"
	"github.com/pt9912/m-trace/apps/api/hexagon/application"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
	"github.com/pt9912/m-trace/apps/api/internal/storage"
)

const (
	serviceName       = "m-trace-api"
	serviceVersion    = "0.1.2"
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

	otelProviders, err := telemetry.Setup(context.Background(), serviceName, serviceVersion)
	if err != nil {
		logger.Error("otel setup failed", "error", err)
		os.Exit(1)
	}

	otelTelemetry, err := telemetry.NewOTelTelemetry(otelProviders.Meter.Meter(telemetry.MeterName))
	if err != nil {
		logger.Error("otel telemetry adapter init failed", "error", err)
		os.Exit(1)
	}

	persistCtx := context.Background()
	persist, err := newPersistence(persistCtx, logger)
	if err != nil {
		logger.Error("persistence init failed", "error", err)
		os.Exit(1)
	}
	defer persist.Close()
	repo := persist.events
	sessions := persist.sessions
	sequencer := persist.sequencer

	resolver := auth.NewStaticProjectResolver(map[string]auth.ProjectConfig{
		"demo": {
			Token: "demo-token",
			AllowedOrigins: []string{
				"http://localhost:5173",
				"http://localhost:3000",
				"http://dashboard-e2e:5173",
			},
		},
	})
	limiter := ratelimit.NewTokenBucketRateLimiter(rateLimitCapacity, rateLimitRefill, time.Now)
	publisher := metrics.NewPrometheusPublisher(metrics.WithActiveSessionsFunc(func() float64 {
		// Prometheus-Scrape-Pfad: on-demand-Lookup über das adapter-
		// agnostische CountByState. SQLite-Adapter macht ein
		// `SELECT COUNT(*) WHERE state='active'`; InMemory-Adapter
		// loopt über die in-memory Map. Errors werden geloggt und
		// auf 0 gemappt — der Gauge bleibt damit immer lesbar.
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		n, err := sessions.CountByState(ctx, domain.SessionStateActive)
		if err != nil {
			logger.Warn("active sessions count failed", "error", err)
			return 0
		}
		return float64(n)
	}))
	analyzer := newAnalyzer(logger)

	useCase := application.NewRegisterPlaybackEventBatchUseCase(
		resolver, limiter, repo, sessions, publisher, otelTelemetry, analyzer, sequencer, time.Now,
	)

	processID, err := newProcessInstanceID()
	if err != nil {
		logger.Error("process_instance_id generation failed", "error", err)
		os.Exit(1)
	}
	logger.Info("process instance allocated", "process_instance_id", string(processID))
	sessionsService := application.NewSessionsService(sessions, repo, processID)
	sessionsSweeper := application.NewSessionsSweeper(sessions, time.Now, logger)

	analysisService := application.NewAnalyzeManifestUseCase(analyzer)

	tracer := otelProviders.Tracer.Tracer(telemetry.TracerName)
	router := apihttp.NewRouter(useCase, sessionsService, analysisService, resolver, publisher.Handler(), publisher, tracer, logger)
	router = apihttp.RequestMetricsMiddleware(router, publisher)
	addr := listenAddr()

	srv := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go sessionsSweeper.Run(ctx)

	go func() {
		logger.Info("server starting",
			"addr", srv.Addr,
			"service", serviceName,
			"version", serviceVersion,
		)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
	if err := otelProviders.Shutdown(shutdownCtx); err != nil {
		logger.Error("otel shutdown failed", "error", err)
	}
	logger.Info("server stopped")
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

// newProcessInstanceID erzeugt eine 16-Byte-Zufalls-ID und gibt sie als
// Hex-String zurück (32 Zeichen). Verwendet als domain.ProcessInstanceID
// im Cursor-Vertrag aus plan-0.1.0.md §5.1.
func newProcessInstanceID() (domain.ProcessInstanceID, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return domain.ProcessInstanceID(hex.EncodeToString(b[:])), nil
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
