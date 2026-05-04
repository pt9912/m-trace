#!/usr/bin/env bash
set -euo pipefail

# plan-0.4.0 §6.4 — Tempo-Smoke für drei Startzustände.
#
# Pflichten aus §6 DoD:
#   1. Core (Default-Compose ohne `OTEL_*`-Werte) löst KEINEN
#      OTLP-Verbindungsversuch aus.
#   2. observability-Profil (ohne Tempo) liefert OTLP an den Collector
#      und produziert keine Tempo-Exportfehler.
#   3. tempo-Profil aktiv: `correlation_id`-Roundtrip ist in Tempo
#      auffindbar (über das Span-Attribut `mtrace.session.correlation_id`,
#      siehe `spec/telemetry-model.md` §2.6).
#
# Aufruf:
#   SMOKE_STATE=core         scripts/smoke-tempo.sh    # Stack: `make dev`
#   SMOKE_STATE=observability scripts/smoke-tempo.sh   # Stack: `make dev-observability`
#   SMOKE_STATE=tempo        scripts/smoke-tempo.sh    # Stack: `make dev-tempo` (Default)

API_URL="${API_URL:-http://localhost:8080}"
TEMPO_URL="${TEMPO_URL:-http://localhost:3200}"
OTEL_HEALTH_URL="${OTEL_HEALTH_URL:-http://localhost:13133}"
SMOKE_STATE="${SMOKE_STATE:-tempo}"
TEMPO_INGEST_WAIT_SECONDS="${TEMPO_INGEST_WAIT_SECONDS:-20}"
TEMPO_SEARCH_LOOKBACK_SECONDS="${TEMPO_SEARCH_LOOKBACK_SECONDS:-300}"
TOKEN="${TOKEN:-demo-token}"
PROJECT_ID="${PROJECT_ID:-demo}"

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

wait_for_status() {
  local url="$1" expected="$2" label="$3"
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

post_event() {
  local session_id="$1"
  local body
  body=$(cat <<JSON
{
  "schema_version": "1.0",
  "events": [
    {
      "event_name": "manifest_loaded",
      "project_id": "${PROJECT_ID}",
      "session_id": "${session_id}",
      "client_timestamp": "$(date -u +%Y-%m-%dT%H:%M:%S.000Z)",
      "sdk": {"name": "smoke-tempo", "version": "0.4.0"}
    }
  ]
}
JSON
)
  curl -sS -o "$tmpdir/post.body" -w '%{http_code}' \
    -H "Content-Type: application/json" \
    -H "X-MTrace-Token: ${TOKEN}" \
    -X POST "${API_URL}/api/playback-events" \
    -d "$body"
}

read_correlation_id() {
  local session_id="$1"
  curl -sS \
    -H "X-MTrace-Token: ${TOKEN}" \
    "${API_URL}/api/stream-sessions/${session_id}" |
    node -e 'const p=JSON.parse(require("fs").readFileSync(0,"utf8")); process.stdout.write(p.session?.correlation_id ?? "")'
}

urlencode() {
  node -e 'process.stdout.write(encodeURIComponent(process.argv[1]))' "$1"
}

tempo_trace_count() {
  node -e 'try { const p=JSON.parse(require("fs").readFileSync(0,"utf8")); process.stdout.write(String((p.traces||[]).length)); } catch { process.stdout.write("0"); }'
}

case "$SMOKE_STATE" in
  core)
    # State 1: Core ohne OTLP-Werte.
    # api-Container hat in docker-compose.yml OTEL_*-Defaults auf "" —
    # daher kein OTLP-Versuch. Smoke prüft, dass die API auch ohne
    # observability-Stack lebt.
    wait_for_status "${API_URL}/api/health" "200" "api-health-core"
    echo "smoke-tempo[core]: API up, no OTLP collector — keine Tempo-Pipeline aktiv"
    ;;

  observability)
    # State 2: observability-Profil, kein Tempo.
    # Collector lädt config.yaml (Default-Pfad in COLLECTOR_CONFIG),
    # Traces gehen an `debug`-Exporter — KEIN Tempo-Verbindungsversuch.
    wait_for_status "${API_URL}/api/health" "200" "api-health-obs"
    wait_for_status "${OTEL_HEALTH_URL}/" "200" "otel-health-obs"
    # Sanity: Tempo darf NICHT erreichbar sein (Profil ist nicht aktiv).
    # Wenn diese Probe 200 sieht, läuft fast immer ein stale Tempo-
    # Container aus einem früheren `make dev-tempo`; die Diagnose nennt
    # den Cleanup-Pfad explizit, statt einen Collector-Fehler zu
    # suggerieren.
    if curl -sS -o /dev/null -w '%{http_code}' --max-time 2 "${TEMPO_URL}/ready" | grep -q "^200$"; then
      echo "smoke-tempo[observability]: Tempo ist erreichbar, obwohl das tempo-Profil inaktiv sein soll" >&2
      echo "smoke-tempo[observability]: bitte stale Tempo-Container mit 'make stop' beenden und dev-observability neu starten" >&2
      exit 1
    fi
    echo "smoke-tempo[observability]: API + Collector up, Tempo-Profil inaktiv (kein Verbindungsversuch)"
    ;;

  tempo)
    # State 3: Tempo-Profil aktiv.
    wait_for_status "${API_URL}/api/health" "200" "api-health-tempo"
    wait_for_status "${OTEL_HEALTH_URL}/" "200" "otel-health-tempo"
    wait_for_status "${TEMPO_URL}/ready" "200" "tempo-ready"

    search_start=$(($(date +%s) - TEMPO_SEARCH_LOOKBACK_SECONDS))
    session_id="smoke-tempo-$(date +%s)-$$"
    status=$(post_event "$session_id")
    if [ "$status" != "202" ]; then
      echo "smoke-tempo[tempo]: POST /api/playback-events expected 202, got $status" >&2
      cat "$tmpdir/post.body" >&2 || true
      exit 1
    fi
    echo "smoke-tempo[tempo]: posted single-session batch for ${session_id}"

    sleep 2
    correlation_id=$(read_correlation_id "$session_id")
    if [ -z "$correlation_id" ]; then
      echo "smoke-tempo[tempo]: empty correlation_id from API read" >&2
      exit 1
    fi
    echo "smoke-tempo[tempo]: correlation_id=${correlation_id}"

    echo "smoke-tempo[tempo]: waiting ${TEMPO_INGEST_WAIT_SECONDS}s for Tempo to ingest the batch span"
    sleep "$TEMPO_INGEST_WAIT_SECONDS"

    search_end=$(date +%s)

    # Primary Tempo-Search: TraceQL mit explizitem start/end-Fenster.
    # Legacy `tags=` bleibt nur Fallback für ältere Tempo-Setups.
    # Spec-Anker: `spec/telemetry-model.md` §2.6.
    traceql_query="{ span.mtrace.session.correlation_id = \"${correlation_id}\" }"
    encoded_query=$(urlencode "$traceql_query")
    search_response=$(curl -sS "${TEMPO_URL}/api/search?q=${encoded_query}&start=${search_start}&end=${search_end}")
    trace_count=$(printf '%s' "$search_response" | tempo_trace_count)
    if [ "$trace_count" -le 0 ]; then
      encoded_tag=$(urlencode "mtrace.session.correlation_id=${correlation_id}")
      legacy_response=$(curl -sS "${TEMPO_URL}/api/search?tags=${encoded_tag}&start=${search_start}&end=${search_end}")
      legacy_count=$(printf '%s' "$legacy_response" | tempo_trace_count)
      if [ "$legacy_count" -gt 0 ]; then
        search_response="$legacy_response"
        trace_count="$legacy_count"
      fi
    fi
    if [ "$trace_count" -le 0 ]; then
      echo "smoke-tempo[tempo]: no traces found for correlation_id=${correlation_id} in window ${search_start}..${search_end}" >&2
      printf '%s\n' "$search_response" >&2
      exit 1
    fi
    echo "smoke-tempo[tempo]: Tempo returned ${trace_count} trace(s) for correlation_id=${correlation_id}"
    echo "smoke-tempo[tempo]: roundtrip OK (RAK-31 trace search via mtrace.session.correlation_id)"
    ;;

  *)
    echo "smoke-tempo: unknown SMOKE_STATE='$SMOKE_STATE' (expected core|observability|tempo)" >&2
    exit 1
    ;;
esac
