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
  
  // Auto-reload stats when filters change
  $: if (clientFilter !== undefined && modelFilter !== undefined) {
    loadStats();
  }
</script>

<div class="p-8">
  <h1 class="text-3xl font-bold mb-8">Usage Statistics</h1>
  
  <div class="bg-white p-6 rounded-lg shadow mb-8">
    <h2 class="text-xl font-bold mb-4">Filters</h2>
    <p class="text-sm text-gray-600 mb-3">Statistics will update automatically as you type</p>
    <div class="grid grid-cols-2 gap-4">
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Client ID</label>
        <input 
          type="text" 
          bind:value={clientFilter} 
          placeholder="Filter by client ID..." 
          class="w-full p-2 border rounded focus:outline-none focus:ring-2 focus:ring-blue-500" 
        />
      </div>
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Model</label>
        <input 
          type="text" 
          bind:value={modelFilter} 
          placeholder="Filter by model..." 
          class="w-full p-2 border rounded focus:outline-none focus:ring-2 focus:ring-blue-500" 
        />
      </div>
    </div>
    {#if clientFilter || modelFilter}
      <div class="mt-3 text-sm text-gray-600">
        <span class="font-medium">Active filters:</span>
        {#if clientFilter}
          <span class="ml-2 px-2 py-1 bg-blue-100 text-blue-800 rounded">Client: {clientFilter}</span>
        {/if}
        {#if modelFilter}
          <span class="ml-2 px-2 py-1 bg-green-100 text-green-800 rounded">Model: {modelFilter}</span>
        {/if}
        <button 
          on:click={() => { clientFilter = ''; modelFilter = ''; }} 
          class="ml-3 text-blue-600 hover:text-blue-800 underline"
        >
          Clear all
        </button>
      </div>
    {/if}
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
