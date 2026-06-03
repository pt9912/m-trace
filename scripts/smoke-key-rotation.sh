#!/usr/bin/env bash
# smoke-key-rotation.sh — Reproduzierbarer Lab-Smoke für die
# Multi-Key-Signing-Rotation aus `0.12.5` (RAK-78).
#
# Verifiziert das Operator-Workflow-Verhalten aus `auth.md` §5.3.1:
# 1. Token unter `kid=A` signieren (ACTIVE=kid_a, Key-Ring kid_a+kid_b).
# 2. ACTIVE auf `kid=B` umschalten (zweiter Resolver-Bau, gleiches
# ENV-Schema).
# 3. Altes Token muss weiterhin verifizieren (kid_a bleibt im
# Verify-Set).
# 4. Neue Tokens werden mit kid_b signiert.
#
# Implementation: ruft den End-to-End-Unit-Test
# `TestParseSigningKeysEnv_RotationEndToEnd` in
# `apps/api/adapters/driven/auth/` über das golang:1.26-Docker-Image
# auf. Der Test deckt den semantischen Rotation-Kern ab (ENV-Parser
# → Resolver → Signer → Verify nach Rotation). Eine echte
# Compose-/API-Restart-Variante ist Folge-Item, sobald ein Multi-
# Replica-Compose-Smoke gebraucht wird (siehe `R-17` Folge-Scope).
#
# Konvention (siehe Geschwister-Smokes):
# - eigener Docker-Run, keine globalen Volumes
# - opt-in (nicht in `make gates`)
# - exit 0 bei grünem Test, exit 1 sonst

set -euo pipefail

GO_IMAGE="${GO_IMAGE:-golang:1.26.3}"
REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
API_DIR="${REPO_ROOT}/apps/api"

if [[ ! -d "${API_DIR}" ]]; then
  echo "[smoke-key-rotation] apps/api nicht gefunden — bitte aus dem Repo-Root aufrufen." >&2
  exit 2
fi

echo "▶ smoke-key-rotation: signiere Token unter kid_a, rotiere ACTIVE auf kid_b, verifiziere."
echo "  Driver: go test -run TestParseSigningKeysEnv_RotationEndToEnd (RAK-78)"

if ! docker run --rm \
  -v "${API_DIR}:/src" \
  -w /src \
  -e CGO_ENABLED=0 \
  "${GO_IMAGE}" \
  go test -run TestParseSigningKeysEnv_RotationEndToEnd -v -count=1 ./adapters/driven/auth/...
then
  echo "[smoke-key-rotation] FAIL — Rotation-End-to-End-Test ist rot." >&2
  exit 1
fi

echo "✔ smoke-key-rotation ok (Multi-Key-Rotation-Code-Pfad RAK-78 grün)."
