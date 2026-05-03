// Package application holds the use-case services. This is the
// inner-hexagon layer that orchestrates domain logic via driven ports.
//
// Per docs/planning/done/plan-spike.md §5.2, application code may import domain and
// driven/driving ports but no adapter (HTTP, JSON, Prometheus, OTel).
package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driving"
)

// SupportedSchemaVersion is the only schema_version the API accepts.
const SupportedSchemaVersion = "1.0"

// MaxBatchSize is the upper bound on events per request (Spec §6.1).
const MaxBatchSize = 100

// TimeSkewThreshold ist die Konstante aus spec/telemetry-model.md §5.3:
// liegt `|client_timestamp - server_received_at|` über diesem Wert,
// markiert der Use-Case den Batch als skew-warned (kein Configuration-
// Item in 0.4.0).
const TimeSkewThreshold = 60 * time.Second

// RegisterPlaybackEventBatchUseCase validates and persists a batch of
// player events. It implements driving.PlaybackEventInbound.
type RegisterPlaybackEventBatchUseCase struct {
	projects  driven.ProjectResolver
	limiter   driven.RateLimiter
	events    driven.EventRepository
	sessions  driven.SessionRepository
	metrics   driven.MetricsPublisher
	telemetry driven.Telemetry
	analyzer  driven.StreamAnalyzer
	sequencer driven.IngestSequencer
	now       func() time.Time
}

// NewRegisterPlaybackEventBatchUseCase wires the use case with its
// driven ports. If now is nil, time.Now is used. analyzer ist die
// F-22-Architektur-Vorbereitung (siehe plan-0.1.0.md §5.1, F-22-Item):
// der Slot wird gesetzt, AnalyzeBatch jedoch erst ab 0.3.0 produktiv
// aufgerufen; bis dahin trägt main.go einen NoopStreamAnalyzer ein.
// sequencer liefert die serverseitige ingest_sequence pro Event vor
// dem Append (plan-0.1.0.md §5.1). sessions hält den aggregierten
// Session-State, der nach jedem akzeptierten Batch via
// UpsertFromEvents fortgeschrieben wird.
func NewRegisterPlaybackEventBatchUseCase(
	projects driven.ProjectResolver,
	limiter driven.RateLimiter,
	events driven.EventRepository,
	sessions driven.SessionRepository,
	metrics driven.MetricsPublisher,
	telemetry driven.Telemetry,
	analyzer driven.StreamAnalyzer,
	sequencer driven.IngestSequencer,
	now func() time.Time,
) *RegisterPlaybackEventBatchUseCase {
	if now == nil {
		now = time.Now
	}
	return &RegisterPlaybackEventBatchUseCase{
		projects:  projects,
		limiter:   limiter,
		events:    events,
		sessions:  sessions,
		metrics:   metrics,
		telemetry: telemetry,
		analyzer:  analyzer,
		sequencer: sequencer,
		now:       now,
	}
}

// RegisterPlaybackEventBatch implements the validation order of
// spec/backend-api-contract.md §5 from step 3 onwards. Steps 1
// (X-MTrace-Token header presence) and 2 (body size) are the HTTP
// adapter's responsibility.
//
// Counter semantics (API-Kontrakt §7, harmonisiert mit Lastenheft 1.1.2
// §7.9): mtrace_invalid_events_total zählt nur Validierungs-Rejects
// mit Status 400/422 (Schema, Batch-Form, Event-Felder). Auth-Fehler
// (401) — sowohl ResolveByToken als auch Token/Project-Bindung —
// inkrementieren keinen Counter. mtrace_dropped_events_total ist auf
// interne Backpressure-Drops beschränkt; synchron fehlgeschlagenes
// Append ist kein Drop und inkrementiert den Counter nicht.
func (u *RegisterPlaybackEventBatchUseCase) RegisterPlaybackEventBatch(
	ctx context.Context, in driving.BatchInput,
) (driving.BatchResult, error) {
	// OTel-Counter: vor Auth zählen, damit auch fehlgeschlagene Auth-
	// Requests im received-Counter erscheinen (siehe Telemetry-Port-Doc).
	u.telemetry.BatchReceived(ctx, len(in.Events))

	project, err := u.authorizeAndAdmit(ctx, in)
	if err != nil {
		return driving.BatchResult{}, err
	}

	parsed, timeSkewWarning, err := u.parseEvents(in, project.ID)
	if err != nil {
		return driving.BatchResult{}, err
	}

	correlations, err := u.resolveCorrelationIDs(ctx, parsed)
	if err != nil {
		return driving.BatchResult{}, err
	}
	for i := range parsed {
		parsed[i].CorrelationID = correlations[parsed[i].SessionID]
	}

	// Step 10 — persist + accept. Synchron fehlgeschlagenes Append ist
	// kein Backpressure-Drop und inkrementiert dropped_events nicht;
	// Sichtbarkeit erfolgt über HTTP-5xx-Histogramm und Logs. Session-
	// Aggregation läuft erst nach erfolgreichem Append, damit die
	// Sessions-Sicht nicht mit Events divergiert, die der Repository-
	// Append nicht akzeptiert hat.
	if err := u.events.Append(ctx, parsed); err != nil {
		return driving.BatchResult{}, err
	}
	if err := u.sessions.UpsertFromEvents(ctx, parsed); err != nil {
		return driving.BatchResult{}, err
	}

	u.metrics.EventsAccepted(len(parsed))
	u.publishPlaybackMetrics(parsed)
	return driving.BatchResult{
		Accepted:             len(parsed),
		ProjectID:            project.ID,
		SessionCount:         len(correlations),
		SessionCorrelationID: singleSessionCorrelationID(correlations),
		TimeSkewWarning:      timeSkewWarning,
	}, nil
}

// authorizeAndAdmit deckt API-Kontrakt §5 Steps 3–7 ab: Token →
// Project, Origin-Allowlist, Rate-Limit, Schema-Version und Batch-
// Größenbedingungen. Counter-Semantik (API-Kontrakt §7): Auth-Fehler
// zählen nicht in invalid_events; Schema-/Größen-Rejects zählen die
// volle Batch-Größe. Origin="" überspringt Project-Bindung
// (CLI/curl/Lab-Flow, plan-0.1.0.md §5.1).
func (u *RegisterPlaybackEventBatchUseCase) authorizeAndAdmit(
	ctx context.Context, in driving.BatchInput,
) (domain.Project, error) {
	project, err := u.projects.ResolveByToken(ctx, in.AuthToken)
	if err != nil {
		return domain.Project{}, err
	}
	if !project.IsOriginAllowed(in.Origin) {
		return domain.Project{}, domain.ErrOriginNotAllowed
	}
	limitKey := driven.RateLimitKey{
		ProjectID: project.ID,
		ClientIP:  in.ClientIP,
		Origin:    in.Origin,
	}
	if err := u.limiter.Allow(ctx, limitKey, len(in.Events)); err != nil {
		u.metrics.RateLimitedEvents(len(in.Events))
		return domain.Project{}, err
	}
	if in.SchemaVersion != SupportedSchemaVersion {
		u.metrics.InvalidEvents(len(in.Events))
		return domain.Project{}, domain.ErrSchemaVersionMismatch
	}
	// Empty batch: 422 ohne Counter-Increment (Lastenheft §7.9 — der
	// Counter zählt Events, nicht Batches; n=0 also kein Increment).
	if len(in.Events) == 0 {
		return domain.Project{}, domain.ErrBatchEmpty
	}
	if len(in.Events) > MaxBatchSize {
		u.metrics.InvalidEvents(len(in.Events))
		return domain.Project{}, domain.ErrBatchTooLarge
	}
	return project, nil
}

// parseEvents deckt API-Kontrakt §5 Steps 8–9 ab: per-Event-Feldcheck,
// Token/Project-Bindung und Time-Skew-Detection (telemetry-model.md
// §5.3, Schwelle 60 s). Liefert die domain-PlaybackEvent-Liste plus
// das Skew-Warning-Flag, das einmal aktiv für den ganzen Batch gilt.
func (u *RegisterPlaybackEventBatchUseCase) parseEvents(
	in driving.BatchInput, projectID string,
) ([]domain.PlaybackEvent, bool, error) {
	now := u.now().UTC()
	parsed := make([]domain.PlaybackEvent, 0, len(in.Events))
	timeSkewWarning := false
	for _, e := range in.Events {
		if !hasRequiredFields(e) {
			u.metrics.InvalidEvents(len(in.Events))
			return nil, false, domain.ErrInvalidEvent
		}
		// Token/Project-Bindung: Auth-Fehler (401) zählt nicht in
		// invalid_events — Counter ist auf 400/422 beschränkt.
		if e.ProjectID != projectID {
			return nil, false, domain.ErrUnauthorized
		}
		ts, err := time.Parse(time.RFC3339Nano, e.ClientTimestamp)
		if err != nil {
			u.metrics.InvalidEvents(len(in.Events))
			return nil, false, domain.ErrInvalidEvent
		}
		if now.Sub(ts).Abs() > TimeSkewThreshold {
			timeSkewWarning = true
		}
		parsed = append(parsed, domain.PlaybackEvent{
			EventName:        e.EventName,
			ProjectID:        e.ProjectID,
			SessionID:        e.SessionID,
			ClientTimestamp:  ts,
			ServerReceivedAt: now,
			IngestSequence:   u.sequencer.Next(),
			SequenceNumber:   e.SequenceNumber,
			SDK: domain.SDKInfo{
				Name:    e.SDK.Name,
				Version: e.SDK.Version,
			},
			Meta:    domain.EventMeta(copyEventMeta(e.Meta)),
			TraceID: in.Trace.TraceID,
			SpanID:  in.Trace.SpanID,
		})
	}
	return parsed, timeSkewWarning, nil
}

// resolveCorrelationIDs liefert für jede distinct session_id im Batch
// die zugehörige CorrelationID — entweder aus der bereits bekannten
// Session oder neu generiert. Verhalten:
//   - Bekannte Session mit nicht-leerer CorrelationID → übernimm.
//   - Bekannte Session ohne CorrelationID (Legacy / Test) → generiere
//     eine neue (Self-Healing für Daten von vor §3.2-Closeout).
//   - Unbekannte Session → generiere eine neue. UpsertFromEvents
//     übernimmt sie aus dem Event-Wert beim Insert.
func (u *RegisterPlaybackEventBatchUseCase) resolveCorrelationIDs(
	ctx context.Context, events []domain.PlaybackEvent,
) (map[string]string, error) {
	out := make(map[string]string)
	for _, e := range events {
		if _, ok := out[e.SessionID]; ok {
			continue
		}
		existing, err := u.sessions.Get(ctx, e.SessionID)
		switch {
		case errors.Is(err, domain.ErrSessionNotFound):
			cid, gErr := newCorrelationID()
			if gErr != nil {
				return nil, gErr
			}
			out[e.SessionID] = cid
		case err != nil:
			return nil, err
		default:
			if existing.CorrelationID != "" {
				out[e.SessionID] = existing.CorrelationID
				continue
			}
			cid, gErr := newCorrelationID()
			if gErr != nil {
				return nil, gErr
			}
			out[e.SessionID] = cid
		}
	}
	return out, nil
}

// singleSessionCorrelationID liefert die einzige CorrelationID, wenn
// alle Events im Batch dieselbe Session teilen — sonst leer. Adapter
// setzt das Span-Attribut `mtrace.session.correlation_id` nur, wenn
// der Wert nicht leer ist.
func singleSessionCorrelationID(correlations map[string]string) string {
	if len(correlations) != 1 {
		return ""
	}
	for _, v := range correlations {
		return v
	}
	return ""
}

// newCorrelationID generiert eine UUIDv4 (RFC 4122) als Hex-String mit
// Bindestrichen. crypto/rand garantiert Kryptosicherheit, was hier zwar
// nicht strikt nötig ist, aber kein zusätzlicher Aufwand und vermeidet
// Kollisionen bei sehr hoher Session-Frequenz.
func newCorrelationID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("application: correlation_id rand: %w", err)
	}
	// RFC 4122 §4.4: version 4 in highest 4 bits of byte 6.
	b[6] = (b[6] & 0x0f) | 0x40
	// RFC 4122 §4.4: variant 10 in highest 2 bits of byte 8.
	b[8] = (b[8] & 0x3f) | 0x80
	hexed := hex.EncodeToString(b[:])
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		hexed[0:8], hexed[8:12], hexed[12:16], hexed[16:20], hexed[20:32]), nil
}


func hasRequiredFields(e driving.EventInput) bool {
	return strings.TrimSpace(e.EventName) != "" &&
		strings.TrimSpace(e.ProjectID) != "" &&
		strings.TrimSpace(e.SessionID) != "" &&
		strings.TrimSpace(e.ClientTimestamp) != "" &&
		strings.TrimSpace(e.SDK.Name) != "" &&
		strings.TrimSpace(e.SDK.Version) != ""
}

func (u *RegisterPlaybackEventBatchUseCase) publishPlaybackMetrics(events []domain.PlaybackEvent) {
	var playbackErrors int
	var rebufferEvents int
	for _, e := range events {
		switch e.EventName {
		case "playback_error":
			playbackErrors++
		case "rebuffer_started":
			rebufferEvents++
		case "startup_time_measured":
			if duration, ok := numericMeta(e.Meta, "duration_ms"); ok {
				u.metrics.StartupTimeMS(duration)
			}
		}
	}
	u.metrics.PlaybackErrors(playbackErrors)
	u.metrics.RebufferEvents(rebufferEvents)
}

func numericMeta(meta domain.EventMeta, key string) (float64, bool) {
	if len(meta) == 0 {
		return 0, false
	}
	switch v := meta[key].(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case jsonNumber:
		f, err := v.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}

type jsonNumber interface {
	Float64() (float64, error)
}

func copyEventMeta(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// Compile-time check: the use case implements the inbound port.
var _ driving.PlaybackEventInbound = (*RegisterPlaybackEventBatchUseCase)(nil)
