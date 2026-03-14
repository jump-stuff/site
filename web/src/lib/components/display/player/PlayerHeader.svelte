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
  import confetti from 'canvas-confetti';
  import { onDestroy } from 'svelte';
  import { browser } from '$app/environment';

  type Props = {
    player: Player;
    class_pref?: string;
  };

  let { player, class_pref = $bindable(player.class_pref) }: Props = $props();

  const hasGift = browser && !localStorage.getItem('unwrapped-gift');
  // svelte-ignore state_referenced_locally
  if (player.alias === 'mur' && browser && hasGift) {
    const myCanvas = document.createElement('canvas');
    document.body.appendChild(myCanvas);
    myCanvas.classList.add('fixed', 'top-0', 'h-full', 'w-full', 'z-60', 'pointer-events-none');
    const myConfetti = confetti.create(myCanvas, {
      resize: true,
      useWorker: true
    });
    const duration = 3 * 1000;
    const end = Date.now() + duration;
    const frame = () => {
      myConfetti({
        particleCount: 5,
        angle: 60,
        spread: 55,
        origin: { x: 0 }
      });
      myConfetti({
        particleCount: 5,
        angle: 120,
        spread: 55,
        origin: { x: 1 }
      });
      if (Date.now() < end) requestAnimationFrame(frame);
    };
    frame();
    onDestroy(() => {
      myCanvas.remove();
    });
  }
</script>

<div class="relative z-10 h-56 flex-col overflow-hidden bg-base-900">
  {#if player.map_pref}
    <div
      class="h-36 w-full mask-b-from-98% bg-cover bg-center"
      style:background-image={`url("https://tempusplaza.com/map-backgrounds/${player.map_pref}.webp")`}>
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
    class={[
      'absolute top-22 left-4 z-10 h-24 rounded-box object-cover',
      player.alias === 'mur' &&
        'animate-[color-party_7s_infinite,float_5s_infinite] shadow-[0px_0px_12px_2px_var(--color-success)] ring-2 shadow-success ring-success ring-offset-2 ring-offset-base-900'
    ]}
    src={player.avatar_url}
    alt="" />
  {#if player.alias === 'mur'}
    <div class="relative animate-[color-party_7s_infinite,float_5s_infinite]">
      <div
        class="pointer-events-none absolute -top-27 left-3 animate-[float_5s_infinite] text-5xl text-success filter-[drop-shadow(0px_0px_1px_var(--color-base-900))_drop-shadow(0px_0px_20px_var(--color-base-900))] before:icon-[mdi--balloon]">
      </div>
      <div
        class="pointer-events-none absolute -top-28 left-16 animate-[float_6s_infinite] text-5xl text-success filter-[drop-shadow(0px_0px_1px_var(--color-base-900))_drop-shadow(0px_0px_20px_var(--color-base-900))] before:icon-[mdi--balloon]">
      </div>
    </div>
  {/if}
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
      {#if player.alias === 'mur' && hasGift}
        <a
          href="/secret"
          class="mt-0.5 cursor-pointer px-2 pt-2 text-5xl opacity-25 transition-opacity before:icon-[mdi--gift-outline] hover:opacity-75"
          aria-label="secret">
        </a>
      {/if}
      <ClassSelect bind:player_class={class_pref} />
    </div>
  </div>
</div>

<!-- TODO: Dedupe this. -->
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
