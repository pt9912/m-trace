# Implementation Plan — `0.16.0` (Selected Product Slice)

> **Status**: 🟡 aktiv seit 2026-05-12 — Tranche 0 abgeschlossen,
> Szenario B (`NF-13` HTTP-Range-/Byte-Range-Slice) gewaehlt.
> Umsetzung erst nach Scope-/Contract-Haertung in Tranche 1.
>
> **Vorgänger**: `0.15.0` (Product Scope / Analyzer Boundary),
> released 2026-05-12 in
> [`../done/plan-0.15.0.md`](../done/plan-0.15.0.md).
>
> **Release-Typ**: Minor-Release mit Lastenheft-Patch `1.1.21`,
> neuer RAK-Gruppe `RAK-106`..`RAK-110` und geplantem Tag
> `v0.16.0`.
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
> [`../done/plan-0.15.0.md`](../done/plan-0.15.0.md).
>
> **Nachfolger**:
> [`../open/plan-0.17.0.md`](../open/plan-0.17.0.md) vorbereitet.

## 0. Konvention

DoD-Checkboxen tracken den Lieferstand:

- `[x]` ausgeliefert mit Commit-Hash.
- `[ ]` offen.
- `[!]` blockiert durch ADR- oder Scope-Entscheidung.
- 🟡 in Arbeit.

### 0.1 Scope-Definition

`0.16.0` ist kein zweiter Decision-Plan. Er importiert die `0.15.0`-
Entscheidung und setzt hoechstens einen klar begrenzten Folgepfad um.

In Scope nach Aktivierung:

- Szenario B: ein kleiner Analyzer-Folge-Slice aus `NF-13`.
- Konkret: HTTP-Range-/Byte-Range-Loader fuer manifest-referenzierte
  CMAF-Init-/Media-Segmente, abgeleitet aus `0.15.0` RAK-104.
- Tranche 1 muss Byte-Range-Scope, Fixture-Plan, SSRF-/Redirect-/
  Timeout-/Groessen-/Range-Grenzen, Laufzeitbudget und Compatibility-
  Nachweis festlegen, bevor Code umgesetzt wird.

Vorlaeufig out of scope:

- Keine parallele Umsetzung von Analyzer-API, Control-Plane und
  Backend-Store in einem Release.
- Keine externe Analyzer-API.
- Kein Production-Control-Plane-Betrieb.
- Kein Multi-Tenant-SaaS-Produkt ohne eigenen Folgeplan.
- Kein Postgres-Default und keine automatische SQLite-Migration ohne
  ausgelösten Trigger und eigenen Migrations-/Rollback-Nachweis.
- Kein Hochvolumen-Analytics-Pflichtbackend ohne Workload-Trigger,
  Owner, Zeitgrenze und Abbruchkriterien.
- Kein Production-Kubernetes-Ausbau als Nebenprodukt dieses Plans.
- Kein Low-Latency-CMAF, keine vollstaendige Segmentset-Abdeckung,
  kein Codec-Decoding und kein Player-SDK-CMAF-Laufzeitpfad.

### 0.2 Vorgänger-Gate

Vor Aktivierung von `0.16.0` muessen diese Bedingungen erfuellt sein:

- [x] `0.15.0` ist released und als
  `docs/planning/done/plan-0.15.0.md` archiviert.
- [x] Roadmap zeigt `0.16.0` als aktive Folgephase oder begruendet
  einen anderen Nachfolger.
- [x] Die `0.15.0`-RAK-Matrix ist geschlossen oder enthaelt explizite
  Defer-/Blockerstatus.
- [x] Genau ein Implementierungs- oder POC-Pfad wurde freigegeben,
  oder Tranche 0 waehlt bewusst ein Trigger-/Defer-Release.
- [x] Der freigegebene Pfad hat Owner, Gating, Abbruchkriterien und
  Backwards-Compat-Grenzen.

Uebergangsausnahme bei Release-Freeze oder blockiertem Vorgaenger:
Tranche 0 darf `0.16.0` nur als **draft/in-progress Planning** starten,
wenn Roadmap und Blocker-Log den fehlenden `0.15.0`-Closeout
begründen, alle aus `0.15.0` importierten Entscheidungen als `[!]`
blockiert markiert sind und kein Tag/Release fuer `0.16.0` erstellt
wird, bevor `0.15.0` archiviert oder ausdruecklich durch einen neuen
Plan ersetzt ist.

### 0.3 Lastenheft-Patch `1.1.21`

Die finale RAK-Gruppe fuer `0.16.0` ist `RAK-106`..`RAK-110`.

| RAK | Thema | Bedingung |
| --- | --- | --- |
| RAK-106 | Import der `0.15.0`-Entscheidung | Immer aktiv; Szenario B wird gewaehlt, alle anderen Pfade bleiben deferred. |
| RAK-107 | HTTP-Range-/Byte-Range Analyzer-Slice (`NF-13`) | Nur fuer manifest-referenzierte CMAF-Init-/Media-Segmente. |
| RAK-108 | Contract-/Compatibility-Nachweis | Pflicht, sobald Analyzer-Code, Result-Schema oder Doku-Surface geaendert wird. |
| RAK-109 | Operational-/Security-Grenzen | Pflicht fuer Fetch-Pfade: SSRF, Redirects, Timeout, Groessen und Range-Anzahl. |
| RAK-110 | Closeout und Folge-Trigger | Release-Nachweis, Defer-Trigger und naechster Planpfad. |

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
| B | Analyzer-Folge-Slice (`NF-13`) | Stream-Analyzer Feature-Slice | `0.15.0` priorisiert den HTTP-Range-/Byte-Range-Loader mit Test- und Fetch-Grenzen |
| C | Control-Plane POC (`F-132`) | Platform-Prep ohne Production-Zusage | Der RAK-103-Defer aus `0.15.0` wurde durch konkreten Betreiber-/Auth-/Tenant-Trigger wieder geoeffnet |
| D | Ops-Backend POC (`MVP-40`/`MVP-41`) | Storage-/Analytics-POC | Trigger wurden nachweisbar erreicht und eigener Rollback-/Abbruchpfad existiert |
| E | Trigger-/Defer-Release | dokumentarisch | Kein Pfad erfuellt die Go-Bedingungen |

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Erwartetes Ergebnis | Eingang | Ausgang | Status |
| --- | --- | --- | --- | --- | --- |
| 0 | Aktivierung und Szenario-Import | Ein `0.15.0`-Folgepfad verbindlich gewaehlt | `0.15.0` released | Szenario B | ✅ |
| 1 | Scope- und Contract-Haertung | Minimaler Lieferumfang, Nicht-Ziele und Gates stehen | gewaehltes Szenario | Slice-Spezifikation | ⬜ |
| 2 | Implementierung oder POC | Code-/Doku-/POC-Artefakt fuer genau einen Pfad | Slice-Spezifikation | nachweisbarer Lieferstand | ⬜ |
| 3 | Tests, Security und Operational Boundaries | Gates und Risikoabgrenzung abgeschlossen | Implementierung/POC | Verifikationsnachweis | ⬜ |
| 4 | Release-Closeout | RAK-Matrix, Version, Changelog, Roadmap, Tag | alle aktiven Tranchen | Tag `v0.16.0` | ⬜ |

## 2. Tranche 0 — Aktivierung und Szenario-Import

Ziel: `0.16.0` uebernimmt genau einen freigegebenen Pfad aus
`0.15.0` und verhindert Scope-Drift in mehrere Plattformrichtungen.

DoD:

- [x] Plan von `docs/planning/open/plan-0.16.0.md` nach
  `docs/planning/in-progress/plan-0.16.0.md` verschoben.
- [x] Ausgangszustand von `git status --short` dokumentiert.
- [x] `0.15.0`-Closeout gelesen und Szenario A/B/C/D/E ausgewaehlt.
- [x] Lastenheft-Patch mit finaler RAK-Range ergaenzt.
- [x] Roadmap auf `0.16.0` als aktive Folgephase umgestellt.
- [x] Risks-Backlog aktualisiert, falls ein neuer POC ein Risiko
  ausloest oder schliesst.
- [x] Nicht gewaehlt Pfade explizit deferred:
  - Analyzer-API,
  - alle Analyzer-Folge-Slices ausser HTTP-Range-/Byte-Range,
  - Control-Plane,
  - Postgres,
  - Analytics-Backend.
- [x] Aktivierungsnotiz enthaelt `What aendert sich` /
  `What bleibt unveraendert` und benennt nicht aktivierte Pfade.

### 2.1 Aktivierungsnotiz

| Feld | Wert |
| --- | --- |
| Aktivierungsdatum | 2026-05-12 |
| Ausgangs-Commit | `cdf72ec` (`main`, `origin/main`, `v0.15.0`) |
| Ausgangszustand | `git status --short --branch`: `## main...origin/main` vor dem Plan-Move |
| Gewaehltes Szenario | B — Analyzer-Folge-Slice (`NF-13`) |
| Uebernommene 0.15-Entscheidung | `RAK-104` empfiehlt HTTP-Range-/Byte-Range-Loader fuer manifest-referenzierte CMAF-Init-/Media-Segmente als einzigen kleinen Folge-Slice. `RAK-102` deferred externe Analyzer-API, `RAK-103` deferred Control-Plane, `RAK-105` deferred Postgres/Analytics ohne erreichte Trigger. |
| Explizit deferred | Externe Analyzer-API, Control-Plane, Postgres, Analytics-Backend, Production-K8s, Low-Latency-CMAF, vollstaendige Segmentsets, Codec-Decoding, Player-SDK-CMAF-Laufzeitpfade. |
| Blocker | Keine offenen Blocker fuer Tranche 0. Tranche 1 muss Byte-Range-Scope, Fixtures, SSRF-/Redirect-/Timeout-/Groessen-/Range-Limits und Compatibility-Gates festlegen. |
| Required Gates | Tranche 0 docs-only: `make docs-check`. Ab Tranche 2 je nach Artefakt: passender Stream-Analyzer-/TS-Test, Contract-/Fixture-Drift-Check, `make security-gates` oder gruenes CI-Äquivalent fuer Fetch-Security. |

### 2.2 Aktivierungsentscheid

What aendert sich:

- `0.16.0` ist aktiv und nicht mehr nur vorbereitet.
- Szenario B ist der einzige Go-Pfad: HTTP-Range-/Byte-Range-Loader
  fuer manifest-referenzierte CMAF-Init-/Media-Segmente.
- Lastenheft-Patch `1.1.21` reserviert `RAK-106`..`RAK-110` fuer
  Import, Slice, Compatibility, Security/Ops und Closeout.
- Tranche 1 wird zum zwingenden Scope- und Gate-Filter vor Code.

What bleibt unveraendert:

- Interner `apps/analyzer-service` plus `@npm9912/stream-analyzer`
  Library/CLI bleiben der aktuelle Analyzer-Standardpfad.
- Keine externe Analyzer-API, keine Control-Plane, kein Postgres-
  Default, kein Analytics-Pflichtbackend und kein Production-K8s.
- Der CMAF-Scope bleibt klein: keine LL-CMAF-Unterstuetzung, keine
  vollstaendige Segmentset-Abdeckung, kein Codec-Decoding, kein
  Player-Laufzeitpfad.
- Bestehende Wire-/Runtime-Defaults bleiben durch Tranche 0
  unveraendert.

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
- Analyzer-Folge-Slice: HTTP-Range-/Byte-Range-Scope, Fixture-Plan,
  Fetch-Grenzen, Segmentauswahl, Laufzeitbudget.
- Control-Plane: nach `0.15.0` RAK-103 nicht freigegeben, solange
  kein konkreter Betreiber-/Auth-/Tenant-Trigger vorliegt; kein
  Production-Admin-Versprechen, keine User-/Org-Verwaltung ohne
  eigenen Plan.
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
- B: Stream-Analyzer-Code, Fixtures und User-Doku fuer den HTTP-
  Range-/Byte-Range-Slice.
- C: nur bei wieder geoeffnetem RAK-103-Trigger: Control-Plane-POC-
  Artefakt oder Plan-/ADR-Geruest ohne Production-Zusage.
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

## 7. RAK-Verifikationsmatrix

Die finalen RAK-IDs fuer `0.16.0` sind mit Lastenheft-Patch `1.1.21`
vergeben.

| RAK | Prioritaet | Nachweis | Akzeptanz | Status |
| --- | --- | --- | --- | --- |
| RAK-106 | Muss | `0.15.0`-Closeout, Szenario-Import | Genau ein Folgepfad ist gewaehlt; Szenario B ist aktiv, alle anderen grossen Pfade bleiben deferred | [x] |
| RAK-107 | Muss | Slice-Spezifikation und Artefaktnachweis | HTTP-Range-/Byte-Range-Loader fuer manifest-referenzierte CMAF-Init-/Media-Segmente ist begrenzt geliefert oder bewusst deferred | [ ] |
| RAK-108 | Konditional Muss | Contract-/Compat-Tests oder Doku-Gate | Analyzer-Result-Schema-/API-Kompatibilitaet ist belegt oder unveraendert | [ ] |
| RAK-109 | Muss | Security-/Ops-Grenzen, Risks-Backlog | Fetch-Risiken sind kontrolliert; keine externe API-/Control-Plane-/Backend-Zusage entsteht nebenbei | [ ] |
| RAK-110 | Muss | Closeout, Roadmap, Changelog, Tag | Release ist abgeschlossen; nicht gewaehlt Pfade bleiben sichtbar deferred | [ ] |

Sofort nutzbares Verifikationsmapping:

| RAK | Primaere Datei(en) | Datum | Owner | Status |
| --- | --- | --- | --- | --- |
| RAK-106 | `docs/planning/in-progress/plan-0.16.0.md`, `docs/planning/done/plan-0.15.0.md` §6.3/§9, `spec/lastenheft.md` §13.20 | 2026-05-12 | Product/PM | ✅ |
| RAK-107 | `docs/planning/in-progress/plan-0.16.0.md`, spaeter Stream-Analyzer-Artefakte | TBD | Platform/Analyzer | ⬜ |
| RAK-108 | `docs/planning/in-progress/plan-0.16.0.md`, spaeter Contract-/Fixture-Nachweise | TBD | Platform/QA | ⬜ |
| RAK-109 | `docs/planning/in-progress/plan-0.16.0.md`, `docs/planning/in-progress/risks-backlog.md`, spaeter Security-Gates | TBD | Platform/Ops | ⬜ |
| RAK-110 | `docs/planning/in-progress/plan-0.16.0.md`, `CHANGELOG.md`, Roadmap, Tag `v0.16.0` | TBD | Platform/CI | ⬜ |

## 7.1 Blocker-Log

| Blocker | Betroffene Tranche | Status |
| --- | --- | --- |
| `0.15.0` noch nicht released | alle | ✅ geschlossen: `v0.15.0` released und Plan archiviert |
| Kein freigegebener Folgepfad aus `0.15.0` | Tranche 0/2 | ✅ geschlossen: Szenario B aus `RAK-104` importiert |
| Mehrere konkurrierende Go-Pfade | Tranche 0 | ✅ geschlossen: nur Szenario B aktiv, Rest deferred |
| RAK-Range noch offen | Tranche 0/4 | ✅ geschlossen: `RAK-106`..`RAK-110` in Lastenheft `1.1.21` |

## 8. Folge-Scope nach `0.16.0`

- Spaeter: naechster Product-/Analyzer-/Platform-Slice, falls `0.16.0`
  nur einen POC liefert.
- Spaeter: Production-Backends nur bei ausgelösten Triggern und eigenem
  Migrations-/Rollback-Plan.
- Spaeter: Control-Plane-Produktisierung nur nach Zielgruppen-,
  Auth-/Policy- und Betriebsmodellentscheidung.
