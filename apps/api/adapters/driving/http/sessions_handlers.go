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
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driving"
)

// SessionsListHandler implementiert GET /api/stream-sessions
// (plan-0.1.0.md §5.1). Span-Konvention siehe telemetry-model.md §2.1
// — Span-Name "http.handler GET /api/stream-sessions" mit
// http.method/http.route/http.status_code-Attributen.
//
// Ab plan-0.4.0 §4.2 (mit dem §4.3-vorgezogenen Auth-Anteil) ist der
// Endpoint tokenpflichtig: ohne `X-MTrace-Token` oder bei unbekanntem
// Token wird `401 Unauthorized` zurückgegeben; der aufgelöste
// `project_id` filtert die Sessions-Liste.
type SessionsListHandler struct {
	UseCase  driving.SessionsInbound
	Resolver driven.ProjectResolver
	Tracer   trace.Tracer
	Logger   *slog.Logger
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

	projectID, ok := resolveProjectFromToken(ctx, w, r, h.Resolver)
	if !ok {
		return
	}

	limit, err := parseLimit(r.URL.Query().Get("limit"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "limit_invalid"})
		return
	}

	cursor, err := decodeListSessionsCursor(r.URL.Query().Get("cursor"), projectID)
	if err != nil {
		writeCursorError(w, err)
		return
	}

	res, err := h.UseCase.ListSessions(ctx, driving.ListSessionsInput{
		ProjectID: projectID,
		Limit:     limit,
		After:     cursor,
	})
	if err != nil {
		h.Logger.Error("ListSessions failed", "error", err)
		writeStatus(w, http.StatusInternalServerError)
		return
	}

	body := struct {
		Sessions   []sessionWire `json:"sessions"`
		NextCursor string        `json:"next_cursor,omitempty"`
	}{
		Sessions: toSessionWireListWithBoundaries(res.Sessions, res.Boundaries),
	}
	if res.NextCursor != nil {
		next, err := encodeListSessionsCursor(res.NextCursor, projectID)
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
// Tokenpflichtig ab plan-0.4.0 §4.2 (siehe SessionsListHandler-Doc).
type SessionsGetHandler struct {
	UseCase  driving.SessionsInbound
	Resolver driven.ProjectResolver
	Tracer   trace.Tracer
	Logger   *slog.Logger
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

	projectID, ok := resolveProjectFromToken(ctx, w, r, h.Resolver)
	if !ok {
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

	cursor, err := decodeSessionEventsCursor(r.URL.Query().Get("events_cursor"), projectID, id)
	if err != nil {
		writeCursorError(w, err)
		return
	}

	res, err := h.UseCase.GetSession(ctx, driving.GetSessionInput{
		ProjectID:   projectID,
		SessionID:   id,
		EventsLimit: limit,
		EventsAfter: cursor,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrSessionNotFound):
			writeStatus(w, http.StatusNotFound)
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
		Session: toSessionWireWithBoundaries(res.Session, res.Boundaries),
		Events:  toEventWireList(res.Events),
	}
	if res.NextCursor != nil {
		next, err := encodeSessionEventsCursor(res.NextCursor, projectID, id)
		if err != nil {
			h.Logger.Error("encode events cursor failed", "error", err)
			writeStatus(w, http.StatusInternalServerError)
			return
		}
		body.NextCursor = next
	}
	writeJSON(w, http.StatusOK, body)
}

// resolveProjectFromToken liest `X-MTrace-Token`, löst ihn über den
// Project-Resolver auf und schreibt im Fehlerfall direkt eine 401-
// Antwort. Rückgabe: (projectID, true) bei Erfolg, ("", false) wenn
// der Caller den Request abbrechen soll. Ab plan-0.4.0 §4.2 sind
// Session-/Event-Read-Endpunkte tokenpflichtig (siehe
// SessionsListHandler/SessionsGetHandler-Doc-Strings).
func resolveProjectFromToken(ctx context.Context, w http.ResponseWriter, r *http.Request, resolver driven.ProjectResolver) (string, bool) {
	token := r.Header.Get("X-MTrace-Token")
	if token == "" {
		writeStatus(w, http.StatusUnauthorized)
		return "", false
	}
	project, err := resolver.ResolveByToken(ctx, token)
	if err != nil {
		writeStatus(w, http.StatusUnauthorized)
		return "", false
	}
	return project.ID, true
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
// `correlation_id` ist ab `0.4.0` (§3.7.1 im API-Kontrakt) Teil der
// Read-Antwort — bei Sessions, die vor §3.2-Closeout angelegt
// wurden, kann es leer sein und wird dann als JSON-`""` ausgeliefert
// (kein `omitempty`, damit Clients das Feld klar als „nicht gesetzt"
// erkennen).
//
// `network_signal_absent` ist ab `0.4.0` (§4.4 D3) im Read-Shape,
// Default `[]`. Reihenfolge: kind asc, adapter asc, reason asc;
// Tripel-Dedup erfolgt im Repository.
type sessionWire struct {
	ID                  string                       `json:"session_id"`
	ProjectID           string                       `json:"project_id"`
	State               string                       `json:"state"`
	StartedAt           string                       `json:"started_at"`
	LastEventAt         string                       `json:"last_event_at"`
	EndedAt             *string                      `json:"ended_at,omitempty"`
	EventCount          int64                        `json:"event_count"`
	CorrelationID       string                       `json:"correlation_id"`
	NetworkSignalAbsent []networkSignalAbsentWire    `json:"network_signal_absent"`
}

// networkSignalAbsentWire ist das Read-Shape eines
// `session_boundaries[]`-Tripels gemäß spec/backend-api-contract.md
// §3.7.1: `kind` ist der Netzwerksignal-Typ
// (`manifest`/`segment`) — nicht der Boundary-`kind`-Wert
// (`network_signal_absent`), der für alle Einträge in dieser Liste
// implizit gilt.
type networkSignalAbsentWire struct {
	Kind    string `json:"kind"`
	Adapter string `json:"adapter"`
	Reason  string `json:"reason"`
}

func toSessionWireWithBoundaries(s domain.StreamSession, boundaries []domain.SessionBoundary) sessionWire {
	out := sessionWire{
		ID:                  s.ID,
		ProjectID:           s.ProjectID,
		State:               string(s.State),
		StartedAt:           s.StartedAt.UTC().Format(time.RFC3339Nano),
		LastEventAt:         s.LastEventAt.UTC().Format(time.RFC3339Nano),
		EventCount:          s.EventCount,
		CorrelationID:       s.CorrelationID,
		NetworkSignalAbsent: toNetworkSignalAbsentWires(boundaries),
	}
	if s.EndedAt != nil {
		ended := s.EndedAt.UTC().Format(time.RFC3339Nano)
		out.EndedAt = &ended
	}
	return out
}

// toSessionWireListWithBoundaries projiziert Sessions parallel zu
// einer Boundary-Slice. `boundaries` darf kürzer/`nil` sein — fehlende
// Indizes erhalten ein leeres `network_signal_absent[]` (Default `[]`).
func toSessionWireListWithBoundaries(in []domain.StreamSession, boundaries [][]domain.SessionBoundary) []sessionWire {
	out := make([]sessionWire, len(in))
	for i, s := range in {
		var b []domain.SessionBoundary
		if i < len(boundaries) {
			b = boundaries[i]
		}
		out[i] = toSessionWireWithBoundaries(s, b)
	}
	return out
}

// toNetworkSignalAbsentWires mappt persistierte Boundaries auf das
// Read-Shape-Tripel (kind/adapter/reason). `domain.SessionBoundary.
// NetworkKind` wird auf das `kind`-Feld der Wire-Antwort projiziert
// (Spec §3.7.1: kind ∈ {manifest, segment}). Default-Rückgabe für
// `nil`/leer ist eine leere Slice (`[]`), damit JSON kein `null` ausgibt.
func toNetworkSignalAbsentWires(in []domain.SessionBoundary) []networkSignalAbsentWire {
	out := make([]networkSignalAbsentWire, 0, len(in))
	for _, b := range in {
		out = append(out, networkSignalAbsentWire{
			Kind:    b.NetworkKind,
			Adapter: b.Adapter,
			Reason:  b.Reason,
		})
	}
	return out
}

// eventWire ist die JSON-Antwortform für ein Event im Detail-Response.
// `correlation_id` und `trace_id` ab `0.4.0` (§3.7 im API-Kontrakt):
//   - `correlation_id` ist Pflichtfeld in 0.4.0+-Read-Antworten (kein
//     `omitempty`); Empty-String nur bei Legacy-Events, die vor
//     §3.2-Closeout persistiert wurden.
//   - `trace_id` ist optional (`omitempty`); fehlt bei Events ohne
//     gültigen Trace-Kontext (Edge-Case: Server-Span ohne Trace-ID).
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
	CorrelationID    string         `json:"correlation_id"`
	TraceID          string         `json:"trace_id,omitempty"`
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
			Meta:          e.Meta,
			CorrelationID: e.CorrelationID,
			TraceID:       e.TraceID,
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

// writeCursorError mappt die Cursor-Fehlerklassen aus cursor.go auf
// HTTP-Status und Body gemäß API-Kontrakt §10.3 / ADR-0004 §6.
// Unbekannte Errors fallen auf 500 zurück (sollte nicht erreichbar
// sein, weil decode nur die drei Klassen liefert).
func writeCursorError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, errCursorInvalidLegacy):
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":  "cursor_invalid_legacy",
			"reason": "process_instance_id-based cursor no longer supported; reload snapshot",
		})
	case errors.Is(err, errCursorInvalidMalformed):
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":  "cursor_invalid_malformed",
			"reason": "cursor token failed to decode or violates v:2 schema",
		})
	case errors.Is(err, errCursorExpired):
		writeJSON(w, http.StatusGone, map[string]string{
			"error":  "cursor_expired",
			"reason": "cursor target no longer in storage; reload snapshot",
		})
	default:
		writeStatus(w, http.StatusInternalServerError)
	}
}
