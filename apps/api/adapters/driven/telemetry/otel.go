// Package telemetry holds a minimal OpenTelemetry setup. Per
// docs/spike/0001-backend-stack.md §6.7 the Spike requires only that
// OTel is wired into the build and code path; no exporter, no
// production collector, no full trace correlation.
//
// Bewertet wird die Ergonomie der Integration (Spec §9), nicht die
// produktive Telemetrie.
package telemetry

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// MeterName is the OTel scope used for spike instrumentation.
const MeterName = "github.com/example/m-trace/apps/api"

// Setup initializes a minimal OTel MeterProvider with a process-local
// resource and registers it as the global provider. The provider is
// returned so the caller can Shutdown it on graceful shutdown.
//
// No exporter is configured — OTel is "wired but silent" per
// Spec §6.7. A future MVP can swap in OTLP without touching call sites.
func Setup(serviceName, serviceVersion string) (*sdkmetric.MeterProvider, error) {
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

	mp := sdkmetric.NewMeterProvider(sdkmetric.WithResource(res))
	otel.SetMeterProvider(mp)
	return mp, nil
}

// Meter returns the spike's named meter from the globally registered
// provider. Call Setup first.
func Meter() metric.Meter {
	return otel.GetMeterProvider().Meter(MeterName)
}
