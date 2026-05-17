# Implementation Plan — `0.22.1` (devalue Security-Patch + Nightly-Audit)

> **Status**: ✅ released und archiviert in `done/`.
>
> **Vorgänger**: `0.22.0` ist als Quality-Gates-Follow-up
> veröffentlicht und archiviert in
> [`../done/plan-0.22.0.md`](../done/plan-0.22.0.md).
>
> **Auslöser**: GHSA-77vg-94rm-hx3p (`devalue` DoS via sparse-array
> deserialization, Patched-Range `>=5.8.1`) wurde am 2026-05-17
> publiziert — vier Tage nach dem `0.22.0`-Tag. Die TS-Audit-Stage
> in `build.yml::Security gates` hat die Vulnerability beim ersten
> Push nach Veröffentlichung blockiert; das `mtrace-dashboard:0.22.0`-
> Image enthält die verwundbare `devalue@5.7.1` weiterhin. Parallel
> haben zwei Bug-/Refactor-Items im Benchmark-Nightly auf den
> Folge-Patch gewartet.
>
> **Release-Typ**: Patch-/Tooling-Release ohne Lastenheft-Patch,
> ohne Runtime-, Wire-, Public-API-, Persistenz- oder
> Analyzer-Schema-Änderung. Versionstragend, weil
> `image-publish`/`package-publish` die fixed-`devalue`-Closure
> als neues Image-Tag veröffentlicht.

## 0. Scope

In Scope:

- `pnpm.overrides` für `devalue@^5.8.1` in der Workspace-Root
  `package.json`, lockfile-only via `make lock-refresh`.
- Nightly-Mirror der drei Push-Security-Gates
  (`make vuln-check` / `make audit-ts` / `make image-scan`) als
  eigenen Workflow `.github/workflows/security-audit.yml` plus
  Helper-Script `scripts/open-security-audit-issue.sh`.
- Path-Fix im `Benchmark regression (nightly)`-Workflow: das
  redirect-Ziel war `apps/.tmp/bench/current.txt` statt
  `.tmp/bench/current.txt` am Repo-Root; Output über `tee` plus
  `set -o pipefail`, damit `b.Fatalf`-/`--- FAIL`-Zeilen wieder
  im Workflow-Log erscheinen.
- Refactor des `Open regression issue`-Steps: 35-zeiliger
  HEREDOC-Block wandert nach
  `scripts/open-bench-regression-issue.sh`; YAML-Indent leakt nicht
  mehr ins Markdown.
- Doku: `extra-gates.md §3.7` (Nightly-Security-Audit-Mirror),
  Roadmap-Eintrag, Changelog.

Nicht in Scope:

- `@sveltejs/kit`-Upgrade (Override reicht; Major-Bump später).
- Dependabot-Aktivierung als zusätzliche Schicht.
- Runtime-/Analyzer-Funktionalität, Lastenheft-Patch.
- Anpassung der Benchmark-Budgets oder neue Bench-Targets.

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Ergebnis |
| --- | --- | --- |
| 1 | Benchmark-Workflow-Fixes (Refactor + Path/Tee) | CI-Run 25983254894 grün, Artefakt enthält `current.txt` (3850 B vs. 149 B vorher) |
| 2 | devalue-Override + lock-refresh | `make audit-ts` zeigt 0 high; Build-Run 25983535978 grün |
| 3 | Nightly-Security-Audit-Workflow + Helper-Script + extra-gates §3.7 | Manueller Run 25983690742 grün (govulncheck 44s, pnpm audit 16s, trivy 1m33s) |
| 4 | Release-Closeout | Versions-Sweep auf `0.22.1`, Plan archiviert, Tag/Release |

## 2. Tranche 1 — Benchmark-Workflow-Fixes

Commit-IDs: `735ca62` (Refactor), `0a8a4f1` (Path/Tee).

DoD:

- [x] `Open regression issue (on bench compare failure)`-Step ist
  auf einen Einzeiler reduziert; HEREDOC lebt in
  `scripts/open-bench-regression-issue.sh`.
- [x] `Run Go benchmarks`-Step schreibt nach
  `../../.tmp/bench/current.txt` (Repo-Root), nicht mehr
  `../.tmp/bench/`.
- [x] Bench-Output ist im Workflow-Log sichtbar (`tee` +
  `set -o pipefail`); ein zukünftiger Bench-Failure produziert
  diagnostisches Output statt fünf Minuten Stille.
- [x] Verifikation: manueller Workflow-Run
  [`25983254894`](https://github.com/pt9912/m-trace/actions/runs/25983254894)
  ist `success` in 6m18s; Artefakt enthält
  `current.txt` (60 Bench-Einträge, 6 Targets × 10 Iterationen).

## 3. Tranche 2 — devalue-Override

Commit-ID: `4a1163c`.

Hintergrund: `apps/dashboard > @sveltejs/kit@2.58.0 > devalue@5.7.1`
fällt unter den Vulnerability-Range `>=5.6.3 <=5.8.0`. Der bestehende
`picomatch`-Override im Root-`package.json` ist die Vorlage für
denselben Eingriffspunkt.

DoD:

- [x] `pnpm.overrides.devalue = "^5.8.1"` in der Workspace-Root
  `package.json`.
- [x] `make lock-refresh` hebt beide transitiven Resolutions
  (Kit + Svelte) auf `devalue@5.8.1`.
- [x] `make audit-ts` zeigt 0 high (5 vulns total: 1 low,
  4 moderate); Verifikation lokal im Docker-Audit-Stage.
- [x] CI-Bestätigung: Build-Run
  [`25983535978`](https://github.com/pt9912/m-trace/actions/runs/25983535978)
  `success` in 7m24s.

## 4. Tranche 3 — Nightly-Security-Audit-Mirror

Commit-ID: `02ad6ae`.

Hintergrund: Push-/PR-Audit fängt nur Advisories, die zum Zeitpunkt
des Pushes existieren. Vier-Tage-Gap zwischen Release `0.22.0` und
Devalue-Advisory war unsichtbar. Nightly-Mirror schließt die Lücke
mit täglicher Cadence.

DoD:

- [x] Workflow `.github/workflows/security-audit.yml` mit Cron
  `57 1 * * *` Europe/Berlin (gestaffelt: 07 observation, 17 webrtc,
  27 benchmark, 37 fuzz, 47 mutation, 57 security-audit).
- [x] Drei Steps mit `continue-on-error: true` und konsolidiertem
  Issue (`scripts/open-security-audit-issue.sh`), Labels
  `security,audit,plan-0.8.5`.
- [x] Run scheitert explizit am Ende, damit GitHub-UI rot bleibt
  bis das Issue geschlossen ist.
- [x] `docs/planning/in-progress/extra-gates.md §3.7` dokumentiert
  Entscheidung, Scope, Policy, DoD; verweist auf GHSA-77vg-94rm-hx3p
  als Auslöser.
- [x] Manueller Verifikations-Run
  [`25983690742`](https://github.com/pt9912/m-trace/actions/runs/25983690742)
  ist `success` in 2m39s; alle drei Audit-Steps grün, Issue-/Fail-
  Steps korrekt geskippt.

## 5. Tranche 4 — Release-Closeout

DoD:

- [x] Versions-Sweep `0.22.0` → `0.22.1` an allen 5× `package.json`,
  `apps/api/cmd/api/main.go::serviceVersion`,
  `packages/player-sdk/src/version.ts::PLAYER_SDK_VERSION`,
  `contracts/sdk-compat.json::sdk_version`, 20 Analyzer-Fixtures,
  20 testdata-Kopien (`make sync-contract-fixtures`),
  Test-Strings (`http_test.go`, `version.test.ts`).
- [x] `CHANGELOG.md`: `[Unreleased]` → `[0.22.1] - 2026-05-17`.
- [x] `docs/planning/in-progress/roadmap.md`: 0.22.1-Eintrag im
  Lieferstand, im Phase-Block und in der Release-Tabelle.
- [x] Wave-2-Quality-Gates-Verifikation laut
  `docs/user/releasing.md §3.1` zitiert die jüngsten Nightly-Run-IDs
  (siehe §6).
- [x] `make gates` und `make release-guard` lokal grün.
- [x] Plan direkt in `done/plan-0.22.1.md` archiviert, Status
  auf "released" gesetzt.
- [x] Tag `v0.22.1`, Push optional vom Operator.

## 6. Verifikations-Run-IDs (für Tag-Annotation)

| Gate | Run-ID | Status |
| --- | --- | --- |
| Build (devalue-Fix) | `25983535978` | success |
| Build (security-audit added) | `25983675570` | success |
| Benchmark regression (manual) | `25983254894` | success |
| Security audit nightly (manual) | `25983690742` | success |
| Benchmark smoke (latest schedule) | `25976389964` | success |
| WebRTC drift smoke (latest schedule) | `25976517253` | success |
| Fuzz nightly (latest schedule) | `25977954286` | success |
| Mutation nightly (latest schedule) | `25978037105` | success |
