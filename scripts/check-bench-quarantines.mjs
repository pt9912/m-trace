#!/usr/bin/env node
// plan-0.9.5 §3 Tranche 2 (DoD-Item §3-6) — Quarantäne-Tag-Check
// für Bench-Suites. Scant Bench-Files nach
// `// bench:quarantine YYYY-MM-DD reason: <text>` direkt über
// `func BenchmarkX(...)` (Go) oder `bench("...", ...)` (TS).
//
// Verhalten:
//
//   - exit 1 wenn ein Tag älter als die maximale Quarantäne-Dauer
//     ist (Default 30 Tage; Plan-DoD-Wartungsregel). Operator muss
//     das Tag entweder verlängern (Plan-DoD-Item-Änderung im
//     Folge-Plan, kein stiller Re-Skip) oder den Bench fixen.
//   - exit 0 sonst; gibt die Liste der aktiven Quarantänen entweder
//     auf stdout oder als JSON-File aus (`--output <path>`).
//
// Usage:
//   node scripts/check-bench-quarantines.mjs apps/api packages/stream-analyzer
//   node scripts/check-bench-quarantines.mjs --output .tmp/bench/quarantine.json apps/api packages/stream-analyzer
//   node scripts/check-bench-quarantines.mjs --max-age-days 30 apps/api
//
// Ausgabe-Format (JSON):
//   [{ name: "BenchmarkX", file: "apps/api/.../foo_bench_test.go",
//      expires: "2026-06-06", reason: "flaky on CI" }, ...]

import { readdirSync, readFileSync, statSync, writeFileSync, mkdirSync } from "node:fs";
import { argv, exit, stderr, stdout } from "node:process";
import { dirname, join } from "node:path";

const args = parseArgs(argv.slice(2));
if (args.roots.length === 0) {
  stderr.write(
    "[bench-quarantine] usage: check-bench-quarantines.mjs " +
      "[--max-age-days N] [--output <path>] <dir>...\n"
  );
  exit(2);
}

const maxAgeDays = args.maxAgeDays;
const today = new Date();
today.setUTCHours(0, 0, 0, 0);

const benchFiles = [];
for (const root of args.roots) {
  walk(root, benchFiles);
}

const quarantines = [];
const expired = [];

for (const file of benchFiles) {
  const text = readFileSync(file, "utf8");
  const lines = text.split(/\r?\n/);
  for (let i = 0; i < lines.length; i++) {
    const m = lines[i].match(
      /^\s*\/\/\s*bench:quarantine\s+(\d{4}-\d{2}-\d{2})\s+reason:\s*(.+?)\s*$/
    );
    if (!m) continue;
    const startDate = m[1];
    const reason = m[2];
    const benchName = findBenchName(lines, i);
    if (benchName === null) {
      stderr.write(
        `[bench-quarantine] WARN tag in ${file}:${i + 1} has no following bench/Benchmark; skipped\n`
      );
      continue;
    }
    const start = parseUtcDate(startDate);
    if (start === null) {
      stderr.write(
        `[bench-quarantine] WARN tag in ${file}:${i + 1} has invalid date ${startDate}; skipped\n`
      );
      continue;
    }
    const ageDays = Math.floor((today.getTime() - start.getTime()) / (24 * 3_600_000));
    const expiresAt = new Date(start.getTime() + maxAgeDays * 24 * 3_600_000);
    const expires = isoDate(expiresAt);
    const entry = { name: benchName, file, expires, reason, started: startDate, ageDays };
    quarantines.push(entry);
    if (ageDays > maxAgeDays) {
      expired.push(entry);
    }
  }
}

stdout.write(`[bench-quarantine] === Quarantänen-Check (max ${maxAgeDays} Tage) ===\n`);
for (const q of quarantines) {
  const status = q.ageDays > maxAgeDays ? "EXPIRED" : "active";
  stdout.write(
    `[bench-quarantine]   ${status}: ${q.name} (started ${q.started}, age ${q.ageDays}d, expires ${q.expires}) — ${q.reason}\n`
  );
}
if (quarantines.length === 0) {
  stdout.write("[bench-quarantine]   (no quarantine tags found)\n");
}

if (args.output) {
  mkdirSync(dirname(args.output), { recursive: true });
  writeFileSync(args.output, JSON.stringify(quarantines, null, 2) + "\n", "utf8");
  stdout.write(`[bench-quarantine] wrote ${args.output}\n`);
}

if (expired.length > 0) {
  stderr.write("[bench-quarantine] === EXPIRED QUARANTINES ===\n");
  for (const e of expired) {
    stderr.write(
      `[bench-quarantine] FAIL ${e.name}: started ${e.started}, age ${e.ageDays}d > ${maxAgeDays}d max\n` +
        `[bench-quarantine]   file: ${e.file}\n` +
        `[bench-quarantine]   reason: ${e.reason}\n`
    );
  }
  stderr.write(
    "[bench-quarantine] Operator action: either fix the bench and remove the tag,\n" +
      "[bench-quarantine] or extend the tag with a Plan-DoD-Item update (no silent re-skip).\n"
  );
  exit(1);
}

exit(0);

function walk(dir, out) {
  let entries;
  try {
    entries = readdirSync(dir, { withFileTypes: true });
  } catch {
    return;
  }
  for (const entry of entries) {
    const path = join(dir, entry.name);
    if (entry.isDirectory()) {
      if (entry.name === "node_modules" || entry.name === "dist" || entry.name.startsWith(".")) {
        continue;
      }
      walk(path, out);
    } else if (entry.isFile()) {
      if (entry.name.endsWith("_bench_test.go") || entry.name.endsWith(".bench.ts")) {
        out.push(path);
      }
    }
  }
}

function findBenchName(lines, idx) {
  // Schau bis zu 5 Zeilen nach unten nach `func BenchmarkX(` (Go)
  // oder `bench("X"` (TS). Ein Quarantäne-Tag hängt an genau einer
  // Bench-Definition; weiter entfernte Tags sind nicht zulässig.
  for (let j = idx + 1; j < Math.min(lines.length, idx + 6); j++) {
    const goMatch = lines[j].match(/^\s*func\s+(Benchmark\w+)/);
    if (goMatch) return goMatch[1];
    const tsMatch = lines[j].match(/^\s*bench\s*\(\s*["'](.+?)["']/);
    if (tsMatch) return tsMatch[1];
  }
  return null;
}

function parseUtcDate(s) {
  const [y, m, d] = s.split("-").map(Number);
  if (!y || !m || !d) return null;
  return new Date(Date.UTC(y, m - 1, d));
}

function isoDate(date) {
  const y = date.getUTCFullYear();
  const m = String(date.getUTCMonth() + 1).padStart(2, "0");
  const d = String(date.getUTCDate()).padStart(2, "0");
  return `${y}-${m}-${d}`;
}

function parseArgs(arr) {
  const out = { roots: [], maxAgeDays: 30, output: null };
  for (let i = 0; i < arr.length; i++) {
    const a = arr[i];
    if (a === "--max-age-days") {
      out.maxAgeDays = Number(arr[++i]);
    } else if (a === "--output") {
      out.output = arr[++i];
    } else if (a.startsWith("-")) {
      stderr.write(`[bench-quarantine] unknown flag: ${a}\n`);
      exit(2);
    } else {
      out.roots.push(a);
    }
  }
  return out;
}
