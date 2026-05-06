// Tests laufen im internen Package, weil sie auf die nicht-exportierte
// `webrtcMetrics`-Struktur und ihre `record`-Methode zugreifen — die
// Delta-Counter-Semantik ist absichtlich kein Public-Vertrag.
//
//nolint:testpackage // siehe oben: interner Zugriff auf newWebRTCMetrics ist beabsichtigt.
package metrics

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

func counterValue(t *testing.T, c prometheus.Metric) float64 {
	t.Helper()
	var m dto.Metric
	if err := c.Write(&m); err != nil {
		t.Fatalf("counter write: %v", err)
	}
	return m.GetCounter().GetValue()
}

// plan-0.8.0 §4 Tranche 3 — Delta-Counter-Semantik aus
// spec/telemetry-model.md §3.5.1.

func sample(runID string, sampleID, packetsLost, bytesReceived, bytesSent int64) driven.WebRTCSampleSnapshot {
	return driven.WebRTCSampleSnapshot{
		ProjectID:       "demo",
		SessionID:       "session-1",
		RunID:           runID,
		SampleID:        sampleID,
		ConnectionState: "connected",
		IceState:        "completed",
		DtlsState:       "connected",
		PacketsLost:     packetsLost,
		BytesReceived:   bytesReceived,
		BytesSent:       bytesSent,
	}
}

func TestWebRTCMetrics_FirstSampleIsBaselineNoDelta(t *testing.T) {
	t.Parallel()
	m := newWebRTCMetrics()
	m.record(sample("run-A", 0, 5, 1000, 500))

	if got := counterValue(t, m.packetsLostTotal); got != 0 {
		t.Fatalf("expected packets_lost_total to remain at baseline (0), got %v", got)
	}
	if got := counterValue(t, m.bytesReceivedTotal); got != 0 {
		t.Fatalf("expected bytes_received_total baseline 0, got %v", got)
	}
	// State-Counter inkrementieren immer — auch beim ersten Sample.
	if got := counterValue(t, m.connectionStateTotal.WithLabelValues("connected")); got != 1 {
		t.Fatalf("expected connection_state_total to count the sample, got %v", got)
	}
}

func TestWebRTCMetrics_DeltaIncrementsOnlyPositiveProgress(t *testing.T) {
	t.Parallel()
	m := newWebRTCMetrics()
	m.record(sample("run-A", 0, 0, 0, 0))
	m.record(sample("run-A", 1, 5, 1000, 500))
	m.record(sample("run-A", 2, 5, 1500, 1500))

	if got := counterValue(t, m.packetsLostTotal); got != 5 {
		t.Fatalf("expected packets_lost_total=5 after baseline+5, got %v", got)
	}
	if got := counterValue(t, m.bytesReceivedTotal); got != 1500 {
		t.Fatalf("expected bytes_received_total=1500, got %v", got)
	}
	if got := counterValue(t, m.bytesSentTotal); got != 1500 {
		t.Fatalf("expected bytes_sent_total=1500, got %v", got)
	}
}

func TestWebRTCMetrics_NegativeDeltaDoesNotIncrement(t *testing.T) {
	t.Parallel()
	m := newWebRTCMetrics()
	m.record(sample("run-A", 0, 100, 100, 100))
	// Counter-Reset / negative Delta — alle Werte fallen.
	m.record(sample("run-A", 1, 50, 50, 50))

	if got := counterValue(t, m.packetsLostTotal); got != 0 {
		t.Fatalf("expected packets_lost_total to stay at 0 after negative delta, got %v", got)
	}
}

func TestWebRTCMetrics_DuplicateSampleIDIgnored(t *testing.T) {
	t.Parallel()
	m := newWebRTCMetrics()
	m.record(sample("run-A", 0, 0, 0, 0))
	m.record(sample("run-A", 1, 10, 10, 10)) // delta 10
	m.record(sample("run-A", 1, 99, 99, 99)) // duplicate sample_id
	m.record(sample("run-A", 0, 99, 99, 99)) // older sample_id

	if got := counterValue(t, m.packetsLostTotal); got != 10 {
		t.Fatalf("expected packets_lost_total=10 (duplicate/older ignored), got %v", got)
	}
}

func TestWebRTCMetrics_NewRunIDStartsFreshBaseline(t *testing.T) {
	t.Parallel()
	m := newWebRTCMetrics()
	m.record(sample("run-A", 0, 0, 0, 0))
	m.record(sample("run-A", 1, 100, 100, 100)) // run-A-Delta = 100
	// Reconnect: neue run-id, baseline wird wieder 0.
	m.record(sample("run-B", 0, 1000, 1000, 1000))

	if got := counterValue(t, m.packetsLostTotal); got != 100 {
		t.Fatalf("expected packets_lost_total=100 (run-B baseline-only), got %v", got)
	}
	// State-Counter zählt aber alle drei Samples.
	if got := counterValue(t, m.connectionStateTotal.WithLabelValues("connected")); got != 3 {
		t.Fatalf("expected connection_state_total=3, got %v", got)
	}
}

func TestWebRTCMetrics_StateCountersAreIndependent(t *testing.T) {
	t.Parallel()
	m := newWebRTCMetrics()
	m.record(sample("run-A", 0, 0, 0, 0))
	s := sample("run-A", 1, 1, 1, 1)
	s.ConnectionState = "disconnected"
	s.IceState = "failed"
	s.DtlsState = "failed"
	m.record(s)

	if got := counterValue(t, m.connectionStateTotal.WithLabelValues("connected")); got != 1 {
		t.Fatalf("expected connected sample to count once, got %v", got)
	}
	if got := counterValue(t, m.connectionStateTotal.WithLabelValues("disconnected")); got != 1 {
		t.Fatalf("expected disconnected sample to count once, got %v", got)
	}
	if got := counterValue(t, m.iceStateTotal.WithLabelValues("failed")); got != 1 {
		t.Fatalf("expected ice failed sample, got %v", got)
	}
}

func TestWebRTCMetrics_NoForbiddenLabels(t *testing.T) {
	t.Parallel()
	// State-Counter dürfen nur ihr State-Label tragen; Byte-/Loss-
	// Counter sind label-frei (außer Target-Metadaten).
	m := newWebRTCMetrics()
	m.record(sample("run-A", 0, 1, 1, 1))
	m.record(sample("run-A", 1, 2, 2, 2))

	desc := m.connectionStateTotal.WithLabelValues("connected").Desc().String()
	if !strings.Contains(desc, "connection_state") {
		t.Fatalf("desc missing connection_state label: %s", desc)
	}
	for _, forbidden := range []string{"session_id", "project_id", "peer_connection_run_id", "user_agent"} {
		if strings.Contains(desc, forbidden) {
			t.Fatalf("connection_state desc must not contain %q: %s", forbidden, desc)
		}
	}
}
