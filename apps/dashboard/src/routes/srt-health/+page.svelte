<script lang="ts">
  import { onMount } from "svelte";
  import {
    formatBandwidthMbps,
    formatTime,
    isSrtSampleStale,
    listSrtHealth,
    type SrtHealthItem
  } from "$lib/api";

  // SRT-Health-Übersicht (plan-0.6.0 §6 Tranche 5, RAK-43/RAK-44).
  // Zeigt pro Stream den jüngsten persistierten Health-Sample mit
  // Health-Badge, vier Pflichtmetriken und Freshness-Hinweis.
  // Live-Updates via Polling (alle 5 s); Cursor-Pagination ist als
  // Folge-Item dokumentiert.

  let items: SrtHealthItem[] = [];
  let loading = true;
  let error = "";
  let pollHandle: ReturnType<typeof setInterval> | undefined;

  async function refresh(): Promise<void> {
    try {
      const res = await listSrtHealth();
      items = res.items;
      error = "";
    } catch (err) {
      // getJSON wirft Error-Instanzen; Fallback-String wäre toter Code.
      error = (err as Error).message;
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    void refresh();
    pollHandle = setInterval(() => {
      void refresh();
    }, 5_000);
    return () => {
      if (pollHandle) {
        clearInterval(pollHandle);
      }
    };
  });

  function pillClass(item: SrtHealthItem): string {
    if (isSrtSampleStale(item)) {
      return "pill stale";
    }
    return `pill ${item.health_state}`;
  }

  function pillLabel(item: SrtHealthItem): string {
    if (isSrtSampleStale(item) && item.health_state !== "unknown") {
      return `${item.health_state} (stale)`;
    }
    return item.health_state;
  }
</script>

<section class="page-head">
  <div class="page-title">
    <h1>SRT health</h1>
    <p>Latest persisted sample per stream — RTT, packet loss, retransmissions, available bandwidth.</p>
  </div>
  <div class="toolbar">
    <button class="button" on:click={refresh} disabled={loading}>Refresh</button>
  </div>
</section>

{#if error}
  <p class="error">{error}</p>
{/if}

<section class="panel">
  <div class="table-wrap">
    <table>
      <thead>
        <tr>
          <th>Stream</th>
          <th>Health</th>
          <th>RTT</th>
          <th>Loss (total)</th>
          <th>Retrans (total)</th>
          <th>Avail. bandwidth</th>
          <th>Source</th>
          <th>Last update</th>
        </tr>
      </thead>
      <tbody>
        {#each items as item (item.stream_id)}
          <tr>
            <td><a href={`/srt-health/${encodeURIComponent(item.stream_id)}`}>{item.stream_id}</a></td>
            <td>
              <span class={pillClass(item)}>{pillLabel(item)}</span>
              {#if item.source_status !== "ok"}
                <span class="muted source-status" title={item.source_error_code}>
                  source: {item.source_status}
                </span>
              {/if}
            </td>
            <td>{item.metrics.rtt_ms.toFixed(2)} ms</td>
            <td>{item.metrics.packet_loss_total}</td>
            <td>{item.metrics.retransmissions_total}</td>
            <td>{formatBandwidthMbps(item.metrics.available_bandwidth_bps)}</td>
            <td>
              {#if item.source_status === "ok"}
                ok
              {:else}
                {item.source_error_code}
              {/if}
            </td>
            <td>
              {formatTime(item.freshness.ingested_at)}
              <span class="muted age">({Math.round(item.freshness.sample_age_ms / 1000)}s ago)</span>
            </td>
          </tr>
        {:else}
          <tr>
            <td colspan="8" class="muted">
              {#if loading}Loading…{:else}No SRT streams reported. Collector may be disabled (set MTRACE_SRT_SOURCE_URL).{/if}
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</section>

<style>
  .source-status {
    margin-left: 6px;
    font-size: 0.85em;
  }
  .age {
    margin-left: 4px;
    font-size: 0.85em;
  }
  .pill.stale {
    background: var(--color-stale-bg, #fde68a);
    color: var(--color-stale-fg, #78350f);
  }
</style>
