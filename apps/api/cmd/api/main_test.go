package main

import (
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/ratelimit"
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

// TestBuildIngestRateLimiter deckt den Backend-Selektor
// `MTRACE_RATE_LIMIT_BACKEND` (R-26 b) ab: Default/`memory` liefern den
// unveränderten In-Process-Bucket, `redis` den shared Adapter, und der
// Boot-Validator lehnt `sqlite`/`memcached`/Unbekanntes mit präziser
// Begründung ab (RAK-90-Stil).
func TestBuildIngestRateLimiter(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	t.Run("default is the in-process bucket", func(t *testing.T) {
		t.Setenv(envRateLimitBackend, "")
		l, err := buildIngestRateLimiter(logger)
		if err != nil {
			t.Fatalf("default backend: %v", err)
		}
		if _, ok := l.(*ratelimit.TokenBucketRateLimiter); !ok {
			t.Fatalf("default backend: got %T, want *ratelimit.TokenBucketRateLimiter", l)
		}
	})

	t.Run("memory explicit", func(t *testing.T) {
		t.Setenv(envRateLimitBackend, "memory")
		l, err := buildIngestRateLimiter(logger)
		if err != nil {
			t.Fatalf("memory backend: %v", err)
		}
		if _, ok := l.(*ratelimit.TokenBucketRateLimiter); !ok {
			t.Fatalf("memory backend: got %T", l)
		}
	})

	t.Run("redis requires MTRACE_REDIS_ADDR", func(t *testing.T) {
		t.Setenv(envRateLimitBackend, "redis")
		t.Setenv(envRedisAddr, "")
		if _, err := buildIngestRateLimiter(logger); err == nil ||
			!strings.Contains(err.Error(), envRedisAddr) {
			t.Fatalf("redis without addr: err=%v, want mention of %s", err, envRedisAddr)
		}
	})

	t.Run("redis constructs the shared adapter", func(t *testing.T) {
		t.Setenv(envRateLimitBackend, "Redis") // case-insensitiv
		t.Setenv(envRedisAddr, "127.0.0.1:1")  // unerreichbar: Konstruktion bleibt non-fatal
		l, err := buildIngestRateLimiter(logger)
		if err != nil {
			t.Fatalf("redis backend: %v", err)
		}
		if _, ok := l.(*ratelimit.RedisTokenBucketRateLimiter); !ok {
			t.Fatalf("redis backend: got %T, want *ratelimit.RedisTokenBucketRateLimiter", l)
		}
	})

	t.Run("rejected and unknown values fail loud", func(t *testing.T) {
		for backend, wantSubstr := range map[string]string{
			"sqlite":    "not Multi-Host-safe",
			"memcached": "follow-up item",
			"etcd":      "valid: memory|redis",
		} {
			t.Setenv(envRateLimitBackend, backend)
			_, err := buildIngestRateLimiter(logger)
			if err == nil || !strings.Contains(err.Error(), wantSubstr) {
				t.Errorf("backend %q: err=%v, want substring %q", backend, err, wantSubstr)
			}
		}
	})
}

// TestRateLimitFailClosedOptIn: nur explizit-truthy Werte schalten auf
// fail-closed; Default (und alles andere) bleibt fail-open (§4.3 des
// R-26-b-Plans — bewusst anders als der geteilte Auth-Schalter).
func TestRateLimitFailClosedOptIn(t *testing.T) {
	for raw, want := range map[string]bool{
		"": false, "0": false, "no": false, "off": false,
		"1": true, "true": true, "YES": true,
	} {
		t.Setenv(envRateLimitFailClosed, raw)
		if got := rateLimitFailClosedOptIn(); got != want {
			t.Errorf("%s=%q: got %v, want %v", envRateLimitFailClosed, raw, got, want)
		}
	}
}
