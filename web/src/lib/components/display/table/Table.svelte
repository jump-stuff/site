<script lang="ts">
  import type { Snippet } from 'svelte';

  type Props = {
    data: (object & { onmouseover?: (e: Event) => void; onmouseout?: (e: Event) => void })[];
    header: Snippet;
    row: Snippet<[any, number]>;
  };

  let { data, header, row }: Props = $props();
</script>

<table class="w-full border-separate border-spacing-y-px rounded-layout text-sm">
  <!-- header -->
  <thead class="text-nowrap text-primary/75">
    <tr>
      <!-- <th class="w-number">{content}</th> -->
      {@render header()}
    </tr>
  </thead>

  <!-- rows -->
  <tbody>
    {#each data as object, i (i)}
      <tr
        onmouseover={object.onmouseover}
        onfocus={object.onmouseover}
        onmouseout={object.onmouseout}
        onblur={object.onmouseout}>
        {@render row(object, i + 1)}
        <!-- <td>{content}</td> -->
      </tr>
    {/each}
  </tbody>
</table>
