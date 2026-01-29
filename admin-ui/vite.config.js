import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'

// https://vite.dev/config/
export default defineConfig({
  plugins: [svelte()],
  server: {
    port: 5173,
    host: true,
    proxy: {
      // Optional: Proxy API calls to backend during development
      // '/api': {
      //   target: 'http://localhost:8080',
      //   changeOrigin: true,
      // }
    }
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
