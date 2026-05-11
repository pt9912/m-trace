#!/usr/bin/env node
// plan-0.9.5 §2 Tranche 1 (DoD-Item §2-4) — Budget-Validator für die
// Bench-Suite. Liest text-Output (über stdin) und prüft gegen die
// Budgets aus `docs/perf/budgets.md` §3 / §4. Verletzungen werden
// als `[bench-budget] FAIL <name>: ist=<X> soll=<Y>` auf stderr
// gemeldet; Exit-Code 1 bei mind. einer Verletzung.
//
// Zwei Eingabeformate:
//
//   - `--kind ts`: Vitest-Bench-Tabelle aus stdout. Format pro Zeile
//     (vitest 4.1):
//         `   · <name>  <hz>  <min>  <max>  <mean>  <p75>  <p99>  ...`
//     Werte sind in Millisekunden (vitest-bench-Native-Einheit). Wir
//     prüfen `mean` und `p99` gegen das Budget.
//   - `--kind go`: Go-Bench-Output (`go test -bench=. -benchmem`).
//     Format pro Zeile:
//         `BenchmarkX-12   123456 ns/op   456 B/op   7 allocs/op`
//     `ns/op` wird in Millisekunden konvertiert (`ns / 1_000_000`).
//
// Vitest-JSON-Reporter wird **nicht** verwendet — vitest 4.1 wirft
// bei `--reporter=json --outputFile=...` einen internen Server-
// Setup-Fehler. Text-Parsing ist resilient und versionsstabil.
//
// Usage:
//   node scripts/check-bench-budgets.mjs --kind ts < analyzer-bench.txt
//   node scripts/check-bench-budgets.mjs --kind go < api-bench.txt
//
// Beobachtungsphase (Plan-DoD §2-6): das Skript wird zunächst
// nicht-blockierend im Nightly-Workflow `benchmark-observation.yml`
// aufgerufen (`continue-on-error: true`). Nach N=3-5 grünen
// CI-Läufen wird das `continue-on-error` entfernt und der Smoke
// landet PR-blockierend in `make gates`.

import { argv, exit, stderr, stdin, stdout } from "node:process";

// Budget-Tabelle (Single-Source-of-Truth: docs/perf/budgets.md;
// hier als maschinenlesbares Mapping, weil Markdown-Tabellen nicht
// stabil zu parsen sind). Beim Schärfen eines Budgets müssen beide
// Stellen synchron gehalten werden — Plan-DoD §2-4 + docs/perf/
// budgets.md §5 Wartung.
//
// Zeitwerte sind alle in **Millisekunden** (vitest-bench-Native-
// Einheit für `mean`/`p75`/`p99`). Go-Bench-`ns/op` wird in
// Millisekunden konvertiert (`ns / 1_000_000`).
const TS_BUDGETS_MS = {
  "HLS Master klein (5 Variants + 1 Rendition, ≤ 5 ms)": 5,
  "HLS Master groß (50 Variants + 20 Renditions, ≤ 25 ms)": 25,
  "HLS Media (1.000 Segmente, ≤ 50 ms)": 50,
  "DASH-MPD VOD (1 Period / 2 AdaptationSets, ≤ 5 ms)": 5,
  "DASH-MPD Live (3 AdaptationSets, ≤ 10 ms)": 10,
  "Detector über 256-KiB-Body (≤ 500 µs)": 0.5,
  "SSRF-URL-Klassifizierung (100 Calls, ≤ 5 ms)": 5
};

// Go-Benchmark-Names werden über substring-match gegen die Tabelle
// geprüft (Go-Bench-Names enthalten Suffix `-N` für CPU-Count).
const GO_BUDGETS_MS = {
  "BenchmarkRegisterPlaybackEventBatch_Typical": 10,
  "BenchmarkRegisterPlaybackEventBatch_MaxBatch": 25,
  "BenchmarkEventRepository_AppendBatch_100": 100,
  "BenchmarkSessionsService_ListSessions_DefaultPage": 50,
  // plan-0.12.6 Tranche 5 / R-7: Bulk-Boundary-Read auf der Hard-Cap-
  // Page (1000 Sessions). Pre-T5 (N+1): ~1000 boundary-Roundtrips pro
  // Page. Post-T5: ein einziger IN-Clause-Call. Budget aus dem Plan-
  // DoD: < 200 ms p95.
  "BenchmarkSessionsService_ListSessions_MaxPage_BulkBoundaries": 200,
  "BenchmarkCursorEncodeDecode_Pair": 0.25
};

const args = parseArgs(argv.slice(2));

if (args.kind === "ts") {
  await checkVitestText();
} else if (args.kind === "go") {
  await checkGoBench();
} else {
  stderr.write("[bench-budget] usage: --kind ts (stdin)  OR  --kind go (stdin)\n");
  exit(2);
}

async function checkVitestText() {
  const text = await readStdin();
  const lines = text.split(/\r?\n/);
  const violations = [];
  const checks = [];
  // Vitest-Bench-Tabellenzeilen beginnen mit `   · <name>` und
  // enden mit `±<rme>%   <samples>`. Werte zwischen sind hz, min,
  // max, mean, p75, p99, p995, p999. Komma-Tausender-Trenner in hz
  // werden mit removeCommas() entfernt.
  const benchLine = /^\s*·\s+(.+?)\s+([\d,.]+)\s+([\d.]+)\s+([\d.]+)\s+([\d.]+)\s+([\d.]+)\s+([\d.]+)\s+([\d.]+)\s+([\d.]+)\s+±[\d.]+%\s+\d+\s*$/;
  for (const line of lines) {
    const m = line.match(benchLine);
    if (!m) continue;
    const name = m[1].trim();
    const mean = Number(m[5]);
    const p99 = Number(m[7]);
    const budget = TS_BUDGETS_MS[name];
    if (budget === undefined) {
      stderr.write(`[bench-budget] WARN unknown bench name in budget table: ${name}\n`);
      continue;
    }
    checks.push({ name, mean, p99, budget });
    if (Number.isFinite(p99) && p99 > budget) {
      violations.push({ name, ist: p99, soll: budget, kind: "p99" });
    } else if (Number.isFinite(mean) && mean > budget) {
      violations.push({ name, ist: mean, soll: budget, kind: "mean" });
    }
  }

  reportChecks(checks, "ts");
  if (checks.length === 0) {
    stderr.write("[bench-budget] FAIL no vitest-bench rows recognised on stdin\n");
    exit(2);
  }
  if (violations.length > 0) {
    reportViolations(violations, "ms");
    exit(1);
  }
  stdout.write(`[bench-budget] OK — all ${checks.length} TS bench(es) under budget\n`);
}

async function checkGoBench() {
  const text = await readStdin();
  const lines = text.split(/\r?\n/);
  const violations = [];
  const checks = [];
  for (const line of lines) {
    // Go-Bench output format:
    //   BenchmarkX-12   123456 ns/op   456 B/op   7 allocs/op
    // We accept any whitespace and ignore the trailing CPU-count
    // suffix (`-12`).
    const m = line.match(
      /^(Benchmark\S+?)(?:-\d+)?\s+\d+\s+([\d.]+)\s+ns\/op/
    );
    if (!m) continue;
    const name = m[1];
    const nsPerOp = Number(m[2]);
    const budget = GO_BUDGETS_MS[name];
    if (budget === undefined) {
      stderr.write(`[bench-budget] WARN unknown go-bench name: ${name}\n`);
      continue;
    }
    const ms = nsPerOp / 1_000_000;
    checks.push({ name, ms, budget });
    if (ms > budget) {
      violations.push({ name, ist: ms, soll: budget, kind: "ns/op→ms" });
    }
  }

  reportChecks(checks, "go");
  if (checks.length === 0) {
    stderr.write("[bench-budget] FAIL no go-bench rows recognised on stdin\n");
    exit(2);
  }
  if (violations.length > 0) {
    reportViolations(violations, "ms");
    exit(1);
  }
  stdout.write(`[bench-budget] OK — all ${checks.length} Go bench(es) under budget\n`);
}

function reportChecks(checks, kind) {
  stdout.write(`[bench-budget] === ${kind.toUpperCase()} bench checks ===\n`);
  for (const c of checks) {
    if (kind === "ts") {
      stdout.write(
        `[bench-budget]   ${c.name}\n` +
          `[bench-budget]     mean=${c.mean.toFixed(4)} ms p99=${c.p99.toFixed(4)} ms (budget ${c.budget} ms)\n`
      );
    } else {
      stdout.write(
        `[bench-budget]   ${c.name}: ${c.ms.toFixed(4)} ms (budget ${c.budget} ms)\n`
      );
    }
  }
}

function reportViolations(violations, unit) {
  stderr.write("[bench-budget] === BUDGET VIOLATIONS ===\n");
  for (const v of violations) {
    const overBy = ((v.ist - v.soll) / v.soll) * 100;
    stderr.write(
      `[bench-budget] FAIL ${v.name}: ist=${v.ist.toFixed(4)} ${unit} ` +
        `soll=${v.soll} ${unit} (over by ${overBy.toFixed(1)}%, source=${v.kind})\n`
    );
  }
}

function parseArgs(arr) {
  const out = { kind: null };
  for (let i = 0; i < arr.length; i++) {
    const a = arr[i];
    if (a === "--kind") out.kind = arr[++i];
  }
  return out;
}

async function readStdin() {
  const chunks = [];
  for await (const chunk of stdin) chunks.push(chunk);
  return Buffer.concat(chunks).toString("utf8");
}
