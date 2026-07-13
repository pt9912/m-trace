#!/usr/bin/env bash
# smoke-cutover.sh — opt-in-Smoke für den SQLite->Postgres-Cutover (ADR-0007).
#
# Verifiziert doctor + Phase 0 (profile) + Phase 1 (bulk) von
# scripts/cutover-sqlite-postgres.sh gegen ein ephemeres Lab:
#   1. doctor gegen frische, per PG-DDL angelegte Ziel-DB + gesunde Quelle
#      -> alle Pre-Flight-Checks grün (Exit 0).
#   2. doctor gegen eine READ-ONLY-Quelle (Datei+Verzeichnis nicht schreibbar)
#      -> grün (Exit 0); belegt die gelockerte Probe (d-migrate >= 0.9.12
#      öffnet die Quelle read-only, keine Schreibbarkeits-Anforderung mehr).
#   3. profile gegen die gesunde Quelle -> grün (Exit 0); Cross-Type-Warnings
#      sind Info, kein Abbruch.
#   4. profile gegen die READ-ONLY-Quelle -> grün (Exit 0); belegt die
#      read-only-Öffnung (file:?mode=ro, keine -wal/-shm-Nebendateien).
#   5. profile gegen eine KORRUPTE Quelle (ein Text-Wert in einer INTEGER-
#      Spalte) -> Abbruch (Exit 3); belegt das (b)-Tripwire.
#   6. bulk gegen das frische Ziel -> grün (Exit 0); Row-Count-Parität je
#      Tabelle + Sequenz-Erhalt (kein PK-Kollision).
#   7. bulk erneut gegen das nicht-leere Ziel -> Abbruch (Exit 1); belegt den
#      --on-conflict-abort-Guard (kein Doppel-Load).
#   8. incremental nach neuem Quell-Delta -> grün (Exit 0); zieht das Delta nach,
#      Parität + duplikatfrei.
#   9. incremental erneut -> grün (Exit 0); idempotent (keine Duplikate).
#  10. switch nach mutabler Änderung + append-only Delta -> grün (Exit 0);
#      belegt, dass der finale Re-Sync die Mutation propagiert (Design a).
#
# Ephemere Ressourcen (eigenes Netz + PG-Container, trap-cleanup). Opt-in
# (NICHT in `make gates`). Exit 0 nur, wenn alle Erwartungen zutreffen.

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
  chmod -R u+w "${WORK}" 2>/dev/null || true   # read-only-Quelle-Case (555-Dir)
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
INSERT INTO stream_sessions(session_id,project_id,state,started_at,last_seen_at,event_count,sample_rate_ppm) VALUES ('sess-1','demo','active','t','t',5,1000000);
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
# READ-ONLY-Quelle: Kopie in nicht-schreibbarem Verzeichnis (Datei 444, Dir 555)
# — unter d-migrate < 0.9.12 starb daran profile/transfer (SQLITE_READONLY).
mkdir "${WORK}/ro"
cp "${WORK}/live.db" "${WORK}/ro/live.db"
chmod 444 "${WORK}/ro/live.db"
chmod 555 "${WORK}/ro"

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

echo "  [1/10] doctor gegen gesunde Quelle + frisches (leeres) Ziel (erwartet Exit 0)"
run_case "doctor" 0 env SQLITE_DB="${WORK}/live.db" PG_DSN="${DSN}" PG_NETWORK="${NET}"
echo "  [2/10] doctor gegen READ-ONLY-Quelle (erwartet Exit 0, gelockerte Probe)"
run_case "doctor" 0 env SQLITE_DB="${WORK}/ro/live.db" PG_DSN="${DSN}" PG_NETWORK="${NET}"
echo "  [3/10] profile gegen gesunde Quelle (erwartet Exit 0)"
run_case "profile" 0 env SQLITE_DB="${WORK}/live.db"
echo "  [4/10] profile gegen READ-ONLY-Quelle (erwartet Exit 0, mode=ro-Öffnung)"
run_case "profile" 0 env SQLITE_DB="${WORK}/ro/live.db"
echo "  [5/10] profile gegen KORRUPTE Quelle (erwartet Exit 3, (b)-Tripwire)"
run_case "profile" 3 env SQLITE_DB="${WORK}/corrupt.db"
echo "  [6/10] bulk gegen frisches Ziel (erwartet Exit 0: Parität + Sequenz-Erhalt)"
run_case "bulk" 0 env SQLITE_DB="${WORK}/live.db" PG_DSN="${DSN}" PG_NETWORK="${NET}"
echo "  [7/10] bulk erneut → Ziel nicht leer, --on-conflict abort (erwartet Exit 1)"
run_case "bulk" 1 env SQLITE_DB="${WORK}/live.db" PG_DSN="${DSN}" PG_NETWORK="${NET}"

# Delta in der Quelle erzeugen (neue playback_events + srt-Row), dann inkrementell.
docker run --rm --user "$(id -u):$(id -g)" --entrypoint sqlite3 -v "${WORK}:/work" "$SQLITE_IMAGE" /work/live.db "
INSERT INTO playback_events(project_id,session_id,event_name,client_timestamp,server_received_at,sequence_number,sdk_name,sdk_version,schema_version,delivery_status,time_skew_warning)
VALUES ('demo','s-1','seek','t','t',3,'js','1.0','1','accepted',0),('demo','s-1','end','t','t',4,'js','1.0','1','accepted',0);
INSERT INTO srt_health_samples(project_id,stream_id,connection_id,collected_at,ingested_at,rtt_ms,packet_loss_total,retransmissions_total,available_bandwidth_bps,source_status,source_error_code,connection_state,health_state)
VALUES ('demo','s-1','c1','t','t',9.0,0,0,1000000,'ok','none','connected','healthy');"
echo "  [8/10] incremental zieht das Delta nach (erwartet Exit 0: Parität + duplikatfrei)"
run_case "incremental" 0 env SQLITE_DB="${WORK}/live.db" PG_DSN="${DSN}" PG_NETWORK="${NET}"
echo "  [9/10] incremental erneut → idempotent (erwartet Exit 0, keine Duplikate)"
run_case "incremental" 0 env SQLITE_DB="${WORK}/live.db" PG_DSN="${DSN}" PG_NETWORK="${NET}"

# Mutable Änderung + append-only Delta erzeugen. Der --on-conflict-skip-
# incremental fängt die Mutation NICHT — erst der quiescte Switch (Design a).
docker run --rm --user "$(id -u):$(id -g)" --entrypoint sqlite3 -v "${WORK}:/work" "$SQLITE_IMAGE" /work/live.db "
UPDATE stream_sessions SET event_count=99, state='ended' WHERE session_id='sess-1';
INSERT INTO playback_events(project_id,session_id,event_name,client_timestamp,server_received_at,sequence_number,sdk_name,sdk_version,schema_version,delivery_status,time_skew_warning)
VALUES ('demo','s-1','stop','t','t',5,'js','1.0','1','accepted',0);"
echo "  [10/10] switch → append-only Delta + mutable Re-Sync (erwartet Exit 0)"
run_case "switch" 0 env SQLITE_DB="${WORK}/live.db" PG_DSN="${DSN}" PG_NETWORK="${NET}"
# Beleg für Design (a): der Switch propagiert die mutable Änderung, die der
# skip-incremental NICHT gefangen hätte.
ss="$(docker run --rm -i --network "${NET}" "$PG_IMAGE" psql "$DSN" -tAc "SELECT event_count FROM stream_sessions WHERE session_id='sess-1';" 2>/dev/null || echo '?')"
if [ "$ss" = "99" ]; then
  echo "  ✔ Switch propagierte die mutable Änderung (stream_sessions.event_count=99)"
else
  echo "  ✘ Switch: stream_sessions.event_count=${ss} (erwartet 99)" >&2; fail=1
fi

if [[ "$fail" -ne 0 ]]; then
  echo "✘ smoke-cutover: mindestens eine Erwartung verfehlt." >&2
  exit 1
fi
echo "✔ smoke-cutover: doctor (inkl. read-only-Quelle) + profile (gesund/read-only/Tripwire) + bulk (Parität/Sequenz + abort-Guard) + incremental (Delta + Idempotenz) + switch (append-only Delta + mutable Re-Sync) — alle Erwartungen erfüllt."
