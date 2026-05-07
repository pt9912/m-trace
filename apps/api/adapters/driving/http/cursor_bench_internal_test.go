package http

import (
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/port/driving"
)

// plan-0.9.5 §2 Tranche 1 — Cursor-Hot-Path-Bench für
// `make api-benchmark-smoke`.
//
// Budget aus `docs/perf/budgets.md` §3 (initial, Tranche-0-Stand):
//   - cursor.Encode/Decode (Cursor-v3 inkl. Process-Instance-Stamp):
//     ≤ 250 µs / Pair (typische Pagination-Latenz, kein HMAC-Sign).

// BenchmarkCursorEncodeDecode_Pair pinnt einen kompletten Encode→
// Decode-Roundtrip eines `ListSessionsCursor`. Cursor-v3 ist seit
// plan-0.4.0 produktiv; Drift im Encode/Decode-Pfad wirkt direkt
// auf jede Sessions-Listing-Antwort.
func BenchmarkCursorEncodeDecode_Pair(b *testing.B) {
	cursor := &driving.ListSessionsCursor{
		StartedAt: time.Date(2026, 4, 28, 12, 34, 56, 789_012_345, time.UTC),
		SessionID: "sess-xyz",
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encoded, err := encodeListSessionsCursor(cursor, testProjectID)
		if err != nil {
			b.Fatalf("encode: %v", err)
		}
		if _, err := decodeListSessionsCursor(encoded, testProjectID); err != nil {
			b.Fatalf("decode: %v", err)
		}
	}
}
