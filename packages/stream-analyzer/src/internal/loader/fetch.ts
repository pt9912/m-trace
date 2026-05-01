import { AnalysisError } from "../../types/error.js";
import type { LoaderRuntime, ResolvedAddress } from "./runtime.js";
import { parseIPv4, parseIPv6, validateResolvedIp, validateUrl } from "./ssrf.js";

export interface LoadOptions {
  readonly runtime: LoaderRuntime;
  readonly timeoutMs: number;
  readonly maxBytes: number;
  readonly maxRedirects: number;
}

export interface LoadResult {
  readonly text: string;
  /** URL nach allen Redirects, geeignet als Base-URL für relative URIs. */
  readonly finalUrl: string;
}

const ALLOWED_CONTENT_TYPES: ReadonlySet<string> = new Set([
  "application/vnd.apple.mpegurl",
  "application/x-mpegurl",
  "audio/mpegurl",
  "text/plain"
]);

/**
 * Lädt ein HLS-Manifest unter den Tranche-2-Schutzregeln.
 *
 * Ablauf pro Hop:
 *   1. URL-Form prüfen (Schema, Credentials, Host).
 *   2. Hostname klassifizieren: IPv4-/IPv6-Literal direkt gegen die
 *      Sperrliste prüfen; Domain-Hostname per `runtime.resolveHost`
 *      auflösen und jeden zurückgegebenen Eintrag prüfen. Schon ein
 *      blockierter Eintrag bricht ab, damit Mehrfach-A/AAAA-Antworten
 *      keine Lücken aufmachen.
 *   3. fetch mit AbortController-Timeout und manuellem Redirect.
 *      Der Timer läuft über die gesamte Hop-Dauer (Header **und**
 *      Body), ein Slow-Loris-Server kann ihn nicht aushebeln.
 *   4. Bei 3xx: Location interpretieren, Hop-Zähler prüfen, von vorn.
 *   5. Bei 2xx: Content-Type prüfen, Body bis maxBytes streamen,
 *      andernfalls AnalysisError(manifest_too_large).
 *
 * `maxRedirects` zählt die zulässige Anzahl gefolgter Redirects;
 * danach kommt noch der finale Hop. Bei `maxRedirects = 5` sind also
 * höchstens 6 fetches insgesamt.
 */
export async function loadHlsManifest(url: string, options: LoadOptions): Promise<LoadResult> {
  let currentUrl = url;
  for (let hop = 0; hop <= options.maxRedirects; hop++) {
    const next = await fetchHop(currentUrl, options, hop);
    if (next.kind === "final") {
      return { text: next.text, finalUrl: currentUrl };
    }
    currentUrl = next.location;
  }
  throw new AnalysisError(
    "fetch_blocked",
    `Redirect-Limit von ${options.maxRedirects} überschritten.`,
    { lastUrl: currentUrl, maxRedirects: options.maxRedirects }
  );
}

type HopResult = { kind: "final"; text: string } | { kind: "redirect"; location: string };

async function fetchHop(rawUrl: string, options: LoadOptions, hop: number): Promise<HopResult> {
  const parsed = parseUrl(rawUrl);
  const urlDecision = validateUrl(parsed);
  if (!urlDecision.ok) {
    throw new AnalysisError(
      "fetch_blocked",
      `URL-Schutzregel verletzt: ${urlDecision.reason}.`,
      { hop, url: rawUrl, ...(urlDecision.detail ?? {}) }
    );
  }

  const candidates = await collectAddressCandidates(parsed.hostname, options.runtime, hop);
  if (candidates.length === 0) {
    throw new AnalysisError("fetch_blocked", "DNS-Auflösung lieferte keine Einträge.", {
      hop,
      host: parsed.hostname
    });
  }
  for (const entry of candidates) {
    const decision = validateResolvedIp(entry.address, entry.family);
    if (!decision.ok) {
      throw new AnalysisError(
        "fetch_blocked",
        `Aufgelöste IP-Adresse verletzt SSRF-Sperrliste: ${decision.reason}.`,
        { hop, host: parsed.hostname, ...(decision.detail ?? {}) }
      );
    }
  }

  const controller = new AbortController();
  let timedOut = false;
  const timer = setTimeout(() => {
    timedOut = true;
    controller.abort();
  }, options.timeoutMs);

  try {
    let response;
    try {
      response = await options.runtime.fetch(rawUrl, {
        signal: controller.signal,
        redirect: "manual",
        headers: { accept: "application/vnd.apple.mpegurl,application/x-mpegurl,audio/mpegurl,text/plain;q=0.9" }
      });
    } catch (error) {
      if (timedOut) {
        throw new AnalysisError("fetch_failed", `Timeout nach ${options.timeoutMs} ms.`, {
          hop,
          url: rawUrl,
          timeoutMs: options.timeoutMs
        });
      }
      throw new AnalysisError("fetch_failed", `Netzwerkfehler beim Laden: ${describeError(error)}.`, {
        hop,
        url: rawUrl
      });
    }

    if (response.status >= 300 && response.status < 400) {
      const location = response.headers.get("location");
      if (location === null || location === "") {
        throw new AnalysisError("fetch_failed", `Redirect ${response.status} ohne Location-Header.`, {
          hop,
          url: rawUrl,
          status: response.status
        });
      }
      return { kind: "redirect", location: new URL(location, rawUrl).toString() };
    }

    if (response.status < 200 || response.status >= 300) {
      throw new AnalysisError(
        "fetch_failed",
        `HTTP-Statuscode ${response.status} ist kein Erfolgsstatus.`,
        { hop, url: rawUrl, status: response.status }
      );
    }

    const contentType = (response.headers.get("content-type") ?? "").toLowerCase();
    const mainType = contentType.split(";")[0].trim();
    if (mainType !== "" && !ALLOWED_CONTENT_TYPES.has(mainType)) {
      throw new AnalysisError("fetch_failed", `Content-Type "${mainType}" ist kein HLS-Manifest.`, {
        hop,
        url: rawUrl,
        contentType: mainType
      });
    }

    const text = await readBody(response.body, options, rawUrl, hop, () => timedOut);
    return { kind: "final", text };
  } finally {
    clearTimeout(timer);
  }
}

async function collectAddressCandidates(
  hostname: string,
  runtime: LoaderRuntime,
  hop: number
): Promise<ResolvedAddress[]> {
  const literal = classifyHostnameLiteral(hostname);
  if (literal !== null) {
    return [literal];
  }
  return safeResolve(hostname, runtime, hop);
}

function classifyHostnameLiteral(hostname: string): ResolvedAddress | null {
  if (hostname.startsWith("[") && hostname.endsWith("]")) {
    const inner = hostname.slice(1, -1);
    if (parseIPv6(inner) !== null) {
      return { address: inner, family: 6 };
    }
    return null;
  }
  if (parseIPv4(hostname) !== null) {
    return { address: hostname, family: 4 };
  }
  return null;
}

function parseUrl(value: string): URL {
  try {
    return new URL(value);
  } catch {
    throw new AnalysisError("invalid_input", "URL ist nicht parseable.", { url: value });
  }
}

async function safeResolve(hostname: string, runtime: LoaderRuntime, hop: number): Promise<ResolvedAddress[]> {
  try {
    return await runtime.resolveHost(hostname);
  } catch (error) {
    throw new AnalysisError(
      "fetch_blocked",
      `DNS-Auflösung fehlgeschlagen: ${describeError(error)}.`,
      { hop, host: hostname }
    );
  }
}

async function readBody(
  body: AsyncIterable<Uint8Array> | null,
  options: LoadOptions,
  url: string,
  hop: number,
  timedOutRef: () => boolean
): Promise<string> {
  if (body === null) return "";
  let received = 0;
  const chunks: Uint8Array[] = [];
  try {
    for await (const chunk of body) {
      if (timedOutRef()) {
        throw new AnalysisError("fetch_failed", `Timeout nach ${options.timeoutMs} ms im Body-Stream.`, {
          hop,
          url,
          timeoutMs: options.timeoutMs
        });
      }
      received += chunk.byteLength;
      if (received > options.maxBytes) {
        throw new AnalysisError(
          "manifest_too_large",
          `Manifest überschreitet das Größenlimit von ${options.maxBytes} Bytes.`,
          { hop, url, maxBytes: options.maxBytes }
        );
      }
      chunks.push(chunk);
    }
  } catch (error) {
    if (error instanceof AnalysisError) throw error;
    if (timedOutRef()) {
      throw new AnalysisError("fetch_failed", `Timeout nach ${options.timeoutMs} ms im Body-Stream.`, {
        hop,
        url,
        timeoutMs: options.timeoutMs
      });
    }
    throw new AnalysisError("fetch_failed", `Body-Lesefehler: ${describeError(error)}.`, {
      hop,
      url
    });
  }
  const total = new Uint8Array(received);
  let offset = 0;
  for (const chunk of chunks) {
    total.set(chunk, offset);
    offset += chunk.byteLength;
  }
  return new TextDecoder("utf-8").decode(total);
}

function describeError(error: unknown): string {
  if (error instanceof Error) return error.message;
  return String(error);
}
