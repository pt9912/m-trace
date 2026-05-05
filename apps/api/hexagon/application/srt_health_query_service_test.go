package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/application"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// queryRepo mockt SrtHealthRepository minimal für die Query-Tests.
type queryRepo struct {
	latest    []domain.SrtHealthSample
	latestErr error

	history    []domain.SrtHealthSample
	historyErr error

	gotHistoryQuery driven.SrtHealthHistoryQuery
}

func (r *queryRepo) Append(_ context.Context, _ []domain.SrtHealthSample) error {
	return errors.New("not used")
}
func (r *queryRepo) LatestByStream(_ context.Context, _ string) ([]domain.SrtHealthSample, error) {
	return r.latest, r.latestErr
}
func (r *queryRepo) HistoryByStream(_ context.Context, q driven.SrtHealthHistoryQuery) (driven.SrtHealthHistoryPage, error) {
	r.gotHistoryQuery = q
	if r.historyErr != nil {
		return driven.SrtHealthHistoryPage{}, r.historyErr
	}
	return driven.SrtHealthHistoryPage{Items: r.history}, nil
}

var _ driven.SrtHealthRepository = (*queryRepo)(nil)

func newQuerySvc(t *testing.T, repo *queryRepo, fixed time.Time) *application.SrtHealthQueryService {
	t.Helper()
	svc, err := application.NewSrtHealthQueryService(repo, func() time.Time { return fixed }, application.DefaultThresholds())
	if err != nil {
		t.Fatalf("NewSrtHealthQueryService: %v", err)
	}
	return svc
}

func sampleAt(streamID string, ingested time.Time) domain.SrtHealthSample {
	return domain.SrtHealthSample{
		ProjectID:             "demo",
		StreamID:              streamID,
		ConnectionID:          "c1",
		CollectedAt:           ingested.Add(-100 * time.Millisecond),
		IngestedAt:            ingested,
		RTTMillis:             0.5,
		AvailableBandwidthBPS: 4_000_000_000,
		SourceStatus:          domain.SourceStatusOK,
		SourceErrorCode:       domain.SourceErrorCodeNone,
		ConnectionState:       domain.ConnectionStateConnected,
		HealthState:           domain.HealthStateHealthy,
	}
}

// LatestByStream: pro Stream ein Sample, abgeleitete Felder
// (sample_age_ms, stale_after_ms, headroom) berechnet.
func TestQueryService_LatestByStream(t *testing.T) {
	now := time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC)
	required := int64(2_000_000_000)
	sample := sampleAt("srt-test", now.Add(-2*time.Second))
	sample.RequiredBandwidthBPS = &required
	repo := &queryRepo{latest: []domain.SrtHealthSample{sample}}
	svc := newQuerySvc(t, repo, now)

	got, err := svc.LatestByStream(context.Background(), "demo")
	if err != nil {
		t.Fatalf("LatestByStream: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(got))
	}
	s := got[0]
	if s.SampleAgeMillis != 2000 {
		t.Errorf("SampleAgeMillis = %d, want 2000", s.SampleAgeMillis)
	}
	if s.StaleAfterMillis != application.DefaultThresholds().StaleAfterMillis {
		t.Errorf("StaleAfterMillis mismatch: %d vs threshold default", s.StaleAfterMillis)
	}
	if s.BandwidthHeadroom == nil || *s.BandwidthHeadroom != 2.0 {
		t.Errorf("BandwidthHeadroom = %v, want 2.0", s.BandwidthHeadroom)
	}
}

// LatestByStream: ohne RequiredBandwidth bleibt headroom nil.
func TestQueryService_LatestByStream_NoRequired(t *testing.T) {
	now := time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC)
	repo := &queryRepo{latest: []domain.SrtHealthSample{sampleAt("a", now)}}
	svc := newQuerySvc(t, repo, now)

	got, err := svc.LatestByStream(context.Background(), "demo")
	if err != nil || len(got) != 1 {
		t.Fatalf("LatestByStream: items=%d err=%v", len(got), err)
	}
	if got[0].BandwidthHeadroom != nil {
		t.Errorf("expected nil headroom without required, got %v", *got[0].BandwidthHeadroom)
	}
}

// LatestByStream: leere Liste bleibt leer (kein Stream → kein Fehler).
func TestQueryService_LatestByStream_EmptyOK(t *testing.T) {
	now := time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC)
	repo := &queryRepo{latest: nil}
	svc := newQuerySvc(t, repo, now)

	got, err := svc.LatestByStream(context.Background(), "demo")
	if err != nil {
		t.Fatalf("LatestByStream: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected 0 summaries, got %d", len(got))
	}
}

// LatestByStream: leere ProjectID → Validierungsfehler.
func TestQueryService_LatestByStream_RequiresProjectID(t *testing.T) {
	repo := &queryRepo{}
	svc := newQuerySvc(t, repo, time.Now())
	if _, err := svc.LatestByStream(context.Background(), ""); err == nil {
		t.Fatal("expected error for empty projectID, got nil")
	}
}

// HistoryByStream: Sample-Liste wird als History-Items zurückgegeben.
func TestQueryService_HistoryByStream(t *testing.T) {
	now := time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC)
	repo := &queryRepo{
		history: []domain.SrtHealthSample{
			sampleAt("srt-test", now.Add(-1*time.Second)),
			sampleAt("srt-test", now.Add(-3*time.Second)),
		},
	}
	svc := newQuerySvc(t, repo, now)

	got, err := svc.HistoryByStream(context.Background(), "demo", "srt-test", 50)
	if err != nil {
		t.Fatalf("HistoryByStream: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 items, got %d", len(got))
	}
	if got[0].SampleAgeMillis != 1000 || got[1].SampleAgeMillis != 3000 {
		t.Errorf("SampleAgeMillis mismatch: got %d, %d", got[0].SampleAgeMillis, got[1].SampleAgeMillis)
	}
	if repo.gotHistoryQuery.Limit != 50 {
		t.Errorf("limit pass-through: got %d, want 50", repo.gotHistoryQuery.Limit)
	}
}

// HistoryByStream: leere Liste → ErrSrtHealthStreamUnknown.
func TestQueryService_HistoryByStream_UnknownStream(t *testing.T) {
	now := time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC)
	repo := &queryRepo{history: nil}
	svc := newQuerySvc(t, repo, now)
	_, err := svc.HistoryByStream(context.Background(), "demo", "missing", 0)
	if !errors.Is(err, application.ErrSrtHealthStreamUnknown) {
		t.Fatalf("expected ErrSrtHealthStreamUnknown, got %v", err)
	}
}

// HistoryByStream: Limit-Clamping.
func TestQueryService_HistoryByStream_LimitClamping(t *testing.T) {
	now := time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		in, want int
	}{
		{0, application.DefaultSrtHealthHistoryLimit},
		{-5, application.DefaultSrtHealthHistoryLimit},
		{50, 50},
		{2000, application.MaxSrtHealthHistoryLimit},
	}
	for _, tc := range cases {
		repo := &queryRepo{history: []domain.SrtHealthSample{sampleAt("a", now)}}
		svc := newQuerySvc(t, repo, now)
		_, err := svc.HistoryByStream(context.Background(), "demo", "a", tc.in)
		if err != nil {
			t.Fatalf("HistoryByStream(in=%d): %v", tc.in, err)
		}
		if repo.gotHistoryQuery.Limit != tc.want {
			t.Errorf("limit clamp: in=%d got=%d want=%d", tc.in, repo.gotHistoryQuery.Limit, tc.want)
		}
	}
}
