#!/usr/bin/env bash
# smoke-mediaserver-provision.sh — Reproduzierbarer Lab-Smoke fuer
# den externen Media-Server-Provisioner aus 
# (RAK-87 / R-15).
#
# Verifiziert den `MediaMTXProvisioner`-Adapter und den Use-Case-
# Pfad:
# 1. Apply 200 OK → State `applied` (happy path).
# 2. Apply 409 Conflict → State `applied` (idempotent).
# 3. Apply 401 Unauthorized → State `failed` mit ErrorCode
# `auth_failure`.
# 4. Apply 500 Internal → State `failed` mit ErrorCode
# `server_status_500`.
# 5. Apply unreachable → State `failed` mit ErrorCode `unreachable`.
# 6. Use-Case `Provision=false` laesst MediaServerState leer
# (byte-stabil zum 0.11.0-Format).
# 7. Use-Case `Provision=true` ohne Adapter → State `disabled` +
# Operator-Hint.
#
# Implementation: wrapt `TestMediaMTX_*` und
# `TestIngestControlService_CreateStream_Provision*` ueber das
# `golang:1.26.3`-Docker-Image. Kein echter MediaMTX-Server noetig
# (`httptest.Server`-Mock); eine produktive MediaMTX-Anbindung mit
# `examples/srt/compose.yaml` ist Folge-Item.

set -euo pipefail

GO_IMAGE="${GO_IMAGE:-golang:1.26.3}"
REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
API_DIR="${REPO_ROOT}/apps/api"

if [[ ! -d "${API_DIR}" ]]; then
  echo "[smoke-mediaserver-provision] apps/api nicht gefunden — bitte aus dem Repo-Root aufrufen." >&2
  exit 2
fi

echo "▶ smoke-mediaserver-provision: MediaMTX-Adapter + Use-Case-Pfad mit httptest-Mock."
echo "  Driver: go test -run 'TestMediaMTX_|TestIngestControlService_CreateStream_Provision' (RAK-87, R-15)"

if ! docker run --rm \
  -v "${API_DIR}:/src" \
  -w /src \
  -e CGO_ENABLED=0 \
  "${GO_IMAGE}" \
  go test -run "TestMediaMTX_|TestIngestControlService_CreateStream_Provision" -v -count=1 ./adapters/driven/mediaserver/... ./hexagon/application/...
then
  echo "[smoke-mediaserver-provision] FAIL — MediaMTX/Provision-Tests sind rot." >&2
  exit 1
fi

echo "✔ smoke-mediaserver-provision ok (MediaMTX-Adapter + Use-Case RAK-87 gruen; R-15 in 0.12.6 strukturell aufgeloest)."
