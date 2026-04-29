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
// /api/metrics-Endpoint und prüft, dass die vier Pflicht-Counter aus
// API-Kontrakt §7 exposiert sind, jeweils mit ihren initialen Werten
// (0 nach NewPrometheusPublisher) sowie nach gezielten Increments.
func TestPrometheusPublisher_HandlerExposesAllCounters(t *testing.T) {
	t.Parallel()
	p := metrics.NewPrometheusPublisher()

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
	} {
		if !strings.Contains(scrape, name) {
			t.Errorf("scrape missing %q", name)
		}
	}
}
