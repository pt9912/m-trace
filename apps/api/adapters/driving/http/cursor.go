package http

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driving"
)

// errInvalidCursor wird intern verwendet, damit der Handler den
// 400-Pfad mit body {"error":"cursor_invalid"} bedienen kann. Der Use
// Case wirft denselben Domain-Fehler, wenn der Cursor mit fremder
// process_instance_id ankommt — beide Fälle münden im selben Mapping.
var errInvalidCursor = errors.New("cursor invalid")

// wireListSessionsCursor ist die JSON-Form von
// driving.ListSessionsCursor. Felder bewusst kurz, damit der Cursor
// im URL-Query klein bleibt.
type wireListSessionsCursor struct {
	PID string `json:"pid"`
	SA  string `json:"sa"`  // started_at, RFC3339Nano
	SID string `json:"sid"` // session_id
}

// wireSessionEventsCursor ist die JSON-Form von
// driving.SessionEventsCursor.
type wireSessionEventsCursor struct {
	PID string `json:"pid"`
	RCV string `json:"rcv"` // server_received_at, RFC3339Nano
	SEQ *int64 `json:"seq,omitempty"`
	ING int64  `json:"ing"` // ingest_sequence
}

// encodeListSessionsCursor liefert den base64-url-encodierten Cursor
// für die Sessions-Listenpage. Verwendet base64.RawURLEncoding (kein
// Padding, URL-sicher).
func encodeListSessionsCursor(c *driving.ListSessionsCursor) (string, error) {
	if c == nil {
		return "", nil
	}
	wire := wireListSessionsCursor{
		PID: string(c.ProcessInstanceID),
		SA:  c.StartedAt.UTC().Format(time.RFC3339Nano),
		SID: c.SessionID,
	}
	raw, err := json.Marshal(wire)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

// decodeListSessionsCursor parst einen URL-Cursor zurück in den
// typisierten Cursor. Leerer String → nil. Defekter Wert →
// errInvalidCursor; das ist semantisch derselbe Fall wie
// domain.ErrCursorInvalid (siehe Plan 0.1.0 §5.1).
func decodeListSessionsCursor(s string) (*driving.ListSessionsCursor, error) {
	if s == "" {
		return nil, nil
	}
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, errInvalidCursor
	}
	var wire wireListSessionsCursor
	if err := json.Unmarshal(raw, &wire); err != nil {
		return nil, errInvalidCursor
	}
	startedAt, err := time.Parse(time.RFC3339Nano, wire.SA)
	if err != nil {
		return nil, errInvalidCursor
	}
	if wire.PID == "" || wire.SID == "" {
		return nil, errInvalidCursor
	}
	return &driving.ListSessionsCursor{
		ProcessInstanceID: domain.ProcessInstanceID(wire.PID),
		StartedAt:         startedAt,
		SessionID:         wire.SID,
	}, nil
}

// encodeSessionEventsCursor liefert den base64-url-encodierten
// Event-Cursor.
func encodeSessionEventsCursor(c *driving.SessionEventsCursor) (string, error) {
	if c == nil {
		return "", nil
	}
	wire := wireSessionEventsCursor{
		PID: string(c.ProcessInstanceID),
		RCV: c.ServerReceivedAt.UTC().Format(time.RFC3339Nano),
		SEQ: c.SequenceNumber,
		ING: c.IngestSequence,
	}
	raw, err := json.Marshal(wire)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

// decodeSessionEventsCursor ist das Pendant zu encodeSessionEventsCursor.
func decodeSessionEventsCursor(s string) (*driving.SessionEventsCursor, error) {
	if s == "" {
		return nil, nil
	}
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, errInvalidCursor
	}
	var wire wireSessionEventsCursor
	if err := json.Unmarshal(raw, &wire); err != nil {
		return nil, errInvalidCursor
	}
	receivedAt, err := time.Parse(time.RFC3339Nano, wire.RCV)
	if err != nil {
		return nil, errInvalidCursor
	}
	if wire.PID == "" {
		return nil, errInvalidCursor
	}
	return &driving.SessionEventsCursor{
		ProcessInstanceID: domain.ProcessInstanceID(wire.PID),
		ServerReceivedAt:  receivedAt,
		SequenceNumber:    wire.SEQ,
		IngestSequence:    wire.ING,
	}, nil
}
