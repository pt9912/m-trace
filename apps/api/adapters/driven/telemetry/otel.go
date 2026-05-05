// Package telemetry implements the driven.Telemetry port via the
// OpenTelemetry SDK. Per spec/architecture.md §3.4 this is one of two
// places (alongside adapters/driving/http) where OTel imports are
// allowed; hexagon/ stays OTel-frei (boundary-test in
// scripts/check-architecture.sh).
//
// The Setup function returns Providers with a graceful Shutdown(ctx)
// that combines the meter- and tracer-provider shutdown errors. Reader
// und Span-Exporter werden via autoexport mit explizitem No-Op-Fallback
// aufgelöst — siehe spec/architecture.md §5.3.
package telemetry

import (
	"context"
	"errors"
	"os"

	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// MeterName is the OTel scope used for instrumentation.
const MeterName = "github.com/pt9912/m-trace/apps/api"

// TracerName is the OTel scope used for span creation in the HTTP
// adapter (ServeHTTP wraps requests in a span).
const TracerName = "github.com/pt9912/m-trace/apps/api"

// counterBatchesReceived is the OTel counter name for BatchReceived
// calls. Per spec/telemetry-model.md §2.2 it appears in Prometheus as
// mtrace_api_batches_received after the OTel→Prom translation
// (`.` → `_`).
const counterBatchesReceived = "mtrace.api.batches.received"

// Providers bundles the meter- and tracer-provider returned by Setup.
// A single Shutdown(ctx) collapses both lifecycle calls and joins
// their errors.
type Providers struct {
	Meter  *sdkmetric.MeterProvider
	Tracer *sdktrace.TracerProvider
}

// Shutdown stops both providers and joins any errors. Safe to call
// when only one provider is set; nil-safe per provider.
func (p *Providers) Shutdown(ctx context.Context) error {
	var errs []error
	if p.Meter != nil {
		if err := p.Meter.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	if p.Tracer != nil {
		if err := p.Tracer.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// Setup initializes a MeterProvider and a TracerProvider with a
// process-local resource and registers them as the globals.
//
// Reader und Span-Exporter werden über autoexport mit explizitem
// No-Op-Fallback aufgelöst — ohne `OTEL_TRACES_EXPORTER=otlp` /
// `OTEL_METRICS_EXPORTER=otlp` (oder andere `OTEL_*`-Env-Vars) bleibt
// das Setup silent. autoexport defaultet ohne Fallback auf OTLP, was
// lokale Dev-Setups gegen einen nicht vorhandenen Collector pushen
// ließe; deshalb der explizite Fallback.
func Setup(ctx context.Context, serviceName, serviceVersion string) (*Providers, error) {
	unsetBlankOTelEnv()

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
			attribute.String("mtrace.component", "api"),
		),
	)
	if err != nil {
		return nil, err
	}

	// Metric reader with no-op fallback when no OTEL_METRICS_EXPORTER
	// is set.
	reader, err := autoexport.NewMetricReader(
		ctx,
		autoexport.WithFallbackMetricReader(func(_ context.Context) (sdkmetric.Reader, error) {
			return sdkmetric.NewManualReader(), nil
		}),
	)
	if err != nil {
		return nil, err
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(reader),
	)
	otel.SetMeterProvider(mp)

	// Span exporter with no-op fallback when no OTEL_TRACES_EXPORTER
	// is set.
	spanExporter, err := autoexport.NewSpanExporter(
		ctx,
		autoexport.WithFallbackSpanExporter(func(_ context.Context) (sdktrace.SpanExporter, error) {
			return noopSpanExporter{}, nil
		}),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(spanExporter),
	)
	otel.SetTracerProvider(tp)

	return &Providers{Meter: mp, Tracer: tp}, nil
}

func unsetBlankOTelEnv() {
	for _, key := range []string{
		"OTEL_EXPORTER_OTLP_ENDPOINT",
		"OTEL_EXPORTER_OTLP_PROTOCOL",
		"OTEL_TRACES_EXPORTER",
		"OTEL_METRICS_EXPORTER",
	} {
		if os.Getenv(key) == "" {
			_ = os.Unsetenv(key)
		}
	}
}

// noopSpanExporter discards all spans. Used as autoexport fallback.
type noopSpanExporter struct{}

func (noopSpanExporter) ExportSpans(_ context.Context, _ []sdktrace.ReadOnlySpan) error {
	return nil
}

func (noopSpanExporter) Shutdown(_ context.Context) error {
	return nil
}

// Meter returns the spike's named meter from the globally registered
// provider. Call Setup first.
func Meter() metric.Meter {
	return otel.GetMeterProvider().Meter(MeterName)
}

// OTelTelemetry implements driven.Telemetry by mapping BatchReceived
// onto an OTel Int64Counter. The counter is created lazily on the
// first call so that Setup may not yet have run during construction
// in tests; production code wires Setup before calling the use case.
type OTelTelemetry struct {
	counter metric.Int64Counter
}

// NewOTelTelemetry returns a telemetry implementation that uses the
// given meter to create the Int64Counter `mtrace.api.batches.received`.
// If meter is nil, a no-op meter is used (useful for tests that do
// not wire the SDK).
func NewOTelTelemetry(meter metric.Meter) (*OTelTelemetry, error) {
	if meter == nil {
		meter = noop.NewMeterProvider().Meter(MeterName)
	}
	counter, err := meter.Int64Counter(
		counterBatchesReceived,
		metric.WithDescription("Anzahl der via POST /api/playback-events empfangenen Batches"),
	)
	if err != nil {
		return nil, err
	}
	return &OTelTelemetry{counter: counter}, nil
}

// BatchReceived increments the counter by 1. The counter is label-free
// — `batch.size` would create an unbounded label domain because the
// counter runs in use case Step 0, before the MaxBatchSize=100
// validation; a rejected batch with events.length=250 would emit
// `batch_size="250"` to Prometheus. The per-request batch size lives on
// the `http.handler POST /api/playback-events` span instead (see
// adapters/driving/http/handler.go:73). See spec/telemetry-model.md
// §2.2 / §3.1 and docs/planning/in-progress/plan-0.4.0.md §8.2 for the
// full cardinality contract.
func (t *OTelTelemetry) BatchReceived(ctx context.Context, _ int) {
	t.counter.Add(ctx, 1)
}

// Compile-time check: OTelTelemetry implements driven.Telemetry.
var _ driven.Telemetry = (*OTelTelemetry)(nil)
