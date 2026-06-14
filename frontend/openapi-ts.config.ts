import { defineConfig } from '@hey-api/openapi-ts'

// Generates the typed client + TanStack Query options from the backend's spec.
// Run `just openapi` at the repo root first to (re)create ../openapi.yaml.
export default defineConfig({
  input: '../openapi.yaml',
  output: 'src/client',
  plugins: ['@hey-api/client-fetch', '@tanstack/react-query'],
})
