#!/usr/bin/env bash
set -euo pipefail

# plan-0.6.0 §3 Tranche 2 + §8 Tranche 7 — SRT-Health-Smoke
# (RAK-41/42, plus opt-in m-trace-API-Probe für RAK-43).
#
# Pflichtpfad (immer aktiv):
#   1. HLS-Manifest erreichbar (Publish-Baseline aus smoke-srt).
#   2. MediaMTX Control-API antwortet `200`.
#   3. `items[]` enthält mindestens eine Verbindung mit
#      `path=srt-test` und `state=publish`.
#   4. Vier RAK-43-Pflichtwerte sind numerisch gesetzt:
#      `msRTT`, `packetsReceivedLoss`, `packetsReceivedRetrans`,
#      `mbpsLinkCapacity` — letzteres muss `> 0` sein, der Rest
#      darf 0 sein (gesundes Lab).
#
# Optionaler m-trace-API-Pfad (`SMOKE_INCLUDE_MTRACE_API=1`):
#   5. `GET /api/srt/health/{stream_id}` antwortet `200`.
#   6. Wire-Format aus spec/backend-api-contract.md §7a.2 enthält
#      die vier RAK-43-Pflichtwerte unter `metrics.{rtt_ms,
#      packet_loss_total, retransmissions_total,
#      available_bandwidth_bps}`.
#   Voraussetzung: m-trace-API läuft auf $MTRACE_API_URL (Default
#   http://localhost:8080) mit aktivem Collector
#   (MTRACE_SRT_SOURCE_URL gesetzt). Der Pfad ist opt-in, weil
#   examples/srt/compose.yaml apps/api nicht startet — der
#   Operator fährt das im Release-Closeout aus releasing.md §2.1
#   gegen ein laufendes `make dev`.
#
# Konvention (examples/README.md):
#   - Project-Name `mtrace-srt`; Smoke räumt nur dieses Compose-
#     Projekt auf.
#   - opt-in (nicht in `make gates`).
#
# Manueller Aufruf (Stack vorher gestartet):
#   docker compose -p mtrace-srt -f examples/srt/compose.yaml up -d --build
#   SMOKE_SRT_AUTOSTART=0 scripts/smoke-srt-health.sh

PROJECT="${PROJECT:-mtrace-srt}"
COMPOSE_FILE="${COMPOSE_FILE:-examples/srt/compose.yaml}"
HLS_URL="${HLS_URL:-http://localhost:8889/srt-test/index.m3u8}"
API_URL="${API_URL:-http://localhost:9998/v3/srtconns/list}"
EXPECTED_PATH="${EXPECTED_PATH:-srt-test}"
EXPECTED_STATE="${EXPECTED_STATE:-publish}"
WAIT_SECONDS="${WAIT_SECONDS:-45}"
SMOKE_SRT_AUTOSTART="${SMOKE_SRT_AUTOSTART:-1}"

# Opt-in m-trace-API-Probe (RAK-43-Pfad gegen apps/api). 0=skip,
# 1=zusätzlich zum MediaMTX-Pfad probe gegen /api/srt/health/{stream_id}.
SMOKE_INCLUDE_MTRACE_API="${SMOKE_INCLUDE_MTRACE_API:-0}"
MTRACE_API_URL="${MTRACE_API_URL:-http://localhost:8080}"
MTRACE_API_TOKEN="${MTRACE_API_TOKEN:-demo-token}"
MTRACE_API_STREAM_ID="${MTRACE_API_STREAM_ID:-$EXPECTED_PATH}"

for dep in curl docker python3; do
  if ! command -v "$dep" >/dev/null 2>&1; then
    echo "[smoke-srt-health] missing dependency: $dep" >&2
    exit 2
  fi
done

cleanup() {
  if [ "$SMOKE_SRT_AUTOSTART" = "1" ]; then
    echo "[smoke-srt-health] cleanup: docker compose -p $PROJECT down"
    docker compose -p "$PROJECT" -f "$COMPOSE_FILE" down >/dev/null 2>&1 || true
  fi
}

if [ "$SMOKE_SRT_AUTOSTART" = "1" ]; then
  trap cleanup EXIT
  echo "[smoke-srt-health] starting compose project $PROJECT"
  docker compose -p "$PROJECT" -f "$COMPOSE_FILE" up -d --build >/dev/null
fi

# 1) HLS-Manifest erreichbar (Baseline aus smoke-srt)
status=""
for _ in $(seq 1 "$WAIT_SECONDS"); do
  status="$(curl -sS -L -o /dev/null -w '%{http_code}' "$HLS_URL" 2>/dev/null || true)"
  if [ "$status" = "200" ]; then
    break
  fi
  sleep 1
done
if [ "$status" != "200" ]; then
  echo "[smoke-srt-health] HLS manifest unreachable at $HLS_URL (last status: ${status:-none})" >&2
  echo "[smoke-srt-health] diagnose:" >&2
  echo "  docker compose -p $PROJECT logs srt-publisher | tail -20" >&2
  echo "  docker compose -p $PROJECT logs mediamtx | tail -20" >&2
  exit 1
fi
echo "[smoke-srt-health] hls-status OK ($status @ $HLS_URL)"

# 2) MediaMTX-API erreichbar — Status und Body in einem curl-Call
#    erfassen, damit kein Race zwischen zwei Requests entsteht.
api_status=""
api_body=""
api_tmp="$(mktemp)"
trap 'rm -f "$api_tmp"; cleanup' EXIT
for _ in $(seq 1 "$WAIT_SECONDS"); do
  : >"$api_tmp"
  api_status="$(curl -sS -o "$api_tmp" -w '%{http_code}' "$API_URL" 2>/dev/null || true)"
  if [ "$api_status" = "200" ]; then
    api_body="$(cat "$api_tmp")"
    if [ -n "$api_body" ]; then
      break
    fi
  fi
  sleep 1
done
if [ "$api_status" != "200" ] || [ -z "$api_body" ]; then
  echo "[smoke-srt-health] MediaMTX-API unreachable or empty body (status: ${api_status:-none}, body length: ${#api_body})" >&2
  echo "[smoke-srt-health] URL: $API_URL" >&2
  echo "[smoke-srt-health] hint: examples/srt/mediamtx.yml braucht authInternalUsers mit action=api." >&2
  echo "[smoke-srt-health] diagnose:" >&2
  echo "  docker compose -p $PROJECT logs mediamtx | tail -30" >&2
  exit 1
fi
echo "[smoke-srt-health] api-status OK ($api_status @ $API_URL, body=${#api_body}b)"

# 3+4) Items-Validierung: Pfad + State + vier RAK-43-Pflichtwerte.
# Hinweis: `python3 -` mit Heredoc belegt stdin mit dem Script —
# deshalb wird der API-Body über die schon vorhandene Tempdatei
# `$api_tmp` als argv[1] übergeben.
EXPECTED_PATH="$EXPECTED_PATH" EXPECTED_STATE="$EXPECTED_STATE" python3 - "$api_tmp" <<'PYEOF'
import json, os, sys

expected_path = os.environ.get("EXPECTED_PATH", "srt-test")
expected_state = os.environ.get("EXPECTED_STATE", "publish")

api_body_path = sys.argv[1]
with open(api_body_path, "r", encoding="utf-8") as fh:
    raw = fh.read()
try:
    data = json.loads(raw)
except Exception as e:
    print(f"[smoke-srt-health] api response is not valid JSON: {e}", file=sys.stderr)
    print(f"  raw: {raw[:200]}", file=sys.stderr)
    sys.exit(1)

items = data.get("items") or []
if not items:
    print("[smoke-srt-health] /v3/srtconns/list returned no items[] — Publisher noch nicht verbunden?", file=sys.stderr)
    sys.exit(1)

target = next(
    (it for it in items if it.get("path") == expected_path and it.get("state") == expected_state),
    None,
)
if target is None:
    seen = ", ".join(f"{it.get('path')}/{it.get('state')}" for it in items)
    print(f"[smoke-srt-health] no item with path={expected_path} state={expected_state} (got: {seen})", file=sys.stderr)
    sys.exit(1)

# RAK-43-Pflichtwerte. msRTT/Loss/Retrans dürfen 0 sein (gesundes
# Lab); mbpsLinkCapacity muss > 0 sein, sonst keine Bandbreiten-
# Schätzung verfügbar.
required = [
    ("msRTT",                  lambda v: isinstance(v, (int, float)) and v >= 0),
    ("packetsReceivedLoss",    lambda v: isinstance(v, int) and v >= 0),
    ("packetsReceivedRetrans", lambda v: isinstance(v, int) and v >= 0),
    ("mbpsLinkCapacity",       lambda v: isinstance(v, (int, float)) and v > 0),
]
for field, check in required:
    if field not in target:
        print(f"[smoke-srt-health] missing field: {field}", file=sys.stderr)
        sys.exit(1)
    if not check(target[field]):
        print(f"[smoke-srt-health] field {field}={target[field]!r} failed validation", file=sys.stderr)
        sys.exit(1)

print(
    "[smoke-srt-health] item OK "
    f"(path={target['path']} state={target['state']} "
    f"msRTT={target['msRTT']} mbpsLinkCapacity={target['mbpsLinkCapacity']} "
    f"packetsReceivedLoss={target['packetsReceivedLoss']} "
    f"packetsReceivedRetrans={target['packetsReceivedRetrans']})"
)
PYEOF

# 5+6) Opt-in: m-trace-API-Read-Pfad gegen /api/srt/health/{stream_id}.
# Setzt voraus, dass apps/api separat läuft (z. B. `make dev`) und
# der Collector aktiv ist (`MTRACE_SRT_SOURCE_URL` gesetzt). Wir
# warten bis $WAIT_SECONDS auf den ersten persistierten Sample,
# weil der Collector-Polling-Default 5s ist und Backoff bei Source-
# Fehlern ihn zusätzlich verzögern kann.
if [ "$SMOKE_INCLUDE_MTRACE_API" = "1" ]; then
  mtrace_url="${MTRACE_API_URL%/}/api/srt/health/${MTRACE_API_STREAM_ID}"
  mtrace_status=""
  mtrace_body=""
  mtrace_tmp="$(mktemp)"
  trap 'rm -f "$api_tmp" "$mtrace_tmp"; cleanup' EXIT
  for _ in $(seq 1 "$WAIT_SECONDS"); do
    : >"$mtrace_tmp"
    mtrace_status="$(curl -sS -o "$mtrace_tmp" -w '%{http_code}' \
      -H "X-MTrace-Token: $MTRACE_API_TOKEN" \
      "$mtrace_url" 2>/dev/null || true)"
    if [ "$mtrace_status" = "200" ]; then
      mtrace_body="$(cat "$mtrace_tmp")"
      if [ -n "$mtrace_body" ]; then
        break
      fi
    fi
    sleep 1
  done
  if [ "$mtrace_status" != "200" ] || [ -z "$mtrace_body" ]; then
    echo "[smoke-srt-health] m-trace-API unreachable or empty body (status: ${mtrace_status:-none}, body length: ${#mtrace_body})" >&2
    echo "[smoke-srt-health] URL: $mtrace_url" >&2
    echo "[smoke-srt-health] hint: m-trace-API muss laufen (make dev) und der Collector aktiv sein (MTRACE_SRT_SOURCE_URL gesetzt)." >&2
    exit 1
  fi
  echo "[smoke-srt-health] mtrace-api-status OK ($mtrace_status @ $mtrace_url, body=${#mtrace_body}b)"

  # Wire-Format-Pflichtwerte aus spec/backend-api-contract.md §7a.2.
  python3 - "$mtrace_tmp" <<'PYEOF'
import json, sys

with open(sys.argv[1], "r", encoding="utf-8") as fh:
    data = json.load(fh)

items = data.get("items") or []
if not items:
    print("[smoke-srt-health] m-trace api: items[] leer — Collector noch ohne Samples?", file=sys.stderr)
    sys.exit(1)

latest = items[0]
metrics = latest.get("metrics") or {}
required = [
    ("rtt_ms",                 lambda v: isinstance(v, (int, float)) and v >= 0),
    ("packet_loss_total",      lambda v: isinstance(v, int) and v >= 0),
    ("retransmissions_total",  lambda v: isinstance(v, int) and v >= 0),
    ("available_bandwidth_bps", lambda v: isinstance(v, int) and v > 0),
]
for field, check in required:
    if field not in metrics:
        print(f"[smoke-srt-health] m-trace api: missing metrics.{field}", file=sys.stderr)
        sys.exit(1)
    if not check(metrics[field]):
        print(f"[smoke-srt-health] m-trace api: metrics.{field}={metrics[field]!r} failed validation", file=sys.stderr)
        sys.exit(1)

print(
    "[smoke-srt-health] mtrace-api item OK "
    f"(stream_id={latest.get('stream_id')} health_state={latest.get('health_state')} "
    f"rtt_ms={metrics['rtt_ms']} available_bandwidth_bps={metrics['available_bandwidth_bps']})"
)
PYEOF
else
  echo "[smoke-srt-health] mtrace-api skipped (set SMOKE_INCLUDE_MTRACE_API=1 to enable)"
fi

echo "[smoke-srt-health] all checks passed"
