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
> **Ziel**: Die in `0.13.0` getroffenen Ops-Backend-Entscheidungen
> in konkrete, lieferbare Umsetzungsslices überführen. `0.14.0`
> ist kein erneuter Bewertungsplan, sondern der erste Umsetzungsplan
> für die in `0.13.0` freigegebenen Pfade: Postgres-Migrationsslice,
> Analytics-Backend-POC/-Slice, K8s-/NF-18-Option, Devcontainer-
> Reproduzierbarkeit und Release-Automations-Guards.
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

Scope-Regel:

- `0.14.0` übernimmt nur Pfade, die im `0.13.0`-Closeout als
  `proceed`, `seed` oder `POC` markiert sind.
- Pfade mit `defer` bleiben dokumentiert, erhalten aber nur Trigger-
  Pflege und keine Implementierungstranche.
- Pfade mit `blocked` bleiben in der Tranche-Tabelle sichtbar und
  müssen vor Aktivierung aufgelöst oder ausdrücklich aus dem Release
  gestrichen werden.

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
| Postgres `seed`/`POC` | Tranche 1 aktiv | `[!]` blockiert | Migrations-/Rollback-Entscheidung, SQLite-Kompatibilitätsgrenze |
| Postgres `defer` | Tranche 1 nur Trigger-Pflege | nicht implementieren | Defer-Trigger mit Schwellwert und Owner |
| Analytics `proceed`/`POC` | Tranche 2 aktiv | `[!]` blockiert | Backend-Wahl, Datenmodell, Erfolg-/Abbruchkriterien |
| Analytics `defer` | Tranche 2 nur Trigger-Pflege | nicht implementieren | Vergleichsmatrix + Reaktivierungsbedingung |
| K8s `option` | Tranche 3 aktiv | `[!]` blockiert | R-9-Entscheidung und mindestens zwei Gegenmaßnahmen |
| K8s `defer` | Tranche 3 nur Dokumentationspflege | nicht implementieren | klare Nicht-Production-Ready-Abgrenzung |
| Devcontainer `seed` | Tranche 4 DevEx aktiv | `[!]` blockiert | lokale Standardentwicklungs-Abgrenzung |
| Release-Automation `guard` | Tranche 4 Release aktiv | `[!]` blockiert | Human-Approval-Gate und Dry-Run-Nachweis |

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

Jede aktive Umsetzungstranche liefert mindestens:

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
| A | Postgres-Slice + Release-Guards | fokussierter Storage-/CI-Release | RAK-91 und RAK-95 geben Umsetzung frei |
| B | Analytics-POC + optionale K8s-Doku | POC-/Decision-Release | RAK-92 gibt POC frei, RAK-93 ist nicht blockiert |
| C | K8s/DevEx/Release-Guard-Slice | Ops-Enablement-Release | RAK-93..RAK-95 geben Optionpfade frei |
| D | Trigger-/Defer-Release | rein dokumentarisch | 0.13 deferred alle Implementierungspfade |

Mehr als zwei große Implementierungspfade in einem `0.14.0`-Release
gelten als No-Go, außer Tranche 0 dokumentiert explizit zusätzliche
Kapazität und getrennte Gate-Nachweise.

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Erwartetes Ergebnis | Eingang | Ausgang | Status |
| --- | --- | --- | --- | --- | --- |
| 0 | Aktivierung, RAK-Zuschnitt und Vorgänger-Entscheidungen | Scope aus `0.13.0` verbindlich übernommen | `0.13.0` released | Finaler 0.14-Scope | ⬜ |
| 1 | Postgres-Migrations-/Adapter-Slice | Implementierter, POC-fähiger oder final deferred Postgres-Pfad | RAK-91-Ergebnis | Migrations-/Rollback-/Trigger-Nachweis | ⬜ |
| 2 | Analytics-Backend-Slice oder POC | Datenmodell-, Query- und Kostenentscheidung umgesetzt | RAK-92-Ergebnis | POC-Report, Adapter-Slice oder Defer-Notiz | ⬜ |
| 3 | K8s-/NF-18-Optionpfad und R-9 | Optionaler K8s-Pfad ohne Production-Ready-Zusage | RAK-93-Ergebnis | Manifest-/Smoke-/Risiko-Nachweis | ⬜ |
| 4 | Devcontainer und Release-Automations-Guards | Reproduzierbare DevEx und sichere Release-Dry-Runs | RAK-94/95-Ergebnis | Runbook-/Guard-Artefakte | ⬜ |
| 5 | Gates, RAK-Matrix, Versions-Bump, Closeout und Tag | Release nachweisbar abgeschlossen | letzte aktive Tranche | Tag `v0.14.0` | ⬜ |

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
- [ ] Aktivierungsszenario A/B/C/D ausgewählt und begründet.
- [ ] Aktive und deferred Tranchen für das gewählte Szenario in einer
  Tranchenmatrix festgelegt.
- [ ] Nicht übernommene `0.13.0`-Pfade bleiben als Defer-Trigger
  sichtbar, werden aber nicht stillschweigend implementiert.
- [ ] Lastenheft-Patch mit finaler RAK-Range ergänzt.
- [ ] Roadmap auf `0.14.0` als aktive Folgephase umgestellt.
- [ ] Risiken-Backlog aktualisiert, insbesondere R-9 und alle durch
  Postgres/Analytics/K8s neu ausgelösten Betriebsrisiken.
- [ ] No-Go-Liste geprüft:
  - unklare Backend-Pflichtabhängigkeit,
  - fehlender Rollbackpfad,
  - fehlende Human Approval im Release-Pfad,
  - K8s-Production-Ready-Sprache ohne Betriebsplan.

### 2.1 Aktivierungsnotiz (Template)

Bei Aktivierung ausfüllen:

| Feld | Wert |
| --- | --- |
| Aktivierungsdatum | TBD |
| Ausgangs-Commit | TBD |
| Gewähltes Szenario | TBD |
| Übernommene 0.13-Pfade | TBD |
| Explizit deferred | TBD |
| Blocker | TBD |
| Required Gates | TBD |

## 3. Tranche 1 — Postgres-Folgepfad

Ziel: Der aus `0.13.0` übernommene Postgres-Pfad wird entweder
umgesetzt, als zeitbegrenzter POC gefahren oder final deferred.

DoD:

- [!] `0.13.0`-Entscheidung zu `MVP-40` liegt vor.
- [ ] Entscheiden, ob `0.14.0` einen POC, einen schmalen
  produktionsnahen Adapter-Slice oder nur Trigger-Pflege liefert.
- [ ] Migrationsmodell definiert: `migrate up`, `rollback`, `replay`
  und Kompatibilitätsgrenze zu SQLite.
- [ ] Schema-Differenzen zwischen SQLite und Postgres dokumentiert
  (Zeittypen, IDs, Constraints, Transaktionen, Pagination-Sortierung).
- [ ] Adapter-Scope auf minimale Ports und Queries begrenzt.
- [ ] Contract- und Regressionstests belegen, dass SQLite der lokale
  Default bleibt.
- [ ] Backup-/Restore- und Ausfallverhalten dokumentiert.
- [ ] Reaktivierungs- oder Defer-Trigger mit Owner und Messwerten
  aktualisiert.

Go/No-Go:

- **Go:** genau definierter Datenbereich, reproduzierbare Migration,
  keine Änderung am Default-Store.
- **No-Go:** vollständige Store-Ablösung, implizite Postgres-Pflicht
  für lokale Tests, ungetestete Rollbackannahmen.

Vorläufige Artefakte:

- `docs/adr/` oder Plan-Entscheidungsnotiz für den Postgres-Slice.
- Migrations-/Rollback-Dokumentation.
- Adapter-/Repository-Tests oder POC-Report.

## 4. Tranche 2 — Analytics-Backend-Folgepfad

Ziel: Der in `0.13.0` gewählte Analytics-Pfad wird mit begrenztem
Datenmodell, klaren Abbruchkriterien und Query-Nachweisen konkret.

DoD:

- [!] `0.13.0`-Entscheidung zu `MVP-41` liegt vor.
- [ ] Zielbackend oder POC-Variante final bestätigt.
- [ ] Datenmodell und Retention-Grenzen definiert.
- [ ] Query-Workloads mit erwarteter Last und Kostenannahmen
  dokumentiert.
- [ ] Datenfluss klar geschnitten: Realtime-Ingest, Batch-Export,
  Replikation oder synthetischer POC-Load.
- [ ] Ingest-/Exportpfad bleibt optional und führt keine lokale
  Pflichtabhängigkeit ein.
- [ ] POC-Report oder Implementierungsnachweis enthält
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

Go/No-Go:

- **Go:** optionale Manifeste, isolierter Smoke, dokumentierte
  Observability-Label-Strategie.
- **No-Go:** verpflichtender Cluster für Standardtests, Production-
  Betriebsversprechen, unkontrollierte Label-Cardinality.

Vorläufige Artefakte:

- `deploy/`- oder `examples/`-Optionpfad, falls freigegeben.
- Risks-Backlog-Update zu R-9.
- README-/User-Doku-Abgrenzung.

## 6. Tranche 4 — Devcontainer und Release-Automation

Ziel: Reproduzierbarkeit und Release-Sicherheit werden dort umgesetzt,
wo `0.13.0` mehr als Runbook-only freigibt.

DoD:

- [!] `0.13.0`-Entscheidungen zu `MVP-43` und `MVP-44` liegen vor.
- [ ] Devcontainer-Pfad ist implementiert oder mit Triggern deferred.
- [ ] Devcontainer enthält nur reproduzierbare Entwicklungs-
  Hilfsmittel und ersetzt nicht die dokumentierten Docker-/Make-Pfade.
- [ ] Release-Automations-Guard ist als Dry-Run oder CI-/Local-Runbook
  nachweisbar.
- [ ] Human-Approval-Gate bleibt verpflichtend und technisch oder
  prozessual verankert.
- [ ] Guard-Fehler liefern einen sicheren Abbruch ohne Tag-/Release-
  Seiteneffekte.
- [ ] Rollback- und Notfallpfad ist im Release-Runbook beschrieben.

Go/No-Go:

- **Go:** Dry-Run reproduzierbar, Owner/RACI klar, Human Approval
  zwingend.
- **No-Go:** automatisches Taggen/Publishen ohne Review, Devcontainer
  als versteckte Pflichtumgebung.

Vorläufige Artefakte:

- `.devcontainer/` nur bei freigegebenem DevEx-Slice.
- Release-Runbook-Update.
- Dry-Run- oder Guard-Test.

## 7. Tranche 5 — Release-Closeout und Abschluss

Ziel: Alle übernommenen Pfade sind nachweisbar abgeschlossen,
deferred oder blockiert, und der Release kann sauber getaggt werden.

DoD:

- [ ] RAK-Verifikationsmatrix vollständig ausgefüllt.
- [ ] Jede aktive Tranche enthält einen `What ändert sich` /
  `What bleibt unverändert`-Block mit Dateinachweis.
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
gefüllt. Bis dahin dienen die folgenden Zeilen als Zuschnittsvorschlag.

| RAK | Priorität | Nachweis | Akzeptanz | Status |
| --- | --- | --- | --- | --- |
| RAK-TBD-1 | Muss | `0.13.0`-Closeout, Postgres-Entscheidungsnotiz, Migration/POC/Defer-Trigger | Postgres-Folgepfad ist umgesetzt, POC-fähig abgegrenzt oder final deferred; SQLite bleibt Default | [ ] |
| RAK-TBD-2 | Muss | Analytics-POC-Report oder Defer-Notiz, Query-/Kostenmatrix | Analytics-Pfad hat ein Zielbackend, klare Workloads und Erfolg-/Abbruchkriterien oder messbare Defer-Trigger | [ ] |
| RAK-TBD-3 | Muss | K8s-/NF-18-Notiz, R-9-Risiko-Update, optionale Manifeste/Smoke | K8s bleibt optional; Observability-Label-Risiken sind kontrolliert oder Smoke ist deferred | [ ] |
| RAK-TBD-4 | Konditional Muss | Devcontainer-Artefakt oder Defer-Notiz | DevEx-Reproduzierbarkeit ist verbessert, ohne den Standardpfad zu ersetzen | [ ] |
| RAK-TBD-5 | Muss | Release-Runbook, Guard-/Dry-Run-Test, RACI | Release-Automation bleibt freigabepflichtig und erzeugt keine unreviewten Publish-/Tag-Seiteneffekte | [ ] |

Sofort nutzbares Verifikationsmapping (bei Aktivierung auszufüllen):

| RAK | Primäre Datei(en) | Datum | Owner | Status |
| --- | --- | --- | --- | --- |
| RAK-TBD-1 | TBD | TBD | Platform/Storage | ⬜ |
| RAK-TBD-2 | TBD | TBD | Platform/QA | ⬜ |
| RAK-TBD-3 | TBD | TBD | Platform/Ops | ⬜ |
| RAK-TBD-4 | TBD | TBD | Platform/DevEx | ⬜ |
| RAK-TBD-5 | TBD | TBD | Platform/CI | ⬜ |

## 8.1 Blocker-Log (Startzustand)

| Blocker | Betroffene Tranche | Erwartete Auflösung |
| --- | --- | --- |
| `0.13.0` noch nicht released | alle | Vorgänger-Gate in §0.2 schließen |
| RAK-Range noch offen | Tranche 0/5 | Lastenheft-Patch bei Aktivierung vergeben |
| Backend-Entscheidungen noch offen | Tranche 1/2/3/4 | Entscheidungsimport aus §0.1a befüllen |

## 9. Folge-Scope nach `0.14.0`

- Später: vollständige Production-Kubernetes- und Observability-
  Standardisierung, falls K8s in `0.14.0` nur optional bleibt.
- Später: SLO-gesteuerte Storage-Strategien und Ops-Runbooks über den
  ersten Postgres-/Analytics-Slice hinaus.
- Später: dediziertes Secret-Management oder Cloud-spezifische
  Betriebsprofile, falls reale Betreiberanforderungen dies auslösen.
