#!/usr/bin/env bash
# smoke-issuance-replica.sh — Reproduzierbarer Lab-Smoke für den
# Shared-State-Issuance-Limiter aus `0.12.5` Tranche 2 (RAK-77, R-17).
#
# Verifiziert das semantische Multi-Replica-Verhalten aus
# `docs/user/auth.md` §5.4:
#   1. Zwei `*sql.DB`-Verbindungen auf dieselbe SQLite-Datei
#      (Single-Host + Shared-Volume-Pfad) öffnen.
#   2. Replica A verbraucht das Project-Bucket bis zur Kapazität.
#   3. Replica B muss den nächsten Allow als „denied" sehen —
#      Shared-State greift über den `auth_issuance_counters`-Tisch
#      (Migration V5, BEGIN IMMEDIATE-serialisiert).
#   4. Andere Projects bleiben unabhängig (Bucket-Key-Isolation).
#
# Implementation: ruft den End-to-End-Unit-Test
# `TestSqliteIssuanceRateLimiter_SharedAcrossInstances` über das
# `golang:1.26.3`-Docker-Image auf. Der Test deckt den semantischen
# Sharing-Kern ab (zwei Adapter-Instances → dieselbe DB → geteiltes
# Bucket). Eine echte Compose-/K8s-Multi-Container-Variante mit zwei
# laufenden API-Prozessen ist Folge-Item, sobald ein Multi-Replica-
# Compose-Smoke gebraucht wird — der Limiter-Kern ist hiermit aber
# vollständig abgedeckt.
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
  echo "[smoke-issuance-replica] apps/api nicht gefunden — bitte aus dem Repo-Root aufrufen." >&2
  exit 2
fi

echo "▶ smoke-issuance-replica: Replica A verbraucht Project-Bucket, Replica B muss denied sehen."
echo "  Driver: go test -run TestSqliteIssuanceRateLimiter_SharedAcrossInstances (RAK-77, R-17)"

if ! docker run --rm \
  -v "${API_DIR}:/src" \
  -w /src \
  -e CGO_ENABLED=0 \
  "${GO_IMAGE}" \
  go test -run TestSqliteIssuanceRateLimiter_SharedAcrossInstances -v -count=1 ./adapters/driven/auth/...
then
  echo "[smoke-issuance-replica] FAIL — Shared-State-Test ist rot." >&2
  exit 1
fi

echo "✔ smoke-issuance-replica ok (Shared-State-Limiter RAK-77 grün; R-17 teilweise gelöst)."
