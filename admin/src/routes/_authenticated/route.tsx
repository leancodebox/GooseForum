import { createFileRoute, redirect } from '@tanstack/react-router'
import { AuthenticatedLayout } from '@/components/layout/authenticated-layout'
import { useAuthStore } from '@/stores/auth-store'
import axios from 'axios'

export const Route = createFileRoute('/_authenticated')({
  beforeLoad: async ({ location }) => {
    const { auth } = useAuthStore.getState()
    
    // 如果本地没有用户信息，尝试通过 Cookie 从后端同步状态
    if (!auth.user) {
      try {
        const response = await axios.get('/api/get-user-info')
        if (response.data.code === 0) {
          const userData = response.data.result
          auth.setUser({
            accountNo: userData.userId.toString(),
            email: userData.email,
            role: userData.isAdmin ? ['admin'] : ['user'],
            exp: Date.now() + 24 * 60 * 60 * 1000,
          })
          // 成功同步用户信息，继续加载页面
          return
        }
      } catch (error) {
        // 请求失败说明未登录或 Cookie 已失效
      }

      // 只有在尝试同步失败后才重定向到登录页
      throw redirect({
        to: '/sign-in',
        search: {
          redirect: location.href,
        },
      })
    }
  },
  component: AuthenticatedLayout,
})
