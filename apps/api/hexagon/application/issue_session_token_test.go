package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/application"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driving"
)

// `0.12.0` Tranche 2: Issuance-Service ohne HTTP/Storage/Crypto-
// Library. Stubs implementieren die Driven-Ports.

type stubPolicies struct {
	policy domain.ProjectPolicy
	err    error
}

func (s stubPolicies) ResolvePolicy(_ context.Context, _ string) (domain.ProjectPolicy, error) {
	return s.policy, s.err
}

type issuanceStubLimiter struct {
	allow bool
	err   error
	calls int
}

func (s *issuanceStubLimiter) Allow(_ context.Context, _ string) (bool, error) {
	s.calls++
	return s.allow, s.err
}

type stubSigner struct {
	out  string
	err  error
	last domain.SessionTokenClaims
}

func (s *stubSigner) Sign(claims domain.SessionTokenClaims) (string, error) {
	s.last = claims
	return s.out, s.err
}

func (s *stubSigner) Verify(_ string) (domain.SessionTokenClaims, error) {
	return domain.SessionTokenClaims{}, errors.New("not used")
}

type stubIDs struct {
	id  string
	err error
}

func (s stubIDs) NewTokenID() (string, error) { return s.id, s.err }

func defaultPolicy() domain.ProjectPolicy {
	return domain.ProjectPolicy{
		ProjectID:        "demo",
		AllowedAudiences: []domain.SessionTokenAudience{domain.SessionTokenAudiencePlaybackEvents},
	}
}

func newService(policies stubPolicies, limiter *issuanceStubLimiter, signer *stubSigner, ids stubIDs, now time.Time) *application.IssueSessionTokenService {
	svc := application.NewIssueSessionTokenService(policies, limiter, signer, ids)
	svc.Now = func() time.Time { return now }
	return svc
}

func TestIssueSessionToken_HappyPath(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	signer := &stubSigner{out: "mtr_st_signed"}
	limiter := &issuanceStubLimiter{allow: true}
	svc := newService(stubPolicies{policy: defaultPolicy()}, limiter, signer, stubIDs{id: "st_001"}, now)

	out, err := svc.IssueSessionToken(context.Background(), driving.IssueSessionTokenRequest{
		ResolvedProjectID:   "demo",
		Audience:            "playback-events",
		RequestedTTLSeconds: 60,
		SessionID:           "sess_a",
		Origin:              "http://localhost:5173",
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if out.Value != "mtr_st_signed" {
		t.Errorf("Value: want signed token, got %q", out.Value)
	}
	if out.TokenID != "st_001" {
		t.Errorf("TokenID: want st_001, got %q", out.TokenID)
	}
	if out.ProjectID != "demo" {
		t.Errorf("ProjectID: want demo, got %q", out.ProjectID)
	}
	if out.Audience != domain.SessionTokenAudiencePlaybackEvents {
		t.Errorf("Audience: want playback-events, got %q", out.Audience)
	}
	if !out.ExpiresAt.Equal(now.Add(60 * time.Second)) {
		t.Errorf("ExpiresAt: want now+60s, got %v", out.ExpiresAt)
	}
	if out.SessionID != "sess_a" {
		t.Errorf("SessionID: want sess_a, got %q", out.SessionID)
	}
	if signer.last.JTI != "st_001" {
		t.Errorf("Signer claims jti: want st_001, got %q", signer.last.JTI)
	}
	if signer.last.Sub != "demo" {
		t.Errorf("Signer claims sub: want demo, got %q", signer.last.Sub)
	}
	if limiter.calls != 1 {
		t.Errorf("limiter must be called exactly once, got %d", limiter.calls)
	}
}

func TestIssueSessionToken_ProjectMismatchFromRequestBody(t *testing.T) {
	t.Parallel()
	svc := newService(stubPolicies{policy: defaultPolicy()}, &issuanceStubLimiter{allow: true}, &stubSigner{out: "x"}, stubIDs{id: "st"}, time.Now())
	_, err := svc.IssueSessionToken(context.Background(), driving.IssueSessionTokenRequest{
		ResolvedProjectID:   "demo",
		RequestProjectID:    "other",
		Audience:            "playback-events",
		RequestedTTLSeconds: 60,
	})
	if !errors.Is(err, domain.ErrAuthProjectMismatch) {
		t.Errorf("want ErrAuthProjectMismatch, got %v", err)
	}
}

func TestIssueSessionToken_MissingResolvedProjectID(t *testing.T) {
	t.Parallel()
	svc := newService(stubPolicies{policy: defaultPolicy()}, &issuanceStubLimiter{allow: true}, &stubSigner{out: "x"}, stubIDs{id: "st"}, time.Now())
	_, err := svc.IssueSessionToken(context.Background(), driving.IssueSessionTokenRequest{Audience: "playback-events", RequestedTTLSeconds: 60})
	if !errors.Is(err, domain.ErrAuthTokenMissing) {
		t.Errorf("want ErrAuthTokenMissing, got %v", err)
	}
}

func TestIssueSessionToken_AudienceDeniedByPolicy(t *testing.T) {
	t.Parallel()
	svc := newService(stubPolicies{policy: defaultPolicy()}, &issuanceStubLimiter{allow: true}, &stubSigner{out: "x"}, stubIDs{id: "st"}, time.Now())
	_, err := svc.IssueSessionToken(context.Background(), driving.IssueSessionTokenRequest{
		ResolvedProjectID:   "demo",
		Audience:            "admin",
		RequestedTTLSeconds: 60,
	})
	if !errors.Is(err, domain.ErrAuthSessionScopeDenied) {
		t.Errorf("want ErrAuthSessionScopeDenied, got %v", err)
	}
}

func TestIssueSessionToken_AudienceMissing(t *testing.T) {
	t.Parallel()
	svc := newService(stubPolicies{policy: defaultPolicy()}, &issuanceStubLimiter{allow: true}, &stubSigner{out: "x"}, stubIDs{id: "st"}, time.Now())
	_, err := svc.IssueSessionToken(context.Background(), driving.IssueSessionTokenRequest{
		ResolvedProjectID:   "demo",
		Audience:            "   ",
		RequestedTTLSeconds: 60,
	})
	if !errors.Is(err, domain.ErrAuthSessionScopeDenied) {
		t.Errorf("want ErrAuthSessionScopeDenied, got %v", err)
	}
}

func TestIssueSessionToken_TTLTooLargeNoSilentClamp(t *testing.T) {
	t.Parallel()
	svc := newService(stubPolicies{policy: defaultPolicy()}, &issuanceStubLimiter{allow: true}, &stubSigner{out: "x"}, stubIDs{id: "st"}, time.Now())
	_, err := svc.IssueSessionToken(context.Background(), driving.IssueSessionTokenRequest{
		ResolvedProjectID:   "demo",
		Audience:            "playback-events",
		RequestedTTLSeconds: 1000,
	})
	if !errors.Is(err, domain.ErrAuthTokenTTLTooLarge) {
		t.Errorf("want ErrAuthTokenTTLTooLarge, got %v", err)
	}
}

func TestIssueSessionToken_TTLDefaultsToProjectMax(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	policy := defaultPolicy()
	policy.ProjectMaxTTLSeconds = 600
	signer := &stubSigner{out: "x"}
	svc := newService(stubPolicies{policy: policy}, &issuanceStubLimiter{allow: true}, signer, stubIDs{id: "st"}, now)

	out, err := svc.IssueSessionToken(context.Background(), driving.IssueSessionTokenRequest{
		ResolvedProjectID: "demo",
		Audience:          "playback-events",
		// RequestedTTLSeconds == 0
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !out.ExpiresAt.Equal(now.Add(600 * time.Second)) {
		t.Errorf("ExpiresAt: want now+600s, got %v", out.ExpiresAt)
	}
}

func TestIssueSessionToken_RateLimited(t *testing.T) {
	t.Parallel()
	svc := newService(stubPolicies{policy: defaultPolicy()}, &issuanceStubLimiter{allow: false}, &stubSigner{out: "x"}, stubIDs{id: "st"}, time.Now())
	_, err := svc.IssueSessionToken(context.Background(), driving.IssueSessionTokenRequest{
		ResolvedProjectID:   "demo",
		Audience:            "playback-events",
		RequestedTTLSeconds: 60,
	})
	if !errors.Is(err, domain.ErrAuthIssuanceRateLimited) {
		t.Errorf("want ErrAuthIssuanceRateLimited, got %v", err)
	}
}

func TestIssueSessionToken_LimiterErrorPropagates(t *testing.T) {
	t.Parallel()
	limErr := errors.New("limiter down")
	svc := newService(stubPolicies{policy: defaultPolicy()}, &issuanceStubLimiter{allow: false, err: limErr}, &stubSigner{out: "x"}, stubIDs{id: "st"}, time.Now())
	_, err := svc.IssueSessionToken(context.Background(), driving.IssueSessionTokenRequest{
		ResolvedProjectID:   "demo",
		Audience:            "playback-events",
		RequestedTTLSeconds: 60,
	})
	if !errors.Is(err, limErr) {
		t.Errorf("want limiter error to propagate, got %v", err)
	}
}

func TestIssueSessionToken_PolicyResolveErrorPropagates(t *testing.T) {
	t.Parallel()
	svc := newService(stubPolicies{err: domain.ErrAuthPolicyDenied}, &issuanceStubLimiter{allow: true}, &stubSigner{out: "x"}, stubIDs{id: "st"}, time.Now())
	_, err := svc.IssueSessionToken(context.Background(), driving.IssueSessionTokenRequest{
		ResolvedProjectID:   "demo",
		Audience:            "playback-events",
		RequestedTTLSeconds: 60,
	})
	if !errors.Is(err, domain.ErrAuthPolicyDenied) {
		t.Errorf("want ErrAuthPolicyDenied, got %v", err)
	}
}

func TestIssueSessionToken_NoSessionIDClaimWhenEmpty(t *testing.T) {
	t.Parallel()
	signer := &stubSigner{out: "x"}
	svc := newService(stubPolicies{policy: defaultPolicy()}, &issuanceStubLimiter{allow: true}, signer, stubIDs{id: "st"}, time.Now())
	out, err := svc.IssueSessionToken(context.Background(), driving.IssueSessionTokenRequest{
		ResolvedProjectID:   "demo",
		Audience:            "playback-events",
		RequestedTTLSeconds: 60,
		// SessionID and Origin empty.
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if out.SessionID != "" {
		t.Errorf("SessionID: want empty, got %q", out.SessionID)
	}
	if signer.last.SessionID != nil {
		t.Errorf("claims session_id: want nil, got %v", signer.last.SessionID)
	}
	if signer.last.Origin != nil {
		t.Errorf("claims origin: want nil, got %v", signer.last.Origin)
	}
}
