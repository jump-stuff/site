import config from '$lib/api/config';
import { Client } from '$lib/api/api';

import type { Handle } from '@sveltejs/kit';
import { ApiPaths } from '$lib/schema';

export const handle: Handle = async ({ event, resolve }) => {
  // internal api request
  if (event.url.pathname.startsWith('/internal')) {
    const result = fetch(new URL(event.url.pathname, config.apiBaseUrl), event.request);
    return result;
  }

  // check for session before making a request
  // set session to Promise<Session> if not
  if (!event.locals.session) {
    const { data } = await Client.GET(ApiPaths.get_session, {
      baseUrl: config.apiBaseUrl,
      headers: event.request.headers,
      credentials: 'include'
    });
    event.locals.session = data;
  }

  return await resolve(event, {
    filterSerializedResponseHeaders(name) {
      return name === 'content-length';
    }
  });
};
