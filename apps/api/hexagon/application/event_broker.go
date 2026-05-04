package application

import (
	"context"
	"sync"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// EventAppendedFrame ist das Mindest-Wire-Format für SSE-Live-Updates
// (plan-0.4.0 §5 H4; spec/backend-api-contract.md §10a). Konsumenten
// laden den vollen Event-/Session-Read-Shape per REST nach.
type EventAppendedFrame struct {
	ProjectID      string
	SessionID      string
	EventName      string
	IngestSequence int64
}

// EventBroker ist der in-process Pub/Sub-Hub für SSE-Live-Updates.
// Der Append-Use-Case ruft `Publish` nach erfolgreichem
// `EventRepository.Append`, der SSE-Handler ruft `Subscribe` für die
// Lebenszeit einer Connection.
//
// Skalierung: in-process reicht für `0.4.0` (single-instance API,
// siehe ADR-0002). Multi-Instance-Fanout ist 0.6.0+-Folgepunkt.
type EventBroker struct {
	mu          sync.RWMutex
	subscribers map[int]subscriber
	nextID      int
}

type subscriber struct {
	projectID string
	ch        chan EventAppendedFrame
}

// NewEventBroker konstruiert einen leeren Broker.
func NewEventBroker() *EventBroker {
	return &EventBroker{subscribers: make(map[int]subscriber)}
}

// Subscribe öffnet einen project-skopierten Channel. Der Caller liest
// daraus, bis `ctx` abgebrochen wird. `Subscribe` registriert einen
// Cleanup über den Context, sobald `<-ctx.Done()` feuert. Der Cleanup
// entfernt den Subscriber aus dem Broker, schließt aber nicht den
// Channel: `Publish` sendet nach einem lockless Snapshot, daher könnte
// ein paralleles `close(ch)` sonst einen Send auf einen geschlossenen
// Channel auslösen. Buffer-Größe 64: bei langsamen Konsumenten werden
// ältere Frames gedroppt (analog zum SSE-Protokoll, das den Konsumenten
// via `Last-Event-ID`-Reconnect zur Lücken-Schließung auffordert).
func (b *EventBroker) Subscribe(ctx context.Context, projectID string) <-chan EventAppendedFrame {
	ch := make(chan EventAppendedFrame, 64)
	b.mu.Lock()
	id := b.nextID
	b.nextID++
	b.subscribers[id] = subscriber{projectID: projectID, ch: ch}
	b.mu.Unlock()

	go func() {
		<-ctx.Done()
		b.mu.Lock()
		delete(b.subscribers, id)
		b.mu.Unlock()
	}()
	return ch
}

// Publish dispatched ein Append-Event an alle Subscriber des
// referenzierten Projects. Slow Subscribers droppen den Frame
// (Buffer voll → Default-Branch); SSE-Reconnect mit `Last-Event-ID`
// schließt die Lücke. Aufruf ist non-blocking.
//
// Lock-Strategie: Subscriber-Liste wird unter `RLock` einmal in
// einen lokalen Slice kopiert; der eigentliche Channel-Send läuft
// danach lockless. So blockt ein neuer `Subscribe`-Caller nicht auf
// den ganzen Fanout, sondern nur auf den Snapshot-Schritt.
func (b *EventBroker) Publish(events []domain.PlaybackEvent) {
	if len(events) == 0 {
		return
	}
	subs := b.snapshotSubscribers()
	if len(subs) == 0 {
		return
	}
	for _, e := range events {
		frame := EventAppendedFrame{
			ProjectID:      e.ProjectID,
			SessionID:      e.SessionID,
			EventName:      e.EventName,
			IngestSequence: e.IngestSequence,
		}
		for _, sub := range subs {
			if sub.projectID != e.ProjectID {
				continue
			}
			select {
			case sub.ch <- frame:
			default:
				// Slow consumer: drop. Konsument schließt die Lücke
				// per Last-Event-ID-Reconnect.
			}
		}
	}
}

func (b *EventBroker) snapshotSubscribers() []subscriber {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if len(b.subscribers) == 0 {
		return nil
	}
	out := make([]subscriber, 0, len(b.subscribers))
	for _, sub := range b.subscribers {
		out = append(out, sub)
	}
	return out
}

// SubscriberCount ist ein Test-Hook und gibt die aktuelle Anzahl
// aktiver Subscriber zurück.
func (b *EventBroker) SubscriberCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.subscribers)
}
