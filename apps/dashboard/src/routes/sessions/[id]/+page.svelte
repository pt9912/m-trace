<script lang="ts">
  import { onMount } from "svelte";
  import { page } from "$app/stores";
  import {
    formatTime,
    getSession,
    isErrorEvent,
    type NetworkSignalAbsentEntry,
    type PlaybackEvent,
    type StreamSession
  } from "$lib/api";

  let session: StreamSession | undefined;
  let events: PlaybackEvent[] = [];
  let nextCursor: string | undefined;
  let loading = true;
  let loadingMore = false;
  let error = "";

  $: sessionId = $page.params.id ?? "";
  $: errorCount = events.filter(isErrorEvent).length;
  $: rebufferCount = events.filter((event) => event.event_name === "rebuffer_started").length;
  $: networkSignalAbsent = (session?.network_signal_absent ?? []) as NetworkSignalAbsentEntry[];

  async function refresh() {
    loading = true;
    error = "";
    nextCursor = undefined;
    if (!sessionId) {
      error = "Missing session id";
      loading = false;
      return;
    }
    try {
      const res = await getSession(sessionId);
      session = res.session;
      events = res.events;
      nextCursor = res.next_cursor;
    } catch (err) {
      error = err instanceof Error ? err.message : "Could not load session";
    } finally {
      loading = false;
    }
  }

  async function loadMore() {
    if (!nextCursor || loadingMore || !sessionId) {
      return;
    }
    loadingMore = true;
    try {
      const res = await getSession(sessionId, 200, nextCursor);
      events = [...events, ...res.events];
      nextCursor = res.next_cursor;
    } catch (err) {
      error = err instanceof Error ? err.message : "Could not load more events";
    } finally {
      loadingMore = false;
    }
  }

  function eventCategory(name: string): "manifest" | "segment" | "lifecycle" | "playback" {
    if (name === "manifest_loaded") return "manifest";
    if (name === "segment_loaded") return "segment";
    if (name === "playback_started" || name === "session_ended" || name === "startup_completed") {
      return "lifecycle";
    }
    return "playback";
  }

  onMount(refresh);
</script>

<section class="page-head">
  <div class="page-title">
    <h1>Session detail</h1>
    <p>{sessionId}</p>
  </div>
  <button class="button" on:click={refresh} disabled={loading}>Refresh</button>
</section>

{#if error}
  <p class="error">{error}</p>
{/if}

<section class="stats">
  <div class="stat">
    <span>State</span>
    <strong>
      {session?.state ?? "n/a"}
      {#if session?.end_source}
        <span
          class="muted end-source"
          title={session.end_source === "client"
            ? "Ended by explicit session_ended event"
            : "Ended by lifecycle sweeper after timeout"}
        >
          via {session.end_source}
        </span>
      {/if}
    </strong>
  </div>
  <div class="stat">
    <span>Events</span>
    <strong>{session?.event_count ?? 0}</strong>
  </div>
  <div class="stat">
    <span>Rebuffers</span>
    <strong>{rebufferCount}</strong>
  </div>
  <div class="stat">
    <span>Errors</span>
    <strong>{errorCount}</strong>
  </div>
</section>

{#if networkSignalAbsent.length > 0}
  <section class="panel" style="margin-top: 18px;">
    <div class="panel-head">
      <h2>Network signal absent</h2>
      <span class="muted">{networkSignalAbsent.length} marker</span>
    </div>
    <p class="muted">
      The SDK reported a session capability boundary — typically Native HLS without
      hls.js timing data — instead of a synthetic manifest/segment event.
    </p>
    <div class="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Kind</th>
            <th>Adapter</th>
            <th>Reason</th>
          </tr>
        </thead>
        <tbody>
          {#each networkSignalAbsent as entry (`${entry.kind}|${entry.adapter}|${entry.reason}`)}
            <tr>
              <td>{entry.kind}</td>
              <td>{entry.adapter}</td>
              <td>{entry.reason}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  </section>
{/if}

<section class="panel" style="margin-top: 18px;">
  <div class="panel-head">
    <h2>Event timeline</h2>
    <span class="muted">{events.length} loaded</span>
  </div>
  <div class="timeline">
    {#each events as event (event.ingest_sequence)}
      <div class="timeline-row" data-category={eventCategory(event.event_name)}>
        <span class="muted">{formatTime(event.server_received_at)}</span>
        <span class={`category-tag ${eventCategory(event.event_name)}`}>
          {eventCategory(event.event_name)}
        </span>
        <strong>{event.event_name}</strong>
        <span class="muted">seq {event.sequence_number ?? event.ingest_sequence}</span>
        {#if event.delivery_status && event.delivery_status !== "accepted"}
          <span
            class={`delivery-pill ${event.delivery_status}`}
            title={event.delivery_status === "duplicate_suspected"
              ? "Likely duplicate — same sequence_number observed before"
              : "Replayed event — admitted out of order or after restart"}
          >
            {event.delivery_status.replace("_", " ")}
          </span>
        {/if}
      </div>
    {:else}
      <div class="timeline-row">
        <span class="muted">No events for this session.</span>
      </div>
    {/each}
  </div>
  {#if nextCursor}
    <button class="button" on:click={loadMore} disabled={loadingMore} style="margin-top: 12px;">
      {loadingMore ? "Loading…" : "Load more events"}
    </button>
  {/if}
</section>

<style>
  .end-source {
    margin-left: 6px;
    font-size: 0.85em;
    font-weight: normal;
  }
  .category-tag {
    padding: 2px 6px;
    border-radius: 4px;
    font-size: 0.8em;
    text-transform: uppercase;
    background: #eef;
  }
  .category-tag.manifest {
    background: #def;
  }
  .category-tag.segment {
    background: #efe;
  }
  .category-tag.lifecycle {
    background: #fed;
  }
  .delivery-pill {
    padding: 2px 6px;
    border-radius: 4px;
    font-size: 0.8em;
    background: #fee;
    color: #a33;
  }
  .delivery-pill.replayed {
    background: #ffe;
    color: #a73;
  }
</style>
