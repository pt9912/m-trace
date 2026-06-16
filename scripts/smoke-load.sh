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
# ENV: MODE, VUS, DURATION, BATCH_SIZE, CAP_CAPACITY, CAP_REFILL,
#      BASE_URL, PROJECT_TOKEN, SESSION_PREFIX, SMOKE_LOAD_AUTOSTART,
#      K6_IMAGE.
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
# Fehler-Obergrenze (Status != 202/429) als Anteil aller Events. An der
# SQLite-Sättigung sind einzelne explizite Fehler erwartbar; nur eine
# katastrophale Quote bricht den Smoke. Der harte Gate bleibt die
# Reconciliation (kein stiller Verlust).
MAX_ERROR_PCT="${MAX_ERROR_PCT:-5}"

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

echo "[load-smoke] k6: ${VUS} VUs / ${DURATION}, batch=${BATCH_SIZE}"
docker run --rm --network host \
  --user "$(id -u):$(id -g)" \
  -v "$ROOT_DIR/scripts/load:/scripts:ro" \
  -v "$tmpdir:/work" \
  "$K6_IMAGE" run \
  --vus "$VUS" --duration "$DURATION" \
  -e BASE_URL="$BASE_URL" -e PROJECT_TOKEN="$PROJECT_TOKEN" \
  -e BATCH_SIZE="$BATCH_SIZE" -e SESSION_PREFIX="$SESSION_PREFIX" \
  /scripts/playback-events.k6.js 2>&1 | tee "$tmpdir/k6.log"

[ -f "$tmpdir/summary.json" ] || {
  echo "[load-smoke] keine k6-summary.json erzeugt" >&2
  exit 1
}

# Readback: die TATSÄCHLICH persistierten Events je Lauf-Session zählen
# — die Länge des `events[]`-Arrays des Detail-Endpoints (kommt aus
# playback_events), Cursor-paginiert mit events_limit=1000, summiert über
# genau VUS gezielte Sessions (PREFIX-1..PREFIX-VUS).
# Bewusst NICHT der Session-`event_count` (wird im Upsert VOR dem Append
# getickt -> kein Persistenz-Beleg, macht den Verlust-Gate tot) und NICHT
# die Cursor-paginierte Listen-API (truncatet -> Falschalarm). Per-Session
# ist zudem immun gegen fremde/Alt-Sessions.
persisted="$(
  VUS="$VUS" BASE_URL="$BASE_URL" PROJECT_TOKEN="$PROJECT_TOKEN" \
    SESSION_PREFIX="$SESSION_PREFIX" python3 - <<'PY'
import os, json, urllib.request, urllib.parse
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
        except Exception:
            break  # 404 / Session nicht angelegt -> 0 fuer diese Session
        total += len(d.get("events", []))
        cursor = d.get("next_cursor")
        if not cursor:
            break
print(total)
PY
)"

python3 - "$tmpdir/summary.json" "$persisted" "$MODE" "$SESSION_PREFIX" "$MAX_ERROR_PCT" <<'PY'
import sys, json
summ_path, persisted_s, mode, prefix, max_err_s = sys.argv[1:6]
persisted = int(persisted_s)
max_err_pct = float(max_err_s)
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

if fail:
    print("[load-smoke] FAIL: " + "; ".join(fail), file=sys.stderr)
    sys.exit(1)
note = "" if ambiguous == 0 else f" (+{ambiguous} at-least-once unter Last)"
print(f"[load-smoke] OK -- kein stiller Verlust (persisted {persisted} >= accepted {accepted}{note}); "
      f"Fehlerquote {err_pct:.2f}% <= {max_err_pct}% (graceful).")
PY

echo "[load-smoke] done."
