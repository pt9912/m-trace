package sqlite_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/persistence/contract"
	"github.com/pt9912/m-trace/apps/api/adapters/driven/persistence/sqlite"
	"github.com/pt9912/m-trace/apps/api/internal/storage"
)

// TestContract verifiziert, dass die SQLite-Adapter den gemeinsamen
// Persistence-Vertrag aus `persistence/contract` erfüllen — identisch
// zum InMemory-Test in `inmemory_test`. Jeder Sub-Test bekommt eine
// frische SQLite-Datei in einem t.TempDir().
func TestContract(t *testing.T) {
	t.Parallel()
	contract.RunAll(t, func(t *testing.T) contract.Repos {
		ctx := context.Background()
		path := filepath.Join(t.TempDir(), "m-trace.db")
		db, err := storage.Open(ctx, path)
		if err != nil {
			t.Fatalf("storage.Open: %v", err)
		}
		t.Cleanup(func() { _ = db.Close() })
		seq, err := sqlite.NewIngestSequencer(ctx, db)
		if err != nil {
			t.Fatalf("NewIngestSequencer: %v", err)
		}
		return contract.Repos{
			Sessions:  sqlite.NewSessionRepository(db),
			Events:    sqlite.NewEventRepository(db),
			Sequencer: seq,
		}
	})
}
