#!/usr/bin/env bash
set -euo pipefail

# — SRS-Beispiel-Smoke (RAK-57, Kann).
#
# Verifiziert den SRS-Pfad funktional: examples/srs/compose.yaml
# startet einen SRS-Container mit RTMP-Listener + HTTP-API + HTTP-FLV-
# Egress plus einen FFmpeg-Publisher, der per RTMP einen synthetischen
# Stream einliefert. Der Smoke prüft endpoint-/compose-only:
#
# 1. HTTP-API antwortet 200 auf `/api/v1/streams/`.
# 2. Stream `live/srs-test` ist registriert mit `publish.active=true`.
# 3. HTTP-FLV-Egress liefert 200 plus FLV-Magic-Header (`FLV`) für
# `/live/srs-test.flv` — d. h. RTMP-Ingress + FLV-Egress fließen.
#
# Nicht bewiesen: Playback-Qualität (Codec-Verhandlung, Latenz),
# HLS-Output (SRS unterstützt das, aber dafür gibt's andere
# Beispiele), Player-SDK-/Dashboard-Anbindung. Der Pfad ist analog
# zu `make smoke-srt` — endpoint-only, keine Telemetrie.
#
# Konvention (examples/README.md):
# - eigene Compose-Datei → eigener Project-Name `mtrace-srs`.
# - Smoke startet/stoppt nur diesen Project-Namen, räumt keine
# fremden Volumes/Container auf.
# - opt-in (nicht in `make gates`).
#
# Manueller Aufruf möglich (Compose-Stack vorher gestartet):
# docker compose -p mtrace-srs -f examples/srs/compose.yaml up -d --build
# SMOKE_SRS_AUTOSTART=0 scripts/smoke-srs.sh

PROJECT="${PROJECT:-mtrace-srs}"
COMPOSE_FILE="${COMPOSE_FILE:-examples/srs/compose.yaml}"
API_BASE="${API_BASE:-http://localhost:1985}"
FLV_URL="${FLV_URL:-http://localhost:8088/live/srs-test.flv}"
STREAM_APP="${STREAM_APP:-live}"
STREAM_NAME="${STREAM_NAME:-srs-test}"
WAIT_SECONDS="${WAIT_SECONDS:-30}"
SMOKE_SRS_AUTOSTART="${SMOKE_SRS_AUTOSTART:-1}"

if ! command -v curl >/dev/null 2>&1; then
  echo "[smoke-srs] missing dependency: curl" >&2
  exit 2
fi
if ! command -v docker >/dev/null 2>&1; then
  echo "[smoke-srs] missing dependency: docker" >&2
  exit 2
fi

cleanup() {
  if [ "$SMOKE_SRS_AUTOSTART" = "1" ]; then
    echo "[smoke-srs] cleanup: docker compose -p $PROJECT down"
    docker compose -p "$PROJECT" -f "$COMPOSE_FILE" down >/dev/null 2>&1 || true
  fi
}

diagnose_logs() {
  echo "[smoke-srs] diagnose:" >&2
  echo "  docker compose -p $PROJECT ps" >&2
  echo "  docker compose -p $PROJECT logs srs | tail -20" >&2
  echo "  docker compose -p $PROJECT logs srs-publisher | tail -20" >&2
}

if [ "$SMOKE_SRS_AUTOSTART" = "1" ]; then
  trap cleanup EXIT
  echo "[smoke-srs] starting compose project $PROJECT"
  if ! docker compose -p "$PROJECT" -f "$COMPOSE_FILE" up -d --build >/dev/null 2>&1; then
    echo "[smoke-srs] docker compose up failed (port conflict on 1935/1985/8088?)" >&2
    echo "[smoke-srs] hint: ss -tulpn | grep -E ':(1935|1985|8088)'" >&2
    exit 1
  fi
fi

# 1) HTTP-API antwortet 200 auf /api/v1/streams/
api_status=""
for _ in $(seq 1 "$WAIT_SECONDS"); do
  api_status="$(curl -sS -o /dev/null -w '%{http_code}' "${API_BASE}/api/v1/streams/" 2>/dev/null || true)"
  if [ "$api_status" = "200" ]; then
    break
  fi
  sleep 1
done
if [ "$api_status" != "200" ]; then
  echo "[smoke-srs] SRS HTTP-API unreachable at ${API_BASE}/api/v1/streams/ (last status: ${api_status:-connection-refused})" >&2
  echo "[smoke-srs] hint: SRS braucht typisch 2–5 s bis der HTTP-API-Listener aktiv ist." >&2
  diagnose_logs
  exit 1
fi
echo "[smoke-srs] api-status OK ($api_status @ ${API_BASE}/api/v1/streams/)"

# 2) Stream `live/srs-test` ist registriert mit publish.active=true
# SRS antwortet im Format {"code":0,"streams":[{"app":"live","name":"srs-test","publish":{"active":true,...},...}]}
stream_ready=""
for _ in $(seq 1 "$WAIT_SECONDS"); do
  body="$(curl -sS "${API_BASE}/api/v1/streams/" 2>/dev/null || true)"
  if printf '%s' "$body" | grep -q "\"app\":\"${STREAM_APP}\"" \
     && printf '%s' "$body" | grep -q "\"name\":\"${STREAM_NAME}\"" \
     && printf '%s' "$body" | grep -q '"active":true'; then
    stream_ready="1"
    break
  fi
  sleep 1
done
if [ "$stream_ready" != "1" ]; then
  echo "[smoke-srs] stream '${STREAM_APP}/${STREAM_NAME}' not active in SRS (FFmpeg-Publisher liefert keinen Stream?)" >&2
  echo "[smoke-srs] hint: typisch 3–8 s nach 'up -d' bis zum ersten RTMP-Connect." >&2
  echo "[smoke-srs] /api/v1/streams/ body:" >&2
  printf '%s\n' "${body:-<empty>}" | head -5 >&2
  diagnose_logs
  exit 1
fi
echo "[smoke-srs] stream-ready OK (${STREAM_APP}/${STREAM_NAME} publish.active=true)"

# 3) HTTP-FLV-Egress liefert 200 + FLV-Magic-Header `FLV`
# `--max-time 3` schneidet das Streaming nach 3 s ab; wir lesen
# nur ein paar Bytes vom Anfang, um den Magic-Header zu prüfen.
flv_dump="$(mktemp)"
trap 'rm -f "$flv_dump"' RETURN 2>/dev/null || true
flv_status="$(curl -sS -o "$flv_dump" -w '%{http_code}' --max-time 3 "$FLV_URL" 2>/dev/null || true)"
if [ "$flv_status" != "200" ]; then
  # `curl` mit `--max-time 3` bricht das Streaming nach 3 s ab; wenn
  # SRS bereits Bytes geliefert hat, ist das ein Erfolg, auch wenn
  # curl mit Exit-Code 28 (Timeout) abbricht. status=000 = kein
  # Response-Header empfangen, das wäre tatsächlich Fehler.
  if [ -s "$flv_dump" ]; then
    flv_status="200"
  else
    echo "[smoke-srs] HTTP-FLV egress unreachable at $FLV_URL (status: ${flv_status:-none}, no body)" >&2
    diagnose_logs
    rm -f "$flv_dump"
    exit 1
  fi
fi
flv_magic="$(head -c 3 "$flv_dump" 2>/dev/null || true)"
rm -f "$flv_dump"
if [ "$flv_magic" != "FLV" ]; then
  echo "[smoke-srs] HTTP-FLV body missing FLV magic header (got: ${flv_magic:-<empty>})" >&2
  diagnose_logs
  exit 1
fi
echo "[smoke-srs] flv-magic OK (FLV header @ $FLV_URL)"

echo "[smoke-srs] all checks passed"
