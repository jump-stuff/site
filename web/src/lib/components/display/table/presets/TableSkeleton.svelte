<script lang="ts">
  import Table from '../Table.svelte';

  type Props = {
    rows: number;
  };

  let { rows }: Props = $props();

  const data: object[] = $derived.by(() => {
    const a: object[] = [];
    for (let i = 0; i < rows; i++) {
      a.push({});
    }
    return a;
  });
</script>

<Table {data}>
  {#snippet header()}
    <th></th>
  {/snippet}
  {#snippet row()}
    <td class="animate-[pulse_3s_infinite]"></td>{/snippet}
</Table>
