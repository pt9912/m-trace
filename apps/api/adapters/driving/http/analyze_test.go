package http_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apihttp "github.com/pt9912/m-trace/apps/api/adapters/driving/http"
	"github.com/pt9912/m-trace/apps/api/hexagon/application"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

type stubAnalysisInbound struct {
	called int
	gotReq domain.StreamAnalysisRequest
	result domain.StreamAnalysisResult
	err    error
}

func (s *stubAnalysisInbound) AnalyzeManifest(_ context.Context, req domain.StreamAnalysisRequest) (domain.StreamAnalysisResult, error) {
	s.called++
	s.gotReq = req
	return s.result, s.err
}

func newAnalyzeHandler(stub *stubAnalysisInbound) http.Handler {
	return &apihttp.AnalyzeHandler{UseCase: stub}
}

type recordedOutcome struct {
	outcome string
	code    string
}

type stubAnalyzeMetrics struct {
	recorded []recordedOutcome
}

func (s *stubAnalyzeMetrics) AnalyzeRequest(outcome, code string) {
	s.recorded = append(s.recorded, recordedOutcome{outcome: outcome, code: code})
}

func newAnalyzeHandlerWithMetrics(stub *stubAnalysisInbound, m *stubAnalyzeMetrics) http.Handler {
	return &apihttp.AnalyzeHandler{UseCase: stub, Metrics: m}
}

func TestAnalyzeHandler_Success_URL(t *testing.T) {
	t.Parallel()
	stub := &stubAnalysisInbound{
		result: domain.StreamAnalysisResult{
			AnalyzerVersion: "0.3.0",
			PlaylistType:    domain.PlaylistTypeMaster,
			Summary:         domain.StreamAnalysisSummary{ItemCount: 3},
			Findings: []domain.StreamAnalysisFinding{
				{Code: "playlist_ambiguous", Level: domain.FindingLevelWarning, Message: "ambiguous"},
			},
			EncodedDetails: []byte(`{"variants":[]}`),
		},
	}
	server := httptest.NewServer(newAnalyzeHandler(stub))
	defer server.Close()

	res, err := http.Post(server.URL, "application/json", strings.NewReader(`{"kind":"url","url":"https://example.test/m.m3u8"}`))
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status: want 200, got %d", res.StatusCode)
	}
	body, _ := io.ReadAll(res.Body)
	got := string(body)
	if !strings.Contains(got, `"playlistType":"master"`) || !strings.Contains(got, `"variants":[]`) {
		t.Errorf("response missing fields: %s", got)
	}
	if !strings.Contains(got, `"summary":{"itemCount":3}`) {
		t.Errorf("response missing summary.itemCount=3: %s", got)
	}
	if stub.gotReq.ManifestURL != "https://example.test/m.m3u8" {
		t.Errorf("URL not propagated: %+v", stub.gotReq)
	}
}

func TestAnalyzeHandler_Success_TextWithBaseURL(t *testing.T) {
	t.Parallel()
	stub := &stubAnalysisInbound{
		result: domain.StreamAnalysisResult{
			AnalyzerVersion: "0.3.0",
			PlaylistType:    domain.PlaylistTypeMedia,
		},
	}
	server := httptest.NewServer(newAnalyzeHandler(stub))
	defer server.Close()

	body := `{"kind":"text","text":"#EXTM3U\n","baseUrl":"https://cdn.test/"}`
	res, err := http.Post(server.URL, "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status: want 200, got %d", res.StatusCode)
	}
	if stub.gotReq.ManifestText != "#EXTM3U\n" || stub.gotReq.BaseURL != "https://cdn.test/" {
		t.Errorf("text/baseUrl not propagated: %+v", stub.gotReq)
	}
}

func TestAnalyzeHandler_NullDetailsAreEmittedAsNull(t *testing.T) {
	t.Parallel()
	stub := &stubAnalysisInbound{
		result: domain.StreamAnalysisResult{
			AnalyzerVersion: "0.3.0",
			PlaylistType:    domain.PlaylistTypeUnknown,
		},
	}
	server := httptest.NewServer(newAnalyzeHandler(stub))
	defer server.Close()

	res, err := http.Post(server.URL, "application/json", strings.NewReader(`{"kind":"text","text":"#EXTM3U\n"}`))
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if !strings.Contains(string(body), `"details":null`) {
		t.Errorf("details should be null, got %s", body)
	}
}

func TestAnalyzeHandler_RejectsNonJSONContentType(t *testing.T) {
	t.Parallel()
	stub := &stubAnalysisInbound{}
	server := httptest.NewServer(newAnalyzeHandler(stub))
	defer server.Close()

	res, err := http.Post(server.URL, "text/plain", strings.NewReader("{}"))
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusUnsupportedMediaType {
		t.Errorf("status: want 415, got %d", res.StatusCode)
	}
}

func TestAnalyzeHandler_RejectsMalformedJSON(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(newAnalyzeHandler(&stubAnalysisInbound{}))
	defer server.Close()

	res, err := http.Post(server.URL, "application/json", strings.NewReader(`{not json`))
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("status: want 400, got %d", res.StatusCode)
	}
}

func TestAnalyzeHandler_RejectsInvalidRequestShapes(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		body string
	}{
		{"missing kind", `{}`},
		{"empty url", `{"kind":"url","url":""}`},
		{"empty text", `{"kind":"text","text":""}`},
		{"unknown kind", `{"kind":"binary"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(newAnalyzeHandler(&stubAnalysisInbound{}))
			defer server.Close()
			res, err := http.Post(server.URL, "application/json", strings.NewReader(tc.body))
			if err != nil {
				t.Fatalf("post failed: %v", err)
			}
			defer res.Body.Close()
			if res.StatusCode != http.StatusBadRequest {
				t.Errorf("status: want 400, got %d", res.StatusCode)
			}
			body, _ := io.ReadAll(res.Body)
			if !strings.Contains(string(body), `"code":"invalid_request"`) {
				t.Errorf("body missing invalid_request code: %s", body)
			}
		})
	}
}

func TestAnalyzeHandler_RejectsOversizedBody(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(newAnalyzeHandler(&stubAnalysisInbound{}))
	defer server.Close()

	big := `{"kind":"text","text":"` + strings.Repeat("x", 2*1024*1024) + `"}`
	res, err := http.Post(server.URL, "application/json", strings.NewReader(big))
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("status: want 413, got %d", res.StatusCode)
	}
}

func TestAnalyzeHandler_MapsErrEmptyTo400(t *testing.T) {
	t.Parallel()
	stub := &stubAnalysisInbound{err: application.ErrAnalyzeManifestEmpty}
	server := httptest.NewServer(newAnalyzeHandler(stub))
	defer server.Close()

	res, err := http.Post(server.URL, "application/json", strings.NewReader(`{"kind":"text","text":"#EXTM3U\n"}`))
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("status: want 400, got %d", res.StatusCode)
	}
}

func TestAnalyzeHandler_MapsTransportErrorTo502(t *testing.T) {
	t.Parallel()
	stub := &stubAnalysisInbound{err: errors.New("analyzer-service unreachable")}
	server := httptest.NewServer(newAnalyzeHandler(stub))
	defer server.Close()

	res, err := http.Post(server.URL, "application/json", strings.NewReader(`{"kind":"url","url":"https://example.test/m.m3u8"}`))
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadGateway {
		t.Errorf("status: want 502, got %d", res.StatusCode)
	}
	body, _ := io.ReadAll(res.Body)
	if !strings.Contains(string(body), `"code":"analyzer_unavailable"`) {
		t.Errorf("body missing analyzer_unavailable code: %s", body)
	}
	// Defense-in-Depth: die rohe Adapter-Fehler-Message darf nicht in
	// den Antwort-Body durchsickern (Info-Leak).
	if strings.Contains(string(body), "analyzer-service unreachable") {
		t.Errorf("body leaks adapter error message: %s", body)
	}
}

func TestAnalyzeHandler_MapsDomainErrorsByCode(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name         string
		code         domain.StreamAnalysisErrorCode
		wantStatus   int
		wantCodeBody string
	}{
		{"invalid_input → 400", domain.StreamAnalysisInvalidInput, http.StatusBadRequest, "invalid_input"},
		{"fetch_blocked → 400", domain.StreamAnalysisFetchBlocked, http.StatusBadRequest, "fetch_blocked"},
		{"manifest_not_hls → 422", domain.StreamAnalysisManifestNotHLS, http.StatusUnprocessableEntity, "manifest_not_hls"},
		{"fetch_failed → 502", domain.StreamAnalysisFetchFailed, http.StatusBadGateway, "fetch_failed"},
		{"manifest_too_large → 502", domain.StreamAnalysisManifestTooLarge, http.StatusBadGateway, "manifest_too_large"},
		{"internal_error → 502", domain.StreamAnalysisInternalError, http.StatusBadGateway, "internal_error"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			stub := &stubAnalysisInbound{err: &domain.StreamAnalysisDomainError{
				Code:    tc.code,
				Message: "expected message",
				Details: map[string]any{"hint": "structured detail"},
			}}
			server := httptest.NewServer(newAnalyzeHandler(stub))
			defer server.Close()

			res, err := http.Post(server.URL, "application/json", strings.NewReader(`{"kind":"url","url":"https://example.test/m.m3u8"}`))
			if err != nil {
				t.Fatalf("post failed: %v", err)
			}
			defer res.Body.Close()
			if res.StatusCode != tc.wantStatus {
				t.Errorf("status: want %d, got %d", tc.wantStatus, res.StatusCode)
			}
			body, _ := io.ReadAll(res.Body)
			if !strings.Contains(string(body), `"code":"`+tc.wantCodeBody+`"`) {
				t.Errorf("body missing code %q: %s", tc.wantCodeBody, body)
			}
			if !strings.Contains(string(body), `"hint":"structured detail"`) {
				t.Errorf("body should pass through structured details: %s", body)
			}
		})
	}
}

func TestAnalyzeHandler_RecordsOutcomeMetrics(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		stubResult  domain.StreamAnalysisResult
		stubErr     error
		body        string
		contentType string
		wantOutcome string
		wantCode    string
	}{
		{
			name:        "success → outcome=ok, code=ok",
			stubResult:  domain.StreamAnalysisResult{AnalyzerVersion: "0.3.0", PlaylistType: domain.PlaylistTypeMaster},
			body:        `{"kind":"text","text":"#EXTM3U\n"}`,
			contentType: "application/json",
			wantOutcome: "ok",
			wantCode:    "ok",
		},
		{
			name:        "domain error → outcome=error, code=manifest_not_hls",
			stubErr:     &domain.StreamAnalysisDomainError{Code: domain.StreamAnalysisManifestNotHLS, Message: "x"},
			body:        `{"kind":"text","text":"<html>"}`,
			contentType: "application/json",
			wantOutcome: "error",
			wantCode:    "manifest_not_hls",
		},
		{
			name:        "transport error → outcome=error, code=analyzer_unavailable",
			stubErr:     errors.New("dial tcp: i/o timeout"),
			body:        `{"kind":"url","url":"https://example.test/m.m3u8"}`,
			contentType: "application/json",
			wantOutcome: "error",
			wantCode:    "analyzer_unavailable",
		},
		{
			name:        "invalid_request → outcome=error, code=invalid_request",
			body:        `{}`,
			contentType: "application/json",
			wantOutcome: "error",
			wantCode:    "invalid_request",
		},
		{
			name:        "unsupported_media_type → outcome=error, code=unsupported_media_type",
			body:        `{}`,
			contentType: "text/plain",
			wantOutcome: "error",
			wantCode:    "unsupported_media_type",
		},
		{
			name:        "invalid_json → outcome=error, code=invalid_json",
			body:        `{not json`,
			contentType: "application/json",
			wantOutcome: "error",
			wantCode:    "invalid_json",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			stub := &stubAnalysisInbound{result: tc.stubResult, err: tc.stubErr}
			metrics := &stubAnalyzeMetrics{}
			server := httptest.NewServer(newAnalyzeHandlerWithMetrics(stub, metrics))
			defer server.Close()

			res, err := http.Post(server.URL, tc.contentType, strings.NewReader(tc.body))
			if err != nil {
				t.Fatalf("post failed: %v", err)
			}
			res.Body.Close()

			if len(metrics.recorded) != 1 {
				t.Fatalf("expected exactly one metric record, got %d: %+v", len(metrics.recorded), metrics.recorded)
			}
			got := metrics.recorded[0]
			if got.outcome != tc.wantOutcome || got.code != tc.wantCode {
				t.Errorf("metric outcome/code: want (%q,%q), got (%q,%q)", tc.wantOutcome, tc.wantCode, got.outcome, got.code)
			}
		})
	}
}

func TestAnalyzeHandler_RecordsNothingWhenMetricsNil(t *testing.T) {
	t.Parallel()
	stub := &stubAnalysisInbound{result: domain.StreamAnalysisResult{AnalyzerVersion: "0.3.0", PlaylistType: domain.PlaylistTypeUnknown}}
	server := httptest.NewServer(newAnalyzeHandler(stub))
	defer server.Close()
	res, err := http.Post(server.URL, "application/json", strings.NewReader(`{"kind":"text","text":"#EXTM3U\n"}`))
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	res.Body.Close()
	// Kein Panic — der Handler verträgt nil-Metrics.
}

func TestAnalyzeHandler_DomainErrorWithoutDetailsOmitsDetails(t *testing.T) {
	t.Parallel()
	stub := &stubAnalysisInbound{err: &domain.StreamAnalysisDomainError{
		Code:    domain.StreamAnalysisManifestNotHLS,
		Message: "not HLS",
	}}
	server := httptest.NewServer(newAnalyzeHandler(stub))
	defer server.Close()

	res, err := http.Post(server.URL, "application/json", strings.NewReader(`{"kind":"text","text":"<html>"}`))
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if strings.Contains(string(body), `"details"`) {
		t.Errorf("body should omit details key when domain error has none: %s", body)
	}
}
