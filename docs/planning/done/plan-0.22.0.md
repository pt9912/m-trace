# Implementation Plan — `0.22.0` (Quality-Gates Follow-up)

> **Status**: ✅ released und archiviert in `done/`.
>
> **Vorgänger**: `0.21.0` ist als OCI-Image-Publishing-Release
> veröffentlicht und archiviert in
> [`../done/plan-0.21.0.md`](../done/plan-0.21.0.md).
>
> **Auslöser**: `extra-gates.md` hält die Benchmark-/Mutation-Gates
> aus Wave 2 als offene Nachschärfung. Die Benchmark-Beobachtung hat
> fünf aufeinanderfolgende grüne Läufe; Mutation hat noch keine
> belastbare Score-Reihe nach der Package-Scope-Migration.
>
> **Release-Typ**: Patch-/Tooling-Release ohne Lastenheft-Patch, ohne
> Runtime-, Wire-, Public-API-, Persistenz- oder Analyzer-Schema-
> Änderung. Der Release ist trotzdem versionstragend (`0.22.0`),
> weil `release.published` npm-Pakete und GHCR-Images veröffentlicht.

## 0. Scope

In Scope:

- `make benchmark-smoke` nach fünf grünen Beobachtungsläufen in
  `make gates` aufnehmen.
- `benchmark-observation.yml` von warning-only auf hartes Nightly-
  Ergebnis umstellen.
- Mutation-Nightly auf den aktuellen TS-Package-Scope
  `@pt9912/player-sdk` korrigieren.
- Doku aktualisieren: `extra-gates.md`, `docs/perf/budgets.md`,
  Mutation-Runbook, Roadmap und Changelog.

Nicht in Scope:

- Mutation-Testing PR-blockierend machen.
- Benchmark-Budgets schärfen oder neue Hot-Paths ergänzen.
- Lastenheft-Patch.
- neue Runtime-Artefakte jenseits des Release-Bumps für Package- und
  Image-Publishing.
- neue Runtime-, Wire-, API-, Persistenz- oder Analyzer-Schema-
  Funktionalität.

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Ergebnis |
| --- | --- | --- |
| 0 | Evidence | fünf grüne Benchmark-Beobachtungsläufe belegt |
| 1 | Benchmark-Promotion | `benchmark-smoke` in `make gates`, Nightly hart |
| 2 | Mutation-Messbarkeit | TS-Package-Filter korrigiert, Blockierung deferred |
| 3 | Doku/Gates | Planungsdoku aktualisiert, lokale Verifikation grün |
| 4 | Release-Closeout | Versionsträger `0.22.0`, Plan archiviert, Tag/Release |

## 2. Tranche 0 — Evidence

DoD:

- [x] `benchmark-observation.yml` hat fünf grüne Läufe:
  [`25592982776`](https://github.com/pt9912/m-trace/actions/runs/25592982776),
  [`25621106187`](https://github.com/pt9912/m-trace/actions/runs/25621106187),
  [`25643426077`](https://github.com/pt9912/m-trace/actions/runs/25643426077),
  [`25704811721`](https://github.com/pt9912/m-trace/actions/runs/25704811721),
  [`25769811661`](https://github.com/pt9912/m-trace/actions/runs/25769811661).
- [x] `benchmark.yml` hat parallel grüne Nightly-Regression-Runs:
  [`25643634634`](https://github.com/pt9912/m-trace/actions/runs/25643634634),
  [`25705097012`](https://github.com/pt9912/m-trace/actions/runs/25705097012),
  [`25769998744`](https://github.com/pt9912/m-trace/actions/runs/25769998744).
- [x] Mutation bleibt deferred: Workflow-Erfolg ist wegen
  `continue-on-error` und bisherigem TS-Filter nicht als
  >70%-Score-Reihe verwertbar.

## 3. Tranche 1 — Benchmark-Promotion

DoD:

- [x] `make gates` enthält `benchmark-smoke`.
- [x] Root-`Makefile` beschreibt Benchmark-Smoke nicht mehr als
  Opt-in/Observation-Gate.
- [x] `.github/workflows/benchmark-observation.yml` läuft ohne
  `continue-on-error`.
- [x] Build-Workflow-Timeout ist auf den längeren Gates-Pfad
  angepasst.
- [x] Build-Workflow installiert Node/pnpm vor `make gates`, weil
  `benchmark-smoke` den TS-Bench hostseitig ausführt.

## 4. Tranche 2 — Mutation-Messbarkeit

DoD:

- [x] `.github/workflows/mutation.yml` nutzt
  `pnpm --filter @pt9912/player-sdk run mutation`.
- [x] Der TS-Mutation-Step prüft den Package-Filter vor dem
  bewusst nicht-blockierenden Stryker-Lauf mit
  `pnpm --filter @pt9912/player-sdk exec pwd`.
- [x] `docs/dev/mutation-testing.md` stellt klar, dass ältere grüne
  TS-Nightly-Runs nach Scope-Migration keine reale Score-Reihe
  belegen.
- [x] Mutation-PR-Blockierung bleibt an drei reale Nightly-Reports
  mit >70 % Score pro Modul gebunden.

## 5. Tranche 3 — Doku / Gates

DoD:

- [x] `extra-gates.md` dokumentiert Benchmark-Promotion und
  Mutation-Defer.
- [x] `docs/perf/budgets.md` markiert die aktuelle Budget-Tabelle als
  PR-blockierend.
- [x] `CHANGELOG.md` und Roadmap nennen den `0.22.0`-Quality-Gates-
  Follow-up.
- [x] `make benchmark-smoke`
- [x] `pnpm --filter @pt9912/player-sdk exec pwd`
- [x] `make docs-check`
- [x] `make gates`
- [x] Remote-CI-Fix: `Build` run
  [`25806628250`](https://github.com/pt9912/m-trace/actions/runs/25806628250)
  zeigte fehlendes hostseitiges `pnpm` im neuen Gate-Pfad; Workflow
  korrigiert.
- [x] Remote-CI-Fix: `Build` run
  [`25807084194`](https://github.com/pt9912/m-trace/actions/runs/25807084194)
  zeigte farbigen Vitest-4-Bench-Output im neuen Gate-Pfad; der
  Budget-Parser entfernt ANSI-Sequenzen vor dem Tabellen-Match.
- [x] `make analyzer-benchmark-smoke` nach Parser-Fix.
- [x] Remote-CI: `Build` run
  [`25807641979`](https://github.com/pt9912/m-trace/actions/runs/25807641979)
  ist nach Rerun grün. Der erste Versuch scheiterte an einem
  transienten GHCR-Timeout beim Pull von `d-migrate:0.9.5`; der
  Rerun bestand `Quality gates`, Build und CLI-Smoke.

## 6. Tranche 4 — Release-Closeout

DoD:

- [x] Versionsträger auf `0.22.0` synchronisiert:
  npm-Workspaces, API-Service-Version, SDK-/Analyzer-Version,
  Analyzer-Fixtures, Contract-Testdata, `sdk-compat.json` und
  K8s-Image-Tags.
- [x] `CHANGELOG.md` datiert `0.22.0` auf 2026-05-13.
- [x] Roadmap verweist auf den archivierten Plan.
- [x] `make docs-check`, `make ts-test`, `make api-test`.
- [x] `make package-publish-dry-run`.
- [x] `make image-publish-dry-run VER=0.22.0`.
- [x] Tag `v0.22.0` und GitHub-Release aus diesem Closeout-Commit.

Hinweis: Die Package-/Image-Publish-Workflows starten erst durch
`release.published`; sie werden nach der GitHub-Release-Erstellung
über Actions geprüft.

Closeout-Verdict:

- Benchmark-Smoke: ✅ Implementiert und lokal verifiziert; alle 6
  Go- und 7 TS-Budget-Smokes liegen unter der dokumentierten Schwelle.
- Mutation: ✅ Messbarkeit korrigiert, Blockierung bewusst deferred.
