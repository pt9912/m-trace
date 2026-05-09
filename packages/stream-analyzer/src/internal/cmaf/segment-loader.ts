import { AnalysisError } from "../../types/error.js";
import type { CmafFailureCode } from "../../types/result.js";
import { parseIPv4, parseIPv6, validateResolvedIp, validateUrl } from "../loader/ssrf.js";
import type { LoaderRuntime, ResolvedAddress } from "../loader/runtime.js";

/**
 * Bounded binary segment loader für die binäre CMAF-
 * Konformitätsprüfung (`0.10.0` Tranche 4, NF-13 / RAK-64).
 *
 * Bewusst getrennt vom Manifest-Loader (`internal/loader/fetch.ts`):
 *
 *  - liefert `Uint8Array`, kein UTF-8-Text;
 *  - akzeptiert MP4-/Byte-Content-Types (`video/mp4`, `audio/mp4`,
 *    `application/mp4`, `video/iso.segment`, `audio/iso.segment`,
 *    `application/iso.segment`, `application/octet-stream` und leeren
 *    Content-Type);
 *  - sendet einen MP4-orientierten `Accept`-Header;
 *  - erzwingt `maxSegmentBytes` pro Segment statt das Manifest-Body-
 *    Limit zu recyceln;
 *  - mappt Fetch-/Read-Fehler auf strukturierte
 *    `SegmentLoaderError`-Ergebnisse mit `CmafFailureCode`, statt
 *    in einen Top-Level-Analysefehler zu eskalieren.
 *
 * SSRF-/DNS-/Redirect-/Timeout-Mechanik werden eins zu eins vom
 * Manifest-Loader übernommen (gleiche `LoaderRuntime`,
 * gleiche `validateUrl`/`validateResolvedIp`-Pfade), damit eine
 * Lockerung wie `allowPrivateNetworks` für beide Loader synchron
 * gilt.
 */

const ALLOWED_SEGMENT_CONTENT_TYPES: ReadonlySet<string> = new Set([
  "video/mp4",
  "audio/mp4",
  "application/mp4",
  "video/iso.segment",
  "audio/iso.segment",
  "application/iso.segment",
  "application/octet-stream"
]);

const SEGMENT_ACCEPT_HEADER =
  "video/mp4,audio/mp4,application/mp4," +
  "video/iso.segment,audio/iso.segment,application/iso.segment," +
  "application/octet-stream;q=0.9";

export interface SegmentLoadOptions {
  readonly runtime: LoaderRuntime;
  readonly timeoutMs: number;
  readonly maxSegmentBytes: number;
  readonly maxRedirects: number;
  readonly allowPrivateNetworks: boolean;
}

export interface SegmentLoadOk {
  readonly ok: true;
  readonly bytes: Uint8Array;
  readonly contentType: string;
  readonly finalUrl: string;
}

export interface SegmentLoadFailed {
  readonly ok: false;
  readonly code: CmafFailureCode;
  readonly message: string;
}

export type SegmentLoadResult = SegmentLoadOk | SegmentLoadFailed;

/**
 * Lädt ein einzelnes Segment unter denselben Schutzregeln wie
 * `loadManifest`, aber mit Bytes-Return und MP4-Content-Type-
 * Allowlist. Liefert ein strukturiertes Result statt eine Exception
 * für domänenrelevante Fehler — damit der Verifier sie als
 * `segmentsChecked[]`-Einträge mit Failure-Code abbilden kann.
 *
 * SSRF-/DNS-Fehler aus den geteilten Validators werden auf
 * `segment_uri_blocked` gemappt; Timeouts auf `segment_fetch_failed`;
 * inkompatibler Content-Type auf `segment_content_type_unsupported`;
 * Größenlimit auf `segment_too_large`.
 */
export async function loadSegment(
  url: string,
  options: SegmentLoadOptions
): Promise<SegmentLoadResult> {
  let currentUrl = url;
  for (let hop = 0; hop <= options.maxRedirects; hop++) {
    let next: HopResult;
    try {
      next = await fetchHop(currentUrl, options, hop);
    } catch (error) {
      if (error instanceof SegmentFetchError) {
        return { ok: false, code: error.code, message: error.message };
      }
      throw error;
    }
    if (next.kind === "final") {
      return {
        ok: true,
        bytes: next.bytes,
        contentType: next.contentType,
        finalUrl: currentUrl
      };
    }
    currentUrl = next.location;
  }
  return {
    ok: false,
    code: "segment_uri_blocked",
    message: `Redirect-Limit von ${options.maxRedirects} überschritten.`
  };
}

class SegmentFetchError extends Error {
  constructor(
    readonly code: CmafFailureCode,
    message: string
  ) {
    super(message);
  }
}

type HopResult =
  | { kind: "final"; bytes: Uint8Array; contentType: string }
  | { kind: "redirect"; location: string };

async function fetchHop(rawUrl: string, options: SegmentLoadOptions, hop: number): Promise<HopResult> {
  await validateAndResolveTarget(rawUrl, options, hop);
  const controller = new AbortController();
  const timedOutBox = { value: false };
  const timer = setTimeout(() => {
    timedOutBox.value = true;
    controller.abort();
  }, options.timeoutMs);
  try {
    const response = await executeFetch(rawUrl, options, controller, timedOutBox);
    return await dispatchResponse(response, rawUrl, options, () => timedOutBox.value);
  } finally {
    clearTimeout(timer);
  }
}

async function validateAndResolveTarget(
  rawUrl: string,
  options: SegmentLoadOptions,
  hop: number
): Promise<void> {
  const parsed = tryParseUrl(rawUrl);
  if (parsed === null) {
    throw new SegmentFetchError(
      "segment_uri_blocked",
      `Segment-URL ist nicht parseable: ${rawUrl}.`
    );
  }
  const urlDecision = validateUrl(parsed);
  if (!urlDecision.ok) {
    throw new SegmentFetchError(
      "segment_uri_blocked",
      `Segment-URL-Schutzregel verletzt: ${urlDecision.reason}.`
    );
  }
  const candidates = await collectAddressCandidates(parsed.hostname, options.runtime);
  if (candidates.length === 0) {
    throw new SegmentFetchError(
      "segment_uri_blocked",
      `DNS-Auflösung lieferte keine Einträge für ${parsed.hostname}.`
    );
  }
  if (options.allowPrivateNetworks) return;
  for (const entry of candidates) {
    const decision = validateResolvedIp(entry.address, entry.family);
    if (!decision.ok) {
      throw new SegmentFetchError(
        "segment_uri_blocked",
        `Aufgelöste IP für ${parsed.hostname} verletzt SSRF-Sperrliste: ${decision.reason} (hop ${hop}).`
      );
    }
  }
}

async function executeFetch(
  rawUrl: string,
  options: SegmentLoadOptions,
  controller: AbortController,
  timedOutBox: { value: boolean }
): Promise<Awaited<ReturnType<LoaderRuntime["fetch"]>>> {
  try {
    return await options.runtime.fetch(rawUrl, {
      signal: controller.signal,
      redirect: "manual",
      headers: { accept: SEGMENT_ACCEPT_HEADER }
    });
  } catch (error) {
    if (timedOutBox.value) {
      throw new SegmentFetchError(
        "segment_fetch_failed",
        `Segment-Fetch-Timeout nach ${options.timeoutMs} ms.`
      );
    }
    throw new SegmentFetchError(
      "segment_fetch_failed",
      `Segment-Fetch-Netzwerkfehler: ${describeError(error)}.`
    );
  }
}

async function dispatchResponse(
  response: Awaited<ReturnType<LoaderRuntime["fetch"]>>,
  rawUrl: string,
  options: SegmentLoadOptions,
  timedOutFn: () => boolean
): Promise<HopResult> {
  if (response.status >= 300 && response.status < 400) {
    const location = response.headers.get("location");
    if (location === null || location === "") {
      throw new SegmentFetchError(
        "segment_fetch_failed",
        `Redirect ${response.status} ohne Location-Header.`
      );
    }
    return { kind: "redirect", location: new URL(location, rawUrl).toString() };
  }
  if (response.status < 200 || response.status >= 300) {
    throw new SegmentFetchError(
      "segment_fetch_failed",
      `HTTP-Statuscode ${response.status} ist kein Erfolgsstatus.`
    );
  }
  const contentType = (response.headers.get("content-type") ?? "").toLowerCase();
  const mainType = contentType.split(";")[0].trim();
  if (mainType !== "" && !ALLOWED_SEGMENT_CONTENT_TYPES.has(mainType)) {
    throw new SegmentFetchError(
      "segment_content_type_unsupported",
      `Segment-Content-Type "${mainType}" ist kein MP4-/Byte-kompatibler Type.`
    );
  }
  const bytes = await readBody(response.body, options, timedOutFn);
  return { kind: "final", bytes, contentType: mainType };
}

async function collectAddressCandidates(
  hostname: string,
  runtime: LoaderRuntime
): Promise<ResolvedAddress[]> {
  const literal = classifyHostnameLiteral(hostname);
  if (literal !== null) return [literal];
  return safeResolve(hostname, runtime);
}

function classifyHostnameLiteral(hostname: string): ResolvedAddress | null {
  if (hostname.startsWith("[") && hostname.endsWith("]")) {
    const inner = hostname.slice(1, -1);
    if (parseIPv6(inner) !== null) return { address: inner, family: 6 };
    return null;
  }
  if (parseIPv4(hostname) !== null) return { address: hostname, family: 4 };
  return null;
}

async function safeResolve(hostname: string, runtime: LoaderRuntime): Promise<ResolvedAddress[]> {
  try {
    return await runtime.resolveHost(hostname);
  } catch (error) {
    throw new SegmentFetchError(
      "segment_uri_blocked",
      `DNS-Auflösung fehlgeschlagen: ${describeError(error)}.`
    );
  }
}

async function readBody(
  body: AsyncIterable<Uint8Array> | null,
  options: SegmentLoadOptions,
  timedOutRef: () => boolean
): Promise<Uint8Array> {
  if (body === null) return new Uint8Array(0);
  let received = 0;
  const chunks: Uint8Array[] = [];
  try {
    for await (const chunk of body) {
      if (timedOutRef()) {
        throw new SegmentFetchError(
          "segment_fetch_failed",
          `Segment-Body-Timeout nach ${options.timeoutMs} ms.`
        );
      }
      received += chunk.byteLength;
      if (received > options.maxSegmentBytes) {
        throw new SegmentFetchError(
          "segment_too_large",
          `Segment überschreitet maxSegmentBytes=${options.maxSegmentBytes}.`
        );
      }
      chunks.push(chunk);
    }
  } catch (error) {
    if (error instanceof SegmentFetchError) throw error;
    if (error instanceof AnalysisError) throw error;
    if (timedOutRef()) {
      throw new SegmentFetchError(
        "segment_fetch_failed",
        `Segment-Body-Timeout nach ${options.timeoutMs} ms.`
      );
    }
    throw new SegmentFetchError(
      "segment_fetch_failed",
      `Segment-Body-Lesefehler: ${describeError(error)}.`
    );
  }
  const total = new Uint8Array(received);
  let offset = 0;
  for (const chunk of chunks) {
    total.set(chunk, offset);
    offset += chunk.byteLength;
  }
  return total;
}

function tryParseUrl(value: string): URL | null {
  try {
    return new URL(value);
  } catch {
    return null;
  }
}

function describeError(error: unknown): string {
  if (error instanceof Error) return error.message;
  return String(error);
}
