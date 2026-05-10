#!/usr/bin/env bash
set -euo pipefail

# plan-0.9.0 §2 Tranche 1 (RAK-56) — Browser-Drift-Smoke für
# WebRTC-`getStats()`-Schema-Drift. Schließt R-12 als „automatisiert
# detektiert"; löst die manuelle Drift-Review-Pflicht aus
# `releasing.md` ab.
#
# Was bewiesen wird:
#   1. `mtrace-webrtc`-Lab-Stack läuft (analog smoke-webrtc-prep,
#      delegiert).
#   2. Echte Browser-Versionen (Chromium und Firefox per Default,
#      WebKit/Safari opt-in über `MTRACE_WEBRTC_DRIFT_BROWSERS`)
#      schließen einen WHEP-Handshake gegen
#      `http://localhost:8892/webrtc-test/whep` und liefern alle
#      Muss-Felder pro `RTCStatsType`-Gruppe aus
#      `spec/telemetry-model.md` §3.5.2.
#   3. Alle gemeldeten `connectionState`/`iceConnectionState`/
#      `dtlsState`-Werte liegen in der §1.4-Allowlist.
#
# Was NICHT bewiesen wird:
#   - Soll-Felder sind opt-in pro Engine (§3.5.3); fehlende Soll-
#     Felder werden geloggt, brechen den Smoke aber nicht.
#   - Track-Mounting auf <video>, Codec-Verhandlung mit echten
#     Decodern und Playback-Qualität — dafür ist der manuelle
#     Browser-Handcheck (RAK-50, examples/webrtc/README.md).
#
# Konvention (examples/README.md, plan-0.9.0 §0.5):
#   - opt-in (NICHT in `make gates`); eigener Make-Target
#     `make smoke-webrtc-stats-drift`.
#   - Default-Browser = chromium,firefox; WebKit nur wenn explizit
#     gesetzt: `MTRACE_WEBRTC_DRIFT_BROWSERS=chromium,firefox,webkit`.
#   - Stack-Lifecycle ist auto-up/auto-down auf Project `mtrace-
#     webrtc`; SMOKE_WEBRTC_AUTOSTART=0 hält den Stack offen.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

PROJECT="${PROJECT:-mtrace-webrtc}"
COMPOSE_FILE="${COMPOSE_FILE:-examples/webrtc/compose.yaml}"
WHEP_URL="${MTRACE_WEBRTC_DRIFT_WHEP_URL:-http://localhost:8892/webrtc-test/whep}"
DRIFT_BROWSERS="${MTRACE_WEBRTC_DRIFT_BROWSERS:-chromium,firefox}"
SMOKE_WEBRTC_AUTOSTART="${SMOKE_WEBRTC_AUTOSTART:-1}"
PLAYWRIGHT_TEST_RESULTS_DIR="${PLAYWRIGHT_TEST_RESULTS_DIR:-${TMPDIR:-/tmp}/mtrace-webrtc-drift-results-$$}"
PLAYWRIGHT_IMAGE="${PLAYWRIGHT_IMAGE:-mcr.microsoft.com/playwright:v1.59.1-noble}"

if ! command -v docker >/dev/null 2>&1; then
  echo "[drift-smoke] missing dependency: docker" >&2
  exit 2
fi

cleanup() {
  if [ "$SMOKE_WEBRTC_AUTOSTART" = "1" ]; then
    echo "[drift-smoke] cleanup: docker compose -p $PROJECT down"
    docker compose -p "$PROJECT" -f "$COMPOSE_FILE" down >/dev/null 2>&1 || true
  fi
}

if [ "$SMOKE_WEBRTC_AUTOSTART" = "1" ]; then
  trap cleanup EXIT
  echo "[drift-smoke] preparing mtrace-webrtc stack via scripts/smoke-webrtc-prep.sh"
  # smoke-webrtc-prep startet den Stack, prüft Endpoint-Status und
  # räumt im eigenen EXIT-Trap wieder auf. Wir brauchen den Stack
  # *länger* offen — also delegieren wir nur die Endpoint-Probe und
  # halten den Stack via SMOKE_WEBRTC_AUTOSTART=0 selbst offen.
  if ! docker compose -p "$PROJECT" -f "$COMPOSE_FILE" up -d --build >/dev/null 2>&1; then
    echo "[drift-smoke] docker compose up failed (port conflict on 8892/8189/9999?)" >&2
    echo "[drift-smoke] hint: ss -tulpn | grep -E ':(8892|8189|9999)'" >&2
    exit 1
  fi
  SMOKE_WEBRTC_AUTOSTART=0 bash "$ROOT_DIR/scripts/smoke-webrtc-prep.sh"
fi

echo "[drift-smoke] target WHEP url: $WHEP_URL"
echo "[drift-smoke] driver browsers: $DRIFT_BROWSERS"
echo "[drift-smoke] playwright output: $PLAYWRIGHT_TEST_RESULTS_DIR"
echo "[drift-smoke] playwright image: $PLAYWRIGHT_IMAGE"

project_args=()
IFS=',' read -ra browser_list <<<"$DRIFT_BROWSERS"
for browser in "${browser_list[@]}"; do
  browser_trimmed="$(echo "$browser" | xargs)"
  if [ -z "$browser_trimmed" ]; then
    continue
  fi
  project_args+=(--project="$browser_trimmed")
done

if [ ${#project_args[@]} -eq 0 ]; then
  echo "[drift-smoke] MTRACE_WEBRTC_DRIFT_BROWSERS empty — at least one browser required" >&2
  exit 2
fi

# Spec ist via `MTRACE_WEBRTC_STATS_DRIFT=1` opt-in (verhindert,
# dass `make browser-e2e` den drift-smoke versehentlich mitläuft).
docker run --rm \
  --network host \
  -e CI="${CI:-}" \
  -e PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1 \
  -e MTRACE_WEBRTC_STATS_DRIFT=1 \
  -e MTRACE_WEBRTC_DRIFT_WHEP_URL="$WHEP_URL" \
  -e PLAYWRIGHT_TEST_RESULTS_DIR="$PLAYWRIGHT_TEST_RESULTS_DIR" \
  -v "$ROOT_DIR:/work" \
  -v /work/node_modules \
  -v /work/apps/dashboard/node_modules \
  -v /work/apps/analyzer-service/node_modules \
  -v /work/packages/player-sdk/node_modules \
  -v /work/packages/stream-analyzer/node_modules \
  -w /work \
  "$PLAYWRIGHT_IMAGE" \
  /bin/bash -lc 'corepack enable && pnpm install --frozen-lockfile --config.engine-strict=false && pnpm exec playwright test tests/e2e/webrtc-stats-drift.spec.ts "$@"' \
  bash \
  "${project_args[@]}"

echo "[drift-smoke] all checks passed"
