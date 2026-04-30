#!/usr/bin/env bash
set -euo pipefail

API_URL="${API_URL:-http://localhost:8080}"

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

wait_for_status() {
  url="$1"
  expected="$2"
  label="$3"

  for _ in $(seq 1 60); do
    status="$(curl -sS -o "$tmpdir/${label}.body" -w '%{http_code}' "$url" || true)"
    if [ "$status" = "$expected" ]; then
      echo "$label: $status"
      return 0
    fi
    sleep 1
  done

  echo "$label: expected $expected, got ${status:-none}" >&2
  cat "$tmpdir/${label}.body" >&2 || true
  return 1
}

wait_for_status "${API_URL}/api/health" "200" "api-health"
scripts/seed-rak9.sh --base-url "$API_URL" --sessions 1 --events-per-session 1

for _ in $(seq 1 20); do
  docker compose logs api --tail=240 > "$tmpdir/api.logs"
  if grep -Fq '"Name":"http.handler POST /api/playback-events"' "$tmpdir/api.logs"; then
    if grep -Fq '"Key":"http.status_code"' "$tmpdir/api.logs" && grep -Fq '"Value":202' "$tmpdir/api.logs"; then
      echo "rak10-console-span: found"
      exit 0
    fi
  fi
  sleep 1
done

echo "rak10-console-span: missing request span in api logs" >&2
tail -n 120 "$tmpdir/api.logs" >&2 || true
exit 1
