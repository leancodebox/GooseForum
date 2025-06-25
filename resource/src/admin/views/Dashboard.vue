<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-base-content">仪表盘</h1>
        <p class="text-base-content/70 mt-1">欢迎回来，{{ authStore.user?.username }}！</p>
      </div>
      <div class="text-sm text-base-content/60">
        最后登录：{{ formatDate(new Date()) }}
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
      <div class="stats shadow bg-base-100">
        <div class="stat">
          <div class="stat-figure text-primary">
            <UsersIcon class="w-8 h-8" />
          </div>
          <div class="stat-title">总用户数</div>
          <div class="stat-value text-primary">{{ stats.userCount }}</div>
          <div class="stat-desc">↗︎ 本月新增 {{ stats.userMonthCount }}</div>
        </div>
      </div>

      <div class="stats shadow bg-base-100">
        <div class="stat">
          <div class="stat-figure text-secondary">
            <DocumentTextIcon class="w-8 h-8" />
          </div>
          <div class="stat-title">总帖子数</div>
          <div class="stat-value text-secondary">{{ stats.articleCount }}</div>
          <div class="stat-desc">↗︎ 本月新增 {{ stats.articleMonthCount }}</div>
        </div>
      </div>

      <div class="stats shadow bg-base-100">
        <div class="stat">
          <div class="stat-figure text-accent">
            <ChatBubbleLeftRightIcon class="w-8 h-8" />
          </div>
          <div class="stat-title">总评论数</div>
          <div class="stat-value text-accent">{{ stats.reply }}</div>
          <div class="stat-desc">↗︎ 今日新增 x </div>
        </div>
      </div>

      <div class="stats shadow bg-base-100">
        <div class="stat">
          <div class="stat-figure text-info">
            <EyeIcon class="w-8 h-8" />
          </div>
          <div class="stat-title">总访问量</div>
          <div class="stat-value text-info">x </div>
          <div class="stat-desc">↗︎ 今日访问 x </div>
        </div>
      </div>
    </div>

    <!-- 图表和活动 -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- 访问趋势图 -->
      <div class="lg:col-span-2">
        <div class="card bg-base-100 shadow">
          <div class="card-body">
            <h2 class="card-title">访问趋势</h2>
            <div class="h-64 flex items-center justify-center border-2 border-dashed border-base-300 rounded-lg">
              <div class="text-center text-base-content/50">
                <ChartBarIcon class="w-12 h-12 mx-auto mb-2" />
                <p>图表组件待集成</p>
                <p class="text-sm">可集成 Chart.js 或 ECharts</p>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 最近活动 -->
      <div class="card bg-base-100 shadow">
        <div class="card-body">
          <h2 class="card-title">最近活动</h2>
          <div class="space-y-4 max-h-64 overflow-y-auto">
            <div v-for="(activity, index) in recentActivities" :key="index" class="flex items-start gap-3">
              <div class="avatar placeholder">
                <div class="bg-neutral text-neutral-content rounded-full w-8 h-8">
                  <span class="text-xs">{{ activity.user.charAt(0).toUpperCase() }}</span>
                </div>
              </div>
              <div class="flex-1 min-w-0">
                <p class="text-sm font-medium text-base-content truncate">{{ activity.user }}</p>
                <p class="text-xs text-base-content/70 truncate">{{ activity.action }}</p>
                <p class="text-xs text-base-content/50">{{ formatRelativeTime(activity.time) }}</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 系统状态和快捷操作 -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- 系统状态 -->
      <div class="card bg-base-100 shadow">
        <div class="card-body">
          <h2 class="card-title">系统状态</h2>
          <div class="space-y-4">
            <div class="flex items-center justify-between">
              <span class="text-sm font-medium">服务器状态</span>
              <div class="badge badge-success gap-2">
                <div class="w-2 h-2 bg-success rounded-full"></div>
                正常运行
              </div>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-sm font-medium">数据库连接</span>
              <div class="badge badge-success gap-2">
                <div class="w-2 h-2 bg-success rounded-full"></div>
                连接正常
              </div>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-sm font-medium">内存使用</span>
              <div class="flex items-center gap-2">
                <progress class="progress progress-primary w-20" value="65" max="100"></progress>
                <span class="text-sm">65%</span>
              </div>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-sm font-medium">磁盘使用</span>
              <div class="flex items-center gap-2">
                <progress class="progress progress-warning w-20" value="78" max="100"></progress>
                <span class="text-sm">78%</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 快捷操作 -->
      <div class="card bg-base-100 shadow">
        <div class="card-body">
          <h2 class="card-title">快捷操作</h2>
          <div class="grid grid-cols-2 gap-4">
            <router-link to="/admin/posts" class="btn btn-outline btn-primary">
              <DocumentTextIcon class="w-4 h-4" />
              管理帖子
            </router-link>
            <router-link to="/admin/users" class="btn btn-outline btn-secondary">
              <UsersIcon class="w-4 h-4" />
              管理用户
            </router-link>
            <router-link to="/admin/categories" class="btn btn-outline btn-accent">
              <TagIcon class="w-4 h-4" />
              管理分类
            </router-link>
            <router-link to="/admin/settings" class="btn btn-outline btn-info">
              <CogIcon class="w-4 h-4" />
              系统设置
            </router-link>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import {
  UsersIcon,
  DocumentTextIcon,
  ChatBubbleLeftRightIcon,
  EyeIcon,
  ChartBarIcon,
  TagIcon,
  CogIcon
} from '@heroicons/vue/24/outline'
import { api } from '../utils/axiosInstance'
import { getSiteStatistics } from '../utils/adminService'

const authStore = useAuthStore()

// 统计数据
const stats = ref({
  userCount: 12,
  userMonthCount: 4,
  articleCount: 199,
  articleMonthCount: 10,
  reply: 14,
  linksCount: 7,
})

// 最近活动
const recentActivities = ref([
  {
    user: 'admin',
    action: '发布了新帖子《Vue3 最佳实践》',
    time: new Date(Date.now() - 10 * 60 * 1000)
  },
  {
    user: 'user123',
    action: '注册了新账户',
    time: new Date(Date.now() - 30 * 60 * 1000)
  },
  {
    user: 'moderator',
    action: '审核通过了帖子《Go语言入门》',
    time: new Date(Date.now() - 60 * 60 * 1000)
  },
  {
    user: 'editor',
    action: '更新了分类设置',
    time: new Date(Date.now() - 2 * 60 * 60 * 1000)
  },
  {
    user: 'user456',
    action: '发表了评论',
    time: new Date(Date.now() - 3 * 60 * 60 * 1000)
  }
])

// 格式化数字
const formatNumber = (num: number) => {
  if (num >= 1000000) {
    return (num / 1000000).toFixed(1) + 'M'
  } else if (num >= 1000) {
    return (num / 1000).toFixed(1) + 'K'
  }
  return num.toString()
}

// 格式化日期
const formatDate = (date: Date) => {
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

// 格式化相对时间
const formatRelativeTime = (date: Date) => {
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const minutes = Math.floor(diff / (1000 * 60))
  const hours = Math.floor(diff / (1000 * 60 * 60))
  const days = Math.floor(diff / (1000 * 60 * 60 * 24))

  if (minutes < 1) {
    return '刚刚'
  } else if (minutes < 60) {
    return `${minutes}分钟前`
  } else if (hours < 24) {
    return `${hours}小时前`
  } else {
    return `${days}天前`
  }
}

// 获取统计数据
const fetchStats = async () => {
  try {
    const response = await getSiteStatistics()
    stats.value = response.result
  } catch (error) {
    console.error('获取统计数据失败:', error)
  }
}

// 获取最近活动
const fetchRecentActivities = async () => {
  try {
    
  } catch (error) {
    console.error('获取最近活动失败:', error)
  }
}

// 组件挂载时获取数据
onMounted(() => {
  fetchStats()
  fetchRecentActivities()
})
</script>

<style scoped>
/* 自定义样式 */
.stats {
  border: 1px solid hsl(var(--bc) / 0.1);
}

.card {
  border: 1px solid hsl(var(--bc) / 0.1);
}

.progress {
  height: 0.5rem;
}

/* 悬停效果 */
.btn:hover {
  transform: translateY(-1px);
  transition: transform 0.2s ease;
}

/* 活动列表滚动条 */
.overflow-y-auto::-webkit-scrollbar {
  width: 4px;
}

.overflow-y-auto::-webkit-scrollbar-track {
  background: hsl(var(--b2));
}

.overflow-y-auto::-webkit-scrollbar-thumb {
  background: hsl(var(--bc) / 0.3);
  border-radius: 2px;
}
</style>