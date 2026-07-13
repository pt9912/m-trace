# Roadmap

> **Stand**: 2026-07-13 — **`0.25.0` Minor released** (Multi-Tenant-Fairness R-26 b + SQLite→Postgres-Cutover R-29; Lastenheft `1.1.26`, RAK-131..135).
>
> **2026-07-13**: (1) **d-migrate `0.9.12` eingearbeitet** (`1f1be65`): das
> Read-only-Enhancement kam upstream breiter als geplant (`--read-only`
> Default für `profile`/`export` **und** die Transfer-Quellseite) — Pin
> gebumpt, `doctor`-Write-Probe durch reine Lese-Probe ersetzt (keine
> Cutover-Phase braucht mehr eine schreibbare Quelle), `make smoke-cutover`
> auf **10 Cases** erweitert (read-only-Quelle), `make gates` grün.
> (2) **R-26 b GELIEFERT — plan-0.25.0 T1–T3 gebaut + Nachweis erbracht,
> R-26 → 🟢**:
> [`done/plan-0.25.0-shared-ingest-limiter.md`](../done/plan-0.25.0-shared-ingest-limiter.md)
> — am selben Tag geschnitten, owner-reviewt (zwei substanzielle Befunde
> eingearbeitet: Skew-Refill-Inflation → monotones `last_at` + Skew-Test;
> Fail-Mode-Abwägung inkl. Gegen-Präzedenz), **gefirmt** (§8:
> fail-open-to-memory + `MTRACE_RATE_LIMIT_FAIL_CLOSED`,
> `MTRACE_RATE_LIMIT_BACKEND`, uniforme Caps, eigener Minor `0.25.0` inkl.
> wartendem Cutover, kein ADR) und gebaut: T1 `RedisTokenBucketRateLimiter`
> (Lua n-Token/3-Dimensionen/all-or-nothing, miniredis-Suite inkl.
> Skew-Test), T2 Multi-Tenant-Lab (`MTRACE_LAB_PROJECTS`, k6-Fan-out,
> `make smoke-load-multi-tenant`; **Befund gefixt**: Ingest-`client_ip`
> folgt jetzt der XFF-Trust-Boundary — hinter LB/Proxy sonst globale
> statt per-Client-Drossel), T3 Scale-out-Nachweis
> (`make smoke-scaleout-fairness`): **Fairness-Inversion throttled 1→2
> Replicas 0,96× statt 2,01×** (budgets.md §9), Noisy-Neighbor über den
> LB isoliert, Korrektheits-Gates unverändert. **Offen: T4 = Release
> `0.25.0`** (nimmt den Cutover mit; Lastenheft-Patch mit neuer
> RAK-Gruppe) — bis dahin CHANGELOG `[Unreleased]`.
>
> **Feierabend 2026-07-12**: **SQLite→Postgres-Cutover (plan-0.24.0)
> implementiert + auf `origin/main`** — 4 Phasen
> (`doctor`/`profile`/`bulk`/`incremental`/`switch`), **jede code-reviewt**,
> `make smoke-cutover` (8 Cases) grün, Operator-Runbook
> [`../../ops/postgres-cutover.md`](../../ops/postgres-cutover.md), `R-29` 🟢,
> ADR-0007 „geliefert". **Kein eigenes 0.24.0-Tag** (reine Ops-Tooling ohne
> Runtime-/Package-Diff) → rollt in den nächsten Release (CHANGELOG
> `[Unreleased]`). Außerdem heute: Load-Smoke-CI-Fix (Exit-Code-Contract +
> `R-24`-Debounce, `456fc3a`); d-migrate `0.9.11` gepinnt (fixte `data profile`);
> Read-only-Profile-Enhancement für d-migrate geschrieben (Scratchpad, ephemer)
> → wenn eingearbeitet, `doctor`-RW-Probe für den `profile`-Pfad lockerbar.
> **Offen (nächste größere Tranche): `R-26 b`** (repliken-übergreifend fairer
> Ingest-Limiter, shared Redis — seit 2026-07-13 skizziert, s. oben).
>
> **Phase**: ✅ `0.23.0` Postgres Scale-out (Minor) **released** (Tag
> `v0.23.0`, GHCR- + npm-Publish grün, 2026-07-11), Lastenheft-Patch
> `1.1.25` (RAK-126..RAK-130, §13.24). Liefert den **optionalen**
> Postgres-Runtime-Adapter (`MTRACE_PERSISTENCE=postgres`), den
> DB-autoritativen Ingest-Sequencer (R-28), die Multi-Replica-Harness mit
> über `pg_advisory_lock` serialisierter Startup-Migration und den
> Scale-out-Lasttest, der `R-26 c` belegt (0 Verlust / 0 Duplikate über 2
> Replicas @ ~1,4 Mio Events; Durchsatz store-gebunden = Single-Postgres-
> Decke, ehrlich attribuiert). `SQLite` bleibt Default. `R-26 b`
> (repliken-übergreifende Multi-Tenant-Fairness) bleibt offen.
> Vorher: ✅ `0.22.4` Security-/Tooling-Patch released (Tag `v0.22.4`,
> GHCR- + npm-Publish grün, 2026-06-23).
> Vorheriger Stand: ✅ `0.22.3` Security-/CI-Sammel-Patch released
> ([`done/plan-0.22.3-webrtc-drift.md`](../done/plan-0.22.3-webrtc-drift.md)).
>
> **Aktuell / letzte Releases:**
> - `v0.22.4` Security-/Tooling-Patch (released 2026-06-23, kein
>   Lastenheft-Patch): `golang.org/x/net 0.53.0 → 0.56.0`
>   (sechs HIGH-CVEs im api-gobinary, Trivy-Image-Scan, Issue #9),
>   `undici`-`pnpm.overrides` `^7.28.0` (GHSA-vmh5-mc38-953g), Trivy-Pin
>   `0.71.0 → 0.71.2` und der ENV-konfigurierbare Ingest-Rate-Limiter
>   (Default 100/100 unverändert). `make vuln-check` + `make image-scan`
>   + `make gates` lokal grün.
> - `v0.22.3` Security-/CI-Sammel-Patch (kein Lastenheft-Patch):
>   vier Nightly-Security-Treffer behoben — Trivy `0.59.1 → 0.71.0`
>   + `CVE-2026-45447`/`-48959`-Ignores, esbuild `^0.28.1`-Override,
>   libsqlite3-FTS5-Ignores (`CVE-2026-11822`/`-11824`, kein
>   Upstream-Fix), vite `^8.0.16` (GHSA-fx2h-pf6j-xcff). Plus
>   GitHub-Actions-Node-24-Runtime, WebRTC-Drift-Test-Fix und ein
>   Flaky-CORS-Preflight-Test (geteilter `http.DefaultClient`). Plan
>   archiviert in
>   [`done/plan-0.22.3-webrtc-drift.md`](../done/plan-0.22.3-webrtc-drift.md).
> - `v0.22.2` Go-Stdlib-Security-Patch + perl-base-Trivy-Ignores
>   (Patch-Release, kein Lastenheft-Patch): Build-Image
>   `golang:1.26.3 → 1.26.4` (sechs Stellen) schließt GO-2026-5037
>   + GO-2026-5039; `.security/vulnignore.yaml` mit fünf neuen
>   `perl-base`-Einträgen für dashboard/analyzer-service (kein
>   exploitable Pfad); bilingualer README-Split. Plan archiviert in
>   [`done/plan-0.22.2.md`](../done/plan-0.22.2.md).
> - `v0.22.1` devalue-Security-Patch + Nightly-Audit-Mirror
>   (Patch-Release, kein Lastenheft-Patch): `pnpm.overrides` hebt
>   `devalue` auf `^5.8.1`; neuer `security-audit.yml`-Nightly mit
>   konsolidiertem Auto-Issue; Benchmark-Workflow-Pfadfix +
>   Issue-Body-Auslagerung in `scripts/`. Plan archiviert in
>   [`done/plan-0.22.1.md`](../done/plan-0.22.1.md).
> - `v0.22.0` Quality-Gates Follow-up (released, kein Lastenheft-Patch):
>   Benchmark-Smoke in `make gates`, harter Benchmark-Nightly,
>   Mutation-TS-Scope auf `@pt9912/player-sdk` korrigiert.
>   Plan archiviert in [`done/plan-0.22.0.md`](../done/plan-0.22.0.md).
> - `v0.21.0` OCI-Image-Publishing-Minor (Lastenheft `1.1.24`,
>   RAK-121..RAK-125 in §13.23): GHCR-Namensschema, Make-Targets,
>   Publish-Workflow, Release-Doku und erster GHCR-Publish für drei
>   Runtime-Images. Plan archiviert in
>   [`done/plan-0.21.0.md`](../done/plan-0.21.0.md).
> - `v0.20.0` Package-Publishing-Minor (Lastenheft `1.1.23`,
>   RAK-116..RAK-120 in §13.22): erster GitHub-Packages-Publish für
>   `@pt9912/player-sdk` und `@pt9912/stream-analyzer`, Scope-
>   Migration von `@npm9912` auf `@pt9912`, Publish-Workflow und
>   Release-Doku. Plan archiviert in
>   [`done/plan-0.20.0.md`](../done/plan-0.20.0.md).
> - `v0.18.0` Offene-Risiken-/Trigger-Re-Eval-Minor (kein
>   Lastenheft-Patch; normativer Stand bleibt `1.1.22`): `R-9`,
>   `R-12` und `R-13` bleiben offen, aber mit praezisierten
>   Folge-Triggern in Plan, Roadmap und Risks-Backlog. Keine Runtime-,
>   Wire-, Persistenz-, Public-API-, Schema- oder Default-Aenderung
>   ueber den Release-Bump und versionstragende Test-/Fixture-Assets
>   hinaus. Plan archiviert in
>   [`done/plan-0.18.0.md`](../done/plan-0.18.0.md).
> - `v0.17.0` Hardening-/Evidence-Review-Minor (Lastenheft `1.1.22`,
>   RAK-111..RAK-115 in §13.21): Szenario D Hardening-only,
>   Evidence Review, Doku-/Defer-Entscheid und No-change-Gate-Nachweis;
>   keine Productization, kein Next Slice, kein Switch und keine Runtime-/
>   Public-API-/Schema-Aenderung ueber den versionstragenden Test-/
>   Fixture-Asset-Bump hinaus. Plan archiviert in
>   [`done/plan-0.17.0.md`](../done/plan-0.17.0.md).
> - `v0.16.0` Selected-Product-Slice-/Analyzer-Range-Fetch-Minor
>   (Lastenheft `1.1.21`, RAK-106..RAK-110 in §13.20):
>   HLS-Range-Fetch fuer explizite Byte-Range-Offsets, Gate-Closeout
>   und Tag `v0.16.0`. Plan archiviert in
>   [`done/plan-0.16.0.md`](../done/plan-0.16.0.md).
> - `v0.15.0` Product-Scope-/Analyzer-Boundary-Minor (Lastenheft
>   `1.1.20`, RAK-101..RAK-105 in §13.19): Zielgruppe geschärft,
>   externe Analyzer-API deferred, Control-Plane deferred,
>   HTTP-Range-/Byte-Range-Loader als bevorzugter `NF-13`-Folgeslice,
>   Postgres/Analytics weiter triggerbasiert deferred. Plan archiviert
>   in [`done/plan-0.15.0.md`](../done/plan-0.15.0.md).
> - `v0.14.0` Ops-Backend-Follow-up-Minor (Lastenheft `1.1.19`,
>   RAK-96..RAK-100 in §13.18): Postgres und Analytics bleiben
>   triggerbasiert deferred, K8s-/Devcontainer-Seeds sind clusterfrei
>   validiert, Release-Guard-Fehlerpfade getestet. Plan archiviert in
>   [`done/plan-0.14.0.md`](../done/plan-0.14.0.md).
> - `v0.13.0` Production-/Ops-Backends Decision-and-Seed-Minor
>   (Lastenheft `1.1.18`, RAK-91..RAK-95 in §13.17):
>   ADR 0005 entscheidet Postgres/Analytics als deferred mit
>   messbaren Triggern, optionale K8s-Beispiele unter `deploy/k8s/`,
>   Devcontainer-Seed und Release-Guard. Plan archiviert in
>   [`done/plan-0.13.0.md`](../done/plan-0.13.0.md).
> - `v0.12.6` Auth-/Ingest-Folge-Items-Minor (Lastenheft `1.1.17`, RAK-83..RAK-90 in §13.16); Time-Skew-Persistenz (R-5, V6-Migration), `ListSessions`-Bulk-Read-Port (R-7), Sample-Rate-PPM (R-10, V7-Migration), SRT-Health-Cursor-Pagination v3 (R-11), Trivy-Re-Review (R-13, Expiry 2026-11-02), mediamtx-Provisioner (R-15, additives `?provision=`), Redis-Multi-Host-Issuance-Limiter (R-17 final), Vault-AppRole + KMS-Skeleton (R-20 final), Origin-/IP-Rate-Limiter (R-22); neue Smokes `smoke-srt-health-pagination`, `smoke-origin-rate-limit`, `smoke-issuance-multi-host`, `smoke-vault-approle`, `smoke-kms-skeleton`, `smoke-mediamtx-provision`. Plan archiviert in [`done/plan-0.12.6.md`](../done/plan-0.12.6.md).
> - `v0.12.5` Auth-/Ingest-Adapter-Minor (Lastenheft `1.1.16`, RAK-77..RAK-82 in §13.15); `MultiKeySigningResolver`-Code-Pfad (R-18), `SqliteIssuanceRateLimiter` mit Migration V5 (R-17 teilweise), `AuthSecretBackend`-Port + Vault-Skelett (R-20 teilweise), `BrowserIngestPolicy` mit Origin-Pin/CSRF (R-21), `MediaMTXAuthHookHandler` als `externalAuth`-Bridge (R-14), `OutboundWebhookDispatcher` mit HMAC + Retry (R-16); fünf neue opt-in Smokes (`smoke-key-rotation`, `-issuance-replica`, `-browser-ingest`, `-mediamtx-auth`, `-outbound-webhook`). Plan archiviert in [`done/plan-0.12.5.md`](../done/plan-0.12.5.md).
> - `v0.12.1` Trigger-Re-Eval + Operator-Doku (Patch nach `0.12.0`, kein Lastenheft-Patch); Trigger-Stand pro aktivem R-N-Item, Multi-Key-Signing-Rotation-Operator-Runbook in `auth.md` §5.3.1, OS-1..OS-5 als ⬛ Duplikate in §1.2 abgelegt, OS-6 zu R-22 konvertiert; Plan in [`done/plan-0.12.1.md`](../done/plan-0.12.1.md).
> - `v0.12.0` Auth / Token Lifecycle (F-111..F-113, RAK-71..RAK-76 in §13.14, Lastenheft `1.1.15`); kurzlebige Session Tokens, rotierbare Project-Token-Generationen, tenant-spezifische Ingest Policies; Plan in [`done/plan-0.12.0.md`](../done/plan-0.12.0.md).
> - `v0.11.0` Ingest-Gateway / Stream Control (F-46..F-51, MVP-38, RAK-65..RAK-70 in §13.13, Lastenheft `1.1.14`); lokaler/lab-naher Stream-Control-Pfad, CSPRNG-Stream-Keys, MediaMTX-Konfig-Generator, Lifecycle-Hooks; Plan in [`done/plan-0.11.0.md`](../done/plan-0.11.0.md).
> - `v0.10.0` CMAF-Analyse (Lastenheft `1.1.13`); Plan in [`done/plan-0.10.0.md`](../done/plan-0.10.0.md).
> - `v0.9.6` Lastenheft-Konvergenz; Plan in [`done/plan-0.9.6.md`](../done/plan-0.9.6.md).
> - `v0.9.5` Quality-Gates Wave 2 · `v0.9.1` Drift-Smoke-Robustheit · `v0.9.0` Drift-Smoke + SRS-Lab + DASH-Analyse (Lastenheft-Patch `1.1.11` §13.11); Plan in [`done/plan-0.9.0.md`](../done/plan-0.9.0.md).
> - Frühere Tags: `v0.8.5` (`ce05e3b`, Quality-Gates Wave 1), `v0.8.0` (`8df263a`, Player-SDK-WebRTC-Adapter), `v0.7.0` (`11a3368`), `v0.6.0` (`d08a89f`), `v0.5.0` (`a56dc0b`).
> **Bezug**: `spec/lastenheft.md` RAK-1..RAK-46 (Release-Plan, normativ),
> `spec/architecture.md` (Zielbild),
> Plan-Dokumente pro Release in `docs/planning/plan-X.Y.Z.md`,
> ADRs in `docs/adr/`.

Dieses Dokument ist die **Statusseite** des Projekts. Es duplikiert nicht
die Anforderungen pro Release (die stehen normativ im Release-Plan des
Lastenheft), sondern verfolgt: *Wo sind wir, was kommt als nächstes,
welche Risiken und Folge-Entscheidungen liegen vor uns.*

Wartungsregel: nach jedem Release-Bump und nach jedem Folge-ADR
aktualisieren.

---

## 1. Aktueller Stand (2026-07-13 — `0.25.0` Minor released)

### 1.1 Lieferstand

| Status | Bereich                             | Ergebnis                                                                                                                     | Verweise                                                               |
| ------ | ----------------------------------- | ---------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------- |
| ✅      | Lastenheft                          | `v0.7.0` mit verbindlichem Release-Plan; aktuell `1.1.26` (RAK-1..RAK-135, §13.25 Multi-Tenant-Fairness + SQLite→Postgres-Cutover für `0.25.0`; Patch aktiv).                       | `spec/lastenheft.md`                                                   |
| ✅      | Architektur + ADRs                  | `0001` Backend-Stack (Go) Accepted; `0002` Persistenz Accepted: SQLite als lokaler Durable-Store; `0005` Production-/Ops-Backends Accepted: Postgres/Analytics deferred mit Triggern, K8s/Devcontainer/Release-Guard als Seeds; `0006` Accepted, **R-26 c belegt (2026-07-11)**: reaktiviert den Postgres-Teil von `0005` (RAK-91 → proceed/optional, SQLite bleibt Default); `plan-0.23.0-postgres-scaleout` komplett (6 PG-Ports, DB-autoritativer Sequencer R-28, R-27-Read-Wasserzeichen, Multi-Replica-Harness mit `pg_advisory_lock`-Migration, Scale-out-Lasttest). Nachweis: 0 Verlust/0 Duplikate über 2 Replicas @ ~1,4 Mio Events; Durchsatz-Skalierung flaschenhals-abhängig (app-gebunden linear, store-gebunden = Single-PG-Decke) — s. `budgets.md` §8. `0007` Accepted + **geliefert (`0.25.0`)**: SQLite→Postgres-Cutover als opt-in Ops-Werkzeug (R-29 🟢). **R-26 vollständig belegt** (a Single-Instance-Headroom, b Multi-Tenant-Fairness via shared Redis-Ingest-Limiter — Inversion 0,96–1,0× statt 2,01×, s. `budgets.md` §9 —, c Scale-out-Korrektheit).     | `docs/adr/0001-backend-stack.md`, `docs/adr/0002-persistence-store.md`, `docs/adr/0005-production-ops-backends.md`, `docs/adr/0006-postgres-scaleout-adapter.md`, `docs/adr/0007-sqlite-postgres-data-cutover.md` |
| ✅      | Backend Core (`0.1.0`)              | API-Skelett, Compose-Lab, RAK-1/3/4/6/8.                                                                                     | [`plan-0.1.0.md`](../done/plan-0.1.0.md)                               |
| ✅      | Player-SDK + Dashboard (`0.1.1`)    | Dashboard, Demo-Player, hls.js-Adapter, Session-Ansicht.                                                                     | [`plan-0.1.1.md`](../done/plan-0.1.1.md)                               |
| ✅      | Observability (`0.1.2`)             | Prometheus + Grafana + OTel-Collector als Profil; RAK-9, RAK-10.                                                             | [`plan-0.1.2.md`](../done/plan-0.1.2.md)                               |
| ✅      | Publizierbares Player-SDK (`0.2.0`) | `@pt9912/player-sdk` mit ESM/CJS/IIFE, Pack-Smokes, Browser-Support-Matrix; RAK-11..RAK-21.                                 | [`plan-0.2.0.md`](../done/plan-0.2.0.md)                               |
| ✅      | Stream-Analyzer (`0.3.0`)           | `@pt9912/stream-analyzer` (Library + CLI), `analyzer-service` (interner HTTP-Wrapper), `POST /api/analyze`; RAK-22..RAK-28. | [`plan-0.3.0.md`](../done/plan-0.3.0.md)                               |
| ✅      | Erweiterte Trace-Korrelation (`0.4.0`) | SQLite-Persistenz, `correlation_id`/`trace_id`-Trennung, Dashboard-Session-Timeline (SSE + Polling-Fallback), optionales Tempo-Profil, Aggregat-Metriken-Sichtbarkeit, Cardinality-/Sampling-Doku; RAK-29..RAK-35 erfüllt. | [`plan-0.4.0.md`](../done/plan-0.4.0.md)                            |
| ✅      | Multi-Protocol Lab (`0.5.0`)        | `examples/`-Konventions-Index plus MediaMTX-/SRT-/DASH-Beispiele und WebRTC-Vorbereitungspfad; opt-in Smokes `make smoke-mediamtx`/`smoke-srt`/`smoke-dash`. RAK-36..RAK-40 erfüllt. | [`plan-0.5.0.md`](../done/plan-0.5.0.md)                            |
| ✅      | SRT Health View (`0.6.0`)           | MediaMTX-API als CGO-freie SRT-Quelle (R-2 aufgelöst), durabler Health-Store, Read-API + Dashboard-Route, Operator-Doku. RAK-41..RAK-46 erfüllt; opt-in Smoke `make smoke-srt-health`. | [`plan-0.6.0.md`](../done/plan-0.6.0.md)                            |
| ✅      | WebRTC-Lab-Erweiterung (`0.7.0`)    | Lab-Compose `examples/webrtc/` (Project `mtrace-webrtc`) mit MediaMTX-WHIP/-WHEP und FFmpeg-RTSP-Publisher; opt-in Smoke `make smoke-webrtc-prep` (endpoint-only); WebRTC-Telemetrie-Vorbereitung in `spec/telemetry-model.md` §3.5; R-12 als Schema-Drift-Review-Gate. RAK-47..RAK-50 erfüllt; RAK-51 deferred. | [`plan-0.7.0.md`](../done/plan-0.7.0.md)                            |
| ✅      | Player-SDK-WebRTC-Adapter (`0.8.0`) | Produktiver `attachWebRtc`-Adapter in `@pt9912/player-sdk` (additiv zu `attachHlsJs`); reservierter `webrtc.*`-Meta-Namespace mit harter API-Validation; sechs `mtrace_webrtc_*`-Counter mit Delta-Semantik (Server-side Sample-State, Sample-ID-Idempotenz); `scripts/smoke-observability.sh` spiegelt §3.1-Forbidden und §3.2-Allowlist; R-12 release-blockierend ab nächstem Browser-Major-Bump. Browser-Support-Matrix Chromium 120+/Firefox 120+ Required, Safari 17+ Best-effort. RAK-51..RAK-55 erfüllt. | [`plan-0.8.0.md`](../done/plan-0.8.0.md)                            |
| ✅      | Quality-Gates Wave 1 (`0.8.5`)      | Erstmaliger Patch-Release im Repo: Security-Gates (`vuln-check`/`audit-ts`/`image-scan`/`security-gates`) als zweiter PR-blockierender CI-Job parallel zu `build`; Generated-Artifact-Drift-Gate (`make generated-drift-check`) als Bestandteil von `make gates`; Migrations-Konsolidierung als rolling V1; Image-Hardening auf `node:22-trixie-slim` mit `pnpm deploy --prod`-Snip; OpenTelemetry-Stack-Bump als `GO-2026-4394`-Fix; Patch-Release-Konvention in `docs/user/releasing.md` §3.1 verankert. Keine User-Surface-Änderung. | [`plan-0.8.5.md`](../done/plan-0.8.5.md)                            |
| ✅      | Drift-Smoke + SRS + DASH (`0.9.0`)  | Browser-`getStats()`-Drift-Smoke mit Nightly-Workflow `webrtc-drift.yml` (R-12 von release-blockierend auf automatisiert detektiert); SRS-Lab `examples/srs/` als fünftes Multi-Protocol-Beispiel (MVP-36 eingelöst); DASH-Manifest-Analyse im `@pt9912/stream-analyzer` mit `analyzerKind:"dash"`/`playlistType:"dash"`, Detector + regex-basierter MPD-Parser, `manifest_not_supported` als additiver Public-Code, CLI-Dispatch (NF-12 erfüllt; MVP-37 hochgestuft auf Muss). Lastenheft-Patch `1.1.11` aktiv. RAK-56..RAK-59 erfüllt. | [`plan-0.9.0.md`](../done/plan-0.9.0.md)                            |
| ✅      | Quality-Gates Wave 2 (`0.9.5`)      | Patch-Release ohne User-Surface. Benchmark-Smoke (Go + TS) mit Single-Source-Budgets in `docs/perf/budgets.md` und Beobachtungs-Nightly `benchmark-observation.yml` (Cron 02:30); Nightly-`benchstat`-Regressionen `benchmark.yml` (Cron 04:00) gegen orphan-Branch `benchmark-baseline`, Schwelle +15 % auf p<0.05, Auto-Issue plus Quarantäne-Tag-Mechanik (max. 30 Tage); selektives Fuzzing mit sechs Go-Fuzz-Targets und drei TS-Property-Test-Suites via `fast-check@4.4.0` plus Nightly `fuzz.yml` (Cron 05:00) — Erstfund über `FuzzMapMediaMtxItem` (`mbpsLinkCapacity=-1` leakte als negativer `AvailableBandwidthBPS`, Fix in `apps/api/.../mediamtxclient/mapping.go`); Mutation-Testing mit gremlins (Go) + StrykerJS (TS) als nicht-blockierender Nightly-Report `mutation.yml` (Cron 06:00). Operator-Doku in `docs/dev/fuzzing.md` und `docs/dev/mutation-testing.md`. Kein Lastenheft-Patch (Quality-Gates, keine User-Surface). | [`plan-0.9.5.md`](../done/plan-0.9.5.md)                            |
| ✅      | Lastenheft-Konvergenz (`0.9.6`)     | Patch-Release; fehlende Muss-Repo-Artefakte (`CONTRIBUTING.md`, `SECURITY.md`, `.env.example`, `deploy/`-Struktur), Lastenheft-Patch `1.1.12` (F-7-Status, neue Pflichtdokumente-Kennung `F-131`, NF-13/NF-18 harmonisieren, MVP-19..MVP-26 redaktionell entzerren), Go-Stdlib-Bump `golang:1.26.3` (GO-2026-4982/4980/4971/4918). Keine User-Surface-Änderung. | [`plan-0.9.6.md`](../done/plan-0.9.6.md) |
| ✅      | CMAF-Analyse (`0.10.0`)             | Minor-Release. NF-13-Vollumsetzung im Stream-Analyzer-Scope: manifestbasierte HLS-/DASH-CMAF-Signale (additives `details.cmaf` ohne neuen `analyzerKind`) plus begrenzte binäre CMAF-Konformitätsprüfung (ISO-BMFF-Box-Parser, bounded Segment-Loader; Brand-Allowlist `cmfc`/`cmf2`/`cmfs`/`cmff`; Defaults `maxSegmentBytes=2_000_000`/`maxBinarySegments=6`). Lastenheft-Patch `1.1.13` mit RAK-60..RAK-64 in §13.12. | [`plan-0.10.0.md`](../done/plan-0.10.0.md) |
| ✅      | Ingest-Gateway / Stream Control (`0.11.0`) | Minor-Release. F-46..F-51 + MVP-38 als lokaler/lab-naher Stream-Control-Pfad: CSPRNG-Stream-Keys (nur `key_hash` persistiert), `srt`/`rtmp`-Endpunkte, 1:1-Routing, deterministischer MediaMTX-Konfig-Generator, Lifecycle-Hooks `POST /api/ingest/hooks/stream-{started,ended}` mit Source-Allowlist, `make smoke-ingest-control`. Variante B (Modul in `apps/api`). Lastenheft-Patch `1.1.14` mit RAK-65..RAK-70 in §13.13. | [`plan-0.11.0.md`](../done/plan-0.11.0.md) |
| ✅      | Auth / Token Lifecycle (`0.12.0`)   | Minor-Release. F-111..F-113 als zusammenhängender Auth-/Security-Scope: kurzlebige HMAC-SHA-256-signierte Session Tokens (`POST /api/auth/session-tokens`, Konsum via `Authorization: Bearer mtr_st_*` / `X-MTrace-Session-Token`), rotierbare `mtr_pt_*`-Project-Token-Generationen (V4-SQLite-Migration, `grace_until`), Project-gebundene Ingest Policies + §3.9-konformer CORS-Preflight (`204` minimal). Lastenheft-Patch `1.1.15` mit RAK-71..RAK-76 in §13.14 + neunstufige Auth-Fehlerpräzedenz und zehn `auth_*`-Codes. RAK-74-Scope-Cut: `/api/ingest/*` bleibt `0.11.0`-Token-only. | [`plan-0.12.0.md`](../done/plan-0.12.0.md) |
| ✅      | Trigger-Re-Eval + Operator-Doku (`0.12.1`) | Patch-Release nach `0.12.0`, kein Lastenheft-Patch. Trigger-Stand-Notizen pro aktivem R-N (R-5/R-7/R-9/R-10/R-11/R-12/R-13/R-14/R-15/R-16/R-17/R-18/R-20/R-21, alle „nicht ausgelöst" zum 2026-05-10), Multi-Key-Signing-Rotation-Operator-Runbook in `docs/user/auth.md` §5.3.1 (Soll-Workflow; Code-Pfad in `0.12.5`), OS-1..OS-5 als ⬛ Duplikate zu R-14/R-17/R-18/R-20 in `risks-backlog.md` §1.2 abgelegt, OS-6 zu **R-22** in §1.1 konvertiert (Origin-/IP-naher Rate-Limiter, Auflösungspfad `plan-0.13.x`); R-19 als ⬛ historischer Marker. „Teilweise gelöst"-Konvention im Backlog §2 Wartung gepinnt. | [`plan-0.12.1.md`](../done/plan-0.12.1.md) |
| ✅      | Auth-/Ingest-Adapter-Minor (`0.12.5`) | Minor-Release am 2026-05-11. Lastenheft-Patch `1.1.16` mit RAK-77..RAK-82 in §13.15. Sechs Code-Pfade ausgeliefert: `MultiKeySigningResolver` + ENV-Parser (R-18, RAK-78), `SqliteIssuanceRateLimiter` mit Migration V5 + ENV-Selektor (R-17 teilweise, RAK-77), `AuthSecretBackend`-Driven-Port + ENV/Vault-Adapter-Skelett (R-20 teilweise, RAK-79), `BrowserIngestPolicy` mit Preflight-Handler + POST-Enforcement-Middleware (R-21, RAK-80), `MediaMTXAuthHookHandler` als `externalAuth`-Bridge (R-14, RAK-81), `OutboundWebhookDispatcher` mit HMAC-SHA-256-Signatur + Exponential-Backoff-Retry (R-16, RAK-82). Fünf neue opt-in Smokes (`smoke-key-rotation`/`-issuance-replica`/`-browser-ingest`/`-mediamtx-auth`/`-outbound-webhook`). R-18/R-21/R-14/R-16 in §1.2 nach 🟢 verschoben; R-17/R-20 bleiben in §1.1 mit „teilweise gelöst"-Markierung. | [`plan-0.12.5.md`](../done/plan-0.12.5.md) |
| ✅      | Auth-/Ingest-Folge-Items-Minor (`0.12.6`) | Minor-Release am 2026-05-12. Lastenheft-Patch `1.1.17` mit RAK-83..RAK-90 in §13.16. Neun R-N-Items adressiert: Time-Skew-Persistenz (R-5 🟢, RAK-83, V6-Migration + `event.time_skew_warning`-Wire + Dashboard-Pin), `ListSessions`-Bulk-Read-Port (R-7 🟢, RAK-84, `BoundaryStore.ListBoundariesForSessions`), Sample-Rate-PPM (R-10 🟢 minus Heuristik, RAK-85, V7-Migration + `session.sample_rate_ppm` + Dashboard-Banner), SRT-Health-Cursor-Pagination v3 (R-11 🟢, RAK-86, `samples_cursor`/`next_samples_cursor`), Trivy-Re-Review (R-13 🟢-Wartung, Expiry 2026-11-02 für CVE-2025-69720/CVE-2026-29111/CVE-2026-4878), mediamtx-Provisioner (R-15 🟢, RAK-87, additives `?provision=mediamtx`), Redis-Multi-Host-Issuance-Limiter (R-17 🟢 final, RAK-88), Vault-AppRole + KMS-Skeleton (R-20 🟢 final, RAK-89), Origin-/IP-Rate-Limiter (R-22 🟢, RAK-90). Sechs neue opt-in Smokes. R-17/R-20-Resttrigger aus `0.12.5` geschlossen. | [`plan-0.12.6.md`](../done/plan-0.12.6.md) |
| ✅      | Production / Ops Backends (`0.13.0`) | Minor-Release am 2026-05-12. Lastenheft-Patch `1.1.18` mit RAK-91..RAK-95 in §13.17. Decision-and-Seed-Scope: ADR 0005 deferred Postgres und Analytics-Backends mit messbaren Triggern; optionale Kubernetes-Beispiele unter `deploy/k8s/`; Devcontainer-Seed; Release-Guard mit manueller Freigabe. | [`plan-0.13.0.md`](../done/plan-0.13.0.md) |
| ✅      | Ops Backend Follow-up (`0.14.0`) | Minor-Release am 2026-05-12. Lastenheft-Patch `1.1.19` mit RAK-96..RAK-100 in §13.18. Szenario C: K8s-/Devcontainer-/Release-Guard-Hardening; Postgres und Analytics bleiben Trigger-/Defer-Pfade ohne neue Pflichtabhängigkeit. | [`plan-0.14.0.md`](../done/plan-0.14.0.md) |
| ✅      | Product Scope / Analyzer Boundary (`0.15.0`) | Released 2026-05-12. Lastenheft-Patch `1.1.20` mit RAK-101..RAK-105 in §13.19. Szenario A: Zielgruppe + Analyzer-Boundary; Tranche 1 entscheidet Selbsthoster/kleine bis mittlere Teams/Broadcaster-Labs/technische Media-Teams als Primärziel. Tranche 2 deferred externe `apps/analyzer-api`; interner `apps/analyzer-service` plus Library/CLI bleiben Standard. Tranche 3 deferred `apps/control-plane` ohne POC. Tranche 4 empfiehlt HTTP-Range-/Byte-Range-Loader als einzigen kleinen `NF-13`-Folgeslice. Tranche 5 hält Postgres als `defer-with-migration-seed` und Analytics als `defer`. | [`plan-0.15.0.md`](../done/plan-0.15.0.md) |
| ✅      | Selected Product Slice / Analyzer Range Fetch (`0.16.0`) | Released 2026-05-12. Lastenheft-Patch `1.1.21` mit RAK-106..RAK-110 in §13.20. Szenario B: HTTP-Range-/Byte-Range-Loader fuer manifest-referenzierte CMAF-Init-/Media-Segmente. Tranche 0 schließt RAK-106; Tranche 1 begrenzt den Lieferumfang auf HLS-CMAF-Byte-Ranges, No-new-public-schema und Fetch-Security-Grenzen. Tranche 2 liefert den HLS-Range-Fetch fuer explizite `EXT-X-MAP:BYTERANGE`-/`#EXT-X-BYTERANGE`-Offsets samt Contract-Fixtures. Tranche 3 schließt RAK-109 mit TS-, Drift-, Doku- und Security-Gates. Tranche 4 schließt RAK-110 mit Versions-Bump, Changelog, Plan-Archiv und Tag `v0.16.0`. | [`done/plan-0.16.0.md`](../done/plan-0.16.0.md) |
| ✅      | Hardening / Evidence Review (`0.17.0`) | Released 2026-05-13. Lastenheft-Patch `1.1.22` mit RAK-111..RAK-115 in §13.21. Szenario D: Hardening-only. Tranchen 0–4 erledigen Import, Evidence Review, Doku-/Defer-Entscheid, No-change-Gate-Nachweis und Release-Closeout: kein Productization-/Next-Slice-/Switch-Trigger, keine Runtime-/Public-API-/Schema-Aenderung ueber den versionstragenden Test-/Fixture-Asset-Bump hinaus. | [`done/plan-0.17.0.md`](../done/plan-0.17.0.md) |
| ✅      | Offene Risiken / Trigger-Re-Eval (`0.18.0`) | Released 2026-05-13. Decision-Closeout archiviert in `done/`: `R-9` bleibt K8s-Smoke-/Scrape-Policy-Trigger, `R-12` bleibt Nightly-detektierter WebRTC-Drift-Pfad ohne Safari-/WebKit-Pflicht, `R-13` bleibt Trivy-`continued` mit Re-Review-Artefakt und `expires` `2026-11-02`. Keine Runtime-, Wire-, Persistenz-, Public-API-, Schema- oder Default-Aenderung ueber den Release-Bump hinaus; K8s-Seed-Images wurden auf `0.18.0` synchronisiert. | [`done/plan-0.18.0.md`](../done/plan-0.18.0.md) |
| ✅      | Roadmap-Trigger-Nacharbeit (`0.19.0`) | Decision-only-Plan archiviert: Postgres, CORS-Variante A, externe Analyzer-API und Control-Plane bleiben bis zu konkreten Betreiber-/Multi-Tenant-/Audit-Triggern deferred. Kein Release-Tag, kein Versions-Bump. | [`done/plan-0.19.0.md`](../done/plan-0.19.0.md) |
| ✅      | Package Publishing (`0.20.0`) | Released 2026-05-13. Lastenheft-Patch `1.1.23` mit RAK-116..RAK-120: GitHub-Packages-Publish für `@pt9912/player-sdk` und `@pt9912/stream-analyzer`, Scope-Migration auf `@pt9912`, Release-Doku und Publish-Workflow. | [`done/plan-0.20.0.md`](../done/plan-0.20.0.md) |
| ✅      | OCI Image Publishing (`0.21.0`) | Released 2026-05-13. Lastenheft-Patch `1.1.24` mit RAK-121..RAK-125: versionierte GHCR-Images für API, Dashboard und Analyzer-Service, Make-Dry-Run, approval-gated Publish und Release-Hook. | [`done/plan-0.21.0.md`](../done/plan-0.21.0.md) |
| ✅      | Quality-Gates Follow-up (`0.22.0`) | Released 2026-05-13. Kein Lastenheft-Patch: `make benchmark-smoke` ist nach fünf grünen Beobachtungsläufen PR-blockierend über `make gates`; `benchmark-observation.yml` läuft hart; `mutation.yml` korrigiert den TS-Scope auf `@pt9912/player-sdk`, bleibt aber nicht-blockierend. | [`done/plan-0.22.0.md`](../done/plan-0.22.0.md) |
| ✅      | devalue-Security-Patch + Nightly-Audit (`0.22.1`) | Released 2026-05-17. Kein Lastenheft-Patch. `pnpm.overrides` hebt `devalue` auf `^5.8.1` (GHSA-77vg-94rm-hx3p, vier Tage nach `0.22.0`-Tag publiziert); neuer `security-audit.yml`-Nightly spiegelt `vuln-check`/`audit-ts`/`image-scan` täglich mit konsolidiertem Auto-Issue; Benchmark-Workflow-Pfadfix (`apps/.tmp/bench/` → `.tmp/bench/`) plus `tee`/`pipefail`-Logging; Issue-Body als `scripts/open-bench-regression-issue.sh`/`open-security-audit-issue.sh` ausgelagert. | [`done/plan-0.22.1.md`](../done/plan-0.22.1.md) |
| ✅      | Go-Stdlib-Security-Patch (`0.22.2`) | Released 2026-06-03. Kein Lastenheft-Patch. Erster echter Nightly-`security-audit.yml`-Treffer (Issue #3): `golang:1.26.3 → 1.26.4` an sechs Build-/Test-Image-Stellen (`apps/api/Dockerfile`, `Makefile::vuln-check`, `apps/api/Makefile::{arch-check,benchmark-smoke,fuzz-check,mutation-report}`) schließt GO-2026-5039 (`net/textproto`-Error-Echo via `auth.VaultSecretBackend.LoadSigningKeys`) und GO-2026-5037 (`crypto/x509`-Hostname-Parsing via `auth.NewRedisIssuanceRateLimiter`). Plus fünf bereits dokumentierte `perl-base`-Trivy-Ignores aus den Trivy-DB-Updates 2026-05-28/31 (CVE-2026-42496/42497/8376/9538/48962) für dashboard/analyzer-service und bilingualer README-Split. `make vuln-check` zeigt „No vulnerabilities found.". | [`done/plan-0.22.2.md`](../done/plan-0.22.2.md) |
| ✅      | Security-/CI-Sammel-Patch (`0.22.3`) | Released 2026-06-16. Kein Lastenheft-Patch. Bündelt vier aufeinanderfolgende `security-audit.yml`-Nightly-Treffer: Trivy `0.59.1 → 0.71.0` + `CVE-2026-45447`/`-48959`-Ignores (Issue #4), esbuild `^0.28.1`-Override (GHSA-gv7w-rqvm-qjhr, Issue #5), libsqlite3-FTS5-Ignores `CVE-2026-11822`/`-11824` (kein Upstream-Fix, `expires` 2026-09-12, Issue #6), vite `^8.0.16` (GHSA-fx2h-pf6j-xcff, Issue #7). Plus GitHub-Actions-Node-24-Runtime, WebRTC-Drift-Test-Fix (rollt aus `plan-0.22.3-webrtc-drift` ein, dort ohne eigenen Tag) und Flaky-CORS-Preflight-Test-Fix (geteilter `http.DefaultClient` → `srv.Client()`). `make audit-ts` + `make image-scan` lokal grün. | [`done/plan-0.22.3-webrtc-drift.md`](../done/plan-0.22.3-webrtc-drift.md) |
| ✅      | Security-/Tooling-Patch (`0.22.4`) | **Released 2026-06-23** (Tag `v0.22.4`, GHCR- + npm-Publish grün — publish-images `28007976324`, publish-packages `28007976457`). Kein Lastenheft-Patch. `golang.org/x/net 0.53.0 → 0.56.0` (transitiv `x/sys 0.46.0`, `x/text 0.38.0`) behebt sechs vom Trivy-Image-Scan im `usr/local/bin/api`-gobinary gemeldete HIGH-CVEs (`CVE-2026-25680`/`-25681`/`-27136`/`-39821`/`-42502`/`-42506`; `govulncheck` grün — Call-Graph erreicht die Pfade nicht, Trivy scannt den Modulgraphen unabhängig; Issue #9, Nightly-Lauf `27996614696`); `undici`-`pnpm.overrides` `^7.28.0` (GHSA-vmh5-mc38-953g); Trivy-Pin `0.71.0 → 0.71.2`. Plus ENV-konfigurierbarer Ingest-Rate-Limiter (`MTRACE_RATE_LIMIT_CAPACITY`/`-REFILL`, Default 100/100 unverändert) und die Load-Smoke-Readback-`COUNT(*)`-Reconciliation. `make vuln-check` + `make image-scan` + `make gates` lokal grün. | [`done/plan-0.22.4.md`](../done/plan-0.22.4.md), Issue #9 |
| ✅      | Postgres Scale-out (`0.23.0`)       | **Released 2026-07-11** (Tag `v0.23.0`, GHCR + npm publiziert). Lastenheft-Patch `1.1.25` mit RAK-126..RAK-130: optionaler Postgres-Runtime-Adapter (`MTRACE_PERSISTENCE=postgres`), DB-autoritativer Ingest-Sequencer (R-28), R-27-Read-Wasserzeichen, Multi-Replica-Harness (`pg_advisory_lock`-Migration), Scale-out-Lasttest → R-26 c belegt (0 Verlust/0 Dups über 2 Replicas @ ~1,4 Mio Events; Durchsatz flaschenhals-abhängig, `budgets.md` §8). SQLite bleibt Default. | [`done/plan-0.23.0-postgres-scaleout.md`](../done/plan-0.23.0-postgres-scaleout.md) |
| ✅      | Multi-Tenant-Fairness + Cutover (`0.25.0`) | **Released 2026-07-13**. Lastenheft-Patch `1.1.26` mit RAK-131..RAK-135: shared Redis-Ingest-Limiter (`MTRACE_RATE_LIMIT_BACKEND=redis`, port-erhaltend, memory-Default; fail-open-to-memory + `MTRACE_RATE_LIMIT_FAIL_CLOSED`), `client_ip`-XFF-Trust-Boundary (IP-validiert), Fairness-Nachweis `make smoke-scaleout-fairness` (Inversion 0,96–1,0× statt 2,01×, Noisy-Neighbor-Isolation, Korrektheits-Gates unverändert — `budgets.md` §9) → **R-26 b 🟢 (R-26 komplett)**; plus SQLite→Postgres-Cutover als opt-in Ops-Werkzeug (`make cutover`, 4 Phasen, Quelle read-only, Runbook `docs/ops/postgres-cutover.md`, `make smoke-cutover` 10 Cases) → **R-29 🟢** (ADR-0007; kein eigener 0.24.0-Tag, rollte hierin). T1–T3 code-reviewt (21 Findings gefixt). Kein 0.24.0-Tag vergeben (`0.24.0` blieb Plan-Identifier). | [`plan-0.25.0-shared-ingest-limiter.md`](../done/plan-0.25.0-shared-ingest-limiter.md), [`plan-0.24.0-sqlite-postgres-cutover.md`](../done/plan-0.24.0-sqlite-postgres-cutover.md) |

### 1.2 Nächste Phase

`0.25.0` ist **veröffentlicht** (2026-07-13): shared Redis-Ingest-Limiter
(R-26 b — damit ist `R-26` über alle drei Achsen belegt) plus die
SQLite→Postgres-Datenmigration (R-29, ADR-0007). Lastenheft-Patch `1.1.26`
(RAK-131..RAK-135), RAK-Verifikationsmatrix in Plan §6.1; beide Pläne
(`plan-0.24.0`, `plan-0.25.0`) wandern mit dem Tag nach `done/`.

Es liegt derzeit **keine geschnittene größere Folge-Tranche** vor. Bekannte
Folge-Kandidaten (alle mit dokumentierten Triggern, keine aktiven Blocker):
`R-30` (SSE-Backfill-Skip-Exposition unter Multi-Replica — Prüfpunkt beim
nächsten Scale-out-Lasttest-Ausbau), `R-24` (Load-Smoke-Nightly-Debounce
beobachten), policy-getriebene Per-Projekt-Limiter-Buckets (RAK-74-Anschluss),
Redis-Cluster-Tauglichkeit der Multi-Key-Lua-Limiter, Durchsatz-Scaling
jenseits eines Single-Postgres (`budgets.md` §8-Konsequenz). Mutation-
Blockierung bleibt deferred, bis echte >70%-Score-Reihen vorliegen.

---

## 2. Nächste Schritte

Statusspalte: ✅ abgeschlossen · 🟡 in Arbeit · ⬜ geplant.

Verweise nutzen die Lastenheft-Kennungen (`F-`, `NF-`, `MVP-`, `AK-`)
wo sie existieren; Plan- und ADR-Sektionsnummern werden behalten,
weil dort kein ID-System existiert. Granularer Lieferstand pro Release
steht in den jeweiligen Plan-Dateien mit DoD-Checkboxen und
Commit-Hashes, z. B. [`docs/planning/done/plan-0.3.0.md`](../done/plan-0.3.0.md).

| #   | Status | Schritt                                                                                                               | Trigger                                                         | Verweis                                                       |
| --- | ------ | --------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------- | ------------------------------------------------------------- |
| 1   | ✅      | `spike/go-api` → `apps/api` auf `main` integrieren                                                                    | Sofort                                                          | MVP-2; OE-9; SP-41                                            |
| 2   | ✅      | Lastenheft auf `1.0.0` heben                                                                                          | Nach Schritt 1                                                  | OE-2; OE-9; SP-41                                             |
| 3   | ✅      | README Tech-Overview anpassen                                                                                         | Nach Schritt 2                                                  | MVP-17; SP-41                                                 |
| 4   | ✅      | Phase-2-Risiken in `docs/planning/in-progress/risks-backlog.md`                                                              | Nach Schritt 3                                                  | SP-41                                                         |
| 5   | ✅      | `spec/architecture.md` schreiben                                                                                      | Vor `0.1.0`-DoD                                                 | AK-3, AK-10                                                   |
| 6   | ✅      | `spec/telemetry-model.md` schreiben (Datenmodell, Wire-Format, Cardinality — kein Observability-Setup)                | Vor `0.1.0`-DoD                                                 | F-91, F-92, F-95..F-105, F-106..F-115, F-118..F-130, AK-9     |
| 7   | ✅      | `docs/user/local-development.md` schreiben                                                                            | Vor `0.1.0`-DoD                                                 | AK-1, AK-2                                                    |
| 8   | ✅      | Dashboard-App (`apps/dashboard`) anlegen — `0.1.1` (siehe `plan-0.1.1.md`)                                            | Nach `0.1.0`-Release                                            | MVP-3; F-23..F-28                                             |
| 9   | ✅      | Player-SDK (`packages/player-sdk`) anlegen — `0.1.1` (siehe `plan-0.1.1.md`)                                          | Nach `0.1.0`-Release                                            | MVP-5; F-63..F-67                                             |
| 10  | ✅      | Docker-Compose-Lab inkl. MediaMTX + FFmpeg (Core in `0.1.0`, `dashboard` in `0.1.1`, observability-Profil in `0.1.2`) | Core: vor `0.1.0`-DoD; Erweiterungen mit jeweiligem Sub-Release | MVP-7..MVP-9; F-82..F-88                                      |
| 11  | ✅      | Observability-Stack (Prometheus + optional Grafana, OTel-Collector) — `0.1.2` (siehe `plan-0.1.2.md`)                 | Nach `0.1.1`-Release                                            | MVP-10, MVP-15; F-89..F-94                                    |
| 12  | ✅      | `docs/planning/done/plan-0.2.0.md` anlegen und `0.2.0`-Scope in umsetzbare Tranchen schneiden                         | Nach `0.1.2`-Release                                            | RAK-11..RAK-21                                                |
| 13  | ✅      | Player-SDK-Paketierung und Public API stabilisieren                                                                   | Nach Schritt 12                                                 | RAK-11, RAK-12                                                |
| 14  | ✅      | Event-Schema-Versionierung und SDK↔Schema-Kompatibilitätscheck in CI planen                                           | Nach Schritt 12                                                 | RAK-13, RAK-21                                                |
| 15  | ✅      | hls.js-Adapter, HTTP-Transport sowie Batching/Sampling/Retry-Grenzen testbar absichern                                | Nach Schritt 12                                                 | RAK-14, RAK-15, RAK-17                                        |
| 16  | ✅      | OTel-Transport-Option bewerten und Performance-Budget nachweisen                                                      | Nach Schritt 15                                                 | RAK-16, RAK-18                                                |
| 17  | ✅      | Browser-Support-Matrix und Demo-Integrationsdoku erstellen                                                            | Nach Schritt 16                                                 | RAK-19, RAK-20                                                |
| 18  | ✅      | OE-3-Folge-ADR für Persistenz vorbereiten                                                                             | Parallel zu `0.2.0`-Planung                                     | OE-3; MVP-16                                                  |
| 19  | ✅      | `docs/planning/done/plan-0.3.0.md` anlegen und `0.3.0`-Scope in umsetzbare Tranchen schneiden                         | Nach `0.2.0`-Release                                            | RAK-22..RAK-28                                                |
| 20  | ✅      | Stream-Analyzer-Paket `packages/stream-analyzer` anlegen                                                              | Nach Schritt 19                                                 | RAK-22..RAK-26; MVP-33                                        |
| 21  | ✅      | HLS-Manifest laden und Master-/Media-Playlist-Erkennung umsetzen                                                      | Nach Schritt 20                                                 | RAK-22, RAK-23, RAK-24                                        |
| 22  | ✅      | Segment-Dauern prüfen und JSON-Ergebnisformat stabilisieren                                                           | Nach Schritt 21                                                 | RAK-25, RAK-26                                                |
| 23  | ✅      | API-Anbindung über bestehenden StreamAnalyzer-Port umsetzen                                                           | Nach Schritt 22                                                 | RAK-27; F-22, F-33                                            |
| 24  | ✅      | CLI-Grundlage für den Stream Analyzer schaffen                                                                        | Nach Schritt 22                                                 | RAK-28; MVP-34                                                |
| 25  | ✅      | OE-3/Persistenz nach ADR-Draft neu bewerten — Entscheidung getroffen: SQLite (ADR-0002 `Accepted`, RAK-32-getrieben)  | Vor `0.4.0`-Scope-Cut                                           | OE-3; MVP-16; ADR-0002                                        |
| 26  | ✅      | OE-5/Live-Updates entscheiden — SSE mit Polling-Fallback, WebSocket deferred                                          | Vor `0.4.0`-Scope-Cut                                           | OE-5; MVP-31; ADR-0003                                        |
| 27  | ✅      | `docs/planning/done/plan-0.4.0.md` anlegen und `0.4.0`-Scope in Tranchen schneiden                             | Nach Schritt 26                                                 | RAK-29..RAK-35                                                |
| 28  | ✅      | SQLite-Persistenz, durable Cursor und Cursor-Kompatibilitätsmatrix umsetzen                                           | Nach Schritt 27                                                 | RAK-32; ADR-0002; plan-0.4.0 Tranche 1                        |
| 29  | ✅      | SOLID-nahes `golangci-lint`-Zusatzprofil konfigurieren und Lint-Findings abarbeiten                                   | Nach Lastenheft-/Quality-Doku-Festlegung                        | `spec/lastenheft.md` §10.1; `docs/user/quality.md` §1.2       |
| 30  | ✅      | SOLID-nahes TypeScript-/Svelte-Lintprofil für Apps und Packages festlegen, konfigurieren und Findings abarbeiten      | Nach Schritt 29 oder parallel bei Workspace-Lint-Ausbau         | `spec/lastenheft.md` §10.2–§10.4; `docs/user/quality.md` §1.1 |
| 31  | ✅      | Tempo-unabhängiges Session-Trace-Modell mit lokaler `trace_id`/`correlation_id` festlegen und testen                  | Nach Schritt 30                                                 | RAK-29; RAK-32; plan-0.4.0 Tranche 2 (§3.1–§3.4c, abgeschlossen) |
| 32  | ✅      | Manifest-, Segment- und Player-Ereignisse in gemeinsamen Trace-/Korrelationskontext integrieren                       | Nach Schritt 31                                                 | RAK-30; plan-0.4.0 Tranche 3                                  |
| 33  | ✅      | Dashboard-Session-Verlauf ohne Tempo inkl. SSE, Backfill, Polling-Fallback und SQLite-Restart-Test umsetzen           | Nach Schritt 30                                                 | RAK-32; ADR-0003; plan-0.4.0 Tranche 4                        |
| 34  | ✅      | Optionales Tempo-Profil anbinden, ohne RAK-29/RAK-32 vom Trace-Backend abhängig zu machen                             | Nach Schritt 31                                                 | RAK-31; plan-0.4.0 Tranche 5                                  |
| 35  | ✅      | Aggregat-Metriken, Drop-/Invalid-/Rate-Limit-Sichtbarkeit und Cardinality-/Sampling-Doku abschließen                  | Parallel zu Schritten 30–33                                     | RAK-33..RAK-35; plan-0.4.0 Tranchen 6 (✅) und 7 (✅)           |
| 36  | ✅      | Release-Akzeptanzkriterien `0.4.0` verifizieren und Roadmap auf `0.5.0` umstellen                                     | Nach Schritten 30–35                                            | RAK-29..RAK-35; plan-0.4.0 Tranche 8; Tag `v0.4.0` auf `9e4fdb3`, CI grün                                       |
| 37  | ✅      | Multi-Protocol-Lab (`examples/`) plus opt-in Smokes ausliefern und Roadmap auf `0.6.0` umstellen                      | Nach Schritt 36                                                 | RAK-36..RAK-40; plan-0.5.0 Tranchen 0–6; Tag `v0.5.0` auf `a56dc0b`, CI-Run 25364250989 grün                      |
| 38  | ✅      | SRT Health View (`0.6.0`) mit MediaMTX-API als Quelle plus Read-API/Dashboard ausliefern                              | Nach Schritt 37                                                 | RAK-41..RAK-46; plan-0.6.0 Tranchen 0–7; Tag `v0.6.0` auf `d08a89f`, CI-Run 25380938222 grün                      |
| 39  | ✅      | WebRTC-Lab-Erweiterung (`0.7.0`) mit Lab-Compose, opt-in Smoke und Telemetrie-Vorbereitung ausliefern                 | Nach Schritt 38                                                 | RAK-47..RAK-50; plan-0.7.0 Tranchen 0–5; Tag `v0.7.0` (Closeout-Commit)                                          |
| 40  | ✅      | Lastenheft-Patch `1.1.10` schreiben — RAK-51 von „Kann" auf „Muss" hochgezogen + neue RAK-52..RAK-55 in §13.10 für Public-API/hls.js-Trennung, produktive WebRTC-Telemetrie und Compat-Tests definiert     | Vor Tranche-0-Aktivierung von `0.8.0`                            | RAK-51, MVP-24; [`plan-0.8.0.md`](../done/plan-0.8.0.md) §0.2; Patch-Log §4a.13 in [`plan-0.1.0.md`](../done/plan-0.1.0.md)            |
| 41  | ✅      | `0.8.0` Player-SDK-WebRTC-Adapter ausliefern: Public-API + hls.js-Trennung, WHEP-Adapter gegen `examples/webrtc/`, produktive WebRTC-Telemetrie auf `spec/telemetry-model.md` §3.2/§3.5-Allowlist (R-12 release-blockierend), Compat-Tests | Nach Schritt 40                                                  | RAK-51..RAK-55 (Lastenheft `1.1.10` §13.10); [`plan-0.8.0.md`](../done/plan-0.8.0.md) Tranchen 0–5; Tag `v0.8.0` (Release-Gate-Fix nach Closeout) |
| 42  | ✅      | Lastenheft-Patch `1.1.11` schreiben — neuer §13.11 mit RAK-56 (Drift-Smoke, Soll), RAK-57 (SRS-Lab, Kann), RAK-58 (DASH-Manifest-Analyse, Muss) und RAK-59 (DASH-CLI, Kann); §12.3 MVP-37 von „Kann" auf „Muss" entsprechend NF-12 hochgezogen | Vor Tranchen 1–4 von `0.9.0`                                     | RAK-56..RAK-59, MVP-36, MVP-37, NF-12; [`plan-0.9.0.md`](../done/plan-0.9.0.md) §0.2; Patch-Log §4a.14 in [`plan-0.1.0.md`](../done/plan-0.1.0.md) |
| 43  | ✅      | `0.9.0` Drift-Smoke + SRS-Lab + DASH-Analyse ausliefern: Browser-Drift-Smoke gegen `examples/webrtc/`-Lab plus Nightly-CI (R-12 wandert auf „automatisiert detektiert"), `examples/srs/`-Lab analog der anderen Multi-Protocol-Beispiele, DASH-MPD-Pfad im `@pt9912/stream-analyzer` mit `analyzerKind: "dash"` und CLI-Dispatcher | Nach Schritt 42                                                  | RAK-56..RAK-59 (Lastenheft `1.1.11` §13.11); [`plan-0.9.0.md`](../done/plan-0.9.0.md) Tranchen 0–5; Tag `v0.9.0` |
| 44  | ✅      | `0.9.6` Lastenheft-Konvergenz-Patch ausliefern: fehlende Muss-Repo-Artefakte (`CONTRIBUTING.md`, `SECURITY.md`, `.env.example`, `deploy/`-Struktur), Lastenheft-Patch `1.1.12` (F-7-Status, neue Pflichtdokumente-Kennung `F-131`, NF-13/NF-18 harmonisieren, MVP-19..MVP-26 redaktionell entzerren) und Go-Stdlib-Bump `golang:1.26.3` (GO-2026-4982/4980/4971/4918); keine User-Surface-Änderung | Nach Schritt 43                                                  | F-7, F-131 (neu), NF-13, NF-18, NF-25, NF-29, MVP-19..MVP-26, MVP-40..MVP-42; [`plan-0.9.6.md`](../done/plan-0.9.6.md) Tranchen 0–4 |
| 45  | ✅      | `0.10.0` CMAF-Analyse ausgeliefert (NF-13-Vollumsetzung im Stream-Analyzer-Scope): manifestbasierte HLS-/DASH-CMAF-Signale plus begrenzte binäre CMAF-Konformitätsprüfung ausgewählter Init-/Media-Segmente; Lastenheft-Patch `1.1.13` mit RAK-60..RAK-64 in §13.12; additives `details.cmaf`-Schema unter HLS-/DASH-Detail-Objekten ohne neuen `analyzerKind`; ISO-BMFF-Box-Parser und bounded Segment-Loader (Brand-Allowlist `cmfc`/`cmf2`/`cmfs`/`cmff`; Defaults `maxSegmentBytes=2_000_000`/`maxBinarySegments=6`) | Nach Schritt 44 | NF-13, RAK-60..RAK-64; [`done/plan-0.10.0.md`](../done/plan-0.10.0.md) Tranchen 0–6 |
| 46  | ✅      | `0.11.0` Ingest-Gateway / Stream Control ausgeliefert: F-46..F-51 aus dem Lastenheft (Patch `1.1.14` Hochstufung von Kann auf Release-Muss) in einen umsetzbaren Stream-Control-Pfad geschnitten — CSPRNG-Stream-Keys (nur `key_hash` persistiert), `srt`/`rtmp`-Endpunkte, 1:1-Routing-Regeln, deterministischer MediaMTX-Konfigurations-Generator und lokal reproduzierbares Lifecycle-Eventmodell mit `evt_`-IDs und Source-Allowlist `local-smoke`/`mediamtx-hook`. Architektur Variante B (Modul in `apps/api`, kein eigener `apps/ingest-gateway`-Service). Lastenheft-Patch `1.1.14` mit RAK-65..RAK-70 in §13.13. | Nach Schritt 45 | F-46..F-51, MVP-38, RAK-65..RAK-70; [`done/plan-0.11.0.md`](../done/plan-0.11.0.md) Tranchen 0–6; Tag `v0.11.0` |
| 47  | ✅      | `0.12.0` Auth / Token Lifecycle ausgeliefert: F-111..F-113 als zusammenhängender Security-/Auth-Scope — kurzlebige HMAC-SHA-256-signierte Session Tokens (`POST /api/auth/session-tokens` + Konsum via `Authorization: Bearer mtr_st_*` / `X-MTrace-Session-Token`), rotierbare `mtr_pt_*`-Project-Token-Generationen mit persistiertem `grace_until` (V4-Migration), tenant-spezifische Ingest Policies + §3.9-konformer CORS-Preflight (204 mit minimaler Signalisierung). Lastenheft-Patch `1.1.15` mit RAK-71..RAK-76 in §13.14. Tranchen 0–6 ausgeliefert 2026-05-10. | Nach Schritt 46 | F-111..F-113, RAK-71..RAK-76; [`done/plan-0.12.0.md`](../done/plan-0.12.0.md) |
| 47.5 | ✅    | `0.12.1` Trigger-Re-Eval + Operator-Doku als Patch-Release ausgeliefert (2026-05-10): pro aktivem `R-N`-Item im Backlog (R-5/R-7/R-9/R-10/R-11/R-12/R-13/R-14/R-15/R-16/R-17/R-18/R-20/R-21) Trigger-Status-Notiz (alle „nicht ausgelöst"), Operator-Runbook für Multi-Key-Signing-Rotation in `auth.md` §5.3.1 (Soll-Workflow als Doku; Code-Pfad in 0.12.5), OS-1..OS-5 als ⬛ Duplikate in §1.2 abgelegt, OS-6 zu R-22 konvertiert. Kein Lastenheft-Patch, keine RAK-Matrix, keine neue User-Surface. | Nach Schritt 47 | R-5..R-21; [`done/plan-0.12.1.md`](../done/plan-0.12.1.md) |
| 47.6 | ✅    | `0.12.5` Auth-/Ingest-Adapter-Minor ausgeliefert (2026-05-11): `MultiKeySigningResolver` + ENV-Parser (R-18), `SqliteIssuanceRateLimiter` mit Migration V5 + ENV-Selektor (R-17 teilweise), `AuthSecretBackend`-Driven-Port + ENV/Vault-Adapter-Skelett (R-20 teilweise), `BrowserIngestPolicy` mit Preflight-Handler + POST-Enforcement (R-21), `MediaMTXAuthHookHandler` als `externalAuth`-Bridge (R-14), `OutboundWebhookDispatcher` mit HMAC + Exponential-Backoff-Retry (R-16). Lastenheft-Patch `1.1.16` mit RAK-77..RAK-82 in §13.15. Fünf neue opt-in Smokes. | Nach Schritt 47.5 | R-14, R-16, R-17, R-18, R-20, R-21; RAK-77..RAK-82; [`done/plan-0.12.5.md`](../done/plan-0.12.5.md) |
| 47.7 | ✅    | `0.12.6` Auth-/Ingest-Folge-Items-Minor ausgeliefert (2026-05-12): alle 9 R-N-Items adressiert — Time-Skew-Persistenz (R-5 🟢, RAK-83, V6 + `event.time_skew_warning`), `ListSessions`-Bulk-Read (R-7 🟢, RAK-84), Sampling-ppm-Marker (R-10 🟢 minus Heuristik, RAK-85, V7 + Banner), SRT-Cursor-Pagination via `samples_cursor`/`next_samples_cursor` (R-11 🟢, RAK-86, Wire-Codec v3), Trivy-Re-Review (R-13 🟢-Wartung, Expiry 2026-11-02), mediamtx-Provisioner mit additivem `?provision=mediamtx` (R-15 🟢, RAK-87), Multi-Host-Limiter via Redis (R-17 🟢 final, RAK-88), Vault-AppRole + KMS-Skeleton (R-20 🟢 final, RAK-89), Origin-/IP-Rate-Limiter (R-22 🟢, RAK-90). Lastenheft-Patch `1.1.17` mit RAK-83..RAK-90 in §13.16. Sechs neue opt-in Smokes. | Nach Schritt 47.6 | R-5/R-7/R-10/R-11/R-13/R-15/R-17/R-20/R-22; RAK-83..RAK-90; [`done/plan-0.12.6.md`](../done/plan-0.12.6.md) |
| 48  | ✅      | `0.13.0` Production / Ops Backends ausgeliefert: Postgres und Analytics-Backends als deferred mit Triggern entschieden, optionale Kubernetes-Manifeste, Devcontainer und Release-Guard geliefert. NF-18 mit MVP-42 harmonisiert. Minor-Release mit Lastenheft-Patch `1.1.18` und RAK-91..RAK-95 in §13.17. | Nach Schritt 47.7 | RAK-91..RAK-95 in `spec/lastenheft.md` §13.17; NF-18, MVP-40..MVP-44; [`done/plan-0.13.0.md`](../done/plan-0.13.0.md) |
| 49  | ✅      | `0.14.0` Ops Backend Follow-up ausgeliefert: Szenario C importiert K8s-/Devcontainer-/Release-Guard-Seeds aus `0.13.0` für Hardening/Validation. Postgres bleibt `defer-with-migration-seed`, Analytics bleibt `defer`; keine neue lokale Pflichtabhängigkeit. Lastenheft-Patch `1.1.19` mit RAK-96..RAK-100 in §13.18. | Nach Schritt 48 | RAK-96..RAK-100 in `spec/lastenheft.md` §13.18; MVP-40..MVP-44; [`done/plan-0.14.0.md`](../done/plan-0.14.0.md) |
| 50  | ✅      | `0.15.0` Product Scope / Analyzer Boundary released: Szenario A fokussiert Zielgruppe + Analyzer-Boundary, bevor externe Analyzer-API, Control-Plane, Postgres/Analytics oder Production-K8s in Implementierung gehen. Tranche 1 schließt RAK-101 mit Selbsthoster-/kleine-Team-/Broadcaster-Lab-Fokus. Tranche 2 schließt RAK-102: externe Analyzer-API deferred, interner `apps/analyzer-service` plus Library/CLI bleibt Standard. Tranche 3 schließt RAK-103: Control-Plane deferred, kein POC ohne Betreiber-/Auth-/Tenant-Trigger. Tranche 4 schließt RAK-104: HTTP-Range-/Byte-Range-Loader als einziger kleiner `NF-13`-Folgeslice empfohlen. Tranche 5 schließt RAK-105: Postgres bleibt `defer-with-migration-seed`, Analytics bleibt `defer`. Lastenheft-Patch `1.1.20` mit RAK-101..RAK-105 in §13.19. | Nach Schritt 49 | RAK-101..RAK-105 ✅; `spec/lastenheft.md` §7.5.5/§7.5.6/§8.3/§12.1/§13.19/§16.1; MVP-20, F-132, NF-13, MVP-40/MVP-41; [`done/plan-0.15.0.md`](../done/plan-0.15.0.md) |
| 51  | ✅      | `0.16.0` Selected Product Slice / Analyzer Range Fetch released: Szenario B importiert `RAK-104` als einzigen Go-Pfad. Tranche 1 begrenzt den Lieferumfang auf HLS-CMAF-Byte-Ranges (`#EXT-X-MAP` mit `BYTERANGE`-Attribut und erstes `#EXT-X-BYTERANGE`-fMP4-Media-Segment), No-new-public-schema und Fetch-Security-Grenzen. Tranche 2 liefert den HLS-Range-Fetch fuer explizite Offsets im bestehenden Binary-Check-Pfad. Tranche 3 schließt RAK-109 mit `make security-gates` plus TS-/Doku-/Drift-Gates. Tranche 4 schließt RAK-110 mit Version `0.16.0`, Changelog, Roadmap, Plan-Archiv und Tag `v0.16.0`; externe Analyzer-API, Control-Plane, Postgres/Analytics, Production-K8s, LL-CMAF, vollständige Segmentsets, Codec-Decoding und Player-Laufzeitpfade bleiben deferred. Lastenheft-Patch `1.1.21` mit RAK-106..RAK-110 in §13.20. | Nach Schritt 50 | RAK-106..RAK-110 ✅; `spec/lastenheft.md` §13.20; NF-13; [`done/plan-0.16.0.md`](../done/plan-0.16.0.md); Tag `v0.16.0` |
| 52  | ✅      | `0.17.0` Hardening / Evidence Review released: `0.16.0`-Closeout importiert, Szenario D gewaehlt, Lastenheft-Patch `1.1.22` mit RAK-111..RAK-115 vergeben, Evidence geprueft, Tranche 2 als Doku-/Defer-Artefakt ohne Code-/Runtime-Aenderung geschlossen, Tranche 3 als No-change-Gate-Nachweis abgeschlossen und Tranche 4 mit Version `0.17.0`, versionstragendem Test-/Fixture-Asset-Bump, Changelog, Roadmap, Plan-Archiv und Tag `v0.17.0` geschlossen. | Nach Schritt 51 | RAK-111..RAK-115 ✅; `spec/lastenheft.md` §13.21; [`done/plan-0.17.0.md`](../done/plan-0.17.0.md); Tag `v0.17.0` |
| 53  | ✅      | `0.18.0` Offene Risiken / Trigger-Re-Eval released: `R-9`, `R-12` und `R-13` ohne Implementierungs-Trigger entschieden, Resttrigger in Backlog und Roadmap praezisiert, K8s-Seed-Images und Versionstraeger auf `0.18.0` synchronisiert. Kein Lastenheft-Patch; normativer Stand bleibt `1.1.22`. | Nach Schritt 52 | [`done/plan-0.18.0.md`](../done/plan-0.18.0.md); Tag `v0.18.0` |
| 54  | ✅      | `0.20.0` Package Publishing released: publishbare npm-Pakete auf `@pt9912/player-sdk` und `@pt9912/stream-analyzer` umgestellt, GitHub-Packages-Workflow ergänzt, Release-Doku erweitert und erster Package-Publish vorbereitet/ausgeführt. Dashboard und Analyzer-Service bleiben `private: true`. | Nach Schritt 53 und Decision-only `0.19.0` | RAK-116..RAK-120 ✅; `spec/lastenheft.md` §13.22; [`done/plan-0.20.0.md`](../done/plan-0.20.0.md); Tag `v0.20.0` |
| 55  | ✅      | `0.21.0` OCI Image Publishing released: GHCR-Namensschema, Make-Targets, Publish-Workflow und Release-Doku für drei Runtime-Images geliefert; kein `latest`, kein Production-K8s-Go. | Nach Schritt 54 und ausgelöstem Container-Publishing-Trigger | RAK-121..RAK-125 ✅; `spec/lastenheft.md` §13.23; [`done/plan-0.21.0.md`](../done/plan-0.21.0.md); Tag `v0.21.0` |
| 56  | ✅      | `0.23.0` Postgres Scale-out released (2026-07-11): optionaler Postgres-Runtime-Adapter (`MTRACE_PERSISTENCE=postgres`), DB-autoritativer Ingest-Sequencer (R-28), R-27-Read-Wasserzeichen, Multi-Replica-Harness mit `pg_advisory_lock`-serialisierter Startup-Migration, Scale-out-Lasttest → **R-26 c belegt** (0 Verlust/0 Dup über 2 Replicas @ ~1,4 Mio Events; Durchsatz store-gebunden = Single-Postgres-Decke ~12k ev/s, ehrlich attribuiert in `budgets.md` §8). SQLite bleibt Default. Lastenheft-Patch `1.1.25` mit RAK-126..RAK-130 in §13.24; GHCR + npm publish grün. Zwischen 0.21.0 und hier: Patch-Releases 0.22.0–0.22.4 (Security/Tooling, §1-Historie). | Nach Schritt 55 und ausgelöstem Multi-Replica-/Scale-out-Trigger (R-26 c) | RAK-126..RAK-130 ✅; `spec/lastenheft.md` §13.24; [`done/plan-0.23.0-postgres-scaleout.md`](../done/plan-0.23.0-postgres-scaleout.md); Tag `v0.23.0` |
| 57  | ✅      | **Nächste größere Tranche wählen** (Folge-Scope von `0.23.0`; beide Optionen geliefert und mit `0.25.0` released 2026-07-13): **(A) R-26 b** — repliken-übergreifend fairer Ingest-Limiter (shared Redis): **GELIEFERT** (2026-07-13, [`done/plan-0.25.0-shared-ingest-limiter.md`](../done/plan-0.25.0-shared-ingest-limiter.md) T1–T3: Redis-Adapter port-erhaltend, Multi-Tenant-Lab, **Fairness-Inversion 0,96× statt 2,01×** gemessen + Noisy-Neighbor-Isolation über den LB, budgets.md §9; R-26 → 🟢; getaggter Release `0.25.0` pending). **(B) `plan-0.24.0-sqlite-postgres-cutover`** — SQLite→Postgres-Datenmigration bestehender Läufe; **Implementierung + Closeout komplett** (2026-07-12: 4 Phasen `doctor`/`profile`/`bulk`/`incremental`/`switch` gebaut + alle code-reviewt, `make cutover`, `make smoke-cutover` 8 Cases grün, Runbook `docs/ops/postgres-cutover.md`, ADR-0007 „geliefert", R-29 🟢). **Kein eigenes 0.24.0-Tag** (reine Ops-Tooling ohne Runtime-/Package-Änderung, Owner-Entscheidung 2026-07-12) — liegt auf `main` + CHANGELOG `[Unreleased]`, rollt in den nächsten Release. | Nach `0.23.0`-Release; Trigger = Betreiber-Bedarf (Multi-Tenant-Fairness bzw. Bestandsdaten-Migration) | R-26 b: [`risks-backlog.md`](risks-backlog.md); R-29: [`in-progress/plan-0.24.0-sqlite-postgres-cutover.md`](../done/plan-0.24.0-sqlite-postgres-cutover.md), [ADR-0007](../../adr/0007-sqlite-postgres-data-cutover.md) |

---

## 3. Release-Übersicht

Statusspalte: ✅ abgeschlossen · 🟡 in Arbeit · ⬜ geplant.

| Version | Titel                        | Status | Akzeptanzkriterien                                                                                    |
| ------- | ---------------------------- | ------ | ----------------------------------------------------------------------------------------------------- |
| `0.0.x` | Spike + Planungsphase        | ✅      | —                                                                                                     |
| `0.1.0` | Backend Core + Demo-Lab      | ✅      | RAK-1, RAK-3, RAK-4, RAK-6, RAK-8 (initial); DoD-Tracking in [`plan-0.1.0.md`](../done/plan-0.1.0.md) |
| `0.1.1` | Player-SDK + Dashboard       | ✅      | RAK-2, RAK-5, RAK-7; DoD-Tracking in [`plan-0.1.1.md`](../done/plan-0.1.1.md)                         |
| `0.1.2` | Observability-Stack          | ✅      | RAK-9, RAK-10; DoD-Tracking in [`plan-0.1.2.md`](../done/plan-0.1.2.md)                               |
| `0.2.0` | Publizierbares Player SDK    | ✅      | RAK-11..RAK-21                                                                                        |
| `0.3.0` | Stream Analyzer              | ✅      | RAK-22..RAK-28; DoD-Tracking in [`plan-0.3.0.md`](../done/plan-0.3.0.md)                              |
| `0.4.0` | Erweiterte Trace-Korrelation | ✅      | RAK-29..RAK-35; Tag `v0.4.0` auf `9e4fdb3`, CI-Run 25359933129 grün                                   |
| `0.5.0` | Multi-Protocol Lab           | ✅      | RAK-36..RAK-40; Tag `v0.5.0` auf `a56dc0b`, CI-Run 25364250989 grün                                   |
| `0.6.0` | SRT Health View              | ✅      | RAK-41..RAK-46; DoD-Tracking in [`done/plan-0.6.0.md`](../done/plan-0.6.0.md)                        |
| `0.7.0` | WebRTC-Lab-Erweiterung       | ✅      | RAK-47..RAK-50; RAK-51 deferred / Folgeplan; DoD-Tracking in [`done/plan-0.7.0.md`](../done/plan-0.7.0.md)               |
| `0.8.0` | Player-SDK-WebRTC-Adapter    | ✅      | RAK-51..RAK-55; DoD-Tracking in [`done/plan-0.8.0.md`](../done/plan-0.8.0.md)                                                                              |
| `0.8.5` | Quality-Gates Wave 1 (Patch) | ✅      | Security-Gates (`vuln-check`/`audit-ts`/`image-scan`) als PR-blockierender CI-Job parallel zu `build`; Generated-Artifact-Drift-Gate Teil von `make gates`; Migrations-Konsolidierung als rolling V1; Image-Hardening auf `node:22-trixie-slim`; OTel-Stack-Bump als Vuln-Fix-Folge. Erster Patch-Release im Repo; Patch-Release-Konvention in `docs/user/releasing.md` §3.1. DoD-Tracking in [`done/plan-0.8.5.md`](../done/plan-0.8.5.md). |
| `0.9.0` | Drift-Smoke + SRS + DASH     | ✅      | Drift-Smoke (Nightly-Workflow `webrtc-drift.yml`, R-12 automatisiert detektiert) + SRS-Lab `examples/srs/` (MVP-36 eingelöst) + DASH-Manifest-Analyse im `@pt9912/stream-analyzer` (NF-12 erfüllt; MVP-37 hochgestuft auf Muss). RAK-56..RAK-59 (Lastenheft `1.1.11` §13.11). DoD-Tracking in [`done/plan-0.9.0.md`](../done/plan-0.9.0.md). |
| `0.9.1` | Drift-Smoke-Robustheit (Patch) | ✅      | Wartungs-Patch nach `0.9.0` ohne eigenen Plan-File: WebRTC-Drift-Smoke robuster gegen reale Browser-Eigenheiten (WHEP-POST aus Node-Kontext, Firefox audio-only, fehlende `transport`-Reports als `[drift-soll]` statt Fail); Spec-Korrekturen in `spec/telemetry-model.md` §3.5.2/§3.5.3; Pfad-Korrekturen nach dem `plan-0.9.0`-Closeout. CHANGELOG-`[0.9.1]`-Block. Kein Lastenheft-Patch. |
| `0.9.5` | Quality-Gates Wave 2 (Patch) | ✅      | Patch-Release am 2026-05-07. Plan in [`done/plan-0.9.5.md`](../done/plan-0.9.5.md). Lieferungen: Benchmark-Smoke (PR-Pfad opt-in mit Beobachtungs-Nightly `benchmark-observation.yml`); Nightly-`benchstat`-Regressionen mit Quarantäne-Mechanik (`benchmark.yml`); sechs Go-Fuzz-Targets + drei TS-Property-Test-Suites via `fast-check` (`make fuzz-check` + Nightly `fuzz.yml`) inkl. Erstfund + Fix `mbpsLinkCapacity=-1` in `apps/api/.../mediamtxclient/mapping.go`; Mutation-Testing mit gremlins (Go) + StrykerJS (TS) als Nightly-Report (`mutation.yml`). Single-Source-Budgets in [`docs/perf/budgets.md`](../../perf/budgets.md); Operator-Doku in [`docs/dev/fuzzing.md`](../../dev/fuzzing.md) und [`docs/dev/mutation-testing.md`](../../dev/mutation-testing.md). Kein Lastenheft-Patch. |
| `0.9.6` | Lastenheft-Konvergenz (Patch) | ✅     | Patch-Release am 2026-05-08. Plan in [`done/plan-0.9.6.md`](../done/plan-0.9.6.md). Lieferungen: fehlende Muss-Repo-Artefakte (`CONTRIBUTING.md`, `SECURITY.md`, `.env.example`, `deploy/`-Struktur), Lastenheft-Patch `1.1.12` (F-7-Status, neue Pflichtdokumente-Kennung `F-131`, NF-13/NF-18 harmonisieren, MVP-19..MVP-26 redaktionell entzerren) und Go-Stdlib-Bump `golang:1.26.3` als Folge der GO-2026-4982/4980/4971/4918-CVE-Fixes (analog `0.8.5`-OTel-Bump). Keine User-Surface- oder Wire-Vertragsänderung. |
| `0.10.0` | CMAF-Analyse | ✅     | Minor-Release am 2026-05-09. Plan in [`done/plan-0.10.0.md`](../done/plan-0.10.0.md). NF-13-Vollumsetzung im Stream-Analyzer-Scope: manifestbasierte HLS-/DASH-CMAF-Signale (`details.cmaf` additiv unter HLS-/DASH-Detail-Objekten, kein neuer `analyzerKind`) plus begrenzte binäre CMAF-Konformitätsprüfung ausgewählter Init-/Media-Segmente (ISO-BMFF-Box-Parser, bounded Segment-Loader). Brand-Allowlist `cmfc`/`cmf2` (Init-`ftyp`) und `cmfs`/`cmff`/`cmfc`/`cmf2` (Media-`styp`); Defaults `maxSegmentBytes=2_000_000`, `maxBinarySegments=6`. Lastenheft-Patch `1.1.13` mit RAK-60..RAK-64 in §13.12. Out of scope: vollständige Segmentset-Abdeckung, Codec-Decoding, Low-Latency-CMAF, Player-Laufzeitpfade. |
| `0.11.0` | Ingest-Gateway / Stream Control | ✅ | Minor-Release am 2026-05-09. Plan archiviert in [`done/plan-0.11.0.md`](../done/plan-0.11.0.md). Variante B (Modul in `apps/api`). Lastenheft-Patch `1.1.14` mit RAK-65..RAK-70 in §13.13 hebt `F-46`..`F-51` und `MVP-38` für den lokalen/lab-nahen Ingest-Control-Pfad auf Release-Muss: CSPRNG-Stream-Keys (nur `key_hash` persistiert; Klartext nur in Create-/Rotate-Antworten), `srt`/`rtmp`-Endpunkte, 1:1-Routing, deterministischer MediaMTX-Konfigurations-Generator + Beispiel-Stack `examples/ingest-control/`, Lifecycle-Hook-Endpoints `POST /api/ingest/hooks/stream-{started,ended}` mit Source-Allowlist `local-smoke`/`mediamtx-hook`, `make smoke-ingest-control` als Lab-Verifikation. Wire-Vertrag in [`spec/backend-api-contract.md`](../../../spec/backend-api-contract.md) §2 + §3.8. **Out of scope:** Multi-Tenant-Control-Plane, KMS/Vault, produktive Auth-Hooks, externe Provisionierung, K8s-Operator, ausgehende produktive Webhook-Zustellung. |
| `0.12.0` | Auth / Token Lifecycle | ✅ | Minor-Release am 2026-05-10. Plan archiviert in [`done/plan-0.12.0.md`](../done/plan-0.12.0.md). F-111..F-113 als zusammenhängender Auth-/Security-Scope: kurzlebige HMAC-SHA-256-signierte Session Tokens (Wire-Skizze in [`spec/backend-api-contract.md`](../../../spec/backend-api-contract.md) §3.9), rotierbare `mtr_pt_*`-Project-Token-Generationen mit V4-SQLite-Migration und persistiertem `grace_until`, Project-gebundene Ingest Policies + §3.9-konformer CORS-Preflight (`204` mit minimaler Signalisierung statt Pre-`0.12.0`-`403`). Lastenheft-Patch `1.1.15` mit RAK-71..RAK-76 in §13.14 plus neunstufige Auth-Fehlerpräzedenz und zehn `auth_*`-Codes. RAK-74-Scope-Cut: `/api/ingest/*` bleibt `0.11.0`-Token-only (RAK-65, Lab-Workflow); R-21 trackt Future-Browser-Konsumenten. **Out of scope:** OAuth/OIDC/SSO, User-/Org-Verwaltung, Admin-UI, KMS/Vault, produktive MediaMTX-/SRS-Auth-Hooks (R-14), Multi-Replica-Issuance-Limiter (R-17), Multi-Key-Rotation-Workflow (R-18), Production-Secret-Backends (R-20). |
| `0.12.1` | Trigger-Re-Eval + Operator-Doku (Patch) | ✅ | Patch-Release am 2026-05-10. Plan archiviert in [`done/plan-0.12.1.md`](../done/plan-0.12.1.md). Patch-Release im Sinne von `releasing.md` §3.1 — keine neue User-Surface, kein Lastenheft-Patch, keine RAK-Matrix. Inhalt: Trigger-Re-Eval pro aktivem R-N-Item (R-5/R-7/R-9/R-10/R-11/R-12/R-13/R-14/R-15/R-16/R-17/R-18/R-20/R-21, alle „nicht ausgelöst"), Operator-Runbook für Multi-Key-Signing-Rotation in `auth.md` §5.3.1 (Soll-Workflow; Code-Pfad in `0.12.5`), Trigger-Schärfung der `OS-1..OS-6` aus `done/plan-0.12.0.md` §10 (OS-1..OS-5 als ⬛ Duplikate in §1.2; OS-6 → **R-22** in §1.1; Done-Plan unverändert). **Out of scope:** alle Adapter-Implementierungen — die wandern in `0.12.5`. |
| `0.12.5` | Auth-/Ingest-Adapter-Minor | ✅ | Minor-Release am 2026-05-11. Plan archiviert in [`done/plan-0.12.5.md`](../done/plan-0.12.5.md). Lastenheft-Patch `1.1.16` mit RAK-77..RAK-82 in §13.15. Inhalt: `MultiKeySigningResolver`-Code-Pfad (R-18 🟢, RAK-78), `SqliteIssuanceRateLimiter` mit Migration V5 (R-17 ⬜ teilweise, RAK-77, Single-Host-Shared-Volume), `AuthSecretBackend`-Port + Vault-Skelett (R-20 ⬜ teilweise, RAK-79), `BrowserIngestPolicy` mit Origin-Pin/CSRF (R-21 🟢, RAK-80, RAK-74-Scope-Cut bei aktivierter Policy aufgehoben), `MediaMTXAuthHookHandler` (R-14 🟢, RAK-81), `OutboundWebhookDispatcher` mit HMAC-SHA-256 + 3-stufiger Exponential-Backoff (R-16 🟢, RAK-82). Fünf neue opt-in Smokes. |
| `0.12.6` | Auth-/Ingest-Folge-Items-Minor | ✅ | Minor-Release am 2026-05-12. Plan archiviert in [`done/plan-0.12.6.md`](../done/plan-0.12.6.md). Lastenheft-Patch `1.1.17` mit RAK-83..RAK-90 in §13.16. Alle neun R-N-Items adressiert: Time-Skew-Persistenz (R-5 🟢, RAK-83, V6 + Dashboard-Pin), `ListSessions`-Bulk-Read-Port (R-7 🟢, RAK-84), Sample-Rate-PPM (R-10 🟢 minus Heuristik, RAK-85, V7 + Banner), SRT-Cursor-Pagination v3 (R-11 🟢, RAK-86), Trivy-Re-Review (R-13 🟢-Wartung, Expiry 2026-11-02), mediamtx-Provisioner mit additivem `?provision=mediamtx` (R-15 🟢, RAK-87), Multi-Host-Limiter via Redis (R-17 🟢 final, RAK-88), Vault-AppRole + KMS-Skeleton (R-20 🟢 final, RAK-89), Origin-/IP-Rate-Limiter (R-22 🟢, RAK-90). Sechs neue opt-in Smokes (`smoke-srt-health-pagination`/`smoke-origin-rate-limit`/`smoke-issuance-multi-host`/`smoke-vault-approle`/`smoke-kms-skeleton`/`smoke-mediamtx-provision`). |
| `0.13.0` | Production / Ops Backends | ✅ | Released 2026-05-12. Plan in [`done/plan-0.13.0.md`](../done/plan-0.13.0.md). Decision-and-Seed-Scope: `MVP-40` Postgres und `MVP-41` Analytics deferred mit Triggern; `MVP-42` Kubernetes-Manifeste optional; `MVP-43` Devcontainer; `MVP-44` Release-Guard mit manueller Freigabe. Lastenheft-Patch `1.1.18` + RAK-91..RAK-95 in §13.17 + Tag `v0.13.0`. |
| `0.14.0` | Ops Backend Follow-up | ✅ | Released 2026-05-12. Plan in [`done/plan-0.14.0.md`](../done/plan-0.14.0.md). Szenario C: K8s-/Devcontainer-/Release-Guard-Hardening; Postgres/Analytics nur Triggerpflege. Lastenheft-Patch `1.1.19` + RAK-96..RAK-100 in §13.18 + Tag `v0.14.0`. |
| `0.15.0` | Product Scope / Analyzer Boundary | ✅ | Released 2026-05-12. Plan in [`done/plan-0.15.0.md`](../done/plan-0.15.0.md). Szenario A: Zielgruppe + Analyzer-Boundary; Tranche 1 erledigt RAK-101 und schärft die Primärzielgruppe. Tranche 2 erledigt RAK-102 und deferred eine externe Analyzer-API bis zu konkretem Konsumenten, Auth-/Rate-Limit-/SSRF-/Retention-/Contract-Nachweis und Folgeplan. Tranche 3 erledigt RAK-103 und deferred Control-Plane ohne POC bis zu Betreiber-/Auth-/Tenant-/Audit-Triggern. Tranche 4 erledigt RAK-104 und empfiehlt HTTP-Range-/Byte-Range-Loader als einzigen kleinen `NF-13`-Folgeslice. Tranche 5 erledigt RAK-105: Postgres bleibt `defer-with-migration-seed`, Analytics bleibt `defer`. Lastenheft-Patch `1.1.20` + RAK-101..RAK-105 in §13.19 + Tag `v0.15.0`. |
| `0.16.0` | Selected Product Slice / Analyzer Range Fetch | ✅ | Released 2026-05-12. Plan in [`done/plan-0.16.0.md`](../done/plan-0.16.0.md). Szenario B: HTTP-Range-/Byte-Range-Loader fuer manifest-referenzierte CMAF-Init-/Media-Segmente. Tranche 0 erledigt RAK-106; Tranche 1 definiert RAK-107..RAK-109 als HLS-CMAF-Byte-Range-Scope mit No-new-public-schema und Fetch-Security-Grenzen; Tranche 2 erledigt RAK-107/RAK-108 mit HLS-Range-Fetch-Code und aktualisierten Contract-Fixtures; Tranche 3 erledigt RAK-109 mit TS-/Doku-/Drift-/Security-Gates; Tranche 4 erledigt RAK-110 mit Versions-Bump, Changelog, Roadmap, Plan-Archiv und Tag `v0.16.0`. Lastenheft-Patch `1.1.21` + RAK-106..RAK-110 in §13.20. |
| `0.17.0` | Hardening / Evidence Review | ✅ | Released 2026-05-13. Plan in [`done/plan-0.17.0.md`](../done/plan-0.17.0.md). Szenario D: Hardening-only. Tranche 0 erledigt RAK-111 mit Import des `0.16.0`-Closeouts, Lastenheft-Patch `1.1.22` und Defer-Matrix. Tranche 1 erledigt RAK-112 mit Evidence Review, `make ts-test`, `make generated-drift-check` und der Entscheidung, Productization/Next Slice/Switch weiter deferred zu halten. Tranche 2 schliesst als Doku-/Defer-Artefakt ohne Code-/Runtime-Aenderung. Tranche 3 erledigt RAK-113/RAK-114 mit No-change-Gate-Nachweis. Tranche 4 erledigt RAK-115 mit Versions-Bump, versionstragendem Test-/Fixture-Asset-Bump, Changelog, Roadmap, Plan-Archiv und Tag `v0.17.0`. |
| `0.18.0` | Offene Risiken / Trigger-Re-Eval | ✅ | Released 2026-05-13. Decision-Closeout in [`done/plan-0.18.0.md`](../done/plan-0.18.0.md): `R-9` bleibt K8s-Smoke-/Scrape-Policy-Trigger, `R-12` bleibt Nightly-detektierter WebRTC-Drift-Pfad ohne Safari-/WebKit-Pflicht, `R-13` bleibt Trivy-`continued` mit Re-Review-Artefakt und `expires` `2026-11-02`. Version `0.18.0`, Changelog, Roadmap, Plan-Archiv und Tag `v0.18.0`; kein Lastenheft-Patch. |
| `0.19.0` | Roadmap-Trigger-Nacharbeit (`MVP-40`, Variante A) | ✅ | Decision-only-Plan archiviert in [`done/plan-0.19.0.md`](../done/plan-0.19.0.md). Postgres bleibt `deferred` mit ADR-0005-Schwellen, CORS-Preflight-Variante A bleibt deferred bis zu echtem Multi-Tenant-/Project-in-URL-Trigger; `apps/analyzer-api` und `apps/control-plane` sind als Decision-Records mit Proceed-/POC-/Defer-Kriterien dokumentiert. Kein Release-Tag, kein Versions-Bump. |
| `0.20.0` | Package Publishing | ✅ | Released 2026-05-13. Plan in [`done/plan-0.20.0.md`](../done/plan-0.20.0.md). Lastenheft-Patch `1.1.23` mit RAK-116..RAK-120 in §13.22. Inhalt: Scope-Migration auf `@pt9912`, GitHub-Packages-Workflow, Release-Doku und Publish von `@pt9912/player-sdk` plus `@pt9912/stream-analyzer`; Apps bleiben private Workspace-Pakete. |
| `0.21.0` | OCI Image Publishing | ✅ | Released 2026-05-13. Plan in [`done/plan-0.21.0.md`](../done/plan-0.21.0.md). Lastenheft-Patch `1.1.24` mit RAK-121..RAK-125 in §13.23. Inhalt: versionierte GHCR-Images für API, Dashboard und Analyzer-Service, Make-Dry-Run, approval-gated Publish, Release-Hook und Rollback-Doku; kein `latest`, kein Production-K8s-Go. |
| `0.22.0` | Quality-Gates Follow-up | ✅ | Released 2026-05-13. Plan in [`done/plan-0.22.0.md`](../done/plan-0.22.0.md). Kein Lastenheft-Patch: Benchmark-Smoke ist nach fünf grünen Beobachtungsläufen PR-blockierend über `make gates`; Benchmark-Nightly läuft hart; Mutation-TS-Filter nutzt `@pt9912/player-sdk`, bleibt aber bis zu echten >70%-Score-Reihen nicht-blockierend. |
| `0.22.1` | devalue-Security-Patch + Nightly-Audit | ✅ | Released 2026-05-17. Plan in [`done/plan-0.22.1.md`](../done/plan-0.22.1.md). Kein Lastenheft-Patch. Inhalt: `pnpm.overrides` hebt `devalue` auf `^5.8.1` (GHSA-77vg-94rm-hx3p, vier Tage nach `0.22.0`-Tag publiziert); neuer `security-audit.yml`-Nightly mit konsolidiertem Auto-Issue (`scripts/open-security-audit-issue.sh`) spiegelt die drei Push-Security-Gates auf 24-h-Cadence; Benchmark-Workflow-Pfadfix (`apps/.tmp/bench/` → `.tmp/bench/`) plus `tee`/`pipefail`-Logging; `Open regression issue`-HEREDOC in `scripts/open-bench-regression-issue.sh` ausgelagert. `extra-gates.md §3.7` dokumentiert den Nightly-Mirror. |
| `0.22.2` | Go-Stdlib-Security-Patch | ✅ | Released 2026-06-03. Plan in [`done/plan-0.22.2.md`](../done/plan-0.22.2.md). Kein Lastenheft-Patch. Inhalt: erster echter Treffer des `security-audit.yml`-Nightly (Issue #3); `golang:1.26.3 → 1.26.4` an sechs Build-/Test-Image-Stellen schließt GO-2026-5039 (`net/textproto`-Error-Echo via `auth.VaultSecretBackend.LoadSigningKeys`) und GO-2026-5037 (`crypto/x509`-Hostname-Parsing via `auth.NewRedisIssuanceRateLimiter`). Plus fünf neue `perl-base`-Trivy-Ignores (CVE-2026-42496/42497/8376/9538/48962) für dashboard/analyzer-service mit dokumentierten `expires`-Terminen und bilingualer README-Split. `make vuln-check` lokal grün. |
| `0.22.3` | Security-/CI-Sammel-Patch | ✅ | Released 2026-06-16. Plan in [`done/plan-0.22.3-webrtc-drift.md`](../done/plan-0.22.3-webrtc-drift.md). Kein Lastenheft-Patch. Inhalt: vier aufeinanderfolgende `security-audit.yml`-Nightly-Treffer behoben — Trivy `0.59.1 → 0.71.0` + `CVE-2026-45447`/`-48959`-Ignores (Issue #4), esbuild `^0.28.1`-`pnpm.overrides` (GHSA-gv7w-rqvm-qjhr, Issue #5), libsqlite3-FTS5-Ignores `CVE-2026-11822`/`-11824` (kein Upstream-Fix, Debian `no-dsa`, `expires` 2026-09-12, Issue #6), vite `^8.0.10 → ^8.0.16` (GHSA-fx2h-pf6j-xcff, Issue #7). Plus GitHub-Actions-Node-24-Runtime-Bump, WebRTC-`getStats()`-Drift-Test-Fix (aus `plan-0.22.3-webrtc-drift`, dort kein eigener Tag) und ein Flaky-CORS-Preflight-Test-Fix (paralleler `http.DefaultClient`-Pool → `srv.Client()`, verifiziert via `make api-race`). `make audit-ts` + `make image-scan` lokal grün. |

`0.1.x` ist seit Lastenheft-Patch `1.1.0` in drei Sub-Releases
geschnitten (Variante 2-A); RAK-1..RAK-10 sind dort verteilt.

DoD für die erste Phase ist über **AK-1..AK-11** abgedeckt
(Lastenheft-übergreifend, nicht Release-spezifisch). Detaillierter
Lieferstand pro Tranche steht in den drei `0.1.x`-Plan-Dokumenten;
Release-Vorgehen in [`docs/user/releasing.md`](../../user/releasing.md).

---

## 4. Folge-ADRs

Aus `docs/adr/0001-backend-stack.md` §8 erwartete Folge-ADRs.
Die zugehörigen Risiken stehen in `docs/planning/in-progress/risks-backlog.md`;
erledigte oder obsolete Einträge sind nach §7-Wartungsregel entfernt
(beschlossene ADRs siehe [`docs/adr/`](../../adr/)).

| Erwartete ADR / Decision-Track                         | Trigger-Release                            | Begründung                                                                                                                                                                                                                                                                                                                                        |
| ------------------------------------------------------- | ------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Postgres als produktionsnaher Store (**MVP-40**)        | offen, zuletzt geprüft in `0.19.0`          | ADR-0005 bleibt gueltig: Postgres wird nur bei Multi-Replica-Store, verbindlichem Recovery-SLO oder Retention-/Read-Last-Schwelle reaktiviert. Decision-Record: [`done/plan-0.19.0.md`](../done/plan-0.19.0.md) §3.1. |
| Strengere CORS-Preflight-Project-Isolation (Variante A) | offen, zuletzt geprüft in `0.19.0`          | Variante B bleibt Standard; Variante A braucht echten Multi-Tenant-Browser-Client, Project-in-URL-Kontrakt oder Security-/Audit-Befund, der Preflight-Isolation vor Request erzwingt. Decision-Record: [`done/plan-0.19.0.md`](../done/plan-0.19.0.md) §3.2. |
| Decision-Tracks ohne eigene R-N-ID `RAK-102` / `RAK-103` | offen, zuletzt geprüft in `0.19.0`          | Externe `apps/analyzer-api` (`RAK-102`) und `apps/control-plane` (`RAK-103`) bleiben triggerbasiert deferred; Proceed/POC braucht externen Konsumenten bzw. Betreiber-/Tenant-/Audit-Bedarf plus Owner und Folgeplan. Decision-Record: [`done/plan-0.19.0.md`](../done/plan-0.19.0.md) §3.3. |

Neue Folge-ADRs werden hier ergänzt, sobald der Bedarf entsteht oder
ein Issue darauf hinweist.

---

## 5. Offene Entscheidungen

Verbleibende Lastenheft-`OE-X`; aufgelöste Einträge sind nach §7-Wartungsregel entfernt. Derzeit keine offenen `OE-X` in der Roadmap — historische `OE-X` sind im [Lastenheft](../../../spec/lastenheft.md) als `resolved` geführt.

---

## 6. Lessons-learned aus Spike (Verdichtung)

Vollständige Notizen in `docs/spike/backend-stack-results.md`. Hier nur
die für `0.1.0`+ relevanten Punkte:

- **Hexagon ohne DI-Container-Druck**: Go braucht keine
  Annotation-Magie; `var _ Interface = (*Impl)(nil)`-Compile-Time-Checks
  pro Adapter reichen. Beibehalten.
- **Test-Stack einheitlich**: `testing` + `httptest` deckt Unit und
  Integration ab. Keine externen Test-Frameworks erforderlich.
- **Linting**: `golangci-lint` mit Default-Lintern
  (`govet`, `errcheck`, `staticcheck`, `unused`, `ineffassign`).
  `make lint` als Soll-Target im Dockerfile.
- **Docker-only-Workflow**: alle Build-/Test-/Lint-Schritte über
  `docker build --target ...`. Lokales Go ist optional. Pattern aus
  `docs/planning/done/plan-spike.md` §14.11 wird beibehalten.
- **CI-Artifacts** (SP-41 Lessons-learned): Test-Results,
  Coverage-Reports, Lint-Reports beim CI-Setup hochladen — Pattern
  analog zu `d-migrate/.github/workflows/build.yml`.
- **Multi-Modul-Aufteilung erst on demand**: bei wachsender
  Codebase `apps/api/` per `go.work` oder Sub-Modul-Splits aufteilen.
  Im Spike bewusst Single-Modul für Übersicht.

---

## 7. Wartung dieses Dokuments

- Statusspalten in §2 und §3 nach jedem abgeschlossenen Schritt
  bzw. neuen Release-Tag aktualisieren (✅).
- Nach jedem neuen Folge-ADR Eintrag in §4 ergänzen oder erledigte
  ADRs aus §4 herausnehmen.
- Nach jeder gelösten offenen Entscheidung Eintrag in §5 entfernen
  und (falls strukturell) in das Lastenheft übernehmen.
- §1 Aktueller Stand wird nach jedem signifikanten Meilenstein neu
  geschrieben (nicht inkrementell — die Liste bleibt kurz).

### 7.1 Source-of-Truth-Konvention bei Lastenheft-Widersprüchen

Lastenheft ist die normative Anforderungsquelle. Bei **interner**
Inkonsistenz zwischen einer F-Kennung (Anforderungs-Detail in §7) und
einer MVP-Kennung (Release-Prioritäts-Klassifikation in §12) gewinnt
**keine** Seite automatisch:

1. Plan-Dokumente (`plan-X.Y.Z.md`) markieren betroffene DoD-Items mit
   Status `[!]` (statt `[ ]` oder `[x]`) und beschreiben die
   Inkonsistenz in einem kurzen Hinweis.
2. Auflösung erfolgt durch einen **Lastenheft-Patch**: betroffene
   F- oder MVP-Kennung wird angepasst, Lastenheft-Header-Version
   bekommt einen Patch-Level-Bump (`1.0.0` → `1.0.1` → `1.0.2` …).
3. Der Patch wird im jeweiligen Plan-Dokument unter der dortigen
   Tranche „Lastenheft-Patches" (z. B. `plan-0.1.0.md` Tranche 0c)
   getrackt — mit Verweis auf die geänderten F-/MVP-Kennungen und
   den Begründungs-Pfad (Code-Review-Finding, ADR, Diskussion).
4. Bezug-Listen in den Soll-Dokumenten (`architecture.md`,
   `plan-X.Y.Z.md`, `README.md`) werden auf die neue Patch-Version
   gepinnt; historische Verweise (frühere Plan-Stände, ADRs,
   Spike-Doku) bleiben auf der ursprünglichen Version.

Diese Konvention verhindert, dass der Plan eigenmächtig zugunsten
einer der widersprüchlichen Quellen entscheidet und damit eine
normative Anforderung des Lastenhefts unterläuft.
