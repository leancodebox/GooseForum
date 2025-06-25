import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { axiosInstance } from '../utils/axiosInstance'
import { getUserInfo } from '../utils/adminService'

export interface User {
  userId: number
  username: string
  email: string
  avatar?: string
  role: string
  isAdmin: boolean
}

export interface LoginCredentials {
  username: string
  password: string
  captchaId?: string
  captchaCode?: string
}

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const loading = ref(true) // 初始状态为加载中
  const error = ref<string | null>(null)
  const initialized = ref(false)
  
  // 获取用户信息
  const fetchUserInfo = async () => {
    loading.value = true
    try {
      const response = await getUserInfo()
      user.value = response.result
    } catch (err: any) {
      console.error('获取用户信息失败:', err)
      // 如果 token 无效，清除认证状态
      if (err.response?.status === 401) {
        await logout()
      }
    } finally {
      loading.value = false
      initialized.value = true
    }
  }
  
  // 初始化认证状态
  const initAuth = async () => {
    if (!initialized.value) {
      await fetchUserInfo()
    }
  }
  
  // 立即初始化
  initAuth()
  // 计算属性
  const isAuthenticated = computed(() => {
    return !!user.value
  })

  const isAdmin = computed(() => {
    return user.value?.isAdmin || false
  })

  // 登录
  const login = async (credentials: LoginCredentials) => {
    loading.value = true
    error.value = null

    try {
      const response = await axiosInstance.post('/login', credentials)
      if (response.data.code === 0) {
        await fetchUserInfo()
        return { success: true }
      } else {
        return { success: false, error: response.data.msg }
      }
    } catch (err: any) {
      console.error(err)
      error.value = err.response?.data?.message || '登录失败'
      return { success: false, error: error.value }
    } finally {
      loading.value = false
    }
  }

  // 退出登录
  const logout = async () => {
    try {
      await axiosInstance.post('/logout')
    } catch (err) {
      console.warn('退出登录请求失败:', err)
    } finally {
      // 清除本地状态
      user.value = null
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
    loading,
    error,
    initialized,

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