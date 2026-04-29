// Package streamanalyzer hält die F-22-Architektur-Vorbereitung. Der
// produktive Pfad kommt frühestens in Phase 0.3.0; bis dahin reicht
// der hier implementierte NoopStreamAnalyzer als Slot-Füller im Use
// Case (siehe plan-0.1.0.md §5.1, F-22-Item).
package streamanalyzer

import (
	"context"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

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

// Compile-time check: NoopStreamAnalyzer implements driven.StreamAnalyzer.
var _ driven.StreamAnalyzer = (*NoopStreamAnalyzer)(nil)
