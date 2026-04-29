package http_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apihttp "github.com/pt9912/m-trace/apps/api/adapters/driving/http"
)

// TestHealthHandler_ReturnsOK verifiziert den Pflicht-Smoke-Test (Spec
// §6.1, AK-1/AK-2): GET /api/health → 200 mit JSON-Body
// {"status":"ok"}.
func TestHealthHandler_ReturnsOK(t *testing.T) {
	t.Parallel()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	apihttp.HealthHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status=%d want 200", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type=%q want application/json", got)
	}
	body, _ := io.ReadAll(rec.Body)
	if !strings.Contains(string(body), `"status":"ok"`) {
		t.Errorf("body=%q want to contain status:ok", body)
	}
}
