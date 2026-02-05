<script>
  import { apiKey } from '../lib/stores.js';
  import AdminAPI from '../lib/api.js';
  
  let keyInput = '';
  let loading = false;
  let error = null;
  
  async function handleLogin() {
    if (!keyInput.trim()) return;
    
    loading = true;
    error = null;
    
    try {
      // Test the API key by making a simple request
      const testApi = new AdminAPI(keyInput.trim());
      await testApi.getProviderStatus();
      
      // If we get here, the key is valid
      apiKey.set(keyInput.trim());
    } catch (err) {
      // Key is invalid
      error = 'Invalid API key. Please check and try again.';
      console.error('Login failed:', err);
    } finally {
      loading = false;
    }
  }
</script>

<div class="min-h-screen flex items-center justify-center bg-gray-900">
  <div class="bg-white p-8 rounded-lg shadow-lg w-96">
    <h2 class="text-2xl font-bold mb-6 text-center">LLM-Proxy Admin</h2>
    
    {#if error}
      <div class="mb-4 p-3 bg-red-100 border border-red-400 text-red-700 rounded">
        {error}
      </div>
    {/if}
    
    <form on:submit|preventDefault={handleLogin}>
      <label class="block mb-2 text-sm font-medium">Admin API Key</label>
      <input 
        type="password" 
        bind:value={keyInput} 
        class="w-full p-2 border rounded mb-4" 
        placeholder="Enter your admin API key" 
        disabled={loading}
        required 
      />
      <button 
        type="submit" 
        class="w-full bg-blue-600 text-white p-2 rounded hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed flex items-center justify-center"
        disabled={loading}
      >
        {#if loading}
          <svg class="animate-spin h-5 w-5 mr-2" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          Validating...
        {:else}
          Login
        {/if}
      </button>
    </form>
    <p class="mt-4 text-xs text-gray-500 text-center">
      Get your API key from the .env file or contact your administrator
    </p>
  </div>
</div>
