<script>
  import { Client } from '$lib/api/api';
  import EventHeader from '$lib/components/display/EventHeader.svelte';
  import Content from '$lib/components/layout/Content.svelte';
  import Section from '$lib/components/layout/Section.svelte';
  import { ApiPaths } from '$lib/schema';
  import no_map from '$lib/assets/no_map.png';

  let { data } = $props();
</script>

<div class="h-36 w-full">
  <div class="relative flex size-full overflow-hidden">
    {@render nomap()}
    <div class="absolute z-10 flex flex-col px-3 pt-6">
      <h2 class="text-lg text-primary">a home for some tf2 jump stuff</h2>
      {#if data.stats}
        <div class="flex gap-1">
          <span class="text-content/75">managing</span>
          {data.stats.event_count} events
          <span class="text-content/75">with</span>
          {data.stats.times_count} times
          <span class="text-content/75">for</span>
          {data.stats.player_count} players
        </div>
      {/if}
    </div>
  </div>
</div>

<Content>
  <Section label="latest events">
    {#each data.events as ewl}
      {#if ewl.event.kind === 'motw'}
        {#await Client.GET( ApiPaths.get_timeslot_info, { params: { path: { event_id: ewl.event.id } } } )}
          <EventHeader event={ewl} />
        {:then { data: timeslots }}
          <EventHeader event={ewl} {timeslots} />
        {/await}
      {:else}
        <!-- get prizepool info for non-motw -->
        {#await Client.GET( ApiPaths.get_prizepool_total, { params: { path: { event_id: ewl.event.id } } } )}
          <EventHeader event={ewl} />
        {:then { data: prizepoolTotal }}
          <EventHeader event={ewl} prizepool={prizepoolTotal} />
        {/await}
      {/if}
    {/each}
  </Section>
</Content>

{#snippet nomap()}
  <div class="h-48 w-full bg-base-900">
    <div class="size-full mask-x-from-50% mask-x-to-95%">
      <div
        class=" filter-purelavender size-[1476px] rotate-5 animate-[nomap_360s_linear_infinite] bg-size-[30%] bg-repeat"
        style:background-image={`url(${no_map})`}>
      </div>
    </div>
  </div>
{/snippet}
