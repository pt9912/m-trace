package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestWriteCursorError_MapsAllSentinelErrors verifiziert das Mapping
// der drei Cursor-Sentinel-Errors auf HTTP-Status und Body gemäß
// API-Kontrakt §10.3 / ADR-0004 §6. Der `errCursorExpired`-Pfad ist
// aus dem Decode-Pfad heute nicht erreichbar (Retention-Folge-Arbeit
// in 0.4.0+), wird aber hier am Mapping-Helper verifiziert, damit der
// 410-Vertrag eingefroren bleibt.
func TestWriteCursorError_MapsAllSentinelErrors(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		err      error
		wantCode int
		wantErr  string
	}{
		{"legacy", errCursorInvalidLegacy, http.StatusBadRequest, "cursor_invalid_legacy"},
		{"malformed", errCursorInvalidMalformed, http.StatusBadRequest, "cursor_invalid_malformed"},
		{"expired", errCursorExpired, http.StatusGone, "cursor_expired"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			rec := httptest.NewRecorder()
			writeCursorError(rec, c.err)
			if rec.Code != c.wantCode {
				t.Errorf("status = %d, want %d", rec.Code, c.wantCode)
			}
			if got := rec.Header().Get("Content-Type"); got != "application/json" {
				t.Errorf("content-type = %q, want application/json", got)
			}
			var body map[string]string
			if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body["error"] != c.wantErr {
				t.Errorf("body[error] = %q, want %q", body["error"], c.wantErr)
			}
			if body["reason"] == "" {
				t.Errorf("body[reason] missing — API-Kontrakt §10.3 fordert eine Erklärung")
			}
			// Kein Retry-After in der Cursor-Fehler-Antwort
			// (ADR-0004 §6 Recovery-Verhalten).
			if got := rec.Header().Get("Retry-After"); got != "" {
				t.Errorf("Retry-After = %q, want empty", got)
			}
		})
	}
}

// TestWriteCursorError_UnknownErrorFallsBackTo500 verifiziert den
// Default-Branch: ein unbekannter Error landet bei 500. Sollte nicht
// erreichbar sein (decode-Pfad liefert nur die drei Sentinels), aber
// der Fallback ist als Defense-in-Depth da.
func TestWriteCursorError_UnknownErrorFallsBackTo500(t *testing.T) {
	t.Parallel()
	rec := httptest.NewRecorder()
	writeCursorError(rec, errors.New("unexpected"))
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}
