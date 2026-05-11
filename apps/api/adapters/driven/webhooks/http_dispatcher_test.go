package webhooks_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/webhooks"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

func newSilentLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

func sampleEvent() driven.OutboundWebhookEvent {
	return driven.OutboundWebhookEvent{
		EventID:    "evt_123",
		Kind:       domain.StreamLifecycleEventStarted,
		ProjectID:  "p1",
		StreamID:   "ing_xyz",
		ObservedAt: "2026-05-11T12:00:00Z",
		Source:     domain.StreamLifecycleSourceSmoke,
	}
}

func TestOutboundWebhook_DisabledIsNoop(t *testing.T) {
	t.Parallel()
	d := webhooks.NewHTTPDispatcher("", nil, newSilentLogger())
	if err := d.Dispatch(context.Background(), sampleEvent()); err != nil {
		t.Errorf("disabled dispatcher must be no-op, got %v", err)
	}
}

func TestOutboundWebhook_NilDispatcherIsNoop(t *testing.T) {
	t.Parallel()
	var d *webhooks.HTTPDispatcher
	if err := d.Dispatch(context.Background(), sampleEvent()); err != nil {
		t.Errorf("nil receiver must be no-op, got %v", err)
	}
}

func TestOutboundWebhook_DefaultHelpersAreSafe(t *testing.T) {
	t.Parallel()
	// Construct via zero value to exercise the default-helper branches
	// (`client()`/`requestTimeout()`/`now()` falling back to defaults
	// when the optional fields are not set explicitly).
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	d := &webhooks.HTTPDispatcher{
		Endpoint: srv.URL,
		Secret:   []byte("s"),
		// HTTPClient, Now, RequestTimeout, BaseBackoff intentionally unset
	}
	if err := d.Dispatch(context.Background(), sampleEvent()); err != nil {
		t.Errorf("default-helper dispatcher should succeed: %v", err)
	}
}

func TestOutboundWebhook_HappyPath(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	d := webhooks.NewHTTPDispatcher(srv.URL, []byte("secret"), newSilentLogger())
	if err := d.Dispatch(context.Background(), sampleEvent()); err != nil {
		t.Fatalf("happy path: %v", err)
	}
	if got := calls.Load(); got != 1 {
		t.Errorf("want 1 call, got %d", got)
	}
}

func TestOutboundWebhook_HMACSignatureMatches(t *testing.T) {
	t.Parallel()
	var (
		gotSig  string
		gotBody []byte
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotSig = r.Header.Get("X-MTrace-Signature")
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	secret := []byte("topsecret")
	d := webhooks.NewHTTPDispatcher(srv.URL, secret, newSilentLogger())
	if err := d.Dispatch(context.Background(), sampleEvent()); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if !strings.HasPrefix(gotSig, "sha256=") {
		t.Fatalf("signature header missing sha256= prefix: %q", gotSig)
	}
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write(gotBody)
	want := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	if gotSig != want {
		t.Errorf("HMAC mismatch:\n  got  %s\n  want %s", gotSig, want)
	}
}

func TestOutboundWebhook_RetrySucceedsAfterTransientFailure(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n < 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	d := webhooks.NewHTTPDispatcher(srv.URL, []byte("s"), newSilentLogger())
	d.BaseBackoff = 5 * time.Millisecond
	if err := d.Dispatch(context.Background(), sampleEvent()); err != nil {
		t.Fatalf("retry-success: %v", err)
	}
	if got := calls.Load(); got != 2 {
		t.Errorf("want 2 calls (1 fail + 1 success), got %d", got)
	}
}

func TestOutboundWebhook_DeadLetterAfterMaxAttempts(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	d := webhooks.NewHTTPDispatcher(srv.URL, []byte("s"), newSilentLogger())
	d.BaseBackoff = 5 * time.Millisecond
	d.MaxAttempts = 3
	err := d.Dispatch(context.Background(), sampleEvent())
	if !errors.Is(err, webhooks.ErrOutboundWebhookExhausted) {
		t.Errorf("want ErrOutboundWebhookExhausted, got %v", err)
	}
	if got := calls.Load(); got != 3 {
		t.Errorf("want 3 calls (exhausted), got %d", got)
	}
}

func TestOutboundWebhook_BodyShapeIsStable(t *testing.T) {
	t.Parallel()
	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &got)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	d := webhooks.NewHTTPDispatcher(srv.URL, []byte("s"), newSilentLogger())
	if err := d.Dispatch(context.Background(), sampleEvent()); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	expected := []string{"event_id", "type", "project_id", "stream_id", "observed_at", "source"}
	for _, key := range expected {
		if _, ok := got[key]; !ok {
			t.Errorf("body missing required key %q (got %v)", key, got)
		}
	}
	if got["type"] != "stream_started" {
		t.Errorf("type field: want stream_started, got %v", got["type"])
	}
	// Klartext-Stream-Key darf NIE in der Payload landen.
	body, _ := json.Marshal(got)
	if strings.Contains(strings.ToLower(string(body)), "stream_key") {
		t.Errorf("payload must not leak stream_key field, got %s", body)
	}
}

func TestOutboundWebhook_ContextCancelStopsRetry(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(srv.Close)

	d := webhooks.NewHTTPDispatcher(srv.URL, []byte("s"), newSilentLogger())
	d.BaseBackoff = 50 * time.Millisecond
	d.MaxAttempts = 5

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	err := d.Dispatch(ctx, sampleEvent())
	if err == nil {
		t.Fatalf("expected error after ctx cancel, got nil")
	}
	if got := calls.Load(); got > 2 {
		t.Errorf("ctx cancel should stop early, got %d calls", got)
	}
}
