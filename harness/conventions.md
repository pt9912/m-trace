# Harness conventions

## Purpose

This file declares m-trace-specific structure rules and adaptations to the
adopted harness baseline. It supplements the baseline without copying it.

## Baseline

m-trace adopts `grundlagen-konventionen.md` from ai-harness-course at commit
[`d2f60dae33516ef09e84930bb6bf28e31b275ade`](https://github.com/pt9912/ai-harness-course/blob/d2f60dae33516ef09e84930bb6bf28e31b275ade/lab/regelwerk/grundlagen-konventionen.md),
retrieved 2026-07-14.

## Spec strata

| Stratum | Files | Role |
|---|---|---|
| Contract | `spec/lastenheft.md` | Binding requirements and acceptance criteria |
| Technical | `spec/backend-api-contract.md`, `spec/browser-support.md`, `spec/player-sdk.md`, `spec/telemetry-model.md` | Binding technical detail that refines the contract |
| View | `spec/architecture.md` | Derived component, dependency and data-flow view |

Normative reference stability is `Contract > Technical > View > ADR >
Planning`. References point upward. Spec documents do not use ADR or planning
artifacts as normative sources. Downward provenance is permitted only in
designated history sections.

## Adaptations

### MR-001 - Repository paths

- **Date:** 2026-07-14
- **Scope:** ADR and planning directories
- **Baseline difference:** m-trace uses `docs/adr/` instead of
  `docs/plan/adr/`, and `docs/planning/` instead of
  `docs/plan/planning/`.
- **Reason:** Established public repository layout with extensive stable links.
- **Resolution trigger:** Permanent unless a separately reviewed repository
  layout migration is approved.

### MR-002 - Accepted ADR grandfathering

- **Date:** 2026-07-14
- **Scope:** `docs/adr/0001-*.md` through `docs/adr/0007-*.md`
- **Baseline difference:** These accepted Brownfield records contain historic
  plan provenance outside a designated history section.
- **Reason:** Accepted ADRs are immutable under the adopted baseline and are
  not rewritten solely to retrofit the convention.
- **Resolution trigger:** Permanent historical exception. New ADRs receive no
  exemption; future accepted-ADR changes are blocked by the ADR immutability
  sensor.

### MR-003 - Requirement ID families

- **Date:** 2026-07-14
- **Scope:** Contract, plans, commits and reviews
- **Baseline difference:** m-trace predates the `LH-*` example family and uses
  `F-*`, `NF-*`, `MVP-*`, `AK-*`, `RAK-*` and `R-*`.
- **Reason:** The identifiers are part of the established contract and release
  history.
- **Resolution trigger:** Permanent. New requirement families must be declared
  here before use.

### MR-004 - WSL host-path examples

- **Date:** 2026-07-14
- **Scope:** Three WSL troubleshooting examples in
  `docs/user/local-development.md`
- **Baseline difference:** The `hostpaths` sensor normally rejects host-local
  absolute paths.
- **Reason:** These paths are the subject of the operator guidance and cannot
  be replaced by repository-relative paths without losing the diagnosis.
- **Resolution trigger:** Permanent while WSL2 remains supported. The
  `hostpaths` sensor therefore gates `/Users` and `/Development`; `/mnt` and
  `/home` are intentionally outside its configured prefix set.

## Sensor binding classes

m-trace currently uses requirement binding (`F-*`, `NF-*`, `MVP-*`, `AK-*`,
`RAK-*`, `R-*`), ADR binding (`ADR-NNNN`) and reproducibility binding through
immutable image digests.

## Modes

| Sub-area | Mode | Graduation condition |
|---|---|---|
| Spec reference direction | Greenfield | Enforced by `make docs-check`; no open reconciliation findings |
| Existing accepted ADRs | Brownfield, grandfathered | Historical files remain immutable; every new ADR follows the baseline |
| Commit traceability | Greenfield for new pull requests | PR ranges pass `make docs-commits`; pre-adoption history remains unchanged |
| Requirement coverage | Brownfield, observable | Every required requirement has a slice or curated coverage reference and `make doc-complete` passes |
| Requirement links | Brownfield | Stable per-requirement anchors exist and the ID-link gate can be enabled without ambiguous root links |

## Requirement-coverage convergence

d-check v0.43.0 reads the existing `Kennung`/`Prioritaet`/`Anforderung` and
`Akzeptanzkriterium` tables natively. `make doc-trace` is the advisory sensor.
The historical `RAK-51` redefinition uses the explicit `duplicate-ids: last`
policy because the later row raises its modality from Kann to Muss.

`make doc-complete` is not bound into CI while required requirements without a
slice or curated coverage reference remain. Graduation requires triaging those
entries, adding truthful upward coverage references or explicit curated
coverage, and reaching a passing gate without weakening the modality policy.

## Requirement-link convergence

The d-check `ids` module is not enabled yet. Requirements currently live in
Markdown table rows, so automatic repair would turn 340 bare identifiers into
links to the beginning of `spec/lastenheft.md`, not to the individual
definition. That creates link-shaped ambiguity and is not accepted as
convergence.

Graduation requires these steps:

1. Give each normative requirement a stable, directly addressable definition
   anchor without changing its identifier or contractual wording.
2. Verify that d-check resolves every configured ID family to that definition,
   not merely to the containing file.
3. Migrate upward references in active technical, ADR, planning and user
   documentation; historical release records remain unchanged where required.
4. Enable `ids` only after its advisory run reports no ambiguous target and the
   resulting links pass `links` and `anchors`.
