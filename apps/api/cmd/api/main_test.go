package main

import (
	"io"
	"log/slog"
	"testing"
)

// TestParseIngestRateLimit deckt den ENV-Override des Ingest-Rate-
// Limiters ab: Default bei fehlender ENV, gültiger Override, und
// Fallback auf Default bei ungültigen Werten. Der Override existiert
// für den Kapazitäts-Modus eines Last-Smoke.
func TestParseIngestRateLimit(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	t.Run("default when unset", func(t *testing.T) {
		t.Setenv(envRateLimitCapacity, "")
		t.Setenv(envRateLimitRefill, "")
		gotCap, gotRefill := parseIngestRateLimit(logger)
		if gotCap != defaultRateLimitCapacity || gotRefill != defaultRateLimitRefill {
			t.Fatalf("want default %d/%v, got %d/%v",
				defaultRateLimitCapacity, defaultRateLimitRefill, gotCap, gotRefill)
		}
	})

	t.Run("valid override applied", func(t *testing.T) {
		t.Setenv(envRateLimitCapacity, "5000")
		t.Setenv(envRateLimitRefill, "4000.5")
		gotCap, gotRefill := parseIngestRateLimit(logger)
		if gotCap != 5000 || gotRefill != 4000.5 {
			t.Fatalf("want 5000/4000.5, got %d/%v", gotCap, gotRefill)
		}
	})

	t.Run("invalid falls back to default", func(t *testing.T) {
		t.Setenv(envRateLimitCapacity, "-1")
		t.Setenv(envRateLimitRefill, "abc")
		gotCap, gotRefill := parseIngestRateLimit(logger)
		if gotCap != defaultRateLimitCapacity || gotRefill != defaultRateLimitRefill {
			t.Fatalf("invalid input should keep default, got %d/%v", gotCap, gotRefill)
		}
	})

	t.Run("partial override keeps default for the other", func(t *testing.T) {
		t.Setenv(envRateLimitCapacity, "2000")
		t.Setenv(envRateLimitRefill, "")
		gotCap, gotRefill := parseIngestRateLimit(logger)
		if gotCap != 2000 || gotRefill != defaultRateLimitRefill {
			t.Fatalf("want 2000/%v, got %d/%v", defaultRateLimitRefill, gotCap, gotRefill)
		}
	})
}
