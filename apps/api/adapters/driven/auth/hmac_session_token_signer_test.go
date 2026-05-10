package auth_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/auth"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// `0.12.0` Tranche 2 / RAK-72: Sign-/Verify-Round-Trip,
// `kid`-Lookup, restart-stabile Verifikation und Tampering-
// Detection.

func newSigner(t *testing.T, activeKID domain.SigningKeyID, keys ...domain.SessionSigningKey) *auth.HMACSessionTokenSigner {
	t.Helper()
	r, err := auth.NewStaticSigningKeyResolver(activeKID, keys...)
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	return auth.NewHMACSessionTokenSigner(r)
}

func sampleClaims(now time.Time) domain.SessionTokenClaims {
	sess := "sess_1"
	origin := "http://localhost:5173"
	return domain.SessionTokenClaims{
		Iss:       "m-trace",
		Sub:       "demo",
		Aud:       domain.SessionTokenAudiencePlaybackEvents,
		Iat:       now,
		Nbf:       now,
		Exp:       now.Add(time.Minute),
		JTI:       "st_test",
		SessionID: &sess,
		Origin:    &origin,
	}
}

func TestHMACSigner_RoundTrip(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC().Truncate(time.Second)
	key := domain.SessionSigningKey{
		KID: "kid_1", Algorithm: domain.SigningKeyAlgorithmHS256,
		Secret: []byte("super-secret"),
	}
	s := newSigner(t, "kid_1", key)
	tok, err := s.Sign(sampleClaims(now))
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	if !strings.HasPrefix(tok, "mtr_st_") {
		t.Errorf("missing prefix: %q", tok)
	}
	parts := strings.Split(strings.TrimPrefix(tok, "mtr_st_"), ".")
	if len(parts) != 3 {
		t.Fatalf("want 3 segments, got %d", len(parts))
	}
	out, err := s.Verify(tok)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if out.JTI != "st_test" || out.Sub != "demo" {
		t.Errorf("claims mismatch: %+v", out)
	}
	if out.SessionID == nil || *out.SessionID != "sess_1" {
		t.Errorf("session id: %v", out.SessionID)
	}
	if out.Origin == nil || *out.Origin != "http://localhost:5173" {
		t.Errorf("origin: %v", out.Origin)
	}
}

func TestHMACSigner_VerifyRejectsTamperedSignature(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC().Truncate(time.Second)
	s := newSigner(t, "kid_1", domain.SessionSigningKey{
		KID: "kid_1", Algorithm: domain.SigningKeyAlgorithmHS256,
		Secret: []byte("secret"),
	})
	tok, err := s.Sign(sampleClaims(now))
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	tampered := tok[:len(tok)-2] + "AA"
	if _, err := s.Verify(tampered); !errors.Is(err, domain.ErrAuthTokenInvalid) {
		t.Errorf("want ErrAuthTokenInvalid, got %v", err)
	}
}

func TestHMACSigner_VerifyRejectsUnknownKID(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC().Truncate(time.Second)
	signer := newSigner(t, "kid_a", domain.SessionSigningKey{
		KID: "kid_a", Algorithm: domain.SigningKeyAlgorithmHS256,
		Secret: []byte("secret-a"),
	})
	tok, _ := signer.Sign(sampleClaims(now))
	// Verifier without that KID.
	verifier := newSigner(t, "kid_b", domain.SessionSigningKey{
		KID: "kid_b", Algorithm: domain.SigningKeyAlgorithmHS256,
		Secret: []byte("secret-b"),
	})
	if _, err := verifier.Verify(tok); !errors.Is(err, domain.ErrAuthTokenInvalid) {
		t.Errorf("want ErrAuthTokenInvalid, got %v", err)
	}
}

func TestHMACSigner_RestartStableAcrossKeyResolverReinitialization(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC().Truncate(time.Second)
	secret := []byte("stable-secret")
	signer1 := newSigner(t, "kid_a", domain.SessionSigningKey{
		KID: "kid_a", Algorithm: domain.SigningKeyAlgorithmHS256, Secret: secret,
	})
	tok, err := signer1.Sign(sampleClaims(now))
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	// Operator switches active key but keeps old verify key.
	signer2 := newSigner(t, "kid_b",
		domain.SessionSigningKey{KID: "kid_b", Algorithm: domain.SigningKeyAlgorithmHS256, Secret: []byte("new-secret")},
		domain.SessionSigningKey{KID: "kid_a", Algorithm: domain.SigningKeyAlgorithmHS256, Secret: secret},
	)
	if _, err := signer2.Verify(tok); err != nil {
		t.Fatalf("verify after rotation must succeed: %v", err)
	}
}

func TestHMACSigner_VerifyRejectsMissingPrefix(t *testing.T) {
	t.Parallel()
	s := newSigner(t, "kid_1", domain.SessionSigningKey{
		KID: "kid_1", Algorithm: domain.SigningKeyAlgorithmHS256, Secret: []byte("s"),
	})
	if _, err := s.Verify("eyJhbGciOiJIUzI1NiJ9.payload.sig"); !errors.Is(err, domain.ErrAuthTokenInvalid) {
		t.Errorf("want ErrAuthTokenInvalid, got %v", err)
	}
}

func TestHMACSigner_VerifyRejectsMalformedStructure(t *testing.T) {
	t.Parallel()
	s := newSigner(t, "kid_1", domain.SessionSigningKey{
		KID: "kid_1", Algorithm: domain.SigningKeyAlgorithmHS256, Secret: []byte("s"),
	})
	if _, err := s.Verify("mtr_st_only.two"); !errors.Is(err, domain.ErrAuthTokenInvalid) {
		t.Errorf("want ErrAuthTokenInvalid, got %v", err)
	}
	if _, err := s.Verify("mtr_st_"); !errors.Is(err, domain.ErrAuthTokenInvalid) {
		t.Errorf("want ErrAuthTokenInvalid, got %v", err)
	}
}

// erroringKeyResolver liefert ActiveSigningKey- bzw.
// AllVerifyKeys-Fehler — verifiziert die Error-Propagation in den
// HMACSessionTokenSigner-Pfaden Sign/Verify.
type erroringKeyResolver struct {
	signErr   error
	verifyErr error
}

func (e erroringKeyResolver) ActiveSigningKey() (domain.SessionSigningKey, error) {
	return domain.SessionSigningKey{}, e.signErr
}

func (e erroringKeyResolver) AllVerifyKeys() ([]domain.SessionSigningKey, error) {
	return nil, e.verifyErr
}

func TestHMACSigner_SignReturnsActiveKeyError(t *testing.T) {
	t.Parallel()
	want := errors.New("key store down")
	s := auth.NewHMACSessionTokenSigner(erroringKeyResolver{signErr: want})
	if _, err := s.Sign(domain.SessionTokenClaims{}); !errors.Is(err, want) {
		t.Errorf("ActiveSigningKey err must propagate: got %v", err)
	}
}

func TestHMACSigner_SignRejectsUnknownAlgorithm(t *testing.T) {
	t.Parallel()
	r, err := auth.NewStaticSigningKeyResolver("kid_a", domain.SessionSigningKey{
		KID: "kid_a", Algorithm: domain.SigningKeyAlgorithm("RS256"), Secret: []byte("s"),
	})
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	s := auth.NewHMACSessionTokenSigner(r)
	if _, err := s.Sign(domain.SessionTokenClaims{}); !errors.Is(err, domain.ErrAuthTokenInvalid) {
		t.Errorf("unknown alg on Sign: want ErrAuthTokenInvalid, got %v", err)
	}
}

func TestHMACSigner_VerifyPropagatesAllVerifyKeysError(t *testing.T) {
	t.Parallel()
	want := errors.New("verify keys unavailable")
	s := auth.NewHMACSessionTokenSigner(erroringKeyResolver{verifyErr: want})
	// Token-String muss Header-Form überstehen, damit Verify zum
	// Key-Ring-Lookup kommt.
	tok := "mtr_st_eyJhbGciOiJIUzI1NiIsImtpZCI6ImtpZF9hIn0.eyJzdWIiOiJkZW1vIn0.AAAA"
	if _, err := s.Verify(tok); !errors.Is(err, want) {
		t.Errorf("AllVerifyKeys err must propagate: got %v", err)
	}
}

func TestStaticSigningKeyResolver_EnforcesActiveKey(t *testing.T) {
	t.Parallel()
	if _, err := auth.NewStaticSigningKeyResolver("absent",
		domain.SessionSigningKey{KID: "kid_a", Algorithm: domain.SigningKeyAlgorithmHS256, Secret: []byte("s")},
	); err == nil {
		t.Errorf("missing active kid must error")
	}
	if _, err := auth.NewStaticSigningKeyResolver(""); err == nil {
		t.Errorf("empty active kid must error")
	}
	if _, err := auth.NewStaticSigningKeyResolver("kid_a",
		domain.SessionSigningKey{KID: "kid_a", Algorithm: domain.SigningKeyAlgorithmHS256, Secret: []byte("a")},
		domain.SessionSigningKey{KID: "kid_a", Algorithm: domain.SigningKeyAlgorithmHS256, Secret: []byte("b")},
	); err == nil {
		t.Errorf("duplicate kid must error")
	}
}
