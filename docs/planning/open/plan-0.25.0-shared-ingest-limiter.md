# Implementation Plan — `0.25.0` Repliken-übergreifend fairer Ingest-Limiter (R-26 b)

> **Status**: **Skizze (2026-07-13, ungefirmt)** — Owner-Review ausstehend; die
> offenen Fragen in §8 sind Owner-Calls. Die Versionsnummer `0.25.0` ist
> provisorisch (Plan-Identifier). Hinweis Release-Zuschnitt: der nächste
> getaggte Release nimmt den bereits gelieferten Cutover
> ([`plan-0.24.0`](../in-progress/plan-0.24.0-sqlite-postgres-cutover.md),
> CHANGELOG `[Unreleased]`) mit.
>
> **Bezug**: **R-26 b** in [`risks-backlog.md`](../in-progress/risks-backlog.md)
> (einziger offener Teil von R-26; a/c belegt); Messbeleg der Lücke in
> [`docs/perf/budgets.md`](../../perf/budgets.md) §8 (throttled 2,01× über 2
> Replicas = Limiter liegt pro Replica); RAK-129 (lässt R-26 b explizit
> offen); RAK-90/R-22 (Origin-Limiter: das zu spiegelnde Redis-Adapter-Muster);
> RAK-88/R-17 (Issuance-Redis, geteilter Redis-Server); RAK-74/F-113
> (Ingest-Policies; Projekt-Buckets s. §8.3);
> [ADR-0005](../../adr/0005-production-ops-backends.md) (Ops-Backends),
> [ADR-0006](../../adr/0006-postgres-scaleout-adapter.md) (Scale-out-Kontext).

## 1. Ziel

Das **Per-Projekt-Ingest-Rate-Limit soll über N API-Replicas als EIN
gemeinsames Budget wirken**. Heute ist der Ingest-Limiter ein In-Process-
Token-Bucket **pro Replica** → die effektive Per-Projekt-Decke ist
`N × Capacity` und Fairness gilt nicht repliken-übergreifend. Das ist seit
dem 0.23.0-Scale-out-Lasttest **gemessen** (budgets.md §8: throttled skaliert
1→2 Replicas linear 2,01×, Flaschenhals „Per-Replica-In-Memory-Limiter").

Ziel-Evidenz (Inversion des Befunds): derselbe throttled Lauf mit shared
Limiter skaliert **~1,0×** — plus Multi-Tenant-Nachweis (Per-Projekt-Isolation,
Noisy-Neighbor) unter Scale-out.

Nicht-Ziel: Default-Deployments ändern. Der In-Memory-Limiter bleibt Default;
Redis bleibt **optional** (keine neue Pflichtabhängigkeit — konsistent mit
RAK-90/RAK-88, wo Redis ebenfalls opt-in ist).

## 2. Scope / Abgrenzung

**In Scope:**

- **Redis-Adapter** für den bestehenden Driven-Port `driven.RateLimiter`
  (**port-erhaltend**, wie beim R-28-Sequencer: Call-Sites unangetastet).
- **ENV-Selektor** `MTRACE_RATE_LIMIT_BACKEND=memory|redis` (Default
  `memory`), Muster = `buildOriginRateLimiter`/`buildIssuanceRateLimiter`.
- **Multi-Tenant-Lab-Tooling**: env-getriebenes Seeding von N Lab-Projekten
  (additiv zum `demo`-Default, byte-stabil ohne ENV), k6-Token-Fan-out,
  eigene `make`-Modes (aus der R-26-Machbarkeitsnotiz, dort Teil (b)).
- **Fairness-Nachweis unter Scale-out**: Erweiterung `smoke-scaleout-load`
  (Redis im Compose-Lab, backend=redis-Modus) + Messdokumentation in
  budgets.md; **R-26 b schließen**.

**Nicht in Scope:**

- **Policy-getriebene Per-Projekt-Buckets** (RAK-74-Anschluss): die heutigen
  uniformen ENV-Caps (`MTRACE_RATE_LIMIT_CAPACITY`/`_REFILL`, gleich für alle
  Projekte) bleiben — s. Owner-Frage §8.3.
- **Redis-Cluster-Betrieb**: die Multi-Key-Lua-Scripts setzen (wie schon das
  Issuance-Script mit global+projekt-Key) einen Single-Redis voraus;
  Cluster-Tauglichkeit (Hash-Tags) ist Folge-Scope.
- **TLS/mTLS für `MTRACE_REDIS_*`** — existiert heute für Origin/Issuance
  ebenso nicht; separater Ops-Scope, wird hier nicht nebenbei eingeführt.
- **Durchsatz-Scaling jenseits Single-Postgres** (budgets.md §8 „Konsequenz")
  — eigener Folge-Scope, unabhängig von Fairness.

## 3. Ist-Stand (kartiert 2026-07-13)

**Ingest-Limiter (in-process):**

- Port: `apps/api/hexagon/port/driven/rate_limiter.go` —
  `RateLimiter.Allow(ctx, key RateLimitKey, n int) error`;
  `RateLimitKey{ProjectID, ClientIP, Origin}`. Rückgabe nur `error`
  (`domain.ErrRateLimited`), anders als Origin/Issuance (`(bool, error)`).
- Adapter: `apps/api/adapters/driven/ratelimit/token_bucket.go` — drei
  unabhängige Bucket-Dimensionen (project/ip/origin), **all-or-nothing**
  (erst alle prüfen, dann alle abbuchen, unter einem Mutex).
- Aufruf **per Batch** (`n = len(in.Events)`):
  `apps/api/hexagon/application/register_playback_event_batch.go:256`, nach
  Token-/Origin-Check, vor Schema-Validierung; bei Deny
  `metrics.RateLimitedEvents(n)` → `mtrace_rate_limited_events_total`.
- HTTP-Mapping: 429 mit statischem `Retry-After: 1`
  (`adapters/driving/http/handler.go:284`).
- Konfig: `MTRACE_RATE_LIMIT_CAPACITY` (Default 100) /
  `MTRACE_RATE_LIMIT_REFILL` (Default 100/s), geparst in `main.go`
  (`parseIngestRateLimit`); Konstruktion unbedingt in `buildHandler`
  (`main.go:353`) — **kein Backend-Selektor**.

**Wiederverwendbares Redis-Muster (Origin R-22 / Issuance R-17):**

- `apps/api/adapters/driven/auth/redis_origin_rate_limiter.go` und
  `redis_issuance_rate_limiter.go`: **atomarer Token-Bucket via Lua**
  (Hash `tokens`+`last_at`, Refill = `elapsed × refill` capped, `EXPIRE`),
  `SCRIPT LOAD`→`EVALSHA` mit `EVAL`-Fallback bei `NOSCRIPT`. Zeitquelle ist
  **Client-Zeit als ARGV** (injizierbare Uhr → deterministische Tests);
  Origin-Limiter läuft damit heute schon repliken-übergreifend.
- Fail-Mode-Vertrag: **fail-closed default** (Redis-Fehler → deny, nie 500),
  **fail-open opt-in** (`MTRACE_AUTH_ISSUANCE_FAIL_OPEN`) mit Delegation an
  In-Memory-Fallback pro Replica.
- Client: `go-redis/v9`, `buildRedisClient()` (`main.go:1034`) mit
  `MTRACE_REDIS_ADDR`/`_AUTH`/`_DB` — geteilter Redis-Server, Key-Prefixe
  `mtrace:origin:` / `mtrace:issuance:…`.
- Testharness: **miniredis in-process** (`miniredis.RunT`, kein CI-Service),
  inkl. Cross-Instance-Sharing- und Outage-Tests
  (`redis_*_rate_limiter_test.go`); Smoke `make smoke-issuance-multi-host`.
- **Einschränkungen der bestehenden Scripts** für unseren Fall: sie
  konsumieren fix **1 Token pro Call** (Batch-`n` muss in ARGV) und decken
  1–2 Buckets ab (wir brauchen bis zu 3, all-or-nothing).

**Load-Tooling (Lücken für Multi-Tenant):**

- `scripts/load/playback-events.k6.js`: **ein** Token/Projekt
  (`PROJECT_TOKEN`, Default `demo-token`); Profile `closed`/`open`
  (`LOAD_PROFILE`, `TARGET_EVENT_RATE`, `P95_BUDGET_MS`); Zähler
  `mtrace_events_sent/_accepted/_rate_limited/_rejected`.
- `scripts/smoke-scaleout-load.sh` + `docker-compose.scaleout.yml`:
  2 Replicas + geteilter Postgres + nginx-LB, psql-Readback
  (`persisted==accepted`, `distinct==COUNT(*)`); **kein Redis** im Lab,
  single-token.
- Lab-Seeding: `projectConfigs` in `main.go:320` — **hartkodiert 1 Projekt**
  (`demo`); kein env-getriebenes N-Projekt-Seeding.

## 4. Design

### 4.1 Adapter `ratelimit.RedisTokenBucketRateLimiter`

Neuer Adapter im bestehenden Package `adapters/driven/ratelimit`, implementiert
`driven.RateLimiter` unverändert (`var _`-Assertion). **Ein** Lua-Script pro
`Allow`:

- `KEYS[1..k]` = Bucket-Hashes der **nicht-leeren** Dimensionen
  (`mtrace:ingest:project:<id>`, `mtrace:ingest:ip:<ip>`,
  `mtrace:ingest:origin:<hash>`), `k ∈ 1..3`.
- `ARGV` = `n` (Batch-Größe — neu ggü. Origin/Issuance), `capacity`,
  `refill`, `now` (unix-nano, Client-Uhr wie im etablierten Muster),
  `ttl`.
- Semantik identisch zum In-Memory-Adapter: **alle Dimensionen prüfen, dann
  alle abbuchen** (all-or-nothing, atomar durch das eine Script — kein
  Refund-Tanz nötig); Deny → Rückgabe 0 → Adapter mappt auf
  `domain.ErrRateLimited`. Leerer Key (alle Dimensionen leer) bleibt No-op
  wie im In-Memory.
- Refill-Klausel gegen Uhren-Drift der Replicas: `elapsed < 0` auf 0 clampen
  (NTP-Skew darf Tokens weder schenken noch fressen).

### 4.2 Wiring + Boot-Validation

`buildIngestRateLimiter` in `main.go`, Spiegel von `buildOriginRateLimiter`:
Switch auf `MTRACE_RATE_LIMIT_BACKEND` (`memory` Default | `redis`);
unbekannte Werte lehnt der Boot-Validator mit präziser Begründung ab
(RAK-90-Stil, inkl. „warum kein `sqlite`": nicht Multi-Host-tauglich für
einen Hot-Path-Bucket). `redis` nutzt `buildRedisClient()` und damit den
bestehenden `MTRACE_REDIS_*`-Block (geteilter Server mit Origin/Issuance,
eigener Key-Prefix `mtrace:ingest:`).

### 4.3 Fail-Mode (Owner-Call, §8.1)

Redis wird mit `backend=redis` zur **Hot-Path-Abhängigkeit des Ingest**.
Optionen bei Redis-Ausfall:

- **fail-closed** (Präzedenz Origin/Issuance): deny → 429. Schützt das
  Budget strikt, kostet aber **Telemetrie-Verfügbarkeit** (Events werden
  client-seitig nach Retries verworfen).
- **fail-open auf In-Memory-Fallback**: Degradation **exakt auf den heutigen
  Zustand** (per-Replica-Bucket) — verliert vorübergehend die
  repliken-übergreifende Fairness, nie die Limitierung selbst und nie
  Verfügbarkeit. Eintritt/Austritt als WARN-Log sichtbar machen.

**Empfehlung**: fail-open-to-memory als Default für den *Ingest*-Limiter
(anderes Schutzgut als Auth-Flutung bei Origin/Issuance), fail-closed als
opt-in-Flag. Entscheidung beim Owner (§8.1).

### 4.4 Performance-Betrachtung

Ein `EVALSHA` **pro Batch-Request** (heute: 0 I/O). Bei der gemessenen
Referenzdecke (~12k ev/s, Batch 20) sind das ~600 Script-Calls/s — weit
unter Redis-Kapazität; Latenzbeitrag sub-ms bis low-ms im 1s-p95-Budget
(NF-20). Der Default-Pfad (`memory`) bleibt byte-identisch; der
store-gebundene unthrottled Fall läuft ohnehin mit hochgesetzten Caps.
Neuer Vorbehalt für budgets.md: bei app-gebundener Last mit `redis` ist der
**Single-Redis die neue geteilte Decke** — im Nachweis-Lauf attribuieren
(docker stats), analog zur §8-Ehrlichkeit beim Single-Postgres.

### 4.5 Observability

`mtrace_rate_limited_events_total` bleibt unverändert (der Use-Case
emittiert, nicht der Adapter). Zusätzlich nur: WARN-Log bei
Fallback-Eintritt/-Austritt (fail-open) bzw. bei Redis-Fehlern (fail-closed),
damit ein Fairness-/Verfügbarkeitsverlust nicht still ist.

## 5. Tranchen

**T1 — Redis-Adapter + Selektor + Tests.** Adapter (§4.1), Wiring +
Boot-Validation (§4.2), Fail-Mode gemäß §8.1-Entscheidung. Tests via
miniredis (Muster `redis_issuance_rate_limiter_test.go`): Cross-Instance-
Sharing (zwei Adapter-Instanzen, ein Budget — der eigentliche R-26-b-Kern als
Unit-Beleg), n-Token-Batch, all-or-nothing über 3 Dimensionen,
Capacity/Refill-Verhalten, Outage (fail-closed und fail-open), Context-Cancel,
Empty-Key-No-op. Läuft in `make gates` (in-process, kein Service). ENV-Doku.

**T2 — Multi-Tenant-Lab-Tooling (single-instance).** Env-getriebenes
N-Projekt-Seeding (additiv zu `demo`, byte-stabiler Default ohne ENV);
k6-Token-Fan-out (N Tokens, per-Projekt-Zähler via Tags); `make`-Mode
(z. B. `smoke-load MODE=multi-tenant`): Per-Projekt-Isolation +
Noisy-Neighbor auf **einer** Instanz (Baseline, limiter-backend-unabhängig).
Schwellen für das Noisy-Neighbor-Gate hier festlegen (Opfer-Projekt:
p95-Delta + accepted-Rate-Minimum).

**T3 — Scale-out-Fairness-Nachweis.** `docker-compose.scaleout.yml` + 
`smoke-scaleout-load.sh` um Redis-Service + `backend=redis`-Modus erweitern.
Messungen (Setup identisch budgets.md §8: 50 VUs / 60 s / Batch 20):
(a) **Fairness-Inversion**: throttled 1→2 Replicas ≈ **1,0×** statt 2,01×;
(b) **Noisy-Neighbor über Replicas**: Projekt A saturiert (429s), Projekt B
hält p95/accepted innerhalb der T2-Schwellen; (c) Korrektheits-Gates
unverändert (`persisted==accepted`, `distinct==COUNT(*)`). Ergebnisse als
budgets.md-§9 dokumentieren (inkl. Redis-Attribution §4.4).

**T4 — Closeout.** R-26 → 🟢 (b geschlossen, letzter offener Teil),
roadmap Schritt 57, CHANGELOG `[Unreleased]`; Release-Zuschnitt +
Lastenheft-Patch (neue RAK-Gruppe analog RAK-126..130) gemäß §8.4;
ADR-Frage gemäß §8.5.

## 6. DoD

- **Fairness-Inversion gemessen**: throttled Skalierung 1→2 Replicas mit
  `backend=redis` ≤ **1,15×** (statt 2,01×) im §8-Referenz-Setup; Zahl in
  budgets.md.
- **Noisy-Neighbor-Gate grün** (Schwellen aus T2, über 2 Replicas).
- **Korrektheits-Gates unverändert grün** (kein Verlust, keine Duplikate).
- **Default byte-identisch**: ohne `MTRACE_RATE_LIMIT_BACKEND` ändert sich
  nichts (memory, kein Redis nötig); Boot-Validator-Fehlertexte präzise.
- **Cross-Instance-Sharing als Unit-Beleg** (miniredis) in `make gates`;
  Outage-Fail-Modes getestet.
- Doku: ENV-Referenz, budgets.md-§9, risks-backlog (R-26 🟢), CHANGELOG.

## 7. Risiken

- **Redis als Hot-Path-SPOF** (nur bei `backend=redis`): Fail-Mode-Politik
  §8.1; Degradationsverhalten dokumentiert + getestet, WARN-Logs.
- **Multi-Key-Lua vs. Redis Cluster**: Keys verschiedener Dimensionen landen
  in verschiedenen Hash-Slots → Script bricht im Cluster-Mode. Gleiches gilt
  heute schon fürs Issuance-Script; Single-Redis-Annahme wird explizit
  dokumentiert (Nicht-Ziel §2).
- **Uhren-Drift der Replicas** (Client-`now` als ARGV, etabliertes Muster):
  für 1s-Granularität-Buckets unkritisch bei NTP; negative elapsed werden
  geclampt (§4.1). Alternative (Redis-`TIME`) würde die deterministische
  Testbarkeit (injizierte Uhr) kosten — bewusst nicht gewählt.
- **„Fairness" präzise halten**: der shared Bucket garantiert die **globale
  Per-Projekt-Decke** (FCFS über Replicas), keine Gleichverteilung pro
  Replica — die Nachweis-Formulierung (T3) misst die Decke und die
  Nachbar-Isolation, nicht Scheduling-Fairness.
- **Messmethodik k6**: per-Projekt-Metriken brauchen Tag-basierte Zähler;
  open-loop-Fan-out über N Tokens muss die Ziel-Rate pro Projekt (nicht
  global) treiben, sonst misst der Lauf das Falsche.

## 8. Offene Fragen (Owner-Calls)

- **§8.1 Fail-Mode-Default** für `backend=redis`: fail-open-to-memory
  (Empfehlung, §4.3) vs. fail-closed (Origin/Issuance-Präzedenz)? Eigenes
  Flag `MTRACE_RATE_LIMIT_FAIL_OPEN` oder (nicht empfohlen, anderes
  Schutzgut) Wiederverwendung von `MTRACE_AUTH_ISSUANCE_FAIL_OPEN`?
- **§8.2 ENV-Name/Default**: `MTRACE_RATE_LIMIT_BACKEND=memory|redis`,
  Default `memory` — ok? (Kein `disabled`: der Ingest-Limiter ist heute
  immer an, das bleibt.)
- **§8.3 Per-Projekt-Buckets** (RAK-74/F-113-Anschluss): uniforme ENV-Caps
  beibehalten (Empfehlung — Parität zum heutigen In-Memory-Verhalten,
  kleiner Scope) oder policy-getriebene Projekt-Buckets im selben Zug
  (Use-Case-/Policy-Resolution-Umbau, deutlich größer)?
- **§8.4 Release-Zuschnitt**: eigener Minor-Release (nimmt den wartenden
  Cutover aus `[Unreleased]` mit) + Lastenheft-Patch mit neuer RAK-Gruppe
  analog `0.23.0` (RAK-126..130)?
- **§8.5 ADR ja/nein**: Empfehlung **nein** — erweitert das bestehende
  Redis-Ops-Muster (ADR-0005, RAK-88/90) port-erhaltend auf einen weiteren
  Adapter; Entscheidung + Vorbehalte leben im Plan und in R-26. Alternativ
  ein kurzer ADR „Ingest-Limiter-Backend", falls die Hot-Path-Abhängigkeit
  ADR-würdig erscheint.
