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
//
// Ab plan-0.4.0 §4.2 nutzt das Repository den projekt-skopierten
// Composite-Key `(project_id, session_id)`; alle Reads filtern in
// WHERE-Clauses nach `project_id`, der `ON CONFLICT`-Pfad zielt auf
// den Composite-PK.
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

	// `ON CONFLICT(project_id, session_id) DO NOTHING` schützt vor einem
	// UNIQUE-Verstoß auf dem Composite-PK, wenn zwei parallele
	// Use-Case-Aufrufe für dieselbe noch unbekannte
	// (project_id, session_id) beide nach Get → ErrSessionNotFound
	// springen und jeweils eine eigene UUIDv4 für `correlation_id`
	// zuweisen. Mit ON CONFLICT bleibt der Sieger durchgehen; der
	// Verlust-Race-Aufruf signalisiert das via `RowsAffected() == 0`,
	// woraufhin upsertSessionFromEventTx die DB-finale `correlation_id`
	// nachliest und an UpsertFromEvents zurückreicht. Damit landen die
	// Events des Verlust-Race-Aufrufs nach dem Use-Case-Enrichment auch
	// mit der Sieger-CorrelationID in `playback_events` — R-6 ist
	// technisch geschlossen (§4.2 C2).
	insertSessionSQL = `
INSERT INTO stream_sessions(
    session_id, project_id, state, started_at, last_seen_at, ended_at,
    event_count, correlation_id, end_source
) VALUES (?, ?, ?, ?, ?, ?, 1, ?, NULL)
ON CONFLICT(project_id, session_id) DO NOTHING`

	// Last-Seen + Event-Count werden auch dann inkrementiert, wenn die
	// Session bereits Ended ist — verspätet eintreffende Events werden
	// gezählt; nur der State-Switch zu Ended ist idempotent.
	updateSessionTickSQL = `
UPDATE stream_sessions
SET last_seen_at = ?, event_count = event_count + 1
WHERE project_id = ? AND session_id = ?`

	markSessionEndedSQL = `
UPDATE stream_sessions
SET state = 'ended', ended_at = ?, end_source = ?
WHERE project_id = ? AND session_id = ? AND state != 'ended'`

	selectSessionByCompositeKeySQL = `
SELECT session_id, project_id, state, started_at, last_seen_at, ended_at,
       event_count, correlation_id, end_source
FROM stream_sessions
WHERE project_id = ? AND session_id = ?`

	selectSessionByCorrelationIDSQL = `
SELECT session_id, project_id, state, started_at, last_seen_at, ended_at,
       event_count, correlation_id, end_source
FROM stream_sessions
WHERE project_id = ? AND correlation_id = ?
LIMIT 1`

	listSessionsBaseSQL = `
SELECT session_id, project_id, state, started_at, last_seen_at, ended_at,
       event_count, correlation_id, end_source
FROM stream_sessions
WHERE project_id = ?`

	listSessionsCursorSQL = `
AND (started_at < ?
     OR (started_at = ? AND session_id > ?))`

	listSessionsOrderSQL = `
ORDER BY started_at DESC, session_id ASC LIMIT ?`

	sweepStalledSQL = `
UPDATE stream_sessions
SET state = 'stalled'
WHERE state = 'active' AND last_seen_at < ?`

	sweepEndedSQL = `
UPDATE stream_sessions
SET state = 'ended', ended_at = ?, end_source = 'sweeper'
WHERE state = 'stalled' AND last_seen_at < ?`

	// Tranche 3 §4.4 D2: Insert-or-Refresh für `session_boundaries[]`.
	// Read-Shape (§3.7.1) dedupliziert per Tripel
	// `(kind, network_kind, adapter, reason)`; ON CONFLICT auf dem PK
	// hält die Persistenz idempotent — Mehrfach-Sends derselben Tripel
	// erzeugen keine zusätzlichen Datensätze und erneuern lediglich
	// die Timestamps.
	upsertSessionBoundarySQL = `
INSERT INTO stream_session_boundaries(
    project_id, session_id, kind, network_kind, adapter, reason,
    client_timestamp, server_received_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(project_id, session_id, kind, network_kind, adapter, reason)
DO UPDATE SET
    client_timestamp   = excluded.client_timestamp,
    server_received_at = excluded.server_received_at`

	// Read-Shape-Sortierung gemäß spec/backend-api-contract.md §3.7.1:
	// stabile Sortierung nach (kind, adapter, reason). Doppelte Tripel
	// gibt es per Composite-PK nicht; eine zusätzliche DISTINCT-Klausel
	// ist nicht nötig.
	listBoundariesForSessionSQL = `
SELECT kind, network_kind, adapter, reason, client_timestamp, server_received_at
FROM stream_session_boundaries
WHERE project_id = ? AND session_id = ?
ORDER BY kind ASC, adapter ASC, reason ASC`
)

// UpsertFromEvents legt unbekannte Sessions an und aktualisiert bekannte
// LastEventAt + EventCount. session_ended-Events schalten den State
// idempotent auf Ended (zweimaliges session_ended verändert ended_at
// nicht); LastEventAt + EventCount werden bei jedem Event aktualisiert,
// auch nachdem die Session beendet wurde — verspätete Events bleiben
// gezählt. Spiegelt das InMemory-Verhalten 1:1.
//
// Rückgabe (R-6-Fix, plan-0.4.0 §4.2 C2): map[sessionID]canonicalCID
// — die DB-finale CorrelationID jeder Session. Bei einem Race auf
// einer noch unbekannten (project_id, session_id) liefert
// `RowsAffected() == 0` aus dem ON-CONFLICT-Insert das Signal, die
// Sieger-CorrelationID nachzulesen und zurückzugeben, sodass der Use-
// Case Events des Verlust-Race-Aufrufs vor dem Append damit enrichen
// kann. Damit ist
// `playback_events.correlation_id == stream_sessions.correlation_id`
// auch unter Concurrency garantiert.
func (r *SessionRepository) UpsertFromEvents(ctx context.Context, events []domain.PlaybackEvent) (map[string]string, error) {
	if len(events) == 0 {
		return map[string]string{}, nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("sqlite: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	canonical := make(map[string]string, len(events))
	for _, e := range events {
		cid, err := upsertSessionFromEventTx(ctx, tx, e)
		if err != nil {
			return nil, err
		}
		canonical[e.SessionID] = cid
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("sqlite: commit: %w", err)
	}
	return canonical, nil
}

// upsertSessionFromEventTx wendet ein einzelnes Event auf den
// Session-State an: Project-Upsert (FK-Vorbedingung), dann je nach
// existierender Session entweder Insert (neu) oder Tick (bekannt).
// Liefert die DB-finale CorrelationID der Session zurück, damit
// UpsertFromEvents sie an den Use-Case zurückreichen kann.
func upsertSessionFromEventTx(ctx context.Context, tx *sql.Tx, e domain.PlaybackEvent) (string, error) {
	if _, err := tx.ExecContext(ctx, upsertProjectSQL, e.ProjectID); err != nil {
		return "", fmt.Errorf("sqlite: upsert project: %w", err)
	}
	existing, err := readSessionTx(ctx, tx, e.ProjectID, e.SessionID)
	switch {
	case errors.Is(err, domain.ErrSessionNotFound):
		return insertNewSessionTx(ctx, tx, e)
	case err != nil:
		return "", err
	default:
		if err := tickExistingSessionTx(ctx, tx, e); err != nil {
			return "", err
		}
		return existing.CorrelationID, nil
	}
}

// tickExistingSessionTx aktualisiert LastEventAt + EventCount einer
// bekannten Session und schaltet bei session_ended-Events zusätzlich
// den State idempotent auf Ended.
func tickExistingSessionTx(ctx context.Context, tx *sql.Tx, e domain.PlaybackEvent) error {
	if _, err := tx.ExecContext(ctx, updateSessionTickSQL,
		formatTime(e.ServerReceivedAt), e.ProjectID, e.SessionID); err != nil {
		return fmt.Errorf("sqlite: tick session: %w", err)
	}
	if e.EventName != persistence.SessionEndedEventName {
		return nil
	}
	if _, err := tx.ExecContext(ctx, markSessionEndedSQL,
		formatTime(e.ServerReceivedAt),
		string(domain.SessionEndSourceClient),
		e.ProjectID, e.SessionID); err != nil {
		return fmt.Errorf("sqlite: mark session ended: %w", err)
	}
	return nil
}

// insertNewSessionTx legt eine bisher unbekannte Session an und liefert
// die DB-finale CorrelationID zurück. Beim allerersten Event genügt ein
// Insert mit State=Active, EventCount=1 und der vom Use-Case
// zugewiesenen CorrelationID; ist das erste Event session_ended, wird
// unmittelbar danach der State-Switch ausgeführt.
//
// R-6-Fix: Bei einem Race greift `ON CONFLICT(project_id, session_id)
// DO NOTHING`. `RowsAffected() == 0` heißt: ein konkurrenter Aufruf
// hat die Session bereits angelegt; wir müssen dessen CorrelationID
// einlesen und zurückgeben, damit der Use-Case unsere Events damit
// enricht. `RowsAffected() == 1` heißt: wir haben die Session
// angelegt; unsere Kandidat-CorrelationID ist die Sieger-CID, und der
// optionale `markSessionEndedSQL` wird auf der von uns angelegten Zeile
// ausgeführt. Im Verlust-Fall überspringen wir `markSessionEndedSQL`,
// damit der Verlust-Aufruf nicht den Endzustand des Siegers
// überschreibt.
func insertNewSessionTx(ctx context.Context, tx *sql.Tx, e domain.PlaybackEvent) (string, error) {
	res, err := tx.ExecContext(ctx, insertSessionSQL,
		e.SessionID,
		e.ProjectID,
		string(domain.SessionStateActive),
		formatTime(e.ServerReceivedAt),
		formatTime(e.ServerReceivedAt),
		nullableTime(nil),
		nullableString(e.CorrelationID),
	)
	if err != nil {
		return "", fmt.Errorf("sqlite: insert session: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return "", fmt.Errorf("sqlite: insert session rows-affected: %w", err)
	}
	if rows == 0 {
		// Race-Verlust: ein konkurrenter Insert hat bereits commited (oder
		// liegt im selben Snapshot vor), unser ON CONFLICT DO NOTHING ist
		// zur No-op geworden. CorrelationID des Siegers nachlesen.
		winner, readErr := readSessionTx(ctx, tx, e.ProjectID, e.SessionID)
		if readErr != nil {
			return "", fmt.Errorf("sqlite: read canonical correlation_id after race: %w", readErr)
		}
		return winner.CorrelationID, nil
	}
	// Wir haben gewonnen — Kandidat-CorrelationID ist DB-final.
	if e.EventName == persistence.SessionEndedEventName {
		if _, err := tx.ExecContext(ctx, markSessionEndedSQL,
			formatTime(e.ServerReceivedAt),
			string(domain.SessionEndSourceClient),
			e.ProjectID, e.SessionID); err != nil {
			return "", fmt.Errorf("sqlite: mark session ended: %w", err)
		}
	}
	return e.CorrelationID, nil
}

// Sweep schaltet zeitbasierte Lifecycle-Übergänge:
//   - active  + (now - last_seen_at) > stalledAfter → stalled
//   - stalled + (now - last_seen_at) > endedAfter   → ended (ended_at=now)
//
// Idempotent (bereits Ended-Sessions werden nicht angefasst). Project-
// agnostisch, weil der Sweeper alle Sessions bedient.
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

// Get liefert eine einzelne Session per (projectID, sessionID).
// ErrSessionNotFound wenn die Session in diesem Project nicht
// existiert; ein Treffer in einem anderen Project gilt als nicht
// gefunden (PK-Lookup ist project-skopiert).
func (r *SessionRepository) Get(ctx context.Context, projectID, sessionID string) (domain.StreamSession, error) {
	row := r.db.QueryRowContext(ctx, selectSessionByCompositeKeySQL, projectID, sessionID)
	return scanSessionRow(row)
}

// GetByCorrelationID liefert die Session mit der gegebenen
// CorrelationID innerhalb des Projects. Leerwerte (Legacy-Sessions
// vor §3.2-Closeout) zählen nicht als Treffer; Cross-Project-Treffer
// liefern ErrSessionNotFound (project-skopierte WHERE-Clause + LIMIT 1).
func (r *SessionRepository) GetByCorrelationID(ctx context.Context, projectID, correlationID string) (domain.StreamSession, error) {
	if correlationID == "" {
		return domain.StreamSession{}, domain.ErrSessionNotFound
	}
	row := r.db.QueryRowContext(ctx, selectSessionByCorrelationIDSQL, projectID, correlationID)
	return scanSessionRow(row)
}

// CountByState zählt Sessions im gegebenen Lifecycle-State über einen
// einfachen `SELECT COUNT(*)` mit Filter; das reicht für den
// Prometheus-Active-Sessions-Gauge (Scrape-on-demand, keine
// Hot-Path-Last) und vermeidet ein In-Memory-Snapshot über alle
// Sessions wie im InMemory-Adapter. Project-agnostisch (Cardinality-
// Regel telemetry-model §3 verbietet `project_id` als Prom-Label).
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
// session_id asc) mit Cursor-Pagination zurück, gefiltert nach
// q.ProjectID.
func (r *SessionRepository) List(ctx context.Context, q driven.SessionListQuery) (driven.SessionPage, error) {
	if q.Limit <= 0 {
		return driven.SessionPage{Sessions: []domain.StreamSession{}}, nil
	}

	args := []any{q.ProjectID}
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
	defer func() { _ = rows.Close() }()

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
func readSessionTx(ctx context.Context, tx *sql.Tx, projectID, sessionID string) (domain.StreamSession, error) {
	row := tx.QueryRowContext(ctx, selectSessionByCompositeKeySQL, projectID, sessionID)
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
		endSource     sql.NullString
	)
	err := row.Scan(&id, &project, &state, &startedAt, &lastSeen, &endedAt,
		&eventCount, &correlationID, &endSource)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.StreamSession{}, domain.ErrSessionNotFound
	}
	if err != nil {
		return domain.StreamSession{}, fmt.Errorf("sqlite: scan session: %w", err)
	}
	return decodeSession(id, project, state, startedAt, lastSeen, endedAt,
		eventCount, correlationID, endSource)
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
		endSource     sql.NullString
	)
	if err := rows.Scan(&id, &project, &state, &startedAt, &lastSeen, &endedAt,
		&eventCount, &correlationID, &endSource); err != nil {
		return domain.StreamSession{}, fmt.Errorf("sqlite: scan session: %w", err)
	}
	return decodeSession(id, project, state, startedAt, lastSeen, endedAt,
		eventCount, correlationID, endSource)
}

func decodeSession(id, project, state, startedAtRaw, lastSeenRaw string,
	endedAtRaw sql.NullString, eventCount int64, correlationID sql.NullString,
	endSource sql.NullString,
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
		EndSource:     domain.SessionEndSource(stringFromNull(endSource)),
	}, nil
}

// AppendBoundaries persistiert `session_boundaries[]`-Einträge in die
// durable `stream_session_boundaries`-Tabelle (plan-0.4.0 §4.4 D2;
// V3-Migration). Mehrfach-Sends derselben Tripel
// `(kind, network_kind, adapter, reason)` für eine Session sind
// idempotent — die Insert-Or-Refresh-Klausel hält die Tripel-Eindeutig-
// keit pro Session und aktualisiert nur `client_timestamp` und
// `server_received_at`. Eine leere Liste ist no-op.
func (r *SessionRepository) AppendBoundaries(ctx context.Context, boundaries []domain.SessionBoundary) error {
	if len(boundaries) == 0 {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("sqlite: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	for _, b := range boundaries {
		if _, err := tx.ExecContext(ctx, upsertSessionBoundarySQL,
			b.ProjectID, b.SessionID, b.Kind, b.NetworkKind, b.Adapter, b.Reason,
			formatTime(b.ClientTimestamp), formatTime(b.ServerReceivedAt),
		); err != nil {
			return fmt.Errorf("sqlite: append boundary: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("sqlite: commit boundaries: %w", err)
	}
	return nil
}

// ListBoundariesForSession liefert die persistierten Boundaries einer
// `(projectID, sessionID)`-Partition in stabiler Read-Shape-Sortierung
// (kind, adapter, reason) — spec/backend-api-contract.md §3.7.1.
// Cross-Project-Treffer sind ausgeschlossen (project-skopiertes WHERE).
// Eine Session ohne Boundaries liefert eine leere Slice (`nil`).
func (r *SessionRepository) ListBoundariesForSession(ctx context.Context, projectID, sessionID string) ([]domain.SessionBoundary, error) {
	rows, err := r.db.QueryContext(ctx, listBoundariesForSessionSQL, projectID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("sqlite: query boundaries: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var out []domain.SessionBoundary
	for rows.Next() {
		b, scanErr := scanBoundary(rows, projectID, sessionID)
		if scanErr != nil {
			return nil, scanErr
		}
		out = append(out, b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sqlite: iterate boundaries: %w", err)
	}
	return out, nil
}

func scanBoundary(rows *sql.Rows, projectID, sessionID string) (domain.SessionBoundary, error) {
	var (
		kind        string
		networkKind string
		adapter     string
		reason      string
		clientTS    string
		serverTS    string
	)
	if err := rows.Scan(&kind, &networkKind, &adapter, &reason, &clientTS, &serverTS); err != nil {
		return domain.SessionBoundary{}, fmt.Errorf("sqlite: scan boundary: %w", err)
	}
	clientTime, err := parseTime(clientTS)
	if err != nil {
		return domain.SessionBoundary{}, err
	}
	serverTime, err := parseTime(serverTS)
	if err != nil {
		return domain.SessionBoundary{}, err
	}
	return domain.SessionBoundary{
		Kind:             kind,
		ProjectID:        projectID,
		SessionID:        sessionID,
		NetworkKind:      networkKind,
		Adapter:          adapter,
		Reason:           reason,
		ClientTimestamp:  clientTime,
		ServerReceivedAt: serverTime,
	}, nil
}

var _ driven.SessionRepository = (*SessionRepository)(nil)
