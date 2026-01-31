<script>
  import { onMount } from 'svelte';
  import { apiKey } from '../lib/stores.js';
  import AdminAPI from '../lib/api.js';
  
  let stats = null;
  let cache = null;
  let providers = null;
  let filterStats = null;
  let loading = true;
  let showBlockedContentModal = false;
  let blockedContentList = [];
  let loadingBlockedContent = false;
  
  async function loadData() {
    loading = true;
    const api = new AdminAPI($apiKey);
    try {
      [stats, cache, providers, filterStats] = await Promise.all([
        api.getUsageStats(),
        api.getCacheStats(),
        api.getProviderStatus(),
        api.getFilterStats()
      ]);
    } catch (e) {
      console.error(e);
    }
    loading = false;
  }
  
  async function loadBlockedContent() {
    loadingBlockedContent = true;
    const api = new AdminAPI($apiKey);
    try {
      const result = await api.getFilterMatches(100);
      blockedContentList = result.matches || [];
    } catch (e) {
      console.error(e);
      blockedContentList = [];
    }
    loadingBlockedContent = false;
  }
  
  function openBlockedContentModal() {
    showBlockedContentModal = true;
    loadBlockedContent();
  }
  
  function formatDateTime(dateStr) {
    return new Date(dateStr).toLocaleString();
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
    
    <div class="grid grid-cols-3 gap-6">
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
      
      <div class="bg-white p-6 rounded-lg shadow">
        <h3 class="text-xl font-bold mb-4">Content Filtering</h3>
        <p>Total Filters: {filterStats?.total_filters || 0}</p>
        <p>Enabled: {filterStats?.enabled_filters || 0}</p>
        <p class="text-red-600 font-semibold">Blocked Content: {filterStats?.total_matches || 0}</p>
        {#if filterStats?.total_matches > 0}
          <button 
            on:click={openBlockedContentModal}
            class="mt-3 px-4 py-2 bg-red-600 text-white text-sm rounded hover:bg-red-700"
          >
            View Details
          </button>
        {/if}
      </div>
    </div>
  {/if}
</div>

<!-- Blocked Content Modal -->
{#if showBlockedContentModal}
  <div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" on:click={() => showBlockedContentModal = false}>
    <div class="bg-white rounded-lg shadow-xl max-w-6xl w-full max-h-[90vh] overflow-hidden" on:click|stopPropagation>
      <div class="p-6 border-b border-gray-200 flex justify-between items-center">
        <h2 class="text-2xl font-bold text-gray-900">Blocked Content Details</h2>
        <button on:click={() => showBlockedContentModal = false} class="text-gray-500 hover:text-gray-700 text-2xl">&times;</button>
      </div>
      
      <div class="p-6 overflow-y-auto max-h-[calc(90vh-140px)]">
        {#if loadingBlockedContent}
          <div class="text-center py-8">
            <div class="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
            <p class="mt-4 text-gray-600">Loading blocked content...</p>
          </div>
        {:else if blockedContentList.length === 0}
          <div class="text-center py-8 text-gray-500">
            <p>No blocked content found</p>
          </div>
        {:else}
          <div class="space-y-4">
            {#each blockedContentList as match}
              <div class="border border-gray-200 rounded-lg p-4 hover:border-red-300 transition-colors">
                <div class="flex justify-between items-start mb-3">
                  <div>
                    <span class="inline-block px-3 py-1 bg-red-100 text-red-800 text-sm font-semibold rounded">
                      {match.replacement}
                    </span>
                    {#if match.client_name}
                      <span class="ml-2 text-sm text-gray-600">by {match.client_name}</span>
                    {/if}
                  </div>
                  <div class="text-right text-xs text-gray-500">
                    {formatDateTime(match.created_at)}
                  </div>
                </div>
                
                <div class="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                  <div>
                    <span class="font-medium text-gray-700">Model:</span>
                    <span class="ml-2 text-gray-600">{match.model}</span>
                  </div>
                  <div>
                    <span class="font-medium text-gray-700">Provider:</span>
                    <span class="ml-2 text-gray-600">{match.provider}</span>
                  </div>
                  <div>
                    <span class="font-medium text-gray-700">Matches:</span>
                    <span class="ml-2 text-gray-600">{match.match_count}</span>
                  </div>
                  <div>
                    <span class="font-medium text-gray-700">IP:</span>
                    <span class="ml-2 text-gray-600">{match.ip_address || 'N/A'}</span>
                  </div>
                </div>
                
                <div class="mt-3 text-xs text-gray-500">
                  <span class="font-medium">Request ID:</span> {match.request_id}
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>
      
      <div class="p-6 border-t border-gray-200 flex justify-end">
        <button 
          on:click={() => showBlockedContentModal = false}
          class="px-6 py-2 bg-gray-600 text-white rounded hover:bg-gray-700"
        >
          Close
        </button>
      </div>
    </div>
  </div>
{/if}
