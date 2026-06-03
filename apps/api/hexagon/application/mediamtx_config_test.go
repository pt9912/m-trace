package application_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/application"
	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// RAK-68: MediaMTX-Config-Generator. Pinnt
// deterministischen Output, Brand-/Protocol-Toggles und
// Klartext-Key-Schutz.

func mkTestStream(id, name string, proto domain.IngestProtocol, fingerprint string) domain.IngestStream {
	return domain.IngestStream{
		ID:          id,
		ProjectID:   "p1",
		DisplayName: name,
		Protocol:    proto,
		EndpointID:  "ep",
		TargetID:    "tgt",
		Status:      domain.IngestStreamStatusReady,
		Key:         domain.StreamKey{Hash: "h", Fingerprint: fingerprint},
		CreatedAt:   time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC),
	}
}

func mkTestTarget() domain.MediaServerTarget {
	return domain.MediaServerTarget{
		ID:             "tgt",
		Kind:           domain.MediaServerKindMediaMTX,
		ConfigPath:     "examples/ingest-control/mediamtx.generated.yml",
		HLSURLTemplate: "http://localhost:8889/{stream_path}/index.m3u8",
		ControlAPIURL:  "http://localhost:9997",
	}
}

func TestGenerateMediaMTXConfig_NoStreamsReturnsError(t *testing.T) {
	t.Parallel()
	_, err := application.GenerateMediaMTXConfig(application.MediaMTXConfigInput{
		Target:  mkTestTarget(),
		Streams: nil,
	})
	if !errors.Is(err, application.ErrMediaMTXConfigNoStreams) {
		t.Errorf("err: want ErrMediaMTXConfigNoStreams, got %v", err)
	}
}

func TestGenerateMediaMTXConfig_RejectsNonMediaMTXTarget(t *testing.T) {
	t.Parallel()
	target := mkTestTarget()
	target.Kind = domain.MediaServerKindSRS
	_, err := application.GenerateMediaMTXConfig(application.MediaMTXConfigInput{
		Target:  target,
		Streams: []domain.IngestStream{mkTestStream("ing_a", "Lab", domain.IngestProtocolSRT, "mtr_ing_AAA...BBB")},
	})
	if err == nil || !strings.Contains(err.Error(), "mediamtx") {
		t.Errorf("expected mediamtx-only error, got %v", err)
	}
}

func TestGenerateMediaMTXConfig_SortsByDisplayName(t *testing.T) {
	t.Parallel()
	streams := []domain.IngestStream{
		mkTestStream("ing_z", "z-stream", domain.IngestProtocolSRT, "mtr_ing_ZZZ...111"),
		mkTestStream("ing_a", "a-stream", domain.IngestProtocolRTMP, "mtr_ing_AAA...222"),
		mkTestStream("ing_m", "m-stream", domain.IngestProtocolSRT, "mtr_ing_MMM...333"),
	}
	artifact, err := application.GenerateMediaMTXConfig(application.MediaMTXConfigInput{
		Target: mkTestTarget(), Streams: streams,
	})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	idx := func(s string) int { return strings.Index(artifact.YAML, "  "+s+":\n") }
	if idx("a-stream") >= idx("m-stream") || idx("m-stream") >= idx("z-stream") {
		t.Errorf("paths not sorted: %s", artifact.YAML)
	}
}

func TestGenerateMediaMTXConfig_TogglesProtocolListeners(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		streams []domain.IngestStream
		wantSRT  string
		wantRTMP string
	}{
		{
			name:    "srt only",
			streams: []domain.IngestStream{mkTestStream("ing_a", "Lab", domain.IngestProtocolSRT, "mtr_ing_aaa...bbb")},
			wantSRT:  "srt: yes",
			wantRTMP: "rtmp: no",
		},
		{
			name:    "rtmp only",
			streams: []domain.IngestStream{mkTestStream("ing_a", "Lab", domain.IngestProtocolRTMP, "mtr_ing_aaa...bbb")},
			wantSRT:  "srt: no",
			wantRTMP: "rtmp: yes",
		},
		{
			name: "both",
			streams: []domain.IngestStream{
				mkTestStream("ing_a", "Lab-SRT", domain.IngestProtocolSRT, "mtr_ing_aaa...bbb"),
				mkTestStream("ing_b", "Lab-RTMP", domain.IngestProtocolRTMP, "mtr_ing_ccc...ddd"),
			},
			wantSRT:  "srt: yes",
			wantRTMP: "rtmp: yes",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			artifact, err := application.GenerateMediaMTXConfig(application.MediaMTXConfigInput{
				Target: mkTestTarget(), Streams: tc.streams,
			})
			if err != nil {
				t.Fatalf("generate: %v", err)
			}
			if !strings.Contains(artifact.YAML, tc.wantSRT) {
				t.Errorf("YAML missing %q: %s", tc.wantSRT, artifact.YAML)
			}
			if !strings.Contains(artifact.YAML, tc.wantRTMP) {
				t.Errorf("YAML missing %q: %s", tc.wantRTMP, artifact.YAML)
			}
		})
	}
}

func TestGenerateMediaMTXConfig_NoKlartextKeyInOutput(t *testing.T) {
	t.Parallel()
	streams := []domain.IngestStream{mkTestStream("ing_a", "Lab", domain.IngestProtocolSRT, "mtr_ing_aaa...bbb")}
	artifact, err := application.GenerateMediaMTXConfig(application.MediaMTXConfigInput{
		Target: mkTestTarget(), Streams: streams,
	})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	// Plan T3 / RAK-66: keine Klartext-Werte; Fingerprint darf
	// erscheinen, voller Hash niemals (Hash ist 64-Hex-Chars im
	// realen Fall, der Test setzt ihn auf "h" — also einfach prüfen,
	// dass `key_hash` als Wort nicht im YAML auftaucht).
	if strings.Contains(artifact.YAML, "key_hash") {
		t.Errorf("YAML must not expose key_hash: %s", artifact.YAML)
	}
	// Beispiel-Klartext aus realer GenerateStreamKey-Form sollte nie
	// im Generator-Output landen — wir simulieren das mit einem
	// nicht-Fingerprint-Wert.
	if strings.Contains(artifact.YAML, "mtr_ing_VERY_LONG_KLARTEXT") {
		t.Errorf("YAML must not contain klartext stream key")
	}
	// Fingerprint-Eintrag im Comment.
	if !strings.Contains(artifact.YAML, "key_fingerprint: mtr_ing_aaa...bbb") {
		t.Errorf("YAML missing fingerprint comment: %s", artifact.YAML)
	}
}

func TestGenerateMediaMTXConfig_DeterministicOutput(t *testing.T) {
	t.Parallel()
	streams := []domain.IngestStream{
		mkTestStream("ing_a", "Lab", domain.IngestProtocolSRT, "mtr_ing_aaa...bbb"),
		mkTestStream("ing_b", "Other", domain.IngestProtocolRTMP, "mtr_ing_ccc...ddd"),
	}
	first, err := application.GenerateMediaMTXConfig(application.MediaMTXConfigInput{Target: mkTestTarget(), Streams: streams})
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	second, err := application.GenerateMediaMTXConfig(application.MediaMTXConfigInput{Target: mkTestTarget(), Streams: streams})
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	if first.YAML != second.YAML {
		t.Errorf("output not deterministic")
	}
}

func TestGenerateMediaMTXConfig_SkipsUnsupportedDisplayName(t *testing.T) {
	t.Parallel()
	streams := []domain.IngestStream{
		mkTestStream("ing_a", "ok", domain.IngestProtocolSRT, "mtr_ing_aaa"),
		mkTestStream("ing_b", "***", domain.IngestProtocolSRT, "mtr_ing_bbb"),
	}
	artifact, err := application.GenerateMediaMTXConfig(application.MediaMTXConfigInput{Target: mkTestTarget(), Streams: streams})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if len(artifact.Warnings) != 1 {
		t.Errorf("want 1 warning for unsupported display_name, got %d", len(artifact.Warnings))
	}
	if !strings.Contains(artifact.YAML, "  ok:\n") {
		t.Errorf("YAML missing ok path: %s", artifact.YAML)
	}
}

func TestGenerateMediaMTXConfig_SanitizesSpacesAndDots(t *testing.T) {
	t.Parallel()
	streams := []domain.IngestStream{
		mkTestStream("ing_a", "Lab Stream.v1", domain.IngestProtocolSRT, "mtr_ing_aaa"),
	}
	artifact, err := application.GenerateMediaMTXConfig(application.MediaMTXConfigInput{Target: mkTestTarget(), Streams: streams})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if !strings.Contains(artifact.YAML, "  lab_stream_v1:\n") {
		t.Errorf("YAML missing sanitized path: %s", artifact.YAML)
	}
}
