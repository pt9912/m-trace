# Implementation Plan — `0.25.0` Repliken-übergreifend fairer Ingest-Limiter (R-26 b)

> **Status**: **Gefirmt (2026-07-13, Owner-Review); Tranche 1 GEBAUT
> (2026-07-13)** — die §8-Fragen sind
> entschieden (s. §8). Die Versionsnummer `0.25.0` ist bestätigt (§8.4):
> der Release nimmt den bereits gelieferten Cutover
> ([`plan-0.24.0`](plan-0.24.0-sqlite-postgres-cutover.md),
> CHANGELOG `[Unreleased]`) mit.
>
> **Bau-Amendment 2026-07-13 (T1).** `RedisTokenBucketRateLimiter`
> (`apps/api/adapters/driven/ratelimit/redis_token_bucket.go`) gemäß §4.1:
> EIN Lua-Script (n-Token, bis zu 3 Buckets, check-then-debit
> all-or-nothing, Deny persistiert nur den Refill-Stand), elapsed-Clamp +
> **monotones `last_at`** (`max(stored, now)`), Origin-Hash
> (SHA-256/hex, 32 Zeichen); EVALSHA→EVAL-Fallback. Fail-open-to-memory
> Default + `MTRACE_RATE_LIMIT_FAIL_CLOSED`-Opt-in (§8.1), Degradations-
> Ein-/Austritt je einmal WARN (§4.5); abgebrochener Context wird als
> Outage behandelt (Port kennt nur `error`; `ctx.Err()` würde an der
> Call-Site als rate-limited gezählt und als 500 enden). Wiring
> `buildIngestRateLimiter` + Boot-Validation (§4.2, `sqlite`/`memcached`
> explizit abgelehnt); Default `memory` byte-identisch (kein neuer
> Log-Output auf dem Default-Pfad). **10 miniredis-Tests grün** inkl.
> Cross-Instance-Sharing, **Skew-Test** (versetzte Uhren, exakt Capacity
> — keine Refill-Inflation) und Outage-Recovery (deckt zugleich den
> NOSCRIPT-Pfad). ENV-Doku als `docs/user/auth.md` §5.10 (inkl. der
> dokumentationspflichtigen gemischten Fail-Modi, §8.1).
>
> **Bezug**: **R-26 b** in [`risks-backlog.md`](risks-backlog.md)
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
  `memory`; entschieden, §8.2), Muster =
  `buildOriginRateLimiter`/`buildIssuanceRateLimiter`.
- **Multi-Tenant-Lab-Tooling**: env-getriebenes Seeding von N Lab-Projekten
  (additiv zum `demo`-Default, byte-stabil ohne ENV), k6-Token-Fan-out,
  eigene `make`-Modes (aus der R-26-Machbarkeitsnotiz, dort Teil (b)).
- **Fairness-Nachweis unter Scale-out**: Erweiterung `smoke-scaleout-load`
  (Redis im Compose-Lab, backend=redis-Modus) + Messdokumentation in
  budgets.md; **R-26 b schließen**.

**Nicht in Scope:**

- **Policy-getriebene Per-Projekt-Buckets** (RAK-74-Anschluss): die heutigen
  uniformen ENV-Caps (`MTRACE_RATE_LIMIT_CAPACITY`/`_REFILL`, gleich für alle
  Projekte) bleiben (entschieden, §8.3).
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
- Key-Bildung: `project_id` und `client_ip` sind längenbegrenzt/validiert
  und gehen raw in den Key; **nur Origin wird gehasht** — der Header ist
  client-kontrolliert und unbegrenzt lang, der Hash bounded die Key-Länge
  (fester Hash, z. B. SHA-256/hex verkürzt; Namensgebung analog der
  RAK-90-Formulierung „`Origin`-Header-Hash" — der bestehende
  Origin-Limiter selbst keyed auf Client-IP, gibt also kein Muster vor).
- `ARGV` = `n` (Batch-Größe — neu ggü. Origin/Issuance), `capacity`,
  `refill`, `now` (unix-nano, Client-Uhr wie im etablierten Muster),
  `ttl`.
- Semantik identisch zum In-Memory-Adapter: **alle Dimensionen prüfen, dann
  alle abbuchen** (all-or-nothing, atomar durch das eine Script — kein
  Refund-Tanz nötig); Deny → Rückgabe 0 → Adapter mappt auf
  `domain.ErrRateLimited`. Leerer Key (alle Dimensionen leer) bleibt No-op
  wie im In-Memory.
- Refill-Klausel gegen Uhren-Drift der Replicas (Client-`now` kommt von N
  Uhren): **(a)** `elapsed < 0` auf 0 clampen (kein Token-Fressen) **und
  (b)** `last_at` darf **nie regressieren** — gespeichert wird
  `max(stored_last_at, now)`, nicht `now`. (b) ist der wesentliche Teil:
  das etablierte Lua-Muster schreibt `last_at = now` unconditional (auch im
  Deny-Pfad) — eine Replica mit nachgehender Uhr regressiert damit
  `last_at`, und der nächste Call der vorgehenden Replica bekommt die
  Skew-Differenz **erneut** als Refill gutgeschrieben. Bei alternierenden
  Calls ist das systematische Refill-Inflation (~`skew × refill` Extra-Tokens
  pro Slow→Fast-Wechsel); bei ~600 Calls/s Hot-Path-Frequenz (§4.4) kann
  schon 10 ms Skew ein Mehrfaches des intendierten Budgets schenken —
  für Origin/Issuance (Auth-Frequenz) tolerierbar, hier material. Mit
  (a)+(b) ist Skew ein **einmaliger, begrenzter** Effekt statt
  kontinuierlicher Inflation.

### 4.2 Wiring + Boot-Validation

`buildIngestRateLimiter` in `main.go`, Spiegel von `buildOriginRateLimiter`:
Switch auf `MTRACE_RATE_LIMIT_BACKEND` (`memory` Default | `redis`;
Name entschieden, §8.2); unbekannte Werte lehnt der
Boot-Validator mit präziser Begründung ab (RAK-90-Stil: `sqlite` und
`memcached` werden — wie in den beiden bestehenden Limiter-Validatoren —
explizit mit Begründungs- bzw. „Folge-Item"-Wording abgelehnt,
symmetrische Fehlertexte). `redis` nutzt `buildRedisClient()` und damit den
bestehenden `MTRACE_REDIS_*`-Block (geteilter Server mit Origin/Issuance,
eigener Key-Prefix `mtrace:ingest:`).

### 4.3 Fail-Mode (entschieden: fail-open-to-memory, §8.1)

Redis wird mit `backend=redis` zur **Hot-Path-Abhängigkeit des Ingest**.
Optionen bei Redis-Ausfall:

- **fail-closed** (Präzedenz Origin/Issuance): deny → 429. Schützt das
  Budget strikt, kostet aber **Telemetrie-Verfügbarkeit** (Events werden
  client-seitig nach Retries verworfen).
- **fail-open auf In-Memory-Fallback**: Degradation **exakt auf den heutigen
  Zustand** (per-Replica-Bucket) — verliert vorübergehend die
  repliken-übergreifende Fairness, nie die Limitierung selbst und nie
  Verfügbarkeit. Eintritt/Austritt als WARN-Log sichtbar machen.

**Entschieden (Owner-Review 2026-07-13, §8.1)**: fail-open-to-memory ist
Default für den *Ingest*-Limiter (anderes Schutzgut als Auth-Flutung bei
Origin/Issuance); fail-closed als Opt-in via
`MTRACE_RATE_LIMIT_FAIL_CLOSED=1`. `MTRACE_AUTH_ISSUANCE_FAIL_OPEN` wird
**nicht** wiederverwendet — die bewusste Abweichung vom geteilten
Fail-Mode-Schalter der Auth-Limiter (`main.go:1090`) inkl. der daraus
folgenden gemischten Fail-Modi auf demselben Redis ist Teil der
Entscheidung (volle Abwägung in §8.1) und gehört in die Operator-Doku.

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
Boot-Validation (§4.2), Fail-Mode gemäß §4.3 (fail-open-to-memory Default
+ `MTRACE_RATE_LIMIT_FAIL_CLOSED`-Opt-in). Tests via
miniredis (Muster `redis_issuance_rate_limiter_test.go`): Cross-Instance-
Sharing (zwei Adapter-Instanzen, ein Budget — der eigentliche R-26-b-Kern als
Unit-Beleg), n-Token-Batch, all-or-nothing über 3 Dimensionen,
Capacity/Refill-Verhalten, **Uhren-Drift** (zwei Adapter-Instanzen mit
gegeneinander versetzten injizierten Uhren, alternierende Calls →
Assertion: kein Refill-Überschuss über das Budget; **nur hier belegbar** —
der T3-Compose-Lab-Nachweis läuft auf einem Host mit einer Uhr und kann
Skew-Inflation prinzipiell nicht fangen), Outage (fail-closed und
fail-open), Context-Cancel, Empty-Key-No-op. Läuft in `make gates`
(in-process, kein Service). ENV-Doku.

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
roadmap Schritt 57, CHANGELOG `[Unreleased]`; Release `0.25.0` inkl.
wartendem Cutover + Lastenheft-Patch (neue RAK-Gruppe analog
RAK-126..130) gemäß §8.4; kein ADR (§8.5).

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
  ohne Gegenmaßnahme nicht nur Rausch-, sondern **Inflations**-Risiko — das
  unconditional `last_at = now` des etablierten Musters schenkt der
  vorgehenden Replica die Skew-Differenz wiederholt als Refill (§4.1).
  Mitigiert durch elapsed-Clamp **+ monotones `last_at`** (§4.1) →
  einmaliger, begrenzter Effekt; belegt per miniredis-Skew-Test in T1
  (T3 kann es nicht fangen, eine Host-Uhr). Alternative (Redis-`TIME`)
  würde die deterministische Testbarkeit (injizierte Uhr) kosten —
  bewusst nicht gewählt.
- **„Fairness" präzise halten**: der shared Bucket garantiert die **globale
  Per-Projekt-Decke** (FCFS über Replicas), keine Gleichverteilung pro
  Replica — die Nachweis-Formulierung (T3) misst die Decke und die
  Nachbar-Isolation, nicht Scheduling-Fairness.
- **Messmethodik k6**: per-Projekt-Metriken brauchen Tag-basierte Zähler;
  open-loop-Fan-out über N Tokens muss die Ziel-Rate pro Projekt (nicht
  global) treiben, sonst misst der Lauf das Falsche.

## 8. Entschiedene Owner-Fragen (Owner-Review 2026-07-13)

- **§8.1 Fail-Mode** für `backend=redis` — zwei in sich kohärente
  Varianten standen zur Wahl: **(i)** Default **fail-open-to-memory** +
  Opt-in-Flag `MTRACE_RATE_LIMIT_FAIL_CLOSED` (Schutzgut ist
  Telemetrie-Verfügbarkeit, nicht Auth-Flutung; Degradation = exakt der
  heutige Per-Replica-Zustand). **(ii)** Default **fail-closed** unter
  **Wiederverwendung des bestehenden geteilten Schalters**
  `MTRACE_AUTH_ISSUANCE_FAIL_OPEN`. (ii) ist der **dokumentierte
  Code-Präzedenzfall**: Origin- und Issuance-Limiter teilen den
  Fail-Mode-Schalter bewusst, „damit ein Operator nicht versehentlich
  einen halb-fail-closed Pfad konstruiert" (`main.go:1090`). Die Spannung
  ist real: (i) optimiert das Schutzgut pro Limiter, erzeugt aber genau
  die heterogene Fail-Mode-Landschaft (ein Redis, drei Limiter, gemischte
  Modi), vor der der bestehende Kommentar warnt; (ii) hält die Landschaft
  homogen und opfert dafür die Schutzgut-Differenzierung.
  **Entschieden: (i)** — die Schutzgut-Differenzierung wiegt schwerer;
  die gemischten Fail-Modi sind bewusst in Kauf genommen und werden in
  §4.3 und der Operator-Doku explizit dokumentiert.
- **§8.2 ENV-Name/Default**: **Entschieden:
  `MTRACE_RATE_LIMIT_BACKEND=memory|redis`**, Default `memory` (kein
  `disabled`: der Ingest-Limiter ist heute immer an, das bleibt) —
  konsistent mit der bestehenden `MTRACE_RATE_LIMIT_*`-Familie
  (Capacity/Refill), die der Operator ohnehin zusammen konfiguriert.
  Alternative `MTRACE_INGEST_RATE_LIMITER` (Selektor-Konvention
  `MTRACE_ORIGIN_RATE_LIMITER` / `MTRACE_AUTH_ISSUANCE_LIMITER`)
  verworfen.
- **§8.3 Per-Projekt-Buckets** (RAK-74/F-113-Anschluss): **Entschieden:
  uniforme ENV-Caps bleiben** (Parität zum heutigen In-Memory-Verhalten,
  kleiner Scope); policy-getriebene Projekt-Buckets
  (Use-Case-/Policy-Resolution-Umbau, deutlich größer) bleiben
  RAK-74-Folge-Scope.
- **§8.4 Release-Zuschnitt**: **Entschieden: eigener Minor-Release
  `0.25.0`** (nimmt den wartenden Cutover aus `[Unreleased]` mit) +
  Lastenheft-Patch mit neuer RAK-Gruppe analog `0.23.0` (RAK-126..130).
- **§8.5 ADR ja/nein**: **Entschieden: nein** — erweitert das bestehende
  Redis-Ops-Muster (ADR-0005, RAK-88/90) port-erhaltend auf einen weiteren
  Adapter; Entscheidung + Vorbehalte leben im Plan und in R-26, die
  RAK-Gruppe (§8.4) macht das Zielbild normativ.
