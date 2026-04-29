package driven

import (
	"context"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// StreamAnalyzer ist die Architektur-Vorbereitung für F-22 (Lastenheft
// 1.1.3 §7.3). Der Use Case verdrahtet den Port im Konstruktor; eine
// produktive Implementierung folgt erst ab Phase 0.3.0. Bis dahin
// genügt der NoopStreamAnalyzer aus adapters/driven/streamanalyzer.
//
// Bewusst kein leeres Marker-Interface: in Go würde das von jedem Typ
// erfüllt — der Vertrag wäre wertlos. Stattdessen kleine, no-op-fähige
// Methode, deren Compile-Time-Check echte Erweiterungs-Fixpunkte gibt.
type StreamAnalyzer interface {
	AnalyzeBatch(ctx context.Context, events []domain.PlaybackEvent) error
}
