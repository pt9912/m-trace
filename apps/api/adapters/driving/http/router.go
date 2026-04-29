package http

import (
	"log/slog"
	"net/http"

	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"

	"github.com/pt9912/m-trace/apps/api/hexagon/port/driving"
)

// NewRouter wires the three pflicht endpoints onto a single mux.
// Method routing uses Go 1.22 method-aware patterns ("POST /path"),
// so non-matching methods fall through to a 404 from the mux.
//
// Tracer is used to wrap POST /api/playback-events in a request-span
// (docs/architecture.md §5.3 — der HTTP-Adapter ist neben
// adapters/driven/telemetry der einzige Ort mit OTel-Imports). A nil
// tracer falls back to a no-op tracer so tests can wire the router
// without an OTel SDK setup.
func NewRouter(
	useCase driving.PlaybackEventInbound,
	metricsHandler http.Handler,
	tracer trace.Tracer,
	logger *slog.Logger,
) http.Handler {
	if tracer == nil {
		tracer = tracenoop.NewTracerProvider().Tracer("noop")
	}

	mux := http.NewServeMux()

	playback := &PlaybackEventsHandler{
		UseCase: useCase,
		Tracer:  tracer,
		Logger:  logger,
	}

	mux.Handle("POST /api/playback-events", playback)
	mux.HandleFunc("GET /api/health", HealthHandler)
	mux.Handle("GET /api/metrics", metricsHandler)

	return mux
}
