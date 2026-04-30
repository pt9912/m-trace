#!/usr/bin/env bash
set -euo pipefail

COMPOSE="${COMPOSE:-docker compose}"
API_URL="${API_URL:-http://localhost:8080}"

tmpdir="$(mktemp -d)"
trap 'status=$?; $COMPOSE --profile test down >/dev/null 2>&1 || true; rm -rf "$tmpdir"; exit $status' EXIT

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

$COMPOSE --profile test up -d --build api mediamtx stream-generator dashboard-e2e

wait_for_status "$API_URL/api/health" "200" "health"

$COMPOSE --profile test run --rm browser-e2e "$@"
