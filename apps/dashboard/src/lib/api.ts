export interface StreamSession {
  session_id: string;
  project_id: string;
  state: "active" | "stalled" | "ended" | string;
  started_at: string;
  last_event_at: string;
  ended_at?: string;
  event_count: number;
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
}

export interface SessionsResponse {
  sessions: StreamSession[];
  next_cursor?: string;
}

export interface SessionDetailResponse {
  session: StreamSession;
  events: PlaybackEvent[];
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

export async function getSession(sessionId: string, eventsLimit = 200): Promise<SessionDetailResponse> {
  return getJSON<SessionDetailResponse>(
    `${apiBaseUrl}/api/stream-sessions/${encodeURIComponent(sessionId)}?events_limit=${eventsLimit}`
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
  const res = await fetch(url, {
    headers: { Accept: "application/json" },
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
