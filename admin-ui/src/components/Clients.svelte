<script>
  import { onMount } from 'svelte';
  import { apiKey } from '../lib/stores.js';
  import AdminAPI from '../lib/api.js';
  
  let clients = [];
  let availableModels = []; // All models from all providers
  let showForm = false;
  let editMode = false;
  let currentClient = null;
  let formData = { 
    client_id: '', 
    client_secret: '', 
    name: '', 
    grant_types: ['client_credentials'], 
    default_scope: 'read write',
    allowed_models: null  // null = all models allowed
  };
  let modelSelectionMode = 'all'; // 'all', 'none', 'specific'
  let selectedModels = []; // For 'specific' mode
  
  // Reset Secret Modal
  let showResetSecretModal = false;
  let resetSecretClient = null;
  let resetSecretMode = 'generate'; // 'generate' or 'custom'
  let customSecret = '';
  let newSecretResult = null;
  let resetSecretLoading = false;
  let secretCopied = false;
  
  async function loadClients() {
    const api = new AdminAPI($apiKey);
    const data = await api.listClients();
    clients = data.clients || [];
  }
  
  async function loadAvailableModels() {
    const api = new AdminAPI($apiKey);
    try {
      const data = await api.getProviderDetails();
      // Flatten all models from all providers into one array
      const allModels = [];
      if (data.providers) {
        data.providers.forEach(provider => {
          if (provider.models) {
            provider.models.forEach(model => {
              allModels.push({
                id: model,
                name: model,
                provider: provider.name
              });
            });
          }
        });
      }
      availableModels = allModels.sort((a, b) => a.name.localeCompare(b.name));
    } catch (err) {
      console.error('Failed to load models:', err);
    }
  }
  
  async function createClient() {
    const api = new AdminAPI($apiKey);
    
    // Convert selection mode to allowed_models value
    let allowed_models = null;
    if (modelSelectionMode === 'none') {
      allowed_models = [];
    } else if (modelSelectionMode === 'specific') {
      allowed_models = selectedModels;
    }
    // 'all' mode stays null
    
    await api.createClient({
      ...formData,
      allowed_models
    });
    closeForm();
    await loadClients();
  }
  
  async function updateClient() {
    const api = new AdminAPI($apiKey);
    
    // Convert selection mode to allowed_models value
    let allowed_models = null;
    if (modelSelectionMode === 'none') {
      allowed_models = [];
    } else if (modelSelectionMode === 'specific') {
      allowed_models = selectedModels;
    }
    
    await api.updateClient(currentClient.client_id, {
      name: formData.name,
      default_scope: formData.default_scope,
      grant_types: formData.grant_types,
      allowed_models
    });
    closeForm();
    await loadClients();
  }
  
  async function toggleEnabled(client) {
    const api = new AdminAPI($apiKey);
    await api.updateClient(client.client_id, {
      enabled: !client.enabled
    });
    await loadClients();
  }
  
  async function deleteClient(id) {
    if (confirm('Delete this client? This action cannot be undone.')) {
      const api = new AdminAPI($apiKey);
      await api.deleteClient(id);
      await loadClients();
    }
  }
  
  function openCreateForm() {
    editMode = false;
    currentClient = null;
    formData = { 
      client_id: '', 
      client_secret: '', 
      name: '', 
      grant_types: ['client_credentials'], 
      default_scope: 'read write',
      allowed_models: null
    };
    modelSelectionMode = 'all';
    selectedModels = [];
    showForm = true;
  }
  
  function openEditForm(client) {
    editMode = true;
    currentClient = client;
    formData = {
      client_id: client.client_id,
      client_secret: '', // Don't show existing secret
      name: client.name,
      grant_types: client.grant_types,
      default_scope: client.default_scope,
      allowed_models: client.allowed_models
    };
    
    // Determine selection mode based on allowed_models
    if (client.allowed_models === null || client.allowed_models === undefined) {
      modelSelectionMode = 'all';
      selectedModels = [];
    } else if (client.allowed_models.length === 0) {
      modelSelectionMode = 'none';
      selectedModels = [];
    } else {
      modelSelectionMode = 'specific';
      selectedModels = [...client.allowed_models];
    }
    
    showForm = true;
  }
  
  function closeForm() {
    showForm = false;
    editMode = false;
    currentClient = null;
    formData = { 
      client_id: '', 
      client_secret: '', 
      name: '', 
      grant_types: ['client_credentials'], 
      default_scope: 'read write',
      allowed_models: null
    };
    modelSelectionMode = 'all';
    selectedModels = [];
  }
  
  function handleSubmit() {
    if (editMode) {
      updateClient();
    } else {
      createClient();
    }
  }
  
  function toggleModel(modelId) {
    if (selectedModels.includes(modelId)) {
      selectedModels = selectedModels.filter(m => m !== modelId);
    } else {
      selectedModels = [...selectedModels, modelId];
    }
  }
  
  function selectAllModels() {
    selectedModels = availableModels.map(m => m.id);
  }
  
  function deselectAllModels() {
    selectedModels = [];
  }
  
  function openResetSecretModal(client) {
    resetSecretClient = client;
    resetSecretMode = 'generate';
    customSecret = '';
    newSecretResult = null;
    secretCopied = false;
    showResetSecretModal = true;
  }
  
  function closeResetSecretModal() {
    showResetSecretModal = false;
    resetSecretClient = null;
    newSecretResult = null;
    customSecret = '';
    secretCopied = false;
  }
  
  async function resetSecret() {
    resetSecretLoading = true;
    secretCopied = false;
    try {
      const api = new AdminAPI($apiKey);
      const secret = resetSecretMode === 'custom' ? customSecret : '';
      const result = await api.resetClientSecret(resetSecretClient.client_id, secret);
      newSecretResult = result.new_secret;
    } catch (err) {
      alert('Failed to reset secret: ' + err.message);
    } finally {
      resetSecretLoading = false;
    }
  }
  
  async function copySecret() {
    if (newSecretResult) {
      await navigator.clipboard.writeText(newSecretResult);
      secretCopied = true;
      setTimeout(() => { secretCopied = false; }, 3000);
    }
  }
  
  onMount(() => {
    loadClients();
    loadAvailableModels();
  });
</script>

<div class="p-8">
  <div class="flex justify-between items-center mb-8">
    <div>
      <h1 class="text-3xl font-bold text-gray-900">API Clients</h1>
      <p class="text-gray-600 mt-1">Manage API clients and their model access permissions</p>
    </div>
    <button on:click={openCreateForm} class="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700">
      + New Client
    </button>
  </div>
  
  {#if showForm}
    <div class="bg-white p-6 rounded-lg shadow mb-8">
      <h2 class="text-xl font-bold mb-4">{editMode ? 'Edit Client' : 'Create New Client'}</h2>
      <form on:submit|preventDefault={handleSubmit}>
        <div class="grid grid-cols-2 gap-4 mb-4">
          <div>
            <label class="block text-sm font-medium mb-1">Client ID</label>
            <input type="text" bind:value={formData.client_id} 
                   class="w-full p-2 border rounded" 
                   required disabled={editMode} />
          </div>
          {#if !editMode}
            <div>
              <label class="block text-sm font-medium mb-1">Client Secret</label>
              <input type="password" bind:value={formData.client_secret} 
                     class="w-full p-2 border rounded" 
                     required />
            </div>
          {/if}
          <div>
            <label class="block text-sm font-medium mb-1">Name</label>
            <input type="text" bind:value={formData.name} 
                   class="w-full p-2 border rounded" 
                   required />
          </div>
        </div>
        
        <!-- Model Access Control -->
        <div class="mb-6 p-4 bg-gray-50 rounded-lg border border-gray-200">
          <h3 class="text-lg font-semibold mb-3 flex items-center gap-2">
            <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
              <path d="M10 2a8 8 0 100 16 8 8 0 000-16zM9 9a1 1 0 112 0v4a1 1 0 11-2 0V9zm1-4a1 1 0 100 2 1 1 0 000-2z"/>
            </svg>
            Model Access Control
          </h3>
          <p class="text-sm text-gray-600 mb-4">
            Control which LLM models this client can access. Unauthorized requests will return a 403 error.
          </p>
          
          <div class="space-y-3">
            <label class="flex items-center cursor-pointer">
              <input type="radio" bind:group={modelSelectionMode} value="all" class="mr-2" />
              <span class="font-medium">All Models</span>
              <span class="ml-2 text-sm text-gray-600">(Default - no restrictions)</span>
            </label>
            
            <label class="flex items-center cursor-pointer">
              <input type="radio" bind:group={modelSelectionMode} value="specific" class="mr-2" />
              <span class="font-medium">Specific Models Only</span>
              <span class="ml-2 text-sm text-gray-600">(Whitelist selected models)</span>
            </label>
            
            <label class="flex items-center cursor-pointer">
              <input type="radio" bind:group={modelSelectionMode} value="none" class="mr-2" />
              <span class="font-medium">No Models</span>
              <span class="ml-2 text-sm text-gray-600">(Block all requests)</span>
            </label>
          </div>
          
          {#if modelSelectionMode === 'specific'}
            <div class="mt-4 p-4 bg-white rounded border border-gray-300">
              <div class="flex justify-between items-center mb-3">
                <span class="font-medium">
                  Select Models ({selectedModels.length} of {availableModels.length})
                </span>
                <div class="space-x-2">
                  <button type="button" on:click={selectAllModels} 
                          class="text-sm text-blue-600 hover:underline">
                    Select All
                  </button>
                  <button type="button" on:click={deselectAllModels} 
                          class="text-sm text-gray-600 hover:underline">
                    Clear
                  </button>
                </div>
              </div>
              
              <div class="max-h-64 overflow-y-auto space-y-2">
                {#each availableModels as model}
                  <label class="flex items-start p-2 hover:bg-gray-50 rounded cursor-pointer">
                    <input 
                      type="checkbox" 
                      checked={selectedModels.includes(model.id)}
                      on:change={() => toggleModel(model.id)}
                      class="mt-1 mr-3"
                    />
                    <div class="flex-1">
                      <div class="font-medium text-sm">{model.name}</div>
                      <div class="text-xs text-gray-500">{model.provider}</div>
                    </div>
                  </label>
                {/each}
              </div>
            </div>
          {/if}
        </div>
        
        <div class="flex justify-end gap-2">
          <button type="button" on:click={closeForm} class="px-4 py-2 bg-gray-300 rounded hover:bg-gray-400">
            Cancel
          </button>
          <button type="submit" class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700">
            {editMode ? 'Update' : 'Create'} Client
          </button>
        </div>
      </form>
    </div>
  {/if}
  
  <div class="bg-white rounded-lg shadow overflow-hidden">
    <table class="w-full">
      <thead class="bg-gray-50">
        <tr>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Name</th>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Client ID</th>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Model Access</th>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Actions</th>
        </tr>
      </thead>
      <tbody class="divide-y divide-gray-200">
        {#each clients as client}
          <tr>
            <td class="px-6 py-4">{client.name}</td>
            <td class="px-6 py-4">
              <code class="text-sm bg-gray-100 px-2 py-1 rounded">{client.client_id}</code>
            </td>
            <td class="px-6 py-4">
              {#if client.allowed_models === null || client.allowed_models === undefined}
                <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                  All Models
                </span>
              {:else if client.allowed_models.length === 0}
                <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-100 text-red-800">
                  No Access
                </span>
              {:else}
                <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                  {client.allowed_models.length} Models
                </span>
                <div class="mt-1 text-xs text-gray-500">
                  {client.allowed_models.slice(0, 3).join(', ')}
                  {#if client.allowed_models.length > 3}
                    <span class="text-gray-400">+{client.allowed_models.length - 3} more</span>
                  {/if}
                </div>
              {/if}
            </td>
            <td class="px-6 py-4">
              <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium {client.enabled ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'}">
                {client.enabled ? 'Enabled' : 'Disabled'}
              </span>
            </td>
            <td class="px-6 py-4">
              <div class="flex gap-2">
                <button on:click={() => openEditForm(client)} 
                        class="text-blue-600 hover:text-blue-800">
                  Edit
                </button>
                <button on:click={() => openResetSecretModal(client)} 
                        class="text-orange-600 hover:text-orange-800">
                  Reset Secret
                </button>
                <button on:click={() => toggleEnabled(client)} 
                        class="text-yellow-600 hover:text-yellow-800">
                  {client.enabled ? 'Disable' : 'Enable'}
                </button>
                <button on:click={() => deleteClient(client.client_id)} 
                        class="text-red-600 hover:text-red-800">
                  Delete
                </button>
              </div>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
    
    {#if clients.length === 0}
      <div class="p-8 text-center text-gray-500">
        <p>No clients found. Create your first API client to get started.</p>
      </div>
    {/if}
  </div>

  <!-- Reset Secret Modal -->
  {#if showResetSecretModal && resetSecretClient}
    <div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white rounded-lg shadow-xl max-w-lg w-full mx-4">
        <div class="p-6 border-b border-gray-200 flex justify-between items-center">
          <h2 class="text-xl font-bold text-gray-900">Reset Client Secret</h2>
          <button on:click={closeResetSecretModal} class="text-gray-500 hover:text-gray-700 text-2xl font-bold">&times;</button>
        </div>
        
        <div class="p-6">
          <div class="mb-4 p-3 bg-gray-50 rounded">
            <div class="text-sm text-gray-600">Client</div>
            <div class="font-semibold">{resetSecretClient.name}</div>
            <div class="text-sm text-gray-500 font-mono">{resetSecretClient.client_id}</div>
          </div>
          
          {#if !newSecretResult}
            <!-- Secret Mode Selection -->
            <div class="mb-4">
              <div class="text-sm font-medium text-gray-700 mb-2">How would you like to set the new secret?</div>
              <div class="space-y-2">
                <label class="flex items-center cursor-pointer p-2 rounded hover:bg-gray-50">
                  <input type="radio" bind:group={resetSecretMode} value="generate" class="mr-3" />
                  <div>
                    <div class="font-medium">Auto-generate</div>
                    <div class="text-sm text-gray-500">Generate a secure random secret</div>
                  </div>
                </label>
                <label class="flex items-center cursor-pointer p-2 rounded hover:bg-gray-50">
                  <input type="radio" bind:group={resetSecretMode} value="custom" class="mr-3" />
                  <div>
                    <div class="font-medium">Custom secret</div>
                    <div class="text-sm text-gray-500">Enter your own secret (min. 16 characters)</div>
                  </div>
                </label>
              </div>
            </div>
            
            {#if resetSecretMode === 'custom'}
              <div class="mb-4">
                <label class="block text-sm font-medium text-gray-700 mb-1">New Secret</label>
                <input 
                  type="text" 
                  bind:value={customSecret} 
                  placeholder="Enter at least 16 characters..."
                  class="w-full p-2 border rounded font-mono text-sm {customSecret.length > 0 && customSecret.length < 16 ? 'border-red-500' : ''}"
                />
                {#if customSecret.length > 0 && customSecret.length < 16}
                  <div class="text-red-500 text-xs mt-1">Secret must be at least 16 characters ({customSecret.length}/16)</div>
                {/if}
              </div>
            {/if}
            
            <div class="bg-yellow-50 border border-yellow-200 rounded p-3 mb-4">
              <div class="text-sm text-yellow-800">
                <strong>Warning:</strong> This will invalidate the current secret. Any applications using the old secret will stop working immediately.
              </div>
            </div>
            
            <div class="flex justify-end gap-2">
              <button on:click={closeResetSecretModal} class="px-4 py-2 bg-gray-300 rounded hover:bg-gray-400">
                Cancel
              </button>
              <button 
                on:click={resetSecret}
                disabled={resetSecretLoading || (resetSecretMode === 'custom' && customSecret.length < 16)}
                class="px-4 py-2 bg-orange-600 text-white rounded hover:bg-orange-700 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {resetSecretLoading ? 'Resetting...' : 'Reset Secret'}
              </button>
            </div>
          {:else}
            <!-- Secret Result -->
            <div class="mb-4 p-4 bg-green-50 border border-green-200 rounded">
              <div class="text-sm font-medium text-green-800 mb-2">New secret generated successfully!</div>
              <div class="text-sm text-green-700 mb-3">Copy this secret now. It will not be shown again.</div>
              <div class="flex items-center gap-2">
                <code class="flex-1 p-3 bg-white border border-green-300 rounded font-mono text-sm break-all select-all">
                  {newSecretResult}
                </code>
                <button 
                  on:click={copySecret}
                  class="px-3 py-3 rounded text-sm font-medium {secretCopied ? 'bg-green-600 text-white' : 'bg-gray-200 hover:bg-gray-300 text-gray-700'}"
                >
                  {secretCopied ? 'Copied!' : 'Copy'}
                </button>
              </div>
            </div>
            
            <div class="flex justify-end">
              <button on:click={closeResetSecretModal} class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700">
                Done
              </button>
            </div>
          {/if}
        </div>
      </div>
    </div>
  {/if}
</div>
