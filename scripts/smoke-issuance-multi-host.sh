#!/usr/bin/env bash
# smoke-issuance-multi-host.sh — Reproduzierbarer Lab-Smoke für den
# Redis-Network-Backend-Limiter aus `0.12.6` Tranche 7 (RAK-88 / R-17).
#
# Verifiziert das semantische Multi-Host-Verhalten aus
# `docs/user/auth.md` §5.4:
#   1. miniredis startet einen In-Process-Redis-Mock; zwei
#      `RedisIssuanceRateLimiter`-Instanzen (analog zwei API-
#      Replicas auf verschiedenen Hosts) verbinden sich dagegen.
#   2. Replica A verbraucht das Project-Bucket bis zur Kapazität.
#   3. Replica B muss den nächsten Allow als „denied" sehen —
#      Shared-State greift über das Redis-`HSET`/`EXPIRE`-Tupel,
#      atomar via Lua-EVAL.
#   4. Refund-Pfad bei Project-Deny verbraucht keinen globalen
#      Token (Lua-Script-internes Refund).
#   5. Fail-Mode (closed/open) bei simuliertem Redis-Outage.
#
# Im Gegensatz zu `smoke-issuance-replica` (SQLite Single-Host) deckt
# dieser Smoke das **Multi-Host**-Setup ab — Redis ist der einzige
# Network-Backend-Pfad in `0.12.6`. Eine echte
# Compose-/K8s-Variante mit Redis-Server bleibt Folge-Item.
#
# Implementation: ruft die End-to-End-Unit-Tests
# `TestRedisIssuance_*` und `TestRedisOrigin_*` über das
# `golang:1.26.3`-Docker-Image auf.

set -euo pipefail

GO_IMAGE="${GO_IMAGE:-golang:1.26.3}"
REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
API_DIR="${REPO_ROOT}/apps/api"

if [[ ! -d "${API_DIR}" ]]; then
  echo "[smoke-issuance-multi-host] apps/api nicht gefunden — bitte aus dem Repo-Root aufrufen." >&2
  exit 2
fi

echo "▶ smoke-issuance-multi-host: zwei Adapter-Instances teilen sich Redis-Buckets."
echo "  Driver: go test -run 'TestRedisIssuance|TestRedisOrigin' (RAK-88, R-17 / R-22)"

if ! docker run --rm \
  -v "${API_DIR}:/src" \
  -w /src \
  -e CGO_ENABLED=0 \
  "${GO_IMAGE}" \
  go test -run "TestRedisIssuance|TestRedisOrigin" -v -count=1 ./adapters/driven/auth/...
then
  echo "[smoke-issuance-multi-host] FAIL — Redis-Multi-Host-Test ist rot." >&2
  exit 1
fi

echo "✔ smoke-issuance-multi-host ok (Redis-Network-Backend RAK-88 grün; R-17 + R-22 strukturell aufgelöst)."
