package http_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apihttp "github.com/pt9912/m-trace/apps/api/adapters/driving/http"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driving"
)

// stubValidateOnly ist ein minimaler `IngestControlInbound`-Stub für
// die Auth-Bridge-Tests — nur `ValidateKey` wird gebraucht, der Rest
// liefert Fehler/leere Antworten.
type stubValidateOnly struct {
	want struct {
		project, stream, key string
	}
	valid bool
	err   error
}

func (s *stubValidateOnly) CreateStream(_ context.Context, _ driving.CreateStreamRequest) (driving.CreateStreamResult, error) {
	return driving.CreateStreamResult{}, errors.New("not implemented")
}
func (s *stubValidateOnly) ListStreams(_ context.Context, _ string) ([]domain.IngestStream, error) {
	return nil, errors.New("not implemented")
}
func (s *stubValidateOnly) GetStreamDetail(_ context.Context, _, _ string) (driving.StreamDetail, error) {
	return driving.StreamDetail{}, errors.New("not implemented")
}
func (s *stubValidateOnly) RotateKey(_ context.Context, _, _ string) (driving.RotateKeyResult, error) {
	return driving.RotateKeyResult{}, errors.New("not implemented")
}
func (s *stubValidateOnly) ValidateKey(_ context.Context, projectID, streamID, key string) (driving.ValidateKeyResult, error) {
	if s.err != nil {
		return driving.ValidateKeyResult{}, s.err
	}
	if projectID == s.want.project && streamID == s.want.stream && key == s.want.key {
		return driving.ValidateKeyResult{Valid: s.valid, StreamID: streamID, KeyFingerprint: "fp"}, nil
	}
	return driving.ValidateKeyResult{Valid: false}, nil
}
func (s *stubValidateOnly) RecordLifecycleEvent(_ context.Context, _ driving.LifecycleHookRequest) (driving.LifecycleEventResult, error) {
	return driving.LifecycleEventResult{}, errors.New("not implemented")
}
func (s *stubValidateOnly) GetMediaServerConfig(_ context.Context, _, _ string) (driving.MediaServerConfigResult, error) {
	return driving.MediaServerConfigResult{}, errors.New("not implemented")
}

func newAuthHookServer(t *testing.T, stub *stubValidateOnly) *httptest.Server {
	t.Helper()
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	handler := &apihttp.MediaMTXAuthHookHandler{UseCase: stub, Logger: logger}
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

func TestMediaMTXAuthHook_AllowOnValidKey(t *testing.T) {
	t.Parallel()
	stub := &stubValidateOnly{valid: true}
	stub.want.project = "p1"
	stub.want.stream = "ing_xyz"
	stub.want.key = "secret"
	srv := newAuthHookServer(t, stub)

	body := strings.NewReader("user=p1&password=secret&action=publish&path=ing_xyz")
	req, _ := http.NewRequest(http.MethodPost, srv.URL, body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("valid key: want 200, got %d", resp.StatusCode)
	}
}

func TestMediaMTXAuthHook_DenyOnInvalidKey(t *testing.T) {
	t.Parallel()
	stub := &stubValidateOnly{valid: false}
	stub.want.project = "p1"
	stub.want.stream = "ing_xyz"
	stub.want.key = "secret"
	srv := newAuthHookServer(t, stub)

	body := strings.NewReader("user=p1&password=wrong&action=publish&path=ing_xyz")
	req, _ := http.NewRequest(http.MethodPost, srv.URL, body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("invalid key: want 403, got %d", resp.StatusCode)
	}
}

func TestMediaMTXAuthHook_DenyOnReadAction(t *testing.T) {
	t.Parallel()
	srv := newAuthHookServer(t, &stubValidateOnly{valid: true})

	body := strings.NewReader("user=p1&password=s&action=read&path=ing_xyz")
	req, _ := http.NewRequest(http.MethodPost, srv.URL, body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("read action: want 403, got %d", resp.StatusCode)
	}
}

func TestMediaMTXAuthHook_DenyOnMissingField(t *testing.T) {
	t.Parallel()
	srv := newAuthHookServer(t, &stubValidateOnly{valid: true})

	body := strings.NewReader("user=&password=s&action=publish&path=ing_xyz")
	req, _ := http.NewRequest(http.MethodPost, srv.URL, body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("missing user: want 403, got %d", resp.StatusCode)
	}
}

func TestMediaMTXAuthHook_RejectsNonFormContentType(t *testing.T) {
	t.Parallel()
	srv := newAuthHookServer(t, &stubValidateOnly{valid: true})

	body := strings.NewReader(`{"user":"p1"}`)
	req, _ := http.NewRequest(http.MethodPost, srv.URL, body)
	req.Header.Set("Content-Type", "application/json")
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("non-form content type: want 400, got %d", resp.StatusCode)
	}
}

func TestMediaMTXAuthHook_RejectsNonPOST(t *testing.T) {
	t.Parallel()
	srv := newAuthHookServer(t, &stubValidateOnly{valid: true})

	req, _ := http.NewRequest(http.MethodGet, srv.URL, nil)
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("GET: want 405, got %d", resp.StatusCode)
	}
}

func TestMediaMTXAuthHook_DenyOnValidateError(t *testing.T) {
	t.Parallel()
	stub := &stubValidateOnly{err: errors.New("repo down")}
	srv := newAuthHookServer(t, stub)

	body := strings.NewReader("user=p1&password=s&action=publish&path=ing_xyz")
	req, _ := http.NewRequest(http.MethodPost, srv.URL, body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("validate error must fail-closed: want 403, got %d", resp.StatusCode)
	}
}
