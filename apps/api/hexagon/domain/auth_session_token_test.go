package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// RAK-72: Session-Token-Claims, Audience-,
// Project-, Session- und Origin-Bindung, Signing-Key-Lookup und
// Konstantzeit-Vergleich. Tests laufen ohne HTTP, JSON, Storage oder
// Crypto-Library.

func fixedNow() time.Time {
	return time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
}

func ptr(s string) *string { return &s }

func baseClaims(t *testing.T) domain.SessionTokenClaims {
	t.Helper()
	now := fixedNow()
	return domain.BuildSessionTokenClaims(domain.SessionTokenIssuanceInput{
		ProjectID:  "demo",
		Audience:   domain.SessionTokenAudiencePlaybackEvents,
		TTLSeconds: 60,
		SessionID:  ptr("sess_a"),
		Origin:     ptr("http://localhost:5173"),
	}, "st_test_001", "", now)
}

func TestSessionTokenAudience_IsKnown(t *testing.T) {
	t.Parallel()
	if !domain.SessionTokenAudiencePlaybackEvents.IsKnown() {
		t.Errorf("playback-events must be known")
	}
	if domain.SessionTokenAudience("admin").IsKnown() {
		t.Errorf("unknown audience must not be allowlisted")
	}
	if domain.SessionTokenAudience("").IsKnown() {
		t.Errorf("empty audience must not be allowlisted")
	}
}

func TestSigningKeyAlgorithm_IsKnown(t *testing.T) {
	t.Parallel()
	if !domain.SigningKeyAlgorithmHS256.IsKnown() {
		t.Errorf("HS256 must be allowlisted")
	}
	if domain.SigningKeyAlgorithm("RS256").IsKnown() {
		t.Errorf("RS256 must not be allowlisted in 0.12.0")
	}
}

func TestBuildSessionTokenClaims_Defaults(t *testing.T) {
	t.Parallel()
	c := baseClaims(t)
	if c.Iss != domain.DefaultSessionTokenIssuer {
		t.Errorf("issuer default: want %q, got %q", domain.DefaultSessionTokenIssuer, c.Iss)
	}
	if c.Sub != "demo" {
		t.Errorf("sub: want demo, got %q", c.Sub)
	}
	if c.Aud != domain.SessionTokenAudiencePlaybackEvents {
		t.Errorf("aud: want playback-events, got %q", c.Aud)
	}
	if !c.Iat.Equal(fixedNow()) {
		t.Errorf("iat: want fixedNow, got %v", c.Iat)
	}
	if !c.Nbf.Equal(fixedNow()) {
		t.Errorf("nbf: want fixedNow, got %v", c.Nbf)
	}
	want := fixedNow().Add(60 * time.Second)
	if !c.Exp.Equal(want) {
		t.Errorf("exp: want %v, got %v", want, c.Exp)
	}
	if c.JTI != "st_test_001" {
		t.Errorf("jti: want st_test_001, got %q", c.JTI)
	}
	if c.SessionID == nil || *c.SessionID != "sess_a" {
		t.Errorf("session_id: want sess_a, got %v", c.SessionID)
	}
	if c.Origin == nil || *c.Origin != "http://localhost:5173" {
		t.Errorf("origin: want localhost:5173, got %v", c.Origin)
	}
}

func TestBuildSessionTokenClaims_TokenIDEqualsJTI(t *testing.T) {
	t.Parallel()
	c := baseClaims(t)
	if c.TokenID() != c.JTI {
		t.Errorf("token_id must equal jti: token_id=%q jti=%q", c.TokenID(), c.JTI)
	}
}

func TestBuildSessionTokenClaims_OptionalFieldsAreCloned(t *testing.T) {
	t.Parallel()
	sess := "sess_a"
	origin := "http://localhost:5173"
	in := domain.SessionTokenIssuanceInput{
		ProjectID:  "demo",
		Audience:   domain.SessionTokenAudiencePlaybackEvents,
		TTLSeconds: 60,
		SessionID:  &sess,
		Origin:     &origin,
	}
	c := domain.BuildSessionTokenClaims(in, "st_001", "", fixedNow())
	sess = "mutated"
	origin = "http://other"
	if c.SessionID == nil || *c.SessionID != "sess_a" {
		t.Errorf("session_id must be cloned: got %v", c.SessionID)
	}
	if c.Origin == nil || *c.Origin != "http://localhost:5173" {
		t.Errorf("origin must be cloned: got %v", c.Origin)
	}
}

func TestBuildSessionTokenClaims_CustomIssuer(t *testing.T) {
	t.Parallel()
	c := domain.BuildSessionTokenClaims(domain.SessionTokenIssuanceInput{
		ProjectID:  "demo",
		Audience:   domain.SessionTokenAudiencePlaybackEvents,
		TTLSeconds: 60,
	}, "st_001", "auth.example.com", fixedNow())
	if c.Iss != "auth.example.com" {
		t.Errorf("custom issuer: want auth.example.com, got %q", c.Iss)
	}
}

func TestValidateClaimsTime_Boundaries(t *testing.T) {
	t.Parallel()
	c := baseClaims(t)
	now := fixedNow()
	cases := []struct {
		name string
		when time.Time
		want error
	}{
		{"now == nbf is valid", now, nil},
		{"between nbf and exp is valid", now.Add(30 * time.Second), nil},
		{"now == exp is expired", now.Add(60 * time.Second), domain.ErrAuthTokenExpired},
		{"after exp is expired", now.Add(61 * time.Second), domain.ErrAuthTokenExpired},
		{"before nbf is not yet valid", now.Add(-1 * time.Second), domain.ErrAuthTokenNotYetValid},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := domain.ValidateClaimsTime(c, tc.when)
			if tc.want == nil {
				if err != nil {
					t.Errorf("unexpected err: %v", err)
				}
				return
			}
			if !errors.Is(err, tc.want) {
				t.Errorf("err: want %v, got %v", tc.want, err)
			}
		})
	}
}

func TestValidateClaimsAudience(t *testing.T) {
	t.Parallel()
	c := baseClaims(t)
	if err := domain.ValidateClaimsAudience(c, domain.SessionTokenAudiencePlaybackEvents); err != nil {
		t.Errorf("matching aud must succeed: %v", err)
	}
	if err := domain.ValidateClaimsAudience(c, domain.SessionTokenAudience("admin")); !errors.Is(err, domain.ErrAuthSessionScopeDenied) {
		t.Errorf("unknown expected aud: want ErrAuthSessionScopeDenied, got %v", err)
	}
	// Claim mit unbekannter Audience.
	cBad := c
	cBad.Aud = domain.SessionTokenAudience("admin")
	if err := domain.ValidateClaimsAudience(cBad, domain.SessionTokenAudiencePlaybackEvents); !errors.Is(err, domain.ErrAuthSessionScopeDenied) {
		t.Errorf("unknown claim aud: want ErrAuthSessionScopeDenied, got %v", err)
	}
}

func TestValidateClaimsProject(t *testing.T) {
	t.Parallel()
	c := baseClaims(t)
	if err := domain.ValidateClaimsProject(c, "demo"); err != nil {
		t.Errorf("matching project must succeed: %v", err)
	}
	if err := domain.ValidateClaimsProject(c, "other"); !errors.Is(err, domain.ErrAuthProjectMismatch) {
		t.Errorf("mismatch: want ErrAuthProjectMismatch, got %v", err)
	}
	cBad := c
	cBad.Sub = ""
	if err := domain.ValidateClaimsProject(cBad, "demo"); !errors.Is(err, domain.ErrAuthProjectMismatch) {
		t.Errorf("empty sub: want ErrAuthProjectMismatch, got %v", err)
	}
}

func TestValidateClaimsSession(t *testing.T) {
	t.Parallel()
	c := baseClaims(t)
	if err := domain.ValidateClaimsSession(c, "sess_a"); err != nil {
		t.Errorf("matching session must succeed: %v", err)
	}
	if err := domain.ValidateClaimsSession(c, "sess_b"); !errors.Is(err, domain.ErrAuthSessionScopeDenied) {
		t.Errorf("session mismatch: want ErrAuthSessionScopeDenied, got %v", err)
	}
	if err := domain.ValidateClaimsSession(c, ""); !errors.Is(err, domain.ErrAuthSessionScopeDenied) {
		t.Errorf("missing request session when claim binds session: want ErrAuthSessionScopeDenied, got %v", err)
	}
	cNoSession := c
	cNoSession.SessionID = nil
	if err := domain.ValidateClaimsSession(cNoSession, ""); err != nil {
		t.Errorf("unbound session token must accept any session: %v", err)
	}
	if err := domain.ValidateClaimsSession(cNoSession, "sess_b"); err != nil {
		t.Errorf("unbound session token must accept any session: %v", err)
	}
}

func TestValidateClaimsOrigin(t *testing.T) {
	t.Parallel()
	c := baseClaims(t)
	if err := domain.ValidateClaimsOrigin(c, "http://localhost:5173"); err != nil {
		t.Errorf("matching origin must succeed: %v", err)
	}
	if err := domain.ValidateClaimsOrigin(c, "http://other"); !errors.Is(err, domain.ErrAuthSessionScopeDenied) {
		t.Errorf("origin mismatch: want ErrAuthSessionScopeDenied, got %v", err)
	}
	if err := domain.ValidateClaimsOrigin(c, ""); !errors.Is(err, domain.ErrAuthSessionScopeDenied) {
		t.Errorf("missing origin when claim binds origin: want ErrAuthSessionScopeDenied, got %v", err)
	}
	cNoOrigin := c
	cNoOrigin.Origin = nil
	if err := domain.ValidateClaimsOrigin(cNoOrigin, "http://anything"); err != nil {
		t.Errorf("unbound origin token must accept any origin: %v", err)
	}
}

func TestLookupSigningKey(t *testing.T) {
	t.Parallel()
	keys := []domain.SessionSigningKey{
		{KID: "key_2026_05", Algorithm: domain.SigningKeyAlgorithmHS256, Secret: []byte("a")},
		{KID: "key_2026_04", Algorithm: domain.SigningKeyAlgorithmHS256, Secret: []byte("b")},
	}
	got, err := domain.LookupSigningKey(keys, "key_2026_05")
	if err != nil {
		t.Fatalf("known kid: %v", err)
	}
	if got.KID != "key_2026_05" {
		t.Errorf("got %q, want key_2026_05", got.KID)
	}
	if _, err := domain.LookupSigningKey(keys, "unknown"); !errors.Is(err, domain.ErrAuthTokenInvalid) {
		t.Errorf("unknown kid: want ErrAuthTokenInvalid, got %v", err)
	}
	if _, err := domain.LookupSigningKey(keys, ""); !errors.Is(err, domain.ErrAuthTokenInvalid) {
		t.Errorf("empty kid: want ErrAuthTokenInvalid, got %v", err)
	}
}

func TestConstantTimeEqualSignature(t *testing.T) {
	t.Parallel()
	a := []byte("abc123")
	b := []byte("abc123")
	if !domain.ConstantTimeEqualSignature(a, b) {
		t.Errorf("identical bytes must be equal")
	}
	if domain.ConstantTimeEqualSignature(a, []byte("abc124")) {
		t.Errorf("differing bytes must not be equal")
	}
	// Unterschiedliche Längen: subtle.ConstantTimeCompare liefert 0,
	// also false. Wichtig, weil ein Längenmismatch sonst eine
	// Side-Channel-Lücke wäre.
	if domain.ConstantTimeEqualSignature(a, []byte("abc")) {
		t.Errorf("differing lengths must not be equal")
	}
	// Pinne das Verhalten an den Empty-Slice-Grenzen: zwei leere
	// Slices sind unter `subtle.ConstantTimeCompare` gleich (Rückgabe
	// 1). Das ist für Auth-Verify-Pfade akzeptabel, weil ein vom
	// Aufrufer übergebener leerer Signature-Bytes-Wert vorher im
	// Adapter abgefangen wird (Token-Decode liefert dann
	// `ErrAuthTokenInvalid`).
	if !domain.ConstantTimeEqualSignature(nil, nil) {
		t.Errorf("two nil slices: subtle.ConstantTimeCompare returns 1")
	}
	if !domain.ConstantTimeEqualSignature([]byte{}, []byte{}) {
		t.Errorf("two empty slices: subtle.ConstantTimeCompare returns 1")
	}
	if domain.ConstantTimeEqualSignature(nil, []byte("x")) {
		t.Errorf("nil vs non-empty must not be equal")
	}
}

func TestHasSessionTokenPrefix(t *testing.T) {
	t.Parallel()
	if !domain.HasSessionTokenPrefix("mtr_st_abc.def.ghi") {
		t.Errorf("must detect session token prefix")
	}
	if domain.HasSessionTokenPrefix("mtr_pt_abc") {
		t.Errorf("must not match project token prefix")
	}
	if domain.HasSessionTokenPrefix("demo-token") {
		t.Errorf("must not match legacy token")
	}
	if domain.HasSessionTokenPrefix("") {
		t.Errorf("must not match empty")
	}
}
