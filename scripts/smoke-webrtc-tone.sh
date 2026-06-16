#!/usr/bin/env bash
set -euo pipefail

# WebRTC-Ton-Smoke (Folge zu RAK-56).
#
# Verifiziert, dass die WebRTC-Lab-Medien-Pipeline einen sauberen
# 1-kHz-Sinuston bis zum Egress traegt — die automatisierte
# Entsprechung der manuellen „1-kHz-Sinuston hoerbar"-Abnahme aus
# docs/user/releasing.md. Der Lab-Publisher erzeugt
# `sine=frequency=1000:sample_rate=48000` (Opus, examples/webrtc/
# ffmpeg-rtsp-loop.sh); dieser Smoke zieht denselben Stream und prueft
# per Goertzel-Einzel-Bin-DFT, dass die Zielfrequenz dominiert.
#
# Mechanik (host-seitig kein RTSP erreichbar — 8554 ist
# Container-intern): ein einmaliger ffmpeg-Container im
# `mtrace-webrtc`-Netz zieht `rtsp://mediamtx:8554/<stream>`, dekodiert
# nach Mono-48k-f32le und pipet das PCM an `scripts/check-tone.mjs` auf
# dem Host. Kein Browser noetig — geprueft wird die Pipeline bis
# MediaMTX (Publisher → RTSP → MediaMTX-Egress), komplementaer zum
# getStats-Drift-Smoke, der nur `bytesReceived>0`, nicht die
# Tonqualitaet verifiziert.
#
# Abgrenzung: Dies ersetzt NICHT die perzeptuelle Operator-Abnahme
# („klingt/sieht das ganze Demo im echten Browser richtig"), sondern
# nur den eng definierten „ist ein sauberer 1-kHz-Ton da"-Teil.
#
# Gate-Typ: opt-in lokal + Nightly (webrtc-drift.yml), NICHT
# PR-blockierend — das WebRTC-Lab ist unter Last flaky. Spiegelt das
# Lifecycle-Muster von scripts/smoke-webrtc-stats-drift.sh.
#
# Manueller Aufruf gegen einen bereits laufenden Stack:
#   docker compose -p mtrace-webrtc -f examples/webrtc/compose.yaml up -d --build
#   SMOKE_WEBRTC_AUTOSTART=0 scripts/smoke-webrtc-tone.sh

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

PROJECT="${PROJECT:-mtrace-webrtc}"
COMPOSE_FILE="${COMPOSE_FILE:-examples/webrtc/compose.yaml}"
STREAM="${STREAM:-webrtc-test}"
RTSP_URL="${RTSP_URL:-rtsp://mediamtx:8554/${STREAM}}"
CAPTURE_SECONDS="${CAPTURE_SECONDS:-3}"
RATE="${RATE:-48000}"
FREQ="${FREQ:-1000}"
MIN_FRACTION="${MIN_FRACTION:-0.2}"
MIN_RMS="${MIN_RMS:-0.01}"
SMOKE_WEBRTC_AUTOSTART="${SMOKE_WEBRTC_AUTOSTART:-1}"

if ! command -v docker >/dev/null 2>&1; then
  echo "[tone-smoke] missing dependency: docker" >&2
  exit 2
fi
if ! command -v node >/dev/null 2>&1; then
  echo "[tone-smoke] missing dependency: node" >&2
  exit 2
fi

tmpdir="$(mktemp -d)"
cleanup() {
  status=$?
  rm -rf "$tmpdir"
  if [ "$SMOKE_WEBRTC_AUTOSTART" = "1" ]; then
    echo "[tone-smoke] cleanup: docker compose -p $PROJECT down"
    docker compose -p "$PROJECT" -f "$COMPOSE_FILE" down >/dev/null 2>&1 || true
  fi
  exit $status
}
trap cleanup EXIT

if [ "$SMOKE_WEBRTC_AUTOSTART" = "1" ]; then
  echo "[tone-smoke] starting lab: docker compose -p $PROJECT up -d --build"
  if ! docker compose -p "$PROJECT" -f "$COMPOSE_FILE" up -d --build >/dev/null 2>&1; then
    echo "[tone-smoke] docker compose up failed (port conflict on 8892/8189/9999?)" >&2
    exit 1
  fi
  # Readiness (Compose-Stack + Publisher-Pfad ready) ueber den
  # vorhandenen prep-Smoke; haelt den Stack offen (AUTOSTART=0).
  echo "[tone-smoke] waiting for lab readiness via smoke-webrtc-prep..."
  SMOKE_WEBRTC_AUTOSTART=0 bash "$ROOT_DIR/scripts/smoke-webrtc-prep.sh"
fi

echo "[tone-smoke] capturing ${CAPTURE_SECONDS}s of ${RTSP_URL} (mono ${RATE}Hz f32le) via in-network ffmpeg..."
# ffmpeg laeuft im mtrace-webrtc-Netz (--no-deps: mediamtx laeuft schon),
# damit der Service-Name `mediamtx` aufgeloest wird. Stdout (rohes PCM)
# in eine Host-Tempdatei, damit kein Compose-Diagnose-Text das Binaer-
# Format verunreinigt.
pcm="$tmpdir/tone.f32"
if ! docker compose -p "$PROJECT" -f "$COMPOSE_FILE" run --rm --no-deps -T \
  --entrypoint ffmpeg webrtc-publisher \
  -hide_banner -loglevel error -rtsp_transport tcp \
  -i "$RTSP_URL" \
  -t "$CAPTURE_SECONDS" -vn -ac 1 -ar "$RATE" -f f32le - \
  > "$pcm"; then
  echo "[tone-smoke] ffmpeg RTSP capture failed" >&2
  exit 1
fi

pcm_bytes="$(wc -c < "$pcm")"
echo "[tone-smoke] captured ${pcm_bytes} bytes ($((pcm_bytes / 4)) samples)"
if [ "$pcm_bytes" -lt $((RATE * 4 / 2)) ]; then
  echo "[tone-smoke] captured < 0.5s of audio — egress likely not flowing" >&2
  exit 1
fi

node "$ROOT_DIR/scripts/check-tone.mjs" \
  --rate "$RATE" --freq "$FREQ" \
  --min-fraction "$MIN_FRACTION" --min-rms "$MIN_RMS" \
  < "$pcm"

echo "[tone-smoke] OK -- clean ${FREQ}Hz tone verified through the WebRTC lab pipeline."
