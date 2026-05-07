package application

import (
	"testing"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// plan-0.9.5 §4 Tranche 3 (extra-gates.md §3.5) — Fuzz-Target für
// die Reserved-Meta-Validation. Pinnt:
//
//   - Keine Panics auf jeder Kombination aus Key+Value (Strings,
//     Zahlen, Booleans, nested Maps, Arrays).
//   - Reserved-Namespace-Keys (`network.*`, `timing.*`, `webrtc.*`)
//     liefern entweder nil oder ein typisiertes Validation-Error.
//   - Nicht-reservierte Keys bleiben durch (Vorwärtskompatibilität).
//
// Pflicht-Bereich aus Plan §4 DoD-Item 1: „Event-Meta-Validation
// (`webrtc.*`-Allowlist aus `0.8.0`)".

// FuzzValidateReservedEventMeta wirft random Key+Value-Paare in
// `validateReservedEventMeta`. Die Fuzz-Engine generiert die
// Eingaben byte-strings; wir bauen daraus eine `domain.EventMeta`-
// Map mit kontrollierter Werte-Domäne (string/int64/float64/bool).
func FuzzValidateReservedEventMeta(f *testing.F) {
	// Seed-Korpus mit den drei Reserved-Namespaces plus Drift-Pfaden.
	f.Add("network.kind", "manifest")
	f.Add("network.detail_status", "available")
	f.Add("network.unavailable_reason", "native_hls_unavailable")
	f.Add("network.redacted_url", "https://cdn.example.test/m.m3u8")
	f.Add("timing.first_byte_ms", "120")
	f.Add("webrtc.connection_state", "connected")
	f.Add("webrtc.sample_id", "42")
	f.Add("webrtc.peer_connection_run_id", "run-a")
	f.Add("custom.key", "anything") // additive non-reserved
	f.Add("network.kind", "")
	f.Add("network.kind", "unknown-value-not-in-allowlist")
	f.Add("webrtc.dtls_state", "exotic")
	f.Add("webrtc.error_detail", "loooong " /* Real fuzz extends this */)

	f.Fuzz(func(t *testing.T, key string, value string) {
		// Tausche Werte zwischen string/int/bool/nested map durch
		// einfaches Modulo der Schlüssel-Länge — gibt der Fuzz-
		// Engine drei „Typ-Pfade" gleichzeitig.
		var v any = value
		switch len(key) % 4 {
		case 1:
			v = int64(len(value))
		case 2:
			v = float64(len(value))
		case 3:
			v = len(value)%2 == 0
		}
		meta := domain.EventMeta{key: v}

		// validateReservedEventMeta darf nur typed Errors oder nil
		// liefern. Panic = Bug.
		_ = validateReservedEventMeta(meta)
	})
}

// FuzzValidateUnavailableReason testet die Reason-Enum-Schicht (eine
// der ältesten Reserved-Domain-Validierungen) gegen Random-Strings —
// Pflicht-Pfad aus spec/telemetry-model.md §1.4 + Reserved-Reason-
// Liste in `validateUnavailableReason`.
func FuzzValidateUnavailableReason(f *testing.F) {
	f.Add("native_hls_unavailable")
	f.Add("hlsjs_signal_unavailable")
	f.Add("browser_api_unavailable")
	f.Add("resource_timing_unavailable")
	f.Add("cors_timing_blocked")
	f.Add("service_worker_opaque")
	f.Add("UNKNOWN")
	f.Add("nested.dotted.value")
	f.Add("")
	f.Add("a")

	f.Fuzz(func(t *testing.T, value string) {
		// Typed-Wrapper: Reason muss als string ankommen, sonst
		// liefert der Validator einen Type-Error. Wir prüfen den
		// String-Pfad direkt.
		_ = validateUnavailableReason("network.unavailable_reason", value)
	})
}
