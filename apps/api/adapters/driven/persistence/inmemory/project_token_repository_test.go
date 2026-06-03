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

// RAK-73: Project-Token-Generationen-Repo
// (InMemory-Variante). Spiegelnde Suite läuft in `sqlite_test`.

func mintGeneration(t *testing.T, projectID, tokenID string, createdAt time.Time) domain.ProjectTokenGeneration {
	t.Helper()
	m, err := domain.GenerateProjectToken(tokenID, projectID, createdAt, nil, nil, nil, createdAt)
	if err != nil {
		t.Fatalf("GenerateProjectToken: %v", err)
	}
	return m.Generation
}

func TestInMemoryProjectTokenRepo_CreateAndList(t *testing.T) {
	t.Parallel()
	r := inmemory.NewProjectTokenRepository()
	ctx := context.Background()
	now := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	a := mintGeneration(t, "demo", "pt_a", now)
	b := mintGeneration(t, "demo", "pt_b", now.Add(time.Minute))
	if err := r.Create(ctx, a); err != nil {
		t.Fatalf("create a: %v", err)
	}
	if err := r.Create(ctx, b); err != nil {
		t.Fatalf("create b: %v", err)
	}
	gens, err := r.ListByProject(ctx, "demo")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(gens) != 2 {
		t.Fatalf("len: want 2, got %d", len(gens))
	}
	// Sortierung: aufsteigend nach CreatedAt.
	if gens[0].TokenID != "pt_a" || gens[1].TokenID != "pt_b" {
		t.Errorf("order: %v / %v", gens[0].TokenID, gens[1].TokenID)
	}
}

func TestInMemoryProjectTokenRepo_HashUnique(t *testing.T) {
	t.Parallel()
	r := inmemory.NewProjectTokenRepository()
	ctx := context.Background()
	now := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	a := mintGeneration(t, "demo", "pt_a", now)
	if err := r.Create(ctx, a); err != nil {
		t.Fatalf("create a: %v", err)
	}
	dup := a
	dup.TokenID = "pt_a_dup"
	if err := r.Create(ctx, dup); !errors.Is(err, domain.ErrAuthTokenInvalid) {
		t.Errorf("duplicate hash: want ErrAuthTokenInvalid, got %v", err)
	}
}

func TestInMemoryProjectTokenRepo_FindByHash(t *testing.T) {
	t.Parallel()
	r := inmemory.NewProjectTokenRepository()
	ctx := context.Background()
	a := mintGeneration(t, "demo", "pt_a", time.Now().UTC())
	if err := r.Create(ctx, a); err != nil {
		t.Fatalf("create: %v", err)
	}
	got, err := r.FindByHash(ctx, a.KeyHash)
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if got.TokenID != "pt_a" {
		t.Errorf("got %q", got.TokenID)
	}
	if _, err := r.FindByHash(ctx, "nope"); !errors.Is(err, driven.ErrProjectTokenNotFound) {
		t.Errorf("unknown: want ErrProjectTokenNotFound, got %v", err)
	}
}

func TestInMemoryProjectTokenRepo_SetGraceAndRevoke(t *testing.T) {
	t.Parallel()
	r := inmemory.NewProjectTokenRepository()
	ctx := context.Background()
	a := mintGeneration(t, "demo", "pt_a", time.Now().UTC())
	if err := r.Create(ctx, a); err != nil {
		t.Fatalf("create: %v", err)
	}
	graceUntil := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	if err := r.SetGraceUntil(ctx, "demo", "pt_a", graceUntil); err != nil {
		t.Fatalf("grace: %v", err)
	}
	got, err := r.FindByHash(ctx, a.KeyHash)
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if got.GraceUntil == nil || !got.GraceUntil.Equal(graceUntil) {
		t.Errorf("GraceUntil: %v", got.GraceUntil)
	}
	revokedAt := time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC)
	if err := r.Revoke(ctx, "demo", "pt_a", revokedAt); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	got, err = r.FindByHash(ctx, a.KeyHash)
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if got.RevokedAt == nil || !got.RevokedAt.Equal(revokedAt) {
		t.Errorf("RevokedAt: %v", got.RevokedAt)
	}
	if err := r.SetGraceUntil(ctx, "demo", "ghost", time.Now()); !errors.Is(err, driven.ErrProjectTokenNotFound) {
		t.Errorf("ghost grace: want ErrProjectTokenNotFound, got %v", err)
	}
	if err := r.Revoke(ctx, "demo", "ghost", time.Now()); !errors.Is(err, driven.ErrProjectTokenNotFound) {
		t.Errorf("ghost revoke: want ErrProjectTokenNotFound, got %v", err)
	}
}

func TestInMemoryProjectTokenRepo_DefensiveClone(t *testing.T) {
	t.Parallel()
	r := inmemory.NewProjectTokenRepository()
	ctx := context.Background()
	now := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	originalGrace := now.Add(time.Hour)
	mutableGrace := originalGrace
	gen := domain.ProjectTokenGeneration{
		TokenID: "pt_a", ProjectID: "demo",
		KeyHash:     "abc",
		Fingerprint: "abc...",
		NotBefore:   now,
		GraceUntil:  &mutableGrace,
		CreatedAt:   now,
	}
	if err := r.Create(ctx, gen); err != nil {
		t.Fatalf("create: %v", err)
	}
	// Mutate caller-side pointer-target. Repo must not follow.
	mutableGrace = originalGrace.Add(24 * time.Hour)
	got, err := r.FindByHash(ctx, "abc")
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if got.GraceUntil == nil || !got.GraceUntil.Equal(originalGrace) {
		t.Errorf("repo state must not follow caller mutation: got %v, want %v", got.GraceUntil, originalGrace)
	}
}

func TestInMemoryProjectTokenRepo_RejectsEmptyFields(t *testing.T) {
	t.Parallel()
	r := inmemory.NewProjectTokenRepository()
	ctx := context.Background()
	cases := []domain.ProjectTokenGeneration{
		{},
		{ProjectID: "demo"},
		{ProjectID: "demo", TokenID: "x"},
	}
	for i, c := range cases {
		if err := r.Create(ctx, c); !errors.Is(err, domain.ErrAuthTokenInvalid) {
			t.Errorf("[%d]: want ErrAuthTokenInvalid, got %v", i, err)
		}
	}
}

func TestInMemoryProjectTokenRepo_DuplicateTokenID(t *testing.T) {
	t.Parallel()
	r := inmemory.NewProjectTokenRepository()
	ctx := context.Background()
	a := mintGeneration(t, "demo", "pt_a", time.Now().UTC())
	if err := r.Create(ctx, a); err != nil {
		t.Fatalf("create: %v", err)
	}
	other := mintGeneration(t, "demo", "pt_a", time.Now().UTC().Add(time.Second))
	if err := r.Create(ctx, other); !errors.Is(err, domain.ErrAuthTokenInvalid) {
		t.Errorf("duplicate token-id: want ErrAuthTokenInvalid, got %v", err)
	}
}
