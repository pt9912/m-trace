import type { PlayerTracker } from "../../core/tracker";
import type { EventMeta } from "../../types/events";

/**
 * `getStats()`-Sampling für den WebRTC-Adapter (`plan-0.8.0` Tranche 3).
 *
 * Pollt `pc.getStats()` in einem festen Intervall, extrahiert die
 * Muss-Felder aus `spec/telemetry-model.md` §3.5.2 und sendet sie als
 * `metrics_sampled`-Event mit reserviertem `webrtc.*`-Meta-Namespace
 * (siehe §1.4 / §3.5).
 *
 * **Schema-Drift-Strategie** (§3.5.3): Fehlt ein Muss-Feld
 * (`connectionState`, `dtlsState`, `iceState`, `packetsLost`,
 * `bytesReceived`, `bytesSent`), wird das Sample **gar nicht**
 * emittiert — kein `unknown`-Surrogat. Ein Sample-Tick ohne
 * `RTCPeerConnection.getStats()` wird ebenfalls leise verworfen.
 */

export interface SamplingDeps {
  /** Polling-Interval in Millisekunden. Default 1000 ms. */
  intervalMs?: number;
  /** Test-Override für `setInterval`/`clearInterval`. */
  setInterval?: typeof setInterval;
  clearInterval?: typeof clearInterval;
}

export interface SampleAggregate {
  connectionState: ConnectionState;
  iceState: IceState;
  dtlsState: DtlsState;
  packetsLost: number;
  bytesReceived: number;
  bytesSent: number;
}

type ConnectionState = "new" | "connecting" | "connected" | "disconnected" | "failed" | "closed";
type IceState = "new" | "checking" | "connected" | "completed" | "failed" | "disconnected" | "closed";
type DtlsState = "new" | "connecting" | "connected" | "closed" | "failed";

const VALID_CONNECTION_STATES: readonly ConnectionState[] = [
  "new",
  "connecting",
  "connected",
  "disconnected",
  "failed",
  "closed"
];
const VALID_ICE_STATES: readonly IceState[] = [
  "new",
  "checking",
  "connected",
  "completed",
  "failed",
  "disconnected",
  "closed"
];
const VALID_DTLS_STATES: readonly DtlsState[] = ["new", "connecting", "connected", "closed", "failed"];

interface AggregateAccumulator {
  candidatePairState?: CandidatePairState;
  dtlsState?: DtlsState;
  packetsLost: number;
  bytesReceived: number;
  bytesSent: number;
  hasInbound: boolean;
  hasOutbound: boolean;
}

type CandidatePairState = "frozen" | "waiting" | "in-progress" | "failed" | "succeeded";

const VALID_CANDIDATE_PAIR_STATES: readonly CandidatePairState[] = [
  "frozen",
  "waiting",
  "in-progress",
  "failed",
  "succeeded"
];

/** Wandelt ein `RTCStatsReport` in das Aggregat aus §3.5.2 um. */
export function collectAggregate(
  stats: RTCStatsReport,
  connectionState: string,
  iceConnectionState?: string
): SampleAggregate | null {
  if (!isConnectionState(connectionState)) {
    return null;
  }
  const acc: AggregateAccumulator = {
    packetsLost: 0,
    bytesReceived: 0,
    bytesSent: 0,
    hasInbound: false,
    hasOutbound: false
  };

  stats.forEach((report) => {
    if (typeof report !== "object" || report === null) {
      return;
    }
    accumulateStat(acc, report as Record<string, unknown>);
  });

  const iceState: IceState | undefined =
    typeof iceConnectionState === "string" && isIceState(iceConnectionState)
      ? iceConnectionState
      : mapCandidatePairState(acc.candidatePairState);
  if (!iceState || !acc.dtlsState) {
    return null;
  }
  if (!acc.hasInbound && !acc.hasOutbound) {
    return null;
  }
  return {
    connectionState,
    iceState,
    dtlsState: acc.dtlsState,
    packetsLost: acc.packetsLost,
    bytesReceived: acc.bytesReceived,
    bytesSent: acc.bytesSent
  };
}

function accumulateStat(acc: AggregateAccumulator, r: Record<string, unknown>): void {
  const type = typeof r.type === "string" ? r.type : "";
  if (type === "transport") {
    if (typeof r.dtlsState === "string" && isDtlsState(r.dtlsState)) {
      acc.dtlsState = r.dtlsState;
    }
    return;
  }
  if (type === "candidate-pair") {
    accumulateCandidatePair(acc, r);
    return;
  }
  if (type === "inbound-rtp") {
    acc.hasInbound = true;
    acc.packetsLost += toNonNegativeInt(r.packetsLost);
    acc.bytesReceived += toNonNegativeInt(r.bytesReceived);
    return;
  }
  if (type === "outbound-rtp") {
    acc.hasOutbound = true;
    acc.bytesSent += toNonNegativeInt(r.bytesSent);
  }
}

function accumulateCandidatePair(acc: AggregateAccumulator, r: Record<string, unknown>): void {
  const state = typeof r.state === "string" ? r.state : "";
  if (!isCandidatePairState(state)) {
    return;
  }
  // Aggregat: nominated/selected pair gewinnt; sonst der erste valide.
  if (r.nominated === true || r.selected === true) {
    acc.candidatePairState = state;
    return;
  }
  if (!acc.candidatePairState) {
    acc.candidatePairState = state;
  }
}

function isConnectionState(s: string): s is ConnectionState {
  return (VALID_CONNECTION_STATES as readonly string[]).includes(s);
}
function isIceState(s: string): s is IceState {
  return (VALID_ICE_STATES as readonly string[]).includes(s);
}
function isDtlsState(s: string): s is DtlsState {
  return (VALID_DTLS_STATES as readonly string[]).includes(s);
}
function isCandidatePairState(s: string): s is CandidatePairState {
  return (VALID_CANDIDATE_PAIR_STATES as readonly string[]).includes(s);
}
function mapCandidatePairState(s?: CandidatePairState): IceState | undefined {
  switch (s) {
    case "succeeded":
      return "connected";
    case "in-progress":
      return "checking";
    case "waiting":
    case "frozen":
      return "new";
    case "failed":
      return "failed";
    default:
      return undefined;
  }
}
function toNonNegativeInt(v: unknown): number {
  if (typeof v !== "number" || !Number.isFinite(v) || v < 0) {
    return 0;
  }
  return Math.floor(v);
}

/**
 * Startet die `metrics_sampled`-Sample-Loop. Returns cleanup-Funktion,
 * die das Intervall stoppt.
 */
export function startSampling(
  pc: RTCPeerConnection,
  tracker: PlayerTracker,
  runId: string,
  deps: SamplingDeps = {}
): () => void {
  const intervalMs = deps.intervalMs ?? 1000;
  const setIntervalFn = deps.setInterval ?? globalThis.setInterval;
  const clearIntervalFn = deps.clearInterval ?? globalThis.clearInterval;
  let sampleId = 0;

  const tick = (): void => {
    if (typeof pc.getStats !== "function") {
      return;
    }
    void pc
      .getStats()
      .then((stats) => {
        const aggregate = collectAggregate(stats, pc.connectionState, pc.iceConnectionState);
        if (!aggregate) {
          return;
        }
        const meta: EventMeta = {
          "webrtc.peer_connection_run_id": runId,
          "webrtc.sample_id": sampleId,
          "webrtc.connection_state": aggregate.connectionState,
          "webrtc.ice_state": aggregate.iceState,
          "webrtc.dtls_state": aggregate.dtlsState,
          "webrtc.packets_lost": aggregate.packetsLost,
          "webrtc.bytes_received": aggregate.bytesReceived,
          "webrtc.bytes_sent": aggregate.bytesSent
        };
        sampleId += 1;
        tracker.track({ eventName: "metrics_sampled", meta });
      })
      .catch(() => {
        // getStats() kann in degradierten Browser-Zuständen werfen —
        // schlucken statt einen Schema-Drift-Surrogat-Wert zu emittieren.
      });
  };

  const handle = setIntervalFn(tick, intervalMs);
  return () => clearIntervalFn(handle);
}

/** Generiert eine `peer_connection_run_id` über Browser-Crypto. */
export function newPeerConnectionRunId(): string {
  if (typeof globalThis.crypto?.randomUUID === "function") {
    return globalThis.crypto.randomUUID();
  }

  const bytes = new Uint8Array(16);
  if (typeof globalThis.crypto?.getRandomValues === "function") {
    globalThis.crypto.getRandomValues(bytes);
  } else {
    for (let i = 0; i < bytes.length; i += 1) {
      bytes[i] = Math.floor(Math.random() * 256);
    }
  }

  bytes[6] = (bytes[6] & 0x0f) | 0x40;
  bytes[8] = (bytes[8] & 0x3f) | 0x80;

  const hex = Array.from(bytes, (byte) => byte.toString(16).padStart(2, "0")).join("");
  return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20)}`;
}
