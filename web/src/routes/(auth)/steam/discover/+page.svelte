<script lang="ts">
  import { browser } from '$app/environment';
  import { Client } from '$lib/api/api';
  import { ApiPaths } from '$lib/schema';

  async function handleSignin() {
    const { data } = await Client.GET(ApiPaths.steam_discover, {
      fetch: fetch
    });
    if (data?.url && browser) {
      window.location.assign(data.url);
    }
  }

  handleSignin();
</script>
