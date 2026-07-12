#!/usr/bin/env bash
# smoke-cutover.sh — opt-in-Smoke für den SQLite->Postgres-Cutover (ADR-0007).
#
# Verifiziert doctor + Phase 0 (profile) + Phase 1 (bulk) von
# scripts/cutover-sqlite-postgres.sh gegen ein ephemeres Lab:
#   1. doctor gegen frische, per PG-DDL angelegte Ziel-DB + gesunde Quelle
#      -> alle Pre-Flight-Checks grün (Exit 0).
#   2. profile gegen die gesunde Quelle -> grün (Exit 0); Cross-Type-Warnings
#      sind Info, kein Abbruch.
#   3. profile gegen eine KORRUPTE Quelle (ein Text-Wert in einer INTEGER-
#      Spalte) -> Abbruch (Exit 3); belegt das (b)-Tripwire.
#   4. bulk gegen das frische Ziel -> grün (Exit 0); Row-Count-Parität je
#      Tabelle + Sequenz-Erhalt (kein PK-Kollision).
#   5. bulk erneut gegen das nicht-leere Ziel -> Abbruch (Exit 1); belegt den
#      --on-conflict-abort-Guard (kein Doppel-Load).
#
# Ephemere Ressourcen (eigenes Netz + PG-Container, trap-cleanup). Opt-in
# (NICHT in `make gates`). Exit 0 nur, wenn alle drei Erwartungen zutreffen.

set -euo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
PG_IMAGE="${PG_IMAGE:-postgres:17-alpine}"
SQLITE_IMAGE="${SQLITE_IMAGE:-keinos/sqlite3}"
V1_SQLITE="${REPO_ROOT}/apps/api/internal/storage/migrations/V1__m_trace.sql"
V1_PG="${REPO_ROOT}/apps/api/internal/storage/migrations/postgres/V1__m_trace.sql"

command -v docker  >/dev/null 2>&1 || { echo "[smoke-cutover] docker fehlt"  >&2; exit 2; }
command -v python3 >/dev/null 2>&1 || { echo "[smoke-cutover] python3 fehlt" >&2; exit 2; }
[[ -f "$V1_SQLITE" ]] || { echo "[smoke-cutover] $V1_SQLITE fehlt" >&2; exit 2; }
[[ -f "$V1_PG" ]]     || { echo "[smoke-cutover] $V1_PG fehlt" >&2; exit 2; }

RUN_ID="$$"
NET="mtrace-cutover-net-${RUN_ID}"
DB="mtrace-cutover-pg-${RUN_ID}"
WORK="$(mktemp -d)"
PG_USER="mtrace"; PG_PASS="labpass"; PG_DB="mtrace"

cleanup() {
  docker rm -f "${DB}" >/dev/null 2>&1 || true
  docker network rm "${NET}" >/dev/null 2>&1 || true
  rm -rf "${WORK}"
}
trap cleanup EXIT

echo "▶ smoke-cutover: cutover doctor + profile gegen ephemeres Lab"

# --- Quellen materialisieren (gesund + korrupt) ---------------------------
cp "$V1_SQLITE" "${WORK}/V1.sql"
docker run --rm --user "$(id -u):$(id -g)" --entrypoint sh -v "${WORK}:/work" "$SQLITE_IMAGE" \
  -c 'sqlite3 /work/live.db < /work/V1.sql'
docker run --rm --user "$(id -u):$(id -g)" --entrypoint sqlite3 -v "${WORK}:/work" "$SQLITE_IMAGE" /work/live.db "
INSERT INTO projects(project_id) VALUES ('demo') ON CONFLICT DO NOTHING;
INSERT INTO playback_events(project_id,session_id,event_name,client_timestamp,server_received_at,sequence_number,sdk_name,sdk_version,schema_version,delivery_status,time_skew_warning)
VALUES ('demo','s-1','play','2026-07-12T00:00:01Z','2026-07-12T00:00:01Z',1,'js','1.0','1','accepted',0),
       ('demo','s-1','pause','2026-07-12T00:00:02Z','2026-07-12T00:00:02Z',2,'js','1.0','1','accepted',0);
INSERT INTO srt_health_samples(project_id,stream_id,connection_id,collected_at,ingested_at,rtt_ms,packet_loss_total,retransmissions_total,available_bandwidth_bps,source_status,source_error_code,connection_state,health_state)
VALUES ('demo','s-1','c1','t','t',12.5,0,0,1000000,'ok','none','connected','healthy'),
       ('demo','s-1','c1','t','t',13.0,1,0,1000000,'ok','none','connected','healthy');"
cp "${WORK}/live.db" "${WORK}/corrupt.db"
# Text-Wert in die INTEGER-Spalte sequence_number (SQLite-Affinity lässt ihn Text).
docker run --rm --user "$(id -u):$(id -g)" --entrypoint sqlite3 -v "${WORK}:/work" "$SQLITE_IMAGE" /work/corrupt.db "
INSERT INTO playback_events(project_id,session_id,event_name,client_timestamp,server_received_at,sequence_number,sdk_name,sdk_version,schema_version,delivery_status,time_skew_warning)
VALUES ('demo','s-1','x','2026-07-12T00:00:03Z','2026-07-12T00:00:03Z','BROKEN','js','1.0','1','accepted',0);"

# --- Ziel-PG (frisch, PG-DDL angewendet) ----------------------------------
docker network create "${NET}" >/dev/null
docker run -d --name "${DB}" --network "${NET}" \
  -e POSTGRES_USER="${PG_USER}" -e POSTGRES_PASSWORD="${PG_PASS}" -e POSTGRES_DB="${PG_DB}" \
  "${PG_IMAGE}" >/dev/null
echo "  warte auf stabiles Postgres (übersteht initdb-Restart) ..."
ready=0
for _ in $(seq 1 40); do
  if docker exec "${DB}" psql -U "${PG_USER}" -d "${PG_DB}" -tAc 'select 1' 2>/dev/null | grep -q 1; then
    sleep 1
    docker exec "${DB}" psql -U "${PG_USER}" -d "${PG_DB}" -tAc 'select 1' 2>/dev/null | grep -q 1 && { ready=1; break; }
  fi
  sleep 1
done
[[ "${ready}" -eq 1 ]] || { echo "[smoke-cutover] Postgres nicht bereit" >&2; exit 1; }
docker cp "$V1_PG" "${DB}:/tmp/V1_pg.sql"
docker exec "${DB}" psql -U "${PG_USER}" -d "${PG_DB}" -q -f /tmp/V1_pg.sql >/dev/null

DSN="postgres://${PG_USER}:${PG_PASS}@${DB}:5432/${PG_DB}?sslmode=disable"
CUTOVER="${REPO_ROOT}/scripts/cutover-sqlite-postgres.sh"
fail=0
run_case() { # name expected_rc  ENV...
  local name="$1" want="$2"; shift 2
  set +e
  ( "$@" bash "${CUTOVER}" "${name}" >"${WORK}/case.out" 2>&1 )
  local rc=$?
  set -e
  sed 's/^/    /' "${WORK}/case.out"
  if [[ "$rc" -eq "$want" ]]; then
    echo "  ✔ ${name} -> Exit ${rc} (erwartet ${want})"
  else
    echo "  ✘ ${name} -> Exit ${rc} (erwartet ${want})" >&2; fail=1
  fi
}

echo "  [1/5] doctor gegen gesunde Quelle + frisches (leeres) Ziel (erwartet Exit 0)"
run_case "doctor" 0 env SQLITE_DB="${WORK}/live.db" PG_DSN="${DSN}" PG_NETWORK="${NET}"
echo "  [2/5] profile gegen gesunde Quelle (erwartet Exit 0)"
run_case "profile" 0 env SQLITE_DB="${WORK}/live.db"
echo "  [3/5] profile gegen KORRUPTE Quelle (erwartet Exit 3, (b)-Tripwire)"
run_case "profile" 3 env SQLITE_DB="${WORK}/corrupt.db"
echo "  [4/5] bulk gegen frisches Ziel (erwartet Exit 0: Parität + Sequenz-Erhalt)"
run_case "bulk" 0 env SQLITE_DB="${WORK}/live.db" PG_DSN="${DSN}" PG_NETWORK="${NET}"
echo "  [5/5] bulk erneut → Ziel nicht leer, --on-conflict abort (erwartet Exit 1)"
run_case "bulk" 1 env SQLITE_DB="${WORK}/live.db" PG_DSN="${DSN}" PG_NETWORK="${NET}"

if [[ "$fail" -ne 0 ]]; then
  echo "✘ smoke-cutover: mindestens eine Erwartung verfehlt." >&2
  exit 1
fi
echo "✔ smoke-cutover: doctor + profile (gesund/Korrupt-Tripwire) + bulk (Parität/Sequenz-Erhalt + abort-Guard) — alle Erwartungen erfüllt."
