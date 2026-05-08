package application_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/application"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driving"
)

// plan-0.9.5 §2 Tranche 1 (RAK-Wave-2 / extra-gates.md §3.2) —
// API-Hot-Path-Benchmarks für `make api-benchmark-smoke`.
//
// Budgets aus `docs/perf/budgets.md` §3 (initial, Tranche-0-Stand;
// noch nicht mess-basiert, sondern Architektur-basierte Obergrenzen):
//
//   - typische 100-Event-Batch (In-Memory-Repo): ≤ 10 ms / Batch
//   - maximale 100-Event-Batch (volle Validation):  ≤ 25 ms / Batch
//
// Beobachtungsphase laut Plan §2 DoD: erste N=3-5 grüne CI-Läufe
// bleiben non-blocking; danach landen die Smokes via
// `make benchmark-smoke` PR-blockierend in `make gates`.

// BenchmarkRegisterPlaybackEventBatch_Typical pinnt den typischen
// In-Memory-Pfad: 100 Events, einfache rebuffer_started-Form ohne
// reservierte Meta-Keys, default time-skew-frei. Die Stubs
// (`stubRepo`/`stubLimiter`/...) leben in
// register_playback_event_batch_test.go und werden hier
// wiederverwendet.
func BenchmarkRegisterPlaybackEventBatch_Typical(b *testing.B) {
	uc, _, _, _, _, _, _, _ := newUseCase()
	batch := makeBatch(100, false)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := uc.RegisterPlaybackEventBatch(ctx, batch)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

// BenchmarkRegisterPlaybackEventBatch_MaxBatch pinnt den maximalen
// Batch-Pfad (spec/telemetry-model.md §4.1: 100 Events / 256 KiB
// Body) inklusive Per-Event-Meta-Validation für
// `network.*`-/`webrtc.*`-Reserve-Namespace-Keys.
func BenchmarkRegisterPlaybackEventBatch_MaxBatch(b *testing.B) {
	uc, _, _, _, _, _, _, _ := newUseCase()
	batch := makeBatch(application.MaxBatchSize, true)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := uc.RegisterPlaybackEventBatch(ctx, batch)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

// makeBatch erzeugt einen Batch mit n Events. `withReservedMeta`
// schaltet `network.*`-Keys hinzu, damit der Per-Event-Meta-
// Validation-Pfad mitgemessen wird (Worst-Case-Ingest aus dem
// 0.4.0-Trace-Korrelations-Pfad).
func makeBatch(n int, withReservedMeta bool) driving.BatchInput {
	events := make([]driving.EventInput, 0, n)
	base := time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC)
	for i := 0; i < n; i++ {
		event := driving.EventInput{
			EventName: pickEventName(i, withReservedMeta),
			ProjectID: "demo",
			// Synthetische Session-IDs in ULID-Form sind Pflicht,
			// damit der Validator nicht früh ablehnt.
			SessionID:       fmt.Sprintf("01J7K9X4Z2QHB6V3WS5R8Y%03dF", i%1000),
			ClientTimestamp: base.Add(time.Duration(i) * time.Millisecond).Format(time.RFC3339Nano),
			SequenceNumber:  int64Ptr(int64(i + 1)),
			SDK:             driving.SDKInput{Name: "@npm9912/player-sdk", Version: "0.9.6"},
		}
		if withReservedMeta {
			event.Meta = map[string]any{
				"network.kind":           "manifest",
				"network.detail_status":  "available",
				"network.redacted_url":   "https://cdn.example.test/manifest.m3u8",
			}
		}
		events = append(events, event)
	}
	return driving.BatchInput{
		SchemaVersion: application.SupportedSchemaVersion,
		AuthToken:     "demo-token",
		Events:        events,
	}
}

func pickEventName(i int, withReservedMeta bool) string {
	if withReservedMeta {
		return "manifest_loaded"
	}
	switch i % 3 {
	case 0:
		return "playback_started"
	case 1:
		return "rebuffer_started"
	default:
		return "playback_paused"
	}
}

func int64Ptr(v int64) *int64 { return &v }
