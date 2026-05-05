// Package mediamtxclient ist der HTTP-Adapter, der den
// driven.SrtSource-Port gegen die MediaMTX-Control-API
// (`GET /v3/srtconns/list`) realisiert (plan-0.6.0 §4 Sub-3.4,
// spec/architecture.md §3.4 / §5.4).
//
// `apps/api` bleibt CGO-frei: der Adapter spricht ausschließlich
// HTTP+JSON, keine libsrt-Bindings (R-2 ist mit Sub-1.3 als
// CGO-frei aufgelöst).
//
// Auth: MediaMTX 1.14+ ist standardmäßig auth-pflichtig. Aufrufer
// (cmd/api) liest Username/Password aus ENV
// (`MTRACE_SRT_SOURCE_USER`, `MTRACE_SRT_SOURCE_PASS`) und übergibt
// sie via WithBasicAuth.
package mediamtxclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

const (
	defaultRequestTimeout = 5 * time.Second
	// defaultMaxResponseBytes schützt den Collector gegen
	// pathologische Antworten der Quelle (Defense-in-Depth). 1 MiB
	// reicht für mehrere Hundert SRT-Verbindungen.
	defaultMaxResponseBytes int64 = 1 * 1024 * 1024

	// path für `GET /v3/srtconns/list` (MediaMTX 1.x; siehe
	// plan-0.6.0 Sub-1.2 Probe-Befund §2.4 / Fixture).
	srtConnsListPath = "/v3/srtconns/list"
)

// Sentinel-Fehler werden in `port/driven/srt_errors.go` gepflegt
// (`driven.ErrSrtSourceUnauthorized` / `…Unavailable` / `…ParseError`).
// Der Adapter wrappt seine HTTP-/Parse-Fehler darauf, damit der Use
// Case ohne Adapter-Import via `errors.Is` klassifizieren kann
// (Hexagon-Boundary).

// HTTPSrtSource implementiert driven.SrtSource gegen MediaMTX.
type HTTPSrtSource struct {
	client               *http.Client
	baseURL              string
	username             string
	password             string
	maxResponseSize      int64
	requiredBandwidthBPS *int64
	now                  func() time.Time
}

// Option justiert den Adapter beim Konstruieren.
type Option func(*HTTPSrtSource)

// WithHTTPClient erlaubt Tests, einen eigenen *http.Client mit
// httptest-Server-Round-Tripper zu injizieren.
func WithHTTPClient(c *http.Client) Option {
	return func(s *HTTPSrtSource) {
		if c != nil {
			s.client = c
		}
	}
}

// WithBasicAuth setzt MediaMTX-`authInternalUsers`-Credentials.
// Leere Werte bedeuten kein `Authorization`-Header — Lab-Default
// für `examples/srt/mediamtx.yml` (any/empty) deckt das ab.
func WithBasicAuth(user, pass string) Option {
	return func(s *HTTPSrtSource) {
		s.username = user
		s.password = pass
	}
}

// WithMaxResponseBytes überschreibt das Defense-in-Depth-Limit.
func WithMaxResponseBytes(n int64) Option {
	return func(s *HTTPSrtSource) {
		if n > 0 {
			s.maxResponseSize = n
		}
	}
}

// WithRequiredBandwidthBPS setzt die erwartete Stream-Bandbreite (in
// bit/s), die der Adapter pro Sample als `RequiredBandwidthBPS`
// emittiert. Ohne diese Schwelle bleibt das Feld nil und Health-
// Bewertung wertet die Bandbreite nicht (spec/telemetry-model.md
// §7.4: „Ohne Schwelle darf available_bandwidth_bps angezeigt, aber
// nicht als Engpass bewertet werden"). Werte ≤ 0 bleiben no-op.
func WithRequiredBandwidthBPS(bps int64) Option {
	return func(s *HTTPSrtSource) {
		if bps > 0 {
			value := bps
			s.requiredBandwidthBPS = &value
		}
	}
}

// WithNow erlaubt Tests, einen festen `time.Now`-Provider zu
// injizieren — der Adapter setzt `CollectedAt` zum Polling-
// Zeitpunkt und braucht für deterministische Tests einen
// kontrollierbaren Clock.
func WithNow(now func() time.Time) Option {
	return func(s *HTTPSrtSource) {
		if now != nil {
			s.now = now
		}
	}
}

// New erzeugt einen Adapter gegen `baseURL` (z. B.
// `http://localhost:9998`).
func New(baseURL string, opts ...Option) *HTTPSrtSource {
	s := &HTTPSrtSource{
		client:          &http.Client{Timeout: defaultRequestTimeout},
		baseURL:         strings.TrimRight(baseURL, "/"),
		maxResponseSize: defaultMaxResponseBytes,
		now:             time.Now,
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

// SnapshotConnections liest `/v3/srtconns/list` und mappt jeden
// Eintrag auf einen `domain.SrtConnectionSample`. Fehlt ein
// Pflichtfeld in einem Item, wird der ConnectionState als
// `unknown` markiert — die Health-Bewertung in der Application-
// Schicht klassifiziert das später als `partial`.
func (s *HTTPSrtSource) SnapshotConnections(ctx context.Context) ([]domain.SrtConnectionSample, error) {
	body, err := s.fetchSrtConnsList(ctx)
	if err != nil {
		return nil, err
	}

	var resp srtConnsListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("%w: %v", driven.ErrSrtSourceParseError, err)
	}

	collectedAt := s.now().UTC()
	out := make([]domain.SrtConnectionSample, 0, len(resp.Items))
	for _, it := range resp.Items {
		out = append(out, mapItem(it, collectedAt, s.requiredBandwidthBPS))
	}
	return out, nil
}

func (s *HTTPSrtSource) fetchSrtConnsList(ctx context.Context) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.baseURL+srtConnsListPath, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: build request: %v", driven.ErrSrtSourceUnavailable, err)
	}
	if s.username != "" || s.password != "" {
		req.SetBasicAuth(s.username, s.password)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: do: %v", driven.ErrSrtSourceUnavailable, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, s.maxResponseSize))
	if err != nil {
		return nil, fmt.Errorf("%w: read body: %v", driven.ErrSrtSourceUnavailable, err)
	}

	switch {
	case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
		return nil, fmt.Errorf("%w: status=%d", driven.ErrSrtSourceUnauthorized, resp.StatusCode)
	case resp.StatusCode >= http.StatusBadRequest:
		return nil, fmt.Errorf("%w: status=%d body=%s", driven.ErrSrtSourceUnavailable, resp.StatusCode, truncate(body, 200))
	}

	return body, nil
}

func truncate(b []byte, limit int) string {
	if len(b) <= limit {
		return string(b)
	}
	return string(b[:limit]) + "…"
}
