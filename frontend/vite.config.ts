import react from '@vitejs/plugin-react'
// In newer @tanstack/router-plugin this export may be named `tanstackRouter`.
import { TanStackRouterVite } from '@tanstack/router-plugin/vite'
import { defineConfig } from 'vite'

// The Go backend listens on :8081 locally (SERVER_PORT in .env). In Docker the
// compose file sets VITE_API_TARGET to the backend service URL.
const apiTarget = process.env.VITE_API_TARGET ?? 'http://localhost:8081'

// Proxy API calls to the backend so the client uses relative URLs and we avoid
// CORS. `/api/movies` → `<apiTarget>/movies`. Shared by the dev server and the
// `vite preview` server used in the production container.
const proxy = {
  '/api': {
    target: apiTarget,
    changeOrigin: true,
    rewrite: (path: string) => path.replace(/^\/api/, ''),
  },
}

export default defineConfig({
  plugins: [
    // The router plugin must run before the React plugin.
    TanStackRouterVite(),
    react(),
  ],
  server: { proxy },
  preview: { proxy },
})
