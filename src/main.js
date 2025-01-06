import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './route/router'

// 创建应用实例
const app = createApp(App)

// 创建并安装 Pinia（在其他插件之前）
const pinia = createPinia()
app.use(pinia)

// 安装路由
app.use(router)

// 挂载应用
app.mount('#app') 