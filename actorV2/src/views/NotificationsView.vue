<script setup lang="ts">
import {onMounted, ref} from 'vue';
import {useRouter} from 'vue-router';
import {NButton, NCard, NEmpty, NPagination, NSpace, NTab, NTabs} from 'naive-ui';
import {getNotificationList, markAllAsRead, markAsRead} from "@/utils/articleService.ts";
import type {Notifications} from '@/types/articleInterfaces';

const router = useRouter();
const notifications = ref<Notifications[]>([]);
const loading = ref(false);
const currentPage = ref(1);
const pageSize = ref(10);
const total = ref(0);
const activeTab = ref('unread'); // 'all' 或 'unread'

const loadNotifications = async () => {
  loading.value = true;
  try {
    const response = await getNotificationList(
        currentPage.value,
        pageSize.value,
        activeTab.value === 'unread'
    );

    if (response.code === 0) {
      notifications.value = response.result.list;
      total.value = response.result.total;
    } else {
      console.error('加载通知失败:', response.message);
    }
  } catch (error) {
    console.error('加载通知失败:', error);
  } finally {
    loading.value = false;
  }
};

const handleTabChange = (tabName: string) => {
  activeTab.value = tabName;
  currentPage.value = 1; // 切换 tab 时重置页码
  loadNotifications();
};

const handleMarkAsRead = async (id: number) => {
  try {
    const response = await markAsRead(id);
    if (response.code === 0) {
      await loadNotifications();
    } else {
      console.error('标记已读失败:', response.message);
    }
  } catch (error) {
    console.error('标记已读失败:', error);
  }
};

const handleMarkAllAsRead = async () => {
  try {
    const response = await markAllAsRead();
    if (response.code === 0) {
      await loadNotifications();
    } else {
      console.error('标记全部已读失败:', response.message);
    }
  } catch (error) {
    console.error('标记全部已读失败:', error);
  }
};

const handlePageChange = (page: number) => {
  currentPage.value = page;
  loadNotifications();
};

const getNotificationIcon = (type: string) => {
  switch (type) {
    case 'comment':
      return '💬';
    case 'reply':
      return '❤️';
    case 'system':
      return '🔔';
    case 'follow':
      return '👥';
    default:
      return '📌';
  }
};

onMounted(() => {
  loadNotifications();
});

function openPost(url: string) {
  window.open(url, '_blank');
}
</script>

<template>
  <div class="notifications-page">
    <div class="notifications-header">
      <h1>消息通知</h1>
    </div>

    <div class="notifications-actions">
      <NTabs v-model:value="activeTab" @update:value="handleTabChange">
        <NTab name="unread" tab="未读消息"/>
        <NTab name="all" tab="全部消息"/>
        <template #suffix>
          <NButton type="primary" size="small" @click="handleMarkAllAsRead">
            全部标记为已读
          </NButton>
        </template>
      </NTabs>
    </div>

    <div class="notifications-list" v-if="notifications.length > 0">
      <NCard v-for="notification in notifications"
             :key="notification.id"
             :class="{ 'unread': !notification.isRead, 'read': notification.isRead }">
        <div class="notification-item">
          <div class="notification-icon">
            {{ getNotificationIcon(notification.eventType) }}
          </div>
          <div class="notification-content">
            <h3>{{ notification.payload.title }}</h3>
            <p>{{notification.payload.actorName}} 对  <a :href="`/post/${notification.payload.articleId}`">{{ notification.payload.articleTitle }}</a> 评论： {{ notification.payload.content }}</p>

            <span class="notification-time">{{ notification.createdAt }}</span>
          </div>
          <div class="notification-actions">
            <NSpace>
              <NButton v-if="!notification.isRead"
                       size="small"
                       @click="handleMarkAsRead(notification.id)">
                标记已读
              </NButton>
              <NButton v-if="notification.payload.articleId"
                       size="small"
                       @click="openPost(`/post/${notification.payload.articleId}`)">
                查看详情
              </NButton>
            </NSpace>
          </div>
        </div>
      </NCard>

      <div class="pagination-wrapper" v-if="total > pageSize">
        <NPagination
            v-model:page="currentPage"
            :page-size="pageSize"
            :item-count="total"
            :page-slot="5"
            size="medium"
            @update:page="handlePageChange"
        />
      </div>
    </div>

    <NEmpty v-else description="暂无通知消息"/>
  </div>
</template>

<style scoped>
.notifications-page {
  max-width: 700px;
  margin: 0 auto;
  padding: 20px;
}

.notifications-header {
  margin-bottom: 16px;
}

.notifications-header h1 {
  margin: 0;
  font-size: 1.5rem;
}

.notifications-actions {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: 16px;
}

/* 移除重复的 tabs 样式 */
.n-tabs {
  margin-bottom: 0;
  padding-bottom: 0;
  border-bottom: none;
}

.n-tabs {
  margin-bottom: 16px;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--border-color);
}

.notifications-list {
  margin-top: 12px;
}

.n-card {
  margin-bottom: 4px;
  transition: background-color 0.2s;
}

.notification-item {
  display: flex;
  gap: 8px;
  align-items: flex-start;
}

.read {
  opacity: 0.8;
  background-color: var(--card-bg);
}

.unread {
  background-color: var(--color-background-soft);
  border-left: 2px solid var(--primary-color);
}

.notification-icon {
  font-size: 1.2rem;
  min-width: 24px;
}

.notification-content {
  flex: 1;
  padding: 0;
}

.notification-content h3 {
  margin: 0 0 2px 0;
  font-size: 0.95rem;
  line-height: 1.4;
}

.notification-content p {
  margin: 0;
  line-height: 1.4;
}

.notification-time {
  margin-top: 2px;
  font-size: 0.8rem;
}

.notification-actions {
  margin-left: 8px;
  padding-left: 8px;
}

.pagination-wrapper {
  margin-top: 16px;
  padding: 8px;
}
</style>
