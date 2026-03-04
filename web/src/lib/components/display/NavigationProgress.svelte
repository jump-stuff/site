<script>
  // credit: https://rachidboudjelida.com/articles/smooth-navigation-progress-bar-sveltekit
  import { navigating } from '$app/state';
  import { cubicOut } from 'svelte/easing';
  import { Tween } from 'svelte/motion';
  import { fade } from 'svelte/transition';
  import { untrack } from 'svelte';

  // We use Tween to animate the value change.
  const p = new Tween(0, {
    duration: 200,
    easing: cubicOut
  });

  /**
   * @type {ReturnType<typeof setTimeout> | null}
   */
  let timeout = null; // we use it to track our pending timeouts
  let isVisible = $state(false); // we use it to track the visibility of the progress bar
  function reset() {
    if (timeout) {
      clearTimeout(timeout);
      timeout = null;
    }
    p.set(0, { duration: 0 });
  }

  function increase() {
    const progressLeft = 1 - p.current;
    p.set(p.current + progressLeft * 0.04);
    if (p.current > 1) {
      p.set(1);
    }
    if (navigating.complete) {
      timeout = setTimeout(increase, 50);
    } else {
      p.set(1);
      timeout = setTimeout(() => {
        isVisible = false;
        p.set(0, { duration: 0 });
      }, 150);
    }
  }

  $effect(() => {
    if (navigating.complete) {
      console.log('calling increase');
      untrack(() => {
        isVisible = true;
        reset();
        increase();
      });
    }
  });
</script>

{#if isVisible}
  <!-- we use the fade animation from svelte/transition  and use the tween.current for the value -->
  <div style:width={`${p.current * 100}%`} in:fade={{ duration: 350 }} out:fade={{ duration: 350 }}>
  </div>
{/if}

<style>
  div {
    position: fixed;
    top: 0;
    z-index: 99999;
    left: 0;
    height: 2px;
    width: 0%;
    appearance: none;
    border: none;
    outline: none; /* firefox has a default outline */
    background-color: #b4befe;
  }
</style>
