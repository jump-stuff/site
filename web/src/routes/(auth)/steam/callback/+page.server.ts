import type { PageServerLoad } from './$types';
import { Client } from '$lib/api/api';
import { ApiPaths } from '$lib/schema';
import { PUBLIC_JUMP_SESSION_COOKIE_SECURE } from '$env/static/public';

export const load: PageServerLoad = async ({ cookies, url }) => {
  const queryParams = url.searchParams;
  const { data } = await Client.GET(ApiPaths.steam_callback, {
    params: { query: Object.fromEntries(queryParams) }
  });
  if (data) {
    cookies.set('sessionid', data.sessionToken, {
      path: '/',
      secure: PUBLIC_JUMP_SESSION_COOKIE_SECURE === 'true',
      sameSite: 'strict',
      maxAge: data.maxAge,
      expires: new Date(data.expiresAt)
    });
  }
};
