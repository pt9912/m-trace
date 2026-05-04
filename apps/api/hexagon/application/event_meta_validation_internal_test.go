package application

import (
	"errors"
	"strings"
	"testing"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

func TestValidateReservedEventMeta_AcceptsValidNetworkMeta(t *testing.T) {
	t.Parallel()
	meta := domain.EventMeta{
		"network.kind":          "manifest",
		"network.detail_status": "available",
		"network.redacted_url":  "https://cdn.example.test/playlists/main.m3u8",
		"timing.fetch_ms":       float64(123),
		// nicht-reservierter Key bleibt erlaubt (Vorwärtskompatibilität)
		"buffered_seconds": 1.8,
	}
	if err := validateReservedEventMeta(meta); err != nil {
		t.Fatalf("expected valid meta, got %v", err)
	}
}

func TestValidateReservedEventMeta_AcceptsUnavailableWithReason(t *testing.T) {
	t.Parallel()
	meta := domain.EventMeta{
		"network.kind":               "segment",
		"network.detail_status":      "network_detail_unavailable",
		"network.unavailable_reason": "cors_timing_blocked",
	}
	if err := validateReservedEventMeta(meta); err != nil {
		t.Fatalf("expected valid meta, got %v", err)
	}
}

func TestValidateReservedEventMeta_RejectsForbiddenDomain(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		meta domain.EventMeta
	}{
		{"unknown network.kind", domain.EventMeta{"network.kind": "audio"}},
		{"unknown detail_status", domain.EventMeta{"network.detail_status": "ok"}},
		{"unknown reason enum", domain.EventMeta{
			"network.detail_status":      "network_detail_unavailable",
			"network.unavailable_reason": "totally_made_up",
		}},
		{"reason violates pattern", domain.EventMeta{
			"network.detail_status":      "network_detail_unavailable",
			"network.unavailable_reason": "UPPER_CASE",
		}},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := validateReservedEventMeta(tc.meta)
			if !errors.Is(err, domain.ErrInvalidEvent) {
				t.Fatalf("expected ErrInvalidEvent, got %v", err)
			}
		})
	}
}

func TestValidateReservedEventMeta_RejectsObjectAndArrayValues(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		v    any
	}{
		{"object value", map[string]any{"nested": true}},
		{"array value", []any{"manifest"}},
		{"int value", 42},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := validateReservedEventMeta(domain.EventMeta{
				"network.kind": tc.v,
			})
			if !errors.Is(err, domain.ErrInvalidEvent) {
				t.Fatalf("expected ErrInvalidEvent, got %v", err)
			}
		})
	}
}

func TestValidateReservedEventMeta_ReasonRequiresUnavailable(t *testing.T) {
	t.Parallel()
	meta := domain.EventMeta{
		"network.detail_status":      "available",
		"network.unavailable_reason": "browser_api_unavailable",
	}
	err := validateReservedEventMeta(meta)
	if !errors.Is(err, domain.ErrInvalidEvent) {
		t.Fatalf("expected ErrInvalidEvent, got %v", err)
	}
	if !strings.Contains(err.Error(), "requires") {
		t.Fatalf("expected requires-mismatch hint, got %q", err.Error())
	}
}

func TestValidateReservedEventMeta_RejectsRawRedactedURL(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		raw  string
	}{
		{"with query", "https://cdn.example.test/playlist.m3u8?token=abc"},
		{"with fragment", "https://cdn.example.test/playlist.m3u8#frag"},
		{"with userinfo", "https://user:pw@cdn.example.test/playlist.m3u8"},
		{"with token segment", "https://cdn.example.test/" + strings.Repeat("a", 32) + "/playlist.m3u8"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := validateReservedEventMeta(domain.EventMeta{
				"network.redacted_url": tc.raw,
			})
			if !errors.Is(err, domain.ErrInvalidEvent) {
				t.Fatalf("expected ErrInvalidEvent for %s, got %v", tc.name, err)
			}
		})
	}
}

func TestValidateReservedEventMeta_TimingValues(t *testing.T) {
	t.Parallel()
	good := domain.EventMeta{
		"timing.start":    "2026-04-28T12:00:00.000Z",
		"timing.duration": float64(150),
		"timing.queued":   int64(7),
	}
	if err := validateReservedEventMeta(good); err != nil {
		t.Fatalf("expected valid timing meta, got %v", err)
	}
	bad := []struct {
		name string
		meta domain.EventMeta
	}{
		{"object timing", domain.EventMeta{"timing.start": map[string]any{}}},
		{"array timing", domain.EventMeta{"timing.start": []any{"x"}}},
		{"empty string", domain.EventMeta{"timing.start": ""}},
		{"non-rfc3339", domain.EventMeta{"timing.start": "not-a-time"}},
	}
	for _, tc := range bad {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if err := validateReservedEventMeta(tc.meta); !errors.Is(err, domain.ErrInvalidEvent) {
				t.Fatalf("expected ErrInvalidEvent, got %v", err)
			}
		})
	}
}

func TestValidateReservedEventMeta_ForwardCompatibility(t *testing.T) {
	t.Parallel()
	// Unbekannte additive Keys außerhalb der reservierten Domänen
	// dürfen nicht abgelehnt werden — Vorwärtskompatibilität nach
	// API-Kontrakt §3.4.
	meta := domain.EventMeta{
		"network.foo":    "bar",
		"network.kind":   "manifest",
		"future_marker":  true,
		"experimental":   "anything-goes",
		"timing.unknown": "2026-04-28T12:00:00Z",
	}
	if err := validateReservedEventMeta(meta); err != nil {
		t.Fatalf("expected forward-compat acceptance, got %v", err)
	}
}
