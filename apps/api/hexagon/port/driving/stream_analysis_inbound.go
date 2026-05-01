package driving

import (
	"context"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// StreamAnalysisInbound ist der Driving-Port für den Analyzer-Pfad
// (plan-0.3.0 §7 Tranche 6). Der HTTP-Adapter dekodiert die Anfrage,
// reicht sie als domain.StreamAnalysisRequest weiter und mappt das
// domain.StreamAnalysisResult auf JSON; Application-Layer bleibt frei
// von HTTP-Annahmen.
type StreamAnalysisInbound interface {
	AnalyzeManifest(ctx context.Context, req domain.StreamAnalysisRequest) (domain.StreamAnalysisResult, error)
}
