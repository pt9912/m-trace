package inmemory_test

import (
	"testing"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/persistence/contract"
	"github.com/pt9912/m-trace/apps/api/adapters/driven/persistence/inmemory"
)

// TestContract verifiziert, dass die InMemory-Adapter den
// gemeinsamen Persistence-Vertrag aus
// `apps/api/adapters/driven/persistence/contract` erfüllen.
// Spiegelnde Suite läuft in `sqlite_test`.
func TestContract(t *testing.T) {
	t.Parallel()
	contract.RunAll(t, func(_ *testing.T) contract.Repos {
		return contract.Repos{
			Sessions:  inmemory.NewSessionRepository(),
			Events:    inmemory.NewEventRepository(),
			Sequencer: inmemory.NewIngestSequencer(),
		}
	})
}
