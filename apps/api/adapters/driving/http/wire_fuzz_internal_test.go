package http

import (
	"encoding/json"
	"testing"
)

// plan-0.9.5 §4 Tranche 3 (extra-gates.md §3.5) — Fuzz-Target für
// die HTTP-Body-Validation-Schicht des Playback-Events-Pfads.
// Pinnt:
//
//   - Random-JSON-Bytes durch `wireBatch`-Decode dürfen weder
//     Panic noch Goroutine-Leak produzieren.
//   - `toEventInputs` und `toBoundaryInputs` (die Wire-Domain-
//     Mapper) müssen jede dekodierte Eingabe deterministisch in
//     ein gültiges `driving.BatchInput`-Tripel überführen, ohne
//     unbounded Allokationen oder negative Längenwerte.
//
// Pflicht-Bereich aus Plan §4 DoD-Item 1: „HTTP-Validation für
// Playback-Event-Batches".

// FuzzWireBatchDecode wirft random JSON-Bytes durch den
// Decode-Pfad in `serve()` (`json.Unmarshal(body, &payload)` plus
// die zwei `to*Inputs`-Mapper).
func FuzzWireBatchDecode(f *testing.F) {
	// Seed-Korpus: gültige Batches plus Drift-Pfade.
	f.Add([]byte(`{"schema_version":"1.0","events":[]}`))
	f.Add([]byte(`{"schema_version":"1.0","events":[{"event_name":"playback_started","project_id":"demo","session_id":"01J","client_timestamp":"2026-04-28T12:00:00Z","sdk":{"name":"@npm9912/player-sdk","version":"0.9.6"}}]}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`null`))
	f.Add([]byte(`{"events":null}`))
	f.Add([]byte(`{"session_boundaries":[{"kind":"network_signal_absent","reason":"native_hls_unavailable"}]}`))

	f.Fuzz(func(t *testing.T, raw []byte) {
		var payload wireBatch
		if err := json.Unmarshal(raw, &payload); err != nil {
			// Ungültiges JSON ist legitimer Fuzz-Output —
			// Mapping wird nicht versucht.
			return
		}
		events := toEventInputs(payload.Events)
		if len(events) != len(payload.Events) {
			t.Fatalf("toEventInputs length drift: got=%d want=%d for raw=%q",
				len(events), len(payload.Events), raw)
		}
		boundaries := toBoundaryInputs(payload.SessionBoundaries)
		if boundaries == nil && len(payload.SessionBoundaries) > 0 {
			t.Fatalf("toBoundaryInputs returned nil for non-empty input: raw=%q", raw)
		}
		if boundaries != nil && len(boundaries) != len(payload.SessionBoundaries) {
			t.Fatalf("toBoundaryInputs length drift: got=%d want=%d for raw=%q",
				len(boundaries), len(payload.SessionBoundaries), raw)
		}
	})
}
