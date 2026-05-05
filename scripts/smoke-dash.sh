#!/usr/bin/env bash
set -euo pipefail

# plan-0.5.0 §5 Tranche 4 — DASH-Beispiel-Smoke (RAK-38).
#
# Verifiziert den DASH-Pfad funktional: examples/dash/compose.yaml
# startet einen FFmpeg-Generator, der einen synthetischen Stream
# als DASH-Live-Output in ein shared Volume schreibt; nginx
# serviert das Volume statisch auf Host-Port 8891.
# Der Smoke prüft, dass die MPD erreichbar ist und mindestens ein
# referenziertes Segment lokal abrufbar ist — d. h. der DASH-Origin
# läuft komplett offline ohne CDN.
#
# Konvention (examples/README.md):
#   - eigene Compose-Datei → eigener Project-Name `mtrace-dash`.
#   - Smoke startet/stoppt nur diesen Project-Namen, räumt keine
#     fremden Volumes/Container auf.
#   - opt-in (nicht in `make gates`).
#
# Manueller Aufruf möglich (Compose-Stack vorher gestartet):
#   docker compose -p mtrace-dash -f examples/dash/compose.yaml up -d --build
#   SMOKE_DASH_AUTOSTART=0 scripts/smoke-dash.sh

PROJECT="${PROJECT:-mtrace-dash}"
COMPOSE_FILE="${COMPOSE_FILE:-examples/dash/compose.yaml}"
DASH_BASE="${DASH_BASE:-http://localhost:8891}"
MPD_URL="${MPD_URL:-${DASH_BASE}/manifest.mpd}"
WAIT_SECONDS="${WAIT_SECONDS:-45}"
SMOKE_DASH_AUTOSTART="${SMOKE_DASH_AUTOSTART:-1}"

if ! command -v curl >/dev/null 2>&1; then
  echo "[smoke-dash] missing dependency: curl" >&2
  exit 2
fi
if ! command -v docker >/dev/null 2>&1; then
  echo "[smoke-dash] missing dependency: docker" >&2
  exit 2
fi

cleanup() {
  if [ "$SMOKE_DASH_AUTOSTART" = "1" ]; then
    echo "[smoke-dash] cleanup: docker compose -p $PROJECT down"
    docker compose -p "$PROJECT" -f "$COMPOSE_FILE" down --volumes >/dev/null 2>&1 || true
  fi
}

if [ "$SMOKE_DASH_AUTOSTART" = "1" ]; then
  trap cleanup EXIT
  echo "[smoke-dash] starting compose project $PROJECT"
  docker compose -p "$PROJECT" -f "$COMPOSE_FILE" up -d --build >/dev/null
fi

# 1) MPD erreichbar
status=""
for _ in $(seq 1 "$WAIT_SECONDS"); do
  status="$(curl -sS -L -o /dev/null -w '%{http_code}' "$MPD_URL" 2>/dev/null || true)"
  if [ "$status" = "200" ]; then
    break
  fi
  sleep 1
done
if [ "$status" != "200" ]; then
  echo "[smoke-dash] MPD unreachable at $MPD_URL (last status: ${status:-none})" >&2
  echo "[smoke-dash] hint: FFmpeg-DASH-Generator braucht ~10–20s bis zum ersten Manifest." >&2
  echo "[smoke-dash] diagnose:" >&2
  echo "  docker compose -p $PROJECT logs dash-generator | tail -20" >&2
  echo "  docker compose -p $PROJECT logs dash-server | tail -20" >&2
  exit 1
fi
echo "[smoke-dash] mpd-status OK ($status @ $MPD_URL)"

# 2) MPD-Body sinnvoll
body="$(curl -sS -L "$MPD_URL" 2>/dev/null || true)"
if ! printf '%s' "$body" | grep -q '<MPD'; then
  echo "[smoke-dash] MPD body missing <MPD root element — got:" >&2
  printf '%s\n' "${body:-<empty>}" | head -10 >&2
  exit 1
fi
echo "[smoke-dash] mpd-body OK (<MPD present)"

# 3) Mindestens ein referenziertes Segment HEAD-erreichbar.
# Wir nutzen den `media`-Attribut-Pfad aus SegmentTemplate. FFmpeg
# erzeugt typische Pfade wie `chunk-stream0-NNNNN.m4s` und
# `init-stream0.m4s`. Wir suchen nach einem .m4s/.mp4-Eintrag im MPD.
segment_ref="$(printf '%s' "$body" | grep -oE '(initialization|media)="[^"]+"' | head -1 || true)"
if [ -z "$segment_ref" ]; then
  echo "[smoke-dash] MPD has no SegmentTemplate initialization/media reference" >&2
  exit 1
fi
# FFmpeg-DASH nutzt `init-stream$RepresentationID$.m4s`-Templates.
# Für den Smoke nutzen wir eine konkrete Init-Datei.
init_path="init-stream0.m4s"
init_status="$(curl -sS -L -o /dev/null -w '%{http_code}' "$DASH_BASE/$init_path" 2>/dev/null || true)"
if [ "$init_status" != "200" ]; then
  echo "[smoke-dash] init segment unreachable at $DASH_BASE/$init_path (status: ${init_status:-none})" >&2
  echo "[smoke-dash] MPD reference:" >&2
  printf '  %s\n' "$segment_ref" >&2
  exit 1
fi
echo "[smoke-dash] init-segment OK ($init_status @ $DASH_BASE/$init_path)"

echo "[smoke-dash] all checks passed"
