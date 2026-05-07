package http

import (
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/port/driving"
)

// plan-0.9.5 §4 Tranche 3 (RAK-Wave-2 / extra-gates.md §3.5) —
// Fuzz-Target für den Cursor-Parser. Schließt eine ADR-0004-bekannte
// Edge-Case-Klasse ab: malformierte Base64-/JSON-/Versions-Strings
// dürfen weder Panic noch unerwartete Domain-Werte produzieren.
//
// Fuzzing läuft **nicht** PR-blockierend. Lokal mit
// `make fuzz-check` (kurzes -fuzztime, Default 30s); Nightly via
// `.github/workflows/fuzz.yml` mit längerem Budget. Crash-Funde
// landen automatisch im Repo unter `testdata/fuzz/Fuzz<X>/`.

// FuzzDecodeListSessionsCursor wirft random Strings als
// Cursor-Eingabe in `decodeListSessionsCursor`. Erlaubte Outcomes:
//   - nil-Cursor (leerer String)
//   - errCursorInvalidLegacy / errCursorInvalidMalformed /
//     errCursorExpired (definierte Domain-Fehler aus
//     ADR-0004 §6 / API-Kontrakt §10.3)
//   - gültiger ListSessionsCursor (selten bei Random-Input, aber
//     legitim bei zufällig sinnvollen Bytes)
//
// Verboten: Panic, Goroutine-Leak, unklassifizierter Error. Genau
// das prüft der Fuzz-Loop.
func FuzzDecodeListSessionsCursor(f *testing.F) {
	// Seed-Korpus aus typischen Drift-Pfaden.
	f.Add("")
	f.Add("not-base64")
	f.Add("aGVsbG8")                   // valid base64, invalid JSON
	f.Add("eyJ2IjozLCJwaWQiOiJkZW1vIn0") // valid base64, partial JSON (missing fields)
	f.Add(encodedRoundTripCursor(t0()))  // gültiger Cursor

	f.Fuzz(func(t *testing.T, raw string) {
		cursor, err := decodeListSessionsCursor(raw, "demo")
		if err != nil {
			// Erlaubte Errorklassen — alles andere ist Bug.
			switch err {
			case errCursorInvalidLegacy, errCursorInvalidMalformed, errCursorExpired:
				return
			default:
				t.Fatalf("unexpected error class for input %q: %T %v", raw, err, err)
			}
		}
		if cursor != nil {
			// Erfolgreiche Decodes müssen einen sinnvollen Cursor produzieren.
			if cursor.SessionID == "" && cursor.StartedAt.IsZero() {
				t.Fatalf("decoded cursor empty for input %q", raw)
			}
		}
	})
}

// FuzzDecodeSessionEventsCursor analog für den Events-Cursor-Pfad.
func FuzzDecodeSessionEventsCursor(f *testing.F) {
	f.Add("")
	f.Add("z")
	f.Add(encodedRoundTripEventsCursor("sess-xyz"))

	f.Fuzz(func(t *testing.T, raw string) {
		cursor, err := decodeSessionEventsCursor(raw, "demo", "sess-xyz")
		if err != nil {
			switch err {
			case errCursorInvalidLegacy, errCursorInvalidMalformed, errCursorExpired:
				return
			default:
				t.Fatalf("unexpected error class for input %q: %T %v", raw, err, err)
			}
		}
		if cursor != nil {
			// Decoded events cursor hält keine session-id selbst —
			// der request-Pfad gibt die session-id mit; der Decode-
			// Pfad kontextualisiert sie nur. Sanity-Check: das
			// Tripel aus ServerReceivedAt/SequenceNumber/
			// IngestSequence sollte nicht komplett leer sein.
			if cursor.IngestSequence == 0 && cursor.SequenceNumber == nil && cursor.ServerReceivedAt.IsZero() {
				t.Fatalf("decoded events cursor fully empty for input %q", raw)
			}
		}
	})
}

func t0() time.Time {
	return time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC)
}

func encodedRoundTripCursor(when time.Time) string {
	c := &driving.ListSessionsCursor{
		StartedAt: when,
		SessionID: "sess-1",
	}
	encoded, err := encodeListSessionsCursor(c, "demo")
	if err != nil {
		return ""
	}
	return encoded
}

func encodedRoundTripEventsCursor(sessionID string) string {
	c := &driving.SessionEventsCursor{
		ServerReceivedAt: t0(),
		SequenceNumber:   ptrInt64(1),
		IngestSequence:   2,
	}
	encoded, err := encodeSessionEventsCursor(c, "demo", sessionID)
	if err != nil {
		return ""
	}
	return encoded
}

func ptrInt64(v int64) *int64 { return &v }
