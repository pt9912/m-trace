package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"sync/atomic"

	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// IngestSequencer hält den ingest_sequence-Counter im RAM, lädt aber
// beim Konstruktor den höchsten persistierten Wert aus der
// playback_events-Tabelle. Damit wird der Sequencer nach API-Restart
// konsistent fortgesetzt — Cursor-Stabilität (ADR-0004 §5) verlässt
// sich auf diese Eigenschaft.
//
// Adapter-Inserts persistieren den vom Use Case zugewiesenen Wert
// explizit (`INSERT INTO playback_events(ingest_sequence, ...)`); das
// AUTOINCREMENT auf der PK-Spalte wirkt zusätzlich als Defense-in-Depth
// gegen Reuse nach Rollback.
//
// Safe für nebenläufige Aufrufe.
type IngestSequencer struct {
	counter atomic.Int64
}

// NewIngestSequencer initialisiert den Counter aus
// `SELECT MAX(ingest_sequence) FROM playback_events`. Ist die Tabelle
// leer, beginnt Next() bei 1.
func NewIngestSequencer(ctx context.Context, db *sql.DB) (*IngestSequencer, error) {
	var maxSeq sql.NullInt64
	if err := db.QueryRowContext(ctx,
		"SELECT MAX(ingest_sequence) FROM playback_events").Scan(&maxSeq); err != nil {
		return nil, fmt.Errorf("sqlite: read max(ingest_sequence): %w", err)
	}
	s := &IngestSequencer{}
	if maxSeq.Valid {
		s.counter.Store(maxSeq.Int64)
	}
	return s, nil
}

// Next gibt die nächste ingest_sequence zurück. Atomar inkrementiert.
func (s *IngestSequencer) Next() int64 {
	return s.counter.Add(1)
}

var _ driven.IngestSequencer = (*IngestSequencer)(nil)
