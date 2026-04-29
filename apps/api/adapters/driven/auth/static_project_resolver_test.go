package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/auth"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// TestStaticProjectResolver_ResolveByToken_KnownAndUnknown deckt beide
// Code-Pfade ab: Token bekannt → Project mit AllowedOrigins; unbekannt
// → ErrUnauthorized.
func TestStaticProjectResolver_ResolveByToken_KnownAndUnknown(t *testing.T) {
	t.Parallel()
	r := auth.NewStaticProjectResolver(map[string]auth.ProjectConfig{
		"demo": {
			Token:          "demo-token",
			AllowedOrigins: []string{"http://a.example", "http://b.example"},
		},
	})

	got, err := r.ResolveByToken(context.Background(), "demo-token")
	if err != nil {
		t.Fatalf("expected resolve, got %v", err)
	}
	if got.ID != "demo" {
		t.Errorf("ID=%q want demo", got.ID)
	}
	if len(got.AllowedOrigins) != 2 {
		t.Errorf("AllowedOrigins len=%d want 2", len(got.AllowedOrigins))
	}

	if _, err := r.ResolveByToken(context.Background(), "wrong-token"); !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

// TestStaticProjectResolver_IsOriginInGlobalUnion verifiziert die
// Aggregation der Allowed-Origins über alle Projects (Pflicht für
// CORS-Preflight, plan-0.1.0.md §5.1).
func TestStaticProjectResolver_IsOriginInGlobalUnion(t *testing.T) {
	t.Parallel()
	r := auth.NewStaticProjectResolver(map[string]auth.ProjectConfig{
		"a": {Token: "t1", AllowedOrigins: []string{"http://a.example"}},
		"b": {Token: "t2", AllowedOrigins: []string{"http://b.example"}},
	})

	cases := []struct {
		origin string
		want   bool
	}{
		{"http://a.example", true},
		{"http://b.example", true},
		{"http://c.example", false},
		{"", false}, // leerer Origin gehört nicht in die Union
	}
	for _, tc := range cases {
		if got := r.IsOriginInGlobalUnion(tc.origin); got != tc.want {
			t.Errorf("IsOriginInGlobalUnion(%q)=%v want %v", tc.origin, got, tc.want)
		}
	}
}

// TestStaticProjectResolver_AllowedOriginsAreCopied verifiziert, dass
// der Resolver die übergebene Slice nicht teilt — eine Mutation an der
// ursprünglichen Slice darf den Project-State nicht ändern.
func TestStaticProjectResolver_AllowedOriginsAreCopied(t *testing.T) {
	t.Parallel()
	origins := []string{"http://a.example"}
	r := auth.NewStaticProjectResolver(map[string]auth.ProjectConfig{
		"a": {Token: "t1", AllowedOrigins: origins},
	})
	origins[0] = "http://mutated.example"

	got, err := r.ResolveByToken(context.Background(), "t1")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if got.AllowedOrigins[0] != "http://a.example" {
		t.Errorf("AllowedOrigins shared with caller (got %q)", got.AllowedOrigins[0])
	}
}
