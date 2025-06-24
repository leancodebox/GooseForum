<template>
  <div class="space-y-6">
    <!-- 页面标题和操作 -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold text-base-content">用户管理</h1>
        <p class="text-base-content/70 mt-1">管理系统中的所有用户</p>
      </div>
      <button class="btn btn-primary" @click="openCreateModal">
        <PlusIcon class="w-4 h-4" />
        添加用户
      </button>
    </div>

    <!-- 搜索和筛选 -->
    <div class="card bg-base-100 shadow">
      <div class="card-body">
        <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div class="form-control">
            <label class="label">
              <span class="label-text">搜索用户</span>
            </label>
            <div class="relative">
              <input 
                v-model="searchQuery" 
                type="text" 
                placeholder="用户名、邮箱" 
                class="input input-bordered w-full pl-10"
                @input="handleSearch"
              />
              <MagnifyingGlassIcon class="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-base-content/50" />
            </div>
          </div>
          
          <div class="form-control">
            <label class="label">
              <span class="label-text">角色筛选</span>
            </label>
            <select v-model="filters.role" class="select select-bordered" @change="handleFilter">
              <option value="">全部角色</option>
              <option value="admin">管理员</option>
              <option value="moderator">版主</option>
              <option value="user">普通用户</option>
            </select>
          </div>
          
          <div class="form-control">
            <label class="label">
              <span class="label-text">状态筛选</span>
            </label>
            <select v-model="filters.status" class="select select-bordered" @change="handleFilter">
              <option value="">全部状态</option>
              <option value="active">正常</option>
              <option value="banned">已封禁</option>
              <option value="pending">待激活</option>
            </select>
          </div>
          
          <div class="form-control">
            <label class="label">
              <span class="label-text">注册时间</span>
            </label>
            <select v-model="filters.dateRange" class="select select-bordered" @change="handleFilter">
              <option value="">全部时间</option>
              <option value="today">今天</option>
              <option value="week">本周</option>
              <option value="month">本月</option>
            </select>
          </div>
        </div>
      </div>
    </div>

    <!-- 用户列表 -->
    <div class="card bg-base-100 shadow">
      <div class="card-body p-0">
        <div class="overflow-x-auto">
          <table class="table table-zebra">
            <thead>
              <tr>
                <th>
                  <label>
                    <input 
                      type="checkbox" 
                      class="checkbox" 
                      :checked="isAllSelected"
                      @change="toggleSelectAll"
                    />
                  </label>
                </th>
                <th>用户信息</th>
                <th>角色</th>
                <th>状态</th>
                <th>注册时间</th>
                <th>最后登录</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="user in users" :key="user.id">
                <td>
                  <label>
                    <input 
                      type="checkbox" 
                      class="checkbox" 
                      :value="user.id"
                      v-model="selectedUsers"
                    />
                  </label>
                </td>
                <td>
                  <div class="flex items-center gap-3">
                    <div class="avatar">
                      <div class="mask mask-squircle w-12 h-12">
                        <img :src="user.avatar || '/static/pic/default-avatar.png'" :alt="user.username" />
                      </div>
                    </div>
                    <div>
                      <div class="font-bold">{{ user.username }}</div>
                      <div class="text-sm opacity-50">{{ user.email }}</div>
                    </div>
                  </div>
                </td>
                <td>
                  <div class="badge" :class="getRoleBadgeClass(user.role)">
                    {{ getRoleText(user.role) }}
                  </div>
                </td>
                <td>
                  <div class="badge" :class="getStatusBadgeClass(user.status)">
                    {{ getStatusText(user.status) }}
                  </div>
                </td>
                <td>{{ formatDate(user.createdAt) }}</td>
                <td>{{ user.lastLoginAt ? formatDate(user.lastLoginAt) : '从未登录' }}</td>
                <td>
                  <div class="dropdown dropdown-end">
                    <div tabindex="0" role="button" class="btn btn-ghost btn-xs">
                      <EllipsisVerticalIcon class="w-4 h-4" />
                    </div>
                    <ul tabindex="0" class="dropdown-content z-[1] menu p-2 shadow bg-base-100 rounded-box w-52">
                      <li><a @click="editUser(user)">编辑</a></li>
                      <li><a @click="resetPassword(user)">重置密码</a></li>
                      <li v-if="user.status === 'active'">
                        <a @click="banUser(user)" class="text-warning">封禁用户</a>
                      </li>
                      <li v-else>
                        <a @click="unbanUser(user)" class="text-success">解除封禁</a>
                      </li>
                      <li><a @click="deleteUser(user)" class="text-error">删除</a></li>
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

    <!-- 批量操作 -->
    <div v-if="selectedUsers.length > 0" class="fixed bottom-6 left-1/2 transform -translate-x-1/2 z-50">
      <div class="card bg-base-100 shadow-lg border border-base-300">
        <div class="card-body p-4">
          <div class="flex items-center gap-4">
            <span class="text-sm font-medium">已选择 {{ selectedUsers.length }} 个用户</span>
            <div class="flex gap-2">
              <button class="btn btn-sm btn-warning" @click="batchBan">
                批量封禁
              </button>
              <button class="btn btn-sm btn-error" @click="batchDelete">
                批量删除
              </button>
              <button class="btn btn-sm btn-ghost" @click="clearSelection">
                取消选择
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>

  <!-- 用户编辑/创建模态框 -->
  <dialog ref="userModal" class="modal">
    <div class="modal-box w-11/12 max-w-2xl">
      <h3 class="font-bold text-lg mb-4">{{ isEditing ? '编辑用户' : '添加用户' }}</h3>
      
      <form @submit.prevent="saveUser" class="space-y-4">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div class="form-control">
            <label class="label">
              <span class="label-text">用户名 *</span>
            </label>
            <input 
              v-model="userForm.username" 
              type="text" 
              class="input input-bordered" 
              required
            />
          </div>
          
          <div class="form-control">
            <label class="label">
              <span class="label-text">邮箱 *</span>
            </label>
            <input 
              v-model="userForm.email" 
              type="email" 
              class="input input-bordered" 
              required
            />
          </div>
          
          <div class="form-control" v-if="!isEditing">
            <label class="label">
              <span class="label-text">密码 *</span>
            </label>
            <input 
              v-model="userForm.password" 
              type="password" 
              class="input input-bordered" 
              :required="!isEditing"
            />
          </div>
          
          <div class="form-control">
            <label class="label">
              <span class="label-text">角色</span>
            </label>
            <select v-model="userForm.role" class="select select-bordered">
              <option value="user">普通用户</option>
              <option value="moderator">版主</option>
              <option value="admin">管理员</option>
            </select>
          </div>
          
          <div class="form-control">
            <label class="label">
              <span class="label-text">状态</span>
            </label>
            <select v-model="userForm.status" class="select select-bordered">
              <option value="active">正常</option>
              <option value="banned">已封禁</option>
              <option value="pending">待激活</option>
            </select>
          </div>
        </div>
        
        <div class="modal-action">
          <button type="button" class="btn" @click="closeModal">取消</button>
          <button type="submit" class="btn btn-primary" :disabled="loading">
            {{ loading ? '保存中...' : '保存' }}
          </button>
        </div>
      </form>
    </div>
    <form method="dialog" class="modal-backdrop">
      <button @click="closeModal">close</button>
    </form>
  </dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import {
  PlusIcon,
  MagnifyingGlassIcon,
  EllipsisVerticalIcon
} from '@heroicons/vue/24/outline'
import { api } from '../utils/axiosInstance'

// 用户数据类型
interface User {
  id: number
  username: string
  email: string
  avatar?: string
  role: string
  status: string
  createdAt: string
  lastLoginAt?: string
}

// 响应式数据
const users = ref<User[]>([])
const loading = ref(false)
const searchQuery = ref('')
const selectedUsers = ref<number[]>([])
const userModal = ref<HTMLDialogElement>()
const isEditing = ref(false)

// 筛选条件
const filters = reactive({
  role: '',
  status: '',
  dateRange: ''
})

// 分页
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

// 用户表单
const userForm = reactive({
  id: 0,
  username: '',
  email: '',
  password: '',
  role: 'user',
  status: 'active'
})

// 计算属性
const isAllSelected = computed(() => {
  return users.value.length > 0 && selectedUsers.value.length === users.value.length
})

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
const fetchUsers = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      pageSize: pagination.pageSize,
      search: searchQuery.value,
      ...filters
    }
    
    const response = await api.get('/api/admin/users', params)
    users.value = response.data.data.users
    pagination.total = response.data.data.total
  } catch (error) {
    console.error('获取用户列表失败:', error)
    // 使用模拟数据
    users.value = generateMockUsers()
    pagination.total = 100
  } finally {
    loading.value = false
  }
}

// 生成模拟数据
const generateMockUsers = (): User[] => {
  const roles = ['admin', 'moderator', 'user']
  const statuses = ['active', 'banned', 'pending']
  const mockUsers: User[] = []
  
  for (let i = 1; i <= pagination.pageSize; i++) {
    mockUsers.push({
      id: (pagination.page - 1) * pagination.pageSize + i,
      username: `user${i}`,
      email: `user${i}@example.com`,
      avatar: `/static/pic/default-avatar.png`,
      role: roles[Math.floor(Math.random() * roles.length)],
      status: statuses[Math.floor(Math.random() * statuses.length)],
      createdAt: new Date(Date.now() - Math.random() * 30 * 24 * 60 * 60 * 1000).toISOString(),
      lastLoginAt: Math.random() > 0.3 ? new Date(Date.now() - Math.random() * 7 * 24 * 60 * 60 * 1000).toISOString() : undefined
    })
  }
  
  return mockUsers
}

const handleSearch = () => {
  pagination.page = 1
  fetchUsers()
}

const handleFilter = () => {
  pagination.page = 1
  fetchUsers()
}

const changePage = (page: number) => {
  pagination.page = page
  fetchUsers()
}

const toggleSelectAll = () => {
  if (isAllSelected.value) {
    selectedUsers.value = []
  } else {
    selectedUsers.value = users.value.map(user => user.id)
  }
}

const clearSelection = () => {
  selectedUsers.value = []
}

const openCreateModal = () => {
  isEditing.value = false
  resetUserForm()
  userModal.value?.showModal()
}

const editUser = (user: User) => {
  isEditing.value = true
  Object.assign(userForm, user)
  userModal.value?.showModal()
}

const closeModal = () => {
  userModal.value?.close()
  resetUserForm()
}

const resetUserForm = () => {
  Object.assign(userForm, {
    id: 0,
    username: '',
    email: '',
    password: '',
    role: 'user',
    status: 'active'
  })
}

const saveUser = async () => {
  loading.value = true
  try {
    if (isEditing.value) {
      await api.put(`/api/admin/users/${userForm.id}`, userForm)
    } else {
      await api.post('/api/admin/users', userForm)
    }
    
    closeModal()
    fetchUsers()
  } catch (error) {
    console.error('保存用户失败:', error)
  } finally {
    loading.value = false
  }
}

const resetPassword = async (user: User) => {
  if (confirm(`确定要重置用户 ${user.username} 的密码吗？`)) {
    try {
      await api.post(`/api/admin/users/${user.id}/reset-password`)
      alert('密码重置成功')
    } catch (error) {
      console.error('重置密码失败:', error)
    }
  }
}

const banUser = async (user: User) => {
  if (confirm(`确定要封禁用户 ${user.username} 吗？`)) {
    try {
      await api.post(`/api/admin/users/${user.id}/ban`)
      fetchUsers()
    } catch (error) {
      console.error('封禁用户失败:', error)
    }
  }
}

const unbanUser = async (user: User) => {
  if (confirm(`确定要解除用户 ${user.username} 的封禁吗？`)) {
    try {
      await api.post(`/api/admin/users/${user.id}/unban`)
      fetchUsers()
    } catch (error) {
      console.error('解除封禁失败:', error)
    }
  }
}

const deleteUser = async (user: User) => {
  if (confirm(`确定要删除用户 ${user.username} 吗？此操作不可恢复！`)) {
    try {
      await api.delete(`/api/admin/users/${user.id}`)
      fetchUsers()
    } catch (error) {
      console.error('删除用户失败:', error)
    }
  }
}

const batchBan = async () => {
  if (confirm(`确定要封禁选中的 ${selectedUsers.value.length} 个用户吗？`)) {
    try {
      await api.post('/api/admin/users/batch-ban', { userIds: selectedUsers.value })
      clearSelection()
      fetchUsers()
    } catch (error) {
      console.error('批量封禁失败:', error)
    }
  }
}

const batchDelete = async () => {
  if (confirm(`确定要删除选中的 ${selectedUsers.value.length} 个用户吗？此操作不可恢复！`)) {
    try {
      await api.post('/api/admin/users/batch-delete', { userIds: selectedUsers.value })
      clearSelection()
      fetchUsers()
    } catch (error) {
      console.error('批量删除失败:', error)
    }
  }
}

// 工具函数
const getRoleBadgeClass = (role: string) => {
  const classes = {
    admin: 'badge-error',
    moderator: 'badge-warning',
    user: 'badge-info'
  }
  return classes[role as keyof typeof classes] || 'badge-ghost'
}

const getRoleText = (role: string) => {
  const texts = {
    admin: '管理员',
    moderator: '版主',
    user: '普通用户'
  }
  return texts[role as keyof typeof texts] || role
}

const getStatusBadgeClass = (status: string) => {
  const classes = {
    active: 'badge-success',
    banned: 'badge-error',
    pending: 'badge-warning'
  }
  return classes[status as keyof typeof classes] || 'badge-ghost'
}

const getStatusText = (status: string) => {
  const texts = {
    active: '正常',
    banned: '已封禁',
    pending: '待激活'
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
  fetchUsers()
})
</script>

<style scoped>
/* 自定义样式 */
.table th {
  background-color: hsl(var(--b2));
  font-weight: 600;
}

.modal-box {
  max-height: 90vh;
  overflow-y: auto;
}

/* 批量操作栏动画 */
.fixed {
  animation: slideUp 0.3s ease-out;
}

@keyframes slideUp {
  from {
    transform: translate(-50%, 100%);
    opacity: 0;
  }
  to {
    transform: translate(-50%, 0);
    opacity: 1;
  }
}
</style>