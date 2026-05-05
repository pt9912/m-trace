/** Tripel aus dem Read-Shape `network_signal_absent[]` (Backend-API
 *  §3.7.1, plan-0.4.0 §4.4). Markiert Stellen, an denen das SDK kein
 *  Manifest-/Segment-Signal beobachten konnte (Native HLS, CORS-
 *  Block, Resource-Timing-Lücke). `kind` ist der Netzwerksignal-Typ,
 *  nicht der Boundary-`kind`-Wert. */
export interface NetworkSignalAbsentEntry {
  kind: "manifest" | "segment";
  adapter: "hls.js" | "native_hls" | "unknown";
  reason: string;
}

export interface StreamSession {
  session_id: string;
  project_id: string;
  /** Backend-Domain: typischerweise "active" | "stalled" | "ended"
   *  (siehe domain.SessionState in apps/api), kann sich aber bei
   *  Schema-Erweiterungen vergrößern — daher als Top-Type. */
  state: string;
  started_at: string;
  last_event_at: string;
  ended_at?: string;
  event_count: number;
  /** Server-vergeben ab 0.4.0 §3.2-Closeout; Empty-String bei
   *  Legacy-Sessions vor dem Closeout (siehe API-Kontrakt §3.7.1). */
  correlation_id?: string;
  /** Default `[]`; siehe API-Kontrakt §3.7.1, plan-0.4.0 §4.4. */
  network_signal_absent: NetworkSignalAbsentEntry[];
  /** Auslöser des Endzustands (plan-0.4.0 §5 H1):
   *  - `"client"`  bei explizitem `session_ended`-Event
   *  - `"sweeper"` bei zeitbasiertem Sweeper-Ende
   *  - `null` für aktive Sessions oder Legacy-Einträge */
  end_source: "client" | "sweeper" | null;
}

export interface PlaybackEvent {
  event_name: string;
  project_id: string;
  session_id: string;
  client_timestamp: string;
  server_received_at: string;
  ingest_sequence: number;
  sequence_number?: number;
  sdk: {
    name: string;
    version: string;
  };
  meta?: Record<string, unknown>;
  /** Server-vergeben ab 0.4.0 §3.2-Closeout (API-Kontrakt §3.7.1). */
  correlation_id?: string;
  /** W3C-Trace-ID des Batches (32 Hex), optional. */
  trace_id?: string;
  /** Klassifikation des Events nach API-Kontrakt §10.2:
   *  `"accepted"` (Default), `"duplicate_suspected"` oder
   *  `"replayed"`. Vor §2.3-Closeout liefern Read-Antworten das
   *  Feld nicht; daher optional. */
  delivery_status?: "accepted" | "duplicate_suspected" | "replayed";
}

export interface SessionsResponse {
  sessions: StreamSession[];
  next_cursor?: string;
}

export interface SessionDetailResponse {
  session: StreamSession;
  events: PlaybackEvent[];
  /** Cursor für die nächste Event-Seite; fehlt bei letzter Seite
   *  (API-Kontrakt §10.3 Cursor v3). */
  next_cursor?: string;
}

export interface HealthStatus {
  ok: boolean;
  status: number;
  text: string;
}

export async function listSessions(limit = 100): Promise<SessionsResponse> {
  return getJSON<SessionsResponse>(`${apiBaseUrl}/api/stream-sessions?limit=${limit}`);
}

export async function getSession(
  sessionId: string,
  eventsLimit = 200,
  eventsCursor?: string
): Promise<SessionDetailResponse> {
  const params = new URLSearchParams();
  params.set("events_limit", String(eventsLimit));
  if (eventsCursor) {
    params.set("events_cursor", eventsCursor);
  }
  return getJSON<SessionDetailResponse>(
    `${apiBaseUrl}/api/stream-sessions/${encodeURIComponent(sessionId)}?${params.toString()}`
  );
}

// SRT-Health-Wire-Format aus spec/backend-api-contract.md §7a.2
// (plan-0.6.0 §5 Tranche 4). Felder kommen direkt aus dem
// Go-Adapter `srtHealthWireItem` und sind hier 1:1 typisiert,
// inklusive der drei Sub-Blöcke `metrics`/`derived`/`freshness`.

export type SrtHealthState = "healthy" | "degraded" | "critical" | "unknown";
export type SrtSourceStatus =
  | "ok"
  | "unavailable"
  | "partial"
  | "stale"
  | "no_active_connection";
export type SrtSourceErrorCode =
  | "none"
  | "source_unavailable"
  | "no_active_connection"
  | "partial_sample"
  | "stale_sample"
  | "parse_error";
export type SrtConnectionState = "connected" | "no_active_connection" | "unknown";

export interface SrtHealthMetrics {
  rtt_ms: number;
  packet_loss_total: number;
  packet_loss_rate?: number;
  retransmissions_total: number;
  available_bandwidth_bps: number;
  throughput_bps?: number;
  required_bandwidth_bps?: number;
}

export interface SrtHealthDerived {
  bandwidth_headroom_factor?: number;
}

export interface SrtHealthFreshness {
  source_observed_at: string | null;
  source_sequence?: string;
  collected_at: string;
  ingested_at: string;
  sample_age_ms: number;
  stale_after_ms: number;
}

export interface SrtHealthItem {
  stream_id: string;
  connection_id: string;
  health_state: SrtHealthState;
  source_status: SrtSourceStatus;
  source_error_code: SrtSourceErrorCode;
  connection_state: SrtConnectionState;
  metrics: SrtHealthMetrics;
  derived: SrtHealthDerived;
  freshness: SrtHealthFreshness;
}

export interface SrtHealthListResponse {
  items: SrtHealthItem[];
}

export interface SrtHealthDetailResponse {
  stream_id: string;
  items: SrtHealthItem[];
}

/** Liefert pro StreamID den jüngsten persistierten Health-Sample
 *  (`GET /api/srt/health`). 404/500 propagieren als Error vom
 *  `getJSON`-Pfad. */
export async function listSrtHealth(): Promise<SrtHealthListResponse> {
  return getJSON<SrtHealthListResponse>(`${apiBaseUrl}/api/srt/health`);
}

/** Liefert die letzten `samplesLimit` Samples für eine StreamID
 *  (`GET /api/srt/health/{stream_id}`). 404 → `streamUnknown` */
export async function getSrtHealthDetail(
  streamId: string,
  samplesLimit = 100
): Promise<SrtHealthDetailResponse> {
  const params = new URLSearchParams({ samples_limit: String(samplesLimit) });
  return getJSON<SrtHealthDetailResponse>(
    `${apiBaseUrl}/api/srt/health/${encodeURIComponent(streamId)}?${params.toString()}`
  );
}

/** Stale-Bewertung im UI: ein Sample ist stale, wenn entweder
 *  source_status === "stale" persistiert ist (Server hat Drift
 *  erkannt) ODER sample_age_ms > stale_after_ms zum Lesezeitpunkt.
 *  Die zweite Variante deckt den Fall ab, dass der Collector
 *  zwischen zwei Polls keinen frischen Sample erzeugt hat. */
export function isSrtSampleStale(item: SrtHealthItem): boolean {
  if (item.source_status === "stale") {
    return true;
  }
  return item.freshness.sample_age_ms > item.freshness.stale_after_ms;
}

/** Formatiert eine Bandbreite in bit/s als Mbit/s mit drei Nach-
 *  kommastellen (Lab-Werte sind im Gbps-Bereich; Mbit/s ist die
 *  Operator-Lesbarkeit). */
export function formatBandwidthMbps(bps: number): string {
  return `${(bps / 1_000_000).toFixed(3)} Mbit/s`;
}

export async function getHealth(): Promise<HealthStatus> {
  try {
    const res = await fetch(`${apiBaseUrl}/api/health`, { cache: "no-store" });
    return { ok: res.ok, status: res.status, text: await res.text() };
  } catch (error) {
    return { ok: false, status: 0, text: error instanceof Error ? error.message : "unreachable" };
  }
}

async function getJSON<T>(url: string): Promise<T> {
  const headers: Record<string, string> = { Accept: "application/json" };
  // X-MTrace-Token ist ab plan-0.4.0 §4.2/§4.3 für alle Read-Endpunkte
  // Pflicht. Der Token stammt aus PUBLIC_API_TOKEN; ohne Token wird
  // der Header weggelassen — die API antwortet dann mit 401 und der
  // Caller-Wrapper wirft den Fehler-Pfad unten.
  if (apiToken) {
    headers["X-MTrace-Token"] = apiToken;
  }
  try {
    const res = await fetch(url, {
      headers,
      cache: "no-store"
    });
    if (!res.ok) {
      const err = new Error(`${url} returned ${res.status}`);
      recordReadError(url, err);
      throw err;
    }
    return (await res.json()) as T;
  } catch (err) {
    // Network/timeout/unreachable: nur einmal recorden, nicht doppelt
    // bei thrown HTTP-Errors (die haben recordReadError schon oben).
    if (!(err instanceof Error) || !err.message.includes("returned")) {
      recordReadError(url, err);
    }
    throw err;
  }
}

export function formatTime(value: string | undefined): string {
  if (!value) {
    return "n/a";
  }
  return new Intl.DateTimeFormat(undefined, {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit"
  }).format(new Date(value));
}

export function isErrorEvent(event: PlaybackEvent): boolean {
  return event.event_name.includes("error") || event.event_name.includes("warning");
}
import { env } from "$env/dynamic/public";
import { recordReadError } from "./status";

const apiBaseUrl = (env.PUBLIC_API_BASE_URL ?? "").replace(/\/$/, "");
const apiToken = env.PUBLIC_API_TOKEN ?? "";
