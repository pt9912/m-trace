package http

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/pt9912/m-trace/apps/api/hexagon/application"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driving"
)

// maxAnalyzeRequestBytes ist die Defense-in-Depth-Grenze für den
// Request-Body im API-Tier. Das eigentliche Manifest-Loader-Limit
// (FetchOptions.maxBytes) gilt zusätzlich beim analyzer-service —
// hier geht es nur darum, dass die Go-API selbst nicht durch
// pathologisch große Anfragen ausgehungert wird.
const maxAnalyzeRequestBytes = 1 * 1024 * 1024

// AnalyzeHandler bedient POST /api/analyze: Manifest-Input → Analyzer-
// Result. Erfolg gibt das vollständige domain.StreamAnalysisResult
// als JSON zurück; Fehler werden in eine Problem-Shape (RFC 7807-nah)
// gemappt.
type AnalyzeHandler struct {
	UseCase driving.StreamAnalysisInbound
	Logger  *slog.Logger
}

type analyzeRequestPayload struct {
	Kind    string `json:"kind"`
	URL     string `json:"url,omitempty"`
	Text    string `json:"text,omitempty"`
	BaseURL string `json:"baseUrl,omitempty"`
}

type analyzeResponseEnvelope struct {
	Status          string                  `json:"status"`
	AnalyzerVersion string                  `json:"analyzerVersion"`
	AnalyzerKind    string                  `json:"analyzerKind"`
	PlaylistType    string                  `json:"playlistType"`
	Findings        []analyzeFindingPayload `json:"findings"`
	Details         json.RawMessage         `json:"details"`
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
		return
	}

	limited := io.LimitReader(r.Body, maxAnalyzeRequestBytes+1)
	raw, err := io.ReadAll(limited)
	if err != nil {
		writeAnalyzeProblem(w, http.StatusBadRequest, "invalid_request", "Request-Body konnte nicht gelesen werden.", nil)
		return
	}
	if int64(len(raw)) > maxAnalyzeRequestBytes {
		writeAnalyzeProblem(w, http.StatusRequestEntityTooLarge, "payload_too_large", "Request-Body überschreitet das API-Limit.", nil)
		return
	}

	var payload analyzeRequestPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		writeAnalyzeProblem(w, http.StatusBadRequest, "invalid_json", "Body ist kein gültiges JSON.", nil)
		return
	}
	req, problem := payloadToRequest(payload)
	if problem != nil {
		writeAnalyzeProblem(w, problem.status, problem.code, problem.message, nil)
		return
	}

	result, err := h.UseCase.AnalyzeManifest(r.Context(), req)
	if err != nil {
		mapAndWriteUseCaseError(w, h.Logger, err)
		return
	}

	resp := analyzeResponseEnvelope{
		Status:          "ok",
		AnalyzerVersion: result.AnalyzerVersion,
		// AnalyzerKind ist heute eine HLS-Konstante. Wenn DASH/CMAF
		// (F-73) eingeführt werden, übernimmt das die Domain (per-kind
		// Result-Variant) und das Feld kommt aus result.AnalyzerKind.
		AnalyzerKind: "hls",
		PlaylistType: string(result.PlaylistType),
		Findings:     findingsToPayload(result.Findings),
	}
	if len(result.EncodedDetails) > 0 {
		resp.Details = json.RawMessage(result.EncodedDetails)
	} else {
		resp.Details = json.RawMessage(`null`)
	}
	if h.Logger != nil {
		h.Logger.Info("analyze ok",
			"playlistType", string(result.PlaylistType),
			"findingCount", len(result.Findings),
			"analyzerVersion", result.AnalyzerVersion,
		)
	}
	writeJSON(w, http.StatusOK, resp)
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
		return domain.StreamAnalysisRequest{ManifestText: p.Text, BaseURL: p.BaseURL}, nil
	case "url":
		if p.URL == "" {
			return domain.StreamAnalysisRequest{}, &validationProblem{
				status: http.StatusBadRequest, code: "invalid_request",
				message: "kind=\"url\" verlangt ein nicht-leeres `url`-Feld.",
			}
		}
		return domain.StreamAnalysisRequest{ManifestURL: p.URL}, nil
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
func mapAndWriteUseCaseError(w http.ResponseWriter, logger *slog.Logger, err error) {
	if errors.Is(err, application.ErrAnalyzeManifestEmpty) {
		logWarn(logger, "analyze rejected: empty input", "code", "invalid_request")
		writeAnalyzeProblem(w, http.StatusBadRequest, "invalid_request", err.Error(), nil)
		return
	}

	var domainErr *domain.StreamAnalysisDomainError
	if errors.As(err, &domainErr) {
		status := domainHTTPStatus(domainErr.Code)
		logWarn(logger, "analyze rejected by analyzer",
			"code", string(domainErr.Code),
			"status", status,
		)
		writeAnalyzeProblem(w, status, string(domainErr.Code), domainErr.Message, domainErr.Details)
		return
	}

	logWarn(logger, "analyzer transport error",
		"code", "analyzer_unavailable",
		"status", http.StatusBadGateway,
		"error", err.Error(),
	)
	writeAnalyzeProblem(w, http.StatusBadGateway, "analyzer_unavailable",
		"Analyzer-Service hat den Aufruf nicht erfolgreich beantwortet.", nil)
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
