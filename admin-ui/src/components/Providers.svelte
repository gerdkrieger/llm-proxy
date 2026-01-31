<script>
  import { onMount } from 'svelte';
  import { apiKey } from '../lib/stores.js';
  import AdminAPI from '../lib/api.js';
  
  let providerStatus = null;
  let providerDetails = null;
  let loading = true;
  let error = null;
  
  // Modal states
  let showConfigModal = false;
  let showTestModal = false;
  let showModelsModal = false;
  let selectedProvider = null;
  let configData = null;
  let testResult = null;
  let testLoading = false;
  let modelsData = null;
  let modelsLoading = false;
  let modelsSaving = false;
  
  async function loadProviderStatus() {
    loading = true;
    error = null;
    try {
      const api = new AdminAPI($apiKey);
      [providerStatus, providerDetails] = await Promise.all([
        api.getProviderStatus(),
        api.getProviderDetails()
      ]);
    } catch (err) {
      error = err.message;
      console.error('Failed to load provider status:', err);
    } finally {
      loading = false;
    }
  }
  
  async function refreshStatus() {
    await loadProviderStatus();
  }
  
  function getProviderIcon(type) {
    const icons = {
      'claude': 'C',
      'openai': 'AI',
      'default': '?'
    };
    return icons[type] || icons['default'];
  }
  
  function getProviderColor(type) {
    const colors = {
      'claude': 'from-purple-500 to-pink-500',
      'openai': 'from-green-500 to-teal-500',
      'default': 'from-gray-500 to-gray-700'
    };
    return colors[type] || colors['default'];
  }
  
  // View Config
  async function viewConfig(provider) {
    selectedProvider = provider;
    const api = new AdminAPI($apiKey);
    try {
      configData = await api.getProviderConfig(provider.id);
      showConfigModal = true;
    } catch (err) {
      alert('Failed to load config: ' + err.message);
    }
  }
  
  // Test Connection
  async function testConnection(provider) {
    selectedProvider = provider;
    showTestModal = true;
    testLoading = true;
    testResult = null;
    
    const api = new AdminAPI($apiKey);
    try {
      testResult = await api.testProvider(provider.id);
    } catch (err) {
      testResult = { status: 'error', error: err.message };
    } finally {
      testLoading = false;
    }
  }
  
  // Toggle Provider
  async function toggleProvider(provider) {
    const newState = !provider.enabled;
    const confirmMsg = newState 
      ? `Enable ${provider.name}?` 
      : `Disable ${provider.name}? This will stop all requests to this provider.`;
    
    if (!confirm(confirmMsg)) return;
    
    const api = new AdminAPI($apiKey);
    try {
      await api.toggleProvider(provider.id, newState);
      await refreshStatus();
      alert(`${provider.name} ${newState ? 'enabled' : 'disabled'} successfully`);
    } catch (err) {
      alert('Failed to toggle provider: ' + err.message);
    }
  }

  // Model Management
  async function openModelsModal(provider) {
    selectedProvider = provider;
    modelsLoading = true;
    showModelsModal = true;
    modelsData = null;
    
    const api = new AdminAPI($apiKey);
    try {
      const response = await api.getProviderModels(provider.id);
      modelsData = response;
    } catch (err) {
      alert('Failed to load models: ' + err.message);
      showModelsModal = false;
    } finally {
      modelsLoading = false;
    }
  }

  async function saveModelConfiguration() {
    if (!modelsData || !modelsData.models) return;
    
    modelsSaving = true;
    const enabledModels = modelsData.models
      .filter(m => m.enabled)
      .map(m => m.id);
    
    const api = new AdminAPI($apiKey);
    try {
      await api.configureProviderModels(selectedProvider.id, enabledModels);
      alert(`Model configuration saved! ${enabledModels.length} models enabled.`);
      showModelsModal = false;
      await refreshStatus();
    } catch (err) {
      alert('Failed to save configuration: ' + err.message);
    } finally {
      modelsSaving = false;
    }
  }

  function toggleModel(model) {
    model.enabled = !model.enabled;
    modelsData = modelsData; // Trigger reactivity
  }

  function selectAllModels() {
    if (modelsData && modelsData.models) {
      modelsData.models.forEach(m => m.enabled = true);
      modelsData = modelsData;
    }
  }

  function deselectAllModels() {
    if (modelsData && modelsData.models) {
      modelsData.models.forEach(m => m.enabled = false);
      modelsData = modelsData;
    }
  }
  
  onMount(loadProviderStatus);
  
  // Refresh every 30 seconds
  let refreshInterval;
  onMount(() => {
    refreshInterval = setInterval(refreshStatus, 30000);
    return () => clearInterval(refreshInterval);
  });
</script>

<div class="p-8">
  <div class="flex justify-between items-center mb-8">
    <div>
      <h1 class="text-3xl font-bold text-gray-900">LLM Providers</h1>
      <p class="text-gray-600 mt-1">Manage and monitor connected LLM providers</p>
    </div>
    <button on:click={refreshStatus} class="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700 flex items-center gap-2">
      <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
        <path fill-rule="evenodd" d="M4 2a1 1 0 011 1v2.101a7.002 7.002 0 0111.601 2.566 1 1 0 11-1.885.666A5.002 5.002 0 005.999 7H9a1 1 0 010 2H4a1 1 0 01-1-1V3a1 1 0 011-1zm.008 9.057a1 1 0 011.276.61A5.002 5.002 0 0014.001 13H11a1 1 0 110-2h5a1 1 0 011 1v5a1 1 0 11-2 0v-2.101a7.002 7.002 0 01-11.601-2.566 1 1 0 01.61-1.276z" clip-rule="evenodd" />
      </svg>
      Refresh
    </button>
  </div>

  {#if error}
    <div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
      <strong>Error:</strong> {error}
    </div>
  {/if}

  {#if loading}
    <div class="text-center py-12">
      <div class="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      <p class="mt-4 text-gray-600">Loading provider status...</p>
    </div>
  {:else if providerStatus}
    <!-- Overall Status Card -->
    <div class="bg-white rounded-lg shadow mb-6">
      <div class="p-6">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-4">
            <div class="flex-shrink-0">
              {#if providerStatus.healthy}
                <div class="h-16 w-16 bg-green-100 rounded-full flex items-center justify-center">
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 text-green-600" viewBox="0 0 20 20" fill="currentColor">
                    <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                  </svg>
                </div>
              {:else}
                <div class="h-16 w-16 bg-red-100 rounded-full flex items-center justify-center">
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 text-red-600" viewBox="0 0 20 20" fill="currentColor">
                    <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
                  </svg>
                </div>
              {/if}
            </div>
            <div>
              <h2 class="text-2xl font-bold text-gray-900">
                {providerStatus.healthy ? 'All Systems Operational' : 'System Issues Detected'}
              </h2>
              <p class="text-gray-600">
                {providerStatus.provider_count} provider{providerStatus.provider_count !== 1 ? 's' : ''} connected
              </p>
            </div>
          </div>
          <div class="text-right">
            <div class="text-sm text-gray-500">Last updated</div>
            <div class="text-lg font-semibold text-gray-900">{new Date().toLocaleTimeString()}</div>
          </div>
        </div>
        
        {#if providerStatus.error}
          <div class="mt-4 p-3 bg-red-50 border border-red-200 rounded">
            <p class="text-red-800 text-sm"><strong>Error:</strong> {providerStatus.error}</p>
          </div>
        {/if}
      </div>
    </div>

    <!-- Available Models Card -->
    <div class="bg-white rounded-lg shadow mb-6">
      <div class="p-6">
        <h3 class="text-xl font-bold text-gray-900 mb-4 flex items-center gap-2">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          Available Models
        </h3>
        
        {#if providerStatus.models && providerStatus.models.length > 0}
          <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {#each providerStatus.models as model}
              <div class="border border-gray-200 rounded-lg p-4 hover:border-blue-400 hover:shadow-md transition-all">
                <div class="flex items-start justify-between">
                  <div class="flex-1">
                    <h4 class="font-semibold text-gray-900 mb-1">{model}</h4>
                    <p class="text-sm text-gray-600">
                      {#if model.includes('opus')}
                        Most capable model
                      {:else if model.includes('sonnet')}
                        Balanced performance
                      {:else if model.includes('haiku')}
                        Fast and efficient
                      {:else if model.includes('gpt-4-turbo')}
                        Most advanced GPT-4 model
                      {:else if model.includes('gpt-4')}
                        Advanced reasoning
                      {:else if model.includes('gpt-3.5')}
                        Fast and cost-effective
                      {:else}
                        LLM Model
                      {/if}
                    </p>
                  </div>
                  <div class="flex-shrink-0 ml-3">
                    <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                      Active
                    </span>
                  </div>
                </div>
                
                <div class="mt-3 flex items-center gap-2 text-xs text-gray-500">
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
                    <path d="M2 11a1 1 0 011-1h2a1 1 0 011 1v5a1 1 0 01-1 1H3a1 1 0 01-1-1v-5zM8 7a1 1 0 011-1h2a1 1 0 011 1v9a1 1 0 01-1 1H9a1 1 0 01-1-1V7zM14 4a1 1 0 011-1h2a1 1 0 011 1v12a1 1 0 01-1 1h-2a1 1 0 01-1-1V4z" />
                  </svg>
                  Provider: {#if model.includes('claude')}Claude{:else if model.includes('gpt')}OpenAI{:else}Unknown{/if}
                </div>
              </div>
            {/each}
          </div>
        {:else}
          <p class="text-gray-500">No models available</p>
        {/if}
      </div>
    </div>

    <!-- Provider Details Card -->
    <div class="bg-white rounded-lg shadow">
      <div class="p-6">
        <h3 class="text-xl font-bold text-gray-900 mb-4 flex items-center gap-2">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-purple-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
          </svg>
          Provider Configuration
        </h3>
        
        <div class="space-y-4">
          {#if providerDetails && providerDetails.providers && providerDetails.providers.length > 0}
            <!-- Dynamic Provider Cards -->
            {#each providerDetails.providers as provider}
              <div class="border border-gray-200 rounded-lg p-4">
                <div class="flex items-center justify-between mb-3">
                  <div class="flex items-center gap-3">
                    <div class="h-12 w-12 bg-gradient-to-br {getProviderColor(provider.type)} rounded-lg flex items-center justify-center text-white font-bold text-xl">
                      {getProviderIcon(provider.type)}
                    </div>
                    <div>
                      <h4 class="font-semibold text-gray-900">{provider.name}</h4>
                      <p class="text-sm text-gray-600">
                        {#if provider.type === 'claude'}
                          AI Assistant Provider
                        {:else if provider.type === 'openai'}
                          Language Model Provider
                        {:else}
                          LLM Provider
                        {/if}
                      </p>
                    </div>
                  </div>
                  <div>
                    <span class="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium {provider.status === 'healthy' ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}">
                      <span class="mr-1.5 h-2 w-2 {provider.status === 'healthy' ? 'bg-green-500' : 'bg-red-500'} rounded-full"></span>
                      {provider.enabled ? 'Enabled' : 'Disabled'}
                    </span>
                  </div>
                </div>
                
                <div class="grid grid-cols-2 md:grid-cols-4 gap-4 mt-4 pt-4 border-t border-gray-100">
                  <div>
                    <div class="text-xs text-gray-500 uppercase mb-1">Models</div>
                    <div class="text-lg font-semibold text-gray-900">{provider.models?.length || 0}</div>
                  </div>
                  <div>
                    <div class="text-xs text-gray-500 uppercase mb-1">Status</div>
                    <div class="text-lg font-semibold {provider.status === 'healthy' ? 'text-green-600' : 'text-red-600'}">
                      {provider.status === 'healthy' ? 'Healthy' : 'Unhealthy'}
                    </div>
                  </div>
                  <div>
                    <div class="text-xs text-gray-500 uppercase mb-1">API Keys</div>
                    <div class="text-lg font-semibold text-gray-900">{provider.api_keys || 0}</div>
                  </div>
                  <div>
                    <div class="text-xs text-gray-500 uppercase mb-1">Requests</div>
                    <div class="text-lg font-semibold text-blue-600">Active</div>
                  </div>
                </div>
                
                <div class="mt-4 flex gap-2">
                  <button 
                    on:click={() => viewConfig(provider)}
                    class="px-3 py-1.5 text-sm bg-blue-600 text-white rounded hover:bg-blue-700"
                  >
                    View Config
                  </button>
                  <button 
                    on:click={() => testConnection(provider)}
                    class="px-3 py-1.5 text-sm bg-green-600 text-white rounded hover:bg-green-700"
                  >
                    Test Connection
                  </button>
                  <button 
                    on:click={() => openModelsModal(provider)}
                    class="px-3 py-1.5 text-sm bg-purple-600 text-white rounded hover:bg-purple-700"
                  >
                    Manage Models
                  </button>
                  <button 
                    on:click={() => toggleProvider(provider)}
                    class="px-3 py-1.5 text-sm {provider.enabled ? 'bg-red-600 hover:bg-red-700' : 'bg-green-600 hover:bg-green-700'} text-white rounded"
                  >
                    {provider.enabled ? 'Disable' : 'Enable'}
                  </button>
                </div>
              </div>
            {/each}
            
            <!-- Add More Providers Placeholder -->
            <div class="border border-dashed border-gray-300 rounded-lg p-6 text-center bg-gray-50">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-12 w-12 mx-auto text-gray-400 mb-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
              </svg>
              <h4 class="text-gray-700 font-medium mb-1">Add More Providers</h4>
              <p class="text-sm text-gray-500 mb-3">Connect additional LLM providers like Google Gemini, Cohere, or Ollama</p>
              <button class="px-4 py-2 bg-gray-300 text-gray-500 rounded cursor-not-allowed" disabled title="Feature coming soon">
                Add Provider (Coming Soon)
              </button>
            </div>
          {:else}
            <!-- No Providers Configured -->
            <div class="border border-dashed border-gray-300 rounded-lg p-6 text-center bg-gray-50">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-12 w-12 mx-auto text-gray-400 mb-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <h4 class="text-gray-700 font-medium mb-1">No Providers Configured</h4>
              <p class="text-sm text-gray-500">Configure LLM providers in your config.yaml file</p>
            </div>
          {/if}
        </div>
      </div>
    </div>
  {/if}
</div>

<!-- Config Modal -->
{#if showConfigModal && configData}
  <div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" on:click={() => showConfigModal = false}>
    <div class="bg-white rounded-lg shadow-xl max-w-2xl w-full max-h-[80vh] overflow-hidden" on:click|stopPropagation>
      <div class="p-6 border-b border-gray-200 flex justify-between items-center">
        <h2 class="text-2xl font-bold text-gray-900">{configData.provider_name} Configuration</h2>
        <button on:click={() => showConfigModal = false} class="text-gray-500 hover:text-gray-700 text-2xl">&times;</button>
      </div>
      
      <div class="p-6 overflow-y-auto max-h-[calc(80vh-140px)]">
        <div class="space-y-4">
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-700">Provider ID</label>
              <div class="mt-1 text-gray-900">{configData.provider_id}</div>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700">Status</label>
              <div class="mt-1">
                <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium {configData.enabled ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}">
                  {configData.enabled ? 'Enabled' : 'Disabled'}
                </span>
              </div>
            </div>
          </div>
          
          <div>
            <label class="block text-sm font-medium text-gray-700">API Keys Configured</label>
            <div class="mt-1 text-gray-900">{configData.api_keys}</div>
          </div>
          
          <div>
            <label class="block text-sm font-medium text-gray-700">Available Models</label>
            <div class="mt-2 space-y-2">
              {#each configData.models as model}
                <div class="px-3 py-2 bg-gray-50 rounded border border-gray-200">
                  {model}
                </div>
              {/each}
            </div>
          </div>
        </div>
      </div>
      
      <div class="p-6 border-t border-gray-200 flex justify-end">
        <button 
          on:click={() => showConfigModal = false}
          class="px-6 py-2 bg-gray-600 text-white rounded hover:bg-gray-700"
        >
          Close
        </button>
      </div>
    </div>
  </div>
{/if}

<!-- Test Connection Modal -->
{#if showTestModal}
  <div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" on:click={() => showTestModal = false}>
    <div class="bg-white rounded-lg shadow-xl max-w-2xl w-full max-h-[80vh] overflow-hidden" on:click|stopPropagation>
      <div class="p-6 border-b border-gray-200 flex justify-between items-center">
        <h2 class="text-2xl font-bold text-gray-900">Test {selectedProvider?.name} Connection</h2>
        <button on:click={() => showTestModal = false} class="text-gray-500 hover:text-gray-700 text-2xl">&times;</button>
      </div>
      
      <div class="p-6 overflow-y-auto max-h-[calc(80vh-140px)]">
        {#if testLoading}
          <div class="text-center py-8">
            <div class="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
            <p class="mt-4 text-gray-600">Testing connection...</p>
          </div>
        {:else if testResult}
          <div class="space-y-4">
            <div class="flex items-center gap-2">
              <span class="text-lg font-semibold">Status:</span>
              {#if testResult.status === 'success'}
                <span class="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-green-100 text-green-800">
                  ✓ Success
                </span>
              {:else}
                <span class="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-red-100 text-red-800">
                  ✗ Failed
                </span>
              {/if}
            </div>
            
            {#if testResult.models}
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">Available Models ({testResult.models.length})</label>
                <div class="space-y-2">
                  {#each testResult.models as model}
                    <div class="px-3 py-2 bg-green-50 rounded border border-green-200">
                      {model}
                    </div>
                  {/each}
                </div>
              </div>
            {/if}
            
            {#if testResult.error}
              <div class="mt-4 p-4 bg-red-50 border border-red-200 rounded">
                <p class="text-sm text-red-800">{testResult.error}</p>
              </div>
            {/if}
          </div>
        {/if}
      </div>
      
      <div class="p-6 border-t border-gray-200 flex justify-end">
        <button 
          on:click={() => showTestModal = false}
          class="px-6 py-2 bg-gray-600 text-white rounded hover:bg-gray-700"
        >
          Close
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  /* Custom scrollbar for better UX */
  :global(body) {
    scrollbar-width: thin;
    scrollbar-color: #cbd5e0 #f7fafc;
  }
  
  :global(body::-webkit-scrollbar) {
    width: 8px;
  }
  
  :global(body::-webkit-scrollbar-track) {
    background: #f7fafc;
  }
  
  :global(body::-webkit-scrollbar-thumb) {
    background-color: #cbd5e0;
    border-radius: 4px;
  }
</style>

<!-- Model Management Modal -->
{#if showModelsModal}
  <div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
    <div class="bg-white rounded-lg p-6 max-w-3xl w-full max-h-[80vh] overflow-y-auto m-4">
      <h3 class="text-xl font-semibold mb-4">
        Manage Models - {selectedProvider?.name}
      </h3>
      
      {#if modelsLoading}
        <div class="flex justify-center py-8">
          <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
        </div>
      {:else if modelsData}
        <div class="mb-4 flex justify-between items-center">
          <div class="text-sm text-gray-600">
            {modelsData.models.filter(m => m.enabled).length} of {modelsData.total} models enabled
          </div>
          <div class="space-x-2">
            <button
              on:click={selectAllModels}
              class="px-3 py-1 text-sm bg-blue-600 text-white rounded hover:bg-blue-700"
            >
              Select All
            </button>
            <button
              on:click={deselectAllModels}
              class="px-3 py-1 text-sm bg-gray-600 text-white rounded hover:bg-gray-700"
            >
              Deselect All
            </button>
          </div>
        </div>

        <div class="space-y-2 max-h-96 overflow-y-auto border rounded p-3">
          {#each modelsData.models as model}
            <label class="flex items-start space-x-3 p-2 hover:bg-gray-50 rounded cursor-pointer">
              <input
                type="checkbox"
                bind:checked={model.enabled}
                class="mt-1 h-4 w-4 text-blue-600"
              />
              <div class="flex-1">
                <div class="font-medium text-sm">{model.name}</div>
                <div class="text-xs text-gray-500 font-mono">{model.id}</div>
                {#if model.capabilities && model.capabilities.length > 0}
                  <div class="flex gap-1 mt-1">
                    {#each model.capabilities as cap}
                      <span class="px-2 py-0.5 text-xs bg-blue-100 text-blue-700 rounded">
                        {cap}
                      </span>
                    {/each}
                  </div>
                {/if}
              </div>
            </label>
          {/each}
        </div>

        <div class="mt-6 flex justify-end space-x-3">
          <button
            on:click={() => showModelsModal = false}
            class="px-4 py-2 bg-gray-200 text-gray-700 rounded hover:bg-gray-300"
            disabled={modelsSaving}
          >
            Cancel
          </button>
          <button
            on:click={saveModelConfiguration}
            class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
            disabled={modelsSaving}
          >
            {modelsSaving ? 'Saving...' : 'Save Configuration'}
          </button>
        </div>
      {/if}
    </div>
  </div>
{/if}
