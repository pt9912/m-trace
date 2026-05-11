# Implementation Plan — `0.12.6` (Folge-Items-Sammlung nach `0.12.5`)

> **Status**: 🟡 Tranche 0 aktiv (aktiviert 2026-05-11). Vorgänger
> ist `0.12.5` (released 2026-05-11, Tag `v0.12.5`; Plan in
> [`plan-0.12.5.md`](../done/plan-0.12.5.md)).
>
> **Release-Typ**: **Minor-Release** (`0.12.6`) gemäß
> [`docs/user/releasing.md`](../../user/releasing.md) §3.1 —
> T0-Entscheidung 2026-05-11: alle neun R-N-Items haben einen
> Code-Pfad in diesem Release; Lastenheft-Patch `1.1.17` mit
> RAK-83..RAK-90 in §13.16. Sequenziell vor `0.13.0` (Production /
> Ops Backends). RAK-Range schiebt `0.13.0` auf `RAK-91+`.
>
> **Ziel**: Die nach `0.12.5` offen gebliebenen R-N-Items aus
> [`risks-backlog.md`](./risks-backlog.md) §1.1 systematisch
> adressieren. Plan enthält **eine Tranche pro R-N**; T0 hat alle
> neun aktiviert (Default-A aus §0.3 §2 Tranchen-Auswahl).
>
> **Bezug**:
> [`risks-backlog.md`](./risks-backlog.md) §1.1
> R-5/R-7/R-9/R-10/R-11/R-13/R-15/R-17/R-20/R-22 (R-9 wird im
> `0.13.0`-Plan bearbeitet, weil K8s-bezogen);
> [`done/plan-0.4.0.md`](../done/plan-0.4.0.md) §3.1/§4.4/§8.3
> (R-5/R-7/R-10);
> [`done/plan-0.6.0.md`](../done/plan-0.6.0.md) §4 Sub-3.3 (R-11);
> [`done/plan-0.8.5.md`](../done/plan-0.8.5.md) §2 Tranche 1 (R-13);
> [`done/plan-0.11.0.md`](../done/plan-0.11.0.md) §0.1 + §0.6
> (R-15);
> [`done/plan-0.12.5.md`](../done/plan-0.12.5.md) Tranche 2 + §11
> Folge-Scope (R-17, R-20, R-22).
>
> **Nachfolger**: [`plan-0.13.0.md`](../open/plan-0.13.0.md)
> (Production / Ops Backends, MVP-40..MVP-44). RAK-Range
> verschiebt sich: `0.12.6` belegt `RAK-83..RAK-90` in §13.16,
> `0.13.0` bekommt voraussichtlich `RAK-91..RAK-95` in §13.17.

## 0. Konvention

DoD-Checkboxen tracken den Lieferstand:

- `[x]` ausgeliefert mit Commit-Hash.
- `[ ]` offen.
- `[!]` blockiert durch ADR-/Scope-Entscheidung.
- 🟡 in Arbeit.

### 0.1 Scope-Definition

`0.12.6` ist eine **Folge-Items-Sammlung** für R-N-Einträge, die
nach dem `0.12.5`-Release in [`risks-backlog.md`](../in-progress/risks-backlog.md)
§1.1 offen oder „teilweise gelöst" sind. Pro R-N gibt es eine
Tranche, die das Item entweder strukturell auflöst (🟢) oder mit
geschärftem Trigger und konkreterem Folge-Pfad weiterträgt.

Tranchen pro R-N im Plan (Tranche-Aktivierung in T0):

| R-N  | Charakter                          | User-Surface?                     | Default-Empfehlung   |
| ---- | ---------------------------------- | --------------------------------- | -------------------- |
| R-5  | Time-Skew-Persistenz + Dashboard   | Read-Pfad-Detail-Spalte           | Patch (Doku/Schema)  |
| R-7  | `ListSessions` N+1 → Bulk-Read     | nein (interne Performance)        | Patch                |
| R-10 | Sampling-Lücken-Marker             | Read-Pfad-Detail                  | Patch (Doku/Marker)  |
| R-11 | SRT-Health-Cursor-Pagination       | Wire-Vertrag schon spec'd         | Patch (Adapter-Code) |
| R-13 | Trivy-Ignore Re-Review 2026-08-04  | nein (Wartung)                    | Patch (CI-Wartung)   |
| R-15 | Externe Media-Server-Provisionierung | neue Wire-Endpoints             | Minor (RAK-Gruppe)   |
| R-17 | Multi-Host-Issuance-Limiter        | neuer Network-Backend-Adapter     | Minor (RAK)          |
| R-20 | Produktiver Vault/KMS-Adapter      | neue Vault-Auth-Mechanismen + KMS | Minor (RAK)          |
| R-22 | Origin-/IP-Rate-Limiting           | neuer Driven-Port + ENV           | Minor (RAK)          |

In Scope (T0-Entscheidung):

- Eine **Untermenge** der neun R-N-Tranchen wird aktiviert. Items,
  die nicht aktiviert werden, bleiben im Backlog stehen — mit
  geschärftem Trigger oder Re-Eval-Notiz, falls inhaltlich
  nötig.
- Pro aktivierter Tranche: Code-Pfad (sofern relevant), Tests,
  Doku-Update, Risks-Backlog-Status-Move.

Out of Scope (auch bei voller Aktivierung):

- **R-9** (K8s-Smoke-Stage-Whitelist) — wird im `0.13.0`-Plan
  (MVP-42 K8s-Manifeste) bearbeitet, nicht hier.
- Production-Backends (`MVP-40` Postgres, `MVP-41`
  ClickHouse/VictoriaMetrics) — `0.13.0`-Scope.
- OAuth/OIDC/SSO + User-/Org-Verwaltung — RAK-71-Out-of-Scope
  bleibt normativ.
- Neue Funktionalität jenseits der R-N-Trigger.

### 0.2 Vorgänger-Gate

- `0.12.5` ist released (Tag `v0.12.5`); Lastenheft-Patch
  `1.1.16` mit RAK-77..RAK-82 in §13.15 persistiert.
- `risks-backlog.md` §1.1 enthält die hier adressierten R-N-Items
  mit aktuellem Trigger-Stand.
- Memory-Lehren aus `0.12.5`: T0-Closeout-Commit zuerst, dann
  `make gates` (Drift-Gate-Reihenfolge); Lab-Smokes nicht in
  `make gates`, sondern als opt-in Make-Target pro Adapter.

### 0.2.1 Smoke-Targets-Konvention

Die Tranche-DoDs nennen mehrere neue `make smoke-*`-Targets
(`smoke-srt-health-pagination`, `smoke-origin-rate-limit`,
`smoke-issuance-multi-host`, `smoke-vault-approle`,
`smoke-kms-skeleton`, `smoke-mediaserver-provision`). **Keines
davon existiert heute** im [`Makefile`](../../../Makefile) — sie
werden **zusammen mit dem Adapter-Code in der jeweiligen Tranche
neu angelegt**, analog zum Pattern aus `0.12.5`:

- `scripts/smoke-<name>.sh` als Bash-Wrapper, der den End-to-End-
  Test (oder reduzierte Lab-Variante) via `golang:1.26.3`-Docker
  triggert (Vorbilder: `scripts/smoke-key-rotation.sh`,
  `scripts/smoke-issuance-replica.sh`,
  `scripts/smoke-browser-ingest.sh` aus `0.12.5`).
- `Makefile`-Target plus `.PHONY`-Liste plus Help-Eintrag.
- Opt-in (NICHT in `make gates`).

DoD-Klauseln in den Tranchen lesen sich deshalb als
„`make smoke-<name>` als Operator-Smoke" und schließen
implizit das Anlegen von Script + Target ein. Sollte eine
Tranche nur ein **Test-Wrapper-Pattern** liefern (statt echtem
Compose-/Container-Smoke), wird das im Tranche-DoD explizit
markiert (vgl. `smoke-issuance-replica` aus `0.12.5` T2, das
zwei `*sql.DB`-Verbindungen statt zwei API-Prozessen nutzt).

### 0.3 Architektur-/Scope-Entscheidungen (T0)

Die T0-Aktivierung trifft mindestens diese Entscheidungen
**bevor** der Plan nach `in-progress/` wandert:

1. **Release-Typ**:
   - **Patch** `0.12.6` — wenn nur Tranchen ohne neue User-Surface
     aktiviert werden (R-5/R-7/R-10/R-11/R-13). Kein Lastenheft-
     Patch, keine RAK-Matrix.
   - **Minor** — wenn mindestens eine Tranche mit neuer Wire-/
     Adapter-Surface aktiviert wird (R-15/R-17/R-20/R-22). Dann
     Lastenheft-Patch (vermutlich `1.1.17`) mit neuer RAK-Gruppe.

2. **RAK-Range-Vorbesetzung** (nur bei Minor): Default-Empfehlung
   ist `RAK-83`..`RAK-N`, weil RAK-77..RAK-82 mit `0.12.5` belegt
   sind. Falls `0.13.0` vor `0.12.6` aktiviert wird, schiebt sich
   die Range entsprechend nach hinten — die T0-Aktivierung gegen
   den dann aktuellen Lastenheft-Stand validieren.

3. **Tranchen-Auswahl** pro R-N: aktivieren / deferr / als-Backlog-
   Eintrag-verschärfen. Default-Empfehlung pro R-N steht in der
   §0.1-Tabelle, finale Entscheidung im T0-Closeout-Commit-Body.

4. **Vorgänger-Plan-Reihenfolge gegenüber `0.13.0`**: ist `0.12.6`
   ein Zwischen-Patch vor `0.13.0` oder ein paralleler Minor?
   - **Sequenziell** (`0.12.6` vor `0.13.0`): einfacher, weil
     Lastenheft-Patches in Reihenfolge wandern und R-N-Adressierung
     vor Production-Ops kommt.
   - **Parallel**: nur sinnvoll, wenn `0.13.0` Production-Backends
     liefert, die für die R-N-Tranchen Voraussetzung sind (z. B.
     Multi-Host für R-17, KMS für R-20). Default-Empfehlung
     sequenziell.

### 0.4 Lastenheft-Patch (Vorschlag, nur bei Minor-Variante)

Bei Minor-Aktivierung ergänzt der Plan eine neue RAK-Gruppe in
einem neuen Lastenheft-Abschnitt. Genaue Patch-Version und §-
Nummer werden zur T0-Aktivierung gegen den dann aktuellen
Lastenheft-Stand bestimmt — Vorschlag bei Aktivierung **vor**
`0.13.0`:

| RAK (Platzhalter) | Bereich                | Anforderung                                                                                       |
| ----------------- | ---------------------- | ------------------------------------------------------------------------------------------------- |
| RAK-83            | Telemetry/Skew-Read    | Time-Skew-Markierung im Read-Pfad sichtbar (R-5); persistente Spalte + Dashboard-Indikator.       |
| RAK-84            | Sessions/Bulk-Read     | `ListSessions`-Performance-Garantie unter Multi-Hundert-Sessions-Hard-Cap (R-7).                  |
| RAK-85            | Telemetry/Sampling     | Sampling-Lücken-Marker im Read-Pfad (R-10); Schema-Spalte + Dashboard-Banner.                     |
| RAK-86            | SRT-Health/Cursor      | Cursor-Pagination für `GET /api/srt/health/{stream_id}` (R-11); Wire-Vertrag schon in §7a.3.       |
| RAK-87            | Ingest/Provisionierung | Optionaler MediaMTX-/SRS-Provisionierungs-Adapter (R-15).                                         |
| RAK-88            | Auth/Multi-Host-Limiter| Network-Backend-Adapter (Redis/Memcached) für Multi-Host-Issuance-Limiter (R-17 Resttrigger).      |
| RAK-89            | Auth/Vault-Production  | Produktive Vault-Authentifizierung (AppRole/IAM/K8s-ServiceAccount) plus KMS-Adapter (R-20 Resttrigger). |
| RAK-90            | Auth/IP-Limiter        | Origin-/IP-nahes Rate-Limiting als Driven-Port-Adapter (R-22).                                    |

> Die genaue RAK-Anzahl und §-Nummer hängt von der T0-aktivierten
> Tranchen-Untermenge ab. Patch-Block-Wording analog `1.1.16`
> aus `0.12.5`.

## 1. Tranchen-Übersicht

| Tranche | R-N  | Inhalt                                          | Charakter   | Status |
| ------- | ---- | ----------------------------------------------- | ----------- | ------ |
| 0       | —    | Plan-Aktivierung, Release-Typ-Entscheidung (Minor), Tranchen-Auswahl (Option A — alle 9), Lastenheft-Patch `1.1.17` §13.16 RAK-83..RAK-90, Roadmap-Insert | T0          | 🟡      |
| 1       | R-13 | Trivy-Ignore-Re-Review 2026-08-04 (Wartungspflicht) | CI-Wartung  | 🟡      |
| 2       | R-11 | SRT-Health-Detail-Cursor-Pagination             | Adapter-Code | 🟡      |
| 3       | R-5  | Time-Skew-Persistenz + Dashboard-Marker         | Schema + UI | 🟡      |
| 4       | R-10 | Sampling-Vollständigkeits-Marker                | Schema + UI | ⬜      |
| 5       | R-7  | `ListSessions` Bulk-Read-Port                   | Performance | ⬜      |
| 6       | R-22 | Origin-/IP-Rate-Limiter (Driven-Port)           | Adapter     | ⬜      |
| 7       | R-17 | Multi-Host-Issuance-Limiter (Network-Backend)   | Adapter     | ⬜      |
| 8       | R-20 | Produktiver Vault/KMS-Adapter                   | Adapter     | ⬜      |
| 9       | R-15 | Externe Media-Server-Provisionierung            | Adapter     | ⬜      |
| 10      | —    | Closeout: Versions-Bump, CHANGELOG, Plan-Move, Tag, Wave-2-Verdict | Closeout | ⬜ |

---

## 2. Tranche 0 — Aktivierung

Ziel: Release-Typ-Entscheidung, Tranchen-Auswahl, ggf.
Lastenheft-Patch + RAK-Matrix-Skelett vor erster Code-Lieferung.

DoD:

- [x] Plan von `docs/planning/open/plan-0.12.6.md` nach
  `docs/planning/in-progress/plan-0.12.6.md` verschoben
  (T0-Commit).
- [x] **Release-Typ fixiert**: **Minor `0.12.6`** (T0-Entscheidung
  2026-05-11). Begründung: R-15/R-17/R-20/R-22 sind neue
  User-Surface-Features (`provision=true`-Wire-Erweiterung,
  Redis-Backend, AppRole/KMS-Auth, Origin-Limiter-Driven-Port);
  damit greift der Regelrahmen aus `releasing.md` §3.1 für
  Lastenheft-Patch zwingend.
- [x] **Tranchen-Auswahl: Option A — alle 9 Tranchen** (T0-
  Entscheidung 2026-05-11). Begründung: vergleichbarer Scope
  zu `0.12.5` (5 Adapter-Tranchen + Closeout); Adapter sind
  hexagonal unabhängig und können parallelisiert werden;
  weniger Folge-Patches und technische Schulden, klare
  Planbarkeit für `0.12.7`/`0.13.0`. Tranchen-Liste in §1
  Übersicht spiegelt die Auswahl.
- [x] **Sequenzierung sequenziell vor `0.13.0`** (T0-Entscheidung
  2026-05-11). Begründung: vermeidet Integrationskonflikte;
  Production-Backends in `0.13.0` brauchen die R-17-Multi-Host-
  und R-20-Vault-Vorarbeit aus `0.12.6` als Voraussetzung.
- [x] Lastenheft-Patch `1.1.17` mit §13.16 RAK-83..RAK-90
  persistiert; Patch-Block-Wording in `spec/lastenheft.md`
  Frontmatter oberhalb `1.1.16`. RAK-Matrix-Skelett pro
  RAK in §13.16 ausgefüllt.
- [x] Roadmap-Insert (Closeout-Commit): §1 Phase auf
  🟡 `0.12.6` Tranche 0 aktiv; §1.1 Lastenheft-Eintrag auf
  `1.1.17`; §1.2 Nächste Phase auf `0.12.6`; §2 Schritt 47.7
  ergänzt; §3 Release-Übersicht 0.12.6-Zeile 🟡, 0.13.0-Zeile
  RAK-91+ statt 83+.
- [x] Vorgänger-Gate verifiziert: `git tag --list v0.12.5`
  liefert den Tag.
- [x] `make docs-check` grün.

## 3. Tranche 1 — R-13 Trivy-Ignore Re-Review (Wartungspflicht)

Ziel: `.security/vulnignore.yaml`-`expires` 2026-08-04 vor Ablauf
re-reviewen; entweder Verlängerung mit Begründung, Trixie-Point-
Release-Fix oder Base-Image-Wechsel (z. B. Distroless).

DoD:

- [x] Trivy-Scan gegen `node:22-trixie-slim` (Dashboard +
  Analyzer-Service) mit aktuellem Vulnerability-Stand
  (Re-Review 2026-05-11): 6 HIGH-Findings, 0 CRITICAL — exakt
  die drei bekannten CVEs aus `.security/vulnignore.yaml`,
  verteilt auf 6 Packages (libtinfo6/ncurses-base/ncurses-bin
  für CVE-2025-69720, libsystemd0/libudev1 für CVE-2026-29111,
  libcap2 für CVE-2026-4878). `Fixed Version`-Spalte leer in
  allen Treffern → **kein Upstream-Fix in Debian Trixie**
  verfügbar.
- [x] Pro CVE Status-Eintrag (Re-Review 2026-05-11):
  - `CVE-2025-69720` (ncurses): kein Trixie-Point-Release-Fix;
    Container weiterhin ohne TTY-Pfad → Vektor nicht
    erreichbar; **`expires` um 90 Tage verlängert** auf
    `2026-11-02`.
  - `CVE-2026-29111` (systemd): kein Trixie-Point-Release-Fix;
    Container weiterhin ohne systemd/IPC-Mounts → Vektor nicht
    aktivierbar; **`expires` um 90 Tage verlängert** auf
    `2026-11-02`.
  - `CVE-2026-4878` (libcap): kein Trixie-Point-Release-Fix;
    Container weiterhin als unprivileged `USER node` ohne
    setcap-Pfad; **`expires` um 90 Tage verlängert** auf
    `2026-11-02`.
- [x] `.security/vulnignore.yaml` aktualisiert: drei
  `expires`-Werte von `2026-08-04` auf `2026-11-02`; pro
  CVE-Block neuer Re-Review-Kommentar-Block mit Datum
  `2026-05-11`, Plan-Bezug `plan-0.12.6 Tranche 1`, Stand
  „kein Upstream-Fix" und Verlängerungs-Begründung.
- [x] `scripts/render-trivyignore.sh mtrace-dashboard` grün
  (3 entries gerendert, kein `expires`-Überschreitungs-Fehler).
- [!] Optional: ADR-Draft für Distroless-Base-Image-Wechsel vor
  `1.0` — bewusst **deferred**. 90 Tage Reserve bis nächstem
  Re-Review-Termin (`2026-11-02`) reichen, um den
  Distroless-Switch ohne Zeitdruck strukturell zu evaluieren;
  ein vorab-`gcr.io/distroless/nodejs22-debian12`-Wechsel
  würde den `node:22-trixie-slim`-glibc-Pfad vor einem
  konkreten Operator-Trigger durcheinander bringen.
- [x] Risks-Backlog R-13: Trigger-Stand-Eintrag aktualisiert
  mit neuem `expires`-Datum (`2026-11-02`) und Re-Review-Notiz
  (Stand `0.12.6` Tranche 1); Status bleibt ⬜ („nicht
  ausgelöst") — die Verlängerung verschiebt den nächsten
  Re-Review, löst R-13 aber nicht strukturell auf.

## 4. Tranche 2 — R-11 SRT-Health-Detail-Cursor-Pagination

Ziel: `GET /api/srt/health/{stream_id}` liefert Cursor-Pagination
über `samples_limit` hinaus.

**Wire-Vertrags-Pin** (Review-Finding 1, 2026-05-11): die
Pagination-Semantik ist in
[`spec/backend-api-contract.md`](../../../spec/backend-api-contract.md)
§7a.3 bereits präzise spezifiziert — **`samples_cursor`** als
Query-Param, **nicht** `cursor`. Plan-DoD-Wording entsprechend
gepinnt; jede Abweichung (z. B. zusätzlicher `cursor`-Alias mit
Deprecation) wäre eine Spec-Änderung mit eigenem RAK und ist
explizit **nicht** Teil dieser Tranche. Bestehende Clients/Tests
gegen `samples_cursor`/`next_cursor` bleiben unverändert
funktional.

Fehlerverhalten ist ebenfalls schon in §7a.4 normiert:
`400 cursor_invalid` bei `process_instance_id`-Mismatch (analog
§10.3). Tranche 2 implementiert die Spec, sie ändert sie nicht.

**Spec-Konsistenz-Fix** (während T2 erkannt, 2026-05-11): §7a.3 +
§7a.4 trugen Pre-§4.3-Wording (`process_instance_id`-Cursor +
monolithischer `cursor_invalid`-Body), das §10.3 mit seinem
v3-Cursor und der Zwei-Klassen-Reject-Tabelle (`cursor_invalid_legacy`
/ `cursor_invalid_malformed` / `cursor_expired`) widersprach. T2
zieht §7a.3/§7a.4 auf §10.3-Konvention nach — die Pagination
implementiert die Spec, sie ändert sie nicht: §10.3 ist die
normative Quelle, §7a.3/§7a.4 verweisen jetzt explizit darauf.

DoD:

- [x] Adapter-Implementation in
  `apps/api/adapters/driven/persistence/sqlite/srt_health_repository.go`:
  Keyset-Pagination über `(ingested_at, id)` (TEXT-Spalte mit
  RFC3339Nano UTC ist byte-lexikografisch zeit-sortiert; Index
  `idx_srt_health_samples_stream_ingested` matched WHERE+ORDER
  direkt). `limit+1`-Probe für NextAfter-Detection.
- [x] HTTP-Handler `GET /api/srt/health/{stream_id}` akzeptiert
  Query-Param **`samples_cursor`** (gemäß §7a.3) und liefert
  **`next_cursor`**-Feld in der Antwort (`omit-empty` auf der
  letzten Seite). `cursor`-Param wird **nicht** akzeptiert — kein
  Alias, kein silent-Fallback.
- [x] Cursor-Wire-Codec v3 in
  `apps/api/adapters/driving/http/cursor.go`: Token kapselt
  Collection-Scope `(pid, sid)` plus Storage-Position `(ing, id)`
  analog dem v3-Event-Cursor aus §10.3. Reject-Klassen-Mapping:
  - `v` fehlt / `v ∈ {1, 2}` → `400 cursor_invalid_legacy`.
  - Alles andere (Base64-/JSON-Decode-Fehler, unbekannter `v`,
    fehlende Pflichtfelder, fremder Project- oder Stream-Scope,
    unbekannte Zusatzfelder) → `400 cursor_invalid_malformed`.
  - Beide Bodies tragen `{"error":"<klasse>","reason":"<kurze
    Erklärung>"}` — keine Information über die konkrete
    Fehlerursache leaken (Reason ist plan-stabil, nicht
    user-input-abhängig).
- [x] Wire-Vertrag-Update in
  [`spec/backend-api-contract.md`](../../../spec/backend-api-contract.md)
  §7a.3/§7a.4: Pre-§4.3-Wording durch §10.3-konformes v3-Wording
  ersetzt (siehe **Spec-Konsistenz-Fix** oben). Body-Schema für
  beide Reject-Klassen explizit; `cursor_expired` (410) für
  Konsistenz mit §10.3 ergänzt, auch wenn ohne Retention nur via
  `make wipe` erreichbar.
- [x] Unit-Test + Adapter-Test:
  - `TestSrtHealth_HistoryCursorWalksAllPages`: 1500 Samples, 4
    Pages mit limit=400, kein Duplikat, keine Lücke,
    Sort-Invariante stabil.
  - `TestSrtHealth_HistoryCursorScopeIsolation`: Adapter liefert
    nur Samples der WHERE-Stream — kein Cross-Stream-Bleed selbst
    bei „falschem" Cursor (Defense-in-Depth; HTTP-Codec rejected
    den Pfad schon vorher).
  - HTTP-Tests: `TestSrtHealthDetail_CursorRoundTrip` (encode→
    decode via API-Roundtrip), `TestSrtHealthDetail_CursorInvalidLegacy`
    (v=2), `TestSrtHealthDetail_CursorInvalidMalformed` (sechs
    Sub-Cases: malformed Base64, unknown version, foreign
    project, foreign stream, missing fields, unknown field).
  - **Contract-Fixtures** in `spec/contract-fixtures/api/` für
    beide Reject-Klassen:
    `srt-health-cursor-invalid-legacy.json` und
    `srt-health-cursor-invalid-malformed.json` (testdata-Sync via
    `make sync-contract-fixtures`, Drift-Check in
    `make generated-drift-check`).
- [x] Smoke `make smoke-srt-health-pagination` neu angelegt:
  Wrapper-Target setzt `SMOKE_INCLUDE_MTRACE_API=1` und
  `MTRACE_SRT_HEALTH_PAGINATION=1` und ruft das existierende
  `scripts/smoke-srt-health.sh` (Konvention §0.2.1). Smoke deckt
  (a) erste Page mit `samples_limit=1` und Existenz von
  `next_cursor`, (b) Folge-Page mit gepacktem Cursor → 200, (c)
  malformed Cursor → 400 `cursor_invalid_malformed`. (a)+(b) sind
  konditional auf ≥ 2 Samples in DB (Lab-Skip ohne Fail);
  (c) ist deterministisch.
- [x] Risks-Backlog R-11: Status 🟢 mit Auflösungspfad „Cursor-
  Pagination in 0.12.6 Tranche 2 (`samples_cursor`/`next_cursor` +
  `cursor_invalid_legacy`/`cursor_invalid_malformed` gemäß
  §7a.3/§7a.4/§10.3)"; Wieder-Eröffnung bei Operator-Report über
  Inkonsistenz im Cursor-Wandern oder Schema-Drift.

## 5. Tranche 3 — R-5 Time-Skew-Persistenz + Dashboard-Marker

Ziel: `mtrace.time.skew_warning=true`-Events sind im Read-Pfad
(Dashboard ohne Tempo) sichtbar markiert.

**Implementierungs-Note** (während T3 erkannt, 2026-05-11): die
Skew-Detection lief schon seit `0.4.0` pro Event in `parseEvents`
(`now.Sub(ts).Abs() > TimeSkewThreshold`), das Ergebnis wurde aber
nur zum Batch-Flag aggregiert und ans Server-Span-Attribut gehängt
— nicht auf die Domain-Event geschrieben. T3 schreibt das gleiche
Pro-Event-Ergebnis zusätzlich aufs `domain.PlaybackEvent.
TimeSkewWarning` und persistiert es via V6-Spalte; das spec'd
„`mtrace.time.skew_warning`-Attribut tragen" aus dem alten DoD-
Wording bezog sich also auf die **Detection-Bedingung**, nicht auf
ein Wire-Feld im Ingest-Body. Wire-In bleibt unverändert; nur das
**Wire-Out** des Read-Pfads bekommt das neue Feld.

DoD:

- [x] SQLite-Schema-Erweiterung Migration `V6` mit Spalte
  `time_skew_warning INTEGER NOT NULL DEFAULT 0` an
  `playback_events` (SQLite hat keinen nativen BOOLEAN-Typ;
  Go `boolToInt`-Helper im SQLite-Adapter). Hand-gepflegt analog
  V2..V5 (schema.yaml bleibt V1-Source-of-Truth).
- [x] Ingest-Pfad: `parseEvents` setzt `TimeSkewWarning` pro
  Event basierend auf der bestehenden 60-s-Schwelle
  (`TimeSkewThreshold`); Batch-Flag (Span-Attribut) bleibt
  aggregiert wie bisher. **Wire-In bleibt unverändert** —
  Detection ist Server-seitig, kein neues Body-Feld.
- [x] Read-Pfad: `GetSessionDetail` (`eventWire.time_skew_warning`)
  und SSE-`event_appended`-Frame (`frameWire.time_skew_warning`)
  echo'en das Flag; beide mit `omitempty` (default `false`).
  `ListSessions` braucht das Feld nicht (Session-Aggregat, kein
  Event-Detail) — bei Read von Event-Detail ist es sichtbar.
- [x] Dashboard-UI: `⏱ skew`-Pin in der Session-Timeline pro Event
  mit `time_skew_warning=true`; Tooltip nennt die 60-s-Schwelle und
  das verwandte Span-Attribut. `data-testid="time-skew-indicator"`
  als Stable-Hook für künftige E2E-Tests.
- [x] Doku in [`spec/telemetry-model.md`](../../../spec/telemetry-model.md)
  §5.3 erweitert um die Persistenz auf Event-Ebene (Migration V6,
  Read-Pfad-Verhalten, Dashboard-Anzeige); §2.5-Tabellen-Eintrag
  für `mtrace.time.skew_warning` ergänzt um den `0.12.6`-
  Persistenz-Hinweis.
- [x] Tests:
  - Use-Case `TestRegisterPlaybackEventBatch_TimeSkew` erweitert
    um Pro-Event-Snapshot-Assertion.
  - `TestRegisterPlaybackEventBatch_TimeSkewPerEvent`: Mixed-Batch
    mit einem Skew-Event und einem In-Range-Event, Pro-Event-
    Flag-Mapping geprüft.
  - Adapter `TestRestartPreservesTimeSkewWarning`: Roundtrip
    durch SQLite inkl. DB-Close/Re-Open.
  - E2E `TestE2E_TimeSkewPersistedPerEvent`: HTTP-Detail-Body
    liefert `time_skew_warning=true` für skew-Event und
    `omitempty` für in-range Event.
- [x] Risks-Backlog R-5: Status 🟢 mit Auflösungspfad „Migration
  V6 + Pro-Event-Persistenz + Read-Pfad-Echo + Dashboard-Pin in
  0.12.6 Tranche 3"; Wieder-Eröffnungs-Trigger bei Operator-
  Report über fehlende Skew-Sichtbarkeit oder strukturelle
  Schwellen-Re-Eval (z. B. configurable per Project).

## 6. Tranche 4 — R-10 Sampling-Vollständigkeits-Marker

Ziel: Sampled Sessions (mit `sampleRate < 1`) sind im Read-Pfad
serverseitig erkennbar markiert — der Operator muss die
Inkompletheit nicht aus der Konfig ableiten.

**Quelle des `sample_rate`-Werts** (Review-Finding 2 / OQ2,
2026-05-11): **dediziertes Event-Feld + Immutability**, **nicht**
„erstes Event"-Heuristik. Begründung: das erste Event kann ein
beliebiger Pfad sein (Player-Start, später Buffer-Stall etc.) —
fehlt das Feld dort, würde die Session fälschlich als
voll-gesampelt markiert. Stattdessen:

- **Wire-Anforderung** an das Player-SDK: Sessions, die mit
  `sampleRate < 1` betrieben werden, müssen das in **jedem**
  Event-Body als Pflicht-Feld `meta.session_sample_rate` (oder
  analog) liefern. Voll-gesampelte Sessions dürfen das Feld
  weglassen (Default-Konstante `SAMPLE_RATE_FULL`, siehe unten).
- **Server-Verhalten**: erstmaliger Wert pro `session_id` wird in
  der Session-Metadaten-Zeile persistiert (immutable nach erstem
  gültigen Setzen). Spätere Events mit abweichendem Wert lösen
  einen Warning-Log und einen `mtrace_sample_rate_drift_total`-
  Counter aus — das Risiko wäre, dass das Operator-SDK mitten in
  einer Session den `sampleRate` ändert, was die Lücken-Erkennung
  inkorrekt machen würde.
- **Fallback**: wenn keines der eingegangenen Events das Feld
  setzt, bleibt der Wert auf `SAMPLE_RATE_FULL` — Session gilt
  als voll gesampelt (Backwards-Compat zum heutigen Verhalten).

**Präzision/Persistenz** (Review-Finding 3, 2026-05-11): keine
Float-Spalte. Stattdessen **Integer-ppm** (parts per million)
für deterministische Vergleiche:

- Persistenz-Spalte: `sample_rate_ppm INTEGER NOT NULL DEFAULT
  1000000`. Konstante `SAMPLE_RATE_FULL = 1_000_000` für „voll
  gesampelt" (entspricht 1.0). Bereich `[1, 1_000_000]`.
- Wire-Konvertierung: SDK schickt weiterhin Float `0.0 < x ≤ 1.0`
  im JSON; der Ingest-Adapter normalisiert auf Integer-ppm via
  `round(x * 1_000_000)`. Werte außerhalb `(0, 1]` werden mit
  einem `mtrace_sample_rate_invalid_total`-Counter geloggt und
  als `SAMPLE_RATE_FULL` behandelt (Fallback).
- Immutability-Check: exakter Integer-Vergleich
  `WHERE sample_rate_ppm = 1000000` (SQL-`=` auf Integer ist
  deterministisch, kein Float-Drift).
- Drift-Vergleich: `incoming_ppm != stored_ppm` mit konfig-
  baren Toleranz-Bändern (z. B. ±100 ppm = ±0.01%) im
  `mtrace_sample_rate_drift_total`-Counter; jenseits davon
  zählt es als Drift, innerhalb davon als „SDK-Rundungsartefakt"
  (silent).
- Read-API liefert beides: `sample_rate_ppm` (raw) und
  `sample_rate` (`= ppm / 1_000_000` als Float für Dashboard-
  Display). Wire-API-Konsumenten dürfen den Float-Wert nicht
  für `==`-Vergleiche nutzen — Doku-Hinweis im Wire-Block.

DoD:

- [ ] Domain-Erweiterung: `Session` bekommt ein Feld
  `SampleRatePPM int` (Default `1000000`, exported konstante
  `domain.SampleRateFull`); Persistenz immutable nach erstem
  nicht-Default-Wert. Helper `SampleRatePPMFromFloat(x float64)
  (int, error)` mit Validierung `(0, 1]`-Range.
- [ ] SDK-/Wire-Erweiterung: Player-SDK setzt
  `meta.session_sample_rate` als Float (Pflicht-Feld bei
  `sampleRate < 1`, Default-weglass bei `sampleRate == 1`).
  Schema-Eintrag in
  [`contracts/event-schema.json`](../../../contracts/event-schema.json)
  mit Session-Scope-Hinweis (kein per-Event-Wert) und Range-
  Constraint `0 < x ≤ 1.0`. Server-Adapter konvertiert via
  `SampleRatePPMFromFloat` auf Integer-ppm.
- [ ] SQLite-Schema-Migration ergänzt die Session-Spalte
  `sample_rate_ppm INTEGER NOT NULL DEFAULT 1000000`
  (`V7` o. ä.).
- [ ] Ingest-Pfad: erster Event mit normalisiertem
  `sample_rate_ppm < 1_000_000` schreibt den Wert in die
  Session-Zeile (`UPDATE … WHERE sample_rate_ppm = 1000000`
  für Immutability via Integer-Vergleich); spätere Drift wird
  geloggt und gezählt, aber **nicht** überschreibt.
- [ ] Neuer Counter `mtrace_sample_rate_drift_total{project_id}`
  in `apps/api/adapters/driven/metrics/` mit Cardinality-Limit
  (project_id).
- [ ] Read-Pfad markiert Sessions mit `sample_rate_ppm <
  SAMPLE_RATE_FULL` (`< 1_000_000`) als „sampled" — explizit
  als API-Feld plus Dashboard-Banner. Wire-Antwort liefert
  beides: `sample_rate_ppm` (int) und abgeleiteter
  `sample_rate` (float, nur Display).
- [ ] Sampling-Lücken-Heuristik (optional, T0-Entscheidung):
  Server vergleicht erwartete vs. tatsächliche Event-Anzahl bei
  bekanntem `sample_rate_ppm` und markiert auffällige Sessions
  als `"possible_loss"`. Heuristik-Schwellen sind Folge-Tuning.
  Berechnung der Erwartungswerte in Integer-Arithmetik
  (`expected = total_events * 1_000_000 / sample_rate_ppm`) —
  kein Float in der Server-Side-Logik.
- [ ] Doku-Update in
  [`spec/telemetry-model.md`](../../../spec/telemetry-model.md)
  §8.3 plus
  [`contracts/event-schema.json`](../../../contracts/event-schema.json)
  (`meta`-Allowlist + `reserved_meta_keys`) für das neue
  `session_sample_rate`-Feld.
- [ ] Tests: erster-Wert-immutable, Drift-Log, Default-`1.0`-
  Pfad, Read-Pfad-Markierung. Plus contract-fixture für
  `meta.session_sample_rate < 1`.
- [ ] Risks-Backlog R-10: Status 🟢 mit Auflösungspfad
  „dediziertes Meta-Feld + Immutability in `0.12.6` Tranche 4";
  oder geschärfter Resttrigger, wenn Lücken-Heuristik deferred
  wird.

## 7. Tranche 5 — R-7 `ListSessions` Bulk-Read-Port

Ziel: N+1-Latenz bei `network_signal_absent[]`-Read durch einen
Bulk-Read-Port (`ListBoundariesForSessions(ctx, ids)`) eliminiert.

DoD:

- [ ] Neuer Port-Methode in
  `apps/api/hexagon/port/driven/session_repository.go`:
  `ListBoundariesForSessions(ctx, sessionIDs []string)
  (map[string][]Boundary, error)`.
- [ ] SQLite-Adapter implementiert die Methode mit einer einzigen
  Query (`IN`-Clause + sortiertes Result, gruppiert nach
  `session_id`).
- [ ] `SessionsService.ListSessions` nutzt die neue Methode statt
  pro-Eintrag-Aufruf.
- [ ] Performance-Benchmark in `apps/api/.../session*_bench_test.go`:
  1000 Sessions in einer Page brauchen < 200 ms p95.
- [ ] Race-Test (`make api-race`) bleibt grün.
- [ ] Risks-Backlog R-7: Status 🟢.

## 8. Tranche 6 — R-22 Origin-/IP-Rate-Limiter (Driven-Port)

Ziel: optionaler IP-/Origin-Bucket-Limiter als Driven-Port-Adapter
(analog `IssuanceLimiterPort` aus RAK-77, aber pro `client_ip`
oder `Origin`-Header-Hash).

**Backend-Strategie** (Review-Finding 1 / OQ1, 2026-05-11):
**kein** SQLite-Backend für diesen Limiter. SQLite via Shared-
Volume hat dieselben Single-Host-Beschränkungen wie der
`0.12.5`-Issuance-Limiter — Origin-/IP-Limits sind aber typisch
Multi-Host-Konzern (Edge/Reverse-Proxy/LB vor mehreren API-
Replicas). Ein SQLite-Pfad würde False-Negative-Limits erzeugen,
sobald die Replicas auf verschiedenen Hosts laufen. Backend-
Optionen:

- `memory` — In-Process Token-Bucket. Misst pro Replica;
  geeignet für Single-Replica-Lab und als Defense-in-Depth-
  Ergänzung zum Edge-Layer-Limit (CDN/Reverse-Proxy).
- `redis` (Network-Backend) — atomarer Token-Bucket via
  `EVAL`-Script. Vorausgesetzte Topologie ist sowieso ein
  geteilter Redis (z. B. mit `R-17` Tranche 7 gemeinsam genutzt).
  **Empfohlen** für Multi-Host-Setups.
- Kein `sqlite` — würde dem Operator suggerieren, dass IP-Limits
  über Hosts hinweg robust sind, was falsch wäre.

DoD:

- [ ] Neuer Driven-Port
  `apps/api/hexagon/port/driven/origin_rate_limiter.go` mit
  `Allow(ctx, key)`-Methode (`key` = `client_ip` oder
  `Origin`-Hash).
- [ ] `InMemoryOriginRateLimiter`-Adapter analog
  `InMemoryIssuanceRateLimiter` (Token-Bucket-Logik
  wiederverwenden; Bucket-Key-Prefix `origin:` bzw. `ip:`).
  **Kein** SQLite-Adapter — bewusste Plan-Entscheidung (siehe
  Backend-Strategie oben).
- [ ] `RedisOriginRateLimiter`-Adapter (zusammen mit
  `R-17`-Network-Backend in Tranche 7 entwickeln; gleicher
  Redis-Server, anderer Bucket-Key-Prefix). Aktivierung erst
  nach R-17 Tranche 7 (oder gemeinsam, wenn Tranche 6 + 7 als
  Bundle aktiviert werden).
- [ ] ENV-Selektor `MTRACE_ORIGIN_RATE_LIMITER=disabled|memory|redis`
  (Default `disabled` — kein Limiter, Backwards-Compat). Andere
  Werte lehnt der Boot-Validator mit klarem Fehler ab:
  - `sqlite` → „nicht Multi-Host-tauglich, siehe
    Plan-0.12.6 §8 Backend-Strategie".
  - `memcached` → „follow-up item, wird gemeinsam mit R-17
    Tranche 7 geliefert, falls Operator-Bedarf für gemeinsamen
    Memcached-Cluster entsteht" (Backend-Konsistenz mit
    Issuance-Limiter aus §9; einzelne Memcached-Adoption ohne
    R-17-Pendant produziert Backend-Fragmentation).
- [ ] Integration vor `POST /api/auth/session-tokens` und
  `POST /api/playback-events`-Handlern (Reihenfolge: erst Origin-
  Limit, dann Project-Limit).
- [ ] `client_ip`-Quelle dokumentiert: standardmäßig
  `r.RemoteAddr`; bei Reverse-Proxy-Setups muss der Operator den
  `X-Forwarded-For`-Trust-Boundary explizit aktivieren (ENV
  `MTRACE_TRUST_FORWARDED_FOR=1` o. ä.) — sonst trifft der
  Limiter den Proxy, nicht den Client.
- [ ] Tests: In-Memory-Adapter (Single-Process) + Redis-Adapter
  (Cross-Instance-Sharing über `miniredis`-Mock oder echten
  Redis-Container). **Kein** Cross-Instance-Test auf SQLite,
  weil der Adapter es nicht gibt.
- [ ] `make smoke-origin-rate-limit` **neu anlegen** (Script +
  Makefile-Target + Help-Eintrag; Konvention siehe §0.2.1).
  Wrapt entweder den In-Memory-Adapter-Test oder den Redis-
  Mock-Test (`miniredis`).
- [ ] Risks-Backlog R-22: Status 🟢 sobald Redis-Adapter steht;
  bei nur In-Memory bleibt es „teilweise gelöst" mit
  Resttrigger „Multi-Host-IP-Limits benötigt Network-Backend".

## 9. Tranche 7 — R-17 Multi-Host-Issuance-Limiter (Network-Backend)

Ziel: Resttrigger aus `0.12.5` Tranche 2 auflösen — echte
Multi-Host-Topologie ohne Shared-Volume durch Network-Backend-
Adapter.

**Backend-Wahl** (Review-Finding 2 / OQ2, 2026-05-11): **Redis
als Plan-Default für `0.12.6`**, gemeinsam mit R-22 Tranche 6.
Begründung: ein einziger Network-Backend-Typ für beide
Multi-Host-Limiter (Issuance + Origin/IP) hält das Operator-
Setup einheitlich und vermeidet zwei getrennte Cluster.
Memcached bleibt **gemeinsames Folge-Item für beide Tranchen** —
sobald Operator-Bedarf entsteht (z. B. wegen existierender
Memcached-Infrastruktur), wird ein Memcached-Adapter parallel
für R-17 und R-22 nachgezogen, **nicht** für eine Tranche
alleine (sonst entsteht Fragmentation zwischen Issuance-Limiter
und Origin-Limiter).

DoD:

- [ ] Neuer Adapter
  `apps/api/adapters/driven/auth/redis_issuance_rate_limiter.go` —
  implementiert `driven.IssuanceRateLimiter` über atomare
  Token-Bucket-Operationen (Redis `EVAL`-Lua-Script mit
  `INCRBYFLOAT`/`SETEX`-Pattern). Bucket-Key-Prefix
  `mtrace:issuance:`.
- [ ] ENV-Selektor `MTRACE_AUTH_ISSUANCE_LIMITER` um `redis`
  erweitert. Pflicht-ENV `MTRACE_REDIS_ADDR` (und optional
  `MTRACE_REDIS_AUTH`/`MTRACE_REDIS_DB`). Wert `memcached` wird
  vom Boot-Validator als „follow-up item, wird gemeinsam mit
  R-22 Tranche 6 geliefert" abgelehnt.
- [ ] Fail-modus: **fail-closed** als Default (Redis-Outage →
  kein Token issuen, `429 auth_issuance_rate_limited`). ENV
  `MTRACE_AUTH_ISSUANCE_FAIL_OPEN=1` für expliziten Operator-
  Opt-in auf fail-open (lokales `memory`-Fallback).
- [ ] Test gegen `miniredis`-Mock + optional gegen echten Redis-
  Container.
- [ ] `make smoke-issuance-multi-host` **neu anlegen** (Script +
  Makefile-Target + Help-Eintrag; Konvention siehe §0.2.1).
  Wrapt den Cross-Instance-Sharing-Test gegen `miniredis` (oder
  echten Redis-Container) — Pattern analog
  `smoke-issuance-replica` aus `0.12.5` T2, aber mit Redis-
  Network-Backend statt SQLite.
- [ ] Risks-Backlog R-17: Status 🟢 mit „Redis-Backend in
  `0.12.6` Tranche 7"; Memcached-Resttrigger explizit als
  Folge-Item dokumentiert (gemeinsam mit R-22).

## 10. Tranche 8 — R-20 Produktive Vault/KMS-Adapter

Ziel: Resttrigger aus `0.12.5` Tranche 3 auflösen — produktive
Vault-Authentifizierung (AppRole/IAM/K8s-ServiceAccount) und
KMS-Adapter (AWS-KMS minimal).

DoD:

- [ ] Vault-Adapter um AppRole-Auth erweitert
  (`role_id` + `secret_id`-Login-Flow); optional Kubernetes-
  ServiceAccount-Auth-Flow (Token aus
  `/var/run/secrets/kubernetes.io/serviceaccount/token`).
- [ ] Neuer KMS-Adapter (AWS-KMS) als zweite externe Backend-
  Option im `AuthSecretBackend`-Port. ENV-Selektor um `kms`
  erweitert (wird heute vom Boot-Validator abgelehnt).
- [ ] Caching/Refresh-TTL-Konfig: ENV
  `MTRACE_AUTH_SECRET_BACKEND_REFRESH_SECONDS` (Default 0 = keine
  Refresh, Boot-Time-Load wie heute).
- [ ] Compliance-Audit-Vorbereitung: Doku zu PCI-/SOC2-relevanten
  Konfigurationspfaden in
  [`docs/user/auth.md`](../../user/auth.md) §5.5.
- [ ] Tests gegen `httptest`-Mock (AppRole) + Mock-KMS-Provider.
- [ ] `make smoke-vault-approle` und `make smoke-kms-skeleton`
  **beide neu anlegen** (Script + Makefile-Target + Help-Eintrag;
  Konvention siehe §0.2.1). Vault-Smoke wrapt entweder einen
  `vault dev`-Server-Lab-Test oder einen `httptest`-Mock mit
  AppRole-Login-Flow; KMS-Smoke wrapt einen Mock-KMS-Provider-
  Test (kein echter AWS-Lab-Pfad).
- [ ] Risks-Backlog R-20: Status 🟢 sobald produktiver Pfad +
  Compliance-Doku stehen; sonst „teilweise gelöst" mit konkretem
  Resttrigger.

## 11. Tranche 9 — R-15 Externe Media-Server-Provisionierung

Ziel: `POST /api/ingest/streams` (oder ein neuer dedizierter
Endpoint) provisioniert optional gegen einen laufenden MediaMTX/
SRS-Server, statt nur eine Konfig-Datei zu schreiben.

**Backwards-Compat-Vertrag** (Review-Findings 3 + 1 /
OQ1 2026-05-11): Wire-Erweiterung **strikt additiv** mit
deterministischem Verhalten — der `provision`-Query-Param ist
die einzige Schaltvariable für Seiteneffekte:

- Neuer Query-Param `provision`:
  - **Default `false`** — alter Pfad bleibt **byte-stabil** zum
    `0.11.0`-Format. Der Server macht **kein** I/O gegen einen
    externen Media-Server, **unabhängig** von der ENV-
    Konfiguration. `MTRACE_MEDIASERVER_PROVISION_URL` wird in
    diesem Pfad **nicht** gelesen; ein konfigurierter Server
    bleibt ohne `provision=true` unberührt. Im Response-Body
    wird `media_server_state` **nicht** emittiert.
  - **`provision=true`** — opt-in: server-seitiges I/O findet
    statt. `media_server_state` ist im Response-Body **immer**
    gesetzt:
    - ENV konfiguriert + Server erreichbar → `"applied"` oder
      `"partial"` (je nach Adapter-Resultat).
    - ENV konfiguriert + Server unreachable → `"failed"` plus
      `error_code` (kein API-State-Rollback; lokale Konfig +
      Stream sind angelegt, HTTP-Status bleibt `201 Created`).
    - ENV **nicht** konfiguriert → `"disabled"` plus Hinweis-
      Body „set `MTRACE_MEDIASERVER_PROVISION_URL` to enable".
      Auch hier kein I/O-Versuch — der Server kann nicht
      provisionieren, was er nicht kennt.

  Damit ist der Body byte-stabil:
  - `provision=false` ODER fehlt → kein neues Feld, alte
    `0.11.0`-Clients (auch mit `additionalProperties: false`)
    bleiben funktional.
  - `provision=true` → neues Feld immer im Body. Neue Clients,
    die den Param setzen, erwarten das Feld; alte Clients
    setzen den Param nicht und sehen es nie.

- Wire-Versions-Pin: Der Wire-Vertrag in
  [`spec/backend-api-contract.md`](../../../spec/backend-api-contract.md)
  §3.8 bekommt einen `0.12.6`-Block mit dem additiv-Hinweis,
  damit zukünftige Reviewer das Backwards-Compat-Versprechen
  sehen. Folge-Endpoint
  `POST /api/ingest/streams/{id}/provision` für nachträglichen
  Server-Sync bleibt Folge-Item nach `0.12.6`.

DoD:

- [ ] Neuer Driven-Port
  `apps/api/hexagon/port/driven/media_server_provisioner.go` mit
  `Apply(ctx, config)` / `Rollback(ctx, ids)`-API.
- [ ] Adapter-Implementation für MediaMTX (HTTP-API `/v3/config/`-
  Pfade); SRS bleibt Folge-Item nach `0.12.6`.
- [ ] ENV-Konfiguration `MTRACE_MEDIASERVER_PROVISION_URL`/`_TOKEN`;
  fehlt + `provision=true` → `media_server_state="disabled"`
  ohne I/O. ENV wird im `provision=false`-Pfad **nie** gelesen
  (hard-not-read; Operator-Setup darf keinen Seiteneffekt durch
  bloße ENV-Setzung erzeugen).
- [ ] Wire-Update auf `POST /api/ingest/streams`: optionaler
  `provision=true`-Query-Param (Default `false` → 1:1 `0.11.0`-
  Verhalten, **kein** I/O, kein neues Feld). Response-Feld
  `media_server_state` ist **strikt** an `provision=true`
  gekoppelt: immer gesetzt bei `provision=true`, nie sonst
  (siehe Backwards-Compat-Vertrag oben).
- [ ] Wire-Vertrag-Erweiterung in
  [`spec/backend-api-contract.md`](../../../spec/backend-api-contract.md)
  §3.8 mit explizitem `0.12.6`-additive-Hinweis und einem
  Beispiel-Body für jeden der drei Pfade (alter Default,
  Provisionierung erfolgreich, Provisionierung fehlgeschlagen).
- [ ] Contract-Test pinnt: Body byte-stabil zum `0.11.0`-Format
  für `provision=false`-Pfad (alte Fixtures bleiben grün).
- [ ] Operator-Doku in `docs/user/ingest-control.md` mit
  Rollback-Pfad bei API-State-vs-Server-State-Diskrepanz und
  expliziter Backwards-Compat-Notiz für alte Clients.
- [ ] Tests gegen `httptest`-MediaMTX-Mock: happy path (200),
  unreachable (`media_server_state: "failed"`), Auth-Failure,
  partial-state (Stream angelegt, Routing-Rule rejected).
- [ ] `make smoke-mediaserver-provision` **neu anlegen** (Script +
  Makefile-Target + Help-Eintrag; Konvention siehe §0.2.1).
  Wrapt den Adapter-Test gegen `httptest`-MediaMTX-Mock.
- [ ] Risks-Backlog R-15: Status 🟢 sobald Adapter geliefert; sonst
  „teilweise gelöst" mit konkretem Operator-Trigger für SRS.

## 12. Tranche 10 — Release-Closeout

DoD (analog `0.12.5`-Closeout-Pattern):

- [ ] `make docs-check` grün.
- [ ] `make gates` grün (Memory-Pattern: Bump committen, dann
  Gates).
- [ ] `make generated-drift-check` grün (Teil von `make gates`).
- [ ] Falls Minor: Lastenheft-Patch persistiert, RAK-Matrix
  vollständig (Code- + Test-Pfade pro RAK).
- [ ] Wave-2-Quality-Gates dokumentiert (`releasing.md` §3.1):
  benchmark.yml, fuzz.yml, fuzz-Issues, mutation.yml — alle vier
  Indikatoren grün.
- [ ] Versions-Bump auf `0.12.6` an allen Stellen aus
  `releasing.md` §3.1.
- [ ] `CHANGELOG.md` `[0.12.6] - YYYY-MM-DD`-Block:
  `### Added` (aktivierte Adapter, ENV-Vars, Smokes),
  `### Changed` (Wire-Updates),
  `### Security` (R-22-Limiter, R-20-Production-Auth),
  `### Fixed` (R-11-Pagination, R-7-Bulk-Read, ggf. R-5/R-10-
  Marker).
- [ ] Roadmap-Status aktualisiert: §1 Phase auf released, §2
  Schritt 47.7 (oder analog) ✅, §3-Zeile `0.12.6` ✅.
- [ ] Plan nach `docs/planning/done/plan-0.12.6.md` verschoben;
  Status-Header `✅ released YYYY-MM-DD (Tag v0.12.6)`; Tranchen-
  Übersicht alle ✅ oder mit dokumentiertem Defer-Eintrag.
- [ ] Annotierter Tag `v0.12.6` mit Lieferzusammenfassung.
- [ ] GitHub-Release `m-trace 0.12.6` mit Notes-File aus dem
  CHANGELOG-Block.

## 13. Folge-Scope nach `0.12.6`

- [`plan-0.13.0.md`](../open/plan-0.13.0.md): Production / Ops Backends
  (`MVP-40` Postgres, `MVP-41` ClickHouse/VictoriaMetrics,
  `MVP-42` K8s-Manifeste, `MVP-43` Devcontainer, `MVP-44`
  Release-Automatisierung). RAK-Range wird bei dessen T0
  bestimmt — verschiebt sich, falls `0.12.6` als Minor RAK-83+
  belegt.
- **R-9** (Observability-Smoke-K8s-Label-Whitelist) wandert mit
  `0.13.0` Tranche 3 (MVP-42 K8s), siehe Backlog-Eintrag.
- Falls in T0 Tranchen explizit deferred wurden, bleibt der
  R-N-Eintrag in `risks-backlog.md` §1.1 mit geschärftem Trigger
  und Verweis auf `plan-0.12.7.md` o. ä. — der Verweis-Plan
  selbst wird erst angelegt, wenn ein konkreter Operator-Bedarf
  greift.

## 14. Qualitätsregeln für `0.12.6`

- Hexagonale Architektur: jeder neue Adapter ist Driven-Port-
  konform; ENV-Selektion ist die einzige Auswahlsteuerung.
- Backwards-Compat: heutige ENV-Variablen-Werte bleiben
  Default-Pfad; neue Werte sind opt-in.
- Lastenheft als normative Quelle: jede neue Verhaltensaussage
  geht zuerst in den RAK-Block, dann in `docs/user/*.md`/Code
  (Memory-Lehre `feedback_lastenheft_normativ.md`).
- Wave-2-Verdict vor Tag dokumentieren (gepinnt im Plan-DoD wie
  `0.12.5` Tranche 6).
- Memory-Pattern „Closeout-Drift-Gate-Reihenfolge": Versions-
  Bump committen, **danach** `make gates` — `generated-drift-
  check` vergleicht Working-Tree gegen HEAD.
