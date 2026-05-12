#!/usr/bin/env bash
set -euo pipefail

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT

fail() {
  echo "release-guard-test: $*" >&2
  exit 1
}

setup_repo() {
  local name="$1"
  local dir="$tmp_dir/$name"
  mkdir -p "$dir/scripts" "$dir/apps/api/cmd/api"
  cp "$root_dir/scripts/release-guard.sh" "$dir/scripts/release-guard.sh"
  cat >"$dir/CHANGELOG.md" <<'EOF'
# Changelog

## [9.9.9] - 2099-01-01
EOF
  cat >"$dir/package.json" <<'EOF'
{
  "name": "release-guard-fixture",
  "version": "9.9.9"
}
EOF
  cat >"$dir/apps/api/cmd/api/main.go" <<'EOF'
package main

const (
	serviceVersion    = "9.9.9"
)
EOF
  (
    cd "$dir"
    git init -q
    git branch -M main
    git add .
    git -c user.name="Release Guard Test" \
      -c user.email="release-guard-test@example.invalid" \
      commit -q -m init
    git remote add origin https://example.invalid/m-trace.git
  )
  printf '%s\n' "$dir"
}

expect_success() {
  local name="$1"
  shift
  local output
  if ! output="$("$@" 2>&1)"; then
    printf '%s\n' "$output" >&2
    fail "$name: expected success"
  fi
}

expect_failure() {
  local name="$1"
  local expected="$2"
  shift 2
  local output
  if output="$("$@" 2>&1)"; then
    printf '%s\n' "$output" >&2
    fail "$name: expected failure"
  fi
  case "$output" in
    *"$expected"*) ;;
    *)
      printf '%s\n' "$output" >&2
      fail "$name: expected output to contain '$expected'"
      ;;
  esac
}

run_guard() {
  local repo="$1"
  shift
  (
    cd "$repo"
    "$@"
  )
}

repo="$(setup_repo success)"
expect_success "approved dry-run" \
  run_guard "$repo" \
  env MTRACE_RELEASE_APPROVED=1 MTRACE_RELEASE_ALLOW_OFFLINE=1 MTRACE_RELEASE_DRY_RUN=1 \
  bash scripts/release-guard.sh 9.9.9

repo="$(setup_repo missing-approval)"
expect_failure "missing approval" "manual approval missing" \
  run_guard "$repo" \
  bash scripts/release-guard.sh 9.9.9

repo="$(setup_repo v-prefix)"
expect_failure "v-prefix" "without v-prefix" \
  run_guard "$repo" \
  env MTRACE_RELEASE_APPROVED=1 MTRACE_RELEASE_ALLOW_OFFLINE=1 \
  bash scripts/release-guard.sh v9.9.9

repo="$(setup_repo non-main)"
(
  cd "$repo"
  git switch -q -c feature/release-guard-test
)
expect_failure "non-main" "release must run on main" \
  run_guard "$repo" \
  env MTRACE_RELEASE_APPROVED=1 MTRACE_RELEASE_ALLOW_OFFLINE=1 \
  bash scripts/release-guard.sh 9.9.9
expect_success "non-main override" \
  run_guard "$repo" \
  env MTRACE_RELEASE_APPROVED=1 MTRACE_RELEASE_ALLOW_NON_MAIN=1 MTRACE_RELEASE_ALLOW_OFFLINE=1 \
  bash scripts/release-guard.sh 9.9.9

repo="$(setup_repo dirty)"
printf '\n' >>"$repo/CHANGELOG.md"
expect_failure "dirty tree" "working tree must be clean" \
  run_guard "$repo" \
  env MTRACE_RELEASE_APPROVED=1 MTRACE_RELEASE_ALLOW_OFFLINE=1 \
  bash scripts/release-guard.sh 9.9.9

repo="$(setup_repo local-tag)"
(
  cd "$repo"
  git tag v9.9.9
)
expect_failure "local tag" "tag already exists locally" \
  run_guard "$repo" \
  env MTRACE_RELEASE_APPROVED=1 MTRACE_RELEASE_ALLOW_OFFLINE=1 \
  bash scripts/release-guard.sh 9.9.9

repo="$(setup_repo package-version)"
cat >"$repo/package.json" <<'EOF'
{
  "name": "release-guard-fixture",
  "version": "9.9.8"
}
EOF
(
  cd "$repo"
  git add package.json
  git -c user.name="Release Guard Test" \
    -c user.email="release-guard-test@example.invalid" \
    commit -q -m "break package version"
)
expect_failure "package version mismatch" "root package.json version is not 9.9.9" \
  run_guard "$repo" \
  env MTRACE_RELEASE_APPROVED=1 MTRACE_RELEASE_ALLOW_OFFLINE=1 \
  bash scripts/release-guard.sh 9.9.9

echo "release-guard-test: ok"
