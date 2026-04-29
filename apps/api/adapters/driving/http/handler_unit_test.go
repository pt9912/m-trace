package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestOutcomeFor_AllBuckets deckt jeden Status-Code-Bucket aus
// docs/telemetry-model.md §2.1 ab. Reine Pure-Function-Tests sind
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

// TestAppendVary deckt alle drei Pfade in appendVary ab: leerer
// Vary-Header, vorhandener Vary ohne Origin, vorhandener Vary mit
// Origin (no-op).
func TestAppendVary(t *testing.T) {
	t.Parallel()

	t.Run("empty header sets full Vary", func(t *testing.T) {
		t.Parallel()
		w := httptest.NewRecorder()
		appendVary(w)
		got := w.Header().Get("Vary")
		if got != varyHeader {
			t.Errorf("Vary=%q want %q", got, varyHeader)
		}
	})

	t.Run("non-empty without Origin appends", func(t *testing.T) {
		t.Parallel()
		w := httptest.NewRecorder()
		w.Header().Set("Vary", "Accept")
		appendVary(w)
		got := w.Header().Get("Vary")
		if got == "Accept" {
			t.Errorf("Vary unchanged (want appended): %q", got)
		}
	})

	t.Run("existing Vary with Origin is no-op", func(t *testing.T) {
		t.Parallel()
		w := httptest.NewRecorder()
		w.Header().Set("Vary", "Origin, Accept")
		before := w.Header().Get("Vary")
		appendVary(w)
		after := w.Header().Get("Vary")
		if before != after {
			t.Errorf("Vary mutated although already contained Origin: %q → %q", before, after)
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
