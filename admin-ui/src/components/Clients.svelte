<script>
  import { onMount } from 'svelte';
  import { apiKey } from '../lib/stores.js';
  import AdminAPI from '../lib/api.js';
  
  let clients = [];
  let showForm = false;
  let formData = { client_id: '', client_secret: '', name: '', grant_types: ['client_credentials'], default_scope: 'read write' };
  
  async function loadClients() {
    const api = new AdminAPI($apiKey);
    const data = await api.listClients();
    clients = data.clients || [];
  }
  
  async function createClient() {
    const api = new AdminAPI($apiKey);
    await api.createClient(formData);
    showForm = false;
    formData = { client_id: '', client_secret: '', name: '', grant_types: ['client_credentials'], default_scope: 'read write' };
    await loadClients();
  }
  
  async function deleteClient(id) {
    if (confirm('Delete this client?')) {
      const api = new AdminAPI($apiKey);
      await api.deleteClient(id);
      await loadClients();
    }
  }
  
  onMount(loadClients);
</script>

<div class="p-8">
  <div class="flex justify-between items-center mb-8">
    <h1 class="text-3xl font-bold">OAuth Clients</h1>
    <button on:click={() => showForm = !showForm} class="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700">
      + New Client
    </button>
  </div>
  
  {#if showForm}
    <div class="bg-white p-6 rounded-lg shadow mb-8">
      <h2 class="text-xl font-bold mb-4">Create New Client</h2>
      <form on:submit|preventDefault={createClient} class="space-y-4">
        <div>
          <label class="block mb-1">Client ID</label>
          <input type="text" bind:value={formData.client_id} class="w-full p-2 border rounded" required />
        </div>
        <div>
          <label class="block mb-1">Client Secret</label>
          <input type="text" bind:value={formData.client_secret} class="w-full p-2 border rounded" required />
        </div>
        <div>
          <label class="block mb-1">Name</label>
          <input type="text" bind:value={formData.name} class="w-full p-2 border rounded" required />
        </div>
        <div>
          <label class="block mb-1">Default Scope</label>
          <input type="text" bind:value={formData.default_scope} class="w-full p-2 border rounded" />
        </div>
        <div class="flex gap-2">
          <button type="submit" class="bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700">Create</button>
          <button type="button" on:click={() => showForm = false} class="bg-gray-400 text-white px-4 py-2 rounded hover:bg-gray-500">Cancel</button>
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
            <td class="px-6 py-4">{client.client_id}</td>
            <td class="px-6 py-4">{client.name}</td>
            <td class="px-6 py-4">{client.default_scope}</td>
            <td class="px-6 py-4">
              <span class:text-green-600={client.enabled} class:text-red-600={!client.enabled}>
                {client.enabled ? 'Enabled' : 'Disabled'}
              </span>
            </td>
            <td class="px-6 py-4">
              <button on:click={() => deleteClient(client.client_id)} class="text-red-600 hover:text-red-800">Delete</button>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>
