import {createApp} from 'vue'
import Page from './notifications.vue'
import './style.css'
import './utils/notification.ts'

const app = createApp(Page)
app.mount('#app')
