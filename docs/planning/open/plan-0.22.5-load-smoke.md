# Implementation Plan (Entwurf) — `0.22.5` Load-/Soak-Smoke

> **Status**: 📝 **Entwurf / nicht gestartet**. Skizze als Antwort auf
> ein externes Tool-Review (2026-06-16), das die **Lastfähigkeit als
> einzigen nicht-belegten Bereich** markierte (🔴). Aktivierung: per
> `git mv` nach [`../in-progress/`](../in-progress/), sobald eingeplant.
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
  [`extra-gates.md`](../in-progress/extra-gates.md), Verweis aus
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

## 5. Tranchen (Skizze)

| Tranche | Inhalt |
| --- | --- |
| 1 | Machbarkeit: k6-Container gegen laufendes Core-Lab, ein Ingest-Szenario mit echten Tokens, Workload-Matrix §4 als Startpunkt; Baseline-Zahlen (noch ohne Schwellen). |
| 2 | Limiter-ENV-Konfig (Code-Change §1) + `scripts/smoke-load.sh` + `make smoke-load` (auto-up/down Core-Lab, beide Auth-Szenarien §3, Readback-Reconciliation); Schwellen in `docs/perf/budgets.md`. |
| 3 | Soak-Variante: **≥ 10 Mio Events**, Retention-/`ListSessions`-p95 gegen **2 s** (ADR-0005 Trigger #3); Nightly-Integration; Doku in `extra-gates.md` + `releasing.md`. |
| 4 | **Load-Readiness-Verdict**: konkrete Zahlen (max. stabile Rate, p99, SQLite-Durchsatz, Drift, Reconciliation-Ergebnis); ADR-0005-Trigger-#3-Stand mit Messwert. |

## 6. DoD (Skizze)

- [ ] Limiter `rateLimitCapacity`/`rateLimitRefill` per ENV
  konfigurierbar (Default 100/s unverändert), mit Test.
- [ ] `make smoke-load` reproduzierbar, opt-in, nicht-blockierend; beide
  Auth-Szenarien (§3) + Workload-Matrix (§4) gepinnt.
- [ ] Schwellwerte als Obergrenzen in `docs/perf/budgets.md`,
  referenziert von NF-20/NF-22/NF-23.
- [ ] „Kein stiller Verlust" über **Readback/Reconciliation** belegt
  (persistiert vs. gesendet, `sequence_number`-Kontinuität +
  HTTP-5xx-Rate) — nicht über Counter-Deltas.
- [ ] Soak hat **≥ 10 Mio Events** erreicht; Retention-p95 gegen **2 s**
  gemessen; ADR-0005-Trigger #3 als ausgelöst / nicht ausgelöst bewertet,
  mit Messwert.
- [ ] **Nightly-Step ist non-blocking (`continue-on-error`), aber der
  Nachweis steckt im Artefakt/Job-Summary**: Step-Outcome, k6-Summary
  (p95/p99, `http_req_failed`) und Reconciliation-Report werden als
  Workflow-Artefakt + Job-Summary persistiert. Ein grüner Job trotz
  fehlgeschlagenem Step zählt **nicht** als Nachweis — das Verdict liest
  sich aus dem Report, nicht aus der Job-Farbe.
- [ ] Load-Readiness-Verdict im Plan-Closeout + CHANGELOG.
- [ ] `extra-gates.md`-Gate-Eintrag + `releasing.md`-Verweis.

## 7. Abgrenzung

Der Smoke beweist **Lab-Lastfähigkeit unter kontrollierter Parallelität**
(NF-20/NF-22/NF-23), nicht produktive Multi-Tenant-/Multi-Replica-
Skalierung. Letzteres bleibt
[ADR-0005](../../adr/0005-production-ops-backends.md)-Trigger-Gebiet. Der
Wert liegt darin, das einzige unbelegte Review-Feld mit Daten zu
schließen — entweder „Lab-Last hält Budget X" oder ehrlich „ab Rate Y
bricht SQLite, ADR-0005-Trigger empfohlen".
