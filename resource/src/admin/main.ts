import { createApp, h } from 'vue'
import { Toaster } from 'vue-sonner'
import AdminApp from './AdminApp.vue'
import './styles/admin.css'
import 'vue-sonner/style.css'
import { readAdminPayload } from '@/admin/runtime/payload'
import { adminRouter } from '@/admin/runtime/router'
import { i18n } from '@/runtime/i18n'
import { configureAdminAccess } from '@/admin/runtime/access'

const payload = readAdminPayload()
configureAdminAccess(payload.layout.viewer.adminPermissions)

document.documentElement.lang = document.documentElement.lang || 'zh-CN'

const app = createApp({
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
})

app.use(adminRouter)
app.use(i18n)
await adminRouter.isReady()
app.mount('#goose-admin-app')
