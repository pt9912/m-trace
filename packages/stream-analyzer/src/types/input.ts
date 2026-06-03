/**
 * Eingabeformen für die Manifestanalyse. legt die
 * Diskriminierung fest; hat das URL-Laden mit Timeout,
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
  /**
   * Zeitlimit pro HTTP-Hop in Millisekunden. Default: 10_000.
   *
   * Gilt für URL-Manifeste **und** für binäre Segment-Fetches aus
   * Text-Inputs mit sicherer HTTP(S)-`baseUrl` (`0.10.0`,
   * NF-13 / RAK-64).
   */
  readonly timeoutMs?: number;
  /**
   * Maximaler Bytes-Cap für den gesamten Manifest-Body. Default:
   * 5_000_000.
   *
   * **Wichtig:** `maxBytes` ist ausschließlich das Manifest-Body-
   * Limit. Binäre Segment-Fetches nutzen
   * `CmafBinaryOptions.maxSegmentBytes` (Default 2_000_000) — siehe
   * `cmaf.binary` unten. Eine Verschiebung von `maxBytes` ändert
   * Segment-Limits nicht.
   */
  readonly maxBytes?: number;
  /**
   * Maximal zulässige Redirect-Hops. Default: 5. Gilt für Manifest-
   * und Segment-Fetches.
   */
  readonly maxRedirects?: number;
  /**
   * Opt-in: lockert die IPv4-/IPv6-Sperrlisten für loopback,
   * private und link-local-Bereiche. Default `false` (Block bleibt).
   * Scheme-Whitelist, Credentials-Block und Größen-/Redirect-Regeln
   * bleiben **unabhängig** vom Flag aktiv.
   *
   * Vorgesehen für streng-vertrauenswürdige Lokal-/Compose-Setups
   * (z. B. analyzer-service mit `ANALYZER_ALLOW_PRIVATE_NETWORKS=true`),
   * in denen interne mediamtx-Streams gegen den Compose-Hostnamen
   * geprüft werden sollen. Produktionsdeployments sollten das Flag
   * NICHT setzen — der SSRF-Schutz vor RFC1918-Adressen ist eine
   * der Kernregeln des Loaders.
   */
  readonly allowPrivateNetworks?: boolean;
}

/**
 * Optionen für die binäre CMAF-Konformitätsprüfung (`0.10.0`,
 * NF-13 / RAK-64). Aktiv nur für Detail-Scopes, in denen Init-/
 * Media-Segment-Referenzen ableitbar sind (HLS Media-Playlist und
 * DASH-MPD).
 */
export interface CmafBinaryOptions {
  /**
   * Schaltet die binäre Prüfung an oder aus. Default: `true`. Bei
   * `false` bleibt `details.cmaf` bestehen (Manifestsignale werden
   * weiterhin emittiert), aber das `binary`-Objekt trägt
   * `status:"skipped"` mit Failure-Code `binary_disabled`. Stiller
   * Fallback auf Defaults ist nicht zulässig.
   */
  readonly enabled?: boolean;
  /**
   * Pro-Segment-Body-Limit in Bytes. Default: 2_000_000 (≈ 2 MB).
   * Segmente, die das Limit überschreiten, werden mit
   * `segment_too_large` als `skipped` berichtet.
   */
  readonly maxSegmentBytes?: number;
  /**
   * Globale Obergrenze für tatsächlich gefetchete Init-/Media-
   * Segmente pro Analyseaufruf. Default: 6 — deckt bewusst bis zu
   * drei DASH-AdaptationSets mit je Init- und Media-Prüfung ab.
   * Größere Pflichtprüfmengen tragen `not_planned_due_to_limit` und
   * verhindern `binary.status:"passed"`, sofern der Aufrufer das
   * Limit nicht erhöht.
   */
  readonly maxBinarySegments?: number;
}

/**
 * Optionssektion für die CMAF-Analyse. Aktuell nur
 * `binary` — manifestbasierte Signal-Erkennung ist immer aktiv,
 * weil sie keinen zusätzlichen Netzwerkverkehr erzeugt.
 */
export interface CmafAnalyzeOptions {
  readonly binary?: CmafBinaryOptions;
}

/**
 * Optionen, die aufruferseitig nicht zwingend gesetzt werden müssen,
 * den Analyseaufruf aber feinjustieren.
 */
export interface AnalyzeOptions {
  /**
   * Optionen für den URL-Loader. Greift bei `kind === "url"` und
   * — (NF-13 / RAK-64) — auch für binäre Segment-Fetches
   * aus Text-Inputs mit sicherer HTTP(S)-`baseUrl`. `fetch.maxBytes`
   * bleibt ausschließlich das Manifest-Body-Limit; Segment-Größen
   * werden über `cmaf.binary.maxSegmentBytes` konfiguriert.
   */
  readonly fetch?: FetchOptions;
  /**
   * CMAF-spezifische Optionen (`0.10.0`, NF-13 / RAK-64).
   */
  readonly cmaf?: CmafAnalyzeOptions;
}
