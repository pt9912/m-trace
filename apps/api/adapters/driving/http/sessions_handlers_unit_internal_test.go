package http

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel/trace/noop"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driving"
)

// erroringSessions liefert für jeden Aufruf einen synthetischen
// non-domain-Fehler — deckt damit den default-Pfad
// (Logger.Error + 500) in den GET-Handlern (plan-0.1.0.md §5.1).
type erroringSessions struct{}

func (erroringSessions) ListSessions(_ context.Context, _ driving.ListSessionsInput) (driving.ListSessionsResult, error) {
	return driving.ListSessionsResult{}, errors.New("synthetic backend failure")
}

func (erroringSessions) GetSession(_ context.Context, _ driving.GetSessionInput) (driving.GetSessionResult, error) {
	return driving.GetSessionResult{}, errors.New("synthetic backend failure")
}

// staticResolver löst einen festen Token auf ein festes Project auf,
// damit die Tests ab plan-0.4.0 §4.2 die Token-Resolution-Stufe der
// Read-Handler erfüllen, ohne ein echtes Auth-Backend zu wiren.
type staticResolver struct {
	token string
	id    string
}

func (s staticResolver) ResolveByToken(_ context.Context, token string) (domain.Project, error) {
	if token != s.token {
		return domain.Project{}, domain.ErrUnauthorized
	}
	return domain.Project{ID: s.id, Token: domain.ProjectToken(token)}, nil
}

const testToken = "test-token"
const testProjectID = "demo"

func newTestResolver() staticResolver {
	return staticResolver{token: testToken, id: testProjectID}
}

func newDiscardLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

// TestSessionsListHandler_BackendErrorMapsTo500 deckt den default-Pfad
// (default-Error → 500) inklusive Logger.Error.
func TestSessionsListHandler_BackendErrorMapsTo500(t *testing.T) {
	t.Parallel()
	h := &SessionsListHandler{
		UseCase:  erroringSessions{},
		Resolver: newTestResolver(),
		Tracer:   noop.NewTracerProvider().Tracer("test"),
		Logger:   newDiscardLogger(),
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/stream-sessions", nil)
	req.Header.Set("X-MTrace-Token", testToken)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status=%d want 500", rec.Code)
	}
}

// TestSessionsGetHandler_BackendErrorMapsTo500 — analog für den
// Detail-Endpoint.
func TestSessionsGetHandler_BackendErrorMapsTo500(t *testing.T) {
	t.Parallel()
	h := &SessionsGetHandler{
		UseCase:  erroringSessions{},
		Resolver: newTestResolver(),
		Tracer:   noop.NewTracerProvider().Tracer("test"),
		Logger:   newDiscardLogger(),
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/stream-sessions/sess-1", nil)
	req.Header.Set("X-MTrace-Token", testToken)
	req.SetPathValue("id", "sess-1")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status=%d want 500", rec.Code)
	}
}

// TestSessionsListHandler_RejectsNonGet deckt die explizite Method-
// Guard. Geht über die Mux-Pattern hinaus (mux würde non-GET nicht
// matchen), aber der Handler ist defensiv.
func TestSessionsListHandler_RejectsNonGet(t *testing.T) {
	t.Parallel()
	h := &SessionsListHandler{
		UseCase: erroringSessions{},
		Tracer:  noop.NewTracerProvider().Tracer("test"),
		Logger:  newDiscardLogger(),
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/stream-sessions", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status=%d want 405", rec.Code)
	}
}
