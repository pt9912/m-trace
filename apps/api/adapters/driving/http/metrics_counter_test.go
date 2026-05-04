package http_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

// plan-0.4.0 §7.2 — Backend-Tests für die vier Pflichtcounter aus
// API-Kontrakt §7. Jeder Test pinnt explizit den Counter-Wert nach
// einem konkreten Request, damit Inkrement- und Null-Inkrement-Pfade
// getrennt nachweisbar sind.
//
// Die Tests laufen jeweils gegen einen frischen `newTestServer` —
// damit startet jede `prometheus.Counter`-Instanz bei 0 und der
// gemessene Delta = neuer Wert.

// scrapeCounterValue liest `/api/metrics` und gibt die Zeile mit
// `<name> <value>` als Substring-Treffer zurück. Tests asserten gegen
// `"<name> <expected>"`-Strings (analog zum Stil in
// `prometheus_publisher_test.go`), weil die Prometheus-Exposition pro
// label-freiem Counter genau eine `<name> <value>`-Zeile produziert.
func scrapeMetrics(t *testing.T, srv *httptest.Server) string {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/api/metrics", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("scrape: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	return string(body)
}

func assertCounter(t *testing.T, scrape, name string, expected int) {
	t.Helper()
	want := name + " "
	for _, line := range strings.Split(scrape, "\n") {
		if strings.HasPrefix(line, "# ") {
			continue
		}
		if strings.HasPrefix(line, want) {
			got := strings.TrimPrefix(line, want)
			expectedStr := strconv.Itoa(expected)
			if got == expectedStr {
				return
			}
			t.Errorf("counter %s = %q, want %q", name, got, expectedStr)
			return
		}
	}
	t.Errorf("counter %s not found in scrape:\n%s", name, scrape)
}

// TestMetrics_AcceptedCounter_HappyPath pinnt: ein 1-Event-Batch mit
// 202 → `mtrace_playback_events_total += 1`.
func TestMetrics_AcceptedCounter_HappyPath(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	resp := postEvents(t, srv, "demo-token", validBody)
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", resp.StatusCode)
	}
	scrape := scrapeMetrics(t, srv)
	assertCounter(t, scrape, "mtrace_playback_events_total", 1)
	assertCounter(t, scrape, "mtrace_invalid_events_total", 0)
	assertCounter(t, scrape, "mtrace_rate_limited_events_total", 0)
	assertCounter(t, scrape, "mtrace_dropped_events_total", 0)
}

// TestMetrics_InvalidCounter_SchemaVersion pinnt: `schema_version`
// abweichend → 400 → `mtrace_invalid_events_total += len(events)`.
func TestMetrics_InvalidCounter_SchemaVersion(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	body := strings.Replace(validBody, `"schema_version": "1.0"`, `"schema_version": "2.0"`, 1)
	resp := postEvents(t, srv, "demo-token", body)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	scrape := scrapeMetrics(t, srv)
	assertCounter(t, scrape, "mtrace_invalid_events_total", 1)
	assertCounter(t, scrape, "mtrace_playback_events_total", 0)
}

// TestMetrics_InvalidCounter_BatchTooLarge pinnt: 101 Events → 422
// (§5 step 5) → `mtrace_invalid_events_total += 101`. Nutzt den
// unlimited-Limiter, damit der Batch-Size-Pfad sichtbar ist (sonst
// würde der Token-Bucket schon bei 101 > 100 ablehnen).
func TestMetrics_InvalidCounter_BatchTooLarge(t *testing.T) {
	t.Parallel()
	srv := newServerWithUnlimitedRate(t)
	resp := postEvents(t, srv, "demo-token", batchOf(101))
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", resp.StatusCode)
	}
	scrape := scrapeMetrics(t, srv)
	assertCounter(t, scrape, "mtrace_invalid_events_total", 101)
	assertCounter(t, scrape, "mtrace_playback_events_total", 0)
}

// TestMetrics_InvalidCounter_MissingField pinnt: ein 1-Event-Batch
// ohne Pflichtfeld → 422 → `mtrace_invalid_events_total += 1`.
func TestMetrics_InvalidCounter_MissingField(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	body := strings.Replace(validBody, `"event_name": "rebuffer_started",`, ``, 1)
	resp := postEvents(t, srv, "demo-token", body)
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", resp.StatusCode)
	}
	scrape := scrapeMetrics(t, srv)
	assertCounter(t, scrape, "mtrace_invalid_events_total", 1)
	assertCounter(t, scrape, "mtrace_playback_events_total", 0)
}

// TestMetrics_InvalidCounter_NoIncrement_OnEmptyBatch pinnt: leerer
// Batch → 422 → `mtrace_invalid_events_total` bleibt 0 (Use-Case ruft
// `InvalidEvents(0)`, der Publisher bewegt nichts).
func TestMetrics_InvalidCounter_NoIncrement_OnEmptyBatch(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	resp := postEvents(t, srv, "demo-token", `{"schema_version":"1.0","events":[]}`)
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", resp.StatusCode)
	}
	scrape := scrapeMetrics(t, srv)
	assertCounter(t, scrape, "mtrace_invalid_events_total", 0)
	assertCounter(t, scrape, "mtrace_playback_events_total", 0)
}

// TestMetrics_InvalidCounter_NoIncrement_OnAuthError pinnt: kein Token
// → 401 → `mtrace_invalid_events_total` bleibt 0 (Header-Auth feuert
// vor dem Use-Case).
func TestMetrics_InvalidCounter_NoIncrement_OnAuthError(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	resp := postEvents(t, srv, "", validBody)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
	scrape := scrapeMetrics(t, srv)
	assertCounter(t, scrape, "mtrace_invalid_events_total", 0)
	assertCounter(t, scrape, "mtrace_playback_events_total", 0)
}

// TestMetrics_InvalidCounter_NoIncrement_OnBodyTooLarge pinnt: Body
// > 256 KB → 413 → `mtrace_invalid_events_total` bleibt 0
// (`MaxBytesReader` failt vor Use-Case-Aufruf).
func TestMetrics_InvalidCounter_NoIncrement_OnBodyTooLarge(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	body := strings.Repeat("X", 256*1024+1)
	resp := postEvents(t, srv, "demo-token", body)
	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", resp.StatusCode)
	}
	scrape := scrapeMetrics(t, srv)
	assertCounter(t, scrape, "mtrace_invalid_events_total", 0)
}

// TestMetrics_InvalidCounter_NoIncrement_OnMalformedJSON pinnt:
// kaputtes JSON → 400 → `mtrace_invalid_events_total` bleibt 0
// (`json.Unmarshal` failt vor Use-Case-Aufruf).
func TestMetrics_InvalidCounter_NoIncrement_OnMalformedJSON(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	resp := postEvents(t, srv, "demo-token", `{"schema_version":"1.0","events":[`)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	scrape := scrapeMetrics(t, srv)
	assertCounter(t, scrape, "mtrace_invalid_events_total", 0)
	assertCounter(t, scrape, "mtrace_playback_events_total", 0)
}

// TestMetrics_RateLimitedCounter_Increments pinnt: 429 → counter
// inkrementiert um `len(events)`. Stuck-Clock-Limiter erschöpft das
// Bucket; das nächste 1-Event-POST trifft den Rate-Limit-Pfad.
func TestMetrics_RateLimitedCounter_Increments(t *testing.T) {
	t.Parallel()
	fixed := time.Date(2026, 5, 4, 12, 0, 0, 0, time.UTC)
	srv := newTestServerWithClock(t, func() time.Time { return fixed })
	// Erst 100 Events durchwinken (Bucket leeren).
	resp := postEvents(t, srv, "demo-token", batchOf(100))
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("first batch: expected 202, got %d", resp.StatusCode)
	}
	// Dann 1 Event in den 429-Pfad.
	resp = postEvents(t, srv, "demo-token", validBody)
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", resp.StatusCode)
	}
	scrape := scrapeMetrics(t, srv)
	assertCounter(t, scrape, "mtrace_rate_limited_events_total", 1)
	assertCounter(t, scrape, "mtrace_playback_events_total", 100)
}

// TestMetrics_DroppedCounter_StaysZero pinnt: API-Kontrakt §7
// erlaubt `mtrace_dropped_events_total = 0` solange kein produktiver
// Drop-Pfad existiert. Plan-0.4.0 §7 (Tranche 6) implementiert keinen
// solchen Pfad — die Metrik muss aber sichtbar sein.
func TestMetrics_DroppedCounter_StaysZero(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	resp := postEvents(t, srv, "demo-token", validBody)
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", resp.StatusCode)
	}
	scrape := scrapeMetrics(t, srv)
	assertCounter(t, scrape, "mtrace_dropped_events_total", 0)
	if !strings.Contains(scrape, "mtrace_dropped_events_total") {
		t.Errorf("dropped-events counter must be exposed even when 0; scrape:\n%s", scrape)
	}
}
