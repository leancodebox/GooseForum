<script setup>
import {
  NAvatar,
  NButton,
  NCard,
  NEmpty,
  NFlex,
  NList,
  NListItem,
  NSpace,
  NTag,
  NText,
  NTime,
  useMessage
} from "naive-ui"
import {onMounted, onUnmounted, ref} from "vue";
import {
  deleteNotification,
  getNotificationList,
  getNotificationTypes,
  markAllAsRead,
  markAsRead
} from "@/service/request";
import {useRouter} from 'vue-router'
import { useNotificationStore } from '@/modules/notification'

const message = useMessage()
const notifications = ref([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = 20
const loading = ref(false)
const showUnreadOnly = ref(false)

// 通知类型配置
const notificationTypes = ref({
  comment: {name: '评论通知', type: 'info'},
  reply: {name: '回复通知', type: 'success'},
  system: {name: '系统通知', type: 'warning'},
  follow: {name: '关注通知', type: 'error'}
})

const options = [
  {
    label: '只看未读',
    key: 'unread'
  },
  {
    label: '全部消息',
    key: 'all'
  }
]

const notificationStore = useNotificationStore()

// 加载通知列表
async function loadNotifications() {
  loading.value = true
  try {
    const res = await getNotificationList({
      page: currentPage.value,
      pageSize: pageSize,
      unreadOnly: showUnreadOnly.value
    })
    if (res.code === 0) {
      notifications.value = res.result.list.map(notification => ({
        ...notification,
        payload: typeof notification.payload === 'string'
            ? JSON.parse(notification.payload)
            : notification.payload
      }))
      total.value = res.result.total
      notificationStore.refreshUnreadCount()
    }
  } catch (err) {
    console.error('Failed to load notifications:', err)
    message.error('加载通知失败')
  } finally {
    loading.value = false
  }
}

// 处理菜单选择
function handleMenuSelect(key) {
  showUnreadOnly.value = key === 'unread'
  currentPage.value = 1
  loadNotifications()
}

// 标记单条通知为已读
async function handleMarkAsRead(notificationId) {
  try {
    const res = await markAsRead({notificationId})
    if (res.code === 0) {
      message.success('已标记为已读')
      await loadNotifications()
      notificationStore.refreshUnreadCount()
    }
  } catch (err) {
    message.error('操作失败')
  }
}

// 标记所有通知为已读
async function handleMarkAllAsRead() {
  try {
    const res = await markAllAsRead()
    if (res.code === 0) {
      message.success('已全部标记为已读')
      await loadNotifications()
      notificationStore.refreshUnreadCount()
    }
  } catch (err) {
    message.error('操作失败')
  }
}

// 删除通知
async function handleDelete(notificationId) {
  try {
    const res = await deleteNotification({notificationId})
    if (res.code === 0) {
      message.success('删除成功')
      loadNotifications()
    }
  } catch (err) {
    message.error('删除失败')
  }
}

// 获取通知类型配置
async function loadNotificationTypes() {
  try {
    const res = await getNotificationTypes()
    if (res.code === 0) {
      notificationTypes.value = res.result.reduce((acc, type) => {
        acc[type.type] = type
        return acc
      }, {})
    }
  } catch (err) {
    console.error('Failed to load notification types:', err)
  }
}

const router = useRouter()

// 获取动作文本
function getActionText(eventType) {
  switch (eventType) {
    case 'comment':
      return '评论了文章'
    case 'reply':
      return '回复了你的评论'
    case 'follow':
      return '关注了你'
    default:
      return '发送了一条通知'
  }
}

// 跳转到文章
function goToArticle(articleId) {
  router.push({
    name: 'articlesDetail',
    query: {id: articleId}
  })
}

onMounted(() => {
  loadNotifications()
  loadNotificationTypes()
})

// 响应式布局
const isSmallScreen = ref(false)

function checkScreenSize() {
  isSmallScreen.value = window.innerWidth < 1200
}

onMounted(() => {
  checkScreenSize()
  window.addEventListener('resize', checkScreenSize)
})

onUnmounted(() => {
  window.removeEventListener('resize', checkScreenSize)
})
</script>

<template>
  <div class="notification-container">
    <n-flex :vertical="isSmallScreen" justify="center">
      <!-- 左侧菜单 -->
      <div class="menu-section">
        <n-card :bordered="false" style="position: sticky; top: 24px;">
          <n-flex vertical class="nav-buttons">
            <n-button class="nav-button" ghost :type="!showUnreadOnly?'info':'tertiary'"
                      @click="handleMenuSelect('all')">
              全部消息
            </n-button>
            <n-button class="nav-button" ghost :type="showUnreadOnly?'info':'tertiary'"
                      @click="handleMenuSelect('unread')">
              只看未读
            </n-button>
            <n-button
                type="warning"
                ghost
                @click="handleMarkAllAsRead"
                class="nav-button"
            >
              全部标记已读
            </n-button>
          </n-flex>
        </n-card>
      </div>

      <!-- 右侧内容 -->
      <div class="content-section">
        <n-card :bordered="false">
          <n-list :loading="loading">
            <template v-if="notifications.length > 0">
              <n-list-item v-for="notification in notifications" :key="notification.id">
                <n-space vertical :size="12">
                  <!-- 通知头部 -->
                  <n-space justify="space-between" align="center">
                    <n-space align="center">
                      <n-tag
                          :type="notificationTypes[notification.eventType]?.type || 'default'"
                          size="large"
                          round
                      >
                        {{ notificationTypes[notification.eventType]?.name || '通知' }}
                      </n-tag>
                    </n-space>
                    <n-space>
                      <n-time :time="new Date(notification.createdAt)"/>
                      <n-button
                          v-if="!notification.isRead"
                          text
                          size="tiny"
                          @click="handleMarkAsRead(notification.id)"
                      >
                        标记已读
                      </n-button>
                      <n-button
                          text
                          size="tiny"
                          @click="handleDelete(notification.id)"
                      >
                        删除
                      </n-button>
                    </n-space>
                  </n-space>

                  <!-- 通知内容 -->
                  <n-card size="small" class="notification-card">
                    <template #header>
                      <n-flex align="center" class="notification-header" style="align-items: center;">
                        <n-avatar
                            round
                            size="small"
                            :src="notification.payload.actorAvatarUrl || '/static/pic/default-avatar.png'"
                        />
                        <n-flex align="center">
                          <n-text strong>{{ notification.payload.actorName }}</n-text>
                          <n-text depth="3">
                            {{ getActionText(notification.eventType) }}
                          </n-text>
                          <n-button
                              v-if="notification.payload.articleId"
                              text
                              type="primary"
                              @click="goToArticle(notification.payload.articleId)"
                          >
                            《{{ notification.payload.articleTitle }}》
                          </n-button>
                        </n-flex>
                      </n-flex>
                    </template>

                    <div class="notification-content">
                      {{ notification.payload.content }}
                    </div>
                  </n-card>
                </n-space>
              </n-list-item>
            </template>
            <template v-else>
              <div class="empty-container">
                <n-empty description="暂无通知"/>
              </div>
            </template>
          </n-list>
        </n-card>
      </div>
    </n-flex>
  </div>
</template>

<style scoped>

.nav-buttons {
  min-width: 240px;
  gap: 8px;
}

.nav-button {
  width: 100%;
}

.notification-header {
  align-items: center; /* 确保头像和文字垂直居中 */
}

.notification-container {
  padding: 24px;
  min-height: 100%;
}

.menu-section {
  width: 240px;
  margin-right: 24px;
  flex-shrink: 0;
}

.menu-container {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

:deep(.notification-menu) {
  width: 100% !important;
}

:deep(.notification-menu .n-menu-item-content) {
  padding: 0 16px !important;
  width: 100% !important;
}

:deep(.notification-menu .n-menu-item) {
  width: 100% !important;
}

:deep(.n-menu) {
  --n-item-height: 40px !important;
}

.content-section {
  width: 800px;
  flex-shrink: 0;
}

.mark-all-button {
  width: 100%;
  margin-top: 16px;
}

/* 响应式布局 */
@media (max-width: 1200px) {
  .notification-container {
    padding: 16px;
  }

  .menu-section {
    width: 100%;
    margin-right: 0;
    margin-bottom: 16px;
  }

  .content-section {
    width: 100%;
    max-width: 900px;
    margin: 0 auto;
  }

  :deep(.n-menu.n-menu--horizontal) {
    width: 100% !important;
  }

  :deep(.n-menu.n-menu--horizontal .n-menu-item) {
    width: auto !important;
    flex: 1;
    min-width: 120px;
  }

  :deep(.n-menu.n-menu--horizontal .n-menu-item-content) {
    justify-content: center;
    white-space: nowrap;
  }

  .mark-all-button {
    margin-top: 12px;
  }
}


.notification-title {
  font-weight: 500;
}

.notification-content {
  color: var(--n-text-color-2);
  font-size: 14px;
  line-height: 1.6;
  padding: 8px 0;
}

.notification-card {
  background-color: var(--n-card-color);
  transition: box-shadow .3s ease;
}

.notification-card:hover {
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
}

.notification-content {
  margin-top: 8px;
  padding: 8px;
  background-color: var(--n-color-hover);
  border-radius: 4px;
  font-size: 14px;
  line-height: 1.6;
}

.n-tag {
  padding: 0 12px;
}

.empty-container {
  min-height: 400px;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: var(--n-card-color);
  border-radius: 8px;
  margin: 16px 0;
}
</style>
