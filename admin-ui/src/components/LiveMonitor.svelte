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
  let connectionStatus = {
    openwebui: { status: 'unknown', last_seen: null, total_requests: 0, failed_requests: 0 },
    other: { status: 'unknown', last_seen: null, total_requests: 0, failed_requests: 0 }
  };
  let isLoading = true;
  let autoRefresh = true;
  let refreshInterval;
  
  async function fetchLogs() {
    try {
      const API_BASE = window.location.origin.replace(':3005', ':8080');
      const response = await fetch(`${API_BASE}/admin/requests?limit=50`, {
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
    // Reset
    connectionStatus = {
      openwebui: { status: 'unknown', last_seen: null, total_requests: 0, failed_requests: 0 },
      other: { status: 'unknown', last_seen: null, total_requests: 0, failed_requests: 0 }
    };
    
    const now = new Date();
    const fiveMinutesAgo = new Date(now.getTime() - 5 * 60 * 1000);
    
    logs.forEach(log => {
      const logTime = new Date(log.created_at);
      const isRecent = logTime > fiveMinutesAgo;
      
      // Identify OpenWebUI by IP (Docker network) or User-Agent
      const isOpenWebUI = 
        log.ip_address === '172.18.0.2' || 
        (log.user_agent && log.user_agent.includes('Python')) ||
        (log.user_agent && log.user_agent.includes('aiohttp'));
      
      if (isOpenWebUI) {
        connectionStatus.openwebui.total_requests++;
        if (log.status_code >= 400) {
          connectionStatus.openwebui.failed_requests++;
        }
        if (isRecent) {
          connectionStatus.openwebui.last_seen = logTime;
          if (log.status_code === 200) {
            connectionStatus.openwebui.status = 'connected';
          } else if (log.status_code === 401) {
            connectionStatus.openwebui.status = 'auth_failed';
          } else {
            connectionStatus.openwebui.status = 'error';
          }
        }
      } else if (log.ip_address && !log.ip_address.startsWith('::1') && !log.ip_address.startsWith('127.')) {
        connectionStatus.other.total_requests++;
        if (log.status_code >= 400) {
          connectionStatus.other.failed_requests++;
        }
        if (isRecent) {
          connectionStatus.other.last_seen = logTime;
          if (log.status_code === 200) {
            connectionStatus.other.status = 'connected';
          } else {
            connectionStatus.other.status = 'error';
          }
        }
      }
    });
    
    // If no recent activity, set to disconnected
    if (!connectionStatus.openwebui.last_seen) {
      connectionStatus.openwebui.status = 'disconnected';
    }
    if (!connectionStatus.other.last_seen) {
      connectionStatus.other.status = 'disconnected';
    }
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
      case 'connected': return '✅';
      case 'auth_failed': return '🔐';
      case 'error': return '❌';
      case 'disconnected': return '⚫';
      default: return '❓';
    }
  }
  
  function getStatusText(status) {
    switch(status) {
      case 'connected': return 'Connected';
      case 'auth_failed': return 'Authentication Failed';
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
        class="px-4 py-2 rounded {autoRefresh ? 'bg-green-600' : 'bg-gray-600'} text-white hover:opacity-80">
        {autoRefresh ? '🔄 Auto-Refresh ON' : '⏸ Auto-Refresh OFF'}
      </button>
      <button 
        on:click={fetchLogs}
        class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700">
        🔄 Refresh Now
      </button>
    </div>
  </div>

  {#if isLoading}
    <div class="text-center py-12">
      <div class="text-4xl mb-4">⏳</div>
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
        <div class="grid grid-cols-2 gap-6">
          <!-- OpenWebUI Status -->
          <div class="border rounded-lg p-4 {connectionStatus.openwebui.status === 'connected' ? 'border-green-500 bg-green-50' : connectionStatus.openwebui.status === 'auth_failed' ? 'border-yellow-500 bg-yellow-50' : 'border-gray-300'}">
            <div class="flex items-center justify-between mb-3">
              <div class="flex items-center gap-2">
                <span class="text-2xl">{getStatusIcon(connectionStatus.openwebui.status)}</span>
                <span class="text-lg font-semibold">OpenWebUI</span>
              </div>
              <span class="{getStatusColor(connectionStatus.openwebui.status)} font-semibold">
                {getStatusText(connectionStatus.openwebui.status)}
              </span>
            </div>
            <div class="space-y-2 text-sm">
              <div class="flex justify-between">
                <span class="text-gray-600">Last Seen:</span>
                <span class="font-medium">{formatTime(connectionStatus.openwebui.last_seen)}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-gray-600">Total Requests:</span>
                <span class="font-medium">{connectionStatus.openwebui.total_requests}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-gray-600">Failed Requests:</span>
                <span class="font-medium text-red-600">{connectionStatus.openwebui.failed_requests}</span>
              </div>
              {#if connectionStatus.openwebui.status === 'auth_failed'}
                <div class="mt-3 p-2 bg-yellow-100 border border-yellow-300 rounded text-xs">
                  ⚠️ API Key authentication failing. Check OpenWebUI configuration.
                </div>
              {/if}
            </div>
          </div>

          <!-- Other Clients Status -->
          <div class="border rounded-lg p-4 {connectionStatus.other.status === 'connected' ? 'border-green-500 bg-green-50' : 'border-gray-300'}">
            <div class="flex items-center justify-between mb-3">
              <div class="flex items-center gap-2">
                <span class="text-2xl">{getStatusIcon(connectionStatus.other.status)}</span>
                <span class="text-lg font-semibold">Other Clients</span>
              </div>
              <span class="{getStatusColor(connectionStatus.other.status)} font-semibold">
                {getStatusText(connectionStatus.other.status)}
              </span>
            </div>
            <div class="space-y-2 text-sm">
              <div class="flex justify-between">
                <span class="text-gray-600">Last Seen:</span>
                <span class="font-medium">{formatTime(connectionStatus.other.last_seen)}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-gray-600">Total Requests:</span>
                <span class="font-medium">{connectionStatus.other.total_requests}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-gray-600">Failed Requests:</span>
                <span class="font-medium text-red-600">{connectionStatus.other.failed_requests}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Recent Requests Table -->
    <div class="bg-white rounded-lg shadow">
      <div class="px-6 py-4 border-b border-gray-200">
        <h2 class="text-xl font-semibold text-gray-800">Recent Requests (Last 50)</h2>
      </div>
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Time</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Method</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Path</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Client</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Duration</th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            {#each logs as log (log.request_id)}
              <tr class="hover:bg-gray-50">
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-600">
                  {formatTime(log.created_at)}
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <span class="px-2 py-1 text-xs font-semibold rounded {getStatusCodeClass(log.status_code)}">
                    {log.status_code}
                  </span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  {log.method}
                </td>
                <td class="px-6 py-4 text-sm text-gray-900 max-w-xs truncate">
                  {log.endpoint}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm">
                  {#if log.ip_address === '172.18.0.2' || (log.user_agent && log.user_agent.includes('Python'))}
                    <span class="px-2 py-1 bg-blue-100 text-blue-800 rounded text-xs font-medium">
                      OpenWebUI
                    </span>
                  {:else if log.ip_address && !log.ip_address.startsWith('::1')}
                    <span class="px-2 py-1 bg-gray-100 text-gray-800 rounded text-xs font-medium">
                      {log.ip_address}
                    </span>
                  {:else}
                    <span class="px-2 py-1 bg-gray-100 text-gray-600 rounded text-xs">
                      Internal
                    </span>
                  {/if}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-600">
                  {log.duration_ms ? `${log.duration_ms}ms` : '-'}
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </div>
  {/if}
</div>

<style>
  /* Add any custom styles here */
</style>
