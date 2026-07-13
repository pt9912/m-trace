package ratelimit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// RedisTokenBucketRateLimiter (R-26 b) implementiert `driven.RateLimiter`
// über einen geteilten Redis: die drei Bucket-Dimensionen (project_id /
// client_ip / origin) liegen als Redis-Hashes auf EINEM Server, sodass
// N API-Replicas ein gemeinsames Per-Projekt-Budget teilen — statt
// N × Capacity beim In-Process-`TokenBucketRateLimiter`.
//
// Semantik identisch zum In-Memory-Adapter: n Tokens werden all-or-nothing
// aus jeder nicht-leeren Dimension konsumiert (ein Deny verbraucht nirgends
// Tokens); leerer Key ist ein No-op. Atomar durch EIN Lua-Script pro Allow
// (kein Refund-Tanz über mehrere Roundtrips).
//
// Fail-Mode-Vertrag (bewusst ANDERS als Origin-/Issuance-Limiter, die
// fail-closed defaulten und einen Schalter teilen — Schutzgut dort ist
// Auth-Flutung, hier Telemetrie-Verfügbarkeit):
//   - fail-open (Default): Redis-Outage → Delegation an einen lokalen
//     In-Process-Fallback-Bucket pro Replica. Degradation exakt auf den
//     Zustand vor R-26 b: Limitierung bleibt, nur die repliken-
//     übergreifende Fairness pausiert. Ein-/Austritt wird je einmal
//     als WARN geloggt (kein Log-Flood auf dem Hot-Path).
//   - fail-closed (opt-in, MTRACE_RATE_LIMIT_FAIL_CLOSED=1): Outage →
//     Deny (`domain.ErrRateLimited`, HTTP 429) — nie ein 500.
//
// Ein abgebrochener Context wird wie eine Outage behandelt (Fallback bzw.
// Deny) statt `ctx.Err()` durch den Port zu reichen: der Port kennt nur
// `error`, und die Call-Site zählt jeden Fehler als rate-limited — ein
// Context-Fehler würde dort Metriken verfälschen und als 500 enden.
type RedisTokenBucketRateLimiter struct {
	client     redis.UniversalClient
	capacity   int
	refill     float64
	ttlSec     int
	now        func() time.Time
	failClosed bool
	fallback   *TokenBucketRateLimiter
	logger     *slog.Logger
	scriptSHA  string
	keyPrefix  string
	degraded   atomic.Bool
}

// redisIngestLuaScript: atomare all-or-nothing-Konsumtion von n Tokens
// aus bis zu drei Buckets.
//
//	KEYS[1..k] = Bucket-Keys der nicht-leeren Dimensionen (k ∈ 1..3)
//	ARGV[1] = n (Batch-Größe in Tokens)
//	ARGV[2] = capacity
//	ARGV[3] = refill per second
//	ARGV[4] = now in unix-nano (Client-Uhr der Replica)
//	ARGV[5] = ttl seconds
//
// Uhren-Drift-Schutz (die Replicas liefern `now` von N Uhren): elapsed < 0
// zählt als 0 (kein Token-Fressen) UND `last_at` regressiert nie — gespeichert
// wird max(stored_last_at, now). Ohne das zweite würde eine nachgehende
// Replica `last_at` zurücksetzen und die vorgehende bekäme die Skew-Differenz
// bei jedem Wechsel erneut als Refill gutgeschrieben (systematische
// Inflation auf dem Hot-Path); so bleibt Skew ein einmaliger, begrenzter
// Effekt. Der Deny-Pfad persistiert den Refill-Stand ohne Abbuchung.
//
// Returns 1 (allowed) oder 0 (denied).
const redisIngestLuaScript = `
local n = tonumber(ARGV[1])
local capacity = tonumber(ARGV[2])
local refill = tonumber(ARGV[3])
local now = tonumber(ARGV[4])
local ttl = tonumber(ARGV[5])

local tokens = {}
local last = {}
for i, key in ipairs(KEYS) do
    local stored = redis.call('HMGET', key, 'tokens', 'last_at')
    local t = tonumber(stored[1])
    local l = tonumber(stored[2])
    if t == nil then
        t = capacity
        l = now
    end
    if l < now then
        t = math.min(capacity, t + ((now - l) / 1000000000.0) * refill)
        l = now
    end
    tokens[i] = t
    last[i] = l
end

local allowed = 1
for i = 1, #KEYS do
    if tokens[i] < n then
        allowed = 0
    end
end

for i, key in ipairs(KEYS) do
    local t = tokens[i]
    if allowed == 1 then
        t = t - n
    end
    redis.call('HSET', key, 'tokens', tostring(t), 'last_at', tostring(last[i]))
    redis.call('EXPIRE', key, ttl)
end
return allowed
`

// RedisTokenBucketConfig bündelt Bucket- und Fail-Mode-Parameter für
// `RedisTokenBucketRateLimiter` (analog `RedisOriginLimiterConfig`).
type RedisTokenBucketConfig struct {
	Capacity        int
	RefillPerSecond float64
	TTLSeconds      int
	FailClosed      bool
	KeyPrefix       string // default "mtrace:ingest"
	Now             func() time.Time
}

// NewRedisTokenBucketRateLimiter konstruiert den Adapter.
func NewRedisTokenBucketRateLimiter(client redis.UniversalClient, cfg RedisTokenBucketConfig, logger *slog.Logger) (*RedisTokenBucketRateLimiter, error) {
	if client == nil {
		return nil, fmt.Errorf("redis-ingest-limiter: client is nil")
	}
	prefix := cfg.KeyPrefix
	if prefix == "" {
		prefix = "mtrace:ingest"
	}
	ttl := cfg.TTLSeconds
	if ttl <= 0 {
		ttl = 600 // 10 Minuten — idle Buckets sind nach capacity/refill Sekunden ohnehin voll.
	}
	nowFn := cfg.Now
	if nowFn == nil {
		nowFn = time.Now
	}
	l := &RedisTokenBucketRateLimiter{
		client:     client,
		capacity:   cfg.Capacity,
		refill:     cfg.RefillPerSecond,
		ttlSec:     ttl,
		now:        nowFn,
		failClosed: cfg.FailClosed,
		logger:     logger,
		keyPrefix:  prefix,
	}
	if !cfg.FailClosed {
		l.fallback = NewTokenBucketRateLimiter(cfg.Capacity, cfg.RefillPerSecond, nowFn)
	}
	if sha, err := client.ScriptLoad(context.Background(), redisIngestLuaScript).Result(); err == nil {
		l.scriptSHA = sha
	} else if logger != nil {
		logger.Warn("redis-ingest-limiter: SCRIPT LOAD failed; falling back to inline EVAL",
			"error", err.Error())
	}
	return l, nil
}

// Allow konsumiert n Tokens aus jeder gesetzten Dimension von key —
// all-or-nothing, atomar über das Lua-Script.
func (l *RedisTokenBucketRateLimiter) Allow(ctx context.Context, key driven.RateLimitKey, n int) error {
	if n <= 0 {
		return nil
	}
	keys := redisBucketKeys(l.keyPrefix, key)
	if len(keys) == 0 {
		return nil
	}
	if err := ctx.Err(); err != nil {
		return l.handleRedisError(ctx, key, n, err)
	}
	args := []any{
		n,
		l.capacity,
		l.refill,
		strconv.FormatInt(l.now().UnixNano(), 10),
		l.ttlSec,
	}
	res, err := l.evalScript(ctx, keys, args)
	if err != nil {
		return l.handleRedisError(ctx, key, n, err)
	}
	l.markRecovered()
	allowed, ok := res.(int64)
	if !ok {
		return l.handleRedisError(ctx, key, n,
			fmt.Errorf("unexpected redis result type %T", res))
	}
	if allowed != 1 {
		return domain.ErrRateLimited
	}
	return nil
}

func (l *RedisTokenBucketRateLimiter) evalScript(ctx context.Context, keys []string, args []any) (any, error) {
	if l.scriptSHA != "" {
		res, err := l.client.EvalSha(ctx, l.scriptSHA, keys, args...).Result()
		if err == nil {
			return res, nil
		}
		if !isNoScriptError(err) {
			return nil, err
		}
	}
	return l.client.Eval(ctx, redisIngestLuaScript, keys, args...).Result()
}

// handleRedisError ist der Outage-Pfad: einmaliges WARN beim Eintritt
// (kein Flood — Allow läuft pro Batch-Request), dann Fail-Mode-Politik.
func (l *RedisTokenBucketRateLimiter) handleRedisError(ctx context.Context, key driven.RateLimitKey, n int, err error) error {
	if l.degraded.CompareAndSwap(false, true) && l.logger != nil {
		l.logger.Warn("redis-ingest-limiter outage — degraded mode begins",
			"error", err.Error(),
			"fail_mode", ingestFailModeName(l.failClosed),
		)
	}
	if l.failClosed || l.fallback == nil {
		return domain.ErrRateLimited
	}
	return l.fallback.Allow(ctx, key, n)
}

// markRecovered loggt das Ende einer Degradations-Phase genau einmal —
// im fail-open-Modus war währenddessen die repliken-übergreifende
// Fairness pausiert (per-Replica-Fallback), das soll nicht still enden.
func (l *RedisTokenBucketRateLimiter) markRecovered() {
	if l.degraded.CompareAndSwap(true, false) && l.logger != nil {
		l.logger.Warn("redis-ingest-limiter recovered — degraded mode ends",
			"fail_mode", ingestFailModeName(l.failClosed),
		)
	}
}

// redisBucketKeys bildet die Redis-Keys der nicht-leeren Dimensionen.
// project_id und client_ip sind längenbegrenzt/validiert und gehen raw
// in den Key; NUR der Origin wird gehasht — der Header ist client-
// kontrolliert und unbegrenzt lang, der Hash bounded die Key-Länge
// (Namensgebung analog RAK-90 „Origin-Header-Hash").
func redisBucketKeys(prefix string, key driven.RateLimitKey) []string {
	out := make([]string, 0, 3)
	if key.ProjectID != "" {
		out = append(out, prefix+":project:"+key.ProjectID)
	}
	if key.ClientIP != "" {
		out = append(out, prefix+":ip:"+key.ClientIP)
	}
	if key.Origin != "" {
		sum := sha256.Sum256([]byte(key.Origin))
		out = append(out, prefix+":origin:"+hex.EncodeToString(sum[:16]))
	}
	return out
}

func ingestFailModeName(failClosed bool) string {
	if failClosed {
		return "fail-closed (deny on outage)"
	}
	return "fail-open (local memory fallback)"
}

func isNoScriptError(err error) bool {
	return err != nil && strings.HasPrefix(err.Error(), "NOSCRIPT")
}

// Compile-time check.
var _ driven.RateLimiter = (*RedisTokenBucketRateLimiter)(nil)
