package domain

import "time"

// HealthState ist der bewertete SRT-Verbindungs-Zustand
// (spec/telemetry-model.md §7.4).
//
//   - Healthy:  alle Pflichtwerte verfügbar, Schwellen unterschritten.
//   - Degraded: einer der Werte zwischen Soll und kritischer Schwelle.
//   - Critical: einer der Werte über kritischer Schwelle, oder
//     verfügbare Bandbreite < required.
//   - Unknown:  source_status != ok, oder Pflichtwerte teilweise
//     fehlen, oder Stale-Erkennung schlägt an.
//
// Die Schwellen sind in plan-0.6.0 §4 Sub-3.2 als Konstanten
// festgelegt (siehe SrtHealthThresholds).
type HealthState string

// Health-Zustände aus spec/telemetry-model.md §7.4. Werte sind als
// bounded Prometheus-Label `health_state` freigegeben (§3.2).
const (
	HealthStateHealthy  HealthState = "healthy"
	HealthStateDegraded HealthState = "degraded"
	HealthStateCritical HealthState = "critical"
	HealthStateUnknown  HealthState = "unknown"
)

// SourceStatus klassifiziert den Zustand der Metrikquelle pro Sample
// (spec/telemetry-model.md §7.5).
type SourceStatus string

// Source-Status-Werte aus spec/telemetry-model.md §7.5.
const (
	SourceStatusOK                 SourceStatus = "ok"
	SourceStatusUnavailable        SourceStatus = "unavailable"
	SourceStatusPartial            SourceStatus = "partial"
	SourceStatusStale              SourceStatus = "stale"
	SourceStatusNoActiveConnection SourceStatus = "no_active_connection"
)

// SourceErrorCode ist die stabile Fehlerklasse zur Source-Status-
// Klassifikation (spec/telemetry-model.md §7.5). `none` bei
// SourceStatusOK; sonst eine der vorgegebenen Codes.
type SourceErrorCode string

// Stabile Fehlerklassen aus spec/telemetry-model.md §7.5.
const (
	SourceErrorCodeNone               SourceErrorCode = "none"
	SourceErrorCodeSourceUnavailable  SourceErrorCode = "source_unavailable"
	SourceErrorCodeNoActiveConnection SourceErrorCode = "no_active_connection"
	SourceErrorCodePartialSample      SourceErrorCode = "partial_sample"
	SourceErrorCodeStaleSample        SourceErrorCode = "stale_sample"
	SourceErrorCodeParseError         SourceErrorCode = "parse_error"
)

// ConnectionState beschreibt den SRT-Verbindungszustand getrennt vom
// Source-Status: eine erreichbare Quelle ohne aktive Verbindung
// ist ein anderer Fall als eine nicht erreichbare Quelle.
type ConnectionState string

// SRT-Verbindungszustände (spec/telemetry-model.md §7.1) — getrennt
// vom SourceStatus, weil eine erreichbare Quelle ohne aktive
// Verbindung ein anderer Fall ist als eine nicht erreichbare Quelle.
const (
	ConnectionStateConnected          ConnectionState = "connected"
	ConnectionStateNoActiveConnection ConnectionState = "no_active_connection"
	ConnectionStateUnknown            ConnectionState = "unknown"
)

// SrtConnectionSample ist die normalisierte Rohdaten-Sicht aus der
// Metrikquelle (spec/telemetry-model.md §7.1). Adapter wandelt
// MediaMTX-spezifische Felder in dieses Domain-Modell um, bevor der
// Use Case Health-Bewertung vornimmt. Counter-Felder sind kumulativ
// ab Verbindungsstart (Reset bei ConnectionID-Wechsel).
type SrtConnectionSample struct {
	StreamID     string
	ConnectionID string

	// SourceObservedAt ist der Source-Sample-Zeitpunkt, falls die
	// Quelle ihn liefert. MediaMTX-API in 0.6.0 liefert ihn nicht —
	// dann ist das Feld die Null-Time und der Adapter setzt
	// SourceSequence als monotones Surrogat.
	SourceObservedAt time.Time
	SourceSequence   string

	CollectedAt time.Time

	RTTMillis              float64
	PacketLossTotal        int64
	PacketLossRate         *float64
	RetransmissionsTotal   int64
	AvailableBandwidthBPS  int64
	ThroughputBPS          *int64
	RequiredBandwidthBPS   *int64
	SampleWindowMillis     *int64

	ConnectionState ConnectionState
}

// SrtHealthSample ist der durable persistierte Sample (Tabelle
// srt_health_samples laut spec/backend-api-contract.md §10.6). Alle
// Felder aus SrtConnectionSample plus Bewertungsergebnis und
// IngestedAt-Timestamp.
type SrtHealthSample struct {
	ProjectID    string
	StreamID     string
	ConnectionID string

	SourceObservedAt time.Time
	SourceSequence   string
	CollectedAt      time.Time
	IngestedAt       time.Time

	RTTMillis              float64
	PacketLossTotal        int64
	PacketLossRate         *float64
	RetransmissionsTotal   int64
	AvailableBandwidthBPS  int64
	ThroughputBPS          *int64
	RequiredBandwidthBPS   *int64
	SampleWindowMillis     *int64

	SourceStatus    SourceStatus
	SourceErrorCode SourceErrorCode
	ConnectionState ConnectionState
	HealthState     HealthState
}

