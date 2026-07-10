#!/usr/bin/env bash
# smoke-pg-lab.sh — PG-Lab-Integrationssmoke für den optionalen
# Postgres-Runtime-Adapter (ADR-0006).
#
# Verifiziert die frische PG-DB gegen den eingecheckten Live-Stand:
# 1. Startet eine ephemere Postgres-DB (eigener Container + Netz, keine
#    globalen Volumes).
# 2. Ruft den Go-Integrationstest `TestOpenPostgres_LiveSchema` über das
#    golang-Image auf. Der Test wendet die eingecheckten Postgres-
#    Migrationen (internal/storage/migrations/postgres/) via
#    `storage.OpenPostgres` an und verifiziert:
#      - alle 13 Live-Tabellen (V1–V7) mit exakter Spaltenzahl,
#      - die zwei 64-bit-PKs (ingest_sequence, srt_health_samples.id)
#        als bigint mit nextval-Default (BIGSERIAL, width=64),
#      - alle 18 benannten CHECK-Constraints (d-migrate-0.9.9-Fix),
#      - Idempotenz des zweiten Laufs + Nutzbarkeit (Insert/CHECK/FK).
#
# Der PG-DDL-Drift gegen den Live-SQLite-Stand wird separat vom
# `make schema-generate-postgres-check`-Gate bewacht; dieser Smoke
# ergänzt die Verifikation, dass eine frische DB das DDL sauber
# materialisiert.
#
# Konvention:
# - eigener Docker-Run, ephemere Ressourcen (trap-cleanup)
# - opt-in (nicht in `make gates`)
# - exit 0 bei grünem Test, exit 1 sonst
#
# Der Go-Integrationstest ist über die Env-Var MTRACE_PG_LAB_DSN
# gated: ohne sie überspringt `make test` ihn (keine PG-DB im
# Coverage-Gate).

set -euo pipefail

GO_IMAGE="${GO_IMAGE:-golang:1.26.5}"
PG_IMAGE="${PG_IMAGE:-postgres:17-alpine}"
REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
API_DIR="${REPO_ROOT}/apps/api"

if [[ ! -d "${API_DIR}" ]]; then
  echo "[smoke-pg-lab] apps/api nicht gefunden — bitte aus dem Repo-Root aufrufen." >&2
  exit 2
fi

RUN_ID="$$"
NET="mtrace-pglab-net-${RUN_ID}"
DB="mtrace-pglab-db-${RUN_ID}"
PG_USER="mtrace"
PG_PASS="labpass"
PG_DB="mtrace"

cleanup() {
  docker rm -f "${DB}" >/dev/null 2>&1 || true
  docker network rm "${NET}" >/dev/null 2>&1 || true
}
trap cleanup EXIT

echo "▶ smoke-pg-lab: frische Postgres-DB + OpenPostgres-Migrationen + Inventar-Check"
echo "  PG_IMAGE=${PG_IMAGE}  GO_IMAGE=${GO_IMAGE}"

docker network create "${NET}" >/dev/null

docker run -d --name "${DB}" --network "${NET}" \
  -e POSTGRES_USER="${PG_USER}" \
  -e POSTGRES_PASSWORD="${PG_PASS}" \
  -e POSTGRES_DB="${PG_DB}" \
  "${PG_IMAGE}" >/dev/null

echo "  warte auf pg_isready ..."
ready=0
for _ in $(seq 1 60); do
  if docker exec "${DB}" pg_isready -U "${PG_USER}" -d "${PG_DB}" >/dev/null 2>&1; then
    ready=1
    break
  fi
  sleep 1
done
if [[ "${ready}" -ne 1 ]]; then
  echo "[smoke-pg-lab] Postgres wurde nicht bereit (60s Timeout)." >&2
  docker logs "${DB}" 2>&1 | tail -20 >&2
  exit 1
fi

DSN="postgres://${PG_USER}:${PG_PASS}@${DB}:5432/${PG_DB}?sslmode=disable"

echo "  PG-Lab-Integrationstests gegen die frische DB ..."
if ! docker run --rm --network "${NET}" \
  -v "${API_DIR}:/src" -w /src \
  -e MTRACE_PG_LAB_DSN="${DSN}" \
  "${GO_IMAGE}" \
  go test -p 1 ./internal/storage/... ./adapters/driven/persistence/postgres/... \
    -run 'TestOpenPostgres_LiveSchema|TestIngestSequencer_PgLab|TestSrtHealthRepository_PgLab|TestProjectTokenRepository_PgLab' \
    -v -count=1; then
  echo "[smoke-pg-lab] Integrationstest FEHLGESCHLAGEN." >&2
  exit 1
fi

echo "✔ smoke-pg-lab: frische PG-DB trägt das volle Live-Inventar (13 Tabellen, bigint-PKs, 18 CHECKs); DB-autoritativer Sequencer ohne Dups über Replicas (R-28)."
