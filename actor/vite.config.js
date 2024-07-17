import {defineConfig} from 'vite'
import vue from '@vitejs/plugin-vue'
import * as path from 'path'

const alias = {
    '@': path.resolve(__dirname, './src'),
}
// https://vitejs.dev/config/
export default defineConfig({
    base: "/actor/",
    resolve: {
        alias,
    },
    plugins: [vue()],
    devServer: {
        proxy: {
            '/api': {
                target: 'http://localhost:80',
                pathRewrite: {},
                ws: true,
                changeOrigin: true
            }
        }
    },
    build: {
        chunkSizeWarningLimit: 750,
        outDir: "../app/assert/frontend/dist"
    }
})
