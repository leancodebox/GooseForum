import { defineConfig } from 'vitest/config'
import { fileURLToPath } from 'node:url'

export default defineConfig({
  root: fileURLToPath(new URL('..', import.meta.url)),
  test: {
    environment: 'jsdom',
    include: ['theme-default/test/**/*.test.{ts,tsx}'],
    setupFiles: ['./theme-default/test/setup.ts'],
    coverage: {
      provider: 'v8',
      include: ['{theme-default,runtime,ui}/src/**/*.{ts,tsx}'],
      reporter: ['text-summary', 'json-summary', 'lcov'],
      reportsDirectory: '../coverage/react-package',
      thresholds: {
        statements: 59,
        branches: 50,
        functions: 62,
        lines: 61,
      },
    },
  },
})
