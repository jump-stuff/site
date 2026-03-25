<script lang="ts">
  import type { EventWithLeaderboards, PrizepoolTotal, TimeslotInfo } from '$lib/schema';
  import ClassImage from './ClassImage.svelte';
  import Div from './Div.svelte';
  import TemporalDate from './TemporalDate.svelte';
  import no_map from '$lib/assets/no_map.png';
  import { PUBLIC_JUMP_IMAGES_URL } from '$env/static/public';

  type Props = {
    event: EventWithLeaderboards;
    href?: string;
    timeslots?: TimeslotInfo | null;
    prizepool?: PrizepoolTotal | null;
  };

  let { event, href = '', timeslots = null, prizepool = null }: Props = $props();

  function leaderbaordToMaps(l: EventWithLeaderboards['leaderboards']) {
    const maps: Map<string, string[]> = new Map();
    if (!l) return maps;

    for (const { div, map } of l) {
      const divs = maps.get(map) ?? [];
      divs.push(div ?? '');
      maps.set(map, divs);
    }
    return maps;
  }

  let maps = $derived(leaderbaordToMaps(event.leaderboards));

  const twCols = new Map([
    [1, 'grid-cols-1'],
    [2, 'grid-cols-2'],
    [3, 'grid-cols-3'],
    [4, 'grid-cols-4'],
    [5, 'grid-cols-5'],
    [6, 'grid-cols-6'],
    [7, 'grid-cols-7'],
    [8, 'grid-cols-8']
  ]);
</script>

<div
  class="group relative grid h-48 w-full items-end justify-center overflow-hidden bg-base-900
  {twCols.get(maps.size)}">
  {#each maps as [map, divisions]}
    <!-- map wrapper -->
    <div class="relative flex size-full items-end justify-center">
      <!-- absolute map bg image -->
      {#if href}
        <a class="relative flex size-full overflow-hidden" href="/{href}/{event.event.kind_id}">
          {#if map}
            {@render mapimage(map)}
          {:else}
            {@render nomap()}
          {/if}
        </a>
      {:else if map}
        {@render mapimage(map)}
      {:else}
        {@render nomap()}
      {/if}
      <div
        class="pointer-events-none absolute z-10 flex flex-col items-center gap-1
      p-2">
        <a
          class="pointer-events-auto z-10 truncate text-lg text-shadow-xs/100 text-shadow-base-900 hover:text-primary hover:underline"
          href="https://tempus2.xyz/maps/{map}"
          target="_blank">{map}</a>
        <div class="pointer-events-auto flex flex-wrap justify-center gap-2">
          {#each divisions as division}
            <Div div={division} />
          {/each}
        </div>
      </div>
    </div>
  {/each}
  <!-- absolute details container -->
  <div
    class="pointer-events-none absolute top-0 z-10 flex w-full justify-between bg-linear-to-b from-base-900/50 from-50% to-base-900/0 p-2 text-shadow-xs/100 text-shadow-base-900">
    <!-- competition name -->
    <div class="pointer-events-auto z-10 flex h-12 items-center gap-1">
      <ClassImage player_class={event.event.player_class} />
      <span class="text-lg">{event.event.kind} #{event.event.kind_id}</span>
    </div>
    <!-- date / prizepool -->
    <div class="pointer-events-auto z-10 flex flex-col items-end">
      {#if timeslots}
        {#each timeslots.timeslots as ts}
          {@const twPlayerTs =
            timeslots.player_timeslot.timeslot_id === ts.id ? 'text-primary' : 'text-content/75'}
          <div class="flex items-center gap-2 text-nowrap {twPlayerTs}">
            <span class="relative z-10 text-end">
              <TemporalDate datetime={ts.starts_at} />
            </span>
            <span>-</span>
            <span class="relative z-10 text-start text-nowrap">
              <TemporalDate datetime={ts.ends_at} />
            </span>
            <span class="mt-auto icon-[mdi--clock-outline]"></span>
          </div>
        {/each}
      {:else}
        <div class="flex items-center gap-2">
          <span class="relative z-10">
            <TemporalDate datetime={event.event.starts_at} />
          </span>
          <span class="mt-auto icon-[mdi--calendar-outline]"></span>
        </div>
        <div class="flex items-center gap-2">
          <span class="relative z-10">
            <TemporalDate datetime={event.event.ends_at} />
          </span>
          <span class="icon-[mdi--clock-outline]"></span>
        </div>
      {/if}
      <!-- prizepool total -->
      {#if prizepool?.total}
        <div class="flex items-center gap-2">
          <span>{prizepool.total} keys</span>
          <span class="mt-auto icon-[mdi--key]"></span>
        </div>
      {/if}
    </div>
  </div>
</div>

{#snippet mapimage(map: string)}
  <img
    class="over map-img absolute z-10 h-48 w-full scale-105 object-cover brightness-75 transition-all select-none not-first:mask-x-from-98% not-last:mask-x-from-98% group-hover:brightness-100"
    src="{PUBLIC_JUMP_IMAGES_URL}/maps/{map}.jpg"
    alt=""
    draggable="false" />
{/snippet}

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
