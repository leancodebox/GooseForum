import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getUserInfo } from '@/service/request'

export const useUserStore = defineStore('user', () => {
  const token = ref('')
  const userInfo = ref({
    userId: '',
    username: '',
    avatarUrl: '',
    email: '',
    nickname: ''
  })

  function setToken(newToken) {
    token.value = newToken
  }

  function setUserInfo(info) {
    userInfo.value = info
  }

  function clearUserInfo() {
    token.value = ''
    userInfo.value = {
      userId: '',
      username: '',
      avatarUrl: '',
      email: '',
      nickname: ''
    }
  }

  // 刷新用户信息
  async function refreshUserInfo() {
    if (token.value) {
      try {
        const res = await getUserInfo()
        setUserInfo(res.result)
      } catch (error) {
        console.error('Failed to refresh user info:', error)
      }
    }
  }

  return {
    token,
    userInfo,
    setToken,
    setUserInfo,
    clearUserInfo,
    refreshUserInfo
  }
})
