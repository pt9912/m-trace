package ratelimit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/pt9912/m-trace/apps/api/adapters/driven/redisutil"
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
// Ein abgebrochener Context wird wie eine Outage ENTSCHIEDEN (Fallback
// bzw. Deny), aber NICHT als Outage GEWERTET: `ctx.Err()` durch den Port
// zu reichen würde an der Call-Site als rate-limited gezählt und als 500
// enden — und ein Client-Disconnect darf weder das Degraded-Signal
// flippen noch False-Outage-WARNs erzeugen (sonst verbraucht ein
// Cancellation-Rauschen die „log once"-Kante eines echten Ausfalls).
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
//	ARGV[4] = now in unix-MILLIsekunden (Client-Uhr der Replica) —
//	          bewusst nicht Nanosekunden wie in den Auth-Limiter-Scripts:
//	          ~1.7e18 ns überschreitet Luas exakten Double-Bereich (2^53)
//	          und tostring() (%.14g) quantisiert auf ~100µs; Millis
//	          (~1.7e12) bleiben bis weit übers Jahr 2200 exakt.
//	ARGV[5] = ttl seconds
//
// Uhren-Drift-Schutz (die Replicas liefern `now` von N Uhren): elapsed < 0
// zählt als 0 (kein Token-Fressen) UND `last_at` regressiert nie — gespeichert
// wird max(stored_last_at, now). Ohne das zweite würde eine nachgehende
// Replica `last_at` zurücksetzen und die vorgehende bekäme die Skew-Differenz
// bei jedem Wechsel erneut als Refill gutgeschrieben (systematische
// Inflation auf dem Hot-Path); so bleibt Skew ein einmaliger, begrenzter
// Effekt. Ein partiell beschädigter Hash (tokens ODER last_at fehlt, z. B.
// durch manuelle Ops-Eingriffe) wird als frischer Bucket behandelt statt
// mit einem Lua-Fehler jeden Request auf dem Key zu killen.
//
// Der Deny-Pfad schreibt NICHT (nur EXPIRE auf bestehende Keys): der
// Refill-Stand ist aus den unveränderten Werten rekonstruierbar, und das
// erspart dem heißesten Pfad (saturierender Noisy) die HSET-Writes.
// Unter Skew ist das zusätzlich konservativer als Schreiben (nie
// inflationär). EXPIRE bleibt, damit ein dauersaturierter Bucket nicht
// mitten im Ansturm expired und voll zurückkommt.
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
local allowed = 1
for i, key in ipairs(KEYS) do
    local stored = redis.call('HMGET', key, 'tokens', 'last_at')
    local t = tonumber(stored[1])
    local l = tonumber(stored[2])
    if t == nil or l == nil then
        t = capacity
        l = now
    end
    if l < now then
        t = math.min(capacity, t + ((now - l) / 1000.0) * refill)
        l = now
    end
    if t < n then
        allowed = 0
    end
    tokens[i] = t
    last[i] = l
end

for i, key in ipairs(KEYS) do
    if allowed == 1 then
        redis.call('HSET', key, 'tokens', tostring(tokens[i] - n), 'last_at', tostring(last[i]))
    end
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
		// Client weg — Fail-Mode-Entscheidung ohne Outage-Wertung.
		return l.failDecision(ctx, key, n)
	}
	args := []any{
		n,
		l.capacity,
		l.refill,
		strconv.FormatInt(l.now().UnixMilli(), 10),
		l.ttlSec,
	}
	res, err := redisutil.Eval(ctx, l.client, l.scriptSHA, redisIngestLuaScript, keys, args)
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

// handleRedisError ist der Outage-Pfad: einmaliges WARN beim Eintritt
// (kein Flood — Allow läuft pro Batch-Request), dann Fail-Mode-Politik.
// Context-Abbrüche (der Request ist tot, Redis ist gesund) zählen NICHT
// als Outage — sie würden das Degraded-Signal verfälschen und die
// „log once"-Kante eines echten Ausfalls verbrauchen.
func (l *RedisTokenBucketRateLimiter) handleRedisError(ctx context.Context, key driven.RateLimitKey, n int, err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return l.failDecision(ctx, key, n)
	}
	if l.degraded.CompareAndSwap(false, true) && l.logger != nil {
		l.logger.Warn("redis-ingest-limiter outage — degraded mode begins",
			"error", err.Error(),
			"fail_mode", redisutil.FailModeLabel(!l.failClosed),
		)
	}
	return l.failDecision(ctx, key, n)
}

// failDecision wendet die Fail-Mode-Politik an: fail-closed → Deny,
// fail-open → lokaler In-Memory-Fallback. Der Konstruktor garantiert
// fallback != nil genau dann, wenn !failClosed.
func (l *RedisTokenBucketRateLimiter) failDecision(ctx context.Context, key driven.RateLimitKey, n int) error {
	if l.failClosed {
		return domain.ErrRateLimited
	}
	return l.fallback.Allow(ctx, key, n)
}

// markRecovered loggt das Ende einer Degradations-Phase genau einmal —
// im fail-open-Modus war währenddessen die repliken-übergreifende
// Fairness pausiert (per-Replica-Fallback), das soll nicht still enden.
// Der Load-Guard hält den Hot-Path frei von gelockten RMW-Ops: das CAS
// läuft nur, wenn tatsächlich eine Degradation zu beenden ist.
func (l *RedisTokenBucketRateLimiter) markRecovered() {
	if !l.degraded.Load() {
		return
	}
	if l.degraded.CompareAndSwap(true, false) && l.logger != nil {
		l.logger.Warn("redis-ingest-limiter recovered — degraded mode ends",
			"fail_mode", redisutil.FailModeLabel(!l.failClosed),
		)
	}
}

// redisBucketKeys bildet die Redis-Keys der nicht-leeren Dimensionen.
// project_id ist längenbegrenzt/validiert und geht raw in den Key;
// client_ip ist auf dem XFF-Trust-Pfad nur deshalb bounded, weil
// xffClientIP (HTTP-Adapter) ausschließlich per net.ParseIP validierte
// Werte durchlässt — reißt diese Invariante, wird der Key unbounded
// client-kontrollierbar. NUR der Origin wird gehasht: der Header ist
// client-kontrolliert und unbegrenzt lang, der Hash bounded die
// Key-Länge (Namensgebung analog RAK-90 „Origin-Header-Hash").
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

// Compile-time check.
var _ driven.RateLimiter = (*RedisTokenBucketRateLimiter)(nil)
