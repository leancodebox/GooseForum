import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getUnreadCount } from '@/service/request'

export const useNotificationStore = defineStore('notification', () => {
  const unreadCount = ref(0)
  let lastFetchTime = 0
  let fetchTimer = null

  // 获取未读消息数量
  const fetchUnreadCount = async (force = false) => {
    const now = Date.now()
    // 如果距离上次获取时间小于1分钟且不是强制刷新，则使用缓存的数量
    if (!force && now - lastFetchTime < 60000) {
      return unreadCount.value
    }

    try {
      const res = await getUnreadCount()
      unreadCount.value = res.result.count || 0
      lastFetchTime = now
      return unreadCount.value
    } catch (error) {
      console.error('获取未读消息数量失败:', error)
      return unreadCount.value
    }
  }

  // 启动定时获取
  const startPolling = () => {
    if (fetchTimer) return // 如果已经存在定时器，则不重复创建

    fetchUnreadCount() // 立即获取一次
    fetchTimer = setInterval(() => {
      fetchUnreadCount()
    }, 60000) // 每分钟获取一次
  }

  // 停止定时获取
  const stopPolling = () => {
    if (fetchTimer) {
      clearInterval(fetchTimer)
      fetchTimer = null
    }
  }

  // 重置未读数量
  const resetUnreadCount = () => {
    unreadCount.value = 0
    lastFetchTime = Date.now()
  }

  // 强制刷新未读数量
  const refreshUnreadCount = () => {
    return fetchUnreadCount(true)
  }

  return {
    unreadCount,
    fetchUnreadCount,
    startPolling,
    stopPolling,
    resetUnreadCount,
    refreshUnreadCount
  }
}) 