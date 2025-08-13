<script setup lang="ts">
import {computed, onMounted, ref} from 'vue'
import {markAllAsRead as markAllAsReadReq, markAsReadById, queryNotificationList} from "@/utils/gooseForumService.ts";
import type {Notifications} from "@/utils/gooseForumInterfaces.ts";

const notificationList = ref<Notifications[]>([])

const queryParams = ref({
  startId: 2147483647,
  pageSize: 10,
  unread: true,
})

async function queryNotification() {
  let resp = await queryNotificationList(queryParams.value.startId, queryParams.value.pageSize, queryParams.value.unread)
  resp.result.list.map(item => {
    if (queryParams.value.startId > item.id) {
      queryParams.value.startId = item.id
    }
  })
  notificationList.value.push(...resp.result.list)
}

function cleanNotification() {
  notificationList.value = []
  queryParams.value.startId = 9007199254740991
}

onMounted(async () => {
  await queryNotification()
})


// 消息数据


// 筛选器
const activeFilter = ref('unread')
const filters = computed(() => {
  return [
    {key: 'unread', label: '未读', count: false},
    {key: 'all', label: '全部', count: false},
    {key: 'comment', label: '评论', count: false},
    {key: 'reply', label: '回复', count: false},
    {key: 'like', label: '点赞', count: false},
    {key: 'follow', label: '关注', count: false},
    {key: 'system', label: '系统', count: false}
  ]
})


// 加载更多
const displayCount = ref(10)


// 方法
const markAsRead = (notification: Notifications) => {
  notification.isRead = true
  markAsReadById(notification.id)
}


const markAllAsRead = () => {
  markAllAsReadReq()
  notificationList.value.forEach(n => {
    n.isRead = true
  })
}

const deleteNotification = (notification: Notifications) => {
  // todo
}

const loadMore = () => {
  queryNotification()
}

// 切换筛选器时重置显示数量
const setFilter = (filterKey) => {
  activeFilter.value = filterKey
  cleanNotification()
  switch (filterKey) {
    case 'unread':
      queryParams.value.unread = true
      queryNotification()
      break
    case 'all':
      queryParams.value.unread = false
      queryNotification()
      break
    case 'comment':
    case 'reply':
    case 'like':
    case 'follow':
    case 'system':
      break
  }
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

const formatDateStr = (timeStr: string) => {
  const time = new Date(timeStr);
  const now = new Date()
  const diff = now.getTime() - time.getTime();
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
    case 'unread':
      return '没有未读消息'
    case 'comment':
      return '没有评论通知'
    case 'like':
      return '没有点赞通知'
    case 'follow':
      return '没有关注通知'
    case 'system':
      return '没有系统通知'
    default:
      return '暂时没有任何消息'
  }
}

</script>
<template>
  <div class="container mx-auto px-4 py-4">
    <div class="max-w-4xl mx-auto">
      <div class="flex justify-between items-center mb-2">
        <h1 class="text-3xl font-normal">消息中心</h1>
        <div class="flex gap-2">
          <button class="btn btn-outline btn-sm" @click="markAllAsRead">
            全部标记为已读
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
              v-if="filter.count"
              class="badge badge-xs badge-primary-content "
              :class="{
              'badge-primary-content bg-primary-content/30 text-primary-content': activeFilter === filter.key,
              'badge-primary text-primary-content': activeFilter !== filter.key
            }"
          >
          </span>
        </button>
      </div>

      <!-- 消息列表 -->
      <ul class="list bg-base-200 rounded-box w-full">
        <li
            v-for="notification in notificationList"
            :key="notification.id"
            class="w-full hover:bg-base-300 transition-colors"
            :class="{
            'bg-primary/10 border-l-4 border-l-primary': !notification.isRead
          }"
        >
          <div class="flex items-center gap-3 p-3 cursor-pointer w-full">
            <!-- 消息图标 -->
            <div class="flex-shrink-0">
              <div class="w-8 h-8 rounded-full bg-neutral text-neutral-content flex items-center justify-center">
                <svg v-if="notification.eventType === 'comment'" xmlns="http://www.w3.org/2000/svg" class="h-4 w-4"
                     fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                        d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"/>
                </svg>
                <svg v-else-if="notification.eventType === 'like'" xmlns="http://www.w3.org/2000/svg" class="h-4 w-4"
                     fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                        d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z"/>
                </svg>
                <svg v-else-if="notification.eventType === 'follow'" xmlns="http://www.w3.org/2000/svg" class="h-4 w-4"
                     fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                        d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"/>
                </svg>
                <svg v-else xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24"
                     stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                        d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
                </svg>
              </div>
            </div>
            <!-- 消息内容 -->
            <div class="flex-1 min-w-0">
              <div class="flex items-start justify-between">
                <div class="flex-1 min-w-0">
                  <div class="flex items-center gap-2">
                    <h4 class="font-normal text-sm truncate" v-if="notification.eventType==='comment'">
                      {{ notification.payload.actorName }} 评论了你的文章 {{ notification.payload.title }} :
                      {{ notification.payload.content }}
                    </h4>
                    <h4 class="font-normal text-sm truncate" v-else>
                      {{ notification.payload.content }}
                    </h4>
                    <div class="badge badge-outline badge-xs flex-shrink-0">{{
                        getTypeLabel(notification.eventType)
                      }}
                    </div>
                    <div v-if="!notification.isRead" class="w-2 h-2 bg-primary rounded-full flex-shrink-0"></div>
                  </div>

                  <a v-if="notification.payload.articleTitle && notification.payload.articleId>0"
                     class="text-xs text-primary hover:underline cursor-pointer mt-1 truncate"
                     :href="'/post/'+notification.payload.articleId"
                  >
                    {{ notification.payload.articleTitle }}
                  </a>
                  <div v-else-if="notification.payload.articleTitle"
                       class="text-xs text-primary hover:underline cursor-pointer mt-1 truncate">
                    {{ notification.payload.articleTitle }}
                  </div>

                  <div class="text-xs text-base-content/60 mt-1">
                    {{ formatDateStr(notification.createdAt) }}
                  </div>
                </div>

                <!-- 操作按钮 -->
                <div class="dropdown dropdown-end flex-shrink-0">
                  <div tabindex="0" role="button" class="btn btn-ghost btn-xs btn-circle" @click.stop>
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24"
                         stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                            d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z"/>
                    </svg>
                  </div>
                  <ul tabindex="0" class="dropdown-content z-[1] menu p-2 shadow bg-base-100 rounded-box w-32">
                    <li v-if="!notification.isRead"><a @click="markAsRead(notification)">标记已读</a></li>
                    <li><a @click="deleteNotification(notification)" class="text-error">删除</a></li>
                  </ul>
                </div>
              </div>
            </div>
          </div>
        </li>
      </ul>

      <!-- 空状态 -->
      <div v-if="notificationList.length === 0" class="text-center py-12">
        <div class="text-6xl mb-4">📭</div>
        <h3 class="text-xl font-normal mb-2">暂无消息</h3>
        <p class="text-base-content/60">{{ getEmptyMessage() }}</p>
      </div>

      <!-- 加载更多按钮 -->
      <div v-else class="flex justify-center mt-6">
        <button class="btn btn-sm btn-outline" @click="loadMore">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24"
               stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 14l-7 7m0 0l-7-7m7 7V3"/>
          </svg>
          加载更多消息
        </button>
      </div>
    </div>
  </div>
</template>
<style scoped>
</style>
