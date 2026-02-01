<script>
  import { onMount } from 'svelte';
  import { apiKey } from '../lib/stores.js';
  import AdminAPI from '../lib/api.js';
  
  let clients = [];
  let showForm = false;
  let editMode = false;
  let currentClient = null;
  let formData = { client_id: '', client_secret: '', name: '', grant_types: ['client_credentials'], default_scope: 'read write' };
  
  async function loadClients() {
    const api = new AdminAPI($apiKey);
    const data = await api.listClients();
    clients = data.clients || [];
  }
  
  async function createClient() {
    const api = new AdminAPI($apiKey);
    await api.createClient(formData);
    closeForm();
    await loadClients();
  }
  
  async function updateClient() {
    const api = new AdminAPI($apiKey);
    await api.updateClient(currentClient.client_id, {
      name: formData.name,
      default_scope: formData.default_scope,
      grant_types: formData.grant_types
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
    formData = { client_id: '', client_secret: '', name: '', grant_types: ['client_credentials'], default_scope: 'read write' };
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
      default_scope: client.default_scope
    };
    showForm = true;
  }
  
  function closeForm() {
    showForm = false;
    editMode = false;
    currentClient = null;
    formData = { client_id: '', client_secret: '', name: '', grant_types: ['client_credentials'], default_scope: 'read write' };
  }
  
  function handleSubmit() {
    if (editMode) {
      updateClient();
    } else {
      createClient();
    }
  }
  
  onMount(loadClients);
</script>

<div class="p-8">
  <div class="flex justify-between items-center mb-8">
    <h1 class="text-3xl font-bold">OAuth Clients</h1>
    <button on:click={openCreateForm} class="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700">
      + New Client
    </button>
  </div>
  
  {#if showForm}
    <div class="bg-white p-6 rounded-lg shadow mb-8">
      <h2 class="text-xl font-bold mb-4">{editMode ? 'Edit Client' : 'Create New Client'}</h2>
      <form on:submit|preventDefault={handleSubmit} class="space-y-4">
        <div>
          <label class="block mb-1 font-medium">Client ID</label>
          <input 
            type="text" 
            bind:value={formData.client_id} 
            class="w-full p-2 border rounded" 
            required 
            disabled={editMode}
            class:bg-gray-100={editMode}
          />
          {#if editMode}
            <p class="text-xs text-gray-500 mt-1">Client ID cannot be changed</p>
          {/if}
        </div>
        
        {#if !editMode}
          <div>
            <label class="block mb-1 font-medium">Client Secret</label>
            <input type="text" bind:value={formData.client_secret} class="w-full p-2 border rounded" required />
          </div>
        {/if}
        
        <div>
          <label class="block mb-1 font-medium">Name</label>
          <input type="text" bind:value={formData.name} class="w-full p-2 border rounded" required />
        </div>
        
        <div>
          <label class="block mb-1 font-medium">Default Scope</label>
          <input type="text" bind:value={formData.default_scope} class="w-full p-2 border rounded" />
        </div>
        
        <div class="flex gap-2">
          <button type="submit" class="bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700">
            {editMode ? 'Update' : 'Create'}
          </button>
          <button type="button" on:click={closeForm} class="bg-gray-400 text-white px-4 py-2 rounded hover:bg-gray-500">
            Cancel
          </button>
        </div>
      </form>
    </div>
  {/if}
  
  <div class="bg-white rounded-lg shadow overflow-hidden">
    <table class="w-full">
      <thead class="bg-gray-50">
        <tr>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Client ID</th>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Name</th>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Scope</th>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Actions</th>
        </tr>
      </thead>
      <tbody class="divide-y">
        {#each clients as client}
          <tr>
            <td class="px-6 py-4 font-mono text-sm">{client.client_id}</td>
            <td class="px-6 py-4">{client.name}</td>
            <td class="px-6 py-4 text-sm text-gray-600">{client.default_scope}</td>
            <td class="px-6 py-4">
              <button 
                on:click={() => toggleEnabled(client)}
                class="px-3 py-1 rounded-full text-sm font-medium transition-colors"
                class:bg-green-100={client.enabled}
                class:text-green-800={client.enabled}
                class:bg-red-100={!client.enabled}
                class:text-red-800={!client.enabled}
                class:hover:bg-green-200={client.enabled}
                class:hover:bg-red-200={!client.enabled}
              >
                {client.enabled ? '✓ Enabled' : '✗ Disabled'}
              </button>
            </td>
            <td class="px-6 py-4">
              <div class="flex gap-3">
                <button 
                  on:click={() => openEditForm(client)} 
                  class="text-blue-600 hover:text-blue-800 font-medium"
                >
                  Edit
                </button>
                <button 
                  on:click={() => deleteClient(client.client_id)} 
                  class="text-red-600 hover:text-red-800 font-medium"
                >
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
        <p>No clients found. Create your first OAuth client to get started.</p>
      </div>
    {/if}
  </div>
</div>
