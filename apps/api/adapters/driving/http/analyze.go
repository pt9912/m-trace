package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/pt9912/m-trace/apps/api/hexagon/application"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driving"
)

// maxAnalyzeRequestBytes ist die Defense-in-Depth-Grenze für den
// Request-Body im API-Tier. Das eigentliche Manifest-Loader-Limit
// (FetchOptions.maxBytes) gilt zusätzlich beim analyzer-service —
// hier geht es nur darum, dass die Go-API selbst nicht durch
// pathologisch große Anfragen ausgehungert wird.
const maxAnalyzeRequestBytes = 1 * 1024 * 1024

// AnalyzeMetrics ist die schmale Metrik-Schnittstelle, die der
// AnalyzeHandler braucht. Implementierungen erhöhen einen Counter
// `outcome` ∈ {ok, error} × `code` ∈ {ok, invalid_request,
// analyzer_unavailable, …}. Cardinality bleibt damit beschränkt
// (plan-0.3.0 §9 Tranche 7.5).
type AnalyzeMetrics interface {
	AnalyzeRequest(outcome, code string)
}

// AnalyzeHandler bedient POST /api/analyze: Manifest-Input → Analyzer-
// Result mit optionaler Session-Verknüpfung. Erfolg liefert die
// `{analysis, session_link}`-Hülle aus API-Kontrakt §3.6 als JSON;
// Fehler werden in eine Problem-Shape (RFC 7807-nah) gemappt.
//
// Auth ist endpoint-spezifisch (API-Kontrakt §4): Requests ohne
// `correlation_id` und ohne `session_id` brauchen kein Token; mit
// einem der beiden Felder ist `X-MTrace-Token` Pflicht und muss auf
// ein bekanntes Project resolvieren — sonst 401, ohne Use-Case-Aufruf.
type AnalyzeHandler struct {
	UseCase  driving.StreamAnalysisInbound
	Resolver driven.ProjectResolver
	Logger   *slog.Logger
	// Metrics ist optional — nil bedeutet "nicht zählen". Tests, die
	// den Counter nicht beobachten wollen, können das Feld weglassen.
	Metrics AnalyzeMetrics
}

type analyzeRequestPayload struct {
	Kind          string `json:"kind"`
	URL           string `json:"url,omitempty"`
	Text          string `json:"text,omitempty"`
	BaseURL       string `json:"baseUrl,omitempty"`
	CorrelationID string `json:"correlation_id,omitempty"`
	SessionID     string `json:"session_id,omitempty"`
}

// analyzeAnalysisPayload spiegelt das `analysis`-Feld der Tranche-3-
// Wrapper-Antwort (API-Kontrakt §3.6). Inhaltlich identisch zum pre-
// §4.5 flachen Wire-Format — aber jetzt unterhalb von `{analysis: ...,
// session_link: ...}`.
type analyzeAnalysisPayload struct {
	AnalyzerVersion string                  `json:"analyzerVersion"`
	AnalyzerKind    string                  `json:"analyzerKind"`
	Input           analyzeInputPayload     `json:"input"`
	PlaylistType    string                  `json:"playlistType"`
	Summary         analyzeSummaryPayload   `json:"summary"`
	Findings        []analyzeFindingPayload `json:"findings"`
	Details         json.RawMessage         `json:"details"`
}

// analyzeSessionLinkPayload ist die Wire-Hülle für `session_link`
// (API-Kontrakt §3.6). Optional-Felder werden bei `status != linked`
// nicht ausgegeben (`omitempty`).
type analyzeSessionLinkPayload struct {
	Status        string `json:"status"`
	ProjectID     string `json:"project_id,omitempty"`
	SessionID     string `json:"session_id,omitempty"`
	CorrelationID string `json:"correlation_id,omitempty"`
}

// analyzeResponseEnvelope ist die Tranche-3-Wrapper-Antwort: alle
// erfolgreichen `POST /api/analyze`-Antworten tragen sie, auch
// detached. `Status="ok"` bleibt für Backward-Compat-Logik.
type analyzeResponseEnvelope struct {
	Status      string                    `json:"status"`
	Analysis    analyzeAnalysisPayload    `json:"analysis"`
	SessionLink analyzeSessionLinkPayload `json:"session_link"`
}

type analyzeInputPayload struct {
	Source  string `json:"source"`
	URL     string `json:"url,omitempty"`
	BaseURL string `json:"baseUrl,omitempty"`
}

type analyzeSummaryPayload struct {
	ItemCount int `json:"itemCount"`
}

type analyzeFindingPayload struct {
	Code    string `json:"code"`
	Level   string `json:"level"`
	Message string `json:"message"`
}

type analyzeProblemBody struct {
	Status  string         `json:"status"`
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

// ServeHTTP implementiert net/http.Handler.
func (h *AnalyzeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	contentType := strings.TrimSpace(r.Header.Get("Content-Type"))
	mainType := strings.ToLower(strings.TrimSpace(strings.SplitN(contentType, ";", 2)[0]))
	if mainType != "application/json" {
		writeAnalyzeProblem(w, http.StatusUnsupportedMediaType, "unsupported_media_type", "Content-Type muss application/json sein.", nil)
		h.recordOutcome("error", "unsupported_media_type")
		return
	}

	limited := io.LimitReader(r.Body, maxAnalyzeRequestBytes+1)
	raw, err := io.ReadAll(limited)
	if err != nil {
		writeAnalyzeProblem(w, http.StatusBadRequest, "invalid_request", "Request-Body konnte nicht gelesen werden.", nil)
		h.recordOutcome("error", "invalid_request")
		return
	}
	if int64(len(raw)) > maxAnalyzeRequestBytes {
		writeAnalyzeProblem(w, http.StatusRequestEntityTooLarge, "payload_too_large", "Request-Body überschreitet das API-Limit.", nil)
		h.recordOutcome("error", "payload_too_large")
		return
	}

	var payload analyzeRequestPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		writeAnalyzeProblem(w, http.StatusBadRequest, "invalid_json", "Body ist kein gültiges JSON.", nil)
		h.recordOutcome("error", "invalid_json")
		return
	}
	req, problem := payloadToRequest(payload)
	if problem != nil {
		writeAnalyzeProblem(w, problem.status, problem.code, problem.message, nil)
		h.recordOutcome("error", problem.code)
		return
	}

	// plan-0.4.0 §4.5 — endpoint-spezifische Auth (API-Kontrakt §4):
	// nur Requests mit `correlation_id` oder `session_id` brauchen
	// einen gültigen `X-MTrace-Token`. Ungebundene Requests bleiben
	// auch ohne Token erfolgreich und liefern `session_link.status=
	// detached`.
	if req.CorrelationID != "" || req.SessionID != "" {
		projectID, ok := h.resolveProjectForLinkedRequest(r.Context(), w, r)
		if !ok {
			return
		}
		req.ProjectID = projectID
	}

	envelope, err := h.UseCase.AnalyzeManifest(r.Context(), req)
	if err != nil {
		h.mapAndWriteUseCaseError(w, err)
		return
	}

	resp := buildAnalyzeResponse(envelope)
	if h.Logger != nil {
		h.Logger.Info("analyze ok",
			"playlistType", string(envelope.Analysis.PlaylistType),
			"findingCount", len(envelope.Analysis.Findings),
			"analyzerVersion", envelope.Analysis.AnalyzerVersion,
			"sessionLinkStatus", string(envelope.SessionLink.Status),
		)
	}
	h.recordOutcome("ok", "ok")
	writeJSON(w, http.StatusOK, resp)
}

// resolveProjectForLinkedRequest erzwingt für Analyze-Requests mit
// gesetzten Link-Feldern den Token-Pflicht-Pfad: fehlender Token,
// unbekannter Token oder fehlender Resolver liefert 401 ohne
// Use-Case-Aufruf (API-Kontrakt §4-Regel "kein Session-Lookup ohne
// Project").
func (h *AnalyzeHandler) resolveProjectForLinkedRequest(
	ctx context.Context, w http.ResponseWriter, r *http.Request,
) (string, bool) {
	token := r.Header.Get("X-MTrace-Token")
	if token == "" || h.Resolver == nil {
		writeAnalyzeProblem(w, http.StatusUnauthorized, "unauthorized",
			"Analyze-Requests mit `correlation_id` oder `session_id` benötigen einen gültigen `X-MTrace-Token`.", nil)
		h.recordOutcome("error", "unauthorized")
		return "", false
	}
	project, err := h.Resolver.ResolveByToken(ctx, token)
	if err != nil {
		writeAnalyzeProblem(w, http.StatusUnauthorized, "unauthorized",
			"`X-MTrace-Token` ist ungültig.", nil)
		h.recordOutcome("error", "unauthorized")
		return "", false
	}
	return project.ID, true
}

// buildAnalyzeResponse mappt das Use-Case-Result auf die
// `{analysis, session_link}`-Wrapper-Antwort aus API-Kontrakt §3.6.
// Für alle erfolgreichen Requests, auch detached.
func buildAnalyzeResponse(envelope domain.AnalyzeManifestResult) analyzeResponseEnvelope {
	result := envelope.Analysis
	analysis := analyzeAnalysisPayload{
		AnalyzerVersion: result.AnalyzerVersion,
		// AnalyzerKind ist heute eine HLS-Konstante. Wenn DASH/CMAF
		// (F-73) eingeführt werden, übernimmt das die Domain
		// (per-kind Result-Variant) und das Feld kommt aus
		// result.AnalyzerKind.
		AnalyzerKind: "hls",
		Input: analyzeInputPayload{
			Source:  result.Input.Source,
			URL:     result.Input.URL,
			BaseURL: result.Input.BaseURL,
		},
		PlaylistType: string(result.PlaylistType),
		Summary:      analyzeSummaryPayload{ItemCount: result.Summary.ItemCount},
		Findings:     findingsToPayload(result.Findings),
	}
	if len(result.EncodedDetails) > 0 {
		analysis.Details = json.RawMessage(result.EncodedDetails)
	} else {
		analysis.Details = json.RawMessage(`null`)
	}
	link := analyzeSessionLinkPayload{Status: string(envelope.SessionLink.Status)}
	if envelope.SessionLink.Status == domain.SessionLinkStatusLinked {
		link.ProjectID = envelope.SessionLink.ProjectID
		link.SessionID = envelope.SessionLink.SessionID
		link.CorrelationID = envelope.SessionLink.CorrelationID
	}
	return analyzeResponseEnvelope{
		Status:      "ok",
		Analysis:    analysis,
		SessionLink: link,
	}
}

func (h *AnalyzeHandler) recordOutcome(outcome, code string) {
	if h.Metrics != nil {
		h.Metrics.AnalyzeRequest(outcome, code)
	}
}

type validationProblem struct {
	status  int
	code    string
	message string
}

func payloadToRequest(p analyzeRequestPayload) (domain.StreamAnalysisRequest, *validationProblem) {
	switch p.Kind {
	case "text":
		if p.Text == "" {
			return domain.StreamAnalysisRequest{}, &validationProblem{
				status: http.StatusBadRequest, code: "invalid_request",
				message: "kind=\"text\" verlangt ein nicht-leeres `text`-Feld.",
			}
		}
		return domain.StreamAnalysisRequest{
			ManifestText:  p.Text,
			BaseURL:       p.BaseURL,
			CorrelationID: p.CorrelationID,
			SessionID:     p.SessionID,
		}, nil
	case "url":
		if p.URL == "" {
			return domain.StreamAnalysisRequest{}, &validationProblem{
				status: http.StatusBadRequest, code: "invalid_request",
				message: "kind=\"url\" verlangt ein nicht-leeres `url`-Feld.",
			}
		}
		return domain.StreamAnalysisRequest{
			ManifestURL:   p.URL,
			CorrelationID: p.CorrelationID,
			SessionID:     p.SessionID,
		}, nil
	default:
		return domain.StreamAnalysisRequest{}, &validationProblem{
			status: http.StatusBadRequest, code: "invalid_request",
			message: "Body muss `kind`=\"text\" oder `kind`=\"url\" angeben.",
		}
	}
}

func findingsToPayload(in []domain.StreamAnalysisFinding) []analyzeFindingPayload {
	if len(in) == 0 {
		return []analyzeFindingPayload{}
	}
	out := make([]analyzeFindingPayload, len(in))
	for i, f := range in {
		out[i] = analyzeFindingPayload{
			Code:    f.Code,
			Level:   string(f.Level),
			Message: f.Message,
		}
	}
	return out
}

// mapAndWriteUseCaseError übersetzt Use-Case- und Adapter-Fehler in
// eine Problem-Shape mit dem passenden HTTP-Statuscode (plan-0.3.0 §7
// Tranche 6). Drei Fehlerklassen werden unterschieden:
//
//  1. Eingabevalidierung gegen den Use Case (ErrAnalyzeManifestEmpty)
//     → 400 invalid_request.
//  2. Domain-Fehler vom Analyzer (StreamAnalysisDomainError) — der
//     Analyzer hat den Aufruf bewusst und mit Code abgelehnt. Mapping
//     je Code: invalid_input/fetch_blocked → 400, manifest_not_hls →
//     422, fetch_failed/manifest_too_large/internal_error → 502.
//  3. Transportfehler (HTTP-Status, Timeout, JSON-Decode) → 502
//     analyzer_unavailable. Die Adapter-Fehler-Message bleibt
//     bewusst aus dem Antwort-Body (Info-Leak); sie wird strukturiert
//     im Log abgelegt.
func (h *AnalyzeHandler) mapAndWriteUseCaseError(w http.ResponseWriter, err error) {
	if errors.Is(err, application.ErrAnalyzeManifestEmpty) {
		logWarn(h.Logger, "analyze rejected: empty input", "code", "invalid_request")
		writeAnalyzeProblem(w, http.StatusBadRequest, "invalid_request", err.Error(), nil)
		h.recordOutcome("error", "invalid_request")
		return
	}

	var domainErr *domain.StreamAnalysisDomainError
	if errors.As(err, &domainErr) {
		status := domainHTTPStatus(domainErr.Code)
		logWarn(h.Logger, "analyze rejected by analyzer",
			"code", string(domainErr.Code),
			"status", status,
		)
		writeAnalyzeProblem(w, status, string(domainErr.Code), domainErr.Message, domainErr.Details)
		h.recordOutcome("error", string(domainErr.Code))
		return
	}

	logWarn(h.Logger, "analyzer transport error",
		"code", "analyzer_unavailable",
		"status", http.StatusBadGateway,
		"error", err.Error(),
	)
	writeAnalyzeProblem(w, http.StatusBadGateway, "analyzer_unavailable",
		"Analyzer-Service hat den Aufruf nicht erfolgreich beantwortet.", nil)
	h.recordOutcome("error", "analyzer_unavailable")
}

func domainHTTPStatus(code domain.StreamAnalysisErrorCode) int {
	switch code {
	case domain.StreamAnalysisInvalidInput, domain.StreamAnalysisFetchBlocked:
		return http.StatusBadRequest
	case domain.StreamAnalysisManifestNotHLS:
		return http.StatusUnprocessableEntity
	case domain.StreamAnalysisFetchFailed, domain.StreamAnalysisManifestTooLarge, domain.StreamAnalysisInternalError:
		return http.StatusBadGateway
	default:
		// Unbekannter Code → konservativ als Gateway-Fehler behandeln,
		// damit ein zukünftiger Analyzer-Code, den wir noch nicht
		// kennen, nicht versehentlich als 4xx (Client-Fehler) gemeldet
		// wird.
		return http.StatusBadGateway
	}
}

func logWarn(logger *slog.Logger, msg string, args ...any) {
	if logger == nil {
		return
	}
	logger.Warn(msg, args...)
}

func writeAnalyzeProblem(w http.ResponseWriter, status int, code, message string, details map[string]any) {
	body := analyzeProblemBody{Status: "error", Code: code, Message: message}
	if len(details) > 0 {
		body.Details = details
	}
	writeJSON(w, status, body)
}
