<script>
  import { apiKey, currentPage } from './lib/stores.js';
  import Login from './components/Login.svelte';
  import Dashboard from './components/Dashboard.svelte';
  import Clients from './components/Clients.svelte';
  import Cache from './components/Cache.svelte';
  import Stats from './components/Stats.svelte';
  import Filters from './components/Filters.svelte';

  let isAuthenticated = false;
  
  apiKey.subscribe(value => {
    isAuthenticated = value && value.length > 0;
  });

  function logout() {
    apiKey.clear();
  }
</script>

<main class="min-h-screen bg-gray-100">
  {#if !isAuthenticated}
    <Login />
  {:else}
    <div class="flex h-screen">
      <!-- Sidebar -->
      <aside class="w-64 bg-gray-900 text-white">
        <div class="p-4">
          <h1 class="text-2xl font-bold mb-8">LLM-Proxy Admin</h1>
          <nav class="space-y-2">
            <button on:click={() => currentPage.set('dashboard')} 
                    class="w-full text-left px-4 py-2 rounded hover:bg-gray-800">
              📊 Dashboard
            </button>
            <button on:click={() => currentPage.set('clients')} 
                    class="w-full text-left px-4 py-2 rounded hover:bg-gray-800">
              👥 Clients
            </button>
            <button on:click={() => currentPage.set('filters')} 
                    class="w-full text-left px-4 py-2 rounded hover:bg-gray-800">
              🔒 Filters
            </button>
            <button on:click={() => currentPage.set('cache')} 
                    class="w-full text-left px-4 py-2 rounded hover:bg-gray-800">
              💾 Cache
            </button>
            <button on:click={() => currentPage.set('stats')} 
                    class="w-full text-left px-4 py-2 rounded hover:bg-gray-800">
              📈 Statistics
            </button>
          </nav>
          <div class="mt-8">
            <button on:click={logout} 
                    class="w-full px-4 py-2 bg-red-600 rounded hover:bg-red-700">
              Logout
            </button>
          </div>
        </div>
      </aside>

      <!-- Main Content -->
      <div class="flex-1 overflow-auto">
        {#if $currentPage === 'dashboard'}
          <Dashboard />
        {:else if $currentPage === 'clients'}
          <Clients />
        {:else if $currentPage === 'filters'}
          <Filters />
        {:else if $currentPage === 'cache'}
          <Cache />
        {:else if $currentPage === 'stats'}
          <Stats />
        {/if}
      </div>
    </div>
  {/if}
</main>

<style>
  :global(body) {
    margin: 0;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  }
</style>
