<script setup lang="ts">
import { ref, reactive,computed } from 'vue'
import {getNotificationList} from "@/utils/articleService.ts";
let r = getNotificationList()
// 消息数据
const notifications = ref([
  {
    id: 1,
    type: 'comment',
    title: '新评论通知',
    content: 'ReactDev 评论了你的文章《Vue 3 组合式 API 深度解析》',
    relatedInfo: 'Vue 3 组合式 API 深度解析',
    createTime: new Date('2024-01-15T10:30:00'),
    isRead: false
  },
  {
    id: 2,
    type: 'like',
    title: '点赞通知',
    content: 'NodeMaster 点赞了你的文章《Nuxt.js 性能优化实战指南》',
    relatedInfo: 'Nuxt.js 性能优化实战指南',
    createTime: new Date('2024-01-15T09:15:00'),
    isRead: false
  },
  {
    id: 3,
    type: 'follow',
    title: '新关注者',
    content: 'VueMaster 关注了你',
    relatedInfo: 'VueMaster',
    createTime: new Date('2024-01-14T16:45:00'),
    isRead: false
  },
  {
    id: 4,
    type: 'system',
    title: '系统通知',
    content: '你的文章《TypeScript 进阶技巧分享》已通过审核并发布',
    relatedInfo: 'TypeScript 进阶技巧分享',
    createTime: new Date('2024-01-14T14:20:00'),
    isRead: true
  },
  {
    id: 5,
    type: 'comment',
    title: '新评论通知',
    content: 'JSExpert 评论了你的文章《JavaScript 异步编程最佳实践》',
    relatedInfo: 'JavaScript 异步编程最佳实践',
    createTime: new Date('2024-01-13T11:30:00'),
    isRead: true
  },
  {
    id: 6,
    type: 'like',
    title: '点赞通知',
    content: 'CSSMaster 点赞了你的文章《CSS Grid 布局完全指南》',
    relatedInfo: 'CSS Grid 布局完全指南',
    createTime: new Date('2024-01-12T15:20:00'),
    isRead: true
  },
  {
    id: 7,
    type: 'system',
    title: '系统维护通知',
    content: '系统将于今晚 23:00-01:00 进行维护，期间可能无法访问',
    relatedInfo: null,
    createTime: new Date('2024-01-12T10:00:00'),
    isRead: true
  }
])

// 筛选器
const activeFilter = ref('all')
const filters = computed(() => {
  const all = notifications.value.length
  const unread = notifications.value.filter(n => !n.isRead).length
  const comment = notifications.value.filter(n => n.type === 'comment').length
  const like = notifications.value.filter(n => n.type === 'like').length
  const follow = notifications.value.filter(n => n.type === 'follow').length
  const system = notifications.value.filter(n => n.type === 'system').length

  return [
    { key: 'all', label: '全部', count: all },
    { key: 'unread', label: '未读', count: unread },
    { key: 'comment', label: '评论', count: comment },
    { key: 'like', label: '点赞', count: like },
    { key: 'follow', label: '关注', count: follow },
    { key: 'system', label: '系统', count: system }
  ]
})

// 计算属性
const totalCount = computed(() => notifications.value.length)
const unreadCount = computed(() => notifications.value.filter(n => !n.isRead).length)
const todayCount = computed(() => {
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  return notifications.value.filter(n => n.createTime >= today).length
})

// 过滤后的消息
const filteredNotifications = computed(() => {
  let filtered = notifications.value

  switch (activeFilter.value) {
    case 'unread':
      filtered = filtered.filter(n => !n.isRead)
      break
    case 'comment':
      filtered = filtered.filter(n => n.type === 'comment')
      break
    case 'like':
      filtered = filtered.filter(n => n.type === 'like')
      break
    case 'follow':
      filtered = filtered.filter(n => n.type === 'follow')
      break
    case 'system':
      filtered = filtered.filter(n => n.type === 'system')
      break
  }

  return filtered.sort((a, b) => b.createTime - a.createTime)
})

// 加载更多
const displayCount = ref(10)
const hasMore = computed(() => displayCount.value < filteredNotifications.value.length)

// 显示的消息列表
const displayedNotifications = computed(() => {
  return filteredNotifications.value.slice(0, displayCount.value)
})

// 方法
const markAsRead = (id) => {
  const notification = notifications.value.find(n => n.id === id)
  if (notification) {
    notification.isRead = true
  }
}

const markAsUnread = (id) => {
  const notification = notifications.value.find(n => n.id === id)
  if (notification) {
    notification.isRead = false
  }
}

const markAllAsRead = () => {
  notifications.value.forEach(n => {
    n.isRead = true
  })
}

const deleteNotification = (id) => {
  const index = notifications.value.findIndex(n => n.id === id)
  if (index > -1) {
    notifications.value.splice(index, 1)
  }
}

const clearAll = () => {
  if (confirm('确定要清空所有消息吗？此操作不可恢复。')) {
    notifications.value = []
  }
}

const loadMore = () => {
  displayCount.value += 10
}

// 切换筛选器时重置显示数量
const setFilter = (filterKey) => {
  activeFilter.value = filterKey
  displayCount.value = 10
}

const formatTime = (time) => {
  const now = new Date()
  const diff = now - time
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)

  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes}分钟前`
  if (hours < 24) return `${hours}小时前`
  if (days < 7) return `${days}天前`

  return time.toLocaleDateString('zh-CN')
}

const getTypeLabel = (type) => {
  const labels = {
    comment: '评论',
    like: '点赞',
    follow: '关注',
    system: '系统'
  }
  return labels[type] || '其他'
}

const getEmptyMessage = () => {
  switch (activeFilter.value) {
    case 'unread': return '没有未读消息'
    case 'comment': return '没有评论通知'
    case 'like': return '没有点赞通知'
    case 'follow': return '没有关注通知'
    case 'system': return '没有系统通知'
    default: return '暂时没有任何消息'
  }
}

</script>
<template>
  <div class="container mx-auto px-4 py-4">
    <div class="max-w-4xl mx-auto">
      <div class="flex justify-between items-center mb-2">
        <h1 class="text-3xl font-bold">消息中心</h1>
        <div class="flex gap-2">
          <button class="btn btn-outline btn-sm" @click="markAllAsRead" >
            全部标记为已读
          </button>
          <button class="btn btn-ghost btn-sm" @click="clearAll">
            清空消息
          </button>
        </div>
      </div>

      <!-- 消息筛选 -->
      <div class="flex flex-wrap gap-2 mb-2 p-4 bg-base-200 rounded-lg">
        <button
          v-for="filter in filters"
          :key="filter.key"
          class="btn btn-sm transition-all duration-200 gap-2"
          :class="{
            'btn-primary text-primary-content shadow-lg': activeFilter === filter.key,
            'btn-ghost hover:btn-outline': activeFilter !== filter.key
          }"
          @click="setFilter(filter.key)"
        >
          {{ filter.label }}
          <span
            v-if="filter.count > 0"
            class="badge badge-xs"
            :class="{
              'badge-primary-content bg-primary-content/30 text-primary-content': activeFilter === filter.key,
              'badge-primary text-primary-content': activeFilter !== filter.key
            }"
          >
            {{ filter.count }}
          </span>
        </button>
      </div>

      <!-- 消息列表 -->
      <ul class="menu bg-base-200 rounded-box w-full">
        <li
          v-for="notification in displayedNotifications"
          :key="notification.id"
          class="w-full hover:bg-base-300 transition-colors"
          :class="{
            'bg-primary/10 border-l-4 border-l-primary': !notification.isRead
          }"
        >
          <div class="flex items-center gap-3 p-3 cursor-pointer w-full" @click="markAsRead(notification.id)">
            <!-- 消息图标 -->
            <div class="flex-shrink-0">
              <div class="w-8 h-8 rounded-full bg-neutral text-neutral-content flex items-center justify-center">
                <svg v-if="notification.type === 'comment'" xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
                </svg>
                <svg v-else-if="notification.type === 'like'" xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
                </svg>
                <svg v-else-if="notification.type === 'follow'" xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                </svg>
                <svg v-else xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
            </div>

            <!-- 消息内容 -->
            <div class="flex-1 min-w-0">
              <div class="flex items-start justify-between">
                <div class="flex-1 min-w-0">
                  <div class="flex items-center gap-2">
                    <h4 class="font-medium text-sm truncate">{{ notification.content }}</h4>
                    <div class="badge badge-outline badge-xs flex-shrink-0">{{ getTypeLabel(notification.type) }}</div>
                    <div v-if="!notification.isRead" class="w-2 h-2 bg-primary rounded-full flex-shrink-0"></div>
                  </div>
                  
                  <div v-if="notification.relatedInfo" class="text-xs text-primary hover:underline cursor-pointer mt-1 truncate">
                    {{ notification.relatedInfo }}
                  </div>
                  
                  <div class="text-xs text-base-content/60 mt-1">
                    {{ formatTime(notification.createTime) }}
                  </div>
                </div>

                <!-- 操作按钮 -->
                <div class="dropdown dropdown-end flex-shrink-0">
                  <div tabindex="0" role="button" class="btn btn-ghost btn-xs btn-circle" @click.stop>
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
                    </svg>
                  </div>
                  <ul tabindex="0" class="dropdown-content z-[1] menu p-2 shadow bg-base-100 rounded-box w-32">
                    <li v-if="!notification.isRead"><a @click="markAsRead(notification.id)">标记已读</a></li>
                    <li v-else><a @click="markAsUnread(notification.id)">标记未读</a></li>
                    <li><a @click="deleteNotification(notification.id)" class="text-error">删除</a></li>
                  </ul>
                </div>
              </div>
            </div>
          </div>
        </li>
      </ul>

      <!-- 空状态 -->
      <div v-if="filteredNotifications.length === 0" class="text-center py-12">
        <div class="text-6xl mb-4">📭</div>
        <h3 class="text-xl font-semibold mb-2">暂无消息</h3>
        <p class="text-base-content/60">{{ getEmptyMessage() }}</p>
      </div>

      <!-- 加载更多按钮 -->
      <div  class="flex justify-center mt-6">
        <button class="btn btn-sm btn-outline" @click="loadMore">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 14l-7 7m0 0l-7-7m7 7V3" />
          </svg>
          加载更多消息
        </button>
      </div>
    </div>
  </div>
</template>
<style scoped>
</style>