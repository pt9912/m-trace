/**
 * Eingabeformen für die Manifestanalyse. Tranche 1 legt die
 * Diskriminierung fest; Tranche 2 hat das URL-Laden mit Timeout,
 * Größenlimit und SSRF-Schutz angeschlossen.
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
   * Quell-URL des Manifests. Loader erzwingt http/https, blockt
   * lokale/private/link-local/loopback/reservierte IP-Bereiche und
   * verbietet Credentials in der URL (siehe `docs/user/stream-
   * analyzer.md` §6).
   */
  readonly url: string;
}

export interface FetchOptions {
  /** Zeitlimit pro HTTP-Hop in Millisekunden. Default: 10_000. */
  readonly timeoutMs?: number;
  /** Maximaler Bytes-Cap für den gesamten Body. Default: 5_000_000. */
  readonly maxBytes?: number;
  /** Maximal zulässige Redirect-Hops. Default: 5. */
  readonly maxRedirects?: number;
}

/**
 * Optionen, die aufruferseitig nicht zwingend gesetzt werden müssen,
 * den Analyseaufruf aber feinjustieren. Tranche 4 ergänzt z. B. eine
 * `segmentDurationToleranceFraction`.
 */
export interface AnalyzeOptions {
  /** Optionen für den URL-Loader (greift nur bei `kind === "url"`). */
  readonly fetch?: FetchOptions;
}
