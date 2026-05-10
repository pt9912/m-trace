package http_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/auth"
	apihttp "github.com/pt9912/m-trace/apps/api/adapters/driving/http"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// `0.12.0` Tranche 2 / RAK-72/RAK-75: Auth-Header-Parser. Tests
// fahren den HMAC-Signer (Sign-Pfad ausschließlich für Testdaten,
// damit Verify einen echten Token bekommt).

const parserSigningSecret = "parser-test-signing-secret"

func newParserStack(t *testing.T) (apihttp.AuthHeaderParser, *auth.HMACSessionTokenSigner) {
	t.Helper()
	resolver := auth.NewStaticProjectResolver(map[string]auth.ProjectConfig{
		"demo":  {Token: "demo-token", AllowedOrigins: []string{"http://localhost:5173"}},
		"other": {Token: "other-token", AllowedOrigins: []string{"http://other.example"}},
	})
	keyRing, err := auth.NewStaticSigningKeyResolver("kid_a", domain.SessionSigningKey{
		KID: "kid_a", Algorithm: domain.SigningKeyAlgorithmHS256,
		Secret: []byte(parserSigningSecret),
	})
	if err != nil {
		t.Fatalf("key ring: %v", err)
	}
	signer := auth.NewHMACSessionTokenSigner(keyRing)
	now := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	parser := apihttp.AuthHeaderParser{
		Resolver: resolver,
		Verifier: signer,
		Projects: resolver,
		Now:      func() time.Time { return now },
		Audience: domain.SessionTokenAudiencePlaybackEvents,
	}
	return parser, signer
}

func mintSessionToken(t *testing.T, signer *auth.HMACSessionTokenSigner, sub string, ttl time.Duration, sessionID, origin string) string {
	t.Helper()
	now := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	var sess, og *string
	if sessionID != "" {
		sess = &sessionID
	}
	if origin != "" {
		og = &origin
	}
	tok, err := signer.Sign(domain.SessionTokenClaims{
		Iss:       "m-trace",
		Sub:       sub,
		Aud:       domain.SessionTokenAudiencePlaybackEvents,
		Iat:       now,
		Nbf:       now,
		Exp:       now.Add(ttl),
		JTI:       "st_test",
		SessionID: sess,
		Origin:    og,
	})
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	return tok
}

func headers(m map[string]string) func(string) string {
	return func(key string) string { return m[key] }
}

func TestAuthHeaderParser_LegacyOnly(t *testing.T) {
	t.Parallel()
	parser, _ := newParserStack(t)
	d, err := parser.Parse(context.Background(), headers(map[string]string{"X-MTrace-Token": "demo-token"}), "")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if d.LegacyToken != "demo-token" {
		t.Errorf("LegacyToken: got %q", d.LegacyToken)
	}
	if d.ResolvedProject != nil {
		t.Errorf("ResolvedProject must stay nil for legacy-only path")
	}
}

func TestAuthHeaderParser_BearerHappyPath(t *testing.T) {
	t.Parallel()
	parser, signer := newParserStack(t)
	tok := mintSessionToken(t, signer, "demo", time.Minute, "", "")
	d, err := parser.Parse(context.Background(), headers(map[string]string{"Authorization": "Bearer " + tok}), "")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if d.ResolvedProject == nil || d.ResolvedProject.ID != "demo" {
		t.Errorf("ResolvedProject: %+v", d.ResolvedProject)
	}
	if d.LegacyToken != "" {
		t.Errorf("LegacyToken must stay empty: %q", d.LegacyToken)
	}
}

func TestAuthHeaderParser_XMTraceSessionTokenHappyPath(t *testing.T) {
	t.Parallel()
	parser, signer := newParserStack(t)
	tok := mintSessionToken(t, signer, "demo", time.Minute, "", "")
	d, err := parser.Parse(context.Background(), headers(map[string]string{"X-MTrace-Session-Token": tok}), "")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if d.ResolvedProject == nil || d.ResolvedProject.ID != "demo" {
		t.Errorf("ResolvedProject: %+v", d.ResolvedProject)
	}
}

func TestAuthHeaderParser_ConflictingBearerAndXMTraceSessionToken(t *testing.T) {
	t.Parallel()
	parser, signer := newParserStack(t)
	tokA := mintSessionToken(t, signer, "demo", time.Minute, "", "")
	tokB := mintSessionToken(t, signer, "demo", 2*time.Minute, "", "")
	if tokA == tokB {
		t.Skip("nondeterministic: tokens equal")
	}
	_, err := parser.Parse(context.Background(), headers(map[string]string{
		"Authorization":          "Bearer " + tokA,
		"X-MTrace-Session-Token": tokB,
	}), "")
	if !errors.Is(err, domain.ErrAuthTokenInvalid) {
		t.Errorf("conflicting bearer + session: want ErrAuthTokenInvalid, got %v", err)
	}
}

func TestAuthHeaderParser_BearerPlusLegacySameProject(t *testing.T) {
	t.Parallel()
	parser, signer := newParserStack(t)
	tok := mintSessionToken(t, signer, "demo", time.Minute, "", "")
	d, err := parser.Parse(context.Background(), headers(map[string]string{
		"Authorization":  "Bearer " + tok,
		"X-MTrace-Token": "demo-token",
	}), "")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if d.ResolvedProject == nil || d.ResolvedProject.ID != "demo" {
		t.Errorf("ResolvedProject: %+v", d.ResolvedProject)
	}
	if d.LegacyToken != "demo-token" {
		t.Errorf("LegacyToken: got %q", d.LegacyToken)
	}
}

func TestAuthHeaderParser_BearerPlusLegacyDifferentProject(t *testing.T) {
	t.Parallel()
	parser, signer := newParserStack(t)
	tok := mintSessionToken(t, signer, "demo", time.Minute, "", "")
	_, err := parser.Parse(context.Background(), headers(map[string]string{
		"Authorization":  "Bearer " + tok,
		"X-MTrace-Token": "other-token",
	}), "")
	if !errors.Is(err, domain.ErrAuthProjectMismatch) {
		t.Errorf("want ErrAuthProjectMismatch, got %v", err)
	}
}

func TestAuthHeaderParser_BearerPlusInvalidLegacy(t *testing.T) {
	t.Parallel()
	parser, signer := newParserStack(t)
	tok := mintSessionToken(t, signer, "demo", time.Minute, "", "")
	_, err := parser.Parse(context.Background(), headers(map[string]string{
		"Authorization":  "Bearer " + tok,
		"X-MTrace-Token": "garbage",
	}), "")
	if !errors.Is(err, domain.ErrAuthTokenInvalid) {
		t.Errorf("invalid legacy plus valid bearer: want ErrAuthTokenInvalid, got %v", err)
	}
}

func TestAuthHeaderParser_MalformedBearerBlocksLegacyFallback(t *testing.T) {
	t.Parallel()
	parser, _ := newParserStack(t)
	// `Bearer mtr_st_*` mit einem strukturell ungültigen Wert: Verify
	// scheitert → ErrAuthTokenInvalid (kein Fallback auf Legacy).
	_, err := parser.Parse(context.Background(), headers(map[string]string{
		"Authorization":  "Bearer mtr_st_garbage",
		"X-MTrace-Token": "demo-token",
	}), "")
	if !errors.Is(err, domain.ErrAuthTokenInvalid) {
		t.Errorf("malformed bearer must block fallback: want ErrAuthTokenInvalid, got %v", err)
	}
}

func TestAuthHeaderParser_ForeignAuthorizationIgnoredWithLegacy(t *testing.T) {
	t.Parallel()
	parser, _ := newParserStack(t)
	d, err := parser.Parse(context.Background(), headers(map[string]string{
		"Authorization":  "Basic dXNlcjpwYXNz",
		"X-MTrace-Token": "demo-token",
	}), "")
	if err != nil {
		t.Fatalf("foreign Authorization plus legacy must pass: %v", err)
	}
	if d.LegacyToken != "demo-token" {
		t.Errorf("LegacyToken: got %q", d.LegacyToken)
	}
}

func TestAuthHeaderParser_ForeignAuthorizationOnlyIsMissing(t *testing.T) {
	t.Parallel()
	parser, _ := newParserStack(t)
	_, err := parser.Parse(context.Background(), headers(map[string]string{
		"Authorization": "Basic dXNlcjpwYXNz",
	}), "")
	if !errors.Is(err, domain.ErrAuthTokenMissing) {
		t.Errorf("foreign-only: want ErrAuthTokenMissing, got %v", err)
	}
}

func TestAuthHeaderParser_NoHeadersReturnsMissing(t *testing.T) {
	t.Parallel()
	parser, _ := newParserStack(t)
	_, err := parser.Parse(context.Background(), headers(nil), "")
	if !errors.Is(err, domain.ErrAuthTokenMissing) {
		t.Errorf("want ErrAuthTokenMissing, got %v", err)
	}
}

func TestAuthHeaderParser_OriginMismatchOnSessionToken(t *testing.T) {
	t.Parallel()
	parser, signer := newParserStack(t)
	tok := mintSessionToken(t, signer, "demo", time.Minute, "", "http://localhost:5173")
	_, err := parser.Parse(context.Background(), headers(map[string]string{
		"Authorization": "Bearer " + tok,
	}), "http://other.example")
	if !errors.Is(err, domain.ErrAuthSessionScopeDenied) {
		t.Errorf("origin mismatch: want ErrAuthSessionScopeDenied, got %v", err)
	}
}

func TestAuthHeaderParser_ExpiredSessionToken(t *testing.T) {
	t.Parallel()
	parser, signer := newParserStack(t)
	// Negative TTL produziert einen Token, dessen Exp vor `Now` liegt.
	tok := mintSessionToken(t, signer, "demo", -time.Minute, "", "")
	_, err := parser.Parse(context.Background(), headers(map[string]string{
		"Authorization": "Bearer " + tok,
	}), "")
	if !errors.Is(err, domain.ErrAuthTokenExpired) {
		t.Errorf("expired: want ErrAuthTokenExpired, got %v", err)
	}
}

func TestAuthHeaderParser_UnknownProjectInClaims(t *testing.T) {
	t.Parallel()
	parser, signer := newParserStack(t)
	tok := mintSessionToken(t, signer, "ghost", time.Minute, "", "")
	_, err := parser.Parse(context.Background(), headers(map[string]string{
		"Authorization": "Bearer " + tok,
	}), "")
	if !errors.Is(err, domain.ErrAuthTokenInvalid) {
		t.Errorf("unknown project in sub: want ErrAuthTokenInvalid, got %v", err)
	}
}

func TestAuthHeaderParser_BearerSchemeIsCaseInsensitive(t *testing.T) {
	t.Parallel()
	parser, signer := newParserStack(t)
	tok := mintSessionToken(t, signer, "demo", time.Minute, "", "")
	cases := []string{"Bearer ", "bearer ", "BEARER ", "BeArEr "}
	for _, scheme := range cases {
		t.Run(scheme, func(t *testing.T) {
			d, err := parser.Parse(context.Background(), headers(map[string]string{"Authorization": scheme + tok}), "")
			if err != nil {
				t.Fatalf("scheme %q: %v", scheme, err)
			}
			if d.ResolvedProject == nil || d.ResolvedProject.ID != "demo" {
				t.Errorf("scheme %q: ResolvedProject %+v", scheme, d.ResolvedProject)
			}
		})
	}
}

type lifecycleStubResolver struct {
	err error
}

func (s lifecycleStubResolver) ResolveByToken(_ context.Context, _ string) (domain.Project, error) {
	return domain.Project{}, s.err
}

func TestAuthHeaderParser_LegacyResolverPropagatesAuthLifecycleErrors(t *testing.T) {
	t.Parallel()
	cases := map[string]error{
		"revoked":          domain.ErrAuthTokenRevoked,
		"expired":          domain.ErrAuthTokenExpired,
		"not_yet_valid":    domain.ErrAuthTokenNotYetValid,
		"scope_denied":     domain.ErrAuthSessionScopeDenied,
		"policy_denied":    domain.ErrAuthPolicyDenied,
	}
	for name, want := range cases {
		t.Run(name, func(t *testing.T) {
			parser := apihttp.AuthHeaderParser{
				Resolver: lifecycleStubResolver{err: want},
			}
			_, err := parser.Parse(context.Background(), headers(map[string]string{"X-MTrace-Token": "anything"}), "")
			if !errors.Is(err, want) {
				t.Errorf("want %v, got %v", want, err)
			}
		})
	}
}

func TestAuthHeaderParser_LegacyResolverGenericErrorMappedToInvalid(t *testing.T) {
	t.Parallel()
	parser := apihttp.AuthHeaderParser{
		Resolver: lifecycleStubResolver{err: domain.ErrUnauthorized},
	}
	_, err := parser.Parse(context.Background(), headers(map[string]string{"X-MTrace-Token": "anything"}), "")
	if !errors.Is(err, domain.ErrAuthTokenInvalid) {
		t.Errorf("generic resolver error must map to ErrAuthTokenInvalid: got %v", err)
	}
}

func TestAuthHeaderParser_DefaultClockResolvesValidToken(t *testing.T) {
	t.Parallel()
	// Parser ohne `Now`-Override → fällt auf `time.Now().UTC()` zurück.
	// Token ist frisch ausgestellt mit Exp = now+1h → muss validieren.
	resolver := auth.NewStaticProjectResolver(map[string]auth.ProjectConfig{
		"demo": {Token: "demo-token"},
	})
	keyRing, err := auth.NewStaticSigningKeyResolver("kid_a", domain.SessionSigningKey{
		KID: "kid_a", Algorithm: domain.SigningKeyAlgorithmHS256, Secret: []byte("secret"),
	})
	if err != nil {
		t.Fatalf("ring: %v", err)
	}
	signer := auth.NewHMACSessionTokenSigner(keyRing)
	parser := apihttp.AuthHeaderParser{
		Resolver: resolver,
		Verifier: signer,
		Projects: resolver,
		// Now bewusst nicht gesetzt → default-Clock-Pfad.
	}
	now := time.Now().UTC()
	tok, err := signer.Sign(domain.SessionTokenClaims{
		Iss: "m-trace", Sub: "demo",
		Aud: domain.SessionTokenAudiencePlaybackEvents,
		Iat: now, Nbf: now, Exp: now.Add(time.Hour), JTI: "st_x",
	})
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	d, err := parser.Parse(context.Background(), headers(map[string]string{"Authorization": "Bearer " + tok}), "")
	if err != nil {
		t.Fatalf("default-clock parse: %v", err)
	}
	if d.ResolvedProject == nil || d.ResolvedProject.ID != "demo" {
		t.Errorf("project: %+v", d.ResolvedProject)
	}
}

func TestAuthHeaderParser_DisabledVerifierRejectsSessionToken(t *testing.T) {
	t.Parallel()
	resolver := auth.NewStaticProjectResolver(map[string]auth.ProjectConfig{
		"demo": {Token: "demo-token"},
	})
	parser := apihttp.AuthHeaderParser{
		Resolver: resolver,
		// No Verifier → session-token path disabled.
	}
	_, err := parser.Parse(context.Background(), headers(map[string]string{
		"Authorization": "Bearer mtr_st_anything",
	}), "")
	if !errors.Is(err, domain.ErrAuthTokenInvalid) {
		t.Errorf("disabled verifier: want ErrAuthTokenInvalid, got %v", err)
	}
}
