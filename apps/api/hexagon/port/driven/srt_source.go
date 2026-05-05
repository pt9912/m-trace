package driven

import (
	"context"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// SrtSource ist der Driven-Port für die SRT-Metrikquelle
// (spec/architecture.md §3.3, plan-0.6.0 §4 Sub-3.2). Adapter
// implementieren das Interface gegen eine konkrete Quelle (MediaMTX-
// Control-API in 0.6.0; Sidecar-Exporter oder libsrt-Binding als
// Folge-Optionen, falls je nötig).
//
// SnapshotConnections liefert den aktuellen Zustand aller bekannten
// SRT-Verbindungen als normalisierte Domain-Samples. Adapter ist für
// die Quellen-spezifische Feld-Übersetzung verantwortlich (z. B.
// MediaMTX `mbpsLinkCapacity` × 1_000_000 → AvailableBandwidthBPS).
//
// Fehler-Konventionen (spec/telemetry-model.md §7.5): der Adapter
// klassifiziert Quellen-Fehler nicht selbst — er gibt entweder die
// erfolgreich gelesenen Samples plus nil zurück oder einen Fehler.
// Der Use Case (SrtHealthCollector) mappt Fehler auf SourceStatus/
// SourceErrorCode beim Erstellen des SrtHealthSample.
type SrtSource interface {
	SnapshotConnections(ctx context.Context) ([]domain.SrtConnectionSample, error)
}
