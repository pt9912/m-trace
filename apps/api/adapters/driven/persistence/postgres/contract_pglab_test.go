package postgres_test

import (
	"context"
	"os"
	"testing"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/persistence/contract"
	"github.com/pt9912/m-trace/apps/api/adapters/driven/persistence/postgres"
	"github.com/pt9912/m-trace/apps/api/internal/storage"
)

// TestContract_PgLab führt die adapter-agnostische Persistence-Contract-
// Suite (Sessions/Events/Sequencer) gegen echte PG aus — dieselbe Suite,
// die InMemory und SQLite erfüllen (ADR-0006: der PG-Adapter muss denselben
// Port-Vertrag zeigen, damit das Wiring ohne Verhaltensbruch wechselt).
//
// Per-Test-Isolation über eine geteilte DB: `TRUNCATE … RESTART IDENTITY
// CASCADE` vor jedem Sub-Test setzt die App-Tabellen **und** die zugehörige
// BIGSERIAL-`ingest_sequence` zurück, sodass „sequencer is monotone and
// starts at one" auch im zweiten Lauf frisch bei 1 beginnt. Die Contract-
// Sub-Tests laufen seriell (RunAll nutzt kein t.Parallel), daher ist der
// TRUNCATE-Reset über die geteilte Verbindung rennfrei. Gated über
// MTRACE_PG_LAB_DSN.
func TestContract_PgLab(t *testing.T) {
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

	contract.RunAll(t, func(t *testing.T) contract.Repos {
		if _, err := db.ExecContext(ctx,
			`TRUNCATE stream_session_boundaries, stream_sessions, playback_events, projects RESTART IDENTITY CASCADE`,
		); err != nil {
			t.Fatalf("reset PG store: %v", err)
		}
		// Kleiner Block, damit der Sequencer bei jedem Sub-Test einen frischen
		// Block aus der zurückgesetzten Sequence zieht (erster Next() == 1).
		seq, err := postgres.NewIngestSequencer(ctx, db, 8)
		if err != nil {
			t.Fatalf("NewIngestSequencer: %v", err)
		}
		return contract.Repos{
			Sessions:  postgres.NewSessionRepository(db),
			Events:    postgres.NewEventRepository(db),
			Sequencer: seq,
		}
	})
}
