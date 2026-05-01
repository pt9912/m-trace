package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/pt9912/m-trace/apps/api/hexagon/application"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

type fakeAnalyzer struct {
	called int
	gotReq domain.StreamAnalysisRequest
	result domain.StreamAnalysisResult
	err    error
}

func (f *fakeAnalyzer) AnalyzeBatch(_ context.Context, _ []domain.PlaybackEvent) error { return nil }
func (f *fakeAnalyzer) AnalyzeManifest(_ context.Context, req domain.StreamAnalysisRequest) (domain.StreamAnalysisResult, error) {
	f.called++
	f.gotReq = req
	return f.result, f.err
}

func TestAnalyzeManifest_PassesRequestThrough(t *testing.T) {
	t.Parallel()
	stub := &fakeAnalyzer{result: domain.StreamAnalysisResult{
		AnalyzerVersion: "0.3.0",
		PlaylistType:    domain.PlaylistTypeMaster,
	}}
	uc := application.NewAnalyzeManifestUseCase(stub)

	got, err := uc.AnalyzeManifest(context.Background(), domain.StreamAnalysisRequest{
		ManifestURL: "https://example.test/m.m3u8",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stub.called != 1 || stub.gotReq.ManifestURL != "https://example.test/m.m3u8" {
		t.Errorf("adapter not called correctly: %+v", stub)
	}
	if got.PlaylistType != domain.PlaylistTypeMaster {
		t.Errorf("result.PlaylistType: want master, got %q", got.PlaylistType)
	}
}

func TestAnalyzeManifest_RejectsEmptyRequest(t *testing.T) {
	t.Parallel()
	stub := &fakeAnalyzer{}
	uc := application.NewAnalyzeManifestUseCase(stub)

	_, err := uc.AnalyzeManifest(context.Background(), domain.StreamAnalysisRequest{})
	if !errors.Is(err, application.ErrAnalyzeManifestEmpty) {
		t.Fatalf("expected ErrAnalyzeManifestEmpty, got %v", err)
	}
	if stub.called != 0 {
		t.Errorf("adapter should not be called when request is empty")
	}
}

func TestAnalyzeManifest_RejectsWhitespaceOnlyRequest(t *testing.T) {
	t.Parallel()
	stub := &fakeAnalyzer{}
	uc := application.NewAnalyzeManifestUseCase(stub)

	_, err := uc.AnalyzeManifest(context.Background(), domain.StreamAnalysisRequest{
		ManifestText: "   \n\t",
		ManifestURL:  "  ",
	})
	if !errors.Is(err, application.ErrAnalyzeManifestEmpty) {
		t.Fatalf("expected ErrAnalyzeManifestEmpty, got %v", err)
	}
}

func TestAnalyzeManifest_PropagatesAdapterError(t *testing.T) {
	t.Parallel()
	want := errors.New("adapter boom")
	stub := &fakeAnalyzer{err: want}
	uc := application.NewAnalyzeManifestUseCase(stub)

	_, err := uc.AnalyzeManifest(context.Background(), domain.StreamAnalysisRequest{
		ManifestText: "#EXTM3U\n",
	})
	if !errors.Is(err, want) {
		t.Fatalf("expected wrapped adapter error, got %v", err)
	}
}
