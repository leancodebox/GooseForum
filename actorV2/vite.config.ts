import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueJsx from '@vitejs/plugin-vue-jsx'
import vueDevTools from 'vite-plugin-vue-devtools'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    vueJsx(),
    vueDevTools(),
    (() => {
      return {
        name: "multi-entry",
        configureServer(server) {
          return () => {
            server.middlewares.use((req, res, next) => {
              const url = req.originalUrl;
              if (url !== undefined && url.startsWith('/app/admin')) {
                req.url = '/admin.html'; // 重写请求路径
              }
              next();
            });
          };
        }
      };
    })()
  ],
  base: "/app/",
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    },
  },
  build: {
    chunkSizeWarningLimit: 750,
    outDir:"../app/assert/frontend/dist",
    rollupOptions: {
      input: {
        main: fileURLToPath(new URL('./index.html', import.meta.url)),
        admin: fileURLToPath(new URL('./admin.html', import.meta.url)),
      }
    },
    emptyOutDir:true,
  },

  server: {
    // 如果使用docker-compose开发模式，设置为false
    proxy: {
      '^/api/.*':'http://localhost:99',
      '^/file/img/.*':'http://localhost:99',
      '^/login':'http://localhost:99',
      '^/logout':'http://localhost:99',
      '^/static/*':'http://localhost:99',
      '^/post/*':'http://localhost:99',
      '^/assets/*':'http://localhost:99',
    }
  },
})
