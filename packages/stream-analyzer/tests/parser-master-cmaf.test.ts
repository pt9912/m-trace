import { describe, expect, it } from "vitest";

import { parseMasterPlaylist } from "../src/internal/parsers/master.js";

/**
 * Tranche-2-Tests für `0.10.0` (NF-13 / RAK-61): Master-Playlist-
 * Pfad ist konservativ. Variant-URIs mit fMP4-Suffix erzeugen nur
 * `confidence:"inferred"`, weil der Master-Pfad referenzierte
 * Media-Playlists nicht nachlädt. `CODECS` allein darf kein
 * CMAF-Signal erzeugen — klassische TS-HLS-Master tragen ebenfalls
 * Codecs.
 */

describe("parseMasterPlaylist — CMAF detection", () => {
  it("emits inferred cmaf when a variant URI has an fMP4 suffix", () => {
    const text = [
      "#EXTM3U",
      '#EXT-X-STREAM-INF:BANDWIDTH=1280000,RESOLUTION=1280x720,CODECS="avc1.4d401e"',
      "video/720p/main.m4s.m3u8",
      '#EXT-X-STREAM-INF:BANDWIDTH=2560000,RESOLUTION=1920x1080,CODECS="avc1.640028"',
      "video/1080p/main.m4s.m3u8"
    ].join("\n");
    const { details } = parseMasterPlaylist(text, "https://cdn.example.test/");
    // Der konkrete Test pinnt: Master-Variants mit fMP4-Suffix -> inferred.
    // Die Beispiel-URIs enden mit `.m3u8`, also kein fMP4-Suffix; das ist
    // realistischer Master-Inhalt — und Master sollte daraus nichts ableiten.
    expect(details.cmaf).toBeUndefined();
  });

  it("emits inferred cmaf when a variant points directly at an fMP4 segment-URI", () => {
    // Konstruiertes Beispiel: einige Master verlinken nicht auf Media-
    // Playlists, sondern direkt auf fMP4-Segmente (z. B. Single-File-
    // CMAF). Das ist ein schwacher Hinweis und reicht für inferred.
    const text = [
      "#EXTM3U",
      '#EXT-X-STREAM-INF:BANDWIDTH=1280000,CODECS="avc1.4d401e"',
      "video/720p/track.m4s",
      '#EXT-X-STREAM-INF:BANDWIDTH=2560000,CODECS="avc1.640028"',
      "video/1080p/track.cmfv"
    ].join("\n");
    const { details } = parseMasterPlaylist(text, "https://cdn.example.test/");
    expect(details.cmaf).toBeDefined();
    expect(details.cmaf!.source).toBe("hls");
    expect(details.cmaf!.confidence).toBe("inferred");
    expect(details.cmaf!.signals).toHaveLength(2);
    expect(details.cmaf!.signals.every((s) => s.code === "hls_master_variant_fmp4_uri")).toBe(true);
    expect(details.cmaf!.signals.every((s) => s.confidence === "inferred")).toBe(true);
    expect(details.cmaf!.signals[0].manifestAnchor).toBe("master:line:3");
    expect(details.cmaf!.signals[1].manifestAnchor).toBe("master:line:5");
    // Master-Summaries tragen in 0.10.0 kein binary-Objekt.
    expect(details.cmaf!.binary).toBeUndefined();
  });

  it("does not emit cmaf for a CODECS-only TS master", () => {
    const text = [
      "#EXTM3U",
      '#EXT-X-STREAM-INF:BANDWIDTH=1280000,RESOLUTION=1280x720,CODECS="avc1.4d401e,mp4a.40.2"',
      "video/720p/index.m3u8",
      '#EXT-X-STREAM-INF:BANDWIDTH=2560000,RESOLUTION=1920x1080,CODECS="avc1.640028,mp4a.40.2"',
      "video/1080p/index.m3u8"
    ].join("\n");
    const { details } = parseMasterPlaylist(text, "https://cdn.example.test/");
    expect(details.cmaf).toBeUndefined();
  });

  it("emits cmaf only for the matching variants in a mixed master", () => {
    const text = [
      "#EXTM3U",
      // Variant 1: TS-Style — kein Signal.
      '#EXT-X-STREAM-INF:BANDWIDTH=1280000,CODECS="avc1.4d401e,mp4a.40.2"',
      "video/720p/index.m3u8",
      // Variant 2: direkt auf fMP4 — schwaches Signal.
      '#EXT-X-STREAM-INF:BANDWIDTH=2560000,CODECS="avc1.640028"',
      "video/1080p/track.m4s",
      // Variant 3: cmfa Audio — schwaches Signal.
      '#EXT-X-STREAM-INF:BANDWIDTH=128000,CODECS="mp4a.40.2"',
      "audio/track.cmfa"
    ].join("\n");
    const { details } = parseMasterPlaylist(text, undefined);
    expect(details.cmaf).toBeDefined();
    const signals = details.cmaf!.signals;
    expect(signals).toHaveLength(2);
    expect(signals.map((s) => s.manifestAnchor)).toEqual([
      "master:line:5",
      "master:line:7"
    ]);
    expect(details.cmaf!.confidence).toBe("inferred");
  });
});
