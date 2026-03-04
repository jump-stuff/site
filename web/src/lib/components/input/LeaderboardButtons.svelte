<script lang="ts">
  import type { Leaderboard } from '$lib/schema';
  import Div from '../display/Div.svelte';

  type Props = {
    leaderboards: Leaderboard[];
    selected: number;
    onclick?: () => void;
  };

  let { leaderboards, selected = $bindable(), onclick = () => {} }: Props = $props();
</script>

<div class="flex grow overflow-hidden rounded-t-box">
  {#each leaderboards as l}
    <button
      class="flex h-9 grow cursor-pointer items-center justify-center transition-all
          {selected === l.id
        ? `bg-div-${l.div?.toLowerCase()}/25`
        : 'bg-base-800 opacity-50 hover:opacity-100'}"
      onclick={() => {
        selected = l.id;
        onclick();
      }}>
      <Div div={l.div} />
    </button>
  {/each}
</div>
