#!/usr/bin/env bash
set -euo pipefail

# Smoke für plan-0.3.0 Tranche 7 (HLS) plus plan-0.9.0 Tranche 3
# Erweiterung um den DASH-Pfad (RAK-58/RAK-59 / NF-12):
#  1. `pnpm --silent m-trace --help` zeigt die Usage und exit 0.
#  2. `pnpm --silent m-trace check <file.m3u8>` analysiert eine
#     Master-HLS-Manifest-Fixture, gibt JSON auf stdout aus, exit 0
#     (`analyzerKind:"hls"`, `playlistType:"master"`).
#  3. `pnpm --silent m-trace check <file.mpd>` analysiert eine VOD-
#     DASH-MPD-Fixture, gibt JSON auf stdout aus, exit 0
#     (`analyzerKind:"dash"`, `playlistType:"dash"`). Ab plan-0.9.0
#     Tranche 3.
#  4. `pnpm --silent m-trace check <not-supported>` (HTML-Body)
#     bricht ab mit exit 1 und JSON, das `status:"error"` mit
#     `code:"manifest_not_supported"` trägt — neuer Detector-Pfad
#     ab plan-0.9.0 Tranche 3.
#  5. `pnpm --silent m-trace check /nonexistent.m3u8` läuft auf IO-Fehler,
#     exit 1, Fehler-Message auf stderr.
#  6. Aufruf ohne Argumente → Usage-Fehler, exit 2.
#  7. URL-Loader-Pfad (SSRF-Block) → exit 1 + fetch_blocked.
#  8. bin-Pfad via consumer .bin → Usage.
#
# Erwartet `make ts-build` als Vorbedingung (das Makefile-Target
# `smoke-cli` hängt schon dran).

if ! command -v jq >/dev/null 2>&1; then
  echo "[smoke-cli] missing dependency: jq" >&2
  exit 2
fi

repo_root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$repo_root"

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

# Master-Fixture
master="$tmpdir/master.m3u8"
cat > "$master" <<'M3U'
#EXTM3U
#EXT-X-VERSION:6
#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="aud-en",NAME="English",DEFAULT=YES,URI="audio/en.m3u8"
#EXT-X-STREAM-INF:BANDWIDTH=1280000,RESOLUTION=1280x720,AUDIO="aud-en"
video/720p.m3u8
M3U

# Non-HLS-/Non-DASH-Fixture (HTML-Body) — wird vom Detector in
# plan-0.9.0 Tranche 3 als `manifest_not_supported` zurückgewiesen.
nothls="$tmpdir/not-hls.txt"
echo "<html><body>not a manifest</body></html>" > "$nothls"

# DASH-VOD-Fixture (plan-0.9.0 Tranche 3)
dash_vod="$tmpdir/vod.mpd"
cat > "$dash_vod" <<'MPD'
<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" type="static" mediaPresentationDuration="PT10M0S" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011">
  <Period>
    <AdaptationSet id="1" contentType="video" mimeType="video/mp4">
      <Representation id="v1" bandwidth="1280000" width="1280" height="720" codecs="avc1.4d401e" frameRate="30"/>
    </AdaptationSet>
  </Period>
</MPD>
MPD

# 1. --help
help_out="$(pnpm --silent m-trace --help 2>&1)"
if ! grep -q "Usage: m-trace check" <<<"$help_out"; then
  echo "[smoke-cli] --help did not produce expected usage:"
  echo "$help_out"
  exit 1
fi
echo "[smoke-cli] --help OK"

# 2. happy path
master_out="$(pnpm --silent m-trace check "$master")"
if ! echo "$master_out" | jq -e '.status == "ok" and .playlistType == "master"' >/dev/null; then
  echo "[smoke-cli] master case did not return ok/master:"
  echo "$master_out"
  exit 1
fi
echo "[smoke-cli] check master fixture OK"

# 3. DASH-VOD → analyzerKind=dash, playlistType=dash, exit 0
dash_out="$(pnpm --silent m-trace check "$dash_vod")"
if ! echo "$dash_out" | jq -e '.status == "ok" and .analyzerKind == "dash" and .playlistType == "dash"' >/dev/null; then
  echo "[smoke-cli] DASH VOD case did not return ok/dash/dash:"
  echo "$dash_out"
  exit 1
fi
if ! echo "$dash_out" | jq -e '.details.adaptationSets | length >= 1' >/dev/null; then
  echo "[smoke-cli] DASH VOD result has no adaptationSets:"
  echo "$dash_out"
  exit 1
fi
echo "[smoke-cli] check DASH VOD fixture OK (analyzerKind=dash, playlistType=dash)"

# 4. unsupported (HTML) → analysis-error manifest_not_supported → exit 1
set +e
nothls_out="$(pnpm --silent m-trace check "$nothls" 2>/dev/null)"
nothls_exit=$?
set -e
if [ "$nothls_exit" != "1" ]; then
  echo "[smoke-cli] unsupported expected exit 1, got $nothls_exit"
  exit 1
fi
if ! echo "$nothls_out" | jq -e '.status == "error" and .code == "manifest_not_supported"' >/dev/null; then
  echo "[smoke-cli] unsupported did not return manifest_not_supported:"
  echo "$nothls_out"
  exit 1
fi
echo "[smoke-cli] check unsupported fixture OK (exit 1, manifest_not_supported)"

# 4. missing file → IO error → exit 1, stderr message
set +e
miss_stderr="$(pnpm --silent m-trace check /nonexistent.m3u8 2>&1 >/dev/null)"
miss_exit=$?
set -e
if [ "$miss_exit" != "1" ]; then
  echo "[smoke-cli] missing-file expected exit 1, got $miss_exit"
  exit 1
fi
if ! grep -q "konnte nicht gelesen werden" <<<"$miss_stderr"; then
  echo "[smoke-cli] missing-file stderr did not mention IO error:"
  echo "$miss_stderr"
  exit 1
fi
echo "[smoke-cli] missing file OK (exit 1, stderr message)"

# 5. no args → usage error → strikt exit 2 (pnpm 10 propagiert
#    Script-Exit-Codes verbatim).
set +e
pnpm --silent m-trace >/dev/null 2>&1
noargs_exit=$?
set -e
if [ "$noargs_exit" != "2" ]; then
  echo "[smoke-cli] no-args expected exit 2, got $noargs_exit"
  exit 1
fi
echo "[smoke-cli] no-args usage error OK (exit 2)"

# 6. URL-Loader-Pfad end-to-end: SSRF-Block gegen RFC1918 → exit 1
#    + JSON mit code:fetch_blocked. Bestätigt, dass URL-Inputs durch
#    den echten analyzeHlsManifest-Loader laufen (nicht nur durch den
#    runCli-Dispatcher). Kein localhost-Server nötig — der SSRF-Block
#    selbst ist der positiv-bestätigte Pfad.
set +e
ssrf_out="$(pnpm --silent m-trace check http://10.0.0.1/m.m3u8)"
ssrf_exit=$?
set -e
if [ "$ssrf_exit" != "1" ]; then
  echo "[smoke-cli] URL SSRF expected exit 1, got $ssrf_exit"
  echo "$ssrf_out"
  exit 1
fi
if ! echo "$ssrf_out" | jq -e '.status == "error" and .code == "fetch_blocked"' >/dev/null; then
  echo "[smoke-cli] URL SSRF did not return fetch_blocked:"
  echo "$ssrf_out"
  exit 1
fi
echo "[smoke-cli] URL → fetch_blocked OK (real loader-Pfad)"

# 7. bin-Pfad: analyzer-service hängt als Workspace-Consumer am
#    stream-analyzer und bekommt das m-trace-Bin in sein node_modules/
#    .bin/ symlinked — exakt die Situation, die published-package-
#    Konsumenten nach `npm install @npm9912/stream-analyzer` haben.
#    `pnpm --filter ... exec m-trace --help` exerciert das Symlink
#    plus den Shebang plus den Executable-Bit in einem Aufruf.
help_via_bin="$(pnpm --silent --filter @npm9912/analyzer-service exec m-trace --help)"
if ! grep -q "Usage: m-trace check" <<<"$help_via_bin"; then
  echo "[smoke-cli] m-trace via consumer node_modules/.bin did not produce usage:"
  echo "$help_via_bin"
  exit 1
fi
echo "[smoke-cli] bin via consumer .bin OK"

echo "[smoke-cli] all checks passed"
