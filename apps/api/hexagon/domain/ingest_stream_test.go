package domain_test

import (
	"errors"
	"testing"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// `0.11.0` Tranche 1 / RAK-65..RAK-67: Domain-Validierungen ohne
// HTTP, Storage oder MediaMTX.

func TestIngestProtocol_IsKnown(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   domain.IngestProtocol
		want bool
	}{
		{domain.IngestProtocolSRT, true},
		{domain.IngestProtocolRTMP, true},
		{"webrtc", false},
		{"hls", false},
		{"", false},
		{"SRT", false}, // case-sensitive: Validate normalisiert vor IsKnown
	}
	for _, tc := range cases {
		t.Run(string(tc.in), func(t *testing.T) {
			if got := tc.in.IsKnown(); got != tc.want {
				t.Errorf("%q IsKnown() = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestValidateIngestProtocol_NormalizesAndChecksAllowlist(t *testing.T) {
	t.Parallel()
	cases := []struct {
		raw      string
		want     domain.IngestProtocol
		wantErr  error
	}{
		{"srt", domain.IngestProtocolSRT, nil},
		{"  rtmp  ", domain.IngestProtocolRTMP, nil},
		{"SRT", domain.IngestProtocolSRT, nil},
		{"Rtmp", domain.IngestProtocolRTMP, nil},
		{"webrtc", "", domain.ErrIngestProtocolUnknown},
		{"", "", domain.ErrIngestProtocolUnknown},
	}
	for _, tc := range cases {
		t.Run(tc.raw, func(t *testing.T) {
			got, err := domain.ValidateIngestProtocol(tc.raw)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("err: want %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestValidateProjectIDConsistency(t *testing.T) {
	t.Parallel()
	resolved := "demo-project"
	cases := []struct {
		name        string
		requestID   string
		wantErr     error
	}{
		{"empty request → fall back to resolved", "", nil},
		{"matching request", "demo-project", nil},
		{"matching with whitespace", "  demo-project  ", nil},
		{"mismatching", "other-project", domain.ErrIngestProjectIDMismatch},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := domain.ValidateProjectIDConsistency(tc.requestID, resolved)
			if tc.wantErr == nil {
				if err != nil {
					t.Errorf("unexpected err: %v", err)
				}
				return
			}
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("err: want %v, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestFilterStreamForProject_BlocksCrossProjectLeak(t *testing.T) {
	t.Parallel()
	stream := &domain.IngestStream{
		ID:        "ing_demo",
		ProjectID: "demo-project",
	}
	got, err := domain.FilterStreamForProject(stream, "demo-project")
	if err != nil {
		t.Fatalf("matching project must succeed: %v", err)
	}
	if got != stream {
		t.Errorf("must return same pointer for matching project")
	}

	// Cross-project: 404-äquivalent (kein Hinweis auf Existenz).
	if _, err := domain.FilterStreamForProject(stream, "other-project"); !errors.Is(err, domain.ErrIngestStreamNotFound) {
		t.Errorf("cross-project: want ErrIngestStreamNotFound, got %v", err)
	}
	// Nil-Eingabe: stable not-found, nicht panic.
	if _, err := domain.FilterStreamForProject(nil, "demo-project"); !errors.Is(err, domain.ErrIngestStreamNotFound) {
		t.Errorf("nil stream: want ErrIngestStreamNotFound, got %v", err)
	}
}

func TestRoutingRuleMode_DefaultIsSingle(t *testing.T) {
	t.Parallel()
	if domain.RoutingRuleModeSingle != "single" {
		t.Errorf("RoutingRuleModeSingle wire-Vertrag: want 'single', got %q",
			domain.RoutingRuleModeSingle)
	}
}

func TestStreamLifecycleEventKind_Constants(t *testing.T) {
	t.Parallel()
	// Wire-Vertrag aus spec/backend-api-contract.md §3.8.
	if domain.StreamLifecycleEventStarted != "stream_started" {
		t.Errorf("started: want stream_started, got %q", domain.StreamLifecycleEventStarted)
	}
	if domain.StreamLifecycleEventEnded != "stream_ended" {
		t.Errorf("ended: want stream_ended, got %q", domain.StreamLifecycleEventEnded)
	}
}

func TestMediaServerKind_Constants(t *testing.T) {
	t.Parallel()
	if domain.MediaServerKindMediaMTX != "mediamtx" {
		t.Errorf("mediamtx: want mediamtx, got %q", domain.MediaServerKindMediaMTX)
	}
	if domain.MediaServerKindSRS != "srs" {
		t.Errorf("srs: want srs, got %q", domain.MediaServerKindSRS)
	}
}

func TestIngestStreamStatus_Constants(t *testing.T) {
	t.Parallel()
	cases := map[domain.IngestStreamStatus]string{
		domain.IngestStreamStatusReady:    "ready",
		domain.IngestStreamStatusLive:     "live",
		domain.IngestStreamStatusEnded:    "ended",
		domain.IngestStreamStatusDisabled: "disabled",
	}
	for got, want := range cases {
		if string(got) != want {
			t.Errorf("status %q: want %q", got, want)
		}
	}
}
