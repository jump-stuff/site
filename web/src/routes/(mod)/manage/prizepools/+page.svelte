<script lang="ts">
  import { Client } from '$lib/api/api';
  import EventHeader from '$lib/components/display/EventHeader.svelte';
  import ExternalLink from '$lib/components/display/ExternalLink.svelte';
  import PlayerHeader from '$lib/components/display/player/PlayerHeader.svelte';
  import TableEvents from '$lib/components/display/table/presets/TableEvents.svelte';
  import TablePlayer from '$lib/components/display/table/TablePlayer.svelte';
  import Button from '$lib/components/input/Button.svelte';
  import Errors from '$lib/components/input/Errors.svelte';
  import Input from '$lib/components/input/Input.svelte';
  import LeaderboardButtons from '$lib/components/input/LeaderboardButtons.svelte';
  import Select from '$lib/components/input/Select.svelte';
  import Section from '$lib/components/layout/Section.svelte';
  import { divs } from '$lib/helpers/divs';
  import { datetimeToMs, validDateTime } from '$lib/helpers/temporal';
  import {
    ApiPaths,
    type Event,
    type EventWithLeaderboards,
    type Leaderboard,
    type Prize
  } from '$lib/schema';
  import { Temporal } from 'temporal-polyfill';

  let event_id: number = $state(0);
  let player_class: 'Soldier' | 'Demo' = $state('Soldier');
  let event_kind: string = $state('event');
  let event_kind_id: number = $state(0);
  let visible_date: string = $state('');
  let visible_time: string = $state('');
  let start_date: string = $state('');
  let start_time: string = $state('');

  const visible_at: string = $derived(validDateTime(`${visible_date}T${visible_time}:00Z`));
  const starts_at: string = $derived(validDateTime(`${start_date}T${start_date}:00Z`));

  const divless: Leaderboard[] = $derived.by(() => {
    return [
      {
        id: 0,
        event_id: event_id,
        map: ''
      }
    ];
  });

  let leaderboards: Leaderboard[] = $state([]);
  let selectedLeaderboardID: number = $derived(leaderboards?.at(0)?.id ?? 0);

  // key to fetch updated events
  let reloadEvents: boolean = $state(true);
  let oerror: OpenAPIError = $state(undefined);

  // update event with the selected event
  function loadEvent({ event: e, leaderboards: l }: EventWithLeaderboards): void {
    event_id = e.id;
    player_class = e.player_class;
    event_kind = e.kind;
    event_kind_id = e.kind_id;
    visible_date = e.visible_at.substring(0, e.visible_at.indexOf('T'));
    visible_time = e.visible_at.substring(
      e.visible_at.indexOf('T') + 1,
      e.visible_at.indexOf('Z') - 3
    );
    start_date = e.starts_at.substring(0, e.starts_at.indexOf('T'));
    start_time = e.starts_at.substring(e.starts_at.indexOf('T') + 1, e.starts_at.indexOf('Z') - 3);
    event.starts_at = e.starts_at;
    event.ends_at = e.ends_at;
    leaderboards = l ?? divless;
    leaderboards[0]!.event_id = e.id;
  }

  const event = $derived.by(() => {
    const event: Event = {
      id: event_id,
      kind: event_kind as Event['kind'],
      kind_id: event_kind_id,
      player_class: player_class,
      visible_at: visible_at,
      starts_at: starts_at,
      ends_at: starts_at,
      created_at: starts_at
    };
    return event;
  });

  let selectedPrizepool: Prize[] | null = $state(null);
</script>

<Section label="update">
  <EventHeader event={{ event: event, leaderboards: leaderboards }} />

  <Errors {oerror} />
</Section>
<Section label={'div prizepools'}>
  <!-- select leaderboard buttons -->
  <LeaderboardButtons
    {leaderboards}
    bind:selected={selectedLeaderboardID}
    onclick={async () => {
      const { data } = await Client.GET(ApiPaths.get_leaderboard_prizepool, {
        params: { path: { leaderboard_id: selectedLeaderboardID } }
      });
      selectedPrizepool = data ?? null;
    }} />
  {#key selectedLeaderboardID}
    {#await Client.GET(ApiPaths.get_prizepool_total, { params: { path: { event_id: event_id } } })}
      <span></span>
    {:then { data: pp }}
      {#if pp && selectedPrizepool}
        <div class="flex gap-1">
          <span class="text-primary">total</span>
          <span>{pp.total} keys</span>
        </div>
      {/if}
    {/await}
    {#if selectedPrizepool !== null}
      <Button
        onsubmit={async () => {
          selectedPrizepool!.push({
            keys: 0,
            leaderboard_id: selectedLeaderboardID,
            position: selectedPrizepool!.length + 1
          });
          return true;
        }}><span>add placement</span></Button>
    {/if}
    {#each selectedPrizepool as p}
      <div class="flex items-center">
        <span class="grid w-6 justify-self-center">{p.position}</span>
        <Input
          max_width={'max-w-40'}
          type="text"
          placeholder={`${p.keys} keys`}
          onsubmit={async (value) => {
            p.keys = parseInt(value);
            return true;
          }} />
        {#if p.player_id}
          {#await Client.GET( ApiPaths.get_full_player, { params: { path: { player_id: p.player_id } } } )}
            <span></span>
          {:then { data: player }}
            {#if player}
              <TablePlayer {player} flag={false} />
              {#if player.trade_token}
                {@const steamID3: number = Number(BigInt(player.id) - 76561197960265728n)}
                <ExternalLink
                  label="send trade offer"
                  href={`https://steamcommunity.com/tradeoffer/new/?partner=${steamID3}&token=${player.trade_token}`} />
              {/if}
            {/if}
          {/await}
        {/if}
      </div>
    {/each}
  {/key}
  {#if selectedPrizepool !== null}
    <Button
      onsubmit={async () => {
        const resp = await Client.POST(ApiPaths.update_leaderboard_prizepool, {
          params: { path: { leaderboard_id: selectedLeaderboardID } },
          body: selectedPrizepool
        });
        oerror = resp.error;
        // reset
        selectedPrizepool = null;
        if (resp.response.ok) {
          reloadEvents = !reloadEvents;
        }
        return resp.response.ok;
      }}>
      <span>update prizepool</span></Button>
  {/if}
</Section>
<Section label={'updatable event prizepools'}>
  {#key reloadEvents}
    {#await Client.GET(ApiPaths.get_full_events)}
      <span></span>
    {:then { data: ewls }}
      {@const now = Temporal.Now.instant().epochMilliseconds}
      {@const editable = ewls?.filter(({ event }) => datetimeToMs(event.ends_at) > now) ?? []}
      <TableEvents
        data={editable}
        onclick={async (ewl) => {
          loadEvent(ewl);
          const { data } = await Client.GET(ApiPaths.get_leaderboard_prizepool, {
            params: { path: { leaderboard_id: selectedLeaderboardID } }
          });
          selectedPrizepool = data ?? null;
        }}>
      </TableEvents>
    {/await}
  {/key}
</Section>
