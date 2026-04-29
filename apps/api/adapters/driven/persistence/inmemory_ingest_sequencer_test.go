package persistence_test

import (
	"sync"
	"testing"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/persistence"
)

// TestInMemoryIngestSequencer_StartsAtOne verifies that the first
// Next() call returns 1 (plan-0.1.0.md §5.1: erste ingest_sequence ist 1).
func TestInMemoryIngestSequencer_StartsAtOne(t *testing.T) {
	t.Parallel()
	s := persistence.NewInMemoryIngestSequencer()
	if got := s.Next(); got != 1 {
		t.Errorf("first Next()=%d want 1", got)
	}
	if got := s.Next(); got != 2 {
		t.Errorf("second Next()=%d want 2", got)
	}
}

// TestInMemoryIngestSequencer_ConcurrentMonotonic verifies that under
// concurrent calls Next() returns 1..N exactly once each (no
// duplicates, no gaps). Required because cursor pagination relies on
// global uniqueness within a process (plan-0.1.0.md §5.1).
func TestInMemoryIngestSequencer_ConcurrentMonotonic(t *testing.T) {
	t.Parallel()
	s := persistence.NewInMemoryIngestSequencer()

	const workers = 8
	const perWorker = 250
	total := workers * perWorker

	out := make(chan int64, total)
	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func() {
			defer wg.Done()
			for i := 0; i < perWorker; i++ {
				out <- s.Next()
			}
		}()
	}
	wg.Wait()
	close(out)

	seen := make(map[int64]struct{}, total)
	var max int64
	for v := range out {
		if _, dup := seen[v]; dup {
			t.Fatalf("duplicate sequence value %d", v)
		}
		seen[v] = struct{}{}
		if v > max {
			max = v
		}
	}
	if max != int64(total) {
		t.Errorf("max=%d want %d (no gaps means max equals count)", max, total)
	}
	if len(seen) != total {
		t.Errorf("len(seen)=%d want %d", len(seen), total)
	}
}
