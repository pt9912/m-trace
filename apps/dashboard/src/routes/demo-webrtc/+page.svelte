<script lang="ts">
  import { onDestroy, onMount } from "svelte";
  import { env } from "$env/dynamic/public";
  import { attachWebRtc, createTracker, type PlayerTracker, type WebRtcAdapter } from "@pt9912/player-sdk";

  //  — Demo-Route gegen das examples/webrtc/-
  // Lab-Compose. Bestehende /demo-Route (hls.js) bleibt unverändert;
  // diese Route ist additiv und nutzt denselben Tracker-Pfad in den
  // API-Ingress.
  const whepUrl = env.PUBLIC_WHEP_URL || "http://localhost:8892/webrtc-test/whep";
  const collectorEndpoint = env.PUBLIC_PLAYER_COLLECTOR_ENDPOINT || "http://localhost:8080/api/playback-events";

  let video: HTMLVideoElement;
  let status = "idle";
  let adapter: WebRtcAdapter | undefined;
  let tracker: PlayerTracker | undefined;

  function start(): void {
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

    try {
      adapter = attachWebRtc(video, { whepUrl }, tracker);
      status = `WebRTC handshake against ${whepUrl}`;
    } catch (err) {
      status = `attach failed: ${err instanceof Error ? err.message : String(err)}`;
    }
  }

  function stop(): void {
    adapter?.destroy();
    adapter = undefined;
    void tracker?.destroy();
    tracker = undefined;
    if (video) {
      video.pause();
      try {
        video.srcObject = null;
      } catch {
        // ignore
      }
    }
  }

  onMount(() => {
    const params = new URLSearchParams(window.location.search);
    if (params.get("autostart") === "1") {
      start();
    }
  });

  onDestroy(stop);
</script>

<section class="page-head">
  <div class="page-title">
    <h1>Demo player (WebRTC / WHEP)</h1>
    <p>WebRTC-Adapter aus dem Player-SDK gegen das examples/webrtc/-Lab.</p>
  </div>
  <div class="toolbar">
    <button class="button" on:click={start}>Start</button>
    <button class="button secondary" on:click={stop}>Stop</button>
  </div>
</section>

<video class="player" bind:this={video} autoplay controls muted playsinline></video>

<p class="muted">Status: {status}</p>
<p class="muted">
  Lab-Stack:
  <code>docker compose -p mtrace-webrtc -f examples/webrtc/compose.yaml up -d --build</code>
  · Default WHEP-URL: <code>{whepUrl}</code>
</p>
