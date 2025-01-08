import { defineStore } from 'pinia'
import { ref, watch } from 'vue'
import { getUserInfo } from '@/service/request'

export const useUserStore = defineStore('user', () => {
  // 从 localStorage 初始化 token
  const token = ref(localStorage.getItem('token') || '')
  // 标记是否已经获取过用户信息
  const hasLoadedUserInfo = ref(false)
  const userInfo = ref({
    userId: '',
    username: '',
    avatarUrl: '',
    email: '',
    nickname: '',
    isAdmin: false
  })

  // 监听 token 变化并保存到 localStorage
  watch(() => token.value, (newToken) => {
    if (newToken) {
      localStorage.setItem('token', newToken)
    } else {
      localStorage.removeItem('token')
      hasLoadedUserInfo.value = false // token 被清除时重置标志
    }
  })

  function setToken(newToken) {
    token.value = newToken
  }

  function setUserInfo(info) {
    userInfo.value = {
      ...info,
      isAdmin: !!info.isAdmin
    }
    hasLoadedUserInfo.value = true
  }

  function clearUserInfo() {
    token.value = ''
    userInfo.value = {
      userId: '',
      username: '',
      avatarUrl: '',
      email: '',
      nickname: '',
      isAdmin: false
    }
    hasLoadedUserInfo.value = false
  }

  // 刷新用户信息
  async function refreshUserInfo() {
    if (token.value && !hasLoadedUserInfo.value) {
      try {
        const res = await getUserInfo()
        setUserInfo(res.result)
      } catch (error) {
        console.error('Failed to refresh user info:', error)
        throw error
      }
    }
  }

  return {
    token,
    userInfo,
    setToken,
    setUserInfo,
    clearUserInfo,
    refreshUserInfo,
    hasLoadedUserInfo
  }
})
