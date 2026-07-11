#!/usr/bin/env bash
# smoke-scaleout-load.sh — Scale-out-Lasttest (ADR-0006, R-26 c): treibt
# k6-Playback-Event-Last gegen den Multi-Replica-Stack
# (docker-compose.scaleout.yml, 2 API-Replicas + geteilter Postgres +
# nginx-LB) und belegt mit Messwerten:
#
#   - Phase A (1 Replica): Last direkt gegen api-1 → Durchsatz-Baseline.
#   - Phase B (2 Replicas): Last über den LB (beide) → horizontale Skalierung.
#   - Readback via psql gegen den GETEILTEN Postgres (kein SQLite-GLOB, kein
#     Volume-Hack): persisted = COUNT(*), distinct = COUNT(DISTINCT
#     ingest_sequence) je Session-Prefix.
#
# Verdict je Phase:
#   - persisted == accepted (kein stiller Verlust über die Replicas),
#   - distinct == persisted (0 Duplikate wasserdicht — der DB-autoritative
#     Sequencer (R-28) schließt store-seitige ingest_sequence-Dups aus; der
#     Distinct-Check belegt es explizit).
# Plus: Phase-B-Durchsatz vs. Phase-A (Skalierungssignal).
#
# opt-in (nicht in `make gates`); ephemer (trap-cleanup, down -v).
set -euo pipefail

ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
cd "$ROOT"
COMPOSE="${COMPOSE:-docker compose}"
FILE="docker-compose.scaleout.yml"
K6_IMAGE="${K6_IMAGE:-grafana/k6}"
LB_PORT="${SCALEOUT_LB_PORT:-8088}"
API1_PORT="${SCALEOUT_API1_PORT:-8089}"
export SCALEOUT_LB_PORT="${LB_PORT}" SCALEOUT_API1_PORT="${API1_PORT}"
VUS="${VUS:-10}"
DURATION="${DURATION:-15s}"
BATCH_SIZE="${BATCH_SIZE:-20}"
TOKEN="${PROJECT_TOKEN:-demo-token}"

cleanup() { ${COMPOSE} -f "${FILE}" down -v --remove-orphans >/dev/null 2>&1 || true; }
trap cleanup EXIT

echo "▶ smoke-scaleout-load: k6-Last gegen 2 Replicas + geteilten Postgres (R-26 c)"
${COMPOSE} -f "${FILE}" up -d --build

wait_health() {
  local url="$1" name="$2"
  for _ in $(seq 1 90); do
    curl -sf -o /dev/null "$url" 2>/dev/null && return 0
    sleep 2
  done
  echo "[scaleout-load] ${name} nicht erreichbar (${url})" >&2
  ${COMPOSE} -f "${FILE}" logs --tail 40 >&2
  exit 1
}
wait_health "http://localhost:${LB_PORT}/api/health" "LB"
wait_health "http://localhost:${API1_PORT}/api/health" "api-1"

psql_q() { ${COMPOSE} -f "${FILE}" exec -T postgres psql -U mtrace -d mtrace -tAc "$1" | tr -d '[:space:]'; }
k6_count() { python3 -c "import json;m=json.load(open('$1'))['metrics'].get('$2',{}).get('values',{});print(int(m.get('count',0)))"; }
k6_rate() { python3 -c "import json;m=json.load(open('$1'))['metrics'].get('$2',{}).get('values',{});print(round(m.get('rate',0),1))"; }

# run_phase gibt NUR die accepted-Rate auf stdout zurück; aller Fortschritt
# geht nach stderr, damit `$(run_phase …)` sauber die Rate erfasst.
run_phase() {
  local base="$1" prefix="$2" label="$3"
  local out; out="$(mktemp -d)"
  echo "  ── Phase ${label}: ${VUS} VUs / ${DURATION} gegen ${base} (prefix ${prefix}-) ──" >&2
  docker run --rm --network host --user "$(id -u):$(id -g)" \
    -v "${ROOT}/scripts/load:/scripts:ro" -v "${out}:/work" \
    "${K6_IMAGE}" run --vus "${VUS}" --duration "${DURATION}" \
    -e BASE_URL="${base}" -e PROJECT_TOKEN="${TOKEN}" \
    -e BATCH_SIZE="${BATCH_SIZE}" -e SESSION_PREFIX="${prefix}" \
    /scripts/playback-events.k6.js >"${out}/k6.log" 2>&1 || {
      echo "[scaleout-load] k6 (${label}) fehlgeschlagen:" >&2; tail -15 "${out}/k6.log" >&2; exit 1; }
  [ -f "${out}/summary.json" ] || { echo "[scaleout-load] keine summary.json (${label})" >&2; exit 1; }

  local accepted rate persisted distinct esc
  accepted="$(k6_count "${out}/summary.json" mtrace_events_accepted)"
  rate="$(k6_rate "${out}/summary.json" mtrace_events_accepted)"
  esc="${prefix//_/\\_}"  # `_` in LIKE literal behandeln
  persisted="$(psql_q "SELECT count(*) FROM playback_events WHERE session_id LIKE '${esc}-%'")"
  distinct="$(psql_q "SELECT count(DISTINCT ingest_sequence) FROM playback_events WHERE session_id LIKE '${esc}-%'")"
  echo "    accepted(202)=${accepted} (${rate}/s) persisted(psql)=${persisted} distinct(ingest_sequence)=${distinct}" >&2

  if [ "${accepted}" -le 0 ]; then
    echo "[scaleout-load] INCONCLUSIVE (${label}): 0 accepted — Auth/Setup? kein Verlust-Urteil" >&2; exit 3; fi
  if [ "${persisted}" != "${accepted}" ]; then
    echo "[scaleout-load] FAIL (${label}): persisted ${persisted} != accepted ${accepted} — Verlust über Replicas" >&2; exit 1; fi
  if [ "${distinct}" != "${persisted}" ]; then
    echo "[scaleout-load] FAIL (${label}): distinct ${distinct} != persisted ${persisted} — Duplikat-ingest_sequence" >&2; exit 1; fi
  echo "    ✓ ${label}: persisted==accepted (kein Verlust), distinct==persisted (0 Dups)" >&2
  echo "${rate}"
}

rate1="$(run_phase "http://localhost:${API1_PORT}" "sol1" "A(1 Replica)")"
rate2="$(run_phase "http://localhost:${LB_PORT}" "sol2" "B(2 Replicas)")"

echo ""
echo "  ── Skalierung ──"
echo "    Durchsatz: 1 Replica ${rate1}/s  →  2 Replicas ${rate2}/s"
speedup="$(python3 -c "print(round(${rate2}/${rate1},2) if ${rate1}>0 else 0)")"
echo "    Speedup (2/1): ${speedup}×"

echo ""
echo "✔ smoke-scaleout-load (R-26 c): über 2 Replicas + geteilten Postgres — persisted==accepted (kein stiller Verlust), COUNT(DISTINCT ingest_sequence)==COUNT(*) (0 Duplikate wasserdicht); Durchsatz 1→2 Replicas ${rate1}→${rate2}/s (${speedup}×)."
