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

// Compile-Time-Check: Mocks erfüllen die Driven-Ports.
var (
	_ driven.SrtSource           = (*mockSrtSource)(nil)
	_ driven.SrtHealthRepository = (*mockSrtHealthRepo)(nil)
)

type mockSrtSource struct {
	samples []domain.SrtConnectionSample
	err     error
	calls   int
}

func (m *mockSrtSource) SnapshotConnections(_ context.Context) ([]domain.SrtConnectionSample, error) {
	m.calls++
	if m.err != nil {
		return nil, m.err
	}
	return m.samples, nil
}

type mockSrtHealthRepo struct {
	latest    []domain.SrtHealthSample
	latestErr error

	appended    []domain.SrtHealthSample
	appendCalls int
	appendErr   error
}

func (m *mockSrtHealthRepo) Append(_ context.Context, samples []domain.SrtHealthSample) error {
	m.appendCalls++
	if m.appendErr != nil {
		return m.appendErr
	}
	m.appended = append(m.appended, samples...)
	return nil
}

func (m *mockSrtHealthRepo) LatestByStream(_ context.Context, _ string) ([]domain.SrtHealthSample, error) {
	if m.latestErr != nil {
		return nil, m.latestErr
	}
	return m.latest, nil
}

func (m *mockSrtHealthRepo) HistoryByStream(_ context.Context, _ driven.SrtHealthHistoryQuery) (driven.SrtHealthHistoryPage, error) {
	return driven.SrtHealthHistoryPage{}, nil
}

func ptr[T any](v T) *T { return &v }

const projectID = "test-project"

func fixedNow() time.Time { return time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC) }

func newCollector(t *testing.T, src *mockSrtSource, repo *mockSrtHealthRepo) *application.SrtHealthCollector {
	t.Helper()
	c, err := application.NewSrtHealthCollector(src, repo, projectID, fixedNow, application.DefaultThresholds())
	if err != nil {
		t.Fatalf("NewSrtHealthCollector: %v", err)
	}
	return c
}

// Healthy: alle Pflichtwerte unter Schwellen, available > required×1.5.
func TestEvaluate_Healthy(t *testing.T) {
	cur := domain.SrtConnectionSample{
		StreamID:              "srt-test",
		ConnectionID:          "c1",
		ConnectionState:       domain.ConnectionStateConnected,
		RTTMillis:             50,
		PacketLossTotal:       0,
		PacketLossRate:        ptr(0.0),
		RetransmissionsTotal:  0,
		AvailableBandwidthBPS: 10_000_000,
		RequiredBandwidthBPS:  ptr(int64(2_000_000)),
		SourceSequence:        "seq-1",
	}
	got := application.Evaluate(application.EvaluateInput{
		Current:    cur,
		Now:        fixedNow(),
		Thresholds: application.DefaultThresholds(),
	})
	if got.HealthState != domain.HealthStateHealthy ||
		got.SourceStatus != domain.SourceStatusOK ||
		got.SourceErrorCode != domain.SourceErrorCodeNone {
		t.Fatalf("expected healthy/ok/none, got %+v", got)
	}
}

// Degraded: RTT in [Warn, Critical).
func TestEvaluate_DegradedHighRTT(t *testing.T) {
	cur := domain.SrtConnectionSample{
		ConnectionState:       domain.ConnectionStateConnected,
		RTTMillis:             150, // zwischen 100 (Warn) und 250 (Critical)
		AvailableBandwidthBPS: 10_000_000,
	}
	got := application.Evaluate(application.EvaluateInput{
		Current: cur, Now: fixedNow(), Thresholds: application.DefaultThresholds(),
	})
	if got.HealthState != domain.HealthStateDegraded {
		t.Fatalf("expected degraded for rtt=150, got %s", got.HealthState)
	}
}

// Critical: RTT >= Critical-Schwelle.
func TestEvaluate_CriticalHighRTT(t *testing.T) {
	cur := domain.SrtConnectionSample{
		ConnectionState:       domain.ConnectionStateConnected,
		RTTMillis:             300,
		AvailableBandwidthBPS: 10_000_000,
	}
	got := application.Evaluate(application.EvaluateInput{
		Current: cur, Now: fixedNow(), Thresholds: application.DefaultThresholds(),
	})
	if got.HealthState != domain.HealthStateCritical {
		t.Fatalf("expected critical for rtt=300, got %s", got.HealthState)
	}
}

// Critical via Loss-Rate.
func TestEvaluate_CriticalHighLossRate(t *testing.T) {
	cur := domain.SrtConnectionSample{
		ConnectionState:       domain.ConnectionStateConnected,
		RTTMillis:             10,
		PacketLossRate:        ptr(0.10), // 10 % Loss → critical
		AvailableBandwidthBPS: 10_000_000,
	}
	got := application.Evaluate(application.EvaluateInput{
		Current: cur, Now: fixedNow(), Thresholds: application.DefaultThresholds(),
	})
	if got.HealthState != domain.HealthStateCritical {
		t.Fatalf("expected critical for loss=10%%, got %s", got.HealthState)
	}
}

// Critical via Bandbreite < required.
func TestEvaluate_CriticalBandwidthBelowRequired(t *testing.T) {
	cur := domain.SrtConnectionSample{
		ConnectionState:       domain.ConnectionStateConnected,
		RTTMillis:             10,
		AvailableBandwidthBPS: 1_000_000,
		RequiredBandwidthBPS:  ptr(int64(2_000_000)),
	}
	got := application.Evaluate(application.EvaluateInput{
		Current: cur, Now: fixedNow(), Thresholds: application.DefaultThresholds(),
	})
	if got.HealthState != domain.HealthStateCritical {
		t.Fatalf("expected critical for bandwidth-below-required, got %s", got.HealthState)
	}
}

// Degraded: Bandbreite zwischen required und required×Headroom.
func TestEvaluate_DegradedTightBandwidth(t *testing.T) {
	cur := domain.SrtConnectionSample{
		ConnectionState:       domain.ConnectionStateConnected,
		RTTMillis:             10,
		AvailableBandwidthBPS: 2_500_000, // zwischen 2M (required) und 3M (×1.5 Headroom)
		RequiredBandwidthBPS:  ptr(int64(2_000_000)),
	}
	got := application.Evaluate(application.EvaluateInput{
		Current: cur, Now: fixedNow(), Thresholds: application.DefaultThresholds(),
	})
	if got.HealthState != domain.HealthStateDegraded {
		t.Fatalf("expected degraded for tight bandwidth, got %s", got.HealthState)
	}
}

// Healthy ohne required: Bandbreite wird angezeigt, nicht bewertet.
func TestEvaluate_HealthyWithoutRequiredBandwidth(t *testing.T) {
	cur := domain.SrtConnectionSample{
		ConnectionState:       domain.ConnectionStateConnected,
		RTTMillis:             10,
		AvailableBandwidthBPS: 100_000, // niedrig, aber kein required → keine Bewertung
		RequiredBandwidthBPS:  nil,
	}
	got := application.Evaluate(application.EvaluateInput{
		Current: cur, Now: fixedNow(), Thresholds: application.DefaultThresholds(),
	})
	if got.HealthState != domain.HealthStateHealthy {
		t.Fatalf("expected healthy without required bandwidth, got %s", got.HealthState)
	}
}

// Unknown: NoActiveConnection.
func TestEvaluate_NoActiveConnection(t *testing.T) {
	cur := domain.SrtConnectionSample{
		ConnectionState:       domain.ConnectionStateNoActiveConnection,
		RTTMillis:             0,
		AvailableBandwidthBPS: 0,
	}
	got := application.Evaluate(application.EvaluateInput{
		Current: cur, Now: fixedNow(), Thresholds: application.DefaultThresholds(),
	})
	if got.HealthState != domain.HealthStateUnknown ||
		got.SourceStatus != domain.SourceStatusNoActiveConnection ||
		got.SourceErrorCode != domain.SourceErrorCodeNoActiveConnection {
		t.Fatalf("expected unknown/no_active_connection, got %+v", got)
	}
}

// Unknown: Partial-Sample (negativer RTT-Wert).
func TestEvaluate_PartialSample(t *testing.T) {
	cur := domain.SrtConnectionSample{
		ConnectionState:       domain.ConnectionStateConnected,
		RTTMillis:             -1, // ungültig
		AvailableBandwidthBPS: 10_000_000,
	}
	got := application.Evaluate(application.EvaluateInput{
		Current: cur, Now: fixedNow(), Thresholds: application.DefaultThresholds(),
	})
	if got.HealthState != domain.HealthStateUnknown ||
		got.SourceStatus != domain.SourceStatusPartial {
		t.Fatalf("expected unknown/partial, got %+v", got)
	}
}

// Stale: identische Source-Sequence über StaleAfterMillis.
func TestEvaluate_Stale(t *testing.T) {
	prev := &domain.SrtHealthSample{
		StreamID:       "srt-test",
		SourceSequence: "seq-frozen",
		IngestedAt:     fixedNow().Add(-30 * time.Second),
	}
	cur := domain.SrtConnectionSample{
		StreamID:              "srt-test",
		ConnectionState:       domain.ConnectionStateConnected,
		RTTMillis:             10,
		AvailableBandwidthBPS: 10_000_000,
		SourceSequence:        "seq-frozen",
	}
	got := application.Evaluate(application.EvaluateInput{
		Current: cur, Previous: prev, Now: fixedNow(), Thresholds: application.DefaultThresholds(),
	})
	if got.HealthState != domain.HealthStateUnknown ||
		got.SourceStatus != domain.SourceStatusStale ||
		got.SourceErrorCode != domain.SourceErrorCodeStaleSample {
		t.Fatalf("expected unknown/stale, got %+v", got)
	}
}

// Stale-Schwelle nicht überschritten: trotz identischer Sequence
// noch healthy.
func TestEvaluate_NotYetStale(t *testing.T) {
	prev := &domain.SrtHealthSample{
		StreamID:       "srt-test",
		SourceSequence: "seq-x",
		IngestedAt:     fixedNow().Add(-5 * time.Second), // unter 15s-Schwelle
	}
	cur := domain.SrtConnectionSample{
		StreamID:              "srt-test",
		ConnectionState:       domain.ConnectionStateConnected,
		RTTMillis:             10,
		AvailableBandwidthBPS: 10_000_000,
		SourceSequence:        "seq-x",
	}
	got := application.Evaluate(application.EvaluateInput{
		Current: cur, Previous: prev, Now: fixedNow(), Thresholds: application.DefaultThresholds(),
	})
	if got.HealthState != domain.HealthStateHealthy {
		t.Fatalf("expected healthy below stale threshold, got %s", got.HealthState)
	}
}

// Collect mit happy-path: Source liefert Sample, Repo persistiert.
func TestCollect_HappyPath(t *testing.T) {
	src := &mockSrtSource{samples: []domain.SrtConnectionSample{{
		StreamID:              "srt-test",
		ConnectionID:          "c1",
		ConnectionState:       domain.ConnectionStateConnected,
		RTTMillis:             10,
		AvailableBandwidthBPS: 10_000_000,
		SourceSequence:        "seq-1",
		CollectedAt:           fixedNow().Add(-1 * time.Second),
	}}}
	repo := &mockSrtHealthRepo{}
	c := newCollector(t, src, repo)

	if err := c.Collect(context.Background()); err != nil {
		t.Fatalf("Collect: %v", err)
	}

	if src.calls != 1 {
		t.Fatalf("expected 1 source call, got %d", src.calls)
	}
	if repo.appendCalls != 1 || len(repo.appended) != 1 {
		t.Fatalf("expected 1 append with 1 sample, got %d calls / %d samples", repo.appendCalls, len(repo.appended))
	}
	got := repo.appended[0]
	if got.ProjectID != projectID || got.StreamID != "srt-test" || got.ConnectionID != "c1" {
		t.Fatalf("unexpected sample identifiers: %+v", got)
	}
	if got.HealthState != domain.HealthStateHealthy {
		t.Fatalf("expected healthy, got %s", got.HealthState)
	}
	if !got.IngestedAt.Equal(fixedNow()) {
		t.Fatalf("expected IngestedAt=%v, got %v", fixedNow(), got.IngestedAt)
	}
}

// Collect mit leerem Source-Snapshot persistiert nichts.
func TestCollect_EmptySnapshot(t *testing.T) {
	src := &mockSrtSource{samples: nil}
	repo := &mockSrtHealthRepo{}
	c := newCollector(t, src, repo)

	if err := c.Collect(context.Background()); err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if repo.appendCalls != 0 {
		t.Fatalf("expected 0 append calls on empty snapshot, got %d", repo.appendCalls)
	}
}

// Collect propagiert Source-Fehler ohne Persistenz.
func TestCollect_SourceError(t *testing.T) {
	src := &mockSrtSource{err: errors.New("boom")}
	repo := &mockSrtHealthRepo{}
	c := newCollector(t, src, repo)

	err := c.Collect(context.Background())
	if err == nil {
		t.Fatal("expected error from source, got nil")
	}
	if repo.appendCalls != 0 {
		t.Fatalf("expected 0 append calls on source error, got %d", repo.appendCalls)
	}
}

// Collect nutzt LatestByStream als Stale-Vorgänger und schreibt
// Stale-Sample, wenn Source-Sequence eingefroren ist.
func TestCollect_StaleViaPreviousLookup(t *testing.T) {
	src := &mockSrtSource{samples: []domain.SrtConnectionSample{{
		StreamID:              "srt-test",
		ConnectionID:          "c1",
		ConnectionState:       domain.ConnectionStateConnected,
		RTTMillis:             10,
		AvailableBandwidthBPS: 10_000_000,
		SourceSequence:        "seq-frozen",
	}}}
	repo := &mockSrtHealthRepo{
		latest: []domain.SrtHealthSample{{
			StreamID:       "srt-test",
			SourceSequence: "seq-frozen",
			IngestedAt:     fixedNow().Add(-30 * time.Second),
		}},
	}
	c := newCollector(t, src, repo)

	if err := c.Collect(context.Background()); err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if len(repo.appended) != 1 {
		t.Fatalf("expected 1 append, got %d", len(repo.appended))
	}
	got := repo.appended[0]
	if got.HealthState != domain.HealthStateUnknown || got.SourceStatus != domain.SourceStatusStale {
		t.Fatalf("expected stale, got health=%s status=%s", got.HealthState, got.SourceStatus)
	}
}

// Konstruktor-Validierung.
func TestNewSrtHealthCollector_Validation(t *testing.T) {
	cases := []struct {
		name      string
		src       driven.SrtSource
		repo      driven.SrtHealthRepository
		projectID string
	}{
		{name: "nil source", src: nil, repo: &mockSrtHealthRepo{}, projectID: "p"},
		{name: "nil repo", src: &mockSrtSource{}, repo: nil, projectID: "p"},
		{name: "empty projectID", src: &mockSrtSource{}, repo: &mockSrtHealthRepo{}, projectID: ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := application.NewSrtHealthCollector(c.src, c.repo, c.projectID, nil, application.DefaultThresholds()); err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}
