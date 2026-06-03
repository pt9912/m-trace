package driven

import (
	"context"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// SessionRepository hält den aggregierten Sessions-Zustand
// Batch auf; List und Get bedienen die Read-Endpoints Sub-Item 4); Sweep wird vom Lifecycle-Sweeper aufgerufen
// ( Sub-Item 8).
//
// Ab sind alle Methoden, die einzelne Sessions
// adressieren oder filtern, projekt-skopiert: dieselbe session_id in
// zwei Projekten ist als zwei getrennte Sessions zu führen, und ein
// Treffer in Project A darf nicht über einen Lookup in Project B
// erreichbar sein.
//
// Implementierungen müssen für nebenläufige Aufrufe sicher sein.
type SessionRepository interface {
	// UpsertFromEvents legt für jede unbekannte (project_id, session_id)
	// eine neue StreamSession (State=Active) an und aktualisiert für
	// bekannte Sessions LastEventAt und EventCount. Trifft ein Event mit
	// event_name=session_ended ein, wird die Session direkt auf
	// State=Ended gesetzt und EndedAt=event.ServerReceivedAt. Events
	// werden anhand ihres ProjectID/SessionID-Paares zugeordnet.
	//
	// Rückgabe ab C2 (R-6-Fix): die DB-finale
	// `correlation_id` jeder Session, gekeyed nach SessionID. Der Use-
	// Case enricht damit die Events vor `EventRepository.Append`, sodass
	// auch bei einem Race auf einer noch unbekannten (project, session)-
	// Partition niemals ein Event mit einer Verlust-CorrelationID
	// persistiert wird. Ein Batch ist single-project (validiert in der
	// Application-Schicht), darum reicht SessionID als Map-Key.
	UpsertFromEvents(ctx context.Context, events []domain.PlaybackEvent) (map[string]string, error)
	// AppendBoundaries persistiert die im Batch übergebenen
	// `session_boundaries[]`-Einträge in einen durable Session-Metadaten-
	// Store ( D2; spec/telemetry-model.md). Aufruf
	// erfolgt im Use-Case nach erfolgreichem UpsertFromEvents und vor
	// EventRepository.Append. Mehrfach-Sends derselben Tripel
	// `(kind, network_kind, adapter, reason)` für eine Session sind
	// idempotent (Read-Pfad dedupliziert per Tripel; Adapter SHOULD
	// ON CONFLICT DO UPDATE auf `client_timestamp`/`server_received_at`).
	// Eine leere Liste ist no-op.
	AppendBoundaries(ctx context.Context, boundaries []domain.SessionBoundary) error
	// ListBoundariesForSession liefert die persistierten
	// `session_boundaries[]`-Einträge einer (projectID, sessionID)-
	// Partition in Read-Shape-Sortierung (kind asc, adapter asc,
	// reason asc) mit Tripel-Dedup über
	// (kind, network_kind, adapter, reason). Keine Boundaries → leere
	// Slice (`nil` oder `[]`); Cross-Project-Treffer liefern `nil`.
	// Spec-Anker spec/backend-api-contract.md
	ListBoundariesForSession(ctx context.Context, projectID, sessionID string) ([]domain.SessionBoundary, error)
	// ListBoundariesForSessions ist die Bulk-Variante von
	// ListBoundariesForSession (R-7). Liefert
	// pro `sessionIDs[i]` die zugehörige `network_signal_absent[]`-
	// Liste in der gleichen Read-Shape-Sortierung. Result-Map ist
	// gekeyed nach SessionID; SessionIDs ohne Boundaries fehlen
	// in der Map (Aufrufer-Default: leerer Slice). Cross-Project-
	// Treffer werden über den Project-Scope-Filter ausgeschlossen.
	// Leere SessionID-Liste → leere Map (no-op).
	//
	// Adapter müssen eine einzelne Query (`IN`-Clause) nutzen, um
	// N+1-Roundtrips auf der `stream_session_boundaries`-Tabelle
	// zu eliminieren.
	ListBoundariesForSessions(ctx context.Context, projectID string, sessionIDs []string) (map[string][]domain.SessionBoundary, error)
	// List gibt Sessions in stabiler Sortierung (started_at desc,
	// session_id asc) zurück, gefiltert nach q.ProjectID. Der Adapter
	// ist für die Sortierung verantwortlich; der Use Case clampt nur
	// Limit und prüft Cursor-Validität.
	List(ctx context.Context, q SessionListQuery) (SessionPage, error)
	// Get liefert eine einzelne Session über ihr (projectID, sessionID)-
	// Paar. ErrSessionNotFound wenn keine Session in diesem Project
	// existiert; ein Treffer in einem anderen Project gilt als nicht
	// gefunden (kein Cross-Project-Read).
	Get(ctx context.Context, projectID, sessionID string) (domain.StreamSession, error)
	// GetByCorrelationID liefert die Session, deren CorrelationID im
	// gegebenen Project gesetzt ist (Analyzer-Linking).
	// Legacy-Sessions ohne CorrelationID liefern keinen Treffer; der
	// Lookup ist project-skopiert und liefert nie eine Session aus einem
	// fremden Project. ErrSessionNotFound, wenn nichts passt.
	GetByCorrelationID(ctx context.Context, projectID, correlationID string) (domain.StreamSession, error)
	// Sweep wertet die zeitbasierten Lifecycle-Übergänge aus:
	//  Active + (now - LastEventAt > stalledAfter) → Stalled
	//  Stalled + (now - LastEventAt > endedAfter) → Ended (EndedAt=now)
	// Bereits beendete Sessions werden nicht erneut angefasst. Idempotent.
	// Sweep ist global — Lifecycle-Übergänge dürfen Sessions aller
	// Projekte ohne Filter erfassen, weil der Sweeper kein Project-
	// Kontext-Fan-out macht.
	Sweep(ctx context.Context, now time.Time, stalledAfter, endedAfter time.Duration) error
	// CountByState liefert die Anzahl der Sessions im gegebenen
	// Lifecycle-State. Wird vom Active-Sessions-Gauge in Prometheus
	// (informational, kein Pflicht-Counter aus API-Kontrakt)
	// adapter-agnostisch verwendet, sodass das Wiring zwischen In-Memory
	// und SQLite ohne Anpassung wechselt. Wie Sweep ist der Counter
	// global, weil die Prometheus-Exposition kein Project-Label trägt
	// (Cardinality-Regel telemetry-model §3).
	CountByState(ctx context.Context, state domain.SessionState) (int64, error)
	// SetSessionSampleRatePPMIfDefault setzt `sample_rate_ppm` für die
	// (projectID, sessionID)-Session **immutable** auf `ppm`, aber nur
	// wenn der bisherige Wert auf der Default-Marke `SampleRateFull`
	// (1_000_000) steht. Liefert den nach dem Aufruf in der DB
	// persistierten Wert plus ein `applied`-Flag:
	//
	//  - `applied=true`: diese Methode hat den Wert gerade gesetzt.
	//  `existingPPM` ist gleich dem Eingabe-`ppm`.
	//  - `applied=false`: der Wert war bereits != Default. `existingPPM`
	//  ist der schon persistierte Wert; der Aufrufer
	//  vergleicht ihn mit `ppm` und entscheidet, ob
	//  ein Drift-Event vorliegt.
	//
	// `ppm` muss im Bereich `[1, SampleRateFull]` liegen; ein Aufruf mit
	// `ppm == SampleRateFull` ist ein Programmierfehler und ein No-Op
	// (Default-Wert; nicht persistiert).
	SetSessionSampleRatePPMIfDefault(ctx context.Context, projectID, sessionID string, ppm int) (existingPPM int, applied bool, err error)
}

// SessionListQuery ist die Eingabe für SessionRepository.List. ProjectID
// ist Pflicht und filtert die Ergebnisse; ein Leerwert ist ein
// Programmierfehler (Adapter dürfen ErrInvalidProjectScope liefern).
type SessionListQuery struct {
	ProjectID string
	Limit     int
	After     *SessionCursorPosition
}

// SessionCursorPosition ist die Repository-Sicht auf den Cursor —
// die Sortier-Felder ohne Wire-Format.
type SessionCursorPosition struct {
	StartedAt time.Time
	SessionID string
}

// SessionPage bündelt eine Page Sessions plus optional die nächste
// Cursor-Position.
type SessionPage struct {
	Sessions  []domain.StreamSession
	NextAfter *SessionCursorPosition
}
