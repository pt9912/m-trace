<script lang="ts">
  import { onMount } from "svelte";
  import { getHealth, type HealthStatus } from "$lib/api";
  import {
    lastReadError,
    observabilityServices,
    probeServices,
    sseConnection
  } from "$lib/status";

  let health: HealthStatus = { ok: false, status: 0, text: "not checked" };
  let loading = true;

  async function refresh() {
    loading = true;
    const [nextHealth, nextLinks] = await Promise.all([
      getHealth(),
      probeServices($observabilityServices)
    ]);
    health = nextHealth;
    observabilityServices.set(nextLinks);
    loading = false;
  }

  function clearReadError() {
    lastReadError.set(null);
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
      <h2>SSE</h2>
      <span class={`pill ${$sseConnection.state}`}>{$sseConnection.state.replace(/_/g, " ")}</span>
    </div>
    <p style="padding: 0 16px 16px;" class="muted">
      {#if $sseConnection.detail}
        {$sseConnection.detail}
      {:else if $sseConnection.state === "not_yet_connected"}
        SSE-Live-Updates werden in Tranche 4 §5 H5 verdrahtet.
      {:else if $sseConnection.changedAt}
        Last change: {$sseConnection.changedAt}
      {/if}
    </p>
  </div>

  <div class="panel">
    <div class="panel-head">
      <h2>Last read error</h2>
      <span class={`pill ${$lastReadError ? "disconnected" : "connected"}`}>
        {$lastReadError ? "error" : "ok"}
      </span>
    </div>
    {#if $lastReadError}
      <div style="padding: 0 16px 16px;">
        <p class="muted">{$lastReadError.occurredAt}</p>
        <p><strong>{$lastReadError.source}</strong></p>
        <p>{$lastReadError.message}</p>
        <button class="button" on:click={clearReadError}>Clear</button>
      </div>
    {:else}
      <p style="padding: 0 16px 16px;" class="muted">No session-read error since dashboard load.</p>
    {/if}
  </div>

  <div class="panel">
    <div class="panel-head">
      <h2>Observability</h2>
      <span class={`pill ${$observabilityServices.some((service) => service.status === "connected") ? "connected" : "inactive"}`}>
        {$observabilityServices.some((service) => service.status === "connected") ? "connected" : "inactive"}
      </span>
    </div>
    <div class="status-list">
      {#each $observabilityServices as service (service.name)}
        <div class="status-row">
          <div>
            <strong>{service.name}</strong>
            <span class="muted">{service.configHint}</span>
          </div>
          <div class="status-actions">
            <span class={`pill ${service.status}`}>{service.status.replace(/_/g, " ")}</span>
            {#if service.openUrl}
              <a href={service.openUrl} target="_blank" rel="noreferrer">Open</a>
            {/if}
          </div>
        </div>
      {/each}
    </div>
  </div>
</section>
