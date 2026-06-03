#!/usr/bin/env bash
# smoke-outbound-webhook.sh — Reproduzierbarer Lab-Smoke für den
# Outbound-Webhook-Dispatcher aus `0.12.5` (RAK-82, R-16).
#
# Verifiziert den HTTP-Dispatcher gegen einen `httptest.Server`-
# Mock-Konsumenten:
# 1. Endpoint leer → Dispatch no-op.
# 2. 200 first try → happy path (1 Call).
# 3. HMAC-Signatur stimmt mit `X-MTrace-Signature: sha256=<hex>`.
# 4. 503 → 200 transient → Retry succeeds (2 Calls).
# 5. 3×500 → ErrOutboundWebhookExhausted (Dead-Letter, 3 Calls).
# 6. Body enthält die Pflichtfelder und keinen Klartext-Stream-Key.
# 7. ctx-Cancel stoppt den Retry-Loop.
#
# Implementation: ruft die End-to-End-Tests `TestOutboundWebhook_*`
# aus `apps/api/adapters/driven/webhooks/http_dispatcher_test.go`
# über das `golang:1.26.3`-Docker-Image auf.
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
  echo "[smoke-outbound-webhook] apps/api nicht gefunden — bitte aus dem Repo-Root aufrufen." >&2
  exit 2
fi

echo "▶ smoke-outbound-webhook: Outbound-Webhook-Dispatcher (RAK-82, R-16)"
echo "  Driver: go test -run TestOutboundWebhook -v ./adapters/driven/webhooks/..."

if ! docker run --rm \
  -v "${API_DIR}:/src" \
  -w /src \
  -e CGO_ENABLED=0 \
  "${GO_IMAGE}" \
  go test -run TestOutboundWebhook -v -count=1 ./adapters/driven/webhooks/...
then
  echo "[smoke-outbound-webhook] FAIL — Outbound-Webhook-Tests sind rot." >&2
  exit 1
fi

echo "✔ smoke-outbound-webhook ok (Outbound-Webhook-Dispatcher RAK-82 grün; R-16 aufgelöst)."
