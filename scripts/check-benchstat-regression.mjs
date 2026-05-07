#!/usr/bin/env node
// plan-0.9.5 §3 Tranche 2 (DoD-Item §3-2) — Regression-Detector
// für `benchstat`-Output. Parst die Tabelle aus
// `benchstat baseline.txt current.txt` und meldet jede Zeile, die
// **statistisch signifikant** mehr als die konfigurierte Schwelle
// regrediert.
//
// Schwelle: > 15 % (Plan-DoD-Default, parallel zu
// `extra-gates.md` §3.3 und `docs/perf/budgets.md` §3 Wartung).
// Override: `--threshold-percent <N>`.
//
// Statistische Signifikanz: nur Zeilen, in denen `benchstat` einen
// p-Wert < 0.05 ausweist; das vermeidet False-Positives durch
// Lauf-zu-Lauf-Rauschen. Zeilen ohne p-Wert (z. B. nur ein
// einzelner Bench-Lauf in der Quelle) zählen als rauscharm und
// werden ignoriert.
//
// Usage:
//   node scripts/check-benchstat-regression.mjs <comparison-file>
//   node scripts/check-benchstat-regression.mjs --threshold-percent 20 file
//
// Exit-Codes:
//   0 — keine signifikante Regression
//   1 — eine oder mehrere Regressionen (mit Detail-Output auf stderr)
//   2 — Aufruf-/IO-Fehler

import { readFileSync } from "node:fs";
import { argv, exit, stderr, stdout } from "node:process";

const args = parseArgs(argv.slice(2));
if (!args.file) {
  stderr.write(
    "[benchstat-check] usage: check-benchstat-regression.mjs " +
      "[--threshold-percent N] <comparison-file>\n"
  );
  exit(2);
}

let text;
try {
  text = readFileSync(args.file, "utf8");
} catch (err) {
  stderr.write(`[benchstat-check] could not read ${args.file}: ${err.message}\n`);
  exit(2);
}

const threshold = args.threshold;
const lines = text.split(/\r?\n/);
const regressions = [];
const checks = [];

// benchstat 0.0.0 (golang.org/x/perf) Tabellenzeile (sec/op-Sektion):
//   `BenchmarkX-12   1.234m ± 1%   1.456m ± 2%   +18.00% (p=0.001 n=10)`
// Der p-Wert liegt in `(p=<num> n=<num>)`. Bei `~` (kein
// signifikantes Ergebnis) druckt benchstat z. B. `~ (p=0.123 n=10)`.
//
// Für Allokations-Sektion (`B/op`, `allocs/op`) fokussieren wir nicht;
// das primäre PR-/Release-Gate ist Wall-Clock.
const pattern = /^(Benchmark\S+?)(?:-\d+)?\s+\S+\s+±\s*\S+\s+\S+\s+±\s*\S+\s+([+\-]\d+\.?\d*)%\s+\(p=([\d.]+)\s+n=\d+\)/;

for (const line of lines) {
  const m = line.match(pattern);
  if (!m) continue;
  const name = m[1];
  const deltaPct = Number(m[2]);
  const pValue = Number(m[3]);
  checks.push({ name, deltaPct, pValue });
  if (pValue < 0.05 && deltaPct > threshold) {
    regressions.push({ name, deltaPct, pValue });
  }
}

stdout.write(
  `[benchstat-check] === Regression scan (threshold +${threshold}% at p<0.05) ===\n`
);
for (const c of checks) {
  const sig = c.pValue < 0.05 ? "sig" : "ns ";
  const dir = c.deltaPct >= 0 ? "+" : "";
  stdout.write(
    `[benchstat-check]   ${sig} ${c.name}: ${dir}${c.deltaPct.toFixed(2)}% (p=${c.pValue.toFixed(3)})\n`
  );
}

if (checks.length === 0) {
  stderr.write(
    "[benchstat-check] WARN no benchstat rows recognised — input may be empty or in unexpected format\n"
  );
}

if (regressions.length === 0) {
  stdout.write(
    `[benchstat-check] OK — no significant regressions over +${threshold}%\n`
  );
  exit(0);
}

stderr.write("[benchstat-check] === REGRESSIONS ===\n");
for (const r of regressions) {
  stderr.write(
    `[benchstat-check] FAIL ${r.name}: +${r.deltaPct.toFixed(2)}% over baseline ` +
      `(p=${r.pValue.toFixed(3)} < 0.05; threshold +${threshold}%)\n`
  );
}
exit(1);

function parseArgs(arr) {
  const out = { file: null, threshold: 15 };
  for (let i = 0; i < arr.length; i++) {
    const a = arr[i];
    if (a === "--threshold-percent") {
      out.threshold = Number(arr[++i]);
    } else if (a.startsWith("-")) {
      // skip unknown flags
    } else {
      out.file = a;
    }
  }
  return out;
}
