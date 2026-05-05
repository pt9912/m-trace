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

prom_first_value() {
  node -e 'const p=JSON.parse(require("fs").readFileSync(0,"utf8")); const r=p.data.result[0]; process.stdout.write(r ? r.value[1] : "0")'
}

wait_for_status "${API_URL}/api/health" "200" "api-health"
wait_for_status "${PROMETHEUS_URL}/-/ready" "200" "prometheus-ready"
wait_for_status "${GRAFANA_URL}/api/health" "200" "grafana-health"
wait_for_status "${OTEL_HEALTH_URL}/" "200" "otel-health"

scripts/seed-rak9.sh --base-url "$API_URL"
sleep "$SCRAPE_WAIT_SECONDS"

playback_value="$(prom_query 'mtrace_playback_events_total' | prom_first_value)"
if [ "${playback_value%.*}" -le 0 ]; then
  echo "prometheus-playback-events: expected >0, got ${playback_value}" >&2
  exit 1
fi
echo "prometheus-playback-events: ${playback_value}"

required_metrics=(
  mtrace_playback_events_total
  mtrace_playback_errors_total
  mtrace_active_sessions
  mtrace_rebuffer_events_total
  mtrace_startup_time_ms
  mtrace_api_requests_total
  mtrace_dropped_events_total
  mtrace_rate_limited_events_total
  mtrace_invalid_events_total
)

for metric in "${required_metrics[@]}"; do
  metric_cardinality="$(prom_query "count(count by (__name__) (${metric}))" | prom_first_value)"
  if [ "${metric_cardinality%.*}" -gt 1 ]; then
    echo "prometheus-required-metric-cardinality: ${metric} expected <=1, got ${metric_cardinality}" >&2
    exit 1
  fi
  if [ "${metric_cardinality%.*}" -lt 1 ]; then
    echo "prometheus-required-metric-cardinality: ${metric} expected present, got ${metric_cardinality}" >&2
    exit 1
  fi
  echo "prometheus-required-metric-cardinality: ${metric}=${metric_cardinality}"
done

# plan-0.4.0 §7.3: pro §7-Pflichtcounter eine strikte Labelset-
# Assertion. Der Counter darf KEINE fachlichen Labels tragen — erlaubt
# sind nur Prometheus-Target-Metadaten (`__name__`, `instance`, `job`).
# Jeder zusätzliche Label-Key ist ein Cardinality-Verstoß und
# release-blockierend (API-Kontrakt §7).
mandatory_counters=(
  mtrace_playback_events_total
  mtrace_invalid_events_total
  mtrace_rate_limited_events_total
  mtrace_dropped_events_total
)

for metric in "${mandatory_counters[@]}"; do
  series_count="$(prom_query "count(${metric})" | prom_first_value)"
  if [ "${series_count%.*}" != "1" ]; then
    echo "prometheus-mandatory-counter-series: ${metric} expected exactly 1 series, got ${series_count}" >&2
    exit 1
  fi
  series_labels="$(curl -sS --get "${PROMETHEUS_URL}/api/v1/series" --data-urlencode "match[]=${metric}")"
  printf '%s' "$series_labels" | node -e '
const p=JSON.parse(require("fs").readFileSync(0,"utf8"));
const allowed=new Set(["__name__","instance","job"]);
const extras=[];
for (const series of p.data) {
  for (const key of Object.keys(series)) {
    if (!allowed.has(key)) extras.push({metric: series.__name__, label: key, value: series[key]});
  }
}
if (extras.length) {
  console.error("mandatory counter has extra labels: " + JSON.stringify(extras, null, 2));
  process.exit(1);
}
'
  echo "prometheus-mandatory-counter-labelset: ${metric} label-free OK"
done

series_json="$(curl -sS --get "${PROMETHEUS_URL}/api/v1/series" --data-urlencode 'match[]={__name__=~"mtrace_.+"}')"
series_count="$(printf '%s' "$series_json" | node -e 'const p=JSON.parse(require("fs").readFileSync(0,"utf8")); process.stdout.write(String(p.data.length))')"
if [ "$series_count" -le 0 ]; then
  echo "prometheus-series: expected non-empty mtrace_* series" >&2
  printf '%s\n' "$series_json" >&2
  exit 1
fi
# plan-0.4.0 §7.3: Forbidden-Labels über alle `mtrace_*`-Serien hinweg.
# Liste deckt §7-Vertrag (project_id/session_id/Token/etc.) plus
# Telemetry-Model §3.1 ab. Andere `mtrace_*`-Metriken dürfen bounded
# Aggregat-Labels (`outcome`, `code`, `event_type`) tragen — das
# Filter ist gezielt forbidden-by-name, nicht allowlist-by-name.
printf '%s' "$series_json" | node -e '
const p=JSON.parse(require("fs").readFileSync(0,"utf8"));
const forbidden=[
  "session_id","user_agent","segment_url","client_ip",
  "project_id","trace_id","span_id","correlation_id",
  "viewer_id","request_id","token","authorization",
  "url","uri","secret",
  // plan-0.4.0 §8.2 / spec/telemetry-model.md §3.1: batch_size is
  // forbidden because mtrace.api.batches.received runs in use case
  // Step 0 (vor MaxBatchSize-Validierung) — a rejected batch with
  // events.length=250 would otherwise emit batch_size="250".
  "batch_size"
];
const forbiddenSuffixes=["_url","_uri","_token","_secret"];
const forbiddenLabels=(series) =>
  Object.keys(series).filter((label) =>
    forbidden.includes(label) ||
    forbiddenSuffixes.some((suffix) => label.endsWith(suffix))
  );
const policyProbe=[
  {__name__:"mtrace_test_total",manifest_url:"x"},
  {__name__:"mtrace_test_total",url:"x"},
  {__name__:"mtrace_test_total",uri:"x"},
  {__name__:"mtrace_test_total",secret:"x"},
  {__name__:"mtrace_test_total",batch_size:"7"}
];
const missed=policyProbe.filter((series) => forbiddenLabels(series).length === 0);
if (missed.length) {
  console.error("forbidden label policy self-test failed: " + JSON.stringify(missed, null, 2));
  process.exit(1);
}
const bad=p.data.filter((series) => forbiddenLabels(series).length > 0);
if (bad.length) {
  console.error("forbidden labels found: " + JSON.stringify(bad, null, 2));
  process.exit(1);
}
'
echo "prometheus-series: ${series_count} mtrace series, forbidden labels absent"

cardinality="$(prom_query 'count(count by (instance, job, __name__) ({__name__=~"mtrace_.+"}))' | prom_first_value)"
if [ "${cardinality%.*}" -ge 50 ]; then
  echo "prometheus-cardinality: expected <50, got ${cardinality}" >&2
  exit 1
fi
echo "prometheus-cardinality: ${cardinality}"

# plan-0.6.0 §4 Sub-3.7: SRT-Health-Allowlist prüfen.
#
# Wenn `mtrace_srt_health_*`-Serien existieren (Collector ist aktiv),
# müssen sie sich auf die in spec/telemetry-model.md §3.2 / §7.7
# freigegebenen bounded Labels (`health_state`, `source_status`,
# `source_error_code`) plus Target-Metadaten (`__name__`, `instance`,
# `job`) beschränken. Andere Labels — insbesondere Per-Verbindung-
# Identifier wie `stream_id`, `connection_id`, MediaMTX-`id`/`path`/
# `remoteAddr`/`state` — sind verboten (Cardinality-Vertrag §7.7).
#
# Wenn keine Serien existieren (Default-Lab ohne SRT-Collector),
# überspringt der Check stillschweigend — der Collector ist opt-in
# über `MTRACE_SRT_SOURCE_URL`.
srt_series_json="$(curl -sS --get "${PROMETHEUS_URL}/api/v1/series" --data-urlencode 'match[]={__name__=~"mtrace_srt_health_.+"}')"
srt_series_count="$(printf '%s' "$srt_series_json" | node -e 'const p=JSON.parse(require("fs").readFileSync(0,"utf8")); process.stdout.write(String(p.data.length))')"
if [ "$srt_series_count" -gt 0 ]; then
  printf '%s' "$srt_series_json" | node -e '
const p=JSON.parse(require("fs").readFileSync(0,"utf8"));
const allowedByMetric={
  "mtrace_srt_health_samples_total":           new Set(["__name__","instance","job","health_state"]),
  "mtrace_srt_health_collector_runs_total":    new Set(["__name__","instance","job","source_status"]),
  "mtrace_srt_health_collector_errors_total":  new Set(["__name__","instance","job","source_error_code"]),
};
const violations=[];
for (const series of p.data) {
  const allowed = allowedByMetric[series.__name__];
  if (!allowed) {
    violations.push({metric: series.__name__, reason: "unexpected mtrace_srt_health_* metric"});
    continue;
  }
  for (const key of Object.keys(series)) {
    if (!allowed.has(key)) {
      violations.push({metric: series.__name__, label: key, value: series[key]});
    }
  }
}
if (violations.length) {
  console.error("srt-health-allowlist violations: " + JSON.stringify(violations, null, 2));
  process.exit(1);
}
'
  echo "prometheus-srt-health-allowlist: ${srt_series_count} mtrace_srt_health_* series, allowlist OK"
else
  echo "prometheus-srt-health-allowlist: skipped (collector not active)"
fi

# plan-0.6.0 §0.1 + spec/telemetry-model.md §7.7:
# Source-Rohmetriken aus MediaMTX/SRT werden nicht vom Projekt-
# Prometheus gescraped. Wir prüfen das, indem wir die Active-Targets-
# Liste durchgehen und nach mediamtx-/srt-typischen Job- oder
# Address-Mustern suchen. Treffer bedeutet: ein Operator hat die
# Forbidden-Klausel ohne separaten Folge-Scope gebrochen.
targets_json="$(curl -sS "${PROMETHEUS_URL}/api/v1/targets" || true)"
if [ -z "$targets_json" ]; then
  echo "prometheus-source-targets: targets endpoint unreachable" >&2
  exit 1
fi
printf '%s' "$targets_json" | node -e '
const p=JSON.parse(require("fs").readFileSync(0,"utf8"));
const active = (p.data && p.data.activeTargets) || [];
const forbidden=[];
for (const t of active) {
  const job   = (t.labels && t.labels.job) || "";
  const inst  = (t.labels && t.labels.instance) || "";
  const url   = t.scrapeUrl || "";
  const haystack = (job + " " + inst + " " + url).toLowerCase();
  if (haystack.includes("mediamtx") || haystack.match(/(^|[^a-z])srt([^a-z]|$)/)) {
    forbidden.push({job, instance: inst, scrapeUrl: url});
  }
}
if (forbidden.length) {
  console.error("source raw metrics scraped by project Prometheus: " + JSON.stringify(forbidden, null, 2));
  process.exit(1);
}
'
echo "prometheus-source-targets: no MediaMTX/SRT raw scrape jobs detected"
