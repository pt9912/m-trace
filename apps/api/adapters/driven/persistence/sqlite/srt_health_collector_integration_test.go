package sqlite_test

import (
	"context"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/persistence/sqlite"
	"github.com/pt9912/m-trace/apps/api/hexagon/application"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
	"github.com/pt9912/m-trace/apps/api/internal/storage"
)

// integrationSrtSource liefert pro Aufruf einen Sample mit
// fortschreitender SourceSequence, damit der Collector zwei
// aufeinanderfolgende Persistierungen sieht (DoD plan-0.6.0 §4.3
// für Sub-3.7: „Smoke- oder Integrationstest weist nach, dass der
// Collector im Lab mindestens zwei aufeinanderfolgende Samples
// importiert und persistiert").
type integrationSrtSource struct {
	calls atomic.Int64
}

func (s *integrationSrtSource) SnapshotConnections(_ context.Context) ([]domain.SrtConnectionSample, error) {
	n := s.calls.Add(1)
	return []domain.SrtConnectionSample{{
		StreamID:              "srt-test",
		ConnectionID:          "c1",
		ConnectionState:       domain.ConnectionStateConnected,
		RTTMillis:             0.5,
		PacketLossTotal:       0,
		RetransmissionsTotal:  0,
		AvailableBandwidthBPS: 4_000_000_000,
		SourceSequence:        "seq-" + strconv.FormatInt(n, 10),
	}}, nil
}

var _ driven.SrtSource = (*integrationSrtSource)(nil)

// TestSrtHealthIntegration_TwoConsecutiveSamplesPersisted verdrahtet
// echten SQLite-Storage mit einem Mock-Source und der Collector-Run-
// Loop. Nach dem Lauf müssen zwei verschiedene Samples in der DB
// stehen (steigender SourceSequence + getrennte Rows wegen
// unterschiedlicher Dedupe-Keys).
func TestSrtHealthIntegration_TwoConsecutiveSamplesPersisted(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "m-trace.db")

	db, err := storage.Open(ctx, path)
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	source := &integrationSrtSource{}
	repo := sqlite.NewSrtHealthRepository(db)

	collector, err := application.NewSrtHealthCollector(
		source, repo, "demo", time.Now, application.DefaultThresholds(),
	)
	if err != nil {
		t.Fatalf("NewSrtHealthCollector: %v", err)
	}
	collector.WithPollInterval(2 * time.Millisecond).WithMaxBackoff(10 * time.Millisecond)

	runCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() {
		collector.Run(runCtx)
		close(done)
	}()

	deadline := time.NewTimer(2 * time.Second)
	defer deadline.Stop()
	for {
		hist, err := repo.HistoryByStream(ctx, driven.SrtHealthHistoryQuery{
			ProjectID: "demo", StreamID: "srt-test", Limit: 10,
		})
		if err == nil && len(hist.Items) >= 2 {
			break
		}
		select {
		case <-deadline.C:
			cancel()
			<-done
			items := 0
			if hist, err := repo.HistoryByStream(ctx, driven.SrtHealthHistoryQuery{
				ProjectID: "demo", StreamID: "srt-test", Limit: 10,
			}); err == nil {
				items = len(hist.Items)
			}
			t.Fatalf("expected ≥2 persisted samples within deadline, got %d (source calls=%d)", items, source.calls.Load())
		case <-time.After(5 * time.Millisecond):
		}
	}
	cancel()
	<-done

	hist, err := repo.HistoryByStream(ctx, driven.SrtHealthHistoryQuery{
		ProjectID: "demo", StreamID: "srt-test", Limit: 10,
	})
	if err != nil {
		t.Fatalf("history: %v", err)
	}
	if len(hist.Items) < 2 {
		t.Fatalf("expected ≥2 persisted samples after run, got %d", len(hist.Items))
	}
	// HistoryByStream sortiert nach IngestedAt desc — die ersten
	// beiden Einträge müssen unterschiedliche SourceSequence haben.
	if hist.Items[0].SourceSequence == hist.Items[1].SourceSequence {
		t.Fatalf("expected progressing source sequence; got %q == %q", hist.Items[0].SourceSequence, hist.Items[1].SourceSequence)
	}
	// Persistierte Samples kommen mit Healthy-Status (alle Pflicht-
	// werte gesetzt; keine required_bandwidth-Schwelle → kein
	// degraded/critical).
	for _, s := range hist.Items {
		if s.HealthState != domain.HealthStateHealthy {
			t.Errorf("expected healthy sample, got %s for seq=%s", s.HealthState, s.SourceSequence)
		}
		if s.ConnectionState != domain.ConnectionStateConnected {
			t.Errorf("expected connected, got %s", s.ConnectionState)
		}
	}
}
