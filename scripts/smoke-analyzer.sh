#!/usr/bin/env bash
set -euo pipefail

# Smoke für plan-0.3.0 Tranche 6:
#  1. analyzer-service /health antwortet 200.
#  2. POST /api/analyze mit Text-Input liefert ein AnalysisResult
#     (status:ok, playlistType:master, analyzerKind:hls) — exercitiert
#     den Pfad API → analyzer-service → @npm9912/stream-analyzer.
#  3. POST /api/analyze mit URL gegen RFC1918-Adresse wird vom SSRF-
#     Schutz im analyzer-service abgelehnt; die API mappt das auf 502.
#
# Anmerkung zu URL-Inputs: docker-bridge-IPs liegen typischerweise in
# 172.16/12 — also exakt im SSRF-Sperrbereich. Für intra-Compose-
# Demos nutzen wir deshalb Text-Inputs; ein "public URL"-Test würde
# unkontrolliert öffentliche Internet-Erreichbarkeit verlangen.
#
# Erwartet: docker-compose-Stack ist hochgefahren ("make dev" oder
# "make smoke-analyzer" via Makefile). Manueller Aufruf möglich:
#   API_URL=http://localhost:8080 ANALYZER_URL=http://localhost:7000 \
#     scripts/smoke-analyzer.sh

API_URL="${API_URL:-http://localhost:8080}"
ANALYZER_URL="${ANALYZER_URL:-http://localhost:7000}"

if ! command -v jq >/dev/null 2>&1; then
  echo "[smoke-analyzer] missing dependency: jq" >&2
  echo "  install via your package manager (apt: 'apt-get install jq')." >&2
  exit 2
fi

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

wait_for_status() {
  url="$1"; expected="$2"; label="$3"
  for _ in $(seq 1 60); do
    status="$(curl -sSL -o "$tmpdir/$label.body" -w '%{http_code}' "$url" || true)"
    if [ "$status" = "$expected" ]; then
      echo "[smoke-analyzer] $label OK ($status)"
      return 0
    fi
    sleep 1
  done
  echo "[smoke-analyzer] $label FAIL — final status $status"
  cat "$tmpdir/$label.body" || true
  return 1
}

# 1. analyzer-service health
wait_for_status "$ANALYZER_URL/health" "200" "analyzer-health"

# 2. API health
wait_for_status "$API_URL/api/health" "200" "api-health"

# 3. POST /api/analyze mit Text-Input — Master Playlist, exerciert den
#    vollen Pfad API → analyzer-service → stream-analyzer.
master_manifest='#EXTM3U
#EXT-X-VERSION:6
#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="aud-en",NAME="English",DEFAULT=YES,URI="audio/en.m3u8"
#EXT-X-STREAM-INF:BANDWIDTH=1280000,RESOLUTION=1280x720,AUDIO="aud-en"
video/720p.m3u8
'
master_body="$(jq -n --arg t "$master_manifest" '{kind:"text", text:$t}')"
echo "[smoke-analyzer] POST /api/analyze (kind=text, master playlist)"
status="$(curl -sSL -o "$tmpdir/master.body" -w '%{http_code}' \
  -X POST "$API_URL/api/analyze" \
  -H 'Content-Type: application/json' \
  -d "$master_body")"
if [ "$status" != "200" ]; then
  echo "[smoke-analyzer] master case: expected 200, got $status"
  cat "$tmpdir/master.body"
  exit 1
fi
if ! grep -qE '"status":"ok"' "$tmpdir/master.body"; then
  echo "[smoke-analyzer] /api/analyze (master) did not return status:ok"
  cat "$tmpdir/master.body"
  exit 1
fi
if ! grep -qE '"playlistType":"master"' "$tmpdir/master.body"; then
  echo "[smoke-analyzer] /api/analyze (master) missing playlistType=master"
  cat "$tmpdir/master.body"
  exit 1
fi
if ! grep -qE '"analyzerKind":"hls"' "$tmpdir/master.body"; then
  echo "[smoke-analyzer] /api/analyze (master) missing analyzerKind=hls"
  cat "$tmpdir/master.body"
  exit 1
fi
echo "[smoke-analyzer] master case OK"

# 4. SSRF-Negativfall: Credentials in URL werden unabhängig vom
#    ALLOW_PRIVATE_NETWORKS-Flag geblockt. Im Compose-Lab steht das
#    Flag auf true (intern erreichbares mediamtx), deshalb ist
#    Credentials der robustere Negativtest — der greift in Lab und
#    Produktion. Der API-Adapter mappt den Domain-Code fetch_blocked
#    auf 400.
status="$(curl -sSL -o "$tmpdir/ssrf.body" -w '%{http_code}' \
  -X POST "$API_URL/api/analyze" \
  -H 'Content-Type: application/json' \
  -d '{"kind":"url","url":"http://user:pass@example.test/m.m3u8"}')"
if [ "$status" != "400" ]; then
  echo "[smoke-analyzer] SSRF case: expected 400 (fetch_blocked), got $status"
  cat "$tmpdir/ssrf.body"
  exit 1
fi
if ! grep -qE '"code":"fetch_blocked"' "$tmpdir/ssrf.body"; then
  echo "[smoke-analyzer] SSRF case: response missing fetch_blocked code"
  cat "$tmpdir/ssrf.body"
  exit 1
fi
echo "[smoke-analyzer] SSRF-block correctly mapped to 400 fetch_blocked"

echo "[smoke-analyzer] all checks passed"
