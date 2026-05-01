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

// TestNoopStreamAnalyzer_AnalyzeManifestReturnsEmptyResult sichert die
// 0.3.0-Tranche-1-Zielsignatur ab: der Slot-Adapter erfüllt die neue
// Methode ohne Seiteneffekt, AnalyzerVersion ist "noop" und
// PlaylistType ist PlaylistTypeUnknown — damit bleibt erkennbar, dass
// kein produktiver Adapter angeschlossen ist.
func TestNoopStreamAnalyzer_AnalyzeManifestReturnsEmptyResult(t *testing.T) {
	t.Parallel()

	a := streamanalyzer.NewNoopStreamAnalyzer()

	cases := []struct {
		name string
		req  domain.StreamAnalysisRequest
	}{
		{name: "empty request", req: domain.StreamAnalysisRequest{}},
		{name: "text request", req: domain.StreamAnalysisRequest{ManifestText: "#EXTM3U\n"}},
		{name: "url request", req: domain.StreamAnalysisRequest{ManifestURL: "https://example.invalid/manifest.m3u8"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := a.AnalyzeManifest(context.Background(), tc.req)
			if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
			if result.AnalyzerVersion != "noop" {
				t.Errorf("AnalyzerVersion: want %q, got %q", "noop", result.AnalyzerVersion)
			}
			if result.PlaylistType != domain.PlaylistTypeUnknown {
				t.Errorf("PlaylistType: want %q, got %q", domain.PlaylistTypeUnknown, result.PlaylistType)
			}
			if len(result.Findings) != 0 {
				t.Errorf("Findings: want empty, got %d", len(result.Findings))
			}
			if len(result.EncodedDetails) != 0 {
				t.Errorf("EncodedDetails: want empty, got %d bytes", len(result.EncodedDetails))
			}
		})
	}
}
