import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { axiosInstance } from '../utils/axiosInstance'

export interface User {
  id: number
  username: string
  email: string
  avatar?: string
  role: string
  isAdmin: boolean
}

export interface LoginCredentials {
  username: string
  password: string
}

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const token = ref<string | null>(localStorage.getItem('admin_token'))
  const loading = ref(false)
  const error = ref<string | null>(null)

  // 计算属性
  const isAuthenticated = computed(() => {
    return !!token.value && !!user.value
  })

  const isAdmin = computed(() => {
    return user.value?.isAdmin || false
  })

  // 登录
  const login = async (credentials: LoginCredentials) => {
    loading.value = true
    error.value = null

    try {
      const response = await axiosInstance.post('/api/admin/login', credentials)
      const { token: newToken, user: userData } = response.data.data

      // 保存 token 和用户信息
      token.value = newToken
      user.value = userData
      localStorage.setItem('admin_token', newToken)

      // 设置 axios 默认 header
      axiosInstance.defaults.headers.common['Authorization'] = `Bearer ${newToken}`

      return { success: true }
    } catch (err: any) {
      error.value = err.response?.data?.message || '登录失败'
      return { success: false, error: error.value }
    } finally {
      loading.value = false
    }
  }

  // 退出登录
  const logout = async () => {
    try {
      await axiosInstance.post('/api/admin/logout')
    } catch (err) {
      console.warn('退出登录请求失败:', err)
    } finally {
      // 清除本地状态
      user.value = null
      token.value = null
      localStorage.removeItem('admin_token')
      delete axiosInstance.defaults.headers.common['Authorization']
    }
  }

  // 获取用户信息
  const fetchUserInfo = async () => {
    if (!token.value) return

    loading.value = true
    try {
      const response = await axiosInstance.get('/api/admin/user/info')
      user.value = response.data.data
    } catch (err: any) {
      console.error('获取用户信息失败:', err)
      // 如果 token 无效，清除认证状态
      if (err.response?.status === 401) {
        await logout()
      }
    } finally {
      loading.value = false
    }
  }

  // 初始化认证状态
  const initAuth = async () => {
    if (token.value) {
      axiosInstance.defaults.headers.common['Authorization'] = `Bearer ${token.value}`
      await fetchUserInfo()
    }
  }

  // 检查权限
  const hasPermission = (permission: string) => {
    if (!user.value) return false
    if (user.value.isAdmin) return true
    // 这里可以根据实际的权限系统进行扩展
    return false
  }

  return {
    // 状态
    user,
    token,
    loading,
    error,
    
    // 计算属性
    isAuthenticated,
    isAdmin,
    
    // 方法
    login,
    logout,
    fetchUserInfo,
    initAuth,
    hasPermission
  }
})