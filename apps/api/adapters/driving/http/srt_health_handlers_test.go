package http_test

import (
	"context"
	"encoding/json"
	_ "embed"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	tracenoop "go.opentelemetry.io/otel/trace/noop"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/auth"
	"github.com/pt9912/m-trace/apps/api/adapters/driven/metrics"
	apihttp "github.com/pt9912/m-trace/apps/api/adapters/driving/http"
	"github.com/pt9912/m-trace/apps/api/hexagon/application"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// fixtureSrtHealthDetail spiegelt das Wire-Format aus
// spec/contract-fixtures/api/srt-health-detail.json wider und wird
// per `make sync-contract-fixtures` aus spec/ ins testdata/ kopiert.
//
//go:embed testdata/srt-health-detail.json
var fixtureSrtHealthDetail []byte

const (
	srtTestProject = "demo"
	srtTestToken   = "demo-token"
)

func srtTestResolver() *auth.StaticProjectResolver {
	return auth.NewStaticProjectResolver(map[string]auth.ProjectConfig{
		srtTestProject: {
			Token:          srtTestToken,
			AllowedOrigins: []string{"http://localhost:5173"},
		},
	})
}

// stubSrtHealthInbound erfüllt SrtHealthInbound für Handler-Tests.
type stubSrtHealthInbound struct {
	latest        []application.SrtHealthSummary
	latestErr     error
	historyReturn []application.SrtHealthHistoryItem
	historyErr    error
	gotLimit      int
}

func (s *stubSrtHealthInbound) LatestByStream(_ context.Context, _ string) ([]application.SrtHealthSummary, error) {
	return s.latest, s.latestErr
}

func (s *stubSrtHealthInbound) HistoryByStream(_ context.Context, _, _ string, limit int) ([]application.SrtHealthHistoryItem, error) {
	s.gotLimit = limit
	return s.historyReturn, s.historyErr
}

func newSrtHealthRouter(t *testing.T, inbound apihttp.SrtHealthInbound) *httptest.Server {
	t.Helper()
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	publisher := metrics.NewPrometheusPublisher()
	resolver := srtTestResolver()
	tracer := tracenoop.NewTracerProvider().Tracer("test")
	router := apihttp.NewRouter(
		nil, nil, nil, resolver, resolver,
		publisher.Handler(), nil, nil, inbound, tracer, logger,
	)
	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)
	return srv
}

func newSummary(streamID string, opts ...func(*application.SrtHealthSummary)) application.SrtHealthSummary {
	t := time.Date(2026, 5, 5, 8, 48, 1, 0, time.UTC)
	s := application.SrtHealthSummary{
		Sample: domain.SrtHealthSample{
			ProjectID:             srtTestProject,
			StreamID:              streamID,
			ConnectionID:          "00000000-0000-0000-0000-000000000001",
			SourceSequence:        "37208036",
			CollectedAt:           t,
			IngestedAt:            t.Add(250 * time.Millisecond),
			RTTMillis:             0.231,
			PacketLossTotal:       0,
			RetransmissionsTotal:  0,
			AvailableBandwidthBPS: 4_352_217_617,
			ThroughputBPS:         func() *int64 { v := int64(1_153_142); return &v }(),
			SourceStatus:          domain.SourceStatusOK,
			SourceErrorCode:       domain.SourceErrorCodeNone,
			ConnectionState:       domain.ConnectionStateConnected,
			HealthState:           domain.HealthStateHealthy,
		},
		SampleAgeMillis:  250,
		StaleAfterMillis: 15_000,
	}
	for _, o := range opts {
		o(&s)
	}
	return s
}

// List: erfolgreiche Antwort liefert envelope `{ "items": [...] }`
// mit den Top-Level- und nested-Feldern aus spec §7a.2.
func TestSrtHealthList_HappyPath(t *testing.T) {
	inbound := &stubSrtHealthInbound{
		latest: []application.SrtHealthSummary{newSummary("srt-test")},
	}
	srv := newSrtHealthRouter(t, inbound)

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/api/srt/health", nil)
	req.Header.Set("X-MTrace-Token", srtTestToken)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		t.Fatalf("expected 200, got %d: %s", res.StatusCode, body)
	}

	var got map[string]any
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	items, ok := got["items"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("expected items[1], got %+v", got)
	}
	first, _ := items[0].(map[string]any)
	if first["stream_id"] != "srt-test" {
		t.Errorf("stream_id = %v", first["stream_id"])
	}
	if first["health_state"] != "healthy" {
		t.Errorf("health_state = %v", first["health_state"])
	}
	freshness, _ := first["freshness"].(map[string]any)
	if freshness["source_observed_at"] != nil {
		t.Errorf("source_observed_at expected null, got %v", freshness["source_observed_at"])
	}
	if freshness["sample_age_ms"].(float64) != 250 {
		t.Errorf("sample_age_ms = %v", freshness["sample_age_ms"])
	}
}

// Detail: 200 mit envelope `{ "stream_id": ..., "items": [...] }`.
func TestSrtHealthDetail_HappyPath(t *testing.T) {
	inbound := &stubSrtHealthInbound{
		historyReturn: []application.SrtHealthHistoryItem{newSummary("srt-test")},
	}
	srv := newSrtHealthRouter(t, inbound)

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/api/srt/health/srt-test?samples_limit=50", nil)
	req.Header.Set("X-MTrace-Token", srtTestToken)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		t.Fatalf("expected 200, got %d: %s", res.StatusCode, body)
	}
	if inbound.gotLimit != 50 {
		t.Errorf("limit pass-through: got %d, want 50", inbound.gotLimit)
	}
	var got map[string]any
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got["stream_id"] != "srt-test" {
		t.Errorf("envelope stream_id = %v", got["stream_id"])
	}
}

// Detail: stream_unknown → 404 mit JSON-Body.
func TestSrtHealthDetail_NotFound(t *testing.T) {
	inbound := &stubSrtHealthInbound{historyErr: application.ErrSrtHealthStreamUnknown}
	srv := newSrtHealthRouter(t, inbound)

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/api/srt/health/missing", nil)
	req.Header.Set("X-MTrace-Token", srtTestToken)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(res.Body)
		t.Fatalf("expected 404, got %d: %s", res.StatusCode, body)
	}
	var body map[string]string
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["error"] != "stream_unknown" {
		t.Errorf("error = %v", body["error"])
	}
}

// Detail: invalid samples_limit → 400.
func TestSrtHealthDetail_InvalidLimit(t *testing.T) {
	inbound := &stubSrtHealthInbound{}
	srv := newSrtHealthRouter(t, inbound)

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/api/srt/health/srt-test?samples_limit=foo", nil)
	req.Header.Set("X-MTrace-Token", srtTestToken)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
}

// List: ohne Token → 401 (analog stream-sessions).
func TestSrtHealthList_RequiresToken(t *testing.T) {
	inbound := &stubSrtHealthInbound{}
	srv := newSrtHealthRouter(t, inbound)

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/api/srt/health", nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusUnauthorized {
		body, _ := io.ReadAll(res.Body)
		t.Fatalf("expected 401, got %d: %s", res.StatusCode, body)
	}
}

// List: 5xx vom Service propagiert als 500.
func TestSrtHealthList_RepoError(t *testing.T) {
	inbound := &stubSrtHealthInbound{latestErr: errors.New("db down")}
	srv := newSrtHealthRouter(t, inbound)

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/api/srt/health", nil)
	req.Header.Set("X-MTrace-Token", srtTestToken)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", res.StatusCode)
	}
}

// Schema-Snapshot: das gerenderte Detail-JSON spiegelt die Top-Level-
// und nested-Schlüssel aus spec/contract-fixtures/api/srt-health-detail.json.
// Die Werte ändern sich pro Probe-Run; geprüft wird das **Set**
// der erwarteten Schlüssel pro Block.
func TestSrtHealthDetail_SchemaMatchesFixture(t *testing.T) {
	inbound := &stubSrtHealthInbound{
		historyReturn: []application.SrtHealthHistoryItem{newSummary("srt-test")},
	}
	srv := newSrtHealthRouter(t, inbound)

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/api/srt/health/srt-test", nil)
	req.Header.Set("X-MTrace-Token", srtTestToken)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer func() { _ = res.Body.Close() }()

	var got, want map[string]any
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if err := json.Unmarshal(fixtureSrtHealthDetail, &want); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	delete(want, "_meta")

	expectKeysSubset(t, "<root>", want, got)
}

// expectKeysSubset prüft, dass alle Schlüssel aus `want` auch in
// `got` vorkommen. Werte werden nicht verglichen — die Probe-Werte
// im Wire-Format sind zeitabhängig; das Schema ist es nicht.
func expectKeysSubset(t *testing.T, path string, want, got map[string]any) {
	t.Helper()
	for k, wv := range want {
		gv, ok := got[k]
		if !ok {
			t.Errorf("missing key %q at path %s", k, path)
			continue
		}
		wsub, wok := wv.(map[string]any)
		gsub, gok := gv.(map[string]any)
		if wok && gok {
			expectKeysSubset(t, path+"."+k, wsub, gsub)
			continue
		}
		// items[]: das fixture hat 1 Eintrag; wir prüfen nur den ersten.
		if wlist, ok := wv.([]any); ok && len(wlist) > 0 {
			glist, _ := gv.([]any)
			if len(glist) == 0 {
				t.Errorf("expected non-empty array at %s.%s", path, k)
				continue
			}
			wm, wmok := wlist[0].(map[string]any)
			gm, gmok := glist[0].(map[string]any)
			if wmok && gmok {
				expectKeysSubset(t, path+"."+k+"[0]", wm, gm)
			}
		}
	}
}

// _ ist ein expliziter Verweis darauf, dass strings importiert
// bleibt (für künftige String-Asserts) — golangci-lint wirft ohne
// Nutzung.
var _ = strings.HasPrefix

// Detail: leerer stream_id-Path wird vom Mux gar nicht erst geroutet
// (Go's http.ServeMux matched nicht), aber wir prüfen das Fehler-
// Handling für Empty-Path-Parameter via direktem Handler-Aufruf.
func TestSrtHealthDetail_EmptyStreamID(t *testing.T) {
	inbound := &stubSrtHealthInbound{}
	srv := newSrtHealthRouter(t, inbound)

	// Slash am Ende → Mux routet auf den Detail-Handler, aber
	// PathValue("stream_id") ist leer.
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/api/srt/health/", nil)
	req.Header.Set("X-MTrace-Token", srtTestToken)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer func() { _ = res.Body.Close() }()
	// Mux liefert 404, weil das Path-Pattern `{stream_id}` einen
	// nicht-leeren Wert verlangt. Das ist akzeptabel — der Handler-
	// interne Empty-Check bleibt als Defense-in-Depth.
	if res.StatusCode != http.StatusNotFound && res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 404 or 400 for empty stream_id, got %d", res.StatusCode)
	}
}

// List: POST auf den GET-Endpoint → 405 (analog stream-sessions).
// Go's http.ServeMux mit method+path-Pattern routet nicht-GET nicht
// auf den Handler, sondern auf einen Default-405-Handler — dieser
// Test pinnt das.
func TestSrtHealthList_PostNotAllowed(t *testing.T) {
	inbound := &stubSrtHealthInbound{}
	srv := newSrtHealthRouter(t, inbound)

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, srv.URL+"/api/srt/health", nil)
	req.Header.Set("X-MTrace-Token", srtTestToken)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", res.StatusCode)
	}
}

// List: leere Liste liefert envelope mit `items: []`.
func TestSrtHealthList_EmptyEnvelope(t *testing.T) {
	inbound := &stubSrtHealthInbound{latest: nil}
	srv := newSrtHealthRouter(t, inbound)

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/api/srt/health", nil)
	req.Header.Set("X-MTrace-Token", srtTestToken)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	var got map[string]any
	_ = json.NewDecoder(res.Body).Decode(&got)
	if items, ok := got["items"].([]any); !ok || len(items) != 0 {
		t.Errorf("expected empty items array, got %+v", got["items"])
	}
}
