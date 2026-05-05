import { createRequire } from "node:module";
import { mkdtempSync, readFileSync } from "node:fs";
import { mkdir, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import path from "node:path";
import vm from "node:vm";
import { fileURLToPath, pathToFileURL } from "node:url";
import { gunzipSync } from "node:zlib";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const packageDir = path.resolve(scriptDir, "..");
const expectedVersion = "0.5.0";
const requiredTarballEntries = [
  "package/dist/index.js",
  "package/dist/index.cjs",
  "package/dist/index.d.ts",
  "package/dist/index.global.js",
  "package/README.md",
  "package/package.json"
];

const tarballPath = process.argv[2] ? path.resolve(packageDir, process.argv[2]) : "";
assert(tarballPath.endsWith(".tgz"), "usage: node scripts/pack-smoke.mjs <player-sdk-tarball.tgz>");

const packageJsonPath = path.join(packageDir, "package.json");
const sourcePackageJson = JSON.parse(readFileSync(packageJsonPath, "utf8"));

assert(sourcePackageJson.version === expectedVersion, `package.json version must be ${expectedVersion}`);
assert(sourcePackageJson.private !== true, "package must not be private");
assert(sourcePackageJson.browser === "./dist/index.global.js", "package.json browser field must point at the IIFE build");

const entries = readTarGz(tarballPath);
for (const entry of requiredTarballEntries) {
  assert(entries.has(entry), `packed tarball is missing ${entry}`);
}

const tempDir = mkdtempSync(path.join(tmpdir(), "m-trace-player-sdk-"));

try {
  const appDir = path.join(tempDir, "consumer");
  const installedPackageDir = path.join(appDir, "node_modules", ...sourcePackageJson.name.split("/"));
  await mkdir(installedPackageDir, { recursive: true });
  await writeFile(path.join(appDir, "package.json"), JSON.stringify({ name: "m-trace-player-sdk-smoke", private: true }, null, 2));

  for (const [entryName, content] of entries.entries()) {
    if (!entryName.startsWith("package/")) {
      continue;
    }
    const relativeName = entryName.slice("package/".length);
    const outputPath = path.join(installedPackageDir, relativeName);
    await mkdir(path.dirname(outputPath), { recursive: true });
    await writeFile(outputPath, content);
  }

  const installedPackageJson = JSON.parse(readFileSync(path.join(installedPackageDir, "package.json"), "utf8"));
  assert(installedPackageJson.version === expectedVersion, `installed package version must be ${expectedVersion}`);

  const esmEntry = path.join(installedPackageDir, installedPackageJson.exports["."].import);
  const esmModule = await import(pathToFileURL(esmEntry).href);
  assert(typeof esmModule.createTracker === "function", "ESM entry must export createTracker");
  assert(typeof esmModule.HttpTransport === "function", "ESM entry must export HttpTransport");

  const require = createRequire(path.join(appDir, "package.json"));
  const cjsModule = require("@npm9912/player-sdk");
  assert(typeof cjsModule.createTracker === "function", "CJS entry must export createTracker");
  assert(typeof cjsModule.HttpTransport === "function", "CJS entry must export HttpTransport");

  const browserEntry = installedPackageJson.browser;
  assert(typeof browserEntry === "string" && browserEntry.length > 0, "installed package must expose a browser entry");
  const browserCode = readFileSync(path.join(installedPackageDir, browserEntry), "utf8");
  const context = {};
  vm.createContext(context);
  vm.runInContext(browserCode, context);
  assert(typeof context.MTracePlayerSDK?.createTracker === "function", "IIFE build must expose MTracePlayerSDK.createTracker");
} finally {
  await rm(tempDir, { recursive: true, force: true });
}

function readTarGz(filePath) {
  const tar = gunzipSync(readFileSync(filePath));
  const entries = new Map();
  let offset = 0;

  while (offset + 512 <= tar.length) {
    const header = tar.subarray(offset, offset + 512);
    if (header.every((byte) => byte === 0)) {
      break;
    }

    const name = readNullTerminatedString(header.subarray(0, 100));
    const prefix = readNullTerminatedString(header.subarray(345, 500));
    const sizeText = readNullTerminatedString(header.subarray(124, 136)).trim();
    const size = Number.parseInt(sizeText || "0", 8);
    const entryName = prefix ? `${prefix}/${name}` : name;
    const contentStart = offset + 512;
    const contentEnd = contentStart + size;

    entries.set(entryName, tar.subarray(contentStart, contentEnd));
    offset = contentStart + Math.ceil(size / 512) * 512;
  }

  return entries;
}

function readNullTerminatedString(buffer) {
  const nullIndex = buffer.indexOf(0);
  const end = nullIndex >= 0 ? nullIndex : buffer.length;
  return buffer.subarray(0, end).toString("utf8");
}

function assert(condition, message) {
  if (!condition) {
    throw new Error(message);
  }
}
