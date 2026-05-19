import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig({
  base: '/assets/',
  plugins: [vue(), tailwindcss()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  build: {
    outDir: 'static/dist',
    assetsDir: 'assets',
    manifest: true,
    emptyOutDir: true,
    rollupOptions: {
      input: {
        main: 'src/main.ts',
      },
    },
  },
  server: {
    port: 3010,
    origin: 'http://localhost:3010',
  },
})
