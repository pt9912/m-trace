package auth

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
	"github.com/pt9912/m-trace/apps/api/hexagon/port/driven"
)

// RedisIssuanceRateLimiter (plan-0.12.6 Tranche 7 / R-17) implementiert
// `driven.IssuanceRateLimiter` über atomare Token-Bucket-Operationen
// auf einem Redis-Server. Mehrere API-Replicas — auch über Hosts
// hinweg — teilen sich denselben Bucket-Counter; der `0.12.5`-SQLite-
// Pfad ist auf Single-Host-Shared-Volume beschränkt.
//
// Atomicity-Vertrag:
//   - Ein einziger `EVAL`-Lua-Script-Aufruf für beide Buckets
//     (global → project, mit Refund bei project-deny) garantiert,
//     dass kein anderer Caller die Buckets zwischen den zwei
//     Refill/Consume-Schritten verändert (Race-frei).
//   - Bucket-Keys: `mtrace:issuance:global` und
//     `mtrace:issuance:project:<projectID>`. TTL hält idle Buckets
//     aus dem Redis-Speicher; Default 24 h analog `0.12.5` SQLite-
//     Limiter.
//
// Fail-Mode-Vertrag:
//   - **fail-closed (Default)**: jede Redis-Connection-/EVAL-Error-
//     Klasse wird als „deny" gemeldet (`(false, nil)`); der HTTP-
//     Handler liefert dann `429 auth_issuance_rate_limited`. Damit
//     wird ein Redis-Outage nie zur Mint-Welle.
//   - **fail-open (opt-in)**: wenn der Operator `FailOpen=true`
//     setzt (ENV `MTRACE_AUTH_ISSUANCE_FAIL_OPEN=1`), fällt der
//     Limiter bei Redis-Outage auf einen lokalen In-Process-
//     Token-Bucket-Fallback zurück (`fallback` Field). Der
//     Fallback misst pro Replica — explizite Operator-Entscheidung
//     gegen den Single-Host-Bucket-Konsens.
type RedisIssuanceRateLimiter struct {
	client       redis.UniversalClient
	now          func() time.Time
	globalCfg    bucketConfig
	projectCfg   bucketConfig
	ttlSeconds   int
	failOpen     bool
	fallback     *InMemoryIssuanceRateLimiter
	logger       *slog.Logger
	scriptSHA    string
	keyPrefix    string
}

// redisIssuanceLuaScript: atomarer Two-Bucket-Check.
//
//	KEYS[1] = global bucket key
//	KEYS[2] = project bucket key
//	ARGV[1] = global capacity
//	ARGV[2] = global refill per second
//	ARGV[3] = project capacity
//	ARGV[4] = project refill per second
//	ARGV[5] = now in unix-nano (string)
//	ARGV[6] = ttl seconds
//
// Returns 1 (allowed) oder 0 (denied). Refund auf global Bucket bei
// project-deny ist in der Lua-Logic eingebaut.
const redisIssuanceLuaScript = `
local function consume(key, capacity, refill, now, ttl)
    if capacity == 0 and refill == 0 then
        return true, capacity
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
    if tokens < 1.0 then
        redis.call('HSET', key, 'tokens', tostring(tokens), 'last_at', tostring(now))
        redis.call('EXPIRE', key, ttl)
        return false, tokens
    end
    tokens = tokens - 1.0
    redis.call('HSET', key, 'tokens', tostring(tokens), 'last_at', tostring(now))
    redis.call('EXPIRE', key, ttl)
    return true, tokens
end

local global_key = KEYS[1]
local project_key = KEYS[2]
local global_cap = tonumber(ARGV[1])
local global_refill = tonumber(ARGV[2])
local project_cap = tonumber(ARGV[3])
local project_refill = tonumber(ARGV[4])
local now = tonumber(ARGV[5])
local ttl = tonumber(ARGV[6])

local g_ok, g_tokens = consume(global_key, global_cap, global_refill, now, ttl)
if not g_ok then
    return 0
end
local p_ok, _ = consume(project_key, project_cap, project_refill, now, ttl)
if not p_ok then
    -- Refund global Bucket nur, wenn er nicht deaktiviert ist
    -- (sonst würde der Refund einen Token "erzeugen").
    if not (global_cap == 0 and global_refill == 0) then
        local refunded = math.min(global_cap, g_tokens + 1.0)
        redis.call('HSET', global_key, 'tokens', tostring(refunded), 'last_at', tostring(now))
        redis.call('EXPIRE', global_key, ttl)
    end
    return 0
end
return 1
`

// RedisIssuanceLimiterConfig bündelt Bucket- und Fail-Mode-Parameter.
type RedisIssuanceLimiterConfig struct {
	GlobalCapacity      int
	GlobalRefillPerSec  float64
	ProjectCapacity     int
	ProjectRefillPerSec float64
	TTLSeconds          int
	FailOpen            bool
	KeyPrefix           string // default "mtrace:issuance"
	Now                 func() time.Time
}

// NewRedisIssuanceRateLimiter konstruiert den Adapter und lädt das
// Lua-Script per `SCRIPT LOAD` (EVALSHA-Pfad). Wenn das Skript zum
// Eval-Zeitpunkt aus dem Redis-Cache gefallen ist, fällt der Adapter
// auf `EVAL` zurück.
func NewRedisIssuanceRateLimiter(client redis.UniversalClient, cfg RedisIssuanceLimiterConfig, logger *slog.Logger) (*RedisIssuanceRateLimiter, error) {
	if client == nil {
		return nil, fmt.Errorf("redis-issuance-limiter: client is nil")
	}
	prefix := cfg.KeyPrefix
	if prefix == "" {
		prefix = "mtrace:issuance"
	}
	ttl := cfg.TTLSeconds
	if ttl <= 0 {
		ttl = 24 * 3600
	}
	nowFn := cfg.Now
	if nowFn == nil {
		nowFn = time.Now
	}
	l := &RedisIssuanceRateLimiter{
		client:     client,
		now:        nowFn,
		globalCfg:  bucketConfig{Capacity: cfg.GlobalCapacity, RefillPerSecond: cfg.GlobalRefillPerSec},
		projectCfg: bucketConfig{Capacity: cfg.ProjectCapacity, RefillPerSecond: cfg.ProjectRefillPerSec},
		ttlSeconds: ttl,
		failOpen:   cfg.FailOpen,
		logger:     logger,
		keyPrefix:  prefix,
	}
	if cfg.FailOpen {
		l.fallback = NewInMemoryIssuanceRateLimiter(
			cfg.GlobalCapacity, cfg.GlobalRefillPerSec,
			cfg.ProjectCapacity, cfg.ProjectRefillPerSec,
		)
	}
	// `SCRIPT LOAD` ist best-effort beim Boot; bei Outage fallen wir
	// im Eval auf den Inline-EVAL-Pfad zurück.
	if sha, err := client.ScriptLoad(context.Background(), redisIssuanceLuaScript).Result(); err == nil {
		l.scriptSHA = sha
	} else if logger != nil {
		logger.Warn("redis-issuance-limiter: SCRIPT LOAD failed; falling back to inline EVAL",
			"error", err.Error())
	}
	return l, nil
}

// Allow prüft beide Buckets atomar via Lua-Script.
func (l *RedisIssuanceRateLimiter) Allow(ctx context.Context, projectID string, projectBucket domain.RateLimitBucket) (bool, error) {
	if l == nil {
		return true, nil
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}
	now := l.now()
	globalKey := l.keyPrefix + ":global"
	projectKey := l.keyPrefix + ":project:" + projectID
	pcfg := l.resolveProjectConfig(projectBucket)

	args := []any{
		l.globalCfg.Capacity,
		l.globalCfg.RefillPerSecond,
		pcfg.Capacity,
		pcfg.RefillPerSecond,
		strconv.FormatInt(now.UnixNano(), 10),
		l.ttlSeconds,
	}
	res, err := l.evalScript(ctx, []string{globalKey, projectKey}, args)
	if err != nil {
		return l.handleRedisError(ctx, projectID, projectBucket, err)
	}
	allowed, ok := res.(int64)
	if !ok {
		return l.handleRedisError(ctx, projectID, projectBucket,
			fmt.Errorf("unexpected redis result type %T", res))
	}
	return allowed == 1, nil
}

// evalScript ruft das Lua-Script per EVALSHA und fällt bei
// `NOSCRIPT`-Error auf `EVAL` zurück.
func (l *RedisIssuanceRateLimiter) evalScript(ctx context.Context, keys []string, args []any) (any, error) {
	if l.scriptSHA != "" {
		res, err := l.client.EvalSha(ctx, l.scriptSHA, keys, args...).Result()
		if err == nil {
			return res, nil
		}
		// Bei NOSCRIPT-Fehler: Script ist aus dem Redis-Cache gefallen,
		// EVAL re-uploaded es implizit.
		if !isNoScriptError(err) {
			return nil, err
		}
	}
	return l.client.Eval(ctx, redisIssuanceLuaScript, keys, args...).Result()
}

// handleRedisError implementiert den Fail-Mode-Vertrag.
func (l *RedisIssuanceRateLimiter) handleRedisError(ctx context.Context, projectID string, projectBucket domain.RateLimitBucket, err error) (bool, error) {
	if l.logger != nil {
		l.logger.Warn("redis-issuance-limiter outage",
			"error", err.Error(),
			"fail_mode", failModeName(l.failOpen),
		)
	}
	if l.failOpen && l.fallback != nil {
		return l.fallback.Allow(ctx, projectID, projectBucket)
	}
	// fail-closed: deny ohne Fehler-Propagation an den Caller,
	// damit der HTTP-Handler 429 (statt 500) liefert.
	return false, nil
}

func (l *RedisIssuanceRateLimiter) resolveProjectConfig(override domain.RateLimitBucket) bucketConfig {
	if override.IsZero() {
		return l.projectCfg
	}
	return bucketConfig{Capacity: override.Capacity, RefillPerSecond: override.RefillPerSecond}
}

// isNoScriptError erkennt den Redis-`NOSCRIPT`-Error.
// go-redis liefert das als Error mit `NOSCRIPT`-Prefix.
func isNoScriptError(err error) bool {
	return err != nil && strings.HasPrefix(err.Error(), "NOSCRIPT")
}

func failModeName(failOpen bool) string {
	if failOpen {
		return "fail-open (local memory fallback)"
	}
	return "fail-closed (deny on outage)"
}

// Compile-time check.
var _ driven.IssuanceRateLimiter = (*RedisIssuanceRateLimiter)(nil)
