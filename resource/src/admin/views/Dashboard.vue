<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-normal text-base-content">仪表盘</h1>
        <p class="text-base-content/70 mt-1">欢迎回来，{{ authStore.user?.username }}！</p>
      </div>
      <div class="text-sm text-base-content/60">
        最后登录：{{ formatDate(new Date()) }}
      </div>
    </div>

    <!-- 图表和统计 -->
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-6">
      <!-- 访问统计图表 -->
      <div class="lg:col-span-9">
        <div class="card bg-base-100 shadow">
          <div class="card-body">
            <h2 class="card-title font-normal">访问统计</h2>
            <div class="h-64 flex items-center justify-center border border-base-300 rounded">
              <div class="text-center">
                <ChartBarIcon class="w-16 h-16 mx-auto text-base-300 mb-4" />
                <p class="text-base-content/50">图表组件待集成</p>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 系统统计 -->
      <div class="lg:col-span-3">
        <div class="card bg-base-100 shadow">
          <div class="card-body">
            <h2 class="card-title font-normal">系统统计</h2>
            <div class="space-y-2 h-64">
              <!-- 用户统计 -->
              <div class="flex items-center justify-between p-2 border border-base-300 rounded">
                <div class="flex items-center space-x-2">
                  <UsersIcon class="w-4 h-4 text-primary" />
                  <div>
                    <p class="text-xs text-base-content/70">用户</p>
                    <p class="text-sm font-bold text-base-content">{{ formatNumber(stats.userCount) }}</p>
                  </div>
                </div>
                <p class="text-xs text-base-content/50">+{{ stats.userMonthCount }}</p>
              </div>

              <!-- 帖子统计 -->
              <div class="flex items-center justify-between p-2 border border-base-300 rounded">
                <div class="flex items-center space-x-2">
                  <DocumentTextIcon class="w-4 h-4 text-secondary" />
                  <div>
                    <p class="text-xs text-base-content/70">帖子</p>
                    <p class="text-sm font-bold text-base-content">{{ formatNumber(stats.articleCount) }}</p>
                  </div>
                </div>
                <p class="text-xs text-base-content/50">+{{ stats.articleMonthCount }}</p>
              </div>

              <!-- 评论统计 -->
              <div class="flex items-center justify-between p-2 border border-base-300 rounded">
                <div class="flex items-center space-x-2">
                  <ChatBubbleLeftRightIcon class="w-4 h-4 text-accent" />
                  <div>
                    <p class="text-xs text-base-content/70">评论</p>
                    <p class="text-sm font-bold text-base-content">{{ formatNumber(stats.reply) }}</p>
                  </div>
                </div>
              </div>

              <!-- 访问统计 -->
              <div class="flex items-center justify-between p-2 border border-base-300 rounded">
                <div class="flex items-center space-x-2">
                  <EyeIcon class="w-4 h-4 text-info" />
                  <div>
                    <p class="text-xs text-base-content/70">访问</p>
                    <p class="text-sm font-bold text-base-content">x</p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 系统状态和快捷操作 -->
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
      <!-- 快捷操作 -->
      <div class="card bg-base-100 shadow lg:col-span-4 h-fit">
        <div class="card-body">
          <h2 class="card-title font-normal">快捷操作</h2>
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

      <!-- 项目版本信息 -->
      <div class="card bg-base-100 shadow lg:col-span-8">
        <div class="card-body">
          <div class="flex items-center justify-between mb-4">
            <h2 class="card-title flex items-center gap-2 font-normal">
              <CodeBracketIcon class="w-5 h-5" />
              项目版本
            </h2>
            <a
              href="https://github.com/leancodebox/GooseForum/releases"
              target="_blank"
              rel="noopener noreferrer"
              class="btn btn-ghost btn-sm"
            >
              <ArrowTopRightOnSquareIcon class="w-4 h-4" />
              查看全部
            </a>
          </div>

          <!-- GitHub Releases 信息 -->
          <div class="h-80 border border-base-300 rounded-lg overflow-hidden bg-base-50">
            <div v-if="loadingReleases" class="flex items-center justify-center h-full">
              <span class="loading loading-spinner loading-md"></span>
              <span class="ml-2">加载中...</span>
            </div>

            <div v-else-if="releases.length > 0" class="p-4 h-full overflow-y-auto">
              <div v-for="release in releases.slice(0, 3)" :key="release.id" class="mb-4 last:mb-0">
                <div class="flex items-start justify-between">
                  <div class="flex-1">
                    <h3 class="font-semibold text-sm flex items-center gap-2">
                      <TagIcon class="w-4 h-4 text-primary" />
                      {{ release.tag_name }}
                      <span v-if="release.prerelease" class="badge badge-warning badge-xs">预发布</span>
                      <span v-else-if="release.draft" class="badge badge-ghost badge-xs">草稿</span>
                    </h3>
                    <p class="text-xs text-base-content/70 mt-1">{{ formatRelativeTime(new Date(release.published_at)) }}</p>
                    <p class="text-sm mt-2 line-clamp-2">{{ release.body || '暂无发布说明' }}</p>
                  </div>
                  <a
                    :href="release.html_url"
                    target="_blank"
                    rel="noopener noreferrer"
                    class="btn btn-ghost btn-xs ml-2"
                  >
                    <ArrowTopRightOnSquareIcon class="w-3 h-3" />
                  </a>
                </div>
                <div class="divider my-2 last:hidden"></div>
              </div>
            </div>

            <div v-else class="flex flex-col items-center justify-center h-full text-base-content/50">
              <CodeBracketIcon class="w-12 h-12 mb-2" />
              <p class="text-sm">暂无发布信息</p>
              <a
                href="https://github.com/leancodebox/GooseForum/releases"
                target="_blank"
                rel="noopener noreferrer"
                class="btn btn-ghost btn-sm mt-2"
              >
                查看GitHub页面
              </a>
            </div>
          </div>

          <!-- 备用信息显示 -->
          <div class="mt-4 text-sm text-base-content/70">
            <div class="flex items-center gap-2">
              <div class="w-2 h-2 bg-success rounded-full"></div>
              <span>当前版本：{{ currentVersion }}</span>
            </div>
            <div class="mt-1">
              <span>最后更新：{{ lastUpdateTime }}</span>
            </div>
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
  CogIcon,
  CodeBracketIcon,
  ArrowTopRightOnSquareIcon
} from '@heroicons/vue/24/outline'
import { api } from '../utils/axiosInstance'
import { getSiteStatistics } from '../utils/adminService'

const authStore = useAuthStore()

// 统计数据
const stats = ref({
  userCount: 0,
  userMonthCount: 0,
  articleCount: 0,
  articleMonthCount: 0,
  reply: 0,
  linksCount: 0,
})

// 版本信息
const currentVersion = ref('v1.0.0')
const lastUpdateTime = ref('2024-01-01')
const releases = ref([])
const loadingReleases = ref(false)

// 获取GitHub Releases
const fetchReleases = async () => {
  loadingReleases.value = true
  try {
    const response = await fetch('https://api.github.com/repos/leancodebox/GooseForum/releases')
    if (response.ok) {
      const data = await response.json()
      releases.value = data
      if (data.length > 0) {
        currentVersion.value = data[0].tag_name
        lastUpdateTime.value = new Date(data[0].published_at).toLocaleDateString('zh-CN')
      }
    }
  } catch (error) {
    console.error('获取releases失败:', error)
  } finally {
    loadingReleases.value = false
  }
}



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



// 组件挂载时获取数据
onMounted(() => {
  fetchStats()
  fetchReleases()
})
</script>

<style scoped>
</style>
