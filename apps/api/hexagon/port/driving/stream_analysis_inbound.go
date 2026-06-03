package driving

import (
	"context"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// StreamAnalysisInbound ist der Driving-Port für den Analyzer-Pfad
// . Der HTTP-Adapter dekodiert die Anfrage,
// reicht sie als domain.StreamAnalysisRequest weiter und mappt das
// Ergebnis auf JSON; Application-Layer bleibt frei von HTTP-Annahmen.
//
// Ab liefert der Port `domain.AnalyzeManifestResult`
// mit `{Analysis, SessionLink}` — der HTTP-Adapter mappt das auf die
// `{analysis, session_link}`-Hülle aus API-Kontrakt
type StreamAnalysisInbound interface {
	AnalyzeManifest(ctx context.Context, req domain.StreamAnalysisRequest) (domain.AnalyzeManifestResult, error)
}
