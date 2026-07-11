#!/usr/bin/env bash
# smoke-scaleout.sh — Multi-Replica-Boot-Smoke (ADR-0006): bringt den
# Scale-out-Stack (2 API-Replicas + Postgres + nginx-LB aus
# docker-compose.scaleout.yml) hoch und verifiziert:
#   - LB-Health grün (GET /api/health über den LB → 200),
#   - beide API-Replicas laufen (beide booteten gegen DENSELBEN Postgres;
#     der Startup-Migrations-Race ist über pg_advisory_lock serialisiert —
#     ein Replica-Crash hieße, der Race-Fix ist kaputt),
#   - geteilter Store: genau eine migrierte Baseline-V1 (kein Race-Duplikat),
#   - Connection-Budget: aktive Verbindungen ≤ max_connections (Headroom).
#
# opt-in (nicht in `make gates`); ephemer (trap-cleanup, down -v).
set -euo pipefail

COMPOSE="${COMPOSE:-docker compose}"
FILE="docker-compose.scaleout.yml"
LB_PORT="${SCALEOUT_LB_PORT:-8088}"
export SCALEOUT_LB_PORT="${LB_PORT}"

cleanup() { ${COMPOSE} -f "${FILE}" down -v --remove-orphans >/dev/null 2>&1 || true; }
trap cleanup EXIT

echo "▶ smoke-scaleout: 2 API-Replicas + Postgres + nginx-LB (Port ${LB_PORT})"
${COMPOSE} -f "${FILE}" up -d --build

echo "  warte auf LB /api/health ..."
ok=0
for _ in $(seq 1 90); do
  if curl -sf -o /dev/null "http://localhost:${LB_PORT}/api/health" 2>/dev/null; then ok=1; break; fi
  sleep 2
done
if [[ "${ok}" -ne 1 ]]; then
  echo "[smoke-scaleout] LB-Health nicht erreichbar (180s Timeout)." >&2
  ${COMPOSE} -f "${FILE}" logs --tail 40 >&2
  exit 1
fi

echo "  laufende Services prüfen ..."
running="$(${COMPOSE} -f "${FILE}" ps --services --filter status=running | sort | tr '\n' ' ')"
echo "    running: ${running}"
for svc in api-1 api-2 lb postgres; do
  echo " ${running} " | grep -q " ${svc} " || { echo "[smoke-scaleout] Service ${svc} läuft nicht." >&2; exit 1; }
done

echo "  geteilter Store + Migrations-Version ..."
ver="$(${COMPOSE} -f "${FILE}" exec -T postgres psql -U mtrace -d mtrace -tAc "SELECT string_agg(version::text, ',') FROM schema_migrations WHERE dirty=0")"
echo "    schema_migrations (dirty=0): ${ver}"
if [[ "$(echo "${ver}" | tr -d '[:space:]')" != "1" ]]; then
  echo "[smoke-scaleout] erwartete genau Baseline-Version 1, fand '${ver}'." >&2
  exit 1
fi

echo "  Connection-Budget ..."
conns="$(${COMPOSE} -f "${FILE}" exec -T postgres psql -U mtrace -d mtrace -tAc "SELECT count(*) FROM pg_stat_activity WHERE datname='mtrace'" | tr -d '[:space:]')"
maxc="$(${COMPOSE} -f "${FILE}" exec -T postgres psql -U mtrace -d mtrace -tAc "SHOW max_connections" | tr -d '[:space:]')"
echo "    aktive Connections: ${conns} / max_connections ${maxc}"
if (( conns > maxc )); then
  echo "[smoke-scaleout] Connections (${conns}) > max_connections (${maxc})." >&2
  exit 1
fi

echo "✔ smoke-scaleout: 2 Replicas teilen den Postgres-Store, LB-Health grün, genau Baseline-V1 migriert (kein Startup-Race), Connections ${conns}/${maxc}, Issuance-Limiter=memory."
