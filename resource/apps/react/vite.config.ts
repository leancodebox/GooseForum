import { fileURLToPath, URL } from 'node:url'
import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import { defineConfig, loadEnv, type ProxyOptions } from 'vite'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const backendOrigin = env.GOOSE_DEV_ORIGIN || 'http://127.0.0.1:5234'
  const backendProxy: ProxyOptions = {
    target: backendOrigin,
    changeOrigin: true,
  }

  return {
    plugins: [react(), tailwindcss()],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
      },
    },
    build: {
      manifest: true,
      outDir: 'dist',
      emptyOutDir: true,
    },
    server: {
      port: 3011,
      strictPort: true,
      proxy: {
        '/__goose_page': {
          ...backendProxy,
          headers: {
            Accept: 'application/json',
            'X-Goose-Page': 'true',
          },
          rewrite: (requestPath) => requestPath.replace(/^\/__goose_page/, '') || '/',
        },
        '/api': backendProxy,
        '/file': backendProxy,
        '/oauth2': backendProxy,
        '/site-theme.css': backendProxy,
        '/static': backendProxy,
      },
    },
  }
})
