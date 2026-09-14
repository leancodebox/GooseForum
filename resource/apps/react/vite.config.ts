import { fileURLToPath, URL } from 'node:url'
import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import { defineConfig, loadEnv, type Plugin, type ProxyOptions } from 'vite'

function adminSpaFallback(): Plugin {
  return {
    name: 'gooseforum-admin-spa-fallback',
    configureServer(server) {
      server.middlewares.use((request, _response, next) => {
        if (request.method !== 'GET' || !request.url) {
          next()
          return
        }

        const accept = request.headers.accept || ''
        const url = new URL(request.url, 'http://vite.local')
        if (accept.includes('text/html') && (url.pathname === '/admin' || url.pathname.startsWith('/admin/'))) {
          request.url = `/admin/index.html${url.search}`
        }
        next()
      })
    },
  }
}

function pagePayloadProxy(backendOrigin: string, backendProxy: ProxyOptions): ProxyOptions {
  return {
    ...backendProxy,
    headers: {
      Accept: 'application/json',
      'X-Goose-Page': 'true',
    },
    rewrite: (requestPath) => requestPath.replace(/^\/__goose_page/, '') || '/',
    configure(proxy) {
      proxy.on('proxyRes', (response) => {
        const locationHeader = response.headers.location
        const location = Array.isArray(locationHeader) ? locationHeader[0] : locationHeader
        if (!location) return

        const redirect = new URL(location, backendOrigin)
        const backend = new URL(backendOrigin)
        const isLoopback = redirect.hostname === 'localhost' ||
          redirect.hostname === '127.0.0.1' ||
          redirect.hostname === '::1'
        if (redirect.origin !== backend.origin && !isLoopback) return

        response.headers.location = `/__goose_page${redirect.pathname}${redirect.search}${redirect.hash}`
      })
    },
  }
}

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const backendOrigin = env.GOOSE_DEV_ORIGIN || 'http://127.0.0.1:5234'
  const backendProxy: ProxyOptions = {
    target: backendOrigin,
    changeOrigin: true,
  }

  return {
    plugins: [adminSpaFallback(), react(), tailwindcss()],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
      },
    },
    build: {
      manifest: true,
      outDir: 'dist',
      emptyOutDir: true,
      rollupOptions: {
        input: {
          site: fileURLToPath(new URL('./index.html', import.meta.url)),
          admin: fileURLToPath(new URL('./admin/index.html', import.meta.url)),
        },
      },
    },
    server: {
      port: 3011,
      strictPort: true,
      proxy: {
        '/__goose_page': pagePayloadProxy(backendOrigin, backendProxy),
        '/api': backendProxy,
        '/file': backendProxy,
        '/oauth2': backendProxy,
        '/site-theme.css': backendProxy,
        '/static': backendProxy,
      },
    },
  }
})
