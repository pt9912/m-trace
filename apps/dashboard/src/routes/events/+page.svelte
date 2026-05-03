<script lang="ts">
  import { onMount } from "svelte";
  import { formatTime, getSession, listSessions, type PlaybackEvent, type StreamSession } from "$lib/api";

  let sessions: StreamSession[] = [];
  let events: PlaybackEvent[] = [];
  let sessionFilter = "all";
  let eventTypeFilter = "all";
  let loading = true;
  let error = "";

  $: eventTypes = Array.from(new Set(events.map((event) => event.event_name))).sort();
  $: visibleEvents = events.filter((event) => {
    const sessionMatches = sessionFilter === "all" || event.session_id === sessionFilter;
    const typeMatches = eventTypeFilter === "all" || event.event_name === eventTypeFilter;
    return sessionMatches && typeMatches;
  });

  async function refresh() {
    loading = true;
    error = "";
    try {
      sessions = (await listSessions(100)).sessions;
      const details = await Promise.all(sessions.slice(0, 25).map((session) => getSession(session.session_id, 100)));
      events = details
        .flatMap((detail) => detail.events)
        .sort((a, b) => Date.parse(b.server_received_at) - Date.parse(a.server_received_at));
    } catch (err) {
      error = err instanceof Error ? err.message : "Could not load events";
    } finally {
      loading = false;
    }
  }

  onMount(refresh);
</script>

<section class="page-head">
  <div class="page-title">
    <h1>Events</h1>
    <p>Playback events across recent sessions.</p>
  </div>
  <div class="toolbar">
    <select bind:value={sessionFilter} aria-label="Session filter">
      <option value="all">All sessions</option>
      {#each sessions as session (session.session_id)}
        <option value={session.session_id}>{session.session_id}</option>
      {/each}
    </select>
    <select bind:value={eventTypeFilter} aria-label="Event type filter">
      <option value="all">All event types</option>
      {#each eventTypes as eventType (eventType)}
        <option value={eventType}>{eventType}</option>
      {/each}
    </select>
    <button class="button" on:click={refresh} disabled={loading}>Refresh</button>
  </div>
</section>

{#if error}
  <p class="error">{error}</p>
{/if}

<section class="panel">
  <div class="panel-head">
    <h2>Event list</h2>
    <span class="muted">{visibleEvents.length} of {events.length} loaded</span>
  </div>
  <div class="table-wrap">
    <table>
      <thead>
        <tr>
          <th>Time</th>
          <th>Event</th>
          <th>Session</th>
          <th>Project</th>
          <th>Sequence</th>
        </tr>
      </thead>
      <tbody>
        {#each visibleEvents as event (`${event.session_id}:${event.ingest_sequence}`)}
          <tr>
            <td>{formatTime(event.server_received_at)}</td>
            <td>{event.event_name}</td>
            <td><a href={`/sessions/${event.session_id}`}>{event.session_id}</a></td>
            <td>{event.project_id}</td>
            <td>{event.sequence_number ?? event.ingest_sequence}</td>
          </tr>
        {:else}
          <tr>
            <td colspan="5" class="muted">No matching events.</td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</section>
