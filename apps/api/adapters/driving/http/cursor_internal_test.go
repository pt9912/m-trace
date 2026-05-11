package http

import (
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driving"
)

// testProjectID lebt in sessions_handlers_unit_internal_test.go
// (gleiche Package, ein Test-Setup).

// TestEncodeDecodeListSessionsCursor_RoundTrip verifiziert, dass
// Encoding und Decoding eines ListSessionsCursor sich gegenseitig
// neutralisieren — keine Verluste in StartedAt (auf nano-Genauigkeit
// im UTC-Frame) und SessionID. Cursor-v3 trägt zusätzlich einen
// Project-Scope (`pid`), der beim Decode gegen den Request-Project-
// Kontext geprüft wird (plan-0.4.0 §4.3 / API-Kontrakt §10.3).
func TestEncodeDecodeListSessionsCursor_RoundTrip(t *testing.T) {
	t.Parallel()
	original := &driving.ListSessionsCursor{
		StartedAt: time.Date(2026, 4, 28, 12, 34, 56, 789_012_345, time.UTC),
		SessionID: "sess-xyz",
	}
	encoded, err := encodeListSessionsCursor(original, testProjectID)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	decoded, err := decodeListSessionsCursor(encoded, testProjectID)
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
	got, err := decodeListSessionsCursor("", testProjectID)
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
	if got, err := encodeListSessionsCursor(nil, testProjectID); got != "" || err != nil {
		t.Errorf("encodeListSessionsCursor(nil) = (%q, %v), want (empty, nil)", got, err)
	}
	if got, err := encodeSessionEventsCursor(nil, testProjectID, "sid"); got != "" || err != nil {
		t.Errorf("encodeSessionEventsCursor(nil) = (%q, %v), want (empty, nil)", got, err)
	}
}

// TestDecodeListSessionsCursor_Malformed deckt die einzelnen Decode-
// Stufen ab (Base64 → JSON → v-Wert → Pflichtfelder → unbekannte
// Felder → Project-Scope-Mismatch), die alle in
// `errCursorInvalidMalformed` münden.
func TestDecodeListSessionsCursor_Malformed(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"not-base64":            "not-base64!",
		"valid-base64-not-json": encodeRaw("AA\xFF"),
		"v=0":                   encodeRaw(`{"v":0,"pid":"demo","sa":"2026-04-28T12:00:00Z","sid":"s1"}`),
		"v=-1":                  encodeRaw(`{"v":-1,"pid":"demo","sa":"2026-04-28T12:00:00Z","sid":"s1"}`),
		"unknown v":             encodeRaw(`{"v":99,"pid":"demo","sa":"2026-04-28T12:00:00Z","sid":"s1"}`),
		"v3 missing pid":        encodeRaw(`{"v":3,"sa":"2026-04-28T12:00:00Z","sid":"s1"}`),
		"v3 missing sa":         encodeRaw(`{"v":3,"pid":"demo","sid":"s1"}`),
		"v3 missing sid":        encodeRaw(`{"v":3,"pid":"demo","sa":"2026-04-28T12:00:00Z"}`),
		"v3 empty sid":          encodeRaw(`{"v":3,"pid":"demo","sa":"2026-04-28T12:00:00Z","sid":""}`),
		"v3 sa not parseable":   encodeRaw(`{"v":3,"pid":"demo","sa":"not-a-time","sid":"s1"}`),
		"v3 unknown field":      encodeRaw(`{"v":3,"pid":"demo","sa":"2026-04-28T12:00:00Z","sid":"s1","extra":"x"}`),
		"v3 foreign project":    encodeRaw(`{"v":3,"pid":"other","sa":"2026-04-28T12:00:00Z","sid":"s1"}`),
	}
	for name, raw := range cases {
		if _, err := decodeListSessionsCursor(raw, testProjectID); !errors.Is(err, errCursorInvalidMalformed) {
			t.Errorf("%s: want errCursorInvalidMalformed, got %v", name, err)
		}
	}
}

// TestDecodeListSessionsCursor_Legacy verifiziert die dauerhafte
// Reject-Klasse: Cursor ohne `v`-Feld oder mit `v=1`/`v=2` aus dem
// `0.1.x`-/`0.2.0`-/`0.3.x`-Format → `errCursorInvalidLegacy`. v=2
// gilt ab plan-0.4.0 §4.3 als Legacy, weil ihm der Project-Scope
// fehlt.
func TestDecodeListSessionsCursor_Legacy(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"v missing":      encodeRaw(`{"sa":"2026-04-28T12:00:00Z","sid":"s1"}`),
		"v=1 explicit":   encodeRaw(`{"v":1,"sa":"2026-04-28T12:00:00Z","sid":"s1"}`),
		"v=1 plus pid":   encodeRaw(`{"v":1,"pid":"x","sa":"2026-04-28T12:00:00Z","sid":"s1"}`),
		"v=2 (no scope)": encodeRaw(`{"v":2,"sa":"2026-04-28T12:00:00Z","sid":"s1"}`),
		"v=2 plus pid":   encodeRaw(`{"v":2,"pid":"x","sa":"2026-04-28T12:00:00Z","sid":"s1"}`),
		"pid only":       encodeRaw(`{"pid":"only"}`),
	}
	for name, raw := range cases {
		if _, err := decodeListSessionsCursor(raw, testProjectID); !errors.Is(err, errCursorInvalidLegacy) {
			t.Errorf("%s: want errCursorInvalidLegacy, got %v", name, err)
		}
	}
}

// TestEncodeDecodeSessionEventsCursor_RoundTrip — analog für die
// Event-Pagination, inkl. optional gesetzter SequenceNumber. Cursor-v3
// trägt zusätzlich Collection-Scope `(pid, sid)`.
func TestEncodeDecodeSessionEventsCursor_RoundTrip(t *testing.T) {
	t.Parallel()
	seq := int64(42)
	original := &driving.SessionEventsCursor{
		ServerReceivedAt: time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC),
		SequenceNumber:   &seq,
		IngestSequence:   99,
	}
	const sessionID = "sess-xyz"
	encoded, err := encodeSessionEventsCursor(original, testProjectID, sessionID)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	decoded, err := decodeSessionEventsCursor(encoded, testProjectID, sessionID)
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
// zum Sessions-Cursor ab plus Collection-Scope-Mismatch (fremdes
// Project oder fremde Session).
func TestDecodeSessionEventsCursor_Malformed(t *testing.T) {
	t.Parallel()
	const sessionID = "sess-xyz"
	cases := map[string]string{
		"v=0":                 encodeRaw(`{"v":0,"pid":"demo","sid":"sess-xyz","rcv":"2026-04-28T12:00:00Z","ing":1}`),
		"unknown v":           encodeRaw(`{"v":7,"pid":"demo","sid":"sess-xyz","rcv":"2026-04-28T12:00:00Z","ing":1}`),
		"v3 missing pid":      encodeRaw(`{"v":3,"sid":"sess-xyz","rcv":"2026-04-28T12:00:00Z","ing":1}`),
		"v3 missing sid":      encodeRaw(`{"v":3,"pid":"demo","rcv":"2026-04-28T12:00:00Z","ing":1}`),
		"v3 missing rcv":      encodeRaw(`{"v":3,"pid":"demo","sid":"sess-xyz","ing":1}`),
		"v3 missing ing":      encodeRaw(`{"v":3,"pid":"demo","sid":"sess-xyz","rcv":"2026-04-28T12:00:00Z"}`),
		"v3 rcv not time":     encodeRaw(`{"v":3,"pid":"demo","sid":"sess-xyz","rcv":"not-a-time","ing":1}`),
		"v3 unknown field":    encodeRaw(`{"v":3,"pid":"demo","sid":"sess-xyz","rcv":"2026-04-28T12:00:00Z","ing":1,"x":"y"}`),
		"v3 foreign project":  encodeRaw(`{"v":3,"pid":"other","sid":"sess-xyz","rcv":"2026-04-28T12:00:00Z","ing":1}`),
		"v3 foreign session":  encodeRaw(`{"v":3,"pid":"demo","sid":"other-session","rcv":"2026-04-28T12:00:00Z","ing":1}`),
	}
	for name, raw := range cases {
		if _, err := decodeSessionEventsCursor(raw, testProjectID, sessionID); !errors.Is(err, errCursorInvalidMalformed) {
			t.Errorf("%s: want errCursorInvalidMalformed, got %v", name, err)
		}
	}
}

// TestDecodeSessionEventsCursor_Legacy — Legacy-Detection analog zum
// Sessions-Cursor; v=2 gilt ab §4.3 als Legacy.
func TestDecodeSessionEventsCursor_Legacy(t *testing.T) {
	t.Parallel()
	const sessionID = "sess-xyz"
	cases := map[string]string{
		"v missing":      encodeRaw(`{"rcv":"2026-04-28T12:00:00Z","ing":1}`),
		"v=1 explicit":   encodeRaw(`{"v":1,"rcv":"2026-04-28T12:00:00Z","ing":1}`),
		"v=1 plus pid":   encodeRaw(`{"v":1,"pid":"x","rcv":"2026-04-28T12:00:00Z","ing":1}`),
		"v=2 (no scope)": encodeRaw(`{"v":2,"rcv":"2026-04-28T12:00:00Z","ing":1}`),
		"v=2 plus pid":   encodeRaw(`{"v":2,"pid":"x","rcv":"2026-04-28T12:00:00Z","ing":1}`),
		"pid only":       encodeRaw(`{"pid":"only"}`),
	}
	for name, raw := range cases {
		if _, err := decodeSessionEventsCursor(raw, testProjectID, sessionID); !errors.Is(err, errCursorInvalidLegacy) {
			t.Errorf("%s: want errCursorInvalidLegacy, got %v", name, err)
		}
	}
}

// encodeRaw ist ein Helper für die obigen Tests — base64-url ohne
// Padding über das stdlib base64-Paket.
func encodeRaw(raw string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

// TestEncodeDecodeSrtHealthCursor_RoundTrip (plan-0.12.6 Tranche 2):
// encode→decode-Roundtrip mit Collection-Scope (pid, sid). Cursor
// muss IngestedAt UTC-normalisiert zurückliefern (RFC3339Nano-
// Serialisierung schluckt Sub-Nano-Genauigkeit nicht).
func TestEncodeDecodeSrtHealthCursor_RoundTrip(t *testing.T) {
	t.Parallel()
	const streamID = "srt-test"
	original := &driven.SrtHealthCursor{
		IngestedAt: time.Date(2026, 5, 5, 8, 48, 1, 250_000_000, time.UTC),
		ID:         4711,
	}
	encoded, err := encodeSrtHealthCursor(original, testProjectID, streamID)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if encoded == "" {
		t.Fatal("encode returned empty token for non-nil cursor")
	}
	decoded, err := decodeSrtHealthCursor(encoded, testProjectID, streamID)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if decoded == nil {
		t.Fatal("decode returned nil")
	}
	if decoded.ID != original.ID {
		t.Errorf("ID = %d, want %d", decoded.ID, original.ID)
	}
	if !decoded.IngestedAt.Equal(original.IngestedAt) {
		t.Errorf("IngestedAt = %v, want %v", decoded.IngestedAt, original.IngestedAt)
	}
}

// TestEncodeSrtHealthCursor_NilReturnsEmpty: encode mit nil-Cursor
// liefert ("", nil) — der Handler nutzt das, um `next_cursor` auf
// der letzten Page wegzulassen.
func TestEncodeSrtHealthCursor_NilReturnsEmpty(t *testing.T) {
	t.Parallel()
	got, err := encodeSrtHealthCursor(nil, testProjectID, "any-stream")
	if got != "" || err != nil {
		t.Fatalf("encode(nil) = (%q, %v), want (\"\", nil)", got, err)
	}
}

// TestDecodeSrtHealthCursor_EmptyReturnsNil: decode mit leerem
// Token liefert (nil, nil) — Pagination beginnt auf der ersten Page.
func TestDecodeSrtHealthCursor_EmptyReturnsNil(t *testing.T) {
	t.Parallel()
	got, err := decodeSrtHealthCursor("", testProjectID, "any-stream")
	if got != nil || err != nil {
		t.Fatalf("decode(\"\") = (%+v, %v), want (nil, nil)", got, err)
	}
}

// TestDecodeSrtHealthCursor_Legacy: v=1/v=2/v fehlt → Legacy-Reject.
func TestDecodeSrtHealthCursor_Legacy(t *testing.T) {
	t.Parallel()
	const streamID = "srt-test"
	cases := map[string]string{
		"v missing":   encodeRaw(`{"pid":"demo","sid":"srt-test","ing":"2026-05-05T08:48:01Z","id":1}`),
		"v=1":         encodeRaw(`{"v":1,"pid":"demo","sid":"srt-test","ing":"2026-05-05T08:48:01Z","id":1}`),
		"v=2":         encodeRaw(`{"v":2,"pid":"demo","sid":"srt-test","ing":"2026-05-05T08:48:01Z","id":1}`),
		"pid only v1": encodeRaw(`{"v":1,"pid":"demo"}`),
	}
	for name, raw := range cases {
		if _, err := decodeSrtHealthCursor(raw, testProjectID, streamID); !errors.Is(err, errCursorInvalidLegacy) {
			t.Errorf("%s: want errCursorInvalidLegacy, got %v", name, err)
		}
	}
}

// TestDecodeSrtHealthCursor_Malformed: alle anderen Reject-Pfade →
// errCursorInvalidMalformed.
func TestDecodeSrtHealthCursor_Malformed(t *testing.T) {
	t.Parallel()
	const streamID = "srt-test"
	cases := map[string]string{
		"base64 garbage":  "!!!notbase64!!!",
		"valid base64 nonjson": base64.RawURLEncoding.EncodeToString([]byte("not json")),
		"v=99":            encodeRaw(`{"v":99,"pid":"demo","sid":"srt-test","ing":"2026-05-05T08:48:01Z","id":1}`),
		"missing pid":     encodeRaw(`{"v":3,"sid":"srt-test","ing":"2026-05-05T08:48:01Z","id":1}`),
		"missing sid":     encodeRaw(`{"v":3,"pid":"demo","ing":"2026-05-05T08:48:01Z","id":1}`),
		"missing ing":     encodeRaw(`{"v":3,"pid":"demo","sid":"srt-test","id":1}`),
		"missing id":      encodeRaw(`{"v":3,"pid":"demo","sid":"srt-test","ing":"2026-05-05T08:48:01Z"}`),
		"foreign project": encodeRaw(`{"v":3,"pid":"other","sid":"srt-test","ing":"2026-05-05T08:48:01Z","id":1}`),
		"foreign stream":  encodeRaw(`{"v":3,"pid":"demo","sid":"other","ing":"2026-05-05T08:48:01Z","id":1}`),
		"ing not time":    encodeRaw(`{"v":3,"pid":"demo","sid":"srt-test","ing":"not-a-time","id":1}`),
		"unknown field":   encodeRaw(`{"v":3,"pid":"demo","sid":"srt-test","ing":"2026-05-05T08:48:01Z","id":1,"x":"y"}`),
	}
	for name, raw := range cases {
		if _, err := decodeSrtHealthCursor(raw, testProjectID, streamID); !errors.Is(err, errCursorInvalidMalformed) {
			t.Errorf("%s: want errCursorInvalidMalformed, got %v", name, err)
		}
	}
}
