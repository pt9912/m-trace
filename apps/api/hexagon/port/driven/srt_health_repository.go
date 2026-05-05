package driven

import (
	"context"

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

	// HistoryByStream liefert die letzten `limit` Samples einer
	// (projectID, streamID)-Kombination, sortiert nach IngestedAt
	// desc. Cursor wird vom Adapter erzeugt; nil = erste Seite,
	// nicht-nil = Folgeseite. Eingang für
	// `GET /api/srt/health/{stream_id}` (spec §7a.1, §7a.3).
	//
	// Bei Cursor-Mismatch (process_instance_id) gibt der Adapter
	// ErrCursorInvalid analog zu EventRepository.ListBySession
	// zurück (spec §10.3 / §7a.4).
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

// SrtHealthCursor kapselt die kanonische Sortier-Position
// `(IngestedAt, ID)` aus spec §7a.3 plus die ProcessInstanceID des
// erzeugenden Prozesses. Adapter serialisieren das als opaker Token.
type SrtHealthCursor struct {
	IngestedAt        int64
	ID                int64
	ProcessInstanceID string
}

// SrtHealthHistoryPage ist die Ausgabe von HistoryByStream.
// NextAfter ist nil, wenn keine Folgeseite existiert.
type SrtHealthHistoryPage struct {
	Items     []domain.SrtHealthSample
	NextAfter *SrtHealthCursor
}
