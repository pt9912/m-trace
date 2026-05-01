/**
 * Eingabeformen für die Manifestanalyse. Tranche 1 legt die
 * Diskriminierung fest; das tatsächliche Laden via URL kommt in
 * Tranche 2 (plan-0.3.0 §3) inklusive Timeout, Größenlimit und
 * SSRF-Schutz.
 */
export type ManifestInput = ManifestTextInput | ManifestUrlInput;

export interface ManifestTextInput {
  readonly kind: "text";
  /** Roher Manifestinhalt. */
  readonly text: string;
  /** Optionale Base-URL zur Auflösung relativer Segment-/Variant-URIs. */
  readonly baseUrl?: string;
}

export interface ManifestUrlInput {
  readonly kind: "url";
  /**
   * Quell-URL des Manifests. Tranche 2 erzwingt http/https und
   * blockiert lokale/private Adressen sowie Credentials in der URL.
   */
  readonly url: string;
}

/**
 * Optionen, die aufruferseitig nicht zwingend gesetzt werden müssen,
 * den Analyseaufruf aber feinjustieren. Konkrete Felder kommen mit
 * den jeweiligen Tranchen hinzu (z. B. `fetchTimeoutMs` in Tranche 2,
 * `segmentDurationToleranceFraction` in Tranche 4); das Interface
 * bleibt bewusst leer, bis das erste Feld einzieht.
 */
// eslint-disable-next-line @typescript-eslint/no-empty-object-type
export interface AnalyzeOptions {}
