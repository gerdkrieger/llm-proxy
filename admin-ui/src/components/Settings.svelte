<script>
  import { onMount } from 'svelte';
  import { apiKey } from '../lib/stores';

  let settings = [];
  let loading = true;
  let saving = false;
  let message = { type: '', text: '' };

  // Settings state
  let captureRequestResponseBodies = true;

  async function fetchSettings() {
    loading = true;
    try {
      const API_BASE = window.location.origin;
      const response = await fetch(`${API_BASE}/admin/settings`, {
        headers: { 'X-Admin-API-Key': $apiKey },
      });
      if (!response.ok) throw new Error('Failed to fetch settings');
      
      const data = await response.json();
      settings = data.settings || [];
      
      // Parse settings
      const captureSetting = settings.find(s => s.key === 'capture_request_response_bodies');
      if (captureSetting) {
        captureRequestResponseBodies = captureSetting.value === 'true';
      }
    } catch (error) {
      showMessage('error', 'Failed to load settings: ' + error.message);
    } finally {
      loading = false;
    }
  }

  async function updateSetting(key, value) {
    saving = true;
    
    try {
      const API_BASE = window.location.origin;
      const response = await fetch(`${API_BASE}/admin/settings`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'X-Admin-API-Key': $apiKey,
        },
        body: JSON.stringify({ key, value: String(value) }),
      });

      if (!response.ok) throw new Error('Failed to update setting');
      
      showMessage('success', 'Setting updated successfully');
      await fetchSettings(); // Refresh settings
    } catch (error) {
      showMessage('error', 'Failed to update setting: ' + error.message);
      // Revert the toggle on error
      await fetchSettings();
    } finally {
      saving = false;
    }
  }

  function handleCaptureToggle() {
    updateSetting('capture_request_response_bodies', captureRequestResponseBodies);
  }

  function showMessage(type, text) {
    message = { type, text };
    setTimeout(() => {
      message = { type: '', text: '' };
    }, 5000);
  }

  onMount(() => {
    fetchSettings();
  });
</script>

<div class="max-w-4xl mx-auto">
  <div class="mb-6">
    <h1 class="text-3xl font-bold text-gray-900">System Settings</h1>
    <p class="text-gray-600 mt-2">Configure system-wide settings and behaviors</p>
  </div>

  {#if message.text}
    <div class="mb-6 p-4 rounded-lg {message.type === 'success' ? 'bg-green-50 border border-green-200 text-green-800' : 'bg-red-50 border border-red-200 text-red-800'}">
      {message.text}
    </div>
  {/if}

  {#if loading}
    <div class="text-center py-12">
      <div class="text-4xl mb-4">⏳</div>
      <div class="text-gray-600">Loading settings...</div>
    </div>
  {:else}
    <!-- Request/Response Body Capture Section -->
    <div class="bg-white rounded-lg shadow-md mb-6">
      <div class="px-6 py-4 border-b border-gray-200">
        <h2 class="text-xl font-semibold text-gray-900">Request & Response Logging</h2>
      </div>
      
      <div class="p-6">
        <div class="flex items-center justify-between">
          <div class="flex-1">
            <h3 class="text-lg font-medium text-gray-900">Capture Request & Response Bodies</h3>
            <p class="text-sm text-gray-600 mt-1">
              When enabled, the system will capture and store request and response bodies for debugging and analysis.
              This helps monitor LLM interactions but increases storage usage.
            </p>
            <div class="mt-3">
              <div class="text-xs text-gray-500">
                <div class="flex items-center gap-2 mb-1">
                  <span class="font-medium">✓ Enabled:</span>
                  <span>Full visibility into requests and responses in Live Monitor</span>
                </div>
                <div class="flex items-center gap-2 mb-1">
                  <span class="font-medium">✗ Disabled:</span>
                  <span>Only metadata captured (headers, status codes, token usage)</span>
                </div>
              </div>
            </div>
          </div>
          
          <div class="ml-6">
            <button
              disabled={saving}
              on:click={() => {
                captureRequestResponseBodies = !captureRequestResponseBodies;
                handleCaptureToggle();
              }}
              class="relative inline-flex h-8 w-14 items-center rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 {captureRequestResponseBodies ? 'bg-blue-600' : 'bg-gray-300'} {saving ? 'opacity-50 cursor-not-allowed' : ''}"
            >
              <span class="sr-only">Toggle body capture</span>
              <span
                class="inline-block h-6 w-6 transform rounded-full bg-white transition-transform {captureRequestResponseBodies ? 'translate-x-7' : 'translate-x-1'}"
              />
            </button>
          </div>
        </div>

        {#if captureRequestResponseBodies}
          <div class="mt-4 p-4 bg-blue-50 border border-blue-200 rounded-lg">
            <div class="flex items-start">
              <div class="text-2xl mr-3">💡</div>
              <div class="text-sm text-blue-800">
                <div class="font-medium mb-1">Body Capture is Active</div>
                <div>
                  Request and response bodies are being captured and stored. You can view them in the Live Monitor 
                  by clicking on any request to see the detail modal with full request/response bodies.
                </div>
              </div>
            </div>
          </div>
        {:else}
          <div class="mt-4 p-4 bg-yellow-50 border border-yellow-200 rounded-lg">
            <div class="flex items-start">
              <div class="text-2xl mr-3">⚠️</div>
              <div class="text-sm text-yellow-800">
                <div class="font-medium mb-1">Body Capture is Disabled</div>
                <div>
                  Only metadata (headers, status codes, token usage) is being captured. 
                  Enable this setting to see full request and response bodies in the Live Monitor.
                </div>
              </div>
            </div>
          </div>
        {/if}
      </div>
    </div>

    <!-- Future Settings Sections -->
    <div class="bg-gray-50 rounded-lg border-2 border-dashed border-gray-300 p-8 text-center">
      <div class="text-4xl mb-3">⚙️</div>
      <div class="text-gray-600">
        <div class="font-medium mb-2">More Settings Coming Soon</div>
        <div class="text-sm">Additional system configuration options will be added here.</div>
      </div>
    </div>
  {/if}
</div>

<style>
  /* Custom styles if needed */
</style>
