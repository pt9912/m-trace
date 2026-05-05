package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// srtHealthRow trägt die Raw-Spalten einer srt_health_samples-Row in
// genau der Reihenfolge, in der scanSrtHealthRows sie liest.
type srtHealthRow struct {
	id                    int64
	projectID             string
	streamID              string
	connectionID          string
	sourceObservedAt      sql.NullString
	sourceSequence        sql.NullString
	collectedAt           string
	ingestedAt            string
	rttMS                 float64
	packetLossTotal       int64
	packetLossRate        sql.NullFloat64
	retransmissionsTotal  int64
	availableBandwidthBPS int64
	throughputBPS         sql.NullInt64
	requiredBandwidthBPS  sql.NullInt64
	sampleWindowMS        sql.NullInt64
	sourceStatus          string
	sourceErrorCode       string
	connectionState       string
	healthState           string
}

// scanSrtHealthRows liest eine Folge von srt_health_samples-Rows in
// Domain-Samples, in der vom SQL bestimmten Reihenfolge. Die Spalten-
// ordnung muss zu latestByStreamSQL / historyByStreamSQL passen —
// beim Ändern eines Statements gehört hier die gleiche Reihenfolge
// nachgezogen.
func scanSrtHealthRows(rows *sql.Rows) ([]domain.SrtHealthSample, error) {
	var out []domain.SrtHealthSample
	for rows.Next() {
		var r srtHealthRow
		if err := rows.Scan(
			&r.id,
			&r.projectID, &r.streamID, &r.connectionID,
			&r.sourceObservedAt, &r.sourceSequence,
			&r.collectedAt, &r.ingestedAt,
			&r.rttMS, &r.packetLossTotal, &r.packetLossRate,
			&r.retransmissionsTotal,
			&r.availableBandwidthBPS, &r.throughputBPS, &r.requiredBandwidthBPS,
			&r.sampleWindowMS,
			&r.sourceStatus, &r.sourceErrorCode, &r.connectionState, &r.healthState,
		); err != nil {
			return nil, fmt.Errorf("srt-health-sqlite: scan row: %w", err)
		}
		sample, err := r.toDomain()
		if err != nil {
			return nil, err
		}
		out = append(out, sample)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("srt-health-sqlite: rows iteration: %w", err)
	}
	return out, nil
}

// toDomain konvertiert die Raw-Row in das Domain-Sample, inklusive
// Time-Parsing und Nullable-Mapping. Getrennte Methode, damit
// scanSrtHealthRows kurz bleibt (gocognit/funlen).
func (r srtHealthRow) toDomain() (domain.SrtHealthSample, error) {
	var observedAt time.Time
	if r.sourceObservedAt.Valid && r.sourceObservedAt.String != "" {
		parsed, err := parseTime(r.sourceObservedAt.String)
		if err != nil {
			return domain.SrtHealthSample{}, err
		}
		observedAt = parsed
	}

	collected, err := parseTime(r.collectedAt)
	if err != nil {
		return domain.SrtHealthSample{}, err
	}
	ingested, err := parseTime(r.ingestedAt)
	if err != nil {
		return domain.SrtHealthSample{}, err
	}

	return domain.SrtHealthSample{
		ProjectID:    r.projectID,
		StreamID:     r.streamID,
		ConnectionID: r.connectionID,

		SourceObservedAt: observedAt,
		SourceSequence:   stringFromNull(r.sourceSequence),

		CollectedAt: collected,
		IngestedAt:  ingested,

		RTTMillis:             r.rttMS,
		PacketLossTotal:       r.packetLossTotal,
		PacketLossRate:        nullFloat64ToPtr(r.packetLossRate),
		RetransmissionsTotal:  r.retransmissionsTotal,
		AvailableBandwidthBPS: r.availableBandwidthBPS,
		ThroughputBPS:         nullInt64ToPtr(r.throughputBPS),
		RequiredBandwidthBPS:  nullInt64ToPtr(r.requiredBandwidthBPS),
		SampleWindowMillis:    nullInt64ToPtr(r.sampleWindowMS),

		SourceStatus:    domain.SourceStatus(r.sourceStatus),
		SourceErrorCode: domain.SourceErrorCode(r.sourceErrorCode),
		ConnectionState: domain.ConnectionState(r.connectionState),
		HealthState:     domain.HealthState(r.healthState),
	}, nil
}

func nullFloat64ToPtr(n sql.NullFloat64) *float64 {
	if !n.Valid {
		return nil
	}
	v := n.Float64
	return &v
}

func nullInt64ToPtr(n sql.NullInt64) *int64 {
	if !n.Valid {
		return nil
	}
	v := n.Int64
	return &v
}
