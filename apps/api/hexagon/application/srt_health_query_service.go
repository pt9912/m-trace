package application

// SRT-Health-Query-Service (plan-0.6.0 §5 Tranche 4 — RAK-43).
//
// Liefert die Read-Use-Cases für `GET /api/srt/health` und
// `GET /api/srt/health/{stream_id}` (spec/backend-api-contract.md
// §7a). Der Service ist pure Application-Schicht: er liest aus
// `SrtHealthRepository`, leitet derived/freshness-Felder ab und
// reicht das Ergebnis an den HTTP-Handler weiter, der das Wire-
// Format kodiert.

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// Limits für Health-History-Reads (spec §7a.3).
const (
	DefaultSrtHealthHistoryLimit = 100
	MaxSrtHealthHistoryLimit     = 1000
)

// ErrSrtHealthStreamUnknown wird vom Detail-Read-Pfad zurückgegeben,
// wenn die `stream_id` keinen Sample im Repository hat. HTTP-
// Adapter mappt das auf `404` (spec §7a.4).
var ErrSrtHealthStreamUnknown = errors.New("srt health: stream unknown")

// SrtHealthSummary ist die abgeleitete Sicht auf einen
// SrtHealthSample plus berechnete Felder (`derived`, `freshness`).
// Der HTTP-Adapter serialisiert das Schema in JSON gemäß spec
// §7a.2.
type SrtHealthSummary struct {
	Sample            domain.SrtHealthSample
	BandwidthHeadroom *float64 // available / required, falls required vorhanden
	SampleAgeMillis   int64    // Zeit seit IngestedAt zum Lesezeitpunkt
	StaleAfterMillis  int64    // Schwelle, ab der ein Sample als stale gilt
}

// SrtHealthHistoryItem trägt die Detail-Sicht: ein Sample plus
// dieselben derived-/freshness-Felder wie die Summary.
type SrtHealthHistoryItem = SrtHealthSummary

// SrtHealthQueryService implementiert die zwei Read-Operationen.
type SrtHealthQueryService struct {
	repo       driven.SrtHealthRepository
	now        func() time.Time
	thresholds SrtHealthThresholds
}

// NewSrtHealthQueryService verdrahtet das Repository plus die
// Schwellen-Konstanten. now=nil → time.Now.
func NewSrtHealthQueryService(repo driven.SrtHealthRepository, now func() time.Time, thresholds SrtHealthThresholds) (*SrtHealthQueryService, error) {
	if repo == nil {
		return nil, errors.New("SrtHealthQueryService: repo is nil")
	}
	if now == nil {
		now = time.Now
	}
	return &SrtHealthQueryService{
		repo:       repo,
		now:        now,
		thresholds: thresholds,
	}, nil
}

// LatestByStream liefert pro StreamID den jüngsten persistierten
// Sample des Projects (spec §7a.1 — `GET /api/srt/health`).
func (s *SrtHealthQueryService) LatestByStream(ctx context.Context, projectID string) ([]SrtHealthSummary, error) {
	if projectID == "" {
		return nil, errors.New("SrtHealthQueryService: projectID is empty")
	}
	samples, err := s.repo.LatestByStream(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("srt-health repo latest: %w", err)
	}
	now := s.now()
	out := make([]SrtHealthSummary, 0, len(samples))
	for _, sample := range samples {
		out = append(out, s.summarize(sample, now))
	}
	return out, nil
}

// HistoryByStream liefert die letzten N Samples einer
// (projectID, streamID)-Kombination (spec §7a.1 —
// `GET /api/srt/health/{stream_id}`). Wenn der Stream nie einen
// Sample hatte, gibt der Service `ErrSrtHealthStreamUnknown` zurück.
func (s *SrtHealthQueryService) HistoryByStream(ctx context.Context, projectID, streamID string, limit int) ([]SrtHealthHistoryItem, error) {
	if projectID == "" || streamID == "" {
		return nil, errors.New("SrtHealthQueryService: projectID/streamID is empty")
	}
	page, err := s.repo.HistoryByStream(ctx, driven.SrtHealthHistoryQuery{
		ProjectID: projectID,
		StreamID:  streamID,
		Limit:     clampHistoryLimit(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("srt-health repo history: %w", err)
	}
	if len(page.Items) == 0 {
		return nil, ErrSrtHealthStreamUnknown
	}
	now := s.now()
	out := make([]SrtHealthHistoryItem, 0, len(page.Items))
	for _, sample := range page.Items {
		out = append(out, s.summarize(sample, now))
	}
	return out, nil
}

// summarize berechnet die abgeleiteten Felder für einen Sample
// gegen die aktuelle Zeit.
func (s *SrtHealthQueryService) summarize(sample domain.SrtHealthSample, now time.Time) SrtHealthSummary {
	headroom := bandwidthHeadroom(sample.AvailableBandwidthBPS, sample.RequiredBandwidthBPS)
	age := now.Sub(sample.IngestedAt).Milliseconds()
	if age < 0 {
		age = 0
	}
	return SrtHealthSummary{
		Sample:            sample,
		BandwidthHeadroom: headroom,
		SampleAgeMillis:   age,
		StaleAfterMillis:  s.thresholds.StaleAfterMillis,
	}
}

// bandwidthHeadroom liefert das Verhältnis available/required, oder
// nil wenn required fehlt oder 0/negativ ist.
func bandwidthHeadroom(available int64, required *int64) *float64 {
	if required == nil || *required <= 0 {
		return nil
	}
	r := float64(available) / float64(*required)
	return &r
}

func clampHistoryLimit(limit int) int {
	switch {
	case limit <= 0:
		return DefaultSrtHealthHistoryLimit
	case limit > MaxSrtHealthHistoryLimit:
		return MaxSrtHealthHistoryLimit
	default:
		return limit
	}
}
