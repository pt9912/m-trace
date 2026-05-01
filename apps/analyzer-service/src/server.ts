import { createServer as createHttpServer, type IncomingMessage, type Server, type ServerResponse } from "node:http";

import { analyzeHlsManifest } from "@npm9912/stream-analyzer";
import type {
  AnalyzeOptions,
  AnalyzeOutput,
  ManifestInput
} from "@npm9912/stream-analyzer";

/**
 * Maximaler Request-Body in Bytes. Schützt den Service gegen einfache
 * Resource-Exhaustion-Versuche; der eigentliche Manifest-Loader hat
 * weiter unten seine eigenen Limits (FetchOptions.maxBytes).
 */
const MAX_REQUEST_BODY_BYTES = 1_000_000;

export interface AnalyzerServerOptions {
  /**
   * Optional injizierter Analyse-Aufruf. Default ruft die Public
   * `analyzeHlsManifest`-Funktion. Tests injizieren einen Stub.
   */
  readonly analyze?: (input: ManifestInput, options?: AnalyzeOptions) => Promise<AnalyzeOutput>;
  /**
   * Wenn `true`, lockert der Loader die SSRF-IP-Sperrlisten (loopback,
   * private, link-local). `main.ts` liest das aus
   * `ANALYZER_ALLOW_PRIVATE_NETWORKS` und reicht das Flag durch.
   */
  readonly allowPrivateNetworks?: boolean;
}

export function createAnalyzerServer(options: AnalyzerServerOptions = {}): Server {
  const analyze = options.analyze ?? analyzeHlsManifest;
  const allowPrivateNetworks = options.allowPrivateNetworks === true;

  return createHttpServer((req, res) => {
    handleRequest(req, res, analyze, allowPrivateNetworks).catch((error) => {
      writeProblem(res, 500, "internal_error", `Unbehandelter Fehler: ${describeError(error)}`);
    });
  });
}

async function handleRequest(
  req: IncomingMessage,
  res: ServerResponse,
  analyze: (input: ManifestInput, options?: AnalyzeOptions) => Promise<AnalyzeOutput>,
  allowPrivateNetworks: boolean
): Promise<void> {
  const url = req.url ?? "/";
  if (req.method === "GET" && url === "/health") {
    writeJson(res, 200, { status: "ok" });
    return;
  }
  if (req.method === "POST" && url === "/analyze") {
    await handleAnalyze(req, res, analyze, allowPrivateNetworks);
    return;
  }
  writeProblem(res, 404, "not_found", `${req.method ?? "?"} ${url} ist nicht definiert.`);
}

async function handleAnalyze(
  req: IncomingMessage,
  res: ServerResponse,
  analyze: (input: ManifestInput, options?: AnalyzeOptions) => Promise<AnalyzeOutput>,
  allowPrivateNetworks: boolean
): Promise<void> {
  const contentType = (req.headers["content-type"] ?? "").toLowerCase();
  const mainType = contentType.split(";")[0].trim();
  if (mainType !== "application/json") {
    writeProblem(res, 415, "unsupported_media_type", "Content-Type muss application/json sein.");
    return;
  }
  const body = await readBody(req);
  if (body === null) {
    writeProblem(res, 413, "payload_too_large", `Request-Body überschreitet ${MAX_REQUEST_BODY_BYTES} Bytes.`);
    return;
  }
  let parsed: unknown;
  try {
    parsed = JSON.parse(body.text);
  } catch (error) {
    writeProblem(res, 400, "invalid_json", `Body ist kein gültiges JSON: ${describeError(error)}`);
    return;
  }
  const input = parseManifestInput(parsed);
  if (input === null) {
    writeProblem(
      res,
      400,
      "invalid_request",
      "Body muss { kind: \"text\", text, baseUrl? } oder { kind: \"url\", url } sein."
    );
    return;
  }
  const fetchOptions = mergeFetchOptions(parseFetchOptions(parsed), allowPrivateNetworks);
  const result = await analyze(input, fetchOptions ? { fetch: fetchOptions } : undefined);
  writeJson(res, 200, result);
}

function mergeFetchOptions(
  fromBody: NonNullable<AnalyzeOptions["fetch"]> | null,
  allowPrivateNetworks: boolean
): NonNullable<AnalyzeOptions["fetch"]> | null {
  // Service-seitige Policy hat Vorrang vor Body-Werten — wenn der
  // Operator das Flag in der Compose-/Container-Config gesetzt hat,
  // gilt es für alle Aufrufe an diesen Service-Prozess. Andernfalls
  // greifen nur die Body-Werte (oder Default-Verhalten).
  if (allowPrivateNetworks) {
    return { ...(fromBody ?? {}), allowPrivateNetworks: true };
  }
  return fromBody;
}

interface ReadResult {
  readonly text: string;
}

async function readBody(req: IncomingMessage): Promise<ReadResult | null> {
  const chunks: Buffer[] = [];
  let received = 0;
  for await (const chunk of req) {
    const buf = chunk instanceof Buffer ? chunk : Buffer.from(chunk);
    received += buf.byteLength;
    if (received > MAX_REQUEST_BODY_BYTES) {
      return null;
    }
    chunks.push(buf);
  }
  return { text: Buffer.concat(chunks).toString("utf8") };
}

function parseManifestInput(value: unknown): ManifestInput | null {
  if (typeof value !== "object" || value === null) return null;
  const v = value as Record<string, unknown>;
  if (v.kind === "text") {
    if (typeof v.text !== "string") return null;
    if (v.baseUrl !== undefined && typeof v.baseUrl !== "string") return null;
    return v.baseUrl !== undefined
      ? { kind: "text", text: v.text, baseUrl: v.baseUrl }
      : { kind: "text", text: v.text };
  }
  if (v.kind === "url") {
    if (typeof v.url !== "string" || v.url.length === 0) return null;
    return { kind: "url", url: v.url };
  }
  return null;
}

function parseFetchOptions(value: unknown): NonNullable<AnalyzeOptions["fetch"]> | null {
  if (typeof value !== "object" || value === null) return null;
  const v = value as Record<string, unknown>;
  const fetch = v.fetch;
  if (typeof fetch !== "object" || fetch === null) return null;
  const f = fetch as Record<string, unknown>;
  const out: { timeoutMs?: number; maxBytes?: number; maxRedirects?: number } = {};
  if (typeof f.timeoutMs === "number" && Number.isFinite(f.timeoutMs) && f.timeoutMs > 0) {
    out.timeoutMs = f.timeoutMs;
  }
  if (typeof f.maxBytes === "number" && Number.isFinite(f.maxBytes) && f.maxBytes > 0) {
    out.maxBytes = f.maxBytes;
  }
  if (typeof f.maxRedirects === "number" && Number.isFinite(f.maxRedirects) && f.maxRedirects >= 0) {
    out.maxRedirects = f.maxRedirects;
  }
  if (Object.keys(out).length === 0) return null;
  return out;
}

function writeJson(res: ServerResponse, status: number, payload: unknown): void {
  const body = JSON.stringify(payload);
  res.writeHead(status, {
    "Content-Type": "application/json; charset=utf-8",
    "Content-Length": Buffer.byteLength(body)
  });
  res.end(body);
}

function writeProblem(res: ServerResponse, status: number, code: string, message: string): void {
  writeJson(res, status, { status: "error", code, message });
}

function describeError(error: unknown): string {
  if (error instanceof Error) return error.message;
  return String(error);
}
