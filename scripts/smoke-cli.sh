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

# 8. CMAF-HLS- und CMAF-DASH-Probe mit lokalem HTTP-Server
#    (plan-0.10.0 Tranche 5, NF-13 / RAK-63 / RAK-64). Erzeugt
#    deterministische CMAF-Init-/Media-Bytes via
#    scripts/cmaf-fixture-builder.mjs, serviert sie über
#    `python3 -m http.server` und ruft die CLI mit
#    `MTRACE_CHECK_ALLOW_PRIVATE_NETWORKS=true` auf. Erwartet
#    `analysis.details.cmaf.binary.status="passed"` für beide
#    Manifest-Formate. Datei-Input mit `file://`-baseUrl gilt laut
#    Plan T5-DoD nicht als Ersatz, weil der Segment-Loader strikt
#    HTTP(S)-SSRF-Regeln nutzt.
if ! command -v python3 >/dev/null 2>&1; then
  echo "[smoke-cli] CMAF probes skipped — python3 not available" >&2
  echo "[smoke-cli] all checks passed (CMAF probes skipped)"
  exit 0
fi

cmaf_dir="$tmpdir/cmaf"
node scripts/cmaf-fixture-builder.mjs "$cmaf_dir" >/dev/null

cat > "$cmaf_dir/master.m3u8" <<'M3U'
#EXTM3U
#EXT-X-VERSION:7
#EXT-X-TARGETDURATION:4
#EXT-X-MEDIA-SEQUENCE:0
#EXT-X-MAP:URI="init.mp4"
#EXTINF:4.0,
seg-0.m4s
#EXT-X-ENDLIST
M3U

# DASH-MPD mit SegmentTemplate (init.mp4 + seg-$Number$.m4s; das
# erste Segment ist seg-1.m4s — Builder erzeugt seg-0.m4s als
# Default. Wir benennen die Datei deshalb nach dem Templating um.)
cp "$cmaf_dir/seg-0.m4s" "$cmaf_dir/seg-1.m4s"
cat > "$cmaf_dir/manifest.mpd" <<'MPD'
<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" type="static" mediaPresentationDuration="PT10M0S" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011">
  <Period>
    <AdaptationSet id="v" mimeType="video/mp4">
      <SegmentTemplate initialization="init.mp4" media="seg-$Number$.m4s" startNumber="1"/>
      <Representation id="v1" bandwidth="1280000" width="1280" height="720" codecs="avc1.4d401e"/>
    </AdaptationSet>
  </Period>
</MPD>
MPD

# Pick a random free port and start python3 -m http.server in
# tmpdir. The server is killed on EXIT via the existing trap; we
# extend that trap to also stop the python pid.
http_port=0
# Try a few candidate ports; bind quickly with a tiny python wrapper.
http_log="$tmpdir/http-server.log"
( cd "$cmaf_dir" && exec python3 -u -m http.server 0 ) >"$http_log" 2>&1 &
http_pid=$!
trap 'rm -rf "$tmpdir"; kill "$http_pid" 2>/dev/null || true' EXIT

# Wait until the server printed its bound port (max ~5s).
for _ in $(seq 1 50); do
  if grep -q "Serving HTTP" "$http_log" 2>/dev/null; then
    break
  fi
  sleep 0.1
done
http_port="$(grep -oE 'port [0-9]+' "$http_log" | awk '{print $2}' | head -n1)"
if [ -z "$http_port" ]; then
  echo "[smoke-cli] failed to bring up python3 http.server on tmpdir:"
  cat "$http_log"
  exit 1
fi
echo "[smoke-cli] CMAF probe HTTP server up on port $http_port"

# 8a. CMAF-HLS-Probe.
cmaf_hls_url="http://127.0.0.1:${http_port}/master.m3u8"
cmaf_hls_out="$(MTRACE_CHECK_ALLOW_PRIVATE_NETWORKS=true pnpm --silent m-trace check "$cmaf_hls_url")"
if ! echo "$cmaf_hls_out" | jq -e '.status == "ok" and .playlistType == "media" and .details.cmaf.binary.status == "passed"' >/dev/null; then
  echo "[smoke-cli] CMAF-HLS probe did not return binary.status=passed:"
  echo "$cmaf_hls_out"
  exit 1
fi
echo "[smoke-cli] CMAF-HLS probe OK (binary.status=passed)"

# 8b. CMAF-DASH-Probe.
cmaf_dash_url="http://127.0.0.1:${http_port}/manifest.mpd"
cmaf_dash_out="$(MTRACE_CHECK_ALLOW_PRIVATE_NETWORKS=true pnpm --silent m-trace check "$cmaf_dash_url")"
if ! echo "$cmaf_dash_out" | jq -e '.status == "ok" and .analyzerKind == "dash" and .details.cmaf.binary.status == "passed"' >/dev/null; then
  echo "[smoke-cli] CMAF-DASH probe did not return binary.status=passed:"
  echo "$cmaf_dash_out"
  exit 1
fi
echo "[smoke-cli] CMAF-DASH probe OK (binary.status=passed)"

# 8c. Doppel-check: ohne MTRACE_CHECK_ALLOW_PRIVATE_NETWORKS muss der
# Loopback-Aufruf weiterhin fetch_blocked liefern (der vorhandene
# URL-SSRF-Smoke aus Schritt 7 schützt RFC1918; hier zusätzlich
# gegen 127.0.0.1).
set +e
loopback_out="$(pnpm --silent m-trace check "$cmaf_hls_url" 2>/dev/null)"
loopback_exit=$?
set -e
if [ "$loopback_exit" != "1" ]; then
  echo "[smoke-cli] loopback without opt-in expected exit 1, got $loopback_exit"
  echo "$loopback_out"
  exit 1
fi
if ! echo "$loopback_out" | jq -e '.status == "error" and .code == "fetch_blocked"' >/dev/null; then
  echo "[smoke-cli] loopback without opt-in did not return fetch_blocked:"
  echo "$loopback_out"
  exit 1
fi
echo "[smoke-cli] loopback without opt-in still blocked (fetch_blocked)"

echo "[smoke-cli] all checks passed"
