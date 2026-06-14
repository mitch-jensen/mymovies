import react from '@vitejs/plugin-react'
// In newer @tanstack/router-plugin this export may be named `tanstackRouter`.
import { TanStackRouterVite } from '@tanstack/router-plugin/vite'
import { defineConfig } from 'vite'

// The Go backend listens on :8081 locally (SERVER_PORT in .env). Override with
// VITE_API_TARGET if yours differs.
const apiTarget = process.env.VITE_API_TARGET ?? 'http://localhost:8081'

export default defineConfig({
  plugins: [
    // The router plugin must run before the React plugin.
    TanStackRouterVite(),
    react(),
  ],
  server: {
    proxy: {
      // Proxy API calls to the backend so the client uses relative URLs and we
      // avoid CORS in dev. `/api/movies` → `http://localhost:8081/movies`.
      '/api': {
        target: apiTarget,
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, ''),
      },
    },
  },
})
