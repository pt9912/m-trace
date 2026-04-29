package telemetry_test

import (
	"context"
	"testing"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/telemetry"
)

// TestOTelTelemetry_BatchReceivedIncrementsCounter verifies that
// BatchReceived increments the OTel counter mtrace.api.batches.received
// once per call. Uses a ManualReader to introspect collected metrics
// without an exporter.
func TestOTelTelemetry_BatchReceivedIncrementsCounter(t *testing.T) {
	t.Parallel()

	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	defer func() { _ = mp.Shutdown(context.Background()) }()

	tel, err := telemetry.NewOTelTelemetry(mp.Meter(telemetry.MeterName))
	if err != nil {
		t.Fatalf("NewOTelTelemetry: %v", err)
	}

	tel.BatchReceived(context.Background(), 3)
	tel.BatchReceived(context.Background(), 7)
	tel.BatchReceived(context.Background(), 1)

	var rm metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &rm); err != nil {
		t.Fatalf("collect: %v", err)
	}

	got := findCounter(t, &rm, "mtrace.api.batches.received")
	// Three calls, each with a different batch.size attribute → three
	// distinct data points, each with Value=1.
	if len(got.DataPoints) != 3 {
		t.Fatalf("expected 3 data points (one per batch.size), got %d", len(got.DataPoints))
	}
	for _, dp := range got.DataPoints {
		if dp.Value != 1 {
			t.Errorf("expected each data point Value=1, got %d", dp.Value)
		}
	}
}

// TestOTelTelemetry_NilMeterIsNoop verifies that constructing
// OTelTelemetry with a nil meter does not panic and the counter calls
// are no-ops.
func TestOTelTelemetry_NilMeterIsNoop(t *testing.T) {
	t.Parallel()

	tel, err := telemetry.NewOTelTelemetry(nil)
	if err != nil {
		t.Fatalf("NewOTelTelemetry(nil): %v", err)
	}
	// Must not panic.
	tel.BatchReceived(context.Background(), 5)
}

// findCounter extracts the named Int64-counter metric from the
// collected ResourceMetrics, failing the test if not found.
func findCounter(t *testing.T, rm *metricdata.ResourceMetrics, name string) metricdata.Sum[int64] {
	t.Helper()
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name != name {
				continue
			}
			sum, ok := m.Data.(metricdata.Sum[int64])
			if !ok {
				t.Fatalf("metric %q is not Sum[int64], got %T", name, m.Data)
			}
			return sum
		}
	}
	t.Fatalf("counter %q not found in collected metrics", name)
	return metricdata.Sum[int64]{}
}
