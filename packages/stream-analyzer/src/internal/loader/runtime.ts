import dns from "node:dns/promises";

/**
 * Auflösungs-Eintrag entsprechend Node `dns.lookup({ all: true })`.
 */
export interface ResolvedAddress {
  readonly address: string;
  readonly family: 4 | 6;
}

/**
 * Runtime-Schicht des Manifest-Loaders. Eine Default-Implementierung
 * (Node `dns.lookup` und globales `fetch`) wird in
 * `defaultLoaderRuntime` angeboten; Tests injizieren eigene
 * Implementierungen, damit SSRF-/Timeout-/Redirect-Verhalten ohne
 * echtes Netzwerk getestet werden kann (plan-0.3.0 §3 DoD: "Der
 * Parser arbeitet deterministisch ohne echte Netzwerkabhängigkeit").
 */
export interface LoaderRuntime {
  resolveHost(hostname: string): Promise<ResolvedAddress[]>;
  fetch(url: string, init: LoaderFetchInit): Promise<LoaderResponse>;
}

export interface LoaderFetchInit {
  readonly signal: AbortSignal;
  /** Loader übernimmt Redirect-Handling selbst. */
  readonly redirect: "manual";
  readonly headers: Readonly<Record<string, string>>;
}

export interface LoaderResponse {
  readonly status: number;
  readonly headers: LoaderHeaders;
  /** Antwortkörper als Lese-Stream; `null` für 3xx. */
  readonly body: AsyncIterable<Uint8Array> | null;
}

export interface LoaderHeaders {
  get(name: string): string | null;
}

export const defaultLoaderRuntime: LoaderRuntime = {
  async resolveHost(hostname) {
    const entries = await dns.lookup(hostname, { all: true, verbatim: true });
    return entries.map((e) => ({
      address: e.address,
      family: e.family === 4 ? 4 : 6
    }));
  },
  async fetch(url, init) {
    const response = await fetch(url, {
      signal: init.signal,
      redirect: init.redirect,
      headers: init.headers
    });
    return {
      status: response.status,
      headers: { get: (name) => response.headers.get(name) },
      body: response.body
        ? (async function* () {
            const reader = response.body!.getReader();
            try {
              while (true) {
                const { done, value } = await reader.read();
                if (done) return;
                yield value;
              }
            } finally {
              reader.releaseLock();
            }
          })()
        : null
    };
  }
};
