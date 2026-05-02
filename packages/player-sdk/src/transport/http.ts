import type { TraceParentProvider, Transport } from "../types/config";
import type { PlaybackEventBatch } from "../types/events";

type FetchFn = (input: string, init: RequestInit) => Promise<Response>;
type SleepFn = (ms: number) => Promise<void>;

export interface HttpTransportOptions {
  fetchFn?: FetchFn;
  sleep?: SleepFn;
  maxAttempts?: number;
  baseDelayMs?: number;
  maxDelayMs?: number;
  timeoutMs?: number;
  /**
   * Optionaler Provider für den W3C-`traceparent`-Header (siehe
   * `PlayerSDKConfig.traceparent`). Wenn gesetzt und der Aufruf
   * nicht-leer zurückgibt, sendet `send()` den Header zusätzlich
   * zu `X-MTrace-Token`. Provider-Fehler (Throw) werden gefangen
   * und still verworfen — Tracing darf den Event-Pfad nicht
   * sabotieren.
   */
  traceparent?: TraceParentProvider;
}

export class HttpTransport implements Transport {
  private readonly fetchFn: FetchFn;
  private readonly sleep: SleepFn;
  private readonly maxAttempts: number;
  private readonly baseDelayMs: number;
  private readonly maxDelayMs: number;
  private readonly timeoutMs: number;
  private readonly traceparent?: TraceParentProvider;

  constructor(
    private readonly endpoint: string,
    private readonly token: string,
    options: HttpTransportOptions = {}
  ) {
    this.fetchFn = options.fetchFn ?? fetch;
    this.sleep = options.sleep ?? ((ms) => new Promise((resolve) => setTimeout(resolve, ms)));
    this.maxAttempts = Math.max(1, Math.floor(options.maxAttempts ?? 3));
    this.baseDelayMs = Math.max(0, Math.floor(options.baseDelayMs ?? 250));
    this.maxDelayMs = Math.max(this.baseDelayMs, Math.floor(options.maxDelayMs ?? 5000));
    this.timeoutMs = Math.max(0, Math.floor(options.timeoutMs ?? 10000));
    this.traceparent = options.traceparent;
  }

  async send(batch: PlaybackEventBatch): Promise<void> {
    const body = JSON.stringify(batch);
    let lastError: unknown;

    for (let attempt = 1; attempt <= this.maxAttempts; attempt += 1) {
      const response = await this.trySend(body).catch((error: unknown) => {
        lastError = error;
        return undefined;
      });

      if (response === undefined) {
        if (attempt >= this.maxAttempts) {
          break;
        }
        await this.sleep(this.backoffMs(attempt));
        continue;
      }

      if (response.ok) {
        return;
      }

      if (!isRetryableStatus(response.status) || attempt >= this.maxAttempts) {
        throw new Error(`m-trace transport failed: ${response.status}`);
      }

      await this.sleep(this.retryDelayMs(response, attempt));
    }

    throw lastError instanceof Error ? lastError : new Error("m-trace transport failed");
  }

  private async trySend(body: string): Promise<Response> {
    const controller = this.timeoutMs > 0 ? new AbortController() : undefined;
    const timeout =
      controller !== undefined
        ? setTimeout(() => {
            controller.abort();
          }, this.timeoutMs)
        : undefined;

    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      "X-MTrace-Token": this.token
    };
    const tp = this.resolveTraceParent();
    if (tp !== undefined && tp !== "") {
      headers.traceparent = tp;
    }

    try {
      return await this.fetchFn(this.endpoint, {
        method: "POST",
        credentials: "omit",
        headers,
        body,
        signal: controller?.signal
      });
    } finally {
      if (timeout !== undefined) {
        clearTimeout(timeout);
      }
    }
  }

  /**
   * resolveTraceParent ruft den optionalen Provider auf und fängt
   * Provider-Throws still — Tracing darf den Event-Pfad nicht
   * sabotieren. Throws sind sehr unwahrscheinlich, aber denkbar
   * (z. B. wenn der Provider auf einen nicht-initialisierten Tracer
   * zugreift und dabei stolpert).
   */
  private resolveTraceParent(): string | undefined {
    if (this.traceparent === undefined) {
      return undefined;
    }
    try {
      return this.traceparent();
    } catch {
      return undefined;
    }
  }

  private retryDelayMs(response: Response, attempt: number): number {
    if (response.status === 429) {
      const retryAfterMs = parseRetryAfterMs(response.headers.get("Retry-After"));
      if (retryAfterMs !== undefined) {
        return retryAfterMs;
      }
    }
    return this.backoffMs(attempt);
  }

  private backoffMs(attempt: number): number {
    return Math.min(this.maxDelayMs, this.baseDelayMs * 2 ** Math.max(0, attempt - 1));
  }
}

function isRetryableStatus(status: number): boolean {
  return status === 429 || status >= 500;
}

function parseRetryAfterMs(value: string | null): number | undefined {
  if (value === null || value.trim() === "") {
    return undefined;
  }

  const seconds = Number(value);
  if (Number.isFinite(seconds) && seconds >= 0) {
    return seconds * 1000;
  }

  const timestamp = Date.parse(value);
  if (Number.isNaN(timestamp)) {
    return undefined;
  }

  return Math.max(0, timestamp - Date.now());
}
