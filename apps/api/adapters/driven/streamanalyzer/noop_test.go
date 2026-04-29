package streamanalyzer_test

import (
	"context"
	"testing"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/streamanalyzer"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// TestNoopStreamAnalyzer_AnalyzeBatchReturnsNil verifies that the
// noop adapter satisfies the F-22 port without doing anything. The
// happy-path (zero-event slice) and the populated-batch case are both
// no-ops; both must return nil so that a use case wiring the slot
// stays unaffected (plan-0.1.0 §5.1 F-22).
func TestNoopStreamAnalyzer_AnalyzeBatchReturnsNil(t *testing.T) {
	t.Parallel()

	a := streamanalyzer.NewNoopStreamAnalyzer()

	if err := a.AnalyzeBatch(context.Background(), nil); err != nil {
		t.Errorf("nil batch: expected nil error, got %v", err)
	}

	events := []domain.PlaybackEvent{{EventName: "rebuffer_started"}}
	if err := a.AnalyzeBatch(context.Background(), events); err != nil {
		t.Errorf("populated batch: expected nil error, got %v", err)
	}
}
