package application

import (
	"context"
	"errors"
	"fmt"
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
	repo  driven.IngestStreamRepository
	clock Clock
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
		return driving.CreateStreamResult{}, errors.New("ingest: display_name must not be empty")
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

// RecordLifecycleEvent persistiert ein Lifecycle-Event in T2 als
// Vorbereitung für T4. `key_fingerprint` wird aus dem aktiven Key
// abgeleitet — Klartext-Werte landen niemals im Eventlog.
func (s *IngestControlService) RecordLifecycleEvent(ctx context.Context, projectID, streamID string, kind domain.StreamLifecycleEventKind, occurredAt time.Time, source domain.StreamLifecycleEventSource) error {
	stream, err := s.repo.GetStreamByID(ctx, projectID, streamID)
	if err != nil {
		return err
	}
	rule, err := s.repo.GetRoutingRuleByID(ctx, projectID, streamID)
	if err != nil {
		return err
	}
	if !rule.Enabled {
		return domain.ErrIngestRoutingRuleDisabled
	}
	return s.repo.AppendLifecycleEvent(ctx, domain.StreamLifecycleEvent{
		Kind:           kind,
		StreamID:       streamID,
		ProjectID:      projectID,
		OccurredAt:     occurredAt.UTC(),
		Source:         source,
		KeyFingerprint: stream.Key.Fingerprint,
	})
}
