package mediaserver_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/mediaserver"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

func sampleStream(id string) domain.IngestStream {
	return domain.IngestStream{
		ID:        id,
		ProjectID: "demo",
		Protocol:  domain.IngestProtocolSRT,
		CreatedAt: time.Now(),
	}
}

// TestMediaMTX_Apply_HappyPath (plan-0.12.6 Tranche 9 / R-15):
// MediaMTX antwortet 200 OK → State `applied`.
func TestMediaMTX_Apply_HappyPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	p, err := mediaserver.New(mediaserver.Config{Endpoint: srv.URL}, nil)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res, err := p.Apply(context.Background(), driven.MediaServerApplyInput{
		ProjectID: "demo", Stream: sampleStream("ing_001"),
	})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if res.State != driven.MediaServerStateApplied {
		t.Errorf("state = %s, want applied", res.State)
	}
	if gotPath != "/v3/config/paths/add/ing_001" {
		t.Errorf("path = %s, want /v3/config/paths/add/ing_001", gotPath)
	}
}

// TestMediaMTX_Apply_IdempotentOnConflict: 409 wird als applied
// behandelt (Path schon angelegt).
func TestMediaMTX_Apply_IdempotentOnConflict(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"error":"already exists"}`))
	}))
	t.Cleanup(srv.Close)
	p, _ := mediaserver.New(mediaserver.Config{Endpoint: srv.URL}, nil)
	res, _ := p.Apply(context.Background(), driven.MediaServerApplyInput{ProjectID: "demo", Stream: sampleStream("ing_dup")})
	if res.State != driven.MediaServerStateApplied {
		t.Errorf("state = %s, want applied (idempotent on 409)", res.State)
	}
}

// TestMediaMTX_Apply_AuthFailure: 401 → State failed mit ErrorCode
// `auth_failure`.
func TestMediaMTX_Apply_AuthFailure(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	t.Cleanup(srv.Close)
	p, _ := mediaserver.New(mediaserver.Config{Endpoint: srv.URL}, nil)
	res, _ := p.Apply(context.Background(), driven.MediaServerApplyInput{ProjectID: "demo", Stream: sampleStream("ing_a")})
	if res.State != driven.MediaServerStateFailed || res.ErrorCode != "auth_failure" {
		t.Errorf("got state=%s code=%s, want failed/auth_failure", res.State, res.ErrorCode)
	}
}

// TestMediaMTX_Apply_ServerError: 5xx → failed mit
// `server_status_5xx`.
func TestMediaMTX_Apply_ServerError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal explosion"))
	}))
	t.Cleanup(srv.Close)
	p, _ := mediaserver.New(mediaserver.Config{Endpoint: srv.URL}, nil)
	res, _ := p.Apply(context.Background(), driven.MediaServerApplyInput{ProjectID: "demo", Stream: sampleStream("ing_b")})
	if res.State != driven.MediaServerStateFailed {
		t.Errorf("state = %s, want failed", res.State)
	}
	if res.ErrorCode != "server_status_500" {
		t.Errorf("error_code = %s, want server_status_500", res.ErrorCode)
	}
}

// TestMediaMTX_Apply_Unreachable: kein Server → failed mit
// `unreachable`.
func TestMediaMTX_Apply_Unreachable(t *testing.T) {
	t.Parallel()
	p, _ := mediaserver.New(mediaserver.Config{
		Endpoint: "http://127.0.0.1:1", // garantiert unreachable
		HTTPClient: &http.Client{Timeout: 250 * time.Millisecond},
	}, nil)
	res, _ := p.Apply(context.Background(), driven.MediaServerApplyInput{ProjectID: "demo", Stream: sampleStream("ing_c")})
	if res.State != driven.MediaServerStateFailed {
		t.Errorf("state = %s, want failed", res.State)
	}
	if res.ErrorCode != "unreachable" {
		t.Errorf("error_code = %s, want unreachable", res.ErrorCode)
	}
}

// TestMediaMTX_AuthToken_Header: konfigurierter AuthToken landet im
// `Authorization`-Header.
func TestMediaMTX_AuthToken_Header(t *testing.T) {
	t.Parallel()
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	p, _ := mediaserver.New(mediaserver.Config{Endpoint: srv.URL, AuthToken: "secret-xyz"}, nil)
	_, _ = p.Apply(context.Background(), driven.MediaServerApplyInput{ProjectID: "demo", Stream: sampleStream("ing_auth")})
	if gotAuth != "Bearer secret-xyz" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer secret-xyz")
	}
}

// TestMediaMTX_Rollback_HappyPath.
func TestMediaMTX_Rollback_HappyPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	p, _ := mediaserver.New(mediaserver.Config{Endpoint: srv.URL}, nil)
	if err := p.Rollback(context.Background(), "demo", "ing_r1"); err != nil {
		t.Errorf("Rollback: %v", err)
	}
	if gotPath != "/v3/config/paths/delete/ing_r1" {
		t.Errorf("path = %s", gotPath)
	}
}

// TestMediaMTX_Rollback_NotFoundIsOK: 404 wird ignoriert (Idempotenz).
func TestMediaMTX_Rollback_NotFoundIsOK(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)
	p, _ := mediaserver.New(mediaserver.Config{Endpoint: srv.URL}, nil)
	if err := p.Rollback(context.Background(), "demo", "ing_gone"); err != nil {
		t.Errorf("Rollback: %v (404 should be tolerated)", err)
	}
}

// TestMediaMTX_Rollback_ServerError: 500 propagiert.
func TestMediaMTX_Rollback_ServerError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)
	p, _ := mediaserver.New(mediaserver.Config{Endpoint: srv.URL}, nil)
	if err := p.Rollback(context.Background(), "demo", "ing_x"); err == nil {
		t.Errorf("expected error on 500")
	}
}

// TestMediaMTX_New_MissingEndpoint: Pflicht-Endpoint fehlt.
func TestMediaMTX_New_MissingEndpoint(t *testing.T) {
	t.Parallel()
	if _, err := mediaserver.New(mediaserver.Config{}, nil); err == nil {
		t.Errorf("expected error on missing endpoint")
	}
}
