<script lang="ts">
  import { onMount } from "svelte";
  import { getHealth, type HealthStatus } from "$lib/api";

  let health: HealthStatus = { ok: false, status: 0, text: "not checked" };
  let loading = true;

  async function refresh() {
    loading = true;
    health = await getHealth();
    loading = false;
  }

  onMount(refresh);
</script>

<section class="page-head">
  <div class="page-title">
    <h1>System status</h1>
    <p>Local service reachability and external consoles.</p>
  </div>
  <button class="button" on:click={refresh} disabled={loading}>Refresh</button>
</section>

<section class="grid">
  <div class="panel">
    <div class="panel-head">
      <h2>API</h2>
      <span class={`pill ${health.ok ? "connected" : "disconnected"}`}>{health.ok ? "connected" : "disconnected"}</span>
    </div>
    <p style="padding: 0 16px 16px;">HTTP {health.status || "n/a"}</p>
  </div>

  <div class="panel">
    <div class="panel-head">
      <h2>Core services</h2>
      <span class="pill warning">linked</span>
    </div>
    <div class="links" style="padding: 16px;">
      <a href="http://localhost:8888/teststream/index.m3u8" target="_blank" rel="noreferrer">MediaMTX HLS</a>
      <a href="http://localhost:9997" target="_blank" rel="noreferrer">MediaMTX API</a>
    </div>
  </div>

  <div class="panel">
    <div class="panel-head">
      <h2>Observability</h2>
      <span class="pill warning">profile optional</span>
    </div>
    <div class="links" style="padding: 16px;">
      <a href="http://localhost:9090" target="_blank" rel="noreferrer">Prometheus</a>
      <a href="http://localhost:3000" target="_blank" rel="noreferrer">Grafana</a>
    </div>
  </div>
</section>
