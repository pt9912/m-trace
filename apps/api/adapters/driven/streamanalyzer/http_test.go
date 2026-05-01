package streamanalyzer_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/streamanalyzer"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// TestHTTPStreamAnalyzer_AnalyzeManifest_Master prüft, dass eine 200-
// Antwort des analyzer-service auf das Domain-Result mit PlaylistType=
// master, Findings und EncodedDetails gemappt wird.
func TestHTTPStreamAnalyzer_AnalyzeManifest_Master(t *testing.T) {
	t.Parallel()

	successBody := `{
		"status": "ok",
		"analyzerVersion": "0.3.0",
		"analyzerKind": "hls",
		"input": { "source": "url", "url": "https://example.test/m.m3u8" },
		"playlistType": "master",
		"summary": { "itemCount": 1 },
		"findings": [
			{ "code": "i_frame_variant_skipped", "level": "info", "message": "skipped" }
		],
		"details": { "variants": [], "renditions": [] }
	}`

	var receivedURL, receivedBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedURL = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(successBody))
	}))
	defer server.Close()

	adapter := streamanalyzer.NewHTTPStreamAnalyzer(server.URL)
	result, err := adapter.AnalyzeManifest(context.Background(), domain.StreamAnalysisRequest{
		ManifestURL: "https://example.test/m.m3u8",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if receivedURL != "/analyze" {
		t.Errorf("path: want /analyze, got %q", receivedURL)
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(receivedBody), &parsed); err != nil {
		t.Fatalf("request body not JSON: %v", err)
	}
	if parsed["kind"] != "url" || parsed["url"] != "https://example.test/m.m3u8" {
		t.Errorf("request body wrong: %v", parsed)
	}
	if result.AnalyzerVersion != "0.3.0" {
		t.Errorf("analyzerVersion: want 0.3.0, got %q", result.AnalyzerVersion)
	}
	if result.PlaylistType != domain.PlaylistTypeMaster {
		t.Errorf("playlistType: want master, got %q", result.PlaylistType)
	}
	if len(result.Findings) != 1 || result.Findings[0].Level != domain.FindingLevelInfo {
		t.Errorf("findings unexpected: %+v", result.Findings)
	}
	if len(result.EncodedDetails) == 0 {
		t.Errorf("EncodedDetails empty for master result")
	}
}

func TestHTTPStreamAnalyzer_AnalyzeManifest_TextWithBaseURL(t *testing.T) {
	t.Parallel()

	var receivedBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","analyzerVersion":"0.3.0","analyzerKind":"hls","playlistType":"unknown","details":null,"findings":[]}`))
	}))
	defer server.Close()

	adapter := streamanalyzer.NewHTTPStreamAnalyzer(server.URL + "/")
	_, err := adapter.AnalyzeManifest(context.Background(), domain.StreamAnalysisRequest{
		ManifestText: "#EXTM3U\n",
		BaseURL:      "https://cdn.example.test/",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(receivedBody), &parsed); err != nil {
		t.Fatalf("request body not JSON: %v", err)
	}
	if parsed["kind"] != "text" || parsed["text"] != "#EXTM3U\n" || parsed["baseUrl"] != "https://cdn.example.test/" {
		t.Errorf("text+baseUrl request wrong: %v", parsed)
	}
}

func TestHTTPStreamAnalyzer_AnalyzeManifest_NullDetailsBecomeEmpty(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","analyzerVersion":"0.3.0","analyzerKind":"hls","playlistType":"unknown","details":null,"findings":[]}`))
	}))
	defer server.Close()

	adapter := streamanalyzer.NewHTTPStreamAnalyzer(server.URL)
	result, err := adapter.AnalyzeManifest(context.Background(), domain.StreamAnalysisRequest{
		ManifestText: "#EXTM3U\n",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(result.EncodedDetails) != 0 {
		t.Errorf("EncodedDetails should be empty for details:null, got %d bytes", len(result.EncodedDetails))
	}
}

func TestHTTPStreamAnalyzer_AnalyzeManifest_RejectsConflictingInput(t *testing.T) {
	t.Parallel()

	adapter := streamanalyzer.NewHTTPStreamAnalyzer("http://unused.test")
	_, err := adapter.AnalyzeManifest(context.Background(), domain.StreamAnalysisRequest{
		ManifestText: "#EXTM3U\n",
		ManifestURL:  "https://example.test/m.m3u8",
	})
	if err == nil {
		t.Fatal("expected error when both ManifestText and ManifestURL are set")
	}
}

func TestHTTPStreamAnalyzer_AnalyzeManifest_RejectsEmptyInput(t *testing.T) {
	t.Parallel()

	adapter := streamanalyzer.NewHTTPStreamAnalyzer("http://unused.test")
	_, err := adapter.AnalyzeManifest(context.Background(), domain.StreamAnalysisRequest{})
	if err == nil {
		t.Fatal("expected error when neither ManifestText nor ManifestURL is set")
	}
}

func TestHTTPStreamAnalyzer_AnalyzeManifest_PropagatesNon2xx(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"status":"error","code":"invalid_request","message":"bad input"}`))
	}))
	defer server.Close()

	adapter := streamanalyzer.NewHTTPStreamAnalyzer(server.URL)
	_, err := adapter.AnalyzeManifest(context.Background(), domain.StreamAnalysisRequest{
		ManifestURL: "https://example.test/m.m3u8",
	})
	if err == nil {
		t.Fatal("expected error for 400 response")
	}
	if !strings.Contains(err.Error(), "400") {
		t.Errorf("error should mention status 400, got: %v", err)
	}
}

func TestHTTPStreamAnalyzer_AnalyzeManifest_TransportError(t *testing.T) {
	t.Parallel()

	// Closed server simuliert Verbindungsfehler.
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	server.Close()

	adapter := streamanalyzer.NewHTTPStreamAnalyzer(server.URL)
	_, err := adapter.AnalyzeManifest(context.Background(), domain.StreamAnalysisRequest{
		ManifestURL: "https://example.test/m.m3u8",
	})
	if err == nil {
		t.Fatal("expected transport error")
	}
}

func TestHTTPStreamAnalyzer_AnalyzeManifest_LimitsResponseSize(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// 5 KB Antwort, Limit aber 1 KB → muss abbrechen.
		_, _ = w.Write([]byte(`{"status":"ok","analyzerVersion":"0.3.0","analyzerKind":"hls","playlistType":"unknown","details":"`))
		_, _ = w.Write([]byte(strings.Repeat("x", 5_000)))
		_, _ = w.Write([]byte(`","findings":[]}`))
	}))
	defer server.Close()

	adapter := streamanalyzer.NewHTTPStreamAnalyzer(server.URL, streamanalyzer.WithMaxResponseBytes(1024))
	_, err := adapter.AnalyzeManifest(context.Background(), domain.StreamAnalysisRequest{
		ManifestURL: "https://example.test/m.m3u8",
	})
	if err == nil {
		t.Fatal("expected error when response exceeds size limit")
	}
}

func TestHTTPStreamAnalyzer_AnalyzeManifest_RejectsMalformedJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{not json`))
	}))
	defer server.Close()

	adapter := streamanalyzer.NewHTTPStreamAnalyzer(server.URL)
	_, err := adapter.AnalyzeManifest(context.Background(), domain.StreamAnalysisRequest{
		ManifestURL: "https://example.test/m.m3u8",
	})
	if err == nil {
		t.Fatal("expected JSON-decode error")
	}
}

func TestHTTPStreamAnalyzer_AnalyzeManifest_ReturnsDomainErrorForStatusError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"error","analyzerVersion":"0.3.0","analyzerKind":"hls","code":"manifest_not_hls","message":"not HLS","details":{"firstLine":"<html>"}}`))
	}))
	defer server.Close()

	adapter := streamanalyzer.NewHTTPStreamAnalyzer(server.URL)
	_, err := adapter.AnalyzeManifest(context.Background(), domain.StreamAnalysisRequest{
		ManifestURL: "https://example.test/m.m3u8",
	})
	if err == nil {
		t.Fatal("expected error for 200 + status=error response")
	}
	var domainErr *domain.StreamAnalysisDomainError
	if !errors.As(err, &domainErr) {
		t.Fatalf("expected *StreamAnalysisDomainError, got %T: %v", err, err)
	}
	if domainErr.Code != domain.StreamAnalysisManifestNotHLS {
		t.Errorf("code: want manifest_not_hls, got %q", domainErr.Code)
	}
	if domainErr.Message != "not HLS" {
		t.Errorf("message not propagated: %q", domainErr.Message)
	}
	if got, ok := domainErr.Details["firstLine"].(string); !ok || got != "<html>" {
		t.Errorf("details not propagated: %v", domainErr.Details)
	}
}

func TestHTTPStreamAnalyzer_AnalyzeManifest_DomainErrorPreservesNullDetails(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"error","code":"invalid_input","message":"empty"}`))
	}))
	defer server.Close()

	adapter := streamanalyzer.NewHTTPStreamAnalyzer(server.URL)
	_, err := adapter.AnalyzeManifest(context.Background(), domain.StreamAnalysisRequest{
		ManifestText: "#EXTM3U\n",
	})
	var domainErr *domain.StreamAnalysisDomainError
	if !errors.As(err, &domainErr) {
		t.Fatalf("expected *StreamAnalysisDomainError, got %T: %v", err, err)
	}
	if domainErr.Code != domain.StreamAnalysisInvalidInput {
		t.Errorf("code: want invalid_input, got %q", domainErr.Code)
	}
	if domainErr.Details != nil {
		t.Errorf("details should be nil when omitted, got %v", domainErr.Details)
	}
}

func TestHTTPStreamAnalyzer_WithHTTPClient(t *testing.T) {
	t.Parallel()

	called := 0
	rt := roundTripperFunc(func(_ *http.Request) (*http.Response, error) {
		called++
		body := `{"status":"ok","analyzerVersion":"0.3.0","analyzerKind":"hls","playlistType":"unknown","details":null,"findings":[]}`
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     http.Header{"Content-Type": {"application/json"}},
		}, nil
	})
	adapter := streamanalyzer.NewHTTPStreamAnalyzer("http://injected.test", streamanalyzer.WithHTTPClient(&http.Client{Transport: rt}))
	_, err := adapter.AnalyzeManifest(context.Background(), domain.StreamAnalysisRequest{ManifestText: "#EXTM3U\n"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called != 1 {
		t.Errorf("expected injected RoundTripper to be called once, got %d", called)
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

// TestHTTPStreamAnalyzer_NoopBatch sichert, dass der HTTP-Adapter den
// AnalyzeBatch-Slot auch in 0.3.0 weiterhin no-op bedient — Tranche 6
// fokussiert auf AnalyzeManifest, AnalyzeBatch ist 0.4.0-Arbeit.
func TestHTTPStreamAnalyzer_NoopBatch(t *testing.T) {
	t.Parallel()
	adapter := streamanalyzer.NewHTTPStreamAnalyzer("http://unused.test")
	if err := adapter.AnalyzeBatch(context.Background(), nil); err != nil && !errors.Is(err, nil) {
		t.Errorf("AnalyzeBatch should be no-op, got %v", err)
	}
}
