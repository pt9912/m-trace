# Implementation Plan — `0.8.5` (Quality-Gates Wave 1: Security + Generated-Artifact-Drift)

> **Status**: ✅ released am 2026-05-07 (Tranchen 0..3 abgeschlossen:
> Plan-Aktivierung, Security-Gates, Generated-Drift-Gate inkl.
> Migrations-Konsolidierung und Image-Hardening, Closeout mit
> Versions-Bump 0.8.0 → 0.8.5 und Tag `v0.8.5`). Plan liegt nach dem
> Closeout-`git mv` unter `docs/planning/done/`. Vorgänger `v0.8.0`
> ist released (Tag `v0.8.0` auf `8df263a`, GitHub-Release
> veröffentlicht). Container-Scanner-Wahl in Tranche 0: **Trivy**
> (Default aus §0.4 — etablierter, breitere Image-Format-
> Unterstützung, bessere Default-Policy als Grype).
>
> **Release-Typ**: erstmaliger **Patch-Release** im m-trace-Repo
> (siehe §0.6). Alle bisherigen Releases (`0.1.0`..`0.8.0`) waren
> Minor; `0.8.5` führt die Patch-Release-Konvention ein und
> dokumentiert sie in `docs/user/releasing.md` §3.
>
> **Lastenheft-Status**: kein Lastenheft-Patch nötig. Quality-Gates
> sind keine Lastenheft-Anforderung mit eigener RAK-Pflicht; die
> Plan-DoD-Items sind ausreichend. Wenn künftig ein Gate als RAK
> verankert werden soll (z. B. `make security-gates` als Pflicht-
> Surface), wird ein eigenständiger Lastenheft-Patch in einem
> Folge-Plan aufgenommen.
>
> **Bezug**: [`extra-gates.md`](../in-progress/extra-gates.md) §3.1
> (`govulncheck` + Container-Scan) und §3.4 (Generated-Artifact-
> Drift-Gate) — Master-Backlog für die sechs vorgeschlagenen
> Quality-Gates; `0.8.5` liefert die deterministisch-schnellen
> Wave-1-Gates, `0.9.5` (Folge-Patch nach `0.9.0`) liefert die
> statistisch-langlaufenden Wave-2-Gates;
> [`done/plan-0.2.0.md`](./plan-0.2.0.md) §6 (Pack-Smoke +
> Public-API-Snapshot — bestehende deterministische Drift-Gates,
> auf denen das Generated-Artifact-Drift-Gate aufsetzt);
> [`docs/user/releasing.md`](../../user/releasing.md) §3 (Release-
> Commit und Tag — wird im Closeout um Patch-Release-Konvention
> erweitert).
>
> **Nachfolger**: [`plan-0.9.0.md`](./plan-0.9.0.md) (Drift-Smoke +
> SRS-Lab + DASH-Analyse, released am 2026-05-07) und
> [`plan-0.9.5.md`](../done/plan-0.9.5.md) (Quality-Gates Wave 2,
> seit 2026-05-07 in Arbeit).

## 0. Konvention

DoD-Checkboxen tracken den Lieferstand analog
[`plan-0.1.0.md`](./plan-0.1.0.md) §0:

- `[x]` ausgeliefert mit Commit-Hash.
- `[ ]` offen.
- `[!]` blockiert durch Lastenheft-Inkonsistenz oder offene
  Entscheidung.
- 🟡 in Arbeit.

Scope-Grenze: dieser Plan liefert zwei deterministische,
PR-blockierende Quality-Gates aus `extra-gates.md`. Beide laufen
schnell (< 30 s zusätzlich pro PR-Run) und benötigen keine
Baseline-Daten. Wave 2 (Benchmark-Smoke + Nightly-benchstat +
Fuzzing + Mutation Testing) ist explizit out-of-scope — siehe
`plan-0.9.5.md`.

### 0.1 Vorgänger-Gate

Voraussetzungen, bevor dieser Plan in `in-progress/` gezogen werden
kann:

- `0.8.0` ist released (Tag `v0.8.0`); produktive WebRTC-Telemetrie
  ist live, sechs `mtrace_webrtc_*`-Counter exportiert.
- Toolchain ist non-EOL: Go 1.26 + golangci-lint v2.12.1 + Node 22
  LTS aus `0.7.0` Tranche 0 sind weiterhin aktuell. Bei Bedarf
  eigene Toolchain-Hardening-Sub-Tranche.

### 0.2 Out-of-Scope-Klauseln (durchgängig)

- Kein Benchmark-Smoke, kein Nightly-benchstat, kein Fuzzing, kein
  Mutation Testing — alle vier sind Wave 2 (`plan-0.9.5.md`).
- Keine neuen Lastenheft-Pflicht-RAK. Quality-Gates sind
  Plan-DoD-Items, keine Akzeptanzkriterien.
- Kein neuer Wire-Vertrag, keine API-Surface-Änderung. `0.8.5` ist
  eine reine CI-/Tooling-Lieferung.
- Keine Multi-Tenant-Erweiterung, keine produktiven WebRTC-/SRT-/
  DASH-Themen — diese bleiben bei `0.9.0`/`0.9.5` und späteren
  Phasen.
- Kein Cardinality-/Forbidden-Label-Spec-Patch in
  `spec/telemetry-model.md` §3.1. Quality-Gates verändern keine
  Telemetrie-Surface.

### 0.3 Sequenzierung und harte Gates

1. Tranche 0 (Plan-Aktivierung) ist Pflicht vor Tranche 1+2.
2. Tranche 1 (Security-Gates) und Tranche 2 (Generated-Drift) sind
   **unabhängig** — Reihenfolge richtet sich nach Operator-Präferenz.
   Default-Empfehlung: Generated-Drift zuerst (kein externes Tool
   nötig, keine Tool-Pinning-Entscheidung), dann Security-Gates.
3. Tranche 3 (Closeout) erst nach Tranche 1+2.
4. Versions-Bump im Closeout: 0.8.0 → 0.8.5. Test-Fixtures, die
   Versions-Strings hartkodieren (siehe `0.7.0`/`0.8.0`-Closeout-
   Bulk-Fix-Pfad), werden mitgezogen, weil `version.test.ts` und
   die Analyzer-Contract-Fixtures sonst auf `0.8.0` festklemmen.

### 0.4 Implementierungsleitplanken

**Security-Gates (Tranche 1)**: Bevorzugte Form ist ein neues
`make vuln-check`-Target (Go-Dependencies via `govulncheck ./...`)
plus ein `make image-scan`-Target (Container-Scanner — Trivy oder
Grype, Auswahl in Tranche 0). Beide Targets bekommen ein
Wrapper-Target `make security-gates`, das beide bündelt. CI-Stage
kommt in `.github/workflows/build.yml` als zusätzlicher Job
(parallel zu `make gates`, damit der Pfad nicht serialisiert).

**Generated-Drift-Gate (Tranche 2)**: Bevorzugte Form ist ein
neues `make generated-drift-check`-Target, das die
Generierungs-/Sync-Targets ausführt und `git diff --exit-code`
auf die erwarteten Pfade aufruft. Geprüfte Artefakte:
`apps/api/internal/storage/schema.sql` (aus `schema.yaml`),
`apps/api/adapters/driven/streamanalyzer/testdata/contract-*.json`
(aus `spec/contract-fixtures/analyzer/*.json` via
`make sync-contract-fixtures`), `packages/player-sdk/scripts/
public-api.snapshot.txt` (aus `index.ts` via existierendem
`check-public-api.mjs`). `contracts/sdk-compat.json` wird **nicht**
generiert, sondern manuell gepflegt — bleibt out-of-scope für
diesen Drift-Check.

### 0.5 Test-Fixture-Versions-Drift bei Patch-Release

Tests in `0.7.0`/`0.8.0` haben SDK-/Analyzer-Versions-Strings
(`"0.8.0"`) hartkodiert. Diese Strings sind via
`xargs sed -i 's/"0\.8\.0"/"0.8.5"/g'` in der Closeout-Tranche
mitzuziehen; betroffene Files sind in `done/plan-0.8.0.md`
Tranche 5 dokumentiert. Folge-Aufgabe (Backlog, nicht in `0.8.5`):
Tests sollten die Version aus `package.json` lesen statt
hartzukodieren — das wäre ein eigenständiger Plan-Punkt.

### 0.6 Patch-Release-Konvention (erstmals im Repo)

`0.8.5` ist der erste Patch-Release im m-trace-Repo. Die Konvention
wird im Closeout in `docs/user/releasing.md` §3 verankert:

- **Patch-Release `0.X.Y`**: Quality-/Security-/Doc-Gates,
  CI-Tooling, Bugfixes ohne neue User-Surface. Kein Lastenheft-
  Patch nötig. Versions-Bump in `package.json`/main.go/version.ts/
  pack-smoke/sdk-compat.json (für Drift-Konsistenz) plus
  Test-Fixture-Bulk-Fix.
- **Minor-Release `0.X.0`**: Feature-Release mit neuer User-
  Surface. Lastenheft-Patch (`1.1.X`) mit neuen RAK-Anforderungen
  ist Pflicht.
- **Major-Release `1.0.0`**: erstmaliges öffentliches Public-API-
  Versprechen; aktuell Folge-ADR-Thema (`docs/adr/`).

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Status |
| ------- | ------ | ------ |
| 0 | Plan-Aktivierung (`open/` → `in-progress/`) + Tool-Pinning-Entscheidung (Trivy für Container-Scan; Toolchain-Check ohne Bump) | ✅ |
| 1 | Security-Gates: `make vuln-check` (govulncheck) + `make audit-ts` (`pnpm audit --audit-level high`) + `make image-scan` (Trivy) + Wrapper `make security-gates`; CI-Stage parallel zu `make gates` | ✅ |
| 2 | Generated-Artifact-Drift-Gate: Sub-2a Migrations-Konsolidierung (V2..V5 in rolling V1, Composite-FK in `schema.yaml`); Sub-2b `make generated-drift-check` (Schema-DDL, Contract-Fixtures, Public-API-Snapshot); CI-Stage in `make gates` | ✅ |
| 3 | Release-Doku, Patch-Release-Konvention in `releasing.md` §3, Versions-Bump 0.8.0 → 0.8.5, Plan nach `done/`, Tag `v0.8.5` | ✅ |

---

## 1a. Tranche 0 — Plan-Aktivierung + Tool-Pinning

Bezug: keine RAK direkt; Wartungs-/Hygiene-Tranche analog
`done/plan-0.8.0.md` §1a.

DoD:

- [x] Plan-Skelett von `docs/planning/open/plan-0.8.5.md` nach
  `docs/planning/in-progress/plan-0.8.5.md` verschoben (Status
  `⬜ → 🟡`); Cross-Refs in `roadmap.md` §1.2/§3 nachgezogen
  (Statusspalte 0.8.5 von ⬜ auf 🟡); `extra-gates.md`-Header
  ist seit Plan-Anlage-Commit konsistent (Master-Backlog-Form mit
  Verweis auf 0.8.5/0.9.5).
- [x] Container-Scanner-Wahl: **Trivy**. Begründung: etabliert in der
  Open-Source-Sicherheits-Pipeline, breitere Image-Format-Unterstützung
  als Grype (zusätzlich Filesystem-/SBOM-Scans), bessere
  Out-of-the-box-Policy für CRITICAL/HIGH/MEDIUM-Klassifizierung.
  Aufruf-Form in Tranche 1: `aquasec/trivy:0.X.Y` Container-Image
  mit gepinnter Version, kein lokaler Binary-Pin nötig.
- [x] Toolchain-Bump-Check: keine Anpassung nötig. Go (`1.26`),
  golangci-lint (`v2.12.1`), Node (`22 LTS`), pnpm sind seit
  `0.7.0` Tranche 0 (Commits `ccf68b1` + `8bfad21`) aktuell und
  non-EOL. Race-Detector-Stage ist in `make gates` enthalten.
  `govulncheck` selbst wird in Tranche 1 mit eigener gepinnter
  Version (`go install golang.org/x/vuln/cmd/govulncheck@vX.Y.Z`)
  installiert; keine Auswirkung auf die Repo-Toolchain.

---

## 2. Tranche 1 — Security-Gates (`govulncheck` + `pnpm audit` + Container-Scan)

Bezug: `extra-gates.md` §3.1.

Ziel: Bekannte CVEs in Go- **und** TypeScript-Dependencies sowie in
den Runtime-Images werden früh erkannt. Alle drei Gates sind
PR-blockierend in CI (parallel zu `make gates`, nicht
serialisiert). `extra-gates.md` §3.1 nannte ursprünglich nur Go +
Container; der `pnpm audit`-Gate ist bewusst Bestandteil der
gleichen Wave, weil ein offener npm-CVE-Pfad sonst die Wirkung der
Go-/Image-Gates relativiert.

DoD:

- [x] `make vuln-check` im Root-`Makefile` (Variable
  `GOVULNCHECK_VERSION ?= v1.1.4`): startet einen `golang:1.26`-
  Container, installiert `govulncheck` aus `golang.org/x/vuln/cmd/
  govulncheck@$(GOVULNCHECK_VERSION)` und ruft es gegen `./...` im
  `apps/api`-Modul auf. Pinning ist Default-Override-fähig
  (`make vuln-check GOVULNCHECK_VERSION=vX.Y.Z`).
- [x] `make audit-ts` im Root-`Makefile`: ruft
  `pnpm audit --audit-level high` gegen den gesamten pnpm-
  Workspace auf (`apps/dashboard`, `apps/analyzer-service`,
  `packages/*`). Schwelle = `high`; `moderate`/`low` werden
  berichtet, brechen aber den Lauf nicht. Pendant zu
  `vuln-check` für die TypeScript-Seite — ohne diesen Gate würde
  eine bekannte CVE in einer Frontend-/SDK-Dependency die
  Security-Wave bestehen.
- [x] `make image-scan` im Root-`Makefile` (Variable
  `TRIVY_IMAGE ?= aquasec/trivy:0.59.1`) baut die drei Runtime-
  Images (`apps/api` `runtime`-Stage als `mtrace-api:scan`,
  `apps/dashboard`/`apps/analyzer-service` jeweils Default-Stage
  nach `pnpm run build`) im selben Lauf und scannt sie sequentiell
  mit dem gepinnten Trivy-Image. Cache-Verzeichnis liegt unter
  `.security/.trivy-cache`, damit lokale Wiederholungen nicht
  jedes Mal die Vuln-DB neu laden müssen.
- [x] Scan-Policy: `--severity CRITICAL,HIGH --exit-code 1` für
  alle drei Image-Scans. `MEDIUM` wird in der Trivy-Default-Output-
  Form mitgemeldet (informativ, nicht blockierend), weil
  `CRITICAL,HIGH` als Severity-Filter nur die Exit-Code-Logik steuert.
- [x] `.security/vulnignore.yaml` mit Schema-Header und
  Begründungs-/`expires`-Pflicht angelegt; initial leer
  (`trivy.ignore: []` und `govulncheck.ignore: []`). Die Wartungs-
  Mechanik (`expires`-Check, automatische Erinnerung) ist als
  Folge-Item für `plan-0.9.5` notiert.
- [x] CI-Workflow `.github/workflows/build.yml` um zweiten Job
  `security` erweitert (parallel zu `build`, eigene
  `permissions: contents: read`); führt `make vuln-check`,
  `make audit-ts` und `make image-scan` aus, lädt den
  `.security/.trivy-cache`-Pfad bei jedem Lauf als Workflow-
  Artefakt hoch (Retention 7 Tage).
  Maschinenlesbare SARIF-Ausgabe ist als Folge-Item dokumentiert
  (kommt mit `tranche-2-Erweiterung` oder im `0.9.5`-Closeout —
  Trivy unterstützt `--format sarif` out-of-the-box; aktuell
  blockiert nur die Tabellen-Default-Form, was für die erste
  Auslieferung ausreicht).
- [x] Wrapper-Target
  `make security-gates: vuln-check audit-ts image-scan` bündelt
  die drei Targets sequentiell; Help-Text-Einträge für alle vier
  neuen Targets (`vuln-check`, `audit-ts`, `image-scan`,
  `security-gates`) in `make help`.
- [x] PR-blockierend in CI: der `security`-Job läuft in
  `pull_request` und `push: branches: main` analog zum `build`-
  Job; ein fehlschlagender CRITICAL/HIGH-Befund stoppt den PR.
  Maintenance-Release-Branches können den Job per `if`-Bedingung
  filtern (Folge-Item, falls Bedarf entsteht — aktuell läuft nur
  ein Branch (`main`), keine Maintenance-Branches im Repo).
- [x] Image-Hardening (Tranche-1-Closeout, getriggert vom ersten
  CI-Lauf des `image-scan`-Targets):
  - Dashboard- und Analyzer-Service-Dockerfile beide auf
    `node:22-trixie-slim` (Debian 13) angehoben — eliminiert
    5 CVEs gegenüber `bookworm-slim`. Analyzer-Service vorher
    `node:22-alpine`; Wechsel von musl zurück zu glibc, weil
    musl bei multi-threaded Workloads (libuv-Worker-Pool,
    V8-GC/JIT) gegenüber glibc spürbar pessimisiert ist.
  - `pnpm deploy --prod --legacy /deploy` als build-Stage in
    beiden Service-Dockerfiles (Dashboard war vorher
    `COPY --from=build /app /app` mit allen dev-deps): schneidet
    `vitest`, `tsup`, `typescript-eslint`, `sucrase` und ihre
    `picomatch`-Range-Deklarationen ab.
  - Runtime-Stages entfernen explizit das gebündelte
    `npm`-Tooling (`rm -rf /usr/local/lib/node_modules/npm
    /usr/local/bin/npm /usr/local/bin/npx`), weil npm im Node-
    Base-Image eine eigene `picomatch@4.0.3`-Kopie mitführt
    (CVE-2026-33671); die Runtime startet direkt mit `node ...`
    und braucht npm nicht.
  - `pnpm.overrides`-Block in Root-`package.json` hebt
    `picomatch` workspace-weit auf `^4.0.4` (Belt-and-
    Suspenders gegenüber dem npm-Removal).
  - `.security/vulnignore.yaml` mit drei dokumentierten Trivy-
    Ignores für die verbleibenden Trixie-OS-CVEs ohne Upstream-
    Fix (`CVE-2025-69720`, `CVE-2026-29111`, `CVE-2026-4878`),
    je mit Begründung, Scope (`mtrace-dashboard,
    mtrace-analyzer-service`) und 90-Tage-`expires`.
  - `scripts/render-trivyignore.sh` rendert
    `.security/.trivyignore` aus `vulnignore.yaml` (Audit-
    Source-of-Truth bleibt YAML; Trivy konsumiert das Plain-
    Text-Format via `--ignorefile`); bricht ab, sobald ein
    `expires` überschritten ist (Wartungsregel).
  - `make image-scan` ruft den Generator vor jedem Trivy-Lauf
    auf; `.trivyignore` ist gitignored.
  - R-13 in `risks-backlog.md` als Folge-Risiko: Trixie-Point-
    Release-Re-Review oder Distroless-Wechsel vor 1.0.

---

## 3. Tranche 2 — Generated-Artifact-Drift-Gate

Bezug: `extra-gates.md` §3.4.

Ziel: Generierte Artefakte bleiben synchron zu ihren Quellen.
Drift bricht den PR mit klarer Regenerierungs-Anweisung.

> Vorab-Befund (vor Implementierung des Gates): `make schema-generate`
> erzeugte ein V1__m_trace.sql, das gegenüber dem committeten Stand
> 50 Zeilen Drift hatte — die historischen V2..V5-Migrationen waren
> inkrementell hinzugefügt worden, statt in V1 zurückzuführen. Da
> noch kein Production-State existiert, wurden V2..V5 in der rolling
> V1 konsolidiert (s. Sub-Tranche 2a). Damit ist `schema.yaml` Single-
> Source-of-Truth; der Drift-Gate kann V1 als generiertes Artefakt
> prüfen, ohne dass inkrementelle Migrationen bewusst auseinander
> driften.

### Sub-Tranche 2a — Migrations-Konsolidierung (Vorbedingung)

DoD:

- [x] Composite-FK `stream_session_boundaries → stream_sessions
  (project_id, session_id) ON DELETE CASCADE` in `schema.yaml` als
  `constraints[]`-Eintrag mit `type: foreign_key` ergänzt — vorher
  lebte er nur in der inkrementellen V3-Migration und wäre beim
  Konsolidieren verlorengegangen. Verifiziert via
  `d-migrate schema reverse` über eine V1+V2+V3+V4+V5-DB: Diff zur
  vorhandenen `schema.yaml` zeigte ausschließlich diesen FK.
- [x] V1__m_trace.sql aus aktualisierter `schema.yaml` regeneriert
  (`make schema-generate`); enthält alle 5 Tabellen plus den neu
  ergänzten Composite-FK.
- [x] V2..V5-Files (`V2__project_session_pk.sql`,
  `V3__session_boundaries.sql`, `V4__session_end_source.sql`,
  `V5__srt_health_samples.sql`) gelöscht (`git rm`); kein
  Production-State erreicht, also legitim.
- [x] Apply-Runner (`apps/api/internal/storage/migrate.go`)
  funktional unverändert: ignoriert applied-Versionen ohne File,
  d. h. bestehende Dev-DBs mit V2..V5 als applied bleiben
  funktional. Ein Fresh-Start läuft genau **eine** Migration an
  (V1 mit Endzustand).
- [x] `migrate_internal_test.go::TestOpen_FreshStart` von
  `len(rows) == 5` auf `len(rows) == 1` angepasst; Kommentar
  dokumentiert die Konsolidierung und den Reset-Pfad-Hinweis.
- [x] `docs/adr/0002-persistence-store.md` §8.2 ergänzt: rolling
  V1 als Pre-Production-Privileg, Hand-pflege erst ab erstem
  Production-Stand, Hinweis auf die historischen V2..V5.
- [x] `make api-test` grün; `make schema-validate` grün.

### Sub-Tranche 2b — Drift-Gate

DoD:

- [x] `make generated-drift-check`-Target im Root-`Makefile`:
  ruft `make schema-generate`, `make sync-contract-fixtures` und
  `pnpm --filter @npm9912/player-sdk exec node scripts/check-public-api.mjs`
  auf und führt anschließend `git diff --exit-code HEAD --` auf
  die generierten Pfade aus. `HEAD` (nicht Index) ist der
  Vergleichspunkt, damit ein vorzeitiges `git add` einen Drift
  nicht maskiert.
- [x] Geprüfte Artefakte:
  `apps/api/internal/storage/migrations/V1__m_trace.sql` (aus
  `schema.yaml`),
  `apps/api/adapters/driven/streamanalyzer/testdata/contract-success-master.json`
  und `contract-error-fetch-blocked.json` (aus
  `spec/contract-fixtures/analyzer/*.json`),
  `apps/api/adapters/driven/srt/mediamtxclient/testdata/mediamtx-srtconns-list.json`
  (aus `spec/contract-fixtures/srt/`),
  `apps/api/adapters/driving/http/testdata/srt-health-detail.json`
  (aus `spec/contract-fixtures/api/`),
  `packages/player-sdk/scripts/public-api.snapshot.txt` (aus
  `packages/player-sdk/src/index.ts`, Verifikation read-only via
  `check-public-api.mjs`).
- [x] Fehlertext nennt den konkreten Regenerier-Befehl pro Pfad
  (z. B. „--> run: make schema-generate") plus Hinweis, das Target
  nach dem Fix erneut auszuführen.
- [x] Gate läuft ohne Netzwerk, sobald die `d-migrate`- und
  `golang:1.26`-Images lokal gepullt sind (CI-Cache trägt das mit).
- [x] In `make gates` zwischen `schema-validate` und
  `sdk-pack-smoke` aufgenommen — deterministisch und schnell
  genug, keine Parallelisierung nötig.
- [x] CI-Workflow erbt die `make gates`-Erweiterung über den
  bestehenden `build`-Job; kein neuer Job nötig.

---

## 4. Tranche 3 — Release-Doku, Patch-Release-Konvention, Closeout

Bezug: §0.6 (Patch-Release-Konvention); `docs/user/releasing.md`;
`README.md`; `roadmap.md`.

Ziel: `0.8.5` ist auffindbar dokumentiert, Patch-Release-Konvention
in `releasing.md` §3 verankert, Tag `v0.8.5` gesetzt.

DoD:

- [x] `docs/user/releasing.md` §3.1 „Patch-Release-Konvention
  (`0.X.Y`, ab `0.8.5`)" mit Tabelle Patch/Minor/Major plus Hinweis,
  dass Patch-Releases keinen Lastenheft-Patch brauchen und keine
  RAK-Verifikationsmatrix führen — bereits mit Plan-Anlage-Commit
  `69c3621` ausgeliefert (vor Tranche-0-Aktivierung verfasst, weil
  die Konvention auch §0.6 dieses Plans speist).
- [x] `README.md` Status-Block erwähnt `0.8.5` als Patch-Release
  mit Quality-Gates Wave 1 (Security + Generated-Drift); kein
  Feature-Release-Hinweis (Closeout-Commit).
- [x] Versions-Bump 0.8.0 → 0.8.5 in allen package.json (root,
  apps, packages) plus `apps/api/cmd/api/main.go` `serviceVersion`,
  `packages/player-sdk/src/version.ts`, `packages/player-sdk/
  scripts/pack-smoke.mjs` `expectedVersion`,
  `contracts/sdk-compat.json` `sdk_version` und allen Test-
  Fixtures, die Versions-Strings hartkodieren (Bulk-Fix per
  `xargs sed -i 's/"0\.8\.0"/"0.8.5"/g'` über die `_test.go`/
  `.test.ts`-Files plus `spec/contract-fixtures/analyzer/*.json`;
  zusätzlich Test-Helper `apps/api/adapters/driven/persistence/
  contract/contract.go` und drei Fehlertext-Strings, die der
  sed nicht erfasste, weil sie unquoted sind) (Closeout-Commit).
- [x] CHANGELOG: [Unreleased]-Block in `[0.8.5] - 2026-05-07`
  umgewandelt; neuer leerer [Unreleased]-Block obenauf
  (Closeout-Commit).
- [x] `./scripts/verify-doc-refs.sh` (`make docs-check`) grün vor
  Closeout-Commit; `make security-gates` grün vor Closeout-Commit;
  `make gates` grün **nach** Closeout-Commit (analog `plan-0.8.0.md`
  §7 Release-Gate-Fix: das `generated-drift-check`-Target vergleicht
  Working-Tree gegen HEAD und wertet einen noch nicht committeten
  Versions-Bump als Drift, obwohl Quelle und generierte Kopie
  synchron auf `0.8.5` sind; nach dem Commit ist `git diff HEAD`
  clean und das Gate grün).
- [x] `plan-0.8.5.md` von `docs/planning/in-progress/` nach
  `docs/planning/done/` verschoben (`git mv`); alle relativen
  Cross-Refs angepasst; Roadmap §3 zeigt `0.8.5` ✅
  (Closeout-Commit).
- [x] Tag `v0.8.5` annotiert; Push opt-in (User-Bestätigung);
  GitHub-Release mit CHANGELOG-`[0.8.5]`-Block als Notes-Body
  (Closeout-Commit).

---

## 5. Wartung

- Beim Auslagern eines `[ ]`-Items in einen Commit: `[ ]` → `[x]`,
  Commit-Hash anhängen (analog `done/plan-0.8.0.md` §7).
- Patch-Release-Versions-Drift: wenn ein Test-Fixture nach
  `xargs sed`-Bulk-Fix immer noch auf `"0.8.0"` zeigt, prüfen ob
  der Fixture-Pfad in der grep-Liste fehlt; Folge-Backlog-Item:
  Tests-aus-package.json-lesen (separater Plan, nicht `0.8.5`).
- `extra-gates.md` ist Master-Backlog. Wenn Tranche 1 oder 2 die
  DoD verschärft (z. B. neue Forbidden-Patterns), bleibt der
  Master-Backlog Quelle der Wahrheit; `0.8.5`-Plan zitiert ihn,
  führt aber keine neuen Backlog-Items.
