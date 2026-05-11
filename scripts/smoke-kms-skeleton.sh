#!/usr/bin/env bash
# smoke-kms-skeleton.sh — Reproduzierbarer Lab-Smoke für den
# KMS-Skelett-Adapter aus `0.12.6` Tranche 8 (RAK-89 / R-20).
#
# Verifiziert das Adapter-Verhalten:
#   1. Stub-Decrypter liefert einen `keys`-String → Adapter parsed
#      ihn über die gemeinsame `ParseSigningKeysEnv`-Validation.
#   2. Decrypter-Fehler propagiert fail-closed.
#   3. Constructor lehnt fehlende Pflicht-ENV ab (Active-KID,
#      Ciphertext).
#   4. Ciphertext-Quelle: ENV-Base64 ODER File-Path.
#   5. `LabPassThroughKMSDecrypter` (Lab-Mock) reicht Ciphertext als
#      Plaintext durch — der `make smoke-kms-skeleton`-Smoke nutzt
#      diesen Pfad, um den End-to-End-Flow ohne AWS-KMS-Konto zu
#      testen.
#
# Produktive AWS-SDK-v2-Anbindung ist Folge-Item — der Adapter ist
# heute über das `KMSDecrypter`-Interface vorbereitet. Operatoren
# mit KMS in Production injizieren einen eigenen `KMSDecrypter` über
# Boot-Time-Wiring.

set -euo pipefail

GO_IMAGE="${GO_IMAGE:-golang:1.26.3}"
REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
API_DIR="${REPO_ROOT}/apps/api"

if [[ ! -d "${API_DIR}" ]]; then
  echo "[smoke-kms-skeleton] apps/api nicht gefunden — bitte aus dem Repo-Root aufrufen." >&2
  exit 2
fi

echo "▶ smoke-kms-skeleton: KMS-Adapter-Pfad mit Stub-Decrypter + LabPassThrough."
echo "  Driver: go test -run 'TestKMSSecretBackend_' (RAK-89, R-20)"

if ! docker run --rm \
  -v "${API_DIR}:/src" \
  -w /src \
  -e CGO_ENABLED=0 \
  "${GO_IMAGE}" \
  go test -run "TestKMSSecretBackend_" -v -count=1 ./adapters/driven/auth/...
then
  echo "[smoke-kms-skeleton] FAIL — KMS-Skelett-Tests sind rot." >&2
  exit 1
fi

echo "✔ smoke-kms-skeleton ok (KMS-Adapter-Skelett RAK-89 grün; production AWS-SDK-wiring ist Folge-Item)."
