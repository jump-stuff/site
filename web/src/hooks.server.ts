import config from '$lib/api/config';
import { Client } from '$lib/api/api';

import type { Handle } from '@sveltejs/kit';
import { ApiPaths } from '$lib/schema';

export const handle: Handle = async ({ event, resolve }) => {
  // internal api request
  if (event.url.pathname.startsWith('/internal')) {
    let url = new URL(event.url.pathname, config.apiBaseUrl);
    url.search = event.url.search;
    // todo: better way of handling auth paths
    const resp = new Request(url, {
      ...event.request,
      credentials: 'include',
      headers: event.request.headers,
      method: event.request.method,
      body: event.request.body,
      // @ts-ignore
      duplex: 'half'
    });
    const result = await fetch(resp);
    return result;
  }

  // check for session before making a request
  // set session to Promise<Session> if not
  if (!event.locals.session) {
    try {
      const { data } = await Client.GET(ApiPaths.get_session, {
        fetch: fetch,
        baseUrl: config.apiBaseUrl,
        headers: event.request.headers,
        credentials: 'include'
      });
      event.locals.session = data;
    } catch (error) {
      console.log('erorr', error);
    }
  }

  return await resolve(event, {
    filterSerializedResponseHeaders(name) {
      return name === 'content-length';
    }
  });
};
