export type PlaybackEventName =
  | "manifest_loaded"
  | "segment_loaded"
  | "bitrate_switch"
  | "rebuffer_started"
  | "rebuffer_ended"
  | "playback_error"
  | "startup_completed";

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
}

export interface PlaybackEventBatch {
  schema_version: "1.0";
  events: PlaybackEvent[];
}

export interface EventDraft {
  eventName: PlaybackEventName;
  timestamp?: Date;
}
