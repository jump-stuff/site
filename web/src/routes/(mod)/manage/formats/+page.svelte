<script lang="ts">
  import { Client } from '$lib/api/api';
  import ClassSelect from '$lib/components/display/ClassSelect.svelte';
  import EventHeader from '$lib/components/display/EventHeader.svelte';
  import TableEvents from '$lib/components/display/table/presets/TableEvents.svelte';
  import TemporalDate from '$lib/components/display/TemporalDate.svelte';
  import Button from '$lib/components/input/Button.svelte';
  import Errors from '$lib/components/input/Errors.svelte';
  import Input from '$lib/components/input/Input.svelte';
  import Select from '$lib/components/input/Select.svelte';
  import Section from '$lib/components/layout/Section.svelte';
  import { datetimeToMs, validDateTime } from '$lib/helpers/temporal';
  import { ApiPaths, type Event, type EventWithLeaderboards, type Leaderboard } from '$lib/schema';
  import { Temporal } from 'temporal-polyfill';

  let event_id: number = $state(0);
  let player_class: 'Soldier' | 'Demo' = $state('Soldier');
  let event_kind: string = $state('monthly');
  let event_kind_id: number = $state(0);
  let visible_date: string = $state('');
  let visible_time: string = $state('');
  let start_date: string = $state('');
  let start_time: string = $state('');

  // non-monthly
  let end_date: string = $state('');
  let end_time: string = $state('');

  const visible_at: string = $derived(validDateTime(`${visible_date}T${visible_time}:00Z`));
  const starts_at: string = $derived(validDateTime(`${start_date}T${start_time}:00Z`));
  const ends_at: string = $derived(
    event_kind === 'monthly' || event_kind === 'motw'
      ? validDateTime(`${start_date}T${start_time}:00Z`)
      : validDateTime(`${end_date}T${end_time}:00Z`)
  );
  let leaderboards: Leaderboard[] = $state([]);

  // key to fetch updated events
  let reloadEvents: boolean = $state(true);
  let mode: 'create' | 'update' = $state('create');
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

    end_date = e.ends_at.substring(0, e.ends_at.indexOf('T'));
    end_time = e.ends_at.substring(e.ends_at.indexOf('T') + 1, e.ends_at.indexOf('Z') - 3);

    event.ends_at = e.ends_at;
    leaderboards = l ?? [];
  }

  const event = $derived.by(() => {
    const event: Event = {
      id: event_id,
      kind: event_kind as Event['kind'],
      kind_id: event_kind_id,
      player_class: player_class,
      visible_at: visible_at,
      starts_at: starts_at,
      ends_at: ends_at,
      created_at: starts_at
    };
    return event;
  });
</script>

<Section label={mode}>
  <EventHeader event={{ event: event, leaderboards: leaderboards }} />
  <div class="flex justify-end gap-1">
    <span class="text-primary">visible at</span>
    <TemporalDate datetime={visible_at} />
    <span class="icon-[mdi--monitor]"></span>
  </div>

  <Errors {oerror} />

  <div class="flex">
    <ClassSelect bind:player_class />
  </div>

  <div class="flex flex-col gap-1">
    <Select
      label="kind"
      type="button"
      options={['monthly', 'archive', 'motw', 'test']}
      bind:value={event_kind}
      withSubmit={false}
      onsubmit={async () => {
        return true;
      }} />

    {#if event_kind === 'motw'}
      <span class="text-content/75"
        >motw start time doesn't matter, as it uses the earliest timeslot's start time.</span>
    {/if}
    <div class="flex justify-between">
      <div class="flex grow flex-col gap-1">
        <div class="flex gap-1">
          <Input
            label="visible date"
            type="date"
            withSubmit={false}
            bind:value={visible_date}
            onsubmit={async () => {
              return true;
            }} />
          <Input
            label="visible time"
            type="time"
            withSubmit={false}
            bind:value={visible_time}
            onsubmit={async () => {
              return true;
            }} />
        </div>

        <div class="flex gap-1">
          <Input
            label="start date"
            type="date"
            withSubmit={false}
            bind:value={start_date}
            onsubmit={async () => {
              return true;
            }} />
          <Input
            label="start time"
            type="time"
            withSubmit={false}
            bind:value={start_time}
            onsubmit={async () => {
              return true;
            }} />
        </div>

        <!-- input end datetime -->
        {#if event_kind !== 'monthly' && event_kind !== 'motw'}
          <div class="flex gap-1">
            <Input
              label="end date"
              type="date"
              withSubmit={false}
              bind:value={end_date}
              onsubmit={async () => {
                return true;
              }} />
            <Input
              label="end time"
              type="time"
              withSubmit={false}
              bind:value={end_time}
              onsubmit={async () => {
                return true;
              }} />
          </div>
        {/if}
      </div>

      <div class="flex flex-col">
        <span
          >input timezone <span class="text-primary">UTC</span>
          <span></span>
        </span>
        <span
          >your timezone <span class="text-primary"
            >{Temporal.Now.timeZoneId()} ({Temporal.Now.zonedDateTimeISO().offset} UTC)</span
          ></span>
      </div>
    </div>

    <!-- buttons -->
    <div class="flex gap-1">
      {#if mode === 'create'}
        <Button
          onsubmit={async () => {
            const resp = await Client.POST(ApiPaths.create_event, {
              body: event
            });
            oerror = resp.error;
            if (resp.response.ok) {
              reloadEvents = !reloadEvents;
            }
            return resp.response.ok;
          }}><span>create</span></Button>
      {:else}
        <Button
          onsubmit={async () => {
            const resp = await Client.POST(ApiPaths.update_event, {
              body: event
            });
            oerror = resp.error;
            if (resp.response.ok) {
              reloadEvents = !reloadEvents;
            }
            return resp.response.ok;
          }}><span>update</span></Button>
        <Button
          onsubmit={async () => {
            const resp = await Client.DELETE(ApiPaths.cancel_event, {
              params: { path: { event_id: event.id } }
            });
            oerror = resp.error;
            if (resp.response.ok) {
              reloadEvents = !reloadEvents;
            }
            return resp.response.ok;
          }}><span>delete</span></Button>
      {/if}
    </div>
  </div>
</Section>

<Section label={'updatable events'}>
  {#key reloadEvents}
    {#await Client.GET(ApiPaths.get_full_events)}
      <span></span>
    {:then { data: ewls }}
      {@const now = Temporal.Now.instant().epochMilliseconds}
      {@const editable = ewls?.filter(({ event }) => datetimeToMs(event.starts_at) > now) ?? []}
      <TableEvents
        data={editable}
        onclick={(ewl) => {
          mode = 'update';
          loadEvent(ewl);
        }}>
      </TableEvents>
    {/await}
  {/key}
</Section>
