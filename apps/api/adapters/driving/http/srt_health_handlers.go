package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/pt9912/m-trace/apps/api/hexagon/application"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// rfc3339Millis ist das einheitliche Output-Format für die SRT-
// Health-Read-Pfade (millisekunden-genau, UTC mit `Z`-Suffix).
const rfc3339Millis = "2006-01-02T15:04:05.000Z07:00"

// SrtHealthInbound ist der Driving-Port für die SRT-Health-Read-
// Pfade (plan-0.6.0 §5 Tranche 4 — RAK-43, spec/backend-api-contract.md
// §7a). Application-Layer implementiert das Interface über
// SrtHealthQueryService.
type SrtHealthInbound interface {
	LatestByStream(ctx context.Context, projectID string) ([]application.SrtHealthSummary, error)
	HistoryByStream(ctx context.Context, projectID, streamID string, limit int) ([]application.SrtHealthHistoryItem, error)
}

const (
	srtHealthListSpan = "http.handler GET /api/srt/health"
	srtHealthGetSpan  = "http.handler GET /api/srt/health/{stream_id}"
)

// SrtHealthListHandler implementiert `GET /api/srt/health`. Liefert
// pro StreamID den jüngsten persistierten Sample mit derived/
// freshness-Block (spec §7a.2).
type SrtHealthListHandler struct {
	UseCase  SrtHealthInbound
	Resolver driven.ProjectResolver
	Tracer   trace.Tracer
	Logger   *slog.Logger
}

func (h *SrtHealthListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.Tracer.Start(r.Context(), srtHealthListSpan,
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(
			attribute.String("http.method", http.MethodGet),
			attribute.String("http.route", "/api/srt/health"),
		),
	)
	defer span.End()

	rec := &statusRecorder{ResponseWriter: w}
	h.serve(ctx, rec, r)

	span.SetAttributes(attribute.Int("http.status_code", rec.statusCode()))
	if rec.statusCode() >= http.StatusInternalServerError {
		span.SetStatus(codes.Error, http.StatusText(rec.statusCode()))
	} else {
		span.SetStatus(codes.Ok, "")
	}
}

func (h *SrtHealthListHandler) serve(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeStatus(w, http.StatusMethodNotAllowed)
		return
	}
	projectID, ok := resolveProjectFromToken(ctx, w, r, h.Resolver)
	if !ok {
		return
	}
	summaries, err := h.UseCase.LatestByStream(ctx, projectID)
	if err != nil {
		h.Logger.Error("srt-health list failed", "error", err)
		writeStatus(w, http.StatusInternalServerError)
		return
	}
	items := make([]srtHealthWireItem, 0, len(summaries))
	for _, s := range summaries {
		items = append(items, encodeSrtHealthSummary(s))
	}
	writeJSON(w, http.StatusOK, srtHealthListResponse{Items: items})
}

// SrtHealthGetHandler implementiert
// `GET /api/srt/health/{stream_id}`. Liefert die letzten N Samples
// einer (project, stream)-Kombination mit derived-/freshness-Feldern.
type SrtHealthGetHandler struct {
	UseCase  SrtHealthInbound
	Resolver driven.ProjectResolver
	Tracer   trace.Tracer
	Logger   *slog.Logger
}

func (h *SrtHealthGetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.Tracer.Start(r.Context(), srtHealthGetSpan,
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(
			attribute.String("http.method", http.MethodGet),
			attribute.String("http.route", "/api/srt/health/{stream_id}"),
		),
	)
	defer span.End()

	rec := &statusRecorder{ResponseWriter: w}
	h.serve(ctx, rec, r)

	span.SetAttributes(attribute.Int("http.status_code", rec.statusCode()))
	if rec.statusCode() >= http.StatusInternalServerError {
		span.SetStatus(codes.Error, http.StatusText(rec.statusCode()))
	} else {
		span.SetStatus(codes.Ok, "")
	}
}

func (h *SrtHealthGetHandler) serve(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeStatus(w, http.StatusMethodNotAllowed)
		return
	}
	projectID, ok := resolveProjectFromToken(ctx, w, r, h.Resolver)
	if !ok {
		return
	}
	streamID := r.PathValue("stream_id")
	if streamID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "stream_id_required"})
		return
	}
	limit, err := parseSrtHealthLimit(r.URL.Query().Get("samples_limit"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "samples_limit_invalid"})
		return
	}
	items, err := h.UseCase.HistoryByStream(ctx, projectID, streamID, limit)
	if err != nil {
		if errors.Is(err, application.ErrSrtHealthStreamUnknown) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "stream_unknown", "stream_id": streamID})
			return
		}
		h.Logger.Error("srt-health detail failed", "error", err, "stream_id", streamID)
		writeStatus(w, http.StatusInternalServerError)
		return
	}
	wire := make([]srtHealthWireItem, 0, len(items))
	for _, s := range items {
		wire = append(wire, encodeSrtHealthSummary(s))
	}
	writeJSON(w, http.StatusOK, srtHealthDetailResponse{StreamID: streamID, Items: wire})
}

// srtHealthListResponse / srtHealthDetailResponse sind die Wire-
// Format-Wrapper. „items" ist auch beim Detail-Endpoint die
// Plural-Form, weil History eine Liste der letzten N Samples
// liefert (spec §7a.3).
type srtHealthListResponse struct {
	Items []srtHealthWireItem `json:"items"`
}

type srtHealthDetailResponse struct {
	StreamID string              `json:"stream_id"`
	Items    []srtHealthWireItem `json:"items"`
}

// srtHealthWireItem ist die JSON-Form aus spec §7a.2: trennt
// metrics/derived/freshness explizit, plus Top-Level health_state /
// source_status / source_error_code / connection_state.
type srtHealthWireItem struct {
	StreamID        string                 `json:"stream_id"`
	ConnectionID    string                 `json:"connection_id"`
	HealthState     domain.HealthState     `json:"health_state"`
	SourceStatus    domain.SourceStatus    `json:"source_status"`
	SourceErrorCode domain.SourceErrorCode `json:"source_error_code"`
	ConnectionState domain.ConnectionState `json:"connection_state"`
	Metrics         srtHealthWireMetrics   `json:"metrics"`
	Derived         srtHealthWireDerived   `json:"derived"`
	Freshness       srtHealthWireFreshness `json:"freshness"`
}

type srtHealthWireMetrics struct {
	RTTMillis             float64  `json:"rtt_ms"`
	PacketLossTotal       int64    `json:"packet_loss_total"`
	PacketLossRate        *float64 `json:"packet_loss_rate,omitempty"`
	RetransmissionsTotal  int64    `json:"retransmissions_total"`
	AvailableBandwidthBPS int64    `json:"available_bandwidth_bps"`
	ThroughputBPS         *int64   `json:"throughput_bps,omitempty"`
	RequiredBandwidthBPS  *int64   `json:"required_bandwidth_bps,omitempty"`
}

type srtHealthWireDerived struct {
	BandwidthHeadroomFactor *float64 `json:"bandwidth_headroom_factor,omitempty"`
}

type srtHealthWireFreshness struct {
	SourceObservedAt *string `json:"source_observed_at"`
	SourceSequence   *string `json:"source_sequence,omitempty"`
	CollectedAt      string  `json:"collected_at"`
	IngestedAt       string  `json:"ingested_at"`
	SampleAgeMillis  int64   `json:"sample_age_ms"`
	StaleAfterMillis int64   `json:"stale_after_ms"`
}

func encodeSrtHealthSummary(s application.SrtHealthSummary) srtHealthWireItem {
	sample := s.Sample
	return srtHealthWireItem{
		StreamID:        sample.StreamID,
		ConnectionID:    sample.ConnectionID,
		HealthState:     sample.HealthState,
		SourceStatus:    sample.SourceStatus,
		SourceErrorCode: sample.SourceErrorCode,
		ConnectionState: sample.ConnectionState,
		Metrics: srtHealthWireMetrics{
			RTTMillis:             sample.RTTMillis,
			PacketLossTotal:       sample.PacketLossTotal,
			PacketLossRate:        sample.PacketLossRate,
			RetransmissionsTotal:  sample.RetransmissionsTotal,
			AvailableBandwidthBPS: sample.AvailableBandwidthBPS,
			ThroughputBPS:         sample.ThroughputBPS,
			RequiredBandwidthBPS:  sample.RequiredBandwidthBPS,
		},
		Derived: srtHealthWireDerived{
			BandwidthHeadroomFactor: s.BandwidthHeadroom,
		},
		Freshness: srtHealthWireFreshness{
			SourceObservedAt: optionalRFC3339(sample.SourceObservedAt),
			SourceSequence:   optionalString(sample.SourceSequence),
			CollectedAt:      sample.CollectedAt.UTC().Format(rfc3339Millis),
			IngestedAt:       sample.IngestedAt.UTC().Format(rfc3339Millis),
			SampleAgeMillis:  s.SampleAgeMillis,
			StaleAfterMillis: s.StaleAfterMillis,
		},
	}
}

// optionalRFC3339 liefert nil für Zero-Time, sonst einen *string mit
// RFC3339-Millis-Format. JSON serialisiert nil als `null`, nicht als
// fehlendes Feld — spec §7a.2 zeigt `source_observed_at: null`
// explizit, weil MediaMTX in 0.6.0 keinen Source-Sample-Timestamp
// liefert.
func optionalRFC3339(t time.Time) *string {
	if t.IsZero() {
		return nil
	}
	s := t.UTC().Format(rfc3339Millis)
	return &s
}

// optionalString liefert nil für Empty-String.
func optionalString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// parseSrtHealthLimit parst `samples_limit` aus der Query. Leerwert
// → 0 (Service nutzt Default). Negative oder nicht-numerische Werte
// → Fehler.
func parseSrtHealthLimit(raw string) (int, error) {
	if raw == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0, errors.New("samples_limit invalid")
	}
	return n, nil
}

// Compile-Time-Check: SrtHealthQueryService implementiert
// SrtHealthInbound.
var _ SrtHealthInbound = (*application.SrtHealthQueryService)(nil)
