<script>
  import { onMount } from 'svelte';
  import { apiKey } from '../lib/stores.js';
  import AdminAPI from '../lib/api.js';
  
  let stats = null;
  let clientFilter = '';
  let modelFilter = '';
  
  async function loadStats() {
    const api = new AdminAPI($apiKey);
    const params = {};
    if (clientFilter) params.client_id = clientFilter;
    if (modelFilter) params.model = modelFilter;
    stats = await api.getUsageStats(params);
  }
  
  onMount(loadStats);
</script>

<div class="p-8">
  <h1 class="text-3xl font-bold mb-8">Usage Statistics</h1>
  
  <div class="bg-white p-6 rounded-lg shadow mb-8">
    <h2 class="text-xl font-bold mb-4">Filters</h2>
    <div class="grid grid-cols-3 gap-4">
      <input type="text" bind:value={clientFilter} placeholder="Client ID" class="p-2 border rounded" />
      <input type="text" bind:value={modelFilter} placeholder="Model" class="p-2 border rounded" />
      <button on:click={loadStats} class="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700">
        Apply Filters
      </button>
    </div>
  </div>
  
  {#if stats}
    <div class="grid grid-cols-3 gap-6 mb-8">
      <div class="bg-white p-6 rounded-lg shadow">
        <h3 class="text-gray-500 text-sm mb-2">Total Requests</h3>
        <p class="text-3xl font-bold">{stats.TotalRequests || 0}</p>
      </div>
      <div class="bg-white p-6 rounded-lg shadow">
        <h3 class="text-gray-500 text-sm mb-2">Total Tokens</h3>
        <p class="text-3xl font-bold">{stats.TotalTokens || 0}</p>
      </div>
      <div class="bg-white p-6 rounded-lg shadow">
        <h3 class="text-gray-500 text-sm mb-2">Total Cost</h3>
        <p class="text-3xl font-bold">${(stats.TotalCost || 0).toFixed(4)}</p>
      </div>
    </div>
    
    <div class="grid grid-cols-3 gap-6">
      <div class="bg-white p-6 rounded-lg shadow">
        <h3 class="text-gray-500 text-sm mb-2">Avg Duration</h3>
        <p class="text-2xl font-bold">{(stats.AvgDuration || 0).toFixed(0)}ms</p>
      </div>
      <div class="bg-white p-6 rounded-lg shadow">
        <h3 class="text-gray-500 text-sm mb-2">Cache Hit Rate</h3>
        <p class="text-2xl font-bold">{(stats.CacheHitRate || 0).toFixed(1)}%</p>
      </div>
      <div class="bg-white p-6 rounded-lg shadow">
        <h3 class="text-gray-500 text-sm mb-2">Error Requests</h3>
        <p class="text-2xl font-bold">{stats.ErrorRequests || 0}</p>
      </div>
    </div>
  {/if}
</div>
