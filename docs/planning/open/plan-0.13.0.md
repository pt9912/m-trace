# Implementation Plan — `0.13.0` (Production / Ops Backends)

> **Status**: ⬜ open — noch nicht aktiviert. Dieser Plan darf erst nach
> explizitem Move nach `docs/planning/in-progress/` umgesetzt werden.
> Vorgänger ist voraussichtlich `0.12.0`; Aktivierung erst nach dessen
> Release-Closeout.
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
> [`spec/lastenheft.md`](../../../spec/lastenheft.md) NF-18,
> MVP-40..MVP-44; [`docs/planning/in-progress/roadmap.md`](../in-progress/roadmap.md)
> Folge-ADRs; [`docs/planning/in-progress/risks-backlog.md`](../in-progress/risks-backlog.md)
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

- `0.12.0` ist released; Roadmap muss auf `0.13.0` als aktive
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

### 0.4 Lastenheft-Patch (Vorschlag — Patch-Nr und RAK-IDs werden bei T0-Aktivierung neu vergeben)

Der Plan ergänzt eine neue RAK-Serie für `MVP-40`..`MVP-44` und
`NF-18` als Mindestverpflichtung für 0.13.

> **Hinweis zur ID-Vergabe (Stand 2026-05-11)**: Die ursprünglich
> in diesem Plan platzhalterhaft vorgesehene Range `RAK-77`..`RAK-81`
> wurde mit der Aktivierung von `0.12.5` (Auth-/Ingest-Adapter,
> Lastenheft-Patch `1.1.16`, §13.15) belegt. Bei der `0.13.0`-T0-
> Aktivierung werden Patch-Nummer und RAK-IDs entsprechend neu
> vergeben — voraussichtlich Lastenheft-Patch `1.1.17` und neue
> RAK-Gruppe `RAK-83`..`RAK-87` in §13.16. Die Inhalts-Tabellen
> unten behalten die ursprünglichen RAK-77..RAK-81-Labels als
> **Platzhalter**, bis T0 die finale Vergabe durchgeführt hat.

| RAK (Platzhalter) | Priorität | Inhalt |
| --- | --- | --- |
| RAK-77 → RAK-83 (vmtl.) | Muss | Operativer Scope für `MVP-40`..`MVP-42`, inkl. klarer Seed-/Defer-Boundaries und Nachweise. |
| RAK-78 → RAK-84 (vmtl.) | Muss | Operativer Scope für `MVP-41`: Vergleichs- und Entscheidungspfad gegen ClickHouse/VictoriaMetrics/Mimir (oder gleichwertige Option). |
| RAK-79 → RAK-85 (vmtl.) | Muss | `MVP-42`/`NF-18` als optionaler K8s-Optionpfad wird normativ begrenzt, keine Vollbereitschaftszusage. |
| RAK-80 → RAK-86 (vmtl.) | Muss | Operativer Scope für `MVP-43`: Devcontainer als Seed oder explizites Defer mit Begründung/Trigger. |
| RAK-81 → RAK-87 (vmtl.) | Muss | Release-Prozess enthält mindestens manuelle Freigabe in allen automationsrelevanten Stufen, inklusive sicherer Rollback-Regeln und Closeout-Nachweis. |

> Die finale Verifizierung der IDs erfolgt gegen `spec/lastenheft.md`
> beim `0.13.0`-T0-Closeout. Bis dahin gelten die obigen Mappings
> als Vorbesetzung, nicht als verbindliche IDs.

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
| 0 | Aktivierung, Lastenheft-Patch, ADR-Schnitt und Scope-Gates | Freigabevoraussetzungen geklärt, Scope fixiert | Tranche-Ready | Freigabe- und Blockerliste | ⬜ |
| 1 | Postgres-Entscheidung und Adapter-Scope (`MVP-40`) | Entscheidung inkl. Seed-Umfang oder strukturierte Defer-Liste | Geklärter Scope | Migrations-/Defer-Entscheidung + Trigger | ⬜ |
| 2 | Analytics-Backend-Entscheidung (`MVP-41`) | Vergleichsmatrix + klare Pfadentscheidung | Persistenz-/Query-Anforderungen | Entscheidungsprotokoll | ⬜ |
| 3 | Kubernetes-/Devcontainer-/NF-18-Risiko (`MVP-42`, `MVP-43`, `R-9`, `NF-18`) | Konkreter Scope-Hebel für Folgephase | Tranche-1/2 Ergebnisse | Operativer Optionspfad + Defer-Regeln | ⬜ |
| 4 | Release-Automatisierung (`MVP-44`) | Sichere Automations- und Freigaberegeln | Zielbild Release-Prozess | Runbook + Gateschema | ⬜ |
| 5 | Gates, RAK-Matrix, Versions-Bump, Closeout und Tag | Abschluss nachweisbar und wiederholbar | Tranche-4 Ergebnis | RAK-Status Grün + Release-Nachweise | ⬜ |

## 2. Tranche 0 — Aktivierung, Scope-Gates und Plan-Härtung

Ziel: Voraussetzungen, Scope und Architekturbezug sind so klar,
 dass keine Entscheidung im mittleren Tranchefluss erneut gestoppt werden
 kann.

DoD:

- [ ] Plan von `docs/planning/open/plan-0.13.0.md` nach
  `docs/planning/in-progress/plan-0.13.0.md` verschoben.
- [ ] Ausgangszustand von `git status --short` dokumentiert und im
  Tranche-0-Notizblock gespeichert.
- [ ] Lastenheft-Patch mit neuer RAK-Gruppe für `MVP-40`..`MVP-44`
  und `NF-18` ergänzt.
- [ ] ADR-Schnitt definiert: Architektur-/Persistenz-/Release-
  Entscheidungen (wir liefern ADR oder Plan-DoD).
  - [ ] `decision-record` enthält mindestens zwei Alternativen je Kernfrage.
- [ ] `docs/planning/in-progress/roadmap.md` auf `0.13.0` als aktive
  Folgephase umgestellt.
  - [ ] Alle offenen Gegenargumente dokumentiert und nicht stillschweigend verworfen.
- [ ] Neue RAK-Range in `spec/lastenheft.md` persistiert und gegen
  bestehende IDs geprüft (Platzhalter-Mapping in §0.4 oben:
  voraussichtlich `RAK-83`..`RAK-87` in §13.16, da `RAK-77`..`RAK-82`
  bereits mit `0.12.5` Lastenheft-Patch `1.1.16` §13.15 vergeben).
- [ ] Risiko-R-9 gegen Observability-/Smoke-Impact geprüft und im
  `risks-backlog.md` abgelegt.
- [ ] Tranche-0-Notiz (kurz) mit Status, offenen Entscheidungen,
  offenen Triggern und Ausnahmen angelegt.
  - [ ] Notiz enthält erwartete Entscheidungstermine und Verantwortliche.

Go/No-Go-Kriterien nach Tranche 0:

- **Go:** Scope-Sätze sind vollständig und keine neue Pflichtabhängigkeit
  (Dependency) wird eingeschoben.
- **No-Go:** Unklare ADR-Zuordnung, ungeklärte R-9-Interaktion oder
  unbewertete Pflicht-Trigger.

## 3. Tranche 1 — Postgres (MVP-40)

Ziel: Entweder erster produktionsnaher Slice wird validiert oder eine
saubere Defer-Entscheidung trifft die Ausnahmen.

DoD:

- [ ] Entscheidung dokumentiert: vollständige Implementierung,
  minimaler Seed-Slice oder Defer.
  - [ ] Entscheidung enthält „Entscheidung“, „Begründung“, „Nicht entschieden“. 
- [ ] Wenn Seed implementiert: mindestens eine deterministische
  Migrationsspur definiert (Schema/Repo/Adapter) inklusive Daten-
  Rückwärtskompatibilitätsprüfung mit SQLite.
  - [ ] Migrationspfad enthält `migrate up`, `rollback`, `replay`.
- [ ] Wenn Seed nicht implementiert: Triggerschwellen (z. B.
  Last, SLA, Betriebszeit, Recovery-Szenarien) festgelegt.
  - [ ] Jeder Trigger hat einen messbaren Schwellwert + Besitzer.
- [ ] Contract-/Integrationstests pinnen SQLite-Verhalten, damit Postgres
  nicht unfreiwillig als versteckte Pflichtwirkung eingeführt wird.
- [ ] Betriebsrisiken dokumentiert: Datenkonsistenz bei Migration,
  Ausfall- und Recoveryverhalten, Backup/Restore-Bezug.
- [ ] Ergebnis entscheidet, welche Pfade in Tranche 5 in RAK/Changelog
  aufgenommen werden.
  - [ ] Entscheidung wird vor Tranche-2-Start freigegeben.

## 4. Tranche 2 — Analytics-Backend (MVP-41)

Ziel: Datenpfadbedarf zwischen „kein Pflichtbackend“, „POC“ und
„konkreter Folgepfad“ trennen.

DoD:

- [ ] ClickHouse/VictoriaMetrics/Mimir-Bedarf anhand aktueller
  Datenpfade, Query-Anforderungen und Datenvolumen bewertet.
- [ ] Vergleichsmatrix liegt vor: Komplexität, Betriebskosten,
  Abfragefähigkeit, Integrationsaufwand, Relevanz der Query-Workloads,
  Migrationsrisiko.
- [ ] Entscheidung dokumentiert als `proceed` / `defer` / `POC`.
  - [ ] Matrix enthält einen „Wenn-Pilot→Go“-Abschnitt mit zeitlicher Grenze.
- [ ] Keine neue lokale Pflichtabhängigkeit.
- [ ] Falls POC: Zielbild, Erfolgskriterien, Erfolgspunkte,
  Abbruchkriterien im Plan verankert.
  - [ ] POC-Report enthält Kostenannahmen, Datenmodell-Deckung und
    Rechenlastabschätzung.

## 5. Tranche 3 — Kubernetes- und Devcontainer-Scope (MVP-42, MVP-43, NF-18, R-9)

Ziel: K8s- und Devcontainer-Scope wird normativ abgegrenzt und ein
entscheidbarer Folgepfad garantiert.

DoD:

- [ ] `NF-18` und `MVP-42` sind harmonisiert: klarer Option-Pfad
  (`konkret + optional` oder `folge`) und dokumentiert.
- [ ] R-9-Auswirkungen auf Observability-Label-Allowlists und
  Smoke-Stage vollständig geprüft; Ergebnis im Risks-Backlog und/oder
  ADR.
  - [ ] Risiko-Matrix enthält mindestens zwei konkrete Gegenmaßnahmen.
- [ ] Devcontainer (`MVP-43`) entschieden: liefern Beispielkonfiguration
  oder deferred mit konkreten Gründen.
- [ ] README-/User-Doku-Abgrenzung bleibt konsistent: keine
  Produktions-Ready-K8s-Zusage.
- [ ] Scope-Transitie auf nachfolgende Tranche (`0.14.0` o. ä.)
  klar beschrieben.
  - [ ] Jede Option hat einen klaren Auslöser für die Folgephase.

## 6. Tranche 4 — Release-Automatisierung (MVP-44)

Ziel: Automations-Mechanik für Release ohne Sicherheitsrisiko und ohne
unbemerkte automatische Veröffentlichung definieren.

DoD:

- [ ] Entscheidbare Automation ausgewählt: was wird automatisiert,
  was bleibt manuell.
- [ ] Automationsumfang in einem RACI-/Owner-Mapping hinterlegt.
- [ ] Sichere Freigabe-Guards definiert:
  - Branch- und Tag-Muster,
  - Freigabekanal (Reviewer/Freigabe-
    Kontrolle),
  - zeitliche oder environment-gesteuerte Restriktionen.
- [ ] Release-Doku bleibt die Source-of-Truth; automatischer Teil nur als
  Ergänzung mit Rückholbarkeit.
- [ ] Automations- oder Dry-Run-Tests sind definiert (ohne reale
  Veröffentlichung).
- [ ] CI-/Local-Runbook beschreibt minimalen sicheren Ausführungsweg.
  - [ ] Runbook enthält Notfallplan bei fehlgeschlagenem Guard-Check.

## 7. Tranche 5 — Release-Closeout und Abschluss

Ziel: Alle Gates sind nachweisbar erfüllt und der Release kann
sauber abgeschlossen werden.

DoD:

- [ ] RAK-Verifikationsmatrix vollständig ausgefüllt.
- [ ] `make docs-check` grün (oder dokumentarischer Äquivalent-Gate bei reiner
  Plan-/ADR-/Spec-Arbeit).
- [ ] Bei codebezogenen Änderungen: `make build` grün.
- [ ] Bei codebezogenen Änderungen: `make gates` grün.
- [ ] Bei codebezogenen/Release-Änderungen: `make security-gates` grün oder CI-Job
  `Security gates` grün dokumentiert.
- [ ] Scope-konformer Ausnahmepfad dokumentiert: Bei plan-/decision-lastigem Scope
  reichen Tranche-Notiz, Entscheidungsnachweis und Risks-/Roadmap-Aktualisierung als
  Gate-Nachweis, solange keine code-seitigen Änderungen vorliegen.
- [ ] Wave-2-Quality-Gates vor Tag geprüft.
- [ ] Versions-Bump auf `0.13.0` vollständig durchgeführt.
- [ ] `CHANGELOG.md` mit `[0.13.0] - YYYY-MM-DD` aktualisiert.
- [ ] Roadmap auf released `0.13.0` und nächste Folgephase umgestellt.
- [ ] Plan nach `docs/planning/done/plan-0.13.0.md` verschoben,
  Status auf `✅ released`.
- [ ] Annotierter Tag `v0.13.0` erstellt.

Abschlusskriterien Tranche 5:

- [ ] Alle `[ ]`-Einträge mit Dateinachweis aktualisiert.
- [ ] Kein kritischer Fehler im Blocker-Log offen.

## 8. RAK-Verifikationsmatrix (Vorschau)

Wird während der Umsetzung gefüllt. Jede RAK-Zeile enthält
`Nachweis`, `Akzeptanz` und `Status`.

| RAK | Priorität | Nachweis | Akzeptanz | Status |
| --- | --- | --- | --- | --- |
| RAK-77 | Muss | Scope-Dokumente, Entscheidungs-ADR,
  Adapter-/Migrationsartefakte oder klare Defer-Trigger | Tranche-0/1 liefern eine verbindliche Seed- oder Defer-Entscheidung mit Triggern | [ ] |
| RAK-78 | Muss | Vergleichsmatrix, Architekturentscheidung,
  ggf. PoC-Bericht | Daten-Pfad-/Kosten-/Migrationsentscheidung ist nachvollziehbar und freigegeben (`proceed`/`defer`/`POC`) | [ ] |
| RAK-79 | Muss | NF-18-Harmonisierungs-Notiz + K8s-Decision-Record
  + README-Abgrenzung | K8s ist klar als optionaler/nachgelagerter Option-Pfad mit „not production-ready“ Zusage dokumentiert | [ ] |
| RAK-80 | Muss | Devcontainer-Entscheidung und ggf. Beispielartefakt
  oder Defer-Justification | MVP-43-Entscheidung liegt mit Begründung inkl. Folge-/Defer-Regeln vor | [ ] |
| RAK-81 | Muss | Freigabe-Guard, Automationsumfang,
  Release-Runbook und Dry-Run-Test | Manuelle Freigabe-Gates sind verbindlich und automatisierte Schritte sind rückholbar | [ ] |

Optionaler Zusatznachweis je RAK:

- `Datei`: exakter Pfad zur Quelle (z. B. `docs/planning/...`, `spec/...`).
- `Datum`: Entscheidungs- oder Prüftermin.
- `Owner`: Verantwortlicher Bereich (z. B. Platform/CI/QA).

Sofort nutzbares Verifikationsmapping (auszufüllen):

| RAK | Primäre Datei(en) | Datum | Owner | Status |
| --- | --- | --- | --- | --- |
| RAK-77 | `spec/lastenheft.md`, `docs/planning/in-progress/plan-0.13.0.md`, `docs/planning/in-progress/roadmap.md` | | Platform/PM | [ ] |
| RAK-78 | `docs/planning/in-progress/plan-0.13.0.md`, `docs/planning/in-progress/risks-backlog.md` | | Platform/QA | [ ] |
| RAK-79 | `spec/lastenheft.md`, `docs/planning/in-progress/roadmap.md`, `docs/planning/in-progress/risks-backlog.md` | | Platform/Ops | [ ] |
| RAK-80 | `docs/planning/in-progress/plan-0.13.0.md` | | Platform/DevEx | [ ] |
| RAK-81 | `docs/planning/in-progress/plan-0.13.0.md`, `CHANGELOG.md` | | Platform/CI | [ ] |

## 9. Folge-Scope nach `0.13.0`

- `0.14.0` oder Nachfolge-Phase: Konkretisierung der gewählten
  Ope-Backends (Postgres-Migrationsslice/Analytics-Backend-Scope/K8s-
  Follow-up nach Entscheidung).
- Später: vollständige Production-Kubernetes- und Observability-
  Standardisierung.
- Später: dediziertes Secret-Management, SLO-gesteuerte Storage-
  Strategien und Ops-Runbooks.
