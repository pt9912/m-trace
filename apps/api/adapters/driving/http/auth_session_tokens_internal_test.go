package http

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// RAK-72: Tabellen-Test für `writeAuthError`.
// Pinnt das §3.9-Mapping für jede der zehn Fehlerklassen plus den
// Default-Fallback und `context.Canceled`/`DeadlineExceeded`. Lebt
// im internal-Test, weil die Funktion package-private ist.

func TestWriteAuthError_AllBranches(t *testing.T) {
	t.Parallel()
	logger := slog.Default()
	cases := []struct {
		name     string
		err      error
		wantCode int
		wantBody string
	}{
		{"missing", domain.ErrAuthTokenMissing, http.StatusUnauthorized, "auth_token_missing"},
		{"invalid", domain.ErrAuthTokenInvalid, http.StatusUnauthorized, "auth_token_invalid"},
		{"revoked", domain.ErrAuthTokenRevoked, http.StatusUnauthorized, "auth_token_revoked"},
		{"expired", domain.ErrAuthTokenExpired, http.StatusUnauthorized, "auth_token_expired"},
		{"not_yet_valid", domain.ErrAuthTokenNotYetValid, http.StatusUnauthorized, "auth_token_not_yet_valid"},
		{"project_mismatch", domain.ErrAuthProjectMismatch, http.StatusUnauthorized, "auth_project_mismatch"},
		{"scope_denied", domain.ErrAuthSessionScopeDenied, http.StatusForbidden, "auth_session_scope_denied"},
		{"policy_denied", domain.ErrAuthPolicyDenied, http.StatusForbidden, "auth_policy_denied"},
		{"ttl_too_large", domain.ErrAuthTokenTTLTooLarge, http.StatusUnprocessableEntity, "auth_token_ttl_too_large"},
		{"rate_limited", domain.ErrAuthIssuanceRateLimited, http.StatusTooManyRequests, "auth_issuance_rate_limited"},
		{"context_cancelled", context.Canceled, http.StatusServiceUnavailable, "service_unavailable"},
		{"deadline_exceeded", context.DeadlineExceeded, http.StatusServiceUnavailable, "service_unavailable"},
		{"unknown_default", errors.New("totally unknown"), http.StatusInternalServerError, "internal_error"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			writeAuthError(rec, logger, tc.err)
			if rec.Code != tc.wantCode {
				t.Errorf("status: want %d, got %d", tc.wantCode, rec.Code)
			}
			var body map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("body decode: %v", err)
			}
			if code, _ := body["code"].(string); code != tc.wantBody {
				t.Errorf("code: want %q, got %q", tc.wantBody, code)
			}
			if status, _ := body["status"].(string); status != "error" {
				t.Errorf("status field: want error, got %q", status)
			}
		})
	}
}

// TestWriteAuthError_NilLoggerSafe stellt sicher, dass der
// Default-Pfad auch ohne Logger nicht panict (Defensive Coding für
// Test-Setups ohne expliziten Logger).
func TestWriteAuthError_NilLoggerSafe(t *testing.T) {
	t.Parallel()
	rec := httptest.NewRecorder()
	writeAuthError(rec, nil, errors.New("totally unknown"))
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("nil logger fallback status: want 500, got %d", rec.Code)
	}
}

// errReader simuliert einen Body-Reader, der vor dem EOF-Marker
// einen IO-Fehler liefert — damit wird der `io.ReadAll`-Error-Branch
// von `readAuthBody` exercisiert.
type errReader struct{}

func (errReader) Read(_ []byte) (int, error) {
	return 0, errors.New("simulated read failure")
}

// TestAuthSessionTokensHandler_RejectsNonPOST exercisiert den
// Defense-in-Depth-Method-Guard direkt über ServeHTTP (statt über
// den Router, der GET schon vorher als 405 mappt). Pinnt, dass der
// Handler-eigene Guard `405 method_not_allowed` liefert.
func TestAuthSessionTokensHandler_RejectsNonPOST(t *testing.T) {
	t.Parallel()
	h := &AuthSessionTokensHandler{Logger: slog.Default()}
	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/auth/session-tokens", nil)
	h.ServeHTTP(rec, r)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET via ServeHTTP: want 405, got %d", rec.Code)
	}
}

func TestReadAuthBody_ReadErrorMapsTo400(t *testing.T) {
	t.Parallel()
	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/auth/session-tokens", errReader{})
	body, err := readAuthBody(rec, r)
	if err == nil {
		t.Fatalf("want err on read failure, got nil")
	}
	if body != nil {
		t.Errorf("body must be nil on error, got %d bytes", len(body))
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status: want 400, got %d", rec.Code)
	}
}
