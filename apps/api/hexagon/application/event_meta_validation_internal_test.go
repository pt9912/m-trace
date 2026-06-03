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
	// API-Kontrakt
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

//  — alle Codes der WebRTC-Fehlercode-Allowlist.
func TestValidateReservedEventMeta_WebRTCErrorCodesAccepted(t *testing.T) {
	t.Parallel()
	codes := []string{
		"whep_signaling_failed",
		"whep_sdp_invalid",
		"webrtc_no_tracks",
		"peer_connection_failed",
		"webrtc_destroyed_before_connected",
	}
	for _, c := range codes {
		t.Run(c, func(t *testing.T) {
			t.Parallel()
			meta := domain.EventMeta{"webrtc.error_code": c}
			if err := validateReservedEventMeta(meta); err != nil {
				t.Fatalf("expected error_code %q to validate, got %v", c, err)
			}
		})
	}
}

//  — webrtc.*-Allowlist-Tests.
func TestValidateReservedEventMeta_WebRTCHappyPath(t *testing.T) {
	t.Parallel()
	meta := domain.EventMeta{
		"webrtc.peer_connection_run_id": "run-id-1",
		"webrtc.sample_id":              int64(7),
		"webrtc.connection_state":       "connected",
		"webrtc.ice_state":              "completed",
		"webrtc.dtls_state":             "connected",
		"webrtc.packets_lost":           int64(2),
		"webrtc.bytes_received":         float64(123456),
		"webrtc.bytes_sent":             int64(78910),
	}
	if err := validateReservedEventMeta(meta); err != nil {
		t.Fatalf("expected webrtc happy path to pass, got %v", err)
	}
}

// Float64-Wire-Repräsentation für Counter-Felder (JSON-Decoder
// liefert Numbers als float64).
func TestValidateReservedEventMeta_WebRTCFloat64Accepted(t *testing.T) {
	t.Parallel()
	meta := domain.EventMeta{
		"webrtc.peer_connection_run_id": "run-a",
		"webrtc.sample_id":              float64(0),
		"webrtc.connection_state":       "connected",
		"webrtc.ice_state":              "completed",
		"webrtc.dtls_state":             "connected",
		"webrtc.packets_lost":           float64(2),
		"webrtc.bytes_received":         float64(123),
		"webrtc.bytes_sent":             float64(456),
	}
	if err := validateReservedEventMeta(meta); err != nil {
		t.Fatalf("expected float64 wire form to pass, got %v", err)
	}
}

// requireBoundedString akzeptiert kurze Strings.
func TestValidateReservedEventMeta_WebRTCErrorDetailWithinLimit(t *testing.T) {
	t.Parallel()
	meta := domain.EventMeta{"webrtc.error_detail": "short detail"}
	if err := validateReservedEventMeta(meta); err != nil {
		t.Fatalf("expected short error_detail to pass, got %v", err)
	}
}

func TestValidateReservedEventMeta_WebRTCRejections(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		meta domain.EventMeta
	}{
		{
			"unknown webrtc.* key",
			domain.EventMeta{"webrtc.unknown_field": "value"},
		},
		{
			"forbidden per-identifier (track_id)",
			domain.EventMeta{"webrtc.track_id": "track-1"},
		},
		{
			"forbidden per-identifier (ssrc)",
			domain.EventMeta{"webrtc.ssrc": int64(12345)},
		},
		{
			"forbidden per-identifier (user_agent)",
			domain.EventMeta{"webrtc.user_agent": "Mozilla/5.0"},
		},
		{
			"connection_state outside enum",
			domain.EventMeta{"webrtc.connection_state": "stalled"},
		},
		{
			"ice_state outside enum",
			domain.EventMeta{"webrtc.ice_state": "frozen"},
		},
		{
			"dtls_state outside enum",
			domain.EventMeta{"webrtc.dtls_state": "negotiating"},
		},
		{
			"connection_state wrong type",
			domain.EventMeta{"webrtc.connection_state": int64(1)},
		},
		{
			"sample_id negative",
			domain.EventMeta{"webrtc.sample_id": int64(-1)},
		},
		{
			"sample_id non-integer float",
			domain.EventMeta{"webrtc.sample_id": float64(3.5)},
		},
		{
			"packets_lost wrong type",
			domain.EventMeta{"webrtc.packets_lost": "5"},
		},
		{
			"bytes_received negative",
			domain.EventMeta{"webrtc.bytes_received": int64(-1)},
		},
		{
			"run_id pattern violation",
			domain.EventMeta{"webrtc.peer_connection_run_id": "INVALID Spaces!"},
		},
		{
			"run_id wrong type",
			domain.EventMeta{"webrtc.peer_connection_run_id": int64(1)},
		},
		{
			"error_code outside enum",
			domain.EventMeta{"webrtc.error_code": "made_up"},
		},
		{
			"error_detail too long",
			domain.EventMeta{"webrtc.error_detail": stringOfLen(257)},
		},
		{
			"error_detail wrong type",
			domain.EventMeta{"webrtc.error_detail": int64(1)},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := validateReservedEventMeta(tc.meta)
			if !errors.Is(err, domain.ErrInvalidEvent) {
				t.Fatalf("expected ErrInvalidEvent for %s, got %v", tc.name, err)
			}
		})
	}
}

func stringOfLen(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = 'a'
	}
	return string(b)
}

// TestValidateSessionSampleRate (R-10): Wire-
// Range-Check `(0, 1]`. Accepted: float64 + int64; rejected: andere
// Typen, Out-of-Range-Werte.
func TestValidateSessionSampleRate(t *testing.T) {
	t.Parallel()
	ok := []domain.EventMeta{
		{"session_sample_rate": 1.0},
		{"session_sample_rate": 0.5},
		{"session_sample_rate": int64(1)},
	}
	for _, m := range ok {
		if err := validateReservedEventMeta(m); err != nil {
			t.Errorf("unexpected error for %v: %v", m, err)
		}
	}

	bad := []domain.EventMeta{
		{"session_sample_rate": "0.5"},      // string
		{"session_sample_rate": 0.0},        // out of range
		{"session_sample_rate": 1.5},        // out of range
		{"session_sample_rate": []any{0.5}}, // unsupported type
	}
	for _, m := range bad {
		if err := validateReservedEventMeta(m); !errors.Is(err, domain.ErrInvalidEvent) {
			t.Errorf("expected ErrInvalidEvent for %v, got %v", m, err)
		}
	}
}
