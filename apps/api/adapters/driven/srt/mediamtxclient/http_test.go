package mediamtxclient_test

import (
	"context"
	_ "embed"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/srt/mediamtxclient"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// Compile-Time-Check: Der HTTP-Adapter erfüllt den Driven-Port.
var _ driven.SrtSource = (*mediamtxclient.HTTPSrtSource)(nil)

// fixtureBody ist die anonymisierte Probe-Antwort aus Sub-1.2 (siehe
// spec/contract-fixtures/srt/mediamtx-srtconns-list.json). `make
// sync-contract-fixtures` kopiert die Spec-Datei in das testdata/-
// Verzeichnis, weil der api-Docker-Build-Context nur apps/api/ kennt.
//
//go:embed testdata/mediamtx-srtconns-list.json
var fixtureBody []byte

func fixedNow() time.Time { return time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC) }

// TestSnapshot_FromFixture: erfolgreicher Round-Trip gegen die
// gecapturte MediaMTX-Antwort aus Sub-1.2.
func TestSnapshot_FromFixture(t *testing.T) {
	t.Parallel()
	body := fixtureBody

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/srtconns/list" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)

	src := mediamtxclient.New(srv.URL, mediamtxclient.WithNow(fixedNow))
	got, err := src.SnapshotConnections(context.Background())
	if err != nil {
		t.Fatalf("SnapshotConnections: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 sample, got %d", len(got))
	}
	s := got[0]
	if s.StreamID != "srt-test" || s.ConnectionID != "00000000-0000-0000-0000-000000000001" {
		t.Errorf("unexpected identifiers: %+v", s)
	}
	if s.ConnectionState != domain.ConnectionStateConnected {
		t.Errorf("expected connected, got %s", s.ConnectionState)
	}
	if s.RTTMillis < 0.3 || s.RTTMillis > 0.4 {
		t.Errorf("rtt out of fixture range: %v", s.RTTMillis)
	}
	// mbpsLinkCapacity = 4352.217... Mbps × 1_000_000.
	if s.AvailableBandwidthBPS < 4_352_000_000 || s.AvailableBandwidthBPS > 4_353_000_000 {
		t.Errorf("available bandwidth out of fixture range: %d", s.AvailableBandwidthBPS)
	}
	if s.PacketLossTotal != 0 || s.RetransmissionsTotal != 0 {
		t.Errorf("expected zero loss/retrans in fixture, got loss=%d retrans=%d", s.PacketLossTotal, s.RetransmissionsTotal)
	}
	// SourceSequence = strconv(BytesReceived).
	if s.SourceSequence != "36414672" {
		t.Errorf("expected sequence='36414672', got %q", s.SourceSequence)
	}
	if !s.CollectedAt.Equal(fixedNow()) {
		t.Errorf("CollectedAt = %v, want %v", s.CollectedAt, fixedNow())
	}
	// Throughput = mbpsReceiveRate × 1_000_000 (1.137 → 1_137_417).
	if s.ThroughputBPS == nil || *s.ThroughputBPS < 1_000_000 || *s.ThroughputBPS > 2_000_000 {
		t.Errorf("throughput out of fixture range: %v", s.ThroughputBPS)
	}
	// Probe-Lab hatte LossRate=0 → adapter liefert nil (default).
	if s.PacketLossRate != nil {
		t.Errorf("expected nil loss rate (fixture has 0), got %v", *s.PacketLossRate)
	}
}

// TestSnapshot_BasicAuth: Adapter setzt Authorization-Header, wenn
// WithBasicAuth gesetzt ist.
func TestSnapshot_BasicAuth(t *testing.T) {
	t.Parallel()
	body := fixtureBody

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != "labuser" || pass != "labpass" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)

	src := mediamtxclient.New(srv.URL, mediamtxclient.WithBasicAuth("labuser", "labpass"), mediamtxclient.WithNow(fixedNow))
	got, err := src.SnapshotConnections(context.Background())
	if err != nil {
		t.Fatalf("SnapshotConnections: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 sample, got %d", len(got))
	}
}

// TestSnapshot_Unauthorized: 401 → ErrSourceUnauthorized.
func TestSnapshot_Unauthorized(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "auth", http.StatusUnauthorized)
	}))
	t.Cleanup(srv.Close)

	src := mediamtxclient.New(srv.URL)
	_, err := src.SnapshotConnections(context.Background())
	if !errors.Is(err, driven.ErrSrtSourceUnauthorized) {
		t.Fatalf("expected ErrSourceUnauthorized, got %v", err)
	}
}

// TestSnapshot_Forbidden: 403 → ErrSourceUnauthorized (analog 401).
func TestSnapshot_Forbidden(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	t.Cleanup(srv.Close)

	src := mediamtxclient.New(srv.URL)
	_, err := src.SnapshotConnections(context.Background())
	if !errors.Is(err, driven.ErrSrtSourceUnauthorized) {
		t.Fatalf("expected ErrSourceUnauthorized, got %v", err)
	}
}

// TestSnapshot_ServerError: 500 → ErrSourceUnavailable.
func TestSnapshot_ServerError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	src := mediamtxclient.New(srv.URL)
	_, err := src.SnapshotConnections(context.Background())
	if !errors.Is(err, driven.ErrSrtSourceUnavailable) {
		t.Fatalf("expected ErrSourceUnavailable, got %v", err)
	}
}

// TestSnapshot_BodyParseError: 200 mit nicht-JSON → ErrSourceParseError.
func TestSnapshot_BodyParseError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("<html>not json</html>"))
	}))
	t.Cleanup(srv.Close)

	src := mediamtxclient.New(srv.URL)
	_, err := src.SnapshotConnections(context.Background())
	if !errors.Is(err, driven.ErrSrtSourceParseError) {
		t.Fatalf("expected ErrSourceParseError, got %v", err)
	}
}

// TestSnapshot_EmptyItems: 200 mit `items: []` → leeres Slice (kein Fehler).
// Der Use Case (Sub-3.5) klassifiziert das als no_active_connection.
func TestSnapshot_EmptyItems(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemCount":0,"pageCount":0,"items":[]}`))
	}))
	t.Cleanup(srv.Close)

	src := mediamtxclient.New(srv.URL)
	got, err := src.SnapshotConnections(context.Background())
	if err != nil {
		t.Fatalf("SnapshotConnections: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected 0 samples, got %d", len(got))
	}
}

// TestSnapshot_ItemUnknownState: ein Item mit unbekanntem `state`
// (oder leerem) wird als ConnectionStateUnknown markiert. Use Case
// (Evaluate) klassifiziert das als partial.
func TestSnapshot_ItemUnknownState(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemCount":1,"pageCount":1,"items":[{"id":"a","path":"x","state":"idle","mbpsLinkCapacity":10}]}`))
	}))
	t.Cleanup(srv.Close)

	src := mediamtxclient.New(srv.URL)
	got, err := src.SnapshotConnections(context.Background())
	if err != nil || len(got) != 1 {
		t.Fatalf("snapshot: items=%d err=%v", len(got), err)
	}
	if got[0].ConnectionState != domain.ConnectionStateUnknown {
		t.Errorf("expected unknown for state='idle', got %s", got[0].ConnectionState)
	}
}

// TestSnapshot_ItemMissingBandwidth: mbpsLinkCapacity=0 ist kein
// gültiger Wert (Loopback-Lab liefert immer >0). Adapter markiert
// ConnectionState als unknown, damit Evaluate `partial` setzt.
func TestSnapshot_ItemMissingBandwidth(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemCount":1,"pageCount":1,"items":[{"id":"a","path":"x","state":"publish","mbpsLinkCapacity":0}]}`))
	}))
	t.Cleanup(srv.Close)

	src := mediamtxclient.New(srv.URL)
	got, err := src.SnapshotConnections(context.Background())
	if err != nil || len(got) != 1 {
		t.Fatalf("snapshot: items=%d err=%v", len(got), err)
	}
	if got[0].ConnectionState != domain.ConnectionStateUnknown {
		t.Errorf("expected unknown for mbpsLinkCapacity=0, got %s", got[0].ConnectionState)
	}
}

// TestSnapshot_ResponseTooLarge: Adapter respektiert
// WithMaxResponseBytes (Defense-in-Depth).
func TestSnapshot_ResponseTooLarge(t *testing.T) {
	t.Parallel()
	// Body 4 KiB, Limit 1 KiB → Truncation, JSON Parse-Error.
	huge := strings.Repeat("a", 4096)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(huge))
	}))
	t.Cleanup(srv.Close)

	src := mediamtxclient.New(srv.URL, mediamtxclient.WithMaxResponseBytes(1024))
	_, err := src.SnapshotConnections(context.Background())
	if !errors.Is(err, driven.ErrSrtSourceParseError) {
		t.Fatalf("expected ErrSourceParseError on truncated body, got %v", err)
	}
}

// TestSnapshot_ContextCancellation: cancelled context → ErrSourceUnavailable
// (HTTP-Client gibt context-Fehler aus, den der Adapter wrappt).
func TestSnapshot_ContextCancellation(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	t.Cleanup(srv.Close)

	src := mediamtxclient.New(srv.URL, mediamtxclient.WithHTTPClient(&http.Client{Timeout: 100 * time.Millisecond}))
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err := src.SnapshotConnections(ctx)
	if !errors.Is(err, driven.ErrSrtSourceUnavailable) {
		t.Fatalf("expected ErrSourceUnavailable on context cancel, got %v", err)
	}
}
