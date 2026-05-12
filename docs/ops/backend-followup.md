# Ops Backend Follow-up

> Status: `0.14.0` trigger and hardening record. This document is a
> planning artifact, not an operator runbook for production backends.
> SQLite and the local Compose lab remain the default paths.

## 1. Postgres Trigger Record

Decision imported from `0.13.0`: Postgres remains
`defer-with-migration-seed`. No runtime adapter, DSN selector, dual-write
path or SQLite export is active in `0.14.0`.

### 1.1 Trigger Status

| Trigger | Threshold | Current `0.14.0` status | Owner | Result |
| --- | --- | --- | --- | --- |
| Multi-replica store | `>= 2` API replicas need the same store without shared-volume SQLite | Not triggered. K8s examples are single-replica and `deploy/k8s/api.yaml` uses `strategy: Recreate`. | Platform/Ops | Keep deferred |
| Recovery SLO | `RPO <= 15 min` or `RTO <= 30 min` becomes binding | Not triggered. No production backup/restore SLO is committed. | Platform/Ops | Keep deferred |
| Retention/read load | `> 10,000,000` events with read p95 `< 2 s` requirement | Not triggered. No retention SLA or high-volume read report exists. | Platform/QA | Keep deferred |

### 1.2 Migration Model

`apps/api/internal/storage/schema.yaml` remains the neutral schema anchor.
A future Postgres implementation must be introduced as a separate slice
with all three paths defined before runtime enablement:

| Path | Required behavior before activation |
| --- | --- |
| `migrate up` | Generate Postgres DDL from `schema.yaml`, review type mapping, then run contract tests against SQLite and Postgres. |
| `rollback` | Disable the Postgres selector before publish or migrate back through a documented export/import snapshot. No rollback may rely on implicit dual-write state. |
| `replay` | Rebuild Postgres state from a SQLite snapshot or API-compatible event/session fixtures; replay must preserve project/session scope and canonical event order. |

Compatibility boundary:

- SQLite stays the local default and test baseline.
- Postgres may only be opt-in until migration, rollback, replay and
  backup/restore are proven in CI or an explicit operator runbook.
- Cursor and pagination contracts remain storage-portable. Ingest order
  must not depend on SQLite `ROWID` semantics.

### 1.3 Schema Differences To Review

| Area | SQLite today | Postgres review requirement |
| --- | --- | --- |
| Time values | Stored through adapter-level time encoding | Confirm timezone handling and ordering for `started_at`, `server_received_at`, `collected_at` and cursor comparisons. |
| Identifiers | `identifier` maps to SQLite autoincrement for `ingest_sequence` and sample IDs | Decide sequence ownership and reset behavior; do not expose backend-specific sequence gaps as API semantics. |
| JSON | `meta` is validated in Go, not via DB-specific JSON checks | Decide `jsonb` vs. text and keep Go validation as the portable contract. |
| Constraints | Check and foreign-key constraints are represented in `schema.yaml` | Verify generated constraint names and deferred/immediate behavior before enabling writes. |
| Transactions | SQLite adapter owns current transactional boundaries | Define equivalent transaction scope per repository method before adding a Postgres adapter. |
| Pagination | Canonical sort uses project/session/time/sequence/ingest sequence | Preserve stable ordering across equal timestamps and null sequence numbers. |

### 1.4 Minimal Adapter Scope If Triggered

The first Postgres slice may implement only the existing persistence
ports needed by current read/write paths. It must not add product
surface area while proving storage portability.

Required before any `proceed` decision:

- contract tests shared with SQLite;
- migration and rollback rehearsal;
- backup/restore note;
- failure-mode note for unavailable Postgres;
- no local test or Compose default changes unless explicitly opted in.

## 2. Analytics Trigger Record

Decision imported from `0.13.0`: ClickHouse, VictoriaMetrics and Mimir
remain deferred. `0.14.0` does not run a POC and does not add an
analytics backend dependency.

### 2.1 Trigger Status

| Trigger | Threshold | Current `0.14.0` status | Owner | Result |
| --- | --- | --- | --- | --- |
| High-volume event analytics | `> 50,000,000` playback/SRT/WebRTC events per day | Not triggered. No production-volume report exists. | Platform/QA | Keep deferred |
| API/Prometheus gap | Required ad-hoc analysis cannot be answered by existing API or Prometheus paths | Not triggered. No named query workload is blocked. | Platform/QA | Keep deferred |
| POC readiness | Named owner, 30-day maximum duration, success and abort criteria | Not triggered. No POC owner or date exists. | Platform/QA | Keep deferred |

### 2.2 Candidate Workloads

| Workload | Candidate backend | Required before POC |
| --- | --- | --- |
| Session/event ad-hoc scans by project, time range, event name and delivery status | ClickHouse | Bounded schema, import path, cost estimate and replay plan. |
| Long-retention technical metrics and rollups | VictoriaMetrics | Metric naming/label cardinality review and retention budget. |
| Multi-tenant metrics at larger scale | Mimir | Tenant model, auth boundary and operational owner. |

### 2.3 Data Model And Retention Boundary

No analytics data model is active. A future POC must define:

- source of truth: API event stream, SQLite snapshot export or batch
  replay fixture;
- retention window and deletion behavior;
- deduplication key and replay idempotency;
- project/tenant scope;
- query list with expected p95 and approximate daily volume;
- abort criteria for cost, operational complexity or data-model gaps.

### 2.4 Data Flow Boundary

Allowed future POC flows:

- batch export from SQLite snapshot;
- synthetic load from contract fixtures;
- isolated replay from API-compatible event/session records.

Disallowed without a new plan decision:

- mandatory dual-write from the API hot path;
- default Compose dependency on ClickHouse, VictoriaMetrics or Mimir;
- unbounded retention;
- more than one analytics backend in the same POC.

## 3. Gate Note

This record is documentation-only. The relevant gate is `make docs-check`.
Runtime gates become mandatory only if a later tranche adds code,
containers or CI wiring.
