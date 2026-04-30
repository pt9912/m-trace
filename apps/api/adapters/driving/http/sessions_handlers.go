package http

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driving"
)

// SessionsListHandler implementiert GET /api/stream-sessions
// (plan-0.1.0.md §5.1). Span-Konvention siehe telemetry-model.md §2.1
// — Span-Name "http.handler GET /api/stream-sessions" mit
// http.method/http.route/http.status_code-Attributen.
type SessionsListHandler struct {
	UseCase driving.SessionsInbound
	Tracer  trace.Tracer
	Logger  *slog.Logger
}

const sessionsListSpan = "http.handler GET /api/stream-sessions"

// ServeHTTP wickelt jeden Request in einen Server-Span und schreibt
// das Status-Code-Attribut nach der inneren Verarbeitung.
func (h *SessionsListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.Tracer.Start(r.Context(), sessionsListSpan,
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(
			attribute.String("http.method", http.MethodGet),
			attribute.String("http.route", "/api/stream-sessions"),
		),
	)
	defer span.End()

	rec := &statusRecorder{ResponseWriter: w}
	h.serve(ctx, rec, r)

	span.SetAttributes(attribute.Int("http.status_code", rec.statusCode()))
	if rec.statusCode() >= http.StatusInternalServerError {
		span.SetStatus(codes.Error, http.StatusText(rec.statusCode()))
	} else {
		span.SetStatus(codes.Ok, "")
	}
}

func (h *SessionsListHandler) serve(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeStatus(w, http.StatusMethodNotAllowed)
		return
	}

	limit, err := parseLimit(r.URL.Query().Get("limit"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "limit_invalid"})
		return
	}

	cursor, err := decodeListSessionsCursor(r.URL.Query().Get("cursor"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":  "cursor_invalid",
			"reason": "malformed",
		})
		return
	}

	res, err := h.UseCase.ListSessions(ctx, driving.ListSessionsInput{
		Limit: limit,
		After: cursor,
	})
	if err != nil {
		if errors.Is(err, domain.ErrCursorInvalid) {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error":  "cursor_invalid",
				"reason": "storage_restart",
			})
			return
		}
		h.Logger.Error("ListSessions failed", "error", err)
		writeStatus(w, http.StatusInternalServerError)
		return
	}

	body := struct {
		Sessions   []sessionWire `json:"sessions"`
		NextCursor string        `json:"next_cursor,omitempty"`
	}{
		Sessions: toSessionWireList(res.Sessions),
	}
	if res.NextCursor != nil {
		next, err := encodeListSessionsCursor(res.NextCursor)
		if err != nil {
			h.Logger.Error("encode list cursor failed", "error", err)
			writeStatus(w, http.StatusInternalServerError)
			return
		}
		body.NextCursor = next
	}
	writeJSON(w, http.StatusOK, body)
}

// SessionsGetHandler implementiert GET /api/stream-sessions/{id}.
type SessionsGetHandler struct {
	UseCase driving.SessionsInbound
	Tracer  trace.Tracer
	Logger  *slog.Logger
}

const sessionsGetSpan = "http.handler GET /api/stream-sessions/{id}"

func (h *SessionsGetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.Tracer.Start(r.Context(), sessionsGetSpan,
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(
			attribute.String("http.method", http.MethodGet),
			attribute.String("http.route", "/api/stream-sessions/{id}"),
		),
	)
	defer span.End()

	rec := &statusRecorder{ResponseWriter: w}
	h.serve(ctx, rec, r)

	span.SetAttributes(attribute.Int("http.status_code", rec.statusCode()))
	if rec.statusCode() >= http.StatusInternalServerError {
		span.SetStatus(codes.Error, http.StatusText(rec.statusCode()))
	} else {
		span.SetStatus(codes.Ok, "")
	}
}

func (h *SessionsGetHandler) serve(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeStatus(w, http.StatusMethodNotAllowed)
		return
	}

	id := r.PathValue("id")
	if strings.TrimSpace(id) == "" {
		writeStatus(w, http.StatusNotFound)
		return
	}

	limit, err := parseLimitWithName(r.URL.Query().Get("events_limit"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "events_limit_invalid"})
		return
	}

	cursor, err := decodeSessionEventsCursor(r.URL.Query().Get("events_cursor"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":  "cursor_invalid",
			"reason": "malformed",
		})
		return
	}

	res, err := h.UseCase.GetSession(ctx, driving.GetSessionInput{
		SessionID:   id,
		EventsLimit: limit,
		EventsAfter: cursor,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrSessionNotFound):
			writeStatus(w, http.StatusNotFound)
			return
		case errors.Is(err, domain.ErrCursorInvalid):
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error":  "cursor_invalid",
				"reason": "storage_restart",
			})
			return
		default:
			h.Logger.Error("GetSession failed", "error", err)
			writeStatus(w, http.StatusInternalServerError)
			return
		}
	}

	body := struct {
		Session    sessionWire `json:"session"`
		Events     []eventWire `json:"events"`
		NextCursor string      `json:"next_cursor,omitempty"`
	}{
		Session: toSessionWire(res.Session),
		Events:  toEventWireList(res.Events),
	}
	if res.NextCursor != nil {
		next, err := encodeSessionEventsCursor(res.NextCursor)
		if err != nil {
			h.Logger.Error("encode events cursor failed", "error", err)
			writeStatus(w, http.StatusInternalServerError)
			return
		}
		body.NextCursor = next
	}
	writeJSON(w, http.StatusOK, body)
}

// parseLimit liest den optionalen `limit`-Query-Parameter. Leer → 0
// (Use Case clampt auf Default). Negativ oder NaN → Fehler.
func parseLimit(s string) (int, error) {
	return parseLimitWithName(s)
}

func parseLimitWithName(s string) (int, error) {
	if s == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return 0, errors.New("invalid limit")
	}
	return n, nil
}

// sessionWire ist die JSON-Antwortform für domain.StreamSession.
type sessionWire struct {
	ID          string  `json:"session_id"`
	ProjectID   string  `json:"project_id"`
	State       string  `json:"state"`
	StartedAt   string  `json:"started_at"`
	LastEventAt string  `json:"last_event_at"`
	EndedAt     *string `json:"ended_at,omitempty"`
	EventCount  int64   `json:"event_count"`
}

func toSessionWire(s domain.StreamSession) sessionWire {
	out := sessionWire{
		ID:          s.ID,
		ProjectID:   s.ProjectID,
		State:       string(s.State),
		StartedAt:   s.StartedAt.UTC().Format(time.RFC3339Nano),
		LastEventAt: s.LastEventAt.UTC().Format(time.RFC3339Nano),
		EventCount:  s.EventCount,
	}
	if s.EndedAt != nil {
		ended := s.EndedAt.UTC().Format(time.RFC3339Nano)
		out.EndedAt = &ended
	}
	return out
}

func toSessionWireList(in []domain.StreamSession) []sessionWire {
	out := make([]sessionWire, len(in))
	for i, s := range in {
		out[i] = toSessionWire(s)
	}
	return out
}

// eventWire ist die JSON-Antwortform für ein Event im Detail-Response.
type eventWire struct {
	EventName        string         `json:"event_name"`
	ProjectID        string         `json:"project_id"`
	SessionID        string         `json:"session_id"`
	ClientTimestamp  string         `json:"client_timestamp"`
	ServerReceivedAt string         `json:"server_received_at"`
	IngestSequence   int64          `json:"ingest_sequence"`
	SequenceNumber   *int64         `json:"sequence_number,omitempty"`
	SDK              sdkWire        `json:"sdk"`
	Meta             map[string]any `json:"meta,omitempty"`
}

type sdkWire struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func toEventWireList(in []domain.PlaybackEvent) []eventWire {
	out := make([]eventWire, len(in))
	for i, e := range in {
		out[i] = eventWire{
			EventName:        e.EventName,
			ProjectID:        e.ProjectID,
			SessionID:        e.SessionID,
			ClientTimestamp:  e.ClientTimestamp.UTC().Format(time.RFC3339Nano),
			ServerReceivedAt: e.ServerReceivedAt.UTC().Format(time.RFC3339Nano),
			IngestSequence:   e.IngestSequence,
			SequenceNumber:   e.SequenceNumber,
			SDK: sdkWire{
				Name:    e.SDK.Name,
				Version: e.SDK.Version,
			},
			Meta: e.Meta,
		}
	}
	return out
}

// writeJSON schreibt v als JSON-Response mit Content-Type-Header.
func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
