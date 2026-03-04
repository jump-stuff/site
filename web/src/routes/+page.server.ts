import type { PageServerLoad } from './$types';
import { Client } from '$lib/api/api';
import { ApiPaths } from '$lib/schema';

export const load: PageServerLoad = async ({ fetch }) => {
  const eventsResponse = await Client.GET(ApiPaths.get_recent_events, {
    fetch: fetch
  });
  const statsResponse = await Client.GET(ApiPaths.get_stats, {
    fetch: fetch
  });
  return { events: eventsResponse.data, stats: statsResponse.data };
};
