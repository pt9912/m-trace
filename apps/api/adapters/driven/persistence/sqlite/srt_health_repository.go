package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// SrtHealthRepository ist die durable Variante des
// driven.SrtHealthRepository-Ports gegen die SQLite-Datei aus
// internal/storage. Application- und Domain-Layer bleiben SQLite-frei
// (ADR-0002 §8.2).
//
// Dedupe-Regel (spec/backend-api-contract.md §10.6 / plan-0.6.0 §4
// Sub-3.3): ein Sample ist eindeutig über
//   (project_id, stream_id, connection_id,
//    COALESCE(source_observed_at, source_sequence)).
// Append macht einen Vorab-Lookup auf den Dedupe-Index und
// überspringt vorhandene Einträge — collected_at allein ist kein
// stabiler Schlüssel.
type SrtHealthRepository struct {
	db *sql.DB
}

// NewSrtHealthRepository konstruiert den Adapter.
func NewSrtHealthRepository(db *sql.DB) *SrtHealthRepository {
	return &SrtHealthRepository{db: db}
}

const (
	// upsertProjectForSrtSQL stellt sicher, dass das Project existiert,
	// bevor ein Health-Sample auf die FK srt_health_samples.project_id
	// verweist. Identisch zum upsertProjectSQL aus session_repository,
	// als eigene Konstante belassen, damit der SRT-Adapter unabhängig
	// vom Session-Pfad lebt.
	upsertProjectForSrtSQL = `
INSERT INTO projects(project_id) VALUES (?)
ON CONFLICT(project_id) DO NOTHING`

	// dedupeLookupSQL prüft, ob ein Sample mit gleichem Dedupe-Key
	// schon persistiert ist. Verwendet wird der Dedupe-Index aus der
	// Migration V5 (idx_srt_health_samples_dedupe).
	dedupeLookupSQL = `
SELECT 1 FROM srt_health_samples
WHERE project_id = ?
  AND stream_id = ?
  AND connection_id = ?
  AND COALESCE(source_observed_at, '') = COALESCE(?, '')
  AND COALESCE(source_sequence, '')   = COALESCE(?, '')
LIMIT 1`

	insertSrtHealthSampleSQL = `
INSERT INTO srt_health_samples(
    project_id, stream_id, connection_id,
    source_observed_at, source_sequence,
    collected_at, ingested_at,
    rtt_ms, packet_loss_total, packet_loss_rate,
    retransmissions_total,
    available_bandwidth_bps, throughput_bps, required_bandwidth_bps,
    sample_window_ms,
    source_status, source_error_code, connection_state, health_state
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	// latestByStreamSQL liest den jüngsten Sample pro StreamID. SQLite
	// unterstützt das Pattern "MAX(ingested_at) per stream_id" via
	// Subquery; die Index-Reihenfolge (project_id, stream_id,
	// ingested_at desc) hält den Plan auf Index-Lookup.
	latestByStreamSQL = `
SELECT s.id,
       s.project_id, s.stream_id, s.connection_id,
       s.source_observed_at, s.source_sequence,
       s.collected_at, s.ingested_at,
       s.rtt_ms, s.packet_loss_total, s.packet_loss_rate,
       s.retransmissions_total,
       s.available_bandwidth_bps, s.throughput_bps, s.required_bandwidth_bps,
       s.sample_window_ms,
       s.source_status, s.source_error_code, s.connection_state, s.health_state
FROM srt_health_samples s
INNER JOIN (
    SELECT project_id, stream_id, MAX(ingested_at) AS max_ingested
    FROM srt_health_samples
    WHERE project_id = ?
    GROUP BY project_id, stream_id
) m ON m.project_id = s.project_id AND m.stream_id = s.stream_id AND m.max_ingested = s.ingested_at
WHERE s.project_id = ?
ORDER BY s.stream_id ASC, s.id DESC`

	historyByStreamSQL = `
SELECT id,
       project_id, stream_id, connection_id,
       source_observed_at, source_sequence,
       collected_at, ingested_at,
       rtt_ms, packet_loss_total, packet_loss_rate,
       retransmissions_total,
       available_bandwidth_bps, throughput_bps, required_bandwidth_bps,
       sample_window_ms,
       source_status, source_error_code, connection_state, health_state
FROM srt_health_samples
WHERE project_id = ? AND stream_id = ?
ORDER BY ingested_at DESC, id DESC
LIMIT ?`
)

// Append persistiert eine Liste von Samples mit Dedupe-Skip.
// Concurrent-Writer halten der DSN-Einstellung `_txlock=immediate`
// gegen — alle Operationen einer Append-Charge laufen in einer
// einzelnen Transaktion.
func (r *SrtHealthRepository) Append(ctx context.Context, samples []domain.SrtHealthSample) error {
	if len(samples) == 0 {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("srt-health-sqlite: begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	for i := range samples {
		s := samples[i]
		if _, err := tx.ExecContext(ctx, upsertProjectForSrtSQL, s.ProjectID); err != nil {
			return fmt.Errorf("srt-health-sqlite: upsert project: %w", err)
		}

		var exists int
		row := tx.QueryRowContext(ctx, dedupeLookupSQL,
			s.ProjectID, s.StreamID, s.ConnectionID,
			nullableTime(timePtrOrNil(s.SourceObservedAt)),
			nullableString(s.SourceSequence),
		)
		if err := row.Scan(&exists); err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("srt-health-sqlite: dedupe lookup: %w", err)
			}
			// no rows → frischer Sample, weiter mit Insert.
		} else {
			// Vorhandener Eintrag — Dedupe-Skip.
			continue
		}

		if _, err := tx.ExecContext(ctx, insertSrtHealthSampleSQL,
			s.ProjectID, s.StreamID, s.ConnectionID,
			nullableTime(timePtrOrNil(s.SourceObservedAt)),
			nullableString(s.SourceSequence),
			formatTime(s.CollectedAt), formatTime(s.IngestedAt),
			s.RTTMillis, s.PacketLossTotal, nullableFloat64(s.PacketLossRate),
			s.RetransmissionsTotal,
			s.AvailableBandwidthBPS, nullableInt64(s.ThroughputBPS), nullableInt64(s.RequiredBandwidthBPS),
			nullableInt64(s.SampleWindowMillis),
			string(s.SourceStatus), string(s.SourceErrorCode),
			string(s.ConnectionState), string(s.HealthState),
		); err != nil {
			return fmt.Errorf("srt-health-sqlite: insert sample: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("srt-health-sqlite: commit: %w", err)
	}
	return nil
}

// LatestByStream liefert pro StreamID den jüngsten Sample des
// Projects, sortiert nach StreamID asc / ID desc als Tie-Breaker.
func (r *SrtHealthRepository) LatestByStream(ctx context.Context, projectID string) ([]domain.SrtHealthSample, error) {
	rows, err := r.db.QueryContext(ctx, latestByStreamSQL, projectID, projectID)
	if err != nil {
		return nil, fmt.Errorf("srt-health-sqlite: latest query: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out, err := scanSrtHealthRows(rows)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// HistoryByStream liefert die letzten `limit` Samples einer
// (projectID, streamID)-Kombination. Cursor ist in 0.6.0 Sub-3.3
// noch nicht implementiert (gibt sich bei After != nil mit einem
// Fehler-Stub als Folge-Item für Sub-3.5/Tranche 4).
func (r *SrtHealthRepository) HistoryByStream(ctx context.Context, q driven.SrtHealthHistoryQuery) (driven.SrtHealthHistoryPage, error) {
	if q.After != nil {
		return driven.SrtHealthHistoryPage{}, errors.New("srt-health-sqlite: cursor pagination not yet implemented (plan-0.6.0 Sub-3.5)")
	}
	limit := q.Limit
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	rows, err := r.db.QueryContext(ctx, historyByStreamSQL, q.ProjectID, q.StreamID, limit)
	if err != nil {
		return driven.SrtHealthHistoryPage{}, fmt.Errorf("srt-health-sqlite: history query: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out, err := scanSrtHealthRows(rows)
	if err != nil {
		return driven.SrtHealthHistoryPage{}, err
	}
	// Cursor kommt in Sub-3.5/Tranche 4 — bis dahin keine Folgeseite.
	return driven.SrtHealthHistoryPage{Items: out}, nil
}
