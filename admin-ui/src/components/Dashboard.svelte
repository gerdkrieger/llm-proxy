<script>
  import { onMount, onDestroy } from 'svelte';
  import { apiKey } from '../lib/stores.js';
  import AdminAPI from '../lib/api.js';
  
  let dashboardData = null;
  let loading = true;
  let error = null;
  let autoRefresh = true;
  let refreshInterval = null;
  let lastUpdate = null;
  
  async function loadData() {
    const api = new AdminAPI($apiKey);
    try {
      error = null;
      dashboardData = await api.getDashboardData();
      lastUpdate = new Date();
    } catch (e) {
      console.error('Dashboard error:', e);
      error = e.message;
    } finally {
      loading = false;
    }
  }
  
  function toggleAutoRefresh() {
    autoRefresh = !autoRefresh;
    if (autoRefresh) {
      startAutoRefresh();
    } else {
      stopAutoRefresh();
    }
  }
  
  function startAutoRefresh() {
    if (refreshInterval) clearInterval(refreshInterval);
    refreshInterval = setInterval(() => {
      loadData();
    }, 30000); // Refresh every 30 seconds
  }
  
  function stopAutoRefresh() {
    if (refreshInterval) {
      clearInterval(refreshInterval);
      refreshInterval = null;
    }
  }
  
  function formatNumber(num) {
    if (num >= 1000000) return (num / 1000000).toFixed(2) + 'M';
    if (num >= 1000) return (num / 1000).toFixed(1) + 'K';
    return num?.toLocaleString() || '0';
  }
  
  function formatCost(cost) {
    return '$' + (cost || 0).toFixed(4);
  }
  
  function formatDateTime(date) {
    return new Date(date).toLocaleString();
  }
  
  function formatDuration(ms) {
    if (ms < 1000) return ms + 'ms';
    return (ms / 1000).toFixed(2) + 's';
  }
  
  function getStatusColor(status) {
    if (status >= 200 && status < 300) return 'text-green-600';
    if (status >= 400) return 'text-red-600';
    return 'text-yellow-600';
  }
  
  onMount(() => {
    loadData();
    if (autoRefresh) startAutoRefresh();
  });
  
  onDestroy(() => {
    stopAutoRefresh();
  });
</script>

<div class="p-8">
  <!-- Header with Refresh Controls -->
  <div class="flex justify-between items-center mb-8">
    <div>
      <h1 class="text-3xl font-bold">Dashboard</h1>
      {#if lastUpdate}
        <p class="text-sm text-gray-500 mt-1">Last updated: {formatDateTime(lastUpdate)}</p>
      {/if}
    </div>
    <div class="flex gap-3">
      <button
        on:click={() => { loading = true; loadData(); }}
        class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 flex items-center gap-2"
        disabled={loading}
      >
        <svg class="w-4 h-4" class:animate-spin={loading} fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path>
        </svg>
        Refresh
      </button>
      <button
        on:click={toggleAutoRefresh}
        class="px-4 py-2 rounded flex items-center gap-2"
        class:bg-green-600={autoRefresh}
        class:text-white={autoRefresh}
        class:bg-gray-200={!autoRefresh}
        class:text-gray-700={!autoRefresh}
      >
        {autoRefresh ? '✓ Auto-Refresh (30s)' : 'Auto-Refresh OFF'}
      </button>
    </div>
  </div>
  
  {#if error}
    <div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-6">
      <strong>Error:</strong> {error}
    </div>
  {/if}
  
  {#if loading && !dashboardData}
    <div class="text-center py-12">
      <div class="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      <p class="mt-4 text-gray-600">Loading dashboard...</p>
    </div>
  {:else if dashboardData}
    <!-- Key Metrics Row -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
      <div class="bg-white p-6 rounded-lg shadow-md border-l-4 border-blue-500">
        <h3 class="text-gray-500 text-sm font-medium mb-2">Total Requests</h3>
        <p class="text-3xl font-bold text-gray-900">{formatNumber(dashboardData.usage?.total_requests || 0)}</p>
        <p class="text-xs text-gray-500 mt-2">
          {formatNumber(dashboardData.usage?.cached_requests || 0)} cached ({(dashboardData.usage?.cache_hit_rate || 0).toFixed(1)}%)
        </p>
      </div>
      
      <div class="bg-white p-6 rounded-lg shadow-md border-l-4 border-green-500">
        <h3 class="text-gray-500 text-sm font-medium mb-2">Total Tokens</h3>
        <p class="text-3xl font-bold text-gray-900">{formatNumber(dashboardData.usage?.total_tokens || 0)}</p>
        <p class="text-xs text-gray-500 mt-2">Avg {formatDuration(dashboardData.usage?.avg_duration_ms || 0)}/request</p>
      </div>
      
      <div class="bg-white p-6 rounded-lg shadow-md border-l-4 border-purple-500">
        <h3 class="text-gray-500 text-sm font-medium mb-2">Total Cost</h3>
        <p class="text-3xl font-bold text-gray-900">{formatCost(dashboardData.usage?.total_cost || 0)}</p>
        <p class="text-xs text-gray-500 mt-2">{formatNumber(dashboardData.client_count || 0)} clients</p>
      </div>
      
      <div class="bg-white p-6 rounded-lg shadow-md border-l-4 border-red-500">
        <h3 class="text-gray-500 text-sm font-medium mb-2">Error Rate</h3>
        <p class="text-3xl font-bold text-gray-900">{(dashboardData.error_rate || 0).toFixed(2)}%</p>
        <p class="text-xs text-gray-500 mt-2">{formatNumber(dashboardData.usage?.error_requests || 0)} errors</p>
      </div>
    </div>
    
    <!-- Provider & Cache Row -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-6 mb-8">
      <!-- Providers -->
      <div class="bg-white p-6 rounded-lg shadow-md">
        <h3 class="text-xl font-bold mb-4 flex items-center justify-between">
          <span>Providers</span>
          <span class="px-3 py-1 rounded-full text-sm font-semibold"
            class:bg-green-100={dashboardData.providers?.healthy}
            class:text-green-800={dashboardData.providers?.healthy}
            class:bg-red-100={!dashboardData.providers?.healthy}
            class:text-red-800={!dashboardData.providers?.healthy}
          >
            {dashboardData.providers?.healthy ? 'Healthy' : 'Unhealthy'}
          </span>
        </h3>
        
        <div class="space-y-3">
          {#each (dashboardData.providers?.providers || []) as provider}
            <div class="flex justify-between items-center p-3 bg-gray-50 rounded">
              <div>
                <div class="font-semibold text-gray-900">{provider.id}</div>
                {#if provider.name && provider.name !== provider.id}
                  <div class="text-sm text-gray-500">{provider.name}</div>
                {/if}
              </div>
              <div class="text-right">
                <div class="text-sm font-medium text-gray-700">
                  <span class="text-green-600 font-bold">{provider.enabled_models}</span> / {provider.total_models} models
                </div>
                {#if provider.health_status}
                  <div class="text-xs" class:text-green-600={provider.health_status === 'healthy'} class:text-red-600={provider.health_status === 'unhealthy'}>
                    {provider.health_status}
                  </div>
                {/if}
              </div>
            </div>
          {/each}
        </div>
      </div>
      
      <!-- Cache Performance -->
      <div class="bg-white p-6 rounded-lg shadow-md">
        <h3 class="text-xl font-bold mb-4">Cache Performance</h3>
        <div class="space-y-4">
          <div class="flex justify-between items-center">
            <span class="text-gray-600">Hits</span>
            <span class="text-2xl font-bold text-green-600">{formatNumber(dashboardData.cache?.hits || 0)}</span>
          </div>
          <div class="flex justify-between items-center">
            <span class="text-gray-600">Misses</span>
            <span class="text-2xl font-bold text-gray-700">{formatNumber(dashboardData.cache?.misses || 0)}</span>
          </div>
          <div class="pt-3 border-t border-gray-200">
            <div class="flex justify-between items-center">
              <span class="text-gray-600">Hit Rate</span>
              <span class="text-3xl font-bold text-blue-600">{(dashboardData.cache?.hit_rate || 0).toFixed(1)}%</span>
            </div>
            <div class="w-full bg-gray-200 rounded-full h-2 mt-2">
              <div class="bg-blue-600 h-2 rounded-full" style="width: {dashboardData.cache?.hit_rate || 0}%"></div>
            </div>
          </div>
        </div>
      </div>
    </div>
    
    <!-- Recent Activity & Filter Stats -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
      <!-- Recent Activity (2/3 width) -->
      <div class="md:col-span-2 bg-white p-6 rounded-lg shadow-md">
        <h3 class="text-xl font-bold mb-4">Recent Activity</h3>
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200">
            <thead>
              <tr class="text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                <th class="pb-3">Time</th>
                <th class="pb-3">Model</th>
                <th class="pb-3">Status</th>
                <th class="pb-3">Tokens</th>
                <th class="pb-3">Cost</th>
                <th class="pb-3">Duration</th>
                <th class="pb-3">Flags</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-200">
              {#each (dashboardData.recent_activity || []) as activity}
                <tr class="text-sm">
                  <td class="py-2 text-gray-600 text-xs">{new Date(activity.timestamp).toLocaleTimeString()}</td>
                  <td class="py-2 font-medium text-gray-900">{activity.model || 'N/A'}</td>
                  <td class="py-2">
                    <span class="font-semibold {getStatusColor(activity.status)}">{activity.status}</span>
                  </td>
                  <td class="py-2 text-gray-600">{formatNumber(activity.tokens)}</td>
                  <td class="py-2 text-gray-600">{formatCost(activity.cost)}</td>
                  <td class="py-2 text-gray-600">{formatDuration(activity.duration)}</td>
                  <td class="py-2 text-xs">
                    {#if activity.cached}
                      <span class="px-2 py-1 bg-blue-100 text-blue-800 rounded mr-1">📦</span>
                    {/if}
                    {#if activity.filtered}
                      <span class="px-2 py-1 bg-yellow-100 text-yellow-800 rounded mr-1">🛡️</span>
                    {/if}
                    {#if activity.error}
                      <span class="px-2 py-1 bg-red-100 text-red-800 rounded">❌</span>
                    {/if}
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      </div>
      
      <!-- Content Filtering (1/3 width) -->
      <div class="bg-white p-6 rounded-lg shadow-md">
        <h3 class="text-xl font-bold mb-4">Content Filtering</h3>
        <div class="space-y-4">
          <div class="text-center p-6 bg-red-50 rounded-lg">
            <div class="text-4xl font-bold text-red-600">{formatNumber(dashboardData.filters?.total_matches || 0)}</div>
            <div class="text-sm text-gray-600 mt-2">Total Blocked</div>
          </div>
          {#if (dashboardData.filters?.total_matches || 0) > 0}
            <a
              href="#/filters"
              class="block w-full text-center px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700 transition-colors"
            >
              View Details
            </a>
          {/if}
        </div>
      </div>
    </div>
  {/if}
</div>

<style>
  .animate-spin {
    animation: spin 1s linear infinite;
  }
  
  @keyframes spin {
    from {
      transform: rotate(0deg);
    }
    to {
      transform: rotate(360deg);
    }
  }
</style>
