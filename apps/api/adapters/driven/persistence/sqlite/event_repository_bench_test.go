package sqlite_test

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/persistence/sqlite"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/internal/storage"
)

// plan-0.9.5 §2 Tranche 1 — SQLite-Persistence-Hot-Path-Bench für
// `make api-benchmark-smoke`.
//
// Budgets aus `docs/perf/budgets.md` §3 (initial, Tranche-0-Stand):
//   - Event-Append + Sequence-Allocation (typische 100-Event-Batch):
//     ≤ 100 ms / Batch (CI-Runner ohne tmpfs-Boost).

// BenchmarkEventRepository_AppendBatch_100 misst den Durable-
// Schreibpfad: NewEventRepository.Append plus IngestSequencer.Allocate
// für jeden Event. Die DB-Datei lebt im b.TempDir() und wird pro
// Iteration neu vorbereitet, damit WAL-Wachstum den Bench nicht
// dominiert.
func BenchmarkEventRepository_AppendBatch_100(b *testing.B) {
	ctx := context.Background()
	base := time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		path := filepath.Join(b.TempDir(), fmt.Sprintf("bench-%d.db", i))
		db, err := storage.Open(ctx, path)
		if err != nil {
			b.Fatalf("storage.Open: %v", err)
		}
		seq, err := sqlite.NewIngestSequencer(ctx, db)
		if err != nil {
			b.Fatalf("NewIngestSequencer: %v", err)
		}
		repo := sqlite.NewEventRepository(db)

		events := make([]domain.PlaybackEvent, 0, 100)
		for j := 0; j < 100; j++ {
			ts := base.Add(time.Duration(j) * time.Millisecond)
			seqNum := int64(j + 1)
			events = append(events, domain.PlaybackEvent{
				EventName:        "playback_started",
				ProjectID:        "demo",
				SessionID:        fmt.Sprintf("01J7K9X4Z2QHB6V3WS5R8Y%03dF", j%1000),
				ClientTimestamp:  ts,
				ServerReceivedAt: ts,
				IngestSequence:   seq.Next(),
				SequenceNumber:   &seqNum,
				SDK:              domain.SDKInfo{Name: "@npm9912/player-sdk", Version: "0.11.0"},
			})
		}
		b.StartTimer()

		if err := repo.Append(ctx, events); err != nil {
			b.Fatalf("Append: %v", err)
		}

		b.StopTimer()
		_ = db.Close()
		b.StartTimer()
	}
}
