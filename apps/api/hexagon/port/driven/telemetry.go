package driven

import (
	"context"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// Telemetry kapselt OTel-Aufrufe hinter einer frameworkneutralen
// Schnittstelle. Damit darf hexagon/ keinen direkten OTel-Import
// tragen — siehe spec/architecture.md §3.4 und der Boundary-Test
// scripts/check-architecture.sh.
//
// Der Use Case ruft BatchReceived am Eintritt jedes
// RegisterPlaybackEventBatch-Aufrufs (vor Step 3 Auth-Token), damit
// der Counter auch auth-fehlgeschlagene Requests zählt — mtrace.api.
// batches.received misst received, nicht validated.
type Telemetry interface {
	BatchReceived(ctx context.Context, size int)

	// SrtSampleRecorded erzeugt einen kurzlebigen Span pro
	// persistiertem SRT-Health-Sample (plan-0.6.0 §4 Sub-3.6,
	// spec/telemetry-model.md §7.8 — Span-Name
	// `mtrace.srt.health.collect`). Span-Attribute sind die
	// bounded Felder aus SrtSampleAttrs; per-Verbindung-Identifier
	// (`stream_id`, `connection_id`) gehen ausschließlich in den
	// Span (sample-basiert), nie in Prometheus-Labels (§7.7).
	SrtSampleRecorded(ctx context.Context, attrs SrtSampleAttrs)
}

// SrtSampleAttrs trägt die Span-Attribute aus
// spec/telemetry-model.md §7.8. Werte stammen aus dem persistierten
// SrtHealthSample; der Adapter mappt sie auf OTel-Attribute der
// Form `mtrace.srt.*`.
type SrtSampleAttrs struct {
	StreamID              string
	ConnectionID          string
	HealthState           domain.HealthState
	SourceStatus          domain.SourceStatus
	RTTMillis             float64
	AvailableBandwidthBPS int64
}
