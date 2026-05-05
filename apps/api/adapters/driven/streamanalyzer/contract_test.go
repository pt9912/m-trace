package streamanalyzer_test

import (
	"context"
	_ "embed"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/streamanalyzer"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// Die testdata/-Dateien sind Kopien aus
// `spec/contract-fixtures/analyzer/`. Quelle bleibt spec/; ein TS-Test
// in `packages/stream-analyzer/tests/contract.test.ts` pinnt, dass
// die Kopien byte-gleich mit der Spec-Quelle sind. Damit sind beide
// Sprachen gegen *dieselbe* Wahrheit getestet (plan-0.3.0 §9
// Tranche 7.5/4).

//go:embed testdata/contract-success-master.json
var contractSuccessMaster []byte

//go:embed testdata/contract-error-fetch-blocked.json
var contractErrorFetchBlocked []byte

// TestHTTPStreamAnalyzer_ContractSuccessFixture verifies that the
// canonical success fixture decodes into the expected
// domain.StreamAnalysisResult fields. Drift in the TS-Producer's
// output that breaks Go-side decoding fails this test on CI.
func TestHTTPStreamAnalyzer_ContractSuccessFixture(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(contractSuccessMaster)
	}))
	defer server.Close()

	adapter := streamanalyzer.NewHTTPStreamAnalyzer(server.URL)
	result, err := adapter.AnalyzeManifest(context.Background(), domain.StreamAnalysisRequest{
		ManifestText: "#EXTM3U\n",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if result.AnalyzerVersion != "0.5.0" {
		t.Errorf("AnalyzerVersion: want 0.5.0, got %q", result.AnalyzerVersion)
	}
	if result.PlaylistType != domain.PlaylistTypeMaster {
		t.Errorf("PlaylistType: want master, got %q", result.PlaylistType)
	}
	if result.Summary.ItemCount != 2 {
		t.Errorf("Summary.ItemCount: want 2 (1 variant + 1 rendition), got %d", result.Summary.ItemCount)
	}
	if len(result.Findings) != 0 {
		t.Errorf("Findings: want empty, got %v", result.Findings)
	}
	if len(result.EncodedDetails) == 0 {
		t.Errorf("EncodedDetails: want non-empty for master result")
	}
}

// TestHTTPStreamAnalyzer_ContractErrorFixture verifies that the
// canonical error fixture decodes into the typed
// domain.StreamAnalysisDomainError with the expected code and the
// passed-through structured details.
func TestHTTPStreamAnalyzer_ContractErrorFixture(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(contractErrorFetchBlocked)
	}))
	defer server.Close()

	adapter := streamanalyzer.NewHTTPStreamAnalyzer(server.URL)
	_, err := adapter.AnalyzeManifest(context.Background(), domain.StreamAnalysisRequest{
		ManifestURL: "https://example.test/m.m3u8",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	domainErr, ok := err.(*domain.StreamAnalysisDomainError)
	if !ok {
		t.Fatalf("expected *StreamAnalysisDomainError, got %T: %v", err, err)
	}
	if domainErr.Code != domain.StreamAnalysisFetchBlocked {
		t.Errorf("code: want fetch_blocked, got %q", domainErr.Code)
	}
	if domainErr.Message == "" {
		t.Errorf("message should be non-empty")
	}
	if got, _ := domainErr.Details["host"].(string); got != "internal.example.test" {
		t.Errorf("details.host: want %q, got %q", "internal.example.test", got)
	}
	if got, _ := domainErr.Details["address"].(string); got != "10.0.0.5" {
		t.Errorf("details.address: want 10.0.0.5, got %q", got)
	}
	if family, _ := domainErr.Details["family"].(float64); family != 4 {
		t.Errorf("details.family: want 4, got %v", domainErr.Details["family"])
	}
}
