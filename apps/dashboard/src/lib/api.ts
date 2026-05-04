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
  const res = await fetch(url, {
    headers,
    cache: "no-store"
  });
  if (!res.ok) {
    throw new Error(`${url} returned ${res.status}`);
  }
  return (await res.json()) as T;
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

const apiBaseUrl = (env.PUBLIC_API_BASE_URL ?? "").replace(/\/$/, "");
const apiToken = env.PUBLIC_API_TOKEN ?? "";
