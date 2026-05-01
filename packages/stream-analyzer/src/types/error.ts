/**
 * Fehler-Codes der Public API. Konsumenten erkennen Fehler an
 * `status === "error"`; das ist die strukturelle Trennung gegenüber
 * dem Erfolgsergebnis (plan-0.3.0 §6 Tranche 5).
 */
export type AnalysisErrorCode =
  | "invalid_input"
  | "manifest_not_hls"
  | "fetch_failed"
  | "fetch_blocked"
  | "manifest_too_large"
  | "internal_error";

export interface AnalysisErrorResult {
  readonly status: "error";
  readonly analyzerVersion: string;
  readonly code: AnalysisErrorCode;
  readonly message: string;
  /**
   * Optionale, maschinenlesbare Zusatzinformationen — z. B. der
   * abgelehnte URL-Host bei `fetch_blocked`. Inhalt wird je
   * Code in der jeweiligen Tranche dokumentiert.
   */
  readonly details?: Readonly<Record<string, unknown>>;
}

/**
 * Fehler-Klasse, die Adapter intern werfen können; `analyzeHlsManifest`
 * fängt sie ab und übersetzt sie in ein `AnalysisErrorResult`. Direkte
 * Konsumenten nutzen normalerweise das Result, nicht den Throw-Pfad.
 */
export class AnalysisError extends Error {
  readonly code: AnalysisErrorCode;
  readonly details?: Readonly<Record<string, unknown>>;

  constructor(code: AnalysisErrorCode, message: string, details?: Readonly<Record<string, unknown>>) {
    super(message);
    this.name = "AnalysisError";
    this.code = code;
    this.details = details;
  }
}
