package http

import (
	"strconv"
	"strings"
)

// W3C Trace Context traceparent header
// (https://www.w3.org/TR/trace-context/#traceparent-header-field-values).
// Format: `version-trace_id-parent_id-flags`
//   - version  : 2 hex; in 0.4.0 wird nur Version "00" akzeptiert.
//   - trace_id : 32 hex (16 Bytes), nicht all-zero.
//   - parent_id: 16 hex (8 Bytes), nicht all-zero.
//   - flags    : 2 hex (sampled-flag etc.); aktuell nicht ausgewertet.
//
// Defensive parser laut spec/telemetry-model.md §2.5: jeder Formatfehler
// → ok=false; Caller setzt das Span-Attribut mtrace.trace.parse_error
// und fällt auf einen Server-Root-Span zurück.

const (
	traceParentLen     = 55 // "00-" (3) + 32 + "-" (1) + 16 + "-" (1) + 2
	traceIDHexLen      = 32
	spanIDHexLen       = 16
	traceParentVersion = "00"
	zeroTraceID        = "00000000000000000000000000000000"
	zeroSpanID         = "0000000000000000"
)

// parseTraceParent parst den Header-Wert. Liefert die hex-codierte
// trace_id, parent_id und das flags-Byte (W3C Trace Context, niedrigste
// 8 Bit; Bit 0 = sampled), wenn der Wert formal gültig ist; sonst
// ok=false und alle Felder zero.
//
// Die `flags` müssen durchgeschleift werden: ohne sie würde
// `Tracer.Start` mit einem `ParentBased(AlwaysSample)`-Sampler den
// Server-Span als not-sampled werten, sobald der Client `flags=00`
// schickt — und sampled-Spans (`flags=01`) gingen in der Test-Suite
// verloren.
func parseTraceParent(raw string) (traceID, parentID string, flags byte, ok bool) {
	if len(raw) != traceParentLen {
		return "", "", 0, false
	}
	parts := strings.Split(raw, "-")
	if len(parts) != 4 {
		return "", "", 0, false
	}
	if parts[0] != traceParentVersion {
		return "", "", 0, false
	}
	if len(parts[1]) != traceIDHexLen || !isLowerHex(parts[1]) {
		return "", "", 0, false
	}
	if len(parts[2]) != spanIDHexLen || !isLowerHex(parts[2]) {
		return "", "", 0, false
	}
	if len(parts[3]) != 2 || !isLowerHex(parts[3]) {
		return "", "", 0, false
	}
	if parts[1] == zeroTraceID || parts[2] == zeroSpanID {
		return "", "", 0, false
	}
	flagsValue, err := strconv.ParseUint(parts[3], 16, 8)
	if err != nil {
		return "", "", 0, false
	}
	return parts[1], parts[2], byte(flagsValue), true
}

// isLowerHex prüft, ob ein String nur lower-case-Hex-Zeichen enthält.
// W3C trace-context fordert lower-case; großgeschriebene Hex-Werte
// sind nicht standardkonform und werden hier als Parse-Error behandelt.
func isLowerHex(s string) bool {
	for _, c := range s {
		switch {
		case c >= '0' && c <= '9':
		case c >= 'a' && c <= 'f':
		default:
			return false
		}
	}
	return true
}
