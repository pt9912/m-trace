package sqlite_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/persistence/sqlite"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
	"github.com/pt9912/m-trace/apps/api/internal/storage"
)

// `0.11.0` Tranche 2 — SQLite-Adapter-Tests gegen die V2__ingest.sql-
// Migration. Pinnt produktive Persistenz, Cross-Project-Leak-Schutz,
// Key-Rotation mit Deaktivierungs-Audit-Trail und Lifecycle-Append.

func newSeededSQLiteRepo(t *testing.T) *sqlite.IngestStreamRepository {
	t.Helper()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "m-trace.db")
	db, err := storage.Open(ctx, path)
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	repo := sqlite.NewIngestStreamRepository(db)
	if err := repo.SeedEndpoint(ctx, domain.IngestEndpoint{
		ID:            "ep-srt",
		Protocol:      domain.IngestProtocolSRT,
		ListenHost:    "127.0.0.1",
		ListenPort:    8890,
		PathTemplate:  "publish:{stream_path}",
		LabStack:      "mtrace-srt",
		PublicURLHint: "srt://localhost:8890",
	}); err != nil {
		t.Fatalf("seed endpoint: %v", err)
	}
	if err := repo.SeedTarget(ctx, domain.MediaServerTarget{
		ID:             "tgt-mediamtx",
		Kind:           domain.MediaServerKindMediaMTX,
		ConfigPath:     "examples/ingest-control/mediamtx.generated.yml",
		HLSURLTemplate: "http://localhost:8889/{stream_path}/index.m3u8",
		ControlAPIURL:  "",
	}); err != nil {
		t.Fatalf("seed target: %v", err)
	}
	return repo
}

func mkSqliteKey(t *testing.T, when time.Time) domain.StreamKey {
	t.Helper()
	material, err := domain.GenerateStreamKey(when)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	return material.ToPersistable()
}

func mkSqliteInput(projectID, displayName string, key domain.StreamKey) driven.CreateStreamInput {
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

func TestSQLiteIngestRepo_CreateStream_HappyPath(t *testing.T) {
	t.Parallel()
	repo := newSeededSQLiteRepo(t)
	now := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)
	stream, err := repo.CreateStream(context.Background(), mkSqliteInput("p1", "Lab", mkSqliteKey(t, now)))
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if stream.ProjectID != "p1" || stream.Status != domain.IngestStreamStatusReady {
		t.Errorf("stream: %+v", stream)
	}
}

func TestSQLiteIngestRepo_CreateStream_RejectsMissingEndpoint(t *testing.T) {
	t.Parallel()
	repo := newSeededSQLiteRepo(t)
	now := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)
	in := mkSqliteInput("p1", "Lab", mkSqliteKey(t, now))
	in.EndpointID = "missing"
	_, err := repo.CreateStream(context.Background(), in)
	if !errors.Is(err, domain.ErrIngestEndpointNotFound) {
		t.Errorf("err: want ErrIngestEndpointNotFound, got %v", err)
	}
}

func TestSQLiteIngestRepo_GetStreamByID_CrossProjectIsNotFound(t *testing.T) {
	t.Parallel()
	repo := newSeededSQLiteRepo(t)
	now := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)
	created, _ := repo.CreateStream(context.Background(), mkSqliteInput("p1", "Lab", mkSqliteKey(t, now)))
	if _, err := repo.GetStreamByID(context.Background(), "p2", created.ID); !errors.Is(err, domain.ErrIngestStreamNotFound) {
		t.Errorf("cross-project: want ErrIngestStreamNotFound, got %v", err)
	}
	got, err := repo.GetStreamByID(context.Background(), "p1", created.ID)
	if err != nil {
		t.Fatalf("same-project: %v", err)
	}
	if got.ID != created.ID || got.Key.Hash != created.Key.Hash {
		t.Errorf("stream mismatch")
	}
}

func TestSQLiteIngestRepo_DuplicateActiveName(t *testing.T) {
	t.Parallel()
	repo := newSeededSQLiteRepo(t)
	now := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)
	if _, err := repo.CreateStream(context.Background(), mkSqliteInput("p1", "Lab", mkSqliteKey(t, now))); err != nil {
		t.Fatalf("first: %v", err)
	}
	_, err := repo.CreateStream(context.Background(), mkSqliteInput("p1", "Lab", mkSqliteKey(t, now.Add(time.Second))))
	if !errors.Is(err, domain.ErrIngestStreamNameConflict) {
		t.Errorf("err: want ErrIngestStreamNameConflict, got %v", err)
	}
}

func TestSQLiteIngestRepo_RotateKey_DeactivatesOld(t *testing.T) {
	t.Parallel()
	repo := newSeededSQLiteRepo(t)
	now := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)
	created, _ := repo.CreateStream(context.Background(), mkSqliteInput("p1", "Lab", mkSqliteKey(t, now)))
	originalHash := created.Key.Hash
	newKey := mkSqliteKey(t, now.Add(time.Hour))
	rotated, err := repo.RotateKey(context.Background(), "p1", created.ID, newKey)
	if err != nil {
		t.Fatalf("rotate: %v", err)
	}
	if rotated.Key.Hash == originalHash {
		t.Errorf("hash must change")
	}
	active, err := repo.FindActiveStreamKey(context.Background(), "p1", created.ID)
	if err != nil {
		t.Fatalf("find active: %v", err)
	}
	if active.Hash != newKey.Hash {
		t.Errorf("active hash mismatch")
	}
}

func TestSQLiteIngestRepo_ListByProject_FiltersByProject(t *testing.T) {
	t.Parallel()
	repo := newSeededSQLiteRepo(t)
	now := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)
	for _, name := range []string{"a", "b"} {
		if _, err := repo.CreateStream(context.Background(), mkSqliteInput("p1", name, mkSqliteKey(t, now))); err != nil {
			t.Fatalf("p1 %q: %v", name, err)
		}
		now = now.Add(time.Second)
	}
	if _, err := repo.CreateStream(context.Background(), mkSqliteInput("p2", "other", mkSqliteKey(t, now))); err != nil {
		t.Fatalf("p2: %v", err)
	}
	streams, err := repo.ListByProject(context.Background(), "p1")
	if err != nil {
		t.Fatalf("list: %v", err)
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

func TestSQLiteIngestRepo_GetEndpointAndTarget_NotFound(t *testing.T) {
	t.Parallel()
	repo := newSeededSQLiteRepo(t)
	if _, err := repo.GetEndpointByID(context.Background(), "missing"); !errors.Is(err, domain.ErrIngestEndpointNotFound) {
		t.Errorf("endpoint: want ErrIngestEndpointNotFound, got %v", err)
	}
	if _, err := repo.GetTargetByID(context.Background(), "missing"); !errors.Is(err, domain.ErrIngestTargetNotFound) {
		t.Errorf("target: want ErrIngestTargetNotFound, got %v", err)
	}
}

func TestSQLiteIngestRepo_RoutingRule_AfterCreate(t *testing.T) {
	t.Parallel()
	repo := newSeededSQLiteRepo(t)
	now := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)
	created, _ := repo.CreateStream(context.Background(), mkSqliteInput("p1", "Lab", mkSqliteKey(t, now)))
	rule, err := repo.GetRoutingRuleByID(context.Background(), "p1", created.ID)
	if err != nil {
		t.Fatalf("rule: %v", err)
	}
	if !rule.Enabled || rule.Mode != domain.RoutingRuleModeSingle {
		t.Errorf("rule: %+v", rule)
	}
}

func TestSQLiteIngestRepo_AppendLifecycleEvent(t *testing.T) {
	t.Parallel()
	repo := newSeededSQLiteRepo(t)
	now := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)
	created, _ := repo.CreateStream(context.Background(), mkSqliteInput("p1", "Lab", mkSqliteKey(t, now)))
	if err := repo.AppendLifecycleEvent(context.Background(), domain.StreamLifecycleEvent{
		EventID: "evt_sqlite1", Kind: domain.StreamLifecycleEventStarted, StreamID: created.ID, ProjectID: "p1",
		OccurredAt: now, Source: domain.StreamLifecycleSourceSmoke, KeyFingerprint: created.Key.Fingerprint,
		ConnectionID: "srtconn-1", Reason: "smoke-test",
	}); err != nil {
		t.Fatalf("append: %v", err)
	}
	// Empty event_id wird abgelehnt — der Service muss event_id setzen.
	if err := repo.AppendLifecycleEvent(context.Background(), domain.StreamLifecycleEvent{
		Kind: domain.StreamLifecycleEventEnded, StreamID: created.ID, ProjectID: "p1",
		OccurredAt: now, Source: domain.StreamLifecycleSourceSmoke,
	}); err == nil {
		t.Errorf("empty event_id must be rejected by sqlite repo")
	}
	// Cross-Project-Append wird abgelehnt.
	if err := repo.AppendLifecycleEvent(context.Background(), domain.StreamLifecycleEvent{
		EventID: "evt_xproj", StreamID: created.ID, ProjectID: "other",
	}); !errors.Is(err, domain.ErrIngestStreamNotFound) {
		t.Errorf("cross-project lifecycle: want ErrIngestStreamNotFound, got %v", err)
	}
}

func TestSQLiteIngestRepo_RestartPreservesKeyHistory(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "m-trace.db")
	db1, err := storage.Open(ctx, path)
	if err != nil {
		t.Fatalf("open #1: %v", err)
	}
	repo1 := sqlite.NewIngestStreamRepository(db1)
	if err := repo1.SeedEndpoint(ctx, domain.IngestEndpoint{ID: "ep-srt", Protocol: domain.IngestProtocolSRT, ListenHost: "127.0.0.1", ListenPort: 8890, PathTemplate: "x", LabStack: "y", PublicURLHint: "z"}); err != nil {
		t.Fatalf("seed endpoint: %v", err)
	}
	if err := repo1.SeedTarget(ctx, domain.MediaServerTarget{ID: "tgt-mediamtx", Kind: domain.MediaServerKindMediaMTX, ConfigPath: "x", HLSURLTemplate: "y", ControlAPIURL: ""}); err != nil {
		t.Fatalf("seed target: %v", err)
	}
	now := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)
	created, err := repo1.CreateStream(ctx, mkSqliteInput("p1", "Lab", mkSqliteKey(t, now)))
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := db1.Close(); err != nil {
		t.Fatalf("close #1: %v", err)
	}
	// Re-open + verify stream + active key persist.
	db2, err := storage.Open(ctx, path)
	if err != nil {
		t.Fatalf("open #2: %v", err)
	}
	defer func() { _ = db2.Close() }()
	repo2 := sqlite.NewIngestStreamRepository(db2)
	got, err := repo2.GetStreamByID(ctx, "p1", created.ID)
	if err != nil {
		t.Fatalf("get after restart: %v", err)
	}
	if got.Key.Hash != created.Key.Hash {
		t.Errorf("key hash must persist across restart")
	}
}
