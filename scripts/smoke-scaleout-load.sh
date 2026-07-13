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
# FAIRNESS=1 (R-26 b) schaltet auf den shared Redis-Ingest-Limiter
# (MTRACE_RATE_LIMIT_BACKEND=redis, Redis-Service im Stack) und misst
# statt der Durchsatz-Skalierung die FAIRNESS:
#   - Phase A/B laufen THROTTLED (Default-Limiter 100/s): mit shared
#     Limiter darf die 2-Replica-Rate NICHT über die 1-Replica-Rate
#     skalieren — Gate: rate2/rate1 <= FAIRNESS_MAX_SCALE (Default 1.15;
#     invertiert den gemessenen 2,01x-Befund des per-Replica-Limiters).
#   - Phase C: Noisy-Neighbor ÜBER den LB (MT_PROJECTS Lab-Projekte,
#     Seeding via MTRACE_LAB_PROJECTS, synthetische Client-IP je Projekt
#     via XFF + MTRACE_TRUST_FORWARDED_FOR=1; der Lab-nginx reicht XFF
#     durch): Victims 0x 429 + p95 im Budget, Noisy wird gedrosselt
#     (k6-Thresholds), Verlust-/Duplikat-Gates unverändert.
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
FAIRNESS="${FAIRNESS:-0}"
FAIRNESS_MAX_SCALE="${FAIRNESS_MAX_SCALE:-1.15}"
MT_PROJECTS="${MT_PROJECTS:-3}"
MT_NOISY_EVENT_RATE="${MT_NOISY_EVENT_RATE:-400}"
MT_VICTIM_EVENT_RATE="${MT_VICTIM_EVENT_RATE:-50}"
P95_BUDGET_MS="${P95_BUDGET_MS:-1000}"

if [ "$FAIRNESS" = "1" ]; then
  # Shared Limiter + Multi-Tenant-Seeding + XFF-Trust in den Stack
  # injizieren (Pass-throughs in docker-compose.scaleout.yml). Throttled:
  # KEIN Capacity-Override — der Default-Limiter (100/s) ist das Messobjekt.
  export MTRACE_RATE_LIMIT_BACKEND=redis
  export MTRACE_REDIS_ADDR=redis:6379
  export MTRACE_LAB_PROJECTS="$MT_PROJECTS"
  export MTRACE_TRUST_FORWARDED_FOR=1
  unset MTRACE_RATE_LIMIT_CAPACITY MTRACE_RATE_LIMIT_REFILL 2>/dev/null || true
fi

cleanup() { ${COMPOSE} -f "${FILE}" down -v --remove-orphans >/dev/null 2>&1 || true; }
trap cleanup EXIT

if [ "$FAIRNESS" = "1" ]; then
  echo "▶ smoke-scaleout-load FAIRNESS: shared Redis-Ingest-Limiter über 2 Replicas (R-26 b)"
else
  echo "▶ smoke-scaleout-load: k6-Last gegen 2 Replicas + geteilten Postgres (R-26 c)"
fi
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

# run_mt_phase: Noisy-Neighbor über den LB (Phase C des Fairness-Modus).
# k6 fährt die Multi-Tenant-Szenarien (MT_PROJECTS, Thresholds aus dem
# Skript); Gates hier: k6-Exit (Thresholds), Victims 0x 429, Noisy
# gedrosselt, plus Verlust-/Duplikat-Readback für den Prefix.
run_mt_phase() {
  local base="$1" prefix="$2"
  local out; out="$(mktemp -d)"
  echo "  ── Phase C: Noisy-Neighbor über LB — noisy ${MT_NOISY_EVENT_RATE} ev/s vs $((MT_PROJECTS - 1)) Victims je ${MT_VICTIM_EVENT_RATE} ev/s, ${DURATION} ──" >&2
  local k6_rc=0
  docker run --rm --network host --user "$(id -u):$(id -g)" \
    -v "${ROOT}/scripts/load:/scripts:ro" -v "${out}:/work" \
    "${K6_IMAGE}" run \
    -e BASE_URL="${base}" -e BATCH_SIZE="${BATCH_SIZE}" \
    -e SESSION_PREFIX="${prefix}" -e MT_PROJECTS="${MT_PROJECTS}" \
    -e MT_NOISY_EVENT_RATE="${MT_NOISY_EVENT_RATE}" \
    -e MT_VICTIM_EVENT_RATE="${MT_VICTIM_EVENT_RATE}" \
    -e DURATION="${DURATION}" -e P95_BUDGET_MS="${P95_BUDGET_MS}" \
    /scripts/playback-events.k6.js >"${out}/k6.log" 2>&1 || k6_rc=$?
  [ -f "${out}/summary.json" ] || { echo "[scaleout-load] keine summary.json (C)" >&2; tail -15 "${out}/k6.log" >&2; exit 1; }

  local v_sent v_acc v_rl n_rl accepted persisted distinct esc
  v_sent="$(k6_count "${out}/summary.json" mtrace_mt_victim_sent)"
  v_acc="$(k6_count "${out}/summary.json" mtrace_mt_victim_accepted)"
  v_rl="$(k6_count "${out}/summary.json" mtrace_mt_victim_rate_limited)"
  n_rl="$(k6_count "${out}/summary.json" mtrace_mt_noisy_rate_limited)"
  accepted="$(k6_count "${out}/summary.json" mtrace_events_accepted)"
  esc="${prefix//_/\\_}"
  persisted="$(psql_q "SELECT count(*) FROM playback_events WHERE session_id LIKE '${esc}-%'")"
  distinct="$(psql_q "SELECT count(DISTINCT ingest_sequence) FROM playback_events WHERE session_id LIKE '${esc}-%'")"
  echo "    victims: sent=${v_sent} accepted=${v_acc} rate_limited=${v_rl}; noisy rate_limited=${n_rl}" >&2
  echo "    accepted(202)=${accepted} persisted(psql)=${persisted} distinct=${distinct}" >&2

  local fail=0
  [ "${v_sent}" -gt 0 ] || { echo "[scaleout-load] INCONCLUSIVE (C): keine Victim-Events (Setup?)" >&2; exit 3; }
  if [ "${v_rl}" != "0" ]; then
    echo "[scaleout-load] FAIL (C): Victims sahen ${v_rl} rate-limited Events über den LB — Isolation verletzt" >&2; fail=1; fi
  if [ "${n_rl}" = "0" ]; then
    echo "[scaleout-load] FAIL (C): Noisy wurde nie gedrosselt — Lauf misst keine Limiter-Kontention" >&2; fail=1; fi
  if [ "${k6_rc}" -ne 0 ]; then
    echo "[scaleout-load] FAIL (C): k6-Thresholds verletzt (exit=${k6_rc}: victim-429/victim-p95/noisy-429)" >&2; fail=1; fi
  if [ "${persisted}" != "${accepted}" ]; then
    echo "[scaleout-load] FAIL (C): persisted ${persisted} != accepted ${accepted}" >&2; fail=1; fi
  if [ "${distinct}" != "${persisted}" ]; then
    echo "[scaleout-load] FAIL (C): distinct ${distinct} != persisted ${persisted}" >&2; fail=1; fi
  [ "${fail}" -eq 0 ] || exit 1
  echo "    ✓ C: Victims isoliert (0x 429), Noisy gedrosselt (${n_rl}x 429), kein Verlust, 0 Dups" >&2
}

rate1="$(run_phase "http://localhost:${API1_PORT}" "sol1" "A(1 Replica)")"
rate2="$(run_phase "http://localhost:${LB_PORT}" "sol2" "B(2 Replicas)")"

echo ""
echo "  ── Skalierung ──"
echo "    Durchsatz: 1 Replica ${rate1}/s  →  2 Replicas ${rate2}/s"
speedup="$(python3 -c "print(round(${rate2}/${rate1},2) if ${rate1}>0 else 0)")"
echo "    Speedup (2/1): ${speedup}×"

if [ "$FAIRNESS" = "1" ]; then
  # Fairness-Inversion (R-26 b): der shared Limiter macht aus dem
  # gemessenen 2,01x des per-Replica-Limiters ein ~1,0x — die zweite
  # Replica darf das Per-Projekt-Budget NICHT verdoppeln.
  if ! python3 -c "import sys; sys.exit(0 if ${speedup} <= ${FAIRNESS_MAX_SCALE} else 1)"; then
    echo "[scaleout-load] FAIL (Fairness): throttled Skalierung ${speedup}x > ${FAIRNESS_MAX_SCALE}x — Budget skaliert mit Replicas (shared Limiter unwirksam?)" >&2
    exit 1
  fi
  echo "    ✓ Fairness-Inversion: throttled 1→2 Replicas ${speedup}× <= ${FAIRNESS_MAX_SCALE}× (EIN Budget statt N×Capacity)"
  run_mt_phase "http://localhost:${LB_PORT}" "sol3"
  echo ""
  echo "✔ smoke-scaleout-load FAIRNESS (R-26 b): shared Redis-Ingest-Limiter über 2 Replicas — throttled ${rate1}→${rate2}/s (${speedup}× <= ${FAIRNESS_MAX_SCALE}×, Inversion des per-Replica-2,01×); Noisy-Neighbor über LB: Victims isoliert, Noisy gedrosselt; kein Verlust, 0 Duplikate."
else
  echo ""
  echo "✔ smoke-scaleout-load (R-26 c): über 2 Replicas + geteilten Postgres — persisted==accepted (kein stiller Verlust), COUNT(DISTINCT ingest_sequence)==COUNT(*) (0 Duplikate wasserdicht); Durchsatz 1→2 Replicas ${rate1}→${rate2}/s (${speedup}×)."
fi
