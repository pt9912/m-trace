package domain

import "time"

// SessionState ist der grobe Lifecycle einer Player-Session
// (plan-0.1.0.md §5.1).
//
//   - Active:  letztes Event innerhalb des Stalled-Schwellwerts.
//   - Stalled: keine Events innerhalb des Schwellwerts (z. B. 60 s),
//     aber noch kein Ended.
//   - Ended:   explizites End-Event aus dem SDK oder Inaktivität jenseits
//     des Stalled-Fensters.
//
// Stalled/Ended-Übergänge übernimmt der Lifecycle-Sweeper aus
// plan-0.1.0.md §5.1 Sub-Item 8 (in 0.1.0 noch ⬜); §5.1 Sub-Item 3
// liefert nur den Zustand „Active" plus die Felder, gegen die der
// Sweeper später entscheidet.
type SessionState string

// Session-Lifecycle-Zustände aus plan-0.1.0.md §5.1 Sub-Item 8.
// `Active` ist der Eintrittszustand beim ersten Event; `Stalled` und
// `Ended` werden vom Sweeper gesetzt (siehe SessionsSweeper).
const (
	SessionStateActive  SessionState = "active"
	SessionStateStalled SessionState = "stalled"
	SessionStateEnded   SessionState = "ended"
)

// StreamSession aggregiert Events mit gleicher session_id (plan-0.1.0.md
// §5.1). Felder werden beim ersten Event auf Default-State Active
// gesetzt; LastEventAt und EventCount tracken folgende Events derselben
// Session und sind die Grundlage für Lifecycle-Übergänge (Sub-Item 8).
//
// EndedAt wird nur gesetzt, wenn State==Ended; bis dahin nil.
type StreamSession struct {
	ID          string
	ProjectID   string
	State       SessionState
	StartedAt   time.Time
	LastEventAt time.Time
	EndedAt     *time.Time
	EventCount  int64
	// CorrelationID ist die Server-generierte, durable Source-of-Truth
	// für die Tempo-unabhängige Dashboard-Korrelation der Session. Wird
	// beim allerersten Event der Session erzeugt (UUIDv4) und über alle
	// Folge-Events konstant gehalten. Source spec/telemetry-model.md §2.5.
	CorrelationID string
}
