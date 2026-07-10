// Package postgres enthält die Postgres-Persistenz-Adapter für die
// Driven-Ports (ADR-0006, optionaler Scale-out-Store). SQLite bleibt
// Default; dieser Adapter wird nur bei MTRACE_PERSISTENCE=postgres
// verdrahtet.
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

const (
	// defaultBlockSize ist die Anzahl ingest_sequence-Werte, die pro
	// DB-Roundtrip reserviert werden (Amortisierung des per-Event-
	// nextval-Roundtrips, R-28-Perf-Vorbehalt). Ein Block ist per
	// Replica reserviert; Werte über Replicas bleiben eindeutig, weil
	// jeder nextval DB-atomar einen frischen Bereich vergibt.
	defaultBlockSize = 512

	// reserveTimeout begrenzt einen einzelnen Block-Reserve-Roundtrip
	// im Next()-Nachfüllpfad (Next() trägt selbst keinen Context).
	reserveTimeout = 2 * time.Second

	// reserveMaxAttempts ist die Zahl der Nachfüll-Versuche, bevor der
	// Sequencer fail-fast panickt (s. Next()).
	reserveMaxAttempts = 3
)

// IngestSequencer ist die **DB-autoritative** Postgres-Variante des
// driven.IngestSequencer (R-28-Redesign, kein Dialekt-Spiegel). Statt
// eines In-Process-`MAX`-Seeds + `atomic.Add` (der über Replicas
// dieselben Werte vergäbe → PK-Kollisionen) zieht sie die Werte aus der
// DB-Sequence hinter playback_events.ingest_sequence via `nextval` —
// jeder Wert ist damit global eindeutig, auch über mehrere API-Replicas.
//
// Port-erhaltend: `Next() int64` bleibt unverändert (SQLite-/InMemory-
// Impl + Call-Site register_playback_event_batch.go unangetastet).
// `IDENTITY`+`RETURNING` ist bewusst vermieden — der Wert wäre erst
// nach dem INSERT bekannt und bräche den Pre-Assign-Flow.
//
// Gegen den per-Event-Roundtrip reserviert der Sequencer Blöcke
// (defaultBlockSize Werte je Roundtrip) und puffert sie; Next() bedient
// aus dem Puffer und füllt nur bei leerem Puffer nach.
//
// Safe für nebenläufige Aufrufe.
type IngestSequencer struct {
	db        *sql.DB
	seqName   string
	blockSize int64

	mu     sync.Mutex
	buffer []int64
}

// NewIngestSequencer löst die Sequence hinter
// playback_events.ingest_sequence auf (robust via
// pg_get_serial_sequence statt hartkodiertem Namen) und reserviert
// eager den ersten Block, sodass der Konstruktor bei unerreichbarer DB
// laut scheitert (statt erst beim ersten Next()). blockSize < 1 nimmt
// defaultBlockSize.
func NewIngestSequencer(ctx context.Context, db *sql.DB, blockSize int64) (*IngestSequencer, error) {
	if blockSize < 1 {
		blockSize = defaultBlockSize
	}
	var seqName sql.NullString
	if err := db.QueryRowContext(ctx,
		"SELECT pg_get_serial_sequence('playback_events', 'ingest_sequence')").
		Scan(&seqName); err != nil {
		return nil, fmt.Errorf("postgres: resolve ingest_sequence sequence: %w", err)
	}
	if !seqName.Valid || seqName.String == "" {
		return nil, errors.New("postgres: no sequence backs playback_events.ingest_sequence")
	}
	s := &IngestSequencer{db: db, seqName: seqName.String, blockSize: blockSize}
	buf, err := s.reserveBlock(ctx)
	if err != nil {
		return nil, err
	}
	s.buffer = buf
	return s, nil
}

// Next gibt die nächste ingest_sequence zurück. Bedient aus dem
// vorreservierten Block; ist der Puffer leer, wird ein neuer Block
// reserviert. Kann der Nachfüll-Roundtrip auch nach reserveMaxAttempts
// nicht gelingen, panickt der Sequencer fail-fast: `Next() int64` kann
// keinen Fehler tragen (port-erhaltend, R-28), und ein Sequencer, der
// keinen eindeutigen Wert mehr liefern kann, ist ein harter
// Invariantenbruch — dasselbe Fail-Fast-Muster wie der RNG-Panic in den
// ingest_stream_repository-Adaptern.
func (s *IngestSequencer) Next() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.buffer) == 0 {
		buf, err := s.reserveBlockWithRetry()
		if err != nil {
			panic("postgres: ingest_sequence block reservation failed: " + err.Error())
		}
		s.buffer = buf
	}
	v := s.buffer[0]
	s.buffer = s.buffer[1:]
	return v
}

// reserveBlock reserviert blockSize eindeutige Werte in EINEM Roundtrip.
// `nextval` ist volatile → generate_series(1,N) liefert N distinkte,
// aufsteigende Werte (DB-atomar; konkurrierende Reserves erhalten
// disjunkte Bereiche → keine Dups über Replicas).
func (s *IngestSequencer) reserveBlock(ctx context.Context) ([]int64, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT nextval($1::regclass) FROM generate_series(1, $2)", s.seqName, s.blockSize)
	if err != nil {
		return nil, fmt.Errorf("postgres: reserve ingest_sequence block: %w", err)
	}
	defer func() { _ = rows.Close() }()
	buf := make([]int64, 0, s.blockSize)
	for rows.Next() {
		var v int64
		if err := rows.Scan(&v); err != nil {
			return nil, fmt.Errorf("postgres: scan ingest_sequence: %w", err)
		}
		buf = append(buf, v)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: reserve ingest_sequence block: %w", err)
	}
	if len(buf) == 0 {
		return nil, errors.New("postgres: ingest_sequence block reservation returned no values")
	}
	return buf, nil
}

// reserveBlockWithRetry ruft reserveBlock mit einem eigenen, kurz
// begrenzten Background-Context (Next() hat keinen) und bounded Retries.
func (s *IngestSequencer) reserveBlockWithRetry() ([]int64, error) {
	var lastErr error
	for attempt := 0; attempt < reserveMaxAttempts; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), reserveTimeout)
		buf, err := s.reserveBlock(ctx)
		cancel()
		if err == nil {
			return buf, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

var _ driven.IngestSequencer = (*IngestSequencer)(nil)
