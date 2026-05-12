#!/usr/bin/env bash
set -euo pipefail

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
k8s_dir="${1:-$root_dir/deploy/k8s}"

fail() {
  echo "k8s-validate: $*" >&2
  exit 1
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || fail "missing required command: $1"
}

require_cmd yq
require_cmd node

[ -d "$k8s_dir" ] || fail "directory not found: $k8s_dir"

expected_version="$(cd "$root_dir" && node -p "require('./package.json').version")"
expected_namespace="m-trace-example"
expected_part="m-trace"

files=(
  "$k8s_dir/namespace.yaml"
  "$k8s_dir/analyzer-service.yaml"
  "$k8s_dir/api.yaml"
  "$k8s_dir/dashboard.yaml"
)

for file in "${files[@]}"; do
  [ -f "$file" ] || fail "missing manifest: ${file#$root_dir/}"
  yq . "$file" >/dev/null || fail "invalid YAML: ${file#$root_dir/}"
done

while IFS= read -r kind; do
  case "$kind" in
    Namespace|PersistentVolumeClaim|Deployment|Service) ;;
    "") ;;
    *) fail "unsupported manifest kind: $kind" ;;
  esac
done < <(yq -r '.kind // ""' "${files[@]}")

while IFS= read -r namespace; do
  [ "$namespace" = "$expected_namespace" ] || \
    fail "non-Namespace object must use namespace $expected_namespace, got ${namespace:-<empty>}"
done < <(yq -r 'select(.kind != "Namespace") | .metadata.namespace // ""' "${files[@]}")

while IFS= read -r part; do
  [ "$part" = "$expected_part" ] || \
    fail "app.kubernetes.io/part-of must be $expected_part, got ${part:-<empty>}"
done < <(yq -r '.. | objects | select(has("labels")) | .labels."app.kubernetes.io/part-of" // ""' "${files[@]}")

while IFS= read -r label_key; do
  case "$label_key" in
    pod|namespace|container)
      fail "example manifests must not add K8s infrastructure label '$label_key'; keep R-9 allowlists profile-specific"
      ;;
  esac
done < <(yq -r '.. | objects | select(has("labels")) | .labels | keys[]' "${files[@]}")

while IFS= read -r replicas; do
  [ "$replicas" = "1" ] || fail "example deployments must stay single-replica, got $replicas"
done < <(yq -r 'select(.kind == "Deployment") | .spec.replicas // ""' "${files[@]}")

while IFS= read -r image; do
  case "$image" in
    *":$expected_version") ;;
    *) fail "image tag must match package version $expected_version, got $image" ;;
  esac
done < <(yq -r 'select(.kind == "Deployment") | .spec.template.spec.containers[].image' "${files[@]}")

if grep -RiqE 'kind:[[:space:]]*(Ingress|HorizontalPodAutoscaler|NetworkPolicy|PodDisruptionBudget)' "$k8s_dir"; then
  fail "production-oriented K8s kinds are out of scope for these example manifests"
fi

grep -qi 'production-ready' "$k8s_dir/README.md" || \
  fail "README must keep the not-production-ready boundary"
grep -q 'R-9' "$k8s_dir/README.md" || \
  fail "README must document the R-9 observability boundary"

echo "k8s-validate: ok for $expected_namespace at image tag $expected_version"
