<script lang="ts">
  import Div from '../Div.svelte';
  import { PUBLIC_JUMP_IMAGES_URL } from '$env/static/public';

  type Props = {
    map: string;
    div?: string;
  };

  let { map, div = '' }: Props = $props();
</script>

<div
  class="group relative grid size-full scale-x-105 place-content-center text-shadow-xs/100 text-shadow-base-900 not-first:mask-l-from-97% not-last:mask-r-from-97%">
  <span class="relative z-10 truncate">{map}</span>
  <img
    class="map-img absolute size-full object-cover brightness-75 transition-all group-hover:brightness-100"
    src="{PUBLIC_JUMP_IMAGES_URL}/maps/{map}.jpg"
    alt=""
    draggable="false"
    loading="lazy" />
  {#if div}
    <div class="absolute ml-4 flex h-full items-center"><Div {div} /></div>
  {/if}
</div>
