import { createApp } from 'vue'
import Guestbook from './Guestbook.vue'
import './style.css'

const app = createApp(Guestbook)
app.mount('#guestbook-app')