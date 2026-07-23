# Implementation Plan — `0.16.0` (Selected Product Slice)

> **Status**: ✅ released 2026-05-12 — Tranchen 0–4
> geschlossen; Release-Tag `v0.16.0`.
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
> [`../../../spec/lastenheft.md`](../../../../spec/lastenheft.md)
> §16.1, `MVP-20`, `F-132`, `NF-13`, `MVP-40`, `MVP-41`,
> [`../done/plan-0.15.0.md`](../done/plan-0.15.0.md).
>
> **Nachfolger**:
> [`../done/plan-0.17.0.md`](../done/plan-0.17.0.md)
> released.

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
| 1 | Scope- und Contract-Haertung | Minimaler Lieferumfang, Nicht-Ziele und Gates stehen | gewaehltes Szenario | Slice-Spezifikation | ✅ |
| 2 | Implementierung oder POC | Code-/Doku-/POC-Artefakt fuer genau einen Pfad | Slice-Spezifikation | nachweisbarer Lieferstand | ✅ |
| 3 | Tests, Security und Operational Boundaries | Gates und Risikoabgrenzung abgeschlossen | Implementierung/POC | Verifikationsnachweis | ✅ |
| 4 | Release-Closeout | RAK-Matrix, Version, Changelog, Roadmap, Tag | alle aktiven Tranchen | Tag `v0.16.0` | ✅ |

## 2. Tranche 0 — Aktivierung und Szenario-Import

Ziel: `0.16.0` uebernimmt genau einen freigegebenen Pfad aus
`0.15.0` und verhindert Scope-Drift in mehrere Plattformrichtungen.

DoD:

- [x] Plan aus `docs/planning/open/plan-0.16.0.md` aktiviert und
  spaeter nach `docs/planning/done/plan-0.16.0.md` archiviert.
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

- [x] Nutzer/Konsument des Slice benannt.
- [x] Wire-/Contract-Aenderungen entweder definiert oder ausgeschlossen.
- [x] Backwards-Compat-Grenzen dokumentiert.
- [x] Security-/Operational-Grenzen dokumentiert.
- [x] Erfolgskriterien und Abbruchkriterien festgelegt.
- [x] Nicht-Ziele in diesem Plan sichtbar.
- [x] Anti-Scope-Drift-Nachweis dokumentiert: alle nicht gewaehlten
  Szenarien bleiben deferred oder blockiert.
- [x] Tranche enthaelt `What aendert sich` /
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

### 3.1 Tranche-1-Entscheidung

| Feld | Wert |
| --- | --- |
| Nutzer/Konsument | `@npm9912/stream-analyzer` Library/CLI und der bestehende interne `apps/analyzer-service`; kein neuer externer API-Konsument. |
| Liefer-Slice | HLS-CMAF-Binary-Verifikation fuer bereits manifest-referenzierte Byte-Ranges: `EXT-X-MAP` mit `BYTERANGE`-Attribut fuer Init-Segmente und `#EXT-X-BYTERANGE` fuer das erste fMP4-Media-Segment. |
| Ausgeschlossene Formate | DASH-Byte-Range-Support, `SegmentTimeline`/`$Time$`, LL-CMAF-Parts, chunked CMAF, vollstaendige Segmentsets, Codec-Decoding, Player-Laufzeitpfade. |
| Code-Pfad | Nur bestehender Stream-Analyzer-CMAF-Binary-Pfad: `packages/stream-analyzer/src/internal/cmaf/binary-verify.ts`, `segment-loader.ts`, HLS-CMAF-Parser-Metadaten. |
| API-/Wire-Entscheidung | Kein neues `analyzerKind`, kein neuer Top-Level-Endpoint, kein neuer `apps/analyzer-api`-Pfad. Bestehendes `details.cmaf.binary` bleibt der einzige Result-Surface. |
| Compatibility-Entscheidung | Additiv/behavioral: valide HLS-Byte-Range-Manifeste, die bisher `hls_map_byterange_unsupported` oder `hls_media_byterange_unsupported` lieferten, duerfen nach Umsetzung `passed`/fachliche Box-Failures liefern. Failure-Code-Domain bleibt stabil; die Unsupported-Codes bleiben fuer nicht umgesetzte/ungueltige Faelle verfuegbar. |
| Abbruchkriterien | Slice abbrechen oder in Doku-only-Defer drehen, wenn Range-Reads nicht ohne zusaetzliche SSRF-/Redirect-/Size-Grenzen testbar sind, wenn ein neues Public-Schema noetig wuerde oder wenn mehr als Init + erstes Media-Segment erforderlich wird. |

### 3.2 Range-Fetch-Scope

Der `0.16.0`-Slice erweitert nur den bereits vorhandenen bounded
CMAF-Binary-Pfad aus `0.10.0`.

Pflichtumfang:

- `#EXT-X-MAP:URI="...",BYTERANGE="<length>[@<offset>]"` fuer das
  Init-Segment.
- `#EXT-X-BYTERANGE:<length>[@<offset>]` direkt vor dem ersten
  fMP4-Media-Segment.
- Offset-loser Media-Range ist nur zulaessig, wenn aus dem HLS-
  Kontext ein vorheriger Range-Offset eindeutig ableitbar ist; fuer
  den ersten zu pruefenden Media-Range ohne ableitbaren Offset bleibt
  der Pfad skipped.
- Range-Ladepfad muss denselben URL-, DNS-, Redirect-, Timeout-,
  Private-Network- und Content-Type-Schutz wie `segment-loader.ts`
  nutzen.
- Range-Bytes zaehlen gegen `cmaf.binary.maxSegmentBytes`; zusaetzlich
  darf pro Segment nur genau ein Range-Request geplant werden.

Nicht-Ziele:

- Kein Multi-Range-Request und keine Wiederaufnahme ueber mehrere
  Requests.
- Keine heuristische Ermittlung weiterer Media-Segmente.
- Kein DASH-Range-/SegmentBase-Ausbau in `0.16.0`.
- Keine Aenderung der Analyzer-Service-Request-Whitelist; `fetch`
  bleibt auf `timeoutMs`, `maxBytes`, `maxRedirects` begrenzt,
  `allowPrivateNetworks` bleibt nur Service-Env.

### 3.3 Security- und Operational-Grenzen

Pflichtgrenzen fuer Tranche 2:

- Header: genau `Range: bytes=<start>-<end>`; `end` ist inklusiv und
  aus `offset + length - 1` berechnet.
- Limit: `length > 0`, `offset >= 0`, `length <= maxSegmentBytes`,
  `offset + length` darf nicht ueber `Number.MAX_SAFE_INTEGER` laufen.
- Status: Erfolgreiche Range-Antworten muessen `206 Partial Content`
  liefern. `200 OK` auf einen Range-Request ist kein stiller Erfolg,
  sondern `segment_fetch_failed` oder ein explizit dokumentierter
  skipped-Fall.
- Redirects: jeder Redirect-Hop validiert die Ziel-URL erneut; der
  `Range`-Header darf nur an den validierten Folge-Hop gehen.
- Body: gelesene Bytes muessen exakt die geplante Range-Laenge
  erreichen oder als Fetch-Failure gelten; Over-Read bricht mit
  `segment_too_large` oder Fetch-Failure ab.
- Logging/Doku: keine rohen Segment-URLs als neue Metriklabels oder
  Persistenzfelder.

### 3.4 Fixture- und Testplan

Tranche 2/3 muss mindestens diese Nachweise liefern:

- Unit-Test fuer `#EXT-X-MAP` mit `BYTERANGE`-Attribut, erfolgreichem
  Init-Range-Fetch und bestehender Box-Validierung.
- Unit-Test fuer `#EXT-X-BYTERANGE` auf dem ersten fMP4-Media-Segment
  mit erfolgreichem Media-Range-Fetch.
- Negativtests fuer `200 OK` statt `206`, Range-Laenge ueber Limit,
  offset-losen ersten Media-Range ohne ableitbaren Offset,
  Redirect-Revalidation und Private-Network-Block ohne Opt-in.
- Contract-Fixtures fuer HLS-Map-ByteRange und HLS-Media-ByteRange
  werden aktualisiert; keine neuen Top-Level-Result-Felder.
- Drift-Nachweis: Go-Testdata-Kopien unter
  `apps/api/adapters/driven/streamanalyzer/testdata/` bleiben
  bytegleich zu den Spec-Fixtures.

### 3.5 What aendert sich

- `RAK-107` ist als HLS-Range-Fetch-Scope spezifiziert.
- `RAK-108` bekommt eine explizite No-new-public-schema-
  Compatibility-Entscheidung.
- `RAK-109` bekommt konkrete Fetch-Security-Grenzen fuer Tranche 2.

### 3.6 What bleibt unveraendert

- Externe Analyzer-API, Control-Plane, Postgres, Analytics-Backend,
  Production-K8s, LL-CMAF, vollstaendige Segmentsets, Codec-Decoding
  und Player-Laufzeitpfade bleiben deferred.
- `details.cmaf.binary` bleibt der einzige Public-Surface fuer den
  Slice.
- `make docs-check` bleibt das einzige Gate fuer Tranche 1; Code-,
  TS-, Contract- und Security-Gates werden erst mit Tranche 2/3
  verpflichtend.

## 4. Tranche 2 — Implementierung oder POC

Ziel: Genau ein begrenzter Pfad wird umgesetzt oder als POC
nachweisbar gemacht.

DoD:

- [x] Implementierung oder POC bleibt innerhalb der Tranche-1-Grenzen.
- [x] Keine nicht gewaehlt Pfade werden nebenbei gebaut.
- [x] Artefakte haben klare Dateipfade.
- [x] Feature/POC bleibt opt-in, wenn neue Infrastruktur benoetigt wird.
- [x] Doku erklaert, was der Pfad nicht leistet.
- [x] Anti-Scope-Drift-Nachweis dokumentiert: keine deferred Pfade
  wurden nebenbei implementiert.
- [x] Tranche enthaelt `What aendert sich` /
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

### 4.1 What aendert sich

- `packages/stream-analyzer/src/internal/cmaf/segment-loader.ts`:
  `loadSegment` akzeptiert optional einen einzelnen HTTP-Range-Scope
  (`offset`, `length`), sendet `Range: bytes=<start>-<end>`, verlangt
  fuer Range-Fetches `206 Partial Content` und validiert exakte
  Body-Laenge gegen `maxSegmentBytes`.
- `packages/stream-analyzer/src/internal/cmaf/binary-verify.ts`:
  HLS-`EXT-X-MAP:BYTERANGE` und erstes `#EXT-X-BYTERANGE` werden bei
  explizitem Offset in den bestehenden Binary-Check-Plan aufgenommen.
  Offset-lose oder ungueltige Ranges bleiben mit den bestehenden
  Unsupported-Codes skipped.
- `packages/stream-analyzer/tests/segment-loader.test.ts` und
  `packages/stream-analyzer/tests/binary-verify.test.ts`: Positive
  HLS-Range-Fetches sowie `200 OK` statt `206`, Short-Read, Over-Read
  und Limit-Ueberschreitung sind abgedeckt.
- `spec/contract-fixtures/analyzer/success-hls-map-byterange.json`
  und `spec/contract-fixtures/analyzer/success-hls-media-byterange.json`
  sowie die Go-Testdata-Kopien unter
  `apps/api/adapters/driven/streamanalyzer/testdata/` zeigen jetzt
  `binary.status:"passed"` fuer explizite Byte-Ranges.
- `docs/user/stream-analyzer.md`, `spec/lastenheft.md`,
  `CHANGELOG.md` und `docs/planning/in-progress/roadmap.md`
  dokumentieren den gelieferten Tranche-2-Stand.

### 4.2 What bleibt unveraendert

- Kein neuer `analyzerKind`, kein neues Result-Top-Level-Feld, kein
  neuer API-Endpoint und keine externe Analyzer-API.
- Kein DASH-Range-/SegmentBase-Ausbau, kein Multi-Range, kein
  LL-CMAF, keine vollstaendige Segmentset-Abdeckung, kein
  Codec-Decoding und kein Player-Laufzeitpfad.
- Derselbe SSRF-, DNS-, Redirect-, Timeout-, Private-Network- und
  Content-Type-Schutz wie beim bestehenden Segment-Loader bleibt
  Pflicht.
- Offset-lose HLS-Byte-Ranges werden nicht heuristisch aufgeloest;
  sie bleiben bewusst skipped.

## 5. Tranche 3 — Tests, Security und Operational Boundaries

Ziel: Der gewaehlte Pfad wird mit passenden Gates abgesichert.

DoD:

- [x] `make docs-check` gruen.
- [x] Bei Szenario E oder dokumentations-only Scope: Code-,
  Contract- und Security-Gates sind als `n/a` mit Begruendung
  dokumentiert; `make docs-check` bleibt Pflicht.
- [x] Bei Go-/Backend-Code: `make api-test` oder `make gates` gruen.
- [x] Bei TypeScript-/Analyzer-Code: `make ts-test` oder passender
  Package-Test gruen.
- [x] Bei Security-relevantem Pfad: `make security-gates` gruen oder
  CI-Job `Security gates` gruen dokumentiert.
- [x] Contract-/Fixture-Drift geprueft, falls Wire- oder Analyzer-
  Result-Schema geaendert wurde.
- [x] Risks-Backlog aktualisiert oder explizit unveraendert markiert.
- [x] Anti-Scope-Drift-Nachweis dokumentiert: Gates beziehen sich nur
  auf das gewaehlt Szenario.
- [x] Tranche enthaelt `What aendert sich` /
  `What bleibt unveraendert` mit Dateinachweis.

### 5.1 Gate-Nachweis

Ausgefuehrt am 2026-05-12:

| Gate | Ergebnis | Nachweis |
| --- | --- | --- |
| `make ts-test` | ✅ gruen | Workspace-TS-Tests inklusive `packages/stream-analyzer`: 19 Testdateien / 393 Tests gruen. |
| `make ts-lint` | ✅ gruen | TypeScript-Lint, `tsc --noEmit`, Boundary- und Public-API-Checks gruen. |
| `make docs-check` | ✅ gruen | `scripts/verify-doc-refs.sh`: alle Doku-Links OK. |
| `make generated-drift-check` | ✅ gruen | Schema-/Contract-Fixture-/Public-API-Drift: OK, keine generierten Drift-Reste. |
| `make security-gates` | ✅ gruen | `govulncheck`: keine Vulnerabilities; `pnpm audit --audit-level high`: gruen; Trivy: `mtrace-api`, `mtrace-dashboard`, `mtrace-analyzer-service` jeweils 0 HIGH/CRITICAL. |
| `git diff --check` | ✅ gruen | Keine Whitespace-/Patch-Format-Fehler. |

`make api-test`/vollstaendiges `make gates` ist fuer Tranche 3 als
separates Muss nicht einschlaegig, weil Tranche 2 keinen Go-/Backend-
Code geaendert hat. Die Go-seitige Contract-Kopplung ist ueber
`make generated-drift-check` und die bytegleichen Testdata-Kopien
abgedeckt; die Security-Gates bauen und scannen die API-Runtime
trotzdem mit.

### 5.2 What aendert sich

- `docs/planning/done/plan-0.16.0.md`: Tranche 3 dokumentiert
  die ausgefuehrten Gates, die Security-Ergebnisse und die n/a-
  Begruendung fuer Go-/Backend-Code.
- `docs/planning/in-progress/risks-backlog.md`: RAK-109 ist als
  kontrolliert markiert; kein neues R-N-Item entsteht aus dem Range-
  Fetch-Slice.
- `docs/planning/in-progress/roadmap.md` und `CHANGELOG.md`:
  Tranche-3-Gate-Nachweis ist sichtbar.

### 5.3 What bleibt unveraendert

- Kein neuer API-/Backend-Pfad, kein Control-Plane-/Postgres-/
  Analytics- oder Production-K8s-Scope.
- Security-Gates beziehen sich auf den gewaehlten HLS-Range-Fetch-
  Slice und die bestehenden Runtime-Images; sie erweitern keine
  Produktzusagen.
- RAK-110/Release-Closeout bleibt offen fuer Tranche 4.

## 6. Tranche 4 — Release-Closeout

Ziel: Der Release ist nachweisbar abgeschlossen oder bewusst deferred.

DoD:

- [x] RAK-Verifikationsmatrix vollstaendig ausgefuellt.
- [x] Jede aktive Tranche enthaelt einen `What aendert sich` /
  `What bleibt unveraendert`-Block mit Dateinachweis.
- [x] Doku-only-/Defer-Release markiert nicht zutreffende Build-,
  Code-, Contract- und Security-Gates explizit als `n/a` mit
  Begruendung.
- [x] Versions-Bump auf `0.16.0` vollstaendig durchgefuehrt.
- [x] `CHANGELOG.md` mit `[0.16.0] - YYYY-MM-DD` aktualisiert.
- [x] Roadmap auf released `0.16.0` und naechste Folgephase
  umgestellt.
- [x] Plan nach `docs/planning/done/plan-0.16.0.md` verschoben,
  Status auf `✅ released`.
- [x] Annotierter Tag `v0.16.0` erstellt.

### 6.1 Release-Nachweis

| Artefakt | Nachweis |
| --- | --- |
| Versionen | `package.json`, `apps/*/package.json`, `packages/*/package.json`, `apps/api/cmd/api/main.go`, `packages/player-sdk/src/version.ts`, `contracts/sdk-compat.json`, Analyzer-Contract-Fixtures jeweils `0.16.0`. |
| Changelog | `CHANGELOG.md` enthaelt `[0.16.0] - 2026-05-12`. |
| Plan | Archiviert als `docs/planning/done/plan-0.16.0.md`. |
| Roadmap | `docs/planning/in-progress/roadmap.md` markiert `0.16.0` released und `0.17.0` als aktive Folgephase. |
| Tag | Annotierter Tag `v0.16.0`. |
| Gates | `make ts-test`, `make ts-lint`, `make docs-check`, `make generated-drift-check`, `make security-gates`, `make build`, `make sdk-performance-smoke` und `make smoke-cli` fuer den Release-Stand. |

### 6.2 What aendert sich

- `0.16.0` ist kein aktiver Plan mehr, sondern archivierter Release-
  Nachweis.
- RAK-110 ist geschlossen; `0.17.0` ist als Folgeplan unter
  `docs/planning/done/plan-0.17.0.md` released.
- Versions- und Contract-Artefakte tragen `0.16.0`.

### 6.3 What bleibt unveraendert

- Keine Aktivierung einer externen Analyzer-API, Control-Plane,
  Postgres-/Analytics-Pflicht, Production-K8s oder weiterer CMAF-
  Slices im Release-Closeout.
- Nicht gewaehlt Pfade bleiben deferred und muessen bei Bedarf in
  `0.17.0` neu getriggert werden.

## 7. RAK-Verifikationsmatrix

Die finalen RAK-IDs fuer `0.16.0` sind mit Lastenheft-Patch `1.1.21`
vergeben.

| RAK | Prioritaet | Nachweis | Akzeptanz | Status |
| --- | --- | --- | --- | --- |
| RAK-106 | Muss | `0.15.0`-Closeout, Szenario-Import | Genau ein Folgepfad ist gewaehlt; Szenario B ist aktiv, alle anderen grossen Pfade bleiben deferred | [x] |
| RAK-107 | Muss | Slice-Spezifikation und Artefaktnachweis | HTTP-Range-/Byte-Range-Loader fuer manifest-referenzierte CMAF-Init-/Media-Segmente ist begrenzt geliefert oder bewusst deferred | ✅ HLS-Slice geliefert |
| RAK-108 | Konditional Muss | Contract-/Compat-Tests oder Doku-Gate | Analyzer-Result-Schema-/API-Kompatibilitaet ist belegt oder unveraendert | ✅ No-new-public-schema, Fixtures aktualisiert |
| RAK-109 | Muss | Security-/Ops-Grenzen, Risks-Backlog | Fetch-Risiken sind kontrolliert; keine externe API-/Control-Plane-/Backend-Zusage entsteht nebenbei | ✅ Gates gruen |
| RAK-110 | Muss | Closeout, Roadmap, Changelog, Tag | Release ist abgeschlossen; nicht gewaehlt Pfade bleiben sichtbar deferred | [x] |

Sofort nutzbares Verifikationsmapping:

| RAK | Primaere Datei(en) | Datum | Owner | Status |
| --- | --- | --- | --- | --- |
| RAK-106 | `docs/planning/done/plan-0.16.0.md`, `docs/planning/done/plan-0.15.0.md` §6.3/§9, `spec/lastenheft.md` §13.20 | 2026-05-12 | Product/PM | ✅ |
| RAK-107 | `packages/stream-analyzer/src/internal/cmaf/segment-loader.ts`, `packages/stream-analyzer/src/internal/cmaf/binary-verify.ts`, `docs/planning/done/plan-0.16.0.md` §4.1 | 2026-05-12 | Platform/Analyzer | ✅ |
| RAK-108 | `spec/contract-fixtures/analyzer/success-hls-map-byterange.json`, `spec/contract-fixtures/analyzer/success-hls-media-byterange.json`, `apps/api/adapters/driven/streamanalyzer/testdata/` | 2026-05-12 | Platform/QA | ✅ |
| RAK-109 | `packages/stream-analyzer/tests/segment-loader.test.ts`, `packages/stream-analyzer/tests/binary-verify.test.ts`, `docs/planning/done/plan-0.16.0.md` §3.3/§5, `docs/planning/in-progress/risks-backlog.md` | 2026-05-12 | Platform/Ops | ✅ |
| RAK-110 | `docs/planning/done/plan-0.16.0.md`, `CHANGELOG.md`, `docs/planning/in-progress/roadmap.md`, Tag `v0.16.0` | 2026-05-12 | Platform/CI | ✅ |

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
