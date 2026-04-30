// Package main is the entry point of the m-trace API spike (Go variant).
//
// Wires the driven adapters (auth, persistence, ratelimit, metrics,
// telemetry) into the application use case and exposes the three
// pflicht endpoints (POST /api/playback-events, GET /api/health,
// GET /api/metrics) over HTTP. See docs/spike/0001-backend-stack.md
// and docs/plan-spike.md for scope.
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/auth"
	"github.com/pt9912/m-trace/apps/api/adapters/driven/metrics"
	"github.com/pt9912/m-trace/apps/api/adapters/driven/persistence"
	"github.com/pt9912/m-trace/apps/api/adapters/driven/ratelimit"
	"github.com/pt9912/m-trace/apps/api/adapters/driven/streamanalyzer"
	"github.com/pt9912/m-trace/apps/api/adapters/driven/telemetry"
	apihttp "github.com/pt9912/m-trace/apps/api/adapters/driving/http"
	"github.com/pt9912/m-trace/apps/api/hexagon/application"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

const (
	serviceName       = "m-trace-api"
	serviceVersion    = "0.1.2"
	defaultListenAddr = ":8080"

	// Spike Spec §6.9: 100 events/sec/project.
	rateLimitCapacity = 100
	rateLimitRefill   = 100.0
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

	repo := persistence.NewInMemoryEventRepository()
	sessions := persistence.NewInMemorySessionRepository()
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
		var active float64
		for _, session := range sessions.Snapshot() {
			if session.State == domain.SessionStateActive {
				active++
			}
		}
		return active
	}))
	analyzer := streamanalyzer.NewNoopStreamAnalyzer()
	sequencer := persistence.NewInMemoryIngestSequencer()

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

	tracer := otelProviders.Tracer.Tracer(telemetry.TracerName)
	router := apihttp.NewRouter(useCase, sessionsService, resolver, publisher.Handler(), tracer, logger)
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
