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

// TestEventRepository_PgLab deckt den Postgres-event-Adapter gegen echte
// PG ab: Append + Dedup-Klassifikation (§8.3), ListBySession-Keyset-
// Pagination (kanonische Sortierung), ListAfterIngestSequence-Backfill.
// Gated über MTRACE_PG_LAB_DSN. Nutzt hohe, eindeutige ingest_sequence-
// Werte, um von den anderen PG-Lab-Tests isoliert zu sein.
func TestEventRepository_PgLab(t *testing.T) {
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

	if _, err := db.ExecContext(ctx,
		"INSERT INTO projects(project_id) VALUES ($1) ON CONFLICT(project_id) DO NOTHING", "ev-lab-proj"); err != nil {
		t.Fatalf("seed project: %v", err)
	}

	repo := postgres.NewEventRepository(db)
	const proj, sess = "ev-lab-proj", "ev-lab-sess"
	base := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)

	events := []domain.PlaybackEvent{
		event(proj, sess, 100_001, 1, base),
		event(proj, sess, 100_002, 2, base.Add(1*time.Second)),
		event(proj, sess, 100_003, 3, base.Add(2*time.Second)),
	}
	if err := repo.Append(ctx, events); err != nil {
		t.Fatalf("Append: %v", err)
	}

	t.Run("list by session keyset pagination", func(t *testing.T) {
		first, err := repo.ListBySession(ctx, driven.EventListQuery{ProjectID: proj, SessionID: sess, Limit: 2})
		if err != nil {
			t.Fatalf("ListBySession page 1: %v", err)
		}
		if len(first.Events) != 2 || first.NextAfter == nil {
			t.Fatalf("page 1: %d Events, NextAfter=%v; want 2 + NextAfter", len(first.Events), first.NextAfter)
		}
		if first.Events[0].IngestSequence != 100_001 || first.Events[1].IngestSequence != 100_002 {
			t.Errorf("page 1 Reihenfolge = %d,%d; want 100001,100002", first.Events[0].IngestSequence, first.Events[1].IngestSequence)
		}
		second, err := repo.ListBySession(ctx, driven.EventListQuery{ProjectID: proj, SessionID: sess, Limit: 2, After: first.NextAfter})
		if err != nil {
			t.Fatalf("ListBySession page 2: %v", err)
		}
		if len(second.Events) != 1 || second.NextAfter != nil {
			t.Fatalf("page 2: %d Events, NextAfter=%v; want 1 + kein NextAfter", len(second.Events), second.NextAfter)
		}
		if second.Events[0].IngestSequence != 100_003 {
			t.Errorf("page 2 = %d; want 100003", second.Events[0].IngestSequence)
		}
	})

	t.Run("list after ingest sequence (backfill)", func(t *testing.T) {
		out, err := repo.ListAfterIngestSequence(ctx, proj, 100_001, 10)
		if err != nil {
			t.Fatalf("ListAfterIngestSequence: %v", err)
		}
		if len(out) != 2 || out[0].IngestSequence != 100_002 || out[1].IngestSequence != 100_003 {
			t.Errorf("backfill = %d Events; want [100002,100003]", len(out))
		}
	})

	t.Run("dedup classifies re-sent sequence as duplicate_suspected", func(t *testing.T) {
		dup := event(proj, sess, 100_004, 1, base.Add(3*time.Second)) // gleiche sequence_number 1
		if err := repo.Append(ctx, []domain.PlaybackEvent{dup}); err != nil {
			t.Fatalf("Append dup: %v", err)
		}
		var status string
		if err := db.QueryRowContext(ctx,
			"SELECT delivery_status FROM playback_events WHERE ingest_sequence = $1", int64(100_004)).Scan(&status); err != nil {
			t.Fatalf("read delivery_status: %v", err)
		}
		if status != "duplicate_suspected" {
			t.Errorf("delivery_status = %q, want duplicate_suspected", status)
		}
	})
}

func event(proj, sess string, ingestSeq, seqNum int64, serverAt time.Time) domain.PlaybackEvent {
	sn := seqNum
	return domain.PlaybackEvent{
		EventName:        "rebuffer_started",
		ProjectID:        proj,
		SessionID:        sess,
		ClientTimestamp:  serverAt,
		ServerReceivedAt: serverAt,
		IngestSequence:   ingestSeq,
		SequenceNumber:   &sn,
		SDK:              domain.SDKInfo{Name: "@pt9912/player-sdk", Version: "0.4.0"},
	}
}
