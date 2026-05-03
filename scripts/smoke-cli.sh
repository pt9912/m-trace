#!/usr/bin/env bash
set -euo pipefail

# Smoke für plan-0.3.0 Tranche 7:
#  1. `pnpm --silent m-trace --help` zeigt die Usage und exit 0.
#  2. `pnpm --silent m-trace check <file>` analysiert eine Master-Manifest-
#     Fixture, gibt JSON auf stdout aus, exit 0.
#  3. `pnpm --silent m-trace check <not-hls>` bricht ab mit exit 1 und JSON,
#     das `status:"error"` mit `code:"manifest_not_hls"` trägt.
#  4. `pnpm --silent m-trace check /nonexistent.m3u8` läuft auf IO-Fehler,
#     exit 1, Fehler-Message auf stderr.
#  5. Aufruf ohne Argumente → Usage-Fehler, exit 2.
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

# Non-HLS-Fixture
nothls="$tmpdir/not-hls.txt"
echo "<html><body>not a manifest</body></html>" > "$nothls"

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

# 3. non-HLS → analysis-error → exit 1
set +e
nothls_out="$(pnpm --silent m-trace check "$nothls" 2>/dev/null)"
nothls_exit=$?
set -e
if [ "$nothls_exit" != "1" ]; then
  echo "[smoke-cli] non-HLS expected exit 1, got $nothls_exit"
  exit 1
fi
if ! echo "$nothls_out" | jq -e '.status == "error" and .code == "manifest_not_hls"' >/dev/null; then
  echo "[smoke-cli] non-HLS did not return manifest_not_hls:"
  echo "$nothls_out"
  exit 1
fi
echo "[smoke-cli] check non-HLS fixture OK (exit 1, manifest_not_hls)"

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
