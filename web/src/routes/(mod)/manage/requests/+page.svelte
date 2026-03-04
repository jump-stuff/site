<script lang="ts">
  import { Client } from '$lib/api/api';
  import PlayerHeader from '$lib/components/display/player/PlayerHeader.svelte';
  import Table from '$lib/components/display/table/Table.svelte';
  import TablePlayer from '$lib/components/display/table/TablePlayer.svelte';
  import TemporalDate from '$lib/components/display/TemporalDate.svelte';
  import Section from '$lib/components/layout/Section.svelte';
  import { ApiPaths, type Player, type RequestWithPlayer } from '$lib/schema';
  import jf from '$lib/assets/logo/jf.png';
  import Input from '$lib/components/input/Input.svelte';
  import Errors from '$lib/components/input/Errors.svelte';
  import Select from '$lib/components/input/Select.svelte';
  import { divs } from '$lib/helpers/divs';
  import Button from '$lib/components/input/Button.svelte';

  let selectedPlayer: Player = $state({
    id: '0',
    role: 'player',
    alias: 'select a request',
    avatar_url: jf,
    class_pref: 'Soldier',
    created_at: '2026-01-01T00:00:00Z'
  });

  let refreshRequests = $state(false);

  let oerror: OpenAPIError = $state(undefined);
</script>

<PlayerHeader player={selectedPlayer} />
{#if selectedPlayer.id !== '0'}
  <Section label="update player">
    <Errors {oerror} />
    <Input
      label="alias"
      type="text"
      placeholder={selectedPlayer.alias}
      onsubmit={async (value) => {
        const resp = await Client.POST(ApiPaths.update_alias, {
          params: { path: { player_id: selectedPlayer.id, alias: value } }
        });
        oerror = resp.error;
        if (resp.response.ok) {
          selectedPlayer.alias = value;
        }
        return resp.response.ok;
      }} />
    <Select
      label="soldier div"
      type="text"
      placeholder={selectedPlayer.soldier_div}
      options={divs.concat('none')}
      onsubmit={async (value) => {
        const resp = await Client.POST(ApiPaths.update_div, {
          params: { path: { player_id: selectedPlayer.id, player_class: 'Soldier', div: value } }
        });
        oerror = resp.error;
        if (resp.response.ok) {
          selectedPlayer.soldier_div = value;
        }
        return resp.response.ok;
      }} />
    <Select
      label="demo div"
      type="text"
      placeholder={selectedPlayer.demo_div}
      options={divs.concat('none')}
      onsubmit={async (value) => {
        const resp = await Client.POST(ApiPaths.update_div, {
          params: { path: { player_id: selectedPlayer.id, player_class: 'Demo', div: value } }
        });
        oerror = resp.error;
        if (resp.response.ok) {
          selectedPlayer.demo_div = value;
        }
        return resp.response.ok;
      }} />
  </Section>
{/if}

<Section label="request list">
  {#key refreshRequests}
    {#await Client.GET(ApiPaths.get_all_requests)}
      <span></span>
    {:then { data: rwps }}
      {#if rwps?.length}
        <Table data={rwps}>
          {#snippet header()}
            <th class="w-32">kind</th>
            <th class="w-div">content</th>
            <th></th>
            <th class="w-date">request created..</th>
            <th class="w-0"></th>
          {/snippet}
          {#snippet row({ request, player }: RequestWithPlayer)}
            <td>{request.kind}</td>
            <td>{request.content}</td>
            <td
              class="cursor-pointer hover:underline"
              onclick={() => {
                selectedPlayer = player;
                // jump to top
                window.scrollTo({ top: 0, behavior: 'smooth' });
              }}><TablePlayer {player} /></td>
            <td class="table-date"><TemporalDate datetime={request.created_at} /></td>
            <td
              ><Button
                table={true}
                onsubmit={async () => {
                  const resp = await Client.POST(ApiPaths.resolve_request, {
                    params: { path: { request_id: request.id } }
                  });
                  oerror = resp.error;
                  if (resp.response.ok) {
                    refreshRequests = !refreshRequests;
                  }
                  return resp.response.ok;
                }}><span class="icon-[mdi--check]"></span></Button
              ></td>
          {/snippet}
        </Table>
      {:else}
        <span>no requests!</span>
      {/if}
    {/await}
  {/key}
</Section>
