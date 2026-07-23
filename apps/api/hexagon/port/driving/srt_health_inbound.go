package driving

import (
	"context"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// SrtHealthSummary ist die abgeleitete Read-Sicht auf ein SRT-Health-Sample
// (Sample + freshness-/derived-Felder). Driving-Port-Typ (slice-004), damit
// der HTTP-Adapter nicht die Application-Schicht importiert.
type SrtHealthSummary struct {
	Sample            domain.SrtHealthSample
	BandwidthHeadroom *float64 // available / required, falls required vorhanden
	SampleAgeMillis   int64    // Zeit seit IngestedAt zum Lesezeitpunkt
	StaleAfterMillis  int64    // Schwelle, ab der ein Sample als stale gilt
}

// SrtHealthHistoryItem trägt die Detail-Sicht: dieselben Felder wie Summary.
type SrtHealthHistoryItem = SrtHealthSummary

// SrtHealthHistoryPage bündelt die Detail-Items einer Page plus den optionalen
// Folge-Cursor. Wire-Codec für `samples_cursor`/`next_cursor` lebt im HTTP-Adapter.
type SrtHealthHistoryPage struct {
	Items     []SrtHealthHistoryItem
	NextAfter *driven.SrtHealthCursor
}

// SrtHealthInbound ist der Driving-Port für die zwei SRT-Health-Read-Operationen.
// Implementiert von application.SrtHealthQueryService.
type SrtHealthInbound interface {
	LatestByStream(ctx context.Context, projectID string) ([]SrtHealthSummary, error)
	HistoryByStream(ctx context.Context, projectID, streamID string, limit int, after *driven.SrtHealthCursor) (SrtHealthHistoryPage, error)
}
