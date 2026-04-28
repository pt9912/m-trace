// Minimal OpenTelemetry setup. Per
// docs/spike/0001-backend-stack.md §6.7 the spike requires only that
// OTel is wired into the build and code path; no exporter, no
// production collector, no full trace correlation. Bewertet wird die
// Ergonomie der Integration (Spec §9), nicht die produktive
// Telemetrie.
package dev.mtrace.api.adapters.driven.telemetry

import io.micronaut.context.annotation.Factory
import io.opentelemetry.api.OpenTelemetry
import io.opentelemetry.api.metrics.Meter
import io.opentelemetry.sdk.OpenTelemetrySdk
import io.opentelemetry.sdk.metrics.SdkMeterProvider
import io.opentelemetry.sdk.resources.Resource
import jakarta.inject.Singleton

const val METER_NAME: String = "dev.mtrace.api"

@Factory
class OtelFactory {

    // Spike scope: no preDestroy hook. JVM shutdown takes care of
    // cleaning up SDK resources; production setup will introduce a
    // proper lifecycle once an OTLP exporter is wired.
    @Singleton
    fun openTelemetry(): OpenTelemetry {
        val resource = Resource.getDefault().merge(
            Resource.builder()
                .put("service.name", "m-trace-api")
                .put("service.version", "0.1.0-spike")
                .put("mtrace.component", "api")
                .build(),
        )
        val meterProvider = SdkMeterProvider.builder()
            .setResource(resource)
            .build()
        return OpenTelemetrySdk.builder()
            .setMeterProvider(meterProvider)
            .build()
    }

    @Singleton
    fun meter(openTelemetry: OpenTelemetry): Meter = openTelemetry.getMeter(METER_NAME)
}
