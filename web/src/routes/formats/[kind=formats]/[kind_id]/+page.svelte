<script lang="ts">
  import { Client } from '$lib/api/api';
  import EventHeader from '$lib/components/display/EventHeader.svelte';
  import { ApiPaths, type Leaderboard, type Player, type TimeWithPlayer } from '$lib/schema';
  import { onMount } from 'svelte';
  import type { PageData } from './$types';
  import Table from '$lib/components/display/table/Table.svelte';
  import Section from '$lib/components/layout/Section.svelte';
  import TablePlayer from '$lib/components/display/table/TablePlayer.svelte';
  import { formatPosition, twTableGradients, validDuration } from '$lib/helpers/times';
  import TableTime from '$lib/components/display/table/TableTime.svelte';
  import TemporalDate from '$lib/components/display/TemporalDate.svelte';
  import Content from '$lib/components/layout/Content.svelte';
  import Button from '$lib/components/input/Button.svelte';
  import Errors from '$lib/components/input/Errors.svelte';
  import Input from '$lib/components/input/Input.svelte';
  import Collapse from '$lib/components/input/Collapse.svelte';
  import { Temporal } from 'temporal-polyfill';
  import LeaderboardButtons from '$lib/components/input/LeaderboardButtons.svelte';
  import TableSkeleton from '$lib/components/display/table/presets/TableSkeleton.svelte';
  import MonthlyChart from './MonthlyChart.svelte';
  import MonthlyChartSkeleton from './MonthlyChartSkeleton.svelte';
  import PrizepoolSkeleton from './PrizepoolSkeleton.svelte';

  type Props = {
    data: PageData;
  };
  let { data }: Props = $props();
  const now = Temporal.Now.instant();

  let selectedLeaderboardID: number = $state(0);

  let playerLeaderboard: Leaderboard | undefined = $state(undefined);
  let prPlayer: Player | undefined = $state(undefined);

  let refreshPR: boolean = $state(true);
  let refreshLeaderboard: boolean = $state(true);
  let showPrizepool: boolean = $state(false);

  let ended_days: number = $derived(
    Math.floor(now.since(data?.ewl?.event.ends_at ?? '').seconds / 60 / 60 / 24)
  );

  let mod: boolean = $derived(
    data.session?.role === 'mod' || data.session?.role === 'admin' || data.session?.role === 'dev'
  );

  let oerror: OpenAPIError = $state(undefined);

  // set selectedLeaderboardID & playerLeaderboard
  onMount(async () => {
    if (data.ewl) {
      selectedLeaderboardID = data.ewl.leaderboards?.at(0)?.id ?? 0;

      // set matching player leaderboard ID
      if (data.session) {
        const { data: player } = await Client.GET(ApiPaths.get_player, {
          params: { path: { player_id: data.session.id } }
        });
        if (player) {
          prPlayer = player;
          const div =
            data.ewl.event.player_class === 'Soldier' ? player.soldier_div : player.demo_div;
          for (const l of data.ewl.leaderboards ?? []) {
            if (l.div === div || !l.div) {
              playerLeaderboard = l;
              break;
            }
          }
        }
      }
    }
  });

  let playerIdHighlighted: string | undefined = $state(undefined);
  let chartPositionReferenceElement: HTMLDivElement | undefined = $state(undefined);
  let isChartPoppedOut = $state(false);
  const updateIsChartPoppedOut = () => {
    if (!chartPositionReferenceElement) return;
    const rect = chartPositionReferenceElement.getBoundingClientRect();
    const isVisible = rect.top >= 0 && rect.bottom <= window.innerHeight;
    isChartPoppedOut = !isVisible;
  };
  setTimeout(updateIsChartPoppedOut); // Pages can restore scroll position on load in some cases... I'm not sure it can happen here or if this would actually patch it, but some handling may be needed.
</script>

<svelte:document onscroll={updateIsChartPoppedOut} />

{#if data.ewl}
  {@const prizepoolTotalPromise = Client.GET(ApiPaths.get_prizepool_total, {
    params: { path: { event_id: data.ewl.event.id } }
  })}
  <!-- consider motw timeslots -->
  {#if data.ewl.event.kind == 'motw' && data.session}
    {#await Client.GET( ApiPaths.get_motw, { params: { path: { event_kind: 'motw', kind_id: data.ewl.event.kind_id } } } )}
      <span></span>
    {:then { data: motw }}
      {#if motw}
        {#await Client.GET( ApiPaths.get_timeslot_info, { params: { path: { event_id: data.ewl.event.id } } } )}
          <span></span>
        {:then { data: timeslotInfo }}
          <EventHeader event={motw} timeslots={timeslotInfo} />
        {/await}
      {/if}
    {/await}
  {:else}
    {#await prizepoolTotalPromise}
      <EventHeader event={data.ewl} />
    {:then { data: prizepoolTotal }}
      <EventHeader event={data.ewl} prizepool={prizepoolTotal} />
    {/await}
  {/if}

  <Content>
    {#if data.session}
      {#if (ended_days > 0 && data.session.role === 'admin') || data.session.role === 'dev'}
        <Section label="admin">
          <Button
            onsubmit={async () => {
              let resp = await Client.POST(ApiPaths.update_event_results, {
                params: { path: { event_id: data.ewl?.event.id ?? 0 } }
              });
              oerror = resp.error;
              if (resp.response.ok) {
                refreshPR = !refreshPR;
              }
              return resp.response.ok;
            }}>refresh event results</Button>
        </Section>
      {/if}

      <Section>
        {#key refreshPR}
          {#await Client.GET( ApiPaths.get_event_pr, { params: { path: { event_id: data.ewl.event.id } } } )}
            <span></span>
          {:then { data: pr }}
            {#if prPlayer && pr}
              <Table data={[pr]}>
                {#snippet header()}
                  <th class="w-rank"></th>
                  <th class="w-time"></th>
                  <th class=""></th>
                  <th class="w-date"></th>
                {/snippet}
                {#snippet row({ player, time, position }: TimeWithPlayer)}
                  <td class={twTableGradients.get(`r${position}`)}>{position}</td>
                  <td class={twTableGradients.get(`t${position}`)}><TableTime {time} /></td>
                  <td><TablePlayer {player} /></td>
                  <td class="table-date"><TemporalDate datetime={time.created_at} /></td>
                {/snippet}
              </Table>
            {/if}
          {/await}
        {/key}

        <Errors {oerror} />

        {#if ended_days < 1 && playerLeaderboard}
          <Button
            grow={true}
            onsubmit={async () => {
              let resp = await Client.POST(ApiPaths.submit_time, {
                params: { path: { leaderboard_id: selectedLeaderboardID } }
              });
              oerror = resp.error;
              if (resp.response.ok) {
                refreshPR = !refreshPR;
              }
              return resp.response.ok;
            }}>refresh PR from Tempus<span class="icon-[mdi--refresh]"></span></Button>

          <Collapse label="manually submit time">
            <Input
              type="text"
              label="manual submit"
              placeholder={'MM:SS.ss'}
              max_width="max-w-58"
              onsubmit={async (value) => {
                const valid = validDuration(value);
                if (valid) {
                  const resp = await Client.POST(ApiPaths.submit_unverified_time, {
                    params: {
                      path: { leaderboard_id: playerLeaderboard!.id, run_time: value }
                    }
                  });
                  oerror = resp.error;
                  if (resp.response.ok) {
                    refreshPR = !refreshPR;
                  }
                  return resp.response.ok;
                } else {
                  oerror = {
                    detail: "time isn't in the expected format. format: MM:SS.ss",
                    type: 'error'
                  };
                }
                return false;
              }} />
          </Collapse>
        {/if}
      </Section>
    {/if}
    <Section label={'leaderboards'}>
      <!-- select leaderboard buttons -->
      {#if data.ewl.leaderboards?.length}
        <div class="flex w-full">
          <LeaderboardButtons
            leaderboards={data.ewl.leaderboards}
            bind:selected={selectedLeaderboardID} />
          <!-- refresh prizepool -->
          {#await prizepoolTotalPromise}
            <span></span>
          {:then { data: prizepoolTotal }}
            {#if prizepoolTotal?.total}
              <Button
                table={true}
                onsubmit={async () => {
                  showPrizepool = !showPrizepool;
                  return true;
                }}><span class="icon-[mdi--key]"></span></Button>
            {/if}
          {/await}
          {#if ended_days <= 0}
            <Button
              table={true}
              onsubmit={async () => {
                refreshLeaderboard = !refreshLeaderboard;
                return true;
              }}><span class="icon-[mdi--refresh]"></span></Button>
          {/if}
        </div>
      {/if}

      {#if showPrizepool}
        {#await prizepoolTotalPromise}
          <PrizepoolSkeleton />
        {:then { data: prizepoolTotal }}
          <div class="my-0.5 bg-base-900/50 p-2">
            {#if prizepoolTotal?.total}
              <Content>
                {#await Client.GET( ApiPaths.get_leaderboard_prizepool, { params: { path: { leaderboard_id: selectedLeaderboardID } } } )}
                  <span></span>
                {:then { data: prizepool }}
                  <div class="mx-auto flex w-fit flex-wrap items-center gap-x-2 gap-y-1 text-sm">
                    {#each prizepool?.toReversed() as prize}
                      <div
                        class={[
                          'flex px-2',
                          {
                            1: 'text-div-gold',
                            2: 'text-div-silver',
                            3: 'text-div-bronze'
                          }[prize.position] ?? 'text-primary'
                        ]}>
                        <span class="rounded-box bg-base-900 px-3 text-left whitespace-nowrap"
                          >{formatPosition(prize.position)}</span>
                        <span class="ml-1.5 whitespace-nowrap"
                          >{prize.keys} key{prize.keys === 1 ? '' : 's'}</span>
                      </div>
                    {/each}
                  </div>
                {/await}
              </Content>
            {:else}
              <div class="mx-auto w-fit opacity-65">no prizes offered</div>
            {/if}
          </div>
        {/await}
      {/if}

      {#key refreshLeaderboard && refreshPR}
        <!-- call motw leaderboards with session if applicable -->
        {@const motwWithSession: boolean = (data.session !== undefined && data.ewl.event.kind === "motw")}
        {@const leaderboardPath = motwWithSession
          ? ApiPaths.get_motw_leaderboard_times
          : ApiPaths.get_leaderboard_times}
        {#await Client.GET( leaderboardPath, { params: { path: { leaderboard_id: selectedLeaderboardID } } } )}
          <MonthlyChartSkeleton />
          <TableSkeleton rows={3} />
        {:then { data: times }}
          <div bind:this={chartPositionReferenceElement}></div>
          <div
            class={[
              'z-20 bg-base-800',
              isChartPoppedOut && 'sticky top-0 border-b-2 border-b-base-800'
            ]}>
            <MonthlyChart times={times ?? []} {playerIdHighlighted} />
          </div>
          <Table
            data={(times ?? []).map((time) => ({
              ...time,
              onmouseover: () => (playerIdHighlighted = time.player.id),
              onmouseout: () => (playerIdHighlighted = undefined)
            }))}>
            {#snippet header()}
              <th class="w-rank"></th>
              <th class="w-time"></th>
              <th class=""></th>
              <th class="w-date"></th>
              {#if mod}
                <th class="w-0"></th>
                <th class="w-0"></th>
              {/if}
            {/snippet}
            {#snippet row({ player, time, position }: TimeWithPlayer)}
              <td class={twTableGradients.get(`r${position}`)}>{position}</td>
              <td class={twTableGradients.get(`t${position}`)}><TableTime {time} /></td>
              <td><TablePlayer {player} link={true} /></td>
              <td class="table-date"><TemporalDate datetime={time.created_at} /></td>
              {#if mod && !time.verified}
                <td
                  ><Button
                    table={true}
                    onsubmit={async () => {
                      const resp = await Client.POST(ApiPaths.verify_player_time, {
                        params: { path: { time_id: time.id } }
                      });
                      oerror = resp.error;
                      if (resp.response.ok) {
                        refreshLeaderboard = !refreshLeaderboard;
                      }
                      return resp.response.ok;
                    }}><span class="icon-[mdi--check]"></span></Button
                  ></td>
              {/if}
              {#if mod && !time.tempus_time_id && ended_days < 7}
                <td
                  ><Button
                    table={true}
                    onsubmit={async () => {
                      const resp = await Client.DELETE(ApiPaths.delete_player_time, {
                        params: { path: { time_id: time.id } }
                      });
                      oerror = resp.error;
                      if (resp.response.ok) {
                        refreshLeaderboard = !refreshLeaderboard;
                      }
                      return resp.response.ok;
                    }}><span class="icon-[mdi--close]"></span></Button
                  ></td>
              {/if}
            {/snippet}
          </Table>
        {/await}
      {/key}
    </Section>
  </Content>
{/if}
