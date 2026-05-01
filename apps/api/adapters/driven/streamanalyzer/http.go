package streamanalyzer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

const (
	defaultRequestTimeout = 30 * time.Second
	// defaultMaxResponseBytes schützt den API-Prozess gegen
	// pathologische Antworten des analyzer-service. Größenlimit auf
	// Manifestebene setzt der Service selbst (FetchOptions.maxBytes,
	// plan-0.3.0 §3); dieses Limit ist zusätzlich Defense-in-Depth.
	defaultMaxResponseBytes = 4 * 1024 * 1024
)

// HTTPStreamAnalyzer ruft den internen analyzer-service über HTTP auf
// (plan-0.3.0 §7 Tranche 6). Der Service akzeptiert das ManifestInput-
// Schema des stream-analyzer-Pakets und liefert das AnalyzeOutput-
// JSON, das der Adapter auf domain.StreamAnalysisResult mappt.
//
// Bewusst keine Auth zwischen API und analyzer-service: die Compose-
// Topologie isoliert den Service auf einem internen Netz; ein
// öffentlich erreichbarer Deploy braucht eine Egress-Firewall oder
// einen separaten Token-Folge-ADR.
type HTTPStreamAnalyzer struct {
	client          *http.Client
	baseURL         string
	maxResponseSize int64
}

// HTTPStreamAnalyzerOption justiert den Adapter beim Konstruieren.
type HTTPStreamAnalyzerOption func(*HTTPStreamAnalyzer)

// WithHTTPClient erlaubt Tests, einen eigenen *http.Client zu
// injizieren (z. B. mit httptest.Server-Round-Tripper).
func WithHTTPClient(c *http.Client) HTTPStreamAnalyzerOption {
	return func(h *HTTPStreamAnalyzer) {
		if c != nil {
			h.client = c
		}
	}
}

// WithMaxResponseBytes überschreibt das Defense-in-Depth-Limit für
// die Antwortgröße.
func WithMaxResponseBytes(n int64) HTTPStreamAnalyzerOption {
	return func(h *HTTPStreamAnalyzer) {
		if n > 0 {
			h.maxResponseSize = n
		}
	}
}

// NewHTTPStreamAnalyzer erzeugt einen Adapter, der gegen baseURL
// (z. B. "http://analyzer-service:7000") spricht.
func NewHTTPStreamAnalyzer(baseURL string, opts ...HTTPStreamAnalyzerOption) *HTTPStreamAnalyzer {
	h := &HTTPStreamAnalyzer{
		client:          &http.Client{Timeout: defaultRequestTimeout},
		baseURL:         strings.TrimRight(baseURL, "/"),
		maxResponseSize: defaultMaxResponseBytes,
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// AnalyzeBatch hält den Slot für F-22 (plan-0.1.0). Die HTTP-Variante
// nutzt ihn nicht produktiv und reicht den Aufruf no-op weiter.
func (*HTTPStreamAnalyzer) AnalyzeBatch(_ context.Context, _ []domain.PlaybackEvent) error {
	return nil
}

// AnalyzeManifest delegiert an den analyzer-service. Fehler werden als
// strukturierte Fehler des Domain-Layers zurückgegeben — das Mapping
// auf HTTP-Statuscodes übernimmt das Driving-HTTP-Adapter (siehe
// plan-0.3.0 §7 Tranche 6, „Fehlerabbildung von Analyzer-Fehlern auf
// HTTP-Status/Problem-Shape").
func (h *HTTPStreamAnalyzer) AnalyzeManifest(ctx context.Context, req domain.StreamAnalysisRequest) (domain.StreamAnalysisResult, error) {
	body, err := buildRequestBody(req)
	if err != nil {
		return domain.StreamAnalysisResult{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, h.baseURL+"/analyze", bytes.NewReader(body))
	if err != nil {
		return domain.StreamAnalysisResult{}, fmt.Errorf("build analyzer request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := h.client.Do(httpReq)
	if err != nil {
		return domain.StreamAnalysisResult{}, fmt.Errorf("call analyzer-service: %w", err)
	}
	defer resp.Body.Close()

	limited := io.LimitReader(resp.Body, h.maxResponseSize+1)
	raw, err := io.ReadAll(limited)
	if err != nil {
		return domain.StreamAnalysisResult{}, fmt.Errorf("read analyzer response: %w", err)
	}
	if int64(len(raw)) > h.maxResponseSize {
		return domain.StreamAnalysisResult{}, fmt.Errorf("analyzer response exceeds %d bytes", h.maxResponseSize)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// analyzer-service liefert bei 4xx/5xx ein Problem-Shape;
		// wir reichen es als Fehler durch, ohne Inhalt zu mutieren.
		return domain.StreamAnalysisResult{}, fmt.Errorf("analyzer-service status %d: %s", resp.StatusCode, truncate(string(raw), 256))
	}

	return parseSuccessResponse(raw)
}

type analyzeRequestBody struct {
	Kind    string `json:"kind"`
	URL     string `json:"url,omitempty"`
	Text    string `json:"text,omitempty"`
	BaseURL string `json:"baseUrl,omitempty"`
}

func buildRequestBody(req domain.StreamAnalysisRequest) ([]byte, error) {
	if req.ManifestURL != "" && req.ManifestText != "" {
		return nil, errors.New("StreamAnalysisRequest: ManifestText und ManifestURL dürfen nicht beide gesetzt sein")
	}
	body := analyzeRequestBody{}
	switch {
	case req.ManifestURL != "":
		body.Kind = "url"
		body.URL = req.ManifestURL
	case req.ManifestText != "":
		body.Kind = "text"
		body.Text = req.ManifestText
		body.BaseURL = req.BaseURL
	default:
		return nil, errors.New("StreamAnalysisRequest: weder ManifestText noch ManifestURL gesetzt")
	}
	return json.Marshal(body)
}

type analyzeResponseEnvelope struct {
	Status          string                  `json:"status"`
	AnalyzerVersion string                  `json:"analyzerVersion"`
	AnalyzerKind    string                  `json:"analyzerKind"`
	PlaylistType    string                  `json:"playlistType"`
	Findings        []analyzeFindingPayload `json:"findings"`
	Details         json.RawMessage         `json:"details"`
	// Fehler-Felder (nur bei status="error" relevant). Das
	// analyzer-service-Wire-Format kapselt Domain-Fehler in 200 +
	// status="error", siehe `@npm9912/stream-analyzer`-Doku §2.3 und
	// `apps/analyzer-service/src/server.ts`.
	Code         string          `json:"code"`
	Message      string          `json:"message"`
	ErrorDetails json.RawMessage `json:"-"`
}

// rawErrorEnvelope ist das Schmal-Schema für Domain-Fehler (`status:
// "error"`, plus optionalen `details`-Block) aus dem analyzer-service.
type rawErrorEnvelope struct {
	Status  string          `json:"status"`
	Code    string          `json:"code"`
	Message string          `json:"message"`
	Details json.RawMessage `json:"details"`
}

type analyzeFindingPayload struct {
	Code    string `json:"code"`
	Level   string `json:"level"`
	Message string `json:"message"`
}

func parseSuccessResponse(raw []byte) (domain.StreamAnalysisResult, error) {
	var env analyzeResponseEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return domain.StreamAnalysisResult{}, fmt.Errorf("decode analyzer response: %w", err)
	}
	if env.Status == "error" {
		// analyzer-service kapselt Domain-Fehler in 200 + status=error.
		// Das ist KEIN Verfügbarkeitsproblem — der Analyzer hat den
		// Aufruf bewusst abgelehnt und einen Code aus einer
		// abgeschlossenen Domäne mitgeliefert. Der Driving-Adapter
		// mappt das anhand des Codes auf den passenden HTTP-Status.
		return domain.StreamAnalysisResult{}, parseDomainError(raw)
	}
	if env.Status != "ok" {
		return domain.StreamAnalysisResult{}, fmt.Errorf("analyzer-service unknown status %q", env.Status)
	}

	result := domain.StreamAnalysisResult{
		AnalyzerVersion: env.AnalyzerVersion,
		PlaylistType:    mapPlaylistType(env.PlaylistType),
		Findings:        mapFindings(env.Findings),
	}
	// `details: null` und fehlendes Feld werden beide als leerer Slice
	// transportiert; jeder andere Inhalt landet als JSON-Bytes in
	// EncodedDetails (Tranche 5: `details` ist diskriminiertes Schema).
	if len(env.Details) > 0 && string(env.Details) != "null" {
		result.EncodedDetails = append([]byte(nil), env.Details...)
	}
	return result, nil
}

func parseDomainError(raw []byte) error {
	var env rawErrorEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return fmt.Errorf("decode analyzer error envelope: %w", err)
	}
	domainErr := &domain.StreamAnalysisDomainError{
		Code:    domain.StreamAnalysisErrorCode(env.Code),
		Message: env.Message,
	}
	if len(env.Details) > 0 && string(env.Details) != "null" {
		var details map[string]any
		if err := json.Unmarshal(env.Details, &details); err == nil {
			domainErr.Details = details
		}
	}
	return domainErr
}

func mapPlaylistType(value string) domain.PlaylistType {
	switch value {
	case "master":
		return domain.PlaylistTypeMaster
	case "media":
		return domain.PlaylistTypeMedia
	default:
		return domain.PlaylistTypeUnknown
	}
}

func mapFindings(in []analyzeFindingPayload) []domain.StreamAnalysisFinding {
	if len(in) == 0 {
		return nil
	}
	out := make([]domain.StreamAnalysisFinding, len(in))
	for i, f := range in {
		out[i] = domain.StreamAnalysisFinding{
			Code:    f.Code,
			Level:   mapFindingLevel(f.Level),
			Message: f.Message,
		}
	}
	return out
}

func mapFindingLevel(value string) domain.FindingLevel {
	switch value {
	case "warning":
		return domain.FindingLevelWarning
	case "error":
		return domain.FindingLevelError
	default:
		return domain.FindingLevelInfo
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// Compile-time check: HTTPStreamAnalyzer implements driven.StreamAnalyzer.
var _ driven.StreamAnalyzer = (*HTTPStreamAnalyzer)(nil)
