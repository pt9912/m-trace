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
		AnalyzerKind:    "hls",
		PlaylistType:    string(result.PlaylistType),
		Findings:        findingsToPayload(result.Findings),
	}
	if len(result.EncodedDetails) > 0 {
		resp.Details = json.RawMessage(result.EncodedDetails)
	} else {
		resp.Details = json.RawMessage(`null`)
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
// eine Problem-Shape mit dem passenden HTTP-Statuscode. Die
// Klassifikation ist absichtlich konservativ: Eingabe-Fehler → 400/415/413,
// alles aus dem analyzer-service-Pfad → 502 (Bad Gateway), unbekannte
// Fehler → 500.
func mapAndWriteUseCaseError(w http.ResponseWriter, logger *slog.Logger, err error) {
	if errors.Is(err, application.ErrAnalyzeManifestEmpty) {
		writeAnalyzeProblem(w, http.StatusBadRequest, "invalid_request", err.Error(), nil)
		return
	}
	if logger != nil {
		logger.Warn("analyze manifest failed", "error", err)
	}
	writeAnalyzeProblem(w, http.StatusBadGateway, "analyzer_unavailable", "Analyzer-Service hat den Aufruf nicht erfolgreich beantwortet.", map[string]any{
		"reason": err.Error(),
	})
}

func writeAnalyzeProblem(w http.ResponseWriter, status int, code, message string, details map[string]any) {
	body := analyzeProblemBody{Status: "error", Code: code, Message: message}
	if len(details) > 0 {
		body.Details = details
	}
	writeJSON(w, status, body)
}
