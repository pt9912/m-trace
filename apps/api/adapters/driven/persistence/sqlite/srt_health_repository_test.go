package sqlite_test

import (
	"context"
	"path/filepath"
	"strconv"
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

// TestSrtHealth_HistoryCursorWalksAllPages (plan-0.12.6 Tranche 2):
// Adapter paginiert eine größere Sample-Menge konsistent mit
// (ingested_at desc, id desc); jede Page hat einen NextAfter-Cursor
// außer der letzten; alle Samples kommen exakt einmal vor und der
// Cursor wandert ohne Lücken.
func TestSrtHealth_HistoryCursorWalksAllPages(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := freshSrtHealthDB(t)

	const total = 1500
	t0 := time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC)

	// 1500 Samples mit unique source_sequence schreiben.
	batch := make([]domain.SrtHealthSample, 0, total)
	for i := range total {
		batch = append(batch, mkSample(
			"srt-test", "c1",
			t0.Add(time.Duration(i)*time.Second),
			"seq-"+strconv.Itoa(i),
			nil, nil,
		))
	}
	if err := repo.Append(ctx, batch); err != nil {
		t.Fatalf("append: %v", err)
	}

	const pageLimit = 400
	seen := make(map[string]bool, total)
	var (
		cursor *driven.SrtHealthCursor
		pages  int
	)
	for {
		pages++
		if pages > 10 {
			t.Fatalf("too many pages (>%d) — likely cursor loop", pages)
		}
		page, err := repo.HistoryByStream(ctx, driven.SrtHealthHistoryQuery{
			ProjectID: "demo", StreamID: "srt-test",
			Limit: pageLimit,
			After: cursor,
		})
		if err != nil {
			t.Fatalf("history page %d: %v", pages, err)
		}
		if len(page.Items) == 0 {
			t.Fatalf("page %d unexpectedly empty", pages)
		}
		for i, s := range page.Items {
			if seen[s.SourceSequence] {
				t.Fatalf("page %d item %d: duplicate seq %s", pages, i, s.SourceSequence)
			}
			seen[s.SourceSequence] = true
		}
		// Sort-Invariante (desc): jede nachfolgende Page-Item darf nicht
		// jünger sein als das vorhergehende.
		for i := 1; i < len(page.Items); i++ {
			if page.Items[i].IngestedAt.After(page.Items[i-1].IngestedAt) {
				t.Fatalf("page %d: items[%d] (%v) is younger than items[%d] (%v) — sort broken",
					pages, i, page.Items[i].IngestedAt, i-1, page.Items[i-1].IngestedAt)
			}
		}
		if page.NextAfter == nil {
			// Letzte Seite.
			if len(page.Items) > pageLimit {
				t.Fatalf("final page has %d items, expected ≤ %d", len(page.Items), pageLimit)
			}
			break
		}
		if len(page.Items) != pageLimit {
			t.Fatalf("intermediate page %d has %d items, expected exactly %d (NextAfter set implies full page)",
				pages, len(page.Items), pageLimit)
		}
		cursor = page.NextAfter
	}
	if len(seen) != total {
		t.Fatalf("expected %d unique samples across pages, got %d (pages=%d)", total, len(seen), pages)
	}
	// 1500 / 400 = 3 volle Pages + 1 Rest-Page = 4 Pages.
	if pages != 4 {
		t.Errorf("expected 4 pages (limit=%d, total=%d), got %d", pageLimit, total, pages)
	}
}

// TestSrtHealth_HistoryCursorScopeIsolation: ein Cursor in
// (project=demo, stream=srt-test) darf weder Samples aus einer
// anderen Stream noch aus einem anderen Project liefern. Die
// Stream-Isolation wird im Adapter durch die WHERE-Klausel
// `project_id = ? AND stream_id = ?` garantiert; der Cursor selbst
// kennt diese Werte gar nicht (Wire-Codec-Verantwortung).
func TestSrtHealth_HistoryCursorScopeIsolation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := freshSrtHealthDB(t)
	t0 := time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC)

	if err := repo.Append(ctx, []domain.SrtHealthSample{
		mkSample("stream-a", "c1", t0, "a-1", nil, nil),
		mkSample("stream-a", "c1", t0.Add(2*time.Second), "a-2", nil, nil),
		mkSample("stream-b", "c1", t0.Add(1*time.Second), "b-1", nil, nil),
		mkSample("stream-b", "c1", t0.Add(3*time.Second), "b-2", nil, nil),
	}); err != nil {
		t.Fatalf("append: %v", err)
	}

	// Cursor an einer „mittleren" Position für stream-a:
	first, err := repo.HistoryByStream(ctx, driven.SrtHealthHistoryQuery{
		ProjectID: "demo", StreamID: "stream-a", Limit: 1,
	})
	if err != nil {
		t.Fatalf("history a page 1: %v", err)
	}
	if first.NextAfter == nil {
		t.Fatal("expected NextAfter after first page of stream-a")
	}

	// Cursor wird in einer stream-b-Query verwendet (was der HTTP-
	// Codec vor §10.3-Scope-Check verhindert). Der Adapter selbst
	// liefert dann nur stream-b-Samples, weil die WHERE-Klausel
	// stream-b filtert — ein Indikator, dass der Adapter keine
	// Cross-Stream-Bleed-Through erzeugt.
	bPage, err := repo.HistoryByStream(ctx, driven.SrtHealthHistoryQuery{
		ProjectID: "demo", StreamID: "stream-b", Limit: 100,
		After: first.NextAfter,
	})
	if err != nil {
		t.Fatalf("history b with stream-a cursor: %v", err)
	}
	for _, s := range bPage.Items {
		if s.StreamID != "stream-b" {
			t.Errorf("cross-stream bleed: stream-b query returned stream_id=%s", s.StreamID)
		}
	}
}

