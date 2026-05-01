package domain_test

import (
	"testing"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

// TestStreamAnalysisDomainError_Error stellt das error()-Format der
// Domain-Fehlerklasse fest. Konsumenten sehen den Code als Prefix
// im Log.
func TestStreamAnalysisDomainError_Error(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		err  *domain.StreamAnalysisDomainError
		want string
	}{
		{
			name: "nil receiver",
			err:  nil,
			want: "<nil>",
		},
		{
			name: "code + message",
			err: &domain.StreamAnalysisDomainError{
				Code:    domain.StreamAnalysisManifestNotHLS,
				Message: "Manifest beginnt nicht mit #EXTM3U.",
			},
			want: "manifest_not_hls: Manifest beginnt nicht mit #EXTM3U.",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.err.Error(); got != tc.want {
				t.Errorf("Error(): want %q, got %q", tc.want, got)
			}
		})
	}
}
