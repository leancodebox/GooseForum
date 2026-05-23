import { createApp, h } from 'vue'
import { Toaster } from 'vue-sonner'
import AdminApp from './AdminApp.vue'
import './styles/admin.css'
import 'vue-sonner/style.css'
import { readAdminPayload } from '@/admin/runtime/payload'

const payload = readAdminPayload()

document.documentElement.lang = document.documentElement.lang || 'zh-CN'

createApp({
  render: () => [
    h(AdminApp, { payload }),
    h(Toaster, {
      duration: 5000,
      position: 'bottom-right',
      richColors: false,
      toastOptions: {
        class: 'border bg-popover text-popover-foreground shadow-lg',
      },
    }),
  ],
}).mount('#goose-admin-app')
