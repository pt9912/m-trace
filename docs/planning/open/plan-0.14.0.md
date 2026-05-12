# Implementation Plan — `0.14.0` (Ops Backend Follow-up)

> **Status**: ⬜ offen — vorbereiteter Folgeplan nach `0.13.0`.
> Aktivierung erst nach Abschluss und Closeout von `0.13.0`.
>
> **Vorgänger**: `0.13.0` (Production / Ops Backends), aktuell in
> Arbeit seit 2026-05-12; Plan in
> `docs/planning/in-progress/plan-0.13.0.md`.
>
> **Release-Typ**: voraussichtlich Minor-Release mit Lastenheft-Patch,
> neuer RAK-Gruppe und Tag `v0.14.0`. RAK-Range noch offen.
>
> **Ziel**: Die in `0.13.0` getroffenen Ops-Backend-Entscheidungen in
> konkrete, lieferbare Umsetzungsslices überführen: Postgres-
> Migrationsslice, Analytics-Backend-Folgepfad und/oder K8s-/DevEx-
> Optionen. Der finale Scope hängt von den `0.13.0`-Entscheidungen ab.
>
> **Bezug**:
> [`../in-progress/roadmap.md`](../in-progress/roadmap.md),
> `docs/planning/in-progress/plan-0.13.0.md`
> §9, [`../../../spec/lastenheft.md`](../../../spec/lastenheft.md)
> §13.17 (`RAK-91`..`RAK-95`).
>
> **Nachfolger**: offen.

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

Vorläufig in Scope:

- Umsetzung eines Postgres-Folgepfads, falls `0.13.0` für Seed oder
  POC entscheidet: Migration, Adapter-Slice, Contract-/Regressionstests
  und Rollback-Nachweis.
- Umsetzung oder POC eines Analytics-Backend-Folgepfads, falls
  `0.13.0` `proceed` oder `POC` entscheidet: begrenztes Datenmodell,
  Ingest-/Exportpfad, Query-Nachweise und Kosten-/Lastgrenzen.
- Konkretisierung des K8s-Optionpfads, falls `0.13.0` ihn freigibt:
  Beispielmanifeste, Smoke-Gate-Entscheidung und Observability-
  Label-Harmonisierung.
- DevEx-Folgepfad aus `MVP-43`, falls `0.13.0` einen Devcontainer-
  Seed oder eine Reaktivierung empfiehlt.
- Release-Automations-Umsetzung aus `MVP-44`, falls `0.13.0` konkrete
  Dry-Run-/Guard-Schritte freigibt.

Vorläufig out of scope:

- Keine verpflichtende Ablösung von SQLite im lokalen Standardbetrieb,
  solange keine `0.13.0`-Entscheidung dies ausdrücklich vorgibt.
- Kein vollständiger Production-Kubernetes-Betrieb.
- Kein Managed-Cloud-Betrieb.
- Kein Multi-Tenant-SaaS-Produkt.
- Keine automatische Veröffentlichung ohne explizite Human Approval.
- Kein Secret-Management-Vollausbau jenseits der bereits gelieferten
  `0.12.x`-Pfade, außer ein `0.13.0`-Closeout zieht ihn ausdrücklich
  als Folge-Scope.

### 0.2 Vorgänger-Gate

Vor Aktivierung von `0.14.0` müssen diese Bedingungen erfüllt sein:

- [ ] `0.13.0` ist released und als
  `docs/planning/done/plan-0.13.0.md` archiviert.
- [ ] Roadmap zeigt `0.14.0` als aktive Folgephase oder begründet
  einen anderen Nachfolger.
- [ ] RAK-91..RAK-95 sind in der `0.13.0`-Verifikationsmatrix
  geschlossen oder mit explizitem Defer-/Blockerstatus versehen.
- [ ] Für jeden übernommenen Pfad existiert eine `0.13.0`-
  Entscheidung mit `Entscheidung`, `Begründung`, `Nicht entschieden`
  und Triggern.
- [ ] Keine neue lokale Pflichtabhängigkeit wird ohne Lastenheft-Patch
  und Migrations-/Rollback-Nachweis eingeführt.

### 0.3 Lastenheft-Patch (TBD)

Die neue RAK-Gruppe wird erst nach dem `0.13.0`-Closeout vergeben.
Vorläufige RAK-Themen:

| Vorläufige Kennung | Thema | Bedingung |
| --- | --- | --- |
| RAK-TBD-1 | Postgres-Folgepfad | Nur bei `0.13.0`-Entscheidung `seed`, `proceed` oder `POC`. |
| RAK-TBD-2 | Analytics-Folgepfad | Nur bei `0.13.0`-Entscheidung `proceed` oder `POC`. |
| RAK-TBD-3 | K8s-/NF-18-Optionpfad | Nur nach R-9-Entscheidung mit Gegenmaßnahmen. |
| RAK-TBD-4 | Devcontainer-/DevEx-Reproduzierbarkeit | Nur falls `MVP-43` nicht vollständig in `0.13.0` geschlossen wird. |
| RAK-TBD-5 | Release-Automations-Guards | Nur falls `0.13.0` Umsetzung statt Runbook-only empfiehlt. |

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Erwartetes Ergebnis | Eingang | Ausgang | Status |
| --- | --- | --- | --- | --- | --- |
| 0 | Aktivierung, RAK-Zuschnitt und Vorgänger-Entscheidungen | Scope aus `0.13.0` verbindlich übernommen | `0.13.0` released | Finaler 0.14-Scope | ⬜ |
| 1 | Postgres-Migrations-/Adapter-Slice | Implementierter oder final deferred Postgres-Pfad | RAK-91-Ergebnis | Migrations-/Rollback-Nachweis | ⬜ |
| 2 | Analytics-Backend-Slice oder POC | Datenmodell-, Query- und Kostenentscheidung umgesetzt | RAK-92-Ergebnis | POC-Report oder Implementierung | ⬜ |
| 3 | K8s-/NF-18-Optionpfad und R-9 | Optionaler K8s-Pfad ohne Production-Ready-Zusage | RAK-93-Ergebnis | Manifest-/Smoke-/Risiko-Nachweis | ⬜ |
| 4 | Devcontainer und Release-Automations-Guards | Reproduzierbare DevEx und sichere Release-Dry-Runs | RAK-94/95-Ergebnis | Runbook-/Guard-Artefakte | ⬜ |
| 5 | Gates, RAK-Matrix, Versions-Bump, Closeout und Tag | Release nachweisbar abgeschlossen | Tranche 4 | Tag `v0.14.0` | ⬜ |

## 2. Tranche 0 — Aktivierung und Scope-Härtung

Ziel: Der offene Plan wird nach `0.13.0` in einen entscheidbaren
Umsetzungsplan überführt, ohne die `0.13.0`-Entscheidungen zu
überstimmen.

DoD:

- [ ] Plan von `docs/planning/open/plan-0.14.0.md` nach
  `docs/planning/in-progress/plan-0.14.0.md` verschoben.
- [ ] Ausgangszustand von `git status --short` dokumentiert.
- [ ] `0.13.0`-Closeout gelesen und alle übernommenen Pfade explizit
  auf `proceed`, `POC`, `defer` oder `blocked` gemappt.
- [ ] Lastenheft-Patch mit finaler RAK-Range ergänzt.
- [ ] Roadmap auf `0.14.0` als aktive Folgephase umgestellt.
- [ ] Risiken-Backlog aktualisiert, insbesondere R-9 und alle durch
  Postgres/Analytics/K8s neu ausgelösten Betriebsrisiken.

## 3. Tranche 1 — Postgres-Folgepfad

Ziel: Der aus `0.13.0` übernommene Postgres-Pfad wird entweder
umgesetzt, als zeitbegrenzter POC gefahren oder final deferred.

DoD:

- [!] `0.13.0`-Entscheidung zu `MVP-40` liegt vor.
- [ ] Migrationsmodell definiert: `migrate up`, `rollback`, `replay`
  und Kompatibilitätsgrenze zu SQLite.
- [ ] Adapter-Scope auf minimale Ports und Queries begrenzt.
- [ ] Contract- und Regressionstests belegen, dass SQLite der lokale
  Default bleibt.
- [ ] Backup-/Restore- und Ausfallverhalten dokumentiert.
- [ ] Reaktivierungs- oder Defer-Trigger mit Owner und Messwerten
  aktualisiert.

## 4. Tranche 2 — Analytics-Backend-Folgepfad

Ziel: Der in `0.13.0` gewählte Analytics-Pfad wird mit begrenztem
Datenmodell, klaren Abbruchkriterien und Query-Nachweisen konkret.

DoD:

- [!] `0.13.0`-Entscheidung zu `MVP-41` liegt vor.
- [ ] Zielbackend oder POC-Variante final bestätigt.
- [ ] Datenmodell und Retention-Grenzen definiert.
- [ ] Query-Workloads mit erwarteter Last und Kostenannahmen
  dokumentiert.
- [ ] Ingest-/Exportpfad bleibt optional und führt keine lokale
  Pflichtabhängigkeit ein.
- [ ] POC-Report oder Implementierungsnachweis enthält
  Erfolgskriterien, Abbruchkriterien und Zeitgrenze.

## 5. Tranche 3 — Kubernetes, NF-18 und R-9

Ziel: K8s bleibt optional, ist aber als reproduzierbarer Optionspfad
konkret genug, um später nicht erneut grundsätzlich entschieden werden
zu müssen.

DoD:

- [!] `0.13.0`-Entscheidung zu `MVP-42`, `NF-18` und R-9 liegt vor.
- [ ] Beispielmanifeste oder Defer-Notiz liegen mit klarer
  Production-Ready-Abgrenzung vor.
- [ ] Observability-Label-Allowlist ist gegen K8s-Smoke-Anforderungen
  geprüft.
- [ ] Mindestens zwei R-9-Gegenmaßnahmen sind dokumentiert und einem
  Owner zugeordnet.
- [ ] Smoke-Stage ist entweder optional implementiert oder mit
  messbarem Trigger deferred.

## 6. Tranche 4 — Devcontainer und Release-Automation

Ziel: Reproduzierbarkeit und Release-Sicherheit werden dort umgesetzt,
wo `0.13.0` mehr als Runbook-only freigibt.

DoD:

- [!] `0.13.0`-Entscheidungen zu `MVP-43` und `MVP-44` liegen vor.
- [ ] Devcontainer-Pfad ist implementiert oder mit Triggern deferred.
- [ ] Release-Automations-Guard ist als Dry-Run oder CI-/Local-Runbook
  nachweisbar.
- [ ] Human-Approval-Gate bleibt verpflichtend und technisch oder
  prozessual verankert.
- [ ] Rollback- und Notfallpfad ist im Release-Runbook beschrieben.

## 7. Tranche 5 — Release-Closeout und Abschluss

Ziel: Alle übernommenen Pfade sind nachweisbar abgeschlossen,
deferred oder blockiert, und der Release kann sauber getaggt werden.

DoD:

- [ ] RAK-Verifikationsmatrix vollständig ausgefüllt.
- [ ] `make docs-check` grün.
- [ ] Bei codebezogenen Änderungen: `make build` grün.
- [ ] Bei codebezogenen Änderungen: `make gates` grün.
- [ ] Bei codebezogenen/Release-Änderungen: `make security-gates`
  grün oder CI-Job `Security gates` grün dokumentiert.
- [ ] Versions-Bump auf `0.14.0` vollständig durchgeführt.
- [ ] `CHANGELOG.md` mit `[0.14.0] - YYYY-MM-DD` aktualisiert.
- [ ] Roadmap auf released `0.14.0` und nächste Folgephase umgestellt.
- [ ] Plan nach `docs/planning/done/plan-0.14.0.md` verschoben,
  Status auf `✅ released`.
- [ ] Annotierter Tag `v0.14.0` erstellt.

## 8. RAK-Verifikationsmatrix (Platzhalter)

Wird bei Aktivierung nach dem `0.13.0`-Closeout mit finalen RAK-IDs
gefüllt.

| RAK | Priorität | Nachweis | Akzeptanz | Status |
| --- | --- | --- | --- | --- |
| TBD | Muss | `0.13.0`-Closeout, Lastenheft-Patch, dieser Plan | Finaler Scope ist aus `0.13.0` übernommen und nicht vorweggenommen | [ ] |

## 9. Folge-Scope nach `0.14.0`

- Später: vollständige Production-Kubernetes- und Observability-
  Standardisierung, falls K8s in `0.14.0` nur optional bleibt.
- Später: SLO-gesteuerte Storage-Strategien und Ops-Runbooks über den
  ersten Postgres-/Analytics-Slice hinaus.
- Später: dediziertes Secret-Management oder Cloud-spezifische
  Betriebsprofile, falls reale Betreiberanforderungen dies auslösen.
