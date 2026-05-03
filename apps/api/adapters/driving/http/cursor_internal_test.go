package http

import (
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/port/driving"
)

// TestEncodeDecodeListSessionsCursor_RoundTrip verifiziert, dass
// Encoding und Decoding eines ListSessionsCursor sich gegenseitig
// neutralisieren — keine Verluste in StartedAt (auf nano-Genauigkeit
// im UTC-Frame) und SessionID. Cursor-v2 enthält **kein** PID mehr
// (siehe ADR-0004 §5).
func TestEncodeDecodeListSessionsCursor_RoundTrip(t *testing.T) {
	t.Parallel()
	original := &driving.ListSessionsCursor{
		StartedAt: time.Date(2026, 4, 28, 12, 34, 56, 789_012_345, time.UTC),
		SessionID: "sess-xyz",
	}
	encoded, err := encodeListSessionsCursor(original)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	decoded, err := decodeListSessionsCursor(encoded)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !decoded.StartedAt.Equal(original.StartedAt) {
		t.Errorf("StartedAt round-trip: got %v want %v", decoded.StartedAt, original.StartedAt)
	}
	if decoded.SessionID != original.SessionID {
		t.Errorf("SessionID round-trip: got %q want %q", decoded.SessionID, original.SessionID)
	}
}

// TestDecodeListSessionsCursor_Empty verifiziert, dass ein leerer
// Cursor-Query als nil-Cursor (= keine Pagination) durchläuft, ohne
// Fehler.
func TestDecodeListSessionsCursor_Empty(t *testing.T) {
	t.Parallel()
	got, err := decodeListSessionsCursor("")
	if err != nil {
		t.Errorf("empty: expected nil error, got %v", err)
	}
	if got != nil {
		t.Errorf("empty: expected nil cursor, got %v", got)
	}
}

// TestEncodeCursor_NilReturnsEmpty verifiziert die Symmetrie zum
// decode-Empty-Pfad: der Handler ruft encode nur, wenn ein Cursor da
// ist, aber der defensive nil-Branch in beiden Encode-Funktionen
// muss leer + nil-Error zurückgeben.
func TestEncodeCursor_NilReturnsEmpty(t *testing.T) {
	t.Parallel()
	if got, err := encodeListSessionsCursor(nil); got != "" || err != nil {
		t.Errorf("encodeListSessionsCursor(nil) = (%q, %v), want (empty, nil)", got, err)
	}
	if got, err := encodeSessionEventsCursor(nil); got != "" || err != nil {
		t.Errorf("encodeSessionEventsCursor(nil) = (%q, %v), want (empty, nil)", got, err)
	}
}

// TestDecodeListSessionsCursor_Malformed deckt die einzelnen Decode-
// Stufen ab (Base64 → JSON → v-Wert → Pflichtfelder → unbekannte
// Felder), die alle in `errCursorInvalidMalformed` münden.
func TestDecodeListSessionsCursor_Malformed(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"not-base64":            "not-base64!",
		"valid-base64-not-json": encodeRaw("AA\xFF"),
		"v=0":                   encodeRaw(`{"v":0,"sa":"2026-04-28T12:00:00Z","sid":"s1"}`),
		"v=-1":                  encodeRaw(`{"v":-1,"sa":"2026-04-28T12:00:00Z","sid":"s1"}`),
		"unknown v":             encodeRaw(`{"v":99,"sa":"2026-04-28T12:00:00Z","sid":"s1"}`),
		"missing sa":            encodeRaw(`{"v":2,"sid":"s1"}`),
		"missing sid":           encodeRaw(`{"v":2,"sa":"2026-04-28T12:00:00Z"}`),
		"empty sid":             encodeRaw(`{"v":2,"sa":"2026-04-28T12:00:00Z","sid":""}`),
		"sa not parseable":      encodeRaw(`{"v":2,"sa":"not-a-time","sid":"s1"}`),
		"unknown extra field":   encodeRaw(`{"v":2,"sa":"2026-04-28T12:00:00Z","sid":"s1","extra":"x"}`),
	}
	for name, raw := range cases {
		if _, err := decodeListSessionsCursor(raw); !errors.Is(err, errCursorInvalidMalformed) {
			t.Errorf("%s: want errCursorInvalidMalformed, got %v", name, err)
		}
	}
}

// TestDecodeListSessionsCursor_Legacy verifiziert die dauerhafte
// Reject-Klasse: Cursor mit `pid`-Feld oder ohne/`v:1` aus dem
// `0.1.x`/`0.2.x`/`0.3.x`-Format → `errCursorInvalidLegacy`
// (ADR-0004 §6). PID-Check hat Vorrang vor v-Wert: ein hybrider
// `v:2`-Cursor mit zusätzlichem `pid` bleibt Legacy.
func TestDecodeListSessionsCursor_Legacy(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"pid present":   encodeRaw(`{"pid":"abc","sa":"2026-04-28T12:00:00Z","sid":"s1"}`),
		"v missing":     encodeRaw(`{"sa":"2026-04-28T12:00:00Z","sid":"s1"}`),
		"v=1 explicit":  encodeRaw(`{"v":1,"sa":"2026-04-28T12:00:00Z","sid":"s1"}`),
		"v=1 plus pid":  encodeRaw(`{"v":1,"pid":"x","sa":"2026-04-28T12:00:00Z","sid":"s1"}`),
		"v=2 plus pid":  encodeRaw(`{"v":2,"pid":"x","sa":"2026-04-28T12:00:00Z","sid":"s1"}`),
		"pid only":      encodeRaw(`{"pid":"only"}`),
	}
	for name, raw := range cases {
		if _, err := decodeListSessionsCursor(raw); !errors.Is(err, errCursorInvalidLegacy) {
			t.Errorf("%s: want errCursorInvalidLegacy, got %v", name, err)
		}
	}
}

// TestEncodeDecodeSessionEventsCursor_RoundTrip — analog für die
// Event-Pagination, inkl. optional gesetzter SequenceNumber.
func TestEncodeDecodeSessionEventsCursor_RoundTrip(t *testing.T) {
	t.Parallel()
	seq := int64(42)
	original := &driving.SessionEventsCursor{
		ServerReceivedAt: time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC),
		SequenceNumber:   &seq,
		IngestSequence:   99,
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

// TestDecodeSessionEventsCursor_Malformed deckt Decode-Fehler analog
// zum Sessions-Cursor ab.
func TestDecodeSessionEventsCursor_Malformed(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"v=0":                 encodeRaw(`{"v":0,"rcv":"2026-04-28T12:00:00Z","ing":1}`),
		"unknown v":           encodeRaw(`{"v":7,"rcv":"2026-04-28T12:00:00Z","ing":1}`),
		"missing rcv":         encodeRaw(`{"v":2,"ing":1}`),
		"missing ing":         encodeRaw(`{"v":2,"rcv":"2026-04-28T12:00:00Z"}`),
		"rcv not parseable":   encodeRaw(`{"v":2,"rcv":"not-a-time","ing":1}`),
		"unknown extra field": encodeRaw(`{"v":2,"rcv":"2026-04-28T12:00:00Z","ing":1,"x":"y"}`),
	}
	for name, raw := range cases {
		if _, err := decodeSessionEventsCursor(raw); !errors.Is(err, errCursorInvalidMalformed) {
			t.Errorf("%s: want errCursorInvalidMalformed, got %v", name, err)
		}
	}
}

// TestDecodeSessionEventsCursor_Legacy — Legacy-Detection analog zum
// Sessions-Cursor.
func TestDecodeSessionEventsCursor_Legacy(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"pid present":  encodeRaw(`{"pid":"abc","rcv":"2026-04-28T12:00:00Z","ing":1}`),
		"v missing":    encodeRaw(`{"rcv":"2026-04-28T12:00:00Z","ing":1}`),
		"v=1 explicit": encodeRaw(`{"v":1,"rcv":"2026-04-28T12:00:00Z","ing":1}`),
		"v=2 plus pid": encodeRaw(`{"v":2,"pid":"x","rcv":"2026-04-28T12:00:00Z","ing":1}`),
	}
	for name, raw := range cases {
		if _, err := decodeSessionEventsCursor(raw); !errors.Is(err, errCursorInvalidLegacy) {
			t.Errorf("%s: want errCursorInvalidLegacy, got %v", name, err)
		}
	}
}

// encodeRaw ist ein Helper für die obigen Tests — base64-url ohne
// Padding über das stdlib base64-Paket.
func encodeRaw(raw string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}
