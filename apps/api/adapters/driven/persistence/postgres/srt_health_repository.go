package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// historyClampLimit ist das obere Limit pro Page (spec §7a.3).
const historyClampLimit = 1000

// SrtHealthRepository ist die Postgres-Variante des
// driven.SrtHealthRepository-Ports (ADR-0006, Dialekt-Spiegel des
// SQLite-Adapters). Query-Konstanten sind mit dem SQLite-Adapter
// identisch (dialekt-neutrale `?`-Platzhalter), rebind() übersetzt sie
// zur Laufzeit auf `$n`; die Spalten-Typen sind dank reversiertem
// Schema gleich (datetime-als-TEXT, RFC3339Nano).
//
// Dedupe-Regel (spec/backend-api-contract.md): ein Sample ist eindeutig
// über (project_id, stream_id, connection_id,
// COALESCE(source_observed_at, source_sequence)). Append macht einen
// Vorab-Lookup und überspringt vorhandene Einträge.
type SrtHealthRepository struct {
	db *sql.DB
}

// NewSrtHealthRepository konstruiert den Adapter.
func NewSrtHealthRepository(db *sql.DB) *SrtHealthRepository {
	return &SrtHealthRepository{db: db}
}

const (
	upsertProjectForSrtSQL = `
INSERT INTO projects(project_id) VALUES (?)
ON CONFLICT(project_id) DO NOTHING`

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

	historyByStreamAfterSQL = `
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
  AND (ingested_at < ? OR (ingested_at = ? AND id < ?))
ORDER BY ingested_at DESC, id DESC
LIMIT ?`
)

// Append persistiert eine Liste von Samples mit Dedupe-Skip. Alle
// Operationen einer Charge laufen in einer einzelnen Transaktion
// (READ COMMITTED); Concurrent-Writer serialisiert Postgres über den
// Row-Lock der jeweiligen Insert-Transaktion.
func (r *SrtHealthRepository) Append(ctx context.Context, samples []domain.SrtHealthSample) error {
	if len(samples) == 0 {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("srt-health-postgres: begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	upsertQ := rebind(upsertProjectForSrtSQL)
	dedupeQ := rebind(dedupeLookupSQL)
	insertQ := rebind(insertSrtHealthSampleSQL)

	for i := range samples {
		s := samples[i]
		if _, err := tx.ExecContext(ctx, upsertQ, s.ProjectID); err != nil {
			return fmt.Errorf("srt-health-postgres: upsert project: %w", err)
		}

		var exists int
		row := tx.QueryRowContext(ctx, dedupeQ,
			s.ProjectID, s.StreamID, s.ConnectionID,
			nullableTime(timePtrOrNil(s.SourceObservedAt)),
			nullableString(s.SourceSequence),
		)
		if err := row.Scan(&exists); err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("srt-health-postgres: dedupe lookup: %w", err)
			}
			// no rows → frischer Sample, weiter mit Insert.
		} else {
			// Vorhandener Eintrag — Dedupe-Skip.
			continue
		}

		if _, err := tx.ExecContext(ctx, insertQ,
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
			return fmt.Errorf("srt-health-postgres: insert sample: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("srt-health-postgres: commit: %w", err)
	}
	return nil
}

// LatestByStream liefert pro StreamID den jüngsten Sample des Projects,
// sortiert nach StreamID asc / ID desc als Tie-Breaker.
func (r *SrtHealthRepository) LatestByStream(ctx context.Context, projectID string) ([]domain.SrtHealthSample, error) {
	rows, err := r.db.QueryContext(ctx, rebind(latestByStreamSQL), projectID, projectID)
	if err != nil {
		return nil, fmt.Errorf("srt-health-postgres: latest query: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out, _, err := scanSrtHealthRows(rows)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// HistoryByStream liefert bis zu `Limit` Samples einer
// (projectID, streamID)-Kombination, sortiert nach (ingested_at desc,
// id desc). Bei q.After != nil setzt die WHERE-Klausel die
// Storage-Position fort (Keyset-Pagination). Der Adapter fetched
// limit+1 Rows; bei genau limit+1 existiert eine Folgeseite und
// NextAfter zeigt auf die letzte zurückgegebene Row.
func (r *SrtHealthRepository) HistoryByStream(ctx context.Context, q driven.SrtHealthHistoryQuery) (driven.SrtHealthHistoryPage, error) {
	limit := q.Limit
	if limit <= 0 || limit > historyClampLimit {
		limit = 100
	}
	probe := limit + 1

	var (
		rows *sql.Rows
		err  error
	)
	if q.After == nil {
		rows, err = r.db.QueryContext(ctx, rebind(historyByStreamSQL), q.ProjectID, q.StreamID, probe)
	} else {
		afterTS := formatTime(q.After.IngestedAt)
		rows, err = r.db.QueryContext(ctx, rebind(historyByStreamAfterSQL),
			q.ProjectID, q.StreamID,
			afterTS, afterTS, q.After.ID,
			probe,
		)
	}
	if err != nil {
		return driven.SrtHealthHistoryPage{}, fmt.Errorf("srt-health-postgres: history query: %w", err)
	}
	defer func() { _ = rows.Close() }()

	samples, ids, err := scanSrtHealthRows(rows)
	if err != nil {
		return driven.SrtHealthHistoryPage{}, err
	}

	page := driven.SrtHealthHistoryPage{Items: samples}
	if len(samples) > limit {
		page.Items = samples[:limit]
		page.NextAfter = &driven.SrtHealthCursor{
			IngestedAt: page.Items[limit-1].IngestedAt,
			ID:         ids[limit-1],
		}
	}
	return page, nil
}

var _ driven.SrtHealthRepository = (*SrtHealthRepository)(nil)
