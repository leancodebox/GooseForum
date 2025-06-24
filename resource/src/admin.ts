import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './Admin.vue'
import router from './admin/router/index'
import './admin.css'

// 创建应用实例
const app = createApp(App)

// 使用 Pinia 状态管理
app.use(createPinia())

// 使用路由
app.use(router)

// 挂载应用
app.mount('#app')
