package domain_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// RAK-73: Project-Token-Generationen, Status-
// Evaluation, Hash-Vergleich in konstanter Zeit, Klartext-/Hash-/
// Fingerprint-Trennung. Tests laufen ohne HTTP, Storage oder Crypto-
// Library.

func mustGenerateGen(t *testing.T) domain.ProjectTokenMaterial {
	t.Helper()
	now := fixedNow()
	m, err := domain.GenerateProjectToken(
		"pt_001",
		"demo",
		now,
		nil,
		nil,
		nil,
		now,
	)
	if err != nil {
		t.Fatalf("GenerateProjectToken: %v", err)
	}
	return m
}

func TestGenerateProjectToken_FormatAndPrefix(t *testing.T) {
	t.Parallel()
	m := mustGenerateGen(t)
	if !strings.HasPrefix(m.Value, "mtr_pt_") {
		t.Errorf("Value missing prefix: %q", m.Value)
	}
	// Prefix (7) + Base64(32 Bytes) ohne Padding (43) = 50.
	if len(m.Value) != 50 {
		t.Errorf("Value length: want 50, got %d", len(m.Value))
	}
	if len(m.Generation.KeyHash) != 64 {
		t.Errorf("KeyHash length: want 64 hex chars, got %d", len(m.Generation.KeyHash))
	}
	// Fingerprint: 8 head + 3 dots + 4 tail = 15 Zeichen.
	if len(m.Generation.Fingerprint) != 15 {
		t.Errorf("Fingerprint length: want 15, got %d", len(m.Generation.Fingerprint))
	}
	if !strings.HasPrefix(m.Generation.Fingerprint, "mtr_pt_") {
		t.Errorf("Fingerprint missing prefix: %q", m.Generation.Fingerprint)
	}
	if m.Generation.TokenID != "pt_001" {
		t.Errorf("TokenID: want pt_001, got %q", m.Generation.TokenID)
	}
	if m.Generation.ProjectID != "demo" {
		t.Errorf("ProjectID: want demo, got %q", m.Generation.ProjectID)
	}
	if m.Generation.RevokedAt != nil {
		t.Errorf("RevokedAt must default to nil")
	}
}

func TestGenerateProjectToken_PersistedFieldsExcludePlaintext(t *testing.T) {
	t.Parallel()
	m := mustGenerateGen(t)
	// Persistente Sicht hat per Konstruktion kein Klartext-Feld.
	// Dieser Test pinnt, dass Hash und Fingerprint sich aus dem
	// Klartext ableiten lassen, aber die Persistenzsicht den Klartext
	// nicht trägt.
	if strings.Contains(m.Generation.KeyHash, m.Value) {
		t.Errorf("KeyHash leaks plaintext")
	}
	if m.Generation.Fingerprint == m.Value {
		t.Errorf("Fingerprint leaks full plaintext")
	}
}

func TestGenerateProjectToken_Uniqueness(t *testing.T) {
	t.Parallel()
	const N = 500
	values := make(map[string]struct{}, N)
	hashes := make(map[string]struct{}, N)
	now := fixedNow()
	for i := 0; i < N; i++ {
		m, err := domain.GenerateProjectToken("id", "demo", now, nil, nil, nil, now)
		if err != nil {
			t.Fatalf("[%d]: %v", i, err)
		}
		if _, dup := values[m.Value]; dup {
			t.Fatalf("duplicate value at %d", i)
		}
		values[m.Value] = struct{}{}
		if _, dup := hashes[m.Generation.KeyHash]; dup {
			t.Fatalf("duplicate hash at %d", i)
		}
		hashes[m.Generation.KeyHash] = struct{}{}
	}
}

func TestGenerateProjectToken_LifecycleFieldsAreCloned(t *testing.T) {
	t.Parallel()
	now := fixedNow()
	grace := now.Add(time.Hour)
	exp := now.Add(24 * time.Hour)
	rotated := "pt_old"
	m, err := domain.GenerateProjectToken("pt_001", "demo", now, &grace, &exp, &rotated, now)
	if err != nil {
		t.Fatalf("GenerateProjectToken: %v", err)
	}
	grace = grace.Add(time.Hour)
	exp = exp.Add(time.Hour)
	rotated = "pt_other"
	if !m.Generation.GraceUntil.Equal(now.Add(time.Hour)) {
		t.Errorf("GraceUntil not cloned: %v", m.Generation.GraceUntil)
	}
	if !m.Generation.ExpiresAt.Equal(now.Add(24 * time.Hour)) {
		t.Errorf("ExpiresAt not cloned: %v", m.Generation.ExpiresAt)
	}
	if m.Generation.RotatedFrom == nil || *m.Generation.RotatedFrom != "pt_old" {
		t.Errorf("RotatedFrom not cloned: %v", m.Generation.RotatedFrom)
	}
}

func TestEvaluateProjectTokenStatus_TimeMatrix(t *testing.T) {
	t.Parallel()
	now := fixedNow()
	past := now.Add(-time.Hour)
	future := now.Add(time.Hour)
	cases := []struct {
		name string
		gen  domain.ProjectTokenGeneration
		want domain.ProjectTokenGenerationStatus
	}{
		{
			name: "active when only NotBefore in past",
			gen:  domain.ProjectTokenGeneration{NotBefore: past},
			want: domain.ProjectTokenStatusActive,
		},
		{
			name: "not_yet_valid when NotBefore in future",
			gen:  domain.ProjectTokenGeneration{NotBefore: future},
			want: domain.ProjectTokenStatusNotYetValid,
		},
		{
			name: "expired when ExpiresAt at or before now",
			gen:  domain.ProjectTokenGeneration{NotBefore: past, ExpiresAt: &now},
			want: domain.ProjectTokenStatusExpired,
		},
		{
			name: "revoked overrides grace",
			gen: domain.ProjectTokenGeneration{
				NotBefore:  past,
				GraceUntil: &future,
				RevokedAt:  &past,
			},
			want: domain.ProjectTokenStatusRevoked,
		},
		{
			name: "grace when GraceUntil in future and not revoked",
			gen:  domain.ProjectTokenGeneration{NotBefore: past, GraceUntil: &future},
			want: domain.ProjectTokenStatusGrace,
		},
		{
			name: "grace expired without ExpiresAt becomes expired",
			gen:  domain.ProjectTokenGeneration{NotBefore: past, GraceUntil: &past},
			want: domain.ProjectTokenStatusExpired,
		},
		{
			name: "future-revoked is still active until revocation kicks in",
			gen:  domain.ProjectTokenGeneration{NotBefore: past, RevokedAt: &future},
			want: domain.ProjectTokenStatusActive,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := domain.EvaluateProjectTokenStatus(tc.gen, now); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestStatusToAuthError(t *testing.T) {
	t.Parallel()
	cases := map[domain.ProjectTokenGenerationStatus]error{
		domain.ProjectTokenStatusActive:      nil,
		domain.ProjectTokenStatusGrace:       nil,
		domain.ProjectTokenStatusRevoked:     domain.ErrAuthTokenRevoked,
		domain.ProjectTokenStatusExpired:     domain.ErrAuthTokenExpired,
		domain.ProjectTokenStatusNotYetValid: domain.ErrAuthTokenNotYetValid,
		domain.ProjectTokenGenerationStatus("garbage"): domain.ErrAuthTokenInvalid,
	}
	for status, want := range cases {
		got := domain.StatusToAuthError(status)
		if want == nil {
			if got != nil {
				t.Errorf("status %q: want nil, got %v", status, got)
			}
			continue
		}
		if !errors.Is(got, want) {
			t.Errorf("status %q: want %v, got %v", status, want, got)
		}
	}
}

func TestProjectTokenStatus_CanAuthenticate(t *testing.T) {
	t.Parallel()
	cases := map[domain.ProjectTokenGenerationStatus]bool{
		domain.ProjectTokenStatusActive:       true,
		domain.ProjectTokenStatusGrace:        true,
		domain.ProjectTokenStatusRevoked:      false,
		domain.ProjectTokenStatusExpired:      false,
		domain.ProjectTokenStatusNotYetValid:  false,
		domain.ProjectTokenGenerationStatus("?"): false,
	}
	for status, want := range cases {
		if got := status.CanAuthenticate(); got != want {
			t.Errorf("status %q: want %v, got %v", status, want, got)
		}
	}
}

func TestValidateProjectTokenString_Roundtrip(t *testing.T) {
	t.Parallel()
	now := fixedNow()
	a, _ := domain.GenerateProjectToken("pt_a", "demo", now, nil, nil, nil, now)
	b, _ := domain.GenerateProjectToken("pt_b", "demo", now, nil, nil, nil, now)
	gens := []domain.ProjectTokenGeneration{a.Generation, b.Generation}
	gen, status, err := domain.ValidateProjectTokenString(b.Value, gens, now)
	if err != nil {
		t.Fatalf("matching token must validate: %v", err)
	}
	if gen.TokenID != "pt_b" {
		t.Errorf("expected pt_b, got %q", gen.TokenID)
	}
	if status != domain.ProjectTokenStatusActive {
		t.Errorf("expected active, got %q", status)
	}
}

func TestValidateProjectTokenString_RejectsMissingPrefix(t *testing.T) {
	t.Parallel()
	now := fixedNow()
	a, _ := domain.GenerateProjectToken("pt_a", "demo", now, nil, nil, nil, now)
	gens := []domain.ProjectTokenGeneration{a.Generation}
	if _, _, err := domain.ValidateProjectTokenString("legacy-demo-token", gens, now); !errors.Is(err, domain.ErrAuthTokenInvalid) {
		t.Errorf("missing prefix: want ErrAuthTokenInvalid, got %v", err)
	}
	if _, _, err := domain.ValidateProjectTokenString("", gens, now); !errors.Is(err, domain.ErrAuthTokenInvalid) {
		t.Errorf("empty: want ErrAuthTokenInvalid, got %v", err)
	}
}

func TestValidateProjectTokenString_RejectsUnknownHash(t *testing.T) {
	t.Parallel()
	now := fixedNow()
	a, _ := domain.GenerateProjectToken("pt_a", "demo", now, nil, nil, nil, now)
	gens := []domain.ProjectTokenGeneration{a.Generation}
	if _, _, err := domain.ValidateProjectTokenString("mtr_pt_AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", gens, now); !errors.Is(err, domain.ErrAuthTokenInvalid) {
		t.Errorf("unknown hash: want ErrAuthTokenInvalid, got %v", err)
	}
}

func TestValidateProjectTokenString_PropagatesLifecycleErrors(t *testing.T) {
	t.Parallel()
	now := fixedNow()
	past := now.Add(-time.Hour)
	a, _ := domain.GenerateProjectToken("pt_revoked", "demo", past, nil, nil, nil, past)
	a.Generation.RevokedAt = &past
	gens := []domain.ProjectTokenGeneration{a.Generation}
	_, status, err := domain.ValidateProjectTokenString(a.Value, gens, now)
	if !errors.Is(err, domain.ErrAuthTokenRevoked) {
		t.Errorf("revoked: want ErrAuthTokenRevoked, got %v", err)
	}
	if status != domain.ProjectTokenStatusRevoked {
		t.Errorf("status: want revoked, got %q", status)
	}
}

func TestValidateProjectTokenString_GraceTokenAuthenticates(t *testing.T) {
	t.Parallel()
	now := fixedNow()
	past := now.Add(-time.Hour)
	graceUntil := now.Add(time.Hour)
	a, _ := domain.GenerateProjectToken("pt_old", "demo", past, &graceUntil, nil, nil, past)
	gens := []domain.ProjectTokenGeneration{a.Generation}
	_, status, err := domain.ValidateProjectTokenString(a.Value, gens, now)
	if err != nil {
		t.Fatalf("grace token must authenticate: %v", err)
	}
	if status != domain.ProjectTokenStatusGrace {
		t.Errorf("status: want grace, got %q", status)
	}
}

func TestHasProjectTokenPrefix(t *testing.T) {
	t.Parallel()
	if !domain.HasProjectTokenPrefix("mtr_pt_abcdef") {
		t.Errorf("must detect project token prefix")
	}
	if domain.HasProjectTokenPrefix("mtr_st_abcdef") {
		t.Errorf("must not match session token prefix")
	}
	if domain.HasProjectTokenPrefix("demo-token") {
		t.Errorf("must not match legacy token")
	}
}

func TestFingerprintProjectTokenValue_DoesNotLeakPlaintext(t *testing.T) {
	t.Parallel()
	m := mustGenerateGen(t)
	fp := domain.FingerprintProjectTokenValue(m.Value)
	if fp == m.Value {
		t.Errorf("fingerprint leaks plaintext")
	}
	if fp != m.Generation.Fingerprint {
		t.Errorf("on-the-fly fingerprint must equal stored fingerprint: %q vs %q", fp, m.Generation.Fingerprint)
	}
	if !strings.Contains(fp, "...") {
		t.Errorf("fingerprint must contain ellipsis: %q", fp)
	}
}
