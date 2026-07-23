package application

// SRT-Health-Query-Service (RAK-43).
//
// Liefert die Read-Use-Cases für `GET /api/srt/health` und
// `GET /api/srt/health/{stream_id}` (spec/backend-api-contract.md
// . Der Service ist pure Application-Schicht: er liest aus
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
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driving"
)

// Limits für Health-History-Reads (spec §7a.3).
const (
	DefaultSrtHealthHistoryLimit = 100
	MaxSrtHealthHistoryLimit     = 1000
)

// SrtHealthQueryService erfüllt den SrtHealthInbound-Driving-Port (slice-004).
var _ driving.SrtHealthInbound = (*SrtHealthQueryService)(nil)

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
func (s *SrtHealthQueryService) LatestByStream(ctx context.Context, projectID string) ([]driving.SrtHealthSummary, error) {
	if projectID == "" {
		return nil, errors.New("SrtHealthQueryService: projectID is empty")
	}
	samples, err := s.repo.LatestByStream(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("srt-health repo latest: %w", err)
	}
	now := s.now()
	out := make([]driving.SrtHealthSummary, 0, len(samples))
	for _, sample := range samples {
		out = append(out, s.summarize(sample, now))
	}
	return out, nil
}

// HistoryByStream liefert die nächsten N Samples einer
// (projectID, streamID)-Kombination (spec §7a.1/§7a.3 —
// `GET /api/srt/health/{stream_id}`). `after` ist nil für die erste
// Seite; danach trägt der Aufrufer den NextAfter-Cursor zurück. Wenn
// auf der ersten Seite (after == nil) keine Samples existieren,
// liefert der Service `ErrSrtHealthStreamUnknown` (HTTP 404). Eine
// leere Folgeseite (after != nil, len(Items) == 0) ist hingegen
// kein Fehler — sie signalisiert „Stream existiert, keine weiteren
// Samples".
func (s *SrtHealthQueryService) HistoryByStream(ctx context.Context, projectID, streamID string, limit int, after *driven.SrtHealthCursor) (driving.SrtHealthHistoryPage, error) {
	if projectID == "" || streamID == "" {
		return driving.SrtHealthHistoryPage{}, errors.New("SrtHealthQueryService: projectID/streamID is empty")
	}
	page, err := s.repo.HistoryByStream(ctx, driven.SrtHealthHistoryQuery{
		ProjectID: projectID,
		StreamID:  streamID,
		Limit:     clampHistoryLimit(limit),
		After:     after,
	})
	if err != nil {
		return driving.SrtHealthHistoryPage{}, fmt.Errorf("srt-health repo history: %w", err)
	}
	if after == nil && len(page.Items) == 0 {
		return driving.SrtHealthHistoryPage{}, domain.ErrSrtHealthStreamUnknown
	}
	now := s.now()
	out := make([]driving.SrtHealthHistoryItem, 0, len(page.Items))
	for _, sample := range page.Items {
		out = append(out, s.summarize(sample, now))
	}
	return driving.SrtHealthHistoryPage{Items: out, NextAfter: page.NextAfter}, nil
}

// summarize berechnet die abgeleiteten Felder für einen Sample
// gegen die aktuelle Zeit.
func (s *SrtHealthQueryService) summarize(sample domain.SrtHealthSample, now time.Time) driving.SrtHealthSummary {
	headroom := bandwidthHeadroom(sample.AvailableBandwidthBPS, sample.RequiredBandwidthBPS)
	age := now.Sub(sample.IngestedAt).Milliseconds()
	if age < 0 {
		age = 0
	}
	return driving.SrtHealthSummary{
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
