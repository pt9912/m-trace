import { readFile } from "node:fs/promises";
import { fileURLToPath } from "node:url";
import path from "node:path";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const packageDir = path.resolve(scriptDir, "..");
const indexPath = path.join(packageDir, "src", "index.ts");
const snapshotPath = path.join(scriptDir, "public-api.snapshot.txt");

const [indexSource, snapshot] = await Promise.all([
  readFile(indexPath, "utf8"),
  readFile(snapshotPath, "utf8")
]);

const publicExports = indexSource
  .split("\n")
  .filter((line) => line.startsWith("export "))
  .join("\n")
  .trim();

const expectedExports = snapshot.trim();

if (publicExports !== expectedExports) {
  console.error("Public API snapshot mismatch.");
  console.error(`Update ${path.relative(packageDir, snapshotPath)} intentionally when changing package exports.`);
  process.exit(1);
}
