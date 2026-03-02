<script lang="ts">
  import Div from '$lib/components/display/Div.svelte';
  import Launcher from '$lib/components/display/Launcher.svelte';
  import Table from '$lib/components/display/table/Table.svelte';
  import TablePlayer from '$lib/components/display/table/TablePlayer.svelte';
  import TemporalDate from '$lib/components/display/TemporalDate.svelte';
  import Content from '$lib/components/layout/Content.svelte';
  import Section from '$lib/components/layout/Section.svelte';
  import { comparePlayers } from '$lib/helpers/divs';
  import { twTableGradients } from '$lib/helpers/times';
  import type { Player } from '$lib/schema';
  import type { PageData } from './$types';

  type Props = {
    data: PageData;
  };
  let { data }: Props = $props();
  let sort_class: string = $state('Soldier');

  function sortPlayers(players: Player[]): Player[] {
    return players.sort((a, b) => comparePlayers(a, b, sort_class));
  }
</script>

{#if data.players}
  <Content>
    <Section label="ladder" indent="/">
      <span class="text-content/75">players are sorted by div, then alphabetically.</span>
      {#key sort_class}
        <Table data={sortPlayers(data.players)}>
          {#snippet header()}
            <th class="w-rank"></th>
            <th></th>
            {#if sort_class === 'Soldier'}
              <th class="w-div"></th>
            {/if}
            <th
              class="w-div cursor-pointer text-start hover:text-primary {sort_class === 'Soldier'
                ? 'text-primary  after:font-normal after:text-content after:content-["_v"]'
                : ''}"
              onclick={() => {
                sort_class = 'Soldier';
              }}>soldier</th>
            <th
              class="w-div cursor-pointer text-start hover:text-primary {sort_class === 'Demo'
                ? 'text-primary  after:font-normal after:text-content after:content-["_v"]'
                : ''}"
              onclick={() => {
                sort_class = 'Demo';
              }}>demo</th>
            <th class="w-date">here since..</th>
          {/snippet}
          {#snippet row(player: Player, i)}
            <td
              class={twTableGradients.get(
                `r${sort_class === 'Soldier' ? player.soldier_div?.toLowerCase() : player.demo_div?.toLowerCase()}`
              )}></td>
            <td
              class={twTableGradients.get(
                `t${sort_class === 'Soldier' ? player.soldier_div?.toLowerCase() : player.demo_div?.toLowerCase()}`
              )}><TablePlayer {player} link={true} /></td>
            {#if sort_class === 'Soldier'}
              <td class="h-6"><Launcher launcher={player.launcher_pref ?? ''} /></td>
            {/if}
            <td class="text-start"><Div div={player.soldier_div} /></td>
            <td class="text-start"><Div div={player.demo_div} /></td>
            <td class="table-date"><TemporalDate datetime={player.created_at} player={true} /></td>
          {/snippet}
        </Table>
      {/key}
    </Section>
  </Content>
{/if}
