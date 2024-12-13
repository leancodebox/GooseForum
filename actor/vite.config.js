import {defineConfig} from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'
import * as path from 'path'

const alias = {
    '@': path.resolve(__dirname, './src'),
}
// https://vitejs.dev/config/
export default defineConfig({
    plugins: [
        vue(),
        vueDevTools(),
    ],
    base: "/dt/actor/",
    resolve: {
        alias,
    },

    build: {
        chunkSizeWarningLimit: 750,
        outDir: "../app/assert/frontend/dist"
    },

    server: {
        // 如果使用docker-compose开发模式，设置为false
        proxy: {
            '^/api/.*':'http://localhost:99',
        }
    },
})
