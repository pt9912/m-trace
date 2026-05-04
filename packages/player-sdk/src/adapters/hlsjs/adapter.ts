import type Hls from "hls.js";
import { SessionMetrics } from "../../core/session-metrics";
import type { PlayerTracker } from "../../core/tracker";
import type { EventMeta } from "../../types/events";
import { redactUrl } from "./redact";

interface HlsEventEmitter {
  on(event: string, callback: (...args: unknown[]) => void): void;
  off(event: string, callback: (...args: unknown[]) => void): void;
}

export interface HlsJsAdapter {
  destroy(): void;
}

// hls.js fragment shape we rely on. Properties are documented in
// hls.js core types; we only consume the dedup-relevant fields plus
// the URL for redaction. Never trust the type-cast — fields may be
// undefined at runtime depending on hls.js version.
interface HlsFragmentLike {
  sn?: number | "initSegment";
  cc?: number;
  type?: string;
  level?: number;
  url?: string;
  relurl?: string;
}

interface HlsLevelLoadedPayload {
  level?: number;
  details?: {
    url?: string;
    live?: boolean;
  };
}

const NETWORK_KIND_MANIFEST = "manifest";
const NETWORK_KIND_SEGMENT = "segment";
const DETAIL_AVAILABLE = "available";

export function attachHlsJs(video: HTMLVideoElement, hls: Hls, tracker: PlayerTracker): HlsJsAdapter {
  const emitter = hls as unknown as HlsEventEmitter;
  const startedAt = performance.now();
  const metrics = new SessionMetrics(startedAt);

  // Dedup-State für `segment_loaded`: hls.js feuert unter normalen
  // Umständen `FRAG_LOADED` nicht doppelt, aber doppelte Listener
  // oder verschachtelte Player-Setups können das tun. Die Set-
  // basierte Dedup-Strategie folgt der Mapping-Tabelle in der
  // README: `(level, type, sn, cc, isInit)`.
  //
  // Für `manifest_loaded` gibt es bewusst keinen Dedup: jedes
  // `MANIFEST_LOADED`/`LEVEL_LOADED` (Live-Refresh, Level-Switch,
  // Reload) erzeugt ein eigenes Event. Doppelt gefeuerte
  // Manifest-Events sind in der Praxis selten und tragen den
  // Refresh-Zustand der Session — sie als Duplikate zu droppen
  // würde Live-Refresh-Patterns unsichtbar machen.
  const seenSegments = new Set<string>();

  const onManifest = (...args: unknown[]) => {
    const meta = manifestMeta(args);
    tracker.track({ eventName: "manifest_loaded", meta });
  };
  const onLevelLoaded = (...args: unknown[]) => {
    const payload = extractLevelLoadedPayload(args);
    const meta: EventMeta = {
      "network.kind": NETWORK_KIND_MANIFEST,
      "network.detail_status": DETAIL_AVAILABLE
    };
    const url = payload?.details?.url;
    if (typeof url === "string" && url.length > 0) {
      meta["network.redacted_url"] = redactUrl(url);
    }
    tracker.track({ eventName: "manifest_loaded", meta });
  };
  const onFragmentLoaded = (...args: unknown[]) => {
    const frag = extractFragment(args);
    const dedupKey = fragmentDedupKey(frag);
    if (dedupKey !== undefined && seenSegments.has(dedupKey)) {
      return;
    }
    if (dedupKey !== undefined) {
      seenSegments.add(dedupKey);
    }
    tracker.track({ eventName: "segment_loaded", meta: segmentMeta(frag) });
  };
  const onLevelSwitched = () => tracker.track({ eventName: "bitrate_switch" });
  const onError = () => tracker.track({ eventName: "playback_error" });
  const onWaiting = () => {
    if (metrics.startRebuffer(performance.now())) {
      tracker.track({ eventName: "rebuffer_started" });
    }
  };
  const onPlaying = () => {
    const rebuffer = metrics.endRebuffer(performance.now());
    if (rebuffer) {
      tracker.track({
        eventName: "rebuffer_ended",
        meta: {
          duration_ms: rebuffer.durationMs,
          rebuffer_count: rebuffer.rebufferCount,
          total_rebuffer_duration_ms: rebuffer.totalRebufferDurationMs
        }
      });
      return;
    }

    reportStartup();
  };
  const onLoadedData = () => reportStartup();

  const reportStartup = () => {
    const startupTimeMs = metrics.completeStartup(performance.now());
    if (startupTimeMs !== undefined) {
      tracker.track({ eventName: "playback_started" });
      tracker.track({ eventName: "startup_time_measured", meta: { duration_ms: startupTimeMs } });
    }
  };

  emitter.on("hlsManifestLoaded", onManifest);
  emitter.on("hlsLevelLoaded", onLevelLoaded);
  emitter.on("hlsFragLoaded", onFragmentLoaded);
  emitter.on("hlsLevelSwitched", onLevelSwitched);
  emitter.on("hlsError", onError);
  video.addEventListener("waiting", onWaiting);
  video.addEventListener("playing", onPlaying);
  video.addEventListener("loadeddata", onLoadedData, { once: true });

  return {
    destroy() {
      emitter.off("hlsManifestLoaded", onManifest);
      emitter.off("hlsLevelLoaded", onLevelLoaded);
      emitter.off("hlsFragLoaded", onFragmentLoaded);
      emitter.off("hlsLevelSwitched", onLevelSwitched);
      emitter.off("hlsError", onError);
      video.removeEventListener("waiting", onWaiting);
      video.removeEventListener("playing", onPlaying);
      video.removeEventListener("loadeddata", onLoadedData);
    }
  };
}

// manifestMeta sammelt die §1.4-`network.*`-Meta-Keys für ein
// initiales `MANIFEST_LOADED`-Event. URL wird hier nicht redigiert,
// weil der initiale Manifest-Lookup in hls.js das URL-Feld nicht
// konsistent als zweites Argument übergibt — `LEVEL_LOADED` füllt das
// Feld zuverlässiger.
function manifestMeta(args: unknown[]): EventMeta {
  const meta: EventMeta = {
    "network.kind": NETWORK_KIND_MANIFEST,
    "network.detail_status": DETAIL_AVAILABLE
  };
  const payload = extractManifestLoadedPayload(args);
  const url = payload?.url;
  if (typeof url === "string" && url.length > 0) {
    meta["network.redacted_url"] = redactUrl(url);
  }
  return meta;
}

// segmentMeta liefert die §1.4-`network.*`-Meta-Keys für ein
// `FRAG_LOADED`-Event. Wenn das URL-Feld fehlt, wird das Event ohne
// `network.redacted_url` rausgereicht — die Server-Side-Validation
// akzeptiert das Fehlen (Vorwärtskompatibilität).
function segmentMeta(frag: HlsFragmentLike | undefined): EventMeta {
  const meta: EventMeta = {
    "network.kind": NETWORK_KIND_SEGMENT,
    "network.detail_status": DETAIL_AVAILABLE
  };
  if (frag === undefined) {
    return meta;
  }
  if (frag.sn === "initSegment") {
    meta.is_init = true;
  }
  if (typeof frag.url === "string" && frag.url.length > 0) {
    meta["network.redacted_url"] = redactUrl(frag.url);
  }
  return meta;
}

// fragmentDedupKey baut den per-session Dedup-Schlüssel ausschließlich
// aus hls.js-nativer Fragment-Identität: `(level, type, sn, cc,
// isInit)`. Init-Segmente bekommen einen eigenen Marker, damit der
// Init-Frag und der erste Media-Frag mit `sn=0` unterscheidbar
// bleiben. Wenn die Felder fehlen (defensive für hls.js-Versionen,
// die das Schema ändern), liefert die Funktion `undefined` und
// schaltet Dedup für dieses Event aus.
function fragmentDedupKey(frag: HlsFragmentLike | undefined): string | undefined {
  if (frag === undefined) {
    return undefined;
  }
  const sn = frag.sn;
  if (sn === undefined) {
    return undefined;
  }
  const level = typeof frag.level === "number" ? frag.level : -1;
  const type = typeof frag.type === "string" ? frag.type : "";
  const cc = typeof frag.cc === "number" ? frag.cc : -1;
  const isInit = sn === "initSegment" ? "init" : "media";
  return `${level}|${type}|${cc}|${isInit}|${sn}`;
}

// firstObjectArg liefert das erste Objekt-Argument der hls.js-Callback-
// Args. hls.js's Event-Form ist meist `(eventName: string, data:
// object)` — wir akzeptieren beide Reihenfolgen, damit Version-Drift
// (Event-Name als erstes Arg vs. Data als erstes Arg) nicht zum
// Shape-Bruch führt.
function firstObjectArg(args: unknown[]): Record<string, unknown> | undefined {
  for (const arg of args) {
    if (typeof arg === "object" && arg !== null) {
      return arg as Record<string, unknown>;
    }
  }
  return undefined;
}

// extractFragment liest das `frag`-Property aus dem
// `FRAG_LOADED`-Payload (hls.js: `{ frag: Fragment, ... }`).
function extractFragment(args: unknown[]): HlsFragmentLike | undefined {
  const data = firstObjectArg(args);
  const candidate = data?.frag;
  if (typeof candidate === "object" && candidate !== null) {
    // HlsFragmentLike has only optional fields → assignable from
    // narrowed `object` without an explicit assertion.
    return candidate;
  }
  return undefined;
}

// extractManifestLoadedPayload liest das URL-Feld aus dem
// `MANIFEST_LOADED`-Payload. hls.js's Payload-Form variiert je
// Version; wir nehmen das erste Object-Argument und akzeptieren ein
// fehlendes `url`-Feld.
function extractManifestLoadedPayload(args: unknown[]): { url?: string } | undefined {
  return firstObjectArg(args);
}

// extractLevelLoadedPayload liest `level` und `details.url` aus dem
// `LEVEL_LOADED`-Payload. hls.js liefert das in modernen Versionen
// als `{ level: number, details: { url: string, ... } }`.
function extractLevelLoadedPayload(args: unknown[]): HlsLevelLoadedPayload | undefined {
  return firstObjectArg(args);
}
