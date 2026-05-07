#!/usr/bin/env bash
set -euo pipefail

# plan-0.9.5 §2 Tranche 1 DoD-Item 7 — Bench-Runs drucken Runner-OS,
# CPU-Modell und Runtime-Versionen, damit Budget-Failures
# einordenbar bleiben.

echo "[bench-runner] === Runner Info ==="
echo "[bench-runner] Date           $(date -u +%Y-%m-%dT%H:%M:%SZ)"
if command -v uname >/dev/null 2>&1; then
  echo "[bench-runner] Kernel         $(uname -srm)"
fi
if [ -r /etc/os-release ]; then
  os_pretty="$(grep '^PRETTY_NAME=' /etc/os-release | cut -d= -f2- | tr -d '"')"
  if [ -n "$os_pretty" ]; then
    echo "[bench-runner] OS             $os_pretty"
  fi
fi
if [ -r /proc/cpuinfo ]; then
  cpu_model="$(grep -m 1 '^model name' /proc/cpuinfo | cut -d: -f2- | sed -e 's/^ *//')"
  cpu_cores="$(grep -c '^processor' /proc/cpuinfo)"
  if [ -n "$cpu_model" ]; then
    echo "[bench-runner] CPU            $cpu_model (${cpu_cores} cores)"
  fi
fi
if command -v node >/dev/null 2>&1; then
  echo "[bench-runner] Node           $(node --version)"
fi
if command -v pnpm >/dev/null 2>&1; then
  echo "[bench-runner] pnpm           $(pnpm --version)"
fi
if command -v go >/dev/null 2>&1; then
  echo "[bench-runner] Go (host)      $(go version)"
fi
echo "[bench-runner] === End Runner Info ==="
