import { fileURLToPath, URL } from 'node:url'
import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import { defineConfig, loadEnv, type ProxyOptions } from 'vite'

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

export default defineConfig(({ command, mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const backendOrigin = env.GOOSE_DEV_ORIGIN || 'http://127.0.0.1:5234'
  const backendProxy: ProxyOptions = {
    target: backendOrigin,
    changeOrigin: true,
  }

  return {
    plugins: [react(), tailwindcss()],
    // Go may serve these chunks from a CDN with its own path prefix.
    base: command === 'build' ? './' : '/',
    resolve: {
      alias: [
        { find: '@', replacement: fileURLToPath(new URL('./src', import.meta.url)) },
        // Serve workspace source so rebuilding client/dist does not trigger reloads.
        ...(command === 'serve' ? [
          {
            find: /^@gooseforum\/client$/,
            replacement: fileURLToPath(new URL('../../packages/client/src/index.ts', import.meta.url)),
          },
          {
            find: /^@gooseforum\/client\//,
            replacement: fileURLToPath(new URL('../../packages/client/src/', import.meta.url)),
          },
        ] : []),
      ],
    },
    optimizeDeps: {
      include: [
        'lucide-react',
        '@dnd-kit/core',
        '@dnd-kit/sortable',
        '@dnd-kit/utilities',
        '@gooseforum/theme-default > cropperjs',
        '@gooseforum/theme-default > qrcode.react',
        '@gooseforum/theme-default > @hcaptcha/react-hcaptcha',
        '@gooseforum/theme-default > react-google-recaptcha',
        '@gooseforum/theme-default > react-google-recaptcha-v3',
      ],
    },
    build: {
      manifest: true,
      outDir: '../../static/dist/react',
      emptyOutDir: true,
    },
    server: {
      port: 3011,
      strictPort: true,
      watch: {
        ignored: ['**/packages/client/dist/**', '**/static/dist/**'],
      },
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
