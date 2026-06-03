package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driving"
)

// IssueSessionTokenService implementiert
// `driving.AuthSessionInbound` (RAK-72). Der Service
// orchestriert Project-Policy-Resolve, Audience-/TTL-Validierung,
// Issuance-Rate-Limit, Token-ID-Generierung, Claim-Build und
// Signatur — ohne selbst HTTP, JSON, Storage oder Crypto-Library zu
// kennen.
//
// Reihenfolge der Schritte (entspricht der Validierungsreihenfolge
// aus `spec/backend-api-contract.md` §3.9):
//
//  1. Project-ID-Konsistenz (Body vs. Token).
//  2. Project-Policy laden.
//  3. Audience normalisieren + gegen Allowlist prüfen.
//  4. TTL gegen wirksame Project-Grenze auflösen (kein stiller
//  Clamp).
//  5. Issuance-Rate-Limit (global + Project) prüfen.
//  6. Token-ID generieren.
//  7. Claims bauen + signieren.
//  8. Result zurückgeben (Klartext-Token genau einmal).
type IssueSessionTokenService struct {
	Policies driven.ProjectPolicyResolver
	Limiter  driven.IssuanceRateLimiter
	Signer   driven.SessionTokenSigner
	IDs      driven.TokenIDGenerator
	Now      func() time.Time
	Issuer   string
}

// NewIssueSessionTokenService konstruiert den Service mit Default-
// Clock (`time.Now.UTC`) und Default-Issuer
// (`domain.DefaultSessionTokenIssuer`). Tests können beide Werte
// überschreiben.
func NewIssueSessionTokenService(
	policies driven.ProjectPolicyResolver,
	limiter driven.IssuanceRateLimiter,
	signer driven.SessionTokenSigner,
	ids driven.TokenIDGenerator,
) *IssueSessionTokenService {
	return &IssueSessionTokenService{
		Policies: policies,
		Limiter:  limiter,
		Signer:   signer,
		IDs:      ids,
		Now:      func() time.Time { return time.Now().UTC() },
		Issuer:   domain.DefaultSessionTokenIssuer,
	}
}

// Compile-time check dass der Service den Driving-Port erfüllt.
var _ driving.AuthSessionInbound = (*IssueSessionTokenService)(nil)

// IssueSessionToken implementiert `driving.AuthSessionInbound`.
func (s *IssueSessionTokenService) IssueSessionToken(ctx context.Context, req driving.IssueSessionTokenRequest) (driving.IssueSessionTokenResult, error) {
	resolved := strings.TrimSpace(req.ResolvedProjectID)
	if resolved == "" {
		return driving.IssueSessionTokenResult{}, domain.ErrAuthTokenMissing
	}
	if err := domain.ValidateProjectIDConsistency(req.RequestProjectID, resolved); err != nil {
		// Der HTTP-Adapter mappt `ErrIngestProjectIDMismatch` heute auf
		// `400 project_id_mismatch`. Im Auth-Pfad wird daraus
		// `401 auth_project_mismatch` — wir liefern den Auth-Fehler
		// direkt, ohne den ingest-spezifischen Marker zu re-mappen.
		if errors.Is(err, domain.ErrIngestProjectIDMismatch) {
			return driving.IssueSessionTokenResult{}, domain.ErrAuthProjectMismatch
		}
		return driving.IssueSessionTokenResult{}, err
	}

	policy, err := s.Policies.ResolvePolicy(ctx, resolved)
	if err != nil {
		return driving.IssueSessionTokenResult{}, err
	}

	audience, err := normalizeAudience(req.Audience)
	if err != nil {
		return driving.IssueSessionTokenResult{}, err
	}
	originBound := strings.TrimSpace(req.Origin)
	if err := domain.ValidateAudience(policy, audience, originBound); err != nil {
		return driving.IssueSessionTokenResult{}, err
	}

	ttl, err := domain.ResolveTTLSeconds(policy, req.RequestedTTLSeconds)
	if err != nil {
		return driving.IssueSessionTokenResult{}, err
	}

	// Issuance-Bucket: Project-Policy hat Vorrang vor dem Adapter-
	// Default. `IsZero` heißt „nimm den Default" (RAK-74).
	allow, err := s.Limiter.Allow(ctx, resolved, policy.RateLimit.IssuanceBucket)
	if err != nil {
		return driving.IssueSessionTokenResult{}, err
	}
	if !allow {
		return driving.IssueSessionTokenResult{}, domain.ErrAuthIssuanceRateLimited
	}

	tokenID, err := s.IDs.NewTokenID()
	if err != nil {
		return driving.IssueSessionTokenResult{}, err
	}

	now := s.now()
	claims := domain.BuildSessionTokenClaims(domain.SessionTokenIssuanceInput{
		ProjectID:  resolved,
		Audience:   audience,
		TTLSeconds: ttl,
		SessionID:  optionalString(req.SessionID),
		Origin:     optionalString(originBound),
	}, tokenID, s.Issuer, now)

	signed, err := s.Signer.Sign(claims)
	if err != nil {
		return driving.IssueSessionTokenResult{}, err
	}

	out := driving.IssueSessionTokenResult{
		Value:     signed,
		TokenID:   claims.JTI,
		ProjectID: claims.Sub,
		Audience:  claims.Aud,
		ExpiresAt: claims.Exp,
	}
	if claims.SessionID != nil {
		out.SessionID = *claims.SessionID
	}
	return out, nil
}

// now liefert die aktuelle Clock — Tests injizieren eine deterministische
// Funktion, Produktion verwendet `time.Now.UTC`. Eine `nil`-Clock
// fällt auf den Default zurück, damit ein direkt instanziiertes
// Service-Objekt nicht panict.
func (s *IssueSessionTokenService) now() time.Time {
	if s.Now == nil {
		return time.Now().UTC()
	}
	return s.Now()
}

// normalizeAudience lehnt leere oder unbekannte Audiences ab und
// liefert sonst den getrimmten Wert. Der HTTP-Adapter prüft das
// zusätzlich gegen die Project-Policy-Allowlist über
// `domain.ValidateAudience`.
func normalizeAudience(raw string) (domain.SessionTokenAudience, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", domain.ErrAuthSessionScopeDenied
	}
	candidate := domain.SessionTokenAudience(trimmed)
	if !candidate.IsKnown() {
		return "", domain.ErrAuthSessionScopeDenied
	}
	return candidate, nil
}

// optionalString konvertiert einen leeren String in `nil` und einen
// gesetzten in einen *string. Wichtig, weil `domain.SessionTokenClaims`
// optionale Felder als String-Pointer modelliert (Unterscheidung
// „nicht gesetzt" vs. „leerer String").
func optionalString(s string) *string {
	t := strings.TrimSpace(s)
	if t == "" {
		return nil
	}
	return &t
}
