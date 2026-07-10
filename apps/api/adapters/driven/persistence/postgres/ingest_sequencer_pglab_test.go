package postgres_test

import (
	"context"
	"database/sql"
	"os"
	"sync"
	"testing"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/persistence/postgres"
	"github.com/pt9912/m-trace/apps/api/internal/storage"
)

// TestIngestSequencer_PgLab ist der PG-Lab-Integrationstest für den
// DB-autoritativen Postgres-Sequencer (R-28). Übersprungen ohne
// MTRACE_PG_LAB_DSN (siehe scripts/smoke-pg-lab.sh).
func TestIngestSequencer_PgLab(t *testing.T) {
	dsn := os.Getenv("MTRACE_PG_LAB_DSN")
	if dsn == "" {
		t.Skip("MTRACE_PG_LAB_DSN nicht gesetzt — PG-Lab-Integrationstest übersprungen")
	}
	ctx := context.Background()

	// Schema anwenden, damit die BIGSERIAL-Sequence existiert.
	db, err := storage.OpenPostgres(ctx, dsn)
	if err != nil {
		t.Fatalf("OpenPostgres: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	t.Run("monotonic and unique across block boundary", func(t *testing.T) {
		assertMonotonicUnique(ctx, t, db)
	})
	t.Run("R-28: no duplicates across N replicas", func(t *testing.T) {
		assertNoDupAcrossReplicas(ctx, t, db)
	})
}

// assertMonotonicUnique: ein einzelner Sequencer liefert über mehrere
// Block-Grenzen hinweg strikt aufsteigende, eindeutige Werte. Kleiner
// blockSize (4) erzwingt mehrere Reserve-Roundtrips.
func assertMonotonicUnique(ctx context.Context, t *testing.T, db *sql.DB) {
	t.Helper()
	seq, err := postgres.NewIngestSequencer(ctx, db, 4)
	if err != nil {
		t.Fatalf("NewIngestSequencer: %v", err)
	}
	const n = 20
	seen := make(map[int64]bool, n)
	var prev int64
	for i := 0; i < n; i++ {
		v := seq.Next()
		if seen[v] {
			t.Fatalf("Next() lieferte Duplikat %d (Iteration %d)", v, i)
		}
		seen[v] = true
		if i > 0 && v <= prev {
			t.Fatalf("Next() nicht strikt aufsteigend: %d nach %d (Iteration %d)", v, prev, i)
		}
		prev = v
	}
}

// assertNoDupAcrossReplicas: der R-28-Kern-Check. N unabhängige
// Sequencer-Instanzen (= N Replicas, je eigener In-Memory-Block-Puffer)
// ziehen gleichzeitig je M Werte; über alle N*M Werte darf es KEIN
// Duplikat geben — das ist genau die Eigenschaft, die der In-Process-
// `MAX`+`atomic.Add`-Sequencer über Replicas verletzt hätte (jede
// Instanz startete beim selben MAX). Der `nextval` ist DB-atomar, daher
// erhalten die Instanzen disjunkte Blöcke — schon über einen geteilten
// Pool, denn die Atomizität sitzt in der DB, nicht in der Verbindung.
func assertNoDupAcrossReplicas(ctx context.Context, t *testing.T, db *sql.DB) {
	t.Helper()
	const replicas = 4
	const perReplica = 60

	results := make([][]int64, replicas)
	errs := make([]error, replicas)
	var wg sync.WaitGroup
	for r := 0; r < replicas; r++ {
		wg.Add(1)
		go func(r int) {
			defer wg.Done()
			seq, err := postgres.NewIngestSequencer(ctx, db, 16)
			if err != nil {
				errs[r] = err
				return
			}
			vals := make([]int64, perReplica)
			for i := range vals {
				vals[i] = seq.Next()
			}
			results[r] = vals
		}(r)
	}
	wg.Wait()

	seen := make(map[int64]bool, replicas*perReplica)
	total := 0
	for r := 0; r < replicas; r++ {
		if errs[r] != nil {
			t.Fatalf("Replica %d: %v", r, errs[r])
		}
		for _, v := range results[r] {
			if seen[v] {
				t.Errorf("Duplikat %d über Replicas (Replica %d) — R-28 verletzt", v, r)
			}
			seen[v] = true
			total++
		}
	}
	if len(seen) != total {
		t.Errorf("eindeutige Werte %d != gezogene %d", len(seen), total)
	}
	if total != replicas*perReplica {
		t.Errorf("gezogene Werte %d != erwartete %d", total, replicas*perReplica)
	}
}
