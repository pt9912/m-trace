package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestOriginRateLimitMiddleware_NilLimiterPassesThrough (plan-0.12.6
// Tranche 6 / R-22): wenn `limiter=nil`, ist die Middleware ein
// 1:1-Pass-Through (kein Wrap; Disabled-Pfad aus dem Boot-Validator).
func TestOriginRateLimitMiddleware_NilLimiterPassesThrough(t *testing.T) {
	t.Parallel()
	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	wrapped := originRateLimitMiddleware(inner, nil, false, nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))
	if rec.Code != http.StatusTeapot {
		t.Errorf("status = %d, want %d (pass-through)", rec.Code, http.StatusTeapot)
	}
}

// TestClientIPFromRemoteAddr_EdgeCases (plan-0.12.6 Tranche 6 / R-22):
// leere RemoteAddr → Empty-String (No-Op-Key); host-only RemoteAddr
// → `ip:<host>`-Key; host:port-Form → entkoppelter Host als Key.
func TestClientIPFromRemoteAddr_EdgeCases(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		remoteAddr string
		want       string
	}{
		{"empty remote addr", "", ""},
		{"host:port form", "1.2.3.4:54321", "ip:1.2.3.4"},
		{"host-only form (test server)", "1.2.3.4", "ip:1.2.3.4"},
		{"ipv6 with port", "[::1]:54321", "ip:::1"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := httptest.NewRequest(http.MethodGet, "/x", nil)
			r.RemoteAddr = tc.remoteAddr
			got := clientIPFromRemoteAddr(r)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
