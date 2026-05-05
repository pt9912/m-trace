package sqlite_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/persistence/sqlite"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
	"github.com/pt9912/m-trace/apps/api/internal/storage"
)

// Compile-Time-Check: SQLite-Adapter erfüllt den Driven-Port.
var _ driven.SrtHealthRepository = (*sqlite.SrtHealthRepository)(nil)

// freshSrtHealthDB öffnet eine frische SQLite-Datei mit allen
// Migrationen (inkl. V5 srt_health_samples) und liefert das
// Repository.
func freshSrtHealthDB(t *testing.T) *sqlite.SrtHealthRepository {
	t.Helper()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "m-trace.db")
	db, err := storage.Open(ctx, path)
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return sqlite.NewSrtHealthRepository(db)
}

func mkSample(streamID, connID string, ingested time.Time, sequence string, lossRate *float64, throughput *int64) domain.SrtHealthSample {
	return domain.SrtHealthSample{
		ProjectID:             "demo",
		StreamID:              streamID,
		ConnectionID:          connID,
		SourceObservedAt:      time.Time{}, // MediaMTX-API liefert keinen
		SourceSequence:        sequence,
		CollectedAt:           ingested.Add(-100 * time.Millisecond),
		IngestedAt:            ingested,
		RTTMillis:             0.5,
		PacketLossTotal:       0,
		PacketLossRate:        lossRate,
		RetransmissionsTotal:  0,
		AvailableBandwidthBPS: 4_000_000_000,
		ThroughputBPS:         throughput,
		RequiredBandwidthBPS:  nil,
		SampleWindowMillis:    nil,
		SourceStatus:          domain.SourceStatusOK,
		SourceErrorCode:       domain.SourceErrorCodeNone,
		ConnectionState:       domain.ConnectionStateConnected,
		HealthState:           domain.HealthStateHealthy,
	}
}

// TestSrtHealth_AppendAndLatest persistiert zwei Samples für denselben
// Stream und prüft, dass LatestByStream den jüngsten zurückliefert.
func TestSrtHealth_AppendAndLatest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := freshSrtHealthDB(t)
	t0 := time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC)

	if err := repo.Append(ctx, []domain.SrtHealthSample{
		mkSample("srt-test", "c1", t0, "seq-1", nil, nil),
		mkSample("srt-test", "c1", t0.Add(5*time.Second), "seq-2", nil, nil),
	}); err != nil {
		t.Fatalf("append: %v", err)
	}

	got, err := repo.LatestByStream(ctx, "demo")
	if err != nil {
		t.Fatalf("latest: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 latest sample, got %d", len(got))
	}
	if got[0].StreamID != "srt-test" || got[0].SourceSequence != "seq-2" {
		t.Fatalf("expected latest with seq-2, got %+v", got[0])
	}
	if !got[0].IngestedAt.Equal(t0.Add(5 * time.Second)) {
		t.Fatalf("expected ingested=%v, got %v", t0.Add(5*time.Second), got[0].IngestedAt)
	}
}

// TestSrtHealth_DedupeSkipsIdenticalKey: zweimaliges Append desselben
// Dedupe-Keys (gleicher source_sequence) führt zu einem einzigen Row.
func TestSrtHealth_DedupeSkipsIdenticalKey(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := freshSrtHealthDB(t)
	t0 := time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC)

	s := mkSample("srt-test", "c1", t0, "seq-frozen", nil, nil)
	if err := repo.Append(ctx, []domain.SrtHealthSample{s}); err != nil {
		t.Fatalf("append #1: %v", err)
	}
	// Zweiter Append mit anderem IngestedAt aber gleichem Dedupe-Key.
	dup := s
	dup.IngestedAt = t0.Add(10 * time.Second)
	dup.CollectedAt = t0.Add(9 * time.Second)
	if err := repo.Append(ctx, []domain.SrtHealthSample{dup}); err != nil {
		t.Fatalf("append #2: %v", err)
	}

	got, err := repo.HistoryByStream(ctx, driven.SrtHealthHistoryQuery{
		ProjectID: "demo", StreamID: "srt-test",
	})
	if err != nil {
		t.Fatalf("history: %v", err)
	}
	if len(got.Items) != 1 {
		t.Fatalf("expected 1 row after dedupe, got %d", len(got.Items))
	}
	// Original-Sample bleibt; Dup wird verworfen.
	if !got.Items[0].IngestedAt.Equal(t0) {
		t.Fatalf("expected original IngestedAt, got %v", got.Items[0].IngestedAt)
	}
}

// TestSrtHealth_LatestByStreamMultipleStreams: Latest liefert pro
// StreamID den jüngsten Sample.
func TestSrtHealth_LatestByStreamMultipleStreams(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := freshSrtHealthDB(t)
	t0 := time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC)

	if err := repo.Append(ctx, []domain.SrtHealthSample{
		mkSample("stream-a", "c1", t0, "a-1", nil, nil),
		mkSample("stream-a", "c1", t0.Add(2*time.Second), "a-2", nil, nil),
		mkSample("stream-b", "c1", t0.Add(1*time.Second), "b-1", nil, nil),
	}); err != nil {
		t.Fatalf("append: %v", err)
	}

	got, err := repo.LatestByStream(ctx, "demo")
	if err != nil {
		t.Fatalf("latest: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 latest samples (stream-a, stream-b), got %d", len(got))
	}
	bySeq := map[string]string{}
	for _, s := range got {
		bySeq[s.StreamID] = s.SourceSequence
	}
	if bySeq["stream-a"] != "a-2" {
		t.Fatalf("stream-a latest expected a-2, got %s", bySeq["stream-a"])
	}
	if bySeq["stream-b"] != "b-1" {
		t.Fatalf("stream-b latest expected b-1, got %s", bySeq["stream-b"])
	}
}

// TestSrtHealth_HistoryOrderingDescending: HistoryByStream sortiert
// nach IngestedAt desc, ID desc.
func TestSrtHealth_HistoryOrderingDescending(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := freshSrtHealthDB(t)
	t0 := time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC)

	if err := repo.Append(ctx, []domain.SrtHealthSample{
		mkSample("srt-test", "c1", t0, "1", nil, nil),
		mkSample("srt-test", "c1", t0.Add(2*time.Second), "2", nil, nil),
		mkSample("srt-test", "c1", t0.Add(4*time.Second), "3", nil, nil),
	}); err != nil {
		t.Fatalf("append: %v", err)
	}

	got, err := repo.HistoryByStream(ctx, driven.SrtHealthHistoryQuery{
		ProjectID: "demo", StreamID: "srt-test",
	})
	if err != nil {
		t.Fatalf("history: %v", err)
	}
	if len(got.Items) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(got.Items))
	}
	wantSeqs := []string{"3", "2", "1"}
	for i, want := range wantSeqs {
		if got.Items[i].SourceSequence != want {
			t.Errorf("got.Items[%d].SourceSequence = %s, want %s", i, got.Items[i].SourceSequence, want)
		}
	}
}

// TestSrtHealth_RestartPreservesData verifiziert Restart-Stabilität:
// nach Close + Re-Open derselben SQLite-Datei sind Samples
// erhalten und LatestByStream/HistoryByStream funktionieren.
func TestSrtHealth_RestartPreservesData(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "m-trace.db")
	t0 := time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC)

	// Pass 1: zwei Samples schreiben.
	db1, err := storage.Open(ctx, path)
	if err != nil {
		t.Fatalf("open #1: %v", err)
	}
	repo1 := sqlite.NewSrtHealthRepository(db1)
	if err := repo1.Append(ctx, []domain.SrtHealthSample{
		mkSample("srt-test", "c1", t0, "seq-1", nil, nil),
		mkSample("srt-test", "c1", t0.Add(5*time.Second), "seq-2", nil, nil),
	}); err != nil {
		t.Fatalf("append #1: %v", err)
	}
	if err := db1.Close(); err != nil {
		t.Fatalf("close #1: %v", err)
	}

	// Pass 2: gleiche Datei öffnen, Lese-Pfade prüfen.
	db2, err := storage.Open(ctx, path)
	if err != nil {
		t.Fatalf("open #2: %v", err)
	}
	t.Cleanup(func() { _ = db2.Close() })
	repo2 := sqlite.NewSrtHealthRepository(db2)

	got, err := repo2.LatestByStream(ctx, "demo")
	if err != nil {
		t.Fatalf("latest after restart: %v", err)
	}
	if len(got) != 1 || got[0].SourceSequence != "seq-2" {
		t.Fatalf("expected restored latest seq-2, got %+v", got)
	}

	hist, err := repo2.HistoryByStream(ctx, driven.SrtHealthHistoryQuery{
		ProjectID: "demo", StreamID: "srt-test",
	})
	if err != nil {
		t.Fatalf("history after restart: %v", err)
	}
	if len(hist.Items) != 2 {
		t.Fatalf("expected 2 rows after restart, got %d", len(hist.Items))
	}
}

// TestSrtHealth_OptionalFieldsRoundTrip prüft, dass alle Nullable-
// Felder (PacketLossRate, ThroughputBPS, RequiredBandwidthBPS,
// SampleWindowMillis, SourceObservedAt) korrekt durch SQLite gehen.
func TestSrtHealth_OptionalFieldsRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := freshSrtHealthDB(t)
	t0 := time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC)
	observed := t0.Add(-200 * time.Millisecond)

	lossRate := 0.012
	throughput := int64(1_500_000)
	required := int64(2_000_000)
	window := int64(5_000)

	full := domain.SrtHealthSample{
		ProjectID:             "demo",
		StreamID:              "srt-test",
		ConnectionID:          "c1",
		SourceObservedAt:      observed,
		SourceSequence:        "seq-full",
		CollectedAt:           t0.Add(-100 * time.Millisecond),
		IngestedAt:            t0,
		RTTMillis:             1.234,
		PacketLossTotal:       42,
		PacketLossRate:        &lossRate,
		RetransmissionsTotal:  17,
		AvailableBandwidthBPS: 5_000_000_000,
		ThroughputBPS:         &throughput,
		RequiredBandwidthBPS:  &required,
		SampleWindowMillis:    &window,
		SourceStatus:          domain.SourceStatusOK,
		SourceErrorCode:       domain.SourceErrorCodeNone,
		ConnectionState:       domain.ConnectionStateConnected,
		HealthState:           domain.HealthStateHealthy,
	}
	if err := repo.Append(ctx, []domain.SrtHealthSample{full}); err != nil {
		t.Fatalf("append: %v", err)
	}

	got, err := repo.LatestByStream(ctx, "demo")
	if err != nil || len(got) != 1 {
		t.Fatalf("latest: items=%d err=%v", len(got), err)
	}
	r := got[0]
	if !r.SourceObservedAt.Equal(observed) {
		t.Errorf("SourceObservedAt = %v, want %v", r.SourceObservedAt, observed)
	}
	if r.PacketLossRate == nil || *r.PacketLossRate != lossRate {
		t.Errorf("PacketLossRate = %v, want %v", r.PacketLossRate, lossRate)
	}
	if r.ThroughputBPS == nil || *r.ThroughputBPS != throughput {
		t.Errorf("ThroughputBPS = %v, want %v", r.ThroughputBPS, throughput)
	}
	if r.RequiredBandwidthBPS == nil || *r.RequiredBandwidthBPS != required {
		t.Errorf("RequiredBandwidthBPS = %v, want %v", r.RequiredBandwidthBPS, required)
	}
	if r.SampleWindowMillis == nil || *r.SampleWindowMillis != window {
		t.Errorf("SampleWindowMillis = %v, want %v", r.SampleWindowMillis, window)
	}
}

// TestSrtHealth_HistoryCursorNotImplemented: Cursor-Pagination ist
// in 0.6.0 Sub-3.3 noch nicht implementiert; Adapter gibt einen
// expliziten Fehler zurück, statt stillschweigend nur die erste
// Seite zu liefern.
func TestSrtHealth_HistoryCursorNotImplemented(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := freshSrtHealthDB(t)
	_, err := repo.HistoryByStream(ctx, driven.SrtHealthHistoryQuery{
		ProjectID: "demo", StreamID: "srt-test",
		After: &driven.SrtHealthCursor{IngestedAt: 1, ID: 1, ProcessInstanceID: "x"},
	})
	if err == nil {
		t.Fatal("expected error for cursor pagination, got nil")
	}
}
