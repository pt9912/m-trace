# Harness

## Purpose

This directory records the repository-level precedence, conventions and real
quality sensors used by humans and automation. It points to authoritative
sources and does not duplicate their contracts.

## Source precedence

When two sources conflict, the higher-ranked source wins and the lower-ranked
source must be corrected:

1. `spec/lastenheft.md` (contract)
2. technical specifications under `spec/`
3. `spec/architecture.md` (derived architectural view)
4. accepted records under `docs/adr/`
5. `docs/planning/in-progress/roadmap.md`
6. user and operations documentation under `docs/user/` and `docs/ops/`
7. root README files
8. this harness documentation

The exact classification and repository-specific adaptations are declared in
[`conventions.md`](conventions.md).

## Guides

| Guide | Role |
|---|---|
| [`spec/lastenheft.md`](../spec/lastenheft.md) | Contract: what m-trace must provide |
| [`spec/architecture.md`](../spec/architecture.md) | Derived component and dependency view |
| [`docs/adr/`](../docs/adr/) | Rationale for accepted architectural choices |
| [`docs/planning/in-progress/roadmap.md`](../docs/planning/in-progress/roadmap.md) | Delivery status and sequencing |

## Sensors

| Target | Checks |
|---|---|
| `make docs-check` | Markdown references, spans, tracked targets, code paths and document direction |
| `make doc-trace` | Advisory requirements matrix from the native Lastenheft tables and planning references |
| `make docs-immutable STAGED=1` | Accepted-ADR core against the staged diff |
| `make docs-commits RANGE=base..head` | Commit-message traceability over a pull-request range |
| `make gates` | Required repository quality gates |
| `make build` | Buildable release artifacts |

Only commands that exist in the Makefile are listed here. Run status belongs
to CI and is not recorded in this document.

## Traceability rules

Requirements use the existing `F-*`, `NF-*`, `MVP-*`, `AK-*`, `RAK-*` and
`R-*` families. Architecture decisions use `ADR-NNNN`. New normative
references point from volatile to stable sources; downward provenance is
confined to designated history sections.

Commit-message enforcement applies to pull-request ranges. Documentation,
test, build, CI and maintenance commits are exempt; feature and fix commits
carry a requirement, decision or plan identifier.

## Safety and scope boundaries

- Never weaken a higher-ranked contract to match an implementation drift.
- Never claim a gate that has no executable target.
- Existing accepted ADRs are historical records; changes require the
  documented decision process.

## Minimal agent workflow

1. Read the highest-ranked source relevant to the task.
2. Inspect existing code and tests before changing behavior.
3. Link the change to an existing requirement, decision, test or gate.
4. Implement the smallest coherent change.
5. Run focused tests.
6. Run `make docs-check` for documentation changes.
7. Run the proportionate aggregate gate.
8. Update lower-ranked status or user documentation without changing the
   higher-ranked contract implicitly.
