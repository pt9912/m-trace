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

func TestAnalyzeHandler_Success_URL(t *testing.T) {
	t.Parallel()
	stub := &stubAnalysisInbound{
		result: domain.StreamAnalysisResult{
			AnalyzerVersion: "0.3.0",
			PlaylistType:    domain.PlaylistTypeMaster,
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

func TestAnalyzeHandler_MapsAdapterErrorTo502(t *testing.T) {
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
}
