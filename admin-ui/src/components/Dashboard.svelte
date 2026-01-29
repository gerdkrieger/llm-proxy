<script>
  import { onMount } from 'svelte';
  import { apiKey } from '../lib/stores.js';
  import AdminAPI from '../lib/api.js';
  
  let stats = null;
  let cache = null;
  let providers = null;
  let loading = true;
  
  async function loadData() {
    loading = true;
    const api = new AdminAPI($apiKey);
    try {
      [stats, cache, providers] = await Promise.all([
        api.getUsageStats(),
        api.getCacheStats(),
        api.getProviderStatus()
      ]);
    } catch (e) {
      console.error(e);
    }
    loading = false;
  }
  
  onMount(loadData);
</script>

<div class="p-8">
  <h1 class="text-3xl font-bold mb-8">Dashboard</h1>
  
  {#if loading}
    <p>Loading...</p>
  {:else}
    <div class="grid grid-cols-3 gap-6 mb-8">
      <div class="bg-white p-6 rounded-lg shadow">
        <h3 class="text-gray-500 text-sm mb-2">Total Requests</h3>
        <p class="text-3xl font-bold">{stats?.TotalRequests || 0}</p>
      </div>
      <div class="bg-white p-6 rounded-lg shadow">
        <h3 class="text-gray-500 text-sm mb-2">Total Tokens</h3>
        <p class="text-3xl font-bold">{stats?.TotalTokens || 0}</p>
      </div>
      <div class="bg-white p-6 rounded-lg shadow">
        <h3 class="text-gray-500 text-sm mb-2">Total Cost</h3>
        <p class="text-3xl font-bold">${(stats?.TotalCost || 0).toFixed(4)}</p>
      </div>
    </div>
    
    <div class="grid grid-cols-2 gap-6">
      <div class="bg-white p-6 rounded-lg shadow">
        <h3 class="text-xl font-bold mb-4">Cache Performance</h3>
        <p>Hits: {cache?.hits || 0}</p>
        <p>Misses: {cache?.misses || 0}</p>
        <p>Hit Rate: {(cache?.hit_rate || 0).toFixed(1)}%</p>
      </div>
      
      <div class="bg-white p-6 rounded-lg shadow">
        <h3 class="text-xl font-bold mb-4">Provider Status</h3>
        <p>Status: <span class:text-green-600={providers?.healthy} class:text-red-600={!providers?.healthy}>
          {providers?.healthy ? 'Healthy' : 'Unhealthy'}
        </span></p>
        <p>Providers: {providers?.provider_count || 0}</p>
        <p>Models: {providers?.models?.length || 0}</p>
      </div>
    </div>
  {/if}
</div>
