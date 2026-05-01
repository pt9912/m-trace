import type { AnalyzeOptions, FetchOptions, ManifestInput } from "./types/input.js";
import type { AnalysisInputMetadata, AnalysisResult } from "./types/result.js";
import type { AnalysisErrorResult } from "./types/error.js";
import { AnalysisError } from "./types/error.js";
import { STREAM_ANALYZER_VERSION } from "./version.js";
import { analyzeHlsManifestText } from "./internal/parsers/hls.js";
import { loadHlsManifest } from "./internal/loader/fetch.js";
import { defaultLoaderRuntime, type LoaderRuntime } from "./internal/loader/runtime.js";

const FETCH_DEFAULTS: Required<FetchOptions> = {
  timeoutMs: 10_000,
  maxBytes: 5_000_000,
  maxRedirects: 5,
  allowPrivateNetworks: false
};

/**
 * Public Entry Point. Liefert je nach Eingabe entweder ein
 * `AnalysisResult` (Erfolg) oder ein `AnalysisErrorResult` (Fehler).
 */
export async function analyzeHlsManifest(
  input: ManifestInput,
  options: AnalyzeOptions = {}
): Promise<AnalysisResult | AnalysisErrorResult> {
  return analyzeWithRuntime(input, options, defaultLoaderRuntime);
}

/**
 * Internal Entry Point für Tests: erlaubt Injektion einer Runtime,
 * damit SSRF-/Loader-Verhalten ohne echtes Netzwerk geprüft werden
 * kann. Nicht Teil der publizierten API.
 */
export async function analyzeWithRuntime(
  input: ManifestInput,
  options: AnalyzeOptions,
  runtime: LoaderRuntime
): Promise<AnalysisResult | AnalysisErrorResult> {
  const validation = validateInput(input);
  if (validation.kind === "error") {
    return toErrorResult(validation.error);
  }
  const validInput = validation.input;
  if (validInput.kind === "url") {
    return loadAndAnalyzeUrl(validInput.url, options, runtime);
  }
  const inputMeta: AnalysisInputMetadata =
    validInput.baseUrl !== undefined
      ? { source: "text", baseUrl: validInput.baseUrl }
      : { source: "text" };
  return runParser(validInput.text, inputMeta);
}

async function loadAndAnalyzeUrl(
  url: string,
  options: AnalyzeOptions,
  runtime: LoaderRuntime
): Promise<AnalysisResult | AnalysisErrorResult> {
  const fetchOpts: Required<FetchOptions> = {
    ...FETCH_DEFAULTS,
    ...(options.fetch ?? {})
  };
  let loaded;
  try {
    loaded = await loadHlsManifest(url, {
      runtime,
      timeoutMs: fetchOpts.timeoutMs,
      maxBytes: fetchOpts.maxBytes,
      maxRedirects: fetchOpts.maxRedirects,
      allowPrivateNetworks: fetchOpts.allowPrivateNetworks
    });
  } catch (error) {
    if (error instanceof AnalysisError) {
      return toErrorResult(error);
    }
    throw error;
  }
  const inputMeta: AnalysisInputMetadata = {
    source: "url",
    url,
    baseUrl: loaded.finalUrl
  };
  return runParser(loaded.text, inputMeta);
}

function runParser(text: string, inputMeta: AnalysisInputMetadata): AnalysisResult | AnalysisErrorResult {
  try {
    return analyzeHlsManifestText(text, inputMeta, STREAM_ANALYZER_VERSION);
  } catch (error) {
    if (error instanceof AnalysisError) {
      return toErrorResult(error);
    }
    throw error;
  }
}

type Validation = { kind: "ok"; input: ManifestInput } | { kind: "error"; error: AnalysisError };

function validateInput(input: ManifestInput): Validation {
  if (input.kind === "text") {
    if (typeof input.text !== "string") {
      return {
        kind: "error",
        error: new AnalysisError("invalid_input", "ManifestInput.text muss ein String sein.")
      };
    }
    return { kind: "ok", input };
  }
  if (input.kind === "url") {
    if (typeof input.url !== "string" || input.url.length === 0) {
      return {
        kind: "error",
        error: new AnalysisError("invalid_input", "ManifestInput.url darf nicht leer sein.")
      };
    }
    return { kind: "ok", input };
  }
  return {
    kind: "error",
    error: new AnalysisError("invalid_input", "ManifestInput.kind muss 'text' oder 'url' sein.")
  };
}

function toErrorResult(error: AnalysisError): AnalysisErrorResult {
  return {
    status: "error",
    analyzerVersion: STREAM_ANALYZER_VERSION,
    analyzerKind: "hls",
    code: error.code,
    message: error.message,
    ...(error.details !== undefined ? { details: error.details } : {})
  };
}
