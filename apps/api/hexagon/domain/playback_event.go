// Package domain holds the framework-free fact-types of the spike.
//
// Per docs/plan-spike.md §5.2/§14.6 nothing in this package may import
// HTTP, JSON, OpenTelemetry, Prometheus, or any other adapter concern.
package domain

import "time"

// PlaybackEvent is a normalized player-side event accepted by the API.
// The wire-format counterpart lives in
// hexagon/port/driving (BatchInput / EventInput).
type PlaybackEvent struct {
	EventName        string
	ProjectID        string
	SessionID        string
	ClientTimestamp  time.Time
	ServerReceivedAt time.Time
	SequenceNumber   *int64
	SDK              SDKInfo
}

// SDKInfo identifies the producing player SDK.
type SDKInfo struct {
	Name    string
	Version string
}
