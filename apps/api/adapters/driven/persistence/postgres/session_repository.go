package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/persistence"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// SessionRepository ist die Postgres-Variante des
// driven.SessionRepository-Ports (ADR-0006, Dialekt-Spiegel des
// SQLite-Adapters). Query-Konstanten sind mit dem SQLite-Adapter
// identisch (dialekt-neutrale `?`-Platzhalter); rebind() übersetzt sie
// zur Laufzeit auf `$n`. Die Spalten-Typen sind dank reversiertem Schema
// gleich (datetime-als-TEXT/RFC3339Nano, event_count/sample_rate_ppm als
// INTEGER).
//
// Idempotenz aus §8.3 (identisch zum SQLite-Adapter):
//   - UpsertFromEvents legt unbekannte Sessions an, aktualisiert bekannte
//     LastEventAt/EventCount, und schaltet auf Ended bei
//     `event_name == "session_ended"`. Ein zweiter Upsert mit demselben
//     session_ended-Event lässt EndedAt unverändert.
//   - Sweep ist idempotent: bereits Ended-Sessions bleiben unangetastet.
//
// Dialekt-Unterschied nur beim Platzhalter (rebind). Der R-6-Race wird —
// wie im SQLite-Adapter — rein über `ON CONFLICT DO NOTHING` +
// `RowsAffected` + Readback der Sieger-CorrelationID geschlossen (kein
// SQLSTATE-Mapping): unter Postgres' READ-COMMITTED-Default wartet die
// spekulative Insertion des Verlust-Aufrufs auf den Commit des Siegers,
// danach liest der Readback die kanonische correlation_id.
type SessionRepository struct {
	db *sql.DB
}

// NewSessionRepository konstruiert den Adapter.
func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

var _ driven.SessionRepository = (*SessionRepository)(nil)

const (
	upsertProjectForSessionSQL = `
INSERT INTO projects(project_id) VALUES (?)
ON CONFLICT(project_id) DO NOTHING`

	// `ON CONFLICT(project_id, session_id) DO NOTHING` schützt vor einem
	// UNIQUE-Verstoß auf dem Composite-PK, wenn zwei parallele
	// Use-Case-Aufrufe für dieselbe noch unbekannte
	// (project_id, session_id) beide nach Get → ErrSessionNotFound springen
	// und je eine eigene UUIDv4 für `correlation_id` zuweisen. Der Sieger
	// geht durch; der Verlust-Race-Aufruf signalisiert das via
	// `RowsAffected == 0`, woraufhin insertNewSessionTx die DB-finale
	// `correlation_id` nachliest und an UpsertFromEvents zurückreicht —
	// R-6 ist damit technisch geschlossen (§4.2 C2).
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
       event_count, correlation_id, end_source, sample_rate_ppm
FROM stream_sessions
WHERE project_id = ? AND session_id = ?`

	selectSessionByCorrelationIDSQL = `
SELECT session_id, project_id, state, started_at, last_seen_at, ended_at,
       event_count, correlation_id, end_source, sample_rate_ppm
FROM stream_sessions
WHERE project_id = ? AND correlation_id = ?
LIMIT 1`

	listSessionsBaseSQL = `
SELECT session_id, project_id, state, started_at, last_seen_at, ended_at,
       event_count, correlation_id, end_source, sample_rate_ppm
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

	// Insert-or-Refresh für `session_boundaries[]`. Read-Shape (§3.7.1)
	// dedupliziert per Tripel `(kind, network_kind, adapter, reason)`;
	// ON CONFLICT auf dem Composite-PK hält die Persistenz idempotent —
	// Mehrfach-Sends derselben Tripel erneuern nur die Timestamps.
	upsertSessionBoundarySQL = `
INSERT INTO stream_session_boundaries(
    project_id, session_id, kind, network_kind, adapter, reason,
    client_timestamp, server_received_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(project_id, session_id, kind, network_kind, adapter, reason)
DO UPDATE SET
    client_timestamp   = excluded.client_timestamp,
    server_received_at = excluded.server_received_at`

	// Read-Shape-Sortierung gemäß spec/backend-api-contract.md: stabile
	// Sortierung nach (kind, adapter, reason). Doppelte Tripel gibt es per
	// Composite-PK nicht; kein zusätzliches DISTINCT nötig.
	listBoundariesForSessionSQL = `
SELECT kind, network_kind, adapter, reason, client_timestamp, server_received_at
FROM stream_session_boundaries
WHERE project_id = ? AND session_id = ?
ORDER BY kind ASC, adapter ASC, reason ASC`

	setSampleRateIfDefaultSQL = `
UPDATE stream_sessions
SET sample_rate_ppm = ?
WHERE project_id = ? AND session_id = ? AND sample_rate_ppm = ?`

	selectSampleRateSQL = `
SELECT sample_rate_ppm FROM stream_sessions
WHERE project_id = ? AND session_id = ?`
)

// UpsertFromEvents legt unbekannte Sessions an und aktualisiert bekannte
// LastEventAt + EventCount. session_ended-Events schalten den State
// idempotent auf Ended; LastEventAt + EventCount werden bei jedem Event
// aktualisiert, auch nach dem Ende. Spiegelt das InMemory-/SQLite-
// Verhalten 1:1.
//
// Rückgabe (R-6-Fix, C2): map[sessionID]canonicalCID — die DB-finale
// CorrelationID jeder Session, damit der Use-Case die Events des
// Verlust-Race-Aufrufs vor `EventRepository.Append` mit der Sieger-CID
// enrichen kann.
func (r *SessionRepository) UpsertFromEvents(ctx context.Context, events []domain.PlaybackEvent) (map[string]string, error) {
	if len(events) == 0 {
		return map[string]string{}, nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("session-postgres: begin tx: %w", err)
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
		return nil, fmt.Errorf("session-postgres: commit: %w", err)
	}
	return canonical, nil
}

// upsertSessionFromEventTx wendet ein einzelnes Event auf den
// Session-State an: Project-Upsert (FK-Vorbedingung), dann je nach
// existierender Session entweder Insert (neu) oder Tick (bekannt).
// Liefert die DB-finale CorrelationID der Session zurück.
func upsertSessionFromEventTx(ctx context.Context, tx *sql.Tx, e domain.PlaybackEvent) (string, error) {
	if _, err := tx.ExecContext(ctx, rebind(upsertProjectForSessionSQL), e.ProjectID); err != nil {
		return "", fmt.Errorf("session-postgres: upsert project: %w", err)
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
// bekannten Session und schaltet bei session_ended-Events zusätzlich den
// State idempotent auf Ended.
func tickExistingSessionTx(ctx context.Context, tx *sql.Tx, e domain.PlaybackEvent) error {
	if _, err := tx.ExecContext(ctx, rebind(updateSessionTickSQL),
		formatTime(e.ServerReceivedAt), e.ProjectID, e.SessionID); err != nil {
		return fmt.Errorf("session-postgres: tick session: %w", err)
	}
	if e.EventName != persistence.SessionEndedEventName {
		return nil
	}
	if _, err := tx.ExecContext(ctx, rebind(markSessionEndedSQL),
		formatTime(e.ServerReceivedAt),
		string(domain.SessionEndSourceClient),
		e.ProjectID, e.SessionID); err != nil {
		return fmt.Errorf("session-postgres: mark session ended: %w", err)
	}
	return nil
}

// insertNewSessionTx legt eine bisher unbekannte Session an und liefert
// die DB-finale CorrelationID zurück.
//
// R-6-Fix: Bei einem Race greift `ON CONFLICT(project_id, session_id)
// DO NOTHING`. `RowsAffected == 0` heißt: ein konkurrenter Aufruf hat die
// Session bereits angelegt; wir lesen dessen CorrelationID nach und geben
// sie zurück (und überspringen `markSessionEndedSQL`, damit der
// Verlust-Aufruf den Endzustand des Siegers nicht überschreibt).
// `RowsAffected == 1` heißt: unsere Kandidat-CID ist die Sieger-CID.
func insertNewSessionTx(ctx context.Context, tx *sql.Tx, e domain.PlaybackEvent) (string, error) {
	res, err := tx.ExecContext(ctx, rebind(insertSessionSQL),
		e.SessionID,
		e.ProjectID,
		string(domain.SessionStateActive),
		formatTime(e.ServerReceivedAt),
		formatTime(e.ServerReceivedAt),
		nullableTime(nil),
		nullableString(e.CorrelationID),
	)
	if err != nil {
		return "", fmt.Errorf("session-postgres: insert session: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return "", fmt.Errorf("session-postgres: insert session rows-affected: %w", err)
	}
	if rows == 0 {
		// Race-Verlust: ein konkurrenter Insert hat bereits committed,
		// unser ON CONFLICT DO NOTHING ist zur No-op geworden. CID des
		// Siegers nachlesen.
		winner, readErr := readSessionTx(ctx, tx, e.ProjectID, e.SessionID)
		if readErr != nil {
			return "", fmt.Errorf("session-postgres: read canonical correlation_id after race: %w", readErr)
		}
		return winner.CorrelationID, nil
	}
	// Wir haben gewonnen — Kandidat-CID ist DB-final.
	if e.EventName == persistence.SessionEndedEventName {
		if _, err := tx.ExecContext(ctx, rebind(markSessionEndedSQL),
			formatTime(e.ServerReceivedAt),
			string(domain.SessionEndSourceClient),
			e.ProjectID, e.SessionID); err != nil {
			return "", fmt.Errorf("session-postgres: mark session ended: %w", err)
		}
	}
	return e.CorrelationID, nil
}

// Sweep schaltet zeitbasierte Lifecycle-Übergänge (active→stalled→ended).
// Idempotent (bereits Ended-Sessions bleiben unangetastet), project-
// agnostisch.
func (r *SessionRepository) Sweep(ctx context.Context, now time.Time, stalledAfter, endedAfter time.Duration) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("session-postgres: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stalledThreshold := now.Add(-stalledAfter)
	endedThreshold := now.Add(-endedAfter)

	if _, err := tx.ExecContext(ctx, rebind(sweepStalledSQL),
		formatTime(stalledThreshold)); err != nil {
		return fmt.Errorf("session-postgres: sweep stalled: %w", err)
	}
	if _, err := tx.ExecContext(ctx, rebind(sweepEndedSQL),
		formatTime(now), formatTime(endedThreshold)); err != nil {
		return fmt.Errorf("session-postgres: sweep ended: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("session-postgres: commit: %w", err)
	}
	return nil
}

// Get liefert eine einzelne Session per (projectID, sessionID).
// ErrSessionNotFound bei Nichtexistenz; ein Treffer in einem anderen
// Project gilt als nicht gefunden (project-skopierter PK-Lookup).
func (r *SessionRepository) Get(ctx context.Context, projectID, sessionID string) (domain.StreamSession, error) {
	row := r.db.QueryRowContext(ctx, rebind(selectSessionByCompositeKeySQL), projectID, sessionID)
	s, err := scanSession(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.StreamSession{}, domain.ErrSessionNotFound
	}
	if err != nil {
		return domain.StreamSession{}, err
	}
	return s, nil
}

// GetByCorrelationID liefert die Session mit der gegebenen CorrelationID
// innerhalb des Projects. Leerwerte zählen nicht als Treffer;
// Cross-Project-Treffer liefern ErrSessionNotFound.
func (r *SessionRepository) GetByCorrelationID(ctx context.Context, projectID, correlationID string) (domain.StreamSession, error) {
	if correlationID == "" {
		return domain.StreamSession{}, domain.ErrSessionNotFound
	}
	row := r.db.QueryRowContext(ctx, rebind(selectSessionByCorrelationIDSQL), projectID, correlationID)
	s, err := scanSession(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.StreamSession{}, domain.ErrSessionNotFound
	}
	if err != nil {
		return domain.StreamSession{}, err
	}
	return s, nil
}

// CountByState zählt Sessions im gegebenen Lifecycle-State. Project-
// agnostisch (Cardinality-Regel telemetry-model §3).
func (r *SessionRepository) CountByState(ctx context.Context, state domain.SessionState) (int64, error) {
	var n int64
	if err := r.db.QueryRowContext(ctx,
		rebind("SELECT COUNT(*) FROM stream_sessions WHERE state = ?"),
		string(state)).Scan(&n); err != nil {
		return 0, fmt.Errorf("session-postgres: count sessions by state: %w", err)
	}
	return n, nil
}

// SetSessionSampleRatePPMIfDefault implementiert den Immutability-Set für
// die Pro-Session-Sampling-Rate (R-10): UPDATE nur, wenn sample_rate_ppm
// aktuell auf SampleRateFull (Default) steht. RowsAffected==1 →
// applied=true; ==0 → bereits gesetzt, existingPPM via Folge-SELECT.
func (r *SessionRepository) SetSessionSampleRatePPMIfDefault(ctx context.Context, projectID, sessionID string, ppm int) (int, bool, error) {
	if ppm == domain.SampleRateFull {
		// No-Op: Default-Wert. Spec sagt: bei `sampleRate == 1` weglassen.
		return domain.SampleRateFull, false, nil
	}
	res, err := r.db.ExecContext(ctx, rebind(setSampleRateIfDefaultSQL),
		int64(ppm), projectID, sessionID, int64(domain.SampleRateFull))
	if err != nil {
		return 0, false, fmt.Errorf("session-postgres: set sample_rate_ppm: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return 0, false, fmt.Errorf("session-postgres: set sample_rate_ppm rows-affected: %w", err)
	}
	if rows == 1 {
		return ppm, true, nil
	}
	// rows == 0 → bereits != Default. Bestehenden Wert lesen für Drift-
	// Vergleich.
	var existing int64
	if err := r.db.QueryRowContext(ctx, rebind(selectSampleRateSQL), projectID, sessionID).Scan(&existing); err != nil {
		return 0, false, fmt.Errorf("session-postgres: read existing sample_rate_ppm: %w", err)
	}
	return int(existing), false, nil
}

// List gibt Sessions in stabiler Sortierung (started_at desc, session_id
// asc) mit Cursor-Pagination zurück, gefiltert nach q.ProjectID. Der
// dynamisch zusammengesetzte Query (Base + optionaler Cursor + Order)
// wird als Ganzes rebindet, damit die `$n`-Nummerierung über alle
// Platzhalter korrekt läuft.
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
		rebind(listSessionsBaseSQL+" "+cursorClause+" "+listSessionsOrderSQL),
		args...)
	if err != nil {
		return driven.SessionPage{}, fmt.Errorf("session-postgres: query sessions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]domain.StreamSession, 0, q.Limit)
	for rows.Next() {
		s, err := scanSession(rows)
		if err != nil {
			return driven.SessionPage{}, err
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return driven.SessionPage{}, fmt.Errorf("session-postgres: scan sessions: %w", err)
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

// readSessionTx liest eine Session über den aktuellen Tx-Handle, damit die
// Lesung den eigenen In-Flight-Insert der UpsertFromEvents-Tx sieht.
func readSessionTx(ctx context.Context, tx *sql.Tx, projectID, sessionID string) (domain.StreamSession, error) {
	row := tx.QueryRowContext(ctx, rebind(selectSessionByCompositeKeySQL), projectID, sessionID)
	s, err := scanSession(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.StreamSession{}, domain.ErrSessionNotFound
	}
	if err != nil {
		return domain.StreamSession{}, err
	}
	return s, nil
}

// scanSession liest eine Row (`*sql.Row` oder `*sql.Rows`) in eine
// domain.StreamSession. Rohfehler (inkl. sql.ErrNoRows) werden
// durchgereicht; die Aufrufer mappen ErrNoRows auf ErrSessionNotFound.
func scanSession(rs rowScanner) (domain.StreamSession, error) {
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
		sampleRatePPM int64
	)
	if err := rs.Scan(&id, &project, &state, &startedAt, &lastSeen, &endedAt,
		&eventCount, &correlationID, &endSource, &sampleRatePPM); err != nil {
		return domain.StreamSession{}, err
	}
	startedAtT, err := parseTime(startedAt)
	if err != nil {
		return domain.StreamSession{}, err
	}
	lastSeenT, err := parseTime(lastSeen)
	if err != nil {
		return domain.StreamSession{}, err
	}
	var endedAtP *time.Time
	if endedAt.Valid {
		t, err := parseTime(endedAt.String)
		if err != nil {
			return domain.StreamSession{}, err
		}
		endedAtP = &t
	}
	return domain.StreamSession{
		ID:            id,
		ProjectID:     project,
		State:         domain.SessionState(state),
		StartedAt:     startedAtT,
		LastEventAt:   lastSeenT,
		EndedAt:       endedAtP,
		EventCount:    eventCount,
		CorrelationID: stringFromNull(correlationID),
		EndSource:     domain.SessionEndSource(stringFromNull(endSource)),
		SampleRatePPM: int(sampleRatePPM),
	}, nil
}

// AppendBoundaries persistiert `session_boundaries[]`-Einträge idempotent
// (Insert-or-Refresh je Tripel). Eine leere Liste ist no-op.
func (r *SessionRepository) AppendBoundaries(ctx context.Context, boundaries []domain.SessionBoundary) error {
	if len(boundaries) == 0 {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("session-postgres: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	boundaryQ := rebind(upsertSessionBoundarySQL)
	for _, b := range boundaries {
		if _, err := tx.ExecContext(ctx, boundaryQ,
			b.ProjectID, b.SessionID, b.Kind, b.NetworkKind, b.Adapter, b.Reason,
			formatTime(b.ClientTimestamp), formatTime(b.ServerReceivedAt),
		); err != nil {
			return fmt.Errorf("session-postgres: append boundary: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("session-postgres: commit boundaries: %w", err)
	}
	return nil
}

// ListBoundariesForSession liefert die persistierten Boundaries einer
// (projectID, sessionID)-Partition in stabiler Read-Shape-Sortierung
// (kind, adapter, reason). Cross-Project-Treffer sind ausgeschlossen; eine
// Session ohne Boundaries liefert eine leere Slice (`nil`).
func (r *SessionRepository) ListBoundariesForSession(ctx context.Context, projectID, sessionID string) ([]domain.SessionBoundary, error) {
	rows, err := r.db.QueryContext(ctx, rebind(listBoundariesForSessionSQL), projectID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("session-postgres: query boundaries: %w", err)
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
		return nil, fmt.Errorf("session-postgres: iterate boundaries: %w", err)
	}
	return out, nil
}

// ListBoundariesForSessions ist die Bulk-Variante: eine einzige Query mit
// `IN (?, ?, ?)`-Clause ersetzt N+1-Roundtrips (R-7). Der dynamisch
// zusammengesetzte Query wird als Ganzes rebindet. Result-Map ist gekeyt
// nach SessionID; SessionIDs ohne Boundaries fehlen in der Map.
func (r *SessionRepository) ListBoundariesForSessions(ctx context.Context, projectID string, sessionIDs []string) (map[string][]domain.SessionBoundary, error) {
	if len(sessionIDs) == 0 {
		return map[string][]domain.SessionBoundary{}, nil
	}
	placeholders := make([]string, len(sessionIDs))
	args := make([]any, 0, len(sessionIDs)+1)
	args = append(args, projectID)
	for i, id := range sessionIDs {
		placeholders[i] = "?"
		args = append(args, id)
	}
	query := rebind(fmt.Sprintf(`
SELECT session_id, kind, network_kind, adapter, reason, client_timestamp, server_received_at
FROM stream_session_boundaries
WHERE project_id = ? AND session_id IN (%s)
ORDER BY session_id ASC, kind ASC, adapter ASC, reason ASC`,
		strings.Join(placeholders, ", ")))
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("session-postgres: bulk query boundaries: %w", err)
	}
	defer func() { _ = rows.Close() }()
	out := make(map[string][]domain.SessionBoundary, len(sessionIDs))
	for rows.Next() {
		var (
			sessionID   string
			kind        string
			networkKind string
			adapter     string
			reason      string
			clientTS    string
			serverTS    string
		)
		if err := rows.Scan(&sessionID, &kind, &networkKind, &adapter, &reason, &clientTS, &serverTS); err != nil {
			return nil, fmt.Errorf("session-postgres: scan bulk boundary: %w", err)
		}
		clientTime, err := parseTime(clientTS)
		if err != nil {
			return nil, err
		}
		serverTime, err := parseTime(serverTS)
		if err != nil {
			return nil, err
		}
		out[sessionID] = append(out[sessionID], domain.SessionBoundary{
			Kind:             kind,
			ProjectID:        projectID,
			SessionID:        sessionID,
			NetworkKind:      networkKind,
			Adapter:          adapter,
			Reason:           reason,
			ClientTimestamp:  clientTime,
			ServerReceivedAt: serverTime,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("session-postgres: iterate bulk boundaries: %w", err)
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
		return domain.SessionBoundary{}, fmt.Errorf("session-postgres: scan boundary: %w", err)
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
