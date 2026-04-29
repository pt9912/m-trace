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

// TestSetup_NoEnvVarsReturnsProvidersAndShutsDown verifiziert den
// No-Op-Fallback-Pfad aus plan-0.1.0.md §4.3: ohne `OTEL_*`-Env-Vars
// liefert Setup einsatzbereite Provider. Erzeugt einen Span (über den
// TracerProvider) und beendet beide Provider sauber via Shutdown —
// damit decken wir noopSpanExporter.{ExportSpans,Shutdown} ab.
func TestSetup_NoEnvVarsReturnsProvidersAndShutsDown(t *testing.T) {
	// Nicht parallel: der Setup-Pfad registriert globale Provider via
	// otel.SetMeterProvider/SetTracerProvider — andere Tests dürfen
	// dabei nicht queren.
	ctx := context.Background()

	providers, err := telemetry.Setup(ctx, "test-service", "test-version")
	if err != nil {
		t.Fatalf("Setup: %v", err)
	}
	if providers == nil || providers.Meter == nil || providers.Tracer == nil {
		t.Fatalf("Setup returned incomplete providers: %#v", providers)
	}

	// Span-Erzeugung + Force-Flush triggert noopSpanExporter.ExportSpans.
	tracer := providers.Tracer.Tracer(telemetry.TracerName)
	_, span := tracer.Start(ctx, "test-span")
	span.End()
	if err := providers.Tracer.ForceFlush(ctx); err != nil {
		t.Errorf("ForceFlush: %v", err)
	}

	// Meter()-Helper: deckt den Default-Pfad über den globalen Provider.
	if telemetry.Meter() == nil {
		t.Errorf("Meter() returned nil")
	}

	// Shutdown deckt den combined-Shutdown-Pfad inkl.
	// noopSpanExporter.Shutdown.
	if err := providers.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown: %v", err)
	}
}

// TestProviders_ShutdownNilSafe deckt den Pfad ab, in dem ein
// Provider-Bundle teilweise unvollständig ist — z. B. wenn Setup nach
// Meter-Provider-Erfolg vor Tracer-Provider-Setup abgebrochen wäre.
// Beide Felder gleichzeitig nil → Shutdown ist no-op und gibt nil
// zurück.
func TestProviders_ShutdownNilSafe(t *testing.T) {
	t.Parallel()
	p := &telemetry.Providers{}
	if err := p.Shutdown(context.Background()); err != nil {
		t.Errorf("nil providers Shutdown: expected nil, got %v", err)
	}
}
