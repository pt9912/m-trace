# Implementation Plan — `0.12.6` (Folge-Items-Sammlung nach `0.12.5`)

> **Status**: ⬜ open — noch nicht aktiviert. Dieser Plan darf erst
> nach explizitem Move nach `docs/planning/in-progress/` umgesetzt
> werden. Vorgänger ist `0.12.5` (released 2026-05-11; Plan in
> [`done/plan-0.12.5.md`](../done/plan-0.12.5.md)).
>
> **Release-Typ**: **T0-Entscheidung** — Patch (`0.12.6`) vs. Minor
> (`0.13.0` parallel oder vor diesem Plan). Default-Empfehlung: die
> Tranchen ohne neue User-Surface bleiben Patch
> (R-5/R-7/R-10/R-11/R-13); die Tranchen mit neuer Wire-/Adapter-
> Surface (R-15/R-17/R-20/R-22) machen aus `0.12.6` einen Minor mit
> Lastenheft-Patch und neuer RAK-Gruppe. Die Auswahl ist T0-Decision
> — siehe §0.3.
>
> **Ziel**: Die nach `0.12.5` offen gebliebenen R-N-Items aus
> [`risks-backlog.md`](../in-progress/risks-backlog.md) §1.1
> systematisch adressieren — entweder mit Code-Pfad (🟢) oder mit
> explizit dokumentierter Defer-Entscheidung und geschärftem
> Trigger. Plan enthält **eine Tranche pro R-N**; T0 entscheidet,
> welche Tranchen für `0.12.6` aktiviert werden, welche nach
> `0.13.x` verschoben werden und welche im Backlog bleiben.
>
> **Bezug**:
> [`risks-backlog.md`](../in-progress/risks-backlog.md) §1.1
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
> **Nachfolger**: [`plan-0.13.0.md`](./plan-0.13.0.md) (Production /
> Ops Backends, MVP-40..MVP-44).

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
| 0       | —    | Plan-Aktivierung, Release-Typ-Entscheidung, ggf. Lastenheft-Patch + RAK-Matrix-Skelett, Roadmap-Insert | T0          | ⬜      |
| 1       | R-13 | Trivy-Ignore-Re-Review 2026-08-04 (Wartungspflicht) | CI-Wartung  | ⬜      |
| 2       | R-11 | SRT-Health-Detail-Cursor-Pagination             | Adapter-Code | ⬜      |
| 3       | R-5  | Time-Skew-Persistenz + Dashboard-Marker         | Schema + UI | ⬜      |
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

- [ ] Plan von `docs/planning/open/plan-0.12.6.md` nach
  `docs/planning/in-progress/plan-0.12.6.md` verschoben.
- [ ] **Release-Typ fixiert** (Patch vs. Minor) — Begründung im
  T0-Commit-Body.
- [ ] **Tranchen-Auswahl** pro R-N (aktivieren / deferr /
  Backlog-Schärfung) — im T0-Commit-Body und in §1 oben dokumentiert.
- [ ] Falls Minor: Lastenheft-Patch (Version + §-Nummer aktuell
  bestimmt) mit RAK-Range für die aktivierten R-N-Items ergänzt.
- [ ] Falls Minor: RAK-Matrix-Skelett in §2 unten pro RAK.
- [ ] Roadmap-Insert: §1 Phase auf `0.12.6` aktiv; §2 Schritt
  (z. B. 47.7) ergänzt; §3 Release-Übersicht-Zeile `0.12.6`.
- [ ] Vorgänger-Gate verifiziert: `git tag --list v0.12.5`.
- [ ] `make docs-check` grün.

## 3. Tranche 1 — R-13 Trivy-Ignore Re-Review (Wartungspflicht)

Ziel: `.security/vulnignore.yaml`-`expires` 2026-08-04 vor Ablauf
re-reviewen; entweder Verlängerung mit Begründung, Trixie-Point-
Release-Fix oder Base-Image-Wechsel (z. B. Distroless).

DoD:

- [ ] Trivy-Scan gegen `node:22-trixie-slim` (Dashboard + Analyzer-
  Service) mit aktuellem Vulnerability-Stand; Output in
  `.tmp/audit/`.
- [ ] Pro CVE (`CVE-2025-69720` ncurses, `CVE-2026-29111` systemd,
  `CVE-2026-4878` libcap): Status-Eintrag (Trixie-Point-Release
  hat Fix? `expires`-Verlängerung um 90 Tage? Base-Image-Wechsel?).
- [ ] `.security/vulnignore.yaml` aktualisiert (Verlängerung mit
  neuem Datum bzw. Eintrag entfernt).
- [ ] `scripts/render-trivyignore.sh` testet, dass `expires` nicht
  überschritten ist.
- [ ] Optional: ADR-Draft für Distroless-Base-Image-Wechsel vor
  `1.0`, falls Re-Review die strukturelle Lösung empfiehlt.
- [ ] Risks-Backlog R-13: Trigger-Stand-Eintrag aktualisiert mit
  neuem `expires`-Datum oder strukturellem Wechsel; ggf. Status auf
  🟢 wenn Trixie-Fix verfügbar.

## 4. Tranche 2 — R-11 SRT-Health-Detail-Cursor-Pagination

Ziel: `GET /api/srt/health/{stream_id}` liefert Cursor-Pagination
über `samples_limit` hinaus.

DoD:

- [ ] Adapter-Implementation in
  `apps/api/adapters/driven/persistence/sqlite/srt_health_repository.go`:
  Cursor-Token analog `EventRepository`-Pattern
  (`process_instance_id + (ingested_at, id)`-Position als opaker
  Token).
- [ ] HTTP-Handler `GET /api/srt/health/{stream_id}` akzeptiert
  Query-Param `cursor=<opake_token>` und liefert
  `next_cursor`-Feld in der Antwort.
- [ ] Wire-Vertrag-Update in
  [`spec/backend-api-contract.md`](../../../spec/backend-api-contract.md)
  §7a.3 (nur Anwendungs-Notiz, weil das Wire-Format dort bereits
  spezifiziert ist).
- [ ] Unit-Test + Adapter-Test: Pagination liefert konsistente,
  überlappungsfreie Pages über 1500+ Samples.
- [ ] Smoke `make smoke-srt-health-pagination` (oder erweiterte
  Variante des existierenden `smoke-srt-health`).
- [ ] Risks-Backlog R-11: Status 🟢 mit Auflösungspfad „Cursor-
  Pagination in 0.12.6 Tranche 2"; Wieder-Eröffnung bei
  Operator-Report über Inkonsistenz im Cursor-Wandern.

## 5. Tranche 3 — R-5 Time-Skew-Persistenz + Dashboard-Marker

Ziel: `mtrace.time.skew_warning=true`-Events sind im Read-Pfad
(Dashboard ohne Tempo) sichtbar markiert.

DoD:

- [ ] SQLite-Schema-Erweiterung (Migration `V6` o. ä.): Spalte
  `time_skew_warning BOOLEAN NOT NULL DEFAULT 0` an
  `playback_events`.
- [ ] Ingest-Pfad (`POST /api/playback-events`-Handler oder
  Application-Service) setzt das Bit, wenn das eingehende Event
  ein `mtrace.time.skew_warning`-Attribut trägt.
- [ ] Read-Pfad: `ListSessions` und `GetSessionDetail` liefern die
  Spalte mit; SSE-Frames echo'en sie.
- [ ] Dashboard-UI: Indikator-Pin (z. B. Skew-Symbol) auf dem
  betroffenen Event in der Timeline.
- [ ] Doku in [`spec/telemetry-model.md`](../../../spec/telemetry-model.md)
  §2.5/§5.3 aktualisiert: Spalte ist persistent, Read-Pfad-
  Verhalten beschrieben.
- [ ] Tests: Ingest-Roundtrip, List-Read, SSE-Frame, Dashboard-
  Render.
- [ ] Risks-Backlog R-5: Status 🟢 oder Wieder-Eröffnungs-Trigger
  (Operator-Report) dokumentiert.

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
  weglassen (Default `1.0`).
- **Server-Verhalten**: erstmaliger Wert pro `session_id` wird in
  der Session-Metadaten-Zeile persistiert (immutable nach erstem
  gültigen Setzen). Spätere Events mit abweichendem Wert lösen
  einen Warning-Log und einen `mtrace_sample_rate_drift_total`-
  Counter aus — das Risiko wäre, dass das Operator-SDK mitten in
  einer Session den `sampleRate` ändert, was die Lücken-Erkennung
  inkorrekt machen würde.
- **Fallback**: wenn keines der eingegangenen Events das Feld
  setzt, bleibt der Wert auf `1.0` — Session gilt als voll
  gesampelt (Backwards-Compat zum heutigen Verhalten).

DoD:

- [ ] Domain-Erweiterung: `Session` bekommt ein Feld
  `sample_rate float` (Default `1.0`); Persistenz immutable
  nach erstem nicht-Default-Wert.
- [ ] SDK-/Wire-Erweiterung: Player-SDK setzt
  `meta.session_sample_rate` (Pflicht-Feld bei `sampleRate < 1`,
  Default-weglass bei `sampleRate == 1`). Schema-Eintrag in
  `contracts/event-schema.json` mit deutlichem Hinweis auf den
  Session-Scope (kein per-Event-Wert).
- [ ] SQLite-Schema-Migration ergänzt die Session-Spalte
  (`V7` o. ä.).
- [ ] Ingest-Pfad: erster Event mit `meta.session_sample_rate < 1`
  schreibt den Wert in die Session-Zeile (`UPDATE … WHERE
  sample_rate = 1.0` für Immutability); spätere Drift wird
  geloggt und gezählt, aber **nicht** überschreibt.
- [ ] Neuer Counter `mtrace_sample_rate_drift_total{project_id}`
  in `apps/api/adapters/driven/metrics/` mit Cardinality-Limit
  (project_id).
- [ ] Read-Pfad markiert Sessions mit `sample_rate < 1` als
  „sampled" — explizit als API-Feld plus Dashboard-Banner.
- [ ] Sampling-Lücken-Heuristik (optional, T0-Entscheidung):
  Server vergleicht erwartete vs. tatsächliche Event-Anzahl bei
  bekanntem `sample_rate` und markiert auffällige Sessions als
  `"possible_loss"`. Heuristik-Schwellen sind Folge-Tuning.
- [ ] Doku-Update in `spec/telemetry-model.md` §8.3 plus
  `spec/event-schema.md` (oder analog) für das neue Meta-Feld.
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
  Werte (`sqlite`, `memcached`) lehnt der Boot-Validator mit
  klarem Fehler ab — `sqlite` mit expliziter Begründung „nicht
  Multi-Host-tauglich, siehe Plan-0.12.6 §8 Backend-Strategie".
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
- [ ] `make smoke-origin-rate-limit` als Operator-Smoke.
- [ ] Risks-Backlog R-22: Status 🟢 sobald Redis-Adapter steht;
  bei nur In-Memory bleibt es „teilweise gelöst" mit
  Resttrigger „Multi-Host-IP-Limits benötigt Network-Backend".

## 9. Tranche 7 — R-17 Multi-Host-Issuance-Limiter (Network-Backend)

Ziel: Resttrigger aus `0.12.5` Tranche 2 auflösen — echte
Multi-Host-Topologie ohne Shared-Volume durch Network-Backend-
Adapter (Redis/Memcached).

DoD:

- [ ] Neuer Adapter `apps/api/adapters/driven/auth/redis_issuance_rate_limiter.go`
  (oder Memcached, je nach T0-Entscheidung) — implementiert
  `driven.IssuanceRateLimiter` über Network-Calls mit atomarem
  Token-Bucket (Redis `EVAL`-Script oder Memcached-CAS).
- [ ] ENV-Selektor `MTRACE_AUTH_ISSUANCE_LIMITER` um `redis` (bzw.
  `memcached`) erweitert; Pflicht-ENV `MTRACE_REDIS_ADDR`/`_AUTH`
  bzw. `MTRACE_MEMCACHED_ADDR`.
- [ ] Fail-modus dokumentiert: Network-Outage → fail-closed (kein
  Token issuen) oder fail-open (lokales `memory`-Fallback) —
  T0-Entscheidung.
- [ ] Test gegen `httptest`/`miniredis`-Mock.
- [ ] `make smoke-issuance-multi-host`.
- [ ] Risks-Backlog R-17: Status 🟢 mit dem entsprechenden
  Backend-Hinweis.

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
  Konfigurationspfaden in `auth.md` §5.5.
- [ ] Tests gegen `httptest`-Mock (AppRole) + Mock-KMS-Provider.
- [ ] `make smoke-vault-approle` und `make smoke-kms-skeleton`.
- [ ] Risks-Backlog R-20: Status 🟢 sobald produktiver Pfad +
  Compliance-Doku stehen; sonst „teilweise gelöst" mit konkretem
  Resttrigger.

## 11. Tranche 9 — R-15 Externe Media-Server-Provisionierung

Ziel: `POST /api/ingest/streams` (oder ein neuer dedizierter
Endpoint) provisioniert optional gegen einen laufenden MediaMTX/
SRS-Server, statt nur eine Konfig-Datei zu schreiben.

**Backwards-Compat-Vertrag** (Review-Finding 3, 2026-05-11):
Wire-Erweiterung **strikt additiv**:

- Neuer Query-Param `provision`:
  - Default `false` — alter Pfad bleibt 1:1: Endpoint schreibt die
    Konfig-Datei lokal, **kein** I/O gegen externen Server. Alte
    Clients ohne den Param sehen exakt das `0.11.0`-Verhalten.
  - `provision=true` — opt-in: zusätzlich gegen den konfigurierten
    Media-Server provisioniert.
- Neues Response-Feld `media_server_state`:
  - **Optional/additiv** — wird **nur** im Body gesetzt, wenn der
    Request `provision=true` hatte ODER wenn der Server unter
    `MTRACE_MEDIASERVER_PROVISION_URL` konfiguriert ist und ein
    Provisionierungs-Versuch erfolgte. Standardpfad (alter Client,
    `provision=false`, keine ENV-Konfig) liefert das Feld **nicht**
    — der Body bleibt byte-stabil zum `0.11.0`-Format.
  - JSON-Konvention: alte Clients mit strikter Deserialisierung
    (z. B. `additionalProperties: false`) müssen nichts ändern,
    weil der Server das Feld in dem Pfad gar nicht emittiert.
    Neue Clients mit toleranter Deserialisierung sehen es bei
    `provision=true`.
- Provisionierungs-Fehler (`MTRACE_MEDIASERVER_PROVISION_URL`
  unreachable o. ä.) löst **keinen** API-State-Rollback aus:
  `POST /api/ingest/streams` gibt weiterhin `201 Created`
  (lokale Konfig+Stream sind angelegt), `media_server_state`
  enthält dann `"failed"` plus optionalen `error_code`. Operator
  triggert den Server-Sync manuell nach (separater Endpoint
  `POST /api/ingest/streams/{id}/provision`-Folge-Item, falls
  Bedarf besteht).
- Wire-Versions-Pin: Der Wire-Vertrag in
  [`spec/backend-api-contract.md`](../../../spec/backend-api-contract.md)
  §3.8 bekommt einen `0.12.6`-Block mit dem additiv-Hinweis,
  damit zukünftige Reviewer das Backwards-Compat-Versprechen
  sehen.

DoD:

- [ ] Neuer Driven-Port
  `apps/api/hexagon/port/driven/media_server_provisioner.go` mit
  `Apply(ctx, config)` / `Rollback(ctx, ids)`-API.
- [ ] Adapter-Implementation für MediaMTX (HTTP-API `/v3/config/`-
  Pfade); SRS bleibt Folge-Item nach `0.12.6`.
- [ ] ENV-Konfiguration `MTRACE_MEDIASERVER_PROVISION_URL`/`_TOKEN`;
  fehlt → Adapter no-op (heutiges Verhalten).
- [ ] Wire-Update auf `POST /api/ingest/streams`: optionaler
  `provision=true`-Query-Param (Default `false` → 1:1 `0.11.0`-
  Verhalten). Response-Feld `media_server_state` ist **strikt
  additiv**: nur gesetzt bei `provision=true` oder konfiguriertem
  Provisionier-Server (siehe Backwards-Compat-Vertrag oben).
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
- [ ] `make smoke-mediaserver-provision`.
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

- [`plan-0.13.0.md`](./plan-0.13.0.md): Production / Ops Backends
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
