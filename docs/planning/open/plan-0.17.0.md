# Implementation Plan — `0.17.0` (Productization / Next Slice)

> **Status**: ⬜ offen — vorbereiteter Folgeplan nach `0.16.0`.
> Aktivierung erst nach Abschluss und Closeout von `0.16.0`.
>
> **Vorgänger**: `0.16.0` (Selected Product Slice), vorbereitet in
> [`plan-0.16.0.md`](./plan-0.16.0.md).
>
> **Release-Typ**: voraussichtlich Minor-Release mit Lastenheft-Patch,
> neuer RAK-Gruppe und Tag `v0.17.0`. Erwartete RAK-Range:
> `RAK-111`..`RAK-115`, sofern keine Zwischenreleases RAKs belegen.
>
> **Ziel**: `0.17.0` ist der kontrollierte Anschluss an den in
> `0.16.0` gelieferten Slice. Der Release produktisiert einen
> erfolgreichen POC, liefert den naechsten kleinen Slice desselben
> Pfads oder wechselt auf den naechsten in `0.15.0`/`0.16.0`
> deferred gebliebenen Pfad. Ohne belastbaren Trigger bleibt der
> Release dokumentarisch.
>
> **Bezug**:
> [`../in-progress/roadmap.md`](../in-progress/roadmap.md),
> [`../../../spec/lastenheft.md`](../../../spec/lastenheft.md)
> §16.1, `MVP-20`, `F-132`, `NF-13`, `MVP-40`, `MVP-41`,
> [`../in-progress/plan-0.15.0.md`](../in-progress/plan-0.15.0.md),
> [`plan-0.16.0.md`](./plan-0.16.0.md).
>
> **Nachfolger**: offen.

## 0. Konvention

DoD-Checkboxen tracken den Lieferstand:

- `[x]` ausgeliefert mit Commit-Hash.
- `[ ]` offen.
- `[!]` blockiert durch ADR- oder Scope-Entscheidung.
- 🟡 in Arbeit.

### 0.1 Scope-Definition

`0.17.0` darf nur aktiv werden, wenn `0.16.0` einen klaren
Anschluss liefert. Der Plan ist absichtlich nicht auf einen
bestimmten Produktpfad festgelegt.

Vorlaeufig in Scope, abhaengig vom `0.16.0`-Closeout:

- Productization eines erfolgreichen `0.16.0`-POC, falls
  Erfolgskriterien erfuellt und Abbruchkriterien nicht verletzt sind.
- Naechster kleiner Slice desselben Pfads, falls `0.16.0` bewusst nur
  einen begrenzten ersten Teil geliefert hat.
- Wechsel auf einen deferred gebliebenen Pfad aus `0.15.0`/`0.16.0`,
  falls dieser jetzt einen besseren Trigger oder klareren Nutzen hat.
- Reines Trigger-/Defer-Release, falls kein Pfad die Go-Bedingungen
  erfuellt.

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

- [ ] `0.16.0` ist released und als
  `docs/planning/done/plan-0.16.0.md` archiviert.
- [ ] Roadmap zeigt `0.17.0` als aktive Folgephase oder begruendet
  einen anderen Nachfolger.
- [ ] Die `0.16.0`-RAK-Matrix ist geschlossen oder enthaelt explizite
  Defer-/Blockerstatus.
- [ ] Falls `0.16.0` einen POC lieferte: Erfolgskriterien,
  Abbruchkriterien, Rest-Risiken und offene Gates sind dokumentiert.
- [ ] Falls `0.16.0` ein Defer-Release war: Tranche 0 begruendet,
  warum `0.17.0` trotzdem aktiviert wird.

Uebergangsausnahme bei Release-Freeze oder blockiertem Vorgaenger:
Tranche 0 darf `0.17.0` nur als **draft/in-progress Planning** starten,
wenn Roadmap und Blocker-Log den fehlenden `0.16.0`-Closeout
begründen, alle aus `0.16.0` importierten Entscheidungen als `[!]`
blockiert markiert sind und kein Tag/Release fuer `0.17.0` erstellt
wird, bevor `0.16.0` archiviert oder ausdruecklich durch einen neuen
Plan ersetzt ist.

### 0.3 Lastenheft-Patch (TBD)

Die finale RAK-Gruppe wird erst bei Aktivierung vergeben. Erwartete
RAK-Themen, falls keine Zwischen-Ranges belegt werden:

| Vorlaeufige Kennung | Thema | Bedingung |
| --- | --- | --- |
| RAK-TBD-1 / RAK-111 | Import des `0.16.0`-Ergebnisses | Immer aktiv; Productize, Next Slice, Switch oder Defer wird gewaehlt. |
| RAK-TBD-2 / RAK-112 | Productization-/Next-Slice-Scope | Nur fuer den gewaehlt Hauptpfad. |
| RAK-TBD-3 / RAK-113 | Betriebs-/Security-Haertung | Pflicht, sobald POC-Artefakte stabiler nutzbar werden. |
| RAK-TBD-4 / RAK-114 | Compatibility- und Migration-Gates | Pflicht bei Wire-, API-, Persistenz- oder Runtime-Aenderungen. |
| RAK-TBD-5 / RAK-115 | Closeout und Folgepfad | Release-Nachweis, offene Trigger und naechste Entscheidung. |

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
| 0 | Aktivierung und `0.16.0`-Import | Szenario A/B/C/D/E gewaehlt | `0.16.0` released | finaler `0.17.0`-Scope | ⬜ |
| 1 | Evidence Review und Scope-Haertung | POC-/Slice-Belege ausgewertet, Nicht-Ziele gesetzt | `0.16.0`-Closeout | Scope-Decision | ⬜ |
| 2 | Productization, Next Slice oder Hardening | Genau ein Pfad geliefert oder deferred | Scope-Decision | Artefakt-/POC-Nachweis | ⬜ |
| 3 | Compatibility, Security und Ops Gates | Surface und Betriebsgrenzen nachgewiesen | Tranche 2 | Gate-Nachweis | ⬜ |
| 4 | Release-Closeout | RAK-Matrix, Version, Changelog, Roadmap, Tag | alle aktiven Tranchen | Tag `v0.17.0` | ⬜ |

## 2. Tranche 0 — Aktivierung und `0.16.0`-Import

Ziel: Der Plan wird nur aktiviert, wenn `0.16.0` einen klaren
Folgepfad liefert oder ein bewusst dokumentarischer Release noetig ist.

DoD:

- [ ] Plan von `docs/planning/open/plan-0.17.0.md` nach
  `docs/planning/in-progress/plan-0.17.0.md` verschoben.
- [ ] Ausgangszustand von `git status --short` dokumentiert.
- [ ] `0.16.0`-Closeout gelesen und Szenario A/B/C/D/E ausgewaehlt.
- [ ] Lastenheft-Patch mit finaler RAK-Range ergaenzt.
- [ ] Roadmap auf `0.17.0` als aktive Folgephase umgestellt.
- [ ] Risiken-Backlog aktualisiert, falls ein POC stabilisiert,
  abgebrochen oder deferred wird.
- [ ] Defer-Matrix fuer nicht gewaehlt Pfade ausgefuellt.
- [ ] Aktivierungsnotiz enthaelt `What aendert sich` /
  `What bleibt unveraendert` und benennt nicht aktivierte Pfade.

### 2.1 Aktivierungsnotiz (Template)

Bei Aktivierung ausfuellen:

| Feld | Wert |
| --- | --- |
| Aktivierungsdatum | TBD |
| Ausgangs-Commit | TBD |
| Gewaehltes Szenario | TBD |
| Import aus 0.16.0 | TBD |
| Productize / Next Slice / Switch / Defer | TBD |
| Explizit deferred | TBD |
| Blocker | TBD |
| Required Gates | TBD |

## 3. Tranche 1 — Evidence Review und Scope-Härtung

Ziel: `0.17.0` startet nicht aus Wunschdenken, sondern aus den
nachweisbaren Ergebnissen von `0.16.0`.

DoD:

- [ ] `0.16.0`-Erfolgskriterien gegen tatsaechliche Nachweise
  geprueft.
- [ ] Offene Risiken und Testluecken aus `0.16.0` gelistet.
- [ ] Productization-/Next-Slice-Entscheidung dokumentiert.
- [ ] Compatibility- und Rollback-Grenzen festgelegt.
- [ ] Nicht-Ziele und bewusst deferred Pfade dokumentiert.
- [ ] Anti-Scope-Drift-Nachweis dokumentiert: nicht importierte Pfade
  bleiben deferred oder blockiert.
- [ ] Tranche enthaelt `What aendert sich` /
  `What bleibt unveraendert` mit Dateinachweis.

Szenario-spezifische Fragen:

- Analyzer-API: Ist der Konsument real, und sind Job-/Auth-/Rate-
  Limit-Grenzen stabil genug?
- Analyzer-Slice: Sind Fixtures, Laufzeitbudget und Fetch-Grenzen
  tragfaehig?
- Control-Plane: Ist der POC noch nicht-produktiv abgegrenzt?
- Ops-Backend: Sind Trigger weiter gueltig und Migration/Rollback
  belegbar?

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

## 7. RAK-Verifikationsmatrix (Platzhalter)

Wird bei Aktivierung nach dem `0.16.0`-Closeout mit finalen RAK-IDs
gefuellt.

| RAK | Prioritaet | Nachweis | Akzeptanz | Status |
| --- | --- | --- | --- | --- |
| RAK-TBD-1 | Muss | `0.16.0`-Closeout, Szenario-Import | Productize, Next Slice, Switch, Hardening oder Defer ist eindeutig gewaehlt | [ ] |
| RAK-TBD-2 | Konditional Muss | Scope-Decision, Artefaktnachweis | Der gewaehlt Pfad ist stabilisiert, erweitert oder bewusst abgebrochen | [ ] |
| RAK-TBD-3 | Konditional Muss | Security-/Ops-Notiz, Tests | Neue oder stabilisierte Surface hat Betriebs- und Security-Grenzen | [ ] |
| RAK-TBD-4 | Konditional Muss | Contract-/Compat-Tests oder Doku-Gate | API-/Wire-/Persistenz-Kompatibilitaet ist belegt oder unveraendert | [ ] |
| RAK-TBD-5 | Muss | Closeout, Roadmap, Changelog, Tag | Release ist abgeschlossen; naechster Pfad und Defer-Status sind sichtbar | [ ] |

Sofort nutzbares Verifikationsmapping (bei Aktivierung auszufuellen):

| RAK | Primaere Datei(en) | Datum | Owner | Status |
| --- | --- | --- | --- | --- |
| RAK-TBD-1 | TBD | TBD | Product/PM | ⬜ |
| RAK-TBD-2 | TBD | TBD | Platform | ⬜ |
| RAK-TBD-3 | TBD | TBD | Platform/Ops | ⬜ |
| RAK-TBD-4 | TBD | TBD | Platform/QA | ⬜ |
| RAK-TBD-5 | TBD | TBD | Platform/CI | ⬜ |

## 7.1 Blocker-Log (Startzustand)

| Blocker | Betroffene Tranche | Erwartete Aufloesung |
| --- | --- | --- |
| `0.16.0` noch nicht released | alle | Vorgaenger-Gate in §0.2 schliessen |
| `0.16.0` liefert keinen Anschluss | Tranche 0 | Szenario E waehlen oder `0.17.0` durch anderen Plan ersetzen |
| POC-Erfolg nicht belegt | Tranche 1/2 | Productization blockieren, Hardening oder Defer waehlen |
| Mehrere konkurrierende Folgepfade | Tranche 0 | genau einen Pfad waehlen, Rest deferred |
| RAK-Range noch offen | Tranche 0/4 | Lastenheft-Patch bei Aktivierung vergeben |

## 8. Folge-Scope nach `0.17.0`

- Spaeter: weiterer Slice desselben Pfads, wenn Productization noch
  nicht abgeschlossen ist.
- Spaeter: Wechsel auf einen der weiterhin deferred Product-/Analyzer-/
  Platform-Pfade.
- Spaeter: Production-Backends nur bei ausgelösten Triggern, ADR und
  eigenem Migrations-/Rollback-Plan.
