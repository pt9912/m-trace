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
// daraus, bis `ctx` abgebrochen oder `Unsubscribe` gerufen wird —
// `Subscribe` registriert einen Cleanup über den Context, sobald
// `<-ctx.Done()` feuert. Buffer-Größe 64: bei langsamen Konsumenten
// werden ältere Frames gedroppt (analog zum SSE-Protokoll, das den
// Konsumenten via `Last-Event-ID`-Reconnect zur Lücken-Schließung
// auffordert).
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
		if sub, ok := b.subscribers[id]; ok {
			close(sub.ch)
			delete(b.subscribers, id)
		}
		b.mu.Unlock()
	}()
	return ch
}

// Publish dispatched ein Append-Event an alle Subscriber des
// referenzierten Projects. Slow Subscribers droppen den Frame
// (Buffer voll → Default-Branch); SSE-Reconnect mit `Last-Event-ID`
// schließt die Lücke. Aufruf ist non-blocking.
func (b *EventBroker) Publish(events []domain.PlaybackEvent) {
	if len(events) == 0 {
		return
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, e := range events {
		frame := EventAppendedFrame{
			ProjectID:      e.ProjectID,
			SessionID:      e.SessionID,
			EventName:      e.EventName,
			IngestSequence: e.IngestSequence,
		}
		for _, sub := range b.subscribers {
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

// SubscriberCount ist ein Test-Hook und gibt die aktuelle Anzahl
// aktiver Subscriber zurück.
func (b *EventBroker) SubscriberCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.subscribers)
}
