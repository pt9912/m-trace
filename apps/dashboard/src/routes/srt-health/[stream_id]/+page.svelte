<script lang="ts">
  import { onDestroy, onMount } from "svelte";
  import { page } from "$app/stores";
  import {
    formatBandwidthMbps,
    formatTime,
    getSrtHealthDetail,
    isSrtSampleStale,
    type SrtHealthDetailResponse,
    type SrtHealthItem
  } from "$lib/api";

  // SRT-Health-Detail-Ansicht (plan-0.6.0 §6 Tranche 5, RAK-44).
  // Zeigt den aktuellen Sample plus eine Mini-Timeline der letzten N
  // Samples. Polling alle 5 s, samples_limit=50 als Default.

  let detail: SrtHealthDetailResponse | undefined;
  let loading = true;
  let error = "";
  let notFound = false;
  let pollHandle: ReturnType<typeof setInterval> | undefined;

  $: streamId = $page.params.stream_id ?? "";
  $: latest = detail?.items[0];
  $: history = detail?.items ?? [];

  async function refresh(): Promise<void> {
    if (!streamId) {
      return;
    }
    try {
      detail = await getSrtHealthDetail(streamId, 50);
      error = "";
      notFound = false;
    } catch (err) {
      // getJSON wirft Error-Instanzen; Fallback-String wäre toter Code.
      const msg = (err as Error).message;
      if (msg.includes("returned 404")) {
        notFound = true;
        error = "";
      } else {
        error = msg;
      }
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    void refresh();
    pollHandle = setInterval(() => {
      void refresh();
    }, 5_000);
  });

  onDestroy(() => {
    if (pollHandle) {
      clearInterval(pollHandle);
    }
  });

  function pillClass(item: SrtHealthItem): string {
    if (isSrtSampleStale(item)) {
      return "pill stale";
    }
    return `pill ${item.health_state}`;
  }

  function bandwidth(item: SrtHealthItem): string {
    return formatBandwidthMbps(item.metrics.available_bandwidth_bps);
  }
</script>

<section class="page-head">
  <div class="page-title">
    <h1>SRT health: <code>{streamId}</code></h1>
    <p>Latest sample plus history (max 50 samples).</p>
  </div>
  <div class="toolbar">
    <a class="button" href="/srt-health">Back</a>
    <button class="button" on:click={refresh} disabled={loading}>Refresh</button>
  </div>
</section>

{#if error}
  <p class="error">{error}</p>
{/if}

{#if notFound}
  <section class="panel">
    <p class="muted">Stream <code>{streamId}</code> has no persisted health samples.</p>
  </section>
{:else if loading && !latest}
  <section class="panel"><p class="muted">Loading…</p></section>
{:else if latest}
  <section class="panel">
    <h2>Current</h2>
    <div class="grid">
      <div>
        <div class="label">Health</div>
        <div><span class={pillClass(latest)}>{latest.health_state}</span></div>
      </div>
      <div>
        <div class="label">Source</div>
        <div>
          {latest.source_status}
          {#if latest.source_error_code !== "none"}
            <span class="muted">({latest.source_error_code})</span>
          {/if}
        </div>
      </div>
      <div>
        <div class="label">Connection</div>
        <div>{latest.connection_state}</div>
      </div>
      <div>
        <div class="label">RTT</div>
        <div>{latest.metrics.rtt_ms.toFixed(2)} ms</div>
      </div>
      <div>
        <div class="label">Packet loss (total)</div>
        <div>{latest.metrics.packet_loss_total}</div>
      </div>
      <div>
        <div class="label">Retransmissions (total)</div>
        <div>{latest.metrics.retransmissions_total}</div>
      </div>
      <div>
        <div class="label">Available bandwidth</div>
        <div>{bandwidth(latest)}</div>
      </div>
      <div>
        <div class="label">Required bandwidth</div>
        <div>
          {#if latest.metrics.required_bandwidth_bps !== undefined}
            {formatBandwidthMbps(latest.metrics.required_bandwidth_bps)}
          {:else}
            <span class="muted">unset</span>
          {/if}
        </div>
      </div>
      <div>
        <div class="label">Throughput</div>
        <div>
          {#if latest.metrics.throughput_bps !== undefined}
            {formatBandwidthMbps(latest.metrics.throughput_bps)}
          {:else}
            <span class="muted">unset</span>
          {/if}
        </div>
      </div>
      <div>
        <div class="label">Bandwidth headroom</div>
        <div>
          {#if latest.derived.bandwidth_headroom_factor !== undefined}
            ×{latest.derived.bandwidth_headroom_factor.toFixed(2)}
          {:else}
            <span class="muted">n/a</span>
          {/if}
        </div>
      </div>
      <div>
        <div class="label">Last update</div>
        <div>
          {formatTime(latest.freshness.ingested_at)}
          <span class="muted">({Math.round(latest.freshness.sample_age_ms / 1000)}s ago, stale &gt;{Math.round(latest.freshness.stale_after_ms / 1000)}s)</span>
        </div>
      </div>
      <div>
        <div class="label">Source observed at</div>
        <div>
          {#if latest.freshness.source_observed_at}
            {formatTime(latest.freshness.source_observed_at)}
          {:else}
            <span class="muted">not provided by source</span>
          {/if}
        </div>
      </div>
    </div>
  </section>

  <section class="panel">
    <h2>History</h2>
    <div class="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Ingested</th>
            <th>Health</th>
            <th>RTT (ms)</th>
            <th>Loss (total)</th>
            <th>Retrans (total)</th>
            <th>Avail. bandwidth</th>
          </tr>
        </thead>
        <tbody>
          {#each history as item (item.freshness.ingested_at + ":" + item.connection_id)}
            <tr>
              <td>{formatTime(item.freshness.ingested_at)}</td>
              <td><span class={pillClass(item)}>{item.health_state}</span></td>
              <td>{item.metrics.rtt_ms.toFixed(2)}</td>
              <td>{item.metrics.packet_loss_total}</td>
              <td>{item.metrics.retransmissions_total}</td>
              <td>{bandwidth(item)}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  </section>
{/if}

<style>
  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
    gap: 16px;
  }
  .label {
    font-size: 0.85em;
    color: var(--color-muted, #6b7280);
    margin-bottom: 4px;
  }
  .pill.stale {
    background: var(--color-stale-bg, #fde68a);
    color: var(--color-stale-fg, #78350f);
  }
</style>
