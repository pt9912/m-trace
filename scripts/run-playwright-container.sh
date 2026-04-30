#!/usr/bin/env bash
set -euo pipefail

corepack enable
pnpm install --frozen-lockfile --config.engine-strict=false

wait_for_status() {
  url="$1"
  expected="$2"
  label="$3"

  for _ in $(seq 1 60); do
    status="$(curl -sSL -o /tmp/"$label".body -w '%{http_code}' "$url" || true)"
    if [ "$status" = "$expected" ]; then
      echo "$label: $status"
      return 0
    fi
    sleep 1
  done

  echo "$label: expected $expected, got ${status:-none}" >&2
  cat /tmp/"$label".body >&2 || true
  return 1
}

wait_for_status "${API_URL:-http://api:8080}/api/health" "200" "api-health"
wait_for_status "${DASHBOARD_URL:-http://dashboard-e2e:5173}/demo" "200" "dashboard-demo"

pnpm exec playwright test "$@"
