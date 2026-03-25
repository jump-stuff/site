<script lang="ts">
  import { type Player } from '$lib/schema';
  import ClassSelect from '$lib/components/display/ClassSelect.svelte';
  import Div from '$lib/components/display/Div.svelte';
  import ExternalLink from '$lib/components/display/ExternalLink.svelte';
  import Flag from '$lib/components/display/Flag.svelte';
  import tempus from '$lib/assets/components/profile/tempus.png';
  import plaza from '$lib/assets/components/profile/plaza.png';
  import no_map from '$lib/assets/no_map.png';
  import Launcher from '../Launcher.svelte';
  import { PUBLIC_JUMP_IMAGES_URL } from '$env/static/public';

  type Props = {
    player: Player;
    class_pref?: string;
  };

  let { player, class_pref = $bindable(player.class_pref) }: Props = $props();
</script>

<div class="relative z-10 h-56 flex-col overflow-hidden bg-base-900">
  {#if player.map_pref}
    <div
      class="h-36 w-full mask-b-from-98% bg-cover bg-center"
      style:background-image={`url("${PUBLIC_JUMP_IMAGES_URL}/maps/${player.map_pref}.webp")`}>
    </div>
  {:else}
    <div class="h-36 w-full mask-b-from-98%">
      <div class="size-full mask-x-from-50% mask-x-to-95%">
        <div
          class="filter-purelavender size-[1476px] rotate-5 animate-[nomap_360s_linear_infinite] bg-size-[30%] bg-repeat"
          style:background-image={`url(${no_map})`}>
        </div>
      </div>
    </div>
  {/if}
  <!-- avatar -->
  <img
    class="absolute top-22 left-4 z-10 h-24 rounded-box object-cover"
    src={player.avatar_url}
    alt="" />
  <div class="relative -top-5.5 flex flex-col gap-1">
    <div class="flex">
      <div class="relative w-fit rounded-tr-box bg-base-900 pr-2 pl-30 text-lg">
        <span>{player.alias}</span>
        {#if class_pref === 'Soldier'}
          <div class="absolute -top-3 -right-16 rotate-15">
            <Launcher launcher={player.launcher_pref} />
          </div>
        {/if}
      </div>
    </div>
    <div class="ml-30 flex items-center gap-2">
      <!-- div -->
      {#if class_pref === 'Soldier'}
        {#if player.soldier_div}
          <Div div={player.soldier_div} playerClass="Soldier" />
        {:else}
          <Div div="Divless" playerClass="Soldier" />
        {/if}
      {:else if player.demo_div}
        <Div div={player.demo_div} playerClass="Demo" />
      {:else}
        <Div div="Divless" playerClass="Demo" />
      {/if}
      <div class="flex gap-1">
        <Flag code={player.country_code} country={player.country} />
      </div>
    </div>
    {#if player.tempus_id}
      <div class="mt-2 ml-4 flex gap-2">
        <ExternalLink
          label="Tempus"
          src={tempus}
          href={`https://tempus2.xyz/players/${player.tempus_id}`}
          newTab={true} />
        <ExternalLink
          label="Plaza"
          src={plaza}
          href={`https://tempusplaza.com/players/${player.tempus_id}`}
          newTab={true} />
      </div>
    {/if}

    <!-- class select -->
    <div class="absolute top-5 right-2 ml-auto flex">
      <ClassSelect bind:player_class={class_pref} />
    </div>
  </div>
</div>
