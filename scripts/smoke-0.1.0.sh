#!/usr/bin/env bash
set -euo pipefail

API_URL="${API_URL:-http://localhost:8080}"
HLS_URL="${HLS_URL:-http://localhost:8888/teststream/index.m3u8}"
TOKEN="${MTRACE_DEMO_TOKEN:-demo-token}"
SESSION_ID="${MTRACE_SMOKE_SESSION_ID:-smoke-test-1}"

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

wait_for_status() {
  url="$1"
  expected="$2"
  label="$3"

  for _ in $(seq 1 60); do
    status="$(curl -sSL -o "$tmpdir/$label.body" -w '%{http_code}' "$url" || true)"
    if [ "$status" = "$expected" ]; then
      echo "$label: $status"
      return 0
    fi
    sleep 1
  done

  echo "$label: expected $expected, got ${status:-none}" >&2
  cat "$tmpdir/$label.body" >&2 || true
  return 1
}

wait_for_status "$API_URL/api/health" "200" "health"

post_status="$(
  curl -sS -o "$tmpdir/post.body" -w '%{http_code}' \
    -X POST "$API_URL/api/playback-events" \
    -H 'Content-Type: application/json' \
    -H "X-MTrace-Token: $TOKEN" \
    --data-binary @- <<JSON
{
  "schema_version": "1.0",
  "events": [{
    "event_name": "rebuffer_started",
    "project_id": "demo",
    "session_id": "$SESSION_ID",
    "client_timestamp": "2026-04-29T10:00:00.000Z",
    "sequence_number": 1,
    "sdk": {"name": "@m-trace/player-sdk", "version": "0.1.0"}
  }]
}
JSON
)"

if [ "$post_status" != "202" ]; then
  echo "post-events: expected 202, got $post_status" >&2
  cat "$tmpdir/post.body" >&2
  exit 1
fi
echo "post-events: $post_status"

sessions_status="$(curl -sS -o "$tmpdir/sessions.body" -w '%{http_code}' \
  -H "X-MTrace-Token: $TOKEN" \
  "$API_URL/api/stream-sessions")"
if [ "$sessions_status" != "200" ]; then
  echo "stream-sessions: expected 200, got $sessions_status" >&2
  cat "$tmpdir/sessions.body" >&2
  exit 1
fi
if ! grep -Fq "$SESSION_ID" "$tmpdir/sessions.body"; then
  echo "stream-sessions: session $SESSION_ID not found" >&2
  cat "$tmpdir/sessions.body" >&2
  exit 1
fi
echo "stream-sessions: $sessions_status"

wait_for_status "$HLS_URL" "200" "hls-manifest"
