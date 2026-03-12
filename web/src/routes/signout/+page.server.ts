import type { PageServerLoad } from './$types';
import { Client } from '$lib/api/api';
import { ApiPaths } from '$lib/schema';
import { redirect } from '@sveltejs/kit';

export const load: PageServerLoad = async ({ cookies, fetch }) => {
  const { data } = await Client.GET(ApiPaths.sign_out, { fetch: fetch });
  if (data) {
    cookies.delete('sessionid', {
      path: '/',
      secure: false,
      sameSite: 'strict',
      maxAge: data.maxAge,
      expires: new Date(data.expiresAt)
    });
    redirect(302, '/');
  }
};
