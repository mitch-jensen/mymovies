import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'

// Separate from vite.config.ts (which is a build/preview config) so the test
// Runner is self-contained. jsdom gives component tests a DOM; setup wires the
// Jest-dom matchers into Vitest's expect.
export default defineConfig({
  plugins: [react()],
  test: {
    environment: 'jsdom',
    include: ['src/**/*.test.{ts,tsx}'],
    setupFiles: ['./src/test/setup.ts'],
  },
})
