package inmemory_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/persistence/inmemory"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// `0.11.0` Tranche 2 — InMemory-Adapter-Tests für den
// IngestStreamRepository-Driven-Port. Spiegelt die Verhaltensregeln
// aus `apps/api/hexagon/application/ingest_control_service_test.go`,
// damit der Adapter eigenständig coverable ist und der gemeinsame
// Verhaltens-Vertrag (Cross-Project-Leak-Schutz, Duplicate-Erkennung,
// Key-Rotation) jeweils gegen den konkreten Adapter geprüft wird.

func newSeededInMemRepo(t *testing.T) *inmemory.IngestStreamRepository {
	t.Helper()
	repo := inmemory.NewIngestStreamRepository()
	repo.SeedEndpoint(domain.IngestEndpoint{
		ID:            "ep-srt",
		Protocol:      domain.IngestProtocolSRT,
		ListenHost:    "127.0.0.1",
		ListenPort:    8890,
		PathTemplate:  "publish:{stream_path}",
		LabStack:      "mtrace-srt",
		PublicURLHint: "srt://localhost:8890",
	})
	repo.SeedTarget(domain.MediaServerTarget{
		ID:             "tgt-mediamtx",
		Kind:           domain.MediaServerKindMediaMTX,
		ConfigPath:     "examples/ingest-control/mediamtx.generated.yml",
		HLSURLTemplate: "http://localhost:8889/{stream_path}/index.m3u8",
		ControlAPIURL:  "",
	})
	return repo
}

func mkInMemKey(t *testing.T, when time.Time) domain.StreamKey {
	t.Helper()
	material, err := domain.GenerateStreamKey(when)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	return material.ToPersistable()
}

func mkInMemInput(projectID, displayName string, key domain.StreamKey) driven.CreateStreamInput {
	return driven.CreateStreamInput{
		ProjectID:   projectID,
		DisplayName: displayName,
		Protocol:    domain.IngestProtocolSRT,
		EndpointID:  "ep-srt",
		TargetID:    "tgt-mediamtx",
		InitialKey:  key,
		CreatedAt:   key.CreatedAt,
	}
}

func TestInMemoryIngestRepo_CreateStream_HappyPath(t *testing.T) {
	t.Parallel()
	repo := newSeededInMemRepo(t)
	now := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)
	stream, err := repo.CreateStream(context.Background(), mkInMemInput("p1", "Lab", mkInMemKey(t, now)))
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if stream.ProjectID != "p1" {
		t.Errorf("project_id: want p1, got %q", stream.ProjectID)
	}
	if stream.Status != domain.IngestStreamStatusReady {
		t.Errorf("status: want ready, got %q", stream.Status)
	}
	if stream.Key.Hash == "" {
		t.Errorf("hash must be populated")
	}
}

func TestInMemoryIngestRepo_CreateStream_RejectsMissingEndpoint(t *testing.T) {
	t.Parallel()
	repo := inmemory.NewIngestStreamRepository()
	now := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)
	in := mkInMemInput("p1", "Lab", mkInMemKey(t, now))
	_, err := repo.CreateStream(context.Background(), in)
	if !errors.Is(err, domain.ErrIngestEndpointNotFound) {
		t.Errorf("err: want ErrIngestEndpointNotFound, got %v", err)
	}
}

func TestInMemoryIngestRepo_CreateStream_RejectsMissingTarget(t *testing.T) {
	t.Parallel()
	repo := inmemory.NewIngestStreamRepository()
	repo.SeedEndpoint(domain.IngestEndpoint{ID: "ep-srt", Protocol: domain.IngestProtocolSRT, ListenHost: "127.0.0.1", ListenPort: 8890, PathTemplate: "x", LabStack: "y", PublicURLHint: "z"})
	now := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)
	in := mkInMemInput("p1", "Lab", mkInMemKey(t, now))
	_, err := repo.CreateStream(context.Background(), in)
	if !errors.Is(err, domain.ErrIngestTargetNotFound) {
		t.Errorf("err: want ErrIngestTargetNotFound, got %v", err)
	}
}

func TestInMemoryIngestRepo_DuplicateActiveName(t *testing.T) {
	t.Parallel()
	repo := newSeededInMemRepo(t)
	now := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)
	if _, err := repo.CreateStream(context.Background(), mkInMemInput("p1", "Lab", mkInMemKey(t, now))); err != nil {
		t.Fatalf("first: %v", err)
	}
	_, err := repo.CreateStream(context.Background(), mkInMemInput("p1", "Lab", mkInMemKey(t, now.Add(time.Second))))
	if !errors.Is(err, domain.ErrIngestStreamNameConflict) {
		t.Errorf("err: want ErrIngestStreamNameConflict, got %v", err)
	}
}

func TestInMemoryIngestRepo_GetStreamByID_CrossProjectIsNotFound(t *testing.T) {
	t.Parallel()
	repo := newSeededInMemRepo(t)
	now := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)
	created, _ := repo.CreateStream(context.Background(), mkInMemInput("p1", "Lab", mkInMemKey(t, now)))
	if _, err := repo.GetStreamByID(context.Background(), "p2", created.ID); !errors.Is(err, domain.ErrIngestStreamNotFound) {
		t.Errorf("cross-project: want ErrIngestStreamNotFound, got %v", err)
	}
	got, err := repo.GetStreamByID(context.Background(), "p1", created.ID)
	if err != nil {
		t.Fatalf("same-project: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("id mismatch")
	}
}

func TestInMemoryIngestRepo_ListByProject_FiltersAndSorts(t *testing.T) {
	t.Parallel()
	repo := newSeededInMemRepo(t)
	now := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)
	for _, name := range []string{"a", "b"} {
		if _, err := repo.CreateStream(context.Background(), mkInMemInput("p1", name, mkInMemKey(t, now))); err != nil {
			t.Fatalf("p1 %q: %v", name, err)
		}
		now = now.Add(time.Second)
	}
	if _, err := repo.CreateStream(context.Background(), mkInMemInput("p2", "other", mkInMemKey(t, now))); err != nil {
		t.Fatalf("p2: %v", err)
	}
	streams, err := repo.ListByProject(context.Background(), "p1")
	if err != nil {
		t.Fatalf("list p1: %v", err)
	}
	if len(streams) != 2 {
		t.Errorf("want 2 streams, got %d", len(streams))
	}
	for _, s := range streams {
		if s.ProjectID != "p1" {
			t.Errorf("cross-project leak: %q", s.ProjectID)
		}
	}
}

func TestInMemoryIngestRepo_RotateKey_DeactivatesOld(t *testing.T) {
	t.Parallel()
	repo := newSeededInMemRepo(t)
	now := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)
	created, _ := repo.CreateStream(context.Background(), mkInMemInput("p1", "Lab", mkInMemKey(t, now)))
	originalHash := created.Key.Hash
	newKey := mkInMemKey(t, now.Add(time.Hour))
	rotated, err := repo.RotateKey(context.Background(), "p1", created.ID, newKey)
	if err != nil {
		t.Fatalf("rotate: %v", err)
	}
	if rotated.Key.Hash == originalHash {
		t.Errorf("hash must change after rotate")
	}
	active, err := repo.FindActiveStreamKey(context.Background(), "p1", created.ID)
	if err != nil {
		t.Fatalf("find active: %v", err)
	}
	if active.Hash != newKey.Hash {
		t.Errorf("active hash must equal rotated hash")
	}
}

func TestInMemoryIngestRepo_GetEndpointByID_NotFound(t *testing.T) {
	t.Parallel()
	repo := inmemory.NewIngestStreamRepository()
	_, err := repo.GetEndpointByID(context.Background(), "missing")
	if !errors.Is(err, domain.ErrIngestEndpointNotFound) {
		t.Errorf("err: want ErrIngestEndpointNotFound, got %v", err)
	}
}

func TestInMemoryIngestRepo_GetTargetByID_NotFound(t *testing.T) {
	t.Parallel()
	repo := inmemory.NewIngestStreamRepository()
	_, err := repo.GetTargetByID(context.Background(), "missing")
	if !errors.Is(err, domain.ErrIngestTargetNotFound) {
		t.Errorf("err: want ErrIngestTargetNotFound, got %v", err)
	}
}

func TestInMemoryIngestRepo_RoutingRuleAndLifecycle(t *testing.T) {
	t.Parallel()
	repo := newSeededInMemRepo(t)
	now := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)
	created, _ := repo.CreateStream(context.Background(), mkInMemInput("p1", "Lab", mkInMemKey(t, now)))
	rule, err := repo.GetRoutingRuleByID(context.Background(), "p1", created.ID)
	if err != nil {
		t.Fatalf("rule: %v", err)
	}
	if !rule.Enabled {
		t.Errorf("rule must be enabled by default")
	}
	if rule.Mode != domain.RoutingRuleModeSingle {
		t.Errorf("mode: want single, got %q", rule.Mode)
	}
	// Lifecycle-Event-Append + Snapshot.
	if err := repo.AppendLifecycleEvent(context.Background(), domain.StreamLifecycleEvent{
		Kind: domain.StreamLifecycleEventStarted, StreamID: created.ID, ProjectID: "p1",
		OccurredAt: now, Source: domain.StreamLifecycleSourceSmoke, KeyFingerprint: created.Key.Fingerprint,
	}); err != nil {
		t.Fatalf("append lifecycle: %v", err)
	}
	events := repo.LifecycleSnapshot()
	if len(events) != 1 {
		t.Errorf("want 1 lifecycle event, got %d", len(events))
	}
	// Cross-Project-Lifecycle-Append wird abgelehnt.
	if err := repo.AppendLifecycleEvent(context.Background(), domain.StreamLifecycleEvent{
		StreamID: created.ID, ProjectID: "other",
	}); !errors.Is(err, domain.ErrIngestStreamNotFound) {
		t.Errorf("cross-project lifecycle: want ErrIngestStreamNotFound, got %v", err)
	}
}

func TestInMemoryIngestRepo_FindActiveStreamKey_StreamNotFound(t *testing.T) {
	t.Parallel()
	repo := inmemory.NewIngestStreamRepository()
	_, err := repo.FindActiveStreamKey(context.Background(), "p1", "missing")
	if !errors.Is(err, domain.ErrIngestStreamNotFound) {
		t.Errorf("err: want ErrIngestStreamNotFound, got %v", err)
	}
}
