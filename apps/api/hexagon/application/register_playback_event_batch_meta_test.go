package application_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driving"
)

// plan-0.4.0 §4.4 D1 — End-to-End-Tests für reservierte Meta-Validation
// und URL-Redaction. Use-Case ruft Validation/Redaction in `parseEvents`
// auf; ungültige Meta-Werte führen zu domain.ErrInvalidEvent (422), und
// URL-verdächtige Meta-Keys werden vor dem Persistenz-Append redigiert.

func TestRegisterBatch_AcceptsValidNetworkMeta(t *testing.T) {
	t.Parallel()
	uc, _, repo, _, metrics, _, _, _ := newUseCase()
	in := validBatch()
	in.Events[0].EventName = "manifest_loaded"
	in.Events[0].Meta = map[string]any{
		"network.kind":          "manifest",
		"network.detail_status": "available",
		"network.redacted_url":  "https://cdn.example.test/playlists/main.m3u8",
		"timing.fetch_ms":       float64(123),
	}
	res, err := uc.RegisterPlaybackEventBatch(context.Background(), in)
	if err != nil {
		t.Fatalf("expected accept, got %v", err)
	}
	if res.Accepted != 1 {
		t.Fatalf("expected 1 accepted, got %d", res.Accepted)
	}
	if metrics.invalid != 0 {
		t.Fatalf("invalid_events must not increment on valid meta, got %d", metrics.invalid)
	}
	got := repo.appended[0].Meta["network.redacted_url"].(string)
	if got != "https://cdn.example.test/playlists/main.m3u8" {
		t.Fatalf("pre-redacted url must persist unchanged, got %q", got)
	}
}

func TestRegisterBatch_RejectsInvalidReservedMeta(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		meta map[string]any
	}{
		{"unknown network.kind", map[string]any{"network.kind": "audio"}},
		{"object reserved value", map[string]any{"network.kind": map[string]any{"x": 1}}},
		{"reason without unavailable", map[string]any{
			"network.detail_status":      "available",
			"network.unavailable_reason": "browser_api_unavailable",
		}},
		{"reason violates pattern", map[string]any{
			"network.detail_status":      "network_detail_unavailable",
			"network.unavailable_reason": "BAD-VALUE",
		}},
		{"raw redacted_url with query", map[string]any{
			"network.redacted_url": "https://cdn.example.test/x.m3u8?token=abc",
		}},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			uc, _, repo, _, metrics, _, _, _ := newUseCase()
			in := validBatch()
			in.Events[0].Meta = tc.meta
			_, err := uc.RegisterPlaybackEventBatch(context.Background(), in)
			if !errors.Is(err, domain.ErrInvalidEvent) {
				t.Fatalf("expected ErrInvalidEvent, got %v", err)
			}
			if len(repo.appended) != 0 {
				t.Fatalf("invalid meta must not persist any event, got %d", len(repo.appended))
			}
			if metrics.invalid != len(in.Events) {
				t.Fatalf("invalid_events must increment by batch size, got %d", metrics.invalid)
			}
		})
	}
}

func TestRegisterBatch_RedactsURLishMetaBeforeAppend(t *testing.T) {
	t.Parallel()
	uc, _, repo, _, _, _, _, _ := newUseCase()
	in := validBatch()
	in.Events[0].EventName = "segment_loaded"
	in.Events[0].Meta = map[string]any{
		"segment_url":  "https://user:pw@cdn.example.test/" + strings.Repeat("a", 32) + "/seg.ts?token=secret#frag",
		"manifest_url": "https://cdn.example.test/m.m3u8?sig=xyz",
		"network.url":  "https://cdn.example.test/n.m3u8?token=abc",
		"foo_uri":      "https://cdn.example.test/foo.m3u8?key=k",
	}
	if _, err := uc.RegisterPlaybackEventBatch(context.Background(), in); err != nil {
		t.Fatalf("expected accept, got %v", err)
	}
	if len(repo.appended) != 1 {
		t.Fatalf("expected 1 appended event, got %d", len(repo.appended))
	}
	persisted := repo.appended[0].Meta
	for _, k := range []string{"segment_url", "manifest_url", "network.url", "foo_uri"} {
		s := persisted[k].(string)
		if strings.Contains(s, "?") || strings.Contains(s, "#") || strings.Contains(s, "@") {
			t.Errorf("meta[%q] still raw after redaction: %q", k, s)
		}
		if strings.Contains(s, "secret") || strings.Contains(s, "xyz") || strings.Contains(s, "token=") {
			t.Errorf("meta[%q] still leaks credential: %q", k, s)
		}
	}
	if got := persisted["segment_url"].(string); !strings.Contains(got, ":redacted") {
		t.Fatalf("expected token-segment redaction marker in segment_url, got %q", got)
	}
}

func TestRegisterBatch_ForwardCompatibleUnknownKeys(t *testing.T) {
	t.Parallel()
	// Unbekannte additive Keys außerhalb der reservierten Domänen und
	// ohne URL-Verdacht müssen unverändert durchlaufen — alte Backends
	// dürfen sie ignorieren, neue Backends dürfen sie konsumieren.
	uc, _, repo, _, _, _, _, _ := newUseCase()
	in := validBatch()
	in.Events[0].Meta = map[string]any{
		"experimental":   "anything-goes",
		"future_marker":  true,
		"network.foo":    "bar",
		"buffered_seconds": 1.8,
	}
	if _, err := uc.RegisterPlaybackEventBatch(context.Background(), in); err != nil {
		t.Fatalf("expected accept on forward-compat keys, got %v", err)
	}
	persisted := repo.appended[0].Meta
	if persisted["experimental"] != "anything-goes" || persisted["future_marker"] != true ||
		persisted["network.foo"] != "bar" || persisted["buffered_seconds"] != 1.8 {
		t.Fatalf("forward-compat keys mutated: %+v", persisted)
	}
}

// ensures the EventInput.Meta map is not aliased into the persisted
// event — redaction must mutate the use-case copy only.
func TestRegisterBatch_RedactionDoesNotMutateInputMeta(t *testing.T) {
	t.Parallel()
	uc, _, _, _, _, _, _, _ := newUseCase()
	in := validBatch()
	original := "https://cdn.example.test/x.m3u8?token=abc"
	in.Events[0].Meta = map[string]any{"url": original}
	caller := in.Events[0].Meta
	if _, err := uc.RegisterPlaybackEventBatch(context.Background(), in); err != nil {
		t.Fatalf("expected accept, got %v", err)
	}
	if caller["url"] != original {
		t.Fatalf("caller's meta map was mutated: got %v want %q", caller["url"], original)
	}
}

// keep the import of driving alive even if test signatures change.
var _ driving.EventInput
