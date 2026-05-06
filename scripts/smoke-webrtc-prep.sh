#!/usr/bin/env bash
set -euo pipefail

# plan-0.7.0 §4 Tranche 3 — WebRTC-Lab-Vorbereitungs-Smoke (RAK-48).
#
# Verifiziert die Vorbereitungsgrenze des WebRTC-Lab-Stacks
# (examples/webrtc/compose.yaml) endpoint-/compose-only — ohne
# Browser, ohne Playback, ohne getStats(). Bewiesen wird:
#   1. Compose-Stack läuft (MediaMTX-Control-API antwortet).
#   2. FFmpeg-Publisher hat den Stream registriert (Pfad ready=true).
#   3. WHEP-Endpoint OPTIONS → 204 für aktiven Pfad.
#   4. WHIP-Endpoint OPTIONS → 204 für aktiven Pfad.
#   5. Pfad-Differenzierung greift (unbekannter Pfad → OPTIONS 500).
#
# Nicht bewiesen: Playback-Qualität, ICE-Erfolgsquote, getStats()-
# Stabilität, Codec-Verhandlung mit echten Browsern. Dafür ist der
# manuelle Browser-Handcheck (RAK-50, examples/webrtc/README.md).
#
# Konvention (examples/README.md):
#   - eigene Compose-Datei → eigener Project-Name `mtrace-webrtc`.
#   - Smoke startet/stoppt nur diesen Project-Namen, räumt keine
#     fremden Volumes/Container auf.
#   - opt-in (nicht in `make gates`).
#
# Manueller Aufruf möglich (Compose-Stack vorher gestartet):
#   docker compose -p mtrace-webrtc -f examples/webrtc/compose.yaml up -d --build
#   SMOKE_WEBRTC_AUTOSTART=0 scripts/smoke-webrtc-prep.sh

PROJECT="${PROJECT:-mtrace-webrtc}"
COMPOSE_FILE="${COMPOSE_FILE:-examples/webrtc/compose.yaml}"
STREAM="${STREAM:-webrtc-test}"
WHIP_WHEP_BASE="${WHIP_WHEP_BASE:-http://localhost:8892}"
API_BASE="${API_BASE:-http://localhost:9999}"
API_USER="${API_USER:-any}"
WAIT_SECONDS="${WAIT_SECONDS:-30}"
SMOKE_WEBRTC_AUTOSTART="${SMOKE_WEBRTC_AUTOSTART:-1}"

if ! command -v curl >/dev/null 2>&1; then
  echo "[smoke-webrtc-prep] missing dependency: curl" >&2
  exit 2
fi
if ! command -v docker >/dev/null 2>&1; then
  echo "[smoke-webrtc-prep] missing dependency: docker" >&2
  exit 2
fi

cleanup() {
  if [ "$SMOKE_WEBRTC_AUTOSTART" = "1" ]; then
    echo "[smoke-webrtc-prep] cleanup: docker compose -p $PROJECT down"
    docker compose -p "$PROJECT" -f "$COMPOSE_FILE" down >/dev/null 2>&1 || true
  fi
}

diagnose_logs() {
  echo "[smoke-webrtc-prep] diagnose:" >&2
  echo "  docker compose -p $PROJECT ps" >&2
  echo "  docker compose -p $PROJECT logs mediamtx | tail -20" >&2
  echo "  docker compose -p $PROJECT logs webrtc-publisher | tail -20" >&2
}

if [ "$SMOKE_WEBRTC_AUTOSTART" = "1" ]; then
  trap cleanup EXIT
  echo "[smoke-webrtc-prep] starting compose project $PROJECT"
  if ! docker compose -p "$PROJECT" -f "$COMPOSE_FILE" up -d --build >/dev/null 2>&1; then
    echo "[smoke-webrtc-prep] docker compose up failed (port conflict on 8892/8189/9999?)" >&2
    echo "[smoke-webrtc-prep] hint: ss -tulpn | grep -E ':(8892|8189|9999)'" >&2
    exit 1
  fi
fi

# 1) MediaMTX-Control-API antwortet (Compose-Stack ist oben)
api_status=""
for _ in $(seq 1 "$WAIT_SECONDS"); do
  api_status="$(curl -sS -u "${API_USER}:" -o /dev/null -w '%{http_code}' "${API_BASE}/v3/paths/list" 2>/dev/null || true)"
  if [ "$api_status" = "200" ]; then
    break
  fi
  sleep 1
done
if [ "$api_status" != "200" ]; then
  echo "[smoke-webrtc-prep] MediaMTX-API unreachable at ${API_BASE}/v3/paths/list (last status: ${api_status:-connection-refused})" >&2
  echo "[smoke-webrtc-prep] hint: prüfe Port 9999 auf Konflikte oder ob der mediamtx-Container hochkommt." >&2
  diagnose_logs
  exit 1
fi
echo "[smoke-webrtc-prep] mediamtx-api OK ($api_status @ ${API_BASE}/v3/paths/list)"

# 2) Stream-Pfad ist registriert (FFmpeg-Publisher → MediaMTX RTSP)
stream_ready=""
for _ in $(seq 1 "$WAIT_SECONDS"); do
  body="$(curl -sS -u "${API_USER}:" "${API_BASE}/v3/paths/list" 2>/dev/null || true)"
  if printf '%s' "$body" | grep -q "\"name\":\"${STREAM}\"" \
     && printf '%s' "$body" | grep -q '"ready":true'; then
    stream_ready="1"
    break
  fi
  sleep 1
done
if [ "$stream_ready" != "1" ]; then
  echo "[smoke-webrtc-prep] stream path '${STREAM}' not ready in MediaMTX (FFmpeg-Publisher liefert keinen Stream?)" >&2
  echo "[smoke-webrtc-prep] hint: typisch 3–8s nach 'up -d' bis zum ersten Frame." >&2
  echo "[smoke-webrtc-prep] paths/list:" >&2
  printf '%s\n' "${body:-<empty>}" | head -5 >&2
  diagnose_logs
  exit 1
fi
echo "[smoke-webrtc-prep] stream-ready OK (${STREAM} ready=true)"

# 3) WHEP-Endpoint OPTIONS → 204 (aktiver Pfad, Listener bedient)
whep_status="$(curl -sS -o /dev/null -w '%{http_code}' -X OPTIONS "${WHIP_WHEP_BASE}/${STREAM}/whep" 2>/dev/null || true)"
if [ "$whep_status" != "204" ]; then
  echo "[smoke-webrtc-prep] WHEP OPTIONS unexpected status: got ${whep_status:-none}, expected 204" >&2
  echo "[smoke-webrtc-prep] hint: 500 = Pfad nicht registriert, 405 = falsche Methode am Listener, leer = kein TCP-Port." >&2
  diagnose_logs
  exit 1
fi
echo "[smoke-webrtc-prep] whep-options OK ($whep_status @ ${WHIP_WHEP_BASE}/${STREAM}/whep)"

# 4) WHIP-Endpoint OPTIONS → 204 (Listener bedient Publish-Surface)
whip_status="$(curl -sS -o /dev/null -w '%{http_code}' -X OPTIONS "${WHIP_WHEP_BASE}/${STREAM}/whip" 2>/dev/null || true)"
if [ "$whip_status" != "204" ]; then
  echo "[smoke-webrtc-prep] WHIP OPTIONS unexpected status: got ${whip_status:-none}, expected 204" >&2
  diagnose_logs
  exit 1
fi
echo "[smoke-webrtc-prep] whip-options OK ($whip_status @ ${WHIP_WHEP_BASE}/${STREAM}/whip)"

# 5) Negativ-Probe: unbekannter Pfad muss differenzierbar antworten
#    (500 für unkonfigurierten Pfad bei MediaMTX 1.x). Zeigt, dass der
#    obige 204 wirklich an unseren Stream gebunden ist und nicht ein
#    pauschales Listener-OK ist.
unknown_path="${STREAM}-does-not-exist-$(date +%s)"
neg_status="$(curl -sS -o /dev/null -w '%{http_code}' -X OPTIONS "${WHIP_WHEP_BASE}/${unknown_path}/whep" 2>/dev/null || true)"
if [ "$neg_status" = "204" ]; then
  echo "[smoke-webrtc-prep] negative probe failed: unknown path '${unknown_path}' answered 204 — Listener differenziert nicht" >&2
  exit 1
fi
echo "[smoke-webrtc-prep] negative-probe OK (unknown path → $neg_status, ≠ 204)"

echo "[smoke-webrtc-prep] all checks passed"
