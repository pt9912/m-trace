package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driving"
)

// maxIngestRequestBytes ist die Defense-in-Depth-Grenze für den
// Request-Body im Ingest-Control-Pfad (`0.11.0` Tranche 2,
// NF-13 / RAK-65..RAK-70). Spec §3.8 §6 Schema-Validierung
// erwartet, dass größere Bodies vor dem JSON-Parser abgeschnitten
// werden.
const maxIngestRequestBytes = 1 * 1024 * 1024

// IngestStreamHandler bedient `/api/ingest/streams`-Endpunkte
// (POST/GET, Detail, Rotate, Validate).
//
// Auth: alle Routen sind tokenpflichtig; `project_id` wird
// serverseitig aus `X-MTrace-Token` abgeleitet. Cross-Project-
// Lookups werden wie nicht-existent behandelt — keine Hinweise auf
// Existenz fremder Streams.
type IngestStreamHandler struct {
	UseCase  driving.IngestControlInbound
	Resolver driven.ProjectResolver
	Logger   *slog.Logger
}

// ServeHTTP implementiert das Pattern-Matching für Pfad+Methode
// innerhalb des Sub-Trees. Methoden-Routing über Go 1.22-Patterns
// erfolgt im NewRouter; hier dispatch'en wir nur auf die Sub-Pfade.
func (h *IngestStreamHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/api/ingest/streams":
		h.handleCreate(w, r)
	case r.Method == http.MethodGet && r.URL.Path == "/api/ingest/streams":
		h.handleList(w, r)
	default:
		http.NotFound(w, r)
	}
}

// IngestStreamDetailHandler bedient die Sub-Routen pro Stream-ID
// (`GET /api/ingest/streams/{id}`,
// `POST /api/ingest/streams/{id}/rotate-key`,
// `POST /api/ingest/streams/{id}/validate-key`).
type IngestStreamDetailHandler struct {
	UseCase  driving.IngestControlInbound
	Resolver driven.ProjectResolver
	Logger   *slog.Logger
}

func (h *IngestStreamDetailHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	streamID := r.PathValue("id")
	if streamID == "" {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		h.handleDetail(w, r, streamID)
	case http.MethodPost:
		// Nicht-Stream-Pfade (rotate-key/validate-key) routen über
		// dedizierte Handler unten — dieser Handler matched nur das
		// Detail-Read.
		http.NotFound(w, r)
	default:
		http.NotFound(w, r)
	}
}

// IngestStreamRotateHandler bedient
// `POST /api/ingest/streams/{id}/rotate-key`.
type IngestStreamRotateHandler struct {
	UseCase  driving.IngestControlInbound
	Resolver driven.ProjectResolver
	Logger   *slog.Logger
}

func (h *IngestStreamRotateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	streamID := r.PathValue("id")
	if streamID == "" {
		http.NotFound(w, r)
		return
	}
	if !ensureContentType(w, r) {
		return
	}
	projectID, ok := h.requireProject(w, r)
	if !ok {
		return
	}
	result, err := h.UseCase.RotateKey(r.Context(), projectID, streamID)
	if err != nil {
		writeIngestError(w, h.Logger, "rotate", err)
		return
	}
	writeJSON(w, http.StatusOK, buildStreamWithKeyPayload(result.Stream, &result.Material))
}

// IngestMediaServerConfigHandler bedient
// `GET /api/ingest/media-server-config` (`0.11.0` Tranche 3,
// RAK-68). Antwort enthält das deterministisch generierte
// MediaMTX-YAML-Artefakt; Klartext-Stream-Keys erscheinen niemals.
type IngestMediaServerConfigHandler struct {
	UseCase  driving.IngestControlInbound
	Resolver driven.ProjectResolver
	Logger   *slog.Logger
}

// ServeHTTP implementiert net/http.Handler.
func (h *IngestMediaServerConfigHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	projectID, ok := resolveIngestProject(w, r, h.Resolver)
	if !ok {
		return
	}
	targetID := strings.TrimSpace(r.URL.Query().Get("target_id"))
	result, err := h.UseCase.GetMediaServerConfig(r.Context(), projectID, targetID)
	if err != nil {
		writeMediaServerConfigError(w, h.Logger, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"target_id":   result.TargetID,
		"kind":        string(result.Kind),
		"config_path": result.ConfigPath,
		"config_yaml": result.ConfigYAML,
		"warnings":    result.Warnings,
	})
}

func writeMediaServerConfigError(w http.ResponseWriter, logger *slog.Logger, err error) {
	switch {
	case errors.Is(err, domain.ErrIngestTargetNotFound):
		writeIngestProblem(w, http.StatusNotFound, "target_not_found", "MediaServerTarget nicht im aufgelösten Project gefunden.")
	default:
		// `ErrMediaMTXConfigNoStreams` und mediamtx-only-Konflikte landen
		// hier — nach Plan §3.8 ist das `503 media_server_config_unavailable`.
		if logger != nil {
			logger.Warn("media-server-config error", "err", err.Error())
		}
		writeIngestProblem(w, http.StatusServiceUnavailable, "media_server_config_unavailable",
			"MediaServer-Konfiguration konnte nicht generiert werden.")
	}
}

// IngestStreamValidateHandler bedient
// `POST /api/ingest/streams/{id}/validate-key`.
type IngestStreamValidateHandler struct {
	UseCase  driving.IngestControlInbound
	Resolver driven.ProjectResolver
	Logger   *slog.Logger
}

func (h *IngestStreamValidateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	streamID := r.PathValue("id")
	if streamID == "" {
		http.NotFound(w, r)
		return
	}
	if !ensureContentType(w, r) {
		return
	}
	projectID, ok := h.requireProject(w, r)
	if !ok {
		return
	}
	body, err := readIngestBody(w, r)
	if err != nil {
		return
	}
	var req struct {
		StreamKey string `json:"stream_key"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		writeIngestProblem(w, http.StatusBadRequest, "invalid_json", "Body ist kein gültiges JSON.")
		return
	}
	result, err := h.UseCase.ValidateKey(r.Context(), projectID, streamID, req.StreamKey)
	if err != nil {
		writeIngestError(w, h.Logger, "validate", err)
		return
	}
	if !result.Valid {
		writeJSON(w, http.StatusOK, map[string]any{"valid": false})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"valid":           true,
		"stream_id":       result.StreamID,
		"key_fingerprint": result.KeyFingerprint,
	})
}

// handleCreate implementiert die sieben-stufige Fehlerreihenfolge aus
// `spec/backend-api-contract.md` §3.8.
func (h *IngestStreamHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	if !ensureContentType(w, r) {
		return
	}
	body, err := readIngestBody(w, r)
	if err != nil {
		return
	}
	projectID, ok := h.requireProject(w, r)
	if !ok {
		return
	}
	var req struct {
		DisplayName string `json:"display_name"`
		Protocol    string `json:"protocol"`
		EndpointID  string `json:"endpoint_id"`
		TargetID    string `json:"target_id"`
		ProjectID   string `json:"project_id,omitempty"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		writeIngestProblem(w, http.StatusBadRequest, "invalid_json", "Body ist kein gültiges JSON.")
		return
	}
	result, err := h.UseCase.CreateStream(r.Context(), driving.CreateStreamRequest{
		ResolvedProjectID: projectID,
		RequestProjectID:  req.ProjectID,
		DisplayName:       req.DisplayName,
		Protocol:          req.Protocol,
		EndpointID:        req.EndpointID,
		TargetID:          req.TargetID,
	})
	if err != nil {
		writeIngestError(w, h.Logger, "create", err)
		return
	}
	writeJSON(w, http.StatusCreated, buildStreamWithKeyPayload(result.Stream, &result.Material))
}

func (h *IngestStreamHandler) handleList(w http.ResponseWriter, r *http.Request) {
	projectID, ok := h.requireProject(w, r)
	if !ok {
		return
	}
	streams, err := h.UseCase.ListStreams(r.Context(), projectID)
	if err != nil {
		writeIngestError(w, h.Logger, "list", err)
		return
	}
	out := struct {
		Streams []map[string]any `json:"streams"`
	}{Streams: make([]map[string]any, 0, len(streams))}
	for _, s := range streams {
		out.Streams = append(out.Streams, buildStreamSummaryPayload(s))
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *IngestStreamDetailHandler) handleDetail(w http.ResponseWriter, r *http.Request, streamID string) {
	projectID, ok := h.requireProject(w, r)
	if !ok {
		return
	}
	detail, err := h.UseCase.GetStreamDetail(r.Context(), projectID, streamID)
	if err != nil {
		writeIngestError(w, h.Logger, "detail", err)
		return
	}
	writeJSON(w, http.StatusOK, buildStreamDetailPayload(detail))
}

func (h *IngestStreamHandler) requireProject(w http.ResponseWriter, r *http.Request) (string, bool) {
	return resolveIngestProject(w, r, h.Resolver)
}

func (h *IngestStreamDetailHandler) requireProject(w http.ResponseWriter, r *http.Request) (string, bool) {
	return resolveIngestProject(w, r, h.Resolver)
}

func (h *IngestStreamRotateHandler) requireProject(w http.ResponseWriter, r *http.Request) (string, bool) {
	return resolveIngestProject(w, r, h.Resolver)
}

func (h *IngestStreamValidateHandler) requireProject(w http.ResponseWriter, r *http.Request) (string, bool) {
	return resolveIngestProject(w, r, h.Resolver)
}

func resolveIngestProject(w http.ResponseWriter, r *http.Request, resolver driven.ProjectResolver) (string, bool) {
	token := r.Header.Get("X-MTrace-Token")
	if token == "" || resolver == nil {
		writeIngestProblem(w, http.StatusUnauthorized, "unauthorized",
			"`X-MTrace-Token` fehlt oder ist ungültig.")
		return "", false
	}
	project, err := resolver.ResolveByToken(r.Context(), token)
	if err != nil {
		writeIngestProblem(w, http.StatusUnauthorized, "unauthorized",
			"`X-MTrace-Token` ist ungültig.")
		return "", false
	}
	return project.ID, true
}

func ensureContentType(w http.ResponseWriter, r *http.Request) bool {
	contentType := strings.TrimSpace(r.Header.Get("Content-Type"))
	mainType := strings.ToLower(strings.TrimSpace(strings.SplitN(contentType, ";", 2)[0]))
	if mainType != "application/json" {
		writeIngestProblem(w, http.StatusUnsupportedMediaType, "unsupported_media_type",
			"Content-Type muss application/json sein.")
		return false
	}
	return true
}

func readIngestBody(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	limited := io.LimitReader(r.Body, maxIngestRequestBytes+1)
	raw, err := io.ReadAll(limited)
	if err != nil {
		writeIngestProblem(w, http.StatusBadRequest, "invalid_request", "Request-Body konnte nicht gelesen werden.")
		return nil, err
	}
	if int64(len(raw)) > maxIngestRequestBytes {
		writeIngestProblem(w, http.StatusRequestEntityTooLarge, "payload_too_large",
			"Request-Body überschreitet das API-Limit.")
		return nil, errors.New("payload too large")
	}
	return raw, nil
}

// writeIngestError mappt Domain-/Validation-Fehler auf HTTP-Status
// + JSON-Body. Wichtig: ausschließlich `errors.Is`-Branches —
// String-Heuristiken sind verboten, weil gewrappte Repo-Fehler
// (z. B. `fmt.Errorf("ingest: insert ...: %w", err)` aus dem
// SQLite-Adapter) sonst fälschlich als `400 invalid_request`
// rauskämen statt als `500 internal_error`.
func writeIngestError(w http.ResponseWriter, logger *slog.Logger, op string, err error) {
	switch {
	case errors.Is(err, domain.ErrIngestProtocolUnknown),
		errors.Is(err, domain.ErrIngestDisplayNameRequired):
		writeIngestProblem(w, http.StatusBadRequest, "invalid_request", err.Error())
	case errors.Is(err, domain.ErrIngestProjectIDMismatch):
		writeIngestProblem(w, http.StatusBadRequest, "project_id_mismatch", err.Error())
	case errors.Is(err, domain.ErrIngestStreamNameConflict):
		writeIngestProblem(w, http.StatusConflict, "stream_name_conflict", err.Error())
	case errors.Is(err, domain.ErrIngestRoutingRuleDisabled):
		writeIngestProblem(w, http.StatusConflict, "routing_rule_disabled", err.Error())
	case errors.Is(err, domain.ErrIngestStreamNotFound):
		writeIngestProblem(w, http.StatusNotFound, "stream_not_found",
			"Ingest-Stream existiert nicht im aufgelösten Project.")
	case errors.Is(err, domain.ErrIngestEndpointNotFound):
		writeIngestProblem(w, http.StatusNotFound, "endpoint_not_found",
			"Referenzierter Ingest-Endpunkt fehlt.")
	case errors.Is(err, domain.ErrIngestTargetNotFound):
		writeIngestProblem(w, http.StatusNotFound, "target_not_found",
			"Referenziertes MediaServerTarget fehlt.")
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		writeIngestProblem(w, http.StatusServiceUnavailable, "service_unavailable",
			"Anfrage wurde abgebrochen.")
	default:
		if logger != nil {
			logger.Warn("ingest handler error", "op", op, "err", err.Error())
		}
		writeIngestProblem(w, http.StatusInternalServerError, "internal_error",
			"Ingest-Operation fehlgeschlagen.")
	}
}

func writeIngestProblem(w http.ResponseWriter, status int, code, message string) {
	body := map[string]any{
		"status":  "error",
		"code":    code,
		"message": message,
	}
	writeJSON(w, status, body)
}

func buildStreamWithKeyPayload(stream domain.IngestStream, material *domain.StreamKeyMaterial) map[string]any {
	out := map[string]any{
		"stream": buildStreamCorePayload(stream),
	}
	if material != nil {
		out["stream_key"] = map[string]any{
			"value":       material.Value,
			"fingerprint": material.Fingerprint,
			"created_at":  material.CreatedAt.UTC().Format(time.RFC3339Nano),
		}
	}
	return out
}

func buildStreamSummaryPayload(stream domain.IngestStream) map[string]any {
	out := buildStreamCorePayload(stream)
	out["key_fingerprint"] = stream.Key.Fingerprint
	return out
}

func buildStreamCorePayload(stream domain.IngestStream) map[string]any {
	return map[string]any{
		"id":              stream.ID,
		"project_id":      stream.ProjectID,
		"display_name":    stream.DisplayName,
		"protocol":        string(stream.Protocol),
		"endpoint_id":     stream.EndpointID,
		"target_id":       stream.TargetID,
		"routing_rule_id": stream.RoutingRuleID,
		"status":          string(stream.Status),
		"created_at":      stream.CreatedAt.UTC().Format(time.RFC3339Nano),
		"updated_at":      stream.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func buildStreamDetailPayload(detail driving.StreamDetail) map[string]any {
	return map[string]any{
		"stream": buildStreamSummaryPayload(detail.Stream),
		"endpoint": map[string]any{
			"id":              detail.Endpoint.ID,
			"protocol":        string(detail.Endpoint.Protocol),
			"listen_host":     detail.Endpoint.ListenHost,
			"listen_port":     detail.Endpoint.ListenPort,
			"path_template":   detail.Endpoint.PathTemplate,
			"lab_stack":       detail.Endpoint.LabStack,
			"public_url_hint": detail.Endpoint.PublicURLHint,
		},
		"routing_rule": map[string]any{
			"id":        detail.RoutingRule.ID,
			"stream_id": detail.RoutingRule.StreamID,
			"target_id": detail.RoutingRule.TargetID,
			"mode":      string(detail.RoutingRule.Mode),
			"enabled":   detail.RoutingRule.Enabled,
		},
		"target": map[string]any{
			"id":               detail.Target.ID,
			"kind":             string(detail.Target.Kind),
			"config_path":      detail.Target.ConfigPath,
			"hls_url_template": detail.Target.HLSURLTemplate,
			"control_api_url":  detail.Target.ControlAPIURL,
		},
	}
}
