#!/usr/bin/env bash
set -euo pipefail

API_URL="${API_URL:-http://localhost:8080}"
PROMETHEUS_URL="${PROMETHEUS_URL:-http://localhost:9090}"
GRAFANA_URL="${GRAFANA_URL:-http://localhost:3000}"
OTEL_HEALTH_URL="${OTEL_HEALTH_URL:-http://localhost:13133}"
SCRAPE_WAIT_SECONDS="${SCRAPE_WAIT_SECONDS:-16}"

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

prom_query() {
  query="$1"
  curl -sS --get "${PROMETHEUS_URL}/api/v1/query" --data-urlencode "query=${query}"
}

wait_for_status "${API_URL}/api/health" "200" "api-health"
wait_for_status "${PROMETHEUS_URL}/-/ready" "200" "prometheus-ready"
wait_for_status "${GRAFANA_URL}/api/health" "200" "grafana-health"
wait_for_status "${OTEL_HEALTH_URL}/" "200" "otel-health"

scripts/seed-rak9.sh --base-url "$API_URL"
sleep "$SCRAPE_WAIT_SECONDS"

playback_value="$(prom_query 'mtrace_playback_events_total' | node -e 'const p=JSON.parse(require("fs").readFileSync(0,"utf8")); const r=p.data.result[0]; process.stdout.write(r ? r.value[1] : "0")')"
if [ "${playback_value%.*}" -le 0 ]; then
  echo "prometheus-playback-events: expected >0, got ${playback_value}" >&2
  exit 1
fi
echo "prometheus-playback-events: ${playback_value}"

series_json="$(curl -sS --get "${PROMETHEUS_URL}/api/v1/series" --data-urlencode 'match[]={__name__=~"mtrace_.+"}')"
series_count="$(printf '%s' "$series_json" | node -e 'const p=JSON.parse(require("fs").readFileSync(0,"utf8")); process.stdout.write(String(p.data.length))')"
if [ "$series_count" -le 0 ]; then
  echo "prometheus-series: expected non-empty mtrace_* series" >&2
  printf '%s\n' "$series_json" >&2
  exit 1
fi
printf '%s' "$series_json" | node -e '
const p=JSON.parse(require("fs").readFileSync(0,"utf8"));
const forbidden=["session_id","user_agent","segment_url","client_ip"];
const bad=p.data.filter((series) => forbidden.some((label) => Object.prototype.hasOwnProperty.call(series, label)));
if (bad.length) {
  console.error(JSON.stringify(bad, null, 2));
  process.exit(1);
}
'
echo "prometheus-series: ${series_count} mtrace series, forbidden labels absent"

cardinality="$(prom_query 'count(count by (instance, job, __name__) ({__name__=~"mtrace_.+"}))' | node -e 'const p=JSON.parse(require("fs").readFileSync(0,"utf8")); const r=p.data.result[0]; process.stdout.write(r ? r.value[1] : "0")')"
if [ "${cardinality%.*}" -ge 50 ]; then
  echo "prometheus-cardinality: expected <50, got ${cardinality}" >&2
  exit 1
fi
echo "prometheus-cardinality: ${cardinality}"
