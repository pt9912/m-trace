import type {
  AnalyzeOptions,
  CmafBinaryOptions,
  FetchOptions,
  ManifestInput
} from "./types/input.js";
import type {
  AnalysisInputMetadata,
  AnalysisResult,
  AnalyzerKind,
  CmafBinaryVerification,
  CmafSignalSummary,
  DashAnalysisResult,
  MediaAnalysisResult
} from "./types/result.js";
import type { AnalysisErrorResult } from "./types/error.js";
import { AnalysisError } from "./types/error.js";
import { STREAM_ANALYZER_VERSION } from "./version.js";
import { analyzeHlsManifestText, type HlsAnalyzeOutput } from "./internal/parsers/hls.js";
import {
  analyzeDashManifestText,
  type DashAnalyzeOutput
} from "./internal/parsers/dash.js";
import { detectManifestKind } from "./internal/parsers/detect.js";
import { loadManifest } from "./internal/loader/fetch.js";
import { defaultLoaderRuntime, type LoaderRuntime } from "./internal/loader/runtime.js";
import {
  verifyBinaryCmaf,
  withBinaryConfidenceUpgrade,
  type BinaryVerifyOptions
} from "./internal/cmaf/binary-verify.js";

const FETCH_DEFAULTS: Required<FetchOptions> = {
  timeoutMs: 10_000,
  maxBytes: 5_000_000,
  maxRedirects: 5,
  allowPrivateNetworks: false
};

const CMAF_BINARY_DEFAULTS: Required<CmafBinaryOptions> = {
  enabled: true,
  maxSegmentBytes: 2_000_000,
  maxBinarySegments: 6
};

/**
 * Public Entry Point. Liefert je nach Eingabe entweder ein
 * `AnalysisResult` (Erfolg) oder ein `AnalysisErrorResult` (Fehler).
 *
 * Ab (RAK-58 / NF-12) dispatcht der Analyzer auf
 * dem Manifest-Body intern: HLS-Eingaben (erste Zeile `#EXTM3U`)
 * laufen durch den HLS-Parser, DASH-Eingaben (`<?xml`/`<MPD`-Header)
 * durch den MPD-Parser. Der Funktionsname `analyzeHlsManifest` bleibt
 * aus Backward-Kompat-Gründen erhalten; der generischere Alias
 * `analyzeManifest` ist Public-API.
 */
export async function analyzeHlsManifest(
  input: ManifestInput,
  options: AnalyzeOptions = {}
): Promise<AnalysisResult | AnalysisErrorResult> {
  return analyzeWithRuntime(input, options, defaultLoaderRuntime);
}

/**
 * Public Entry Point ab. Funktionsidentisch zu
 * `analyzeHlsManifest`; der generischere Name spiegelt, dass der
 * Dispatcher seit dieser Phase HLS und DASH unterstützt.
 */
export async function analyzeManifest(
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
    return toErrorResult(validation.error, "hls");
  }
  const validInput = validation.input;
  const fetchOpts: Required<FetchOptions> = {
    ...FETCH_DEFAULTS,
    ...(options.fetch ?? {})
  };
  if (validInput.kind === "url") {
    return loadAndAnalyzeUrl(validInput.url, options, fetchOpts, runtime);
  }
  const inputMeta: AnalysisInputMetadata =
    validInput.baseUrl !== undefined
      ? { source: "text", baseUrl: validInput.baseUrl }
      : { source: "text" };
  return runParserAndVerify(validInput.text, inputMeta, options, fetchOpts, runtime);
}

async function loadAndAnalyzeUrl(
  url: string,
  options: AnalyzeOptions,
  fetchOpts: Required<FetchOptions>,
  runtime: LoaderRuntime
): Promise<AnalysisResult | AnalysisErrorResult> {
  let loaded;
  try {
    loaded = await loadManifest(url, {
      runtime,
      timeoutMs: fetchOpts.timeoutMs,
      maxBytes: fetchOpts.maxBytes,
      maxRedirects: fetchOpts.maxRedirects,
      allowPrivateNetworks: fetchOpts.allowPrivateNetworks
    });
  } catch (error) {
    if (error instanceof AnalysisError) {
      return toErrorResult(error, "hls");
    }
    throw error;
  }
  const inputMeta: AnalysisInputMetadata = {
    source: "url",
    url,
    baseUrl: loaded.finalUrl
  };
  return runParserAndVerify(loaded.text, inputMeta, options, fetchOpts, runtime);
}

async function runParserAndVerify(
  text: string,
  inputMeta: AnalysisInputMetadata,
  options: AnalyzeOptions,
  fetchOpts: Required<FetchOptions>,
  runtime: LoaderRuntime
): Promise<AnalysisResult | AnalysisErrorResult> {
  const detected = detectManifestKind(text);
  if (detected.kind === "unsupported") {
    return toErrorResult(
      new AnalysisError(
        "manifest_not_supported",
        "Manifest-Body wurde weder als HLS (#EXTM3U-Header) noch als DASH (<?xml/<MPD-Header) erkannt.",
        { firstLine: detected.firstLine }
      ),
      "hls"
    );
  }
  try {
    if (detected.kind === "dash") {
      const dashOut = analyzeDashManifestText(text, inputMeta, STREAM_ANALYZER_VERSION);
      return await maybeVerifyDash(dashOut, options, fetchOpts, runtime, inputMeta.baseUrl);
    }
    const hlsOut = analyzeHlsManifestText(text, inputMeta, STREAM_ANALYZER_VERSION);
    return await maybeVerifyHls(hlsOut, options, fetchOpts, runtime, inputMeta.baseUrl);
  } catch (error) {
    if (error instanceof AnalysisError) {
      return toErrorResult(error, detected.kind);
    }
    throw error;
  }
}

async function maybeVerifyHls(
  out: HlsAnalyzeOutput,
  options: AnalyzeOptions,
  fetchOpts: Required<FetchOptions>,
  runtime: LoaderRuntime,
  baseUrl: string | undefined
): Promise<AnalysisResult> {
  if (out.result.playlistType !== "media") return out.result;
  const summary = out.result.details.cmaf;
  if (summary === undefined) return out.result;
  const binary = await verifyBinaryCmaf(
    { source: "hls", mediaCmaf: out.cmafMeta },
    buildVerifyOptions(options, fetchOpts, runtime, baseUrl)
  );
  return attachHlsBinary(out.result, summary, binary);
}

async function maybeVerifyDash(
  out: DashAnalyzeOutput,
  options: AnalyzeOptions,
  fetchOpts: Required<FetchOptions>,
  runtime: LoaderRuntime,
  baseUrl: string | undefined
): Promise<AnalysisResult> {
  const summary = out.result.details.cmaf;
  if (summary === undefined) return out.result;
  const binary = await verifyBinaryCmaf(
    { source: "dash", dashCmaf: out.cmafMeta },
    buildVerifyOptions(options, fetchOpts, runtime, baseUrl)
  );
  return attachDashBinary(out.result, summary, binary);
}

function buildVerifyOptions(
  options: AnalyzeOptions,
  fetchOpts: Required<FetchOptions>,
  runtime: LoaderRuntime,
  baseUrl: string | undefined
): BinaryVerifyOptions {
  return {
    cmafBinary: { ...CMAF_BINARY_DEFAULTS, ...(options.cmaf?.binary ?? {}) },
    timeoutMs: fetchOpts.timeoutMs,
    maxRedirects: fetchOpts.maxRedirects,
    allowPrivateNetworks: fetchOpts.allowPrivateNetworks,
    runtime,
    baseUrl
  };
}

function attachHlsBinary(
  result: MediaAnalysisResult,
  summary: CmafSignalSummary,
  binary: CmafBinaryVerification
): MediaAnalysisResult {
  return {
    ...result,
    details: { ...result.details, cmaf: withBinaryConfidenceUpgrade(summary, binary) }
  };
}

function attachDashBinary(
  result: DashAnalysisResult,
  summary: CmafSignalSummary,
  binary: CmafBinaryVerification
): DashAnalysisResult {
  return {
    ...result,
    details: { ...result.details, cmaf: withBinaryConfidenceUpgrade(summary, binary) }
  };
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

function toErrorResult(error: AnalysisError, analyzerKind: AnalyzerKind): AnalysisErrorResult {
  return {
    status: "error",
    analyzerVersion: STREAM_ANALYZER_VERSION,
    analyzerKind,
    code: error.code,
    message: error.message,
    ...(error.details !== undefined ? { details: error.details } : {})
  };
}
