import type { AnalyzeOptions, ManifestInput } from "./types/input.js";
import type { AnalysisResult } from "./types/result.js";
import type { AnalysisErrorResult } from "./types/error.js";
import { AnalysisError } from "./types/error.js";
import { STREAM_ANALYZER_VERSION } from "./version.js";
import { analyzeHlsManifestText } from "./internal/parsers/hls.js";

/**
 * Public Entry Point. Liefert je nach Eingabe entweder ein
 * `AnalysisResult` (Erfolg) oder ein `AnalysisErrorResult` (Fehler).
 * Damit ist die Erfolg-/Fehlerunterscheidung statisch über
 * `result.status` typisiert (plan-0.3.0 §6 Tranche 5).
 *
 * URL-Input wird in Tranche 2 mit Lade-Politik (Timeout, Größe,
 * SSRF-Schutz) implementiert; Tranche 1 lehnt ihn definiert mit
 * Code `fetch_blocked` ab, damit kein versehentlicher Netzwerkzugriff
 * entstehen kann, bevor die Schutzregeln greifen.
 */
export async function analyzeHlsManifest(
  input: ManifestInput,
  _options: AnalyzeOptions = {}
): Promise<AnalysisResult | AnalysisErrorResult> {
  const validation = validateInput(input);
  if (validation.kind === "error") {
    return toErrorResult(validation.error);
  }
  if (validation.input.kind === "url") {
    return toErrorResult(
      new AnalysisError(
        "fetch_blocked",
        "URL-Laden wird erst in 0.3.0 Tranche 2 freigeschaltet (plan-0.3.0 §3).",
        { url: validation.input.url }
      )
    );
  }
  const inputMeta =
    validation.input.baseUrl !== undefined
      ? { source: "text" as const, baseUrl: validation.input.baseUrl }
      : { source: "text" as const };
  return analyzeHlsManifestText(validation.input.text, inputMeta, STREAM_ANALYZER_VERSION);
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
    code: error.code,
    message: error.message,
    ...(error.details !== undefined ? { details: error.details } : {})
  };
}
