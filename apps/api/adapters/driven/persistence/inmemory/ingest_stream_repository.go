package inmemory

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sort"
	"sync"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// IngestStreamRepository liefert die In-Memory-Variante der
// Ingest-Control-Persistenz (RAK-66/RAK-67).
// Geteilt mit dem SQLite-Adapter über einen gemeinsamen
// Contract-Test-Body; Daten überleben keinen Restart.
//
// Sicherheitsprofil:
//  - `keysByStream` speichert ausschließlich `domain.StreamKey`
//  (Hash + Fingerprint + Lifecycle-Felder). Klartext-Werte
//  gehen nie in dieses Repository.
//  - Cross-Project-Lookups liefern `domain.ErrIngestStreamNotFound`
//  ohne Hinweis auf Existenz.
type IngestStreamRepository struct {
	mu sync.Mutex

	// Project-skopiert: streams[projectID][streamID].
	streams map[string]map[string]ingestStreamRecord
	// Pro Stream: alle jemals erzeugten Stream-Keys (aktuell aktiv
	// + Historie). Das aktive Element ist das mit `deactivatedAt
	// == nil`. SQLite kann das identisch über `stream_keys` mit
	// `deactivated_at IS NULL` modellieren.
	keysByStream map[streamRef][]storedKey

	endpoints      map[string]domain.IngestEndpoint
	targets        map[string]domain.MediaServerTarget
	routingRules   map[streamRef]domain.RoutingRule
	lifecycleLog   []domain.StreamLifecycleEvent
}

type ingestStreamRecord struct {
	stream domain.IngestStream
}

type storedKey struct {
	keyID         string
	hash          string
	fingerprint   string
	createdAt     time.Time
	deactivatedAt *time.Time
}

type streamRef struct {
	projectID string
	streamID  string
}

// NewIngestStreamRepository erzeugt ein leeres Repository. Endpoints
// und Targets können über `SeedEndpoint` / `SeedTarget` populated
// werden; das spiegelt das SQLite-Verhalten, in dem die beiden
// Tabellen vor Stream-Anlage bekannt sein müssen.
func NewIngestStreamRepository() *IngestStreamRepository {
	return &IngestStreamRepository{
		streams:      map[string]map[string]ingestStreamRecord{},
		keysByStream: map[streamRef][]storedKey{},
		endpoints:    map[string]domain.IngestEndpoint{},
		targets:      map[string]domain.MediaServerTarget{},
		routingRules: map[streamRef]domain.RoutingRule{},
	}
}

// SeedEndpoint registriert einen Endpoint. Tests und Lab-Bootstrap
// nutzen das, um die Endpoint-Allowlist analog zum SQLite-INSERT
// vorzubereiten.
func (r *IngestStreamRepository) SeedEndpoint(endpoint domain.IngestEndpoint) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.endpoints[endpoint.ID] = endpoint
}

// SeedTarget registriert ein MediaServerTarget.
func (r *IngestStreamRepository) SeedTarget(target domain.MediaServerTarget) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.targets[target.ID] = target
}

// CreateStream legt Stream + Routing-Regel + initialen Key an.
func (r *IngestStreamRepository) CreateStream(_ context.Context, input driven.CreateStreamInput) (*domain.IngestStream, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.endpoints[input.EndpointID]; !ok {
		return nil, domain.ErrIngestEndpointNotFound
	}
	if _, ok := r.targets[input.TargetID]; !ok {
		return nil, domain.ErrIngestTargetNotFound
	}
	projectStreams, ok := r.streams[input.ProjectID]
	if !ok {
		projectStreams = map[string]ingestStreamRecord{}
		r.streams[input.ProjectID] = projectStreams
	}
	for _, rec := range projectStreams {
		if rec.stream.DisplayName == input.DisplayName && rec.stream.Status != domain.IngestStreamStatusEnded {
			return nil, domain.ErrIngestStreamNameConflict
		}
	}
	streamID := newID("ing_")
	ruleID := newID("route_")
	stream := domain.IngestStream{
		ID:            streamID,
		ProjectID:     input.ProjectID,
		DisplayName:   input.DisplayName,
		Protocol:      input.Protocol,
		EndpointID:    input.EndpointID,
		TargetID:      input.TargetID,
		RoutingRuleID: ruleID,
		Status:        domain.IngestStreamStatusReady,
		Key:           input.InitialKey,
		CreatedAt:     input.CreatedAt,
		UpdatedAt:     input.CreatedAt,
	}
	projectStreams[streamID] = ingestStreamRecord{stream: stream}
	ref := streamRef{projectID: input.ProjectID, streamID: streamID}
	r.routingRules[ref] = domain.RoutingRule{
		ID:       ruleID,
		StreamID: streamID,
		TargetID: input.TargetID,
		Mode:     domain.RoutingRuleModeSingle,
		Enabled:  true,
	}
	r.keysByStream[ref] = []storedKey{{
		keyID:       newID("key_"),
		hash:        input.InitialKey.Hash,
		fingerprint: input.InitialKey.Fingerprint,
		createdAt:   input.InitialKey.CreatedAt,
	}}
	return &stream, nil
}

// GetStreamByID lädt einen Stream mit aktivem Key. Liefert
// `domain.ErrIngestStreamNotFound`, wenn der Stream nicht im
// angefragten Project existiert.
func (r *IngestStreamRepository) GetStreamByID(_ context.Context, projectID, streamID string) (*domain.IngestStream, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	stream, ok := r.lookupLocked(projectID, streamID)
	if !ok {
		return nil, domain.ErrIngestStreamNotFound
	}
	out := stream
	return &out, nil
}

// ListByProject liefert alle Streams eines Projects mit aktivem Key,
// sortiert nach `created_at desc, stream_id asc`.
func (r *IngestStreamRepository) ListByProject(_ context.Context, projectID string) ([]domain.IngestStream, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	projectStreams, ok := r.streams[projectID]
	if !ok {
		return nil, nil
	}
	out := make([]domain.IngestStream, 0, len(projectStreams))
	for id := range projectStreams {
		stream, _ := r.lookupLocked(projectID, id)
		out = append(out, stream)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out, nil
}

// RotateKey deaktiviert den bisher aktiven Stream-Key und legt einen
// neuen an. Aktualisiert `updated_at` auf den `CreatedAt`-Wert des
// neuen Keys.
func (r *IngestStreamRepository) RotateKey(_ context.Context, projectID, streamID string, newKey domain.StreamKey) (*domain.IngestStream, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	projectStreams, ok := r.streams[projectID]
	if !ok {
		return nil, domain.ErrIngestStreamNotFound
	}
	rec, ok := projectStreams[streamID]
	if !ok {
		return nil, domain.ErrIngestStreamNotFound
	}
	ref := streamRef{projectID: projectID, streamID: streamID}
	now := newKey.CreatedAt
	keys := r.keysByStream[ref]
	for i := range keys {
		if keys[i].deactivatedAt == nil {
			t := now
			keys[i].deactivatedAt = &t
		}
	}
	keys = append(keys, storedKey{
		keyID:       newID("key_"),
		hash:        newKey.Hash,
		fingerprint: newKey.Fingerprint,
		createdAt:   newKey.CreatedAt,
	})
	r.keysByStream[ref] = keys
	rec.stream.Key = newKey
	rec.stream.UpdatedAt = newKey.CreatedAt
	projectStreams[streamID] = rec
	out := rec.stream
	return &out, nil
}

// FindActiveStreamKey liefert den aktiven Key des Streams. Liefert
// `domain.ErrIngestKeyInvalid`, wenn kein aktiver Key existiert
// (z. B. nach Rotation, bevor der neue Key persistiert ist — in der
// In-Memory-Variante praktisch nicht erreichbar).
func (r *IngestStreamRepository) FindActiveStreamKey(_ context.Context, projectID, streamID string) (domain.StreamKey, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.lookupLocked(projectID, streamID); !ok {
		return domain.StreamKey{}, domain.ErrIngestStreamNotFound
	}
	ref := streamRef{projectID: projectID, streamID: streamID}
	for _, k := range r.keysByStream[ref] {
		if k.deactivatedAt == nil {
			return domain.StreamKey{
				Hash:        k.hash,
				Fingerprint: k.fingerprint,
				CreatedAt:   k.createdAt,
			}, nil
		}
	}
	return domain.StreamKey{}, domain.ErrIngestKeyInvalid
}

// GetEndpointByID liefert einen Endpoint aus der Seed-Allowlist.
func (r *IngestStreamRepository) GetEndpointByID(_ context.Context, endpointID string) (*domain.IngestEndpoint, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	endpoint, ok := r.endpoints[endpointID]
	if !ok {
		return nil, domain.ErrIngestEndpointNotFound
	}
	out := endpoint
	return &out, nil
}

// GetTargetByID liefert ein MediaServerTarget aus der Seed-Allowlist.
func (r *IngestStreamRepository) GetTargetByID(_ context.Context, targetID string) (*domain.MediaServerTarget, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	target, ok := r.targets[targetID]
	if !ok {
		return nil, domain.ErrIngestTargetNotFound
	}
	out := target
	return &out, nil
}

// GetRoutingRuleByID liefert die Routing-Regel des Streams.
func (r *IngestStreamRepository) GetRoutingRuleByID(_ context.Context, projectID, streamID string) (*domain.RoutingRule, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.lookupLocked(projectID, streamID); !ok {
		return nil, domain.ErrIngestStreamNotFound
	}
	rule, ok := r.routingRules[streamRef{projectID: projectID, streamID: streamID}]
	if !ok {
		return nil, domain.ErrIngestStreamNotFound
	}
	out := rule
	return &out, nil
}

// AppendLifecycleEvent persistiert ein Lifecycle-Event im append-only
// Log. Klartext-Keys werden nicht gespeichert; das Event darf nur
// einen Fingerprint tragen.
func (r *IngestStreamRepository) AppendLifecycleEvent(_ context.Context, event domain.StreamLifecycleEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.lookupLocked(event.ProjectID, event.StreamID); !ok {
		return domain.ErrIngestStreamNotFound
	}
	r.lifecycleLog = append(r.lifecycleLog, event)
	return nil
}

// lookupLocked liefert den Stream inkl. aktivem Key. Caller muss den
// Mutex halten.
func (r *IngestStreamRepository) lookupLocked(projectID, streamID string) (domain.IngestStream, bool) {
	projectStreams, ok := r.streams[projectID]
	if !ok {
		return domain.IngestStream{}, false
	}
	rec, ok := projectStreams[streamID]
	if !ok {
		return domain.IngestStream{}, false
	}
	stream := rec.stream
	for _, k := range r.keysByStream[streamRef{projectID: projectID, streamID: streamID}] {
		if k.deactivatedAt == nil {
			stream.Key = domain.StreamKey{
				Hash:        k.hash,
				Fingerprint: k.fingerprint,
				CreatedAt:   k.createdAt,
			}
			break
		}
	}
	return stream, true
}

// LifecycleSnapshot liefert eine Kopie des append-only Lifecycle-Logs.
// Test-Hilfe; nicht Teil des Driven-Ports.
func (r *IngestStreamRepository) LifecycleSnapshot() []domain.StreamLifecycleEvent {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]domain.StreamLifecycleEvent, len(r.lifecycleLog))
	copy(out, r.lifecycleLog)
	return out
}

func newID(prefix string) string {
	var raw [12]byte
	if _, err := rand.Read(raw[:]); err != nil {
		panic("inmemory: ingest-control id rng failed: " + err.Error())
	}
	return prefix + hex.EncodeToString(raw[:])
}
