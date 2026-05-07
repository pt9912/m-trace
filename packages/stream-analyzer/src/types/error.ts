import type { AnalyzerKind } from "./result.js";

/**
 * Fehler-Codes der Public API. Konsumenten erkennen Fehler an
 * `status === "error"`; das ist die strukturelle Trennung gegenüber
 * dem Erfolgsergebnis (plan-0.3.0 §6 Tranche 5).
 *
 * **`manifest_not_hls`** ist HLS-spezifisch: der HLS-Parser hat das
 * Manifest abgelehnt, weil es nicht mit `#EXTM3U` beginnt oder
 * leer ist. Der Code wird **nur** erzeugt, wenn der Detector den
 * Input als HLS klassifiziert hat (Content-Type-Heuristik / erste
 * Zeile beginnt mit `#EXTM3U`-Präfix).
 *
 * **`manifest_not_supported`** ist die Sammelantwort des Detectors
 * für Eingaben, die weder als HLS noch als DASH erkannt werden
 * (`0.9.0` Tranche 3). Beispiele: HTML-Bodies, JSON-Bodies, leere
 * Texte, beliebige Binärdaten.
 */
export type AnalysisErrorCode =
  | "invalid_input"
  | "manifest_not_hls"
  | "manifest_not_supported"
  | "fetch_failed"
  | "fetch_blocked"
  | "manifest_too_large"
  | "internal_error";

export interface AnalysisErrorResult {
  readonly status: "error";
  readonly analyzerVersion: string;
  /**
   * Identifiziert den ausführenden Analyzer (`"hls"` oder `"dash"`).
   * Bei `manifest_not_supported` ist `analyzerKind` der Stand des
   * Detectors zum Zeitpunkt des Fehlers — typischerweise `"hls"`,
   * weil der Detector HLS als Default zurückgibt, wenn keine
   * DASH-Marker gefunden werden.
   */
  readonly analyzerKind: AnalyzerKind;
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
