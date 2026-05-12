#!/usr/bin/env bash
set -euo pipefail

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
devcontainer_file="$root_dir/.devcontainer/devcontainer.json"

fail() {
  echo "devcontainer-validate: $*" >&2
  exit 1
}

command -v node >/dev/null 2>&1 || fail "missing required command: node"
[ -f "$devcontainer_file" ] || fail "missing .devcontainer/devcontainer.json"

node - "$devcontainer_file" <<'NODE'
const fs = require("fs");
const path = process.argv[2];
const config = JSON.parse(fs.readFileSync(path, "utf8"));

function fail(message) {
  console.error(`devcontainer-validate: ${message}`);
  process.exit(1);
}

const features = config.features || {};
const dockerFeature = "ghcr.io/devcontainers/features/docker-outside-of-docker:1";
const goFeature = "ghcr.io/devcontainers/features/go:1";
const nodeFeature = "ghcr.io/devcontainers/features/node:1";

if (!features[dockerFeature]) {
  fail("docker-outside-of-docker feature is required");
}
if (features[goFeature]?.version !== "1.26.3") {
  fail(`Go feature must pin 1.26.3, got ${features[goFeature]?.version || "<empty>"}`);
}
if (features[nodeFeature]?.version !== "22") {
  fail(`Node feature must pin major 22, got ${features[nodeFeature]?.version || "<empty>"}`);
}
if (config.remoteUser !== "vscode") {
  fail(`remoteUser must remain vscode, got ${config.remoteUser || "<empty>"}`);
}

const postCreate = String(config.postCreateCommand || "");
if (!postCreate.includes("corepack enable") || !postCreate.includes("pnpm@10.18.0")) {
  fail("postCreateCommand must only prepare pinned pnpm via corepack");
}
if (/\b(pnpm|npm|yarn)\s+(install|add|update)\b/.test(postCreate)) {
  fail("postCreateCommand must not install workspace dependencies implicitly");
}

console.log("devcontainer-validate: ok");
NODE
