package http

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driving"
)

// MaxBodyBytes caps the request body at 256 KB
// (docs/spike/backend-api-contract.md §5 step 2).
const MaxBodyBytes = 256 * 1024

// PlaybackEventsHandler implements POST /api/playback-events.
type PlaybackEventsHandler struct {
	UseCase driving.PlaybackEventInbound
	Logger  *slog.Logger
}

// ServeHTTP follows the validation order from
// docs/spike/backend-api-contract.md §5:
//
//	step 1: X-MTrace-Token header presence -> 401
//	step 2: body size                      -> 413
//	(steps 3-10 are inside the use case, mapped from domain errors)
func (h *PlaybackEventsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Method routing is done by the mux (Go 1.22 method-aware patterns)
	// but we keep an explicit guard so the handler is robust if mounted
	// without a method filter.
	if r.Method != http.MethodPost {
		writeStatus(w, http.StatusMethodNotAllowed)
		return
	}

	// Step 1 — Auth-Header presence. Origin-loser Fast-Reject vor dem
	// Body-Read; siehe API-Kontrakt §5 (Auth-vor-Body-Reihenfolge,
	// Patch 40d79d9).
	token := r.Header.Get("X-MTrace-Token")
	if token == "" {
		writeStatus(w, http.StatusUnauthorized)
		return
	}

	// Step 2 — Body size limit. MaxBytesReader wraps r.Body so reads
	// past the limit return *http.MaxBytesError.
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			writeStatus(w, http.StatusRequestEntityTooLarge)
			return
		}
		writeStatus(w, http.StatusBadRequest)
		return
	}

	var payload wireBatch
	if err := json.Unmarshal(body, &payload); err != nil {
		writeStatus(w, http.StatusBadRequest)
		return
	}

	in := driving.BatchInput{
		SchemaVersion: payload.SchemaVersion,
		AuthToken:     token,
		Events:        toEventInputs(payload.Events),
	}

	res, err := h.UseCase.RegisterPlaybackEventBatch(r.Context(), in)
	if err != nil {
		h.writeUseCaseError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]int{"accepted": res.Accepted})
}

func (h *PlaybackEventsHandler) writeUseCaseError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrSchemaVersionMismatch):
		writeStatus(w, http.StatusBadRequest)
	case errors.Is(err, domain.ErrUnauthorized):
		writeStatus(w, http.StatusUnauthorized)
	case errors.Is(err, domain.ErrBatchEmpty),
		errors.Is(err, domain.ErrBatchTooLarge),
		errors.Is(err, domain.ErrInvalidEvent):
		writeStatus(w, http.StatusUnprocessableEntity)
	case errors.Is(err, domain.ErrRateLimited):
		// Spike scope: a static 1-second hint is sufficient. A real
		// implementation would expose remaining budget from the bucket.
		w.Header().Set("Retry-After", "1")
		writeStatus(w, http.StatusTooManyRequests)
	default:
		h.Logger.Error("unhandled use case error", "error", err)
		writeStatus(w, http.StatusInternalServerError)
	}
}

func writeStatus(w http.ResponseWriter, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write([]byte("{}"))
}
