// Package persistence holds in-memory storage for the spike. Per
// docs/spike/0001-backend-stack.md §6.10 there is no on-disk persistence;
// data does not survive a restart, on purpose.
package persistence

import (
	"context"
	"sync"

	"github.com/example/m-trace/apps/api/hexagon/domain"
	"github.com/example/m-trace/apps/api/hexagon/port/driven"
)

// InMemoryEventRepository keeps accepted events in a slice. Safe for
// concurrent use; the spike scope does not require performance tuning.
type InMemoryEventRepository struct {
	mu     sync.Mutex
	events []domain.PlaybackEvent
}

// NewInMemoryEventRepository constructs an empty repository.
func NewInMemoryEventRepository() *InMemoryEventRepository {
	return &InMemoryEventRepository{}
}

// Append stores all events atomically.
func (r *InMemoryEventRepository) Append(_ context.Context, events []domain.PlaybackEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, events...)
	return nil
}

// Snapshot returns a copy of the stored events. Useful for tests.
func (r *InMemoryEventRepository) Snapshot() []domain.PlaybackEvent {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]domain.PlaybackEvent, len(r.events))
	copy(out, r.events)
	return out
}

var _ driven.EventRepository = (*InMemoryEventRepository)(nil)
