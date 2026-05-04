package driving

import (
	"context"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// StreamAnalysisInbound ist der Driving-Port für den Analyzer-Pfad
// (plan-0.3.0 §7 Tranche 6). Der HTTP-Adapter dekodiert die Anfrage,
// reicht sie als domain.StreamAnalysisRequest weiter und mappt das
// Ergebnis auf JSON; Application-Layer bleibt frei von HTTP-Annahmen.
//
// Ab plan-0.4.0 §4.5 liefert der Port `domain.AnalyzeManifestResult`
// mit `{Analysis, SessionLink}` — der HTTP-Adapter mappt das auf die
// `{analysis, session_link}`-Hülle aus API-Kontrakt §3.6.
type StreamAnalysisInbound interface {
	AnalyzeManifest(ctx context.Context, req domain.StreamAnalysisRequest) (domain.AnalyzeManifestResult, error)
}
