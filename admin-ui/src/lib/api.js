// Admin API Client
// PRODUCTION FIX: Use window.location.origin to always use the current domain
// This prevents localhost:8080 issues when deploying from registry
// DEVELOPMENT: Use environment variable or localhost:8080 for local development
const API_BASE_URL = import.meta.env.DEV 
  ? (import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080')
  : window.location.origin;

class AdminAPI {
  constructor(apiKey) {
    this.apiKey = apiKey;
    this.baseURL = API_BASE_URL;
  }

  async request(endpoint, options = {}) {
    const headers = {
      'X-Admin-API-Key': this.apiKey,
      'Content-Type': 'application/json',
      ...options.headers,
    };

    const url = `${this.baseURL}${endpoint}`;
    
    try {
      const response = await fetch(url, {
        ...options,
        headers,
      });

      // Always try to read response body as text first
      const text = await response.text();
      
      // Try to parse JSON if we have content
      let data = null;
      if (text && text.trim()) {
        try {
          data = JSON.parse(text);
        } catch (e) {
          // Not JSON, that's ok for some responses
        }
      }

      if (!response.ok) {
        const message = data?.message || data?.error || `Request failed with status ${response.status}`;
        throw new Error(message);
      }

      return data;
    } catch (error) {
      console.error(`API Error (${endpoint}):`, error);
      throw error;
    }
  }

  // Client Management
  async listClients() {
    return this.request('/admin/clients');
  }

  async getClient(clientId) {
    return this.request(`/admin/clients/${clientId}`);
  }

  async createClient(client) {
    return this.request('/admin/clients', {
      method: 'POST',
      body: JSON.stringify(client),
    });
  }

  async updateClient(clientId, updates) {
    return this.request(`/admin/clients/${clientId}`, {
      method: 'PATCH',
      body: JSON.stringify(updates),
    });
  }

  async deleteClient(clientId) {
    return this.request(`/admin/clients/${clientId}`, {
      method: 'DELETE',
    });
  }

  async resetClientSecret(clientId, newSecret = '') {
    return this.request(`/admin/clients/${clientId}/reset-secret`, {
      method: 'POST',
      body: JSON.stringify({ new_secret: newSecret }),
    });
  }

  // Cache Management
  async getCacheStats() {
    return this.request('/admin/cache/stats');
  }

  async clearCache() {
    return this.request('/admin/cache/clear', {
      method: 'POST',
    });
  }

  async invalidateCacheByModel(model) {
    return this.request(`/admin/cache/invalidate/${model}`, {
      method: 'POST',
    });
  }

  // Dashboard (comprehensive stats in one call)
  async getDashboardData() {
    return this.request('/admin/dashboard');
  }

  // Usage Statistics
  async getUsageStats(params = {}) {
    const query = new URLSearchParams(params).toString();
    const endpoint = query ? `/admin/stats/usage?${query}` : '/admin/stats/usage';
    return this.request(endpoint);
  }

  // Provider Status
  async getProviderStatus() {
    return this.request('/admin/providers/status');
  }

  // Provider Details
  async getProviderDetails() {
    return this.request('/admin/providers');
  }

  // Content Filters
  async listFilters() {
    return this.request('/admin/filters');
  }

  async getFilter(id) {
    return this.request(`/admin/filters/${id}`);
  }

  async createFilter(filter) {
    return this.request('/admin/filters', {
      method: 'POST',
      body: JSON.stringify(filter),
    });
  }

  async updateFilter(id, updates) {
    return this.request(`/admin/filters/${id}`, {
      method: 'PUT',
      body: JSON.stringify(updates),
    });
  }

  async deleteFilter(id) {
    return this.request(`/admin/filters/${id}`, {
      method: 'DELETE',
    });
  }

  async bulkImportFilters(filters) {
    return this.request('/admin/filters/bulk-import', {
      method: 'POST',
      body: JSON.stringify({ filters }),
    });
  }

  async testFilter(id, text) {
    return this.request(`/admin/filters/${id}/test`, {
      method: 'POST',
      body: JSON.stringify({ text }),
    });
  }

  async getFilterStats() {
    return this.request('/admin/filters/stats');
  }

  async getFilterMatches(limit = 100) {
    return this.request(`/admin/filters/matches?limit=${limit}`);
  }

  async refreshFilters() {
    return this.request('/admin/filters/refresh', {
      method: 'POST',
    });
  }

  // Provider Management
  async getProviderConfig(providerId) {
    return this.request(`/admin/providers/${providerId}/config`);
  }

  async testProvider(providerId) {
    return this.request(`/admin/providers/${providerId}/test`, {
      method: 'POST',
    });
  }

  async toggleProvider(providerId, enabled) {
    return this.request(`/admin/providers/${providerId}/toggle`, {
      method: 'PUT',
      body: JSON.stringify({ enabled }),
    });
  }

  // Provider API Key Management
  async listProviderKeys(providerId) {
    return this.request(`/admin/providers/${providerId}/keys`);
  }

  async addProviderKey(providerId, data) {
    return this.request(`/admin/providers/${providerId}/keys`, {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async deleteProviderKey(providerId, keyId) {
    return this.request(`/admin/providers/${providerId}/keys/${keyId}`, {
      method: 'DELETE',
    });
  }

  async toggleProviderKey(providerId, keyId, enabled) {
    return this.request(`/admin/providers/${providerId}/keys/${keyId}/toggle`, {
      method: 'PUT',
      body: JSON.stringify({ enabled }),
    });
  }

  // Model Management
  async getProviderModels(providerId) {
    return this.request(`/admin/providers/${providerId}/models`, {
      method: 'GET',
    });
  }

  async configureProviderModels(providerId, enabledModels) {
    return this.request(`/admin/providers/${providerId}/models/configure`, {
      method: 'POST',
      body: JSON.stringify({ enabled_models: enabledModels }),
    });
  }

  async importProviderModels(providerId) {
    return this.request(`/admin/providers/${providerId}/models/import`, {
      method: 'POST',
    });
  }

  async syncProviderModels() {
    return this.request('/admin/providers/sync-models', {
      method: 'POST',
    });
  }

  // Provider CRUD Operations
  async createProvider(providerData) {
    return this.request('/admin/providers', {
      method: 'POST',
      body: JSON.stringify(providerData),
    });
  }

  async updateProvider(providerId, updates) {
    return this.request(`/admin/providers/${providerId}`, {
      method: 'PUT',
      body: JSON.stringify(updates),
    });
  }

  async deleteProvider(providerId) {
    return this.request(`/admin/providers/${providerId}`, {
      method: 'DELETE',
    });
  }
}

export default AdminAPI;
