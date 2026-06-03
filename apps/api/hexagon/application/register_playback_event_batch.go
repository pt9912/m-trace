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

// TimeSkewThreshold ist die Konstante aus spec/telemetry-model.md:
// liegt `|client_timestamp - server_received_at|` über diesem Wert,
// markiert der Use-Case den Batch als skew-warned (kein Configuration-
// Item in 0.4.0).
const TimeSkewThreshold = 60 * time.Second

// SampleRateDriftTolerancePPM ist das Toleranz-Band (±100 ppm = ±0.01 %)
// fuer den Vergleich zwischen einem eingehenden `session_sample_rate`-
// Wert und dem bereits persistierten Pro-Session-Wert. Innerhalb der
// Toleranz gilt eine Abweichung als SDK-Rundungsartefakt (silent);
// jenseits davon zählt es als Drift und der Use-Case incrementiert
// `mtrace_sample_rate_drift_total{project_id}`.
const SampleRateDriftTolerancePPM = 100

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
	// broker ist optional; wenn gesetzt, ruft der Use-Case
	// `broker.Publish` nach erfolgreichem `EventRepository.Append`
	// und füttert SSE-Subscriber mit Mindestframes (§5 H4).
	broker *EventBroker
}

// NewRegisterPlaybackEventBatchUseCase wires the use case with its
// driven ports. If now is nil, time.Now is used. analyzer ist die
// F-22-Architektur-Vorbereitung (siehe, F-22-Item):
// der Slot wird gesetzt, AnalyzeBatch jedoch erst produktiv
// aufgerufen; bis dahin trägt main.go einen NoopStreamAnalyzer ein.
// sequencer liefert die serverseitige ingest_sequence pro Event vor
// dem Append. sessions hält den aggregierten
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

// WithBroker setzt den optionalen EventBroker für SSE-Live-Updates
// ( H4). Production-Wiring (`cmd/api`) ruft das, Tests
// können den Broker injizieren oder weglassen.
func (u *RegisterPlaybackEventBatchUseCase) WithBroker(broker *EventBroker) *RegisterPlaybackEventBatchUseCase {
	u.broker = broker
	return u
}

// RegisterPlaybackEventBatch implements the validation order of
// spec/backend-api-contract.md from step 3 onwards. Steps 1
// (X-MTrace-Token header presence) and 2 (body size) are the HTTP
// adapter's responsibility.
//
// Counter semantics (API-Kontrakt, harmonisiert mit Lastenheft 1.1.2
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

	//  D2: `session_boundaries[]` atomar mit den Events
	// validieren — invalider Block persistiert weder Events noch
	// Boundaries und erhöht `accepted` nicht (API-Kontrakt). Die
	// Validation läuft nach den Event-Pflichtchecks, damit der
	// Partition-Match (Boundary referenziert eine Session, die im
	// selben Batch ein Event trägt) auf den geparsten Events arbeitet.
	eventSessions := collectEventSessions(parsed)
	boundaries, err := parseAndValidateBoundaries(
		in.Boundaries, project.ID, eventSessions, u.now().UTC(),
	)
	if err != nil {
		u.metrics.InvalidEvents(len(in.Events))
		return driving.BatchResult{}, err
	}

	correlations, err := u.resolveCorrelationIDs(ctx, parsed)
	if err != nil {
		return driving.BatchResult{}, err
	}
	for i := range parsed {
		parsed[i].CorrelationID = correlations[parsed[i].SessionID]
	}

	// Step 10 — persist + accept. Reihenfolge ab C2
	// (R-6-Fix): zuerst Sessions upserten, damit DB-finale
	// `correlation_id` (Sieger des Race auf einer noch unbekannten
	// `(project_id, session_id)`) feststeht; danach Events mit dieser
	// canonical-CID enrichen und appenden. So trägt jedes persistierte
	// Event garantiert dieselbe CorrelationID wie die zugehörige
	// `stream_sessions`-Zeile.
	//
	// Synchron fehlgeschlagenes Append ist weiter kein Backpressure-Drop
	// und inkrementiert `dropped_events` nicht; Sichtbarkeit über
	// HTTP-5xx + Logs. Tradeoff der Reorder: kommt es zwischen Upsert
	// und Append zu einem 5xx, existiert die Session-Zeile mit
	// `event_count=1` ohne korrespondierendes Event. Sweep beendet sie
	// nach Idle-Timeout; ein Retry sieht die Session und tickt — kein
	// CorrelationID-Mismatch entsteht. Dieser Tradeoff ist schwächer als
	// die R-6-Inkonsistenz, die er ersetzt.
	canonical, err := u.sessions.UpsertFromEvents(ctx, parsed)
	if err != nil {
		return driving.BatchResult{}, err
	}
	for i := range parsed {
		if cid, ok := canonical[parsed[i].SessionID]; ok && cid != "" {
			parsed[i].CorrelationID = cid
			correlations[parsed[i].SessionID] = cid
		}
	}
	//  / R-10: persistiere Pro-Session-
	// `sample_rate_ppm` (immutable, erstmaliger Sub-1-Wert) und zähle
	// Drift gegen den bereits persistierten Wert. Reihenfolge zwischen
	// UpsertFromEvents und AppendBoundaries: Boundaries hängen am
	// Session-State, aber nicht an `sample_rate_ppm` — die Reihenfolge
	// ist innerhalb dieses Blocks frei wählbar. Sample-Rate-Update
	// läuft jetzt vor Boundaries, damit ein Boundary-Append-Fehler den
	// Drift-Counter nicht überspringt.
	if err := u.applySessionSampleRate(ctx, parsed); err != nil {
		return driving.BatchResult{}, err
	}
	//  D2: Boundaries nach Sessions-Upsert persistieren,
	// damit der Boundary-Record auf eine in `stream_sessions`-bestätigte
	// Partition verweist. Reihenfolge bleibt: Sessions → Boundaries →
	// Events; ein Boundary-Append-Fehler bricht den Batch ab, ohne
	// dass die Events bereits geschrieben wurden.
	if err := u.sessions.AppendBoundaries(ctx, boundaries); err != nil {
		return driving.BatchResult{}, err
	}
	if err := u.events.Append(ctx, parsed); err != nil {
		return driving.BatchResult{}, err
	}

	//  H4: SSE-Subscriber bekommen einen Mindestframe pro
	// erfolgreich persistiertem Event. Publish ist non-blocking; slow
	// Subscriber droppen den Frame und schließen die Lücke per
	// `Last-Event-ID`-Reconnect.
	if u.broker != nil {
		u.broker.Publish(parsed)
	}

	u.metrics.EventsAccepted(len(parsed))
	u.publishPlaybackMetrics(parsed)
	u.publishWebRTCSamples(parsed)
	return driving.BatchResult{
		Accepted:             len(parsed),
		ProjectID:            project.ID,
		SessionCount:         len(correlations),
		SessionCorrelationID: singleSessionCorrelationID(correlations),
		TimeSkewWarning:      timeSkewWarning,
	}, nil
}

// authorizeAndAdmit deckt API-Kontrakt Steps 3–7 ab: Token →
// Project, Origin-Allowlist, Rate-Limit, Schema-Version und Batch-
// Größenbedingungen. Counter-Semantik (API-Kontrakt): Auth-Fehler
// zählen nicht in invalid_events; Schema-/Größen-Rejects zählen die
// volle Batch-Größe. Origin="" überspringt Project-Bindung
// (CLI/curl/Lab-Flow, ).
func (u *RegisterPlaybackEventBatchUseCase) authorizeAndAdmit(
	ctx context.Context, in driving.BatchInput,
) (domain.Project, error) {
	var project domain.Project
	if in.PreResolvedProject != nil {
		// Pfad: HTTP-Adapter hat den Auth-Header bereits über
		// den Session-Token-Pfad aufgelöst. Use-Case überspringt
		// ResolveByToken; alle nachfolgenden Stufen (Origin-Allowlist,
		// Rate-Limit, Schema, Batch-Größe) gelten unverändert.
		project = *in.PreResolvedProject
	} else {
		var err error
		project, err = u.projects.ResolveByToken(ctx, in.AuthToken)
		if err != nil {
			return domain.Project{}, err
		}
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
	// Empty batch: 422 ohne Counter-Increment (der
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

// parseEvents deckt API-Kontrakt Steps 8–9 ab: per-Event-Feldcheck,
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
		eventSkew := now.Sub(ts).Abs() > TimeSkewThreshold
		if eventSkew {
			timeSkewWarning = true
		}
		//  D1: reservierte Meta-Keys vor Persistenz
		// typvalidieren (422 bei Domänen-/Typ-/Requires-Verstoß), dann
		// URL-Redaction für alle URL-verdächtigen Meta-Keys ausführen.
		// Reihenfolge ist verbindlich: Validation prüft den strikten
		// `network.redacted_url`-Vertrag, bevor die Redaction unbekannte
		// URL-Keys mutiert.
		meta := domain.EventMeta(copyEventMeta(e.Meta))
		if err := validateReservedEventMeta(meta); err != nil {
			u.metrics.InvalidEvents(len(in.Events))
			return nil, false, err
		}
		redactEventMetaURLs(meta)
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
			Meta:            meta,
			TraceID:         in.Trace.TraceID,
			SpanID:          in.Trace.SpanID,
			TimeSkewWarning: eventSkew,
		})
	}
	return parsed, timeSkewWarning, nil
}

// resolveCorrelationIDs liefert für jede distinct session_id im Batch
// die zugehörige CorrelationID — entweder aus der bereits bekannten
// Session oder neu generiert. Verhalten:
//  - Bekannte Session mit nicht-leerer CorrelationID → übernimm.
//  - Bekannte Session ohne CorrelationID (Legacy / Test) → generiere
//  eine neue (Self-Healing für Daten von vor §3.2-Closeout).
//  - Unbekannte Session → generiere eine neue. UpsertFromEvents
//  übernimmt sie aus dem Event-Wert beim Insert.
func (u *RegisterPlaybackEventBatchUseCase) resolveCorrelationIDs(
	ctx context.Context, events []domain.PlaybackEvent,
) (map[string]string, error) {
	out := make(map[string]string)
	for _, e := range events {
		if _, ok := out[e.SessionID]; ok {
			continue
		}
		existing, err := u.sessions.Get(ctx, e.ProjectID, e.SessionID)
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

// collectEventSessions sammelt die Set der `session_id`-Werte, die im
// geparsten Event-Array vorkommen. Boundary-Validation prüft gegen
// dieses Set, dass jede Boundary einer Partition mit mindestens einem
// Event im selben Batch zugeordnet ist (API-Kontrakt).
func collectEventSessions(events []domain.PlaybackEvent) map[string]struct{} {
	out := make(map[string]struct{}, len(events))
	for _, e := range events {
		out[e.SessionID] = struct{}{}
	}
	return out
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

// publishWebRTCSamples ruft `MetricsPublisher.WebRTCSample` für jedes
// `metrics_sampled`-Event mit reservierten webrtc.*-Keys auf
// (spec/telemetry-model.md). Die
// Wire-Validierung in `validateReservedEventMeta` hat zu diesem
// Zeitpunkt schon sichergestellt, dass alle Pflichtfelder typkorrekt
// und in der Allowlist sind; fehlende optionale Felder lassen die
// Snapshot-Konstruktion auf 0 zurückfallen.
func (u *RegisterPlaybackEventBatchUseCase) publishWebRTCSamples(events []domain.PlaybackEvent) {
	for _, e := range events {
		if e.EventName != "metrics_sampled" {
			continue
		}
		snapshot, ok := buildWebRTCSampleSnapshot(e)
		if !ok {
			continue
		}
		u.metrics.WebRTCSample(snapshot)
	}
}

// buildWebRTCSampleSnapshot extrahiert die Sample-Daten aus einem
// validierten `metrics_sampled`-Event. Liefert (snapshot, false), wenn
// das Event keine WebRTC-Samples enthält oder Pflichtfelder fehlen.
func buildWebRTCSampleSnapshot(e domain.PlaybackEvent) (driven.WebRTCSampleSnapshot, bool) {
	runID, ok := e.Meta[metaKeyWebRTCRunID].(string)
	if !ok || runID == "" {
		return driven.WebRTCSampleSnapshot{}, false
	}
	connectionState, ok := e.Meta[metaKeyWebRTCConnectionState].(string)
	if !ok {
		return driven.WebRTCSampleSnapshot{}, false
	}
	iceState, ok := e.Meta[metaKeyWebRTCIceState].(string)
	if !ok {
		return driven.WebRTCSampleSnapshot{}, false
	}
	dtlsState, ok := e.Meta[metaKeyWebRTCDtlsState].(string)
	if !ok {
		return driven.WebRTCSampleSnapshot{}, false
	}
	return driven.WebRTCSampleSnapshot{
		ProjectID:       e.ProjectID,
		SessionID:       e.SessionID,
		RunID:           runID,
		SampleID:        intMeta(e.Meta, metaKeyWebRTCSampleID),
		ConnectionState: connectionState,
		IceState:        iceState,
		DtlsState:       dtlsState,
		PacketsLost:     intMeta(e.Meta, metaKeyWebRTCPacketsLost),
		BytesReceived:   intMeta(e.Meta, metaKeyWebRTCBytesReceived),
		BytesSent:       intMeta(e.Meta, metaKeyWebRTCBytesSent),
	}, true
}

// intMeta liest einen non-negativen Integer-Wert aus dem Meta-Slot.
// Validation hat negative/non-int-Werte schon ausgesondert; hier
// reicht ein bester Pfad pro Wire-Repräsentation.
func intMeta(meta domain.EventMeta, key string) int64 {
	switch v := meta[key].(type) {
	case int64:
		return v
	case float64:
		return int64(v)
	case jsonNumber:
		f, err := v.Float64()
		if err != nil {
			return 0
		}
		return int64(f)
	default:
		return 0
	}
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

// applySessionSampleRate verarbeitet `meta.session_sample_rate` pro
// distinct (project_id, session_id)-Partition im Batch (
// §6 / R-10). Per-Session-Logik:
//
//  - Wert fehlt oder == 1.0 → no-op (Session bleibt voll-gesampelt-
//  Default in der DB).
//  - Wert < 1.0 → normalisiere via `domain.SampleRatePPMFromFloat`.
//  - Server-State == Default: persistiert via Immutability-Set.
//  - Server-State != Default und Abweichung außerhalb
//  `SampleRateDriftTolerancePPM`: incrementiert
//  `mtrace_sample_rate_drift_total{project_id}`; existing-Wert
//  wird NICHT überschrieben.
//
// Der erste session_sample_rate-Wert im Batch pro Session entscheidet —
// spätere Events derselben Session im selben Batch werden ignoriert
// (Set-Logik ist immutable; alle Events sollen ohnehin denselben Wert
// tragen).
func (u *RegisterPlaybackEventBatchUseCase) applySessionSampleRate(ctx context.Context, events []domain.PlaybackEvent) error {
	seen := make(map[string]struct{}, len(events))
	for _, e := range events {
		key := e.ProjectID + "/" + e.SessionID
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		raw, ok := e.Meta[metaKeySessionSampleRate]
		if !ok {
			continue
		}
		var f float64
		switch v := raw.(type) {
		case float64:
			f = v
		case int64:
			f = float64(v)
		case int:
			f = float64(v)
		default:
			continue
		}
		ppm, err := domain.SampleRatePPMFromFloat(f)
		if err != nil || ppm == domain.SampleRateFull {
			continue
		}
		existing, applied, err := u.sessions.SetSessionSampleRatePPMIfDefault(ctx, e.ProjectID, e.SessionID, ppm)
		if err != nil {
			return fmt.Errorf("apply sample_rate: %w", err)
		}
		if !applied && absInt(existing-ppm) > SampleRateDriftTolerancePPM {
			u.metrics.SampleRateDrift(e.ProjectID)
		}
	}
	return nil
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Compile-time check: the use case implements the inbound port.
var _ driving.PlaybackEventInbound = (*RegisterPlaybackEventBatchUseCase)(nil)
