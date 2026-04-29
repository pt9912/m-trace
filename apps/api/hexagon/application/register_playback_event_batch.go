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
	projects driven.ProjectResolver
	limiter  driven.RateLimiter
	events   driven.EventRepository
	metrics  driven.MetricsPublisher
	now      func() time.Time
}

// NewRegisterPlaybackEventBatchUseCase wires the use case with its
// driven ports. If now is nil, time.Now is used.
func NewRegisterPlaybackEventBatchUseCase(
	projects driven.ProjectResolver,
	limiter driven.RateLimiter,
	events driven.EventRepository,
	metrics driven.MetricsPublisher,
	now func() time.Time,
) *RegisterPlaybackEventBatchUseCase {
	if now == nil {
		now = time.Now
	}
	return &RegisterPlaybackEventBatchUseCase{
		projects: projects,
		limiter:  limiter,
		events:   events,
		metrics:  metrics,
		now:      now,
	}
}

// RegisterPlaybackEventBatch implements the validation order of
// docs/spike/backend-api-contract.md §5 from step 2 onwards. Steps 1
// (body size) and the bare presence of the X-MTrace-Token header are
// the HTTP adapter's responsibility.
//
// On error, the corresponding metric counter is incremented so that
// rejection counts are observable through GET /api/metrics.
func (u *RegisterPlaybackEventBatchUseCase) RegisterPlaybackEventBatch(
	ctx context.Context, in driving.BatchInput,
) (driving.BatchResult, error) {
	// Step 2 — auth: resolve token to project.
	project, err := u.projects.ResolveByToken(ctx, in.AuthToken)
	if err != nil {
		return driving.BatchResult{}, err
	}

	// Step 3 — rate limit: charged for the requested batch size, even
	// if the batch later turns out to be malformed. This prevents a
	// caller from probing for validation responses without paying the
	// per-project budget.
	if err := u.limiter.Allow(ctx, project.ID, len(in.Events)); err != nil {
		u.metrics.RateLimitedEvents(len(in.Events))
		return driving.BatchResult{}, err
	}

	// Step 4 — schema version.
	if in.SchemaVersion != SupportedSchemaVersion {
		u.metrics.InvalidEvents(len(in.Events))
		return driving.BatchResult{}, domain.ErrSchemaVersionMismatch
	}

	// Step 5 — batch shape.
	if len(in.Events) == 0 {
		// Empty batch: nothing to count, but the request itself is
		// invalid. Increment by 0 so the counter call still happens
		// (callers may rely on the side effect for tracing).
		u.metrics.InvalidEvents(0)
		return driving.BatchResult{}, domain.ErrBatchEmpty
	}
	if len(in.Events) > MaxBatchSize {
		u.metrics.InvalidEvents(len(in.Events))
		return driving.BatchResult{}, domain.ErrBatchTooLarge
	}

	now := u.now().UTC()
	parsed := make([]domain.PlaybackEvent, 0, len(in.Events))
	for _, e := range in.Events {
		// Step 6 — per-event field check.
		if !hasRequiredFields(e) {
			u.metrics.InvalidEvents(len(in.Events))
			return driving.BatchResult{}, domain.ErrInvalidEvent
		}
		// Step 7 — token/project binding.
		if e.ProjectID != project.ID {
			u.metrics.InvalidEvents(len(in.Events))
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

	// Step 8 — persist + accept.
	if err := u.events.Append(ctx, parsed); err != nil {
		u.metrics.DroppedEvents(len(parsed))
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
