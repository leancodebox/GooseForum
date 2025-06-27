<script setup lang="ts">
import {computed, onMounted, reactive, ref} from 'vue'
import {EllipsisVerticalIcon, PlusIcon} from '@heroicons/vue/24/outline'
import {editArticle, getAdminArticlesList, getArticleEnum} from "@/admin/utils/adminService.ts";
import type {AdminArticlesItem, Label} from "@/admin/utils/adminInterfaces.ts";

const posts = ref<AdminArticlesItem[]>([])
const loading = ref(false)
const searchQuery = ref('')
const typeLabel = ref<Label[]>([])

// 筛选条件
const filters = reactive({
  category: '',
  status: '',
  dateRange: '',
  sortBy: 'created_at'
})

// 分页
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

// 计算属性
const totalPages = computed(() => {
  return Math.ceil(pagination.total / pagination.pageSize)
})

const visiblePages = computed(() => {
  const current = pagination.page
  const total = totalPages.value
  const pages = []
  let start = Math.max(1, current - 2)
  let end = Math.min(total, current + 2)
  for (let i = start; i <= end; i++) {
    pages.push(i)
  }
  return pages
})

// 方法
const fetchPosts = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      pageSize: pagination.pageSize,
      search: searchQuery.value,
      ...filters
    }
    const response = await getAdminArticlesList(pagination.page, pagination.pageSize)
    posts.value = response.result.list
    pagination.total = response.result.total
  } catch (error) {
    console.error('获取帖子列表失败:', error)
    // 使用模拟数据
    pagination.total = 0
  } finally {
    loading.value = false
  }
}


const changePage = (page: number) => {
  pagination.page = page
  fetchPosts()
}

const toggleTop = async (post: AdminArticlesItem) => {
  alert("todo")
}

const toggleRecommend = async (post: AdminArticlesItem) => {
  alert("todo")
}

const approvePost = async (post: AdminArticlesItem) => {
  try {
    const resp = await editArticle(post.id, 0)
    await fetchPosts()
  } catch (error) {
    console.error('恢复:', error)
  }
}

const rejectPost = async (post: AdminArticlesItem) => {
  try {
    const resp = await editArticle(post.id, 1)
    await fetchPosts()
  } catch (error) {
    console.error('封禁:', error)
  }
}


// 工具函数
const getStatusBadgeClass = (status: number) => {
  const classes = {
    0: 'badge-info',
    1: 'badge-success',
  }
  return classes[status as keyof typeof classes] || 'badge-ghost'
}

const getStatusText = (status: number) => {
  const texts = {
    0: '草稿',
    1: '发布'
  }
  return texts[status as keyof typeof texts] || status
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const getTypeLabel = (typeValue: number) => {
  const typeItem = typeLabel.value.find(item => item.value === typeValue)
  console.log(typeLabel.value,typeValue)
  return typeItem?.name??'文章'
}

async function fetchLabel() {
  const resp = await getArticleEnum()
  typeLabel.value = resp.result.type
}

// 组件挂载时获取数据
onMounted(() => {
  fetchPosts()
  fetchLabel()
})

</script>

<template>
  <div class="space-y-6">
    <!-- 页面标题和操作 -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold text-base-content">帖子管理</h1>
        <p class="text-base-content/70 mt-1">管理系统中的所有帖子内容</p>
      </div>
      <div class="flex gap-2">
        <button class="btn btn-outline btn-sm" @click="fetchPosts">
          刷新
        </button>
        <a href="/publish" class="btn btn-primary btn-sm">
          <PlusIcon class="w-4 h-4"/>
          发布帖子
        </a>
      </div>
    </div>

    <!-- 帖子列表 -->
    <div class="card bg-base-100 shadow">
      <div class="card-body p-0">
        <div class="overflow-x-auto">
          <table class="table table-zebra">
            <thead>
            <tr>
              <th>帖子信息</th>
              <th>作者</th>
              <th>分类</th>
              <th>状态</th>
              <th>统计</th>
              <th>发布时间</th>
              <th>操作</th>
            </tr>
            </thead>
            <tbody>
            <tr v-for="post in posts" :key="post.id">
              <td>
                <div class="flex items-start gap-3">
                  <div class="avatar" v-if="post.userAvatarUrl">
                    <div class="mask mask-squircle w-12 h-12">
                      <img :src="post.userAvatarUrl" :alt="post.username"/>
                    </div>
                  </div>
                  <div class="flex-1 min-w-0">
                    <div class="font-bold text-sm line-clamp-2">{{ post.title }}</div>
                    <div class="text-xs text-base-content/70 line-clamp-1 mt-1">
                      {{ post.description || '暂无摘要' }}
                    </div>
                    <div class="flex items-center gap-2 mt-1">
                      <div class="badge badge-xs" v-if="true">
                        置顶
                      </div>
                      <div class="badge badge-xs badge-accent" v-if="true">
                        热门
                      </div>
                      <div class="badge badge-xs badge-info" v-if="true">
                        推荐
                      </div>
                    </div>
                  </div>
                </div>
              </td>
              <td>
                <div class="flex items-center gap-2">
                  <div class="avatar">
                    <div class="w-8 h-8 rounded-full">
                      <img :src="post.userAvatarUrl || '/static/pic/default-avatar.png'" :alt="post.username"/>
                    </div>
                  </div>
                  <div class="text-sm">
                    <div class="font-medium">{{ post.username }}</div>
                  </div>
                </div>
              </td>
              <td>
                <div class="badge badge-outline badge-sm whitespace-nowrap">{{ getTypeLabel(post.type) }}</div>
              </td>
              <td>
                <div class="badge badge-sm whitespace-nowrap" :class="getStatusBadgeClass(post.articleStatus)">
                  {{ getStatusText(post.articleStatus) }}
                </div>
                <div class="badge badge-sm whitespace-nowrap badge-error" v-if="post.processStatus==1">
                  封禁
                </div>
              </td>
              <td>
                <div class="text-xs space-y-1">
                  <div>浏览: {{ post.viewCount }}</div>
                  <div>评论: {{ post.replyCount }}</div>
                  <div>点赞: {{ post.likeCount }}</div>
                </div>
              </td>
              <td class="text-xs">{{ formatDate(post.createdAt) }}</td>
              <td>
                <div class="dropdown dropdown-end">
                  <div tabindex="0" role="button" class="btn btn-ghost btn-xs">
                    <EllipsisVerticalIcon class="w-4 h-4"/>
                  </div>
                  <ul tabindex="0" class="dropdown-content z-[1] menu p-2 shadow bg-base-100 rounded-box w-52">
                    <li><a :href="`/post/${post.id}`" target="_blank">查看</a></li>
                    <li><a @click="toggleTop(post)">{{ false ? '取消置顶' : '置顶' }}</a></li>
                    <li><a @click="toggleRecommend(post)">{{ false ? '取消推荐' : '推荐' }}</a></li>
                    <li v-if="post.processStatus === 1">
                      <a @click="approvePost(post)" class="text-success">恢复</a>
                    </li>
                    <li v-if="post.processStatus === 0">
                      <a @click="rejectPost(post)" class="text-warning">封禁</a>
                    </li>
                  </ul>
                </div>
              </td>
            </tr>
            </tbody>
          </table>
        </div>

        <!-- 分页 -->
        <div class="flex justify-between items-center p-4 border-t border-base-300">
          <div class="text-sm text-base-content/70">
            显示 {{ (pagination.page - 1) * pagination.pageSize + 1 }} -
            {{ Math.min(pagination.page * pagination.pageSize, pagination.total) }}
            共 {{ pagination.total }} 条
          </div>
          <div class="join">
            <button
                class="join-item btn btn-sm"
                :disabled="pagination.page <= 1"
                @click="changePage(pagination.page - 1)"
            >
              上一页
            </button>
            <button
                v-for="page in visiblePages"
                :key="page"
                class="join-item btn btn-sm"
                :class="{ 'btn-active': page === pagination.page }"
                @click="changePage(page)"
            >
              {{ page }}
            </button>
            <button
                class="join-item btn btn-sm"
                :disabled="pagination.page >= totalPages"
                @click="changePage(pagination.page + 1)"
            >
              下一页
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
</style>