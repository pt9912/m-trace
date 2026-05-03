// Package sqlite liefert die durable Variante der Driven-Persistence-
// Ports (Sessions, Events, Ingest-Sequencer) gegen die SQLite-Datei
// aus internal/storage. Application- und Domain-Layer bleiben SQLite-
// frei und sprechen ausschließlich gegen die Hexagon-Ports (ADR-0002).
package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// EventRepository ist die durable Variante des
// driven.EventRepository-Ports gegen die SQLite-Datei aus
// internal/storage. Application- und Domain-Layer bleiben SQLite-frei
// (ADR-0002 §8.2).
//
// Race-Schutz beim Append: jede Append-Operation läuft in einer
// einzigen Transaktion. Die DSN aus internal/storage erzwingt
// `BEGIN IMMEDIATE` für jede `db.BeginTx`-Tx (ADR-0002 §8.3),
// womit zwei konkurrente Append-Calls per DB-Lock serialisiert
// werden — Dedup-Klassifikation bleibt deterministisch.
//
// Event-Dedup folgt §8.3: Events mit gesetzter `sequence_number`
// werden anhand `(project_id, session_id, sequence_number)` mit
// existierenden `delivery_status='accepted'`-Einträgen abgeglichen;
// Treffer landen mit `delivery_status='duplicate_suspected'`. Events
// ohne `sequence_number` sind immer `accepted`.
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
    trace_id, span_id, correlation_id
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	dedupLookupSQL = `
SELECT 1 FROM playback_events
WHERE project_id = ? AND session_id = ? AND sequence_number = ?
  AND delivery_status = 'accepted'
LIMIT 1`

	// listEventsSQL liefert Events einer Session in kanonischer
	// Sortierung. COALESCE bringt NULL-sequence_number auf den
	// Sentinel-Wert nullSeqSentinel und macht den Cursor-Filter sowie
	// die ORDER BY-Klausel ohne CASE-Switching ausdrückbar.
	listEventsSQL = `
SELECT ingest_sequence, project_id, session_id, event_name,
       client_timestamp, server_received_at, sequence_number,
       sdk_name, sdk_version, meta,
       trace_id, span_id, correlation_id
FROM playback_events
WHERE session_id = ? %s
ORDER BY server_received_at ASC,
         COALESCE(sequence_number, ?) ASC,
         ingest_sequence ASC
LIMIT ?`

	cursorFilterSQL = `
AND (
    server_received_at > ?
    OR (server_received_at = ? AND COALESCE(sequence_number, ?) > COALESCE(?, ?))
    OR (server_received_at = ?
        AND COALESCE(sequence_number, ?) = COALESCE(?, ?)
        AND ingest_sequence > ?)
)`
)

// Append serialisiert die Events in die playback_events-Tabelle.
// Pro Event wird die Dedup-Klassifikation aus §8.3 angewandt;
// Insert-Reihenfolge folgt der Slice-Reihenfolge.
func (r *EventRepository) Append(ctx context.Context, events []domain.PlaybackEvent) error {
	if len(events) == 0 {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("sqlite: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	for _, e := range events {
		status, err := classifyDelivery(ctx, tx, e)
		if err != nil {
			return err
		}
		metaJSON, err := encodeMeta(e.Meta)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, insertPlaybackEventSQL,
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
		); err != nil {
			return fmt.Errorf("sqlite: insert event: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("sqlite: commit: %w", err)
	}
	return nil
}

// classifyDelivery entscheidet pro Event, ob es als `accepted` oder
// `duplicate_suspected` persistiert wird. Events ohne sequence_number
// sind immer `accepted` — kein automatischer Dedup ohne expliziten
// Schlüssel (ADR-0002 §8.3).
func classifyDelivery(ctx context.Context, tx *sql.Tx, e domain.PlaybackEvent) (string, error) {
	if e.SequenceNumber == nil {
		return "accepted", nil
	}
	var hit int
	err := tx.QueryRowContext(ctx, dedupLookupSQL,
		e.ProjectID, e.SessionID, *e.SequenceNumber).Scan(&hit)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return "accepted", nil
	case err != nil:
		return "", fmt.Errorf("sqlite: dedup lookup: %w", err)
	default:
		return "duplicate_suspected", nil
	}
}

// ListBySession liefert Events einer Session in kanonischer
// Sortierung mit Limit-/Cursor-basierter Pagination
// (ADR-0002 §8.1, API-Kontrakt §10.4).
func (r *EventRepository) ListBySession(ctx context.Context, q driven.EventListQuery) (driven.EventPage, error) {
	if q.Limit <= 0 {
		return driven.EventPage{Events: []domain.PlaybackEvent{}}, nil
	}

	args := []any{q.SessionID}
	cursorClause := ""
	if q.After != nil {
		cursorClause = cursorFilterSQL
		args = append(args,
			formatTime(q.After.ServerReceivedAt),                // server_received_at >
			formatTime(q.After.ServerReceivedAt),                // server_received_at =
			nullSeqSentinel, nullableInt64(q.After.SequenceNumber), nullSeqSentinel, // COALESCE seq > COALESCE seq
			formatTime(q.After.ServerReceivedAt),                // server_received_at =
			nullSeqSentinel, nullableInt64(q.After.SequenceNumber), nullSeqSentinel, // COALESCE seq = COALESCE seq
			q.After.IngestSequence,                              // ingest_sequence >
		)
	}
	args = append(args, nullSeqSentinel)         // COALESCE in ORDER BY
	args = append(args, q.Limit+1)               // fetch limit+1 to detect NextAfter

	query := fmt.Sprintf(listEventsSQL, cursorClause)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return driven.EventPage{}, fmt.Errorf("sqlite: query events: %w", err)
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
		return driven.EventPage{}, fmt.Errorf("sqlite: scan events: %w", err)
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
	)
	if err := rows.Scan(&ingest, &project, &session, &eventName,
		&clientTS, &serverTS, &seqNumber, &sdkName, &sdkVer, &metaRaw,
		&traceID, &spanID, &correlationID); err != nil {
		return domain.PlaybackEvent{}, fmt.Errorf("sqlite: scan: %w", err)
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
	}, nil
}

var _ driven.EventRepository = (*EventRepository)(nil)
