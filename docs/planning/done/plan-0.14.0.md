# Implementation Plan — `0.14.0` (Ops Backend Follow-up)

> **Status**: ✅ released 2026-05-12 — Tranchen 0..5 geschlossen,
> Release-Tag `v0.14.0`.
>
> **Vorgänger**: `0.13.0` (Production / Ops Backends), released
> 2026-05-12; Plan in
> [`docs/planning/done/plan-0.13.0.md`](../done/plan-0.13.0.md).
>
> **Release-Typ**: Minor-Release mit Lastenheft-Patch, neuer
> RAK-Gruppe und Tag `v0.14.0`. RAK-Range:
> `RAK-96`..`RAK-100`.
>
> **Ziel**: Die in `0.13.0` getroffenen Ops-Backend-Entscheidungen
> in konkrete, lieferbare Umsetzungs-, Hardening- oder Defer-Slices
> überführen. `0.14.0` ist kein erneuter Bewertungsplan, sondern der
> erste Folgeplan für die in `0.13.0` freigegebenen oder
> reaktivierbaren Pfade: Postgres-Migrationsvorbereitung,
> Analytics-Backend-POC/-Trigger, K8s-/NF-18-Option, Devcontainer-
> Reproduzierbarkeit und Release-Automations-Guards.
>
> **Bezug**:
> [`../in-progress/roadmap.md`](../in-progress/roadmap.md),
> [`docs/planning/done/plan-0.13.0.md`](../done/plan-0.13.0.md)
> §9, [`../../../spec/lastenheft.md`](../../../spec/lastenheft.md)
> §13.18 (`RAK-96`..`RAK-100`).
>
> **Nachfolger**: `0.15.0` offen in
> [`../open/plan-0.15.0.md`](../open/plan-0.15.0.md).

## 0. Konvention

DoD-Checkboxen tracken den Lieferstand:

- `[x]` ausgeliefert mit Commit-Hash.
- `[ ]` offen.
- `[!]` blockiert durch ADR- oder Scope-Entscheidung.
- 🟡 in Arbeit.

### 0.1 Scope-Definition

`0.14.0` ist der vorbereitete Folgeplan für die nach `0.13.0`
gewählten Ops-Backend-Pfade. Dieser Plan darf erst aktiviert werden,
wenn `0.13.0` mindestens die Entscheidungen aus `RAK-91`..`RAK-95`
nachweisbar geschlossen hat.

Scope-Regel:

- `0.14.0` übernimmt Implementierungspfade nur, wenn sie im `0.13.0`-
  Closeout als `proceed` oder `POC` markiert sind.
- Bereits gelieferte `seed`-Artefakte werden nur als Hardening-,
  Validierungs- oder Dokumentationsaufgabe übernommen.
- Ein `defer-with-migration-seed` importiert nur Trigger-/DDL-/
  Replay-Vorbereitung und keine Adapter-Implementierung.
- Pfade mit `defer` bleiben dokumentiert, erhalten aber nur Trigger-
  Pflege und keine Implementierungstranche.
- Pfade mit `blocked` bleiben in der Tranche-Tabelle sichtbar und
  müssen vor Aktivierung aufgelöst oder ausdrücklich aus dem Release
  gestrichen werden.

Vorläufig in Scope:

- Umsetzung eines Postgres-Folgepfads, falls `0.13.0` `proceed` oder
  `POC` entscheidet: Migration, Adapter-Slice, Contract-/
  Regressionstests und Rollback-Nachweis. Bei
  `defer-with-migration-seed` bleibt der Scope auf DDL-/Replay-
  Vorbereitung und Trigger-Pflege begrenzt.
- Umsetzung oder POC eines Analytics-Backend-Folgepfads, falls
  `0.13.0` `proceed` oder `POC` entscheidet: begrenztes Datenmodell,
  Ingest-/Exportpfad, Query-Nachweise und Kosten-/Lastgrenzen. Bei
  `defer` bleibt nur Trigger-Pflege in Scope.
- Konkretisierung oder Hardening des K8s-Optionpfads, falls `0.13.0`
  ihn freigibt: fehlende Beispielmanifeste, Smoke-Gate-Entscheidung,
  Seed-Validation und Observability-Label-Harmonisierung.
- DevEx-Folgepfad aus `MVP-43`, falls `0.13.0` keinen vollständigen
  Devcontainer-Seed liefert oder offene Hardening-/Validation-
  Aufgaben empfiehlt.
- Release-Automations-Folgepfad aus `MVP-44`, falls `0.13.0` konkrete
  Dry-Run-/Guard-Schritte freigibt oder offene Guard-Hardening-
  Aufgaben hinterlässt.

Vorläufig out of scope:

- Keine Ablösung von SQLite im lokalen Standardbetrieb in `0.14.0`.
  Postgres darf nur opt-in oder produktionsnaher Zusatzpfad werden;
  eine Änderung des lokalen Standard-Stores braucht einen separaten
  ADR-/Planbeschluss außerhalb dieses vorbereiteten Folgeplans.
- Kein vollständiger Production-Kubernetes-Betrieb.
- Kein Managed-Cloud-Betrieb.
- Kein Multi-Tenant-SaaS-Produkt.
- Keine automatische Veröffentlichung ohne explizite Human Approval.
- Kein Secret-Management-Vollausbau jenseits der bereits gelieferten
  `0.12.x`-Pfade, außer ein `0.13.0`-Closeout zieht ihn ausdrücklich
  als Folge-Scope.
- Keine parallele Einführung mehrerer neuer Betriebsbackends ohne
  explizite Ressourcen- und Risikoentscheidung in Tranche 0.

### 0.1a Entscheidungsimport aus `0.13.0`

Diese Matrix wird bei Aktivierung aus dem `0.13.0`-Closeout befüllt.
Sie ist das Gate gegen Scope-Drift.

| 0.13-Entscheidung | Import nach 0.14 | Default, falls unklar | Pflichtnachweis |
| --- | --- | --- | --- |
| Postgres `proceed`/`POC` | Tranche 1 aktiv | `[!]` blockiert | Migrations-/Rollback-Entscheidung, SQLite-Kompatibilitätsgrenze |
| Postgres `defer-with-migration-seed` | Tranche 1 nur Trigger-/DDL-Vorbereitung | nicht implementieren | Defer-Trigger mit Schwellwert und Owner, bestehender Migrationsanker |
| Postgres `defer` | Tranche 1 nur Trigger-Pflege | nicht implementieren | Defer-Trigger mit Schwellwert und Owner |
| Analytics `proceed`/`POC` | Tranche 2 aktiv | `[!]` blockiert | Backend-Wahl, Datenmodell, Erfolg-/Abbruchkriterien |
| Analytics `defer` | Tranche 2 nur Trigger-Pflege | nicht implementieren | Vergleichsmatrix + Reaktivierungsbedingung |
| K8s `option` ohne Seed | Tranche 3 aktiv | `[!]` blockiert | R-9-Entscheidung und mindestens zwei Gegenmaßnahmen |
| K8s `option` mit Seed | Tranche 3 nur Hardening/Validation | nicht neu implementieren | Seed-Artefakte, Nicht-Production-Ready-Abgrenzung, R-9-Trigger |
| K8s `defer` | Tranche 3 nur Dokumentationspflege | nicht implementieren | klare Nicht-Production-Ready-Abgrenzung |
| Devcontainer `seed` | Tranche 4 nur Hardening/Validation | nicht neu implementieren | lokale Standardentwicklungs-Abgrenzung, Seed-Artefakt |
| Release-Automation `guard` | Tranche 4 nur Hardening/Validation | nicht neu implementieren | Human-Approval-Gate und Dry-Run-Nachweis |

### 0.2 Vorgänger-Gate

Vor Aktivierung von `0.14.0` müssen diese Bedingungen erfüllt sein:

- [x] `0.13.0` ist released und als
  `docs/planning/done/plan-0.13.0.md` archiviert.
- [x] Roadmap zeigt `0.14.0` als aktive Folgephase oder begründet
  einen anderen Nachfolger.
- [x] RAK-91..RAK-95 sind in der `0.13.0`-Verifikationsmatrix
  geschlossen oder mit explizitem Defer-/Blockerstatus versehen.
- [x] Für jeden übernommenen Pfad existiert eine `0.13.0`-
  Entscheidung mit `Entscheidung`, `Begründung`, `Nicht entschieden`
  und Triggern.
- [x] Keine neue lokale Pflichtabhängigkeit wird ohne Lastenheft-Patch
  und Migrations-/Rollback-Nachweis eingeführt.

### 0.3 Lastenheft-Patch (`1.1.19`, RAK-96..RAK-100)

T0-Aktivierung 2026-05-12: `0.13.0` belegt `RAK-91`..`RAK-95`
in §13.17. Damit ist `0.14.0` verbindlich auf Lastenheft-Patch
`1.1.19` und `RAK-96`..`RAK-100` in §13.18 gesetzt.

| RAK | Thema | Bedingung |
| --- | --- | --- |
| RAK-96 | Postgres-Triggerpflege | `0.13.0` entschied `defer-with-migration-seed`; `0.14.0` darf keinen Runtime-Adapter aktivieren, sondern muss DDL-/Replay-/Triggergrenzen konkret halten. |
| RAK-97 | Analytics-Triggerpflege | `0.13.0` entschied `defer`; `0.14.0` hält Query-/Kosten-/POC-Trigger messbar, ohne ein Pflichtbackend einzuführen. |
| RAK-98 | K8s-/NF-18-Seed-Hardening | `0.13.0` lieferte optionale K8s-Beispiele; `0.14.0` validiert und härtet den Optionspfad ohne Production-Ready-Zusage. |
| RAK-99 | Devcontainer-/DevEx-Validation | `0.13.0` lieferte den Devcontainer-Seed; `0.14.0` validiert ihn als Zusatzpfad, nicht als Ersatz für Make/Docker. |
| RAK-100 | Release-Guard-Hardening | `0.13.0` lieferte den lokalen Release-Guard; `0.14.0` validiert Dry-Run, Fehlerfälle und Runbook-Konsistenz. |

### 0.4 Qualitätsregeln für `0.14.0`

- Ein neuer Backend-Pfad darf nur opt-in aktiviert werden, bis
  Contract-, Migration- und Rollback-Nachweise vorliegen.
- SQLite bleibt in Tests und lokaler Standardentwicklung der
  Compatibility-Anker.
- Jeder POC hat eine harte Zeitgrenze und explizite Abbruchkriterien.
- K8s-Manifeste dürfen nicht als Production-Ready-Dokumentation
  formuliert werden, solange kein separater Production-Betriebsplan
  existiert.
- Release-Automation darf Artefakte bauen, prüfen und dry-runnen,
  aber nicht ohne explizite Freigabe veröffentlichen.
- Jede Tranche endet mit einem `What ändert sich` / `What bleibt
  unverändert`-Block.

### 0.5 Tranche-Output-Verpflichtungen

Jede aktive Umsetzungs- oder Hardening-Tranche liefert mindestens:

- **Entscheidungsnachweis**: übernommene `0.13.0`-Entscheidung plus
  lokale `0.14.0`-Scope-Bestätigung.
- **Artefaktnachweis**: Code, Manifest, Runbook, POC-Report oder
  Defer-Notiz mit Dateipfad.
- **Gatenachweis**: passender Test, Smoke, Dry-Run oder begründeter
  Doku-only-Gate.
- **Risikostatus**: Update im Risks-Backlog oder explizite Aussage,
  dass kein neuer Risiko-Trigger ausgelöst wurde.

### 0.6 Aktivierungsszenarien

Tranche 0 wählt genau eines dieser Aktivierungsszenarien:

| Szenario | Inhalt | Release-Charakter | Go-Bedingung |
| --- | --- | --- | --- |
| A | Postgres-Slice + Release-Guard-Hardening | fokussierter Storage-/CI-Release | RAK-91 gibt Umsetzung oder `defer-with-migration-seed` frei; RAK-95 enthält offene Guard-Folgeaufgaben |
| B | Analytics-POC + optionale K8s-Doku | POC-/Decision-Release | RAK-92 gibt POC frei, RAK-93 ist nicht blockiert |
| C | K8s/DevEx/Release-Guard-Hardening | Ops-Enablement-Release ohne neue Pflichtpfade | RAK-93..RAK-95 enthalten offene Folgeaufgaben nach den `0.13.0`-Seeds |
| D | Trigger-/Defer-Release | rein dokumentarisch | 0.13 deferred alle Implementierungspfade |

Mehr als zwei große Implementierungspfade in einem `0.14.0`-Release
gelten als No-Go, außer Tranche 0 dokumentiert explizit zusätzliche
Kapazität und getrennte Gate-Nachweise.

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Erwartetes Ergebnis | Eingang | Ausgang | Status |
| --- | --- | --- | --- | --- | --- |
| 0 | Aktivierung, RAK-Zuschnitt und Vorgänger-Entscheidungen | Scope aus `0.13.0` verbindlich übernommen | `0.13.0` released | Finaler 0.14-Scope | ✅ |
| 1 | Postgres-Migrations-/Adapter-Slice | Trigger-/DDL-/Replay-Grenzen gepflegt; kein Runtime-Adapter | RAK-91-Ergebnis | Migrations-/Rollback-/Trigger-Nachweis | ✅ |
| 2 | Analytics-Backend-Slice, POC oder Trigger-Pflege | Query-/Kosten-/POC-Trigger gepflegt; kein Pflichtbackend | RAK-92-Ergebnis | POC-Report, Adapter-Slice oder Defer-Notiz | ✅ |
| 3 | K8s-/NF-18-Optionpfad und R-9 | Optionaler K8s-Pfad oder Seed-Hardening ohne Production-Ready-Zusage | RAK-93-Ergebnis | Manifest-/Smoke-/Risiko-Nachweis | ✅ |
| 4 | Devcontainer und Release-Automations-Guards | Reproduzierbare DevEx und sichere Release-Dry-Runs oder Seed-Validation | RAK-94/95-Ergebnis | Runbook-/Guard-Artefakte | ✅ |
| 5 | Gates, RAK-Matrix, Versions-Bump, Closeout und Tag | Release nachweisbar abgeschlossen | letzte aktive Tranche | Tag `v0.14.0` | ⬜ |

## 2. Tranche 0 — Aktivierung und Scope-Härtung

Ziel: Der offene Plan wird nach `0.13.0` in einen entscheidbaren
Umsetzungsplan überführt, ohne die `0.13.0`-Entscheidungen zu
überstimmen.

DoD:

- [x] Plan von `docs/planning/open/plan-0.14.0.md` nach
  `docs/planning/done/plan-0.14.0.md` verschoben.
- [x] Ausgangszustand von `git status --short` dokumentiert.
- [x] `0.13.0`-Closeout gelesen und alle übernommenen Pfade explizit
  auf `proceed`, `POC`, `defer` oder `blocked` gemappt.
- [x] Aktivierungsszenario A/B/C/D ausgewählt und begründet.
- [x] Aktive und deferred Tranchen für das gewählte Szenario in einer
  Tranchenmatrix festgelegt.
- [x] Nicht übernommene `0.13.0`-Pfade bleiben als Defer-Trigger
  sichtbar, werden aber nicht stillschweigend implementiert.
- [x] Lastenheft-Patch mit finaler RAK-Range ergänzt.
- [x] Roadmap auf `0.14.0` als aktive Folgephase umgestellt.
- [x] Risiken-Backlog aktualisiert, insbesondere R-9 und alle durch
  Postgres/Analytics/K8s neu ausgelösten Betriebsrisiken.
- [x] No-Go-Liste geprüft:
  - unklare Backend-Pflichtabhängigkeit,
  - fehlender Rollbackpfad,
  - fehlende Human Approval im Release-Pfad,
  - K8s-Production-Ready-Sprache ohne Betriebsplan.

### 2.1 Aktivierungsnotiz — 2026-05-12

| Feld | Wert |
| --- | --- |
| Aktivierungsdatum | 2026-05-12 |
| Ausgangs-Commit | `4178f52` (`v0.13.0`) |
| Gewähltes Szenario | C — K8s/DevEx/Release-Guard-Hardening |
| Übernommene 0.13-Pfade | K8s-Seed, Devcontainer-Seed, Release-Guard-Seed |
| Explizit deferred | Postgres Runtime-Adapter; Analytics Pflichtbackend/POC |
| Blocker | keine release-blockierenden Blocker zum Start |
| Required Gates | `make docs-check`; bei Code-/Guard-Änderungen zusätzlich `make build`, `make gates`, `make security-gates` nach Risiko |

Entscheidungsimport:

| 0.13-Pfad | Import nach 0.14 | Begründung |
| --- | --- | --- |
| Postgres | `defer-with-migration-seed`; Tranche 1 nur Trigger-/DDL-/Replay-Grenzen | ADR 0005 nennt `schema.yaml` als Migrationsanker, aber keinen ausgelösten Multi-Replica-/Recovery-/Retention-Trigger. |
| Analytics | `defer`; Tranche 2 nur Query-/Kosten-/POC-Triggerpflege | ADR 0005 und Plan 0.13.0 schließen ClickHouse/VictoriaMetrics/Mimir ohne Pflichtbackend. |
| K8s | `option with seed`; Tranche 3 aktiv für Seed-Hardening/Validation | `deploy/k8s/` existiert, ist aber bewusst nicht production-ready und braucht R-9-Kontrolle. |
| Devcontainer | `seed`; Tranche 4 aktiv für Validation | `.devcontainer/devcontainer.json` existiert als Zusatzpfad; Make/Docker bleiben Standard. |
| Release-Automation | `guard`; Tranche 4 aktiv für Guard-Hardening | `scripts/release-guard.sh` existiert und soll Fehlerfälle/Dry-Run-Konsistenz behalten. |

Aktive/deferred Tranchen:

| Tranche | T0-Status | Liefergrenze |
| --- | --- | --- |
| 1 Postgres | deferred/Triggerpflege | keine Adapter-Implementierung, keine DSN-Pflicht, keine SQLite-Ablösung |
| 2 Analytics | deferred/Triggerpflege | kein ClickHouse-/VictoriaMetrics-/Mimir-POC ohne neuen Trigger |
| 3 K8s/R-9 | aktiv | Seed-Validation, optionale Smoke-Entscheidung, Label-Risiko; keine Production-Ready-Zusage |
| 4 Devcontainer/Release-Guard | aktiv | DevEx-Validation und Guard-Fehlerfälle; keine automatische Veröffentlichung |

No-Go-Prüfung:

- Keine neue Backend-Pflichtabhängigkeit: erfüllt.
- Rollbackpfad: für Seeds bleibt Entfernen/Deaktivieren vor Publish möglich;
  Runtime-Adapter werden nicht aktiviert.
- Human Approval: Release-Guard bleibt freigabepflichtig.
- K8s-Production-Sprache: verboten; alle K8s-Artefakte bleiben Beispiele.

What ändert sich:

- `0.14.0` ist aktiv und normativ mit RAK-96..RAK-100 verankert.
- Der Release-Scope ist auf Szenario C begrenzt; Postgres/Analytics
  werden nicht stillschweigend implementiert.

What bleibt unverändert:

- SQLite und Compose bleiben Standardpfade.
- K8s, Devcontainer und Release-Guard bleiben optionale bzw.
  freigabepflichtige Zusatzpfade.
- Keine Veröffentlichung ohne explizite menschliche Freigabe.

## 3. Tranche 1 — Postgres-Folgepfad

Ziel: Der aus `0.13.0` übernommene Postgres-Pfad wird entweder
umgesetzt, als zeitbegrenzter POC gefahren oder final deferred.

DoD:

- [x] `0.13.0`-Entscheidung zu `MVP-40` liegt vor.
- [x] Entscheiden, ob `0.14.0` einen POC, einen schmalen
  produktionsnahen Adapter-Slice, eine reine DDL-/Replay-Vorbereitung
  oder nur Trigger-Pflege liefert.
- [x] Migrationsmodell definiert: `migrate up`, `rollback`, `replay`
  und Kompatibilitätsgrenze zu SQLite.
- [x] Schema-Differenzen zwischen SQLite und Postgres dokumentiert
  (Zeittypen, IDs, Constraints, Transaktionen, Pagination-Sortierung).
- [x] Adapter-Scope auf minimale Ports und Queries begrenzt.
- [x] Contract- und Regressionstests belegen, dass SQLite der lokale
  Default bleibt.
- [x] Backup-/Restore- und Ausfallverhalten dokumentiert.
- [x] Reaktivierungs- oder Defer-Trigger mit Owner und Messwerten
  aktualisiert.

Go/No-Go:

- **Go:** genau definierter Datenbereich, reproduzierbare Migration,
  keine Änderung am Default-Store.
- **No-Go:** vollständige Store-Ablösung, implizite Postgres-Pflicht
  für lokale Tests, ungetestete Rollbackannahmen.

Vorläufige Artefakte:

- `docs/adr/` oder Plan-Entscheidungsnotiz für den Postgres-Slice
  oder den `defer-with-migration-seed`-Status.
- Migrations-/Rollback-Dokumentation.
- Adapter-/Repository-Tests oder POC-Report.

### 3.1 Triggerpflege — 2026-05-12

**Entscheidung:** `0.14.0` liefert keinen Postgres-POC und keinen
Runtime-Adapter. Der Pfad bleibt `defer-with-migration-seed`; die
Lieferung ist ein Trigger-/DDL-/Replay-Nachweis.

Artefakt:

- `docs/ops/backend-followup.md` §1

Status:

- Multi-Replica-Trigger nicht ausgelöst: K8s-Beispiele bleiben
  Single-Replica mit `strategy: Recreate`.
- Recovery-SLO nicht ausgelöst: kein verbindliches `RPO <= 15 min`
  oder `RTO <= 30 min`.
- Retention-/Read-Last nicht ausgelöst: kein Bericht über > 10 Mio.
  Events mit p95-Read-Anforderung < 2 s.

Migrationsmodell:

- `migrate up`: Postgres-DDL wird erst in einem Folge-Slice aus
  `apps/api/internal/storage/schema.yaml` erzeugt und gegen SQLite-
  Contract-Tests gespiegelt.
- `rollback`: vor Runtime-Aktivierung ist Rollback das Entfernen des
  Folge-Slices; nach Runtime-Aktivierung braucht es Snapshot-Export/
  Import oder explizit dokumentiertes Zurückschalten.
- `replay`: Folge-Slice muss API-kompatible Event-/Session-Fixtures
  oder SQLite-Snapshots replayen können, ohne Reihenfolge oder
  Project-Scope zu verändern.

What ändert sich:

- Postgres hat einen konkreten `0.14.0`-Trigger- und
  Migration-Grenznachweis.

What bleibt unverändert:

- SQLite bleibt Default; kein DSN-Selector, kein Dual-Write und kein
  Postgres-Adapter landen in diesem Schnitt.

## 4. Tranche 2 — Analytics-Backend-Folgepfad

Ziel: Der in `0.13.0` gewählte Analytics-Pfad wird mit begrenztem
Datenmodell, klaren Abbruchkriterien und Query-Nachweisen konkret.

DoD:

- [x] `0.13.0`-Entscheidung zu `MVP-41` liegt vor.
- [x] Zielbackend oder POC-Variante final bestätigt.
- [x] Datenmodell und Retention-Grenzen definiert.
- [x] Query-Workloads mit erwarteter Last und Kostenannahmen
  dokumentiert.
- [x] Datenfluss klar geschnitten: Realtime-Ingest, Batch-Export,
  Replikation oder synthetischer POC-Load.
- [x] Ingest-/Exportpfad bleibt optional und führt keine lokale
  Pflichtabhängigkeit ein.
- [x] POC-Report oder Implementierungsnachweis enthält
  Erfolgskriterien, Abbruchkriterien und Zeitgrenze.

Go/No-Go:

- **Go:** begrenzter Workload, messbare Query-Anforderung,
  isolierter Betriebspfad.
- **No-Go:** parallele Einführung mehrerer Analytics-Systeme,
  unbounded Retention, Pflichtbetrieb im Standard-Compose-Lab.

Vorläufige Artefakte:

- Vergleichsfortschreibung aus `0.13.0`.
- POC-Report mit Kosten-/Lastannahmen.
- Optionaler Smoke oder synthetischer Load-Nachweis.

### 4.1 Triggerpflege — 2026-05-12

**Entscheidung:** `0.14.0` startet keinen Analytics-POC. ClickHouse,
VictoriaMetrics und Mimir bleiben deferred, bis ein messbarer Workload-
oder Owner-Trigger ausgelöst ist.

Artefakt:

- `docs/ops/backend-followup.md` §2

Status:

- High-Volume-Trigger nicht ausgelöst: kein Bericht über > 50 Mio.
  Events pro Tag.
- API-/Prometheus-Gap nicht ausgelöst: kein benannter Query-Workload
  ist blockiert.
- POC-Readiness nicht ausgelöst: kein Owner, kein maximal 30 Tage
  laufender POC und keine datierten Erfolg-/Abbruchkriterien.

Datenfluss-Grenze:

- Erlaubt für spätere POCs: Batch-Export aus SQLite-Snapshot,
  synthetische Last aus Contract-Fixtures oder isolierter Replay.
- Nicht erlaubt ohne neuen Planbeschluss: Pflicht-Dual-Write im API-
  Hot-Path, Default-Compose-Abhängigkeit, unbounded Retention oder
  mehrere Analytics-Backends im selben POC.

What ändert sich:

- Analytics hat einen konkreten `0.14.0`-Trigger-, Workload- und
  Datenfluss-Nachweis.

What bleibt unverändert:

- Kein ClickHouse-, VictoriaMetrics- oder Mimir-Pflichtbackend wird
  eingeführt.

## 5. Tranche 3 — Kubernetes, NF-18 und R-9

Ziel: K8s bleibt optional, ist aber als reproduzierbarer Optionspfad
konkret genug, um später nicht erneut grundsätzlich entschieden werden
zu müssen.

DoD:

- [x] `0.13.0`-Entscheidung zu `MVP-42`, `NF-18` und R-9 liegt vor.
- [x] Beispielmanifeste, Seed-Hardening-Notiz oder Defer-Notiz liegen
  mit klarer Production-Ready-Abgrenzung vor.
- [x] Observability-Label-Allowlist ist gegen K8s-Smoke-Anforderungen
  geprüft.
- [x] Mindestens zwei R-9-Gegenmaßnahmen sind dokumentiert und einem
  Owner zugeordnet.
- [x] Smoke-Stage ist entweder optional implementiert oder mit
  messbarem Trigger deferred.

Go/No-Go:

- **Go:** optionale Manifeste, isolierter Smoke, dokumentierte
  Observability-Label-Strategie.
- **No-Go:** verpflichtender Cluster für Standardtests, Production-
  Betriebsversprechen, unkontrollierte Label-Cardinality.

Vorläufige Artefakte:

- `deploy/`- oder `examples/`-Optionpfad, falls freigegeben.
- Risks-Backlog-Update zu R-9.
- README-/User-Doku-Abgrenzung.

### 5.1 Seed-Hardening — 2026-05-12

**Entscheidung:** K8s bleibt optionaler Beispielpfad. `0.14.0`
führt keinen Cluster-Smoke als Pflicht-Gate ein, sondern liefert einen
clusterfreien Manifest-Validator.

Artefakte:

- `scripts/validate-k8s-examples.sh`
- `Makefile` Target `k8s-validate`
- `deploy/k8s/README.md` Abschnitt `Validate`
- `docs/planning/in-progress/risks-backlog.md` R-9-Mitigation

Validierte Grenzen:

- erlaubte Ressourcen bleiben `Namespace`, `PersistentVolumeClaim`,
  `Deployment` und `Service`;
- alle Workloads bleiben Single-Replica;
- Image-Tags müssen zur Root-`package.json`-Version passen;
- die gemeinsamen Labels bleiben auf `app.kubernetes.io/name` und
  `app.kubernetes.io/part-of=m-trace` begrenzt;
- Beispielmanifeste dürfen keine Infrastruktur-Labelkeys `pod`,
  `namespace` oder `container` einführen;
- Production-orientierte Kinds wie `Ingress`, `HorizontalPodAutoscaler`,
  `NetworkPolicy` und `PodDisruptionBudget` bleiben out of scope.

R-9-Gegenmaßnahmen:

| Gegenmaßnahme | Owner | Nachweis |
| --- | --- | --- |
| Separater K8s-Allowlist-Modus statt Erweiterung des Compose-Defaults | Platform/Ops | Validator blockiert Infrastruktur-Labelkeys in den Beispielen. |
| Smoke-Scope-Trennung per explizitem Profil/ENV-Gate | Platform/Ops | K8s-Smoke bleibt deferred; `k8s-validate` ist clusterfrei. |
| Nicht-Production-Ready-Abgrenzung im Deploy-Pfad | Platform/Ops | README-Check im Validator erzwingt die Formulierung. |

What ändert sich:

- `make k8s-validate` ist der reproduzierbare Seed-Hardening-Nachweis
  für `deploy/k8s/`.

What bleibt unverändert:

- K8s ist kein Standard-Gate und ersetzt das Compose-Lab nicht.

## 6. Tranche 4 — Devcontainer und Release-Automation

Ziel: Reproduzierbarkeit und Release-Sicherheit werden dort umgesetzt
oder validiert, wo `0.13.0` mehr als Runbook-only freigibt oder
offene Seed-Hardening-Aufgaben hinterlässt.

DoD:

- [x] `0.13.0`-Entscheidungen zu `MVP-43` und `MVP-44` liegen vor.
- [x] Devcontainer-Pfad ist implementiert, als vorhandener Seed validiert
  oder mit Triggern deferred.
- [x] Devcontainer enthält nur reproduzierbare Entwicklungs-
  Hilfsmittel und ersetzt nicht die dokumentierten Docker-/Make-Pfade.
- [x] Release-Automations-Guard ist als vorhandener Seed, Dry-Run oder
  CI-/Local-Runbook nachweisbar.
- [x] Human-Approval-Gate bleibt verpflichtend und technisch oder
  prozessual verankert.
- [x] Guard-Fehler liefern einen sicheren Abbruch ohne Tag-/Release-
  Seiteneffekte.
- [x] Rollback- und Notfallpfad ist im Release-Runbook beschrieben.

Go/No-Go:

- **Go:** Dry-Run reproduzierbar, Owner/RACI klar, Human Approval
  zwingend.
- **No-Go:** automatisches Taggen/Publishen ohne Review, Devcontainer
  als versteckte Pflichtumgebung.

Vorläufige Artefakte:

- `.devcontainer/` nur bei freigegebenem oder zu validierendem DevEx-
  Scope.
- Release-Runbook-Update.
- Dry-Run- oder Guard-Test.

### 6.1 Seed-Validation — 2026-05-12

**Entscheidung:** Devcontainer und Release-Guard bleiben Zusatzpfade.
`0.14.0` validiert beide Seeds, ohne lokale Standardentwicklung oder
Release-Veröffentlichung zu automatisieren.

Artefakte:

- `scripts/validate-devcontainer.sh`
- `scripts/test-release-guard.sh`
- `Makefile` Targets `devcontainer-validate` und `release-guard-test`
- `docs/user/local-development.md` §1.4
- `docs/user/releasing.md` §3.0

Devcontainer-Validation:

- JSON muss parsebar sein;
- Docker-outside-of-Docker bleibt explizites Feature;
- Go ist auf `1.26.3`, Node auf `22`, pnpm auf `10.18.0` gepinnt;
- `postCreateCommand` darf keine Workspace-Dependencies installieren;
- `remoteUser` bleibt `vscode`.

Release-Guard-Validation:

- temporäre Git-Repositories prüfen Erfolgspfad und Fehlerfälle;
- fehlende Freigabe, `v`-Prefix, Non-`main`, Dirty Worktree,
  lokaler Tag und `package.json`-Versionsdrift werden abgefangen;
- Tests nutzen `MTRACE_RELEASE_ALLOW_OFFLINE=1` nur im Testkontext und
  erzeugen keinen Tag.

What ändert sich:

- Seed-Validierung ist per `make devcontainer-validate` und
  `make release-guard-test` reproduzierbar.

What bleibt unverändert:

- Make/Docker bleiben Standardentwicklung.
- Commit, Tag, Push und GitHub-Release bleiben manuelle Schritte mit
  expliziter Human Approval.

## 7. Tranche 5 — Release-Closeout und Abschluss

Ziel: Alle übernommenen Pfade sind nachweisbar abgeschlossen,
deferred oder blockiert, und der Release kann sauber getaggt werden.

DoD:

- [x] RAK-Verifikationsmatrix vollständig ausgefüllt.
- [x] Jede aktive Tranche enthält einen `What ändert sich` /
  `What bleibt unverändert`-Block mit Dateinachweis.
- [x] `make docs-check` grün.
- [x] Bei codebezogenen Änderungen: `make build` grün.
- [x] Bei codebezogenen Änderungen: `make gates` grün.
- [x] Bei codebezogenen/Release-Änderungen: `make security-gates`
  grün oder CI-Job `Security gates` grün dokumentiert.
- [x] Versions-Bump auf `0.14.0` vollständig durchgeführt.
- [x] `CHANGELOG.md` mit `[0.14.0] - 2026-05-12` aktualisiert.
- [x] Roadmap auf released `0.14.0` und nächste Folgephase umgestellt.
- [x] Plan nach `docs/planning/done/plan-0.14.0.md` verschoben,
  Status auf `✅ released`.
- [x] Annotierter Tag `v0.14.0` erstellt.

Verifikation:

| Gate | Ergebnis |
| --- | --- |
| `make api-race` | ✅ grün; validiert GitHub-CI-Flake-Fix für `adapters/driven/webhooks` |
| `make docs-check` | ✅ grün |
| `make k8s-validate` | ✅ grün |
| `make devcontainer-validate` | ✅ grün |
| `make release-guard-test` | ✅ grün |
| `make build` | ✅ grün |
| `make gates` | ✅ grün |
| `make security-gates` | ✅ grün |
| `MTRACE_RELEASE_APPROVED=1 make release-guard VER=0.14.0` | ✅ grün |

## 8. RAK-Verifikationsmatrix

Die RAK-IDs sind mit Lastenheft-Patch `1.1.19` in §13.18
reserviert.

| RAK | Priorität | Nachweis | Akzeptanz | Status |
| --- | --- | --- | --- | --- |
| RAK-96 | Muss | `0.13.0`-Closeout, Postgres-Entscheidungsnotiz, Migration/POC/Defer-Trigger | Postgres-Folgepfad bleibt als `defer-with-migration-seed` vorbereitet oder wird nur bei Trigger umgesetzt; SQLite bleibt Default | [x] |
| RAK-97 | Muss | Analytics-Defer-Notiz, Query-/Kostenmatrix | Analytics-Pfad hat klare Workloads und Erfolg-/Abbruchkriterien oder messbare Defer-Trigger; kein Pflichtbackend | [x] |
| RAK-98 | Muss | K8s-/NF-18-Notiz, R-9-Risiko-Update, optionale Manifeste/Smoke | K8s bleibt optional; vorhandene Seeds sind validiert oder Observability-Label-Risiken sind kontrolliert oder Smoke ist deferred | [x] |
| RAK-99 | Muss | Devcontainer-Artefakt oder Validation-Notiz | DevEx-Reproduzierbarkeit ist verbessert, ohne den Standardpfad zu ersetzen | [x] |
| RAK-100 | Muss | Release-Runbook, Guard-/Dry-Run-Test, RACI | Release-Automation bleibt freigabepflichtig und erzeugt keine unreviewten Publish-/Tag-Seiteneffekte | [x] |

Sofort nutzbares Verifikationsmapping (bei Aktivierung auszufüllen):

| RAK | Primäre Datei(en) | Datum | Owner | Status |
| --- | --- | --- | --- | --- |
| RAK-96 | `docs/adr/0005-production-ops-backends.md`, `docs/ops/backend-followup.md`, `docs/planning/done/plan-0.14.0.md` | 2026-05-12 | Platform/Storage | ✅ |
| RAK-97 | `docs/adr/0005-production-ops-backends.md`, `docs/ops/backend-followup.md`, `docs/planning/done/plan-0.14.0.md` | 2026-05-12 | Platform/QA | ✅ |
| RAK-98 | `scripts/validate-k8s-examples.sh`, `deploy/k8s/README.md`, `deploy/k8s/*.yaml`, `docs/planning/in-progress/risks-backlog.md` | 2026-05-12 | Platform/Ops | ✅ |
| RAK-99 | `.devcontainer/devcontainer.json`, `scripts/validate-devcontainer.sh`, `docs/user/local-development.md`, `docs/planning/done/plan-0.14.0.md` | 2026-05-12 | Platform/DevEx | ✅ |
| RAK-100 | `scripts/release-guard.sh`, `scripts/test-release-guard.sh`, `docs/user/releasing.md`, `docs/planning/done/plan-0.14.0.md` | 2026-05-12 | Platform/CI | ✅ |

## 8.1 Blocker-Log (Startzustand)

| Blocker | Betroffene Tranche | Erwartete Auflösung |
| --- | --- | --- |
| `0.13.0` noch nicht released | alle | ✅ geschlossen: `v0.13.0` auf `4178f52` |
| RAK-Range noch offen | Tranche 0/5 | ✅ geschlossen: `RAK-96`..`RAK-100` |
| Backend-Entscheidungen noch offen | Tranche 1/2/3/4 | ✅ geschlossen: Entscheidungsimport in §2.1 |

## 9. Folge-Scope nach `0.14.0`

- Später: vollständige Production-Kubernetes- und Observability-
  Standardisierung, falls K8s in `0.14.0` nur optional bleibt.
- Später: SLO-gesteuerte Storage-Strategien und Ops-Runbooks über den
  ersten Postgres-/Analytics-Slice hinaus.
- Später: dediziertes Secret-Management oder Cloud-spezifische
  Betriebsprofile, falls reale Betreiberanforderungen dies auslösen.
