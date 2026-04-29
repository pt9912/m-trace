package http

import (
	"log/slog"
	"net/http"

	"github.com/pt9912/m-trace/apps/api/hexagon/port/driving"
)

// NewRouter wires the three pflicht endpoints onto a single mux.
// Method routing uses Go 1.22 method-aware patterns ("POST /path"),
// so non-matching methods fall through to a 404 from the mux.
func NewRouter(
	useCase driving.PlaybackEventInbound,
	metricsHandler http.Handler,
	logger *slog.Logger,
) http.Handler {
	mux := http.NewServeMux()

	playback := &PlaybackEventsHandler{
		UseCase: useCase,
		Logger:  logger,
	}

	mux.Handle("POST /api/playback-events", playback)
	mux.HandleFunc("GET /api/health", HealthHandler)
	mux.Handle("GET /api/metrics", metricsHandler)

	return mux
}
