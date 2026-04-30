package metrics_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/metrics"
)

// TestPrometheusPublisher_HandlerExposesAllCounters scrapt den
// /api/metrics-Endpoint und prüft, dass die aggregierten Pflicht-
// Metriken aus Lastenheft §7.9 exposiert sind, jeweils mit ihren
// initialen Werten sowie nach gezielten Increments.
func TestPrometheusPublisher_HandlerExposesAllCounters(t *testing.T) {
	t.Parallel()
	p := metrics.NewPrometheusPublisher(metrics.WithActiveSessionsFunc(func() float64 { return 2 }))

	// Inkremente decken alle vier Branches; n=0 darf kein Increment sein
	// (call-site-Uniformität, siehe Method-Doc).
	p.EventsAccepted(3)
	p.EventsAccepted(0)
	p.InvalidEvents(2)
	p.InvalidEvents(0)
	p.RateLimitedEvents(1)
	p.RateLimitedEvents(0)
	p.DroppedEvents(4)
	p.DroppedEvents(0)
	p.PlaybackErrors(5)
	p.PlaybackErrors(0)
	p.RebufferEvents(6)
	p.RebufferEvents(0)
	p.APIRequests(7)
	p.APIRequests(0)
	p.StartupTimeMS(1234)

	srv := httptest.NewServer(p.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatalf("scrape: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d want 200", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	scrape := string(body)

	want := map[string]string{
		"mtrace_playback_events_total":     "3",
		"mtrace_invalid_events_total":      "2",
		"mtrace_rate_limited_events_total": "1",
		"mtrace_dropped_events_total":      "4",
		"mtrace_playback_errors_total":     "5",
		"mtrace_rebuffer_events_total":     "6",
		"mtrace_api_requests_total":        "7",
		"mtrace_active_sessions":           "2",
		"mtrace_startup_time_ms":           "1234",
	}
	for name, expected := range want {
		line := name + " " + expected
		if !strings.Contains(scrape, line) {
			t.Errorf("scrape missing %q\nfull scrape:\n%s", line, scrape)
		}
	}
}

// TestPrometheusPublisher_NegativeCallsAreNoop deckt den Branch ab, in
// dem die Counter-Methoden für n<=0 nichts tun — die Pflicht-Counter
// bleiben bei 0.
func TestPrometheusPublisher_NegativeCallsAreNoop(t *testing.T) {
	t.Parallel()
	p := metrics.NewPrometheusPublisher()
	p.EventsAccepted(-5)
	p.InvalidEvents(-1)
	p.RateLimitedEvents(-2)
	p.DroppedEvents(-3)
	p.PlaybackErrors(-4)
	p.RebufferEvents(-5)
	p.APIRequests(-6)
	p.StartupTimeMS(-7)

	srv := httptest.NewServer(p.Handler())
	defer srv.Close()
	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatalf("scrape: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	scrape := string(body)
	for _, name := range []string{
		"mtrace_playback_events_total 0",
		"mtrace_invalid_events_total 0",
		"mtrace_rate_limited_events_total 0",
		"mtrace_dropped_events_total 0",
		"mtrace_playback_errors_total 0",
		"mtrace_rebuffer_events_total 0",
		"mtrace_api_requests_total 0",
		"mtrace_active_sessions 0",
		"mtrace_startup_time_ms 0",
	} {
		if !strings.Contains(scrape, name) {
			t.Errorf("scrape missing %q", name)
		}
	}
}
