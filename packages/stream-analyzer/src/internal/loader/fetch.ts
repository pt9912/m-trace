import { AnalysisError } from "../../types/error.js";
import type { LoaderRuntime } from "./runtime.js";
import { validateResolvedIp, validateUrl } from "./ssrf.js";

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
 *   2. Host auflösen, jeden zurückgegebenen Eintrag gegen die SSRF-
 *      Sperrlisten prüfen. Schon ein blockierter Eintrag bricht ab,
 *      damit Mehrfach-A/AAAA-Antworten keine Lücken aufmachen.
 *   3. fetch mit AbortController-Timeout und manuellem Redirect.
 *   4. Bei 3xx: Location interpretieren, Hop-Zähler prüfen, von vorn.
 *   5. Bei 2xx: Content-Type prüfen, Body bis maxBytes streamen,
 *      andernfalls AnalysisError(manifest_too_large).
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

  const resolved = await safeResolve(parsed.hostname, options.runtime);
  if (resolved.length === 0) {
    throw new AnalysisError(
      "fetch_blocked",
      "DNS-Auflösung lieferte keine Einträge.",
      { hop, host: parsed.hostname }
    );
  }
  for (const entry of resolved) {
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
  const timer = setTimeout(() => controller.abort(), options.timeoutMs);
  let response;
  try {
    response = await options.runtime.fetch(rawUrl, {
      signal: controller.signal,
      redirect: "manual",
      headers: { accept: "application/vnd.apple.mpegurl,application/x-mpegurl,audio/mpegurl,text/plain;q=0.9" }
    });
  } catch (error) {
    clearTimeout(timer);
    if (controller.signal.aborted) {
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
    clearTimeout(timer);
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
    clearTimeout(timer);
    throw new AnalysisError(
      "fetch_failed",
      `HTTP-Statuscode ${response.status} ist kein Erfolgsstatus.`,
      { hop, url: rawUrl, status: response.status }
    );
  }

  const contentType = (response.headers.get("content-type") ?? "").toLowerCase();
  const mainType = contentType.split(";")[0].trim();
  if (mainType !== "" && !ALLOWED_CONTENT_TYPES.has(mainType)) {
    clearTimeout(timer);
    throw new AnalysisError("fetch_failed", `Content-Type "${mainType}" ist kein HLS-Manifest.`, {
      hop,
      url: rawUrl,
      contentType: mainType
    });
  }

  try {
    const text = await readBody(response.body, options.maxBytes, rawUrl, hop);
    return { kind: "final", text };
  } finally {
    clearTimeout(timer);
  }
}

function parseUrl(value: string): URL {
  try {
    return new URL(value);
  } catch {
    throw new AnalysisError("invalid_input", "URL ist nicht parseable.", { url: value });
  }
}

async function safeResolve(hostname: string, runtime: LoaderRuntime) {
  try {
    return await runtime.resolveHost(hostname);
  } catch (error) {
    throw new AnalysisError(
      "fetch_blocked",
      `DNS-Auflösung fehlgeschlagen: ${describeError(error)}.`,
      { host: hostname }
    );
  }
}

async function readBody(
  body: AsyncIterable<Uint8Array> | null,
  maxBytes: number,
  url: string,
  hop: number
): Promise<string> {
  if (body === null) return "";
  let received = 0;
  const chunks: Uint8Array[] = [];
  for await (const chunk of body) {
    received += chunk.byteLength;
    if (received > maxBytes) {
      throw new AnalysisError(
        "manifest_too_large",
        `Manifest überschreitet das Größenlimit von ${maxBytes} Bytes.`,
        { hop, url, maxBytes }
      );
    }
    chunks.push(chunk);
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
