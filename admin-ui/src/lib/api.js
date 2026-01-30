// Admin API Client
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

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

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.message || `HTTP ${response.status}`);
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

  async refreshFilters() {
    return this.request('/admin/filters/refresh', {
      method: 'POST',
    });
  }
}

export default AdminAPI;
