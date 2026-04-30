<script lang="ts">
  import { onMount } from "svelte";
  import { formatTime, getSession, isErrorEvent, listSessions, type PlaybackEvent } from "$lib/api";

  let errors: PlaybackEvent[] = [];
  let loading = true;
  let error = "";

  async function refresh() {
    loading = true;
    error = "";
    try {
      const sessions = (await listSessions(100)).sessions;
      const details = await Promise.all(sessions.slice(0, 25).map((session) => getSession(session.session_id, 100)));
      errors = details.flatMap((detail) => detail.events.filter(isErrorEvent));
    } catch (err) {
      error = err instanceof Error ? err.message : "Could not load errors";
    } finally {
      loading = false;
    }
  }

  onMount(refresh);
</script>

<section class="page-head">
  <div class="page-title">
    <h1>Errors</h1>
    <p>Playback errors and warnings from recent sessions.</p>
  </div>
  <button class="button" on:click={refresh} disabled={loading}>Refresh</button>
</section>

{#if error}
  <p class="error">{error}</p>
{/if}

<section class="panel">
  <div class="table-wrap">
    <table>
      <thead>
        <tr>
          <th>Time</th>
          <th>Event</th>
          <th>Session</th>
          <th>SDK</th>
        </tr>
      </thead>
      <tbody>
        {#each errors as event}
          <tr>
            <td>{formatTime(event.server_received_at)}</td>
            <td>{event.event_name}</td>
            <td><a href={`/sessions/${event.session_id}`}>{event.session_id}</a></td>
            <td>{`${event.sdk.name} ${event.sdk.version}`}</td>
          </tr>
        {:else}
          <tr>
            <td colspan="4" class="muted">No playback errors found.</td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</section>
