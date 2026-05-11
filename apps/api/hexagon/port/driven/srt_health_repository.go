package driven

import (
	"context"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// SrtHealthRepository persistiert SRT-Health-Samples (spec/architecture.md
// §3.3, spec/backend-api-contract.md §10.6). Implementierungen müssen
// safe für Concurrent-Use sein und die Dedupe-Regel aus §10.6
// einhalten:
//
//   Eindeutigkeit über (project_id, stream_id, connection_id,
//   COALESCE(source_observed_at, source_sequence)).
//
// CollectedAt allein ist kein stabiler Dedupe-Schlüssel.
type SrtHealthRepository interface {
	// Append persistiert eine Liste von Samples. Bei Dedupe-Konflikt
	// (siehe §10.6) wird der existierende Eintrag nicht überschrieben;
	// Aufrufer erhält keinen Fehler — der Repository-Adapter loggt
	// den Skip auf Debug-Level.
	Append(ctx context.Context, samples []domain.SrtHealthSample) error

	// LatestByStream liefert pro StreamID den jüngsten Sample des
	// gegebenen Projects, sortiert nach IngestedAt desc. Eingang für
	// `GET /api/srt/health` (spec §7a.1).
	LatestByStream(ctx context.Context, projectID string) ([]domain.SrtHealthSample, error)

	// HistoryByStream liefert bis zu `Limit` Samples einer
	// (projectID, streamID)-Kombination, sortiert nach IngestedAt
	// desc, ID desc (spec §10.4). After ist nil für die erste Seite;
	// danach trägt der Adapter den nächsten After-Cursor in
	// HistoryPage.NextAfter (spec §7a.3). Eingang für
	// `GET /api/srt/health/{stream_id}` (spec §7a.1).
	//
	// Die Scope-Validierung (Cursor aus Project A im Request für
	// Project B, oder Stream X im Request für Stream Y) lebt im
	// HTTP-Adapter (Wire-Codec gemäß §10.3 v3-Cursor); der Port
	// kennt nur die Storage-Position.
	HistoryByStream(ctx context.Context, q SrtHealthHistoryQuery) (SrtHealthHistoryPage, error)
}

// SrtHealthHistoryQuery ist die Eingabe für HistoryByStream.
// ProjectID und StreamID sind Pflicht; ein Leerwert ist ein
// Programmierfehler. After ist nil für die erste Seite; danach hält
// der Adapter den nächsten After-Cursor in HistoryPage.NextAfter.
type SrtHealthHistoryQuery struct {
	ProjectID string
	StreamID  string
	Limit     int
	After     *SrtHealthCursor
}

// SrtHealthCursor ist die Repository-Sicht auf den Cursor: roh die
// kanonische Sortier-Position `(IngestedAt, ID)` aus spec §7a.3 /
// §10.4. Der Wire-Codec lebt im HTTP-Adapter und ergänzt den
// Collection-Scope `(project_id, stream_id)` gemäß §10.3 v3.
type SrtHealthCursor struct {
	IngestedAt time.Time
	ID         int64
}

// SrtHealthHistoryPage ist die Ausgabe von HistoryByStream.
// NextAfter ist nil, wenn keine Folgeseite existiert.
type SrtHealthHistoryPage struct {
	Items     []domain.SrtHealthSample
	NextAfter *SrtHealthCursor
}
