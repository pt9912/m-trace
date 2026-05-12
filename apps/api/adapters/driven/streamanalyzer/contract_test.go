package streamanalyzer_test

import (
	"context"
	_ "embed"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

//go:embed testdata/contract-success-dash-vod.json
var contractSuccessDashVod []byte

//go:embed testdata/contract-success-dash-live.json
var contractSuccessDashLive []byte

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

	if result.AnalyzerVersion != "0.12.6" {
		t.Errorf("AnalyzerVersion: want 0.12.0, got %q", result.AnalyzerVersion)
	}
	if result.AnalyzerKind != domain.AnalyzerKindHLS {
		t.Errorf("AnalyzerKind: want hls, got %q", result.AnalyzerKind)
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

// TestHTTPStreamAnalyzer_ContractDashVodFixture verifiziert, dass die
// VOD-MPD-Fixture (plan-0.9.0 Tranche 3, RAK-58) byte-gleich vom
// Adapter geparsed wird: AnalyzerKind = dash, PlaylistType = dash,
// itemCount = 3 (2 video + 1 audio Representation).
func TestHTTPStreamAnalyzer_ContractDashVodFixture(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(contractSuccessDashVod)
	}))
	defer server.Close()

	adapter := streamanalyzer.NewHTTPStreamAnalyzer(server.URL)
	result, err := adapter.AnalyzeManifest(context.Background(), domain.StreamAnalysisRequest{
		ManifestText: `<?xml version="1.0"?><MPD type="static"></MPD>`,
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if result.AnalyzerKind != domain.AnalyzerKindDASH {
		t.Errorf("AnalyzerKind: want dash, got %q", result.AnalyzerKind)
	}
	if result.PlaylistType != domain.PlaylistTypeDash {
		t.Errorf("PlaylistType: want dash, got %q", result.PlaylistType)
	}
	if result.Summary.ItemCount != 3 {
		t.Errorf("Summary.ItemCount: want 3 (2 video + 1 audio rep), got %d", result.Summary.ItemCount)
	}
	if len(result.EncodedDetails) == 0 {
		t.Errorf("EncodedDetails: want non-empty for DASH result")
	}
}

// TestHTTPStreamAnalyzer_ContractDashLiveFixture verifiziert, dass
// die Live-MPD-Variante (`type=dynamic`, `live=true`,
// `minimumUpdatePeriod`/`availabilityStartTime`) ebenfalls als DASH
// klassifiziert und mit derselben PlaylistType-Konstante geliefert wird.
func TestHTTPStreamAnalyzer_ContractDashLiveFixture(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(contractSuccessDashLive)
	}))
	defer server.Close()

	adapter := streamanalyzer.NewHTTPStreamAnalyzer(server.URL)
	result, err := adapter.AnalyzeManifest(context.Background(), domain.StreamAnalysisRequest{
		ManifestText: `<?xml version="1.0"?><MPD type="dynamic"></MPD>`,
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if result.AnalyzerKind != domain.AnalyzerKindDASH {
		t.Errorf("AnalyzerKind: want dash, got %q", result.AnalyzerKind)
	}
	if result.PlaylistType != domain.PlaylistTypeDash {
		t.Errorf("PlaylistType: want dash, got %q", result.PlaylistType)
	}
	if result.Summary.ItemCount != 1 {
		t.Errorf("Summary.ItemCount: want 1 (single video rep), got %d", result.Summary.ItemCount)
	}
}

// TestHTTPStreamAnalyzer_ContractDashVodCMAFBinarySkipped verifiziert
// (plan-0.10.0 Tranche 5, NF-13 / RAK-63 / RAK-64), dass das additive
// `details.cmaf.binary`-Schema aus T4 unverändert über den Driven-
// Adapter durchgereicht wird. Der Test JSON-decodiert das
// EncodedDetails-Feld des Domain-Results und pinnt den Skipped-Status
// plus den Failure-Code `segment_reference_missing`, weil die VOD-
// Fixture nur MP4-MIME-Signale ohne Init-/Media-Referenzen trägt.
func TestHTTPStreamAnalyzer_ContractDashVodCMAFBinarySkipped(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(contractSuccessDashVod)
	}))
	defer server.Close()

	adapter := streamanalyzer.NewHTTPStreamAnalyzer(server.URL)
	result, err := adapter.AnalyzeManifest(context.Background(), domain.StreamAnalysisRequest{
		ManifestText: `<?xml version="1.0"?><MPD type="static"></MPD>`,
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	var details struct {
		CMAF struct {
			Source     string `json:"source"`
			Confidence string `json:"confidence"`
			Signals    []struct {
				Code string `json:"code"`
			} `json:"signals"`
			Binary struct {
				Status          string `json:"status"`
				SegmentsChecked []struct {
					Kind        string `json:"kind"`
					Status      string `json:"status"`
					FailureCode string `json:"failureCode"`
				} `json:"segmentsChecked"`
				Limits struct {
					RequiredSegmentChecks int `json:"requiredSegmentChecks"`
				} `json:"limits"`
			} `json:"binary"`
		} `json:"cmaf"`
	}
	if err := json.Unmarshal(result.EncodedDetails, &details); err != nil {
		t.Fatalf("decode EncodedDetails: %v", err)
	}

	if details.CMAF.Source != "dash" {
		t.Errorf("cmaf.source: want dash, got %q", details.CMAF.Source)
	}
	if details.CMAF.Confidence != "inferred" {
		t.Errorf("cmaf.confidence: want inferred, got %q", details.CMAF.Confidence)
	}
	if details.CMAF.Binary.Status != "skipped" {
		t.Errorf("cmaf.binary.status: want skipped, got %q", details.CMAF.Binary.Status)
	}
	if details.CMAF.Binary.Limits.RequiredSegmentChecks != 4 {
		t.Errorf("cmaf.binary.limits.requiredSegmentChecks: want 4, got %d", details.CMAF.Binary.Limits.RequiredSegmentChecks)
	}
	if got := len(details.CMAF.Binary.SegmentsChecked); got != 4 {
		t.Errorf("cmaf.binary.segmentsChecked: want 4, got %d", got)
	}
	for i, c := range details.CMAF.Binary.SegmentsChecked {
		if c.Status != "skipped" {
			t.Errorf("segmentsChecked[%d].status: want skipped, got %q", i, c.Status)
		}
		if c.FailureCode != "segment_reference_missing" {
			t.Errorf("segmentsChecked[%d].failureCode: want segment_reference_missing, got %q", i, c.FailureCode)
		}
	}
}

// TestHTTPStreamAnalyzer_CMAFContractFixturesDecode stellt sicher,
// dass alle additiven CMAF-Contract-Fixtures aus plan-0.10.0 über den
// Go-Adapter decodierbar bleiben. Die fachlichen Einzelpfade sind im
// TypeScript-Analyzer gepinnt; hier geht es um den Wire-Vertrag und
// die unveränderte Durchleitung von `details.cmaf` in EncodedDetails.
func TestHTTPStreamAnalyzer_CMAFContractFixturesDecode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		wantCMAF bool
	}{
		{"contract-success-hls-cmaf-vod.json", true},
		{"contract-success-hls-ts-negative.json", false},
		{"contract-success-hls-master-codecs-only.json", false},
		{"contract-success-hls-map-byterange.json", true},
		{"contract-success-hls-media-byterange.json", true},
		{"contract-success-dash-mp4-mime-only.json", true},
		{"contract-success-dash-cmaf-vod.json", true},
		{"contract-success-dash-no-cmaf-signals.json", false},
		{"contract-success-dash-baseurl-inheritance.json", true},
		{"contract-success-dash-segmentlist.json", true},
		{"contract-error-cmaf-binary-validation.json", true},
		{"contract-error-cmaf-invalid-box-structure.json", true},
		{"contract-success-cmaf-skipped-too-large.json", true},
		{"contract-success-cmaf-skipped-content-type.json", true},
		{"contract-success-cmaf-skipped-binary-disabled.json", true},
		{"contract-success-cmaf-skipped-not-planned.json", true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			payload, err := os.ReadFile(filepath.Join("testdata", tt.name))
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write(payload)
			}))
			defer server.Close()

			adapter := streamanalyzer.NewHTTPStreamAnalyzer(server.URL)
			result, err := adapter.AnalyzeManifest(context.Background(), domain.StreamAnalysisRequest{
				ManifestText: "#EXTM3U\n",
			})
			if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
			if len(result.EncodedDetails) == 0 {
				t.Fatal("EncodedDetails: want non-empty")
			}

			var details map[string]json.RawMessage
			if err := json.Unmarshal(result.EncodedDetails, &details); err != nil {
				t.Fatalf("decode EncodedDetails: %v", err)
			}
			_, hasCMAF := details["cmaf"]
			if hasCMAF != tt.wantCMAF {
				t.Fatalf("details.cmaf presence: want %v, got %v", tt.wantCMAF, hasCMAF)
			}
		})
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
