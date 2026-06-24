// In newer @tanstack/router-plugin this export may be named `tanstackRouter`.
import { TanStackRouterVite } from '@tanstack/router-plugin/vite'
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// The Go backend listens on :8081 locally (SERVER_PORT in .env). In Docker the
// Compose file sets VITE_API_TARGET to the backend service URL.
const apiTarget = process.env.VITE_API_TARGET ?? 'http://localhost:8081'

// Proxy API calls to the backend so the client uses relative URLs and we avoid
// CORS. `/api/movies` → `<apiTarget>/movies`. Shared by the dev server and the
// `vite preview` server used in the production container.
const proxy = {
  '/api': {
    changeOrigin: true,
    rewrite: (path: string) => path.replace(/^\/api/u, ''),
    target: apiTarget,
  },
}

export default defineConfig(({ isPreview }) => ({
  // `vite preview` (used by the production container) only serves the prebuilt
  // dist/ and proxies /api, so skip the build-time plugins. They would
  // otherwise try to (re)generate the route tree into src/, which isn't present
  // in the runtime image.
  plugins: isPreview ? [] : [
    // The router plugin must run before the React plugin.
    TanStackRouterVite(),
    react(),
  ],
  preview: { proxy },
  server: { proxy },
}))
