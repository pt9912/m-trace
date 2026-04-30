import { readFileSync, readdirSync, statSync } from "node:fs";
import { join } from "node:path";

const root = new URL("..", import.meta.url).pathname;
const coreDir = join(root, "src", "core");
const forbidden = /from\s+["']\.\.\/adapters\/|from\s+["']hls\.js["']/;

function files(dir) {
  return readdirSync(dir).flatMap((entry) => {
    const path = join(dir, entry);
    return statSync(path).isDirectory() ? files(path) : [path];
  });
}

const offenders = files(coreDir).filter((path) => forbidden.test(readFileSync(path, "utf8")));
if (offenders.length > 0) {
  console.error("core/ must not import browser adapters or hls.js:");
  for (const offender of offenders) {
    console.error(`- ${offender}`);
  }
  process.exit(1);
}
