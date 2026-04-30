#!/usr/bin/env bash
set -euo pipefail

BASE_URL="http://localhost:8080"
PROJECT_ID="demo"
TOKEN="demo-token"
ORIGIN=""
SESSIONS=5
EVENTS_PER_SESSION=10
SKIP_AUTH=false

usage() {
  cat <<'USAGE'
Usage: scripts/seed-rak9.sh [options]

Options:
  --base-url URL             API base URL (default: http://localhost:8080)
  --project-id ID            Project ID (default: demo)
  --token TOKEN              X-MTrace-Token value (default: demo-token)
  --origin ORIGIN            Optional Origin header
  --sessions N               Number of sessions (default: 5)
  --events-per-session N     Events per session (default: 10)
  --skip-auth                Do not send X-MTrace-Token
  -h, --help                 Show this help
USAGE
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --base-url)
      BASE_URL="$2"
      shift 2
      ;;
    --project-id)
      PROJECT_ID="$2"
      shift 2
      ;;
    --token)
      TOKEN="$2"
      shift 2
      ;;
    --origin)
      ORIGIN="$2"
      shift 2
      ;;
    --sessions)
      SESSIONS="$2"
      shift 2
      ;;
    --events-per-session)
      EVENTS_PER_SESSION="$2"
      shift 2
      ;;
    --skip-auth)
      SKIP_AUTH=true
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "unknown option: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

case "$SESSIONS" in
  ''|*[!0-9]*)
    echo "--sessions must be a positive integer" >&2
    exit 2
    ;;
esac
case "$EVENTS_PER_SESSION" in
  ''|*[!0-9]*)
    echo "--events-per-session must be a positive integer" >&2
    exit 2
    ;;
esac
if [ "$SESSIONS" -lt 1 ] || [ "$EVENTS_PER_SESSION" -lt 1 ]; then
  echo "--sessions and --events-per-session must be >= 1" >&2
  exit 2
fi

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

headers=(-H "Content-Type: application/json")
if [ "$SKIP_AUTH" = false ]; then
  headers+=(-H "X-MTrace-Token: $TOKEN")
fi
if [ -n "$ORIGIN" ]; then
  headers+=(-H "Origin: $ORIGIN")
fi

event_name_for() {
  case "$1" in
    1) echo "playback_started" ;;
    2) echo "startup_time_measured" ;;
    3) echo "segment_loaded" ;;
    4) echo "rebuffer_started" ;;
    5) echo "rebuffer_ended" ;;
    6) echo "playback_error" ;;
    *) echo "segment_loaded" ;;
  esac
}

meta_for() {
  event_name="$1"
  sequence="$2"
  case "$event_name" in
    startup_time_measured)
      printf ', "meta": {"duration_ms": %d}' "$((900 + sequence))"
      ;;
    rebuffer_ended)
      printf ', "meta": {"duration_ms": %d, "rebuffer_count": 1}' "$((120 + sequence))"
      ;;
    playback_error)
      printf ', "meta": {"error_code": "rak9_demo"}'
      ;;
    *)
      ;;
  esac
}

posted=0
for session_index in $(seq 1 "$SESSIONS"); do
  session_id="rak9-session-${session_index}"
  events=""
  for event_index in $(seq 1 "$EVENTS_PER_SESSION"); do
    sequence=$(( (session_index - 1) * EVENTS_PER_SESSION + event_index ))
    event_name="$(event_name_for "$event_index")"
    timestamp="$(printf '2026-04-30T10:%02d:%02d.000Z' "$((session_index - 1))" "$((event_index - 1))")"
    meta="$(meta_for "$event_name" "$sequence")"
    if [ -n "$events" ]; then
      events="${events},"
    fi
    events="${events}{\"event_name\":\"${event_name}\",\"project_id\":\"${PROJECT_ID}\",\"session_id\":\"${session_id}\",\"client_timestamp\":\"${timestamp}\",\"sequence_number\":${sequence},\"sdk\":{\"name\":\"@m-trace/player-sdk\",\"version\":\"0.1.1\"}${meta}}"
  done

  payload="{\"schema_version\":\"1.0\",\"events\":[${events}]}"
  status="$(
    curl -sS -o "$tmpdir/post-${session_index}.body" -w '%{http_code}' \
      -X POST "${BASE_URL}/api/playback-events" \
      "${headers[@]}" \
      --data-binary "$payload"
  )"
  if [ "$status" != "202" ]; then
    echo "seed-rak9: session ${session_id} expected 202, got ${status}" >&2
    cat "$tmpdir/post-${session_index}.body" >&2
    exit 1
  fi
  posted=$((posted + EVENTS_PER_SESSION))
done

echo "seed-rak9: posted ${posted} events across ${SESSIONS} sessions"
