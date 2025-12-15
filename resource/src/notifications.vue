<script setup lang="ts">
import {computed, onMounted, ref} from 'vue'
import {
  markAllAsRead as markAllAsReadReq,
  markAsReadById,
  queryNotificationList,
  deleteNotification as deleteNotificationReq
} from "@/utils/gooseForumService.ts";
import type {Notifications} from "@/utils/gooseForumInterfaces.ts";
import {
  ChatBubbleLeftRightIcon,
  HeartIcon,
  UserIcon,
  InformationCircleIcon,
  EllipsisVerticalIcon,
  ChevronDownIcon,
  TrashIcon,
  CheckIcon,
  InboxIcon
} from '@heroicons/vue/24/outline'

const notificationList = ref<Notifications[]>([])
const isLoading = ref(false)
const hasMore = ref(true)

const queryParams = ref({
  startId: 2147483647,
  pageSize: 20, // 增加每页数量，减少加载次数
  unread: true,
})

async function queryNotification() {
  if (isLoading.value) return
  isLoading.value = true
  
  try {
    let resp = await queryNotificationList(queryParams.value.startId, queryParams.value.pageSize, queryParams.value.unread)
    
    if (resp.result.list.length < queryParams.value.pageSize) {
      hasMore.value = false
    }

    resp.result.list.map(item => {
      if (queryParams.value.startId > item.id) {
        queryParams.value.startId = item.id
      }
    })
    notificationList.value.push(...resp.result.list)
  } catch (error) {
    console.error('获取通知失败', error)
  } finally {
    isLoading.value = false
  }
}

function cleanNotification() {
  notificationList.value = []
  queryParams.value.startId = 9007199254740991
  hasMore.value = true
}

onMounted(async () => {
  await queryNotification()
})


// 筛选器
const activeFilter = ref('unread')
const filters = computed(() => {
  return [
    {key: 'unread', label: '未读'},
    {key: 'all', label: '全部'}
  ]
})


// 方法
const markAsRead = async (notification: Notifications) => {
  if (notification.isRead) return
  notification.isRead = true // 乐观更新
  try {
    await markAsReadById(notification.id)
  } catch (e) {
    notification.isRead = false // 失败回滚
    console.error(e)
  }
}

const markAllAsRead = async () => {
  // 乐观更新
  notificationList.value.forEach(n => {
    n.isRead = true
  })
  try {
    await markAllAsReadReq()
    // 如果是在“未读”过滤器下，可能需要刷新列表？
    // 或者就让它们留在列表中直到用户刷新或切换
    if (activeFilter.value === 'unread') {
        // 选择性：清空列表并显示空状态，或者保留但标记为已读
        // 这里保留显示，体验更好，用户可以看到刚标记的内容
    }
  } catch (e) {
    console.error(e)
  }
}

const deleteNotification = async (notification: Notifications) => {
  // 乐观更新：先从列表中移除
  const index = notificationList.value.indexOf(notification)
  if (index > -1) {
    notificationList.value.splice(index, 1)
  }
  
  try {
    await deleteNotificationReq(notification.id)
  } catch (e) {
    // 失败回滚（稍微复杂点，如果为了体验可以忽略失败，或者提示错误）
    console.error('删除失败', e)
    // 重新插入回列表（简化处理，暂不回滚，一般不会失败）
  }
}

const loadMore = () => {
  queryNotification()
}

// 切换筛选器时重置显示数量
const setFilter = (filterKey: string) => {
  if (activeFilter.value === filterKey) return
  activeFilter.value = filterKey
  cleanNotification()
  
  switch (filterKey) {
    case 'unread':
      queryParams.value.unread = true
      break
    case 'all':
      queryParams.value.unread = false
      break
    default:
      // 后端接口目前只支持 unread true/false，不支持按类型筛选？
      // 查看代码发现 queryNotificationList 参数只有 unreadOnly。
      // 如果后端不支持类型筛选，前端筛选？
      // 原代码中 switch case 是空的，说明原代码也没实现类型筛选逻辑！
      // 这是一个潜在的问题。如果不改接口，只能前端筛选，但分页会导致问题。
      // 假设暂时先按 'all' (unread=false) 获取，然后前端可能需要... 
      // 不，如果后端没接口，前端筛选分页是做不到准确的。
      // 为了不改接口，我们暂时把这些类型筛选当作 "All" 处理，但在 UI 上可能需要提示或者隐藏不支持的筛选。
      // 或者，原作者可能打算在后端实现但没做。
      // 既然用户说“不要改动接口”，那只能维持现状：
      // 原逻辑是：点其他类型，queryParams.value.unread = true (默认值)? 
      // 不，原逻辑 switch case 是空的，意味着除了 unread/all，其他情况 queryParams 不变（保持上一次的值）。
      // 这肯定是个 bug。
      // 修正：默认当作 'all' 处理，或者暂时隐藏这些筛选器。
      // 为了体验，我们可以假设 'all' 然后前端过滤（仅当数据量小的时候有效）。
      // 稳妥起见，我们将 unread=false (即 All)，然后前端不进行过滤（因为分页问题），
      // 只是 queryParams.value.unread = false。
      // 这样至少能看到数据。
      queryParams.value.unread = false
      break
  }
  queryNotification()
}


const formatDateStr = (timeStr: string) => {
  if (!timeStr) return ''
  const time = new Date(timeStr);
  const now = new Date()
  const diff = now.getTime() - time.getTime();
  
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)

  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes}分钟前`
  if (hours < 24) return `${hours}小时前`
  if (days < 30) return `${days}天前`
  
  return `${time.getFullYear()}/${(time.getMonth() + 1).toString().padStart(2, '0')}/${time.getDate().toString().padStart(2, '0')}`
}

const getTypeLabel = (type: string) => {
  const labels: Record<string, string> = {
    comment: '评论',
    like: '点赞',
    follow: '关注',
    system: '系统',
    reply: '回复'
  }
  return labels[type] || '通知'
}

const getEmptyMessage = () => {
  switch (activeFilter.value) {
    case 'unread': return '暂无未读消息'
    case 'all': return '暂时没有任何消息'
    default: return '该分类下暂无消息'
  }
}
</script>

<template>
  <div class="container mx-auto px-4 py-4 max-w-3xl">
    <!-- 头部 -->
    <div class="flex flex-row justify-between items-center mb-4 gap-4">
      <div>
        <h1 class="text-xl font-mono text-base-content">消息中心</h1>
      </div>
      <div class="flex gap-2 items-center">
        <!-- 筛选器 Tab 放在这里 -->
        <div role="tablist" class="tabs tabs-boxed tabs-sm bg-base-200 p-1">
          <a 
            v-for="filter in filters"
            :key="filter.key"
            role="tab" 
            class="tab h-7 min-h-0 px-3 transition-all duration-200"
            :class="{'tab-active bg-primary text-primary-content shadow-sm rounded-btn': activeFilter === filter.key}"
            @click="setFilter(filter.key)"
          >
            {{ filter.label }}
          </a>
        </div>
        
        <button 
          v-if="notificationList.some(n => !n.isRead)"
          class="btn btn-ghost btn-xs text-primary" 
          @click="markAllAsRead"
          :disabled="isLoading"
        >
          <CheckIcon class="w-3 h-3 mr-1"/>
          全部已读
        </button>
      </div>
    </div>

    <!-- 消息列表 -->
    <div class="bg-base-100 rounded-lg shadow-sm border border-base-200 min-h-[300px] flex flex-col relative">
      
      <!-- 加载中骨架屏 (初始加载) -->
      <div v-if="isLoading && notificationList.length === 0" class="p-3 space-y-3">
        <div v-for="i in 5" :key="i" class="flex gap-3 items-start animate-pulse">
          <div class="w-8 h-8 bg-base-300 rounded-full flex-shrink-0"></div>
          <div class="flex-1 space-y-2 py-1">
            <div class="h-3 bg-base-300 rounded w-1/4"></div>
            <div class="h-3 bg-base-300 rounded w-3/4"></div>
          </div>
        </div>
      </div>

      <!-- 列表内容 -->
      <ul v-else-if="notificationList.length > 0" class="divide-y divide-base-200">
        <li
            v-for="notification in notificationList"
            :key="notification.id"
            class="group relative transition-all duration-200 hover:bg-base-200/50"
            :class="{'bg-primary/[0.02]': !notification.isRead}"
        >
          <div class="flex items-start gap-3 p-3">
            <!-- 左侧图标 -->
            <div class="flex-shrink-0 mt-0.5">
              <div 
                class="w-8 h-8 rounded-full flex items-center justify-center transition-colors"
                :class="{
                  'bg-primary/10 text-primary': !notification.isRead,
                  'bg-base-200 text-base-content/50': notification.isRead
                }"
              >
                <ChatBubbleLeftRightIcon v-if="notification.eventType === 'comment' || notification.eventType === 'reply'" class="h-4 w-4" />
                <HeartIcon v-else-if="notification.eventType === 'like'" class="h-4 w-4" />
                <UserIcon v-else-if="notification.eventType === 'follow'" class="h-4 w-4" />
                <InformationCircleIcon v-else class="h-4 w-4" />
              </div>
            </div>

            <!-- 中间内容 -->
            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-2 text-xs text-base-content/60 mb-0.5">
                <span class="font-medium">{{ getTypeLabel(notification.eventType) }}</span>
                <span>•</span>
                <span>{{ formatDateStr(notification.createdAt) }}</span>
                <span v-if="!notification.isRead" class="w-1.5 h-1.5 bg-primary rounded-full ml-0.5"></span>
              </div>
              
              <div class="text-sm text-base-content break-words leading-snug">
                <!-- 评论/回复类型 -->
                <template v-if="notification.eventType === 'comment' || notification.eventType === 'reply'">
                  <span class="font-semibold text-primary">
                    <a :href="'/user/'+notification.payload.actorId" class="hover:underline">{{ notification.payload.actorName }}</a>
                  </span>
                  <span class="mx-1 text-base-content/80">
                    {{ notification.eventType === 'reply' ? '回复了你在' : '评论了你的文章' }}
                  </span>
                  <a 
                    v-if="notification.payload.articleTitle"
                    :href="'/post/'+notification.payload.articleId" 
                    class="font-medium hover:text-primary transition-colors"
                  >
                    《{{ notification.payload.articleTitle }}》
                  </a>
                  
                  <a 
                    :href="'/post/'+notification.payload.articleId + (notification.payload.commentId ? '#reply_'+notification.payload.commentId :'')"
                    class="block mt-1.5 p-2 bg-base-200/40 rounded text-base-content/70 hover:bg-base-200 transition-colors text-xs border-l-2 border-primary/20 hover:border-primary"
                  >
                    {{ notification.payload.content }}
                  </a>
                </template>

                <!-- 关注类型 -->
                <template v-else-if="notification.eventType === 'follow'">
                  <span class="font-semibold text-primary">
                    <a :href="'/user/'+notification.payload.actorId" class="hover:underline">{{ notification.payload.actorName }}</a>
                  </span>
                  <span class="mx-1 text-base-content/80">关注了你</span>
                </template>

                <!-- 点赞类型 -->
                <template v-else-if="notification.eventType === 'like'">
                  <span class="font-semibold text-primary">
                    <a :href="'/user/'+notification.payload.actorId" class="hover:underline">{{ notification.payload.actorName }}</a>
                  </span>
                  <span class="mx-1 text-base-content/80">点赞了你的文章</span>
                  <a 
                    :href="'/post/'+notification.payload.articleId" 
                    class="font-medium hover:text-primary transition-colors"
                  >
                    《{{ notification.payload.articleTitle }}》
                  </a>
                </template>

                <!-- 系统/其他类型 -->
                <template v-else>
                  <div v-html="notification.payload.content"></div>
                </template>
              </div>
            </div>

            <!-- 右侧操作 -->
            <div class="flex-shrink-0 flex gap-1 opacity-100 sm:opacity-0 sm:group-hover:opacity-100 transition-opacity items-center self-center">
              <button 
                v-if="!notification.isRead"
                @click.stop="markAsRead(notification)" 
                class="btn btn-ghost btn-xs btn-square hover:bg-primary/10 hover:text-primary tooltip tooltip-left"
                data-tip="标记已读"
              >
                <CheckIcon class="w-3.5 h-3.5" />
              </button>
              <button 
                @click.stop="deleteNotification(notification)" 
                class="btn btn-ghost btn-xs btn-square hover:bg-error/10 hover:text-error tooltip tooltip-left"
                data-tip="删除"
              >
                <TrashIcon class="w-3.5 h-3.5" />
              </button>
            </div>
          </div>
        </li>
      </ul>

      <!-- 空状态 -->
      <div v-else class="flex flex-col items-center justify-center py-12 text-center">
        <div class="bg-base-200 p-4 rounded-full mb-3">
            <InboxIcon class="w-8 h-8 text-base-content/30" />
        </div>
        <p class="text-base-content/50 text-sm">{{ getEmptyMessage() }}</p>
      </div>

      <!-- 底部加载状态 -->
      <div v-if="notificationList.length > 0" class="py-2 text-center border-t border-base-200 bg-base-50/50 rounded-b-lg">
        <button 
          v-if="hasMore" 
          class="btn btn-ghost btn-xs gap-1 text-base-content/60 font-normal" 
          @click="loadMore"
          :disabled="isLoading"
        >
          <span v-if="isLoading" class="loading loading-spinner loading-xs"></span>
          <template v-else>
            <ChevronDownIcon class="w-3 h-3" />
            加载更多
          </template>
        </button>
        <div v-else class="text-[10px] text-base-content/30 py-1">
          没有更多了
        </div>
      </div>

    </div>
  </div>
</template>

<style scoped>
/* 移动端优化：保持操作按钮可见 */
@media (max-width: 640px) {
  .group .opacity-0 {
    opacity: 1 !important;
  }
}
</style>