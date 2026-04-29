#!/usr/bin/env bash
# Architecture-Boundary-Test für apps/api.
#
# Stellt sicher, dass die Hexagon-Pakete keine direkten Imports auf
# Adapter, Infrastruktur-Bibliotheken oder über die Hexagon-Schichten-
# grenzen hinweg haben. Bezug: docs/architecture.md §3.2, §3.4.
#
# Nutzung:
#   cd apps/api
#   bash scripts/check-architecture.sh
#
# Exit 0 = ok, Exit 1 = mindestens ein Verstoß gefunden.
set -euo pipefail

MODULE_PATH="$(go list -m)"
FAILED=0

check_no_direct_imports() {
  local package_pattern="$1"
  local forbidden_regex="$2"
  local reason="$3"

  while IFS= read -r pkg; do
    imports="$(go list -f '{{join .Imports "\n"}}' "$pkg")"

    matches="$(echo "$imports" | grep -E "$forbidden_regex" || true)"

    if [[ -n "$matches" ]]; then
      echo "Architecture violation in package:"
      echo "  $pkg"
      echo
      echo "Reason:"
      echo "  $reason"
      echo
      echo "Forbidden direct imports:"
      echo "$matches" | sed 's/^/  - /'
      echo
      FAILED=1
    fi
  done < <(go list "$package_pattern")
}

check_no_direct_imports \
  "./hexagon/..." \
  "^${MODULE_PATH}/adapters|^go.opentelemetry.io|^github.com/prometheus|^database/sql|^net/http" \
  "hexagon must not directly import adapters or infrastructure libraries"

check_no_direct_imports \
  "./hexagon/domain/..." \
  "^${MODULE_PATH}/hexagon/application|^${MODULE_PATH}/hexagon/port" \
  "domain must not directly import application or ports"

check_no_direct_imports \
  "./hexagon/application/..." \
  "^${MODULE_PATH}/adapters" \
  "application must depend on ports, not adapter implementations"

check_no_direct_imports \
  "./hexagon/port/..." \
  "^${MODULE_PATH}/adapters" \
  "ports must define abstractions, not depend on adapter implementations"

if [[ "$FAILED" -ne 0 ]]; then
  exit 1
fi

echo "Architecture check passed"
