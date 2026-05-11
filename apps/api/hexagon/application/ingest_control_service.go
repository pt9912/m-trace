package application

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driving"
)

// IngestControlService implementiert den `driving.IngestControlInbound`-
// Port (`0.11.0` Tranche 2, NF-13 / RAK-65..RAK-67). Ohne HTTP- und
// SQLite-Annahmen — der Use-Case bekommt das Repository über den
// Driven-Port und einen `Clock`-Hook für deterministische Tests.
//
// Sicherheitsprofil:
//   - Klartext-Stream-Keys leben nur transient zwischen
//     `domain.GenerateStreamKey` und der Service-Antwort. Das
//     Repository bekommt ausschließlich `domain.StreamKey` (Hash +
//     Fingerprint).
//   - Cross-Project-Leak-Schutz: jeder Lookup nutzt die im Aufruf
//     übergebene `projectID` (vom HTTP-Adapter aus dem Token
//     resolved); ein Stream eines fremden Projects wird wie
//     nicht-existent behandelt (`domain.ErrIngestStreamNotFound`).
//   - `ValidateKey` liefert `Valid:false` ohne Stream-ID-Hinweis,
//     wenn entweder Stream nicht im Project ist, der Key Format-
//     verstoß hat oder der Hash nicht passt.
type IngestControlService struct {
	repo     driven.IngestStreamRepository
	clock    Clock
	webhooks driven.OutboundWebhookDispatcher
}

// Clock erlaubt deterministische Tests. Default in Produktion:
// `time.Now`; Tests injizieren einen festen Wert.
type Clock func() time.Time

// NewIngestControlService konstruiert den Service. `clock == nil`
// fällt auf `time.Now` zurück, damit produktive Verdrahtung kompakt
// bleibt.
func NewIngestControlService(repo driven.IngestStreamRepository, clock Clock) *IngestControlService {
	if clock == nil {
		clock = time.Now
	}
	return &IngestControlService{repo: repo, clock: clock}
}

// WithOutboundWebhookDispatcher verdrahtet den optionalen
// Outbound-Webhook-Dispatcher (`0.12.5`/RAK-82, R-16). `nil`
// deaktiviert den Pfad — das Service-Verhalten ist dann unverändert
// zum `0.11.0`-Stand (Lifecycle-Events bleiben rein lokal).
//
// Verdrahtung als Option-Setter und nicht im Konstruktor, damit
// existierende Bootstrap-Aufrufe (insb. die `0.11.0`-Tests)
// unverändert weiterlaufen.
func (s *IngestControlService) WithOutboundWebhookDispatcher(d driven.OutboundWebhookDispatcher) *IngestControlService {
	if s == nil {
		return s
	}
	s.webhooks = d
	return s
}

// CreateStream legt einen neuen Stream samt initialem Klartext-Key
// an. Validierungspipeline:
//   1. Project-ID-Konsistenz (Request vs. Token).
//   2. Display-Name nicht leer.
//   3. Protocol in Allowlist (`srt`/`rtmp`).
//   4. Endpoint und Target existieren.
//   5. Repo-Insert (atomar) — bei Constraint-Verstoß
//      `ErrIngestStreamNameConflict`.
func (s *IngestControlService) CreateStream(ctx context.Context, req driving.CreateStreamRequest) (driving.CreateStreamResult, error) {
	if err := domain.ValidateProjectIDConsistency(req.RequestProjectID, req.ResolvedProjectID); err != nil {
		return driving.CreateStreamResult{}, err
	}
	displayName := strings.TrimSpace(req.DisplayName)
	if displayName == "" {
		return driving.CreateStreamResult{}, domain.ErrIngestDisplayNameRequired
	}
	protocol, err := domain.ValidateIngestProtocol(req.Protocol)
	if err != nil {
		return driving.CreateStreamResult{}, err
	}
	if strings.TrimSpace(req.EndpointID) == "" {
		return driving.CreateStreamResult{}, domain.ErrIngestEndpointNotFound
	}
	if strings.TrimSpace(req.TargetID) == "" {
		return driving.CreateStreamResult{}, domain.ErrIngestTargetNotFound
	}

	now := s.clock().UTC()
	material, err := domain.GenerateStreamKey(now)
	if err != nil {
		return driving.CreateStreamResult{}, fmt.Errorf("ingest: generate stream key: %w", err)
	}

	stream, err := s.repo.CreateStream(ctx, driven.CreateStreamInput{
		ProjectID:   req.ResolvedProjectID,
		DisplayName: displayName,
		Protocol:    protocol,
		EndpointID:  strings.TrimSpace(req.EndpointID),
		TargetID:    strings.TrimSpace(req.TargetID),
		InitialKey:  material.ToPersistable(),
		CreatedAt:   now,
	})
	if err != nil {
		return driving.CreateStreamResult{}, err
	}
	return driving.CreateStreamResult{Stream: *stream, Material: material}, nil
}

// ListStreams liefert alle Streams im aufgelösten Project, sortiert
// nach `created_at desc, stream_id asc`. Reicht das Repo-Result
// unverändert durch — der HTTP-Adapter filtert die Klartext-Felder.
func (s *IngestControlService) ListStreams(ctx context.Context, projectID string) ([]domain.IngestStream, error) {
	streams, err := s.repo.ListByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return streams, nil
}

// GetStreamDetail liefert Stream + Endpoint + Target + Routing-Regel
// für `GET /api/ingest/streams/{id}`. Cross-Project-Lookups schlagen
// als `domain.ErrIngestStreamNotFound` durch.
func (s *IngestControlService) GetStreamDetail(ctx context.Context, projectID, streamID string) (driving.StreamDetail, error) {
	stream, err := s.repo.GetStreamByID(ctx, projectID, streamID)
	if err != nil {
		return driving.StreamDetail{}, err
	}
	endpoint, err := s.repo.GetEndpointByID(ctx, stream.EndpointID)
	if err != nil {
		return driving.StreamDetail{}, err
	}
	target, err := s.repo.GetTargetByID(ctx, stream.TargetID)
	if err != nil {
		return driving.StreamDetail{}, err
	}
	rule, err := s.repo.GetRoutingRuleByID(ctx, projectID, streamID)
	if err != nil {
		return driving.StreamDetail{}, err
	}
	return driving.StreamDetail{
		Stream:      *stream,
		Endpoint:    *endpoint,
		Target:      *target,
		RoutingRule: *rule,
	}, nil
}

// RotateKey deaktiviert den aktuell aktiven Stream-Key, erzeugt ein
// frisches CSPRNG-Klartext-Material und liefert es genau einmal an
// den Aufrufer. Persistenz speichert ausschließlich den neuen Hash
// und Fingerprint.
func (s *IngestControlService) RotateKey(ctx context.Context, projectID, streamID string) (driving.RotateKeyResult, error) {
	if _, err := s.repo.GetStreamByID(ctx, projectID, streamID); err != nil {
		return driving.RotateKeyResult{}, err
	}
	now := s.clock().UTC()
	material, err := domain.GenerateStreamKey(now)
	if err != nil {
		return driving.RotateKeyResult{}, fmt.Errorf("ingest: generate rotated key: %w", err)
	}
	stream, err := s.repo.RotateKey(ctx, projectID, streamID, material.ToPersistable())
	if err != nil {
		return driving.RotateKeyResult{}, err
	}
	return driving.RotateKeyResult{Stream: *stream, Material: material}, nil
}

// ValidateKey vergleicht den Klartext-Kandidaten gegen den aktiven
// Stream-Key. Cross-Project-Leak-Schutz: jeder Fehlerpfad (Stream
// nicht im Project, Format-Verstoß, Hash-Mismatch) liefert
// `Valid:false` ohne Stream-ID-Hinweis.
func (s *IngestControlService) ValidateKey(ctx context.Context, projectID, streamID, candidateKey string) (driving.ValidateKeyResult, error) {
	stream, err := s.repo.GetStreamByID(ctx, projectID, streamID)
	if err != nil {
		if errors.Is(err, domain.ErrIngestStreamNotFound) {
			return driving.ValidateKeyResult{Valid: false}, nil
		}
		return driving.ValidateKeyResult{}, err
	}
	activeKey, err := s.repo.FindActiveStreamKey(ctx, projectID, streamID)
	if err != nil {
		if errors.Is(err, domain.ErrIngestKeyInvalid) {
			return driving.ValidateKeyResult{Valid: false}, nil
		}
		return driving.ValidateKeyResult{}, err
	}
	ok, _ := domain.ValidateStreamKey(candidateKey, activeKey)
	if !ok {
		return driving.ValidateKeyResult{Valid: false}, nil
	}
	return driving.ValidateKeyResult{
		Valid:          true,
		StreamID:       stream.ID,
		KeyFingerprint: activeKey.Fingerprint,
	}, nil
}

// GetMediaServerConfig generiert ein MediaMTX-Konfigurations-
// Artefakt aus den aktiven Streams des Projects. `targetID` ist
// optional — leer wählt das Target aus dem ersten Stream; konkret
// gesetzt prüft Existenz im Repository.
//
// Wenn das Project mehrere distinkte Targets nutzt und der Aufrufer
// keinen `target_id`-Filter setzt, wird der Auto-Pick **nicht
// stillschweigend** durchgeführt: der Result trägt einen Warning,
// der das gewählte Target und alle übrigen Target-IDs nennt — die
// HTTP-Antwort reicht den Warning ungekürzt durch.
//
// Sicherheitsprofil: das generierte YAML enthält ausschließlich
// Display-Name (sanitized) und `key_fingerprint`. Klartext-Stream-
// Keys werden niemals serialisiert — die Funktion ruft den
// Generator-Pfad in `mediamtx_config.go`, der Klartext-Werte gar
// nicht zu Gesicht bekommt (Repository liefert nur
// `domain.StreamKey` ohne Klartext).
func (s *IngestControlService) GetMediaServerConfig(ctx context.Context, projectID, targetID string) (driving.MediaServerConfigResult, error) {
	streams, err := s.repo.ListByProject(ctx, projectID)
	if err != nil {
		return driving.MediaServerConfigResult{}, err
	}
	if len(streams) == 0 {
		return driving.MediaServerConfigResult{}, ErrMediaMTXConfigNoStreams
	}
	requestedTargetID := strings.TrimSpace(targetID)
	resolvedTargetID := requestedTargetID
	autoPickWarning := ""
	if resolvedTargetID == "" {
		resolvedTargetID = streams[0].TargetID
		// Multi-Target-Diagnose: alle distinkten Target-IDs sammeln,
		// damit der Warning den Aufrufer nicht im Dunkeln lässt.
		others := distinctOtherTargetIDs(streams, resolvedTargetID)
		if len(others) > 0 {
			autoPickWarning = fmt.Sprintf(
				"multiple target ids in project [%s]; auto-selected %q — pass ?target_id=... to choose explicitly",
				strings.Join(others, ", "), resolvedTargetID,
			)
		}
	}
	target, err := s.repo.GetTargetByID(ctx, resolvedTargetID)
	if err != nil {
		return driving.MediaServerConfigResult{}, err
	}
	if target.Kind != domain.MediaServerKindMediaMTX {
		return driving.MediaServerConfigResult{}, fmt.Errorf("ingest: media-server config supports only mediamtx, got %q", target.Kind)
	}
	// Pro Target nur passende Streams einbeziehen.
	matching := make([]domain.IngestStream, 0, len(streams))
	for _, stream := range streams {
		if stream.TargetID == target.ID {
			matching = append(matching, stream)
		}
	}
	if len(matching) == 0 {
		return driving.MediaServerConfigResult{}, ErrMediaMTXConfigNoStreams
	}
	artifact, err := GenerateMediaMTXConfig(MediaMTXConfigInput{
		Target:  *target,
		Streams: matching,
	})
	if err != nil {
		return driving.MediaServerConfigResult{}, err
	}
	warnings := artifact.Warnings
	if autoPickWarning != "" {
		warnings = append([]string{autoPickWarning}, warnings...)
	}
	return driving.MediaServerConfigResult{
		TargetID:   target.ID,
		Kind:       target.Kind,
		ConfigPath: target.ConfigPath,
		ConfigYAML: artifact.YAML,
		Warnings:   warnings,
	}, nil
}

// distinctOtherTargetIDs liefert alle Target-IDs aus `streams`, die
// nicht dem `selected`-Wert entsprechen, sortiert und dedupliziert.
// Wird ausschließlich für den Multi-Target-Auto-Pick-Warning genutzt.
func distinctOtherTargetIDs(streams []domain.IngestStream, selected string) []string {
	seen := map[string]struct{}{}
	for _, s := range streams {
		if s.TargetID == "" || s.TargetID == selected {
			continue
		}
		seen[s.TargetID] = struct{}{}
	}
	if len(seen) == 0 {
		return nil
	}
	out := make([]string, 0, len(seen))
	for id := range seen {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

// RecordLifecycleEvent persistiert ein Lifecycle-Event (Plan §0.11.0
// Tranche 4 / RAK-69). Der `event_id`-Wert wird hier serverseitig
// erzeugt und mit dem Event in den Adapter zurückgereicht — der
// HTTP-Hook-Handler echo't ihn als Acknowledgement. `KeyFingerprint`
// wird aus dem aktiven Key abgeleitet; Klartext-Werte landen niemals
// im Eventlog.
//
// Source-Werte werden hier auf die Domain-Allowlist gefiltert. Der
// HTTP-Adapter prüft Längenlimits **bevor** er hierher kommt; der
// Service prüft trotzdem nochmal, weil der Driving-Port auch von
// CLI/Tests direkt aufgerufen werden kann.
func (s *IngestControlService) RecordLifecycleEvent(ctx context.Context, req driving.LifecycleHookRequest) (driving.LifecycleEventResult, error) {
	if req.ObservedAt.IsZero() {
		return driving.LifecycleEventResult{}, domain.ErrIngestLifecycleObservedAtRequired
	}
	if !req.Source.IsKnown() {
		return driving.LifecycleEventResult{}, domain.ErrIngestLifecycleSourceUnknown
	}
	if len(req.ConnectionID) > domain.MaxLifecycleStringField || len(req.Reason) > domain.MaxLifecycleStringField {
		return driving.LifecycleEventResult{}, domain.ErrIngestLifecycleFieldTooLong
	}
	stream, err := s.repo.GetStreamByID(ctx, req.ResolvedProjectID, req.StreamID)
	if err != nil {
		return driving.LifecycleEventResult{}, err
	}
	rule, err := s.repo.GetRoutingRuleByID(ctx, req.ResolvedProjectID, req.StreamID)
	if err != nil {
		return driving.LifecycleEventResult{}, err
	}
	if !rule.Enabled {
		return driving.LifecycleEventResult{}, domain.ErrIngestRoutingRuleDisabled
	}
	eventID, err := newLifecycleEventID()
	if err != nil {
		return driving.LifecycleEventResult{}, err
	}
	observed := req.ObservedAt.UTC()
	if err := s.repo.AppendLifecycleEvent(ctx, domain.StreamLifecycleEvent{
		EventID:        eventID,
		Kind:           req.Kind,
		StreamID:       req.StreamID,
		ProjectID:      req.ResolvedProjectID,
		OccurredAt:     observed,
		Source:         req.Source,
		KeyFingerprint: stream.Key.Fingerprint,
		ConnectionID:   strings.TrimSpace(req.ConnectionID),
		Reason:         strings.TrimSpace(req.Reason),
	}); err != nil {
		return driving.LifecycleEventResult{}, err
	}
	// `0.12.5`/RAK-82: Outbound-Webhook-Dispatch (R-16). Best-Effort,
	// im Hintergrund — der Lifecycle-Pfad failed nicht, wenn ein
	// externer Konsument nicht erreichbar ist. Adapter handlet Retry
	// und Dead-Letter selbst.
	s.dispatchOutboundWebhook(ctx, driven.OutboundWebhookEvent{
		EventID:      eventID,
		Kind:         req.Kind,
		ProjectID:    req.ResolvedProjectID,
		StreamID:     req.StreamID,
		ObservedAt:   observed.Format(time.RFC3339Nano),
		Source:       req.Source,
		ConnectionID: strings.TrimSpace(req.ConnectionID),
		Reason:       strings.TrimSpace(req.Reason),
	})
	return driving.LifecycleEventResult{
		EventID:    eventID,
		StreamID:   req.StreamID,
		Kind:       req.Kind,
		ObservedAt: observed,
	}, nil
}

// dispatchOutboundWebhook ist die zentrale Stelle, an der ein
// Lifecycle-Event an einen optional konfigurierten Webhook-Adapter
// gereicht wird. `nil`-Dispatcher → no-op; ein Adapter-Fehler
// (z. B. Dead-Letter) bleibt im Adapter-Log und blockt den Caller
// nicht.
//
// Aktuell synchron im selben `ctx`, damit der Test-Pfad
// deterministisch bleibt. Eine asynchrone Variante (Goroutine plus
// In-Memory-Queue) ist Folge-Item — sobald der Lifecycle-Pfad an
// einen Hot-Path mit ms-Budget gekoppelt wird.
func (s *IngestControlService) dispatchOutboundWebhook(ctx context.Context, event driven.OutboundWebhookEvent) {
	if s == nil || s.webhooks == nil {
		return
	}
	_ = s.webhooks.Dispatch(ctx, event)
}

// newLifecycleEventID erzeugt einen opaken Event-Identifier mit
// Prefix `evt_` und 12 Byte crypto/rand-Hex (192 Bit Entropie).
// Kollisionsraum genügt für Lab-Smokes; CSPRNG verhindert
// Vorhersagbarkeit, falls Hooks später öffentlich exponiert werden.
func newLifecycleEventID() (string, error) {
	var raw [12]byte
	if _, err := cryptorand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("ingest: lifecycle event_id rand: %w", err)
	}
	return "evt_" + hex.EncodeToString(raw[:]), nil
}
