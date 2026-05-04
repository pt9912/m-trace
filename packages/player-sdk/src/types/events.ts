export type PlaybackEventName =
  | "manifest_loaded"
  | "segment_loaded"
  | "playback_started"
  | "bitrate_switch"
  | "rebuffer_started"
  | "rebuffer_ended"
  | "playback_error"
  | "startup_time_measured"
  | "metrics_sampled"
  | "session_ended"
  | "startup_completed";

export type EventMetaValue = string | number | boolean | null;
export type EventMeta = Record<string, EventMetaValue>;

export interface SDKInfo {
  name: string;
  version: string;
}

export interface PlaybackEvent {
  event_name: PlaybackEventName;
  project_id: string;
  session_id: string;
  client_timestamp: string;
  sequence_number?: number;
  sdk: SDKInfo;
  meta?: EventMeta;
}

export interface PlaybackEventBatch {
  schema_version: "1.0";
  events: PlaybackEvent[];
  /**
   * Optionaler `session_boundaries[]`-Wrapper-Block aus
   * `spec/backend-api-contract.md` §3.4. Maximal 20 pro Batch; jede
   * Boundary muss auf eine `(project_id, session_id)`-Partition
   * verweisen, für die im selben Batch mindestens ein Event
   * existiert. Boundary-only-Batches werden vom Backend mit `422`
   * abgelehnt.
   */
  session_boundaries?: SessionBoundary[];
}

export interface EventDraft {
  eventName: PlaybackEventName;
  timestamp?: Date;
  meta?: EventMeta;
}

/**
 * Network-Signal-Typ einer Session-Boundary. In Tranche 3 sind nur
 * `manifest` und `segment` definiert.
 */
export type BoundaryNetworkKind = "manifest" | "segment";

/**
 * Adapter-Klasse, die eine Boundary gemeldet hat. Tranche 3 erlaubt
 * `hls.js`, `native_hls` oder `unknown`.
 */
export type BoundaryAdapter = "hls.js" | "native_hls" | "unknown";

/**
 * `SessionBoundary` ist ein Wire-Eintrag im
 * `session_boundaries[]`-Block. `kind` ist in Tranche 3 immer
 * `network_signal_absent`; `reason` muss aus dem Reason-Enum von
 * `contracts/event-schema.json#network_unavailable_reasons` kommen
 * und dem Pattern `^[a-z0-9_]{1,64}$` entsprechen — andernfalls
 * lehnt das Backend mit `422` ab.
 */
export interface SessionBoundary {
  kind: "network_signal_absent";
  project_id: string;
  session_id: string;
  network_kind: BoundaryNetworkKind;
  adapter: BoundaryAdapter;
  reason: string;
  client_timestamp: string;
}

/**
 * `BoundaryDraft` ist die Public-API-Form für `tracker.addBoundary`.
 * Der Tracker setzt `kind`, `project_id`, `session_id` und
 * `client_timestamp` selbst — Caller liefern nur Network-Signal-Typ,
 * Adapter und Reason.
 */
export interface BoundaryDraft {
  networkKind: BoundaryNetworkKind;
  adapter: BoundaryAdapter;
  reason: string;
  timestamp?: Date;
}
