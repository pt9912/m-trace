package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/persistence"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// SessionRepository ist die durable Variante des
// driven.SessionRepository-Ports gegen die SQLite-Datei aus
// internal/storage. Application- und Domain-Layer bleiben SQLite-frei
// (ADR-0002 §8.2).
//
// Idempotenz aus §8.3:
//   - UpsertFromEvents legt unbekannte Sessions an, aktualisiert
//     bekannte LastEventAt/EventCount, und schaltet auf Ended bei
//     `event_name == "session_ended"`. Ein zweiter Upsert mit demselben
//     session_ended-Event lässt EndedAt unverändert.
//   - Sweep ist idempotent: bereits Ended-Sessions bleiben unangetastet.
//
// Einzelne Operationen laufen in einer Transaktion (BEGIN IMMEDIATE
// via DSN), damit Concurrent-Reader/Writer eindeutige Snapshots sehen.
// FK auf projects(project_id) wird vom Aufruf-Pfad geprüft —
// UpsertFromEvents stellt sicher, dass das Project vor der Session-
// Insert existiert.
type SessionRepository struct {
	db *sql.DB
}

// NewSessionRepository konstruiert den Adapter.
func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

const (
	upsertProjectSQL = `
INSERT INTO projects(project_id) VALUES (?)
ON CONFLICT(project_id) DO NOTHING`

	insertSessionSQL = `
INSERT INTO stream_sessions(
    session_id, project_id, state, started_at, last_seen_at, ended_at,
    event_count, correlation_id
) VALUES (?, ?, ?, ?, ?, ?, 1, ?)`

	// Last-Seen + Event-Count werden auch dann inkrementiert, wenn die
	// Session bereits Ended ist — verspätet eintreffende Events werden
	// gezählt; nur der State-Switch zu Ended ist idempotent.
	updateSessionTickSQL = `
UPDATE stream_sessions
SET last_seen_at = ?, event_count = event_count + 1
WHERE session_id = ?`

	markSessionEndedSQL = `
UPDATE stream_sessions
SET state = 'ended', ended_at = ?
WHERE session_id = ? AND state != 'ended'`

	selectSessionByIDSQL = `
SELECT session_id, project_id, state, started_at, last_seen_at, ended_at,
       event_count, correlation_id
FROM stream_sessions
WHERE session_id = ?`

	listSessionsBaseSQL = `
SELECT session_id, project_id, state, started_at, last_seen_at, ended_at,
       event_count, correlation_id
FROM stream_sessions`

	listSessionsCursorSQL = `
WHERE started_at < ?
   OR (started_at = ? AND session_id > ?)`

	listSessionsOrderSQL = `
ORDER BY started_at DESC, session_id ASC LIMIT ?`

	sweepStalledSQL = `
UPDATE stream_sessions
SET state = 'stalled'
WHERE state = 'active' AND last_seen_at < ?`

	sweepEndedSQL = `
UPDATE stream_sessions
SET state = 'ended', ended_at = ?
WHERE state = 'stalled' AND last_seen_at < ?`
)

// UpsertFromEvents legt unbekannte Sessions an und aktualisiert bekannte
// LastEventAt + EventCount. session_ended-Events schalten den State
// idempotent auf Ended (zweimaliges session_ended verändert ended_at
// nicht); LastEventAt + EventCount werden bei jedem Event aktualisiert,
// auch nachdem die Session beendet wurde — verspätete Events bleiben
// gezählt. Spiegelt das InMemory-Verhalten 1:1.
func (r *SessionRepository) UpsertFromEvents(ctx context.Context, events []domain.PlaybackEvent) error {
	if len(events) == 0 {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("sqlite: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	for _, e := range events {
		if _, err := tx.ExecContext(ctx, upsertProjectSQL, e.ProjectID); err != nil {
			return fmt.Errorf("sqlite: upsert project: %w", err)
		}
		_, err := readSessionTx(ctx, tx, e.SessionID)
		switch {
		case errors.Is(err, domain.ErrSessionNotFound):
			if err := insertNewSessionTx(ctx, tx, e); err != nil {
				return err
			}
		case err != nil:
			return err
		default:
			if _, err := tx.ExecContext(ctx, updateSessionTickSQL,
				formatTime(e.ServerReceivedAt), e.SessionID); err != nil {
				return fmt.Errorf("sqlite: tick session: %w", err)
			}
			if e.EventName == persistence.SessionEndedEventName {
				if _, err := tx.ExecContext(ctx, markSessionEndedSQL,
					formatTime(e.ServerReceivedAt), e.SessionID); err != nil {
					return fmt.Errorf("sqlite: mark session ended: %w", err)
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("sqlite: commit: %w", err)
	}
	return nil
}

// insertNewSessionTx legt eine bisher unbekannte Session an. Beim
// allerersten Event genügt ein Insert mit State=Active, EventCount=1
// und der vom Use-Case zugewiesenen CorrelationID; ist das erste
// Event session_ended, wird unmittelbar danach der State-Switch
// ausgeführt.
func insertNewSessionTx(ctx context.Context, tx *sql.Tx, e domain.PlaybackEvent) error {
	if _, err := tx.ExecContext(ctx, insertSessionSQL,
		e.SessionID,
		e.ProjectID,
		string(domain.SessionStateActive),
		formatTime(e.ServerReceivedAt),
		formatTime(e.ServerReceivedAt),
		nullableTime(nil),
		nullableString(e.CorrelationID),
	); err != nil {
		return fmt.Errorf("sqlite: insert session: %w", err)
	}
	if e.EventName == persistence.SessionEndedEventName {
		if _, err := tx.ExecContext(ctx, markSessionEndedSQL,
			formatTime(e.ServerReceivedAt), e.SessionID); err != nil {
			return fmt.Errorf("sqlite: mark session ended: %w", err)
		}
	}
	return nil
}

// Sweep schaltet zeitbasierte Lifecycle-Übergänge:
//   - active  + (now - last_seen_at) > stalledAfter → stalled
//   - stalled + (now - last_seen_at) > endedAfter   → ended (ended_at=now)
//
// Idempotent (bereits Ended-Sessions werden nicht angefasst).
func (r *SessionRepository) Sweep(ctx context.Context, now time.Time, stalledAfter, endedAfter time.Duration) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("sqlite: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stalledThreshold := now.Add(-stalledAfter)
	endedThreshold := now.Add(-endedAfter)

	if _, err := tx.ExecContext(ctx, sweepStalledSQL,
		formatTime(stalledThreshold)); err != nil {
		return fmt.Errorf("sqlite: sweep stalled: %w", err)
	}
	if _, err := tx.ExecContext(ctx, sweepEndedSQL,
		formatTime(now), formatTime(endedThreshold)); err != nil {
		return fmt.Errorf("sqlite: sweep ended: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("sqlite: commit: %w", err)
	}
	return nil
}

// Get liefert eine einzelne Session per ID. ErrSessionNotFound wenn
// keine Session existiert.
func (r *SessionRepository) Get(ctx context.Context, id string) (domain.StreamSession, error) {
	row := r.db.QueryRowContext(ctx, selectSessionByIDSQL, id)
	return scanSessionRow(row)
}

// CountByState zählt Sessions im gegebenen Lifecycle-State über einen
// einfachen `SELECT COUNT(*)` mit Filter; das reicht für den
// Prometheus-Active-Sessions-Gauge (Scrape-on-demand, keine
// Hot-Path-Last) und vermeidet ein In-Memory-Snapshot über alle
// Sessions wie im InMemory-Adapter.
func (r *SessionRepository) CountByState(ctx context.Context, state domain.SessionState) (int64, error) {
	var n int64
	if err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM stream_sessions WHERE state = ?",
		string(state)).Scan(&n); err != nil {
		return 0, fmt.Errorf("sqlite: count sessions by state: %w", err)
	}
	return n, nil
}

// List gibt Sessions in stabiler Sortierung (started_at desc,
// session_id asc) mit Cursor-Pagination zurück.
func (r *SessionRepository) List(ctx context.Context, q driven.SessionListQuery) (driven.SessionPage, error) {
	if q.Limit <= 0 {
		return driven.SessionPage{Sessions: []domain.StreamSession{}}, nil
	}

	args := []any{}
	cursorClause := ""
	if q.After != nil {
		cursorClause = listSessionsCursorSQL
		args = append(args,
			formatTime(q.After.StartedAt), // started_at <
			formatTime(q.After.StartedAt), // started_at =
			q.After.SessionID,             // session_id >
		)
	}
	args = append(args, q.Limit+1)

	rows, err := r.db.QueryContext(ctx,
		listSessionsBaseSQL+" "+cursorClause+" "+listSessionsOrderSQL,
		args...)
	if err != nil {
		return driven.SessionPage{}, fmt.Errorf("sqlite: query sessions: %w", err)
	}
	defer rows.Close()

	out := make([]domain.StreamSession, 0, q.Limit)
	for rows.Next() {
		s, err := scanSessionFromRows(rows)
		if err != nil {
			return driven.SessionPage{}, err
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return driven.SessionPage{}, fmt.Errorf("sqlite: scan sessions: %w", err)
	}

	page := driven.SessionPage{Sessions: out}
	if len(out) > q.Limit {
		page.Sessions = out[:q.Limit]
		last := page.Sessions[q.Limit-1]
		page.NextAfter = &driven.SessionCursorPosition{
			StartedAt: last.StartedAt,
			SessionID: last.ID,
		}
	}
	return page, nil
}

// readSessionTx liest eine Session über den aktuellen Tx-Handle.
// Nutzt Get-äquivalentes SQL, aber muss innerhalb der UpsertFromEvents-
// Tx laufen, damit die Lesung den eigenen In-Flight-Insert sieht.
func readSessionTx(ctx context.Context, tx *sql.Tx, id string) (domain.StreamSession, error) {
	row := tx.QueryRowContext(ctx, selectSessionByIDSQL, id)
	return scanSessionRow(row)
}

func scanSessionRow(row *sql.Row) (domain.StreamSession, error) {
	var (
		id            string
		project       string
		state         string
		startedAt     string
		lastSeen      string
		endedAt       sql.NullString
		eventCount    int64
		correlationID sql.NullString
	)
	err := row.Scan(&id, &project, &state, &startedAt, &lastSeen, &endedAt,
		&eventCount, &correlationID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.StreamSession{}, domain.ErrSessionNotFound
	}
	if err != nil {
		return domain.StreamSession{}, fmt.Errorf("sqlite: scan session: %w", err)
	}
	return decodeSession(id, project, state, startedAt, lastSeen, endedAt,
		eventCount, correlationID)
}

func scanSessionFromRows(rows *sql.Rows) (domain.StreamSession, error) {
	var (
		id            string
		project       string
		state         string
		startedAt     string
		lastSeen      string
		endedAt       sql.NullString
		eventCount    int64
		correlationID sql.NullString
	)
	if err := rows.Scan(&id, &project, &state, &startedAt, &lastSeen, &endedAt,
		&eventCount, &correlationID); err != nil {
		return domain.StreamSession{}, fmt.Errorf("sqlite: scan session: %w", err)
	}
	return decodeSession(id, project, state, startedAt, lastSeen, endedAt,
		eventCount, correlationID)
}

func decodeSession(id, project, state, startedAtRaw, lastSeenRaw string,
	endedAtRaw sql.NullString, eventCount int64, correlationID sql.NullString,
) (domain.StreamSession, error) {
	startedAt, err := parseTime(startedAtRaw)
	if err != nil {
		return domain.StreamSession{}, err
	}
	lastSeen, err := parseTime(lastSeenRaw)
	if err != nil {
		return domain.StreamSession{}, err
	}
	var endedAt *time.Time
	if endedAtRaw.Valid {
		t, err := parseTime(endedAtRaw.String)
		if err != nil {
			return domain.StreamSession{}, err
		}
		endedAt = &t
	}
	return domain.StreamSession{
		ID:            id,
		ProjectID:     project,
		State:         domain.SessionState(state),
		StartedAt:     startedAt,
		LastEventAt:   lastSeen,
		EndedAt:       endedAt,
		EventCount:    eventCount,
		CorrelationID: stringFromNull(correlationID),
	}, nil
}

var _ driven.SessionRepository = (*SessionRepository)(nil)
