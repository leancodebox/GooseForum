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
        site: 'src/site/main.ts',
        admin: 'src/admin/main.ts',
      },
      output: {
        manualChunks(id) {
          if (!id.includes('node_modules')) return

          const normalized = id.replace(/\\/g, '/')

          if (
            normalized.includes('/vue/') ||
            normalized.includes('/vue-router/') ||
            normalized.includes('/vue-i18n/')
          ) {
            return 'vue-core'
          }

          if (
            normalized.includes('/@lucide/') ||
            normalized.includes('/reka-ui/') ||
            normalized.includes('/motion-v/') ||
            normalized.includes('/@vueuse/')
          ) {
            return 'ui-vendor'
          }

          if (
            normalized.includes('/@unovis/') ||
            normalized.includes('/simple-icons/')
          ) {
            return 'charts-vendor'
          }

          if (
            normalized.includes('/markdown-it') ||
            normalized.includes('/cropperjs/') ||
            normalized.includes('/compressorjs/') ||
            normalized.includes('/sortablejs/') ||
            normalized.includes('/vuedraggable/')
          ) {
            return 'editor-vendor'
          }

          if (
            normalized.includes('/node-forge/') ||
            normalized.includes('/universal-cookie/') ||
            normalized.includes('/clsx/') ||
            normalized.includes('/class-variance-authority/') ||
            normalized.includes('/tailwind-merge/') ||
            normalized.includes('/tw-animate-css/') ||
            normalized.includes('/@internationalized/date/')
          ) {
            return 'utils-vendor'
          }
        },
      },
    },
  },
  server: {
    port: 3010,
    origin: 'http://localhost:3010',
  },
})
