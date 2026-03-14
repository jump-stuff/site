<script lang="ts">
  import { browser } from '$app/environment';
  import { goto } from '$app/navigation';
  import type { Attachment } from 'svelte/attachments';
  import tree from '$lib/assets/tree.jpg';

  if (browser && localStorage.getItem('unwrapped-gift')) {
    goto('/');
  }

  let isUseItemModalOpen = $state(false);
  let unboxCountdownDigit: number | undefined = $state(undefined);
  const countdown = () => {
    setTimeout(() => {
      if (unboxCountdownDigit === undefined) return;
      unboxCountdownDigit--;
      if (unboxCountdownDigit > 0) {
        countdown();
      } else {
        unboxCountdownDigit = undefined;
        localStorage.setItem('unwrapped-gift', 'true');

        eggAcquired = true;
        setTimeout(() => {
          showTree = true;
          setTimeout(() => {
            showEggMessage = true;
          }, 3000);
        }, 2000);
      }
    }, 1000);
  };

  let eggAcquired = $state(false);
  let showTree = $state(false);
  let showEggMessage = $state(false);

  const typewriter =
    (message: string): Attachment =>
    (element: Element) => {
      if (message === '') {
        element.innerHTML = message;
        return;
      }

      const textChunks = message.split(' ');
      let chunkIndex = 0;
      let charIndex = 0;
      let currentTextWithoutSpans = '';

      const updateInnerHtml = (currentText: string) => {
        currentText = currentText.replaceAll('⏸', '');
        element.innerHTML = '&#8203;' + currentText;
      };

      let timeoutId = -1;
      const run = () => {
        const text = textChunks[chunkIndex]!;
        const delay = 40;

        if (charIndex < text.length) {
          const nextChar = text[charIndex];
          if (nextChar !== '⏸') {
            currentTextWithoutSpans += nextChar;
          }
          charIndex++;
          timeoutId = setTimeout(run, delay) as unknown as number;
          updateInnerHtml(
            `${currentTextWithoutSpans}<span style="visibility: hidden">${text.slice(charIndex)}</span>`
          );
        } else if (chunkIndex + 1 < textChunks.length) {
          charIndex = 0;
          chunkIndex++;
          currentTextWithoutSpans += ' ';
          timeoutId = setTimeout(run, 20) as unknown as number;
          updateInnerHtml(
            `${currentTextWithoutSpans}<span style="visibility: hidden">${textChunks[chunkIndex]}</span>`
          );
        }
      };

      run();

      return () => {
        clearTimeout(timeoutId);
      };
    };
</script>

{#if eggAcquired}
  <div
    class="grid h-screen grid-rows-[1fr_auto] overflow-hidden bg-[black] p-5 font-delta text-[white]">
    {#if showTree}
      <img src={tree} alt="tree" class="m-auto" />
    {/if}
    {#if showEggMessage}
      <div
        class="mx-auto h-60 w-200 max-w-full overflow-hidden border-5 border-[white] px-8 py-5 text-[2.25rem]"
        {@attach typewriter('* (You received an Egg.)')}>
      </div>
    {:else}
      <div class="h-60"></div>
    {/if}
  </div>
{:else}
  <div class="h-screen overflow-hidden bg-[#292526] font-tf2build text-[#fbf1d7]">
    <div
      class="mx-auto grid h-full max-w-5xl grid-rows-[auto_1fr_auto] items-center gap-x-4 overflow-hidden">
      <div class="mt-10 mb-2 grid grid-rows-2 place-items-center gap-10">
        <div class="text-5xl">NEW ITEM ACQUIRED!</div>
        <div class="text-3xl">YOU <span class="text-[#b85345]">FOUND</span>:</div>
      </div>
      <div
        class="relative mx-20 grid h-full grid-cols-2 overflow-hidden rounded-xl border-3 border-[#776b5f] bg-[#332f2e]">
        <div class="absolute top-0 left-1.5 grid grid-rows-[auto-auto]">
          <div class="text-lg">ITEM</div>
          <div class="-mt-2 text-5xl">#1</div>
        </div>
        <div class="absolute top-1 right-2.5 grid size-16 grid-rows-[auto-auto] bg-[black] p-1.75">
          <img
            src="https://avatars.steamstatic.com/e5af6e00bb33691c656311a595f14e95994a6f4e_full.jpg"
            alt="gift receiver" />
        </div>
        <img
          src="https://wiki.teamfortress.com/w/images/d/d2/Backpack_Secret_Saxton.png"
          alt="gift"
          class="m-auto h-full overflow-hidden object-contain" />
        <div class="flex flex-col items-center justify-center overflow-hidden font-tf2secondary">
          <div class="-mb-1.25 font-tf2build text-xl text-[gold]">Secret Saxton</div>
          <div class="text-[#8f8a83]">Level 100 Gift</div>
          <div class="-mt-0.5 font-semibold text-[#7ea9d1]">Imbued with a giftly aura</div>
          <div class="h-5"></div>
          <div class="max-w-80 text-center">
            When used, we're really not sure what happens. No one has opened it before. That would
            defeat the point.
          </div>
          <!-- <div class="h-5"></div>
        <div>( Not Tradable )</div> -->
          <div class="h-5"></div>
          <div class="text-[#00A000]">This is a limited use item. Uses: 1</div>
        </div>
      </div>
      <div class="mt-20 mb-10 grid grid-cols-2 gap-30 text-[22px]">
        <button
          onclick={() => {
            if (history.length === 1) {
              goto('/');
            } else {
              history.back();
            }
          }}
          class="cursor-pointer rounded-lg bg-[#776b5f] pt-3.5 pb-2 hover:bg-[#91493b]"
          >Return</button>
        <button
          onclick={() => {
            isUseItemModalOpen = true;
          }}
          class="cursor-pointer rounded-lg bg-[#776b5f] pt-3.5 pb-2 hover:bg-[#91493b]">Use</button>
      </div>

      {#if isUseItemModalOpen}
        <button
          onclick={() => {
            isUseItemModalOpen = false;
          }}
          class="absolute inset-0"
          type="button"
          aria-label="scrim"></button>
        <div
          class="absolute inset-20 m-auto grid max-h-150 max-w-2xl grid-rows-[auto_1fr_auto] justify-center gap-4 rounded-xl border-3 border-[#776b5f] bg-[#24201B] px-15 py-8">
          <div class="text-center text-[3.25rem]">USE ITEM?</div>
          <div class="text-center text-3xl text-[#b85345]">
            ARE YOU SURE YOU WANT TO USE SECRET SAXTON? IT HAS 1 USE(S) BEFORE IT WILL BE REMOVED
            FROM YOUR INVENTORY.
          </div>
          <div class="-mx-11 -mb-2 grid grid-cols-[1.5fr_1fr] gap-3 text-[22px]">
            <button
              onclick={() => {
                isUseItemModalOpen = false;
                unboxCountdownDigit = 5;
                countdown();
              }}
              class="cursor-pointer rounded-lg bg-[#776b5f] pt-3.5 pb-2 hover:bg-[#91493b]"
              >OK</button>
            <button
              onclick={() => {
                isUseItemModalOpen = false;
              }}
              class="cursor-pointer rounded-lg bg-[#776b5f] pt-3.5 pb-2 hover:bg-[#91493b]"
              >CANCEL</button>
          </div>
        </div>
      {/if}
      {#if unboxCountdownDigit !== undefined}
        <div class="absolute inset-0 bg-[#24201B]/50"></div>
        <div
          class="absolute inset-20 m-auto grid max-h-63 max-w-110 grid-rows-[auto_auto_auto] justify-center gap-3 rounded-xl border-3 border-[#776b5f] bg-[#24201B] px-13 py-15">
          <div class="text-center text-2xl text-[#b85345]">UNWRAPPING YOUR LOOT</div>
          <div class="text-center text-xl text-[#b85345]">
            {'.'.repeat(1 + (unboxCountdownDigit % 3))}
          </div>
          <div class="-mb-3 text-center text-5xl text-[#b85345]">{unboxCountdownDigit}</div>
        </div>
      {/if}
    </div>
  </div>
{/if}
