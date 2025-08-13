import {createApp} from 'vue'
import Login from './login.vue'
import './style.css'
import './utils/notification.ts'

const app = createApp(Login)
app.mount('#login')
