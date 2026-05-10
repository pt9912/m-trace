package domain_test

import (
	"errors"
	"testing"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// `0.12.0` Tranche 1 / RAK-74: Project-Policy-Validierung (Audience,
// Origin, Methode, Header, TTL-Auflösung, Issuance-Quote). Tests
// laufen ohne HTTP, Storage oder Rate-Limit-Adapter.

func TestHTTPMethod_IsKnown(t *testing.T) {
	t.Parallel()
	if !domain.HTTPMethodPOST.IsKnown() {
		t.Errorf("POST must be known")
	}
	if !domain.HTTPMethodOPTIONS.IsKnown() {
		t.Errorf("OPTIONS must be known")
	}
	if domain.HTTPMethod("GET").IsKnown() {
		t.Errorf("GET must not be in 0.12.0 allowlist")
	}
	if domain.HTTPMethod("post").IsKnown() {
		t.Errorf("lowercase post must not be known: case-sensitive wire contract")
	}
}

func TestGlobalPreflightAllowlistIsStable(t *testing.T) {
	t.Parallel()
	headers := domain.GlobalPreflightHeaderAllowlist()
	want := []domain.AllowedRequestHeader{
		domain.HeaderContentType,
		domain.HeaderAuthorization,
		domain.HeaderXMTraceToken,
		domain.HeaderXMTraceSessionToken,
		domain.HeaderTraceparent,
	}
	if len(headers) != len(want) {
		t.Fatalf("len: want %d, got %d", len(want), len(headers))
	}
	for i, h := range headers {
		if h != want[i] {
			t.Errorf("header[%d]: want %q, got %q", i, want[i], h)
		}
	}
	methods := domain.GlobalPreflightMethodAllowlist()
	if len(methods) != 2 || methods[0] != domain.HTTPMethodPOST || methods[1] != domain.HTTPMethodOPTIONS {
		t.Errorf("methods: want [POST OPTIONS], got %v", methods)
	}
}

func TestRateLimitBucket_IsZero(t *testing.T) {
	t.Parallel()
	if !(domain.RateLimitBucket{}).IsZero() {
		t.Errorf("default must be zero")
	}
	if (domain.RateLimitBucket{Capacity: 5}).IsZero() {
		t.Errorf("non-zero capacity must not be zero")
	}
	if (domain.RateLimitBucket{RefillPerSecond: 1.5}).IsZero() {
		t.Errorf("non-zero refill must not be zero")
	}
}

func TestProjectPolicy_EffectiveMaxTTLSeconds(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		set  int
		want int
	}{
		{"unset → default 900", 0, 900},
		{"negative → default 900", -10, 900},
		{"valid lower bound", 1, 1},
		{"valid mid", 300, 300},
		{"exact default", 900, 900},
		{"above default → clamp 900", 9000, 900},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := domain.ProjectPolicy{ProjectMaxTTLSeconds: tc.set}
			if got := p.EffectiveMaxTTLSeconds(); got != tc.want {
				t.Errorf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestProjectPolicy_AllowsOrigin(t *testing.T) {
	t.Parallel()
	p := domain.ProjectPolicy{
		AllowedOrigins: []string{"http://localhost:5173", "https://demo.example.com"},
	}
	if !p.AllowsOrigin("http://localhost:5173") {
		t.Errorf("listed origin must be allowed")
	}
	if !p.AllowsOrigin("") {
		t.Errorf("empty origin must be allowed (CLI/curl pass)")
	}
	if p.AllowsOrigin("http://other") {
		t.Errorf("unlisted origin must not be allowed")
	}
}

func TestProjectPolicy_AllowsMethod(t *testing.T) {
	t.Parallel()
	// Empty list → globaler Default (POST/OPTIONS).
	pEmpty := domain.ProjectPolicy{}
	if !pEmpty.AllowsMethod(domain.HTTPMethodPOST) {
		t.Errorf("empty list: POST must be allowed via global default")
	}
	if pEmpty.AllowsMethod(domain.HTTPMethod("GET")) {
		t.Errorf("empty list: GET must not be allowed (not in global default)")
	}
	// Restrictive list (POST only).
	pStrict := domain.ProjectPolicy{AllowedMethods: []domain.HTTPMethod{domain.HTTPMethodPOST}}
	if !pStrict.AllowsMethod(domain.HTTPMethodPOST) {
		t.Errorf("restricted list: POST must be allowed")
	}
	if pStrict.AllowsMethod(domain.HTTPMethodOPTIONS) {
		t.Errorf("restricted list: OPTIONS must not be allowed")
	}
}

func TestProjectPolicy_AllowsHeader(t *testing.T) {
	t.Parallel()
	pEmpty := domain.ProjectPolicy{}
	for _, h := range []string{"Content-Type", "content-type", "Authorization", "X-MTrace-Token", "X-MTrace-Session-Token", "traceparent"} {
		if !pEmpty.AllowsHeader(h) {
			t.Errorf("empty list: %q must be allowed via global default", h)
		}
	}
	if pEmpty.AllowsHeader("X-Custom") {
		t.Errorf("empty list: X-Custom must not be allowed")
	}
	if pEmpty.AllowsHeader("") {
		t.Errorf("empty header must never be allowed")
	}
	pStrict := domain.ProjectPolicy{
		AllowedRequestHeaders: []domain.AllowedRequestHeader{domain.HeaderContentType, domain.HeaderXMTraceToken},
	}
	if !pStrict.AllowsHeader("content-type") {
		t.Errorf("strict list: case-insensitive content-type must be allowed")
	}
	if pStrict.AllowsHeader("Authorization") {
		t.Errorf("strict list: Authorization must not be allowed when not listed")
	}
}

func TestProjectPolicy_AllowsAudience(t *testing.T) {
	t.Parallel()
	pEmpty := domain.ProjectPolicy{}
	if !pEmpty.AllowsAudience(domain.SessionTokenAudiencePlaybackEvents) {
		t.Errorf("empty list defaults to playback-events")
	}
	if pEmpty.AllowsAudience(domain.SessionTokenAudience("admin")) {
		t.Errorf("empty list must not allow unknown audiences")
	}
	pSet := domain.ProjectPolicy{
		AllowedAudiences: []domain.SessionTokenAudience{domain.SessionTokenAudiencePlaybackEvents},
	}
	if !pSet.AllowsAudience(domain.SessionTokenAudiencePlaybackEvents) {
		t.Errorf("explicit list must allow listed audience")
	}
	if pSet.AllowsAudience(domain.SessionTokenAudience("admin")) {
		t.Errorf("explicit list must not allow unlisted audience")
	}
	// Globally unknown audience must be denied even if Project lists it.
	pGarbage := domain.ProjectPolicy{
		AllowedAudiences: []domain.SessionTokenAudience{"admin"},
	}
	if pGarbage.AllowsAudience(domain.SessionTokenAudience("admin")) {
		t.Errorf("globally unknown audience must be denied even if Project lists it")
	}
}

func TestProjectPolicy_AllowsAudienceForOrigin_RestrictiveOverride(t *testing.T) {
	t.Parallel()
	p := domain.ProjectPolicy{
		AllowedAudiences: []domain.SessionTokenAudience{domain.SessionTokenAudiencePlaybackEvents},
		OriginOverrides: []domain.OriginPolicy{
			{Origin: "http://restricted.example.com", RestrictedAudiences: []domain.SessionTokenAudience{}},
			{Origin: "http://allow.example.com", RestrictedAudiences: []domain.SessionTokenAudience{domain.SessionTokenAudiencePlaybackEvents}},
		},
	}
	// Project allows playback-events for any origin without override.
	if !p.AllowsAudienceForOrigin(domain.SessionTokenAudiencePlaybackEvents, "http://other.example.com") {
		t.Errorf("origin without override must inherit project allowlist")
	}
	// Override with empty restricted list is permissive.
	if !p.AllowsAudienceForOrigin(domain.SessionTokenAudiencePlaybackEvents, "http://restricted.example.com") {
		t.Errorf("empty restriction means inherit project allowlist")
	}
	// Override with explicit listing.
	if !p.AllowsAudienceForOrigin(domain.SessionTokenAudiencePlaybackEvents, "http://allow.example.com") {
		t.Errorf("explicit override must allow listed audience")
	}
}

func TestValidateAudience(t *testing.T) {
	t.Parallel()
	p := domain.ProjectPolicy{
		AllowedAudiences: []domain.SessionTokenAudience{domain.SessionTokenAudiencePlaybackEvents},
	}
	if err := domain.ValidateAudience(p, domain.SessionTokenAudiencePlaybackEvents, ""); err != nil {
		t.Errorf("must succeed: %v", err)
	}
	if err := domain.ValidateAudience(p, domain.SessionTokenAudience("admin"), ""); !errors.Is(err, domain.ErrAuthSessionScopeDenied) {
		t.Errorf("unknown aud: want ErrAuthSessionScopeDenied, got %v", err)
	}
}

func TestValidateOriginAgainstPolicy(t *testing.T) {
	t.Parallel()
	p := domain.ProjectPolicy{
		AllowedOrigins: []string{"http://localhost:5173"},
	}
	if err := domain.ValidateOriginAgainstPolicy(p, "http://localhost:5173"); err != nil {
		t.Errorf("listed origin: %v", err)
	}
	if err := domain.ValidateOriginAgainstPolicy(p, ""); err != nil {
		t.Errorf("empty origin (CLI): %v", err)
	}
	if err := domain.ValidateOriginAgainstPolicy(p, "http://other"); !errors.Is(err, domain.ErrAuthPolicyDenied) {
		t.Errorf("unlisted origin: want ErrAuthPolicyDenied, got %v", err)
	}
}

func TestValidateMethodAgainstPolicy(t *testing.T) {
	t.Parallel()
	pStrict := domain.ProjectPolicy{
		AllowedMethods: []domain.HTTPMethod{domain.HTTPMethodPOST},
	}
	if err := domain.ValidateMethodAgainstPolicy(pStrict, domain.HTTPMethodPOST); err != nil {
		t.Errorf("POST: %v", err)
	}
	if err := domain.ValidateMethodAgainstPolicy(pStrict, domain.HTTPMethodOPTIONS); !errors.Is(err, domain.ErrAuthPolicyDenied) {
		t.Errorf("OPTIONS denied: want ErrAuthPolicyDenied, got %v", err)
	}
}

func TestValidateHeaderAgainstPolicy(t *testing.T) {
	t.Parallel()
	p := domain.ProjectPolicy{
		AllowedRequestHeaders: []domain.AllowedRequestHeader{domain.HeaderXMTraceToken},
	}
	if err := domain.ValidateHeaderAgainstPolicy(p, "X-MTrace-Token"); err != nil {
		t.Errorf("listed header: %v", err)
	}
	if err := domain.ValidateHeaderAgainstPolicy(p, "X-Custom"); !errors.Is(err, domain.ErrAuthPolicyDenied) {
		t.Errorf("unlisted header: want ErrAuthPolicyDenied, got %v", err)
	}
}

func TestResolveTTLSeconds(t *testing.T) {
	t.Parallel()
	p := domain.ProjectPolicy{ProjectMaxTTLSeconds: 600}
	cases := []struct {
		name      string
		requested int
		wantTTL   int
		wantErr   error
	}{
		{"unset → effective max", 0, 600, nil},
		{"valid lower", 60, 60, nil},
		{"exact effective max", 600, 600, nil},
		{"above effective max → 422", 700, 0, domain.ErrAuthTokenTTLTooLarge},
		{"above global max → 422", 1000, 0, domain.ErrAuthTokenTTLTooLarge},
		{"negative → 422", -1, 0, domain.ErrAuthTokenTTLTooLarge},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ttl, err := domain.ResolveTTLSeconds(p, tc.requested)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("err: want %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if ttl != tc.wantTTL {
				t.Errorf("ttl: want %d, got %d", tc.wantTTL, ttl)
			}
		})
	}
}

func TestResolveTTLSeconds_DefaultsTo900WhenUnset(t *testing.T) {
	t.Parallel()
	p := domain.ProjectPolicy{} // ProjectMaxTTLSeconds unset.
	ttl, err := domain.ResolveTTLSeconds(p, 0)
	if err != nil {
		t.Fatalf("unset request: %v", err)
	}
	if ttl != 900 {
		t.Errorf("ttl: want 900, got %d", ttl)
	}
	if _, err := domain.ResolveTTLSeconds(p, 901); !errors.Is(err, domain.ErrAuthTokenTTLTooLarge) {
		t.Errorf("ttl 901 with default project max: want ErrAuthTokenTTLTooLarge, got %v", err)
	}
}

func TestProjectPolicyFromBaseProject(t *testing.T) {
	t.Parallel()
	base := domain.Project{
		ID:             "demo",
		AllowedOrigins: []string{"http://localhost:5173"},
	}
	p := domain.ProjectPolicyFromBaseProject(base)
	if p.ProjectID != "demo" {
		t.Errorf("ProjectID: want demo, got %q", p.ProjectID)
	}
	if len(p.AllowedOrigins) != 1 || p.AllowedOrigins[0] != "http://localhost:5173" {
		t.Errorf("AllowedOrigins: want [localhost:5173], got %v", p.AllowedOrigins)
	}
	// Defaults should match global allowlists.
	if len(p.AllowedMethods) != 2 {
		t.Errorf("default methods: want 2, got %d", len(p.AllowedMethods))
	}
	if len(p.AllowedRequestHeaders) != 5 {
		t.Errorf("default headers: want 5, got %d", len(p.AllowedRequestHeaders))
	}
	if len(p.AllowedAudiences) != 1 || p.AllowedAudiences[0] != domain.SessionTokenAudiencePlaybackEvents {
		t.Errorf("default audiences: want [playback-events], got %v", p.AllowedAudiences)
	}
	// Mutate input — defensive copy must hold.
	base.AllowedOrigins[0] = "http://mutated"
	if p.AllowedOrigins[0] != "http://localhost:5173" {
		t.Errorf("AllowedOrigins must be defensively copied")
	}
}

func TestEffectiveIssuanceQuota(t *testing.T) {
	t.Parallel()
	p := domain.ProjectPolicy{
		RateLimit: domain.RateLimitPolicy{
			IssuanceBucket: domain.RateLimitBucket{Capacity: 10, RefillPerSecond: 1.0},
		},
	}
	q := p.EffectiveIssuanceQuota()
	if q.Bucket.Capacity != 10 {
		t.Errorf("capacity: want 10, got %d", q.Bucket.Capacity)
	}
	if q.Bucket.RefillPerSecond != 1.0 {
		t.Errorf("refill: want 1.0, got %v", q.Bucket.RefillPerSecond)
	}
}
