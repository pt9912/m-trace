package auth

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// RedisOriginRateLimiter (plan-0.12.6 Tranche 7 / R-22-Resttrigger)
// implementiert `driven.OriginRateLimiter` über atomare Token-Bucket-
// Operationen auf Redis — Single-Bucket pro Key, gemeinsam mit dem
// Issuance-Limiter auf demselben Redis-Server, aber mit eigenem
// Key-Prefix (`mtrace:origin`), damit kein Bucket-Konflikt entsteht.
//
// Fail-Mode-Vertrag analog `RedisIssuanceRateLimiter`:
//   - fail-closed (Default): Redis-Outage → `(false, nil)` (HTTP 429).
//   - fail-open (opt-in): `FailOpen=true` aktiviert lokalen
//     In-Process-Fallback-Bucket pro Replica.
type RedisOriginRateLimiter struct {
	client    redis.UniversalClient
	now       func() time.Time
	cfg       bucketConfig
	ttlSec    int
	failOpen  bool
	fallback  *InMemoryOriginRateLimiter
	logger    *slog.Logger
	scriptSHA string
	keyPrefix string
}

// redisOriginLuaScript: atomare Single-Bucket-Operation.
//
//	KEYS[1] = bucket key
//	ARGV[1] = capacity
//	ARGV[2] = refill per second
//	ARGV[3] = now in unix-nano
//	ARGV[4] = ttl seconds
//
// Returns 1 (allowed) oder 0 (denied).
const redisOriginLuaScript = `
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local refill = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local ttl = tonumber(ARGV[4])

if capacity == 0 and refill == 0 then
    return 1
end

local stored = redis.call('HMGET', key, 'tokens', 'last_at')
local tokens = tonumber(stored[1])
local last_at = tonumber(stored[2])
if tokens == nil then
    tokens = capacity
    last_at = now
end

local elapsed_sec = (now - last_at) / 1000000000.0
if elapsed_sec > 0 then
    tokens = math.min(capacity, tokens + elapsed_sec * refill)
end

local allowed = 0
if tokens >= 1.0 then
    tokens = tokens - 1.0
    allowed = 1
end
redis.call('HSET', key, 'tokens', tostring(tokens), 'last_at', tostring(now))
redis.call('EXPIRE', key, ttl)
return allowed
`

// RedisOriginLimiterConfig bündelt Bucket- und Fail-Mode-Parameter
// für `RedisOriginRateLimiter` (analog `RedisIssuanceLimiterConfig`).
type RedisOriginLimiterConfig struct {
	Capacity        int
	RefillPerSecond float64
	TTLSeconds      int
	FailOpen        bool
	KeyPrefix       string // default "mtrace:origin"
	Now             func() time.Time
}

// NewRedisOriginRateLimiter konstruiert den Adapter.
func NewRedisOriginRateLimiter(client redis.UniversalClient, cfg RedisOriginLimiterConfig, logger *slog.Logger) (*RedisOriginRateLimiter, error) {
	if client == nil {
		return nil, fmt.Errorf("redis-origin-limiter: client is nil")
	}
	prefix := cfg.KeyPrefix
	if prefix == "" {
		prefix = "mtrace:origin"
	}
	ttl := cfg.TTLSeconds
	if ttl <= 0 {
		ttl = 600 // 10 Minuten — Origin-Buckets sind kurzlebiger als Issuance.
	}
	nowFn := cfg.Now
	if nowFn == nil {
		nowFn = time.Now
	}
	l := &RedisOriginRateLimiter{
		client:    client,
		now:       nowFn,
		cfg:       bucketConfig{Capacity: cfg.Capacity, RefillPerSecond: cfg.RefillPerSecond},
		ttlSec:    ttl,
		failOpen:  cfg.FailOpen,
		logger:    logger,
		keyPrefix: prefix,
	}
	if cfg.FailOpen {
		l.fallback = NewInMemoryOriginRateLimiter(cfg.Capacity, cfg.RefillPerSecond)
	}
	if sha, err := client.ScriptLoad(context.Background(), redisOriginLuaScript).Result(); err == nil {
		l.scriptSHA = sha
	} else if logger != nil {
		logger.Warn("redis-origin-limiter: SCRIPT LOAD failed; falling back to inline EVAL",
			"error", err.Error())
	}
	return l, nil
}

// Allow prüft das Bucket atomar.
func (l *RedisOriginRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	if l == nil {
		return true, nil
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if key == "" {
		return true, nil
	}
	bucketKey := l.keyPrefix + ":" + key
	args := []any{
		l.cfg.Capacity,
		l.cfg.RefillPerSecond,
		strconv.FormatInt(l.now().UnixNano(), 10),
		l.ttlSec,
	}
	res, err := l.evalScript(ctx, []string{bucketKey}, args)
	if err != nil {
		return l.handleRedisError(ctx, key, err)
	}
	allowed, ok := res.(int64)
	if !ok {
		return l.handleRedisError(ctx, key,
			fmt.Errorf("unexpected redis result type %T", res))
	}
	return allowed == 1, nil
}

func (l *RedisOriginRateLimiter) evalScript(ctx context.Context, keys []string, args []any) (any, error) {
	if l.scriptSHA != "" {
		res, err := l.client.EvalSha(ctx, l.scriptSHA, keys, args...).Result()
		if err == nil {
			return res, nil
		}
		if !isNoScriptError(err) {
			return nil, err
		}
	}
	return l.client.Eval(ctx, redisOriginLuaScript, keys, args...).Result()
}

func (l *RedisOriginRateLimiter) handleRedisError(ctx context.Context, key string, err error) (bool, error) {
	if l.logger != nil {
		l.logger.Warn("redis-origin-limiter outage",
			"error", err.Error(),
			"fail_mode", failModeName(l.failOpen),
		)
	}
	if l.failOpen && l.fallback != nil {
		return l.fallback.Allow(ctx, key)
	}
	return false, nil
}

// Compile-time check.
var _ driven.OriginRateLimiter = (*RedisOriginRateLimiter)(nil)
