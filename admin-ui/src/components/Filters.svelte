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
  let showEditModal = false;
  let showBulkModal = false;
  let selectedFilter = null;
  let editingFilter = null;

  // Replacement Templates
  const replacementTemplates = {
    'custom': 'Custom (type your own)',
    // PII
    '[EMAIL]': '🆔 PII - Email Address',
    '[PHONE]': '🆔 PII - Phone Number',
    '[SSN]': '🆔 PII - Social Security Number',
    '[TAX_ID]': '🆔 PII - Tax ID',
    '[PASSPORT]': '🆔 PII - Passport Number',
    '[DRIVER_LICENSE]': '🆔 PII - Driver License',
    '[NATIONAL_ID]': '🆔 PII - National ID',
    '[MRN]': '🆔 PII - Medical Record Number',
    // Financial
    '[CREDIT_CARD]': '💳 Financial - Credit Card',
    '[CVV]': '💳 Financial - CVV/CVC Code',
    '[IBAN]': '💳 Financial - IBAN',
    '[BIC]': '💳 Financial - BIC/SWIFT',
    '[BANK_ACCOUNT]': '💳 Financial - Bank Account',
    '[ROUTING_NUMBER]': '💳 Financial - Routing Number',
    '[CRYPTO_ADDRESS]': '💳 Financial - Crypto Address',
    // Security
    '[***API_KEY***]': '🔐 Security - API Key',
    '[***API_SECRET***]': '🔐 Security - API Secret',
    '[***AWS_KEY***]': '🔐 Security - AWS Access Key',
    '[***AWS_SECRET***]': '🔐 Security - AWS Secret',
    '[***GOOGLE_API_KEY***]': '🔐 Security - Google API Key',
    '[***GITHUB_TOKEN***]': '🔐 Security - GitHub Token',
    '[***GITLAB_TOKEN***]': '🔐 Security - GitLab Token',
    '[***JWT_TOKEN***]': '🔐 Security - JWT Token',
    '[***SSH_PRIVATE_KEY***]': '🔐 Security - SSH Private Key',
    '[***BEARER_TOKEN***]': '🔐 Security - Bearer Token',
    '[***ACCESS_TOKEN***]': '🔐 Security - Access Token',
    '[***PASSWORD***]': '🔐 Security - Password',
    '[***SLACK_TOKEN***]': '🔐 Security - Slack Token',
    '[***STRIPE_KEY***]': '🔐 Security - Stripe Key',
    '[***TWILIO_SID***]': '🔐 Security - Twilio SID',
    '[***SENDGRID_KEY***]': '🔐 Security - SendGrid Key',
    // Technical
    '[***DB_CONNECTION***]': '🗄️ Technical - DB Connection',
    '[***DB_CREDENTIALS***]': '🗄️ Technical - DB Credentials',
    '[***DB_PASSWORD***]': '🗄️ Technical - DB Password',
    '[INTERNAL_IP]': '🗄️ Technical - Internal IP',
    '[INTERNAL_HOST]': '🗄️ Technical - Internal Hostname',
    '[LOCALHOST]': '🗄️ Technical - Localhost',
    '[***SECRET_KEY***]': '🗄️ Technical - Secret Key',
    '[***ENCRYPTION_KEY***]': '🗄️ Technical - Encryption Key',
    '[***DOCKER_LOGIN***]': '🗄️ Technical - Docker Login',
    // Confidential
    '[CONFIDENTIAL]': '🔒 Confidential - Confidential',
    '[REDACTED]': '🔒 Confidential - Redacted',
    '[CLASSIFIED]': '🔒 Confidential - Classified',
    '[INTERNAL_PROJECT]': '🔒 Confidential - Internal Project',
    '[PROPRIETARY]': '🔒 Confidential - Proprietary',
    '[TRADE_SECRET]': '🔒 Confidential - Trade Secret',
    '[SALARY_INFO]': '🔒 Confidential - Salary Info',
    '[HR_DOCUMENT]': '🔒 Confidential - HR Document',
    '[LEGAL_PRIVILEGE]': '🔒 Confidential - Legal Privilege',
    '[COMPETITOR]': '🔒 Confidential - Competitor Name',
    // Additional
    '[UUID]': '🛡️ Additional - UUID',
    '[LICENSE_KEY]': '🛡️ Additional - License Key',
    '[SESSION_TOKEN]': '🛡️ Additional - Session Token',
    '[CSRF_TOKEN]': '🛡️ Additional - CSRF Token',
  };

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
  let newReplacementMode = 'custom'; // 'custom' or template key

  // Edit Filter Form
  let editFilter = {
    pattern: '',
    replacement: '',
    filter_type: 'word',
    priority: 100,
    description: '',
    case_sensitive: false,
    enabled: true
  };
  let editReplacementMode = 'custom';

  // Bulk Import
  let bulkText = '';
  let bulkResult = null;

  // Test
  let testText = '';
  let testResult = null;

  // Filter & Search
  let filterType = 'all'; // all, word, phrase, regex
  let searchQuery = '';
  let filterCategory = 'all'; // all, pii, financial, security, etc.
  let sortBy = 'priority'; // priority, id, pattern

  // Filtered & Sorted filters
  $: filteredFilters = filters
    .filter(f => {
      // Filter by type
      if (filterType !== 'all' && f.filter_type !== filterType) return false;
      
      // Filter by search query
      if (searchQuery) {
        const query = searchQuery.toLowerCase();
        return f.pattern.toLowerCase().includes(query) ||
               f.replacement.toLowerCase().includes(query) ||
               f.description.toLowerCase().includes(query);
      }
      
      // Filter by category (based on replacement text)
      if (filterCategory !== 'all') {
        const replacement = f.replacement.toLowerCase();
        switch(filterCategory) {
          case 'pii':
            return replacement.includes('email') || replacement.includes('phone') || 
                   replacement.includes('ssn') || replacement.includes('tax_id') ||
                   replacement.includes('passport') || replacement.includes('driver');
          case 'financial':
            return replacement.includes('credit_card') || replacement.includes('cvv') ||
                   replacement.includes('iban') || replacement.includes('bank') ||
                   replacement.includes('crypto');
          case 'security':
            return replacement.includes('api') || replacement.includes('key') ||
                   replacement.includes('token') || replacement.includes('password') ||
                   replacement.includes('secret') || replacement.includes('aws') ||
                   replacement.includes('github') || replacement.includes('jwt');
          case 'technical':
            return replacement.includes('db') || replacement.includes('connection') ||
                   replacement.includes('ip') || replacement.includes('host') ||
                   replacement.includes('localhost');
          case 'confidential':
            return replacement.includes('confidential') || replacement.includes('redacted') ||
                   replacement.includes('classified') || replacement.includes('internal') ||
                   replacement.includes('proprietary');
          default:
            return true;
        }
      }
      
      return true;
    })
    .sort((a, b) => {
      switch(sortBy) {
        case 'priority':
          return b.priority - a.priority; // Highest first
        case 'id':
          return a.id - b.id;
        case 'pattern':
          return a.pattern.localeCompare(b.pattern);
        default:
          return 0;
      }
    });

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

  function openEditModal(filter) {
    editingFilter = filter;
    editFilter = {
      pattern: filter.pattern,
      replacement: filter.replacement,
      filter_type: filter.filter_type,
      priority: filter.priority,
      description: filter.description,
      case_sensitive: filter.case_sensitive,
      enabled: filter.enabled
    };
    
    // Check if replacement matches a template
    if (replacementTemplates[filter.replacement]) {
      editReplacementMode = filter.replacement;
    } else {
      editReplacementMode = 'custom';
    }
    
    showEditModal = true;
  }

  async function updateFilter() {
    if (!editingFilter) return;
    
    loading = true;
    error = null;
    try {
      await api.updateFilter(editingFilter.id, editFilter);
      await loadFilters();
      await loadStats();
      showEditModal = false;
      editingFilter = null;
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
    newReplacementMode = 'custom';
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

  <!-- Filter & Search Controls -->
  <div class="bg-white shadow-md rounded-lg p-4 mb-6">
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
      <!-- Search -->
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Search</label>
        <input 
          type="text" 
          bind:value={searchQuery}
          placeholder="Search pattern, replacement..." 
          class="w-full px-3 py-2 border rounded focus:outline-none focus:ring-2 focus:ring-blue-500" />
      </div>

      <!-- Filter by Type -->
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Filter Type</label>
        <select bind:value={filterType} class="w-full px-3 py-2 border rounded focus:outline-none focus:ring-2 focus:ring-blue-500">
          <option value="all">All Types ({filters.length})</option>
          <option value="word">Word ({filters.filter(f => f.filter_type === 'word').length})</option>
          <option value="phrase">Phrase ({filters.filter(f => f.filter_type === 'phrase').length})</option>
          <option value="regex">Regex ({filters.filter(f => f.filter_type === 'regex').length})</option>
        </select>
      </div>

      <!-- Filter by Category -->
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Category</label>
        <select bind:value={filterCategory} class="w-full px-3 py-2 border rounded focus:outline-none focus:ring-2 focus:ring-blue-500">
          <option value="all">All Categories</option>
          <option value="pii">🆔 PII (Personal Info)</option>
          <option value="financial">💳 Financial Data</option>
          <option value="security">🔐 Security & Credentials</option>
          <option value="technical">🗄️ Technical Secrets</option>
          <option value="confidential">🔒 Confidential</option>
        </select>
      </div>

      <!-- Sort By -->
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Sort By</label>
        <select bind:value={sortBy} class="w-full px-3 py-2 border rounded focus:outline-none focus:ring-2 focus:ring-blue-500">
          <option value="priority">Priority (High to Low)</option>
          <option value="id">ID (Ascending)</option>
          <option value="pattern">Pattern (A-Z)</option>
        </select>
      </div>
    </div>

    <!-- Results Count -->
    <div class="mt-3 text-sm text-gray-600">
      Showing <span class="font-semibold">{filteredFilters.length}</span> of <span class="font-semibold">{filters.length}</span> filters
      {#if searchQuery}
        <span class="ml-2">
          • Search: "<span class="font-semibold">{searchQuery}</span>"
          <button on:click={() => searchQuery = ''} class="ml-1 text-blue-600 hover:text-blue-800">Clear</button>
        </span>
      {/if}
      {#if filterType !== 'all'}
        <span class="ml-2">
          • Type: <span class="font-semibold">{filterType}</span>
          <button on:click={() => filterType = 'all'} class="ml-1 text-blue-600 hover:text-blue-800">Clear</button>
        </span>
      {/if}
      {#if filterCategory !== 'all'}
        <span class="ml-2">
          • Category: <span class="font-semibold">{filterCategory}</span>
          <button on:click={() => filterCategory = 'all'} class="ml-1 text-blue-600 hover:text-blue-800">Clear</button>
        </span>
      {/if}
    </div>
  </div>

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
        {:else if filteredFilters.length === 0}
          <tr>
            <td colspan="8" class="px-6 py-4 text-center text-gray-500">
              {filters.length === 0 ? 'No filters found' : 'No filters match your search criteria'}
            </td>
          </tr>
        {:else}
          {#each filteredFilters as filter}
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
                  on:click={() => openEditModal(filter)}
                  class="text-purple-600 hover:text-purple-900">
                  Edit
                </button>
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
          <label class="block text-sm font-medium mb-1">Replacement Template</label>
          <select 
            bind:value={newReplacementMode} 
            on:change={() => { 
              if (newReplacementMode !== 'custom') {
                newFilter.replacement = newReplacementMode;
              }
            }}
            class="w-full border rounded px-3 py-2 mb-2 focus:outline-none focus:ring-2 focus:ring-green-500">
            <option value="custom">✏️ Custom (type your own)</option>
            <optgroup label="🆔 PII - Personal Identifiable Information">
              <option value="[EMAIL]">[EMAIL] - Email Address</option>
              <option value="[PHONE]">[PHONE] - Phone Number</option>
              <option value="[SSN]">[SSN] - Social Security Number</option>
              <option value="[TAX_ID]">[TAX_ID] - Tax ID</option>
              <option value="[PASSPORT]">[PASSPORT] - Passport Number</option>
              <option value="[DRIVER_LICENSE]">[DRIVER_LICENSE] - Driver License</option>
              <option value="[NATIONAL_ID]">[NATIONAL_ID] - National ID</option>
              <option value="[MRN]">[MRN] - Medical Record Number</option>
            </optgroup>
            <optgroup label="💳 Financial Data">
              <option value="[CREDIT_CARD]">[CREDIT_CARD] - Credit Card</option>
              <option value="[CVV]">[CVV] - CVV/CVC Code</option>
              <option value="[IBAN]">[IBAN] - IBAN</option>
              <option value="[BIC]">[BIC] - BIC/SWIFT</option>
              <option value="[BANK_ACCOUNT]">[BANK_ACCOUNT] - Bank Account</option>
              <option value="[ROUTING_NUMBER]">[ROUTING_NUMBER] - Routing Number</option>
              <option value="[CRYPTO_ADDRESS]">[CRYPTO_ADDRESS] - Crypto Address</option>
            </optgroup>
            <optgroup label="🔐 Security & Credentials">
              <option value="[***API_KEY***]">[***API_KEY***] - API Key</option>
              <option value="[***API_SECRET***]">[***API_SECRET***] - API Secret</option>
              <option value="[***AWS_KEY***]">[***AWS_KEY***] - AWS Access Key</option>
              <option value="[***AWS_SECRET***]">[***AWS_SECRET***] - AWS Secret</option>
              <option value="[***GOOGLE_API_KEY***]">[***GOOGLE_API_KEY***] - Google API Key</option>
              <option value="[***GITHUB_TOKEN***]">[***GITHUB_TOKEN***] - GitHub Token</option>
              <option value="[***GITLAB_TOKEN***]">[***GITLAB_TOKEN***] - GitLab Token</option>
              <option value="[***JWT_TOKEN***]">[***JWT_TOKEN***] - JWT Token</option>
              <option value="[***SSH_PRIVATE_KEY***]">[***SSH_PRIVATE_KEY***] - SSH Private Key</option>
              <option value="[***BEARER_TOKEN***]">[***BEARER_TOKEN***] - Bearer Token</option>
              <option value="[***ACCESS_TOKEN***]">[***ACCESS_TOKEN***] - Access Token</option>
              <option value="[***PASSWORD***]">[***PASSWORD***] - Password</option>
              <option value="[***SLACK_TOKEN***]">[***SLACK_TOKEN***] - Slack Token</option>
              <option value="[***STRIPE_KEY***]">[***STRIPE_KEY***] - Stripe Key</option>
              <option value="[***TWILIO_SID***]">[***TWILIO_SID***] - Twilio SID</option>
              <option value="[***SENDGRID_KEY***]">[***SENDGRID_KEY***] - SendGrid Key</option>
            </optgroup>
            <optgroup label="🗄️ Technical Secrets">
              <option value="[***DB_CONNECTION***]">[***DB_CONNECTION***] - DB Connection</option>
              <option value="[***DB_CREDENTIALS***]">[***DB_CREDENTIALS***] - DB Credentials</option>
              <option value="[***DB_PASSWORD***]">[***DB_PASSWORD***] - DB Password</option>
              <option value="[INTERNAL_IP]">[INTERNAL_IP] - Internal IP</option>
              <option value="[INTERNAL_HOST]">[INTERNAL_HOST] - Internal Hostname</option>
              <option value="[LOCALHOST]">[LOCALHOST] - Localhost</option>
              <option value="[***SECRET_KEY***]">[***SECRET_KEY***] - Secret Key</option>
              <option value="[***ENCRYPTION_KEY***]">[***ENCRYPTION_KEY***] - Encryption Key</option>
              <option value="[***DOCKER_LOGIN***]">[***DOCKER_LOGIN***] - Docker Login</option>
            </optgroup>
            <optgroup label="🔒 Confidential">
              <option value="[CONFIDENTIAL]">[CONFIDENTIAL] - Confidential</option>
              <option value="[REDACTED]">[REDACTED] - Redacted</option>
              <option value="[CLASSIFIED]">[CLASSIFIED] - Classified</option>
              <option value="[INTERNAL_PROJECT]">[INTERNAL_PROJECT] - Internal Project</option>
              <option value="[PROPRIETARY]">[PROPRIETARY] - Proprietary</option>
              <option value="[TRADE_SECRET]">[TRADE_SECRET] - Trade Secret</option>
              <option value="[SALARY_INFO]">[SALARY_INFO] - Salary Info</option>
              <option value="[HR_DOCUMENT]">[HR_DOCUMENT] - HR Document</option>
              <option value="[LEGAL_PRIVILEGE]">[LEGAL_PRIVILEGE] - Legal Privilege</option>
              <option value="[COMPETITOR]">[COMPETITOR] - Competitor Name</option>
            </optgroup>
            <optgroup label="🛡️ Additional">
              <option value="[UUID]">[UUID] - UUID</option>
              <option value="[LICENSE_KEY]">[LICENSE_KEY] - License Key</option>
              <option value="[SESSION_TOKEN]">[SESSION_TOKEN] - Session Token</option>
              <option value="[CSRF_TOKEN]">[CSRF_TOKEN] - CSRF Token</option>
            </optgroup>
          </select>
          
          {#if newReplacementMode === 'custom'}
            <input 
              bind:value={newFilter.replacement} 
              placeholder="e.g., [FILTERED], [REMOVED], ***REDACTED***"
              class="w-full border rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-green-500" />
          {:else}
            <div class="text-sm text-gray-600 p-2 bg-gray-50 rounded border">
              Will replace matches with: <code class="font-mono font-bold text-green-700">{newReplacementMode}</code>
            </div>
          {/if}
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

<!-- Edit Filter Modal -->
{#if showEditModal}
  <div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
    <div class="bg-white rounded-lg p-6 max-w-md w-full">
      <h2 class="text-2xl font-bold mb-4">Edit Filter #{editingFilter?.id}</h2>
      
      <div class="space-y-4">
        <div>
          <label class="block text-sm font-medium mb-1">Pattern</label>
          <input bind:value={editFilter.pattern} class="w-full border rounded px-3 py-2" />
        </div>

        <div>
          <label class="block text-sm font-medium mb-1">Replacement Template</label>
          <select 
            bind:value={editReplacementMode} 
            on:change={() => { 
              if (editReplacementMode !== 'custom') {
                editFilter.replacement = editReplacementMode;
              }
            }}
            class="w-full border rounded px-3 py-2 mb-2 focus:outline-none focus:ring-2 focus:ring-purple-500">
            <option value="custom">✏️ Custom (type your own)</option>
            <optgroup label="🆔 PII - Personal Identifiable Information">
              <option value="[EMAIL]">[EMAIL] - Email Address</option>
              <option value="[PHONE]">[PHONE] - Phone Number</option>
              <option value="[SSN]">[SSN] - Social Security Number</option>
              <option value="[TAX_ID]">[TAX_ID] - Tax ID</option>
              <option value="[PASSPORT]">[PASSPORT] - Passport Number</option>
              <option value="[DRIVER_LICENSE]">[DRIVER_LICENSE] - Driver License</option>
              <option value="[NATIONAL_ID]">[NATIONAL_ID] - National ID</option>
              <option value="[MRN]">[MRN] - Medical Record Number</option>
            </optgroup>
            <optgroup label="💳 Financial Data">
              <option value="[CREDIT_CARD]">[CREDIT_CARD] - Credit Card</option>
              <option value="[CVV]">[CVV] - CVV/CVC Code</option>
              <option value="[IBAN]">[IBAN] - IBAN</option>
              <option value="[BIC]">[BIC] - BIC/SWIFT</option>
              <option value="[BANK_ACCOUNT]">[BANK_ACCOUNT] - Bank Account</option>
              <option value="[ROUTING_NUMBER]">[ROUTING_NUMBER] - Routing Number</option>
              <option value="[CRYPTO_ADDRESS]">[CRYPTO_ADDRESS] - Crypto Address</option>
            </optgroup>
            <optgroup label="🔐 Security & Credentials">
              <option value="[***API_KEY***]">[***API_KEY***] - API Key</option>
              <option value="[***API_SECRET***]">[***API_SECRET***] - API Secret</option>
              <option value="[***AWS_KEY***]">[***AWS_KEY***] - AWS Access Key</option>
              <option value="[***AWS_SECRET***]">[***AWS_SECRET***] - AWS Secret</option>
              <option value="[***GOOGLE_API_KEY***]">[***GOOGLE_API_KEY***] - Google API Key</option>
              <option value="[***GITHUB_TOKEN***]">[***GITHUB_TOKEN***] - GitHub Token</option>
              <option value="[***GITLAB_TOKEN***]">[***GITLAB_TOKEN***] - GitLab Token</option>
              <option value="[***JWT_TOKEN***]">[***JWT_TOKEN***] - JWT Token</option>
              <option value="[***SSH_PRIVATE_KEY***]">[***SSH_PRIVATE_KEY***] - SSH Private Key</option>
              <option value="[***BEARER_TOKEN***]">[***BEARER_TOKEN***] - Bearer Token</option>
              <option value="[***ACCESS_TOKEN***]">[***ACCESS_TOKEN***] - Access Token</option>
              <option value="[***PASSWORD***]">[***PASSWORD***] - Password</option>
              <option value="[***SLACK_TOKEN***]">[***SLACK_TOKEN***] - Slack Token</option>
              <option value="[***STRIPE_KEY***]">[***STRIPE_KEY***] - Stripe Key</option>
              <option value="[***TWILIO_SID***]">[***TWILIO_SID***] - Twilio SID</option>
              <option value="[***SENDGRID_KEY***]">[***SENDGRID_KEY***] - SendGrid Key</option>
            </optgroup>
            <optgroup label="🗄️ Technical Secrets">
              <option value="[***DB_CONNECTION***]">[***DB_CONNECTION***] - DB Connection</option>
              <option value="[***DB_CREDENTIALS***]">[***DB_CREDENTIALS***] - DB Credentials</option>
              <option value="[***DB_PASSWORD***]">[***DB_PASSWORD***] - DB Password</option>
              <option value="[INTERNAL_IP]">[INTERNAL_IP] - Internal IP</option>
              <option value="[INTERNAL_HOST]">[INTERNAL_HOST] - Internal Hostname</option>
              <option value="[LOCALHOST]">[LOCALHOST] - Localhost</option>
              <option value="[***SECRET_KEY***]">[***SECRET_KEY***] - Secret Key</option>
              <option value="[***ENCRYPTION_KEY***]">[***ENCRYPTION_KEY***] - Encryption Key</option>
              <option value="[***DOCKER_LOGIN***]">[***DOCKER_LOGIN***] - Docker Login</option>
            </optgroup>
            <optgroup label="🔒 Confidential">
              <option value="[CONFIDENTIAL]">[CONFIDENTIAL] - Confidential</option>
              <option value="[REDACTED]">[REDACTED] - Redacted</option>
              <option value="[CLASSIFIED]">[CLASSIFIED] - Classified</option>
              <option value="[INTERNAL_PROJECT]">[INTERNAL_PROJECT] - Internal Project</option>
              <option value="[PROPRIETARY]">[PROPRIETARY] - Proprietary</option>
              <option value="[TRADE_SECRET]">[TRADE_SECRET] - Trade Secret</option>
              <option value="[SALARY_INFO]">[SALARY_INFO] - Salary Info</option>
              <option value="[HR_DOCUMENT]">[HR_DOCUMENT] - HR Document</option>
              <option value="[LEGAL_PRIVILEGE]">[LEGAL_PRIVILEGE] - Legal Privilege</option>
              <option value="[COMPETITOR]">[COMPETITOR] - Competitor Name</option>
            </optgroup>
            <optgroup label="🛡️ Additional">
              <option value="[UUID]">[UUID] - UUID</option>
              <option value="[LICENSE_KEY]">[LICENSE_KEY] - License Key</option>
              <option value="[SESSION_TOKEN]">[SESSION_TOKEN] - Session Token</option>
              <option value="[CSRF_TOKEN]">[CSRF_TOKEN] - CSRF Token</option>
            </optgroup>
          </select>
          
          {#if editReplacementMode === 'custom'}
            <input 
              bind:value={editFilter.replacement} 
              placeholder="e.g., [FILTERED], [REMOVED], ***REDACTED***"
              class="w-full border rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-purple-500" />
          {:else}
            <div class="text-sm text-gray-600 p-2 bg-gray-50 rounded border">
              Will replace matches with: <code class="font-mono font-bold text-purple-700">{editReplacementMode}</code>
            </div>
          {/if}
        </div>

        <div>
          <label class="block text-sm font-medium mb-1">Type</label>
          <select bind:value={editFilter.filter_type} class="w-full border rounded px-3 py-2">
            <option value="word">Word</option>
            <option value="phrase">Phrase</option>
            <option value="regex">Regex</option>
          </select>
        </div>

        <div>
          <label class="block text-sm font-medium mb-1">Priority</label>
          <input type="number" bind:value={editFilter.priority} class="w-full border rounded px-3 py-2" />
        </div>

        <div>
          <label class="block text-sm font-medium mb-1">Description</label>
          <input bind:value={editFilter.description} class="w-full border rounded px-3 py-2" />
        </div>

        <div class="flex items-center">
          <input type="checkbox" bind:checked={editFilter.case_sensitive} id="edit-case" class="mr-2" />
          <label for="edit-case" class="text-sm">Case Sensitive</label>
        </div>

        <div class="flex items-center">
          <input type="checkbox" bind:checked={editFilter.enabled} id="edit-enabled" class="mr-2" />
          <label for="edit-enabled" class="text-sm">Enabled</label>
        </div>
      </div>

      <div class="mt-6 flex space-x-2">
        <button on:click={updateFilter} class="flex-1 bg-purple-600 text-white px-4 py-2 rounded hover:bg-purple-700">
          Update
        </button>
        <button on:click={() => { showEditModal = false; editingFilter = null; }} class="flex-1 bg-gray-300 px-4 py-2 rounded hover:bg-gray-400">
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
