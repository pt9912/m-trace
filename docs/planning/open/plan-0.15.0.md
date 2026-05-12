# Implementation Plan — `0.15.0` (Product Scope / Analyzer Boundary)

> **Status**: ⬜ offen — vorbereiteter Folgeplan nach `0.14.0`.
> Aktivierung erst nach Abschluss und Closeout von `0.14.0`.
>
> **Vorgänger**: `0.14.0` (Ops Backend Follow-up), released
> 2026-05-12; Plan in
> [`../done/plan-0.14.0.md`](../done/plan-0.14.0.md).
>
> **Release-Typ**: voraussichtlich Minor-Release mit Lastenheft-Patch,
> neuer RAK-Gruppe und Tag `v0.15.0`. Erwartete RAK-Range:
> `RAK-101`..`RAK-105`, sofern kein Zwischenrelease RAKs belegt.
>
> **Ziel**: `0.15.0` trifft die naechsten Produkt- und
> Architekturgrenzen, bevor neue grosse Backend- oder Plattformpfade
> implementiert werden. Der Release soll Zielgruppe, externe
> Analyzer-Grenze, Control-Plane-Scope, naechsten Analyzer-Folge-Slice
> und Ops-Backend-Trigger zusammenfuehren.
>
> **Bezug**:
> [`../in-progress/roadmap.md`](../in-progress/roadmap.md),
> [`../in-progress/risks-backlog.md`](../in-progress/risks-backlog.md),
> [`../../../spec/lastenheft.md`](../../../spec/lastenheft.md)
> §16.1, `MVP-20`, `F-132`, `NF-13`, `MVP-40`, `MVP-41`.
>
> **Nachfolger**: offen.

## 0. Konvention

DoD-Checkboxen tracken den Lieferstand:

- `[x]` ausgeliefert mit Commit-Hash.
- `[ ]` offen.
- `[!]` blockiert durch ADR- oder Scope-Entscheidung.
- 🟡 in Arbeit.

### 0.1 Scope-Definition

`0.15.0` ist ein Product-Scope- und Boundary-Decision-Release. Er
liefert bewusst keine neue Pflichtplattform, keinen Postgres-Default
und kein Hochvolumen-Analytics-Backend. Der Plan bereitet nur die
naechsten sinnvollen Implementierungsslices vor.

Vorlaeufig in Scope:

- Zielgruppenentscheidung aus Lastenheft §16.1 formalisieren:
  Selbsthoster/kleine Teams/Broadcaster-Labs bleiben Primaerziel oder
  ein Plattform-Betreiber-Pfad wird explizit als spaeterer Scope
  vorbereitet.
- `MVP-20` bewerten: externe `apps/analyzer-api` gegen bestehenden
  internen `apps/analyzer-service` abgrenzen, inklusive Job-Modell,
  Auth, Rate Limits und Ergebnisabruf als Entscheidungsoptionen.
- `F-132` (`apps/control-plane`) nur als Decision-Record zuschneiden:
  Aufgaben, Nicht-Ziele, Trigger, Abhaengigkeiten und minimale
  RACI/Owner klaeren, aber keine Control-Plane bauen.
- `NF-13` Folge-Scope zuschneiden: Low-Latency-CMAF, HTTP-Range/
  Byte-Range, vollstaendigere Segmentabdeckung oder Codec-Decoding
  gegeneinander bewerten und hoechstens einen kleinen Folge-Slice
  fuer einen spaeteren Plan empfehlen.
- `MVP-40`/`MVP-41` Trigger-Re-Eval nach `0.14.0`: pruefen, ob
  Postgres- oder Analytics-Trigger tatsaechlich erreicht wurden.

Vorlaeufig out of scope:

- Keine Implementierung von `apps/control-plane`.
- Keine neue externe `apps/analyzer-api`, solange Tranche 2 keinen
  klaren Go-Scope liefert.
- Kein Postgres-Runtime-Adapter, kein Postgres-Default und keine
  automatische SQLite-zu-Postgres-Migration ohne ausgelösten Trigger.
- Kein ClickHouse-/VictoriaMetrics-/Mimir-Pflichtbackend und kein POC
  ohne konkreten Workload-Trigger.
- Kein vollstaendiger Production-Kubernetes-Betrieb.
- Kein Multi-Tenant-SaaS-Produkt.
- Kein OAuth/OIDC/SSO- oder User-/Org-Verwaltungs-Ausbau ohne
  ausdrueckliche Zielgruppenentscheidung.

### 0.2 Vorgänger-Gate

Vor Aktivierung von `0.15.0` muessen diese Bedingungen erfuellt sein:

- [ ] `0.14.0` ist released und als
  `docs/planning/done/plan-0.14.0.md` archiviert.
- [ ] Roadmap zeigt `0.15.0` als aktive Folgephase oder begruendet
  einen anderen Nachfolger.
- [ ] RAK-96..RAK-100 sind in der `0.14.0`-Verifikationsmatrix
  geschlossen oder mit explizitem Defer-/Blockerstatus versehen.
- [ ] `0.14.0` hat bestaetigt, dass K8s-/Devcontainer-/Release-Guard-
  Hardening abgeschlossen oder bewusst deferred ist.
- [ ] Keine Postgres-/Analytics-/K8s-Trigger wurden stillschweigend
  als Implementierungsfreigabe interpretiert.

Uebergangsausnahme bei Release-Freeze oder blockiertem Vorgaenger:
Tranche 0 darf `0.15.0` nur als **draft/in-progress Planning** starten,
wenn Roadmap und Blocker-Log den fehlenden `0.14.0`-Closeout
begründen, alle aus `0.14.0` importierten RAKs als `[!]` blockiert
markiert sind und kein Tag/Release fuer `0.15.0` erstellt wird, bevor
`0.14.0` archiviert oder ausdruecklich durch einen neuen Plan ersetzt
ist.

### 0.3 Lastenheft-Patch (TBD)

Die finale RAK-Gruppe wird erst bei Aktivierung vergeben. Erwartete
RAK-Themen, falls keine Zwischen-Ranges belegt werden:

| Vorlaeufige Kennung | Thema | Bedingung |
| --- | --- | --- |
| RAK-TBD-1 / RAK-101 | Zielgruppenentscheidung | Immer aktiv; Lastenheft §16.1 muss geschlossen oder bewusst als Produkt-ADR deferred werden. |
| RAK-TBD-2 / RAK-102 | Analyzer-API-Boundary (`MVP-20`) | Bewertung externe API vs. interner Service, mit Go/No-Go und Triggern. |
| RAK-TBD-3 / RAK-103 | Control-Plane-Scope (`F-132`) | Nur Decision-Record; keine Implementierung ohne spaeteren Plan. |
| RAK-TBD-4 / RAK-104 | Analyzer-Folge-Slice (`NF-13`) | Auswahl oder Defer eines kleinen naechsten Analyzer-Slices. |
| RAK-TBD-5 / RAK-105 | Ops-Trigger-Re-Eval (`MVP-40`/`MVP-41`) | Postgres-/Analytics-Trigger pruefen und sichtbar deferred oder aktiviert halten. |

### 0.4 Qualitätsregeln für `0.15.0`

- Jede Entscheidung enthaelt mindestens `Entscheidung`, `Begruendung`,
  `Nicht entschieden`, `Trigger` und `Naechster Planpfad`.
- Keine neue Plattform-App wird gebaut, bevor Zielgruppe und
  Betriebsmodell entschieden sind.
- Analyzer-Scope muss zwischen internem Service, externer API und
  Library/CLI sauber trennen.
- Neue Backends bleiben opt-in und duerfen lokale Entwicklung nicht
  belasten.
- Decision-only-Tranchen liefern trotzdem pruefbare Artefakte:
  Plan-Notiz, ADR, Lastenheft-Patch, Roadmap-Update oder
  Risks-Backlog-Update.
- Jede aktive Tranche endet mit einem `What aendert sich` /
  `What bleibt unveraendert`-Block.

### 0.5 Aktivierungsszenarien

Tranche 0 waehlt genau eines dieser Aktivierungsszenarien:

| Szenario | Inhalt | Release-Charakter | Go-Bedingung |
| --- | --- | --- | --- |
| A | Zielgruppe + Analyzer-Boundary | fokussierter Product-/Analyzer-Decision-Release | `0.14.0` schliesst Ops-Hardening ohne neue Backend-Trigger |
| B | Zielgruppe + Control-Plane-Decision | Product-/Platform-Decision-Release | Stakeholder- oder Operator-Bedarf fuer Admin-/Multi-Project-Scope liegt vor |
| C | Analyzer-Folge-Slice + Zielgruppe | kleiner Analyzer-Feature-Prep-Release | `NF-13`-Folgeslice ist klar begrenzt und ohne neue Plattformabhaengigkeit |
| D | Trigger-/Defer-Release | rein dokumentarisch | Alle Implementierungsoptionen bleiben ohne Trigger |

Mehr als ein grosser Implementierungspfad in `0.15.0` ist No-Go,
ausser Tranche 0 dokumentiert explizit Ressourcen, Risikoentkopplung
und getrennte Gates.

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Erwartetes Ergebnis | Eingang | Ausgang | Status |
| --- | --- | --- | --- | --- | --- |
| 0 | Aktivierung, RAK-Zuschnitt und `0.14.0`-Closeout-Import | Finaler `0.15.0`-Scope | `0.14.0` released | Szenario A/B/C/D | ⬜ |
| 1 | Zielgruppenentscheidung | Primaerziel und spaetere Plattformgrenze entschieden | Lastenheft §16.1 | Product-Decision-Notiz oder ADR | ⬜ |
| 2 | Analyzer-API-Boundary (`MVP-20`) | Externe Analyzer-API: Go, Defer oder Triggerpflege | `apps/analyzer-service` Bestand | Boundary-Decision + Trigger | ⬜ |
| 3 | Control-Plane-Scope (`F-132`) | Control-Plane bleibt deferred oder erhaelt klaren spaeteren Planpfad | Zielgruppenentscheidung | Decision-Record, Nicht-Ziele, Trigger | ⬜ |
| 4 | Analyzer-Folge-Slice (`NF-13`) | Naechster Analyzer-Slice gewaehlt oder bewusst deferred | CMAF-Folge-Scope | Slice-Zuschnitt oder Defer-Notiz | ⬜ |
| 5 | Ops-Trigger-Re-Eval und Closeout | Postgres/Analytics-Triggerstatus, RAK-Matrix, Release | `MVP-40`/`MVP-41` Trigger | Tag `v0.15.0` | ⬜ |

## 2. Tranche 0 — Aktivierung und Scope-Härtung

Ziel: `0.15.0` wird nach dem `0.14.0`-Closeout aktiviert, ohne die
gerade gehaerteten Ops-Pfade erneut als Implementierungsvorwand zu
nutzen.

DoD:

- [ ] Plan von `docs/planning/open/plan-0.15.0.md` nach
  `docs/planning/in-progress/plan-0.15.0.md` verschoben.
- [ ] Ausgangszustand von `git status --short` dokumentiert.
- [ ] `0.14.0`-Closeout gelesen und Folgepunkte explizit importiert.
- [ ] Aktivierungsszenario A/B/C/D ausgewaehlt und begruendet.
- [ ] Lastenheft-Patch mit finaler RAK-Range ergaenzt.
- [ ] Roadmap auf `0.15.0` als aktive Folgephase umgestellt.
- [ ] Risiken-Backlog aktualisiert, falls neue Trigger oder Risiken
  entstehen.
- [ ] Aktivierungsnotiz enthaelt `What aendert sich` /
  `What bleibt unveraendert` und benennt nicht aktivierte Pfade.
- [ ] No-Go-Liste geprueft:
  - Control-Plane-Implementierung ohne Zielgruppenentscheidung,
  - externe Analyzer-API ohne klaren Konsumenten,
  - Postgres/Analytics-Implementation ohne Trigger,
  - Multi-Tenant-/OAuth-Scope ohne Produktentscheidung.

### 2.1 Aktivierungsnotiz (Template)

Bei Aktivierung ausfuellen:

| Feld | Wert |
| --- | --- |
| Aktivierungsdatum | TBD |
| Ausgangs-Commit | TBD |
| Gewaehltes Szenario | TBD |
| Uebernommene 0.14-Folgepunkte | TBD |
| Explizit deferred | TBD |
| Blocker | TBD |
| Required Gates | TBD |

## 3. Tranche 1 — Zielgruppenentscheidung

Ziel: Die offene Produktfrage aus Lastenheft §16.1 wird so weit
entschieden, dass Folgeplaene nicht gleichzeitig Selbsthoster-Lab,
kleine Teams und grosse Plattformbetreiber bedienen muessen.

DoD:

- [ ] Primaerziel fuer die naechsten Minor-Releases entschieden.
- [ ] Plattform-Betreiber-Scope entweder deferred oder als eigener
  spaeterer Planpfad mit Triggern beschrieben.
- [ ] Auswirkungen auf Storage, Sampling, Cardinality, Multi-Tenant,
  Betriebsmodell, Dashboard-Komplexitaet und Alerting dokumentiert.
- [ ] Lastenheft §16.1 aktualisiert oder durch ADR/Plan-Decision
  referenziert.
- [ ] Out-of-Scope-Liste fuer `0.15.0` nachgezogen.
- [ ] Tranche enthaelt `What aendert sich` /
  `What bleibt unveraendert` mit Dateinachweis.

Go/No-Go:

- **Go:** klare Primaerzielgruppe, explizite Nicht-Ziele, kein
  impliziter Plattform-Scope.
- **No-Go:** gleichzeitige Optimierung fuer Solo-Selbsthoster und
  hunderte parallele Plattformstreams ohne Ressourcenschnitt.

Vorlaeufige Artefakte:

- Product-Scope-Decision in diesem Plan oder `docs/adr/`.
- Roadmap-Update mit Zielgruppenstand.
- Optionaler Lastenheft-Patch fuer §16.1.

## 4. Tranche 2 — Analyzer-API-Boundary (`MVP-20`)

Ziel: Die historische externe `apps/analyzer-api`-Anforderung wird
gegen den realen internen `apps/analyzer-service` bewertet.

DoD:

- [ ] Bestehender interner Analyzer-Service-Pfad beschrieben:
  Verantwortlichkeiten, Grenzen, Verbraucher, Failure-Modes.
- [ ] Externe Analyzer-API-Optionen bewertet:
  synchroner HTTP-Pfad, async Job API, nur Library/CLI, oder Defer.
- [ ] Auth, Rate Limits, SSRF-Schutz, Ergebnisabruf, Retention und
  Contract-Fixtures als Pflichtfragen fuer eine externe API
  dokumentiert.
- [ ] Entscheidung `proceed`, `POC`, `defer` oder `drop/anders
  erfuellt` getroffen.
- [ ] Wenn `proceed` oder `POC`: minimaler Folgeslice ohne Control-
  Plane-Abhaengigkeit definiert.
- [ ] Tranche enthaelt `What aendert sich` /
  `What bleibt unveraendert` mit Dateinachweis.

Go/No-Go:

- **Go:** konkreter externer Konsument, klare API-Grenze, keine
  Dopplung zum internen Service.
- **No-Go:** API nur aus historischer `MVP-20`-Formulierung bauen,
  ohne Nutzer oder Betriebsmodell.

Vorlaeufige Artefakte:

- Boundary-Decision-Notiz.
- Optionaler ADR fuer externe Analyzer-API.
- Triggerliste fuer spaeteren Implementierungsplan.

## 5. Tranche 3 — Control-Plane-Scope (`F-132`)

Ziel: `apps/control-plane` bekommt eine belastbare Kennung und
Entscheidungsgrenze, bleibt aber ohne gesonderten Folgeplan
unimplementiert.

DoD:

- [ ] `F-132`-Scope aus Lastenheft §7.5.6 importiert.
- [ ] Aufgaben, Nicht-Ziele und Abhaengigkeiten dokumentiert:
  Multi-Project, Teams, API-Keys, Audit, Admin-UI, User-/Org-Modell,
  OAuth/OIDC, Betriebsprofile.
- [ ] Trigger definiert, wann eine Control-Plane geplant werden darf.
- [ ] Abhaengigkeit zur Zielgruppenentscheidung aus Tranche 1
  explizit beschrieben.
- [ ] Entscheidung `defer`, `POC` oder `proceed in later plan`
  getroffen.
- [ ] Tranche enthaelt `What aendert sich` /
  `What bleibt unveraendert` mit Dateinachweis.

Go/No-Go:

- **Go:** Control-Plane bleibt klarer spaeterer Planpfad mit
  Triggern und Nicht-Zielen.
- **No-Go:** UI/API-Geruest bauen, bevor Zielgruppe, Auth-Modell und
  Betriebsmodell entschieden sind.

Vorlaeufige Artefakte:

- `F-132` Decision-Record.
- Roadmap- oder Risks-Backlog-Update.
- Optionaler ADR-Draft fuer Plattform-/Admin-Scope.

## 6. Tranche 4 — Analyzer-Folge-Slice (`NF-13`)

Ziel: Der bewusst ausgegrenzte CMAF-/Analyzer-Folge-Scope wird in
eine priorisierte Reihenfolge gebracht, ohne den Analyzer unbounded
zu erweitern.

DoD:

- [ ] Optionen verglichen:
  - Low-Latency-CMAF (`#EXT-X-PART`, chunked CMAF),
  - HTTP-Range-/Byte-Range-Loader,
  - vollstaendigere Segmentset-Abdeckung,
  - Codec-Decoding,
  - Player-SDK-CMAF-Laufzeitpfad.
- [ ] Nutzen, Risiko, Testbarkeit, Datenvolumen und Security-Grenzen
  pro Option dokumentiert.
- [ ] Hoestens ein kleiner Folgeslice empfohlen oder alles deferred.
- [ ] SSRF-/Fetch-Grenzen und Contract-Fixture-Auswirkungen benannt.
- [ ] Kein Analyzer-Slice wird an eine externe Analyzer-API gekoppelt,
  solange Tranche 2 diese nicht freigibt.
- [ ] Tranche enthaelt `What aendert sich` /
  `What bleibt unveraendert` mit Dateinachweis.

Go/No-Go:

- **Go:** ein eng begrenzter Analyzer-Slice mit Test- und Fixture-Plan.
- **No-Go:** vollstaendige CMAF-Vollanalyse, Codec-Decoding oder
  Player-Laufzeitpfad ohne getrennten Plan.

Vorlaeufige Artefakte:

- Analyzer-Folge-Scope-Matrix.
- POC-/Defer-Notiz.
- Optionaler Planvorschlag fuer `0.16.0` oder spaeter.

## 7. Tranche 5 — Ops-Trigger-Re-Eval und Release-Closeout

Ziel: `MVP-40` und `MVP-41` bleiben sichtbar getrackt, werden aber
nur bei echten Triggern in Implementierung ueberfuehrt.

DoD:

- [ ] Postgres-Trigger aus ADR 0005 geprueft:
  Multi-Replica-Store, Recovery-SLO, Retention-/Query-Last.
- [ ] Analytics-Trigger geprueft:
  > 50 Mio. Events/Tag, Ad-hoc-Analysebedarf, Owner + Abbruchdatum.
- [ ] Falls kein Trigger erreicht ist: Defer-Status mit Datum und
  Owner aktualisiert.
- [ ] Falls Trigger erreicht ist: Folgeplan statt stiller Umsetzung
  angelegt oder als Blocker fuer `0.15.0` markiert.
- [ ] Tranche enthaelt `What aendert sich` /
  `What bleibt unveraendert` mit Dateinachweis.
- [ ] RAK-Verifikationsmatrix vollstaendig ausgefuellt.
- [ ] `make docs-check` gruen.
- [ ] Bei dokumentarischem Szenario ohne Code-/Release-Skript-
  Aenderungen: Code-, Build- und Security-Gates sind als `n/a` mit
  Begruendung im Closeout markiert.
- [ ] Bei codebezogenen Aenderungen: `make build` und `make gates`
  gruen.
- [ ] Bei codebezogenen/Release-Aenderungen: `make security-gates`
  gruen oder CI-Job `Security gates` gruen dokumentiert.
- [ ] Versions-Bump auf `0.15.0` vollstaendig durchgefuehrt.
- [ ] `CHANGELOG.md` mit `[0.15.0] - YYYY-MM-DD` aktualisiert.
- [ ] Roadmap auf released `0.15.0` und naechste Folgephase
  umgestellt.
- [ ] Plan nach `docs/planning/done/plan-0.15.0.md` verschoben,
  Status auf `✅ released`.
- [ ] Annotierter Tag `v0.15.0` erstellt.

## 8. RAK-Verifikationsmatrix (Platzhalter)

Wird bei Aktivierung nach dem `0.14.0`-Closeout mit finalen RAK-IDs
gefuellt. Bis dahin dienen die folgenden Zeilen als Zuschnittsvorschlag.

| RAK | Prioritaet | Nachweis | Akzeptanz | Status |
| --- | --- | --- | --- | --- |
| RAK-TBD-1 | Muss | Zielgruppen-Decision, Lastenheft §16.1 oder ADR | Primaerzielgruppe ist entschieden; Plattform-Scope ist explizit deferred oder als spaeterer Planpfad beschrieben | [ ] |
| RAK-TBD-2 | Muss | Analyzer-Boundary-Decision zu `MVP-20` | Externe Analyzer-API ist `proceed`, `POC`, `defer` oder `anders erfuellt`; Trigger sind messbar | [ ] |
| RAK-TBD-3 | Muss | `F-132` Control-Plane-Decision | Control-Plane hat Scope, Nicht-Ziele und Trigger; keine Implementierung ohne Folgeplan | [ ] |
| RAK-TBD-4 | Muss | `NF-13` Folge-Scope-Matrix | Naechster Analyzer-Slice ist eng zugeschnitten oder bewusst deferred | [ ] |
| RAK-TBD-5 | Muss | Postgres-/Analytics-Trigger-Re-Eval | `MVP-40`/`MVP-41` bleiben deferred oder bekommen einen separaten Folgeplan bei erreichtem Trigger | [ ] |

Sofort nutzbares Verifikationsmapping (bei Aktivierung auszufuellen):

| RAK | Primaere Datei(en) | Datum | Owner | Status |
| --- | --- | --- | --- | --- |
| RAK-TBD-1 | TBD | TBD | Product/PM | ⬜ |
| RAK-TBD-2 | TBD | TBD | Platform/Analyzer | ⬜ |
| RAK-TBD-3 | TBD | TBD | Platform/Product | ⬜ |
| RAK-TBD-4 | TBD | TBD | Platform/QA | ⬜ |
| RAK-TBD-5 | TBD | TBD | Platform/Ops | ⬜ |

## 8.1 Blocker-Log (Startzustand)

| Blocker | Betroffene Tranche | Erwartete Aufloesung |
| --- | --- | --- |
| `0.14.0` noch nicht released | alle | Vorgaenger-Gate in §0.2 schliessen |
| RAK-Range noch offen | Tranche 0/5 | Lastenheft-Patch bei Aktivierung vergeben |
| Zielgruppe nicht entschieden | Tranche 2/3 | Tranche 1 vor API-/Control-Plane-Go schliessen |
| Kein externer Analyzer-Konsument | Tranche 2 | Analyzer-API deferred oder Trigger definieren |
| Keine Ops-Trigger erreicht | Tranche 5 | Postgres/Analytics weiter deferred, kein Runtime-Scope |

## 9. Folge-Scope nach `0.15.0`

- Spaeter: externer Analyzer-API-Slice, falls `MVP-20` auf `proceed`
  oder `POC` gesetzt wird.
- Spaeter: Control-Plane-Plan, falls `F-132` konkrete Betreiber- oder
  Multi-Project-Anforderungen bekommt.
- Spaeter: begrenzter Analyzer-Folge-Slice aus `NF-13`, falls Tranche 4
  einen priorisierten Scope empfiehlt.
- Spaeter: Postgres- oder Analytics-Backend-Plan, falls die in ADR 0005
  dokumentierten Trigger erreicht werden.
