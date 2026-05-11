#!/usr/bin/env bash
# smoke-vault-approle.sh — Reproduzierbarer Lab-Smoke für den
# Vault-AppRole-Auth-Pfad aus `0.12.6` Tranche 8 (RAK-89 / R-20).
#
# Verifiziert das Adapter-Verhalten:
#   1. AppRole-Login an `/v1/auth/approle/login` mit role_id+secret_id
#      liefert einen Vault `client_token` zurück.
#   2. Adapter nutzt den Token für `GET /v1/<mount>/data/<path>` und
#      parsed `keys`+`active_kid` per gemeinsamem
#      `ParseSigningKeysEnv`-Pfad.
#   3. Fail-Modi (wrong secret → login 401; missing approle env →
#      constructor reject) sind explizit getestet.
#
# Implementation: wrapt `TestVault_AppRoleLogin_*` über das
# `golang:1.26.3`-Docker-Image. Kein echter Vault-Server nötig —
# `httptest.Server` simuliert AppRole-Login + KV-Read. Eine
# vault-dev-CLI-Variante mit echtem Server bleibt Folge-Item.

set -euo pipefail

GO_IMAGE="${GO_IMAGE:-golang:1.26.3}"
REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
API_DIR="${REPO_ROOT}/apps/api"

if [[ ! -d "${API_DIR}" ]]; then
  echo "[smoke-vault-approle] apps/api nicht gefunden — bitte aus dem Repo-Root aufrufen." >&2
  exit 2
fi

echo "▶ smoke-vault-approle: AppRole-Login + KV-Read via httptest-Mock."
echo "  Driver: go test -run 'TestVault_AppRole|TestVault_Kubernetes|TestVault_UnsupportedAuthMethod' (RAK-89, R-20)"

if ! docker run --rm \
  -v "${API_DIR}:/src" \
  -w /src \
  -e CGO_ENABLED=0 \
  "${GO_IMAGE}" \
  go test -run "TestVault_AppRole|TestVault_Kubernetes|TestVault_UnsupportedAuthMethod" -v -count=1 ./adapters/driven/auth/...
then
  echo "[smoke-vault-approle] FAIL — Vault-AppRole/K8s-Tests sind rot." >&2
  exit 1
fi

echo "✔ smoke-vault-approle ok (AppRole + Kubernetes auth-methods RAK-89 grün; R-20 in 0.12.6 strukturell adressiert)."
