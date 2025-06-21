import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue(), tailwindcss(),],
  build: {
    outDir: 'static/dist',
    assetsDir: 'assets',
    manifest: true, // 生成 manifest.json
    rollupOptions: {
      input: {
        main: resolve(__dirname, 'src/main.js'),
        login: resolve(__dirname, 'src/login.js'),
        notifications: resolve(__dirname, 'src/notifications.js'),
        profile: resolve(__dirname, 'src/profile.js'),
        publish: resolve(__dirname, 'src/publish.js'),
        'publish-v2': resolve(__dirname, 'src/publish-v2.js'),
        'submit-link': resolve(__dirname, 'src/submit-link.js')
      },
      output: {
      }
    }
  },
  server: {
    port: 3009,
  },
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src')
    }
  }
})
