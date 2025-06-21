import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import { resolve, sep, basename } from 'path'
import { glob } from 'glob'

// 自动扫描 src 目录下的所有 js 文件作为入口
const getEntries = () => {
  const entries = {}
  const jsFiles = glob.sync('src/*.js', { cwd: __dirname })
  
  jsFiles.forEach(file => {
    // 使用 path.basename 获取文件名，兼容 Windows 路径
    const name = basename(file, '.js')
    entries[name] = resolve(__dirname, file)
  })
  
  return entries
}

export default defineConfig({
  plugins: [vue(), tailwindcss(),],
  build: {
    outDir: 'static/dist',
    assetsDir: 'assets',
    manifest: true, // 生成 manifest.json
    rollupOptions: {
      input: getEntries(),
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
