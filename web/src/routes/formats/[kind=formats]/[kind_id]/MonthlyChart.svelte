<!--
 Future Tasks:
 * Center or randomize vertical positions when vertical space is not all consumed across the chart.
 * Prevent hover labels from escaping horizontally.
 * Potentially highlight table from chart hover.
 * Support half and quarter second tick time steps.
 * Hide hover stems during initial animation.
 * Fix mobile drag actions (context menu and page move).
 * Fix when no space for horizontal outliers to grow.
 * Honor reduced motion.
-->

<script lang="ts">
  import { range } from '$lib/helpers/range';
  import { formatDuration } from '$lib/helpers/times';
  import type { TimeWithPlayer } from '$lib/schema';
  import { onDestroy } from 'svelte';
  import { fly } from 'svelte/transition';

  const {
    times,
    playerIdHighlighted = $bindable()
  }: { times: TimeWithPlayer[]; playerIdHighlighted?: string } = $props();

  const chartPadding = { x: 24, y: 12 };
  const chartHeight = 140;
  const chartHeightWithoutPadding = chartHeight - chartPadding.y * 2;
  let chartWidth = $state(0); // NOTE: Waiting for this to propagate means the chart is partially rendered twice at the start (the intermediate state is not displayed).
  const chartWidthWithoutPadding = $derived(chartWidth - chartPadding.x * 2);
  const animateInDurationMs = 800;

  const maxNameplateWidth = 120;
  const iconSize = (position: number, alias: string) =>
    alias === 'mur' || position <= 3 ? 32 : position <= 10 ? 24 : 10;
  const iconStyle = (position: number, alias: string) =>
    alias === 'mur' || position <= 10 ? 'avatar' : 'dot';
  const iconSizeWithRing = (position: number, alias: string) =>
    iconSize(position, alias) + (iconStyle(position, alias) === 'avatar' ? 9 : 1);
  const iconFloorYoffset = 30;
  const tickLabelWidth = 40 * 3; // TODO: The text is about 40... the multiplier is adjusted arbitrarily for now to produce ok results. This needs investigation.

  const removeHighOutliersByIqr = (times: TimeWithPlayer[], iqrMultiplier = 1.5) => {
    const sortedTimes = times.toSorted((a, b) => a.time.duration - b.time.duration);
    const q1 = sortedTimes[Math.floor(sortedTimes.length * 0.25)]?.time.duration;
    const q3 = sortedTimes[Math.ceil(sortedTimes.length * 0.75)]?.time.duration;
    if (!q1 || !q3) return times;
    const iqr = q3 - q1;
    const maxValue = q3 + iqr * iqrMultiplier;

    return times.filter(({ time: { duration } }) => duration <= maxValue);
  };
  const timesWithoutOutliers = $derived(removeHighOutliersByIqr(times, 0.5));

  const minDuration = $derived(
    Math.min(...timesWithoutOutliers.map(({ time: { duration } }) => duration))
  );
  const maxDuration = $derived(
    Math.max(...timesWithoutOutliers.map(({ time: { duration } }) => duration))
  );

  // TODO: Support half and quarter second step size granularities when needed.
  const wholeNumbersBetween = ({ min, max }: { min: number; max: number }) =>
    range(Math.ceil(min), Math.floor(max));
  const ticks = $derived.by(() => {
    const steps = Math.floor(chartWidthWithoutPadding / tickLabelWidth);
    const wholeNumberRange = wholeNumbersBetween({ min: minDuration, max: maxDuration });
    const stepSize = Math.max(Math.floor(wholeNumberRange.length / steps), 1);
    const filteredRange = wholeNumberRange.filter((_, i) => i % stepSize === 0);

    // TODO: 0.66 is an approximation of text widths and surely won't always get this right.
    const first = filteredRange[0];
    if (first && minDuration - first < stepSize * 0.66) filteredRange.shift();
    const last = filteredRange.at(-1);
    if (last && maxDuration - last < stepSize * 0.66) filteredRange.pop();

    return filteredRange;
  });

  let cursorX: number | undefined = $state(undefined);
  const positions = $derived.by(() => {
    const positions: { x: number; y: number }[] = [];
    for (let i = 0; i < times.length; i++) {
      const entry = times[i]!;
      const x = Math.max(
        0,
        (1 - (entry.time.duration - minDuration) / (maxDuration - minDuration)) *
          chartWidthWithoutPadding
      );

      let y = 0;
      const distToPrevious = Math.abs(x - (positions[i - 1]?.x ?? Number.MIN_SAFE_INTEGER));
      const previousIconSize = times[i - 1]
        ? iconSizeWithRing(times[i - 1]!.position, entry.player.alias)
        : 0;
      const currentIconSize = iconSizeWithRing(entry.position, entry.player.alias);
      // TODO: I thought this should be (previousIconSize + currentIconSize) / 2 for all, but just previousIconSize is producing the desired results.
      if (distToPrevious < (previousIconSize + currentIconSize) / 2) {
        y =
          (positions[i - 1]!.y + previousIconSize) %
          (chartHeightWithoutPadding - previousIconSize / 2 - iconFloorYoffset);
      }

      positions.push({
        x,
        y
      });
    }
    return positions.map(({ x, y }) => ({ x, y: y + chartPadding.y + iconFloorYoffset })); // Internal relative sizes convenient for shifting --> absolute for the chart interior.
  });

  const highlightDistance = iconSize(1, '') / 2;
  const playerIdsHighlighted = $derived.by(() => {
    if (cursorX === undefined || times.length === 0) {
      return playerIdHighlighted ? [playerIdHighlighted] : [];
    }
    let minDist = Number.MAX_VALUE;
    let nearestEntries: (typeof times)[0][] = [];
    for (let i = 0; i < times.length; i++) {
      const dist = Math.abs(positions[i]!.x - cursorX);
      if (dist > highlightDistance) continue;
      if (dist < minDist) {
        minDist = dist;
        nearestEntries = [times[i]!];
      } else if (dist === minDist) {
        nearestEntries.push(times[i]!);
      }
    }
    const nearestIds = nearestEntries.map(({ player }) => player.id);
    return playerIdHighlighted ? [...new Set([playerIdHighlighted, ...nearestIds])] : nearestIds;
  });

  let playerIdHighlightedDebounced: string[] = $state([]);
  let playerIdsHighlightedDebouncedTimeoutId = -1;
  const highlightDebounceMs = 100;
  $effect(() => {
    // Delay release only. Also, not ideal to sync reactively with an effect, but functional for now.
    if (playerIdsHighlighted.length === 0) {
      if (playerIdsHighlightedDebouncedTimeoutId !== -1) return;
      playerIdsHighlightedDebouncedTimeoutId = setTimeout(() => {
        playerIdHighlightedDebounced = playerIdsHighlighted;
        playerIdsHighlightedDebouncedTimeoutId = -1;
      }, highlightDebounceMs) as unknown as number;
    } else {
      playerIdHighlightedDebounced = playerIdsHighlighted;
      clearTimeout(playerIdsHighlightedDebouncedTimeoutId);
      playerIdsHighlightedDebouncedTimeoutId = -1;
    }
  });
  onDestroy(() => clearTimeout(playerIdsHighlightedDebouncedTimeoutId));

  let isAfterStartup = $state(false); // This is used to prevent outliers from transitioning on chart start as well as ticks from transitioning on page resize.
  $effect(() => {
    // eslint-disable-next-line @typescript-eslint/no-unused-expressions
    times;
    const timeoutId = setTimeout(() => (isAfterStartup = true));
    return () => clearTimeout(timeoutId);
  });
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
  class="relative grid bg-linear-to-b from-base-900/50 to-base-900 select-none"
  onpointermove={(e) => {
    const rect = e.currentTarget.getBoundingClientRect();
    cursorX = e.clientX - rect.left - chartPadding.x;
  }}
  onpointerleave={() => (cursorX = undefined)}
  bind:clientWidth={chartWidth}
  style:height="{chartHeight}px"
  style:padding="{chartPadding.y}px {chartPadding.x}px">
  {#if times.length > 0 && chartWidth !== 0}
    {#each times.toReversed() as entry, timesReversedIndex (entry.player.id)}
      {@const hoverState = playerIdHighlightedDebounced.includes(entry.player.id)
        ? 'hovered'
        : playerIdHighlightedDebounced.length > 0
          ? 'other-hovered'
          : 'default'}
      {@const isAboveRange = entry.time.duration > maxDuration}
      {@const timesIndex = times.length - 1 - timesReversedIndex}
      {@const x = positions[timesIndex]?.x ?? 0}
      {@const y = positions[timesIndex]?.y ?? 0}
      {@const xExtensionAboveRange =
        isAboveRange && hoverState === 'hovered'
          ? cursorX !== undefined
            ? Math.min(
                (timesIndex -
                  times.findIndex(({ time: { duration } }) => duration === maxDuration)) *
                  (maxNameplateWidth + 1),
                chartWidthWithoutPadding - maxNameplateWidth / 2
              )
            : maxNameplateWidth / 2 // This state happens from hovering the table.
          : 0}
      <div
        class={[
          'group absolute flex flex-col items-center justify-center gap-1 font-mono text-xs',
          entry.player.alias === 'mur' && 'animate-[color-party_7s_infinite,float_5s_infinite]',
          { 1: 'text-div-gold', 2: 'text-div-silver', 3: 'text-div-bronze' }[entry.position] ??
            'text-content',
          hoverState === 'other-hovered' && 'brightness-50 contrast-90',
          hoverState === 'hovered' && isAboveRange && isAfterStartup
            ? '[transition:filter_150ms,left_50ms]'
            : isAboveRange && isAfterStartup
              ? 'transition-[left_50ms]'
              : hoverState !== 'hovered' && 'transition-[filter]'
        ]}
        style:bottom="{y}px"
        style:left="{(xExtensionAboveRange ? xExtensionAboveRange : x) +
          chartPadding.x -
          iconSize(entry.position, entry.player.alias) / 2}px"
        style:z-index={// z-index cleverness is done to pop hovered times to the top, sort faster times on top, but also sort the stack vertically for ties.
        10 +
          Math.floor(y) +
          chartHeightWithoutPadding * Math.floor(x) +
          (hoverState === 'hovered' ? chartHeightWithoutPadding * chartWidthWithoutPadding : 0)}
        in:fly|global={{
          delay:
            ((x + chartPadding.x - iconSize(entry.position, entry.player.alias) / 2) /
              chartWidthWithoutPadding) *
            (animateInDurationMs / 2),
          y: 10,
          duration: animateInDurationMs / 2
        }}>
        <!-- Crown. -->
        {#if entry.position === 1}
          <div
            class="pointer-events-none absolute -mt-11.5 ml-4.5 rotate-20 text-lg filter-[drop-shadow(0px_0px_1px_var(--color-base-900))_drop-shadow(0px_0px_1px_var(--color-base-900))] before:icon-[mdi--crown-outline]">
          </div>
        {:else if entry.player.alias === 'mur'}
          <div
            class="pointer-events-none absolute -mt-14.5 ml-5.5 animate-[float_5s_infinite] text-3xl text-success filter-[drop-shadow(0px_0px_1px_var(--color-base-900))_drop-shadow(0px_0px_1px_var(--color-base-900))] before:icon-[mdi--balloon]">
          </div>
          <div
            class="pointer-events-none absolute -mt-12 -ml-4.5 animate-[float_7s_infinite] text-3xl text-success filter-[drop-shadow(0px_0px_1px_var(--color-base-900))_drop-shadow(0px_0px_1px_var(--color-base-900))] before:icon-[mdi--balloon]">
          </div>
        {/if}

        <!-- Main icon. -->
        <div
          class={[
            'relative overflow-hidden rounded-full',
            entry.player.alias === 'mur'
              ? 'bg-base-800 shadow-[0px_0px_12px_2px_var(--color-success)] ring-2 shadow-success ring-success ring-offset-2 ring-offset-base-900'
              : ({
                  1: 'bg-base-800 shadow-[0px_0px_10px_2px_var(--color-div-gold)] ring-2 shadow-div-gold ring-div-gold ring-offset-2 ring-offset-base-900',
                  2: 'bg-base-800 shadow-[0px_0px_9px_2px_var(--color-div-silver)] ring-2 shadow-div-gold ring-div-silver ring-offset-2 ring-offset-base-900',
                  3: 'bg-base-800 shadow-[0px_0px_8px_2px_var(--color-div-bronze)] ring-2 shadow-div-gold ring-div-bronze ring-offset-2 ring-offset-base-900'
                }[entry.position] ??
                (iconStyle(entry.position, entry.player.alias) === 'avatar'
                  ? isAboveRange
                    ? 'bg-base-800 ring-2 ring-error ring-offset-2 ring-offset-base-900'
                    : 'bg-base-800 ring-2 ring-[color-mix(in_oklab,var(--color-content),var(--color-base-900)_15%)] ring-offset-2 ring-offset-base-900'
                  : isAboveRange
                    ? 'bg-error'
                    : 'bg-content'))
          ]}
          style:width="{iconSize(entry.position, entry.player.alias)}px"
          style:height="{iconSize(entry.position, entry.player.alias)}px">
          {#if iconStyle(entry.position, entry.player.alias) === 'avatar'}
            <img
              src={entry.player.avatar_url}
              alt={entry.player.alias}
              class="animate-pulse overflow-hidden rounded-full bg-base-700"
              {@attach (element) => {
                // eslint-disable-next-line @typescript-eslint/no-unused-expressions
                entry.player.avatar_url;
                const loadListener = () => {
                  element.classList.remove('bg-base-700', 'animate-pulse');
                };
                const errorListener = () => {
                  element.classList.remove('bg-base-700', 'animate-pulse');
                  element.style.opacity = '0';
                };
                element.addEventListener('load', loadListener);
                element.addEventListener('error', errorListener);
                return () => {
                  element.removeEventListener('load', loadListener);
                  element.removeEventListener('error', errorListener);
                };
              }} />
          {/if}
        </div>

        <!-- Hover nameplate. -->
        {#if hoverState === 'hovered'}
          <!-- This element's height + empty padding. -->
          {@const nameplateHeight = 36 + 4}
          <!-- TODO: Could prevent these from going off the chart edges horizontally too. -->
          <div
            class="pointer-events-none absolute flex justify-center overflow-hidden"
            style:margin-bottom="{(iconSizeWithRing(entry.position, entry.player.alias) +
              nameplateHeight) *
              (y > chartHeightWithoutPadding - nameplateHeight ? -1 : 1)}px"
            style:width="{maxNameplateWidth}px">
            <div
              class="grid w-fit grid-rows-2 overflow-hidden rounded-box bg-base-900 px-1.5 py-0.5 text-center whitespace-nowrap">
              <div class="max-w-full overflow-hidden text-ellipsis whitespace-nowrap">
                {formatDuration(entry.time.duration)}
              </div>
              <div class="max-w-full overflow-hidden text-ellipsis whitespace-nowrap">
                <span class="font-bold">#{entry.position}</span>
                {entry.player.alias}
              </div>
            </div>
          </div>
        {/if}
      </div>

      <!-- Hover stem. -->
      {#if xExtensionAboveRange === 0 && hoverState === 'hovered'}
        <div
          class="absolute bottom-5.5 border-r border-r-base-700"
          style:top="{chartHeightWithoutPadding -
            y +
            iconSize(entry.position, entry.player.alias) / 2 + // TODO: This isn't calculated quite right but... you can't tell for now.
            chartPadding.y}px"
          style:left="{x + chartPadding.x}px">
        </div>
      {/if}
    {/each}

    <!-- Footer. -->
    <div class="relative mt-auto mb-2.5 font-mono text-xs">
      <!-- Base line. -->
      <div
        class="absolute h-px w-full bg-linear-to-r from-transparent via-base-700 to-transparent"
        in:fly|global={{
          y: 5,
          duration: animateInDurationMs / 2
        }}>
      </div>

      <!-- Max time tick. -->
      <div
        class="pointer-events-none absolute left-0 mt-1 flex justify-start"
        in:fly|global={{
          y: 5,
          duration: animateInDurationMs / 2
        }}>
        {formatDuration(maxDuration)}
        <div class="absolute -top-2 h-2 border-r border-r-base-700"></div>
      </div>

      <!-- Min time tick. -->
      <!-- TODO: -mr-px hack means something is off by one. -->
      <div
        class="pointer-events-none absolute right-0 mt-1 -mr-px flex justify-end"
        in:fly|global={{
          delay: animateInDurationMs / 2,
          y: 5,
          duration: animateInDurationMs / 2
        }}>
        {formatDuration(minDuration)}
        <div class="absolute -top-2 h-2 border-r border-r-base-700"></div>
      </div>

      <!-- Non-edge ticks. -->
      {#each ticks as tickDuration, i (i)}
        {@const leftPercent = (maxDuration - tickDuration) / (maxDuration - minDuration)}
        <div
          class="absolute mt-1 flex justify-center"
          style:left="{leftPercent * chartWidthWithoutPadding}px"
          in:fly|global={isAfterStartup
            ? { duration: 0 }
            : {
                delay: leftPercent * (animateInDurationMs / 2),
                y: 5,
                duration: animateInDurationMs / 2
              }}>
          <div class="pointer-events-none absolute">
            <!-- TODO: This won't work for really short times or finer granularity. -->
            {formatDuration(tickDuration).split('.')[0]}
          </div>
          <div class="pointer-events-none absolute -top-2 h-2 border-r border-r-base-700"></div>
        </div>
      {/each}
    </div>
  {:else}
    <div class="flex items-center justify-center opacity-65">no times submitted</div>
  {/if}
</div>

<style lang="postcss">
  @keyframes -global-float {
    0%,
    100% {
      transform: translateY(-5%);
      animation-timing-function: cubic-bezier(0.8, 0, 1, 1);
    }
    50% {
      transform: none;
      animation-timing-function: cubic-bezier(0, 0, 0.2, 1);
    }
  }

  @keyframes -global-color-party {
    0% {
      filter: hue-rotate(0deg);
    }
    50% {
      filter: hue-rotate(180deg);
    }
    100% {
      filter: hue-rotate(360deg);
    }
  }
</style>
