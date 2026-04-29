package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driving"
)

// MaxBodyBytes caps the request body at 256 KB
// (docs/spike/backend-api-contract.md §5 step 2).
const MaxBodyBytes = 256 * 1024

// spanName is the OTel span name for the playback-events handler.
// Mirrors the verbindliche Tabelle in docs/telemetry-model.md §2.1
// (single span per request, scope = HTTP-Adapter).
const spanName = "http.handler POST /api/playback-events"

// PlaybackEventsHandler implements POST /api/playback-events. The
// Tracer wraps each request in a span; ServeHTTP records the
// http.method/route, http.status_code, batch.size and batch.outcome
// attributes documented in docs/telemetry-model.md §2.1.
type PlaybackEventsHandler struct {
	UseCase driving.PlaybackEventInbound
	Tracer  trace.Tracer
	Logger  *slog.Logger
}

// ServeHTTP follows the validation order from
// docs/spike/backend-api-contract.md §5:
//
//	step 1: X-MTrace-Token header presence -> 401
//	step 2: body size                      -> 413
//	(steps 3-10 are inside the use case, mapped from domain errors)
//
// The whole request is wrapped in a single OTel server-span; per
// docs/architecture.md §3.4 the HTTP adapter is one of two places
// allowed to import OTel directly.
func (h *PlaybackEventsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.Tracer.Start(r.Context(), spanName,
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(
			attribute.String("http.method", http.MethodPost),
			attribute.String("http.route", "/api/playback-events"),
		),
	)
	defer span.End()

	rec := &statusRecorder{ResponseWriter: w}
	batchSize := h.serve(ctx, rec, r)

	span.SetAttributes(attribute.Int("http.status_code", rec.statusCode()))
	if batchSize >= 0 {
		span.SetAttributes(attribute.Int("batch.size", batchSize))
	}
	span.SetAttributes(attribute.String("batch.outcome", outcomeFor(rec.statusCode())))
	if rec.statusCode() >= http.StatusInternalServerError {
		span.SetStatus(codes.Error, http.StatusText(rec.statusCode()))
	} else {
		span.SetStatus(codes.Ok, "")
	}
}

// serve runs the request pipeline and returns the parsed batch size,
// or -1 if the request was rejected before JSON parsing.
func (h *PlaybackEventsHandler) serve(ctx context.Context, w http.ResponseWriter, r *http.Request) int {
	// Method routing is done by the mux (Go 1.22 method-aware patterns)
	// but we keep an explicit guard so the handler is robust if mounted
	// without a method filter.
	if r.Method != http.MethodPost {
		writeStatus(w, http.StatusMethodNotAllowed)
		return -1
	}

	// Step 1 — Auth-Header presence. Origin-loser Fast-Reject vor dem
	// Body-Read; siehe API-Kontrakt §5 (Auth-vor-Body-Reihenfolge,
	// Patch 40d79d9).
	token := r.Header.Get("X-MTrace-Token")
	if token == "" {
		writeStatus(w, http.StatusUnauthorized)
		return -1
	}

	// Step 2 — Body size limit. MaxBytesReader wraps r.Body so reads
	// past the limit return *http.MaxBytesError.
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			writeStatus(w, http.StatusRequestEntityTooLarge)
			return -1
		}
		writeStatus(w, http.StatusBadRequest)
		return -1
	}

	var payload wireBatch
	if err := json.Unmarshal(body, &payload); err != nil {
		writeStatus(w, http.StatusBadRequest)
		return -1
	}

	in := driving.BatchInput{
		SchemaVersion: payload.SchemaVersion,
		AuthToken:     token,
		Origin:        r.Header.Get("Origin"),
		Events:        toEventInputs(payload.Events),
	}
	batchSize := len(in.Events)

	res, err := h.UseCase.RegisterPlaybackEventBatch(ctx, in)
	if err != nil {
		h.writeUseCaseError(w, err)
		return batchSize
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]int{"accepted": res.Accepted})
	return batchSize
}

func (h *PlaybackEventsHandler) writeUseCaseError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrSchemaVersionMismatch):
		writeStatus(w, http.StatusBadRequest)
	case errors.Is(err, domain.ErrUnauthorized):
		writeStatus(w, http.StatusUnauthorized)
	case errors.Is(err, domain.ErrOriginNotAllowed):
		writeStatus(w, http.StatusForbidden)
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

// statusRecorder captures the first WriteHeader call so the span can
// record http.status_code without depending on the use case to surface
// it explicitly.
type statusRecorder struct {
	http.ResponseWriter
	code        int
	wroteHeader bool
}

func (r *statusRecorder) WriteHeader(code int) {
	if !r.wroteHeader {
		r.code = code
		r.wroteHeader = true
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.code = http.StatusOK
		r.wroteHeader = true
	}
	return r.ResponseWriter.Write(b)
}

func (r *statusRecorder) statusCode() int {
	if !r.wroteHeader {
		return http.StatusOK
	}
	return r.code
}

// outcomeFor maps HTTP status codes to the small set of batch.outcome
// values used in docs/telemetry-model.md §2.1. Buckets are bounded to
// avoid attribute-cardinality blow-up.
func outcomeFor(code int) string {
	switch {
	case code == http.StatusAccepted:
		return "accepted"
	case code == http.StatusUnauthorized:
		return "unauthorized"
	case code == http.StatusForbidden:
		return "forbidden"
	case code == http.StatusRequestEntityTooLarge:
		return "too_large"
	case code == http.StatusTooManyRequests:
		return "rate_limited"
	case code == http.StatusBadRequest, code == http.StatusUnprocessableEntity:
		return "invalid"
	case code >= http.StatusInternalServerError:
		return "error"
	default:
		return "other"
	}
}
