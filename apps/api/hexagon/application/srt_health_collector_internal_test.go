package application

import (
	"testing"
	"time"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// Internal-Tests für die unexported Helper aus
// `srt_health_collector.go`. Sie pinnen die Schwellen-/Mapping-
// Logik direkt, ohne über den öffentlichen `Run`-Loop zu gehen —
// damit bleiben Coverage-Lücken (Branch-Cap, Status-Rank) klein.

func TestNextBackoff_Branches(t *testing.T) {
	t.Parallel()
	poll := 5 * time.Second
	maxBackoff := 60 * time.Second

	cases := []struct {
		name    string
		current time.Duration
		want    time.Duration
	}{
		{"below pollInterval rises to pollInterval", 1 * time.Second, poll},
		{"doubling stays under cap", 10 * time.Second, 20 * time.Second},
		{"doubling caps at maxBackoff", 40 * time.Second, maxBackoff},
		{"already at cap stays at cap", maxBackoff, maxBackoff},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := nextBackoff(tc.current, poll, maxBackoff)
			if got != tc.want {
				t.Fatalf("nextBackoff(%v, %v, %v) = %v, want %v", tc.current, poll, maxBackoff, got, tc.want)
			}
		})
	}
}

func TestWorseSourceStatus_Rank(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		a, b domain.SourceStatus
		want domain.SourceStatus
	}{
		{"unavailable wins over stale", domain.SourceStatusUnavailable, domain.SourceStatusStale, domain.SourceStatusUnavailable},
		{"stale wins over partial", domain.SourceStatusStale, domain.SourceStatusPartial, domain.SourceStatusStale},
		{"partial wins over no_active_connection", domain.SourceStatusPartial, domain.SourceStatusNoActiveConnection, domain.SourceStatusPartial},
		{"no_active_connection wins over ok", domain.SourceStatusNoActiveConnection, domain.SourceStatusOK, domain.SourceStatusNoActiveConnection},
		{"ok stays ok against ok", domain.SourceStatusOK, domain.SourceStatusOK, domain.SourceStatusOK},
		{"order independent — b worse than a", domain.SourceStatusOK, domain.SourceStatusUnavailable, domain.SourceStatusUnavailable},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := worseSourceStatus(tc.a, tc.b); got != tc.want {
				t.Fatalf("worseSourceStatus(%v, %v) = %v, want %v", tc.a, tc.b, got, tc.want)
			}
		})
	}
}
