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

// `0.12.0` Tranche 3 / RAK-73: Project-Token-Generationen-Repo
// (SQLite-Variante). Spiegelnde Suite zu inmemory; zusätzlich ein
// Restart-Test, der die Persistenz von `grace_until` über einen
// Schließe-Öffne-Zyklus pinnt.

func openProjectTokenDB(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "m-trace.db")
}

func openProjectTokenRepo(t *testing.T, path string) (*sqlite.ProjectTokenRepository, func()) {
	t.Helper()
	ctx := context.Background()
	db, err := storage.Open(ctx, path)
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	repo := sqlite.NewProjectTokenRepository(db)
	cleanup := func() { _ = db.Close() }
	return repo, cleanup
}

func mintGenerationT(t *testing.T, projectID, tokenID string, createdAt time.Time) domain.ProjectTokenGeneration {
	t.Helper()
	m, err := domain.GenerateProjectToken(tokenID, projectID, createdAt, nil, nil, nil, createdAt)
	if err != nil {
		t.Fatalf("GenerateProjectToken: %v", err)
	}
	return m.Generation
}

func TestSQLiteProjectTokenRepo_RoundTrip(t *testing.T) {
	t.Parallel()
	path := openProjectTokenDB(t)
	repo, cleanup := openProjectTokenRepo(t, path)
	defer cleanup()
	ctx := context.Background()
	now := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	gen := mintGenerationT(t, "demo", "pt_a", now)
	if err := repo.Create(ctx, gen); err != nil {
		t.Fatalf("create: %v", err)
	}
	got, err := repo.FindByHash(ctx, gen.KeyHash)
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if got.TokenID != "pt_a" || got.ProjectID != "demo" {
		t.Errorf("got %+v", got)
	}
	if !got.NotBefore.Equal(now) {
		t.Errorf("NotBefore: want %v, got %v", now, got.NotBefore)
	}
	if !got.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt: want %v, got %v", now, got.CreatedAt)
	}
}

func TestSQLiteProjectTokenRepo_HashUnique(t *testing.T) {
	t.Parallel()
	path := openProjectTokenDB(t)
	repo, cleanup := openProjectTokenRepo(t, path)
	defer cleanup()
	ctx := context.Background()
	gen := mintGenerationT(t, "demo", "pt_a", time.Now().UTC())
	if err := repo.Create(ctx, gen); err != nil {
		t.Fatalf("create a: %v", err)
	}
	dup := gen
	dup.TokenID = "pt_a_dup"
	if err := repo.Create(ctx, dup); !errors.Is(err, domain.ErrAuthTokenInvalid) {
		t.Errorf("dup hash: want ErrAuthTokenInvalid, got %v", err)
	}
}

func TestSQLiteProjectTokenRepo_GraceAndRevoke(t *testing.T) {
	t.Parallel()
	path := openProjectTokenDB(t)
	repo, cleanup := openProjectTokenRepo(t, path)
	defer cleanup()
	ctx := context.Background()
	gen := mintGenerationT(t, "demo", "pt_a", time.Now().UTC())
	if err := repo.Create(ctx, gen); err != nil {
		t.Fatalf("create: %v", err)
	}
	graceUntil := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	if err := repo.SetGraceUntil(ctx, "demo", "pt_a", graceUntil); err != nil {
		t.Fatalf("grace: %v", err)
	}
	got, err := repo.FindByHash(ctx, gen.KeyHash)
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if got.GraceUntil == nil || !got.GraceUntil.Equal(graceUntil) {
		t.Errorf("GraceUntil: %v", got.GraceUntil)
	}
	revokedAt := time.Date(2026, 7, 2, 0, 0, 0, 0, time.UTC)
	if err := repo.Revoke(ctx, "demo", "pt_a", revokedAt); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	got, err = repo.FindByHash(ctx, gen.KeyHash)
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if got.RevokedAt == nil || !got.RevokedAt.Equal(revokedAt) {
		t.Errorf("RevokedAt: %v", got.RevokedAt)
	}
}

func TestSQLiteProjectTokenRepo_Restart_GraceUntilPersisted(t *testing.T) {
	t.Parallel()
	// RAK-73 Restart-Test: grace_until lebt über einen Close-Reopen-
	// Zyklus, ohne dass der Caller `RotatedFrom` oder Prozesszustand
	// zur Rekonstruktion braucht.
	path := openProjectTokenDB(t)
	ctx := context.Background()
	now := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	gen := mintGenerationT(t, "demo", "pt_a", now)
	graceUntil := now.Add(2 * time.Hour)

	{
		repo, cleanup := openProjectTokenRepo(t, path)
		if err := repo.Create(ctx, gen); err != nil {
			t.Fatalf("create: %v", err)
		}
		if err := repo.SetGraceUntil(ctx, "demo", "pt_a", graceUntil); err != nil {
			t.Fatalf("grace: %v", err)
		}
		cleanup()
	}

	// Reopen and verify state.
	repo, cleanup := openProjectTokenRepo(t, path)
	defer cleanup()
	got, err := repo.FindByHash(ctx, gen.KeyHash)
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if got.GraceUntil == nil || !got.GraceUntil.Equal(graceUntil) {
		t.Errorf("GraceUntil after restart: want %v, got %v", graceUntil, got.GraceUntil)
	}
	gens, err := repo.ListByProject(ctx, "demo")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(gens) != 1 {
		t.Fatalf("len after restart: want 1, got %d", len(gens))
	}
	if gens[0].GraceUntil == nil || !gens[0].GraceUntil.Equal(graceUntil) {
		t.Errorf("ListByProject GraceUntil: %v", gens[0].GraceUntil)
	}
}

func TestSQLiteProjectTokenRepo_RevokeMissing(t *testing.T) {
	t.Parallel()
	path := openProjectTokenDB(t)
	repo, cleanup := openProjectTokenRepo(t, path)
	defer cleanup()
	if err := repo.Revoke(context.Background(), "demo", "ghost", time.Now()); !errors.Is(err, driven.ErrProjectTokenNotFound) {
		t.Errorf("missing token: want ErrProjectTokenNotFound, got %v", err)
	}
}

func TestSQLiteProjectTokenRepo_FindByHashUnknown(t *testing.T) {
	t.Parallel()
	path := openProjectTokenDB(t)
	repo, cleanup := openProjectTokenRepo(t, path)
	defer cleanup()
	if _, err := repo.FindByHash(context.Background(), "nope"); !errors.Is(err, driven.ErrProjectTokenNotFound) {
		t.Errorf("unknown hash: want ErrProjectTokenNotFound, got %v", err)
	}
}
