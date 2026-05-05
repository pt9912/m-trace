#!/usr/bin/env bash
set -euo pipefail

# plan-0.6.0 §3 Tranche 2 — SRT-Health-Smoke (Vorbereitung RAK-41/42).
#
# Erweitert den smoke-srt-Pfad um eine API-Probe gegen
# MediaMTX `/v3/srtconns/list`. Verifiziert:
#   1. HLS-Manifest erreichbar (Publish-Baseline aus smoke-srt).
#   2. MediaMTX Control-API antwortet `200`.
#   3. `items[]` enthält mindestens eine Verbindung mit
#      `path=srt-test` und `state=publish`.
#   4. Vier RAK-43-Pflichtwerte sind numerisch gesetzt:
#      `msRTT`, `packetsReceivedLoss`, `packetsReceivedRetrans`,
#      `mbpsLinkCapacity` — letzteres muss `> 0` sein, der Rest
#      darf 0 sein (gesundes Lab).
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

echo "[smoke-srt-health] all checks passed"
