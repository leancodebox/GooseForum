import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getUnreadCount } from '@/service/request'

const STORAGE_KEY = {
  COUNT: 'notification_unread_count',
  LAST_FETCH: 'notification_last_fetch_time'
}

const CACHE_DURATION = 20000 // 缓存时间：1分钟

export const useNotificationStore = defineStore('notification', () => {
  // 从 localStorage 初始化状态
  const unreadCount = ref(parseInt(localStorage.getItem(STORAGE_KEY.COUNT) || '0'))
  let fetchTimer = null

  // 检查是否需要更新
  const shouldUpdate = (force = false) => {
    if (force) return true

    const now = Date.now()
    const lastFetch = parseInt(localStorage.getItem(STORAGE_KEY.LAST_FETCH) || '0')
    return now - lastFetch >= CACHE_DURATION
  }

  // 更新本地存储
  const updateLocalStorage = (count) => {
    const now = Date.now()
    localStorage.setItem(STORAGE_KEY.COUNT, count.toString())
    localStorage.setItem(STORAGE_KEY.LAST_FETCH, now.toString())
    unreadCount.value = count
  }

  // 获取未读消息数量
  const fetchUnreadCount = async (force = false) => {
    // 优先检查是否需要更新
    if (!shouldUpdate(force)) {
      return unreadCount.value
    }

    try {
      const res = await getUnreadCount()
      const count = res.result.count || 0
      updateLocalStorage(count)
      return count
    } catch (error) {
      console.error('获取未读消息数量失败:', error)
      return unreadCount.value
    }
  }

  // 启动定时获取
  const startPolling = () => {
    // 如果已经有定时器在运行，直接返回
    if (fetchTimer) return

    // 检查是否需要立即获取
    if (shouldUpdate()) {
      fetchUnreadCount()
    }

    // 设置定时器
    fetchTimer = setInterval(() => {
      // 每次定时器触发时也检查是否需要更新
      if (shouldUpdate()) {
        fetchUnreadCount()
      }
    }, CACHE_DURATION)

    // 监听 storage 事件，在其他标签页更新数据时同步状态
    window.addEventListener('storage', handleStorageChange)
  }

  // 处理 storage 事件
  const handleStorageChange = (event) => {
    if (event.key === STORAGE_KEY.COUNT) {
      unreadCount.value = parseInt(event.newValue || '0')
    }
  }

  // 停止定时获取
  const stopPolling = () => {
    if (fetchTimer) {
      clearInterval(fetchTimer)
      fetchTimer = null
    }
    window.removeEventListener('storage', handleStorageChange)
  }

  // 重置未读数量
  const resetUnreadCount = () => {
    updateLocalStorage(0)
  }

  // 强制刷新未读数量
  const refreshUnreadCount = () => {
    return fetchUnreadCount(true)
  }

  // 清理函数
  const cleanup = () => {
    stopPolling()
  }

  return {
    unreadCount,
    fetchUnreadCount,
    startPolling,
    stopPolling,
    resetUnreadCount,
    refreshUnreadCount,
    cleanup
  }
})
