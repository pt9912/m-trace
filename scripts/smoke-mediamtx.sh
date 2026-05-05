#!/usr/bin/env bash
set -euo pipefail

# plan-0.5.0 §3 Tranche 2 — MediaMTX-Beispiel-Smoke (RAK-36).
#
# Verifiziert den bestehenden Core-Lab-MediaMTX-Pfad funktional über
# die HLS-Auslieferung:
#   1. HLS-Manifest erreichbar unter der dokumentierten URL
#      (HTTP 200 mit bounded Wait — der FFmpeg-Publisher braucht
#      einige Sekunden bis zum ersten Segment).
#   2. Manifest-Body enthält `#EXTM3U` (gültiges M3U8) und mindestens
#      einen Segment-Eintrag (`.ts`, `.m4s` oder `.aac` — also keine
#      leere/initiale Playlist).
#
# Erwartet: Core-Lab läuft (`make dev`). Manueller Aufruf möglich:
#   HLS_URL=http://localhost:8888/teststream/index.m3u8 \
#     scripts/smoke-mediamtx.sh
#
# Hinweis: MediaMTX 1.14+ schaltet die Control-API standardmäßig
# Auth-pflichtig. Dieser Smoke prüft daher absichtlich nicht den
# `:9997`-API-Endpoint, sondern den funktionalen HLS-Pfad — wenn
# HLS ausspielt, ist sowohl MediaMTX als auch der teststream-
# Publisher erreichbar. API-Erreichbarkeit ist eine optionale
# Operator-Diagnose, kein Smoke-Gate.
#
# Konvention: räumt keine fremden Volumes/Container auf — Stop-Pfad
# ist `make stop` durch den Operator.

HLS_URL="${HLS_URL:-http://localhost:8888/teststream/index.m3u8}"
WAIT_SECONDS="${WAIT_SECONDS:-30}"

if ! command -v curl >/dev/null 2>&1; then
  echo "[smoke-mediamtx] missing dependency: curl" >&2
  exit 2
fi

# 1) HLS-Manifest erreichbar
status=""
for _ in $(seq 1 "$WAIT_SECONDS"); do
  status="$(curl -sS -L -o /dev/null -w '%{http_code}' "$HLS_URL" 2>/dev/null || true)"
  if [ "$status" = "200" ]; then
    break
  fi
  sleep 1
done
if [ "$status" != "200" ]; then
  echo "[smoke-mediamtx] HLS manifest unreachable at $HLS_URL (last status: ${status:-none})" >&2
  echo "[smoke-mediamtx] hint: ist 'make dev' gestartet und Port 8888 frei?" >&2
  echo "[smoke-mediamtx] diagnose: 'docker compose logs stream-generator | tail -20' und 'docker compose logs mediamtx | tail -20'" >&2
  exit 1
fi
echo "[smoke-mediamtx] hls-status OK ($status @ $HLS_URL)"

# 2) Manifest-Body sinnvoll
body="$(curl -sS -L "$HLS_URL" 2>/dev/null || true)"
if ! printf '%s' "$body" | grep -q '^#EXTM3U'; then
  echo "[smoke-mediamtx] HLS body missing #EXTM3U header — got:" >&2
  printf '%s\n' "${body:-<empty>}" | head -10 >&2
  exit 1
fi
echo "[smoke-mediamtx] hls-body OK (#EXTM3U present)"

# Master-Playlist verweist auf `.m3u8`-Substreams, Media-Playlist auf
# `.ts`/`.m4s`/`.aac`-Segmente. Beide sind valide Signale, dass der
# teststream läuft und MediaMTX echte Inhalte serviert.
if ! printf '%s' "$body" | grep -qE '\.(m3u8|ts|m4s|aac)(\?|$|[[:space:]])'; then
  echo "[smoke-mediamtx] HLS manifest has no media references (.m3u8/.ts/.m4s/.aac) — teststream not yet publishing?" >&2
  printf '%s\n' "$body" | head -20 >&2
  exit 1
fi
echo "[smoke-mediamtx] hls-segments OK (media references present)"

echo "[smoke-mediamtx] all checks passed"
