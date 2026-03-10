import config from '$lib/api/config';
import { Client } from '$lib/api/api';

import type { Handle } from '@sveltejs/kit';
import { ApiPaths } from '$lib/schema';

export const handle: Handle = async ({ event, resolve }) => {
  // internal api request
  if (event.url.pathname.startsWith('/internal')) {
    let url = new URL(event.url.pathname, config.apiBaseUrl);
    url.search = event.url.search;
    console.log('internal hook new url: ', url);
    const result = await fetch(url, {
      ...event.request,
      redirect: 'manual'
    });
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
