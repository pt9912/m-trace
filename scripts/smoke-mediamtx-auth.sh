#!/usr/bin/env bash
# smoke-mediamtx-auth.sh — Reproduzierbarer Lab-Smoke für die
# MediaMTX-Auth-Bridge aus `0.12.5` (RAK-81, R-14).
#
# Verifiziert das Wire-Verhalten der `POST /api/ingest/auth-hook`-
# Route aus `docs/user/auth.md` §5.7:
# 1. Gültiger Stream-Key + `action=publish` → 200 allow.
# 2. Falscher Stream-Key → 403 deny.
# 3. `action=read` → 403 deny (Read-Auth bleibt Folge-Item).
# 4. Fehlendes `user`/`path`/`password` → 403 deny.
# 5. Falsches Content-Type (z. B. JSON) → 400 bad request.
# 6. GET-Request → 405 method not allowed.
# 7. Backend-Fehler (Repo-Outage) → 403 fail-closed.
#
# Implementation: ruft die End-to-End-Tests
# `TestMediaMTXAuthHook_*` aus
# `apps/api/adapters/driving/http/mediamtx_auth_hook_test.go` über
# das `golang:1.26.3`-Docker-Image auf. Echte Compose-Variante
# (MediaMTX-Container mit externalAuth gegen laufende m-trace-API)
# bleibt Folge-Item — der Wire-Vertrag ist hiermit aber
# vollständig abgedeckt.
#
# Konvention:
# - eigener Docker-Run, keine globalen Volumes
# - opt-in (nicht in `make gates`)
# - exit 0 bei grünem Test, exit 1 sonst

set -euo pipefail

GO_IMAGE="${GO_IMAGE:-golang:1.26.3}"
REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
API_DIR="${REPO_ROOT}/apps/api"

if [[ ! -d "${API_DIR}" ]]; then
  echo "[smoke-mediamtx-auth] apps/api nicht gefunden — bitte aus dem Repo-Root aufrufen." >&2
  exit 2
fi

echo "▶ smoke-mediamtx-auth: MediaMTX-externalAuth Bridge (RAK-81, R-14)"
echo "  Driver: go test -run TestMediaMTXAuthHook -v ./adapters/driving/http/..."

if ! docker run --rm \
  -v "${API_DIR}:/src" \
  -w /src \
  -e CGO_ENABLED=0 \
  "${GO_IMAGE}" \
  go test -run TestMediaMTXAuthHook -v -count=1 ./adapters/driving/http/...
then
  echo "[smoke-mediamtx-auth] FAIL — MediaMTX-Auth-Bridge-Tests sind rot." >&2
  exit 1
fi

echo "✔ smoke-mediamtx-auth ok (MediaMTX-Auth-Bridge RAK-81 grün; R-14 aufgelöst)."
