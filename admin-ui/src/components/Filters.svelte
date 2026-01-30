<script>
  import { apiKey } from '../lib/stores.js';
  import AdminAPI from '../lib/api.js';
  import { onMount } from 'svelte';

  let api;
  let filters = [];
  let stats = {};
  let loading = false;
  let error = null;
  let showCreateModal = false;
  let showBulkModal = false;
  let selectedFilter = null;

  // New Filter Form
  let newFilter = {
    pattern: '',
    replacement: '',
    filter_type: 'word',
    priority: 100,
    description: '',
    case_sensitive: false,
    enabled: true
  };

  // Bulk Import
  let bulkText = '';
  let bulkResult = null;

  // Test
  let testText = '';
  let testResult = null;

  $: {
    if ($apiKey) {
      api = new AdminAPI($apiKey);
      loadFilters();
      loadStats();
    }
  }

  async function loadFilters() {
    loading = true;
    error = null;
    try {
      const data = await api.listFilters();
      filters = data.filters || [];
    } catch (err) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  async function loadStats() {
    try {
      stats = await api.getFilterStats();
    } catch (err) {
      console.error('Failed to load stats:', err);
    }
  }

  async function createFilter() {
    loading = true;
    error = null;
    try {
      await api.createFilter(newFilter);
      await loadFilters();
      await loadStats();
      showCreateModal = false;
      resetForm();
    } catch (err) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  async function deleteFilter(id) {
    if (!confirm('Are you sure you want to delete this filter?')) return;
    
    loading = true;
    error = null;
    try {
      await api.deleteFilter(id);
      await loadFilters();
      await loadStats();
    } catch (err) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  async function toggleFilter(filter) {
    loading = true;
    error = null;
    try {
      await api.updateFilter(filter.id, { enabled: !filter.enabled });
      await loadFilters();
    } catch (err) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  async function bulkImport() {
    loading = true;
    error = null;
    bulkResult = null;

    try {
      // Parse bulk text (CSV format)
      const lines = bulkText.trim().split('\n');
      const filtersToImport = [];

      for (const line of lines) {
        if (!line.trim() || line.startsWith('#')) continue;

        const parts = line.split(',').map(p => p.trim());
        
        if (parts.length >= 3) {
          filtersToImport.push({
            pattern: parts[0],
            replacement: parts[1],
            filter_type: parts[2] || 'word',
            priority: parseInt(parts[3]) || 100,
            description: parts[4] || '',
            case_sensitive: parts[5] === 'true',
            enabled: parts[6] !== 'false'
          });
        }
      }

      if (filtersToImport.length === 0) {
        error = 'No valid filters found in text';
        return;
      }

      const result = await api.bulkImportFilters(filtersToImport);
      bulkResult = result;
      
      await loadFilters();
      await loadStats();
      
      if (result.failed.length === 0) {
        bulkText = '';
        setTimeout(() => {
          showBulkModal = false;
          bulkResult = null;
        }, 2000);
      }
    } catch (err) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  async function testFilterFunc(filterId) {
    if (!testText) return;
    
    loading = true;
    error = null;
    try {
      testResult = await api.testFilter(filterId, testText);
    } catch (err) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  async function refreshCache() {
    loading = true;
    error = null;
    try {
      await api.refreshFilters();
      await loadStats();
      alert('Filter cache refreshed successfully');
    } catch (err) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  function resetForm() {
    newFilter = {
      pattern: '',
      replacement: '',
      filter_type: 'word',
      priority: 100,
      description: '',
      case_sensitive: false,
      enabled: true
    };
  }

  function getFilterTypeBadge(type) {
    const colors = {
      word: 'bg-blue-100 text-blue-800',
      phrase: 'bg-green-100 text-green-800',
      regex: 'bg-purple-100 text-purple-800'
    };
    return colors[type] || 'bg-gray-100 text-gray-800';
  }

  onMount(() => {
    if ($apiKey) {
      loadFilters();
      loadStats();
    }
  });
</script>

<div class="p-8">
  <div class="flex justify-between items-center mb-6">
    <div>
      <h1 class="text-3xl font-bold text-gray-900">Content Filters</h1>
      <p class="text-gray-600 mt-1">Manage content filtering rules</p>
    </div>
    <div class="space-x-2">
      <button on:click={refreshCache} class="px-4 py-2 bg-gray-600 text-white rounded hover:bg-gray-700">
        🔄 Refresh Cache
      </button>
      <button on:click={() => showBulkModal = true} class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700">
        📦 Bulk Import
      </button>
      <button on:click={() => showCreateModal = true} class="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700">
        ➕ Add Filter
      </button>
    </div>
  </div>

  {#if error}
    <div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
      {error}
    </div>
  {/if}

  <!-- Statistics -->
  <div class="grid grid-cols-4 gap-4 mb-6">
    <div class="bg-white p-4 rounded-lg shadow">
      <div class="text-sm text-gray-600">Total Filters</div>
      <div class="text-2xl font-bold">{stats.total_filters || 0}</div>
    </div>
    <div class="bg-white p-4 rounded-lg shadow">
      <div class="text-sm text-gray-600">Enabled</div>
      <div class="text-2xl font-bold text-green-600">{stats.enabled_filters || 0}</div>
    </div>
    <div class="bg-white p-4 rounded-lg shadow">
      <div class="text-sm text-gray-600">Total Matches</div>
      <div class="text-2xl font-bold text-blue-600">{stats.total_matches || 0}</div>
    </div>
    <div class="bg-white p-4 rounded-lg shadow">
      <div class="text-sm text-gray-600">Cache Age</div>
      <div class="text-2xl font-bold">{stats.cache_age_seconds || 0}s</div>
    </div>
  </div>

  {#if stats.by_type}
    <div class="bg-white p-4 rounded-lg shadow mb-6">
      <h3 class="font-semibold mb-2">Filters by Type</h3>
      <div class="flex space-x-4">
        <div>Word: <span class="font-bold">{stats.by_type.word || 0}</span></div>
        <div>Phrase: <span class="font-bold">{stats.by_type.phrase || 0}</span></div>
        <div>Regex: <span class="font-bold">{stats.by_type.regex || 0}</span></div>
      </div>
    </div>
  {/if}

  <!-- Filters List -->
  <div class="bg-white shadow-md rounded-lg overflow-hidden">
    <table class="min-w-full">
      <thead class="bg-gray-50">
        <tr>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">ID</th>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Type</th>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Pattern</th>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Replacement</th>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Priority</th>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Matches</th>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Actions</th>
        </tr>
      </thead>
      <tbody class="bg-white divide-y divide-gray-200">
        {#if loading}
          <tr>
            <td colspan="8" class="px-6 py-4 text-center text-gray-500">Loading...</td>
          </tr>
        {:else if filters.length === 0}
          <tr>
            <td colspan="8" class="px-6 py-4 text-center text-gray-500">No filters found</td>
          </tr>
        {:else}
          {#each filters as filter}
            <tr>
              <td class="px-6 py-4 whitespace-nowrap text-sm">{filter.id}</td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span class="px-2 py-1 text-xs rounded {getFilterTypeBadge(filter.filter_type)}">
                  {filter.filter_type}
                </span>
              </td>
              <td class="px-6 py-4 text-sm font-mono max-w-xs truncate">{filter.pattern}</td>
              <td class="px-6 py-4 text-sm font-mono">{filter.replacement}</td>
              <td class="px-6 py-4 text-sm">{filter.priority}</td>
              <td class="px-6 py-4 text-sm">{filter.match_count || 0}</td>
              <td class="px-6 py-4 whitespace-nowrap">
                <button 
                  on:click={() => toggleFilter(filter)}
                  class="px-2 py-1 text-xs rounded {filter.enabled ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'}">
                  {filter.enabled ? '✓ Enabled' : '✗ Disabled'}
                </button>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm space-x-2">
                <button 
                  on:click={() => { selectedFilter = filter; testText = ''; testResult = null; }}
                  class="text-blue-600 hover:text-blue-900">
                  Test
                </button>
                <button 
                  on:click={() => deleteFilter(filter.id)}
                  class="text-red-600 hover:text-red-900">
                  Delete
                </button>
              </td>
            </tr>
          {/each}
        {/if}
      </tbody>
    </table>
  </div>
</div>

<!-- Create Filter Modal -->
{#if showCreateModal}
  <div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
    <div class="bg-white rounded-lg p-6 max-w-md w-full">
      <h2 class="text-2xl font-bold mb-4">Create New Filter</h2>
      
      <div class="space-y-4">
        <div>
          <label class="block text-sm font-medium mb-1">Pattern</label>
          <input bind:value={newFilter.pattern} class="w-full border rounded px-3 py-2" />
        </div>

        <div>
          <label class="block text-sm font-medium mb-1">Replacement</label>
          <input bind:value={newFilter.replacement} class="w-full border rounded px-3 py-2" />
        </div>

        <div>
          <label class="block text-sm font-medium mb-1">Type</label>
          <select bind:value={newFilter.filter_type} class="w-full border rounded px-3 py-2">
            <option value="word">Word</option>
            <option value="phrase">Phrase</option>
            <option value="regex">Regex</option>
          </select>
        </div>

        <div>
          <label class="block text-sm font-medium mb-1">Priority</label>
          <input type="number" bind:value={newFilter.priority} class="w-full border rounded px-3 py-2" />
        </div>

        <div>
          <label class="block text-sm font-medium mb-1">Description</label>
          <input bind:value={newFilter.description} class="w-full border rounded px-3 py-2" />
        </div>

        <div class="flex items-center">
          <input type="checkbox" bind:checked={newFilter.case_sensitive} id="case" class="mr-2" />
          <label for="case" class="text-sm">Case Sensitive</label>
        </div>

        <div class="flex items-center">
          <input type="checkbox" bind:checked={newFilter.enabled} id="enabled" class="mr-2" />
          <label for="enabled" class="text-sm">Enabled</label>
        </div>
      </div>

      <div class="mt-6 flex space-x-2">
        <button on:click={createFilter} class="flex-1 bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700">
          Create
        </button>
        <button on:click={() => { showCreateModal = false; resetForm(); }} class="flex-1 bg-gray-300 px-4 py-2 rounded hover:bg-gray-400">
          Cancel
        </button>
      </div>
    </div>
  </div>
{/if}

<!-- Bulk Import Modal -->
{#if showBulkModal}
  <div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
    <div class="bg-white rounded-lg p-6 max-w-2xl w-full">
      <h2 class="text-2xl font-bold mb-4">Bulk Import Filters</h2>
      
      <p class="text-sm text-gray-600 mb-4">
        Format: <code>pattern, replacement, type, priority, description, case_sensitive, enabled</code>
      </p>

      <textarea 
        bind:value={bulkText} 
        rows="10"
        placeholder="badword,[FILTERED],word,100,Offensive language,false,true"
        class="w-full border rounded px-3 py-2 font-mono text-sm"></textarea>

      {#if bulkResult}
        <div class="mt-4 p-4 bg-green-50 border border-green-200 rounded">
          <div class="font-semibold">Import Results:</div>
          <div>✓ Success: {bulkResult.success.length}</div>
          <div>✗ Failed: {bulkResult.failed.length}</div>
          {#if bulkResult.failed.length > 0}
            <div class="mt-2 text-sm text-red-600">
              {#each bulkResult.failed as fail}
                <div>• {fail}</div>
              {/each}
            </div>
          {/if}
        </div>
      {/if}

      <div class="mt-6 flex space-x-2">
        <button on:click={bulkImport} class="flex-1 bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700">
          Import
        </button>
        <button on:click={() => { showBulkModal = false; bulkText = ''; bulkResult = null; }} class="flex-1 bg-gray-300 px-4 py-2 rounded hover:bg-gray-400">
          Cancel
        </button>
      </div>
    </div>
  </div>
{/if}

<!-- Test Filter Modal -->
{#if selectedFilter}
  <div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
    <div class="bg-white rounded-lg p-6 max-w-2xl w-full">
      <h2 class="text-2xl font-bold mb-4">Test Filter: {selectedFilter.pattern}</h2>
      
      <div class="mb-4">
        <label class="block text-sm font-medium mb-1">Test Text</label>
        <textarea 
          bind:value={testText} 
          rows="4"
          placeholder="Enter text to test against this filter..."
          class="w-full border rounded px-3 py-2"></textarea>
      </div>

      <button 
        on:click={() => testFilterFunc(selectedFilter.id)} 
        class="w-full bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700 mb-4">
        Run Test
      </button>

      {#if testResult}
        <div class="space-y-4">
          <div>
            <div class="font-semibold text-sm text-gray-600">Original:</div>
            <div class="p-3 bg-gray-50 rounded border">{testResult.original_text}</div>
          </div>
          <div>
            <div class="font-semibold text-sm text-gray-600">Filtered:</div>
            <div class="p-3 bg-green-50 rounded border">{testResult.filtered_text}</div>
          </div>
          <div>
            <div class="font-semibold text-sm text-gray-600">Matches: {testResult.matches.length}</div>
            {#if testResult.matches.length > 0}
              <div class="text-sm text-gray-600">
                {#each testResult.matches as match}
                  <div>• {match.pattern} → {match.replacement} ({match.match_count} times)</div>
                {/each}
              </div>
            {/if}
          </div>
        </div>
      {/if}

      <div class="mt-6">
        <button on:click={() => { selectedFilter = null; testText = ''; testResult = null; }} class="w-full bg-gray-300 px-4 py-2 rounded hover:bg-gray-400">
          Close
        </button>
      </div>
    </div>
  </div>
{/if}
