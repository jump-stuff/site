<script lang="ts">
  import { browser } from '$app/environment';
  import { goto, invalidateAll } from '$app/navigation';
  import { Client } from '$lib/api/api';
  import { ApiPaths } from '$lib/schema';

  async function handleSignOut() {
    const { error } = await Client.POST(ApiPaths.sign_out, {
      fetch: fetch
    });
    if (!error && browser) {
      invalidateAll();
      goto('/');
    } else {
      console.error('error signing out: ', error);
    }
  }

  handleSignOut();
</script>
