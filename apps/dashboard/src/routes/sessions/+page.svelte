<script lang="ts">
  import { onMount } from "svelte";
  import { formatTime, listSessions, type StreamSession } from "$lib/api";

  let sessions: StreamSession[] = [];
  let filter = "all";
  let loading = true;
  let error = "";

  $: visibleSessions = filter === "all" ? sessions : sessions.filter((session) => session.state === filter);

  async function refresh() {
    loading = true;
    error = "";
    try {
      sessions = (await listSessions(250)).sessions;
    } catch (err) {
      error = err instanceof Error ? err.message : "Could not load sessions";
    } finally {
      loading = false;
    }
  }

  onMount(refresh);
</script>

<section class="page-head">
  <div class="page-title">
    <h1>Sessions</h1>
    <p>Stream lifecycle state and event volume.</p>
  </div>
  <div class="toolbar">
    <select bind:value={filter} aria-label="State filter">
      <option value="all">All states</option>
      <option value="active">Active</option>
      <option value="stalled">Stalled</option>
      <option value="ended">Ended</option>
    </select>
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
          <th>Session</th>
          <th>State</th>
          <th>Project</th>
          <th>Events</th>
          <th>Started</th>
          <th>Last event</th>
        </tr>
      </thead>
      <tbody>
        {#each visibleSessions as session (session.session_id)}
          <tr>
            <td><a href={`/sessions/${session.session_id}`}>{session.session_id}</a></td>
            <td>
              <span class={`pill ${session.state}`}>{session.state}</span>
              {#if session.end_source}
                <span
                  class="muted end-source"
                  title={session.end_source === "client"
                    ? "Ended by explicit session_ended event"
                    : "Ended by lifecycle sweeper after timeout"}
                >
                  via {session.end_source}
                </span>
              {/if}
            </td>
            <td>{session.project_id}</td>
            <td>{session.event_count}</td>
            <td>{formatTime(session.started_at)}</td>
            <td>{formatTime(session.last_event_at)}</td>
          </tr>
        {:else}
          <tr>
            <td colspan="6" class="muted">No matching sessions.</td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</section>

<style>
  .end-source {
    margin-left: 6px;
    font-size: 0.85em;
  }
</style>
