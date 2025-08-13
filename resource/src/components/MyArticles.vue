<script setup lang="ts">
import {getUserArticles} from "@/utils/gooseForumService.ts";
import {onMounted, ref, computed} from "vue";
import type {ArticleListItem, UserInfo} from "@/utils/gooseForumInterfaces.ts";
import { notification } from "@/utils/notification";

const props = defineProps<{
  userInfo: UserInfo
}>()

const emit = defineEmits<{
  userInfoUpdated: []
}>()

onMounted(() => {
  refreshUserArticles()
})
let listItem = ref<ArticleListItem[]>([])
let page = ref({
  page: 1,
  size: 10,
  total: 0,
})
let loading = ref(false)

// 计算总页数
const totalPages = computed(() => {
  return Math.ceil(page.value.total / page.value.size)
})

// 计算显示的页码范围
const visiblePages = computed(() => {
  const current = page.value.page
  const total = totalPages.value
  const pages = []

  if (total <= 5) {
    // 如果总页数小于等于5，显示所有页码
    for (let i = 1; i <= total; i++) {
      pages.push(i)
    }
  } else {
    // 如果总页数大于5，显示当前页前后各2页
    let start = Math.max(1, current - 2)
    let end = Math.min(total, current + 2)

    // 确保显示5个页码
    if (end - start < 4) {
      if (start === 1) {
        end = Math.min(total, start + 4)
      } else {
        start = Math.max(1, end - 4)
      }
    }

    for (let i = start; i <= end; i++) {
      pages.push(i)
    }
  }

  return pages
})

// 是否可以上一页
const canPrevious = computed(() => {
  return page.value.page > 1
})

// 是否可以下一页
const canNext = computed(() => {
  return page.value.page < totalPages.value
})

async function refreshUserArticles() {
  if (loading.value) return

  loading.value = true
  try {
    let resp = await getUserArticles(page.value.page, page.value.size)
    listItem.value = resp.result.list
    page.value = {
      page: resp.result.page,
      size: resp.result.size,
      total: resp.result.total,
    }
    // 显示成功通知
    if (resp.result.list.length === 0) {
      notification.info('暂无文章数据')
    }
  } catch (error) {
    console.error('获取用户文章失败:', error)
    notification.error('获取文章列表失败，请稍后重试')
  } finally {
    loading.value = false
  }
}

// 跳转到指定页
async function goToPage(pageNum: number) {
  if (pageNum < 1 || pageNum > totalPages.value || pageNum === page.value.page || loading.value) {
    return
  }

  page.value.page = pageNum
  await refreshUserArticles()
}

// 上一页
async function previousPage() {
  if (canPrevious.value) {
    await goToPage(page.value.page - 1)
  }
}

// 下一页
async function nextPage() {
  if (canNext.value) {
    await goToPage(page.value.page + 1)
  }
}

// 格式化日期
function formatDate(dateString: string) {
  if (!dateString) return ''

  const date = new Date(dateString)
  const now = new Date()
  const diff = now.getTime() - date.getTime()

  // 小于1分钟
  if (diff < 60 * 1000) {
    return '刚刚'
  }

  // 小于1小时
  if (diff < 60 * 60 * 1000) {
    const minutes = Math.floor(diff / (60 * 1000))
    return `${minutes}分钟前`
  }

  // 小于1天
  if (diff < 24 * 60 * 60 * 1000) {
    const hours = Math.floor(diff / (60 * 60 * 1000))
    return `${hours}小时前`
  }

  // 小于7天
  if (diff < 7 * 24 * 60 * 60 * 1000) {
    const days = Math.floor(diff / (24 * 60 * 60 * 1000))
    return `${days}天前`
  }

  // 超过7天，显示具体日期
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hour = String(date.getHours()).padStart(2, '0')
  const minute = String(date.getMinutes()).padStart(2, '0')

  // 如果是今年，不显示年份
  if (year === now.getFullYear()) {
    return `${month}-${day} ${hour}:${minute}`
  }

  return `${year}-${month}-${day} ${hour}:${minute}`
}

</script>
<template>
  <div class="card bg-base-100 shadow-sm">
    <div class="flex justify-between items-center px-4 py-3 border-b border-base-300">
      <h2 class="text-2xl font-normal">我的文章</h2>
      <a href="/publish" class="btn btn-primary btn-sm">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24"
             stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
        </svg>
        写文章
      </a>
    </div>
    <div class="card-body p-0">
      <!-- 加载状态 -->
      <div v-if="loading && listItem.length === 0" class="p-8 text-center">
        <div class="loading loading-spinner loading-lg text-primary"></div>
        <p class="mt-4 text-base-content/60">正在加载文章...</p>
      </div>

      <!-- 空状态 -->
      <div v-else-if="!loading && listItem.length === 0" class="p-8 text-center">
        <div class="text-6xl mb-4">📝</div>
        <p class="text-lg font-normal text-base-content/80 mb-2">还没有发布任何文章</p>
        <p class="text-base-content/60">开始创作你的第一篇文章吧！</p>
      </div>

      <!-- 文章列表 -->
      <ul class="list" v-else>
        <li class="list-row hover:bg-base-200 flex items-center gap-3 px-4 py-2"
            :class="{ 'opacity-50': loading }"
            v-for="item in listItem"
            :key="item.id">
          <!-- 左侧头像 -->
          <a class="avatar" href="">
            <div class="w-10 rounded-full">
              <img :src="props.userInfo.avatarUrl" alt=".Username"/>
            </div>
          </a>
          <!-- 右侧内容 -->
          <div class="flex-1">
            <!-- 标题行 -->
            <div class="flex items-center gap-2 mb-1">
              <div class="badge badge-sm badge-primary flex-shrink-0">{{item.typeStr}}</div>
              <a :href="'/post/'+item.id"
                 class="text-lg font-normal text-base-content hover:text-primary hover:underline flex-1 min-w-0">{{
                  item.title
                }}</a>
            </div>
            <!-- 用户信息行和统计信息合并为一行 -->
            <div class="flex items-center justify-between text-sm text-base-content/60">
              <div class="flex items-center flex-wrap">
                <a :href="'/user/'+props.userInfo.userId" class="mr-3">{{ props.userInfo.username }}</a>
                <span class="mr-3">{{ formatDate(item.createTime) }}</span>
                <span class="badge badge-sm badge-ghost mr-1" v-for="cItem in item.categories" :key="cItem">{{ cItem }}</span>
              </div>
              <div class="flex items-center">
                <div class="flex items-center mr-4">
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1"
                       fill="none"
                       viewBox="0 0 24 24"
                       stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round"
                          stroke-width="2"
                          d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
                    <path stroke-linecap="round" stroke-linejoin="round"
                          stroke-width="2"
                          d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"/>
                  </svg>
                  <span class="flex-shrink-0">{{ item.viewCount }}</span>
                </div>
                <div class="flex items-center mr-4">
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1"
                       fill="none"
                       viewBox="0 0 24 24"
                       stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round"
                          stroke-width="2"
                          d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z"/>
                  </svg>
                  <span class="flex-shrink-0"> {{ item.commentCount }} </span>
                </div>
                <!-- 编辑按钮 -->
                <a :href="'/publish?id=' + item.id"
                   class="btn btn-xs btn-ghost hover:btn-primary flex items-center gap-1"
                   title="编辑文章">
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3"
                       fill="none"
                       viewBox="0 0 24 24"
                       stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round"
                          stroke-width="2"
                          d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/>
                  </svg>
                  <span class="hidden sm:inline">编辑</span>
                </a>
              </div>
            </div>
          </div>
        </li>
      </ul>
    </div>
  </div>
  <!-- 分页 -->
  <div class="flex justify-center mt-8" v-if="totalPages > 1">
    <div class="join bg-base-100 rounded-lg shadow-sm">
      <!-- 上一页按钮 -->
      <button
        class="join-item btn btn-sm"
        :class="{
          'bg-base-100 border-base-300': !canPrevious || loading,
          'btn-disabled': !canPrevious || loading
        }"
        @click="previousPage"
        :disabled="!canPrevious || loading"
      >
        <span v-if="loading" class="loading loading-spinner loading-xs"></span>
        <span v-else>«</span>
      </button>

      <!-- 页码按钮 -->
      <button
        v-for="pageNum in visiblePages"
        :key="pageNum"
        class="join-item btn btn-sm"
        :class="{
          'bg-primary text-primary-content border-primary': pageNum === page.page && !loading,
          'bg-base-100 border-base-300': pageNum !== page.page || loading,
          'btn-disabled': loading
        }"
        @click="goToPage(pageNum)"
        :disabled="loading"
      >
        {{ pageNum }}
      </button>

      <!-- 下一页按钮 -->
      <button
        class="join-item btn btn-sm"
        :class="{
          'bg-base-100 border-base-300': !canNext || loading,
          'btn-disabled': !canNext || loading
        }"
        @click="nextPage"
        :disabled="!canNext || loading"
      >
        <span v-if="loading" class="loading loading-spinner loading-xs"></span>
        <span v-else>»</span>
      </button>
    </div>
  </div>

  <!-- 分页信息 -->
  <div class="flex justify-center mt-4 text-sm text-base-content/60" v-if="totalPages > 0">
    <span>第 {{ page.page }} 页，共 {{ totalPages }} 页，总计 {{ page.total }} 篇文章</span>
  </div>
</template>
