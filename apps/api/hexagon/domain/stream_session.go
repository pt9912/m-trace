package domain

import "time"

// SessionState is the coarse lifecycle of a player session.
// The spike auto-creates sessions on first event and keeps them Active.
// Explicit Ended-transition is bonus scope per Spec §7.
type SessionState string

const (
	SessionStateActive SessionState = "active"
	SessionStateEnded  SessionState = "ended"
)

// StreamSession aggregates events sharing the same session_id.
type StreamSession struct {
	ID        string
	ProjectID string
	State     SessionState
	StartedAt time.Time
	EndedAt   *time.Time
}
