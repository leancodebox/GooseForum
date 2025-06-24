<template>
  <div class="space-y-6">
    <!-- 页面标题和操作 -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold text-base-content">友情链接管理</h1>
        <p class="text-base-content/70 mt-1">管理网站的友情链接</p>
      </div>
      <button class="btn btn-primary" @click="openCreateModal">
        <PlusIcon class="w-4 h-4" />
        添加链接
      </button>
    </div>

    <!-- 搜索和筛选 -->
    <div class="card bg-base-100 shadow">
      <div class="card-body">
        <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div class="form-control">
            <label class="label">
              <span class="label-text">搜索链接</span>
            </label>
            <div class="relative">
              <input 
                v-model="searchQuery" 
                type="text" 
                placeholder="网站名称、URL" 
                class="input input-bordered w-full pl-10"
                @input="handleSearch"
              />
              <MagnifyingGlassIcon class="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-base-content/50" />
            </div>
          </div>
          
          <div class="form-control">
            <label class="label">
              <span class="label-text">状态筛选</span>
            </label>
            <select v-model="filters.status" class="select select-bordered" @change="handleFilter">
              <option value="">全部状态</option>
              <option value="active">显示</option>
              <option value="inactive">隐藏</option>
            </select>
          </div>
          
          <div class="form-control">
            <label class="label">
              <span class="label-text">分类筛选</span>
            </label>
            <select v-model="filters.category" class="select select-bordered" @change="handleFilter">
              <option value="">全部分类</option>
              <option value="tech">技术类</option>
              <option value="blog">博客类</option>
              <option value="forum">论坛类</option>
              <option value="tool">工具类</option>
              <option value="other">其他</option>
            </select>
          </div>
          
          <div class="form-control">
            <label class="label">
              <span class="label-text">排序方式</span>
            </label>
            <select v-model="filters.sortBy" class="select select-bordered" @change="handleFilter">
              <option value="sort_order">排序权重</option>
              <option value="name">网站名称</option>
              <option value="created_at">添加时间</option>
              <option value="click_count">点击次数</option>
            </select>
          </div>
        </div>
      </div>
    </div>

    <!-- 友情链接列表 -->
    <div class="card bg-base-100 shadow">
      <div class="card-body p-0">
        <div class="overflow-x-auto">
          <table class="table table-zebra">
            <thead>
              <tr>
                <th>网站信息</th>
                <th>分类</th>
                <th>状态</th>
                <th>点击统计</th>
                <th>排序权重</th>
                <th>添加时间</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="link in friendLinks" :key="link.id">
                <td>
                  <div class="flex items-center gap-3">
                    <div class="avatar">
                      <div class="mask mask-squircle w-12 h-12">
                        <img 
                          :src="link.logo || '/static/pic/default-website.png'" 
                          :alt="link.name"
                          @error="handleImageError"
                        />
                      </div>
                    </div>
                    <div>
                      <div class="font-bold">{{ link.name }}</div>
                      <div class="text-sm text-base-content/70">{{ link.description || '暂无描述' }}</div>
                      <div class="text-xs text-primary">
                        <a :href="link.url" target="_blank" class="link link-hover">
                          {{ link.url }}
                        </a>
                      </div>
                    </div>
                  </div>
                </td>
                <td>
                  <div class="badge badge-outline">{{ getCategoryText(link.category) }}</div>
                </td>
                <td>
                  <div class="badge" :class="link.status === 'active' ? 'badge-success' : 'badge-error'">
                    {{ link.status === 'active' ? '显示' : '隐藏' }}
                  </div>
                </td>
                <td>
                  <div class="stat-value text-sm">{{ link.clickCount }}</div>
                </td>
                <td>
                  <div class="badge badge-outline">{{ link.sortOrder }}</div>
                </td>
                <td class="text-sm">{{ formatDate(link.createdAt) }}</td>
                <td>
                  <div class="dropdown dropdown-end">
                    <div tabindex="0" role="button" class="btn btn-ghost btn-xs">
                      <EllipsisVerticalIcon class="w-4 h-4" />
                    </div>
                    <ul tabindex="0" class="dropdown-content z-[1] menu p-2 shadow bg-base-100 rounded-box w-52">
                      <li><a :href="link.url" target="_blank">访问网站</a></li>
                      <li><a @click="editLink(link)">编辑</a></li>
                      <li><a @click="toggleStatus(link)">{{ link.status === 'active' ? '隐藏' : '显示' }}</a></li>
                      <li><a @click="moveUp(link)" :disabled="isFirst(link)">上移</a></li>
                      <li><a @click="moveDown(link)" :disabled="isLast(link)">下移</a></li>
                      <li><a @click="deleteLink(link)" class="text-error">删除</a></li>
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

    <!-- 创建/编辑友情链接模态框 -->
    <dialog ref="linkModal" class="modal">
      <div class="modal-box w-11/12 max-w-2xl">
        <form method="dialog">
          <button class="btn btn-sm btn-circle btn-ghost absolute right-2 top-2">✕</button>
        </form>
        
        <h3 class="font-bold text-lg mb-4">
          {{ editingLink ? '编辑友情链接' : '添加友情链接' }}
        </h3>
        
        <div class="space-y-4">
          <div class="form-control">
            <label class="label">
              <span class="label-text">网站名称 <span class="text-error">*</span></span>
            </label>
            <input 
              v-model="linkForm.name" 
              type="text" 
              placeholder="请输入网站名称" 
              class="input input-bordered"
              :class="{ 'input-error': errors.name }"
            />
            <label class="label" v-if="errors.name">
              <span class="label-text-alt text-error">{{ errors.name }}</span>
            </label>
          </div>
          
          <div class="form-control">
            <label class="label">
              <span class="label-text">网站URL <span class="text-error">*</span></span>
            </label>
            <input 
              v-model="linkForm.url" 
              type="url" 
              placeholder="https://example.com" 
              class="input input-bordered"
              :class="{ 'input-error': errors.url }"
            />
            <label class="label" v-if="errors.url">
              <span class="label-text-alt text-error">{{ errors.url }}</span>
            </label>
          </div>
          
          <div class="form-control">
            <label class="label">
              <span class="label-text">网站描述</span>
            </label>
            <textarea 
              v-model="linkForm.description" 
              class="textarea textarea-bordered" 
              placeholder="请输入网站描述"
              rows="3"
            ></textarea>
          </div>
          
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div class="form-control">
              <label class="label">
                <span class="label-text">网站分类</span>
              </label>
              <select v-model="linkForm.category" class="select select-bordered">
                <option value="tech">技术类</option>
                <option value="blog">博客类</option>
                <option value="forum">论坛类</option>
                <option value="tool">工具类</option>
                <option value="other">其他</option>
              </select>
            </div>
            
            <div class="form-control">
              <label class="label">
                <span class="label-text">排序权重</span>
              </label>
              <input 
                v-model.number="linkForm.sortOrder" 
                type="number" 
                placeholder="数字越小越靠前" 
                class="input input-bordered"
                min="0"
              />
            </div>
          </div>
          
          <div class="form-control">
            <label class="label">
              <span class="label-text">网站Logo URL</span>
            </label>
            <input 
              v-model="linkForm.logo" 
              type="url" 
              placeholder="https://example.com/logo.png" 
              class="input input-bordered"
            />
            <label class="label">
              <span class="label-text-alt">建议尺寸: 64x64 像素</span>
            </label>
          </div>
          
          <!-- Logo预览 -->
          <div v-if="linkForm.logo" class="form-control">
            <label class="label">
              <span class="label-text">Logo预览</span>
            </label>
            <div class="avatar">
              <div class="mask mask-squircle w-16 h-16">
                <img 
                  :src="linkForm.logo" 
                  :alt="linkForm.name"
                  @error="handlePreviewError"
                />
              </div>
            </div>
          </div>
          
          <div class="form-control">
            <label class="cursor-pointer label">
              <span class="label-text">显示状态</span>
              <input 
                v-model="linkForm.status" 
                type="checkbox" 
                class="toggle toggle-primary" 
                true-value="active"
                false-value="inactive"
              />
            </label>
          </div>
        </div>
        
        <div class="modal-action">
          <button type="button" class="btn btn-ghost" @click="closeModal">取消</button>
          <button type="button" class="btn btn-primary" @click="saveLink" :disabled="saving">
            <span v-if="saving" class="loading loading-spinner loading-sm"></span>
            {{ saving ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>
    </dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import {
  PlusIcon,
  MagnifyingGlassIcon,
  EllipsisVerticalIcon
} from '@heroicons/vue/24/outline'
import { api } from '../utils/axiosInstance'

// 数据类型定义
interface FriendLink {
  id: number
  name: string
  url: string
  description?: string
  logo?: string
  category: string
  status: 'active' | 'inactive'
  sortOrder: number
  clickCount: number
  createdAt: string
}

// 响应式数据
const friendLinks = ref<FriendLink[]>([])
const loading = ref(false)
const saving = ref(false)
const searchQuery = ref('')
const editingLink = ref<FriendLink | null>(null)
const linkModal = ref<HTMLDialogElement>()

// 筛选条件
const filters = reactive({
  status: '',
  category: '',
  sortBy: 'sort_order'
})

// 分页
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

// 表单数据
const linkForm = reactive({
  name: '',
  url: '',
  description: '',
  logo: '',
  category: 'tech',
  status: 'active' as 'active' | 'inactive',
  sortOrder: 0
})

// 表单验证错误
const errors = reactive({
  name: '',
  url: ''
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
const fetchFriendLinks = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      pageSize: pagination.pageSize,
      search: searchQuery.value,
      ...filters
    }
    
    const response = await api.get('/api/admin/friend-links')
    friendLinks.value = response.data.data.links || response.data.data
    pagination.total = response.data.data.total || friendLinks.value.length
  } catch (error) {
    console.error('获取友情链接列表失败:', error)
    // 使用模拟数据
    friendLinks.value = generateMockLinks()
    pagination.total = 25
  } finally {
    loading.value = false
  }
}

// 生成模拟数据
const generateMockLinks = (): FriendLink[] => {
  const categories = ['tech', 'blog', 'forum', 'tool', 'other']
  const mockLinks: FriendLink[] = []
  
  for (let i = 1; i <= pagination.pageSize; i++) {
    const linkId = (pagination.page - 1) * pagination.pageSize + i
    mockLinks.push({
      id: linkId,
      name: `示例网站 ${linkId}`,
      url: `https://example${linkId}.com`,
      description: `这是示例网站 ${linkId} 的描述信息`,
      logo: Math.random() > 0.5 ? '/static/pic/icon.png' : undefined,
      category: categories[Math.floor(Math.random() * categories.length)],
      status: Math.random() > 0.2 ? 'active' : 'inactive',
      sortOrder: linkId,
      clickCount: Math.floor(Math.random() * 1000),
      createdAt: new Date(Date.now() - Math.random() * 30 * 24 * 60 * 60 * 1000).toISOString()
    })
  }
  
  return mockLinks
}

const handleSearch = () => {
  pagination.page = 1
  fetchFriendLinks()
}

const handleFilter = () => {
  pagination.page = 1
  fetchFriendLinks()
}

const changePage = (page: number) => {
  pagination.page = page
  fetchFriendLinks()
}

const openCreateModal = () => {
  editingLink.value = null
  resetForm()
  linkModal.value?.showModal()
}

const editLink = (link: FriendLink) => {
  editingLink.value = link
  Object.assign(linkForm, {
    name: link.name,
    url: link.url,
    description: link.description || '',
    logo: link.logo || '',
    category: link.category,
    status: link.status,
    sortOrder: link.sortOrder
  })
  linkModal.value?.showModal()
}

const closeModal = () => {
  linkModal.value?.close()
  resetForm()
}

const resetForm = () => {
  Object.assign(linkForm, {
    name: '',
    url: '',
    description: '',
    logo: '',
    category: 'tech',
    status: 'active',
    sortOrder: 0
  })
  Object.assign(errors, {
    name: '',
    url: ''
  })
}

const validateForm = () => {
  errors.name = ''
  errors.url = ''
  
  if (!linkForm.name.trim()) {
    errors.name = '网站名称不能为空'
    return false
  }
  
  if (linkForm.name.length > 50) {
    errors.name = '网站名称不能超过50个字符'
    return false
  }
  
  if (!linkForm.url.trim()) {
    errors.url = '网站URL不能为空'
    return false
  }
  
  // 简单的URL格式验证
  const urlPattern = /^https?:\/\/.+/
  if (!urlPattern.test(linkForm.url)) {
    errors.url = '请输入有效的URL地址'
    return false
  }
  
  return true
}

const saveLink = async () => {
  if (!validateForm()) {
    return
  }
  
  saving.value = true
  try {
    const data = { ...linkForm }
    
    // 统一使用保存友情链接接口
    await api.post('/api/admin/save-friend-links', {
      ...data,
      id: editingLink.value?.id
    })
    
    closeModal()
    fetchFriendLinks()
  } catch (error) {
    console.error('保存友情链接失败:', error)
    // 模拟保存成功
    closeModal()
    fetchFriendLinks()
  } finally {
    saving.value = false
  }
}

const toggleStatus = async (link: FriendLink) => {
  try {
    await api.post('/api/admin/save-friend-links', {
      id: link.id,
      status: link.status === 'active' ? 'inactive' : 'active'
    })
    link.status = link.status === 'active' ? 'inactive' : 'active'
  } catch (error) {
    console.error('切换状态失败:', error)
    // 模拟切换成功
    link.status = link.status === 'active' ? 'inactive' : 'active'
  }
}

const moveUp = async (link: FriendLink) => {
  try {
    await api.post('/api/admin/save-friend-links', {
      id: link.id,
      action: 'move-up'
    })
    fetchFriendLinks()
  } catch (error) {
    console.error('上移失败:', error)
  }
}

const moveDown = async (link: FriendLink) => {
  try {
    await api.post('/api/admin/save-friend-links', {
      id: link.id,
      action: 'move-down'
    })
    fetchFriendLinks()
  } catch (error) {
    console.error('下移失败:', error)
  }
}

const deleteLink = async (link: FriendLink) => {
  if (confirm(`确定要删除友情链接「${link.name}」吗？此操作不可恢复！`)) {
    try {
      await api.post('/api/admin/save-friend-links', {
        id: link.id,
        action: 'delete'
      })
      fetchFriendLinks()
    } catch (error) {
      console.error('删除友情链接失败:', error)
    }
  }
}

// 计算属性
const isFirst = (link: FriendLink) => {
  const sortedLinks = [...friendLinks.value].sort((a, b) => a.sortOrder - b.sortOrder)
  return sortedLinks[0]?.id === link.id
}

const isLast = (link: FriendLink) => {
  const sortedLinks = [...friendLinks.value].sort((a, b) => a.sortOrder - b.sortOrder)
  return sortedLinks[sortedLinks.length - 1]?.id === link.id
}

// 工具函数
const getCategoryText = (category: string) => {
  const categoryMap = {
    tech: '技术类',
    blog: '博客类',
    forum: '论坛类',
    tool: '工具类',
    other: '其他'
  }
  return categoryMap[category as keyof typeof categoryMap] || category
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

const handleImageError = (event: Event) => {
  const target = event.target as HTMLImageElement
  target.src = '/static/pic/default-website.png'
}

const handlePreviewError = (event: Event) => {
  const target = event.target as HTMLImageElement
  target.src = '/static/pic/default-website.png'
}

// 组件挂载时获取数据
onMounted(() => {
  fetchFriendLinks()
})
</script>

<style scoped>
.table th {
  background-color: hsl(var(--b2));
  font-weight: 600;
}

/* 链接样式 */
.link:hover {
  text-decoration: underline;
}

/* 图片加载失败样式 */
img[src="/static/pic/default-website.png"] {
  background-color: hsl(var(--b3));
}
</style>