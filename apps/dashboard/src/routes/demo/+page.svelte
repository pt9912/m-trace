<script lang="ts">
  import { onDestroy, onMount } from "svelte";
  import { env } from "$env/dynamic/public";
  import type Hls from "hls.js";
  import { attachHlsJs, createTracker, type HlsJsAdapter, type PlayerTracker } from "@m-trace/player-sdk";

  const hlsUrl = env.PUBLIC_HLS_URL || "http://localhost:8888/teststream/index.m3u8";
  const collectorEndpoint = env.PUBLIC_PLAYER_COLLECTOR_ENDPOINT || "http://localhost:8080/api/playback-events";

  let video: HTMLVideoElement;
  let status = "idle";
  let hls: Hls | undefined;
  let adapter: HlsJsAdapter | undefined;
  let tracker: PlayerTracker | undefined;

  async function start() {
    stop();
    status = "starting";

    tracker = createTracker({
      endpoint: collectorEndpoint,
      token: "demo-token",
      projectId: "demo",
      sessionId: new URLSearchParams(window.location.search).get("session_id") ?? undefined,
      batchSize: 5,
      flushIntervalMs: 2000
    });

    const HlsModule = await import("hls.js");
    const HlsConstructor = HlsModule.default;

    if (HlsConstructor.isSupported()) {
      hls = new HlsConstructor();
      hls.loadSource(hlsUrl);
      hls.attachMedia(video);
      adapter = attachHlsJs(video, hls, tracker);
      status = "hls.js attached";
      void video.play();
      return;
    }

    if (video.canPlayType("application/vnd.apple.mpegurl")) {
      video.src = hlsUrl;
      status = "native HLS";
      void video.play();
      return;
    }

    status = "HLS unsupported";
  }

  function stop() {
    adapter?.destroy();
    adapter = undefined;
    hls?.destroy();
    hls = undefined;
    void tracker?.destroy();
    tracker = undefined;
    if (video) {
      video.pause();
      video.removeAttribute("src");
      video.load();
    }
  }

  onMount(() => {
    const params = new URLSearchParams(window.location.search);
    if (params.get("autostart") === "1") {
      void start();
    }
  });

  onDestroy(stop);
</script>

<section class="page-head">
  <div class="page-title">
    <h1>Demo player</h1>
    <p>hls.js stream with Player-SDK event ingestion.</p>
  </div>
  <div class="toolbar">
    <button class="button" on:click={start}>Start</button>
    <button class="button secondary" on:click={stop}>Stop</button>
  </div>
</section>

<video class="player" bind:this={video} controls muted playsinline></video>

<p class="muted">Status: {status}</p>
