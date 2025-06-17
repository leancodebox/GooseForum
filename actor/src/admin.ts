import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './admin/App.vue'
import router from './admin/router'

// 引入naive-ui
import naive from 'naive-ui'

// 引入字体
import 'vfonts/Lato.css'
import 'vfonts/FiraCode.css'

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(naive)

app.mount('#app')
