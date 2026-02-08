<script>
  import { onMount, onDestroy } from 'svelte';
  import { apiKey } from '../lib/stores.js';
  
  let logs = [];
  let clients = [];
  let recentActivity = {
    last_minute: 0,
    last_hour: 0,
    filtered_requests: 0,
    error_rate: 0
  };
  let connectionStatus = {};  // Dynamic: keyed by client_name or 'unknown'
  let isLoading = true;
  let autoRefresh = true;
  let refreshInterval;
  
  // Search & filter state
  let searchQuery = '';
  let hideHealthChecks = true;
  let hideAdminRequests = true;
  let showOnlyErrors = false;
  let showOnlyAPI = false;
  
  // Request detail modal state
  let showDetailModal = false;
  let selectedRequest = null;
  let loadingDetails = false;
  
  // Filtered logs (reactive)
  $: filteredLogs = logs.filter(log => {
    // Quick filters
    if (hideHealthChecks && (log.endpoint === '/health' || log.endpoint === '/health/ready' || log.endpoint === '/health/detailed' || log.endpoint === '/metrics')) {
      return false;
    }
    if (hideAdminRequests && log.endpoint && log.endpoint.startsWith('/admin')) {
      return false;
    }
    if (showOnlyErrors && log.status_code < 400) {
      return false;
    }
    if (showOnlyAPI && !(log.endpoint && log.endpoint.startsWith('/v1'))) {
      return false;
    }
    
    // Text search
    if (searchQuery.trim()) {
      const q = searchQuery.toLowerCase().trim();
      const searchFields = [
        log.endpoint,
        log.method,
        log.client_name,
        log.api_key_name,
        log.model,
        log.provider,
        log.ip_address,
        log.auth_type,
        log.error_message,
        String(log.status_code),
      ].filter(Boolean).map(f => f.toLowerCase());
      
      return searchFields.some(f => f.includes(q));
    }
    
    return true;
  });
  
  async function fetchLogs() {
    try {
      const API_BASE = window.location.origin.replace(':3005', ':8080');
      const response = await fetch(`${API_BASE}/admin/requests?limit=100`, {
        headers: {
          'X-Admin-API-Key': $apiKey
        }
      });
      
      if (response.ok) {
        const data = await response.json();
        logs = data.logs || [];
        analyzeConnections();
        analyzeActivity();
      }
    } catch (error) {
      console.error('Error fetching logs:', error);
    } finally {
      isLoading = false;
    }
  }
  
  async function fetchClients() {
    try {
      const API_BASE = window.location.origin.replace(':3005', ':8080');
      const response = await fetch(`${API_BASE}/admin/clients`, {
        headers: {
          'X-Admin-API-Key': $apiKey
        }
      });
      
      if (response.ok) {
        const data = await response.json();
        clients = data.clients || [];
      }
    } catch (error) {
      console.error('Error fetching clients:', error);
    }
  }
  
  function analyzeConnections() {
    // Build dynamic connection status per client
    const statusMap = {};
    
    const now = new Date();
    const fiveMinutesAgo = new Date(now.getTime() - 5 * 60 * 1000);
    
    logs.forEach(log => {
      const logTime = new Date(log.created_at);
      const isRecent = logTime > fiveMinutesAgo;
      
      // Determine client key: use client_name if available, otherwise derive from context
      let clientKey = 'Unknown';
      if (log.client_name) {
        clientKey = log.client_name;
      } else if (log.api_key_name) {
        clientKey = log.api_key_name;
      } else if (log.auth_type === 'oauth') {
        clientKey = 'OAuth Client';
      } else if (log.auth_type === 'admin') {
        clientKey = 'Admin';
      } else if (log.ip_address && !log.ip_address.startsWith('::1') && !log.ip_address.startsWith('127.')) {
        clientKey = log.ip_address;
      }
      
      // Skip internal / noise for connection status
      if (clientKey === 'Unknown' && (log.endpoint === '/health' || (log.endpoint && log.endpoint.startsWith('/admin')))) {
        return;
      }
      
      // Initialize if not yet tracked
      if (!statusMap[clientKey]) {
        statusMap[clientKey] = { status: 'unknown', last_seen: null, total_requests: 0, failed_requests: 0 };
      }
      
      statusMap[clientKey].total_requests++;
      if (log.status_code >= 400) {
        statusMap[clientKey].failed_requests++;
      }
      
      if (isRecent) {
        statusMap[clientKey].last_seen = logTime;
        if (log.status_code >= 200 && log.status_code < 400) {
          statusMap[clientKey].status = 'connected';
        } else if (log.status_code === 401) {
          statusMap[clientKey].status = 'auth_failed';
        } else if (statusMap[clientKey].status !== 'connected') {
          statusMap[clientKey].status = 'error';
        }
      }
    });
    
    // Mark clients without recent activity as disconnected
    Object.keys(statusMap).forEach(key => {
      if (!statusMap[key].last_seen) {
        statusMap[key].status = 'disconnected';
      }
    });
    
    connectionStatus = statusMap;
  }
  
  function analyzeActivity() {
    const now = new Date();
    const oneMinuteAgo = new Date(now.getTime() - 60 * 1000);
    const oneHourAgo = new Date(now.getTime() - 60 * 60 * 1000);
    
    let lastMinute = 0;
    let lastHour = 0;
    let errors = 0;
    
    logs.forEach(log => {
      const logTime = new Date(log.created_at);
      
      if (logTime > oneMinuteAgo) {
        lastMinute++;
      }
      if (logTime > oneHourAgo) {
        lastHour++;
        if (log.status_code >= 400) {
          errors++;
        }
      }
    });
    
    recentActivity = {
      last_minute: lastMinute,
      last_hour: lastHour,
      filtered_requests: 0, // TODO: Get from filter_matches
      error_rate: lastHour > 0 ? (errors / lastHour * 100).toFixed(1) : 0
    };
  }
  
  function getStatusColor(status) {
    switch(status) {
      case 'connected': return 'text-green-500';
      case 'auth_failed': return 'text-yellow-500';
      case 'error': return 'text-red-500';
      case 'disconnected': return 'text-gray-500';
      default: return 'text-gray-400';
    }
  }
  
  function getStatusIcon(status) {
    switch(status) {
      case 'connected': return '●';
      case 'auth_failed': return '▲';
      case 'error': return '✕';
      case 'disconnected': return '○';
      default: return '?';
    }
  }
  
  function getStatusText(status) {
    switch(status) {
      case 'connected': return 'Connected';
      case 'auth_failed': return 'Auth Failed';
      case 'error': return 'Error';
      case 'disconnected': return 'Disconnected';
      default: return 'Unknown';
    }
  }
  
  function formatTime(dateString) {
    if (!dateString) return 'Never';
    const date = new Date(dateString);
    const now = new Date();
    const diff = Math.floor((now - date) / 1000);
    
    if (diff < 60) return `${diff}s ago`;
    if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
    return `${Math.floor(diff / 3600)}h ago`;
  }
  
  function getStatusCodeClass(code) {
    if (code >= 200 && code < 300) return 'text-green-600 bg-green-100';
    if (code >= 400 && code < 500) return 'text-yellow-600 bg-yellow-100';
    if (code >= 500) return 'text-red-600 bg-red-100';
    return 'text-gray-600 bg-gray-100';
  }
  
  function toggleAutoRefresh() {
    autoRefresh = !autoRefresh;
    if (autoRefresh) {
      startAutoRefresh();
    } else {
      stopAutoRefresh();
    }
  }
  
  function startAutoRefresh() {
    refreshInterval = setInterval(() => {
      fetchLogs();
    }, 5000); // Refresh every 5 seconds
  }
  
  function stopAutoRefresh() {
    if (refreshInterval) {
      clearInterval(refreshInterval);
      refreshInterval = null;
    }
  }
  
  async function showRequestDetails(log) {
    showDetailModal = true;
    loadingDetails = true;
    selectedRequest = null;
    
    try {
      const API_BASE = window.location.origin.replace(':3005', ':8080');
      const response = await fetch(`${API_BASE}/admin/requests/${log.id}`, {
        headers: {
          'X-Admin-API-Key': $apiKey
        }
      });
      
      if (response.ok) {
        selectedRequest = await response.json();
      } else {
        console.error('Failed to fetch request details:', response.statusText);
        alert('Failed to load request details');
        showDetailModal = false;
      }
    } catch (error) {
      console.error('Error fetching request details:', error);
      alert('Error loading request details');
      showDetailModal = false;
    } finally {
      loadingDetails = false;
    }
  }
  
  function closeModal() {
    showDetailModal = false;
    selectedRequest = null;
  }
  
  function formatJSON(str) {
    if (!str) return null;
    try {
      const obj = typeof str === 'string' ? JSON.parse(str) : str;
      return JSON.stringify(obj, null, 2);
    } catch (e) {
      return str;
    }
  }
  
  // Extract just the messages content for a readable preview
  function formatRequestBodyPreview(bodyStr) {
    if (!bodyStr) return null;
    try {
      const obj = JSON.parse(bodyStr);
      // For chat completion requests, show a concise summary
      if (obj.model && obj.messages) {
        const lastMsg = obj.messages[obj.messages.length - 1];
        let content = '';
        if (typeof lastMsg.content === 'string') {
          content = lastMsg.content;
        } else if (Array.isArray(lastMsg.content)) {
          content = lastMsg.content.filter(p => p.type === 'text').map(p => p.text).join(' ');
        }
        if (content.length > 200) content = content.substring(0, 200) + '...';
        return `Model: ${obj.model}\nMessages: ${obj.messages.length}\nLast message (${lastMsg.role}): ${content}`;
      }
      return JSON.stringify(obj, null, 2);
    } catch (e) {
      return bodyStr;
    }
  }
  
  onMount(() => {
    fetchLogs();
    fetchClients();
    if (autoRefresh) {
      startAutoRefresh();
    }
  });
  
  onDestroy(() => {
    stopAutoRefresh();
  });
</script>

<div class="p-6">
  <div class="flex justify-between items-center mb-6">
    <h1 class="text-3xl font-bold text-gray-800">Live Monitor</h1>
    <div class="flex items-center gap-4">
      <button 
        on:click={toggleAutoRefresh}
        class="px-4 py-2 rounded text-sm font-medium {autoRefresh ? 'bg-green-600' : 'bg-gray-600'} text-white hover:opacity-80">
        {autoRefresh ? 'Auto-Refresh ON' : 'Auto-Refresh OFF'}
      </button>
      <button 
        on:click={fetchLogs}
        class="px-4 py-2 bg-blue-600 text-white rounded text-sm font-medium hover:bg-blue-700">
        Refresh Now
      </button>
    </div>
  </div>

  {#if isLoading}
    <div class="text-center py-12">
      <div class="text-gray-600">Loading...</div>
    </div>
  {:else}
    <!-- Recent Activity Stats -->
    <div class="grid grid-cols-4 gap-4 mb-6">
      <div class="bg-white p-4 rounded-lg shadow">
        <div class="text-gray-600 text-sm mb-1">Last Minute</div>
        <div class="text-3xl font-bold text-blue-600">{recentActivity.last_minute}</div>
        <div class="text-xs text-gray-500 mt-1">requests</div>
      </div>
      <div class="bg-white p-4 rounded-lg shadow">
        <div class="text-gray-600 text-sm mb-1">Last Hour</div>
        <div class="text-3xl font-bold text-green-600">{recentActivity.last_hour}</div>
        <div class="text-xs text-gray-500 mt-1">requests</div>
      </div>
      <div class="bg-white p-4 rounded-lg shadow">
        <div class="text-gray-600 text-sm mb-1">Error Rate</div>
        <div class="text-3xl font-bold text-red-600">{recentActivity.error_rate}%</div>
        <div class="text-xs text-gray-500 mt-1">last hour</div>
      </div>
      <div class="bg-white p-4 rounded-lg shadow">
        <div class="text-gray-600 text-sm mb-1">Filtered</div>
        <div class="text-3xl font-bold text-purple-600">{recentActivity.filtered_requests}</div>
        <div class="text-xs text-gray-500 mt-1">requests</div>
      </div>
    </div>

    <!-- Connection Status -->
    <div class="bg-white rounded-lg shadow mb-6">
      <div class="px-6 py-4 border-b border-gray-200">
        <h2 class="text-xl font-semibold text-gray-800">Client Connection Status</h2>
      </div>
      <div class="p-6">
        {#if Object.keys(connectionStatus).length === 0}
          <div class="text-center text-gray-500 py-4">No client activity detected yet.</div>
        {:else}
          <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {#each Object.entries(connectionStatus) as [clientName, status]}
              <div class="border rounded-lg p-4 {status.status === 'connected' ? 'border-green-500 bg-green-50' : status.status === 'auth_failed' ? 'border-yellow-500 bg-yellow-50' : status.status === 'error' ? 'border-red-500 bg-red-50' : 'border-gray-300'}">
                <div class="flex items-center justify-between mb-3">
                  <div class="flex items-center gap-2">
                    <span class="text-lg font-bold {getStatusColor(status.status)}">{getStatusIcon(status.status)}</span>
                    <span class="text-lg font-semibold truncate" title={clientName}>{clientName}</span>
                  </div>
                  <span class="{getStatusColor(status.status)} font-semibold text-sm">
                    {getStatusText(status.status)}
                  </span>
                </div>
                <div class="space-y-2 text-sm">
                  <div class="flex justify-between">
                    <span class="text-gray-600">Last Seen:</span>
                    <span class="font-medium">{formatTime(status.last_seen)}</span>
                  </div>
                  <div class="flex justify-between">
                    <span class="text-gray-600">Total Requests:</span>
                    <span class="font-medium">{status.total_requests}</span>
                  </div>
                  <div class="flex justify-between">
                    <span class="text-gray-600">Failed Requests:</span>
                    <span class="font-medium text-red-600">{status.failed_requests}</span>
                  </div>
                  {#if status.status === 'auth_failed'}
                    <div class="mt-3 p-2 bg-yellow-100 border border-yellow-300 rounded text-xs">
                      API Key authentication failing. Check client configuration.
                    </div>
                  {/if}
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>
    </div>

    <!-- Recent Requests Table -->
    <div class="bg-white rounded-lg shadow">
      <div class="px-6 py-4 border-b border-gray-200">
        <div class="flex justify-between items-center mb-3">
          <h2 class="text-xl font-semibold text-gray-800">Recent Requests</h2>
          <span class="text-sm text-gray-500">{filteredLogs.length} of {logs.length} shown</span>
        </div>
        
        <!-- Search Bar -->
        <div class="mb-3">
          <input 
            type="text" 
            bind:value={searchQuery}
            placeholder="Search by path, client, model, status code, IP..." 
            class="w-full px-4 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />
        </div>
        
        <!-- Quick Filters -->
        <div class="flex flex-wrap gap-2">
          <button 
            on:click={() => hideHealthChecks = !hideHealthChecks}
            class="px-3 py-1 rounded-full text-xs font-medium border transition-colors
              {hideHealthChecks ? 'bg-blue-100 text-blue-800 border-blue-300' : 'bg-white text-gray-600 border-gray-300'}">
            {hideHealthChecks ? 'Hiding' : 'Showing'} health checks
          </button>
          <button 
            on:click={() => hideAdminRequests = !hideAdminRequests}
            class="px-3 py-1 rounded-full text-xs font-medium border transition-colors
              {hideAdminRequests ? 'bg-blue-100 text-blue-800 border-blue-300' : 'bg-white text-gray-600 border-gray-300'}">
            {hideAdminRequests ? 'Hiding' : 'Showing'} admin requests
          </button>
          <button 
            on:click={() => { showOnlyAPI = !showOnlyAPI; if (showOnlyAPI) showOnlyErrors = false; }}
            class="px-3 py-1 rounded-full text-xs font-medium border transition-colors
              {showOnlyAPI ? 'bg-green-100 text-green-800 border-green-300' : 'bg-white text-gray-600 border-gray-300'}">
            {showOnlyAPI ? 'API only (/v1)' : 'All paths'}
          </button>
          <button 
            on:click={() => { showOnlyErrors = !showOnlyErrors; if (showOnlyErrors) showOnlyAPI = false; }}
            class="px-3 py-1 rounded-full text-xs font-medium border transition-colors
              {showOnlyErrors ? 'bg-red-100 text-red-800 border-red-300' : 'bg-white text-gray-600 border-gray-300'}">
            {showOnlyErrors ? 'Errors only' : 'All statuses'}
          </button>
        </div>
      </div>
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Time</th>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Method</th>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Path</th>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Client</th>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Model</th>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Duration</th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            {#each filteredLogs as log (log.id)}
              <tr class="hover:bg-blue-50 cursor-pointer transition-colors" on:click={() => showRequestDetails(log)}>
                <td class="px-4 py-3 whitespace-nowrap text-sm text-gray-600">
                  {formatTime(log.created_at)}
                </td>
                <td class="px-4 py-3 whitespace-nowrap">
                  <span class="px-2 py-1 text-xs font-semibold rounded {getStatusCodeClass(log.status_code)}">
                    {log.status_code}
                  </span>
                </td>
                <td class="px-4 py-3 whitespace-nowrap text-sm text-gray-900 font-mono">
                  {log.method}
                </td>
                <td class="px-4 py-3 text-sm text-gray-900 max-w-xs truncate font-mono" title={log.endpoint}>
                  {log.endpoint}
                </td>
                <td class="px-4 py-3 whitespace-nowrap text-sm">
                  {#if log.client_name}
                    <span class="px-2 py-1 bg-blue-100 text-blue-800 rounded text-xs font-medium" title="Client ID: {log.client_id}">
                      {log.client_name}
                    </span>
                  {:else if log.api_key_name}
                    <span class="px-2 py-1 bg-purple-100 text-purple-800 rounded text-xs font-medium">
                      {log.api_key_name}
                    </span>
                  {:else if log.auth_type === 'admin'}
                    <span class="px-2 py-1 bg-orange-100 text-orange-800 rounded text-xs font-medium">
                      Admin
                    </span>
                  {:else if log.ip_address && !log.ip_address.startsWith('127.')}
                    <span class="px-2 py-1 bg-gray-100 text-gray-800 rounded text-xs font-medium">
                      {log.ip_address}
                    </span>
                  {:else}
                    <span class="px-2 py-1 bg-gray-100 text-gray-500 rounded text-xs">
                      -
                    </span>
                  {/if}
                </td>
                <td class="px-4 py-3 whitespace-nowrap text-sm">
                  {#if log.model}
                    <span class="px-2 py-1 bg-indigo-100 text-indigo-800 rounded text-xs font-medium">
                      {log.model}
                    </span>
                  {:else}
                    <span class="text-gray-400">-</span>
                  {/if}
                </td>
                <td class="px-4 py-3 whitespace-nowrap text-sm text-gray-600 font-mono">
                  {log.duration_ms ? `${log.duration_ms}ms` : '-'}
                </td>
              </tr>
            {:else}
              <tr>
                <td colspan="7" class="px-4 py-8 text-center text-gray-500">
                  {#if searchQuery || showOnlyErrors || showOnlyAPI}
                    No requests match your filters.
                  {:else}
                    No requests logged yet.
                  {/if}
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </div>
  {/if}
  
  <!-- Request Detail Modal -->
  {#if showDetailModal}
    <div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50" on:click={closeModal}>
      <div class="bg-white rounded-lg shadow-xl max-w-6xl w-full max-h-[90vh] overflow-y-auto" on:click|stopPropagation>
        <div class="sticky top-0 bg-white border-b border-gray-200 px-6 py-4 flex justify-between items-center z-10">
          <h2 class="text-2xl font-bold text-gray-800">Request Details</h2>
          <button on:click={closeModal} class="text-gray-500 hover:text-gray-700 text-2xl font-bold">&times;</button>
        </div>
        
        {#if loadingDetails}
          <div class="p-12 text-center">
            <div class="text-gray-600">Loading details...</div>
          </div>
        {:else if selectedRequest}
          <div class="p-6 space-y-6">
            <!-- Overview Section -->
            <div class="grid grid-cols-2 gap-6">
              <div class="space-y-3">
                <div>
                  <div class="text-sm font-medium text-gray-500">Request ID</div>
                  <div class="text-sm font-mono text-gray-900">{selectedRequest.request_id}</div>
                </div>
                <div>
                  <div class="text-sm font-medium text-gray-500">Timestamp</div>
                  <div class="text-sm text-gray-900">{new Date(selectedRequest.created_at).toLocaleString()}</div>
                </div>
                <div>
                  <div class="text-sm font-medium text-gray-500">Method & Path</div>
                  <div class="text-sm text-gray-900 font-mono">{selectedRequest.method} {selectedRequest.path}</div>
                </div>
                <div>
                  <div class="text-sm font-medium text-gray-500">Status Code</div>
                  <div>
                    <span class="px-2 py-1 text-xs font-semibold rounded {getStatusCodeClass(selectedRequest.status_code)}">
                      {selectedRequest.status_code}
                    </span>
                  </div>
                </div>
              </div>
              
              <div class="space-y-3">
                <div>
                  <div class="text-sm font-medium text-gray-500">Duration</div>
                  <div class="text-sm text-gray-900">{selectedRequest.duration_ms}ms</div>
                </div>
                <div>
                  <div class="text-sm font-medium text-gray-500">IP Address</div>
                  <div class="text-sm text-gray-900 font-mono">{selectedRequest.ip_address || 'N/A'}</div>
                </div>
                <div>
                  <div class="text-sm font-medium text-gray-500">Auth Type</div>
                  <div class="text-sm text-gray-900">{selectedRequest.auth_type || 'none'}</div>
                </div>
                {#if selectedRequest.client_name}
                  <div>
                    <div class="text-sm font-medium text-gray-500">Client</div>
                    <div class="text-sm text-gray-900">{selectedRequest.client_name}</div>
                  </div>
                {/if}
                {#if selectedRequest.api_key_name}
                  <div>
                    <div class="text-sm font-medium text-gray-500">API Key Name</div>
                    <div class="text-sm text-gray-900">{selectedRequest.api_key_name}</div>
                  </div>
                {/if}
                {#if selectedRequest.model}
                  <div>
                    <div class="text-sm font-medium text-gray-500">Model</div>
                    <div class="text-sm text-gray-900">{selectedRequest.model}</div>
                  </div>
                {/if}
                {#if selectedRequest.provider}
                  <div>
                    <div class="text-sm font-medium text-gray-500">Provider</div>
                    <div class="text-sm text-gray-900">{selectedRequest.provider}</div>
                  </div>
                {/if}
              </div>
            </div>
            
            {#if selectedRequest.error_message}
              <div class="bg-red-50 border border-red-200 rounded-lg p-4">
                <div class="text-sm font-medium text-red-800 mb-2">Error Message</div>
                <div class="text-sm text-red-700 font-mono whitespace-pre-wrap">{selectedRequest.error_message}</div>
              </div>
            {/if}
            
            {#if selectedRequest.was_filtered}
              <div class="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
                <div class="text-sm font-medium text-yellow-800 mb-2">Content Filtered</div>
                <div class="text-sm text-yellow-700">{selectedRequest.filter_reason || 'No reason provided'}</div>
              </div>
            {/if}
            
            {#if selectedRequest.total_tokens}
              <div class="bg-blue-50 border border-blue-200 rounded-lg p-4">
                <h3 class="text-sm font-medium text-blue-800 mb-2">Token Usage</h3>
                <div class="grid grid-cols-3 gap-4 text-sm">
                  <div>
                    <div class="text-blue-600">Prompt Tokens</div>
                    <div class="font-semibold">{selectedRequest.prompt_tokens || 0}</div>
                  </div>
                  <div>
                    <div class="text-blue-600">Completion Tokens</div>
                    <div class="font-semibold">{selectedRequest.completion_tokens || 0}</div>
                  </div>
                  <div>
                    <div class="text-blue-600">Total Tokens</div>
                    <div class="font-semibold">{selectedRequest.total_tokens}</div>
                  </div>
                </div>
                {#if selectedRequest.cost_usd}
                  <div class="mt-3 pt-3 border-t border-blue-200">
                    <div class="text-blue-600">Estimated Cost</div>
                    <div class="font-semibold text-lg">${selectedRequest.cost_usd.toFixed(4)}</div>
                  </div>
                {/if}
              </div>
            {/if}
            
            <!-- Request Body (with readable preview for chat completions) -->
            {#if selectedRequest.request_body}
              <div class="border rounded-lg overflow-hidden">
                <div class="bg-blue-50 px-4 py-3 border-b flex justify-between items-center">
                  <h3 class="text-lg font-semibold text-gray-800">Request Body</h3>
                </div>
                <div class="p-4">
                  <pre class="text-xs font-mono bg-gray-50 p-3 rounded overflow-x-auto max-h-96 whitespace-pre-wrap">{formatJSON(selectedRequest.request_body)}</pre>
                </div>
              </div>
            {:else}
              <div class="border rounded-lg overflow-hidden">
                <div class="bg-gray-50 px-4 py-3 border-b">
                  <h3 class="text-lg font-semibold text-gray-800">Request Body</h3>
                </div>
                <div class="p-4 text-sm text-gray-500 italic">
                  No request body captured. Body capture may be disabled in Settings, or this request had no body.
                </div>
              </div>
            {/if}
            
            <!-- Request Headers -->
            {#if selectedRequest.request_headers && Object.keys(selectedRequest.request_headers).length > 0}
              <div class="border rounded-lg overflow-hidden">
                <div class="bg-gray-50 px-4 py-3 border-b">
                  <h3 class="text-lg font-semibold text-gray-800">Request Headers</h3>
                </div>
                <div class="p-4">
                  <pre class="text-xs font-mono bg-gray-50 p-3 rounded overflow-x-auto">{formatJSON(selectedRequest.request_headers)}</pre>
                </div>
              </div>
            {/if}
            
            <!-- Response Body -->
            {#if selectedRequest.response_body}
              <div class="border rounded-lg overflow-hidden">
                <div class="bg-green-50 px-4 py-3 border-b flex justify-between items-center">
                  <h3 class="text-lg font-semibold text-gray-800">Response Body</h3>
                  {#if selectedRequest.response_size_bytes}
                    <span class="text-sm text-gray-600">{(selectedRequest.response_size_bytes / 1024).toFixed(2)} KB</span>
                  {/if}
                </div>
                <div class="p-4">
                  <pre class="text-xs font-mono bg-gray-50 p-3 rounded overflow-x-auto max-h-96 whitespace-pre-wrap">{formatJSON(selectedRequest.response_body)}</pre>
                </div>
              </div>
            {/if}
            
            <!-- Response Headers -->
            {#if selectedRequest.response_headers && Object.keys(selectedRequest.response_headers).length > 0}
              <div class="border rounded-lg overflow-hidden">
                <div class="bg-gray-50 px-4 py-3 border-b">
                  <h3 class="text-lg font-semibold text-gray-800">Response Headers</h3>
                </div>
                <div class="p-4">
                  <pre class="text-xs font-mono bg-gray-50 p-3 rounded overflow-x-auto">{formatJSON(selectedRequest.response_headers)}</pre>
                </div>
              </div>
            {/if}
          </div>
        {/if}
      </div>
    </div>
  {/if}
</div>
