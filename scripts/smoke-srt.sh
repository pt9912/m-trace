#!/usr/bin/env bash
set -euo pipefail

# plan-0.5.0 §4 Tranche 3 — SRT-Beispiel-Smoke (RAK-37).
#
# Verifiziert den SRT-Pfad funktional: examples/srt/compose.yaml
# startet einen MediaMTX-Container mit SRT-Listener plus einen
# FFmpeg-Publisher, der per SRT einen synthetischen Stream einliefert.
# Der Smoke prüft, dass MediaMTX den Stream als HLS auf dem Host-Port
# 8889 ausspielt — d. h. SRT-Ingress + HLS-Egress beide ok.
#
# Konvention (examples/README.md):
#   - eigene Compose-Datei → eigener Project-Name `mtrace-srt`.
#   - Smoke startet/stoppt nur diesen Project-Namen, räumt keine
#     fremden Volumes/Container auf.
#   - opt-in (nicht in `make gates`).
#
# Manueller Aufruf möglich (Compose-Stack vorher gestartet):
#   docker compose -p mtrace-srt -f examples/srt/compose.yaml up -d --build
#   HLS_URL=http://localhost:8889/srt-test/index.m3u8 \
#     SMOKE_SRT_AUTOSTART=0 scripts/smoke-srt.sh

PROJECT="${PROJECT:-mtrace-srt}"
COMPOSE_FILE="${COMPOSE_FILE:-examples/srt/compose.yaml}"
HLS_URL="${HLS_URL:-http://localhost:8889/srt-test/index.m3u8}"
WAIT_SECONDS="${WAIT_SECONDS:-45}"
SMOKE_SRT_AUTOSTART="${SMOKE_SRT_AUTOSTART:-1}"

if ! command -v curl >/dev/null 2>&1; then
  echo "[smoke-srt] missing dependency: curl" >&2
  exit 2
fi
if ! command -v docker >/dev/null 2>&1; then
  echo "[smoke-srt] missing dependency: docker" >&2
  exit 2
fi

cleanup() {
  if [ "$SMOKE_SRT_AUTOSTART" = "1" ]; then
    echo "[smoke-srt] cleanup: docker compose -p $PROJECT down"
    docker compose -p "$PROJECT" -f "$COMPOSE_FILE" down >/dev/null 2>&1 || true
  fi
}

if [ "$SMOKE_SRT_AUTOSTART" = "1" ]; then
  trap cleanup EXIT
  echo "[smoke-srt] starting compose project $PROJECT"
  docker compose -p "$PROJECT" -f "$COMPOSE_FILE" up -d --build >/dev/null
fi

# 1) HLS-Manifest erreichbar (nach SRT-Publisher-Boot)
status=""
for _ in $(seq 1 "$WAIT_SECONDS"); do
  status="$(curl -sS -L -o /dev/null -w '%{http_code}' "$HLS_URL" 2>/dev/null || true)"
  if [ "$status" = "200" ]; then
    break
  fi
  sleep 1
done
if [ "$status" != "200" ]; then
  echo "[smoke-srt] HLS manifest unreachable at $HLS_URL (last status: ${status:-none})" >&2
  echo "[smoke-srt] hint: SRT-Publisher braucht typisch 10–25s bis zum ersten HLS-Segment." >&2
  echo "[smoke-srt] diagnose:" >&2
  echo "  docker compose -p $PROJECT logs srt-publisher | tail -20" >&2
  echo "  docker compose -p $PROJECT logs mediamtx | tail -20" >&2
  exit 1
fi
echo "[smoke-srt] hls-status OK ($status @ $HLS_URL)"

# 2) Manifest-Body sinnvoll
body="$(curl -sS -L "$HLS_URL" 2>/dev/null || true)"
if ! printf '%s' "$body" | grep -q '^#EXTM3U'; then
  echo "[smoke-srt] HLS body missing #EXTM3U header — got:" >&2
  printf '%s\n' "${body:-<empty>}" | head -10 >&2
  exit 1
fi
echo "[smoke-srt] hls-body OK (#EXTM3U present)"

# Master-Playlist verweist auf `.m3u8`-Substreams, Media-Playlist auf
# `.ts`/`.m4s`/`.aac`-Segmente. Beide sind valide Signale, dass der
# SRT-Publisher mit MediaMTX verbunden ist und Inhalte fließen.
if ! printf '%s' "$body" | grep -qE '\.(m3u8|ts|m4s|aac)(\?|$|[[:space:]])'; then
  echo "[smoke-srt] HLS manifest has no media references (.m3u8/.ts/.m4s/.aac) — SRT publisher not yet connected?" >&2
  printf '%s\n' "$body" | head -20 >&2
  exit 1
fi
echo "[smoke-srt] hls-segments OK (media references present)"

echo "[smoke-srt] all checks passed"
