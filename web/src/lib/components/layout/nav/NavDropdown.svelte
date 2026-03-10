<script lang="ts">
  import { slide } from 'svelte/transition';

  import type { Session } from '$lib/schema';

  type Props = {
    session: Session;
  };

  let { session }: Props = $props();
</script>

<div
  in:slide={{ duration: 200 }}
  class="absolute top-0 right-0 -z-10 flex w-full cursor-default flex-col gap-px overflow-hidden rounded-b-box border border-t-0 border-primary/75 bg-base-900 p-1 pt-22 backdrop-blur-md"
  data-nav="true">
  {@render page(session.alias, `/players/${session.id}`)}
  <hr class="relative left-1/24 my-px w-11/12 text-base-700" />
  {#if session.role === 'admin' || session.role === 'mod' || session.role === 'dev'}
    {@render page('manage', '/manage')}
    <hr class="relative left-1/24 my-px w-11/12 text-base-700" />
  {/if}
  {@render page('settings', '/settings')}
  {@render page('logout', '/signout')}
</div>

{#snippet page(label: string, href: string)}
  <a
    class="truncate rounded-box pl-2 hover:bg-base-800"
    {href}
    data-nav="true"
    data-sveltekit-preload-data={label === 'logout' ? 'tap' : 'hover'}>
    {label}
  </a>
{/snippet}
