<template>
  <div class="space-y-6">
    <!-- 页面标题和操作 -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold text-base-content">帖子管理</h1>
        <p class="text-base-content/70 mt-1">管理系统中的所有帖子内容</p>
      </div>
      <div class="flex gap-2">
        <button class="btn btn-outline" @click="exportPosts">
          <ArrowDownTrayIcon class="w-4 h-4" />
          导出数据
        </button>
        <router-link to="/publish-v3" class="btn btn-primary">
          <PlusIcon class="w-4 h-4" />
          发布帖子
        </router-link>
      </div>
    </div>

    <!-- 搜索和筛选 -->
    <div class="card bg-base-100 shadow">
      <div class="card-body">
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3 sm:gap-4">
          <div class="form-control">
            <label class="label pb-1">
              <span class="label-text text-sm">搜索帖子</span>
            </label>
            <div class="relative">
              <input 
                v-model="searchQuery" 
                type="text" 
                placeholder="帖子标题、作者" 
                class="input input-bordered w-full pl-10"
                @input="handleSearch"
              />
              <MagnifyingGlassIcon class="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-base-content/50" />
            </div>
          </div>
          
          <div class="form-control">
            <label class="label pb-1">
              <span class="label-text text-sm">分类筛选</span>
            </label>
            <select v-model="filters.category" class="select select-bordered w-full" @change="handleFilter">
              <option value="">全部分类</option>
              <option value="tech">技术讨论</option>
              <option value="general">综合讨论</option>
              <option value="help">求助问答</option>
              <option value="share">资源分享</option>
            </select>
          </div>
          
          <div class="form-control">
            <label class="label pb-1">
              <span class="label-text text-sm">状态筛选</span>
            </label>
            <select v-model="filters.status" class="select select-bordered w-full" @change="handleFilter">
              <option value="">全部状态</option>
              <option value="published">已发布</option>
              <option value="draft">草稿</option>
              <option value="hidden">已隐藏</option>
              <option value="deleted">已删除</option>
            </select>
          </div>
          
          <div class="form-control">
            <label class="label pb-1">
              <span class="label-text text-sm">发布时间</span>
            </label>
            <select v-model="filters.dateRange" class="select select-bordered w-full" @change="handleFilter">
              <option value="">全部时间</option>
              <option value="today">今天</option>
              <option value="week">本周</option>
              <option value="month">本月</option>
              <option value="year">本年</option>
            </select>
          </div>
        </div>
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
                    <div class="avatar" v-if="post.thumbnail">
                      <div class="mask mask-squircle w-12 h-12">
                        <img :src="post.thumbnail" :alt="post.title" />
                      </div>
                    </div>
                    <div class="flex-1 min-w-0">
                      <div class="font-bold text-sm line-clamp-2">{{ post.title }}</div>
                      <div class="text-xs text-base-content/70 line-clamp-1 mt-1">
                        {{ post.summary || '暂无摘要' }}
                      </div>
                      <div class="flex items-center gap-2 mt-1">
                        <div class="badge badge-xs" v-if="post.isTop">
                          置顶
                        </div>
                        <div class="badge badge-xs badge-accent" v-if="post.isHot">
                          热门
                        </div>
                        <div class="badge badge-xs badge-info" v-if="post.isRecommended">
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
                        <img :src="post.author.avatar || '/static/pic/default-avatar.png'" :alt="post.author.username" />
                      </div>
                    </div>
                    <div class="text-sm">
                      <div class="font-medium">{{ post.author.username }}</div>
                    </div>
                  </div>
                </td>
                <td>
                  <div class="badge badge-outline badge-sm whitespace-nowrap">{{ post.category.name }}</div>
                </td>
                <td>
                  <div class="badge badge-sm whitespace-nowrap" :class="getStatusBadgeClass(post.status)">
                    {{ getStatusText(post.status) }}
                  </div>
                </td>
                <td>
                  <div class="text-xs space-y-1">
                    <div>浏览: {{ post.views }}</div>
                    <div>评论: {{ post.comments }}</div>
                    <div>点赞: {{ post.likes }}</div>
                  </div>
                </td>
                <td class="text-xs">{{ formatDate(post.createdAt) }}</td>
                <td>
                  <div class="dropdown dropdown-end">
                    <div tabindex="0" role="button" class="btn btn-ghost btn-xs">
                      <EllipsisVerticalIcon class="w-4 h-4" />
                    </div>
                    <ul tabindex="0" class="dropdown-content z-[1] menu p-2 shadow bg-base-100 rounded-box w-52">
                      <li><a :href="`/detail/${post.id}`" target="_blank">查看</a></li>
                      <li><a @click="editPost(post)">编辑</a></li>
                      <li><a @click="toggleTop(post)">{{ post.isTop ? '取消置顶' : '置顶' }}</a></li>
                      <li><a @click="toggleRecommend(post)">{{ post.isRecommended ? '取消推荐' : '推荐' }}</a></li>
                      <li v-if="post.status === 'pending'">
                        <a @click="approvePost(post)" class="text-success">审核通过</a>
                      </li>
                      <li v-if="post.status === 'pending'">
                        <a @click="rejectPost(post)" class="text-warning">审核拒绝</a>
                      </li>
                      <li><a @click="deletePost(post)" class="text-error">删除</a></li>
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

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import {
  PlusIcon,
  MagnifyingGlassIcon,
  EllipsisVerticalIcon,
  ArrowDownTrayIcon
} from '@heroicons/vue/24/outline'
import { api } from '../utils/axiosInstance'

// 数据类型定义
interface Post {
  id: number
  title: string
  summary?: string
  thumbnail?: string
  status: string
  isTop: boolean
  isHot: boolean
  isRecommended: boolean
  views: number
  comments: number
  likes: number
  createdAt: string
  author: {
    id: number
    username: string
    avatar?: string
  }
  category: {
    id: number
    name: string
  }
}

interface Category {
  id: number
  name: string
}

// 响应式数据
const posts = ref<Post[]>([])
const categories = ref<Category[]>([])
const loading = ref(false)
const searchQuery = ref('')

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
    
    const response = await api.post('/api/admin/articles-list', params)
    posts.value = response.data.data.posts
    pagination.total = response.data.data.total
  } catch (error) {
    console.error('获取帖子列表失败:', error)
    // 使用模拟数据
    posts.value = generateMockPosts()
    pagination.total = 150
  } finally {
    loading.value = false
  }
}

const fetchCategories = async () => {
  try {
    const response = await api.post('/api/admin/category-list', {})
    categories.value = response.data.data
  } catch (error) {
    console.error('获取分类列表失败:', error)
    // 使用模拟数据
    categories.value = [
      { id: 1, name: '技术分享' },
      { id: 2, name: '问题求助' },
      { id: 3, name: '项目展示' },
      { id: 4, name: '经验交流' },
      { id: 5, name: '资源分享' }
    ]
  }
}

// 生成模拟数据
const generateMockPosts = (): Post[] => {
  const statuses = ['published', 'draft', 'pending', 'rejected']
  const mockPosts: Post[] = []
  
  for (let i = 1; i <= pagination.pageSize; i++) {
    const postId = (pagination.page - 1) * pagination.pageSize + i
    mockPosts.push({
      id: postId,
      title: `示例帖子标题 ${postId} - 这是一个很长的标题用来测试显示效果`,
      summary: `这是帖子 ${postId} 的摘要内容，用来简要描述帖子的主要内容...`,
      thumbnail: Math.random() > 0.5 ? '/static/pic/icon.png' : undefined,
      status: statuses[Math.floor(Math.random() * statuses.length)],
      isTop: Math.random() > 0.8,
      isHot: Math.random() > 0.7,
      isRecommended: Math.random() > 0.6,
      views: Math.floor(Math.random() * 1000),
      comments: Math.floor(Math.random() * 50),
      likes: Math.floor(Math.random() * 100),
      createdAt: new Date(Date.now() - Math.random() * 30 * 24 * 60 * 60 * 1000).toISOString(),
      author: {
        id: Math.floor(Math.random() * 100) + 1,
        username: `user${Math.floor(Math.random() * 100) + 1}`,
        avatar: '/static/pic/default-avatar.png'
      },
      category: {
        id: Math.floor(Math.random() * 5) + 1,
        name: ['技术分享', '问题求助', '项目展示', '经验交流', '资源分享'][Math.floor(Math.random() * 5)]
      }
    })
  }
  
  return mockPosts
}

const handleSearch = () => {
  pagination.page = 1
  fetchPosts()
}

const handleFilter = () => {
  pagination.page = 1
  fetchPosts()
}

const changePage = (page: number) => {
  pagination.page = page
  fetchPosts()
}



const editPost = (post: Post) => {
  // 跳转到编辑页面
  window.open(`/publish-v3?id=${post.id}`, '_blank')
}

const toggleTop = async (post: Post) => {
  try {
    await api.post(`/api/admin/posts/${post.id}/toggle-top`)
    post.isTop = !post.isTop
  } catch (error) {
    console.error('切换置顶状态失败:', error)
  }
}

const toggleRecommend = async (post: Post) => {
  try {
    await api.post(`/api/admin/posts/${post.id}/toggle-recommend`)
    post.isRecommended = !post.isRecommended
  } catch (error) {
    console.error('切换推荐状态失败:', error)
  }
}

const approvePost = async (post: Post) => {
  try {
    await api.post('/api/admin/article-edit', { id: post.id, action: 'approve' })
    post.status = 'published'
  } catch (error) {
    console.error('审核通过失败:', error)
  }
}

const rejectPost = async (post: Post) => {
  const reason = prompt('请输入拒绝原因:')
  if (reason) {
    try {
      await api.post('/api/admin/article-edit', { id: post.id, action: 'reject', reason })
      post.status = 'rejected'
    } catch (error) {
      console.error('审核拒绝失败:', error)
    }
  }
}

const deletePost = async (post: Post) => {
  if (confirm(`确定要删除帖子「${post.title}」吗？此操作不可恢复！`)) {
    try {
      await api.post('/api/admin/article-edit', { id: post.id, action: 'delete' })
      fetchPosts()
    } catch (error) {
      console.error('删除帖子失败:', error)
    }
  }
}



const exportPosts = async () => {
  try {
    // 导出接口未实现，使用模拟数据
    console.warn('导出接口未实现')
    return
  } catch (error) {
    console.error('导出失败:', error)
    alert('导出功能暂未实现')
  }
}

// 工具函数
const getStatusBadgeClass = (status: string) => {
  const classes = {
    published: 'badge-success',
    draft: 'badge-warning',
    pending: 'badge-info',
    rejected: 'badge-error'
  }
  return classes[status as keyof typeof classes] || 'badge-ghost'
}

const getStatusText = (status: string) => {
  const texts = {
    published: '已发布',
    draft: '草稿',
    pending: '待审核',
    rejected: '已拒绝'
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

// 组件挂载时获取数据
onMounted(() => {
  fetchCategories()
  fetchPosts()
})
</script>

<style scoped>
/* 自定义样式 */
.line-clamp-1 {
  display: -webkit-box;
  -webkit-line-clamp: 1;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.table th {
  background-color: hsl(var(--b2));
  font-weight: 600;
}


</style>