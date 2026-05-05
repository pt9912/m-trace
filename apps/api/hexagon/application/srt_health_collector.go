package application

// SRT-Health-Collector (plan-0.6.0 §4 Sub-3.2).
//
// Der Collector orchestriert den Datenfluss aus
// spec/architecture.md §5.4: Quelle abfragen → Sample bewerten →
// persistieren. Polling-Loop, Backoff und Shutdown sind Sub-3.5;
// hier liegt die framework-freie Bewertungslogik plus die Single-
// Shot-`Collect`-Methode.

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// SrtHealthThresholds bündelt die Schwellen aus
// spec/telemetry-model.md §7.4. Werte sind in 0.6.0 Sub-3.2 als
// Konstanten festgelegt; eine spätere Phase kann das per Project-
// Konfiguration aufweichen, ohne das Domain-Modell zu ändern.
//
// LossWarnRatio/LossCriticalRatio werden gegen
// PacketLossTotal / max(1, PacketLossTotal+PacketsReceivedTotal)
// gerechnet — der Collector hat die Receive-Counter im Sample
// allerdings nicht direkt; in 0.6.0 nutzt er stattdessen
// PacketLossRate als Quelle, falls geliefert. Fehlt PacketLossRate,
// wird Loss nur über den absoluten Counter relativ zum Sample-Window
// bewertet (siehe Evaluate).
type SrtHealthThresholds struct {
	RTTWarnMillis              float64
	RTTCriticalMillis          float64
	LossWarnRatio              float64
	LossCriticalRatio          float64
	BandwidthHeadroomFactor    float64 // healthy verlangt available >= required × Factor
	StaleAfterMillis           int64
}

// DefaultThresholds liefert die Vorschlagswerte aus
// spec/telemetry-model.md §7.4. Tranche 4 finalisiert sie auf
// Basis von Lab-/Operator-Erfahrungen. Funktion statt Var, damit
// gochecknoglobals nicht ausschlägt und Aufrufer eine eigene
// mutable Kopie erhalten.
func DefaultThresholds() SrtHealthThresholds {
	return SrtHealthThresholds{
		RTTWarnMillis:           100,
		RTTCriticalMillis:       250,
		LossWarnRatio:           0.01,
		LossCriticalRatio:       0.05,
		BandwidthHeadroomFactor: 1.5,
		StaleAfterMillis:        15_000,
	}
}

// EvaluateInput trägt die Eingaben für eine Sample-Bewertung. Previous
// ist optional — wenn nicht-nil, vergleicht Evaluate Source-Sequence
// und IngestedAt für Stale-Erkennung (spec §7.6).
type EvaluateInput struct {
	Current    domain.SrtConnectionSample
	Previous   *domain.SrtHealthSample
	Now        time.Time
	Thresholds SrtHealthThresholds
}

// EvaluateResult ist das Bewertungsergebnis ohne Persistenz-Felder.
// CollectAndPersist hängt ProjectID und IngestedAt an, bevor es an
// das Repository übergeben wird.
type EvaluateResult struct {
	HealthState     domain.HealthState
	SourceStatus    domain.SourceStatus
	SourceErrorCode domain.SourceErrorCode
}

// Evaluate berechnet HealthState/SourceStatus/SourceErrorCode für
// einen Sample. Reine Funktion (kein Side-Effect, keine Adapter-
// Aufrufe) — testbar ohne Mocks.
func Evaluate(in EvaluateInput) EvaluateResult {
	if r, ok := evaluateNonOK(in); ok {
		return r
	}
	state := worstHealth(
		evaluateRTTHealth(in.Current.RTTMillis, in.Thresholds),
		evaluateLossHealth(in.Current.PacketLossRate, in.Thresholds),
		evaluateBandwidthHealth(in.Current.AvailableBandwidthBPS, in.Current.RequiredBandwidthBPS, in.Thresholds),
	)
	return EvaluateResult{
		HealthState:     state,
		SourceStatus:    domain.SourceStatusOK,
		SourceErrorCode: domain.SourceErrorCodeNone,
	}
}

// evaluateNonOK deckt die drei Fälle ab, in denen weder eine
// Health-Bewertung noch ein OK-Status sinnvoll ist: stale,
// no_active_connection, partial. Liefert (result, true) bei Treffer,
// sonst (zero, false).
func evaluateNonOK(in EvaluateInput) (EvaluateResult, bool) {
	if isStale(in) {
		return EvaluateResult{
			HealthState:     domain.HealthStateUnknown,
			SourceStatus:    domain.SourceStatusStale,
			SourceErrorCode: domain.SourceErrorCodeStaleSample,
		}, true
	}
	cur := in.Current
	if cur.ConnectionState == domain.ConnectionStateNoActiveConnection {
		return EvaluateResult{
			HealthState:     domain.HealthStateUnknown,
			SourceStatus:    domain.SourceStatusNoActiveConnection,
			SourceErrorCode: domain.SourceErrorCodeNoActiveConnection,
		}, true
	}
	if isPartialSample(cur) {
		return EvaluateResult{
			HealthState:     domain.HealthStateUnknown,
			SourceStatus:    domain.SourceStatusPartial,
			SourceErrorCode: domain.SourceErrorCodePartialSample,
		}, true
	}
	return EvaluateResult{}, false
}

// isStale prüft Source-Sequence-Drift gegen einen vorhergehenden
// Sample (spec §7.6). Ohne Vorgänger oder ohne Source-Sequence kein
// Stale.
func isStale(in EvaluateInput) bool {
	cur := in.Current
	if in.Previous == nil || cur.ConnectionState != domain.ConnectionStateConnected {
		return false
	}
	if cur.SourceSequence == "" || cur.SourceSequence != in.Previous.SourceSequence {
		return false
	}
	age := in.Now.Sub(in.Previous.IngestedAt).Milliseconds()
	return age >= in.Thresholds.StaleAfterMillis
}

// isPartialSample erfasst plausibilisierbar fehlerhafte Pflichtwerte
// (negativ, ConnectionState unbekannt, AvailableBandwidth nicht
// positiv).
func isPartialSample(cur domain.SrtConnectionSample) bool {
	if cur.ConnectionState == domain.ConnectionStateUnknown {
		return true
	}
	if cur.RTTMillis < 0 || cur.PacketLossTotal < 0 || cur.RetransmissionsTotal < 0 {
		return true
	}
	return cur.AvailableBandwidthBPS <= 0
}

// evaluateRTTHealth bewertet RTT-Snapshot gegen die Schwellen.
func evaluateRTTHealth(rttMillis float64, t SrtHealthThresholds) domain.HealthState {
	switch {
	case rttMillis >= t.RTTCriticalMillis:
		return domain.HealthStateCritical
	case rttMillis >= t.RTTWarnMillis:
		return domain.HealthStateDegraded
	default:
		return domain.HealthStateHealthy
	}
}

// evaluateLossHealth bewertet Paketverlust-Rate, wenn die Quelle sie
// liefert. Ohne Rate gibt sie Healthy zurück (Counter-only-Bewertung
// folgt mit Sub-3.5, wenn der Collector Vorgänger-Counter cached).
func evaluateLossHealth(rate *float64, t SrtHealthThresholds) domain.HealthState {
	if rate == nil {
		return domain.HealthStateHealthy
	}
	switch {
	case *rate >= t.LossCriticalRatio:
		return domain.HealthStateCritical
	case *rate >= t.LossWarnRatio:
		return domain.HealthStateDegraded
	default:
		return domain.HealthStateHealthy
	}
}

// evaluateBandwidthHealth bewertet verfügbare gegen erwartete
// Bandbreite. Ohne required-Wert gibt sie Healthy zurück (spec
// §7.4: keine Bandbreiten-Bewertung ohne Schwelle).
func evaluateBandwidthHealth(available int64, required *int64, t SrtHealthThresholds) domain.HealthState {
	if required == nil || *required <= 0 {
		return domain.HealthStateHealthy
	}
	req := *required
	switch {
	case available < req:
		return domain.HealthStateCritical
	case float64(available) < float64(req)*t.BandwidthHeadroomFactor:
		return domain.HealthStateDegraded
	default:
		return domain.HealthStateHealthy
	}
}

// worstHealth fasst mehrere Teil-Bewertungen zur schlechtesten
// zusammen. Reihenfolge: critical > degraded > healthy.
func worstHealth(states ...domain.HealthState) domain.HealthState {
	worst := domain.HealthStateHealthy
	for _, s := range states {
		switch s {
		case domain.HealthStateCritical:
			return domain.HealthStateCritical
		case domain.HealthStateDegraded:
			if worst == domain.HealthStateHealthy {
				worst = domain.HealthStateDegraded
			}
		}
	}
	return worst
}

// SrtHealthCollector orchestriert Snapshot → Bewertung → Persistenz.
// Polling-Loop, Backoff und Shutdown wandern in Sub-3.5; Collect ist
// Single-Shot und thread-safe.
type SrtHealthCollector struct {
	source     driven.SrtSource
	repo       driven.SrtHealthRepository
	now        func() time.Time
	thresholds SrtHealthThresholds
	projectID  string
}

// NewSrtHealthCollector verdrahtet die Driven-Ports. ProjectID ist
// Pflicht — der Collector schreibt alle Samples gegen dieses Project,
// weil MediaMTX-API in 0.6.0 keinen Project-Kontext mitliefert.
// Multi-Project-Support kommt mit einer Folgephase (siehe
// risks-backlog R-9 und plan-0.6.0 §0.1 „Multi-Tenant").
func NewSrtHealthCollector(
	source driven.SrtSource,
	repo driven.SrtHealthRepository,
	projectID string,
	now func() time.Time,
	thresholds SrtHealthThresholds,
) (*SrtHealthCollector, error) {
	if source == nil {
		return nil, errors.New("SrtHealthCollector: source is nil")
	}
	if repo == nil {
		return nil, errors.New("SrtHealthCollector: repo is nil")
	}
	if projectID == "" {
		return nil, errors.New("SrtHealthCollector: projectID is empty")
	}
	if now == nil {
		now = time.Now
	}
	return &SrtHealthCollector{
		source:     source,
		repo:       repo,
		now:        now,
		thresholds: thresholds,
		projectID:  projectID,
	}, nil
}

// Collect liest einen Snapshot, bewertet jede Verbindung gegen die
// vorhergehende und persistiert die resultierenden Samples. Bei
// Source-Fehler wird ein synthetisches `unavailable`-Sample pro
// bekanntem Stream nicht erzeugt (das wäre Folge-Logik in Sub-3.5,
// wo der Collector Vorgänger-Samples cached); stattdessen gibt
// Collect den Fehler zurück und der Aufrufer (Polling-Loop) zählt
// den Fehler für `mtrace_srt_health_collector_errors_total`.
func (c *SrtHealthCollector) Collect(ctx context.Context) error {
	now := c.now()

	samples, err := c.source.SnapshotConnections(ctx)
	if err != nil {
		return fmt.Errorf("srt-source snapshot: %w", err)
	}

	if len(samples) == 0 {
		// Keine Verbindung — Repository erhält keinen Eintrag in
		// 0.6.0 Sub-3.2. Sub-3.5 entscheidet, ob ein synthetisches
		// `no_active_connection`-Sample für historisch bekannte
		// Streams geschrieben wird.
		return nil
	}

	previous, err := c.repo.LatestByStream(ctx, c.projectID)
	if err != nil {
		return fmt.Errorf("srt-health-repo latest: %w", err)
	}

	// Map StreamID → Vorgänger-Sample für Stale-Erkennung.
	prevByStream := make(map[string]*domain.SrtHealthSample, len(previous))
	for i := range previous {
		s := previous[i]
		prevByStream[s.StreamID] = &s
	}

	out := make([]domain.SrtHealthSample, 0, len(samples))
	for _, cur := range samples {
		eval := Evaluate(EvaluateInput{
			Current:    cur,
			Previous:   prevByStream[cur.StreamID],
			Now:        now,
			Thresholds: c.thresholds,
		})

		out = append(out, domain.SrtHealthSample{
			ProjectID:    c.projectID,
			StreamID:     cur.StreamID,
			ConnectionID: cur.ConnectionID,

			SourceObservedAt: cur.SourceObservedAt,
			SourceSequence:   cur.SourceSequence,
			CollectedAt:      cur.CollectedAt,
			IngestedAt:       now,

			RTTMillis:             cur.RTTMillis,
			PacketLossTotal:       cur.PacketLossTotal,
			PacketLossRate:        cur.PacketLossRate,
			RetransmissionsTotal:  cur.RetransmissionsTotal,
			AvailableBandwidthBPS: cur.AvailableBandwidthBPS,
			ThroughputBPS:         cur.ThroughputBPS,
			RequiredBandwidthBPS:  cur.RequiredBandwidthBPS,
			SampleWindowMillis:    cur.SampleWindowMillis,

			SourceStatus:    eval.SourceStatus,
			SourceErrorCode: eval.SourceErrorCode,
			ConnectionState: cur.ConnectionState,
			HealthState:     eval.HealthState,
		})
	}

	if err := c.repo.Append(ctx, out); err != nil {
		return fmt.Errorf("srt-health-repo append: %w", err)
	}
	return nil
}
