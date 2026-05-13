# Implementation Plan — `0.17.0` (Hardening / Evidence Review)

> **Status**: 🟡 in Arbeit seit 2026-05-13 — Tranchen 0–1
> geschlossen, Tranche 2 startet als Hardening-/Defer-Scope nach
> Evidence Review.
>
> **Vorgänger**: `0.16.0` (Selected Product Slice), released und
> archiviert in
> [`../done/plan-0.16.0.md`](../done/plan-0.16.0.md).
>
> **Release-Typ**: Minor-Release mit Lastenheft-Patch `1.1.22`,
> neuer RAK-Gruppe `RAK-111`..`RAK-115` und geplantem Tag
> `v0.17.0`.
>
> **Ziel**: `0.17.0` ist der kontrollierte Anschluss an den in
> `0.16.0` gelieferten HLS-Range-Fetch-Slice. Tranche 0 waehlt
> Szenario D: Hardening-only. Der Release prueft Belege, Gates und
> Restgrenzen des gelieferten Analyzer-Pfads, ohne sofort eine externe
> Analyzer-API, Control-Plane, Ops-Backends oder weitere CMAF-Surface
> zu aktivieren.
>
> **Bezug**:
> [`../in-progress/roadmap.md`](../in-progress/roadmap.md),
> [`../../../spec/lastenheft.md`](../../../spec/lastenheft.md)
> §16.1, `MVP-20`, `F-132`, `NF-13`, `MVP-40`, `MVP-41`,
> [`../done/plan-0.15.0.md`](../done/plan-0.15.0.md),
> [`../done/plan-0.16.0.md`](../done/plan-0.16.0.md).
>
> **Nachfolger**: offen.

## 0. Konvention

DoD-Checkboxen tracken den Lieferstand:

- `[x]` ausgeliefert mit Commit-Hash.
- `[ ]` offen.
- `[!]` blockiert durch ADR- oder Scope-Entscheidung.
- 🟡 in Arbeit.

### 0.1 Scope-Definition

`0.17.0` ist aktiviert, weil `0.16.0` released ist und einen kleinen,
abgeschlossenen Analyzer-Slice geliefert hat. Der Anschluss bleibt
bewusst Hardening-only, bis Tranche 1 einen belastbaren Beleg fuer
einen neuen Productization-, Next-Slice- oder Switch-Pfad findet.

Aktiv in Scope:

- Evidence Review des `0.16.0`-HLS-Range-Fetch-Slice gegen
  tatsaechliche Tests, Fixtures, Security-Gates und Doku.
- Hardening des bestehenden Analyzer-Pfads, falls Tranche 1 konkrete
  Luecken findet.
- Dokumentierte Go-/No-Go-Entscheidung fuer Productization, Next Slice
  oder Switch als Folgepfad, nicht als Nebenprodukt von Tranche 0.
- Triggerpflege fuer weiter deferred Product-/Platform-/Ops-Pfade.

Moegliche Importpfade:

- Analyzer-API (`MVP-20`): vom POC zu stabilem, opt-in API-/Job-Slice.
- Analyzer-Folge-Scope (`NF-13`): zweiter kleiner Analyzer-Slice oder
  Haertung des vorherigen Slice.
- Control-Plane (`F-132`): nur POC-Haertung oder Decision-Vertiefung,
  keine Production-Admin-Plattform.
- Ops-Backend (`MVP-40`/`MVP-41`): nur bei weiterhin ausgelösten
  Triggern und separatem Migrations-/Rollback- oder POC-Nachweis.

Vorlaeufig out of scope:

- Keine gleichzeitige Productization mehrerer `0.16.0`-Pfade.
- Kein Wechsel von POC zu Production, wenn `0.16.0` Erfolgskriterien
  nicht nachweisbar erfuellt.
- Kein Multi-Tenant-SaaS-Produkt ohne eigenen Plan.
- Keine Control-Plane mit echter User-/Org-Verwaltung, OAuth/OIDC
  oder Admin-UI-Zusage ohne Zielgruppen- und Auth-Decision.
- Kein Postgres-Default und kein Hochvolumen-Analytics-Pflichtbackend
  ohne neue ADR-/Planfreigabe.
- Kein Production-Kubernetes-Ausbau als Nebenprodukt.

### 0.2 Vorgänger-Gate

Vor Aktivierung von `0.17.0` muessen diese Bedingungen erfuellt sein:

- [x] `0.16.0` ist released und als
  `docs/planning/done/plan-0.16.0.md` archiviert.
- [x] Roadmap zeigt `0.17.0` als aktive Folgephase oder begruendet
  einen anderen Nachfolger.
- [x] Die `0.16.0`-RAK-Matrix ist geschlossen oder enthaelt explizite
  Defer-/Blockerstatus.
- [x] `0.16.0` war kein offener POC; Erfolgskriterien,
  Rest-Risiken und Gates sind im Release-Closeout dokumentiert.
- [x] `0.16.0` war kein Defer-Release; Tranche 0 begruendet,
  warum `0.17.0` als Hardening-only trotzdem aktiviert wird.

Uebergangsausnahme bei Release-Freeze oder blockiertem Vorgaenger:
Tranche 0 darf `0.17.0` nur als **draft/in-progress Planning** starten,
wenn Roadmap und Blocker-Log den fehlenden `0.16.0`-Closeout
begründen, alle aus `0.16.0` importierten Entscheidungen als `[!]`
blockiert markiert sind und kein Tag/Release fuer `0.17.0` erstellt
wird, bevor `0.16.0` archiviert oder ausdruecklich durch einen neuen
Plan ersetzt ist.

### 0.3 Lastenheft-Patch `1.1.22`

Die finale RAK-Gruppe fuer `0.17.0` ist `RAK-111`..`RAK-115`.

| RAK | Thema | Bedingung |
| --- | --- | --- |
| RAK-111 | Import des `0.16.0`-Ergebnisses | Immer aktiv; Szenario D wird gewaehlt, alle anderen Pfade bleiben deferred. |
| RAK-112 | Evidence Review und Hardening-Scope | Pflicht fuer den bestehenden HLS-Range-Fetch-/Analyzer-Pfad. |
| RAK-113 | Betriebs-/Security-Haertung | Pflicht, falls Tranche 1 neue Luecken in Fetch-, Fixture-, Drift- oder Security-Gates findet. |
| RAK-114 | Compatibility- und Migration-Gates | Pflicht bei Wire-, API-, Persistenz- oder Runtime-Aenderungen; sonst No-change-Nachweis. |
| RAK-115 | Closeout und Folgepfad | Release-Nachweis, offene Trigger und naechste Entscheidung. |

### 0.4 Qualitätsregeln für `0.17.0`

- POC-Productization braucht Belege aus `0.16.0`, nicht nur eine
  Absichtserklaerung.
- Jeder stabilisierte Pfad muss klar als `default`, `opt-in`,
  `experimental` oder `deprecated` gekennzeichnet sein.
- Kein neuer Pflichtdienst im lokalen Standardbetrieb ohne
  Lastenheft-Patch, Rollback und Compatibility-Nachweis.
- API-/Analyzer-Surface braucht Contract-Fixtures oder eine klare
  Aussage, dass kein Wire-Format geaendert wurde.
- Control-Plane-Arbeiten duerfen nur gegen nicht-produktive
  POC-Grenzen laufen, solange kein eigener Plattformplan existiert.
- Alle nicht gewaehlt Pfade bleiben in der Defer-Matrix sichtbar.

### 0.5 Aktivierungsszenarien

Tranche 0 waehlt genau eines dieser Szenarien:

| Szenario | Inhalt | Release-Charakter | Go-Bedingung |
| --- | --- | --- | --- |
| A | Productize `0.16.0` POC | Stabilisierung desselben Pfads | `0.16.0`-POC war erfolgreich und risikoarm fortsetzbar |
| B | Next Slice desselben Pfads | inkrementelles Feature-/API-Release | `0.16.0` lieferte bewusst nur Slice 1 und Slice 2 ist klar begrenzt |
| C | Switch zu deferred Pfad | neuer Selected Slice | Ein anderer Pfad hat inzwischen klareren Trigger/Nutzen |
| D | Hardening-only | Stabilisierung ohne neue Surface | `0.16.0` zeigt technische Schulden oder fehlende Gates |
| E | Trigger-/Defer-Release | dokumentarisch | Kein Implementierungspfad erfuellt Go-Bedingungen |

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Erwartetes Ergebnis | Eingang | Ausgang | Status |
| --- | --- | --- | --- | --- | --- |
| 0 | Aktivierung und `0.16.0`-Import | Szenario D gewaehlt | `0.16.0` released | Hardening-only Scope | ✅ |
| 1 | Evidence Review und Scope-Haertung | Slice-Belege ausgewertet, Nicht-Ziele gesetzt | `0.16.0`-Closeout | Hardening-/Defer-Decision | ✅ |
| 2 | Productization, Next Slice oder Hardening | Genau ein Pfad geliefert oder deferred | Scope-Decision | Artefakt-/POC-Nachweis | 🟡 |
| 3 | Compatibility, Security und Ops Gates | Surface und Betriebsgrenzen nachgewiesen | Tranche 2 | Gate-Nachweis | ⬜ |
| 4 | Release-Closeout | RAK-Matrix, Version, Changelog, Roadmap, Tag | alle aktiven Tranchen | Tag `v0.17.0` | ⬜ |

## 2. Tranche 0 — Aktivierung und `0.16.0`-Import

Ziel: Der Plan wird nur aktiviert, wenn `0.16.0` einen klaren
Folgepfad liefert oder ein bewusst dokumentarischer Release noetig ist.

DoD:

- [x] Plan von `docs/planning/open/plan-0.17.0.md` nach
  `docs/planning/in-progress/plan-0.17.0.md` verschoben.
- [x] Ausgangszustand von `git status --short` dokumentiert.
- [x] `0.16.0`-Closeout gelesen und Szenario A/B/C/D/E ausgewaehlt.
- [x] Lastenheft-Patch mit finaler RAK-Range ergaenzt.
- [x] Roadmap auf `0.17.0` als aktive Folgephase umgestellt.
- [x] Risiken-Backlog aktualisiert, falls ein POC stabilisiert,
  abgebrochen oder deferred wird.
- [x] Defer-Matrix fuer nicht gewaehlt Pfade ausgefuellt.
- [x] Aktivierungsnotiz enthaelt `What aendert sich` /
  `What bleibt unveraendert` und benennt nicht aktivierte Pfade.

### 2.1 Aktivierungsnotiz

`0.17.0` startet als kontrollierter Hardening-/Evidence-Review.

| Feld | Wert |
| --- | --- |
| Aktivierungsdatum | 2026-05-13 |
| Ausgangs-Commit | `2f75331` (`main`, `origin/main`, nach `0.16.0`-Closeout) |
| Ausgangszustand | `git status --short --branch`: `## main...origin/main` vor dem Plan-Move |
| Gewaehltes Szenario | D — Hardening-only |
| Import aus 0.16.0 | `0.16.0` lieferte HLS-CMAF-Byte-Range-Fetches fuer explizite `EXT-X-MAP:BYTERANGE`-/`#EXT-X-BYTERANGE`-Offsets, ohne neues Public-Schema, ohne externe Analyzer-API und mit gruenen TS-/Doku-/Drift-/Security-Gates. |
| Productize / Next Slice / Switch / Defer | Hardening-only: Tranche 1 prueft Belege, Testluecken, Compatibility und Trigger; Tranche 2 darf nur konkrete Hardening-Artefakte liefern oder den Folgepfad deferred halten. |
| Explizit deferred | Productization einer externen Analyzer-API, weiterer CMAF-/DASH-/LL-CMAF-Scope, Control-Plane, Postgres-Default, Analytics-Pflichtbackend, Production-K8s, Codec-Decoding und Player-Laufzeitpfade. |
| Blocker | Keine Blocker fuer Tranche 0. Productization, Next Slice oder Switch bleiben blockiert, bis Tranche 1 einen konkreten Konsumenten, eine Testluecke oder einen belastbaren Trigger nachweist. |
| Required Gates | Tranche 0 docs-only: `make docs-check`. Ab Tranche 2 je nach Artefakt: `make ts-test`, `make generated-drift-check`, `make security-gates` oder begruendete `n/a`-Entscheidung ohne Code-/Wire-Aenderung. |

### 2.2 Aktivierungsentscheid

What aendert sich:

- `0.17.0` ist aktiv und nicht mehr nur vorbereitet.
- Lastenheft-Patch `1.1.22` reserviert `RAK-111`..`RAK-115` fuer
  Import, Evidence Review, Hardening, Compatibility und Closeout.
- Szenario D ist der einzige Go-Pfad: Hardening-only auf Basis des
  bestehenden `0.16.0`-Analyzer-Slice.
- Tranche 1 wird zum zwingenden Gate, bevor ein Productization-,
  Next-Slice- oder Switch-Pfad wieder geoeffnet wird.

What bleibt unveraendert:

- `@npm9912/stream-analyzer` Library/CLI und der interne
  `apps/analyzer-service` bleiben Standardpfade.
- Keine externe Analyzer-API, keine Control-Plane, kein Postgres-
  Default, kein Analytics-Pflichtbackend und kein Production-K8s.
- Kein neues Analyzer-Result-Schema, kein neuer Endpoint und kein
  neuer Runtime-Default durch Tranche 0.
- HLS-Range-Fetch bleibt auf explizite manifest-referenzierte
  Byte-Ranges begrenzt.

### 2.3 Defer-Matrix

| Pfad | Status in `0.17.0` | Begruendung |
| --- | --- | --- |
| Analyzer-API (`MVP-20`) | Deferred | Kein konkreter externer Konsument, kein Job-/Retention-/Auth-/Rate-Limit-/SSRF-Vertrag nach `0.16.0`. |
| Analyzer Next Slice (`NF-13`) | Deferred bis neuer Trigger | Tranche 1 fand keinen Konsumenten-, Fixture- oder Scope-Trigger fuer weiteren DASH-/LL-CMAF-/Segmentset-Scope. |
| Control-Plane (`F-132`) | Deferred | Weiterhin kein Betreiber-, Auth-/Tenant-/Audit-Trigger und keine Production-Admin-Zusage. |
| Postgres (`MVP-40`) | Deferred mit Migration-Seed | ADR 0005 bleibt gueltig; keine neue Last-/Multi-Host-Schwelle erreicht. |
| Analytics-Backend (`MVP-41`) | Deferred | Kein Hochvolumen-Analytics-Trigger und kein Owner-/Kosten-/Rollback-Nachweis. |
| Production-K8s | Deferred | Beispielpfade bleiben optional; kein K8s-Smoke- oder Production-Ready-Gate in Tranche 0. |

## 3. Tranche 1 — Evidence Review und Scope-Härtung

Ziel: `0.17.0` startet nicht aus Wunschdenken, sondern aus den
nachweisbaren Ergebnissen von `0.16.0`.

DoD:

- [x] `0.16.0`-Erfolgskriterien gegen tatsaechliche Nachweise
  geprueft.
- [x] Offene Risiken und Testluecken aus `0.16.0` gelistet.
- [x] Productization-/Next-Slice-Entscheidung dokumentiert.
- [x] Compatibility- und Rollback-Grenzen festgelegt.
- [x] Nicht-Ziele und bewusst deferred Pfade dokumentiert.
- [x] Anti-Scope-Drift-Nachweis dokumentiert: nicht importierte Pfade
  bleiben deferred oder blockiert.
- [x] Tranche enthaelt `What aendert sich` /
  `What bleibt unveraendert` mit Dateinachweis.

Szenario-spezifische Fragen:

- Analyzer-API: Ist der Konsument real, und sind Job-/Auth-/Rate-
  Limit-Grenzen stabil genug?
- Analyzer-Slice: Sind Fixtures, Laufzeitbudget und Fetch-Grenzen
  tragfaehig?
- Control-Plane: Ist der POC noch nicht-produktiv abgegrenzt?
- Ops-Backend: Sind Trigger weiter gueltig und Migration/Rollback
  belegbar?

### 3.1 Evidence Review

Ausgefuehrt und geprueft am 2026-05-13:

| Evidence | Ergebnis | Nachweis |
| --- | --- | --- |
| `0.16.0`-RAK-Matrix | ✅ geschlossen | `docs/planning/done/plan-0.16.0.md` §7: RAK-106..RAK-110 alle ✅/[x]. |
| Range-Fetch-Unit-Tests | ✅ tragfaehig | `packages/stream-analyzer/tests/segment-loader.test.ts`: `206 Partial Content`, `200 OK` auf Range, Short-Read, Over-Read, `maxSegmentBytes`, Redirect-Revalidation und Private-Network-Block sind abgedeckt. |
| Binary-Verifier-Tests | ✅ tragfaehig | `packages/stream-analyzer/tests/binary-verify.test.ts`: `EXT-X-MAP:BYTERANGE` und erstes `#EXT-X-BYTERANGE` mit explizitem Offset passen; offset-lose Ranges bleiben mit `hls_map_byterange_unsupported` / `hls_media_byterange_unsupported` skipped. |
| Contract-Fixtures | ✅ tragfaehig | `spec/contract-fixtures/analyzer/success-hls-map-byterange.json`, `success-hls-media-byterange.json` und Go-Testdata-Kopien unter `apps/api/adapters/driven/streamanalyzer/testdata/`. |
| User-Doku | ✅ tragfaehig | `docs/user/stream-analyzer.md` §`0.16.0` Range-Fetch-Scope nennt Scope, Nicht-Ziele, Security-Grenzen und Unsupported-Codes. |
| Gate-Refresh | ✅ gruen | `make ts-test`: 38 Testdateien / 656 Tests gruen; inklusive `packages/stream-analyzer` 19 Testdateien / 393 Tests. |
| Drift-Refresh | ✅ gruen | `make generated-drift-check`: Schema-/Contract-/Public-API-Drift OK, keine generierten Drift-Reste. |

### 3.2 Scope-Decision

Tranche 1 findet keinen belastbaren Trigger fuer Productization,
Next Slice oder Switch.

Entscheidung:

- Productization einer externen Analyzer-API bleibt blockiert: kein
  konkreter externer Konsument, kein Job-/Retention-Bedarf und kein
  Auth-/Rate-Limit-/SSRF-/Contract-Nachweis.
- Next Slice im Analyzer bleibt blockiert: der naechste naheliegende
  Scope waere DASH-Range-/SegmentBase, LL-CMAF oder weitere
  Segmentset-Abdeckung; dafuer gibt es keinen neuen Konsumenten- oder
  Fixture-Trigger.
- Control-Plane bleibt blockiert: kein Betreiber-/Tenant-/Audit-
  Trigger und kein eigener Plattformplan.
- Ops-Backends bleiben triggerbasiert deferred: ADR 0005 gilt weiter;
  keine neue Last-, Multi-Host-, Analytics- oder Migration-Schwelle.
- Tranche 2 bleibt Hardening-only. Wenn keine neue Luecke beim
  Tranche-2-Start gefunden wird, darf Tranche 2 als Doku-/Defer-
  Artefakt abgeschlossen werden.

### 3.3 Offene Risiken und Testlücken

Keine neue R-N-Risiko-ID entsteht aus Tranche 1.

Bewusst verbleibende Luecken:

- Offset-lose HLS-Byte-Ranges werden nicht heuristisch aufgeloest;
  sie bleiben skipped. Das ist Scope-Grenze, kein Bug.
- DASH-Range-/SegmentBase und LL-CMAF bleiben ohne eigenen
  Folge-Trigger out of scope.
- Es gibt keinen externen Analyzer-API-Vertrag, keine Auth-/Rate-
  Limit-Grenze und keinen Retention-/Job-Lifecycle fuer API-Nutzung.
- Go-/Backend-Code wurde durch `0.16.0` nicht geaendert; die Kopplung
  bleibt ueber Contract-Fixtures und `make generated-drift-check`
  abgesichert.

### 3.4 Compatibility- und Rollback-Grenzen

- Keine Wire-, API-, Persistenz-, Contract- oder Runtime-Aenderung in
  Tranche 1.
- `details.cmaf.binary` bleibt die einzige Analyzer-Surface fuer den
  Range-Fetch-Slice.
- Rollback bleibt trivial: ohne Code-/Runtime-Aenderung ist kein
  Deaktivierungspfad noetig; fuer spaetere Analyzer-Aenderungen bleibt
  `cmaf.binary.enabled:false` der bestehende Caller-Schalter.
- Falls Tranche 2 Code aendert, werden `make ts-test`,
  `make generated-drift-check` und bei Fetch-/Runtime-Relevanz
  `make security-gates` verpflichtend.

### 3.5 Anti-Scope-Drift-Nachweis

Tranche 1 importiert keinen deferred Pfad:

- keine neue `apps/analyzer-api`,
- kein neuer `apps/control-plane`-Pfad,
- kein Postgres-/Analytics-Backend,
- kein K8s-Production-Scope,
- kein DASH-/LL-CMAF-/Codec-/Player-Laufzeit-Scope.

### 3.6 What aendert sich

- `docs/planning/in-progress/plan-0.17.0.md`: RAK-112 ist als
  Evidence Review geschlossen; Tranche 2 startet als Hardening-/
  Defer-Scope ohne neue Product-Surface.
- `docs/planning/in-progress/roadmap.md`, `CHANGELOG.md` und
  `docs/planning/in-progress/risks-backlog.md`: Tranche-1-Stand und
  die No-new-R-N-Entscheidung sind sichtbar.

### 3.7 What bleibt unveraendert

- Keine neue externe Analyzer-API, keine Control-Plane, kein
  Postgres-Default, kein Analytics-Pflichtbackend und kein
  Production-K8s.
- Kein neues Analyzer-Result-Schema, kein neuer Endpoint, keine neue
  Runtime-Abhaengigkeit und kein Versions-Bump durch Tranche 1.
- `0.16.0`-Range-Fetch bleibt auf explizite HLS-Byte-Ranges
  begrenzt; weitere CMAF-/DASH-Ausbaustufen brauchen einen neuen
  Trigger.

## 4. Tranche 2 — Productization, Next Slice oder Hardening

Ziel: Genau ein Pfad wird fortgesetzt, stabilisiert oder bewusst
abgebrochen.

DoD:

- [ ] Umsetzung bleibt innerhalb der Tranche-1-Grenzen.
- [ ] POC-/Experimental-Status ist sichtbar, falls der Pfad nicht
  produktiv ist.
- [ ] Neue Defaults sind explizit begruendet; ansonsten bleibt der
  Pfad opt-in.
- [ ] Doku nennt Nutzer, Grenzen und Nicht-Ziele.
- [ ] Kein deferred Pfad wird nebenbei implementiert.
- [ ] Anti-Scope-Drift-Nachweis dokumentiert: Umsetzung bleibt beim
  gewaehlt Szenario.
- [ ] Tranche enthaelt `What aendert sich` /
  `What bleibt unveraendert` mit Dateinachweis.

Vorlaeufige Artefakte je Szenario:

- Productize: stabilisierte Code-/Doku-/Testpfade fuer den `0.16.0`
  Slice.
- Next Slice: kleiner weiterer Feature-Slice mit Contract-/Fixture-
  Nachweis.
- Switch: neuer kleiner Slice fuer einen bislang deferred Pfad.
- Hardening-only: Tests, Runbooks, Security-/Ops-Grenzen, keine neue
  Surface.
- Defer: Defer-Notiz, Triggerpflege, Roadmap/Risks-Update.

## 5. Tranche 3 — Compatibility, Security und Ops Gates

Ziel: Der fortgesetzte Pfad wird belastbar genug fuer den Release.

DoD:

- [ ] `make docs-check` gruen.
- [ ] Bei Szenario E, Hardening-only ohne Code oder dokumentations-only
  Scope: Code-, Contract- und Security-Gates sind als `n/a` mit
  Begruendung dokumentiert; `make docs-check` bleibt Pflicht.
- [ ] Bei Go-/Backend-Code: `make api-test` oder `make gates` gruen.
- [ ] Bei TypeScript-/Analyzer-Code: `make ts-test` oder passender
  Package-Test gruen.
- [ ] Bei Security-relevantem Pfad: `make security-gates` gruen oder
  CI-Job `Security gates` gruen dokumentiert.
- [ ] Contract-/Fixture-Drift geprueft, falls Wire- oder Analyzer-
  Result-Schema geaendert wurde.
- [ ] Rollback- oder Deaktivierungspfad dokumentiert, falls ein neuer
  opt-in Dienst, Adapter oder API-Pfad entsteht.
- [ ] Risks-Backlog aktualisiert oder explizit unveraendert markiert.
- [ ] Anti-Scope-Drift-Nachweis dokumentiert: Gates beziehen sich nur
  auf den fortgesetzten Pfad.
- [ ] Tranche enthaelt `What aendert sich` /
  `What bleibt unveraendert` mit Dateinachweis.

## 6. Tranche 4 — Release-Closeout

Ziel: Der Release ist nachweisbar abgeschlossen und laesst den
naechsten Pfad klar sichtbar.

DoD:

- [ ] RAK-Verifikationsmatrix vollstaendig ausgefuellt.
- [ ] Jede aktive Tranche enthaelt einen `What aendert sich` /
  `What bleibt unveraendert`-Block mit Dateinachweis.
- [ ] Doku-only-/Defer-Release markiert nicht zutreffende Build-,
  Code-, Contract- und Security-Gates explizit als `n/a` mit
  Begruendung.
- [ ] Versions-Bump auf `0.17.0` vollstaendig durchgefuehrt.
- [ ] `CHANGELOG.md` mit `[0.17.0] - YYYY-MM-DD` aktualisiert.
- [ ] Roadmap auf released `0.17.0` und naechste Folgephase
  umgestellt.
- [ ] Plan nach `docs/planning/done/plan-0.17.0.md` verschoben,
  Status auf `✅ released`.
- [ ] Annotierter Tag `v0.17.0` erstellt.

## 7. RAK-Verifikationsmatrix

Die finalen RAK-IDs fuer `0.17.0` sind mit Lastenheft-Patch `1.1.22`
vergeben.

| RAK | Prioritaet | Nachweis | Akzeptanz | Status |
| --- | --- | --- | --- | --- |
| RAK-111 | Muss | `0.16.0`-Closeout, Szenario-Import | Hardening-only ist eindeutig gewaehlt; nicht importierte Product-/Platform-/Ops-Pfade bleiben deferred | [x] |
| RAK-112 | Muss | Evidence-Review, Scope-Decision | Belege und Testluecken des `0.16.0`-Slice sind geprueft; Tranche 2 hat genau einen Hardening-Scope oder expliziten Defer | [x] |
| RAK-113 | Konditional Muss | Security-/Ops-Notiz, Tests | Neue oder stabilisierte Fetch-/Analyzer-Surface hat Betriebs- und Security-Grenzen; ohne Code-Surface ist `n/a` begruendet | [ ] |
| RAK-114 | Konditional Muss | Contract-/Compat-Tests oder Doku-Gate | API-/Wire-/Persistenz-Kompatibilitaet ist belegt oder unveraendert | [ ] |
| RAK-115 | Muss | Closeout, Roadmap, Changelog, Tag | Release ist abgeschlossen; naechster Pfad und Defer-Status sind sichtbar | [ ] |

Sofort nutzbares Verifikationsmapping (bei Aktivierung auszufuellen):

| RAK | Primaere Datei(en) | Datum | Owner | Status |
| --- | --- | --- | --- | --- |
| RAK-111 | `docs/planning/in-progress/plan-0.17.0.md`, `docs/planning/done/plan-0.16.0.md`, `spec/lastenheft.md` §13.21 | 2026-05-13 | Product/PM | ✅ |
| RAK-112 | `docs/planning/in-progress/plan-0.17.0.md` §3, `make ts-test`, `make generated-drift-check` | 2026-05-13 | Platform/Analyzer | ✅ |
| RAK-113 | `docs/planning/in-progress/plan-0.17.0.md` §5, `docs/planning/in-progress/risks-backlog.md` | TBD | Platform/Ops | ⬜ |
| RAK-114 | Contract-/fixture-/compat Nachweis oder No-change-Notiz | TBD | Platform/QA | ⬜ |
| RAK-115 | `docs/planning/done/plan-0.17.0.md`, `CHANGELOG.md`, `docs/planning/in-progress/roadmap.md`, Tag `v0.17.0` | TBD | Platform/CI | ⬜ |

## 7.1 Blocker-Log

| Blocker | Betroffene Tranche | Status |
| --- | --- | --- |
| `0.16.0` ist released | alle | ✅ geschlossen: `v0.16.0` released und Plan archiviert |
| `0.16.0` liefert keinen Anschluss | Tranche 0 | ✅ geschlossen: Hardening-only importiert den gelieferten HLS-Range-Fetch-Slice als Evidence-Review-Pfad |
| POC-Erfolg nicht belegt | Tranche 1/2 | ✅ geschlossen: `0.16.0` war kein offener POC; Productization bleibt nach Tranche 1 mangels Trigger weiterhin blockiert |
| Mehrere konkurrierende Folgepfade | Tranche 0 | ✅ geschlossen: nur Szenario D aktiv, Rest deferred |
| RAK-Range noch offen | Tranche 0/4 | ✅ geschlossen: `RAK-111`..`RAK-115` in Lastenheft `1.1.22` |

## 8. Folge-Scope nach `0.17.0`

- Spaeter: weiterer Slice desselben Pfads, wenn Productization noch
  nicht abgeschlossen ist.
- Spaeter: Wechsel auf einen der weiterhin deferred Product-/Analyzer-/
  Platform-Pfade.
- Spaeter: Production-Backends nur bei ausgelösten Triggern, ADR und
  eigenem Migrations-/Rollback-Plan.
