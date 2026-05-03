<script lang="ts">
  import { onMount } from "svelte";
  import { getHealth, type HealthStatus } from "$lib/api";

  let health: HealthStatus = { ok: false, status: 0, text: "not checked" };
  let loading = true;

  type ServiceStatus = "connected" | "inactive";

  let observabilityServices: Array<{ name: string; url: string; openUrl?: string; status: ServiceStatus; note: string }> = [
    {
      name: "Prometheus",
      url: "http://localhost:9090/-/ready",
      openUrl: "http://localhost:9090",
      status: "inactive",
      note: "observability profile"
    },
    {
      name: "Grafana",
      url: "http://localhost:3000/api/health",
      openUrl: "http://localhost:3000",
      status: "inactive",
      note: "observability profile"
    },
    {
      name: "OTel Collector",
      url: "http://localhost:13133",
      status: "inactive",
      note: "health endpoint"
    }
  ];

  async function probe(url: string): Promise<ServiceStatus> {
    try {
      await fetch(url, { mode: "no-cors", cache: "no-store" });
      return "connected";
    } catch {
      return "inactive";
    }
  }

  async function refresh() {
    loading = true;
    const [nextHealth, ...statuses] = await Promise.all([getHealth(), ...observabilityServices.map((service) => probe(service.url))]);
    health = nextHealth;
    observabilityServices = observabilityServices.map((service, index) => ({ ...service, status: statuses[index] }));
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
      <span class={`pill ${observabilityServices.some((service) => service.status === "connected") ? "connected" : "inactive"}`}>
        {observabilityServices.some((service) => service.status === "connected") ? "connected" : "inactive"}
      </span>
    </div>
    <div class="status-list">
      {#each observabilityServices as service (service.name)}
        <div class="status-row">
          <div>
            <strong>{service.name}</strong>
            <span class="muted">{service.note}</span>
          </div>
          <div class="status-actions">
            <span class={`pill ${service.status}`}>{service.status}</span>
            {#if service.openUrl}
              <a href={service.openUrl} target="_blank" rel="noreferrer">Open</a>
            {/if}
          </div>
        </div>
      {/each}
    </div>
  </div>
</section>
