package persistence

import (
	"sync/atomic"

	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// InMemoryIngestSequencer hält einen atomaren Counter pro Prozess und
// erfüllt damit den driven.IngestSequencer-Vertrag aus plan-0.1.0.md
// §5.1. Erste Next()-Rückgabe ist 1.
type InMemoryIngestSequencer struct {
	counter atomic.Int64
}

// NewInMemoryIngestSequencer gibt einen einsatzbereiten Sequencer
// zurück, dessen erster Next()-Rückgabewert 1 ist.
func NewInMemoryIngestSequencer() *InMemoryIngestSequencer {
	return &InMemoryIngestSequencer{}
}

// Next erhöht den internen Counter um 1 und gibt den neuen Wert
// zurück. Safe für nebenläufige Aufrufe.
func (s *InMemoryIngestSequencer) Next() int64 {
	return s.counter.Add(1)
}

var _ driven.IngestSequencer = (*InMemoryIngestSequencer)(nil)
