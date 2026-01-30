<script>
  import { apiKey } from '../lib/stores.js';
  let keyInput = 'admin_dev_key_12345678901234567890123456789012';
  
  function handleLogin() {
    if (keyInput.trim()) {
      apiKey.set(keyInput.trim());
    }
  }
</script>

<div class="min-h-screen flex items-center justify-center bg-gray-900">
  <div class="bg-white p-8 rounded-lg shadow-lg w-96">
    <h2 class="text-2xl font-bold mb-6 text-center">LLM-Proxy Admin</h2>
    <form on:submit|preventDefault={handleLogin}>
      <label class="block mb-2 text-sm font-medium">Admin API Key</label>
      <input type="password" bind:value={keyInput} 
             class="w-full p-2 border rounded mb-4" 
             placeholder="Enter your admin API key" required />
      <button type="submit" class="w-full bg-blue-600 text-white p-2 rounded hover:bg-blue-700">
        Login
      </button>
    </form>
    <p class="mt-4 text-xs text-gray-500 text-center">
      Default: admin_dev_key_12345678901234567890123456789012
    </p>
  </div>
</div>
