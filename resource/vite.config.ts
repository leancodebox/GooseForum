import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import { fileURLToPath, URL } from 'node:url'

const vendorChunkGroups = {
  'vue-core': ['/vue/', '/vue-router/', '/vue-i18n/'],
  'ui-vendor': ['/@lucide/', '/reka-ui/', '/motion-v/', '/@vueuse/'],
  'charts-vendor': ['/@unovis/', '/simple-icons/'],
  'editor-vendor': ['/markdown-it', '/cropperjs/', '/compressorjs/', '/sortablejs/', '/vuedraggable/'],
  'utils-vendor': [
    '/node-forge/',
    '/universal-cookie/',
    '/clsx/',
    '/class-variance-authority/',
    '/tailwind-merge/',
    '/tw-animate-css/',
    '/@internationalized/date/',
  ],
} as const

function manualChunks(id: string) {
  const normalized = id.replace(/\\/g, '/')

  if (!normalized.includes('node_modules')) return

  for (const [chunkName, packageMatchers] of Object.entries(vendorChunkGroups)) {
    if (packageMatchers.some((matcher) => normalized.includes(matcher))) {
      return chunkName
    }
  }
}

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
        manualChunks,
      },
    },
  },
  server: {
    port: 3010,
    origin: 'http://localhost:3010',
  },
})
