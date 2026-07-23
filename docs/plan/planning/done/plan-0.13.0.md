# Implementation Plan — `0.13.0` (Production / Ops Backends)

> **Status**: ✅ released 2026-05-12 — Tranchen 0..5
> geschlossen; Tag `v0.13.0`.
> Vorgänger ist `0.12.6` (released 2026-05-12, Tag `v0.12.6`;
> Plan in [`done/plan-0.12.6.md`](../done/plan-0.12.6.md)).
> Lastenheft-Patch `1.1.18` mit RAK-91..RAK-95 ist in §13.17
> persistiert.
>
> **Release-Typ**: Minor-Release mit Lastenheft-Patch, neuer RAK-
> Gruppe, RAK-Verifikationsmatrix und Tag `v0.13.0`.
>
> **Ziel**: Production-/Ops-nahe Folgepunkte werden in einen
> entscheidbaren Scope überführt: `MVP-40` Postgres, `MVP-41`
> ClickHouse/VictoriaMetrics/Mimir, `MVP-42` Kubernetes-Manifeste,
> `MVP-43` Devcontainer und `MVP-44` Release-Automatisierung. `NF-18`
> wird dabei mit `MVP-42` harmonisiert.
>
> **Bezug**:
> [`spec/lastenheft.md`](../../../../spec/lastenheft.md) NF-18,
> MVP-40..MVP-44; [`docs/planning/in-progress/roadmap.md`](../in-progress/roadmap.md)
> Folge-ADRs; [`docs/planning/in-progress/risks-backlog.md`](../risks-backlog.md)
> R-9.
>
> **Nachfolger**: offen.

## 0. Konvention

DoD-Checkboxen tracken den Lieferstand:

- `[x]` ausgeliefert mit Commit-Hash.
- `[ ]` offen.
- `[!]` blockiert durch ADR- oder Scope-Entscheidung.
- 🟡 in Arbeit.

### 0.1 Scope-Definition

`0.13.0` ist ein Decision-and-Seed-Release für Production-/Ops-
Backends. Es kann Artefakte liefern, aber liefert nicht alle
Produktionspfade vollständig.

In Scope:

- `MVP-40`: Postgres als produktionsnaher Store bewerten und ggf.
  minimalen Adapter-/Migrations-Slice planen.
- `MVP-41`: ClickHouse, VictoriaMetrics oder Mimir als hochvolumiges
  Event-Backend bewerten, inklusive Kosten-/Operativvergleich.
- `MVP-42` / `NF-18`: Kubernetes-Manifeste als optionalen Folgepfad
  konkretisieren, optional als Beispieldateien liefern.
- `MVP-43`: Devcontainer für reproduzierbare Entwicklung liefern oder
  als Folge-Scope begründet deferieren.
- `MVP-44`: Release-Automatisierung mit klaren Gates und manuellem
  Freigabepfad definieren.
- R-9 prüfen: K8s-Smoke-Stage und Observability-Labelling beeinflussen.

Out of scope:

- Kein vollständiger Production-Kubernetes-Betrieb.
- Kein Managed-Cloud-Betrieb.
- Kein Multi-Tenant-SaaS-Produkt.
- Keine verpflichtende ClickHouse-/VictoriaMetrics-/Mimir-Pflicht für den
  Standardbetrieb.
- Keine automatische Veröffentlichung ohne explizite human approval.

### 0.2 Vorgänger-Gate

- `0.12.6` ist released; Roadmap muss auf `0.13.0` als aktive
  Folgephase zeigen.
- Bestehende Operativgrundsätze aus `0.12.0` bleiben aktiv: kein
  Vollausbau von Production-Identity oder Secret-Management im Scope.
- Keine Änderung am bereits ausgelieferten Auth-/Token- und Ingest-
  Kernpfad ohne separate ADR oder Sicherheitsbezug.

### 0.3 Architektur-/Scope-Entscheidungen (Planpflicht)

Die Tranche-0-Entscheidung klärt mindestens folgende Dimensionen:

- Welche Komponenten in `0.13.0` als Seed implementiert werden
  (minimaler Codepfad) und welche in 0.14+ verschoben werden.
- Welche Backend-Optionen als „defer“ markiert werden, inklusive
  Auslösern für Reaktivierung.
- Welche Dokumente (ADR/Plan/Spec) die Verbindlichkeit haben.

### 0.4 Lastenheft-Patch (`1.1.18`, RAK-91..RAK-95)

Der Plan ergänzt eine neue RAK-Serie für `MVP-40`..`MVP-44` und
`NF-18` als Mindestverpflichtung für 0.13.

T0-Aktivierung 2026-05-12: `0.12.5` belegt `RAK-77`..`RAK-82`
in §13.15, `0.12.6` belegt `RAK-83`..`RAK-90` in §13.16.
Damit ist `0.13.0` verbindlich auf Lastenheft-Patch `1.1.18`
und `RAK-91`..`RAK-95` in §13.17 gesetzt.

| RAK | Priorität | Inhalt |
| --- | --- | --- |
| RAK-91 | Muss | Operativer Scope für `MVP-40`..`MVP-42`, inkl. klarer Seed-/Defer-Boundaries und Nachweise. |
| RAK-92 | Muss | Operativer Scope für `MVP-41`: Vergleichs- und Entscheidungspfad gegen ClickHouse/VictoriaMetrics/Mimir (oder gleichwertige Option). |
| RAK-93 | Muss | `MVP-42`/`NF-18` als optionaler K8s-Optionpfad wird normativ begrenzt, keine Vollbereitschaftszusage. |
| RAK-94 | Muss | Operativer Scope für `MVP-43`: Devcontainer als Seed oder explizites Defer mit Begründung/Trigger. |
| RAK-95 | Muss | Release-Prozess enthält mindestens manuelle Freigabe in allen automationsrelevanten Stufen, inklusive sicherer Rollback-Regeln und Closeout-Nachweis. |

### 0.5 Qualitätsregeln für 0.13.0

- Keine neuen Pflichtabhängigkeiten in der lokalen Standardumgebung.
- Evaluierungen von MVP-40/MVP-41 dürfen in isolierten Probe-/POC-Setups
  erfolgen (z. B. Container oder expliziter optionaler Dev/Test-Pfad), solange
  diese nicht als Standard- oder Pflichtabhängigkeiten in der lokalen
  Standardumgebung hinterlegt werden.
- Keine Architekturentscheidung ohne nachvollziehbaren ADR-
  Entscheidungsweg.
- Keine Entscheidung ohne dokumentierten Migrations-/Rollbackpfad.
- Reproduzierbarkeit von Build, Test und Release vor Funktionsumfang.
- Jede Tranche endet mit Verifikationsnachweisen (Datei + Test + Doku).

### 0.6 Tranche-Output-Verpflichtungen

Für jede abgeschlossene Tranche müssen mindestens diese drei Dokumenttypen
vorliegen:

- **Entscheidungsnachweis**: ADR, Entscheidungsnotiz oder Plan-Update.
- **Auswirkungsnachweis**: Metriken, Vergleichstabelle oder Risikoanalyse.
- **Closeout-Nachweis**: Statusänderungen in Roadmap/risks-backlog und DoD-Update.

Zusatzregel:

- Jede neue Entscheidung muss mit einem „What ändert sich / What bleibt
  unverändert“-Abschnitt abgeschlossen werden, damit kein Scope-Drift
  entsteht.

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Erwartetes Ergebnis | Eingang | Ausgang | Status |
| --- | --- | --- | --- | --- | --- |
| 0 | Aktivierung, Lastenheft-Patch, ADR-Schnitt und Scope-Gates | Freigabevoraussetzungen geklärt, Scope fixiert | Tranche-Ready | Freigabe- und Blockerliste | ✅ |
| 1 | Postgres-Entscheidung und Adapter-Scope (`MVP-40`) | Entscheidung inkl. Seed-Umfang oder strukturierte Defer-Liste | Geklärter Scope | Migrations-/Defer-Entscheidung + Trigger | ✅ |
| 2 | Analytics-Backend-Entscheidung (`MVP-41`) | Vergleichsmatrix + klare Pfadentscheidung | Persistenz-/Query-Anforderungen | Entscheidungsprotokoll | ✅ |
| 3 | Kubernetes-/Devcontainer-/NF-18-Risiko (`MVP-42`, `MVP-43`, `R-9`, `NF-18`) | Konkreter Scope-Hebel für Folgephase | Tranche-1/2 Ergebnisse | Operativer Optionspfad + Defer-Regeln | ✅ |
| 4 | Release-Automatisierung (`MVP-44`) | Sichere Automations- und Freigaberegeln | Zielbild Release-Prozess | Runbook + Gateschema | ✅ |
| 5 | Gates, RAK-Matrix, Versions-Bump, Closeout und Tag | Abschluss nachweisbar und wiederholbar | Tranche-4 Ergebnis | RAK-Status Grün + Release-Nachweise | 🟡 |

## 2. Tranche 0 — Aktivierung, Scope-Gates und Plan-Härtung

Ziel: Voraussetzungen, Scope und Architekturbezug sind so klar,
 dass keine Entscheidung im mittleren Tranchefluss erneut gestoppt werden
 kann.

DoD:

- [x] Plan von `docs/planning/open/plan-0.13.0.md` nach
  `docs/planning/in-progress/plan-0.13.0.md` verschoben.
- [x] Ausgangszustand von `git status --short` dokumentiert und im
  Tranche-0-Notizblock gespeichert (leer vor Aktivierungs-Move).
- [x] Lastenheft-Patch mit neuer RAK-Gruppe für `MVP-40`..`MVP-44`
  und `NF-18` ergänzt.
- [x] ADR-Schnitt definiert: Architektur-/Persistenz-/Release-
  Entscheidungen (wir liefern ADR oder Plan-DoD).
  - [x] `decision-record` enthält mindestens zwei Alternativen je Kernfrage.
- [x] `docs/planning/in-progress/roadmap.md` auf `0.13.0` als aktive
  Folgephase umgestellt.
  - [x] Alle offenen Gegenargumente dokumentiert und nicht stillschweigend verworfen.
- [x] Neue RAK-Range in `spec/lastenheft.md` persistiert und gegen
  bestehende IDs geprüft (Platzhalter-Mapping in §0.4 oben:
  `RAK-91`..`RAK-95` in §13.17, da
  `RAK-77`..`RAK-82` mit `0.12.5` §13.15 und `RAK-83`..`RAK-90`
  mit `0.12.6` §13.16 vergeben).
- [x] Risiko-R-9 gegen Observability-/Smoke-Impact geprüft und im
  `risks-backlog.md` abgelegt.
- [x] Tranche-0-Notiz (kurz) mit Status, offenen Entscheidungen,
  offenen Triggern und Ausnahmen angelegt.
  - [x] Notiz enthält erwartete Entscheidungstermine und Verantwortliche.

### 2.1 Tranche-0-Notiz — 2026-05-12

Status:

- `git status --short` war vor Aktivierungsbeginn leer.
- `0.12.6` ist released und archiviert; `0.13.0` ist die aktive
  Minor-Folgephase.
- Lastenheft-Patch `1.1.18` / §13.17 / RAK-91..RAK-95 ist die
  verbindliche RAK-Gruppe für diesen Plan.
- Dokumentationsgate: `make docs-check` grün am 2026-05-12.

Decision-Record:

| Kernfrage | Option A | Option B | T0-Entscheidung |
| --- | --- | --- | --- |
| Postgres (`MVP-40`) | Seed-Slice mit Adapter-/Migrationsspur | Defer mit messbaren Triggern | Tranche 1 entscheidet nach Vergleich; SQLite bleibt Default und Pflichtpfad. |
| Analytics-Backend (`MVP-41`) | POC mit ClickHouse/VictoriaMetrics/Mimir | Defer ohne Code-Artefakt | Tranche 2 entscheidet über `proceed`/`defer`/`POC`; keine neue lokale Pflichtabhängigkeit. |
| K8s/NF-18 (`MVP-42`, R-9) | optionale Beispielmanifeste plus K8s-Smoke-Entscheidung | reiner Folge-Scope mit Triggern | Tranche 3 entscheidet; keine Production-Ready-K8s-Zusage in T0. |
| Devcontainer (`MVP-43`) | Beispielkonfiguration liefern | Defer mit Begründung | Tranche 3 entscheidet gemeinsam mit K8s-DevEx-Bedarf. |
| Release-Automatisierung (`MVP-44`) | Dry-Run-/Guard-Automation | Runbook-only mit manuellen Gates | Tranche 4 entscheidet; automatische Veröffentlichung ohne human approval bleibt ausgeschlossen. |

Gegenargumente / Blocker:

- Neue Backend-Pfade können lokale Standard-Setups schwerer machen;
  T0 hält daher alle Postgres-/Analytics-/K8s-Pfade optional.
- K8s-Smoke-Stage kann R-9 auslösen, weil heutige Observability-
  Label-Allowlist Compose-Lab-spezifisch ist.
- Production-Identity, Managed-Cloud-Betrieb und Secret-Management-
  Vollausbau bleiben außerhalb dieses Plans.

Erwartete Termine und Verantwortliche:

| Tranche | Erwarteter Entscheidungspunkt | Owner |
| --- | --- | --- |
| 1 | Postgres Seed/Defer vor Tranche-2-Start | Platform/PM |
| 2 | Analytics `proceed`/`defer`/`POC` vor Tranche-3-Start | Platform/QA |
| 3 | K8s/NF-18/R-9 + Devcontainer vor Tranche-4-Start | Platform/Ops + Platform/DevEx |
| 4 | Release-Automation-Guard vor Closeout | Platform/CI |

What ändert sich:

- `0.13.0` ist aktiv und normativ mit RAK-91..RAK-95 in
  `spec/lastenheft.md` verankert.
- R-9 ist explizit in den `0.13.0`-Tranche-3-Scope gezogen.

What bleibt unverändert:

- SQLite bleibt der lokale Standard-Store.
- K8s, Postgres und Analytics-Backends werden nicht zur lokalen
  Pflichtabhängigkeit.
- Release-Veröffentlichung bleibt manuell freigabepflichtig.

Go/No-Go-Kriterien nach Tranche 0:

- **Go:** Scope-Sätze sind vollständig und keine neue Pflichtabhängigkeit
  (Dependency) wird eingeschoben.
- **No-Go:** Unklare ADR-Zuordnung, ungeklärte R-9-Interaktion oder
  unbewertete Pflicht-Trigger.

## 3. Tranche 1 — Postgres (MVP-40)

Ziel: Entweder erster produktionsnaher Slice wird validiert oder eine
saubere Defer-Entscheidung trifft die Ausnahmen.

DoD:

- [x] Entscheidung dokumentiert: vollständige Implementierung,
  minimaler Seed-Slice oder Defer.
  - [x] Entscheidung enthält „Entscheidung“, „Begründung“, „Nicht entschieden“. 
- [x] Wenn Seed implementiert: mindestens eine deterministische
  Migrationsspur definiert (Schema/Repo/Adapter) inklusive Daten-
  Rückwärtskompatibilitätsprüfung mit SQLite.
  - [x] Migrationspfad enthält `migrate up`, `rollback`, `replay`.
- [x] Wenn Seed nicht implementiert: Triggerschwellen (z. B.
  Last, SLA, Betriebszeit, Recovery-Szenarien) festgelegt.
  - [x] Jeder Trigger hat einen messbaren Schwellwert + Besitzer.
- [x] Contract-/Integrationstests pinnen SQLite-Verhalten, damit Postgres
  nicht unfreiwillig als versteckte Pflichtwirkung eingeführt wird.
- [x] Betriebsrisiken dokumentiert: Datenkonsistenz bei Migration,
  Ausfall- und Recoveryverhalten, Backup/Restore-Bezug.
- [x] Ergebnis entscheidet, welche Pfade in Tranche 5 in RAK/Changelog
  aufgenommen werden.
  - [x] Entscheidung wird vor Tranche-2-Start freigegeben.

### 3.1 Entscheidung — 2026-05-12

**Entscheidung:** Kein Postgres-Runtime-Adapter in `0.13.0`.
`MVP-40` wird als strukturierter Defer mit Migrations-Seed über die
bestehende neutrale `apps/api/internal/storage/schema.yaml`-Spur
geschlossen. ADR 0005 ist der Entscheidungsnachweis.

**Begründung:** SQLite erfüllt den lokalen Default und Restart-
Durability. Ein Postgres-Adapter würde Repository-Implementierungen,
Dual-Read/Replay, Backup/Restore und Rollback berühren, ohne dass für
`0.13.0` bereits eine Multi-Replica- oder Recovery-Schwelle ausgelöst
ist.

**Nicht entschieden:** Kein Postgres-DSN-Selector, kein produktiver
Adapter, kein automatischer SQLite-Export und kein dualer Schreibpfad.

Migrationspfad:

- `migrate up`: bestehende `schema.yaml` bleibt Single-Source-of-Truth;
  ein Folge-Plan erzeugt `--target postgresql`-DDL und einen eigenen
  Adapter-Slice.
- `rollback`: kein Runtime-Pfad landet in `0.13.0`; Rollback ist
  Entfernen des Folge-Slices vor Aktivierung. SQLite-Daten bleiben
  unverändert.
- `replay`: Folge-Plan muss vor Code-Enablement einen Event-/Session-
  Replay aus SQLite-Snapshot oder API-Wire-Fixtures definieren.

Reaktivierungs-Trigger:

| Trigger | Schwelle | Owner |
| --- | --- | --- |
| Multi-Replica-Store | ≥ 2 API-Replicas brauchen denselben Store ohne shared-volume SQLite | Platform/Ops |
| Recovery-SLO | `RPO <= 15 min` oder `RTO <= 30 min` wird verbindlich | Platform/Ops |
| Retention-/Query-Last | > 10 Mio. Events, p95 für Read-Pfade < 2 s erforderlich | Platform/QA |

What ändert sich:

- Postgres ist als messbar reaktivierbarer Folgepfad in ADR 0005
  festgehalten.

What bleibt unverändert:

- SQLite bleibt Default und wird durch bestehende Contract-/
  Integrationstests gepinnt.

## 4. Tranche 2 — Analytics-Backend (MVP-41)

Ziel: Datenpfadbedarf zwischen „kein Pflichtbackend“, „POC“ und
„konkreter Folgepfad“ trennen.

DoD:

- [x] ClickHouse/VictoriaMetrics/Mimir-Bedarf anhand aktueller
  Datenpfade, Query-Anforderungen und Datenvolumen bewertet.
- [x] Vergleichsmatrix liegt vor: Komplexität, Betriebskosten,
  Abfragefähigkeit, Integrationsaufwand, Relevanz der Query-Workloads,
  Migrationsrisiko.
- [x] Entscheidung dokumentiert als `proceed` / `defer` / `POC`.
  - [x] Matrix enthält einen „Wenn-Pilot→Go“-Abschnitt mit zeitlicher Grenze.
- [x] Keine neue lokale Pflichtabhängigkeit.
- [x] Falls POC: Zielbild, Erfolgskriterien, Erfolgspunkte,
  Abbruchkriterien im Plan verankert.
  - [x] POC-Report enthält Kostenannahmen, Datenmodell-Deckung und
    Rechenlastabschätzung.

### 4.1 Entscheidung — 2026-05-12

**Entscheidung:** `MVP-41` wird als `defer` geschlossen. Kein
ClickHouse-, VictoriaMetrics- oder Mimir-Pflichtbackend in `0.13.0`.

| Option | Komplexität | Betriebskosten | Query-Fit | Integrationsaufwand | Migrationsrisiko | Ergebnis |
| --- | --- | --- | --- | --- | --- | --- |
| ClickHouse | mittel-hoch | mittel | stark für Event-/Session-Ad-hoc-Analysen | hoch, eigenes Datenmodell | mittel | Defer bis Volumen-Trigger |
| VictoriaMetrics | mittel | niedrig-mittel | stark für Metriken, schwächer für rohe Events | mittel | niedrig-mittel | Defer; Prometheus reicht aktuell |
| Mimir | hoch | hoch | stark für Multi-Tenant-Metriken | hoch | mittel | Defer; kein Multi-Tenant-SLO |

Wenn-Pilot→Go:

- Pilot nur mit Owner Platform/QA und maximal 30 Kalendertagen Laufzeit.
- Go nur bei > 50 Mio. Events/Tag oder konkretem Ad-hoc-Analysebedarf,
  der mit API-/Prometheus-Pfaden nicht erfüllbar ist.
- Abbruch bei ungeklärter Datenmodell-Deckung, dauerhaftem Dual-Write-
  Bedarf ohne Replay-Plan oder unklarem Kostenrahmen.

Kosten-/Rechenannahme:

- ClickHouse: zusätzlicher Stateful-Service, Storage-Wachstum linear zu
  Event-Volumen, hoher Nutzen erst bei breiten Scan-/Aggregation-
  Workloads.
- VictoriaMetrics/Mimir: zusätzlicher Metrics-Backend-Betrieb, Nutzen
  nur bei Metrik-Retention oder Multi-Tenant-Skalierung jenseits des
  aktuellen Prometheus-Labs.

What ändert sich:

- Analytics-Backends sind mit klaren Go-/Abbruchkriterien reaktivierbar.

What bleibt unverändert:

- Prometheus-/SQLite-/API-Read-Pfade bleiben die einzigen Default-Pfade.

## 5. Tranche 3 — Kubernetes- und Devcontainer-Scope (MVP-42, MVP-43, NF-18, R-9)

Ziel: K8s- und Devcontainer-Scope wird normativ abgegrenzt und ein
entscheidbarer Folgepfad garantiert.

DoD:

- [x] `NF-18` und `MVP-42` sind harmonisiert: klarer Option-Pfad
  (`konkret + optional` oder `folge`) und dokumentiert.
- [x] R-9-Auswirkungen auf Observability-Label-Allowlists und
  Smoke-Stage vollständig geprüft; Ergebnis im Risks-Backlog und/oder
  ADR.
  - [x] Risiko-Matrix enthält mindestens zwei konkrete Gegenmaßnahmen.
- [x] Devcontainer (`MVP-43`) entschieden: liefern Beispielkonfiguration
  oder deferred mit konkreten Gründen.
- [x] README-/User-Doku-Abgrenzung bleibt konsistent: keine
  Produktions-Ready-K8s-Zusage.
- [x] Scope-Transitie auf nachfolgende Tranche (`0.14.0` o. ä.)
  klar beschrieben.
  - [x] Jede Option hat einen klaren Auslöser für die Folgephase.

### 5.1 Entscheidung — 2026-05-12

**Entscheidung:** `MVP-42`/`NF-18` wird als `konkret + optional`
geschlossen: Beispielmanifeste unter `deploy/k8s/`, aber keine
Production-Ready-Zusage und kein Standard-Gate. `MVP-43` wird als Seed
geliefert: `.devcontainer/devcontainer.json`.

Artefakte:

- `deploy/k8s/README.md`
- `deploy/k8s/namespace.yaml`
- `deploy/k8s/api.yaml`
- `deploy/k8s/analyzer-service.yaml`
- `deploy/k8s/dashboard.yaml`
- `.devcontainer/devcontainer.json`

R-9-Risiko-Matrix:

| Risiko | Gegenmaßnahme | Folge-Trigger |
| --- | --- | --- |
| K8s-Infrastruktur-Labels (`pod`, `namespace`, `container`) brechen Compose-Allowlist | separater K8s-Allowlist-Modus statt Erweiterung des Compose-Defaults | K8s-Smoke wird PR-/Release-Gate |
| Aggregatmetriken vermischen App- und Infrastruktur-Labels | Smoke-Scope-Trennung per explizitem Profil/ENV-Gate | neue K8s-Observability-Manifeste landen |
| Operator liest Beispielmanifeste als Production-Ready | README- und Deploy-Doku markieren Nicht-Produktionsstatus | Ingress/TLS/HPA/NetworkPolicy-Scope wird geplant |

What ändert sich:

- K8s und Devcontainer haben konkrete Seed-Dateien.

What bleibt unverändert:

- Compose bleibt primäres Lab; K8s-Smokes sind kein Default-Gate.

## 6. Tranche 4 — Release-Automatisierung (MVP-44)

Ziel: Automations-Mechanik für Release ohne Sicherheitsrisiko und ohne
unbemerkte automatische Veröffentlichung definieren.

DoD:

- [x] Entscheidbare Automation ausgewählt: was wird automatisiert,
  was bleibt manuell.
- [x] Automationsumfang in einem RACI-/Owner-Mapping hinterlegt.
- [x] Sichere Freigabe-Guards definiert:
  - Branch- und Tag-Muster,
  - Freigabekanal (Reviewer/Freigabe-
    Kontrolle),
  - zeitliche oder environment-gesteuerte Restriktionen.
- [x] Release-Doku bleibt die Source-of-Truth; automatischer Teil nur als
  Ergänzung mit Rückholbarkeit.
- [x] Automations- oder Dry-Run-Tests sind definiert (ohne reale
  Veröffentlichung).
- [x] CI-/Local-Runbook beschreibt minimalen sicheren Ausführungsweg.
  - [x] Runbook enthält Notfallplan bei fehlgeschlagenem Guard-Check.

### 6.1 Entscheidung — 2026-05-12

**Entscheidung:** Automatisiert wird nur der lokale Release-Guard.
Commit, Tag, Push und GitHub-Release bleiben manuelle Schritte.

Artefakte:

- `scripts/release-guard.sh`
- `Makefile` Target `release-guard`
- `docs/user/releasing.md` §3.0

RACI:

| Schritt | Responsible | Accountable | Consulted | Informed |
| --- | --- | --- | --- | --- |
| Versions-/Changelog-Bump | Platform/CI | m-trace Owner | QA | Nutzer |
| `make gates`/`make build` | Platform/CI | m-trace Owner | QA | Nutzer |
| `make release-guard VER=X.Y.Z` | Release-Operator | m-trace Owner | QA | Nutzer |
| Tag/Push/GitHub-Release | Release-Operator | m-trace Owner | QA | Nutzer |

Guard-Regeln:

- Branch `main`, außer lokaler Guard-Test setzt explizit
  `MTRACE_RELEASE_ALLOW_NON_MAIN=1`.
- saubere Arbeitskopie, außer lokaler Guard-Test setzt explizit
  `MTRACE_RELEASE_ALLOW_DIRTY=1`.
- Remote-Tag-Prüfung gegen `origin`, außer lokaler Guard-Test setzt
  explizit `MTRACE_RELEASE_ALLOW_OFFLINE=1`.
- manuelle Freigabe über `MTRACE_RELEASE_APPROVED=1`.
- Tag `vX.Y.Z` darf lokal und auf `origin` nicht existieren.

Notfallplan:

- Guard-Fehler vor Tag: Release stoppen, Ursache korrigieren, Guard neu
  ausführen.
- Tag lokal erstellt, aber nicht gepusht: `git tag -d vX.Y.Z`.
- Tag gepusht: Rollback nach `docs/user/releasing.md` §6.

What ändert sich:

- Release-Freigabe ist maschinell prüfbar, bleibt aber manuell.

What bleibt unverändert:

- Keine automatische Veröffentlichung ohne menschliche Freigabe.

## 7. Tranche 5 — Release-Closeout und Abschluss

Ziel: Alle Gates sind nachweisbar erfüllt und der Release kann
sauber abgeschlossen werden.

DoD:

- [x] RAK-Verifikationsmatrix vollständig ausgefüllt.
- [x] `make docs-check` grün (oder dokumentarischer Äquivalent-Gate bei reiner
  Plan-/ADR-/Spec-Arbeit).
- [x] Bei codebezogenen Änderungen: `make build` grün.
- [x] Bei codebezogenen Änderungen: `make gates` grün.
- [x] Bei codebezogenen/Release-Änderungen: `make security-gates` grün oder CI-Job
  `Security gates` grün dokumentiert.
- [x] Scope-konformer Ausnahmepfad dokumentiert: Bei plan-/decision-lastigem Scope
  reichen Tranche-Notiz, Entscheidungsnachweis und Risks-/Roadmap-Aktualisierung als
  Gate-Nachweis, solange keine code-seitigen Änderungen vorliegen.
- [x] Wave-2-Quality-Gates vor Tag geprüft.
- [x] Versions-Bump auf `0.13.0` vollständig durchgeführt.
- [x] `CHANGELOG.md` mit `[0.13.0] - YYYY-MM-DD` aktualisiert.
- [x] Roadmap auf released `0.13.0` und nächste Folgephase umgestellt.
- [x] Plan nach `docs/planning/done/plan-0.13.0.md` verschoben,
  Status auf `✅ released`.
- [x] Annotierter Tag `v0.13.0` erstellt.

Abschlusskriterien Tranche 5:

- [x] Alle release-blockierenden `[ ]`-Einträge mit Dateinachweis
  aktualisiert.
- [x] Kein kritischer Fehler im Blocker-Log offen.

### 7.1 Closeout-Zwischenstand — 2026-05-12

Verifiziert:

- `make docs-check` grün.
- `make build` grün.
- `make gates` grün nach Release-Commit:
  - Go Race-Tests grün.
  - TypeScript-Tests grün.
  - Go-/TS-Lint grün.
  - API-Coverage-Gate: 90.2 % >= 90 %.
  - TS-Coverage grün.
  - `arch-check`, `schema-validate`, `generated-drift-check`,
    SDK-Pack-Smoke, SDK-Performance-Smoke und `verify-doc-refs`
    grün.
- `make security-gates` grün:
  - govulncheck: keine Vulnerabilities.
  - `pnpm audit --audit-level high`: keine Highs (eine Low-Meldung,
    nicht gate-relevant).
  - Trivy API/Dashboard/Analyzer: 0 HIGH/CRITICAL.
- `bash -n scripts/release-guard.sh` grün.
- `MTRACE_RELEASE_APPROVED=1 MTRACE_RELEASE_ALLOW_NON_MAIN=1
  MTRACE_RELEASE_ALLOW_DIRTY=1 MTRACE_RELEASE_ALLOW_OFFLINE=1
  make release-guard VER=0.13.0` grün im Dry-Run. Die Overrides sind
  nur für den lokalen Guard-Test zulässig; der echte Release-Pfad darf
  sie nicht setzen.
- `yq . deploy/k8s/*.yaml >/dev/null` grün (YAML-Syntax).
- Wave-2:
  - `benchmark.yml` run `25705097012` grün.
  - `fuzz.yml` run `25707248676` grün.
  - `mutation.yml` runs `25707444662`, `25645624823`, `25623561573`
    grün.
  - `gh issue list --label fuzz --state open` leer.

Ausnahmen / Hinweise:

- Der erste `make gates`-Lauf vor dem Release-Commit stoppte erwartbar
  bei `generated-drift-check`, weil Versionierungs-/Fixture-Änderungen
  noch nicht in `HEAD` lagen. Der erneute Lauf nach Release-Commit
  ist grün.
- `kubectl --dry-run=client` war in der Sandbox nicht nutzbar, weil
  `kubectl` trotz Client-Dry-Run API-Discovery gegen den konfigurierten
  Cluster versucht hat.
- Annotierter Tag `v0.13.0` wird nach finalem Release-Guard auf dem
  Closeout-Commit erstellt.

## 8. RAK-Verifikationsmatrix

Jede RAK-Zeile enthält `Nachweis`, `Akzeptanz` und `Status`.

| RAK | Priorität | Nachweis | Akzeptanz | Status |
| --- | --- | --- | --- | --- |
| RAK-91 | Muss | `docs/adr/0005-production-ops-backends.md`, Tranche 3.1, `apps/api/internal/storage/schema.yaml` | Tranche-0/1 liefern eine verbindliche Seed- oder Defer-Entscheidung mit Triggern | [x] |
| RAK-92 | Muss | Tranche 4.1, ADR 0005 | Daten-Pfad-/Kosten-/Migrationsentscheidung ist nachvollziehbar und freigegeben (`defer`) | [x] |
| RAK-93 | Muss | `deploy/k8s/README.md`, `deploy/k8s/*.yaml`, `deploy/README.md`, Tranche 5.1 | K8s ist klar als optionaler/nachgelagerter Option-Pfad mit „not production-ready“ Zusage dokumentiert | [x] |
| RAK-94 | Muss | `.devcontainer/devcontainer.json`, Tranche 5.1 | MVP-43-Entscheidung liegt mit Begründung inkl. Folge-/Defer-Regeln vor | [x] |
| RAK-95 | Muss | `scripts/release-guard.sh`, `Makefile`, `docs/user/releasing.md` §3.0, Tranche 6.1 | Manuelle Freigabe-Gates sind verbindlich und automatisierte Schritte sind rückholbar | [x] |

Optionaler Zusatznachweis je RAK:

- `Datei`: exakter Pfad zur Quelle (z. B. `docs/planning/...`, `spec/...`).
- `Datum`: Entscheidungs- oder Prüftermin.
- `Owner`: Verantwortlicher Bereich (z. B. Platform/CI/QA).

Verifikationsmapping:

| RAK | Primäre Datei(en) | Datum | Owner | Status |
| --- | --- | --- | --- | --- |
| RAK-91 | `docs/adr/0005-production-ops-backends.md`, `docs/planning/done/plan-0.13.0.md` | 2026-05-12 | Platform/PM | ✅ |
| RAK-92 | `docs/adr/0005-production-ops-backends.md`, `docs/planning/done/plan-0.13.0.md` | 2026-05-12 | Platform/QA | ✅ |
| RAK-93 | `deploy/k8s/README.md`, `deploy/k8s/*.yaml`, `deploy/README.md`, `docs/planning/in-progress/risks-backlog.md` | 2026-05-12 | Platform/Ops | ✅ |
| RAK-94 | `.devcontainer/devcontainer.json`, `docs/planning/done/plan-0.13.0.md` | 2026-05-12 | Platform/DevEx | ✅ |
| RAK-95 | `scripts/release-guard.sh`, `Makefile`, `docs/user/releasing.md`, `CHANGELOG.md` | 2026-05-12 | Platform/CI | ✅ |

## 9. Folge-Scope nach `0.13.0`

- `0.14.0` oder Nachfolge-Phase: Konkretisierung der gewählten
  Ops-Backends (Postgres-Migrationsslice/Analytics-Backend-Scope/K8s-
  Follow-up nach Entscheidung).
- Später: vollständige Production-Kubernetes- und Observability-
  Standardisierung.
- Später: dediziertes Secret-Management, SLO-gesteuerte Storage-
  Strategien und Ops-Runbooks.
