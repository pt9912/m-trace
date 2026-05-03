<script lang="ts">
  import { onMount } from "svelte";
  import { page } from "$app/stores";
  import { formatTime, getSession, isErrorEvent, type PlaybackEvent, type StreamSession } from "$lib/api";

  let session: StreamSession | undefined;
  let events: PlaybackEvent[] = [];
  let loading = true;
  let error = "";

  $: sessionId = $page.params.id ?? "";
  $: errorCount = events.filter(isErrorEvent).length;
  $: rebufferCount = events.filter((event) => event.event_name === "rebuffer_started").length;

  async function refresh() {
    loading = true;
    error = "";
    if (!sessionId) {
      error = "Missing session id";
      loading = false;
      return;
    }
    try {
      const res = await getSession(sessionId);
      session = res.session;
      events = res.events;
    } catch (err) {
      error = err instanceof Error ? err.message : "Could not load session";
    } finally {
      loading = false;
    }
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
    <strong>{session?.state ?? "n/a"}</strong>
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

<section class="panel" style="margin-top: 18px;">
  <div class="panel-head">
    <h2>Event timeline</h2>
    <span class="muted">{events.length} loaded</span>
  </div>
  <div class="timeline">
    {#each events as event (event.ingest_sequence)}
      <div class="timeline-row">
        <span class="muted">{formatTime(event.server_received_at)}</span>
        <strong>{event.event_name}</strong>
        <span class="muted">seq {event.sequence_number ?? event.ingest_sequence}</span>
      </div>
    {:else}
      <div class="timeline-row">
        <span class="muted">No events for this session.</span>
      </div>
    {/each}
  </div>
</section>
