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
	_ driven.MetricsPublisher    = (*mockMetricsPublisher)(nil)
	_ driven.Telemetry           = (*mockTelemetry)(nil)
)

// mockMetricsPublisher tracks die SRT-spezifischen Counter; alle
// anderen Methoden des MetricsPublisher-Ports sind no-op.
type mockMetricsPublisher struct {
	samples []domain.HealthState
	runs    []domain.SourceStatus
	errors  []domain.SourceErrorCode
}

func (m *mockMetricsPublisher) EventsAccepted(int)            {}
func (m *mockMetricsPublisher) InvalidEvents(int)             {}
func (m *mockMetricsPublisher) RateLimitedEvents(int)         {}
func (m *mockMetricsPublisher) DroppedEvents(int)             {}
func (m *mockMetricsPublisher) PlaybackErrors(int)            {}
func (m *mockMetricsPublisher) RebufferEvents(int)            {}
func (m *mockMetricsPublisher) StartupTimeMS(float64)         {}
func (m *mockMetricsPublisher) SrtHealthSampleAccepted(state domain.HealthState) {
	m.samples = append(m.samples, state)
}
func (m *mockMetricsPublisher) SrtCollectorRun(status domain.SourceStatus) {
	m.runs = append(m.runs, status)
}
func (m *mockMetricsPublisher) SrtCollectorError(code domain.SourceErrorCode) {
	m.errors = append(m.errors, code)
}

// mockTelemetry erfasst SrtSampleRecorded-Aufrufe.
type mockTelemetry struct {
	samples []driven.SrtSampleAttrs
}

func (m *mockTelemetry) BatchReceived(_ context.Context, _ int) {}
func (m *mockTelemetry) SrtSampleRecorded(_ context.Context, attrs driven.SrtSampleAttrs) {
	m.samples = append(m.samples, attrs)
}

type mockSrtSource struct {
	samples    []domain.SrtConnectionSample
	err        error
	calls      int
	snapshotFn func() ([]domain.SrtConnectionSample, error)
}

func (m *mockSrtSource) SnapshotConnections(_ context.Context) ([]domain.SrtConnectionSample, error) {
	m.calls++
	if m.snapshotFn != nil {
		return m.snapshotFn()
	}
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

// Run: zwei aufeinanderfolgende Samples mit fortschreitendem
// SourceSequence werden persistiert (DoD aus plan-0.6.0 §4 für
// Sub-3.5 — „mindestens zwei Samples mit steigender Source-
// Sequence").
func TestRun_AppendsTwoConsecutiveSamples(t *testing.T) {
	src := &mockSrtSource{}
	repo := &mockSrtHealthRepo{}
	c, err := application.NewSrtHealthCollector(src, repo, projectID, fixedNow, application.DefaultThresholds())
	if err != nil {
		t.Fatalf("NewSrtHealthCollector: %v", err)
	}
	c.WithPollInterval(5 * time.Millisecond).WithMaxBackoff(20 * time.Millisecond)

	// Source liefert beim ersten Call Sample mit seq-1, beim zweiten
	// Call Sample mit seq-2 (steigender SourceSequence).
	calls := 0
	src.snapshotFn = func() ([]domain.SrtConnectionSample, error) {
		calls++
		seq := "seq-1"
		if calls == 2 {
			seq = "seq-2"
		}
		return []domain.SrtConnectionSample{{
			StreamID:              "srt-test",
			ConnectionID:          "c1",
			ConnectionState:       domain.ConnectionStateConnected,
			RTTMillis:             10,
			AvailableBandwidthBPS: 10_000_000,
			SourceSequence:        seq,
		}}, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		c.Run(ctx)
		close(done)
	}()

	// Warten, bis Collect mindestens zweimal lief, dann abbrechen.
	deadline := time.NewTimer(500 * time.Millisecond)
	defer deadline.Stop()
	for {
		if len(repo.appended) >= 2 {
			break
		}
		select {
		case <-deadline.C:
			cancel()
			<-done
			t.Fatalf("only %d samples appended within deadline", len(repo.appended))
		case <-time.After(2 * time.Millisecond):
		}
	}
	cancel()
	<-done

	if len(repo.appended) < 2 {
		t.Fatalf("expected >=2 appended samples, got %d", len(repo.appended))
	}
	first, second := repo.appended[0], repo.appended[1]
	if first.SourceSequence != "seq-1" || second.SourceSequence != "seq-2" {
		t.Fatalf("expected seq-1/seq-2 progression, got %q/%q", first.SourceSequence, second.SourceSequence)
	}
}

// Run: Source-Fehler → Backoff verdoppelt sich, Loop bricht aber nicht
// ab; nach Recovery wird wieder persistiert.
func TestRun_BackoffOnSourceError(t *testing.T) {
	src := &mockSrtSource{}
	repo := &mockSrtHealthRepo{}
	c, err := application.NewSrtHealthCollector(src, repo, projectID, fixedNow, application.DefaultThresholds())
	if err != nil {
		t.Fatalf("NewSrtHealthCollector: %v", err)
	}
	c.WithPollInterval(2 * time.Millisecond).WithMaxBackoff(50 * time.Millisecond)

	calls := 0
	src.snapshotFn = func() ([]domain.SrtConnectionSample, error) {
		calls++
		// Erste 2 Calls schlagen fehl, ab Call 3 liefert die Quelle.
		if calls < 3 {
			return nil, errors.New("source down")
		}
		return []domain.SrtConnectionSample{{
			StreamID:              "srt-test",
			ConnectionID:          "c1",
			ConnectionState:       domain.ConnectionStateConnected,
			RTTMillis:             10,
			AvailableBandwidthBPS: 10_000_000,
			SourceSequence:        "seq-recovered",
		}}, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		c.Run(ctx)
		close(done)
	}()

	deadline := time.NewTimer(500 * time.Millisecond)
	defer deadline.Stop()
	for {
		if len(repo.appended) >= 1 {
			break
		}
		select {
		case <-deadline.C:
			cancel()
			<-done
			t.Fatalf("expected recovery within deadline; calls=%d appended=%d", calls, len(repo.appended))
		case <-time.After(2 * time.Millisecond):
		}
	}
	cancel()
	<-done

	if calls < 3 {
		t.Fatalf("expected >=3 source calls (2 errors + 1 success), got %d", calls)
	}
	if len(repo.appended) < 1 || repo.appended[0].SourceSequence != "seq-recovered" {
		t.Fatalf("expected seq-recovered sample, got %+v", repo.appended)
	}
}

// Run: Context-Cancel beendet den Loop sauber.
func TestRun_ShutdownOnCancel(t *testing.T) {
	src := &mockSrtSource{}
	repo := &mockSrtHealthRepo{}
	c, err := application.NewSrtHealthCollector(src, repo, projectID, fixedNow, application.DefaultThresholds())
	if err != nil {
		t.Fatalf("NewSrtHealthCollector: %v", err)
	}
	c.WithPollInterval(50 * time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		c.Run(ctx)
		close(done)
	}()

	// Cancel sofort — Run muss dann zurückkehren, ohne den ersten
	// Tick abzuwarten.
	cancel()
	select {
	case <-done:
		// ok
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Run did not return after cancel")
	}
}

// Sub-3.6: Sample-Counter werden pro persistiertem Sample erhöht;
// Run-Counter steht am Ende auf SourceStatusOK (alle Samples healthy);
// Telemetry-Span wird pro Sample emittiert.
func TestCollect_EmitsMetricsAndSpans(t *testing.T) {
	src := &mockSrtSource{samples: []domain.SrtConnectionSample{{
		StreamID:              "srt-test",
		ConnectionID:          "c1",
		ConnectionState:       domain.ConnectionStateConnected,
		RTTMillis:             10,
		AvailableBandwidthBPS: 10_000_000,
		SourceSequence:        "seq-1",
	}}}
	repo := &mockSrtHealthRepo{}
	metrics := &mockMetricsPublisher{}
	telemetry := &mockTelemetry{}
	c := newCollector(t, src, repo)
	c.WithMetrics(metrics).WithTelemetry(telemetry)

	if err := c.Collect(context.Background()); err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if len(metrics.samples) != 1 || metrics.samples[0] != domain.HealthStateHealthy {
		t.Fatalf("expected 1 healthy sample counter, got %v", metrics.samples)
	}
	if len(metrics.runs) != 1 || metrics.runs[0] != domain.SourceStatusOK {
		t.Fatalf("expected 1 ok run counter, got %v", metrics.runs)
	}
	if len(metrics.errors) != 0 {
		t.Fatalf("expected 0 error counters on healthy run, got %v", metrics.errors)
	}
	if len(telemetry.samples) != 1 || telemetry.samples[0].StreamID != "srt-test" {
		t.Fatalf("expected 1 telemetry sample for srt-test, got %v", telemetry.samples)
	}
}

// Sub-3.6: Bei Source-Fehler wird der run-Counter auf
// SourceStatusUnavailable gesetzt + ein passender Error-Code
// emittiert.
func TestCollect_SourceErrorClassification(t *testing.T) {
	cases := []struct {
		name     string
		err      error
		wantCode domain.SourceErrorCode
	}{
		{"unauthorized → source_unavailable", driven.ErrSrtSourceUnauthorized, domain.SourceErrorCodeSourceUnavailable},
		{"unavailable → source_unavailable", driven.ErrSrtSourceUnavailable, domain.SourceErrorCodeSourceUnavailable},
		{"parse error → parse_error", driven.ErrSrtSourceParseError, domain.SourceErrorCodeParseError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			src := &mockSrtSource{err: tc.err}
			repo := &mockSrtHealthRepo{}
			metrics := &mockMetricsPublisher{}
			c := newCollector(t, src, repo)
			c.WithMetrics(metrics)

			err := c.Collect(context.Background())
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if len(metrics.runs) != 1 || metrics.runs[0] != domain.SourceStatusUnavailable {
				t.Fatalf("expected unavailable run, got %v", metrics.runs)
			}
			if len(metrics.errors) != 1 || metrics.errors[0] != tc.wantCode {
				t.Fatalf("expected error code %s, got %v", tc.wantCode, metrics.errors)
			}
		})
	}
}

// Sub-3.6: Empty-Snapshot wird als no_active_connection-Run gezählt.
func TestCollect_EmptySnapshotReportsNoActiveConnection(t *testing.T) {
	src := &mockSrtSource{samples: nil}
	repo := &mockSrtHealthRepo{}
	metrics := &mockMetricsPublisher{}
	c := newCollector(t, src, repo)
	c.WithMetrics(metrics)

	if err := c.Collect(context.Background()); err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if len(metrics.runs) != 1 || metrics.runs[0] != domain.SourceStatusNoActiveConnection {
		t.Fatalf("expected no_active_connection run, got %v", metrics.runs)
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
