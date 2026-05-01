import { readFileSync, readdirSync, statSync } from "node:fs";
import { join, relative } from "node:path";

const root = new URL("..", import.meta.url).pathname;
const srcDir = join(root, "src");
const internalDir = join(srcDir, "internal");

// Public files (alles direkt unter src/, ohne internal/) dürfen nicht
// implementierungsspezifische Internas re-exportieren. Konsumenten
// müssen über den Package-Entry-Point gehen (plan-0.3.0 §2 Tranche 1:
// "dokumentierte Konsumenten importieren nur über den Package-
// Entry-Point").
const internalImportPattern = /from\s+["'](\.{1,2}\/)*internal\//;

function* walk(dir) {
  for (const entry of readdirSync(dir)) {
    const path = join(dir, entry);
    if (statSync(path).isDirectory()) {
      yield* walk(path);
    } else if (path.endsWith(".ts")) {
      yield path;
    }
  }
}

const offenders = [];
for (const file of walk(srcDir)) {
  if (file.startsWith(internalDir + "/")) continue;
  if (file === join(srcDir, "analyze.ts")) continue; // expliziter Re-Export-Punkt
  const content = readFileSync(file, "utf8");
  if (internalImportPattern.test(content)) {
    offenders.push(relative(root, file));
  }
}

if (offenders.length > 0) {
  console.error("Public modules must not import from internal/ outside the analyze entry point:");
  for (const offender of offenders) {
    console.error(`- ${offender}`);
  }
  process.exit(1);
}
