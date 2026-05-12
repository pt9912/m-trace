#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage: scripts/release-guard.sh X.Y.Z

Required approval:
  MTRACE_RELEASE_APPROVED=1

Optional test overrides:
  MTRACE_RELEASE_ALLOW_NON_MAIN=1
  MTRACE_RELEASE_ALLOW_DIRTY=1
  MTRACE_RELEASE_ALLOW_OFFLINE=1
  MTRACE_RELEASE_DRY_RUN=1
USAGE
}

fail() {
  echo "release-guard: $*" >&2
  exit 1
}

if [ "${1:-}" = "-h" ] || [ "${1:-}" = "--help" ]; then
  usage
  exit 0
fi

version="${1:-}"
[ -n "$version" ] || {
  usage >&2
  exit 2
}

case "$version" in
  v*) fail "pass the version without v-prefix, got $version" ;;
esac

if ! [[ "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  fail "version must look like X.Y.Z, got $version"
fi

tag="v$version"

[ "${MTRACE_RELEASE_APPROVED:-}" = "1" ] || \
  fail "manual approval missing: set MTRACE_RELEASE_APPROVED=1"

branch="$(git rev-parse --abbrev-ref HEAD)"
if [ "$branch" != "main" ] && [ "${MTRACE_RELEASE_ALLOW_NON_MAIN:-}" != "1" ]; then
  fail "release must run on main (current: $branch)"
fi

if [ -n "$(git status --porcelain)" ] && [ "${MTRACE_RELEASE_ALLOW_DIRTY:-}" != "1" ]; then
  fail "working tree must be clean"
fi

if git rev-parse -q --verify "refs/tags/$tag" >/dev/null; then
  fail "tag already exists locally: $tag"
fi

if git ls-remote --exit-code --tags origin "refs/tags/$tag" >/dev/null 2>&1; then
  fail "tag already exists on origin: $tag"
else
  ls_remote_status=$?
  if [ "$ls_remote_status" -ne 2 ] && [ "${MTRACE_RELEASE_ALLOW_OFFLINE:-}" != "1" ]; then
    fail "could not verify tag on origin (set MTRACE_RELEASE_ALLOW_OFFLINE=1 only for local guard tests)"
  fi
fi

if ! grep -q "## \\[$version\\]" CHANGELOG.md; then
  fail "CHANGELOG.md has no section for [$version]"
fi

if ! grep -q "serviceVersion    = \"$version\"" apps/api/cmd/api/main.go; then
  fail "apps/api serviceVersion is not $version"
fi

if ! grep -q "\"version\": \"$version\"" package.json; then
  fail "root package.json version is not $version"
fi

if [ "${MTRACE_RELEASE_DRY_RUN:-}" = "1" ]; then
  echo "release-guard: dry-run ok for $tag"
else
  echo "release-guard: ok for $tag"
fi
