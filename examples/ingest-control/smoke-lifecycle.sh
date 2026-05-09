#!/usr/bin/env bash
# smoke-lifecycle.sh — Reproduzierbarer Lab-Smoke für die
# Lifecycle-Hooks aus `0.11.0` Tranche 4 (RAK-69).
#
# Verifiziert gegen eine **lokal** laufende `apps/api`:
#   1. POST /api/ingest/streams         (Stream anlegen)
#   2. POST /api/ingest/hooks/stream-started   (202 + accepted:true)
#   3. POST /api/ingest/hooks/stream-ended     (202 + accepted:true)
#
# Voraussetzungen:
#   - apps/api ist erreichbar (Default: http://localhost:8080).
#     Empfohlen: `make dev` — der Compose-Stack hängt das `0.11.0`
#     Ingest-Control-Use-Case-Modul automatisch ein, sofern die
#     SQLite-Volumes vorbereitet sind.
#   - MTRACE_API_TOKEN ist gesetzt; default `demo-token` aus dem
#     Static-Resolver.
#
# Der Smoke ist **opt-in** (nicht Teil von `make gates`). Schlägt
# beim ersten unerwarteten HTTP-Status fehl und gibt den Body aus,
# damit Operatoren den Fehler im Klartext sehen.

set -euo pipefail

API_URL="${MTRACE_API_URL:-http://localhost:8080}"
API_TOKEN="${MTRACE_API_TOKEN:-demo-token}"
ENDPOINT_ID="${MTRACE_INGEST_ENDPOINT:-ep-srt}"
TARGET_ID="${MTRACE_INGEST_TARGET:-tgt-mediamtx}"
DISPLAY_NAME="${MTRACE_INGEST_DISPLAY_NAME:-smoke-lifecycle}"

if ! command -v jq >/dev/null 2>&1; then
  echo "smoke-lifecycle: 'jq' fehlt — bitte installieren." >&2
  exit 2
fi

curl_json() {
  local method="$1" path="$2" body="${3:-}"
  local args=(-sS -o /tmp/smoke-lifecycle-body.$$ -w "%{http_code}"
    -X "${method}"
    -H "X-MTrace-Token: ${API_TOKEN}"
    -H "Content-Type: application/json"
    "${API_URL}${path}")
  if [[ -n "${body}" ]]; then
    args+=(--data "${body}")
  fi
  curl "${args[@]}"
}

assert_status() {
  local actual="$1" want="$2" step="$3"
  if [[ "${actual}" != "${want}" ]]; then
    echo "smoke-lifecycle: ${step} — want ${want}, got ${actual}" >&2
    cat /tmp/smoke-lifecycle-body.$$ >&2 || true
    exit 1
  fi
}

echo "▶ create stream …"
status="$(curl_json POST /api/ingest/streams "$(cat <<EOF
{"display_name":"${DISPLAY_NAME}","protocol":"srt","endpoint_id":"${ENDPOINT_ID}","target_id":"${TARGET_ID}"}
EOF
)")"
assert_status "${status}" "201" "create-stream"
stream_id="$(jq -r '.id' /tmp/smoke-lifecycle-body.$$)"
if [[ -z "${stream_id}" || "${stream_id}" == "null" ]]; then
  echo "smoke-lifecycle: create returned no id" >&2
  cat /tmp/smoke-lifecycle-body.$$ >&2
  exit 1
fi
echo "  stream_id=${stream_id}"

echo "▶ stream-started hook …"
status="$(curl_json POST /api/ingest/hooks/stream-started "$(cat <<EOF
{"stream_id":"${stream_id}","observed_at":"$(date -u +%Y-%m-%dT%H:%M:%SZ)","source":"local-smoke","connection_id":"smoke-conn-1"}
EOF
)")"
assert_status "${status}" "202" "stream-started"
accepted="$(jq -r '.accepted' /tmp/smoke-lifecycle-body.$$)"
event_started="$(jq -r '.event_id' /tmp/smoke-lifecycle-body.$$)"
type_started="$(jq -r '.type' /tmp/smoke-lifecycle-body.$$)"
if [[ "${accepted}" != "true" || "${type_started}" != "stream_started" ]]; then
  echo "smoke-lifecycle: stream-started body unerwartet" >&2
  cat /tmp/smoke-lifecycle-body.$$ >&2
  exit 1
fi
echo "  event_id=${event_started}"

echo "▶ stream-ended hook …"
status="$(curl_json POST /api/ingest/hooks/stream-ended "$(cat <<EOF
{"stream_id":"${stream_id}","observed_at":"$(date -u +%Y-%m-%dT%H:%M:%SZ)","source":"local-smoke","connection_id":"smoke-conn-1","reason":"smoke_complete"}
EOF
)")"
assert_status "${status}" "202" "stream-ended"
event_ended="$(jq -r '.event_id' /tmp/smoke-lifecycle-body.$$)"
type_ended="$(jq -r '.type' /tmp/smoke-lifecycle-body.$$)"
if [[ "${type_ended}" != "stream_ended" || "${event_ended}" == "${event_started}" ]]; then
  echo "smoke-lifecycle: stream-ended body unerwartet (Event-IDs müssen unterschiedlich sein)" >&2
  cat /tmp/smoke-lifecycle-body.$$ >&2
  exit 1
fi
echo "  event_id=${event_ended}"

rm -f /tmp/smoke-lifecycle-body.$$
echo "✔ smoke-lifecycle ok (stream=${stream_id}, started=${event_started}, ended=${event_ended})"
