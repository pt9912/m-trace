# Implementation Plan (Entwurf) — `0.23.0` Load-/Soak-Smoke

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
> Prometheus/OTel-Backpressure. Das Review stuft genau das als
> unbelegt ein.

## 0. Versions-Einordnung (Entscheidung offen)

Reine Verifikations-/Tooling-Lieferung **ohne neue User-Surface, ohne
neuen Wire-Vertrag, ohne neue Anforderung** — sie *verifiziert*
bestehende NF-20/NF-22/NF-23. Nach [`releasing.md`](../../user/releasing.md)
§3.1 ist das **Patch-Kategorie** (wie `plan-0.8.5`/`0.9.5`
Quality-Gate-Waves), also natürlicher Weise `0.22.x` ohne
Lastenheft-Patch/§6.1-RAK-Matrix. Der Name `0.23.0` (Minor) ist
vertretbar, **wenn** Load-Readiness bewusst als Meilenstein markiert
werden soll — dann aber ohne neue RAK, weil keine neue Anforderung
entsteht. **Empfehlung**: als Patch ausliefern (z. B. `0.22.5`),
Plan-Name hier nur als Label. Maintainer entscheidet beim Aktivieren.

## 1. Scope

In Scope:

- Opt-in Last-/Soak-Smoke gegen das **Core-Compose-Lab** (`make dev`),
  der NF-20/NF-22/NF-23 empirisch unterlegt: `make smoke-load`
  (`scripts/smoke-load.sh`) fährt einen HTTP-Lastgenerator (**k6**,
  Docker-Image `grafana/k6`, kein Host-Install) gegen
  `POST /api/playback-events` mit rampender Eventrate und N parallelen
  Sessions.
- **Mess-Größen** mit Schwellwerten (Smoke schlägt bei Verletzung an):
  - p95/p99-Ingest-Latenz unter definierter Rate (NF-20/NF-23);
  - **Kein stiller Verlust — über Readback/Reconciliation, NICHT nur
    Counter**: nach dem Lauf werden die gesendeten Events gegen die
    *persistierten* Events (Read-API, `sequence_number`-Kontinuität pro
    Session + Anzahl-Abgleich) abgeglichen, plus HTTP-5xx-Rate.
    Begründung: synchrone Persistenz-Fehler (`500`) landen in **keinem**
    `mtrace_*`-Counter — `mtrace_dropped_events_total` ist laut F-122
    nur für Backpressure-Drops reserviert
    ([spec/telemetry-model.md](../../../spec/telemetry-model.md),
    [spec/architecture.md](../../../spec/architecture.md)). Counter-Deltas
    allein würden einen Verlust übersehen.
  - Limiter-/Validierungs-Verhalten über `mtrace_rate_limited_events_total`
    /`mtrace_invalid_events_total` (ergänzend, siehe §3 Auth-Vertrag);
  - Dashboard-Read-Pfad (`ListSessions`/`GetSessionDetail`) p99 bei M
    aktiven Sessions (NF-22);
  - SQLite-Write-Durchsatz + Latenz-Drift über die Soak-Dauer
    (Single-Writer-Verhalten sichtbar machen).
- **Soak-Variante** als Daten-Lieferant für den
  [ADR-0005](../../adr/0005-production-ops-backends.md)-Postgres-Trigger
  #3, mit **fixierter Schwelle statt vager „lange Laufzeit"**:
  ≥ **10 Millionen** persistierte Events akkumulieren, dann
  Retention-/`ListSessions`-Queries messen und p95 gegen die
  **2-Sekunden**-Grenze (ADR-0005:69) bewerten. Erst diese konkrete
  Menge erlaubt im Closeout ein belastbares „Trigger ausgelöst / nicht
  ausgelöst".
- Nicht-blockierender Nightly-Schritt (eigener Workflow oder
  `benchmark.yml`-Sibling); Gate-Eintrag in
  [`extra-gates.md`](../in-progress/extra-gates.md), Verweis aus
  [`releasing.md`](../../user/releasing.md) §2.

Nicht in Scope:

- **NF-21 bewusst ausgeschlossen**: „Player-SDK darf Playback nicht
  merklich beeinflussen" ist ein Browser-/SDK-Pfad und bereits separat
  über das `0.8.0`-SDK-Bundle-Budget abgedeckt
  (`packages/player-sdk/scripts/performance-smoke.mjs`,
  [docs/perf/budgets.md](../../perf/budgets.md)). Der Last-Smoke trifft
  nur **Backend** (Ingest/Persistenz/Read), nicht den Client-Playback.
  Ein echter Player-SDK-Lastpfad wäre ein separates Folge-Item.
- **Kein** Multi-Tenant-/High-Traffic-Produktionsnachweis. Der Smoke
  belegt Lab-Lastfähigkeit (Single-Replica, SQLite, Compose), nicht
  produktive Skalierung — analog zur Spec-Sprache „lokale Demo-Last"
  (NF-20).
- **Kein** Multi-Replica-/K8s-/Postgres-/Redis-Lasttest (ADR-0005:
  diese Pfade sind deferred; der Smoke *liefert* nur die Trigger-Daten,
  er löst sie nicht aus).
- **Kein** PR-blockierender Gate: lastabhängig hardware-/runner-flaky
  (vgl. WebRTC-Lab in [`plan-0.22.3-webrtc-drift`](../done/plan-0.22.3-webrtc-drift.md)
  §2). Opt-in + Nightly, `continue-on-error`. Nicht in `make gates`.
- **Keine** Duplizierung der Hot-Path-Mikrobenchmarks
  (`benchmark-smoke`) — der Last-Smoke prüft die *End-to-End-Kette
  unter Parallelität*, nicht einzelne Funktions-Budgets.

## 2. Methodik

Mikrobenchmark (vorhanden) misst *eine Funktion isoliert* gegen ein
Budget; der Last-Smoke misst die *gesamte Ingest→Persistenz→Read-Kette
unter Last*. k6 rampt Virtual Users (= parallele Player-Sessions),
jede sendet realistische Event-Batches; k6-`thresholds` (p95/p99,
`http_req_failed`) liefern das Latenz-/Fehler-Pass/Fail.

Der **Verlust-Nachweis läuft über Readback/Reconciliation**, nicht über
Prometheus-Deltas: nach dem Lauf werden die gesendeten Events gegen die
persistierten abgeglichen (`sequence_number`-Kontinuität pro Session,
Gesamt-Anzahl, HTTP-5xx-Rate). Grund: ein synchroner `Append`-Fehler
(`500`) inkrementiert **keinen** `mtrace_*`-Counter (F-122); eine reine
Counter-Bilanz würde stillen Verlust übersehen. Die Counter
(`rate_limited`/`invalid`) bilanzieren ergänzend das Limiter-/Validierungs-
Verhalten. Schwellwerte sind **Obergrenzen** und versioniert in
`docs/perf/budgets.md` (neue Section „Load-Smoke"), analog zur
bestehenden Budget-Single-Source.

## 3. Auth-/Rate-Limit-Vertrag (klärt die Review-Offene-Frage)

`POST /api/playback-events` ist tokenpflichtig (Project- + Session-Token)
und rate-limitiert — beides Teil des bestehenden Modells. Der Smoke läuft
**mit dem echten Auth-Vertrag** (gültige Tokens, kein Bypass), in zwei
bewusst getrennten Szenarien:

- **Kapazitäts-Modus** — Rate-Limit des Test-Projects gezielt hoch
  provisioniert (per ENV/Project-Config angehoben), um die *echte
  Ingest-/Persistenz-Kapazität* (NF-20) zu messen, nicht die
  Limiter-Obergrenze.
- **Vertrags-Modus** — Default-Limits aktiv; verifiziert, dass der
  Limiter unter Last korrekt greift (`429` + `mtrace_rate_limited_events_total`
  steigt) und dabei nichts still verloren geht. Trennt „Limiter
  arbeitet" sauber von „Backend-Kapazität".

Die gesetzten Limits + Token-/Project-Konfiguration werden im
Smoke-Skript **explizit und reproduzierbar gepinnt** (kein impliziter
Default), damit Messwerte vergleichbar bleiben.

## 4. Tranchen (Skizze)

| Tranche | Inhalt |
| --- | --- |
| 1 | Machbarkeit: k6-Container gegen laufendes Core-Lab, ein Ingest-Szenario mit echten Tokens; Baseline-Zahlen (noch ohne Schwellen). |
| 2 | `scripts/smoke-load.sh` + `make smoke-load` (auto-up/down Core-Lab, beide Auth-Szenarien aus §3, Readback-Reconciliation), Schwellen in `docs/perf/budgets.md` verankern. |
| 3 | Soak-Variante: **≥ 10 Mio Events** akkumulieren, Retention-/`ListSessions`-p95 gegen **2 s** (ADR-0005 Trigger #3) messen; Nightly-Integration `continue-on-error`; Doku in `extra-gates.md` + `releasing.md`. |
| 4 | **Load-Readiness-Verdict** dokumentieren: konkrete Zahlen (max. stabile Rate, p99, SQLite-Durchsatz, Drift, Reconciliation-Ergebnis) — wandelt das Review-🔴 in ein belegtes/ehrlich gescoptes Ergebnis und benennt ADR-0005-Trigger-#3-Stand mit Messwert. |

## 5. DoD (Skizze)

- [ ] `make smoke-load` reproduzierbar gegen Core-Lab, opt-in,
  nicht-blockierend; beide Auth-Szenarien (§3) gepinnt.
- [ ] Schwellwerte als Obergrenzen in `docs/perf/budgets.md`
  (Single-Source), referenziert von NF-20/NF-22/NF-23.
- [ ] „Kein stiller Verlust" über **Readback/Reconciliation** belegt
  (persistierte vs. gesendete Events, `sequence_number`-Kontinuität +
  HTTP-5xx-Rate) — nicht über Counter-Deltas.
- [ ] Nightly-Schritt grün/`continue-on-error`.
- [ ] Soak hat **≥ 10 Mio Events** erreicht; Retention-p95 gegen **2 s**
  gemessen; ADR-0005-Trigger #3 explizit als ausgelöst / nicht
  ausgelöst bewertet, mit Messwert.
- [ ] Load-Readiness-Verdict im Plan-Closeout + CHANGELOG.
- [ ] `extra-gates.md`-Gate-Eintrag + `releasing.md`-Verweis.

## 6. Abgrenzung

Der Smoke beweist **Lab-Lastfähigkeit unter kontrollierter Parallelität**
(NF-20/NF-22/NF-23), nicht produktive Multi-Tenant-/Multi-Replica-
Skalierung. Letzteres bleibt ausdrücklich
[ADR-0005](../../adr/0005-production-ops-backends.md)-Trigger-Gebiet
(Postgres/Analytics deferred). Der Wert liegt darin, das einzige
unbelegte Review-Feld mit Daten zu schließen — entweder als „Lab-Last
hält Budget X" oder als ehrlich gescoptes „ab Rate Y bricht SQLite,
ADR-0005-Trigger empfohlen".
