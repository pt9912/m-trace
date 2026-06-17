# Implementation Plan — `0.22.5` Load-/Soak-Smoke

> **Status**: ✅ **Abgeschlossen (2026-06-17)** — Tranche 1–3 implementiert
> + dreifach gereviewt; Tranche 4 (Load-Readiness-Verdict) geliefert: der
> 4h-Dispatch-Soak `27665523746` ergab kein stiller Verlust
> (persisted 55.327.560 == accepted, 3842 ev/s über 4h, 0,01 % Fehler) und
> Read-Retention-p95 12 ms bei 55,3 Mio Events → ADR-0005-Trigger #3 nicht
> ausgelöst. (R-25 hatte den ersten Soak `27628293077` in den 6h-Job-Cap
> laufen lassen — Readback-O(N) — und ist gefixt.)
> Antwort auf ein externes Tool-Review, das die **Lastfähigkeit als
> einzigen nicht-belegten Bereich** markierte (🔴).
> Commits: Limiter-ENV `14f3e64`, k6-Feasibility `e35f5c9`, smoke-load +
> Reconciliation `a7b8b6a`, Review F1–F7 `fa7794a` / F1-b `5c59d2c`,
> open-loop-SLO `e7f3336`, Soak-Probe `1ac6673`, Nightly + Doku `f580726`,
> Tranche-3-Review `d9edc03`. Tranche 4 = ein Nightly-/Dispatch-Soak
> (≥ 10 Mio Events, ~Stunden) entfernt — interaktiv nicht fahrbar.
>
> **Bezug**: NF-20, NF-22, NF-23 (Performance; **NF-21 bewusst nicht** —
> siehe §1), [ADR-0005](../../adr/0005-production-ops-backends.md)
> (Postgres-/Analytics-Trigger).
>
> **Auslöser**: [`docs/perf/budgets.md`](../../perf/budgets.md) sagt
> ausdrücklich „**Keine End-to-End-/Lab-Performance-Smokes** — Compose-
> Stacks bleiben außerhalb des Budget-Smokes". Es gibt Hot-Path-
> Mikrobenchmarks mit Budgets (`make benchmark-smoke`) und einen
> `benchstat`-Regressions-Nightly, **aber keinen** Nachweis unter
> realer Eventrate, parallelen Sessions, lang laufender SQLite-DB oder
> Prometheus/OTel-Backpressure.

## 0. Versions-Einordnung (festgelegt: Patch)

**Entscheidung: Patch `0.22.x`, kein Minor.** Begründung: reine
Verifikations-/Tooling-Lieferung, die bestehende NF-20/NF-22/NF-23
*verifiziert* — keine neue User-Surface, kein neuer Wire-Vertrag, keine
neue Anforderung/RAK. Damit greift [`releasing.md`](../../user/releasing.md)
§3.1 (Patch wie `plan-0.8.5`/`0.9.5`-Quality-Gate-Waves), **kein**
Lastenheft-Patch und **keine** §6.1-RAK-Matrix nötig. Die exakte
Patch-Nummer wird beim Tag bestätigt (`0.22.5` als nächstes freies
Patch-Slot, oder Bündelung mit dem `0.22.4`-Ton-Smoke in einem Tag) —
der Dateiname ist nur das Plan-Label. Die `0.23.0`-Minor-Option ist
damit verworfen.

## 1. Scope

In Scope:

- Opt-in Last-/Soak-Smoke gegen das **Core-Compose-Lab** (`make dev`),
  der NF-20/NF-22/NF-23 empirisch unterlegt: `make smoke-load`
  (`scripts/smoke-load.sh`) fährt einen HTTP-Lastgenerator (**k6**,
  Docker-Image `grafana/k6`, kein Host-Install) gegen
  `POST /api/playback-events` nach der Workload-Matrix (§4).
- **Implementierungs-Voraussetzung (eigener Schritt, NICHT nur Test)**:
  Der Ingest-Rate-Limiter ist heute **hart** auf 100 events/s/project
  codiert (`apps/api/cmd/api/main.go` `rateLimitCapacity`/`rateLimitRefill`),
  der Demo-Project-Resolver ist statisch (ebd.). Der **Kapazitäts-Modus
  (§3) braucht zuerst einen Code-Change**: `rateLimitCapacity`/
  `rateLimitRefill` per ENV konfigurierbar machen (Default unverändert
  100/s, damit kein Verhaltensbruch). Ohne das misst der Kapazitäts-Modus
  nur den bestehenden Limiter, nicht die Ingest-/Persistenz-Kapazität.
- **Mess-Größen** mit Schwellwerten (Smoke schlägt bei Verletzung an):
  - p95/p99-Ingest-Latenz unter definierter Rate (NF-20/NF-23);
  - **Kein stiller Verlust — über Readback/Reconciliation, NICHT nur
    Counter**: nach dem Lauf werden die gesendeten Events gegen die
    *persistierten* abgeglichen (Read-API, `sequence_number`-Kontinuität
    pro Session + Anzahl-Abgleich), plus HTTP-5xx-Rate. Begründung:
    synchrone Persistenz-Fehler (`500`) landen in **keinem**
    `mtrace_*`-Counter — `mtrace_dropped_events_total` ist laut F-122
    nur für Backpressure-Drops
    ([spec/telemetry-model.md](../../../spec/telemetry-model.md),
    [spec/architecture.md](../../../spec/architecture.md)).
  - Limiter-/Validierungs-Verhalten über `mtrace_rate_limited_events_total`
    /`mtrace_invalid_events_total` (ergänzend, siehe §3);
  - Dashboard-Read-Pfad (`ListSessions`/`GetSessionDetail`) p99 bei M
    aktiven Sessions (NF-22);
  - SQLite-Write-Durchsatz + Latenz-Drift über die Soak-Dauer
    (Single-Writer-Verhalten sichtbar machen).
- **Soak-Variante** als Daten-Lieferant für den
  [ADR-0005](../../adr/0005-production-ops-backends.md)-Postgres-Trigger
  #3, **fixierte Schwelle**: ≥ **10 Millionen** persistierte Events
  akkumulieren, dann Retention-/`ListSessions`-Queries messen und p95
  gegen die **2-Sekunden**-Grenze (ADR-0005:69) bewerten.
- Nicht-blockierender Nightly-Schritt; Gate-Eintrag in
  [`extra-gates.md`](extra-gates.md), Verweis aus
  [`releasing.md`](../../user/releasing.md) §2.

Nicht in Scope:

- **NF-21 bewusst ausgeschlossen**: „Player-SDK darf Playback nicht
  merklich beeinflussen" ist ein Browser-/SDK-Pfad, separat über das
  `0.8.0`-SDK-Bundle-Budget abgedeckt
  (`packages/player-sdk/scripts/performance-smoke.mjs`,
  [docs/perf/budgets.md](../../perf/budgets.md)). Der Last-Smoke trifft
  nur **Backend** (Ingest/Persistenz/Read).
- **Kein** Multi-Tenant-/High-Traffic-Produktionsnachweis (Lab-Last:
  Single-Replica, SQLite, Compose — analog NF-20 „lokale Demo-Last").
- **Kein** Multi-Replica-/K8s-/Postgres-/Redis-Lasttest (ADR-0005:
  deferred; der Smoke *liefert* nur die Trigger-Daten).
- **Kein** PR-blockierender Gate (lastabhängig flaky, vgl.
  [`plan-0.22.3-webrtc-drift`](../done/plan-0.22.3-webrtc-drift.md) §2).
  Opt-in + Nightly. Nicht in `make gates`.
- **Keine** Duplizierung der Hot-Path-Mikrobenchmarks
  (`benchmark-smoke`).

## 2. Methodik

Mikrobenchmark misst *eine Funktion isoliert* gegen ein Budget; der
Last-Smoke misst die *gesamte Ingest→Persistenz→Read-Kette unter Last*.
k6 rampt Virtual Users (= parallele Player-Sessions); k6-`thresholds`
(p95/p99, `http_req_failed`) liefern das Latenz-/Fehler-Pass/Fail. Der
**Verlust-Nachweis läuft über Readback/Reconciliation**, nicht über
Prometheus-Deltas (s. §1). Schwellwerte sind **Obergrenzen** und
versioniert in `docs/perf/budgets.md` (neue Section „Load-Smoke").

## 3. Auth-/Rate-Limit-Vertrag

`POST /api/playback-events` ist tokenpflichtig (Project- + Session-Token)
und rate-limitiert. Der Smoke läuft **mit dem echten Auth-Vertrag**
(gültige Tokens, kein Bypass), in zwei gepinnten Szenarien:

- **Kapazitäts-Modus** — Rate-Limit per ENV hochgesetzt (der Code-Change
  aus §1 ist Voraussetzung), um die *echte Ingest-/Persistenz-Kapazität*
  (NF-20) zu messen, nicht die Limiter-Decke.
- **Vertrags-Modus** — Default-Limits (100/s/project) aktiv; verifiziert,
  dass der Limiter unter Last korrekt greift (`429` +
  `mtrace_rate_limited_events_total` steigt) ohne stillen Verlust.

Gesetzte Limits + Token-/Project-Konfiguration werden im Smoke-Skript
**explizit und reproduzierbar gepinnt** (kein impliziter Default).

## 4. Workload-Matrix (Szenarien konkret, Schwellen ggf. erst nach Tranche 1)

Platzhalter-Zahlen als **Startpunkt** — vor Tranche-1-Baseline final zu
bestätigen, aber als benannte, vergleichbare Szenarien fixiert:

| Parameter | Kapazitäts-Modus | Vertrags-Modus | Soak |
| --- | --- | --- | --- |
| VUs / parallele Sessions (N) | Ramp 0→200 (1 min), hold | 200 konstant | 50 konstant |
| aktive Read-Sessions (M) | — | — | vorgeseedet, ≥ 1.000 |
| Eventrate (Ziel) | so hoch wie stabil | > 100/s/project (Limit testen) | moderat, dauerhaft |
| Batch-Größe | 20 Events/Batch | 20 | 20 |
| Dauer | 5 min (nach Warmup) | 3 min | bis ≥ 10 Mio Events |
| Warmup (aus Messung raus) | 30 s | 30 s | 60 s |
| DB | frische DB pro Lauf (Reset) | frische DB | dedizierte Soak-DB, Reset am Start |
| Read-Last | — | — | `ListSessions`/`GetSessionDetail` p99 bei M |
| Runner-Klasse | gepinnt + dokumentiert (Baseline: GitHub `ubuntu-24.04`-Nightly; lokale Läufe vermerken CPU-Klasse, vgl. `scripts/print-bench-runner-info.sh`) ||||

Warmup wird aus der Latenz-Auswertung ausgeschlossen; DB-Reset-/Reuse-
Politik ist pro Szenario oben fix, damit Baseline-Zahlen vergleichbar
bleiben.

### 4.1 Closed-Loop vs. Open-Loop (Review-Entscheidung für die Nightly)

Tranche 1/2 nutzt **closed-loop** (`--vus N`, jeder VU blockiert auf der
Antwort). Das ist für die **Korrektheits-Gates** (kein Verlust,
Fehlerquote, Limiter-Sanity) richtig — die hängen nicht an einer
Zielrate. Die gemessene ~800/s-Decke ist bewusst **Baseline/ADR-0005-
Evidenz, kein Gate**.

Für eine **Nightly-Durchsatz-/Latenz-SLO** (Tranche 3) ist closed-loop
der falsche Aufhänger: die Decke ist N-/hardware-abhängig und unter
Sättigung explodiert p95 (jeder VU blockiert) → flaky. Dann **k6
`constant-arrival-rate`-Executor** (open-loop): die offered load wird
vorgegeben, gemessen wird, ob das System mitkommt (`dropped_iterations`,
p95) → entkoppelt Last von Maschinen-Speed, stabile Schwelle über Runner
hinweg. Zielrate **deutlich unter der Decke** (aus ~800/s z. B.
400–500/s mit p95-Budget) — direkt an der Sättigung ist auch open-loop
instabil. Also zwei Szenarien, ein Skript: closed-loop „Decke finden"
(exploratory) + open-loop „SLO behaupten" (Nightly-Gate).

## 5. Tranchen

| Tranche | Inhalt | Stand |
| --- | --- | --- |
| 1 | Machbarkeit: k6 gegen Core-Lab, Ingest-Szenario mit echten Tokens; Baseline. | ✅ `e35f5c9` |
| 2 | Limiter-ENV (`14f3e64`) + `smoke-load.sh` + `make smoke-load`, beide Auth-Szenarien, Readback gegen echte `playback_events`; budgets.md §7. | ✅ `a7b8b6a` + Review `fa7794a`/`5c59d2c` |
| 3 | Open-loop-SLO (`e7f3336`), Soak-Retention-Probe (`1ac6673`), Nightly `load-smoke.yml` + Doku (`f580726`); Review `d9edc03`. | ✅ |
| 4 | **Load-Readiness-Verdict**: Zahlen (max. stabile Rate, p99, Durchsatz, Reconciliation) + ADR-0005-Trigger-#3-Stand mit Messwert. | ✅ Dispatch-Soak `27665523746` (4h02m, grün): closed-loop 20 VUs / 4h, `accepted=55.327.560` (**3842 ev/s**), `rate_limited=0`, errors `4080` (**0,01 %**), `http_req_duration` p90=214,8ms · **p95=636,0ms** · max=7380,4ms. Reconciliation: **`persisted 55.327.560 == accepted`** (kein stiller Verlust, 0 Überschuss). Retention-p95 (Proxy, indizierte Hot-Reads) list=2,1ms / detail-events=11,8ms bei 55,3 Mio Events → **ADR-0005-Trigger #3 NICHT ausgelöst** (< 2 s, ≥ 10 Mio). (Erster Soak `27628293077` lief in den 6h-Cap → R-25 🟢 gefixt.) |

> **Soak-Dispatch-Log (Tranche 4)** — ausgelöst 2026-06-16 via
> `gh workflow run load-smoke.yml -f mode=soak -f duration=4h`.
> Run `27628293077` (`https://github.com/pt9912/m-trace/actions/runs/27628293077`),
> Start `2026-06-16T15:21:16Z`.
>
> **Ergebnis (geprüft 2026-06-17): FEHLGESCHLAGEN — kein Verdict.** Der
> „Run load smoke"-Step lief 6h (15:21:29→21:21:39Z) und wurde vom
> **GitHub-6h-Job-Cap gecancelt**; der **Verdict-Step wurde dadurch
> skipped**. Ursache ist **kein** SLO-/System-Versagen, sondern ein
> Tooling-Skalierungs-Bug im Readback → **R-25**: Die k6-Lastphase war
> nach **4h00m sauber durch**, danach paginierte die Readback-
> Reconciliation ~45,7k Event-Seiten (1000/Seite) **~2h still** über HTTP,
> bis das 6h-Limit den Step killte (Artefakt-Log endet exakt an der
> k6-Summary; die Retention-Probe-Zeile erscheint nie).
>
> **k6-Ingest-Leg (erfasst, aber KEIN Load-Readiness-Verdict):** closed-
> loop 20 VUs / 4h, `BATCH_SIZE=20`. http_reqs 2.284.278 (158,6/s) →
> **events accepted 45.676.480 (3171,9 ev/s)**, rate_limited 0, rejected
> 9.080 (≈ 0,02 % der 202+Fehler). http_req_duration **p90=239,7ms ·
> p95=836,0ms · max=8276,6ms**. Das ist nur die Ingest-Leg unter
> kontrollierter Parallelität — **Reconciliation (`persisted` vs
> `accepted`) und Retention-Probe-p95 (ADR-0005-Trigger #3) liefen nie**.
>
> **R-25 gefixt (2026-06-17):** Readback im Autostart-Pfad zählt jetzt per
> direktem SQLite-`COUNT(*)` (GLOB auf die Lauf-Sessions) gegen das
> api-Volume statt per HTTP-Pagination — O(1) statt O(N). Lokal verifiziert
> (`make smoke-soak`, DURATION=10s/8 VUs: COUNT instant, `persisted 9900 ==
> accepted 9900`, Retention-Probe erstmals erreicht). Damit passt Last +
> Readback wieder klar in die 6h.
>
> **Re-Dispatch mit Fix — Verdict erreicht (2026-06-17):** Run
> `27665523746` (`https://github.com/pt9912/m-trace/actions/runs/27665523746`,
> headSha `dd37c3d` == Fix), Start `04:20:05Z`, **grün in 4h02m** (vs. 6h-Cancel
> davor). closed-loop 20 VUs / 4h: http_reqs 2.766.582 (192,1/s) → **events
> accepted 55.327.560 (3842,1 ev/s)**, rate_limited 0, rejected 4.080
> (≈ 0,01 %). http_req_duration **p90=214,8ms · p95=636,0ms · max=7380,4ms**.
> Readback-`COUNT(*)` instant: **`persisted 55.327.560 == accepted`** →
> kein stiller Verlust (0 at-least-once-Überschuss). Retention-Probe (Proxy,
> indizierte Hot-Reads) **list-p95 2,1ms / detail-events-p95 11,8ms** bei
> 55,3 Mio Events (Budget 2000ms). **ADR-0005-Trigger #3: NICHT ausgelöst**
> — ≥ 10 Mio erreicht (belastbar, kein INCONCLUSIVE), indizierte Hot-Read-p95
> 12ms ≪ 2 s; größenunabhängig, kein Korpus-Scan-Query in der API
> (Proxy-/Zukunfts-Messung). Messwerte in §5/§6, `CHANGELOG.md` und
> `docs/adr/0005-production-ops-backends.md` nachgetragen.

## 6. DoD

- [x] Limiter `rateLimitCapacity`/`rateLimitRefill` per ENV konfigurierbar
  (Default 100/s unverändert), mit Test (`14f3e64`).
- [x] `make smoke-load` reproduzierbar, opt-in, nicht-blockierend; beide
  Auth-Szenarien (§3) + Profile (closed/open) gepinnt.
- [x] Schwellwerte als Obergrenzen in `docs/perf/budgets.md` §7,
  referenziert von NF-20/NF-22/NF-23.
- [x] „Kein stiller Verlust" über **Readback gegen die echte
  `playback_events`-Tabelle** (`persisted >= accepted`; Lesefehler →
  INCONCLUSIVE, nie Verlust) — nicht `event_count`, nicht Counter-Deltas.
  Autostart-/CI-Pfad zählt seit R-25-Fix per direktem `COUNT(*)` gegen das
  Volume (O(1)); `events[]`-Array-Pagination bleibt Fallback für
  `SMOKE_LOAD_AUTOSTART=0`.
- [x] **Soak hat ≥ 10 Mio Events erreicht** (55,3 Mio im Dispatch-Soak
  `27665523746`); Retention-p95 gegen 2 s gemessen (12ms); ADR-0005-Trigger
  #3 als **nicht ausgelöst** bewertet (Proxy-gescopt). Mechanismus validiert
  (`1ac6673`), Verdikt-Daten geliefert (s. §5 Soak-Dispatch-Log).
- [x] **Nightly non-blocking, Verdikt aus Artefakt/Job-Summary, nicht aus
  der Job-Farbe**: `load-smoke.yml` + Verdict-Step (Job rot nur bei
  Hard-FAIL, grün bei INCONCLUSIVE; Debounce als R-24 offen).
- [x] Load-Readiness-Verdict im Plan-Closeout (§5/Header) + `CHANGELOG.md`
  nachgetragen (Dispatch-Soak `27665523746`).
- [x] `extra-gates.md` §3.9 + `releasing.md` §2.6.

## 7. Abgrenzung

Der Smoke beweist **Lab-Lastfähigkeit unter kontrollierter Parallelität**
(NF-20/NF-22/NF-23), nicht produktive Multi-Tenant-/Multi-Replica-
Skalierung. Letzteres bleibt
[ADR-0005](../../adr/0005-production-ops-backends.md)-Trigger-Gebiet. Der
Wert liegt darin, das einzige unbelegte Review-Feld mit Daten zu
schließen — entweder „Lab-Last hält Budget X" oder ehrlich „ab Rate Y
bricht SQLite, ADR-0005-Trigger empfohlen".
