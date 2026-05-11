package auth_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/auth"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// `0.12.0` Tranche 2: Rand-Pfade der Auth-Adapter — Nil-Receiver,
// deaktivierte Buckets, Fallback auf BaseProject und ID-Format.

func TestRandomTokenIDGenerator_Format(t *testing.T) {
	t.Parallel()
	gen := auth.NewRandomTokenIDGenerator()
	a, err := gen.NewTokenID()
	if err != nil {
		t.Fatalf("NewTokenID: %v", err)
	}
	b, err := gen.NewTokenID()
	if err != nil {
		t.Fatalf("NewTokenID: %v", err)
	}
	if a == b {
		t.Errorf("two consecutive ids must differ")
	}
	if !strings.HasPrefix(a, "st_") {
		t.Errorf("id missing prefix: %q", a)
	}
	if a != strings.ToLower(a) {
		t.Errorf("id must be lower-case: %q", a)
	}
}

func TestInMemoryIssuanceRateLimiter_DisabledBucketsAlwaysAllow(t *testing.T) {
	t.Parallel()
	l := auth.NewInMemoryIssuanceRateLimiter(0, 0, 0, 0)
	for i := 0; i < 5; i++ {
		ok, err := l.Allow(context.Background(), "demo", domain.RateLimitBucket{})
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if !ok {
			t.Errorf("disabled bucket must always allow (iter %d)", i)
		}
	}
}

func TestInMemoryIssuanceRateLimiter_NilReceiver(t *testing.T) {
	t.Parallel()
	var l *auth.InMemoryIssuanceRateLimiter
	ok, err := l.Allow(context.Background(), "demo", domain.RateLimitBucket{})
	if err != nil {
		t.Errorf("nil receiver must not error: %v", err)
	}
	if !ok {
		t.Errorf("nil receiver must allow")
	}
}

func TestInMemoryIssuanceRateLimiter_ContextCancelled(t *testing.T) {
	t.Parallel()
	l := auth.NewInMemoryIssuanceRateLimiter(10, 1, 5, 1)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := l.Allow(ctx, "demo", domain.RateLimitBucket{}); err == nil {
		t.Errorf("cancelled ctx must propagate err")
	}
}

func TestInMemoryIssuanceRateLimiter_ProjectBucketRefundsGlobalOnDeny(t *testing.T) {
	t.Parallel()
	// Global cap=2, project cap=1: nach erstem Allow ist project leer.
	// Zweiter Aufruf scheitert auf project; das globale Bucket muss
	// trotzdem refunded werden, sonst stünde nur noch 1 Token.
	l := auth.NewInMemoryIssuanceRateLimiter(2, 0, 1, 0)
	ok, _ := l.Allow(context.Background(), "demo", domain.RateLimitBucket{})
	if !ok {
		t.Fatalf("first allow")
	}
	ok, _ = l.Allow(context.Background(), "demo", domain.RateLimitBucket{})
	if ok {
		t.Errorf("second allow on demo must be denied (project cap=1)")
	}
	// Anderes Project muss das globale Token noch bekommen — wir
	// pinnen, dass das Refund den globalen Bucket bewahrt hat.
	ok, _ = l.Allow(context.Background(), "other", domain.RateLimitBucket{})
	if !ok {
		t.Errorf("other project must still get a global token after refund")
	}
}

func TestInMemoryIssuanceRateLimiter_NoRefundWhenGlobalDisabled(t *testing.T) {
	t.Parallel()
	// Global deaktiviert (cap=0/refill=0) → consume() ist No-Op und
	// hat keinen Token verbraucht; der Refund-Pfad darf in dem Fall
	// keinen virtuellen Token „erzeugen" (Review-Finding #7).
	l := auth.NewInMemoryIssuanceRateLimiter(0, 0, 1, 0)
	ok, _ := l.Allow(context.Background(), "demo", domain.RateLimitBucket{})
	if !ok {
		t.Fatalf("first allow on disabled global must pass")
	}
	// Project-Bucket erschöpft → zweiter Aufruf wird abgelehnt;
	// globaler Refund-Pfad ist No-Op, kein Token wird „erzeugt".
	ok, _ = l.Allow(context.Background(), "demo", domain.RateLimitBucket{})
	if ok {
		t.Errorf("second allow must be denied")
	}
}

func TestInMemoryIssuanceRateLimiter_ProjectPolicyOverridesDefault(t *testing.T) {
	t.Parallel()
	// Default project cap = 1; Policy-Override schreibt cap = 3 vor.
	// Drei Aufrufe sollen passen, der vierte scheitern.
	l := auth.NewInMemoryIssuanceRateLimiter(100, 0, 1, 0)
	override := domain.RateLimitBucket{Capacity: 3, RefillPerSecond: 0}
	for i := 0; i < 3; i++ {
		ok, _ := l.Allow(context.Background(), "demo", override)
		if !ok {
			t.Fatalf("iter %d: override must allow", i)
		}
	}
	ok, _ := l.Allow(context.Background(), "demo", override)
	if ok {
		t.Errorf("override-cap+1 must be denied")
	}
}

func TestInMemoryProjectPolicyResolver_FallbackToBaseProject(t *testing.T) {
	t.Parallel()
	base := domain.Project{ID: "demo", AllowedOrigins: []string{"http://localhost:5173"}}
	r, err := auth.NewInMemoryProjectPolicyResolver(nil, map[string]domain.Project{"demo": base})
	if err != nil {
		t.Fatalf("ctor: %v", err)
	}
	p, err := r.ResolvePolicy(context.Background(), "demo")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if p.ProjectID != "demo" {
		t.Errorf("ProjectID: want demo, got %q", p.ProjectID)
	}
	if !p.AllowsAudience(domain.SessionTokenAudiencePlaybackEvents) {
		t.Errorf("default policy must allow playback-events")
	}
}

func TestInMemoryProjectPolicyResolver_UnknownProjectDeniesPolicy(t *testing.T) {
	t.Parallel()
	r, err := auth.NewInMemoryProjectPolicyResolver(nil, nil)
	if err != nil {
		t.Fatalf("ctor: %v", err)
	}
	if _, err := r.ResolvePolicy(context.Background(), "ghost"); !errors.Is(err, domain.ErrAuthPolicyDenied) {
		t.Errorf("want ErrAuthPolicyDenied, got %v", err)
	}
	var nilResolver *auth.InMemoryProjectPolicyResolver
	if _, err := nilResolver.ResolvePolicy(context.Background(), "demo"); !errors.Is(err, domain.ErrAuthPolicyDenied) {
		t.Errorf("nil receiver must return policy denied: %v", err)
	}
}

func TestInMemoryProjectPolicyResolver_RejectsTTLExceedingGlobal(t *testing.T) {
	t.Parallel()
	bad := domain.ProjectPolicy{ProjectID: "demo", ProjectMaxTTLSeconds: 1800}
	if _, err := auth.NewInMemoryProjectPolicyResolver(map[string]domain.ProjectPolicy{"demo": bad}, nil); err == nil {
		t.Errorf("ttl above global ceiling must error at ctor")
	}
}

func TestInMemoryProjectPolicyResolver_ExplicitPolicyWins(t *testing.T) {
	t.Parallel()
	explicit := domain.ProjectPolicy{ProjectID: "demo", ProjectMaxTTLSeconds: 120}
	base := domain.Project{ID: "demo", AllowedOrigins: []string{"http://x"}}
	r, err := auth.NewInMemoryProjectPolicyResolver(map[string]domain.ProjectPolicy{"demo": explicit}, map[string]domain.Project{"demo": base})
	if err != nil {
		t.Fatalf("ctor: %v", err)
	}
	p, err := r.ResolvePolicy(context.Background(), "demo")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if p.ProjectMaxTTLSeconds != 120 {
		t.Errorf("explicit policy must win, got TTL %d", p.ProjectMaxTTLSeconds)
	}
}

func TestMultiKeySigningResolver_NilReceiverAndErrors(t *testing.T) {
	t.Parallel()
	var r *auth.MultiKeySigningResolver
	if _, err := r.ActiveSigningKey(); !errors.Is(err, domain.ErrAuthTokenInvalid) {
		t.Errorf("nil receiver ActiveSigningKey: want ErrAuthTokenInvalid, got %v", err)
	}
	if _, err := r.AllVerifyKeys(); !errors.Is(err, domain.ErrAuthTokenInvalid) {
		t.Errorf("nil receiver AllVerifyKeys: want ErrAuthTokenInvalid, got %v", err)
	}
	// Empty kid in key registration.
	if _, err := auth.NewMultiKeySigningResolver("kid", domain.SessionSigningKey{KID: "", Algorithm: domain.SigningKeyAlgorithmHS256, Secret: []byte("s")}); err == nil {
		t.Errorf("empty kid registration must error")
	}
}

func TestHMACSigner_NilReceiverPaths(t *testing.T) {
	t.Parallel()
	var s *auth.HMACSessionTokenSigner
	if _, err := s.Sign(domain.SessionTokenClaims{}); !errors.Is(err, domain.ErrAuthTokenInvalid) {
		t.Errorf("nil signer Sign: want ErrAuthTokenInvalid, got %v", err)
	}
	if _, err := s.Verify("mtr_st_x.y.z"); !errors.Is(err, domain.ErrAuthTokenInvalid) {
		t.Errorf("nil signer Verify: want ErrAuthTokenInvalid, got %v", err)
	}
	// Signer with resolver that has no active key (we go through the
	// constructor with a placeholder, but ActiveSigningKey path is
	// already covered by RoundTrip; here we just ensure Verify against
	// a mismatched kid path stays correct.
	r, _ := auth.NewMultiKeySigningResolver("kid_a", domain.SessionSigningKey{KID: "kid_a", Algorithm: domain.SigningKeyAlgorithmHS256, Secret: []byte("s")})
	signer := auth.NewHMACSessionTokenSigner(r)
	// Tokens with empty kid header → invalid.
	bogus := "mtr_st_eyJhbGciOiJIUzI1NiIsImtpZCI6IiJ9.eyJzdWIiOiJkZW1vIn0.AA"
	if _, err := signer.Verify(bogus); !errors.Is(err, domain.ErrAuthTokenInvalid) {
		t.Errorf("empty kid header: want ErrAuthTokenInvalid, got %v", err)
	}
	// Token with malformed base64 header.
	if _, err := signer.Verify("mtr_st_!!!.payload.sig"); !errors.Is(err, domain.ErrAuthTokenInvalid) {
		t.Errorf("malformed base64 header: want ErrAuthTokenInvalid, got %v", err)
	}
	// Token with malformed sig segment.
	now := time.Now().UTC().Truncate(time.Second)
	tok, err := signer.Sign(domain.SessionTokenClaims{
		Iss: "m-trace", Sub: "demo", Aud: domain.SessionTokenAudiencePlaybackEvents,
		Iat: now, Nbf: now, Exp: now.Add(time.Minute), JTI: "st_x",
	})
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	parts := strings.Split(strings.TrimPrefix(tok, "mtr_st_"), ".")
	corrupted := "mtr_st_" + parts[0] + "." + parts[1] + ".!!!"
	if _, err := signer.Verify(corrupted); !errors.Is(err, domain.ErrAuthTokenInvalid) {
		t.Errorf("corrupted sig: want ErrAuthTokenInvalid, got %v", err)
	}
}

func TestHMACSigner_VerifyRejectsAlgKidMismatch(t *testing.T) {
	t.Parallel()
	// Custom resolver: active kid_a (HS256) plus retired kid_b registered
	// with a *different* algorithm marker so the alg/kid pairing
	// validation kicks in. We only have HS256 so we have to mutate the
	// header manually instead.
	r, _ := auth.NewMultiKeySigningResolver("kid_a", domain.SessionSigningKey{KID: "kid_a", Algorithm: domain.SigningKeyAlgorithmHS256, Secret: []byte("a")})
	signer := auth.NewHMACSessionTokenSigner(r)
	now := time.Now().UTC()
	tok, err := signer.Sign(domain.SessionTokenClaims{
		Iss: "m", Sub: "demo", Aud: domain.SessionTokenAudiencePlaybackEvents,
		Iat: now, Nbf: now, Exp: now.Add(time.Minute), JTI: "st_x",
	})
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	// Header with unknown alg.
	bogus := "mtr_st_eyJhbGciOiJSUzI1NiIsImtpZCI6ImtpZF9hIn0." + strings.Split(strings.TrimPrefix(tok, "mtr_st_"), ".")[1] + ".sig"
	if _, err := signer.Verify(bogus); !errors.Is(err, domain.ErrAuthTokenInvalid) {
		t.Errorf("unknown alg: want ErrAuthTokenInvalid, got %v", err)
	}
}
