package driven

import (
	"context"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// StreamAnalyzer ist der Driven-Port für die Manifestanalyse.
//
// AnalyzeBatch ist die F-22-Architektur-Vorbereitung aus 0.1.0
// (Lastenheft 1.1.3 §7.3). Der Use Case verdrahtet den Slot, ruft ihn
// in 0.1.0/0.2.0 aber nicht produktiv auf.
//
// AnalyzeManifest ist die Zielsignatur ab 0.3.0 (plan-0.3.0 §2
// Tranche 1, RAK-22..RAK-27): die API ruft den Analyzer mit einem
// Manifest-Input und erhält ein domänen-natives Ergebnis. Damit ist
// der Port nicht mehr auf Playback-Event-Batches festgelegt; weitere
// Manifestformate (DASH/CMAF, F-73) sind additiv ergänzbar.
type StreamAnalyzer interface {
	AnalyzeBatch(ctx context.Context, events []domain.PlaybackEvent) error
	AnalyzeManifest(ctx context.Context, req domain.StreamAnalysisRequest) (domain.StreamAnalysisResult, error)
}
