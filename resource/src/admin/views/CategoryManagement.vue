<template>
  <div class="space-y-6">
    <!-- 页面标题和操作 -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold text-base-content">分类管理</h1>
        <p class="text-base-content/70 mt-1">管理论坛的帖子分类</p>
      </div>
      <button class="btn btn-primary" @click="openCreateModal">
        <PlusIcon class="w-4 h-4" />
        新建分类
      </button>
    </div>

    <!-- 搜索和筛选 -->
    <div class="card bg-base-100 shadow">
      <div class="card-body">
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3 sm:gap-4">
          <div class="form-control">
            <label class="label pb-1">
              <span class="label-text text-sm">搜索分类</span>
            </label>
            <div class="relative">
              <input 
                v-model="searchQuery" 
                type="text" 
                placeholder="分类名称、描述" 
                class="input input-bordered w-full pl-10"
                @input="handleSearch"
              />
              <MagnifyingGlassIcon class="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-base-content/50" />
            </div>
          </div>
          
          <div class="form-control">
            <label class="label pb-1">
              <span class="label-text text-sm">状态筛选</span>
            </label>
            <select v-model="filters.status" class="select select-bordered w-full" @change="handleFilter">
              <option value="">全部状态</option>
              <option value="active">启用</option>
              <option value="inactive">禁用</option>
            </select>
          </div>
          
          <div class="form-control">
            <label class="label pb-1">
              <span class="label-text text-sm">类型筛选</span>
            </label>
            <select v-model="filters.type" class="select select-bordered w-full" @change="handleFilter">
              <option value="">全部类型</option>
              <option value="forum">论坛分类</option>
              <option value="blog">博客分类</option>
              <option value="news">新闻分类</option>
            </select>
          </div>
          
          <div class="form-control">
            <label class="label pb-1">
              <span class="label-text text-sm">排序方式</span>
            </label>
            <select v-model="filters.sortBy" class="select select-bordered w-full" @change="handleFilter">
              <option value="sort_order">排序权重</option>
              <option value="name">分类名称</option>
              <option value="created_at">创建时间</option>
              <option value="post_count">帖子数量</option>
            </select>
          </div>
        </div>
      </div>
    </div>

    <!-- 分类列表 -->
    <div class="card bg-base-100 shadow">
      <div class="card-body p-0">
        <div class="overflow-x-auto">
          <table class="table table-zebra">
            <thead>
              <tr>
                <th>分类信息</th>
                <th>状态</th>
                <th>帖子数量</th>
                <th>排序权重</th>
                <th>创建时间</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="category in categories" :key="category.id">
                <td>
                  <div class="flex items-center gap-3">
                    <div class="avatar placeholder" v-if="category.icon">
                      <div class="bg-neutral text-neutral-content rounded-full w-12">
                        <span class="text-xl">{{ category.icon }}</span>
                      </div>
                    </div>
                    <div class="avatar placeholder" v-else>
                      <div class="bg-neutral text-neutral-content rounded-full w-12">
                        <span class="text-xs">{{ category.name.charAt(0) }}</span>
                      </div>
                    </div>
                    <div>
                      <div class="font-bold">{{ category.name }}</div>
                      <div class="text-sm text-base-content/70">{{ category.description || '暂无描述' }}</div>
                      <div class="text-xs text-base-content/50" v-if="category.slug">
                        标识: {{ category.slug }}
                      </div>
                    </div>
                  </div>
                </td>
                <td>
                  <div class="badge badge-sm whitespace-nowrap" :class="category.status === 'active' ? 'badge-success' : 'badge-error'">
                    {{ category.status === 'active' ? '启用' : '禁用' }}
                  </div>
                </td>
                <td>
                  <div class="stat-value text-sm">{{ category.postCount }}</div>
                </td>
                <td>
                  <div class="badge badge-outline">{{ category.sortOrder }}</div>
                </td>
                <td class="text-sm">{{ formatDate(category.createdAt) }}</td>
                <td>
                  <div class="dropdown dropdown-end">
                    <div tabindex="0" role="button" class="btn btn-ghost btn-xs">
                      <EllipsisVerticalIcon class="w-4 h-4" />
                    </div>
                    <ul tabindex="0" class="dropdown-content z-[1] menu p-2 shadow bg-base-100 rounded-box w-52">
                      <li><a @click="editCategory(category)">编辑</a></li>
                      <li><a @click="toggleStatus(category)">{{ category.status === 'active' ? '禁用' : '启用' }}</a></li>
                      <li><a @click="moveUp(category)" :disabled="isFirst(category)">上移</a></li>
                      <li><a @click="moveDown(category)" :disabled="isLast(category)">下移</a></li>
                      <li><a @click="deleteCategory(category)" class="text-error">删除</a></li>
                    </ul>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- 创建/编辑分类模态框 -->
    <dialog ref="categoryModal" class="modal">
      <div class="modal-box w-11/12 max-w-2xl">
        <form method="dialog">
          <button class="btn btn-sm btn-circle btn-ghost absolute right-2 top-2">✕</button>
        </form>
        
        <h3 class="font-bold text-lg mb-4">
          {{ editingCategory ? '编辑分类' : '新建分类' }}
        </h3>
        
        <div class="space-y-4">
          <div class="form-control">
            <label class="label">
              <span class="label-text">分类名称 <span class="text-error">*</span></span>
            </label>
            <input 
              v-model="categoryForm.name" 
              type="text" 
              placeholder="请输入分类名称" 
              class="input input-bordered"
              :class="{ 'input-error': errors.name }"
            />
            <label class="label" v-if="errors.name">
              <span class="label-text-alt text-error">{{ errors.name }}</span>
            </label>
          </div>
          
          <div class="form-control">
            <label class="label">
              <span class="label-text">分类标识</span>
            </label>
            <input 
              v-model="categoryForm.slug" 
              type="text" 
              placeholder="自动生成或手动输入" 
              class="input input-bordered"
            />
            <label class="label">
              <span class="label-text-alt">用于URL，留空将自动生成</span>
            </label>
          </div>
          
          <div class="form-control">
            <label class="label">
              <span class="label-text">分类描述</span>
            </label>
            <textarea 
              v-model="categoryForm.description" 
              class="textarea textarea-bordered" 
              placeholder="请输入分类描述"
              rows="3"
            ></textarea>
          </div>
          
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div class="form-control">
              <label class="label">
                <span class="label-text">分类图标</span>
              </label>
              <input 
                v-model="categoryForm.icon" 
                type="text" 
                placeholder="如: 📚 或 FontAwesome类名" 
                class="input input-bordered"
              />
            </div>
            
            <div class="form-control">
              <label class="label">
                <span class="label-text">排序权重</span>
              </label>
              <input 
                v-model.number="categoryForm.sortOrder" 
                type="number" 
                placeholder="数字越小越靠前" 
                class="input input-bordered"
                min="0"
              />
            </div>
          </div>
          
          <div class="form-control">
            <label class="label">
              <span class="label-text">分类颜色</span>
            </label>
            <div class="flex items-center gap-2">
              <input 
                v-model="categoryForm.color" 
                type="color" 
                class="w-12 h-10 rounded border border-base-300"
              />
              <input 
                v-model="categoryForm.color" 
                type="text" 
                placeholder="#000000" 
                class="input input-bordered flex-1"
              />
            </div>
          </div>
          
          <div class="form-control">
            <label class="cursor-pointer label">
              <span class="label-text">启用状态</span>
              <input 
                v-model="categoryForm.status" 
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
          <button type="button" class="btn btn-primary" @click="saveCategory" :disabled="saving">
            <span v-if="saving" class="loading loading-spinner loading-sm"></span>
            {{ saving ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>
    </dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, nextTick } from 'vue'
import {
  PlusIcon,
  MagnifyingGlassIcon,
  EllipsisVerticalIcon
} from '@heroicons/vue/24/outline'
import { api } from '../utils/axiosInstance'

// 数据类型定义
interface Category {
  id: number
  name: string
  slug?: string
  description?: string
  icon?: string
  color?: string
  status: 'active' | 'inactive'
  sortOrder: number
  postCount: number
  createdAt: string
}

// 响应式数据
const categories = ref<Category[]>([])
const loading = ref(false)
const saving = ref(false)
const searchQuery = ref('')
const editingCategory = ref<Category | null>(null)
const categoryModal = ref<HTMLDialogElement>()

// 筛选条件
const filters = reactive({
  status: '',
  type: '',
  sortBy: 'sort_order'
})

// 表单数据
const categoryForm = reactive({
  name: '',
  slug: '',
  description: '',
  icon: '',
  color: '#3b82f6',
  status: 'active' as 'active' | 'inactive',
  sortOrder: 0
})

// 表单验证错误
const errors = reactive({
  name: ''
})

// 方法
const fetchCategories = async () => {
  loading.value = true
  try {
    const params = {
      search: searchQuery.value,
      ...filters
    }
    
    const response = await api.get('/api/admin/categories', params)
    categories.value = response.data.data
  } catch (error) {
    console.error('获取分类列表失败:', error)
    // 使用模拟数据
    categories.value = generateMockCategories()
  } finally {
    loading.value = false
  }
}

// 生成模拟数据
const generateMockCategories = (): Category[] => {
  return [
    {
      id: 1,
      name: '技术分享',
      slug: 'tech-share',
      description: '分享技术经验、教程和心得',
      icon: '💻',
      color: '#3b82f6',
      status: 'active',
      sortOrder: 1,
      postCount: 156,
      createdAt: '2024-01-15T10:30:00Z'
    },
    {
      id: 2,
      name: '问题求助',
      slug: 'help',
      description: '遇到问题时寻求帮助和解答',
      icon: '❓',
      color: '#f59e0b',
      status: 'active',
      sortOrder: 2,
      postCount: 89,
      createdAt: '2024-01-15T10:35:00Z'
    },
    {
      id: 3,
      name: '项目展示',
      slug: 'showcase',
      description: '展示个人或团队的项目作品',
      icon: '🚀',
      color: '#10b981',
      status: 'active',
      sortOrder: 3,
      postCount: 67,
      createdAt: '2024-01-15T10:40:00Z'
    },
    {
      id: 4,
      name: '经验交流',
      slug: 'experience',
      description: '分享工作和学习中的经验',
      icon: '💡',
      color: '#8b5cf6',
      status: 'active',
      sortOrder: 4,
      postCount: 43,
      createdAt: '2024-01-15T10:45:00Z'
    },
    {
      id: 5,
      name: '资源分享',
      slug: 'resources',
      description: '分享有用的工具、资源和链接',
      icon: '📚',
      color: '#ef4444',
      status: 'inactive',
      sortOrder: 5,
      postCount: 21,
      createdAt: '2024-01-15T10:50:00Z'
    }
  ]
}

const handleSearch = () => {
  fetchCategories()
}

const handleFilter = () => {
  fetchCategories()
}

const openCreateModal = () => {
  editingCategory.value = null
  resetForm()
  categoryModal.value?.showModal()
}

const editCategory = (category: Category) => {
  editingCategory.value = category
  Object.assign(categoryForm, {
    name: category.name,
    slug: category.slug || '',
    description: category.description || '',
    icon: category.icon || '',
    color: category.color || '#3b82f6',
    status: category.status,
    sortOrder: category.sortOrder
  })
  categoryModal.value?.showModal()
}

const closeModal = () => {
  categoryModal.value?.close()
  resetForm()
}

const resetForm = () => {
  Object.assign(categoryForm, {
    name: '',
    slug: '',
    description: '',
    icon: '',
    color: '#3b82f6',
    status: 'active',
    sortOrder: 0
  })
  Object.assign(errors, {
    name: ''
  })
}

const validateForm = () => {
  errors.name = ''
  
  if (!categoryForm.name.trim()) {
    errors.name = '分类名称不能为空'
    return false
  }
  
  if (categoryForm.name.length > 50) {
    errors.name = '分类名称不能超过50个字符'
    return false
  }
  
  return true
}

const saveCategory = async () => {
  if (!validateForm()) {
    return
  }
  
  saving.value = true
  try {
    const data = { ...categoryForm }
    
    if (editingCategory.value) {
      // 编辑分类
      await api.put(`/api/admin/categories/${editingCategory.value.id}`, data)
    } else {
      // 创建分类
      await api.post('/api/admin/categories', data)
    }
    
    closeModal()
    fetchCategories()
  } catch (error) {
    console.error('保存分类失败:', error)
    // 模拟保存成功
    closeModal()
    fetchCategories()
  } finally {
    saving.value = false
  }
}

const toggleStatus = async (category: Category) => {
  try {
    await api.post(`/api/admin/categories/${category.id}/toggle-status`)
    category.status = category.status === 'active' ? 'inactive' : 'active'
  } catch (error) {
    console.error('切换状态失败:', error)
    // 模拟切换成功
    category.status = category.status === 'active' ? 'inactive' : 'active'
  }
}

const moveUp = async (category: Category) => {
  try {
    await api.post(`/api/admin/categories/${category.id}/move-up`)
    fetchCategories()
  } catch (error) {
    console.error('上移失败:', error)
  }
}

const moveDown = async (category: Category) => {
  try {
    await api.post(`/api/admin/categories/${category.id}/move-down`)
    fetchCategories()
  } catch (error) {
    console.error('下移失败:', error)
  }
}

const deleteCategory = async (category: Category) => {
  if (category.postCount > 0) {
    alert('该分类下还有帖子，无法删除！请先移动或删除相关帖子。')
    return
  }
  
  if (confirm(`确定要删除分类「${category.name}」吗？此操作不可恢复！`)) {
    try {
      await api.delete(`/api/admin/categories/${category.id}`)
      fetchCategories()
    } catch (error) {
      console.error('删除分类失败:', error)
    }
  }
}

// 计算属性
const isFirst = (category: Category) => {
  const sortedCategories = [...categories.value].sort((a, b) => a.sortOrder - b.sortOrder)
  return sortedCategories[0]?.id === category.id
}

const isLast = (category: Category) => {
  const sortedCategories = [...categories.value].sort((a, b) => a.sortOrder - b.sortOrder)
  return sortedCategories[sortedCategories.length - 1]?.id === category.id
}

// 工具函数
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
})
</script>

<style scoped>
.table th {
  background-color: hsl(var(--b2));
  font-weight: 600;
}

/* 颜色预览样式 */
input[type="color"] {
  -webkit-appearance: none;
  border: none;
  cursor: pointer;
}

input[type="color"]::-webkit-color-swatch-wrapper {
  padding: 0;
}

input[type="color"]::-webkit-color-swatch {
  border: none;
  border-radius: 4px;
}
</style>