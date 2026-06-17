#!/usr/bin/env bash
set -euo pipefail

# Load-/Soak-Smoke gegen das Core-Lab (opt-in, nicht-blockierend).
#
# Belegt die Lab-Lastfähigkeit der Ingest-/Persistenz-/Read-Kette
# (NF-20/NF-22/NF-23): treibt per k6
# (scripts/load/playback-events.k6.js) Playback-Event-Batches gegen
# /api/playback-events und prüft anschließend per Readback-
# Reconciliation, dass jedes akzeptierte Event auch persistiert wurde
# (kein stiller Verlust). Gezählt werden die TATSÄCHLICH in
# playback_events liegenden Events — das `events[]`-Array des
# Detail-Endpoints (paginiert), NICHT der Session-`event_count` (der
# wird im Upsert VOR dem Event-Append getickt, eigene Tx, taugt nicht
# als Persistenz-Beleg) und NICHT der Prometheus-Counter.
#
# Zwei Modi:
#   MODE=capacity Rate-Limit pro Project hochgesetzt
#                  (MTRACE_RATE_LIMIT_CAPACITY/-REFILL ins Lab injiziert)
#                  -> misst die echte Ingest-/Persistenz-Kapazität.
#   MODE=contract Default-Limit (100/s) aktiv -> verifiziert, dass der
#                  Limiter greift (429) ohne stillen Verlust.
#
# Destruktiv: setzt das Core-Lab-SQLite-Volume zurück (frische DB pro
# Lauf, damit die Reconciliation nur diesen Lauf zählt).
#
# Last-Profil (orthogonal zu MODE):
#   LOAD_PROFILE=closed --vus N (Korrektheits-Gates + Decke finden).
#   LOAD_PROFILE=open constant-arrival-rate (stabile Nightly-SLO);
#                         erfordert MODE=capacity.
#
# ENV: MODE, LOAD_PROFILE, VUS, DURATION, BATCH_SIZE, CAP_CAPACITY,
#      CAP_REFILL, MAX_ERROR_PCT, TARGET_EVENT_RATE, P95_BUDGET_MS,
#      OPEN_MAX_VUS, OPEN_PREALLOC_VUS, BASE_URL, PROJECT_TOKEN,
#      SESSION_PREFIX, SMOKE_LOAD_AUTOSTART, K6_IMAGE, SQLITE_IMAGE,
#      SQLITE_DB_PATH, RETENTION_PROBE, N_PROBES, RETENTION_P95_BUDGET_MS,
#      SOAK_MIN_EVENTS.
#
# Manuell gegen ein bereits laufendes Lab: SMOKE_LOAD_AUTOSTART=0.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

MODE="${MODE:-capacity}"
BASE_URL="${BASE_URL:-http://localhost:8080}"
PROJECT_TOKEN="${PROJECT_TOKEN:-demo-token}"
SESSION_PREFIX="${SESSION_PREFIX:-load-vu}"
VUS="${VUS:-20}"
DURATION="${DURATION:-30s}"
BATCH_SIZE="${BATCH_SIZE:-20}"
CAP_CAPACITY="${CAP_CAPACITY:-1000000}"
CAP_REFILL="${CAP_REFILL:-1000000}"
SMOKE_LOAD_AUTOSTART="${SMOKE_LOAD_AUTOSTART:-1}"
K6_IMAGE="${K6_IMAGE:-grafana/k6}"
# Readback-COUNT (Autostart-Pfad, R-25): eine eigene sqlite3-Instanz zählt
# die persistierten Events direkt im api-Volume (O(1)-Read statt
# O(N)-HTTP-Pagination, die bei Soak-Volumen den CI-Job-Cap sprengte).
# `SQLITE_DB_PATH` muss `MTRACE_SQLITE_PATH` aus docker-compose.yml folgen.
SQLITE_IMAGE="${SQLITE_IMAGE:-keinos/sqlite3}"
SQLITE_DB_PATH="${SQLITE_DB_PATH:-/var/lib/mtrace/m-trace.db}"
# Fehler-Obergrenze (Status != 202/429) als Anteil aller Events. An der
# SQLite-Sättigung sind einzelne explizite Fehler erwartbar; nur eine
# katastrophale Quote bricht den Smoke. Der harte Gate bleibt die
# Reconciliation (kein stiller Verlust).
MAX_ERROR_PCT="${MAX_ERROR_PCT:-5}"
# Last-Profil: closed (--vus, Korrektheits-Gates + Decke finden) oder
# open (constant-arrival-rate, stabile Nightly-SLO). open gibt
# TARGET_EVENT_RATE Events/s vor und prüft p95 < P95_BUDGET_MS +
# dropped_iterations < 1 %.
LOAD_PROFILE="${LOAD_PROFILE:-closed}"
TARGET_EVENT_RATE="${TARGET_EVENT_RATE:-400}"
P95_BUDGET_MS="${P95_BUDGET_MS:-1000}"
OPEN_MAX_VUS="${OPEN_MAX_VUS:-100}"
OPEN_PREALLOC_VUS="${OPEN_PREALLOC_VUS:-50}"
# Soak/Retention-Probe (ADR-0005-Trigger #3): nach dem Load die Read-
# Retention-Latenz (Liste + Detail-mit-Events) messen und gegen 2 s
# bewerten. Belastbar erst ab SOAK_MIN_EVENTS (10 Mio) persistierten
# Events — darunter validiert die Probe nur den Mechanismus (Nightly-Job
# fährt die lange DURATION für >=10 Mio).
RETENTION_PROBE="${RETENTION_PROBE:-0}"
N_PROBES="${N_PROBES:-50}"
RETENTION_P95_BUDGET_MS="${RETENTION_P95_BUDGET_MS:-2000}"
SOAK_MIN_EVENTS="${SOAK_MIN_EVENTS:-10000000}"

for dep in docker curl python3; do
  command -v "$dep" >/dev/null 2>&1 || {
    echo "[load-smoke] missing dependency: $dep" >&2
    exit 2
  }
done
docker compose version >/dev/null 2>&1 || {
  echo "[load-smoke] 'docker compose' (v2) nicht verfügbar (v1 'docker-compose' wird nicht unterstützt)" >&2
  exit 2
}

case "$MODE" in
  capacity | contract) ;;
  *)
    echo "[load-smoke] MODE='$MODE' ungültig — erlaubt: capacity | contract" >&2
    exit 2
    ;;
esac

# Der Limit-Override greift nur über einen frisch erzeugten Container.
# Bei AUTOSTART=0 (gegen ein laufendes Lab) sieht der bestehende
# Container das Export nie -> capacity liefe still als contract durch
# (misleading green). Daher harter Abbruch statt falschem Grün.
if [ "$MODE" = "capacity" ] && [ "$SMOKE_LOAD_AUTOSTART" != "1" ]; then
  echo "[load-smoke] MODE=capacity erfordert SMOKE_LOAD_AUTOSTART=1 (Override braucht frischen Container)" >&2
  exit 2
fi

case "$LOAD_PROFILE" in
  closed | open) ;;
  *)
    echo "[load-smoke] LOAD_PROFILE='$LOAD_PROFILE' ungültig — erlaubt: closed | open" >&2
    exit 2
    ;;
esac
# open offeriert TARGET_EVENT_RATE über der Default-Decke -> braucht das
# angehobene Limit, sonst hängt die Last am 100/s-Limiter (429) und die
# SLO-Messung wäre sinnlos.
if [ "$LOAD_PROFILE" = "open" ] && [ "$MODE" != "capacity" ]; then
  echo "[load-smoke] LOAD_PROFILE=open erfordert MODE=capacity (sonst hängt die offered rate am Limiter)" >&2
  exit 2
fi

tmpdir="$(mktemp -d)"
cleanup() {
  status=$?
  rm -rf "$tmpdir"
  if [ "$SMOKE_LOAD_AUTOSTART" = "1" ]; then
    echo "[load-smoke] cleanup: docker compose down -v"
    (cd "$ROOT_DIR" && docker compose down -v >/dev/null 2>&1) || true
  fi
  exit $status
}
trap cleanup EXIT

if [ "$MODE" = "capacity" ]; then
  export MTRACE_RATE_LIMIT_CAPACITY="$CAP_CAPACITY"
  export MTRACE_RATE_LIMIT_REFILL="$CAP_REFILL"
  echo "[load-smoke] mode=capacity (rate limit raised to ${CAP_CAPACITY}/${CAP_REFILL} per project)"
else
  unset MTRACE_RATE_LIMIT_CAPACITY MTRACE_RATE_LIMIT_REFILL 2>/dev/null || true
  echo "[load-smoke] mode=contract (default 100/s limiter active)"
fi

if [ "$SMOKE_LOAD_AUTOSTART" = "1" ]; then
  echo "[load-smoke] fresh lab: docker compose down -v && up -d --build (SQLite-Volume wird zurückgesetzt)"
  (cd "$ROOT_DIR" && docker compose down -v >/dev/null 2>&1) || true
  if ! (cd "$ROOT_DIR" && docker compose up -d --build) >/dev/null 2>&1; then
    echo "[load-smoke] docker compose up failed (Port 8080 frei?)" >&2
    exit 1
  fi
fi

echo "[load-smoke] warte auf API-Health..."
ok=0
for _ in $(seq 1 60); do
  s="$(curl -sS -o /dev/null -w '%{http_code}' "${BASE_URL}/api/health" 2>/dev/null || echo 000)"
  if [ "$s" = "200" ]; then ok=1; break; fi
  sleep 1
done
[ "$ok" = "1" ] || {
  echo "[load-smoke] API nicht healthy nach 60s" >&2
  exit 1
}

k6_run_args=(run)
k6_env_args=(
  -e BASE_URL="$BASE_URL" -e PROJECT_TOKEN="$PROJECT_TOKEN"
  -e BATCH_SIZE="$BATCH_SIZE" -e SESSION_PREFIX="$SESSION_PREFIX"
)
if [ "$LOAD_PROFILE" = "open" ]; then
  echo "[load-smoke] profile=open (SLO): ~${TARGET_EVENT_RATE} ev/s offered, p95<${P95_BUDGET_MS}ms, ${DURATION}"
  k6_env_args+=(
    -e LOAD_PROFILE=open -e TARGET_EVENT_RATE="$TARGET_EVENT_RATE"
    -e P95_BUDGET_MS="$P95_BUDGET_MS" -e OPEN_MAX_VUS="$OPEN_MAX_VUS"
    -e OPEN_PREALLOC_VUS="$OPEN_PREALLOC_VUS" -e DURATION="$DURATION"
  )
  recon_sessions="$OPEN_MAX_VUS"
else
  echo "[load-smoke] profile=closed: ${VUS} VUs / ${DURATION}, batch=${BATCH_SIZE}"
  k6_run_args+=(--vus "$VUS" --duration "$DURATION")
  recon_sessions="$VUS"
fi

# k6-Exit getrennt erfassen: im open-Profil bedeutet ein Threshold-Bruch
# (p95 / dropped_iterations) Exit != 0 = SLO verfehlt; trotzdem sollen
# Reconciliation + Report noch laufen, statt vorzeitig (set -e) abzubrechen.
set +e
docker run --rm --network host \
  --user "$(id -u):$(id -g)" \
  -v "$ROOT_DIR/scripts/load:/scripts:ro" \
  -v "$tmpdir:/work" \
  "$K6_IMAGE" "${k6_run_args[@]}" "${k6_env_args[@]}" \
  /scripts/playback-events.k6.js 2>&1 | tee "$tmpdir/k6.log"
k6_rc=${PIPESTATUS[0]}
set -e

[ -f "$tmpdir/summary.json" ] || {
  echo "[load-smoke] keine k6-summary.json erzeugt" >&2
  exit 1
}

# Readback: die TATSÄCHLICH in `playback_events` persistierten Events
# dieses Laufs zählen — der harte „kein stiller Verlust"-Beleg
# (persisted >= accepted). Bewusst NICHT der Session-`event_count` (wird
# im Upsert VOR dem Event-Append getickt, eigene Tx -> overcountet, würde
# echten Verlust maskieren) und NICHT die Cursor-paginierte Listen-API
# (truncatet -> Falschalarm).
#
# Zwei Pfade, derselbe Beleg (beide zählen Zeilen in `playback_events`):
#   AUTOSTART=1 (frische, von uns kontrollierte Lab-DB): direkter
#     SQLite-`COUNT(*)` gegen das api-Volume — O(1)-Read. Der frühere
#     HTTP-Readback paginierte den Detail-Endpoint zu je 1000 Events über
#     ALLE Sessions; bei Soak-Volumen (~45 Mio Events) sind das ~45k
#     sequentielle Reads ~2 h -> sprengte den 6h-CI-Job-Cap (R-25). Der
#     COUNT liest dieselbe Tabelle, aus der auch der Detail-Endpoint
#     serviert, nur direkt.
#   AUTOSTART=0 (fremdes, laufendes Lab ohne garantierten Volume-/
#     Container-Zugriff): portabler HTTP-Readback (events[]-Array,
#     Cursor-paginiert, summiert über PREFIX-1..PREFIX-VUS). Für ad-hoc-
#     Läufe gedacht, nicht für Soak-Volumina.
# Lesefehler -> Exit 3 (INCONCLUSIVE), klar getrennt vom Verlust-FAIL
# (Exit 1): ein Readback-Fehler wird NIE als Datenverlust maskiert.
if [ "$SMOKE_LOAD_AUTOSTART" = "1" ]; then
  api_cid="$(cd "$ROOT_DIR" && docker compose ps -q api 2>/dev/null || true)"
  if [ -z "$api_cid" ]; then
    echo "[load-smoke] INCONCLUSIVE: api-Container für Readback-COUNT nicht gefunden (Lesefehler != Datenverlust)" >&2
    exit 3
  fi
  # GLOB statt LIKE: behandelt '_' literal + ist case-sensitiv, matcht also
  # exakt die Lauf-Sessions `${SESSION_PREFIX}-<n>` (frische DB -> keine
  # Fremdsessions). --volumes-from teilt das api-Volume inkl. -wal/-shm;
  # committete WAL-Frames sind für den (jetzt idle) Reader sichtbar.
  # --user 0:0: die Volume-Files gehören dem nonroot-api-User (UID 65532),
  # root liest -wal/-shm permission-frei.
  # --entrypoint sqlite3: keinos/sqlite3 (Default) hat ENTRYPOINT=tini +
  # CMD=sqlite3 — ohne Override ersetzen die Args (DB+SQL) das CMD und tini
  # exec't den DB-Pfad als Programm. Override erzwingt sqlite3 (muss im
  # SQLITE_IMAGE in PATH liegen). stderr getrennt halten, damit eine
  # etwaige WAL-Checkpoint-Warnung die Zahl auf stdout nicht verunreinigt;
  # bei Fehler wird sie in die Diagnose gehoben.
  readback_err="$tmpdir/readback-count.err"
  persisted="$(
    docker run --rm --user 0:0 --entrypoint sqlite3 \
      --volumes-from "$api_cid" "$SQLITE_IMAGE" \
      "$SQLITE_DB_PATH" \
      "SELECT count(*) FROM playback_events WHERE session_id GLOB '${SESSION_PREFIX}-*';" \
      2>"$readback_err" || true
  )"
  persisted="$(printf '%s' "$persisted" | tr -d '[:space:]')"
  if ! printf '%s' "$persisted" | grep -Eq '^[0-9]+$'; then
    echo "[load-smoke] INCONCLUSIVE: Readback-COUNT fehlgeschlagen (sqlite '${SQLITE_IMAGE}' gegen ${SQLITE_DB_PATH}): $(tr '\n' ' ' < "$readback_err" 2>/dev/null) — kein Verlust-Urteil (Lesefehler != Datenverlust)" >&2
    exit 3
  fi
  echo "[load-smoke] readback: SQLite COUNT(*) WHERE session_id GLOB '${SESSION_PREFIX}-*' = ${persisted}"
else
  if ! persisted="$(
    VUS="$recon_sessions" BASE_URL="$BASE_URL" PROJECT_TOKEN="$PROJECT_TOKEN" \
      SESSION_PREFIX="$SESSION_PREFIX" python3 - <<'PY'
import os, sys, json, urllib.request, urllib.parse, urllib.error
base = os.environ["BASE_URL"].rstrip("/")
hdr = {"X-MTrace-Token": os.environ["PROJECT_TOKEN"]}
prefix = os.environ["SESSION_PREFIX"]
vus = int(os.environ["VUS"])
total = 0
for n in range(1, vus + 1):
    sid = f"{prefix}-{n}"
    cursor = None
    while True:
        q = {"events_limit": "1000"}
        if cursor:
            q["events_cursor"] = cursor
        url = f"{base}/api/stream-sessions/{urllib.parse.quote(sid)}?{urllib.parse.urlencode(q)}"
        try:
            with urllib.request.urlopen(urllib.request.Request(url, headers=hdr), timeout=15) as r:
                d = json.load(r)
        except urllib.error.HTTPError as e:
            if e.code == 404:
                break  # Session nie angelegt -> 0, naechste Session
            print(f"readback HTTPError {e.code} fuer {sid}", file=sys.stderr)
            sys.exit(3)  # inconclusive, NICHT als Verlust werten
        except Exception as e:  # Timeout/Reset/JSON -> inconclusive
            print(f"readback error fuer {sid}: {e}", file=sys.stderr)
            sys.exit(3)
        total += len(d.get("events", []))
        cursor = d.get("next_cursor")
        if not cursor:
            break
print(total)
PY
  )"; then
    echo "[load-smoke] INCONCLUSIVE: Readback-Lesefehler — kein Verlust-Urteil möglich (Lesefehler != Datenverlust)" >&2
    exit 3
  fi
fi

python3 - "$tmpdir/summary.json" "$persisted" "$MODE" "$SESSION_PREFIX" "$MAX_ERROR_PCT" "$LOAD_PROFILE" "$k6_rc" "$P95_BUDGET_MS" <<'PY'
import sys, json
summ_path, persisted_s, mode, prefix, max_err_s, profile, k6_rc_s, p95_budget_s = sys.argv[1:9]
persisted = int(persisted_s)
max_err_pct = float(max_err_s)
k6_rc = int(k6_rc_s)
p95_budget = float(p95_budget_s)
with open(summ_path) as f:
    m = json.load(f)["metrics"]

def cnt(name):
    return int(m.get(name, {}).get("values", {}).get("count", 0))

def rate(name):
    return m.get(name, {}).get("values", {}).get("rate", 0.0)

accepted = cnt("mtrace_events_accepted")
rate_limited = cnt("mtrace_events_rate_limited")
rejected = cnt("mtrace_events_rejected")
sent = cnt("mtrace_events_sent")
dropped = cnt("dropped_iterations")
# Fehlerquote über die NICHT-gedrosselten Versuche (202 + echte Fehler).
# 429 bleibt draußen, sonst verdünnen Millionen 429 im contract-Modus die
# Quote bis zur Bedeutungslosigkeit.
attempts = accepted + rejected
err_pct = (100.0 * rejected / attempts) if attempts else 0.0
reqs = m.get("http_reqs", {}).get("values", {})
dur = m.get("http_req_duration", {}).get("values", {})

# Cross-Check: k6-Sendezähler muss der Summe der Status-Zähler
# entsprechen; sonst ist die Summary-Auswertung verrutscht.
if sent and sent != accepted + rate_limited + rejected:
    print(f"[load-smoke] WARN: k6 sent {sent} != "
          f"accepted+rate_limited+rejected {accepted + rate_limited + rejected} "
          f"(Summary-Parse-Drift?)")

print(f"[load-smoke] req/s={reqs.get('rate', 0):.0f}  "
      f"p95={dur.get('p(95)', 0):.2f}ms  p90={dur.get('p(90)', 0):.2f}ms  "
      f"max={dur.get('max', 0):.1f}ms")
print(f"[load-smoke] accepted(202)={accepted} ({rate('mtrace_events_accepted'):.0f}/s)  "
      f"rate_limited(429)={rate_limited}  errors(!=202/429)={rejected} ({err_pct:.2f}% der 202+Fehler)")
ambiguous = persisted - accepted
print(f"[load-smoke] persisted (readback, {prefix}-* sessions)={persisted}  "
      f"(at-least-once-Überschuss: {ambiguous})")
if profile == "open":
    print(f"[load-smoke] SLO (open): p95={dur.get('p(95)', 0):.1f}ms (budget {p95_budget:.0f}ms)  "
          f"dropped_iterations={dropped}  k6_thresholds={'PASS' if k6_rc == 0 else 'FAIL'}")

# Harter Gate: KEIN stiller Verlust -> jedes client-bestätigte (202)
# Event MUSS persistiert sein, also persisted >= accepted. Ein
# Überschuss (persisted > accepted) ist KEIN Verlust, sondern
# at-least-once unter Überlast: der Server persistierte, bevor der
# Client ein Timeout/5xx sah. persisted < accepted wäre echter stiller
# Verlust.
fail = []
if persisted < accepted:
    fail.append(f"STILLER VERLUST: persisted {persisted} < accepted {accepted} "
                f"(Delta {accepted - persisted} client-bestätigte Events fehlen)")
if err_pct > max_err_pct:
    fail.append(f"Fehlerquote {err_pct:.2f}% > {max_err_pct}% "
                f"({rejected} Events mit Status != 202/429) -> nicht graceful")
if mode == "contract" and rate_limited == 0:
    fail.append("contract-Modus: kein 429 -> Limiter hat nicht gegriffen")
if mode == "capacity" and accepted == 0:
    fail.append("capacity-Modus: 0 akzeptiert -> Limit-Override nicht aktiv?")
if profile == "open" and k6_rc != 0:
    fail.append(f"SLO verfehlt (k6 exit={k6_rc}; Threshold p95<{p95_budget:.0f}ms / dropped<1%): "
                f"gemessen p95 {dur.get('p(95)', 0):.1f}ms, dropped_iterations={dropped}")

if fail:
    print("[load-smoke] FAIL: " + "; ".join(fail), file=sys.stderr)
    sys.exit(1)
note = "" if ambiguous == 0 else f" (+{ambiguous} at-least-once unter Last)"
print(f"[load-smoke] OK -- kein stiller Verlust (persisted {persisted} >= accepted {accepted}{note}); "
      f"Fehlerquote {err_pct:.2f}% <= {max_err_pct}% (graceful).")
PY

if [ "$RETENTION_PROBE" = "1" ]; then
  echo "[load-smoke] retention probe (ADR-0005 Trigger #3): ${N_PROBES} Reads, Budget ${RETENTION_P95_BUDGET_MS}ms"
  PERSISTED="$persisted" BASE_URL="$BASE_URL" PROJECT_TOKEN="$PROJECT_TOKEN" \
    SESSION_PREFIX="$SESSION_PREFIX" N_PROBES="$N_PROBES" \
    RETENTION_P95_BUDGET_MS="$RETENTION_P95_BUDGET_MS" SOAK_MIN_EVENTS="$SOAK_MIN_EVENTS" \
    python3 - <<'PY'
import os, sys, time, urllib.request, urllib.parse
base = os.environ["BASE_URL"].rstrip("/")
hdr = {"X-MTrace-Token": os.environ["PROJECT_TOKEN"]}
prefix = os.environ["SESSION_PREFIX"]
persisted = int(os.environ["PERSISTED"])
n = int(os.environ["N_PROBES"])
budget = float(os.environ["RETENTION_P95_BUDGET_MS"])
soak_min = int(os.environ["SOAK_MIN_EVENTS"])

def p95_ms(url):
    samples = []
    for _ in range(n):
        t0 = time.monotonic()
        try:
            with urllib.request.urlopen(urllib.request.Request(url, headers=hdr), timeout=30) as r:
                r.read()
        except Exception as e:
            print(f"[load-smoke] retention probe error: {e}", file=sys.stderr)
            return None
        samples.append((time.monotonic() - t0) * 1000.0)
    samples.sort()
    return samples[min(len(samples) - 1, round(0.95 * (len(samples) - 1)))]

list_p95 = p95_ms(f"{base}/api/stream-sessions?limit=200")
detail_p95 = p95_ms(
    f"{base}/api/stream-sessions/{urllib.parse.quote(prefix + '-1')}?events_limit=1000"
)
lp = f"{list_p95:.1f}ms" if list_p95 is not None else "n/a"
dp = f"{detail_p95:.1f}ms" if detail_p95 is not None else "n/a"
print(f"[load-smoke] retention p95: list={lp} detail-events={dp} (budget {budget:.0f}ms)")
# WICHTIG (Scope/Ehrlichkeit): beide Probes sind keyset-indizierte Reads
# -> Latenz größenunabhängig. Sie sind ein PROXY für ADR-0005 Trigger #3
# ("Queries über >10 Mio Events"), KEIN Korpus-Scan. Die aktuelle
# Read-API serviert keine Full-Scan-/Aggregat-/Time-Range-Query; kommt je
# eine dazu, MUSS die Probe um genau diese ergänzt werden, sonst
# over-claim. Verdikt daher als "(Proxy)" + "indizierte Hot-Read-p95".
vals = [v for v in (list_p95, detail_p95) if v is not None]
worst = max(vals) if vals else None
if persisted < soak_min:
    print(f"[load-smoke] ADR-0005 Trigger #3: INCONCLUSIVE — nur {persisted} Events (< {soak_min}); "
          f"Mechanismus validiert, belastbares Urteil erst im Nightly-Soak.")
elif worst is None:
    print("[load-smoke] ADR-0005 Trigger #3: INCONCLUSIVE — Retention-Probe fehlgeschlagen (kein p95).")
elif worst < budget:
    print(f"[load-smoke] ADR-0005 Trigger #3: NICHT ausgelöst (Proxy) — indizierte Hot-Read-p95 "
          f"{worst:.0f}ms < {budget:.0f}ms bei {persisted} Events. Keyset-indiziert = "
          f"größenunabhängig; kein Korpus-Scan-Query in der API -> Proxy-/Zukunfts-Messung.")
else:
    print(f"[load-smoke] ADR-0005 Trigger #3: AUSGELÖST (Proxy) — indizierte Hot-Read-p95 "
          f"{worst:.0f}ms >= {budget:.0f}ms bei {persisted} Events -> Postgres-Pfad evaluieren.")
PY
fi

echo "[load-smoke] done."
