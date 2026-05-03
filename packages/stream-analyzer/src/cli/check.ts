import { readFile as fsReadFile } from "node:fs/promises";
import type { Writable } from "node:stream";
import { pathToFileURL } from "node:url";

import { analyzeHlsManifest } from "../analyze.js";
import type { AnalyzeOptions, ManifestInput } from "../types/input.js";
import type { AnalysisErrorResult } from "../types/error.js";
import type { AnalysisResult } from "../types/result.js";

type AnalyzeFn = (
  input: ManifestInput,
  options?: AnalyzeOptions
) => Promise<AnalysisResult | AnalysisErrorResult>;

type ReadFileFn = (path: string) => Promise<string>;

export interface RunCliOptions {
  readonly argv: readonly string[];
  readonly stdout: Writable;
  readonly stderr: Writable;
  /** Test-Hook: Default ist `analyzeHlsManifest`. */
  readonly analyze?: AnalyzeFn;
  /** Test-Hook: Default ist `fs/promises.readFile(path, "utf8")`. */
  readonly readFile?: ReadFileFn;
}

/** Exit-Codes — orientiert an klassischer Unix-Konvention. */
export const EXIT_OK = 0;
export const EXIT_FAILURE = 1;
export const EXIT_USAGE = 2;

/**
 * Hauptdispatcher der CLI. Komplett synchron testbar (keine
 * subprocess-spawns nötig); Tests injizieren `analyze` und `readFile`
 * als Stubs, damit der Datei- und URL-Pfad ohne reales Filesystem/
 * Netzwerk geprüft werden können.
 */
export async function runCli(opts: RunCliOptions): Promise<number> {
  const parsed = await parseCliArgs(opts);
  if (parsed.kind === "exit") return parsed.code;

  const readFile = opts.readFile ?? defaultReadFile;
  const input = await loadInput(parsed.target, readFile, opts.stderr);
  if (input === null) return EXIT_FAILURE;

  const analyze = opts.analyze ?? analyzeHlsManifest;
  let result;
  try {
    result = await analyze(input);
  } catch (error) {
    opts.stderr.write(`m-trace check: Analyse fehlgeschlagen: ${describeError(error)}\n`);
    return EXIT_FAILURE;
  }

  opts.stdout.write(JSON.stringify(result, null, 2) + "\n");
  return result.status === "ok" ? EXIT_OK : EXIT_FAILURE;
}

type CliArgsResult =
  | { readonly kind: "exit"; readonly code: number }
  | { readonly kind: "continue"; readonly target: string };

/** Validiert argv und fängt --help / --version / fehlende oder
 * unerwartete Argumente ab. Schreibt Usage/Fehler nach stderr/stdout
 * und liefert entweder einen Exit-Code (Top-Level kehrt sofort zurück)
 * oder das aufgelöste check-Target. */
async function parseCliArgs(opts: RunCliOptions): Promise<CliArgsResult> {
  const args = [...opts.argv];
  if (args.length === 0) {
    printUsage(opts.stderr);
    return { kind: "exit", code: EXIT_USAGE };
  }
  if (args[0] === "--help" || args[0] === "-h") {
    printUsage(opts.stdout);
    return { kind: "exit", code: EXIT_OK };
  }
  if (args[0] === "--version") {
    const { STREAM_ANALYZER_VERSION } = await import("../version.js");
    opts.stdout.write(`${STREAM_ANALYZER_VERSION}\n`);
    return { kind: "exit", code: EXIT_OK };
  }

  const command = args.shift();
  if (command !== "check") {
    opts.stderr.write(`m-trace: unbekanntes Kommando "${command}"\n`);
    printUsage(opts.stderr);
    return { kind: "exit", code: EXIT_USAGE };
  }

  const target = args.shift();
  if (target === undefined || target.length === 0) {
    opts.stderr.write("m-trace check: fehlendes Argument <url-or-file>\n");
    printUsage(opts.stderr);
    return { kind: "exit", code: EXIT_USAGE };
  }

  // Unbekannte Optionen explizit ablehnen, damit sich die CLI hart
  // verhält und keine versehentlich falsch geschriebenen Flags
  // schluckt.
  if (args.length > 0) {
    opts.stderr.write(`m-trace check: unerwartetes Argument "${args[0]}"\n`);
    printUsage(opts.stderr);
    return { kind: "exit", code: EXIT_USAGE };
  }

  return { kind: "continue", target };
}

/** Baut den ManifestInput aus dem aufgelösten Target: HTTP(S) →
 * URL-Input, sonst Datei-Input mit pathToFileURL als baseUrl.
 * Schreibt Lese-Fehler nach stderr und liefert null (Caller exitet
 * mit EXIT_FAILURE). */
async function loadInput(
  target: string,
  readFile: ReadFileFn,
  stderr: Writable
): Promise<ManifestInput | null> {
  if (isHttpUrl(target)) {
    return { kind: "url", url: target };
  }
  try {
    const text = await readFile(target);
    return { kind: "text", text, baseUrl: localBaseUrl(target) };
  } catch (error) {
    stderr.write(`m-trace check: Datei "${target}" konnte nicht gelesen werden: ${describeError(error)}\n`);
    return null;
  }
}

function printUsage(out: Writable): void {
  out.write(
    [
      "Usage: m-trace check <url-or-file>",
      "",
      "Argumente:",
      "  <url-or-file>   HTTP/HTTPS-URL eines HLS-Manifests oder Pfad zu einer",
      "                  lokalen Manifest-Datei (.m3u8). URL-Inputs nutzen den",
      "                  SSRF-Schutz aus @npm9912/stream-analyzer §6.",
      "",
      "Optionen:",
      "  -h, --help      Zeigt diese Hilfe an.",
      "  --version       Gibt die Analyzer-Version aus.",
      "",
      "Exit-Codes:",
      "  0  Analyse erfolgreich (status:\"ok\").",
      "  1  Analyse oder I/O fehlgeschlagen.",
      "  2  Aufruf-/Argumentfehler."
    ].join("\n") + "\n"
  );
}

function isHttpUrl(value: string): boolean {
  // RFC 3986 §3.1: scheme ist case-insensitive. Tolerieren wir, damit
  // copy-paste aus auto-korrigierten Quellen ("HTTP://...") nicht
  // versehentlich als Datei-Pfad behandelt wird.
  return /^https?:\/\//i.test(value);
}

function localBaseUrl(path: string): string | undefined {
  try {
    return pathToFileURL(path).toString();
  } catch {
    return undefined;
  }
}

function defaultReadFile(path: string): Promise<string> {
  return fsReadFile(path, "utf8");
}

function describeError(error: unknown): string {
  if (error instanceof Error) return error.message;
  return String(error);
}
