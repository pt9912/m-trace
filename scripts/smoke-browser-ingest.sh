#!/usr/bin/env bash
# smoke-browser-ingest.sh — Reproduzierbarer Lab-Smoke für die
# Browser-Ingest-Policy aus `0.12.5` Tranche 4 (RAK-80, R-21).
#
# Verifiziert das Wire-Verhalten aus `docs/user/auth.md` §5.6:
#   1. Preflight ohne aktivierte Policy → 204 ohne Allow-Origin
#      (RAK-74-Scope-Cut bleibt strikt).
#   2. Preflight mit aktivierter Policy + Origin in Allowlist
#      → 204 + Allow-Origin/Methods/Headers.
#   3. Preflight mit aktivierter Policy + Origin NICHT in Allowlist
#      → 204 ohne Allow-Origin (kein Enumerations-Leak).
#   4. POST mit Origin-Pin-Mismatch → 403 `ingest_browser_origin_pin_mismatch`.
#   5. POST mit fehlendem CSRF-Header → 403 `ingest_browser_csrf_missing`.
#   6. POST mit allen Checks bestanden → 201 (Stream-Create-Pfad).
#
# Implementation: ruft die End-to-End-Browser-Ingest-Tests aus
# `apps/api/adapters/driving/http/browser_ingest_test.go` über das
# `golang:1.26.3`-Docker-Image auf — derselbe Wire-Vertrag, den der
# Operator gegen die laufende API mit `curl` reproduzieren würde,
# nur als kontrollierter In-Process-`httptest.Server`-Pfad ohne
# zusätzliche Compose-Voraussetzungen.
#
# Konvention:
#   - eigener Docker-Run, keine globalen Volumes
#   - opt-in (nicht in `make gates`)
#   - exit 0 bei grünem Test, exit 1 sonst

set -euo pipefail

GO_IMAGE="${GO_IMAGE:-golang:1.26.3}"
REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
API_DIR="${REPO_ROOT}/apps/api"

if [[ ! -d "${API_DIR}" ]]; then
  echo "[smoke-browser-ingest] apps/api nicht gefunden — bitte aus dem Repo-Root aufrufen." >&2
  exit 2
fi

echo "▶ smoke-browser-ingest: Preflight + POST Browser-Ingest-Policy (RAK-80, R-21)"
echo "  Driver: go test -run TestBrowserIngest -v ./adapters/driving/http/..."

if ! docker run --rm \
  -v "${API_DIR}:/src" \
  -w /src \
  -e CGO_ENABLED=0 \
  "${GO_IMAGE}" \
  go test -run TestBrowserIngest -v -count=1 ./adapters/driving/http/...
then
  echo "[smoke-browser-ingest] FAIL — Browser-Ingest-Tests sind rot." >&2
  exit 1
fi

echo "✔ smoke-browser-ingest ok (Browser-Ingest-Policy RAK-80 grün; R-21 aufgelöst)."
