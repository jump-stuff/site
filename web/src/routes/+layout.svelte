<script lang="ts">
  import './layout.css';
  import favicon from '$lib/assets/favicon.svg';
  import Nav from '$lib/components/layout/nav/Nav.svelte';
  import type { PageData } from './$types';
  import type { Snippet } from 'svelte';
  import Logo from '$lib/components/layout/Logo.svelte';
  import NavigationProgress from '$lib/components/display/NavigationProgress.svelte';

  type Props = {
    data: PageData;
    children: Snippet;
  };
  let { data, children }: Props = $props();
</script>

<svelte:head><link rel="icon" href={favicon} /></svelte:head>

<NavigationProgress />

<div class="pointer-events-none flex justify-center">
  <div class="pointer-events-auto flex w-full max-w-5xl flex-col gap-1 bg-base-900">
    <Nav session={data.session} />
    <main class="relative z-10 flex flex-col bg-base-800 pb-3">
      {@render children()}
    </main>
  </div>
</div>

<div class="fixed -bottom-36 -left-36">
  <Logo />
</div>
