<script>
  import { onMount } from 'svelte';
  import { apiKey } from '../lib/stores.js';
  import AdminAPI from '../lib/api.js';
  
  let stats = null;
  let modelInput = '';
  
  async function loadStats() {
    const api = new AdminAPI($apiKey);
    stats = await api.getCacheStats();
  }
  
  async function clearCache() {
    if (confirm('Clear all cache?')) {
      const api = new AdminAPI($apiKey);
      await api.clearCache();
      alert('Cache cleared!');
      await loadStats();
    }
  }
  
  async function invalidateModel() {
    if (modelInput.trim()) {
      const api = new AdminAPI($apiKey);
      const result = await api.invalidateCacheByModel(modelInput);
      alert(`${result.entries_removed} entries removed`);
      modelInput = '';
      await loadStats();
    }
  }
  
  onMount(loadStats);
</script>

<div class="p-8">
  <h1 class="text-3xl font-bold mb-8">Cache Management</h1>
  
  <div class="bg-white p-6 rounded-lg shadow mb-8">
    <h2 class="text-xl font-bold mb-4">Cache Statistics</h2>
    <div class="grid grid-cols-4 gap-4">
      <div>
        <p class="text-gray-500">Hits</p>
        <p class="text-2xl font-bold">{stats?.hits || 0}</p>
      </div>
      <div>
        <p class="text-gray-500">Misses</p>
        <p class="text-2xl font-bold">{stats?.misses || 0}</p>
      </div>
      <div>
        <p class="text-gray-500">Errors</p>
        <p class="text-2xl font-bold">{stats?.errors || 0}</p>
      </div>
      <div>
        <p class="text-gray-500">Hit Rate</p>
        <p class="text-2xl font-bold">{(stats?.hit_rate || 0).toFixed(1)}%</p>
      </div>
    </div>
  </div>
  
  <div class="grid grid-cols-2 gap-6">
    <div class="bg-white p-6 rounded-lg shadow">
      <h2 class="text-xl font-bold mb-4">Clear All Cache</h2>
      <p class="text-gray-600 mb-4">This will remove all cached responses.</p>
      <button on:click={clearCache} class="bg-red-600 text-white px-4 py-2 rounded hover:bg-red-700">
        Clear All Cache
      </button>
    </div>
    
    <div class="bg-white p-6 rounded-lg shadow">
      <h2 class="text-xl font-bold mb-4">Invalidate by Model</h2>
      <input type="text" bind:value={modelInput} placeholder="claude-3-haiku-20240307" class="w-full p-2 border rounded mb-4" />
      <button on:click={invalidateModel} class="bg-orange-600 text-white px-4 py-2 rounded hover:bg-orange-700">
        Invalidate Model Cache
      </button>
    </div>
  </div>
</div>
