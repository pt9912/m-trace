package http

import (
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// TestOutcomeFor_AllBuckets deckt jeden Status-Code-Bucket aus
// spec/telemetry-model.md ab. Reine Pure-Function-Tests sind
// günstiger als HTTP-Roundtrips, decken aber dieselbe Logik.
func TestOutcomeFor_AllBuckets(t *testing.T) {
	t.Parallel()
	cases := []struct {
		code int
		want string
	}{
		{http.StatusAccepted, "accepted"},
		{http.StatusUnauthorized, "unauthorized"},
		{http.StatusForbidden, "forbidden"},
		{http.StatusRequestEntityTooLarge, "too_large"},
		{http.StatusTooManyRequests, "rate_limited"},
		{http.StatusBadRequest, "invalid"},
		{http.StatusUnprocessableEntity, "invalid"},
		{http.StatusInternalServerError, "error"},
		{http.StatusBadGateway, "error"},
		{http.StatusOK, "other"},
		{http.StatusFound, "other"},
	}
	for _, tc := range cases {
		if got := outcomeFor(tc.code); got != tc.want {
			t.Errorf("outcomeFor(%d)=%q want %q", tc.code, got, tc.want)
		}
	}
}

// TestStatusRecorder_DefaultsAndExplicitWrite deckt die zwei
// Wrote-Header-Pfade in statusRecorder.Write/statusCode ab.
func TestStatusRecorder_DefaultsAndExplicitWrite(t *testing.T) {
	t.Parallel()

	t.Run("write without WriteHeader defaults to 200", func(t *testing.T) {
		t.Parallel()
		w := httptest.NewRecorder()
		rec := &statusRecorder{ResponseWriter: w}
		if _, err := rec.Write([]byte("hi")); err != nil {
			t.Fatalf("Write: %v", err)
		}
		if rec.statusCode() != http.StatusOK {
			t.Errorf("statusCode after Write-without-header=%d want 200", rec.statusCode())
		}
	})

	t.Run("WriteHeader stays sticky across multiple calls", func(t *testing.T) {
		t.Parallel()
		w := httptest.NewRecorder()
		rec := &statusRecorder{ResponseWriter: w}
		rec.WriteHeader(http.StatusBadRequest)
		// Zweiter Call darf den Status nicht überschreiben (defensiver
		// Recorder-Vertrag).
		rec.WriteHeader(http.StatusOK)
		if rec.statusCode() != http.StatusBadRequest {
			t.Errorf("statusCode=%d want 400 (sticky)", rec.statusCode())
		}
	})

	t.Run("statusCode pre-write returns 200", func(t *testing.T) {
		t.Parallel()
		rec := &statusRecorder{ResponseWriter: httptest.NewRecorder()}
		if got := rec.statusCode(); got != http.StatusOK {
			t.Errorf("statusCode pre-write=%d want 200", got)
		}
	})
}

// TestAppendVary pinnt das Verhalten von `appendVary`
// (Review-Finding Y2 — Token-für-Token-Union statt
// `contains "Origin"`-Heuristik). Alle drei Pflicht-Tokens
// (`Origin`, `Access-Control-Request-Method`,
// `Access-Control-Request-Headers`) werden idempotent unioniert.
func TestAppendVary(t *testing.T) {
	t.Parallel()

	wantFull := "Origin, Access-Control-Request-Method, Access-Control-Request-Headers"

	t.Run("empty header sets full Vary", func(t *testing.T) {
		t.Parallel()
		w := httptest.NewRecorder()
		appendVary(w)
		if got := w.Header().Get("Vary"); got != wantFull {
			t.Errorf("Vary=%q want %q", got, wantFull)
		}
	})

	t.Run("non-empty without any vary token appends all three", func(t *testing.T) {
		t.Parallel()
		w := httptest.NewRecorder()
		w.Header().Set("Vary", "Accept")
		appendVary(w)
		got := w.Header().Get("Vary")
		want := "Accept, " + wantFull
		if got != want {
			t.Errorf("Vary=%q want %q", got, want)
		}
	})

	t.Run("existing Vary with Origin alone still appends request-* tokens", func(t *testing.T) {
		t.Parallel()
		// Y2-Fix: vorher hat `appendVary` bei `Origin` allein die
		// request-* Tokens still verschluckt. Jetzt werden sie
		// einzeln unioniert.
		w := httptest.NewRecorder()
		w.Header().Set("Vary", "Origin")
		appendVary(w)
		got := w.Header().Get("Vary")
		want := "Origin, Access-Control-Request-Method, Access-Control-Request-Headers"
		if got != want {
			t.Errorf("Vary=%q want %q", got, want)
		}
	})

	t.Run("existing Vary with all three tokens is no-op", func(t *testing.T) {
		t.Parallel()
		w := httptest.NewRecorder()
		w.Header().Set("Vary", wantFull)
		before := w.Header().Get("Vary")
		appendVary(w)
		if after := w.Header().Get("Vary"); before != after {
			t.Errorf("Vary mutated although fully present: %q → %q", before, after)
		}
	})

	t.Run("token comparison is case-insensitive and respects boundaries", func(t *testing.T) {
		t.Parallel()
		w := httptest.NewRecorder()
		w.Header().Set("Vary", "origin, OriginCustom") // OriginCustom must NOT match Origin
		appendVary(w)
		got := w.Header().Get("Vary")
		// Origin is already present (case-insensitive); the two
		// request-* tokens are appended.
		if got != "origin, OriginCustom, Access-Control-Request-Method, Access-Control-Request-Headers" {
			t.Errorf("Vary=%q (case-insensitive Origin must dedup; OriginCustom must not match)", got)
		}
	})
}

// TestClientIPFromRequest_ParsesRemoteAddr deckt die drei Pfade in
// clientIPFromRequest ab: leer, host:port, kaputt (kein port).
func TestClientIPFromRequest_ParsesRemoteAddr(t *testing.T) {
	t.Parallel()
	cases := []struct {
		remote string
		want   string
	}{
		{"", ""},
		{"127.0.0.1:54321", "127.0.0.1"},
		{"[::1]:8080", "::1"},
		{"raw-without-port", "raw-without-port"},
	}
	for _, tc := range cases {
		req := httptest.NewRequest(http.MethodPost, "/x", nil)
		req.RemoteAddr = tc.remote
		if got := clientIPFromRequest(req); got != tc.want {
			t.Errorf("clientIPFromRequest(%q)=%q want %q", tc.remote, got, tc.want)
		}
	}
}

// TestWriteAuthHeaderError pinnt das §3.9-Mapping vom AuthHeaderParser-
// Fehler auf den HTTP-Status (Body bleibt minimal). Deckt jede der
// neun Fehlerklassen plus den Default-Fallback ab.
func TestWriteAuthHeaderError(t *testing.T) {
	t.Parallel()
	h := &PlaybackEventsHandler{Logger: slog.Default()}
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"missing", domain.ErrAuthTokenMissing, http.StatusUnauthorized},
		{"invalid", domain.ErrAuthTokenInvalid, http.StatusUnauthorized},
		{"revoked", domain.ErrAuthTokenRevoked, http.StatusUnauthorized},
		{"expired", domain.ErrAuthTokenExpired, http.StatusUnauthorized},
		{"not_yet_valid", domain.ErrAuthTokenNotYetValid, http.StatusUnauthorized},
		{"project_mismatch", domain.ErrAuthProjectMismatch, http.StatusUnauthorized},
		{"scope_denied", domain.ErrAuthSessionScopeDenied, http.StatusForbidden},
		{"policy_denied", domain.ErrAuthPolicyDenied, http.StatusForbidden},
		{"unknown_default", errors.New("totally unknown"), http.StatusUnauthorized},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			h.writeAuthHeaderError(rec, tc.err)
			if rec.Code != tc.want {
				t.Errorf("status: want %d, got %d", tc.want, rec.Code)
			}
		})
	}
}

// TestWriteUseCaseError_AuthLifecycleErrors pinnt das §3.9-Mapping
// für Lifecycle-Fehler aus dem RotatingProjectResolver-Pfad
// (RAK-73). Vorher wurden diese auf 500 gemappt; ab sind
// sie 401/403.
func TestWriteUseCaseError_AuthLifecycleErrors(t *testing.T) {
	t.Parallel()
	h := &PlaybackEventsHandler{Logger: slog.Default()}
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"schema_version", domain.ErrSchemaVersionMismatch, http.StatusBadRequest},
		{"unauthorized", domain.ErrUnauthorized, http.StatusUnauthorized},
		{"auth_token_missing", domain.ErrAuthTokenMissing, http.StatusUnauthorized},
		{"auth_token_invalid", domain.ErrAuthTokenInvalid, http.StatusUnauthorized},
		{"auth_token_revoked", domain.ErrAuthTokenRevoked, http.StatusUnauthorized},
		{"auth_token_expired", domain.ErrAuthTokenExpired, http.StatusUnauthorized},
		{"auth_token_not_yet_valid", domain.ErrAuthTokenNotYetValid, http.StatusUnauthorized},
		{"auth_project_mismatch", domain.ErrAuthProjectMismatch, http.StatusUnauthorized},
		{"auth_session_scope_denied", domain.ErrAuthSessionScopeDenied, http.StatusForbidden},
		{"auth_policy_denied", domain.ErrAuthPolicyDenied, http.StatusForbidden},
		{"origin_not_allowed", domain.ErrOriginNotAllowed, http.StatusForbidden},
		{"batch_empty", domain.ErrBatchEmpty, http.StatusUnprocessableEntity},
		{"batch_too_large", domain.ErrBatchTooLarge, http.StatusUnprocessableEntity},
		{"invalid_event", domain.ErrInvalidEvent, http.StatusUnprocessableEntity},
		{"rate_limited", domain.ErrRateLimited, http.StatusTooManyRequests},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			h.writeUseCaseError(rec, tc.err)
			if rec.Code != tc.want {
				t.Errorf("status: want %d, got %d", tc.want, rec.Code)
			}
		})
	}
}

// TestPlaybackClientIP_XFFTrustBoundary (R-26 b): die client_ip-
// Rate-Limit-Dimension folgt der MTRACE_TRUST_FORWARDED_FOR-Boundary —
// mit Opt-in zählt das letzte XFF-Element (hinter LB/Proxy ist
// RemoteAddr die Proxy-IP: EIN Bucket für alle Clients), ohne Opt-in
// bleibt RemoteAddr maßgeblich (XFF wäre client-kontrolliert).
func TestPlaybackClientIP_XFFTrustBoundary(t *testing.T) {
	t.Parallel()
	mk := func(xff string) *http.Request {
		r := httptest.NewRequest(http.MethodPost, "/api/playback-events", nil)
		r.RemoteAddr = "10.0.0.9:4711"
		if xff != "" {
			r.Header.Set("X-Forwarded-For", xff)
		}
		return r
	}
	cases := []struct {
		name  string
		trust bool
		xff   string
		want  string
	}{
		{"untrusted ignores XFF", false, "203.0.113.7", "10.0.0.9"},
		{"trusted uses last XFF hop", true, "198.51.100.2, 203.0.113.7", "203.0.113.7"},
		{"trusted without XFF falls back to RemoteAddr", true, "", "10.0.0.9"},
		// Validierungs-Kontrakt: die XFF-abgeleitete IP landet raw in
		// Redis-Bucket-Keys — Nicht-IP-Werte (Spoofing/kaputter Proxy)
		// dürfen dort NIE ankommen und fallen auf RemoteAddr zurück.
		{"trusted rejects non-IP XFF", true, "evil" + strings.Repeat("x", 512), "10.0.0.9"},
		{"trusted rejects host:port XFF", true, "203.0.113.7:1234", "10.0.0.9"},
		{"trusted canonicalizes IPv6", true, "2001:DB8:0:0:0:0:0:1", "2001:db8::1"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := &PlaybackEventsHandler{TrustForwardedFor: tc.trust}
			if got := h.clientIP(mk(tc.xff)); got != tc.want {
				t.Fatalf("clientIP = %q, want %q", got, tc.want)
			}
		})
	}
}
