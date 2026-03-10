<script lang="ts">
  import { page } from '$app/state';
  import { Client } from '$lib/api/api';
  import { ApiPaths } from '$lib/schema';

  $inspect(page.url.searchParams);

  async function verifyCallback() {
    console.log('TO STRING');
    console.log(page.url.toString());
    console.log(page.url.searchParams);
    const { data } = await Client.GET(ApiPaths.steam_callback, {
      params: { query: { openidurl: encodeURIComponent(page.url.toString()) } }
    });
    console.log(data);
  }

  verifyCallback();
</script>
