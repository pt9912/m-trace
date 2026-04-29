package http

import (
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driving"
)

// TestEncodeDecodeListSessionsCursor_RoundTrip verifiziert, dass
// Encoding und Decoding eines ListSessionsCursor sich gegenseitig
// neutralisieren — keine Verluste in PID, StartedAt (auf nano-Genauigkeit
// im UTC-Frame) und SessionID.
func TestEncodeDecodeListSessionsCursor_RoundTrip(t *testing.T) {
	t.Parallel()
	original := &driving.ListSessionsCursor{
		ProcessInstanceID: domain.ProcessInstanceID("abc123"),
		StartedAt:         time.Date(2026, 4, 28, 12, 34, 56, 789_012_345, time.UTC),
		SessionID:         "sess-xyz",
	}
	encoded, err := encodeListSessionsCursor(original)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	decoded, err := decodeListSessionsCursor(encoded)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if decoded.ProcessInstanceID != original.ProcessInstanceID {
		t.Errorf("PID round-trip: got %q want %q", decoded.ProcessInstanceID, original.ProcessInstanceID)
	}
	if !decoded.StartedAt.Equal(original.StartedAt) {
		t.Errorf("StartedAt round-trip: got %v want %v", decoded.StartedAt, original.StartedAt)
	}
	if decoded.SessionID != original.SessionID {
		t.Errorf("SessionID round-trip: got %q want %q", decoded.SessionID, original.SessionID)
	}
}

// TestDecodeListSessionsCursor_EmptyAndMalformed deckt die zwei Pfade,
// die der HTTP-Handler verzweigt: leerer Cursor (= keine Pagination)
// vs. defekter Cursor (= 400 cursor_invalid).
func TestDecodeListSessionsCursor_EmptyAndMalformed(t *testing.T) {
	t.Parallel()
	got, err := decodeListSessionsCursor("")
	if err != nil {
		t.Errorf("empty: expected nil error, got %v", err)
	}
	if got != nil {
		t.Errorf("empty: expected nil cursor, got %v", got)
	}

	if _, err := decodeListSessionsCursor("not-base64!"); !errors.Is(err, errInvalidCursor) {
		t.Errorf("malformed base64: want errInvalidCursor, got %v", err)
	}
	// Valid base64, but not JSON.
	if _, err := decodeListSessionsCursor("AAEC"); !errors.Is(err, errInvalidCursor) {
		t.Errorf("not JSON: want errInvalidCursor, got %v", err)
	}
}

// TestDecodeListSessionsCursor_EmptyFieldsRejected verifiziert, dass
// ein wohlgeformter base64-JSON-Cursor mit leeren PID- oder SID-Feldern
// als errInvalidCursor abgelehnt wird (defense-in-depth gegen
// gefälschte Cursor).
func TestDecodeListSessionsCursor_EmptyFieldsRejected(t *testing.T) {
	t.Parallel()
	cases := []string{
		`{"pid":"","sa":"2026-04-28T12:00:00Z","sid":"s1"}`,
		`{"pid":"abc","sa":"2026-04-28T12:00:00Z","sid":""}`,
		`{"pid":"abc","sa":"not-a-time","sid":"s1"}`,
	}
	for _, raw := range cases {
		encoded := encodeRaw(raw)
		if _, err := decodeListSessionsCursor(encoded); !errors.Is(err, errInvalidCursor) {
			t.Errorf("decode(%q): expected errInvalidCursor, got %v", raw, err)
		}
	}
}

// TestDecodeSessionEventsCursor_EmptyFieldsRejected — analog für den
// Event-Cursor.
func TestDecodeSessionEventsCursor_EmptyFieldsRejected(t *testing.T) {
	t.Parallel()
	cases := []string{
		`{"pid":"","rcv":"2026-04-28T12:00:00Z","ing":1}`,
		`{"pid":"abc","rcv":"not-a-time","ing":1}`,
	}
	for _, raw := range cases {
		encoded := encodeRaw(raw)
		if _, err := decodeSessionEventsCursor(encoded); !errors.Is(err, errInvalidCursor) {
			t.Errorf("decode(%q): expected errInvalidCursor, got %v", raw, err)
		}
	}
}

// encodeRaw ist ein Helper für die obige Test-Suite — base64-url ohne
// Padding über das stdlib base64-Paket.
func encodeRaw(raw string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

// TestEncodeDecodeSessionEventsCursor_RoundTrip — analog für die
// Event-Pagination, inkl. optional gesetzter SequenceNumber.
func TestEncodeDecodeSessionEventsCursor_RoundTrip(t *testing.T) {
	t.Parallel()
	seq := int64(42)
	original := &driving.SessionEventsCursor{
		ProcessInstanceID: domain.ProcessInstanceID("abc"),
		ServerReceivedAt:  time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC),
		SequenceNumber:    &seq,
		IngestSequence:    99,
	}
	encoded, err := encodeSessionEventsCursor(original)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	decoded, err := decodeSessionEventsCursor(encoded)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if decoded.IngestSequence != 99 {
		t.Errorf("IngestSequence round-trip: got %d want 99", decoded.IngestSequence)
	}
	if decoded.SequenceNumber == nil || *decoded.SequenceNumber != 42 {
		t.Errorf("SequenceNumber round-trip failed, got %v", decoded.SequenceNumber)
	}
	if !decoded.ServerReceivedAt.Equal(original.ServerReceivedAt) {
		t.Errorf("ServerReceivedAt round-trip failed")
	}
}
