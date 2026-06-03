import { AnalysisError } from "../../types/error.js";
import type { CmafFailureCode } from "../../types/result.js";
import { parseIPv4, parseIPv6, validateResolvedIp, validateUrl } from "../loader/ssrf.js";
import type { LoaderRuntime, ResolvedAddress } from "../loader/runtime.js";

/**
 * Bounded binary segment loader für die binäre CMAF-
 * Konformitätsprüfung (NF-13 / RAK-64).
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
  readonly range?: SegmentByteRange;
}

export interface SegmentByteRange {
  readonly offset: number;
  readonly length: number;
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
  const rangeGuard = validateRangeOptions(options.range, options.maxSegmentBytes);
  if (rangeGuard !== null) return rangeGuard;
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

function validateRangeOptions(
  range: SegmentByteRange | undefined,
  maxSegmentBytes: number
): SegmentLoadFailed | null {
  if (range === undefined) return null;
  if (
    !Number.isSafeInteger(range.offset) ||
    !Number.isSafeInteger(range.length) ||
    range.offset < 0 ||
    range.length <= 0 ||
    range.offset > Number.MAX_SAFE_INTEGER - range.length
  ) {
    return {
      ok: false,
      code: "segment_fetch_failed",
      message: "Ungueltiger HTTP-Range-Scope."
    };
  }
  if (range.length > maxSegmentBytes) {
    return {
      ok: false,
      code: "segment_too_large",
      message: `Range-Laenge ${range.length} ueberschreitet maxSegmentBytes=${maxSegmentBytes}.`
    };
  }
  return null;
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
  const headers: Record<string, string> = { accept: SEGMENT_ACCEPT_HEADER };
  if (options.range !== undefined) {
    const end = options.range.offset + options.range.length - 1;
    headers.range = `bytes=${options.range.offset}-${end}`;
  }
  try {
    return await options.runtime.fetch(rawUrl, {
      signal: controller.signal,
      redirect: "manual",
      headers
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
  if (options.range !== undefined && response.status !== 206) {
    throw new SegmentFetchError(
      "segment_fetch_failed",
      `Range-Request erwartete 206 Partial Content, erhielt aber ${response.status}.`
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
  const expectedBytes = options.range?.length;
  if (body === null) {
    return emptyBodyResult(expectedBytes);
  }
  let received = 0;
  const chunks: Uint8Array[] = [];
  const maxAllowedBytes = expectedBytes ?? options.maxSegmentBytes;
  try {
    for await (const chunk of body) {
      assertBodyReadNotTimedOut(timedOutRef, options.timeoutMs);
      received = appendBodyChunk(chunks, chunk, received, maxAllowedBytes, options);
    }
  } catch (error) {
    throw mapBodyReadError(error, timedOutRef, options.timeoutMs);
  }
  assertExpectedRangeBytes(received, expectedBytes);
  return concatChunks(chunks, received);
}

function emptyBodyResult(expectedBytes: number | undefined): Uint8Array {
  if (expectedBytes === undefined) return new Uint8Array(0);
  throw new SegmentFetchError(
    "segment_fetch_failed",
    `Range-Response lieferte keinen Body, erwartet waren ${expectedBytes} Bytes.`
  );
}

function assertBodyReadNotTimedOut(timedOutRef: () => boolean, timeoutMs: number): void {
  if (timedOutRef()) {
    throw new SegmentFetchError(
      "segment_fetch_failed",
      `Segment-Body-Timeout nach ${timeoutMs} ms.`
    );
  }
}

function appendBodyChunk(
  chunks: Uint8Array[],
  chunk: Uint8Array,
  received: number,
  maxAllowedBytes: number,
  options: SegmentLoadOptions
): number {
  const nextReceived = received + chunk.byteLength;
  if (nextReceived > maxAllowedBytes) {
    throw new SegmentFetchError(
      "segment_too_large",
      options.range !== undefined
        ? `Range-Response überschreitet erwartete Laenge ${options.range.length}.`
        : `Segment überschreitet maxSegmentBytes=${options.maxSegmentBytes}.`
    );
  }
  chunks.push(chunk);
  return nextReceived;
}

function mapBodyReadError(
  error: unknown,
  timedOutRef: () => boolean,
  timeoutMs: number
): SegmentFetchError | AnalysisError {
  if (error instanceof SegmentFetchError) return error;
  if (error instanceof AnalysisError) return error;
  if (timedOutRef()) {
    return new SegmentFetchError(
      "segment_fetch_failed",
      `Segment-Body-Timeout nach ${timeoutMs} ms.`
    );
  }
  return new SegmentFetchError(
    "segment_fetch_failed",
    `Segment-Body-Lesefehler: ${describeError(error)}.`
  );
}

function assertExpectedRangeBytes(received: number, expectedBytes: number | undefined): void {
  if (expectedBytes !== undefined && received !== expectedBytes) {
    throw new SegmentFetchError(
      "segment_fetch_failed",
      `Range-Response lieferte ${received} Bytes, erwartet waren ${expectedBytes}.`
    );
  }
}

function concatChunks(chunks: readonly Uint8Array[], received: number): Uint8Array {
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
