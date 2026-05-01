/**
 * Drei-Stufen-Skala aus plan-0.3.0 §5 (Tranche 4). Konsumenten dürfen
 * weitere Stufen behandeln, ohne dass das Schema bricht — additive
 * Erweiterungen bleiben erlaubt.
 */
export type FindingLevel = "info" | "warning" | "error";

export interface AnalysisFinding {
  /** Stabile, maschinenlesbare Kennung, z. B. `not_implemented`. */
  readonly code: string;
  readonly level: FindingLevel;
  /** Menschlich lesbare Begründung. */
  readonly message: string;
}
