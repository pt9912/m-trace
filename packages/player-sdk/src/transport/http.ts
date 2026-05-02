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
   * Optionaler Provider für den W3C-`traceparent`-Header.
   * @see TraceParentProvider für Format, Synchronitäts- und
   *      Fehlersemantik.
   */
  traceparent?: TraceParentProvider;
  /**
   * Unterdrückt die einmalige `console.warn`-Diagnose, wenn der
   * Provider einen Non-String-Wert zurückgibt oder wirft. Default
   * `false`. Vorgesehen für Tests, die das Schluck-Verhalten
   * absichtlich auslösen, ohne die Test-Logs zu verschmutzen.
   */
  silent?: boolean;
}

export class HttpTransport implements Transport {
  private readonly fetchFn: FetchFn;
  private readonly sleep: SleepFn;
  private readonly maxAttempts: number;
  private readonly baseDelayMs: number;
  private readonly maxDelayMs: number;
  private readonly timeoutMs: number;
  private readonly traceparent?: TraceParentProvider;
  private readonly silent: boolean;
  private warnedTraceParentFailure = false;

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
    this.silent = options.silent ?? false;
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
    if (typeof tp === "string") {
      if (tp.length > 0) {
        headers.traceparent = tp;
      }
      // Empty string is a documented "no header" sentinel → no warn.
    } else if (tp !== undefined) {
      this.warnTraceParentOnce(`non-string return (${typeof tp})`);
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
   * Provider-Throws — Tracing darf den Event-Pfad nicht sabotieren.
   * Throws sind unwahrscheinlich, aber denkbar (z. B. wenn der
   * Provider auf einen nicht-initialisierten Tracer zugreift). Der
   * Header-Schreiber in `trySend()` prüft den Rückgabetyp defensiv
   * (`typeof tp === "string"`), damit ein versehentlich Non-String-
   * liefernder Provider nicht in den Header sickert. Beide Fehler-
   * pfade lösen einmal pro Transport-Instanz `warnTraceParentOnce`
   * aus (sofern `silent !== true`), damit Fehlkonfigurationen sicht-
   * bar werden, ohne den Send-Pfad zu blockieren.
   */
  private resolveTraceParent(): string | undefined {
    if (this.traceparent === undefined) {
      return undefined;
    }
    try {
      return this.traceparent();
    } catch (error) {
      this.warnTraceParentOnce(`provider threw: ${describeError(error)}`);
      return undefined;
    }
  }

  private warnTraceParentOnce(detail: string): void {
    if (this.silent || this.warnedTraceParentFailure) {
      return;
    }
    this.warnedTraceParentFailure = true;
    console.warn(`[m-trace] traceparent provider failed (${detail}); header omitted. Further failures on this transport will not be logged.`);
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

function describeError(error: unknown): string {
  if (error instanceof Error) {
    return `${error.name}: ${error.message}`;
  }
  return typeof error === "string" ? error : typeof error;
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
