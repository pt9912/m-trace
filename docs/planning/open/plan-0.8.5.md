# Implementation Plan — `0.8.5` (Quality-Gates Wave 1: Security + Generated-Artifact-Drift)

> **Status**: ⬜ geplant (Plan-Skelett, liegt unter
> `docs/planning/open/`). Vorgänger `v0.8.0` ist released
> (Tag `v0.8.0`). Tranche 0 aktiviert die Phase. Plan wandert dann
> atomar nach `docs/planning/in-progress/`.
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
> **Bezug**: [`extra-gates.md`](./extra-gates.md) §3.1
> (`govulncheck` + Container-Scan) und §3.4 (Generated-Artifact-
> Drift-Gate) — Master-Backlog für die sechs vorgeschlagenen
> Quality-Gates; `0.8.5` liefert die deterministisch-schnellen
> Wave-1-Gates, `0.9.5` (Folge-Patch nach `0.9.0`) liefert die
> statistisch-langlaufenden Wave-2-Gates;
> [`done/plan-0.2.0.md`](../done/plan-0.2.0.md) §6 (Pack-Smoke +
> Public-API-Snapshot — bestehende deterministische Drift-Gates,
> auf denen das Generated-Artifact-Drift-Gate aufsetzt);
> [`docs/user/releasing.md`](../../user/releasing.md) §3 (Release-
> Commit und Tag — wird im Closeout um Patch-Release-Konvention
> erweitert).
>
> **Nachfolger**: [`plan-0.9.0.md`](./plan-0.9.0.md) (Player-SDK-
> WebRTC-Folge-Themen, separat) und
> [`plan-0.9.5.md`](./plan-0.9.5.md) (Quality-Gates Wave 2, nach
> `0.9.0`).

## 0. Konvention

DoD-Checkboxen tracken den Lieferstand analog
[`done/plan-0.1.0.md`](../done/plan-0.1.0.md) §0:

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
| 0 | Plan-Aktivierung (`open/` → `in-progress/`) + Tool-Pinning-Entscheidung (Trivy vs. Grype für Container-Scan) | ⬜ |
| 1 | Security-Gates: `make vuln-check` (govulncheck) + `make image-scan` (Trivy/Grype) + Wrapper `make security-gates`; CI-Stage parallel zu `make gates` | ⬜ |
| 2 | Generated-Artifact-Drift-Gate: `make generated-drift-check` (Schema-DDL, Contract-Fixtures, Public-API-Snapshot); CI-Stage in `make gates` | ⬜ |
| 3 | Release-Doku, Patch-Release-Konvention in `releasing.md` §3, Versions-Bump 0.8.0 → 0.8.5, Plan nach `done/`, Tag `v0.8.5` | ⬜ |

---

## 1a. Tranche 0 — Plan-Aktivierung + Tool-Pinning

Bezug: keine RAK direkt; Wartungs-/Hygiene-Tranche analog
`done/plan-0.8.0.md` §1a.

DoD:

- [ ] Plan-Skelett von `docs/planning/open/plan-0.8.5.md` nach
  `docs/planning/in-progress/plan-0.8.5.md` verschoben (Status
  `⬜ → 🟡`); Cross-Refs in `roadmap.md` §1.2/§3 und
  `extra-gates.md`-Header nachgezogen.
- [ ] Container-Scanner-Wahl entschieden: Trivy oder Grype. Default-
  Empfehlung: Trivy (etablierter, breitere Image-Format-Unterstützung,
  bessere Default-Policy). Falls Grype, Begründung im Plan-Closeout.
- [ ] Toolchain-Bump-Check: Go/Node/golangci-lint sind seit `0.7.0`
  Tranche 0 aktuell — kein Bump nötig. Wenn ein Sicherheits-CVE in
  einer Toolchain-Version vorliegt, eigene Sub-Tranche analog
  `0.7.0`.

---

## 2. Tranche 1 — Security-Gates (`govulncheck` + Container-Scan)

Bezug: `extra-gates.md` §3.1.

Ziel: Bekannte CVEs in Go-Dependencies und Runtime-Images werden
früh erkannt. Beide Gates sind PR-blockierend in CI (parallel zu
`make gates`, nicht serialisiert).

DoD:

- [ ] `make vuln-check`-Target im Root-`Makefile` führt
  `govulncheck ./...` im `apps/api`-Modul aus. `govulncheck`-Version
  ist gepinnt oder reproduzierbar bezogen (z. B. via
  `go install golang.org/x/vuln/cmd/govulncheck@vX.Y.Z`).
- [ ] `make image-scan`-Target im Root-`Makefile` baut die zu
  scannenden Runtime-Images (API, Dashboard, Analyzer-Service) im
  selben Lauf oder konsumiert eindeutig benannte Image-Tags aus
  einem vorangegangenen CI-Step. Scanner-Tool (Trivy oder Grype, aus
  Tranche 0) ist über Docker-Run mit gepinnter Version aufgerufen.
- [ ] Scan-Policy: `CRITICAL` und `HIGH` blockieren PR; `MEDIUM`
  wird im Output gemeldet, aber nicht blockierend.
- [ ] False-Positive-/Ignore-Regeln liegen versioniert vor —
  empfohlener Pfad: `.security/vulnignore.yaml` mit Begründungs-
  Spalte pro Eintrag.
- [ ] CI gibt maschinenlesbare Scan-Artefakte aus (JSON/SARIF),
  hochgeladen als Workflow-Artefakt.
- [ ] Wrapper-Target `make security-gates` ruft `vuln-check` und
  `image-scan` sequentiell auf; eigenes Help-Text-Eintrag.
- [ ] CI-Workflow `.github/workflows/build.yml` (oder neuer
  `security.yml`): zusätzlicher Job `security` parallel zu `build`,
  führt `make security-gates` aus. PR-blockierend; bei
  Maintenance-Release-Branches darf der Job per `if`-Bedingung
  gefiltert werden.

---

## 3. Tranche 2 — Generated-Artifact-Drift-Gate

Bezug: `extra-gates.md` §3.4.

Ziel: Generierte Artefakte bleiben synchron zu ihren Quellen.
Drift bricht den PR mit klarer Regenerierungs-Anweisung.

DoD:

- [ ] `make generated-drift-check`-Target im Root-`Makefile`:
  ruft die Generierungs-/Sync-Targets auf (`make schema-generate`,
  `make sync-contract-fixtures`, `pnpm --filter @npm9912/player-sdk
  exec node scripts/check-public-api.mjs`) und führt anschließend
  `git diff --exit-code` auf die erwarteten Pfade aus.
- [ ] Geprüfte Artefakte: `apps/api/internal/storage/schema.sql`
  (aus `schema.yaml`),
  `apps/api/adapters/driven/streamanalyzer/testdata/contract-*.json`
  (aus `spec/contract-fixtures/analyzer/*.json`),
  `packages/player-sdk/scripts/public-api.snapshot.txt`
  (aus `index.ts`).
- [ ] Fehlertext nennt den konkreten Regenerierungsbefehl pro Pfad
  (z. B. „Run `make schema-generate` to regenerate schema.sql").
- [ ] Gate läuft ohne Netzwerk (Generierung verwendet repo-lokale
  Fixtures plus die bereits gepinnten Tools).
- [ ] In `make gates` aufgenommen: ist deterministisch und schnell
  genug, um nicht parallelisiert werden zu müssen.
- [ ] CI-Workflow erbt die `make gates`-Erweiterung; kein neuer
  Job nötig.

---

## 4. Tranche 3 — Release-Doku, Patch-Release-Konvention, Closeout

Bezug: §0.6 (Patch-Release-Konvention); `docs/user/releasing.md`;
`README.md`; `roadmap.md`.

Ziel: `0.8.5` ist auffindbar dokumentiert, Patch-Release-Konvention
in `releasing.md` §3 verankert, Tag `v0.8.5` gesetzt.

DoD:

- [ ] `docs/user/releasing.md` §3 erweitert um einen kurzen Block
  „Patch-Release-Konvention (`0.X.Y`)" — Definition aus §0.6 plus
  Hinweis, dass Patch-Releases keinen Lastenheft-Patch brauchen
  und keine RAK-Verifikationsmatrix führen.
- [ ] `README.md` Status-Block erwähnt `0.8.5` als Patch-Release
  mit Quality-Gates Wave 1 (Security + Generated-Drift); kein
  Feature-Release-Hinweis.
- [ ] Versions-Bump 0.8.0 → 0.8.5 in allen package.json (root,
  apps, packages) plus `apps/api/cmd/api/main.go` `serviceVersion`,
  `packages/player-sdk/src/version.ts`, `packages/player-sdk/
  scripts/pack-smoke.mjs` `expectedVersion`,
  `contracts/sdk-compat.json` `sdk_version` und allen Test-
  Fixtures, die Versions-Strings hartkodieren (Bulk-Fix per
  `xargs sed -i 's/"0\.8\.0"/"0.8.5"/g'` über die `_test.go`/
  `.test.ts`-Files plus `spec/contract-fixtures/analyzer/*.json`).
- [ ] CHANGELOG: [Unreleased]-Block in `[0.8.5] - YYYY-MM-DD`
  umgewandelt; neuer leerer [Unreleased]-Block obenauf.
- [ ] `./scripts/verify-doc-refs.sh` (`make docs-check`) grün vor
  Closeout-Commit; `make gates` grün; `make security-gates` grün.
- [ ] `plan-0.8.5.md` von `docs/planning/in-progress/` nach
  `docs/planning/done/` verschoben (`git mv`); alle relativen
  Cross-Refs angepasst; Roadmap §3 zeigt `0.8.5` ✅.
- [ ] Tag `v0.8.5` annotiert; Push opt-in (User-Bestätigung);
  GitHub-Release mit CHANGELOG-`[0.8.5]`-Block als Notes-Body.

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
