package domain_test

import (
	"math"
	"testing"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// TestSampleRatePPMFromFloat_HappyCases (R-10):
// Float-in-`(0, 1]`-Range wird auf den deterministisch gerundeten ppm
// abgebildet; SampleRateFull (1.0) bleibt SampleRateFull.
func TestSampleRatePPMFromFloat_HappyCases(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   float64
		want int
	}{
		{"1.0 → SampleRateFull", 1.0, domain.SampleRateFull},
		{"0.5 → 500_000", 0.5, 500_000},
		{"0.1 → 100_000", 0.1, 100_000},
		{"0.001 → 1_000", 0.001, 1_000},
		{"0.000001 → 1", 0.000001, 1},
		{"round-half: 0.4999995 → 500_000", 0.4999995, 500_000},
		{"1e-7 clamps to min 1", 1e-7, 1},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := domain.SampleRatePPMFromFloat(tc.in)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %d, want %d", got, tc.want)
			}
		})
	}
}

// TestSampleRatePPMFromFloat_OutOfRange: Werte
// außerhalb `(0, 1]` und nicht-finite Floats liefern einen Fehler.
func TestSampleRatePPMFromFloat_OutOfRange(t *testing.T) {
	t.Parallel()
	bads := []float64{
		0.0,
		-0.5,
		1.5,
		2.0,
		math.NaN(),
		math.Inf(1),
		math.Inf(-1),
	}
	for _, x := range bads {
		x := x
		t.Run("", func(t *testing.T) {
			t.Parallel()
			if _, err := domain.SampleRatePPMFromFloat(x); err == nil {
				t.Errorf("SampleRatePPMFromFloat(%v) expected error, got nil", x)
			}
		})
	}
}
