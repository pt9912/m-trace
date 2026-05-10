package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/auth"
	"github.com/pt9912/m-trace/apps/api/adapters/driven/persistence/inmemory"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// `0.12.0` Tranche 3 / RAK-73: RotatingProjectResolver — Lifecycle
// (active → grace → revoked) plus Legacy-Fallback auf den Static-
// Resolver für `demo-token`.

func newRotatingResolver(t *testing.T, now time.Time) (*auth.RotatingProjectResolver, *inmemory.ProjectTokenRepository) {
	t.Helper()
	staticResolver := auth.NewStaticProjectResolver(map[string]auth.ProjectConfig{
		"demo": {Token: "demo-token", AllowedOrigins: []string{"http://x"}},
	})
	repo := inmemory.NewProjectTokenRepository()
	resolver := auth.NewRotatingProjectResolver(repo, staticResolver, staticResolver)
	resolver.Now = func() time.Time { return now }
	return resolver, repo
}

func TestRotatingResolver_LegacyFallback(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	resolver, _ := newRotatingResolver(t, now)
	project, err := resolver.ResolveByToken(context.Background(), "demo-token")
	if err != nil {
		t.Fatalf("legacy: %v", err)
	}
	if project.ID != "demo" {
		t.Errorf("project: want demo, got %q", project.ID)
	}
}

func TestRotatingResolver_LegacyUnknownReturnsUnauthorized(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	resolver, _ := newRotatingResolver(t, now)
	if _, err := resolver.ResolveByToken(context.Background(), "garbage"); !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("unknown legacy: want ErrUnauthorized, got %v", err)
	}
}

func TestRotatingResolver_RotationActive(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	resolver, repo := newRotatingResolver(t, now)
	mat, err := domain.GenerateProjectToken("pt_a", "demo", now.Add(-time.Hour), nil, nil, nil, now.Add(-time.Hour))
	if err != nil {
		t.Fatalf("mint: %v", err)
	}
	if err := repo.Create(context.Background(), mat.Generation); err != nil {
		t.Fatalf("create: %v", err)
	}
	project, err := resolver.ResolveByToken(context.Background(), mat.Value)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if project.ID != "demo" {
		t.Errorf("project: %q", project.ID)
	}
}

func TestRotatingResolver_RotationGraceStillAuthenticates(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	resolver, repo := newRotatingResolver(t, now)
	graceUntil := now.Add(time.Hour)
	mat, err := domain.GenerateProjectToken("pt_old", "demo", now.Add(-time.Hour), &graceUntil, nil, nil, now.Add(-time.Hour))
	if err != nil {
		t.Fatalf("mint: %v", err)
	}
	if err := repo.Create(context.Background(), mat.Generation); err != nil {
		t.Fatalf("create: %v", err)
	}
	project, err := resolver.ResolveByToken(context.Background(), mat.Value)
	if err != nil {
		t.Fatalf("grace token must authenticate: %v", err)
	}
	if project.ID != "demo" {
		t.Errorf("project: %q", project.ID)
	}
}

func TestRotatingResolver_RevokedReturnsRevokedError(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	resolver, repo := newRotatingResolver(t, now)
	mat, err := domain.GenerateProjectToken("pt_a", "demo", now.Add(-time.Hour), nil, nil, nil, now.Add(-time.Hour))
	if err != nil {
		t.Fatalf("mint: %v", err)
	}
	if err := repo.Create(context.Background(), mat.Generation); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := repo.Revoke(context.Background(), "demo", "pt_a", now.Add(-time.Minute)); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if _, err := resolver.ResolveByToken(context.Background(), mat.Value); !errors.Is(err, domain.ErrAuthTokenRevoked) {
		t.Errorf("revoked: want ErrAuthTokenRevoked, got %v", err)
	}
}

func TestRotatingResolver_ExpiredReturnsExpiredError(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	resolver, repo := newRotatingResolver(t, now)
	expiresAt := now.Add(-time.Minute)
	mat, err := domain.GenerateProjectToken("pt_a", "demo", now.Add(-time.Hour), nil, &expiresAt, nil, now.Add(-time.Hour))
	if err != nil {
		t.Fatalf("mint: %v", err)
	}
	if err := repo.Create(context.Background(), mat.Generation); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := resolver.ResolveByToken(context.Background(), mat.Value); !errors.Is(err, domain.ErrAuthTokenExpired) {
		t.Errorf("expired: want ErrAuthTokenExpired, got %v", err)
	}
}

func TestRotatingResolver_NotYetValid(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	resolver, repo := newRotatingResolver(t, now)
	mat, err := domain.GenerateProjectToken("pt_future", "demo", now.Add(time.Hour), nil, nil, nil, now)
	if err != nil {
		t.Fatalf("mint: %v", err)
	}
	if err := repo.Create(context.Background(), mat.Generation); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := resolver.ResolveByToken(context.Background(), mat.Value); !errors.Is(err, domain.ErrAuthTokenNotYetValid) {
		t.Errorf("not_yet_valid: want ErrAuthTokenNotYetValid, got %v", err)
	}
}

func TestRotatingResolver_RotationUnknownHash(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	resolver, _ := newRotatingResolver(t, now)
	if _, err := resolver.ResolveByToken(context.Background(), "mtr_pt_AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"); !errors.Is(err, domain.ErrAuthTokenInvalid) {
		t.Errorf("unknown hash: want ErrAuthTokenInvalid, got %v", err)
	}
}

func TestNewRotatingProjectResolver_DefaultClock(t *testing.T) {
	t.Parallel()
	staticResolver := auth.NewStaticProjectResolver(map[string]auth.ProjectConfig{
		"demo": {Token: "demo-token"},
	})
	repo := inmemory.NewProjectTokenRepository()
	r := auth.NewRotatingProjectResolver(repo, staticResolver, staticResolver)
	if r == nil {
		t.Fatal("ctor returned nil")
	}
	// Default-Clock zieht time.Now() — Resolver muss ohne Now-Override
	// funktionieren. Legacy-Pfad hat keine Zeit-Abhängigkeit, deshalb
	// pinnen wir den default über einen Demo-Token-Lookup.
	project, err := r.ResolveByToken(context.Background(), "demo-token")
	if err != nil {
		t.Fatalf("default clock resolve: %v", err)
	}
	if project.ID != "demo" {
		t.Errorf("project: %q", project.ID)
	}
}

func TestNewRotatingProjectResolver_DefaultClockOnRotatingPath(t *testing.T) {
	t.Parallel()
	staticResolver := auth.NewStaticProjectResolver(map[string]auth.ProjectConfig{
		"demo": {Token: "demo-token"},
	})
	repo := inmemory.NewProjectTokenRepository()
	r := auth.NewRotatingProjectResolver(repo, staticResolver, staticResolver)
	// Frische Generation — NotBefore deutlich in der Vergangenheit,
	// damit der default-Clock-Pfad in `r.now()` eine Active-Status-
	// Entscheidung trifft.
	mat, err := domain.GenerateProjectToken("pt_default", "demo", time.Now().UTC().Add(-time.Hour), nil, nil, nil, time.Now().UTC().Add(-time.Hour))
	if err != nil {
		t.Fatalf("mint: %v", err)
	}
	if err := repo.Create(context.Background(), mat.Generation); err != nil {
		t.Fatalf("create: %v", err)
	}
	project, err := r.ResolveByToken(context.Background(), mat.Value)
	if err != nil {
		t.Fatalf("default-clock mtr_pt_* resolve: %v", err)
	}
	if project.ID != "demo" {
		t.Errorf("project: %q", project.ID)
	}
}

func TestRotatingResolver_NilFallbackReturnsUnauthorized(t *testing.T) {
	t.Parallel()
	r := auth.NewRotatingProjectResolver(nil, nil, nil)
	if _, err := r.ResolveByToken(context.Background(), "demo-token"); !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("nil fallback: want ErrUnauthorized, got %v", err)
	}
	// mtr_pt_*-Token mit nil-Repo → invalid.
	if _, err := r.ResolveByToken(context.Background(), "mtr_pt_x"); !errors.Is(err, domain.ErrAuthTokenInvalid) {
		t.Errorf("nil repo on mtr_pt_*: want ErrAuthTokenInvalid, got %v", err)
	}
}

func TestRotatingResolver_RotationProjectGhost(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	resolver, repo := newRotatingResolver(t, now)
	// Token zeigt auf ein nicht im Static-Resolver konfiguriertes Project.
	mat, err := domain.GenerateProjectToken("pt_x", "ghost", now.Add(-time.Hour), nil, nil, nil, now.Add(-time.Hour))
	if err != nil {
		t.Fatalf("mint: %v", err)
	}
	if err := repo.Create(context.Background(), mat.Generation); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := resolver.ResolveByToken(context.Background(), mat.Value); !errors.Is(err, domain.ErrAuthTokenInvalid) {
		t.Errorf("ghost project: want ErrAuthTokenInvalid, got %v", err)
	}
}
