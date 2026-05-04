package application

import (
	"context"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

func TestEventBroker_CancelledSubscriberDoesNotCloseSnapshotChannel(t *testing.T) {
	broker := NewEventBroker()
	ctx, cancel := context.WithCancel(context.Background())
	frames := broker.Subscribe(ctx, "demo")

	staleSnapshot := broker.snapshotSubscribers()
	if len(staleSnapshot) != 1 {
		t.Fatalf("snapshot subscribers = %d want 1", len(staleSnapshot))
	}

	cancel()
	if !waitForSubscriberCount(100*time.Millisecond, broker, 0) {
		t.Fatalf("subscriber not cleaned up, count=%d", broker.SubscriberCount())
	}

	select {
	case _, ok := <-frames:
		if !ok {
			t.Fatal("cancel cleanup closed subscriber channel")
		}
		t.Fatal("unexpected frame before publish")
	default:
	}

	broker.Publish([]domain.PlaybackEvent{{
		ProjectID:      "demo",
		SessionID:      "sess-1",
		EventName:      "manifest_loaded",
		IngestSequence: 1,
	}})
	select {
	case frame := <-frames:
		t.Fatalf("removed subscriber received published frame: %+v", frame)
	default:
	}

	var panicked any
	func() {
		defer func() {
			panicked = recover()
		}()
		select {
		case staleSnapshot[0].ch <- EventAppendedFrame{
			ProjectID:      "demo",
			SessionID:      "sess-1",
			EventName:      "manifest_loaded",
			IngestSequence: 1,
		}:
		default:
		}
	}()
	if panicked != nil {
		t.Fatalf("send to stale snapshot channel panicked: %v", panicked)
	}
}

func waitForSubscriberCount(timeout time.Duration, broker *EventBroker, want int) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if broker.SubscriberCount() == want {
			return true
		}
		time.Sleep(5 * time.Millisecond)
	}
	return broker.SubscriberCount() == want
}
