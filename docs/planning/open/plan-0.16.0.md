# Implementation Plan — `0.16.0` (Selected Product Slice)

> **Status**: ⬜ offen — vorbereiteter Folgeplan nach `0.15.0`.
> Aktivierung erst nach Abschluss und Closeout von `0.15.0`.
>
> **Vorgänger**: `0.15.0` (Product Scope / Analyzer Boundary),
> aktuell in Umsetzung in
> [`../in-progress/plan-0.15.0.md`](../in-progress/plan-0.15.0.md).
>
> **Release-Typ**: voraussichtlich Minor-Release mit Lastenheft-Patch,
> neuer RAK-Gruppe und Tag `v0.16.0`. Erwartete RAK-Range:
> `RAK-106`..`RAK-110`, sofern keine Zwischenreleases RAKs belegen.
>
> **Ziel**: `0.16.0` setzt genau einen in `0.15.0` freigegebenen
> Product-/Analyzer-/Platform-Slice um oder dokumentiert bewusst, dass
> alle Pfade weiter deferred bleiben. Der Release darf nicht mehrere
> grosse Richtungsentscheidungen gleichzeitig in Code ueberfuehren.
>
> **Bezug**:
> [`../in-progress/roadmap.md`](../in-progress/roadmap.md),
> [`../../../spec/lastenheft.md`](../../../spec/lastenheft.md)
> §16.1, `MVP-20`, `F-132`, `NF-13`, `MVP-40`, `MVP-41`,
> [`../in-progress/plan-0.15.0.md`](../in-progress/plan-0.15.0.md).
>
> **Nachfolger**: offen.

## 0. Konvention

DoD-Checkboxen tracken den Lieferstand:

- `[x]` ausgeliefert mit Commit-Hash.
- `[ ]` offen.
- `[!]` blockiert durch ADR- oder Scope-Entscheidung.
- 🟡 in Arbeit.

### 0.1 Scope-Definition

`0.16.0` ist kein zweiter Decision-Plan. Er importiert die `0.15.0`-
Entscheidung und setzt hoechstens einen klar begrenzten Folgepfad um.

Vorlaeufig in Scope, abhaengig vom `0.15.0`-Closeout:

- Externe Analyzer-API als schmaler POC oder erster Slice, falls
  `MVP-20` in `0.15.0` auf `proceed` oder `POC` gesetzt wird.
- Ein kleiner Analyzer-Folge-Slice aus `NF-13`, falls `0.15.0` ihn
  priorisiert und der Slice ohne neue Plattformabhaengigkeit lieferbar
  ist.
- Control-Plane-POC nur als minimales, nicht produktives Geruest,
  falls `F-132` in `0.15.0` ausdruecklich `POC` freigibt und die
  Zielgruppenentscheidung dies rechtfertigt.
- Postgres- oder Analytics-POC nur, falls die in ADR 0005 und
  `0.15.0` dokumentierten Trigger nachweisbar erreicht wurden.
- Trigger-/Defer-Release, falls kein Implementierungspfad freigegeben
  wurde.

Vorlaeufig out of scope:

- Keine parallele Umsetzung von Analyzer-API, Control-Plane und
  Backend-Store in einem Release.
- Kein Production-Control-Plane-Betrieb.
- Kein Multi-Tenant-SaaS-Produkt ohne eigenen Folgeplan.
- Kein Postgres-Default und keine automatische SQLite-Migration ohne
  ausgelösten Trigger und eigenen Migrations-/Rollback-Nachweis.
- Kein Hochvolumen-Analytics-Pflichtbackend ohne Workload-Trigger,
  Owner, Zeitgrenze und Abbruchkriterien.
- Kein Production-Kubernetes-Ausbau als Nebenprodukt dieses Plans.

### 0.2 Vorgänger-Gate

Vor Aktivierung von `0.16.0` muessen diese Bedingungen erfuellt sein:

- [ ] `0.15.0` ist released und als
  `docs/planning/done/plan-0.15.0.md` archiviert.
- [ ] Roadmap zeigt `0.16.0` als aktive Folgephase oder begruendet
  einen anderen Nachfolger.
- [ ] Die `0.15.0`-RAK-Matrix ist geschlossen oder enthaelt explizite
  Defer-/Blockerstatus.
- [ ] Genau ein Implementierungs- oder POC-Pfad wurde freigegeben,
  oder Tranche 0 waehlt bewusst ein Trigger-/Defer-Release.
- [ ] Der freigegebene Pfad hat Owner, Gating, Abbruchkriterien und
  Backwards-Compat-Grenzen.

Uebergangsausnahme bei Release-Freeze oder blockiertem Vorgaenger:
Tranche 0 darf `0.16.0` nur als **draft/in-progress Planning** starten,
wenn Roadmap und Blocker-Log den fehlenden `0.15.0`-Closeout
begründen, alle aus `0.15.0` importierten Entscheidungen als `[!]`
blockiert markiert sind und kein Tag/Release fuer `0.16.0` erstellt
wird, bevor `0.15.0` archiviert oder ausdruecklich durch einen neuen
Plan ersetzt ist.

### 0.3 Lastenheft-Patch (TBD)

Die finale RAK-Gruppe wird erst bei Aktivierung vergeben. Erwartete
RAK-Themen, falls keine Zwischen-Ranges belegt werden:

| Vorlaeufige Kennung | Thema | Bedingung |
| --- | --- | --- |
| RAK-TBD-1 / RAK-106 | Import der `0.15.0`-Entscheidung | Immer aktiv; genau ein Folgeszenario oder bewusstes Defer wird gewaehlt. |
| RAK-TBD-2 / RAK-107 | Ausgewaehlter Product-/Analyzer-Slice | Nur fuer den in Tranche 0 freigegebenen Hauptpfad. |
| RAK-TBD-3 / RAK-108 | Contract-/Compatibility-Nachweis | Pflicht, sobald Code, Wire-Format oder Doku-Surface geaendert wird. |
| RAK-TBD-4 / RAK-109 | Operational-/Security-Grenzen | Pflicht fuer API-, Control-Plane-, Backend- oder Fetch-Pfade. |
| RAK-TBD-5 / RAK-110 | Closeout und Folge-Trigger | Release-Nachweis, Defer-Trigger und naechster Planpfad. |

### 0.4 Qualitätsregeln für `0.16.0`

- Tranche 0 muss ein Szenario waehlen und alle anderen grossen Pfade
  explizit deferred lassen.
- Jeder POC hat Zeitgrenze, Erfolgskriterien und Abbruchkriterien.
- Jeder neue API-/Wire-Pfad braucht Contract-Fixtures oder begruendete
  Doku-only-Abgrenzung.
- Fetch-/Analyzer-Pfade muessen SSRF-, Redirect-, Groessen- und
  Timeout-Grenzen sichtbar halten.
- Control-Plane-Pfade duerfen keine echte User-/Org-Verwaltung,
  OAuth/OIDC oder Admin-UI-Produktzusage implizieren.
- Backend-Pfade bleiben opt-in und duerfen lokale Standardentwicklung
  nicht veraendern.

### 0.5 Aktivierungsszenarien

Tranche 0 waehlt genau eines dieser Szenarien:

| Szenario | Inhalt | Release-Charakter | Go-Bedingung |
| --- | --- | --- | --- |
| A | Analyzer-API POC (`MVP-20`) | externer API-/Job-Slice | `0.15.0` liefert konkreten externen Konsumenten und API-Grenze |
| B | Analyzer-Folge-Slice (`NF-13`) | Stream-Analyzer Feature-Slice | `0.15.0` priorisiert einen engen Analyzer-Scope mit Testplan |
| C | Control-Plane POC (`F-132`) | Platform-Prep ohne Production-Zusage | Zielgruppe und `F-132` erlauben einen nicht produktiven POC |
| D | Ops-Backend POC (`MVP-40`/`MVP-41`) | Storage-/Analytics-POC | Trigger wurden nachweisbar erreicht und eigener Rollback-/Abbruchpfad existiert |
| E | Trigger-/Defer-Release | dokumentarisch | Kein Pfad erfuellt die Go-Bedingungen |

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Erwartetes Ergebnis | Eingang | Ausgang | Status |
| --- | --- | --- | --- | --- | --- |
| 0 | Aktivierung und Szenario-Import | Ein `0.15.0`-Folgepfad verbindlich gewaehlt | `0.15.0` released | Szenario A/B/C/D/E | ⬜ |
| 1 | Scope- und Contract-Haertung | Minimaler Lieferumfang, Nicht-Ziele und Gates stehen | gewaehltes Szenario | Slice-Spezifikation | ⬜ |
| 2 | Implementierung oder POC | Code-/Doku-/POC-Artefakt fuer genau einen Pfad | Slice-Spezifikation | nachweisbarer Lieferstand | ⬜ |
| 3 | Tests, Security und Operational Boundaries | Gates und Risikoabgrenzung abgeschlossen | Implementierung/POC | Verifikationsnachweis | ⬜ |
| 4 | Release-Closeout | RAK-Matrix, Version, Changelog, Roadmap, Tag | alle aktiven Tranchen | Tag `v0.16.0` | ⬜ |

## 2. Tranche 0 — Aktivierung und Szenario-Import

Ziel: `0.16.0` uebernimmt genau einen freigegebenen Pfad aus
`0.15.0` und verhindert Scope-Drift in mehrere Plattformrichtungen.

DoD:

- [ ] Plan von `docs/planning/open/plan-0.16.0.md` nach
  `docs/planning/in-progress/plan-0.16.0.md` verschoben.
- [ ] Ausgangszustand von `git status --short` dokumentiert.
- [ ] `0.15.0`-Closeout gelesen und Szenario A/B/C/D/E ausgewaehlt.
- [ ] Lastenheft-Patch mit finaler RAK-Range ergaenzt.
- [ ] Roadmap auf `0.16.0` als aktive Folgephase umgestellt.
- [ ] Risks-Backlog aktualisiert, falls ein neuer POC ein Risiko
  ausloest oder schliesst.
- [ ] Nicht gewaehlt Pfade explizit deferred:
  - Analyzer-API,
  - Analyzer-Folge-Slice,
  - Control-Plane,
  - Postgres,
  - Analytics-Backend.
- [ ] Aktivierungsnotiz enthaelt `What aendert sich` /
  `What bleibt unveraendert` und benennt nicht aktivierte Pfade.

### 2.1 Aktivierungsnotiz (Template)

Bei Aktivierung ausfuellen:

| Feld | Wert |
| --- | --- |
| Aktivierungsdatum | TBD |
| Ausgangs-Commit | TBD |
| Gewaehltes Szenario | TBD |
| Uebernommene 0.15-Entscheidung | TBD |
| Explizit deferred | TBD |
| Blocker | TBD |
| Required Gates | TBD |

## 3. Tranche 1 — Scope- und Contract-Härtung

Ziel: Der gewaehlt Pfad wird vor Code oder POC auf einen kleinen,
testbaren Lieferumfang begrenzt.

DoD:

- [ ] Nutzer/Konsument des Slice benannt.
- [ ] Wire-/Contract-Aenderungen entweder definiert oder ausgeschlossen.
- [ ] Backwards-Compat-Grenzen dokumentiert.
- [ ] Security-/Operational-Grenzen dokumentiert.
- [ ] Erfolgskriterien und Abbruchkriterien festgelegt.
- [ ] Nicht-Ziele in diesem Plan sichtbar.
- [ ] Anti-Scope-Drift-Nachweis dokumentiert: alle nicht gewaehlten
  Szenarien bleiben deferred oder blockiert.
- [ ] Tranche enthaelt `What aendert sich` /
  `What bleibt unveraendert` mit Dateinachweis.

Szenario-spezifische Pflichtfragen:

- Analyzer-API: sync vs. async Job, Auth, Rate Limit, Retention,
  Ergebnisabruf, SSRF-Schutz.
- Analyzer-Folge-Slice: Fixture-Plan, Fetch-Grenzen, Segmentauswahl,
  Laufzeitbudget.
- Control-Plane: kein Production-Admin-Versprechen, keine User-/Org-
  Verwaltung ohne eigenen Plan.
- Ops-Backend: Migration, Rollback, Replay, Kosten-/Lastgrenzen.

## 4. Tranche 2 — Implementierung oder POC

Ziel: Genau ein begrenzter Pfad wird umgesetzt oder als POC
nachweisbar gemacht.

DoD:

- [ ] Implementierung oder POC bleibt innerhalb der Tranche-1-Grenzen.
- [ ] Keine nicht gewaehlt Pfade werden nebenbei gebaut.
- [ ] Artefakte haben klare Dateipfade.
- [ ] Feature/POC bleibt opt-in, wenn neue Infrastruktur benoetigt wird.
- [ ] Doku erklaert, was der Pfad nicht leistet.
- [ ] Anti-Scope-Drift-Nachweis dokumentiert: keine deferred Pfade
  wurden nebenbei implementiert.
- [ ] Tranche enthaelt `What aendert sich` /
  `What bleibt unveraendert` mit Dateinachweis.

Vorlaeufige Artefakte je Szenario:

- A: `apps/analyzer-api`-POC oder API-Decision-Artefakt plus
  Contract-Fixtures.
- B: Stream-Analyzer-Code, Fixtures und User-Doku fuer den
  ausgewaehlten `NF-13`-Slice.
- C: Control-Plane-POC-Artefakt oder Plan-/ADR-Geruest ohne
  Production-Zusage.
- D: Postgres-/Analytics-POC-Report, Adapter-Slice oder synthetischer
  Load-/Migrationstest.
- E: Defer-Notiz, Triggerpflege und Roadmap/Risks-Update.

## 5. Tranche 3 — Tests, Security und Operational Boundaries

Ziel: Der gewaehlte Pfad wird mit passenden Gates abgesichert.

DoD:

- [ ] `make docs-check` gruen.
- [ ] Bei Szenario E oder dokumentations-only Scope: Code-,
  Contract- und Security-Gates sind als `n/a` mit Begruendung
  dokumentiert; `make docs-check` bleibt Pflicht.
- [ ] Bei Go-/Backend-Code: `make api-test` oder `make gates` gruen.
- [ ] Bei TypeScript-/Analyzer-Code: `make ts-test` oder passender
  Package-Test gruen.
- [ ] Bei Security-relevantem Pfad: `make security-gates` gruen oder
  CI-Job `Security gates` gruen dokumentiert.
- [ ] Contract-/Fixture-Drift geprueft, falls Wire- oder Analyzer-
  Result-Schema geaendert wurde.
- [ ] Risks-Backlog aktualisiert oder explizit unveraendert markiert.
- [ ] Anti-Scope-Drift-Nachweis dokumentiert: Gates beziehen sich nur
  auf das gewaehlt Szenario.
- [ ] Tranche enthaelt `What aendert sich` /
  `What bleibt unveraendert` mit Dateinachweis.

## 6. Tranche 4 — Release-Closeout

Ziel: Der Release ist nachweisbar abgeschlossen oder bewusst deferred.

DoD:

- [ ] RAK-Verifikationsmatrix vollstaendig ausgefuellt.
- [ ] Jede aktive Tranche enthaelt einen `What aendert sich` /
  `What bleibt unveraendert`-Block mit Dateinachweis.
- [ ] Doku-only-/Defer-Release markiert nicht zutreffende Build-,
  Code-, Contract- und Security-Gates explizit als `n/a` mit
  Begruendung.
- [ ] Versions-Bump auf `0.16.0` vollstaendig durchgefuehrt.
- [ ] `CHANGELOG.md` mit `[0.16.0] - YYYY-MM-DD` aktualisiert.
- [ ] Roadmap auf released `0.16.0` und naechste Folgephase
  umgestellt.
- [ ] Plan nach `docs/planning/done/plan-0.16.0.md` verschoben,
  Status auf `✅ released`.
- [ ] Annotierter Tag `v0.16.0` erstellt.

## 7. RAK-Verifikationsmatrix (Platzhalter)

Wird bei Aktivierung nach dem `0.15.0`-Closeout mit finalen RAK-IDs
gefuellt.

| RAK | Prioritaet | Nachweis | Akzeptanz | Status |
| --- | --- | --- | --- | --- |
| RAK-TBD-1 | Muss | `0.15.0`-Closeout, Szenario-Import | Genau ein Folgepfad ist gewaehlt oder alle Pfade sind deferred | [ ] |
| RAK-TBD-2 | Konditional Muss | Slice-Spezifikation und Artefaktnachweis | Der gewaehlt Product-/Analyzer-/Platform-Slice ist begrenzt geliefert oder als POC nachgewiesen | [ ] |
| RAK-TBD-3 | Konditional Muss | Contract-/Compat-Tests oder Doku-Gate | Wire-/Schema-/API-Kompatibilitaet ist belegt oder unveraendert | [ ] |
| RAK-TBD-4 | Muss | Security-/Ops-Grenzen, Risks-Backlog | Neue Risiken sind kontrolliert; keine Production-Zusage ohne Folgeplan | [ ] |
| RAK-TBD-5 | Muss | Closeout, Roadmap, Changelog, Tag | Release ist abgeschlossen; nicht gewaehlt Pfade bleiben sichtbar deferred | [ ] |

Sofort nutzbares Verifikationsmapping (bei Aktivierung auszufuellen):

| RAK | Primaere Datei(en) | Datum | Owner | Status |
| --- | --- | --- | --- | --- |
| RAK-TBD-1 | TBD | TBD | Product/PM | ⬜ |
| RAK-TBD-2 | TBD | TBD | Platform | ⬜ |
| RAK-TBD-3 | TBD | TBD | Platform/QA | ⬜ |
| RAK-TBD-4 | TBD | TBD | Platform/Ops | ⬜ |
| RAK-TBD-5 | TBD | TBD | Platform/CI | ⬜ |

## 7.1 Blocker-Log (Startzustand)

| Blocker | Betroffene Tranche | Erwartete Aufloesung |
| --- | --- | --- |
| `0.15.0` noch nicht released | alle | Vorgaenger-Gate in §0.2 schliessen |
| Kein freigegebener Folgepfad aus `0.15.0` | Tranche 0/2 | Szenario E waehlen oder `0.16.0` durch anderen Plan ersetzen |
| Mehrere konkurrierende Go-Pfade | Tranche 0 | genau einen Pfad waehlen, Rest deferred |
| RAK-Range noch offen | Tranche 0/4 | Lastenheft-Patch bei Aktivierung vergeben |

## 8. Folge-Scope nach `0.16.0`

- Spaeter: naechster Product-/Analyzer-/Platform-Slice, falls `0.16.0`
  nur einen POC liefert.
- Spaeter: Production-Backends nur bei ausgelösten Triggern und eigenem
  Migrations-/Rollback-Plan.
- Spaeter: Control-Plane-Produktisierung nur nach Zielgruppen-,
  Auth-/Policy- und Betriebsmodellentscheidung.
