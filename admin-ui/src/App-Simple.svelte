<script>
  let apiKey = '';
  let isLoggedIn = false;

  function handleLogin() {
    if (apiKey.length > 0) {
      isLoggedIn = true;
      console.log('Login erfolgreich!');
    }
  }

  function handleLogout() {
    isLoggedIn = false;
    apiKey = '';
  }
</script>

<main class="min-h-screen bg-gray-100 p-8">
  {#if !isLoggedIn}
    <!-- Login Screen -->
    <div class="flex items-center justify-center min-h-screen">
      <div class="bg-white p-8 rounded-lg shadow-lg w-96">
        <h1 class="text-3xl font-bold mb-6 text-center">LLM-Proxy Admin</h1>
        
        <div class="mb-4">
          <label class="block mb-2 font-medium">Admin API Key</label>
          <input 
            type="text" 
            bind:value={apiKey}
            class="w-full p-2 border rounded"
            placeholder="admin_dev_key_..."
          />
        </div>
        
        <button 
          on:click={handleLogin}
          class="w-full bg-blue-600 text-white py-2 rounded hover:bg-blue-700"
        >
          Login
        </button>
        
        <div class="mt-4 text-center text-sm text-gray-600">
          Test Key: admin_dev_key_12345678901234567890123456789012
        </div>
      </div>
    </div>
  {:else}
    <!-- Dashboard -->
    <div>
      <div class="bg-white p-6 rounded-lg shadow mb-6">
        <div class="flex justify-between items-center">
          <h1 class="text-2xl font-bold">Dashboard</h1>
          <button 
            on:click={handleLogout}
            class="bg-red-600 text-white px-4 py-2 rounded hover:bg-red-700"
          >
            Logout
          </button>
        </div>
      </div>
      
      <div class="grid grid-cols-3 gap-4">
        <div class="bg-white p-6 rounded-lg shadow">
          <h3 class="text-xl font-bold mb-2">✅ Frontend</h3>
          <p class="text-gray-600">Läuft korrekt!</p>
        </div>
        
        <div class="bg-white p-6 rounded-lg shadow">
          <h3 class="text-xl font-bold mb-2">🔧 Backend</h3>
          <p class="text-gray-600">API verfügbar</p>
        </div>
        
        <div class="bg-white p-6 rounded-lg shadow">
          <h3 class="text-xl font-bold mb-2">🎨 Tailwind</h3>
          <p class="text-gray-600">CSS funktioniert</p>
        </div>
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
