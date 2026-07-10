package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// EventRepository ist die Postgres-Variante des driven.EventRepository-
// Ports (ADR-0006 Dialekt-Spiegel). Query-Strings sind bis auf eine
// Anpassung mit dem SQLite-Adapter identisch (rebind() übersetzt
// `?`→`$n`): `sequence_number` ist im PG-Schema strikt INTEGER (int32),
// SQLite dagegen dynamisch typisiert. Der nullSeqSentinel (int64-min)
// passt nicht in int32, und `COALESCE($a,$b)` aus zwei Bind-Parametern
// hat keinen Typ-Anker (PG inferiert dort `text`). Daher casten die
// Cursor-COALESCEs sowohl die Spalte (`sequence_number::bigint`) als
// auch die Parameter (`?::bigint`) auf bigint — der Keyset-Cursor
// rechnet im bigint-Raum, kanonische Sortierung bleibt gleich.
//
// Event-Dedup folgt §8.3: Events mit gesetzter sequence_number werden
// gegen existierende accepted-Einträge abgeglichen und ggf. als
// duplicate_suspected persistiert.
type EventRepository struct {
	db *sql.DB
}

// NewEventRepository konstruiert den Adapter.
func NewEventRepository(db *sql.DB) *EventRepository {
	return &EventRepository{db: db}
}

const (
	insertPlaybackEventSQL = `
INSERT INTO playback_events(
    ingest_sequence, project_id, session_id, event_name,
    client_timestamp, server_received_at, sequence_number,
    sdk_name, sdk_version, schema_version, meta, delivery_status,
    trace_id, span_id, correlation_id, time_skew_warning
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	dedupLookupSQL = `
SELECT 1 FROM playback_events
WHERE project_id = ? AND session_id = ? AND sequence_number = ?
  AND delivery_status = 'accepted'
LIMIT 1`

	// listEventsSQL trägt ein %s für die optionale Cursor-Klausel;
	// COALESCE bringt NULL-sequence_number auf nullSeqSentinel.
	listEventsSQL = `
SELECT ingest_sequence, project_id, session_id, event_name,
       client_timestamp, server_received_at, sequence_number,
       sdk_name, sdk_version, meta,
       trace_id, span_id, correlation_id, time_skew_warning
FROM playback_events
WHERE project_id = ? AND session_id = ? %s
ORDER BY server_received_at ASC,
         COALESCE(sequence_number::bigint, ?) ASC,
         ingest_sequence ASC
LIMIT ?`

	cursorFilterSQL = `
AND (
    server_received_at > ?
    OR (server_received_at = ? AND COALESCE(sequence_number::bigint, ?) > COALESCE(?::bigint, ?::bigint))
    OR (server_received_at = ?
        AND COALESCE(sequence_number::bigint, ?) = COALESCE(?::bigint, ?::bigint)
        AND ingest_sequence > ?)
)`

	listAfterIngestSequenceSQL = `
SELECT ingest_sequence, project_id, session_id, event_name,
       client_timestamp, server_received_at, sequence_number,
       sdk_name, sdk_version, meta,
       trace_id, span_id, correlation_id, time_skew_warning
FROM playback_events
WHERE project_id = ? AND ingest_sequence > ?
ORDER BY ingest_sequence ASC
LIMIT ?`
)

// Append serialisiert die Events in playback_events. Pro Event greift
// die Dedup-Klassifikation aus §8.3; die Charge läuft in einer
// Transaktion (READ COMMITTED).
func (r *EventRepository) Append(ctx context.Context, events []domain.PlaybackEvent) error {
	if len(events) == 0 {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("event-postgres: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	insertQ := rebind(insertPlaybackEventSQL)
	dedupQ := rebind(dedupLookupSQL)

	for _, e := range events {
		status, err := classifyDelivery(ctx, tx, dedupQ, e)
		if err != nil {
			return err
		}
		metaJSON, err := encodeMeta(e.Meta)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, insertQ,
			e.IngestSequence,
			e.ProjectID,
			e.SessionID,
			e.EventName,
			formatTime(e.ClientTimestamp),
			formatTime(e.ServerReceivedAt),
			nullableInt64(e.SequenceNumber),
			e.SDK.Name,
			e.SDK.Version,
			persistedSchemaVersion,
			metaJSON,
			status,
			nullableString(e.TraceID),
			nullableString(e.SpanID),
			nullableString(e.CorrelationID),
			boolToInt(e.TimeSkewWarning),
		); err != nil {
			return fmt.Errorf("event-postgres: insert event: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("event-postgres: commit: %w", err)
	}
	return nil
}

// classifyDelivery entscheidet pro Event, ob es als accepted oder
// duplicate_suspected persistiert wird. Events ohne sequence_number
// sind immer accepted. dedupQ ist die bereits rebound Query.
func classifyDelivery(ctx context.Context, tx *sql.Tx, dedupQ string, e domain.PlaybackEvent) (string, error) {
	if e.SequenceNumber == nil {
		return "accepted", nil
	}
	var hit int
	err := tx.QueryRowContext(ctx, dedupQ,
		e.ProjectID, e.SessionID, *e.SequenceNumber).Scan(&hit)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return "accepted", nil
	case err != nil:
		return "", fmt.Errorf("event-postgres: dedup lookup: %w", err)
	default:
		return "duplicate_suspected", nil
	}
}

// ListBySession liefert Events einer Session in kanonischer Sortierung
// mit Limit-/Cursor-basierter Pagination (ADR-0002, API-Kontrakt).
func (r *EventRepository) ListBySession(ctx context.Context, q driven.EventListQuery) (driven.EventPage, error) {
	if q.Limit <= 0 {
		return driven.EventPage{Events: []domain.PlaybackEvent{}}, nil
	}

	args := []any{q.ProjectID, q.SessionID}
	cursorClause := ""
	if q.After != nil {
		cursorClause = cursorFilterSQL
		args = append(args,
			formatTime(q.After.ServerReceivedAt),
			formatTime(q.After.ServerReceivedAt),
			nullSeqSentinel, nullableInt64(q.After.SequenceNumber), nullSeqSentinel,
			formatTime(q.After.ServerReceivedAt),
			nullSeqSentinel, nullableInt64(q.After.SequenceNumber), nullSeqSentinel,
			q.After.IngestSequence,
		)
	}
	args = append(args, nullSeqSentinel)
	args = append(args, q.Limit+1)

	query := rebind(fmt.Sprintf(listEventsSQL, cursorClause))
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return driven.EventPage{}, fmt.Errorf("event-postgres: query events: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]domain.PlaybackEvent, 0, q.Limit)
	for rows.Next() {
		e, err := scanEventRow(rows)
		if err != nil {
			return driven.EventPage{}, err
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return driven.EventPage{}, fmt.Errorf("event-postgres: scan events: %w", err)
	}

	page := driven.EventPage{Events: out}
	if len(out) > q.Limit {
		page.Events = out[:q.Limit]
		last := page.Events[q.Limit-1]
		page.NextAfter = &driven.EventCursorPosition{
			ServerReceivedAt: last.ServerReceivedAt,
			SequenceNumber:   last.SequenceNumber,
			IngestSequence:   last.IngestSequence,
		}
	}
	return page, nil
}

func scanEventRow(rows *sql.Rows) (domain.PlaybackEvent, error) {
	var (
		ingest        int64
		project       string
		session       string
		eventName     string
		clientTS      string
		serverTS      string
		seqNumber     sql.NullInt64
		sdkName       string
		sdkVer        string
		metaRaw       sql.NullString
		traceID       sql.NullString
		spanID        sql.NullString
		correlationID sql.NullString
		skewWarning   int64
	)
	if err := rows.Scan(&ingest, &project, &session, &eventName,
		&clientTS, &serverTS, &seqNumber, &sdkName, &sdkVer, &metaRaw,
		&traceID, &spanID, &correlationID, &skewWarning); err != nil {
		return domain.PlaybackEvent{}, fmt.Errorf("event-postgres: scan: %w", err)
	}
	clientAt, err := parseTime(clientTS)
	if err != nil {
		return domain.PlaybackEvent{}, err
	}
	serverAt, err := parseTime(serverTS)
	if err != nil {
		return domain.PlaybackEvent{}, err
	}
	meta, err := decodeMeta(metaRaw)
	if err != nil {
		return domain.PlaybackEvent{}, err
	}
	var seqPtr *int64
	if seqNumber.Valid {
		v := seqNumber.Int64
		seqPtr = &v
	}
	return domain.PlaybackEvent{
		EventName:        eventName,
		ProjectID:        project,
		SessionID:        session,
		ClientTimestamp:  clientAt,
		ServerReceivedAt: serverAt,
		IngestSequence:   ingest,
		SequenceNumber:   seqPtr,
		SDK:              domain.SDKInfo{Name: sdkName, Version: sdkVer},
		Meta:             meta,
		TraceID:          stringFromNull(traceID),
		SpanID:           stringFromNull(spanID),
		CorrelationID:    stringFromNull(correlationID),
		TimeSkewWarning:  skewWarning != 0,
	}, nil
}

// ListAfterIngestSequence implementiert den SSE-Backfill-Hook. Reine
// Read-Operation ohne Tx.
func (r *EventRepository) ListAfterIngestSequence(ctx context.Context, projectID string, afterSeq int64, limit int) ([]domain.PlaybackEvent, error) {
	if limit <= 0 {
		return nil, nil
	}
	rows, err := r.db.QueryContext(ctx, rebind(listAfterIngestSequenceSQL), projectID, afterSeq, limit)
	if err != nil {
		return nil, fmt.Errorf("event-postgres: query backfill: %w", err)
	}
	defer func() { _ = rows.Close() }()
	out := make([]domain.PlaybackEvent, 0, limit)
	for rows.Next() {
		ev, err := scanEventRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, ev)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("event-postgres: iterate backfill: %w", err)
	}
	return out, nil
}

var _ driven.EventRepository = (*EventRepository)(nil)
