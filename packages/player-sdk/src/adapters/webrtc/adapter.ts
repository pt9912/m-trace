import type { PlayerTracker } from "../../core/tracker";

/**
 * Public-API-Skelett des WebRTC-Adapter-Pfads für `0.8.0` Tranche 1.
 *
 * Diese Datei pinnt den TypeScript-Vertrag (Types + Surface), den Tranche 2
 * mit einer WHEP-basierten Implementation füllt. `attachWebRtc` wirft
 * deshalb in Tranche 1 deterministisch — der Vertrag ist Teil der
 * Public-API, die Implementation ist es noch nicht.
 *
 * Bezug: `docs/planning/in-progress/plan-0.8.0.md` §0.5
 * (Implementierungsleitplanken — `attachWebRtc(video, options, tracker)`
 * additiv zu `attachHlsJs(...)`) und §2 Tranche 1.
 *
 * Contract-Entscheidung Tranche 1: rein SDK-intern. Die Adapter-Auswahl
 * ist Sache des SDK-Konsumenten (opt-in pro Player-Instanz); kein
 * Wire-Schema-Patch in `contracts/event-schema.json`,
 * `contracts/sdk-compat.json` oder `spec/backend-api-contract.md`.
 * Tranche 3 erweitert das Wire-Format um den reservierten
 * `webrtc.*`-Meta-Namespace (siehe `spec/telemetry-model.md` §3.5).
 */

/**
 * Lifecycle-Surface eines WebRTC-Adapters. Spiegelung von
 * {@link import("../hlsjs/adapter").HlsJsAdapter HlsJsAdapter} —
 * `destroy()` räumt PeerConnection und WHEP-Resource auf und beendet
 * alle gemounteten Track-Listener.
 */
export interface WebRtcAdapter {
  /** Beendet PeerConnection, entfernt MediaTracks und gibt WHEP-Resource frei. Idempotent. */
  destroy(): void;
}

/**
 * Konfiguration für `attachWebRtc(...)`. Bewusst minimal in Tranche 1
 * gehalten — Tranche 2 darf weitere Optionen ergänzen, sofern sie
 * additiv und backward-compatible sind.
 */
export interface WebRtcAdapterOptions {
  /**
   * WHEP-Endpoint, gegen den der Adapter eine SDP-Offer POSTet
   * und die SDP-Answer entgegennimmt. Pflicht.
   *
   * Beispiel (Lab-Compose aus `examples/webrtc/`):
   * `http://localhost:8892/webrtc-test/whep`.
   */
  whepUrl: string;

  /**
   * Optionale `RTCPeerConnection`-Konfiguration. Standard ist eine
   * leere Konfiguration (`{}`) — für den localhost-Lab-Pfad reichen
   * lokale ICE-Kandidaten. Für LAN-/Public-Internet-Pfade kann der
   * Aufrufer hier STUN-/TURN-Server übergeben.
   */
  peerConnectionConfig?: RTCConfiguration;

  /**
   * Optionales `AbortSignal` zum Abbruch der WHEP-Signalisierung,
   * bevor die Verbindung steht. Nach `destroy()` wird die
   * PeerConnection so oder so geschlossen; das Signal ist für
   * Konsumenten, die den Aufbau frühzeitig stornieren wollen
   * (z. B. Routenwechsel im Dashboard).
   */
  signal?: AbortSignal;
}

/**
 * Aktiviert einen WebRTC-/WHEP-Read-Pfad auf dem übergebenen
 * `<video>`-Element und routet Player-Events in den `tracker`.
 * Pendant zu {@link import("../hlsjs/adapter").attachHlsJs attachHlsJs};
 * der hls.js-Pfad bleibt Default und unverändert.
 *
 * **Tranche 1**: Public-API-Vertrag. Diese Funktion wirft
 * deterministisch `Error("WebRTC adapter not implemented (plan-0.8.0
 * Tranche 2)")`, damit Konsumenten den Pfad nicht versehentlich vor
 * Tranche 2 in Produktion einsetzen.
 *
 * **Tranche 2** (geplant): WHEP-Handshake (SDP-Offer-POST + Answer),
 * `RTCPeerConnection`-Aufbau, Track-Anbindung an `<video>` plus
 * deterministische Fehlercode-Taxonomie analog `playback_error`-Pfad
 * aus `attachHlsJs`.
 *
 * @param video Ziel-`<video>`-Element. Wird nicht erstellt; der
 *   Aufrufer ist für DOM-Mounting und -Cleanup verantwortlich.
 * @param options WHEP-/PeerConnection-Konfiguration.
 * @param tracker `PlayerTracker` aus `createTracker(...)`. Player-
 *   Events werden über denselben Tracker geschickt wie hls.js-Events;
 *   Adapter-Auswahl ist transparent für die Wire-Surface in Tranche 1.
 * @returns {@link WebRtcAdapter} mit `destroy()`-Lifecycle.
 */
export function attachWebRtc(
  video: HTMLVideoElement,
  options: WebRtcAdapterOptions,
  tracker: PlayerTracker
): WebRtcAdapter {
  // Argumente werden bewusst referenziert, damit ungenutzter-Parameter-
  // Lints (`noUnusedParameters`) den Skelett-Vertrag nicht ablehnen.
  void video;
  void options;
  void tracker;
  throw new Error("WebRTC adapter not implemented (plan-0.8.0 Tranche 2)");
}
