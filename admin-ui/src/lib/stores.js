import { writable } from 'svelte/store';

// API Key Store
function createApiKeyStore() {
  // Check if we're in the browser
  const isBrowser = typeof window !== 'undefined';
  const stored = isBrowser ? localStorage.getItem('adminApiKey') : null;
  const { subscribe, set, update } = writable(stored || '');

  return {
    subscribe,
    set: (value) => {
      if (isBrowser) {
        localStorage.setItem('adminApiKey', value);
      }
      set(value);
    },
    clear: () => {
      if (isBrowser) {
        localStorage.removeItem('adminApiKey');
      }
      set('');
    }
  };
}

export const apiKey = createApiKeyStore();

// Navigation Store
export const currentPage = writable('dashboard');

// Loading States
export const loading = writable(false);

// Error Store
export const error = writable(null);

// Data Stores
export const clients = writable([]);
export const cacheStats = writable(null);
export const usageStats = writable(null);
export const providerStatus = writable(null);
