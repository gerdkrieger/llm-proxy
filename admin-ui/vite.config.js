import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'

// https://vite.dev/config/
export default defineConfig({
  plugins: [svelte()],
  
  server: {
    port: 5173,
    host: true,
    
    // Proxy API requests to backend during development
    // This allows frontend on :5173/:3005 to call backend on :8080
    proxy: {
      // Proxy all /admin/* requests to backend
      '/admin': {
        target: 'http://backend:8080',  // Docker service name in dev
        changeOrigin: true,
        secure: false,
        configure: (proxy, options) => {
          // Log proxy requests for debugging
          proxy.on('proxyReq', (proxyReq, req, res) => {
            console.log('Proxying:', req.method, req.url, '→', options.target + req.url);
          });
        },
      },
      // Proxy all /v1/* requests to backend (LLM API)
      '/v1': {
        target: 'http://backend:8080',
        changeOrigin: true,
        secure: false,
      },
      // Proxy /health endpoint
      '/health': {
        target: 'http://backend:8080',
        changeOrigin: true,
        secure: false,
      },
      // Proxy /metrics endpoint
      '/metrics': {
        target: 'http://backend:8080',
        changeOrigin: true,
        secure: false,
      },
    },
  },
  
  preview: {
    port: 3005,
    host: true
  },
  
  build: {
    outDir: 'dist',
    sourcemap: false,
    minify: 'esbuild',
    target: 'esnext'
  }
})
