// Package streamanalyzer enthält den Slot-Füller-Adapter für den
// driven.StreamAnalyzer-Port. Bis ein produktiver Analyzer-Adapter
// (plan-0.3.0 §7 Tranche 6) angeschlossen ist, hält NoopStreamAnalyzer
// die Compile-Time-Garantie für die Use-Case-Verdrahtung aufrecht.
package streamanalyzer

import (
	"context"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// noopAnalyzerVersion markiert Ergebnisse aus dem Slot-Füller, damit
// API-Konsumenten und Tests klar erkennen, dass kein produktiver
// Analyzer angeschlossen ist (plan-0.3.0 §7 Tranche 6).
const noopAnalyzerVersion = "noop"

// NoopStreamAnalyzer erfüllt den Port driven.StreamAnalyzer ohne
// Seiteneffekt. Damit ist die Erweiterungs-Stelle real (mit
// Compile-Time-Garantie via _ = (*NoopStreamAnalyzer)(nil)), ohne dem
// Use Case Produktiv-Last zuzumuten.
type NoopStreamAnalyzer struct{}

// NewNoopStreamAnalyzer gibt einen einsatzbereiten Adapter zurück.
func NewNoopStreamAnalyzer() *NoopStreamAnalyzer {
	return &NoopStreamAnalyzer{}
}

// AnalyzeBatch macht nichts und gibt nil zurück.
func (*NoopStreamAnalyzer) AnalyzeBatch(_ context.Context, _ []domain.PlaybackEvent) error {
	return nil
}

// AnalyzeManifest gibt ein leeres Ergebnis mit Version "noop" zurück.
// Damit bleibt der API-Pfad aufrufbar, bis Tranche 6 den produktiven
// Adapter anschließt; der Aufrufer kann an AnalyzerVersion erkennen,
// dass keine Analyse stattgefunden hat.
func (*NoopStreamAnalyzer) AnalyzeManifest(_ context.Context, _ domain.StreamAnalysisRequest) (domain.StreamAnalysisResult, error) {
	return domain.StreamAnalysisResult{
		AnalyzerVersion: noopAnalyzerVersion,
		PlaylistType:    domain.PlaylistTypeUnknown,
	}, nil
}

// Compile-time check: NoopStreamAnalyzer implements driven.StreamAnalyzer.
var _ driven.StreamAnalyzer = (*NoopStreamAnalyzer)(nil)
