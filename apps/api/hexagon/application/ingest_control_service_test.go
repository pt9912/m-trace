package application_test

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/application"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driving"
)

// fakeIngestRepo ist eine in-test In-Memory-Implementation des
// `driven.IngestStreamRepository`-Ports. Das `application`-Paket darf
// laut `apps/api/.golangci.yml` `application-no-adapters`-Regel keine
// Adapter-Implementierungen importieren — auch nicht für Tests; daher
// die kompakte Doppel-Implementierung. Verhalten spiegelt die echte
// `inmemory.IngestStreamRepository` ein zu eins (Cross-Project-Leak-
// Schutz, Duplikat-Erkennung, Key-Rotation).
type fakeIngestRepo struct {
	mu          sync.Mutex
	streams     map[string]map[string]domain.IngestStream
	keys        map[string][]storedKeyEntry // key: project/stream
	endpoints   map[string]domain.IngestEndpoint
	targets     map[string]domain.MediaServerTarget
	rules       map[string]domain.RoutingRule // key: project/stream
	lifecycle   []domain.StreamLifecycleEvent
	idCounter   int
}

type storedKeyEntry struct {
	hash          string
	fingerprint   string
	createdAt     time.Time
	deactivatedAt *time.Time
}

func newFakeIngestRepo() *fakeIngestRepo {
	return &fakeIngestRepo{
		streams:   map[string]map[string]domain.IngestStream{},
		keys:      map[string][]storedKeyEntry{},
		endpoints: map[string]domain.IngestEndpoint{},
		targets:   map[string]domain.MediaServerTarget{},
		rules:     map[string]domain.RoutingRule{},
	}
}

func (f *fakeIngestRepo) seedEndpoint(e domain.IngestEndpoint) { f.endpoints[e.ID] = e }
func (f *fakeIngestRepo) seedTarget(t domain.MediaServerTarget) { f.targets[t.ID] = t }

func (f *fakeIngestRepo) nextID(prefix string) string {
	f.idCounter++
	return prefix + "test-" + time.Now().Format("150405") + "-" + string(rune('a'+(f.idCounter%26)))
}

func keyRef(projectID, streamID string) string { return projectID + "/" + streamID }

func (f *fakeIngestRepo) CreateStream(_ context.Context, in driven.CreateStreamInput) (*domain.IngestStream, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.endpoints[in.EndpointID]; !ok {
		return nil, domain.ErrIngestEndpointNotFound
	}
	if _, ok := f.targets[in.TargetID]; !ok {
		return nil, domain.ErrIngestTargetNotFound
	}
	ps, ok := f.streams[in.ProjectID]
	if !ok {
		ps = map[string]domain.IngestStream{}
		f.streams[in.ProjectID] = ps
	}
	for _, s := range ps {
		if s.DisplayName == in.DisplayName && s.Status != domain.IngestStreamStatusEnded {
			return nil, domain.ErrIngestStreamNameConflict
		}
	}
	streamID := f.nextID("ing_")
	ruleID := f.nextID("route_")
	stream := domain.IngestStream{
		ID:            streamID,
		ProjectID:     in.ProjectID,
		DisplayName:   in.DisplayName,
		Protocol:      in.Protocol,
		EndpointID:    in.EndpointID,
		TargetID:      in.TargetID,
		RoutingRuleID: ruleID,
		Status:        domain.IngestStreamStatusReady,
		Key:           in.InitialKey,
		CreatedAt:     in.CreatedAt,
		UpdatedAt:     in.CreatedAt,
	}
	ps[streamID] = stream
	f.keys[keyRef(in.ProjectID, streamID)] = []storedKeyEntry{{
		hash: in.InitialKey.Hash, fingerprint: in.InitialKey.Fingerprint, createdAt: in.InitialKey.CreatedAt,
	}}
	f.rules[keyRef(in.ProjectID, streamID)] = domain.RoutingRule{
		ID: ruleID, StreamID: streamID, TargetID: in.TargetID, Mode: domain.RoutingRuleModeSingle, Enabled: true,
	}
	return &stream, nil
}

func (f *fakeIngestRepo) GetStreamByID(_ context.Context, projectID, streamID string) (*domain.IngestStream, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	ps, ok := f.streams[projectID]
	if !ok {
		return nil, domain.ErrIngestStreamNotFound
	}
	s, ok := ps[streamID]
	if !ok {
		return nil, domain.ErrIngestStreamNotFound
	}
	for _, k := range f.keys[keyRef(projectID, streamID)] {
		if k.deactivatedAt == nil {
			s.Key = domain.StreamKey{Hash: k.hash, Fingerprint: k.fingerprint, CreatedAt: k.createdAt}
			break
		}
	}
	return &s, nil
}

func (f *fakeIngestRepo) ListByProject(_ context.Context, projectID string) ([]domain.IngestStream, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	ps := f.streams[projectID]
	out := make([]domain.IngestStream, 0, len(ps))
	for id := range ps {
		s := ps[id]
		for _, k := range f.keys[keyRef(projectID, id)] {
			if k.deactivatedAt == nil {
				s.Key = domain.StreamKey{Hash: k.hash, Fingerprint: k.fingerprint, CreatedAt: k.createdAt}
				break
			}
		}
		out = append(out, s)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out, nil
}

func (f *fakeIngestRepo) RotateKey(_ context.Context, projectID, streamID string, newKey domain.StreamKey) (*domain.IngestStream, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	ps, ok := f.streams[projectID]
	if !ok {
		return nil, domain.ErrIngestStreamNotFound
	}
	s, ok := ps[streamID]
	if !ok {
		return nil, domain.ErrIngestStreamNotFound
	}
	now := newKey.CreatedAt
	keys := f.keys[keyRef(projectID, streamID)]
	for i := range keys {
		if keys[i].deactivatedAt == nil {
			t := now
			keys[i].deactivatedAt = &t
		}
	}
	keys = append(keys, storedKeyEntry{hash: newKey.Hash, fingerprint: newKey.Fingerprint, createdAt: newKey.CreatedAt})
	f.keys[keyRef(projectID, streamID)] = keys
	s.Key = newKey
	s.UpdatedAt = newKey.CreatedAt
	ps[streamID] = s
	return &s, nil
}

func (f *fakeIngestRepo) FindActiveStreamKey(_ context.Context, projectID, streamID string) (domain.StreamKey, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.streams[projectID][streamID]; !ok {
		return domain.StreamKey{}, domain.ErrIngestStreamNotFound
	}
	for _, k := range f.keys[keyRef(projectID, streamID)] {
		if k.deactivatedAt == nil {
			return domain.StreamKey{Hash: k.hash, Fingerprint: k.fingerprint, CreatedAt: k.createdAt}, nil
		}
	}
	return domain.StreamKey{}, domain.ErrIngestKeyInvalid
}

func (f *fakeIngestRepo) GetEndpointByID(_ context.Context, endpointID string) (*domain.IngestEndpoint, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	e, ok := f.endpoints[endpointID]
	if !ok {
		return nil, domain.ErrIngestEndpointNotFound
	}
	return &e, nil
}

func (f *fakeIngestRepo) GetTargetByID(_ context.Context, targetID string) (*domain.MediaServerTarget, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	t, ok := f.targets[targetID]
	if !ok {
		return nil, domain.ErrIngestTargetNotFound
	}
	return &t, nil
}

func (f *fakeIngestRepo) GetRoutingRuleByID(_ context.Context, projectID, streamID string) (*domain.RoutingRule, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.streams[projectID][streamID]; !ok {
		return nil, domain.ErrIngestStreamNotFound
	}
	rule, ok := f.rules[keyRef(projectID, streamID)]
	if !ok {
		return nil, domain.ErrIngestStreamNotFound
	}
	return &rule, nil
}

func (f *fakeIngestRepo) AppendLifecycleEvent(_ context.Context, event domain.StreamLifecycleEvent) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.streams[event.ProjectID][event.StreamID]; !ok {
		return domain.ErrIngestStreamNotFound
	}
	f.lifecycle = append(f.lifecycle, event)
	return nil
}

func (f *fakeIngestRepo) lifecycleSnapshot() []domain.StreamLifecycleEvent {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]domain.StreamLifecycleEvent, len(f.lifecycle))
	copy(out, f.lifecycle)
	return out
}

// `0.11.0` Tranche 2 / RAK-65..RAK-67 / RAK-69 — Use-Case-Tests
// gegen den InMemory-Repo. Cross-Project-Isolation, Duplicate-Name,
// Missing-Endpoint, Disabled-Routing, Validate-Pfad ohne
// Cross-Project-Leak.

const (
	testProjectA = "project-a"
	testProjectB = "project-b"
	testEndpoint = "ep-srt"
	testTarget   = "tgt-mediamtx"
)

func newSeededService(t *testing.T) (*application.IngestControlService, *fakeIngestRepo) {
	t.Helper()
	repo := newFakeIngestRepo()
	repo.seedEndpoint(domain.IngestEndpoint{
		ID:            testEndpoint,
		Protocol:      domain.IngestProtocolSRT,
		ListenHost:    "127.0.0.1",
		ListenPort:    8890,
		PathTemplate:  "publish:{stream_path}",
		LabStack:      "mtrace-srt",
		PublicURLHint: "srt://localhost:8890",
	})
	repo.seedTarget(domain.MediaServerTarget{
		ID:             testTarget,
		Kind:           domain.MediaServerKindMediaMTX,
		ConfigPath:     "examples/ingest-control/mediamtx.generated.yml",
		HLSURLTemplate: "http://localhost:8889/{stream_path}/index.m3u8",
		ControlAPIURL:  "",
	})
	return application.NewIngestControlService(repo, fixedClock(2026, 5, 9)), repo
}

func fixedClock(year int, month time.Month, day int) application.Clock {
	t := time.Date(year, month, day, 10, 0, 0, 0, time.UTC)
	return func() time.Time { return t }
}

func TestIngestControlService_CreateStream_HappyPath(t *testing.T) {
	t.Parallel()
	svc, _ := newSeededService(t)
	result, err := svc.CreateStream(context.Background(), driving.CreateStreamRequest{
		ResolvedProjectID: testProjectA,
		DisplayName:       "Lab SRT",
		Protocol:          "srt",
		EndpointID:        testEndpoint,
		TargetID:          testTarget,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if result.Stream.ProjectID != testProjectA {
		t.Errorf("project_id: want %q, got %q", testProjectA, result.Stream.ProjectID)
	}
	if result.Stream.Status != domain.IngestStreamStatusReady {
		t.Errorf("status: want ready, got %q", result.Stream.Status)
	}
	if result.Material.Value == "" {
		t.Errorf("klartext-key must be present in material output")
	}
	if result.Stream.Key.Hash != result.Material.Hash {
		t.Errorf("hash mismatch material vs persisted stream")
	}
}

func TestIngestControlService_CreateStream_RejectsUnknownProtocol(t *testing.T) {
	t.Parallel()
	svc, _ := newSeededService(t)
	_, err := svc.CreateStream(context.Background(), driving.CreateStreamRequest{
		ResolvedProjectID: testProjectA,
		DisplayName:       "Bad",
		Protocol:          "webrtc",
		EndpointID:        testEndpoint,
		TargetID:          testTarget,
	})
	if !errors.Is(err, domain.ErrIngestProtocolUnknown) {
		t.Errorf("err: want ErrIngestProtocolUnknown, got %v", err)
	}
}

func TestIngestControlService_CreateStream_RejectsProjectIDMismatch(t *testing.T) {
	t.Parallel()
	svc, _ := newSeededService(t)
	_, err := svc.CreateStream(context.Background(), driving.CreateStreamRequest{
		ResolvedProjectID: testProjectA,
		RequestProjectID:  testProjectB,
		DisplayName:       "Lab",
		Protocol:          "srt",
		EndpointID:        testEndpoint,
		TargetID:          testTarget,
	})
	if !errors.Is(err, domain.ErrIngestProjectIDMismatch) {
		t.Errorf("err: want ErrIngestProjectIDMismatch, got %v", err)
	}
}

func TestIngestControlService_CreateStream_RejectsEmptyDisplayName(t *testing.T) {
	t.Parallel()
	svc, _ := newSeededService(t)
	_, err := svc.CreateStream(context.Background(), driving.CreateStreamRequest{
		ResolvedProjectID: testProjectA,
		DisplayName:       "  ",
		Protocol:          "srt",
		EndpointID:        testEndpoint,
		TargetID:          testTarget,
	})
	if !errors.Is(err, domain.ErrIngestDisplayNameRequired) {
		t.Errorf("err: want ErrIngestDisplayNameRequired, got %v", err)
	}
}

func TestIngestControlService_CreateStream_DuplicateActiveName(t *testing.T) {
	t.Parallel()
	svc, _ := newSeededService(t)
	for i := 0; i < 2; i++ {
		_, err := svc.CreateStream(context.Background(), driving.CreateStreamRequest{
			ResolvedProjectID: testProjectA,
			DisplayName:       "Lab SRT",
			Protocol:          "srt",
			EndpointID:        testEndpoint,
			TargetID:          testTarget,
		})
		if i == 0 && err != nil {
			t.Fatalf("first create failed: %v", err)
		}
		if i == 1 && !errors.Is(err, domain.ErrIngestStreamNameConflict) {
			t.Errorf("second create: want ErrIngestStreamNameConflict, got %v", err)
		}
	}
}

func TestIngestControlService_CreateStream_RejectsMissingEndpoint(t *testing.T) {
	t.Parallel()
	svc, _ := newSeededService(t)
	_, err := svc.CreateStream(context.Background(), driving.CreateStreamRequest{
		ResolvedProjectID: testProjectA,
		DisplayName:       "Lab",
		Protocol:          "srt",
		EndpointID:        "no-such-endpoint",
		TargetID:          testTarget,
	})
	if !errors.Is(err, domain.ErrIngestEndpointNotFound) {
		t.Errorf("err: want ErrIngestEndpointNotFound, got %v", err)
	}
}

func TestIngestControlService_CrossProjectIsolation(t *testing.T) {
	t.Parallel()
	svc, _ := newSeededService(t)
	created, err := svc.CreateStream(context.Background(), driving.CreateStreamRequest{
		ResolvedProjectID: testProjectA,
		DisplayName:       "Lab",
		Protocol:          "srt",
		EndpointID:        testEndpoint,
		TargetID:          testTarget,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	// Read aus Projekt B muss not-found liefern (kein Cross-Project-Leak).
	if _, err := svc.GetStreamDetail(context.Background(), testProjectB, created.Stream.ID); !errors.Is(err, domain.ErrIngestStreamNotFound) {
		t.Errorf("cross-project read: want ErrIngestStreamNotFound, got %v", err)
	}
	// Validate aus Projekt B liefert Valid:false ohne Stream-ID-Hinweis.
	res, err := svc.ValidateKey(context.Background(), testProjectB, created.Stream.ID, created.Material.Value)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if res.Valid {
		t.Errorf("cross-project validate must be invalid")
	}
	if res.StreamID != "" {
		t.Errorf("cross-project validate must not leak StreamID, got %q", res.StreamID)
	}
}

func TestIngestControlService_RotateKey_DeactivatesOld(t *testing.T) {
	t.Parallel()
	svc, _ := newSeededService(t)
	created, err := svc.CreateStream(context.Background(), driving.CreateStreamRequest{
		ResolvedProjectID: testProjectA,
		DisplayName:       "Lab",
		Protocol:          "srt",
		EndpointID:        testEndpoint,
		TargetID:          testTarget,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	rotated, err := svc.RotateKey(context.Background(), testProjectA, created.Stream.ID)
	if err != nil {
		t.Fatalf("rotate: %v", err)
	}
	if rotated.Material.Hash == created.Material.Hash {
		t.Errorf("rotated key must differ")
	}
	// Alter Klartext darf nicht mehr validieren.
	res, err := svc.ValidateKey(context.Background(), testProjectA, created.Stream.ID, created.Material.Value)
	if err != nil {
		t.Fatalf("validate old: %v", err)
	}
	if res.Valid {
		t.Errorf("rotated key must invalidate the old klartext")
	}
	// Neuer Klartext muss validieren.
	res, err = svc.ValidateKey(context.Background(), testProjectA, created.Stream.ID, rotated.Material.Value)
	if err != nil {
		t.Fatalf("validate new: %v", err)
	}
	if !res.Valid {
		t.Errorf("rotated key must validate")
	}
}

func TestIngestControlService_ValidateKey_RejectsMalformed(t *testing.T) {
	t.Parallel()
	svc, _ := newSeededService(t)
	created, err := svc.CreateStream(context.Background(), driving.CreateStreamRequest{
		ResolvedProjectID: testProjectA,
		DisplayName:       "Lab",
		Protocol:          "srt",
		EndpointID:        testEndpoint,
		TargetID:          testTarget,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	for _, candidate := range []string{"", "not-a-key", "different_prefix_xyz"} {
		res, err := svc.ValidateKey(context.Background(), testProjectA, created.Stream.ID, candidate)
		if err != nil {
			t.Fatalf("validate %q: %v", candidate, err)
		}
		if res.Valid {
			t.Errorf("candidate %q must be invalid", candidate)
		}
	}
}

func TestIngestControlService_ListStreams_FiltersByProject(t *testing.T) {
	t.Parallel()
	svc, _ := newSeededService(t)
	for i := 0; i < 2; i++ {
		_, _ = svc.CreateStream(context.Background(), driving.CreateStreamRequest{
			ResolvedProjectID: testProjectA,
			DisplayName:       []string{"a", "b"}[i],
			Protocol:          "srt",
			EndpointID:        testEndpoint,
			TargetID:          testTarget,
		})
	}
	_, _ = svc.CreateStream(context.Background(), driving.CreateStreamRequest{
		ResolvedProjectID: testProjectB,
		DisplayName:       "other",
		Protocol:          "srt",
		EndpointID:        testEndpoint,
		TargetID:          testTarget,
	})
	streams, err := svc.ListStreams(context.Background(), testProjectA)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(streams) != 2 {
		t.Errorf("project A: want 2 streams, got %d", len(streams))
	}
	for _, s := range streams {
		if s.ProjectID != testProjectA {
			t.Errorf("cross-project leak: %q", s.ProjectID)
		}
	}
}

func TestIngestControlService_GetStreamDetail_HappyPath(t *testing.T) {
	t.Parallel()
	svc, _ := newSeededService(t)
	created, err := svc.CreateStream(context.Background(), driving.CreateStreamRequest{
		ResolvedProjectID: testProjectA,
		DisplayName:       "Lab",
		Protocol:          "srt",
		EndpointID:        testEndpoint,
		TargetID:          testTarget,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	detail, err := svc.GetStreamDetail(context.Background(), testProjectA, created.Stream.ID)
	if err != nil {
		t.Fatalf("detail: %v", err)
	}
	if detail.Stream.ID != created.Stream.ID {
		t.Errorf("stream id mismatch")
	}
	if detail.Endpoint.ID != testEndpoint {
		t.Errorf("endpoint id mismatch")
	}
	if detail.Target.ID != testTarget {
		t.Errorf("target id mismatch")
	}
	if detail.RoutingRule.StreamID != created.Stream.ID {
		t.Errorf("routing rule stream_id mismatch")
	}
}

func TestIngestControlService_GetStreamDetail_StreamNotFound(t *testing.T) {
	t.Parallel()
	svc, _ := newSeededService(t)
	_, err := svc.GetStreamDetail(context.Background(), testProjectA, "missing")
	if !errors.Is(err, domain.ErrIngestStreamNotFound) {
		t.Errorf("err: want ErrIngestStreamNotFound, got %v", err)
	}
}

func TestIngestControlService_RotateKey_StreamNotFound(t *testing.T) {
	t.Parallel()
	svc, _ := newSeededService(t)
	_, err := svc.RotateKey(context.Background(), testProjectA, "missing")
	if !errors.Is(err, domain.ErrIngestStreamNotFound) {
		t.Errorf("err: want ErrIngestStreamNotFound, got %v", err)
	}
}

func TestIngestControlService_RecordLifecycleEvent_StreamNotFound(t *testing.T) {
	t.Parallel()
	svc, _ := newSeededService(t)
	_, err := svc.RecordLifecycleEvent(context.Background(), driving.LifecycleHookRequest{
		ResolvedProjectID: testProjectA,
		StreamID:          "missing",
		Kind:              domain.StreamLifecycleEventStarted,
		ObservedAt:        time.Now().UTC(),
		Source:            domain.StreamLifecycleSourceSmoke,
	})
	if !errors.Is(err, domain.ErrIngestStreamNotFound) {
		t.Errorf("err: want ErrIngestStreamNotFound, got %v", err)
	}
}

func TestIngestControlService_RecordLifecycleEvent_RejectsZeroObservedAt(t *testing.T) {
	t.Parallel()
	svc, _ := newSeededService(t)
	_, err := svc.RecordLifecycleEvent(context.Background(), driving.LifecycleHookRequest{
		ResolvedProjectID: testProjectA,
		StreamID:          "ing_x",
		Kind:              domain.StreamLifecycleEventStarted,
		Source:            domain.StreamLifecycleSourceSmoke,
	})
	if !errors.Is(err, domain.ErrIngestLifecycleObservedAtRequired) {
		t.Errorf("err: want ErrIngestLifecycleObservedAtRequired, got %v", err)
	}
}

func TestIngestControlService_RecordLifecycleEvent_RejectsUnknownSource(t *testing.T) {
	t.Parallel()
	svc, _ := newSeededService(t)
	_, err := svc.RecordLifecycleEvent(context.Background(), driving.LifecycleHookRequest{
		ResolvedProjectID: testProjectA,
		StreamID:          "ing_x",
		Kind:              domain.StreamLifecycleEventStarted,
		ObservedAt:        time.Now().UTC(),
		Source:            domain.StreamLifecycleEventSource("attacker-controlled"),
	})
	if !errors.Is(err, domain.ErrIngestLifecycleSourceUnknown) {
		t.Errorf("err: want ErrIngestLifecycleSourceUnknown, got %v", err)
	}
}

func TestIngestControlService_RecordLifecycleEvent_RejectsLongFields(t *testing.T) {
	t.Parallel()
	svc, _ := newSeededService(t)
	huge := strings.Repeat("x", domain.MaxLifecycleStringField+1)
	_, err := svc.RecordLifecycleEvent(context.Background(), driving.LifecycleHookRequest{
		ResolvedProjectID: testProjectA,
		StreamID:          "ing_x",
		Kind:              domain.StreamLifecycleEventStarted,
		ObservedAt:        time.Now().UTC(),
		Source:            domain.StreamLifecycleSourceSmoke,
		ConnectionID:      huge,
	})
	if !errors.Is(err, domain.ErrIngestLifecycleFieldTooLong) {
		t.Errorf("err: want ErrIngestLifecycleFieldTooLong, got %v", err)
	}
}

func TestIngestControlService_NewIngestControlService_NilClock(t *testing.T) {
	t.Parallel()
	repo := newFakeIngestRepo()
	repo.seedEndpoint(domain.IngestEndpoint{ID: testEndpoint, Protocol: domain.IngestProtocolSRT, ListenHost: "127.0.0.1", ListenPort: 8890, PathTemplate: "x", LabStack: "y", PublicURLHint: "z"})
	repo.seedTarget(domain.MediaServerTarget{ID: testTarget, Kind: domain.MediaServerKindMediaMTX, ConfigPath: "x", HLSURLTemplate: "y"})
	// nil-Clock fällt auf time.Now zurück; Konstruktor darf nicht
	// panicen.
	svc := application.NewIngestControlService(repo, nil)
	if _, err := svc.CreateStream(context.Background(), driving.CreateStreamRequest{
		ResolvedProjectID: testProjectA,
		DisplayName:       "Lab-NoClock",
		Protocol:          "srt",
		EndpointID:        testEndpoint,
		TargetID:          testTarget,
	}); err != nil {
		t.Fatalf("create with nil clock: %v", err)
	}
}

func TestIngestControlService_GetMediaServerConfig_HappyPath(t *testing.T) {
	t.Parallel()
	svc, _ := newSeededService(t)
	if _, err := svc.CreateStream(context.Background(), driving.CreateStreamRequest{
		ResolvedProjectID: testProjectA,
		DisplayName:       "Lab",
		Protocol:          "srt",
		EndpointID:        testEndpoint,
		TargetID:          testTarget,
	}); err != nil {
		t.Fatalf("create: %v", err)
	}
	result, err := svc.GetMediaServerConfig(context.Background(), testProjectA, "")
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	if result.TargetID != testTarget {
		t.Errorf("target_id: want %q, got %q", testTarget, result.TargetID)
	}
	if result.Kind != domain.MediaServerKindMediaMTX {
		t.Errorf("kind: want mediamtx, got %q", result.Kind)
	}
	if result.ConfigYAML == "" {
		t.Errorf("config_yaml must be present")
	}
}

func TestIngestControlService_GetMediaServerConfig_NoStreams(t *testing.T) {
	t.Parallel()
	svc, _ := newSeededService(t)
	_, err := svc.GetMediaServerConfig(context.Background(), testProjectA, "")
	if !errors.Is(err, application.ErrMediaMTXConfigNoStreams) {
		t.Errorf("err: want ErrMediaMTXConfigNoStreams, got %v", err)
	}
}

func TestIngestControlService_GetMediaServerConfig_TargetNotFound(t *testing.T) {
	t.Parallel()
	svc, _ := newSeededService(t)
	if _, err := svc.CreateStream(context.Background(), driving.CreateStreamRequest{
		ResolvedProjectID: testProjectA,
		DisplayName:       "Lab",
		Protocol:          "srt",
		EndpointID:        testEndpoint,
		TargetID:          testTarget,
	}); err != nil {
		t.Fatalf("create: %v", err)
	}
	_, err := svc.GetMediaServerConfig(context.Background(), testProjectA, "missing-target")
	if !errors.Is(err, domain.ErrIngestTargetNotFound) {
		t.Errorf("err: want ErrIngestTargetNotFound, got %v", err)
	}
}

func TestIngestControlService_GetMediaServerConfig_MultipleTargetsAutoPickWarns(t *testing.T) {
	t.Parallel()
	// Plan-0.11.0-Review-Fix: hat ein Project mehrere distinkte Targets
	// und der Aufrufer setzt **kein** ?target_id=, dann wählt der
	// Service zwar eines aus (das des ersten Streams), gibt aber einen
	// Warning mit den restlichen Target-IDs zurück. So bleibt der
	// Auto-Pick deterministisch und gleichzeitig sichtbar.
	svc, repo := newSeededService(t)
	const secondTarget = "tgt-mediamtx-secondary"
	repo.seedTarget(domain.MediaServerTarget{
		ID:             secondTarget,
		Kind:           domain.MediaServerKindMediaMTX,
		ConfigPath:     "examples/ingest-control/mediamtx.secondary.yml",
		HLSURLTemplate: "http://localhost:18889/{stream_path}/index.m3u8",
	})
	if _, err := svc.CreateStream(context.Background(), driving.CreateStreamRequest{
		ResolvedProjectID: testProjectA,
		DisplayName:       "Primary",
		Protocol:          "srt",
		EndpointID:        testEndpoint,
		TargetID:          testTarget,
	}); err != nil {
		t.Fatalf("create primary: %v", err)
	}
	if _, err := svc.CreateStream(context.Background(), driving.CreateStreamRequest{
		ResolvedProjectID: testProjectA,
		DisplayName:       "Secondary",
		Protocol:          "srt",
		EndpointID:        testEndpoint,
		TargetID:          secondTarget,
	}); err != nil {
		t.Fatalf("create secondary: %v", err)
	}
	result, err := svc.GetMediaServerConfig(context.Background(), testProjectA, "")
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	if result.TargetID != testTarget {
		t.Errorf("auto-pick must use first stream's target; got %q", result.TargetID)
	}
	if len(result.Warnings) == 0 {
		t.Fatalf("expected at least one warning for multi-target auto-pick")
	}
	want := result.Warnings[0]
	if !strings.Contains(want, secondTarget) {
		t.Errorf("warning must name the unselected target %q; got %q", secondTarget, want)
	}
	if !strings.Contains(want, "?target_id=") {
		t.Errorf("warning should hint how to disambiguate; got %q", want)
	}
}

func TestIngestControlService_GetMediaServerConfig_ExplicitTargetSuppressesWarning(t *testing.T) {
	t.Parallel()
	// Setzt der Aufrufer ?target_id=... explizit, gibt es keinen
	// Auto-Pick-Warning — auch wenn andere Targets existieren.
	svc, repo := newSeededService(t)
	const secondTarget = "tgt-mediamtx-secondary"
	repo.seedTarget(domain.MediaServerTarget{
		ID:             secondTarget,
		Kind:           domain.MediaServerKindMediaMTX,
		ConfigPath:     "examples/ingest-control/mediamtx.secondary.yml",
		HLSURLTemplate: "http://localhost:18889/{stream_path}/index.m3u8",
	})
	if _, err := svc.CreateStream(context.Background(), driving.CreateStreamRequest{
		ResolvedProjectID: testProjectA,
		DisplayName:       "Primary",
		Protocol:          "srt",
		EndpointID:        testEndpoint,
		TargetID:          testTarget,
	}); err != nil {
		t.Fatalf("create primary: %v", err)
	}
	if _, err := svc.CreateStream(context.Background(), driving.CreateStreamRequest{
		ResolvedProjectID: testProjectA,
		DisplayName:       "Secondary",
		Protocol:          "srt",
		EndpointID:        testEndpoint,
		TargetID:          secondTarget,
	}); err != nil {
		t.Fatalf("create secondary: %v", err)
	}
	result, err := svc.GetMediaServerConfig(context.Background(), testProjectA, secondTarget)
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	if result.TargetID != secondTarget {
		t.Errorf("target_id: want %q, got %q", secondTarget, result.TargetID)
	}
	for _, w := range result.Warnings {
		if strings.Contains(w, "auto-selected") {
			t.Errorf("explicit target_id must not produce auto-pick warning; got %q", w)
		}
	}
}

// stubOutboundDispatcher captures every Dispatch call so we can
// assert that `RecordLifecycleEvent` routes the event through the
// optional webhook adapter when one is wired up.
type stubOutboundDispatcher struct {
	calls []driven.OutboundWebhookEvent
	err   error
}

func (s *stubOutboundDispatcher) Dispatch(_ context.Context, e driven.OutboundWebhookEvent) error {
	s.calls = append(s.calls, e)
	return s.err
}

func TestIngestControlService_RecordLifecycleEvent_DispatchesToWebhook(t *testing.T) {
	t.Parallel()
	svc, _ := newSeededService(t)
	dispatcher := &stubOutboundDispatcher{}
	svc = svc.WithOutboundWebhookDispatcher(dispatcher)
	created, err := svc.CreateStream(context.Background(), driving.CreateStreamRequest{
		ResolvedProjectID: testProjectA,
		DisplayName:       "Lab",
		Protocol:          "srt",
		EndpointID:        testEndpoint,
		TargetID:          testTarget,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	_, err = svc.RecordLifecycleEvent(context.Background(), driving.LifecycleHookRequest{
		ResolvedProjectID: testProjectA,
		StreamID:          created.Stream.ID,
		Kind:              domain.StreamLifecycleEventStarted,
		ObservedAt:        time.Now().UTC(),
		Source:            domain.StreamLifecycleSourceSmoke,
	})
	if err != nil {
		t.Fatalf("record: %v", err)
	}
	if len(dispatcher.calls) != 1 {
		t.Fatalf("want 1 dispatch call, got %d", len(dispatcher.calls))
	}
	if dispatcher.calls[0].Kind != domain.StreamLifecycleEventStarted {
		t.Errorf("dispatched kind: want stream_started, got %s", dispatcher.calls[0].Kind)
	}
	if dispatcher.calls[0].StreamID != created.Stream.ID {
		t.Errorf("dispatched stream_id mismatch")
	}
}

func TestIngestControlService_RecordLifecycleEvent_WebhookErrorDoesNotFailLifecycle(t *testing.T) {
	t.Parallel()
	svc, _ := newSeededService(t)
	svc = svc.WithOutboundWebhookDispatcher(&stubOutboundDispatcher{
		err: errors.New("downstream unreachable"),
	})
	created, err := svc.CreateStream(context.Background(), driving.CreateStreamRequest{
		ResolvedProjectID: testProjectA,
		DisplayName:       "Lab",
		Protocol:          "srt",
		EndpointID:        testEndpoint,
		TargetID:          testTarget,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	result, err := svc.RecordLifecycleEvent(context.Background(), driving.LifecycleHookRequest{
		ResolvedProjectID: testProjectA,
		StreamID:          created.Stream.ID,
		Kind:              domain.StreamLifecycleEventEnded,
		ObservedAt:        time.Now().UTC(),
		Source:            domain.StreamLifecycleSourceSmoke,
	})
	if err != nil {
		t.Fatalf("lifecycle must not fail when webhook dispatcher errors: %v", err)
	}
	if !strings.HasPrefix(result.EventID, "evt_") {
		t.Errorf("event_id must be issued regardless of webhook error, got %q", result.EventID)
	}
}

func TestIngestControlService_RecordLifecycleEvent_NoKlartextKey(t *testing.T) {
	t.Parallel()
	svc, repo := newSeededService(t)
	created, err := svc.CreateStream(context.Background(), driving.CreateStreamRequest{
		ResolvedProjectID: testProjectA,
		DisplayName:       "Lab",
		Protocol:          "srt",
		EndpointID:        testEndpoint,
		TargetID:          testTarget,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	result, err := svc.RecordLifecycleEvent(context.Background(), driving.LifecycleHookRequest{
		ResolvedProjectID: testProjectA,
		StreamID:          created.Stream.ID,
		Kind:              domain.StreamLifecycleEventStarted,
		ObservedAt:        time.Now().UTC(),
		Source:            domain.StreamLifecycleSourceSmoke,
		ConnectionID:      "srtconn-1",
	})
	if err != nil {
		t.Fatalf("record: %v", err)
	}
	if !strings.HasPrefix(result.EventID, "evt_") {
		t.Errorf("event_id must have evt_ prefix, got %q", result.EventID)
	}
	events := repo.lifecycleSnapshot()
	if len(events) != 1 {
		t.Fatalf("want 1 event, got %d", len(events))
	}
	ev := events[0]
	if ev.EventID != result.EventID {
		t.Errorf("event_id mismatch: persisted %q vs returned %q", ev.EventID, result.EventID)
	}
	if ev.ConnectionID != "srtconn-1" {
		t.Errorf("connection_id: want srtconn-1, got %q", ev.ConnectionID)
	}
	if ev.KeyFingerprint == "" {
		t.Errorf("fingerprint must be present in lifecycle event")
	}
	// Plan T1/T4 DoD: Lifecycle-Events tragen niemals Klartext-Keys.
	if ev.KeyFingerprint == created.Material.Value {
		t.Errorf("lifecycle event must NOT carry klartext key")
	}
	if strings.Contains(ev.Reason, created.Material.Value) || strings.Contains(ev.ConnectionID, created.Material.Value) {
		t.Errorf("optional fields must not echo klartext key value")
	}
}
