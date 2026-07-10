package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/persistence/postgres"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
	"github.com/pt9912/m-trace/apps/api/internal/storage"
)

// TestSrtHealthRepository_PgLab deckt den Postgres-srt_health-Adapter
// gegen eine echte PG-DB ab (Append + Dedupe, LatestByStream,
// HistoryByStream-Keyset-Pagination). Gated über MTRACE_PG_LAB_DSN
// (siehe scripts/smoke-pg-lab.sh). Nutzt einen eindeutigen Projekt-/
// Stream-Präfix, um von den anderen PG-Lab-Tests isoliert zu sein.
func TestSrtHealthRepository_PgLab(t *testing.T) {
	dsn := os.Getenv("MTRACE_PG_LAB_DSN")
	if dsn == "" {
		t.Skip("MTRACE_PG_LAB_DSN nicht gesetzt — PG-Lab-Integrationstest übersprungen")
	}
	ctx := context.Background()
	db, err := storage.OpenPostgres(ctx, dsn)
	if err != nil {
		t.Fatalf("OpenPostgres: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	repo := postgres.NewSrtHealthRepository(db)
	const proj = "srt-lab-proj"
	const stream = "srt-lab-stream"
	base := time.Date(2026, 7, 10, 10, 0, 0, 0, time.UTC)

	// Drei Samples mit unterschiedlichem Dedupe-Key (source_sequence)
	// und aufsteigendem ingested_at.
	samples := []domain.SrtHealthSample{
		sample(proj, stream, "conn-a", "seq-1", base),
		sample(proj, stream, "conn-a", "seq-2", base.Add(1*time.Second)),
		sample(proj, stream, "conn-a", "seq-3", base.Add(2*time.Second)),
	}
	if err := repo.Append(ctx, samples); err != nil {
		t.Fatalf("Append: %v", err)
	}

	t.Run("dedupe skip on re-append", func(t *testing.T) {
		if err := repo.Append(ctx, samples); err != nil {
			t.Fatalf("Append (re): %v", err)
		}
		// Nach dem Re-Append darf es keine Duplikate geben: History über
		// alle drei liefert genau drei Samples.
		page, err := repo.HistoryByStream(ctx, driven.SrtHealthHistoryQuery{
			ProjectID: proj, StreamID: stream, Limit: 100,
		})
		if err != nil {
			t.Fatalf("HistoryByStream: %v", err)
		}
		if len(page.Items) != 3 {
			t.Errorf("nach Re-Append: %d Samples, want 3 (Dedupe verletzt)", len(page.Items))
		}
	})

	t.Run("latest by stream", func(t *testing.T) {
		latest, err := repo.LatestByStream(ctx, proj)
		if err != nil {
			t.Fatalf("LatestByStream: %v", err)
		}
		found := false
		for _, s := range latest {
			if s.StreamID == stream {
				found = true
				if !s.IngestedAt.Equal(base.Add(2 * time.Second)) {
					t.Errorf("jüngster IngestedAt = %v, want %v", s.IngestedAt, base.Add(2*time.Second))
				}
			}
		}
		if !found {
			t.Errorf("Stream %q nicht in LatestByStream", stream)
		}
	})

	t.Run("history keyset pagination", func(t *testing.T) {
		first, err := repo.HistoryByStream(ctx, driven.SrtHealthHistoryQuery{
			ProjectID: proj, StreamID: stream, Limit: 2,
		})
		if err != nil {
			t.Fatalf("HistoryByStream page 1: %v", err)
		}
		if len(first.Items) != 2 || first.NextAfter == nil {
			t.Fatalf("page 1: %d Items, NextAfter=%v; want 2 Items + NextAfter", len(first.Items), first.NextAfter)
		}
		// DESC nach ingested_at: Seite 1 = seq-3, seq-2.
		if !first.Items[0].IngestedAt.Equal(base.Add(2 * time.Second)) {
			t.Errorf("page1[0] ingested_at = %v, want %v", first.Items[0].IngestedAt, base.Add(2*time.Second))
		}

		second, err := repo.HistoryByStream(ctx, driven.SrtHealthHistoryQuery{
			ProjectID: proj, StreamID: stream, Limit: 2, After: first.NextAfter,
		})
		if err != nil {
			t.Fatalf("HistoryByStream page 2: %v", err)
		}
		if len(second.Items) != 1 || second.NextAfter != nil {
			t.Fatalf("page 2: %d Items, NextAfter=%v; want 1 Item + kein NextAfter", len(second.Items), second.NextAfter)
		}
		if !second.Items[0].IngestedAt.Equal(base) {
			t.Errorf("page2[0] ingested_at = %v, want %v", second.Items[0].IngestedAt, base)
		}
	})
}

func sample(proj, stream, conn, seq string, ingestedAt time.Time) domain.SrtHealthSample {
	return domain.SrtHealthSample{
		ProjectID:             proj,
		StreamID:              stream,
		ConnectionID:          conn,
		SourceSequence:        seq,
		CollectedAt:           ingestedAt,
		IngestedAt:            ingestedAt,
		RTTMillis:             12.5,
		PacketLossTotal:       0,
		RetransmissionsTotal:  0,
		AvailableBandwidthBPS: 1_000_000,
		SourceStatus:          domain.SourceStatus("ok"),
		SourceErrorCode:       domain.SourceErrorCode("none"),
		ConnectionState:       domain.ConnectionState("connected"),
		HealthState:           domain.HealthState("healthy"),
	}
}
