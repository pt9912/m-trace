package application

import (
	"context"
	"errors"
	"strings"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driving"
)

// ErrAnalyzeManifestEmpty wird zurückgegeben, wenn weder ManifestText
// noch ManifestURL gesetzt ist; HTTP-Mapping: 400.
var ErrAnalyzeManifestEmpty = errors.New("analyze manifest: weder ManifestText noch ManifestURL gesetzt")

// AnalyzeManifestUseCase orchestriert den Analyzer-Aufruf gegen den
// driven Port. Die Implementierung selbst ist dünn — der eigentliche
// Lade- und Parse-Aufwand liegt im analyzer-service hinter dem
// HTTPStreamAnalyzer-Adapter (plan-0.3.0 §7 Tranche 6).
type AnalyzeManifestUseCase struct {
	analyzer driven.StreamAnalyzer
}

// NewAnalyzeManifestUseCase verdrahtet den Use Case mit dem Port.
func NewAnalyzeManifestUseCase(analyzer driven.StreamAnalyzer) *AnalyzeManifestUseCase {
	return &AnalyzeManifestUseCase{analyzer: analyzer}
}

// Compile-time check.
var _ driving.StreamAnalysisInbound = (*AnalyzeManifestUseCase)(nil)

// AnalyzeManifest validiert die Eingabe und delegiert an den Adapter.
// Die Domain-spezifische Fehlerklasse (ErrAnalyzeManifestEmpty) macht
// dem HTTP-Adapter das Mapping auf 400 möglich, ohne dass der Driving-
// Layer die Adapter-Inhalte interpretieren muss.
func (u *AnalyzeManifestUseCase) AnalyzeManifest(ctx context.Context, req domain.StreamAnalysisRequest) (domain.StreamAnalysisResult, error) {
	text := strings.TrimSpace(req.ManifestText)
	url := strings.TrimSpace(req.ManifestURL)
	if text == "" && url == "" {
		return domain.StreamAnalysisResult{}, ErrAnalyzeManifestEmpty
	}
	return u.analyzer.AnalyzeManifest(ctx, req)
}
