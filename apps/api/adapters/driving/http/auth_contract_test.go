package http_test

import (
	_ "embed"
	"net/http"
	"strings"
	"testing"
)

// `0.12.0` Tranche 5 — Schema-Snapshot-Tests gegen die Auth-
// Fixtures aus spec/contract-fixtures/api/. Die Werte
// (`token_id`, `expires_at`, JWT-`value`) ändern sich pro Lab-Run;
// geprüft wird die Schlüssel-Struktur der Wire-Antwort plus harte
// Asserts auf sicherheitsrelevante Invarianten:
//   - Klartext `session_token.value` darf nur in der Issuance-
//     Antwort erscheinen, nirgends sonst.
//   - Fehler-Bodies tragen nur `status`, `code`, `message` —
//     keine Identifier, die der Aufrufer als Existenz- oder
//     Project-Hinweis interpretieren könnte.
//
// `make sync-contract-fixtures` kopiert die Quell-JSONs aus
// spec/contract-fixtures/api/ in dieses testdata/-Verzeichnis,
// damit der api-Docker-Build-Context (nur apps/api/) sie findet.

//go:embed testdata/auth-session-token-issue.json
var fixtureAuthSessionTokenIssue []byte

//go:embed testdata/auth-error-token-expired.json
var fixtureAuthErrorTokenExpired []byte

//go:embed testdata/auth-error-policy-denied.json
var fixtureAuthErrorPolicyDenied []byte

//go:embed testdata/auth-error-ttl-too-large.json
var fixtureAuthErrorTTLTooLarge []byte

//go:embed testdata/auth-error-issuance-rate-limited.json
var fixtureAuthErrorIssuanceRateLimited []byte

//go:embed testdata/auth-project-token-generation.json
var fixtureAuthProjectTokenGeneration []byte

func TestAuthContract_SessionTokenIssueMatchesFixture(t *testing.T) {
	t.Parallel()
	srv := newAuthSessionTestServer(t)
	resp := postAuthSessionToken(t, srv, `{"audience":"playback-events","ttl_seconds":120,"session_id":"sess_a"}`,
		map[string]string{"X-MTrace-Token": "demo-token"})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status: want 201, got %d", resp.StatusCode)
	}
	assertSchemaMatchesFixture(t, resp.Body, fixtureAuthSessionTokenIssue)
}

func TestAuthContract_TTLTooLargeMatchesFixture(t *testing.T) {
	t.Parallel()
	srv := newAuthSessionTestServer(t)
	resp := postAuthSessionToken(t, srv, `{"audience":"playback-events","ttl_seconds":10000}`,
		map[string]string{"X-MTrace-Token": "demo-token"})
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status: want 422, got %d", resp.StatusCode)
	}
	assertErrorBodyShape(t, resp.Body, fixtureAuthErrorTTLTooLarge, "auth_token_ttl_too_large")
}

func TestAuthContract_IssuanceRateLimitedMatchesFixture(t *testing.T) {
	t.Parallel()
	srv := newAuthSessionTestServer(t, withProjectIssuanceQuota(1, 0))
	body := `{"audience":"playback-events","ttl_seconds":60}`
	// Erstes Issue exhaustet das per-Project-Bucket (cap=1, refill=0).
	first := postAuthSessionToken(t, srv, body, map[string]string{"X-MTrace-Token": "demo-token"})
	if first.StatusCode != http.StatusCreated {
		t.Fatalf("first: want 201, got %d", first.StatusCode)
	}
	_ = first.Body.Close()
	resp := postAuthSessionToken(t, srv, body, map[string]string{"X-MTrace-Token": "demo-token"})
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("status: want 429, got %d", resp.StatusCode)
	}
	assertErrorBodyShape(t, resp.Body, fixtureAuthErrorIssuanceRateLimited, "auth_issuance_rate_limited")
}

// TestAuthContract_PolicyDeniedFixtureSelfValidation prüft nur die
// Fixture-Datei selbst — sie verifiziert nicht, dass eine echte
// API-Antwort zur Fixture passt. Der Trigger-Pfad (origin nicht in
// Project-Allowlist) wird im PlaybackEvents-Pfad
// (`TestCORS_Post_ProjectOriginMismatch_403`) abgedeckt, ohne
// die Fixture als Soll-Wert zu nutzen. Honest-Naming-Convention
// (Code-Review 2026-05-10): `…FixtureSelfValidation` statt
// `…MatchesFixture`, weil hier kein API↔Fixture-Pin passiert.
func TestAuthContract_PolicyDeniedFixtureSelfValidation(t *testing.T) {
	t.Parallel()
	if !strings.Contains(string(fixtureAuthErrorPolicyDenied), `"auth_policy_denied"`) {
		t.Errorf("policy-denied fixture missing expected code marker")
	}
	if !strings.Contains(string(fixtureAuthErrorPolicyDenied), `"status": "error"`) {
		t.Errorf("policy-denied fixture missing status:error marker")
	}
}

// TestAuthContract_TokenExpiredFixtureSelfValidation: analog zu
// Policy-Denied — Fixture-Inhalt-Pin, kein Trigger-Pfad. Der
// Trigger (abgelaufener Session Token) ist in
// `TestValidateClaimsTime_Boundaries` und
// `TestAuthHeaderParser_ExpiredSessionToken` abgedeckt, ohne
// gegen die Fixture zu vergleichen.
func TestAuthContract_TokenExpiredFixtureSelfValidation(t *testing.T) {
	t.Parallel()
	if !strings.Contains(string(fixtureAuthErrorTokenExpired), `"auth_token_expired"`) {
		t.Errorf("token-expired fixture missing expected code marker")
	}
}

// TestAuthContract_ProjectTokenGenerationFixtureShape pinnt die
// persistente Sicht einer `mtr_pt_*`-Generation. **Klartext-Token
// darf in dieser Fixture nicht vorkommen** — nur Hash, Fingerprint
// und Lifecycle-Felder. Verhindert eine Regression, in der jemand
// versehentlich `value` oder einen Klartext-Marker in die
// Persistenz-Sicht aufnimmt.
func TestAuthContract_ProjectTokenGenerationFixtureShape(t *testing.T) {
	t.Parallel()
	body := string(fixtureAuthProjectTokenGeneration)
	wantKeys := []string{
		`"token_id"`, `"project_id"`, `"key_hash"`, `"fingerprint"`,
		`"not_before"`, `"grace_until"`, `"expires_at"`, `"revoked_at"`,
		`"created_at"`, `"rotated_from"`,
	}
	for _, k := range wantKeys {
		if !strings.Contains(body, k) {
			t.Errorf("project-token-generation fixture missing key %s", k)
		}
	}
	// Klartext-Marker darf NICHT vorkommen.
	forbiddenMarkers := []string{`"value"`, `"plain"`, `"secret"`, `"mtr_pt_"`}
	for _, marker := range forbiddenMarkers {
		// `mtr_pt_` darf im Fingerprint vorkommen, aber niemals als
		// Klartext-Wert. Wir prüfen das über das Fehlen eines
		// `"value"`/`"secret"`/`"plain"`-Keys, der eine Klartext-
		// Persistenz andeuten würde.
		if marker == `"mtr_pt_"` {
			continue
		}
		if strings.Contains(body, marker) {
			t.Errorf("project-token-generation fixture must not contain plaintext marker %s", marker)
		}
	}
}
