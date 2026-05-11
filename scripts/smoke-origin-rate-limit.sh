#!/usr/bin/env bash
set -euo pipefail

# plan-0.12.6 Tranche 6 (R-22) — Origin-/IP-Rate-Limiter Smoke.
#
# Pflichtpfad:
#   Setzt einen lokalen Capacity=2-Bucket pro Client-IP (Default-
#   Konstanten via niedrigerer ENV ueberschrieben), feuert drei
#   `POST /api/auth/session-tokens`-Aufrufe in Folge:
#     1. + 2. → erwartete 201.
#     3.      → erwartete 429 mit Body `{"error":"origin_rate_limited"}`.
#
# Voraussetzungen:
#   - m-trace-API laeuft auf $MTRACE_API_URL (Default
#     http://localhost:8080) mit `MTRACE_ORIGIN_RATE_LIMITER=memory`.
#   - Project-Token `demo-token` ist konfiguriert (Compose-Default
#     aus examples/local-dev).
#
# Opt-in (NICHT in `make gates`); braucht `curl`, `python3`.

MTRACE_API_URL="${MTRACE_API_URL:-http://localhost:8080}"
MTRACE_API_TOKEN="${MTRACE_API_TOKEN:-demo-token}"
TOKEN_ENDPOINT="${MTRACE_API_URL%/}/api/auth/session-tokens"

for dep in curl python3; do
  if ! command -v "$dep" >/dev/null 2>&1; then
    echo "[smoke-origin-rate-limit] missing dependency: $dep" >&2
    exit 2
  fi
done

call() {
  local i="$1"
  local tmp
  tmp="$(mktemp)"
  local status
  status="$(curl -sS -o "$tmp" -w '%{http_code}' \
    -H "X-MTrace-Token: $MTRACE_API_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"audience":"playback-events"}' \
    "$TOKEN_ENDPOINT" 2>/dev/null || true)"
  echo "[smoke-origin-rate-limit] call #$i status=$status body=$(cat "$tmp")"
  printf '%s' "$status:$tmp"
}

result1="$(call 1)"
status1="${result1%%:*}"
result2="$(call 2)"
status2="${result2%%:*}"
result3="$(call 3)"
status3="${result3%%:*}"
tmp3="${result3##*:}"

if [ "$status1" != "201" ] || [ "$status2" != "201" ]; then
  echo "[smoke-origin-rate-limit] first 2 calls expected 201, got $status1 + $status2" >&2
  echo "[smoke-origin-rate-limit] hint: ensure MTRACE_ORIGIN_RATE_LIMITER=memory and capacity ≥ 2 is set on the API process" >&2
  exit 1
fi

if [ "$status3" != "429" ]; then
  echo "[smoke-origin-rate-limit] 3rd call expected 429, got $status3" >&2
  echo "[smoke-origin-rate-limit] hint: ensure origin-limiter capacity is exactly 2 (override via build-time constants or use scripted limiter override)" >&2
  exit 1
fi

err="$(python3 -c 'import json,sys; d=json.load(open(sys.argv[1])); print(d.get("error",""))' "$tmp3")"
if [ "$err" != "origin_rate_limited" ]; then
  echo "[smoke-origin-rate-limit] 3rd call error body = $err, want origin_rate_limited" >&2
  exit 1
fi

echo "[smoke-origin-rate-limit] all checks passed (2× 201, then 429 origin_rate_limited)"
