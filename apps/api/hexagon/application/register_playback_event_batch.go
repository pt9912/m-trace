// Package application holds the use-case services. This is the
// inner-hexagon layer that orchestrates domain logic via driven ports.
//
// Per docs/plan-spike.md §5.2, application code may import domain and
// driven/driving ports but no adapter (HTTP, JSON, Prometheus, OTel).
package application

import (
	"context"
	"strings"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driving"
)

// SupportedSchemaVersion is the only schema_version the spike accepts.
const SupportedSchemaVersion = "1.0"

// MaxBatchSize is the upper bound on events per request (Spec §6.1).
const MaxBatchSize = 100

// RegisterPlaybackEventBatchUseCase validates and persists a batch of
// player events. It implements driving.PlaybackEventInbound.
type RegisterPlaybackEventBatchUseCase struct {
	projects  driven.ProjectResolver
	limiter   driven.RateLimiter
	events    driven.EventRepository
	metrics   driven.MetricsPublisher
	telemetry driven.Telemetry
	analyzer  driven.StreamAnalyzer
	now       func() time.Time
}

// NewRegisterPlaybackEventBatchUseCase wires the use case with its
// driven ports. If now is nil, time.Now is used. analyzer ist die
// F-22-Architektur-Vorbereitung (siehe plan-0.1.0.md §5.1, F-22-Item):
// der Slot wird gesetzt, AnalyzeBatch jedoch erst ab 0.3.0 produktiv
// aufgerufen; bis dahin trägt main.go einen NoopStreamAnalyzer ein.
func NewRegisterPlaybackEventBatchUseCase(
	projects driven.ProjectResolver,
	limiter driven.RateLimiter,
	events driven.EventRepository,
	metrics driven.MetricsPublisher,
	telemetry driven.Telemetry,
	analyzer driven.StreamAnalyzer,
	now func() time.Time,
) *RegisterPlaybackEventBatchUseCase {
	if now == nil {
		now = time.Now
	}
	return &RegisterPlaybackEventBatchUseCase{
		projects:  projects,
		limiter:   limiter,
		events:    events,
		metrics:   metrics,
		telemetry: telemetry,
		analyzer:  analyzer,
		now:       now,
	}
}

// RegisterPlaybackEventBatch implements the validation order of
// docs/spike/backend-api-contract.md §5 from step 3 onwards. Steps 1
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

	// Step 3 — auth-token: resolve token to project. Auth-Fehler zählen
	// nicht in invalid_events.
	project, err := u.projects.ResolveByToken(ctx, in.AuthToken)
	if err != nil {
		return driving.BatchResult{}, err
	}

	// Step 4 — rate limit: charged for the requested batch size, even
	// if the batch later turns out to be malformed. This prevents a
	// caller from probing for validation responses without paying the
	// per-project budget.
	if err := u.limiter.Allow(ctx, project.ID, len(in.Events)); err != nil {
		u.metrics.RateLimitedEvents(len(in.Events))
		return driving.BatchResult{}, err
	}

	// Step 5 — schema version.
	if in.SchemaVersion != SupportedSchemaVersion {
		u.metrics.InvalidEvents(len(in.Events))
		return driving.BatchResult{}, domain.ErrSchemaVersionMismatch
	}

	// Step 6 — batch form: empty batch wird mit 422 abgelehnt; der
	// Counter zählt Events, nicht Batches — bei n=0 also kein
	// Counter-Increment (Lastenheft §7.9).
	if len(in.Events) == 0 {
		return driving.BatchResult{}, domain.ErrBatchEmpty
	}
	// Step 7 — batch size: zu viele Events.
	if len(in.Events) > MaxBatchSize {
		u.metrics.InvalidEvents(len(in.Events))
		return driving.BatchResult{}, domain.ErrBatchTooLarge
	}

	now := u.now().UTC()
	parsed := make([]domain.PlaybackEvent, 0, len(in.Events))
	for _, e := range in.Events {
		// Step 8 — per-event field check.
		if !hasRequiredFields(e) {
			u.metrics.InvalidEvents(len(in.Events))
			return driving.BatchResult{}, domain.ErrInvalidEvent
		}
		// Step 9 — token/project binding. Auth-Fehler (401) zählt nicht
		// in invalid_events — Counter ist auf Validierungsfehler 400/422
		// beschränkt.
		if e.ProjectID != project.ID {
			return driving.BatchResult{}, domain.ErrUnauthorized
		}
		ts, err := time.Parse(time.RFC3339Nano, e.ClientTimestamp)
		if err != nil {
			u.metrics.InvalidEvents(len(in.Events))
			return driving.BatchResult{}, domain.ErrInvalidEvent
		}
		parsed = append(parsed, domain.PlaybackEvent{
			EventName:        e.EventName,
			ProjectID:        e.ProjectID,
			SessionID:        e.SessionID,
			ClientTimestamp:  ts,
			ServerReceivedAt: now,
			SequenceNumber:   e.SequenceNumber,
			SDK: domain.SDKInfo{
				Name:    e.SDK.Name,
				Version: e.SDK.Version,
			},
		})
	}

	// Step 10 — persist + accept. Synchron fehlgeschlagenes Append ist
	// kein Backpressure-Drop und inkrementiert dropped_events nicht;
	// Sichtbarkeit erfolgt über HTTP-5xx-Histogramm und Logs.
	if err := u.events.Append(ctx, parsed); err != nil {
		return driving.BatchResult{}, err
	}

	u.metrics.EventsAccepted(len(parsed))
	return driving.BatchResult{Accepted: len(parsed)}, nil
}

func hasRequiredFields(e driving.EventInput) bool {
	return strings.TrimSpace(e.EventName) != "" &&
		strings.TrimSpace(e.ProjectID) != "" &&
		strings.TrimSpace(e.SessionID) != "" &&
		strings.TrimSpace(e.ClientTimestamp) != "" &&
		strings.TrimSpace(e.SDK.Name) != "" &&
		strings.TrimSpace(e.SDK.Version) != ""
}

// Compile-time check: the use case implements the inbound port.
var _ driving.PlaybackEventInbound = (*RegisterPlaybackEventBatchUseCase)(nil)
