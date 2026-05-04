import { sseConnection, type SseConnectionState } from "./status";

/**
 * Fetch-basierter SSE-Client für `GET /api/stream-sessions/stream`.
 *
 * Warum fetch statt nativem `EventSource`: ab plan-0.4.0 §5 ist der
 * Stream tokenpflichtig (`X-MTrace-Token`-Header). Das native
 * `EventSource`-API erlaubt keine Custom-Header — wir lesen das
 * Response-Body manuell als `text/event-stream` und parsen die
 * Frames selbst (Spec `spec/backend-api-contract.md` §10a).
 *
 * Frame-Format (Server-seitig):
 *
 *   id: <ingest_sequence>
 *   event: event_appended | backfill_truncated
 *   data: <JSON>
 *
 * Plus optionale `: heartbeat`-Comment-Frames (Idle-Keepalive). Wir
 * folgen dem WHATWG-Server-Sent-Events-Parsing-Algorithmus minimal:
 * Felder werden aufgesammelt, ein leerer Zeilen-Break dispatched
 * den Frame.
 */

export interface AppendedFrame {
  project_id: string;
  session_id: string;
  ingest_sequence: number;
  event_name: string;
}

export interface TruncatedFrame {
  oldest_ingest_sequence: number;
}

export interface SseClientOptions {
  /** Voller URL des Stream-Endpunkts. */
  url: string;
  /** `X-MTrace-Token`-Wert; ohne Token wirft der Server 401 und der
   *  Client wechselt in den Polling-Fallback. */
  token: string;
  /** Callback bei jedem `event_appended`-Frame. */
  onAppended: (frame: AppendedFrame) => void;
  /** Callback bei `backfill_truncated`-Frame. */
  onTruncated?: (frame: TruncatedFrame) => void;
  /** Polling-Fallback-Callback: wenn SSE persistent ausfällt, ruft
   *  der Client das in regelmäßigen Abständen auf. Defaultet auf
   *  no-op; Konsumenten können hier `listSessions`-Polling
   *  einhängen. */
  onPollingTick?: () => void | Promise<void>;
  /** Optional, defaults: `5000` ms Reconnect-Backoff (initial), max
   *  `30000` ms; `5000` ms Polling-Intervall. */
  reconnectBaseMs?: number;
  reconnectMaxMs?: number;
  pollingIntervalMs?: number;
  /** Test-Hook: per Default `fetch`. Eigene Signatur (ohne den
   *  vollen `URL | RequestInfo`-Union der globalen `typeof fetch`-
   *  Definition), damit Vitest-Mocks ohne `as unknown`-Cast
   *  zuweisbar sind. */
  fetchFn?: (url: string, init?: RequestInit) => Promise<Response>;
  /** Test-Hook: per Default `setTimeout`/`clearTimeout`. */
  schedule?: (cb: () => void, ms: number) => () => void;
}

export interface SseClient {
  /** Schließt Stream + Polling und kippt den `sseConnection`-Store
   *  auf `disabled`. */
  disconnect(): void;
}

const DEFAULT_RECONNECT_BASE_MS = 5_000;
const DEFAULT_RECONNECT_MAX_MS = 30_000;
const DEFAULT_POLLING_INTERVAL_MS = 5_000;
const POLLING_FALLBACK_THRESHOLD = 3;

/**
 * Startet einen SSE-Client. Konsumenten erhalten einen Disconnect-
 * Handle. Der Store `sseConnection` wird vom Client gepflegt:
 * `connecting` beim Verbindungsaufbau, `connected` nach dem ersten
 * Byte, `polling_fallback` bei persistenten SSE-Fehlern,
 * `disabled` bei expliziter Disconnect-Anforderung.
 */
export function startSseClient(opts: SseClientOptions): SseClient {
  const machine = new SseMachine(opts);
  void machine.connect();
  return {
    disconnect() {
      machine.disconnect();
    }
  };
}

class SseMachine {
  private readonly opts: SseClientOptions;
  private readonly fetchFn: NonNullable<SseClientOptions["fetchFn"]>;
  private readonly schedule: NonNullable<SseClientOptions["schedule"]>;
  private readonly reconnectBaseMs: number;
  private readonly reconnectMaxMs: number;
  private readonly pollingIntervalMs: number;
  private closed = false;
  private abort: AbortController | undefined;
  private pollingHandle: (() => void) | undefined;
  private pollingActive = false;
  private reconnectAttempt = 0;
  private lastEventID = "";

  constructor(opts: SseClientOptions) {
    this.opts = opts;
    this.fetchFn = opts.fetchFn ?? fetch;
    this.schedule = opts.schedule ?? defaultSchedule;
    this.reconnectBaseMs = opts.reconnectBaseMs ?? DEFAULT_RECONNECT_BASE_MS;
    this.reconnectMaxMs = opts.reconnectMaxMs ?? DEFAULT_RECONNECT_MAX_MS;
    this.pollingIntervalMs = opts.pollingIntervalMs ?? DEFAULT_POLLING_INTERVAL_MS;
  }

  disconnect(): void {
    this.closed = true;
    this.abort?.abort();
    this.stopPolling();
    this.setState("disabled");
  }

  async connect(): Promise<void> {
    if (this.closed) return;
    const response = await this.openStream();
    if (!response) return;
    this.setState("connected");
    this.reconnectAttempt = 0;
    this.stopPolling();
    if (response.body) {
      await this.drainStream(response.body);
    }
    if (!this.closed) {
      this.handleConnectionError("stream closed by server");
    }
  }

  private setState(newState: SseConnectionState, detail: string | null = null): void {
    sseConnection.set({ state: newState, detail, changedAt: new Date().toISOString() });
  }

  private startPolling(): void {
    if (this.pollingActive || this.closed) return;
    this.pollingActive = true;
    this.setState("polling_fallback", "SSE unavailable; falling back to periodic polling");
    this.pollingHandle = this.schedule(() => {
      void this.pollingTick();
    }, this.pollingIntervalMs);
  }

  private async pollingTick(): Promise<void> {
    if (this.closed || !this.pollingActive) return;
    try {
      await this.opts.onPollingTick?.();
    } catch {
      // Polling-Errors sind dem Caller egal; lastReadError wird
      // beim getJSON-Boundary ohnehin gesetzt.
    }
    this.pollingHandle = this.schedule(() => {
      void this.pollingTick();
    }, this.pollingIntervalMs);
  }

  private stopPolling(): void {
    this.pollingActive = false;
    this.pollingHandle?.();
    this.pollingHandle = undefined;
  }

  private async openStream(): Promise<Response | undefined> {
    if (!this.pollingActive) this.setState("connecting");
    const controller = new AbortController();
    this.abort = controller;
    const headers: Record<string, string> = { Accept: "text/event-stream" };
    if (this.opts.token) headers["X-MTrace-Token"] = this.opts.token;
    if (this.lastEventID) headers["Last-Event-ID"] = this.lastEventID;
    try {
      const response = await this.fetchFn(this.opts.url, { headers, signal: controller.signal });
      if (!response.ok) {
        this.handleConnectionError(`stream returned ${response.status}`);
        return undefined;
      }
      if (!response.body) {
        this.handleConnectionError("stream has no body");
        return undefined;
      }
      return response;
    } catch (err) {
      if (this.closed) return undefined;
      this.handleConnectionError(`fetch failed: ${(err as Error).message}`);
      return undefined;
    }
  }

  private async drainStream(body: ReadableStream<Uint8Array>): Promise<void> {
    const reader = body.getReader();
    const decoder = new TextDecoder("utf-8");
    let buffered = "";
    try {
      while (!this.closed) {
        const { value, done } = await reader.read();
        if (done) break;
        buffered += decoder.decode(value, { stream: true });
        const dispatched = consumeFrames(buffered);
        buffered = dispatched.remainder;
        for (const frame of dispatched.frames) {
          this.dispatchFrame(frame);
        }
      }
    } catch {
      // Read-Error → Reconnect-Pfad im Caller.
    }
  }

  private dispatchFrame(frame: ParsedFrame): void {
    if (frame.id) this.lastEventID = frame.id;
    if (frame.event === "event_appended" && frame.data) {
      const parsed = safeParse<AppendedFrame>(frame.data);
      if (parsed) this.opts.onAppended(parsed);
      return;
    }
    if (frame.event === "backfill_truncated" && frame.data) {
      const parsed = safeParse<TruncatedFrame>(frame.data);
      if (parsed && this.opts.onTruncated) this.opts.onTruncated(parsed);
    }
  }

  private handleConnectionError(detail: string): void {
    if (this.closed) return;
    this.abort = undefined;
    this.reconnectAttempt += 1;
    const backoff = Math.min(
      this.reconnectMaxMs,
      this.reconnectBaseMs * 2 ** Math.max(0, this.reconnectAttempt - 1)
    );
    if (this.reconnectAttempt >= POLLING_FALLBACK_THRESHOLD) {
      this.startPolling();
    } else if (!this.pollingActive) {
      this.setState("connecting", detail);
    }
    this.schedule(() => {
      void this.connect();
    }, backoff);
  }
}

interface ParsedFrame {
  id: string;
  event: string;
  data: string;
}

/**
 * Minimaler WHATWG-SSE-Parser: liest fertig empfangene Frames aus
 * dem Buffer und gibt den unfertigen Rest zurück.
 */
export function consumeFrames(buffered: string): { frames: ParsedFrame[]; remainder: string } {
  const frames: ParsedFrame[] = [];
  let cursor = 0;
  while (cursor < buffered.length) {
    const sep = buffered.indexOf("\n\n", cursor);
    if (sep === -1) break;
    const block = buffered.slice(cursor, sep);
    cursor = sep + 2;
    const frame = parseFrameBlock(block);
    if (frame.event !== "" || frame.data !== "") {
      frames.push(frame);
    }
  }
  return { frames, remainder: buffered.slice(cursor) };
}

function parseFrameBlock(block: string): ParsedFrame {
  const frame: ParsedFrame = { id: "", event: "", data: "" };
  for (const rawLine of block.split("\n")) {
    applyFrameLine(frame, rawLine.endsWith("\r") ? rawLine.slice(0, -1) : rawLine);
  }
  return frame;
}

function applyFrameLine(frame: ParsedFrame, line: string): void {
  if (line === "" || line.startsWith(":")) return;
  const colon = line.indexOf(":");
  const field = colon === -1 ? line : line.slice(0, colon);
  let value = colon === -1 ? "" : line.slice(colon + 1);
  if (value.startsWith(" ")) value = value.slice(1);
  if (field === "id") frame.id = value;
  else if (field === "event") frame.event = value;
  else if (field === "data") {
    frame.data = frame.data === "" ? value : frame.data + "\n" + value;
  }
}

function safeParse<T>(raw: string): T | undefined {
  try {
    return JSON.parse(raw) as T;
  } catch {
    return undefined;
  }
}

function defaultSchedule(cb: () => void, ms: number): () => void {
  const handle = setTimeout(cb, ms);
  return () => clearTimeout(handle);
}
