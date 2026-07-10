package postgres_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/persistence/postgres"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
	"github.com/pt9912/m-trace/apps/api/internal/storage"
)

// TestProjectTokenRepository_PgLab deckt den Postgres-project_token-
// Adapter gegen echte PG ab: Create/FindByHash/ListByProject,
// SetGraceUntil/Revoke (idempotent + NotFound), und den zentralen
// Dialekt-Punkt: Unique-Verletzung → domain.ErrAuthTokenInvalid über
// SQLSTATE 23505 (statt SQLite-Fehlerstring). Gated über MTRACE_PG_LAB_DSN.
func TestProjectTokenRepository_PgLab(t *testing.T) {
	dsn := os.Getenv("MTRACE_PG_LAB_DSN")
	if dsn == "" {
		t.Skip("MTRACE_PG_LAB_DSN nicht gesetzt — PG-Lab-Integrationstest übersprungen")
	}
	ctx := context.Background()
	db, err := storage.OpenPostgres(ctx, dsn)
	if err != nil {
		t.Fatalf("OpenPostgres: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	repo := postgres.NewProjectTokenRepository(db)
	const proj = "pt-lab-proj"
	base := time.Date(2026, 7, 10, 10, 0, 0, 0, time.UTC)

	gen := domain.ProjectTokenGeneration{
		TokenID:     "pt-lab-tok-1",
		ProjectID:   proj,
		KeyHash:     "pt-lab-hash-1",
		Fingerprint: "fp-1",
		NotBefore:   base,
		CreatedAt:   base,
	}
	if err := repo.Create(ctx, gen); err != nil {
		t.Fatalf("Create: %v", err)
	}

	t.Run("find by hash", func(t *testing.T) {
		got, err := repo.FindByHash(ctx, "pt-lab-hash-1")
		if err != nil {
			t.Fatalf("FindByHash: %v", err)
		}
		if got.TokenID != "pt-lab-tok-1" || got.ProjectID != proj {
			t.Errorf("FindByHash = %+v, want token pt-lab-tok-1 / project %s", got, proj)
		}
	})

	t.Run("find by hash not found", func(t *testing.T) {
		_, err := repo.FindByHash(ctx, "pt-lab-hash-missing")
		if !errors.Is(err, driven.ErrProjectTokenNotFound) {
			t.Errorf("FindByHash(missing) err = %v, want ErrProjectTokenNotFound", err)
		}
	})

	t.Run("unique violation maps to ErrAuthTokenInvalid (SQLSTATE 23505)", func(t *testing.T) {
		dup := gen
		dup.TokenID = "pt-lab-tok-2" // andere Token-ID, gleicher key_hash
		err := repo.Create(ctx, dup)
		if !errors.Is(err, domain.ErrAuthTokenInvalid) {
			t.Errorf("Create(dup hash) err = %v, want ErrAuthTokenInvalid", err)
		}
	})

	t.Run("list by project", func(t *testing.T) {
		list, err := repo.ListByProject(ctx, proj)
		if err != nil {
			t.Fatalf("ListByProject: %v", err)
		}
		if len(list) != 1 || list[0].TokenID != "pt-lab-tok-1" {
			t.Errorf("ListByProject = %d Einträge (%+v), want genau pt-lab-tok-1", len(list), list)
		}
	})

	t.Run("set grace + revoke idempotent, NotFound on missing", func(t *testing.T) {
		if err := repo.SetGraceUntil(ctx, proj, "pt-lab-tok-1", base.Add(time.Hour)); err != nil {
			t.Fatalf("SetGraceUntil: %v", err)
		}
		if err := repo.Revoke(ctx, proj, "pt-lab-tok-1", base.Add(2*time.Hour)); err != nil {
			t.Fatalf("Revoke: %v", err)
		}
		got, err := repo.FindByHash(ctx, "pt-lab-hash-1")
		if err != nil {
			t.Fatalf("FindByHash post-update: %v", err)
		}
		if got.GraceUntil == nil || got.RevokedAt == nil {
			t.Errorf("nach SetGrace/Revoke: GraceUntil=%v RevokedAt=%v, beide erwartet gesetzt", got.GraceUntil, got.RevokedAt)
		}
		if err := repo.Revoke(ctx, proj, "pt-lab-missing", base); !errors.Is(err, driven.ErrProjectTokenNotFound) {
			t.Errorf("Revoke(missing) err = %v, want ErrProjectTokenNotFound", err)
		}
	})
}
