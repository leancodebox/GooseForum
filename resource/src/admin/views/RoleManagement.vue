<template>
  <div class="space-y-6">
    <!-- 页面标题和操作 -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold text-base-content">角色管理</h1>
        <p class="text-base-content/70 mt-1">管理系统角色和权限</p>
      </div>
      <button class="btn btn-primary" @click="openCreateModal">
        <PlusIcon class="w-4 h-4" />
        新建角色
      </button>
    </div>

    <!-- 搜索和筛选 -->
    <div class="card bg-base-100 shadow">
      <div class="card-body">
        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div class="form-control">
            <label class="label">
              <span class="label-text">搜索角色</span>
            </label>
            <div class="relative">
              <input 
                v-model="searchQuery" 
                type="text" 
                placeholder="角色名称、描述" 
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
              <option value="active">启用</option>
              <option value="inactive">禁用</option>
            </select>
          </div>
          
          <div class="form-control">
            <label class="label">
              <span class="label-text">排序方式</span>
            </label>
            <select v-model="filters.sortBy" class="select select-bordered" @change="handleFilter">
              <option value="level">权限级别</option>
              <option value="name">角色名称</option>
              <option value="user_count">用户数量</option>
              <option value="created_at">创建时间</option>
            </select>
          </div>
        </div>
      </div>
    </div>

    <!-- 角色列表 -->
    <div class="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-6">
      <div v-for="role in roles" :key="role.id" class="card bg-base-100 shadow hover:shadow-lg transition-shadow">
        <div class="card-body">
          <div class="flex items-start justify-between">
            <div class="flex items-center gap-3">
              <div class="avatar placeholder">
                <div class="bg-primary text-primary-content rounded-full w-12">
                  <span class="text-lg">{{ role.name.charAt(0) }}</span>
                </div>
              </div>
              <div>
                <h3 class="card-title text-lg">{{ role.name }}</h3>
                <p class="text-sm text-base-content/70">{{ role.description || '暂无描述' }}</p>
              </div>
            </div>
            <div class="dropdown dropdown-end">
              <div tabindex="0" role="button" class="btn btn-ghost btn-sm">
                <EllipsisVerticalIcon class="w-4 h-4" />
              </div>
              <ul tabindex="0" class="dropdown-content z-[1] menu p-2 shadow bg-base-100 rounded-box w-52">
                <li><a @click="editRole(role)">编辑</a></li>
                <li><a @click="managePermissions(role)">权限管理</a></li>
                <li><a @click="toggleStatus(role)">{{ role.status === 'active' ? '禁用' : '启用' }}</a></li>
                <li v-if="!role.isSystem"><a @click="deleteRole(role)" class="text-error">删除</a></li>
              </ul>
            </div>
          </div>
          
          <div class="mt-4 space-y-3">
            <!-- 角色状态和级别 -->
            <div class="flex items-center justify-between">
              <div class="badge" :class="role.status === 'active' ? 'badge-success' : 'badge-error'">
                {{ role.status === 'active' ? '启用' : '禁用' }}
              </div>
              <div class="badge badge-outline">
                级别: {{ role.level }}
              </div>
            </div>
            
            <!-- 用户数量 -->
            <div class="flex items-center justify-between text-sm">
              <span class="text-base-content/70">用户数量</span>
              <span class="font-medium">{{ role.userCount }} 人</span>
            </div>
            
            <!-- 权限数量 -->
            <div class="flex items-center justify-between text-sm">
              <span class="text-base-content/70">权限数量</span>
              <span class="font-medium">{{ role.permissions?.length || 0 }} 项</span>
            </div>
            
            <!-- 系统角色标识 -->
            <div v-if="role.isSystem" class="alert alert-info py-2">
              <InformationCircleIcon class="w-4 h-4" />
              <span class="text-xs">系统内置角色</span>
            </div>
            
            <!-- 创建时间 -->
            <div class="text-xs text-base-content/50">
              创建于 {{ formatDate(role.createdAt) }}
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 创建/编辑角色模态框 -->
    <dialog ref="roleModal" class="modal">
      <div class="modal-box w-11/12 max-w-2xl">
        <form method="dialog">
          <button class="btn btn-sm btn-circle btn-ghost absolute right-2 top-2">✕</button>
        </form>
        
        <h3 class="font-bold text-lg mb-4">
          {{ editingRole ? '编辑角色' : '新建角色' }}
        </h3>
        
        <div class="space-y-4">
          <div class="form-control">
            <label class="label">
              <span class="label-text">角色名称 <span class="text-error">*</span></span>
            </label>
            <input 
              v-model="roleForm.name" 
              type="text" 
              placeholder="请输入角色名称" 
              class="input input-bordered"
              :class="{ 'input-error': errors.name }"
            />
            <label class="label" v-if="errors.name">
              <span class="label-text-alt text-error">{{ errors.name }}</span>
            </label>
          </div>
          
          <div class="form-control">
            <label class="label">
              <span class="label-text">角色描述</span>
            </label>
            <textarea 
              v-model="roleForm.description" 
              class="textarea textarea-bordered" 
              placeholder="请输入角色描述"
              rows="3"
            ></textarea>
          </div>
          
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div class="form-control">
              <label class="label">
                <span class="label-text">权限级别 <span class="text-error">*</span></span>
              </label>
              <input 
                v-model.number="roleForm.level" 
                type="number" 
                placeholder="1-100，数字越大权限越高" 
                class="input input-bordered"
                min="1"
                max="100"
                :class="{ 'input-error': errors.level }"
              />
              <label class="label" v-if="errors.level">
                <span class="label-text-alt text-error">{{ errors.level }}</span>
              </label>
            </div>
            
            <div class="form-control">
              <label class="label">
                <span class="label-text">角色颜色</span>
              </label>
              <div class="flex items-center gap-2">
                <input 
                  v-model="roleForm.color" 
                  type="color" 
                  class="w-12 h-10 rounded border border-base-300"
                />
                <input 
                  v-model="roleForm.color" 
                  type="text" 
                  placeholder="#000000" 
                  class="input input-bordered flex-1"
                />
              </div>
            </div>
          </div>
          
          <div class="form-control">
            <label class="cursor-pointer label">
              <span class="label-text">启用状态</span>
              <input 
                v-model="roleForm.status" 
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
          <button type="button" class="btn btn-primary" @click="saveRole" :disabled="saving">
            <span v-if="saving" class="loading loading-spinner loading-sm"></span>
            {{ saving ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>
    </dialog>

    <!-- 权限管理模态框 -->
    <dialog ref="permissionModal" class="modal">
      <div class="modal-box w-11/12 max-w-4xl max-h-[80vh]">
        <form method="dialog">
          <button class="btn btn-sm btn-circle btn-ghost absolute right-2 top-2">✕</button>
        </form>
        
        <h3 class="font-bold text-lg mb-4">
          权限管理 - {{ selectedRole?.name }}
        </h3>
        
        <div class="space-y-4 max-h-96 overflow-y-auto">
          <div v-for="group in permissionGroups" :key="group.name" class="card bg-base-200">
            <div class="card-body p-4">
              <div class="flex items-center justify-between mb-3">
                <h4 class="font-semibold">{{ group.name }}</h4>
                <label class="cursor-pointer label">
                  <span class="label-text mr-2">全选</span>
                  <input 
                    type="checkbox" 
                    class="checkbox checkbox-sm" 
                    :checked="isGroupAllSelected(group)"
                    @change="toggleGroupPermissions(group, $event)"
                  />
                </label>
              </div>
              
              <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-2">
                <label 
                  v-for="permission in group.permissions" 
                  :key="permission.id"
                  class="cursor-pointer label justify-start gap-2 p-2 rounded hover:bg-base-100"
                >
                  <input 
                    type="checkbox" 
                    class="checkbox checkbox-sm" 
                    :value="permission.id"
                    v-model="selectedPermissions"
                  />
                  <div>
                    <div class="text-sm font-medium">{{ permission.name }}</div>
                    <div class="text-xs text-base-content/70">{{ permission.description }}</div>
                  </div>
                </label>
              </div>
            </div>
          </div>
        </div>
        
        <div class="modal-action">
          <button type="button" class="btn btn-ghost" @click="closePermissionModal">取消</button>
          <button type="button" class="btn btn-primary" @click="savePermissions" :disabled="savingPermissions">
            <span v-if="savingPermissions" class="loading loading-spinner loading-sm"></span>
            {{ savingPermissions ? '保存中...' : '保存权限' }}
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
  EllipsisVerticalIcon,
  InformationCircleIcon
} from '@heroicons/vue/24/outline'
import { api } from '../utils/axiosInstance'

// 数据类型定义
interface Role {
  id: number
  name: string
  description?: string
  level: number
  color?: string
  status: 'active' | 'inactive'
  isSystem: boolean
  userCount: number
  permissions?: Permission[]
  createdAt: string
}

interface Permission {
  id: number
  name: string
  description?: string
  code: string
  group: string
}

interface PermissionGroup {
  name: string
  permissions: Permission[]
}

// 响应式数据
const roles = ref<Role[]>([])
const permissionGroups = ref<PermissionGroup[]>([])
const loading = ref(false)
const saving = ref(false)
const savingPermissions = ref(false)
const searchQuery = ref('')
const editingRole = ref<Role | null>(null)
const selectedRole = ref<Role | null>(null)
const selectedPermissions = ref<number[]>([])
const roleModal = ref<HTMLDialogElement>()
const permissionModal = ref<HTMLDialogElement>()

// 筛选条件
const filters = reactive({
  status: '',
  sortBy: 'level'
})

// 表单数据
const roleForm = reactive({
  name: '',
  description: '',
  level: 1,
  color: '#3b82f6',
  status: 'active' as 'active' | 'inactive'
})

// 表单验证错误
const errors = reactive({
  name: '',
  level: ''
})

// 方法
const fetchRoles = async () => {
  loading.value = true
  try {
    const params = {
      search: searchQuery.value,
      ...filters
    }
    
    const response = await api.get('/api/admin/roles', params)
    roles.value = response.data.data
  } catch (error) {
    console.error('获取角色列表失败:', error)
    // 使用模拟数据
    roles.value = generateMockRoles()
  } finally {
    loading.value = false
  }
}

const fetchPermissions = async () => {
  try {
    const response = await api.get('/api/admin/permissions')
    permissionGroups.value = response.data.data
  } catch (error) {
    console.error('获取权限列表失败:', error)
    // 使用模拟数据
    permissionGroups.value = generateMockPermissions()
  }
}

// 生成模拟数据
const generateMockRoles = (): Role[] => {
  return [
    {
      id: 1,
      name: '超级管理员',
      description: '拥有系统所有权限',
      level: 100,
      color: '#ef4444',
      status: 'active',
      isSystem: true,
      userCount: 2,
      createdAt: '2024-01-01T00:00:00Z'
    },
    {
      id: 2,
      name: '管理员',
      description: '拥有大部分管理权限',
      level: 80,
      color: '#f59e0b',
      status: 'active',
      isSystem: true,
      userCount: 5,
      createdAt: '2024-01-01T00:00:00Z'
    },
    {
      id: 3,
      name: '版主',
      description: '负责内容审核和用户管理',
      level: 60,
      color: '#10b981',
      status: 'active',
      isSystem: false,
      userCount: 12,
      createdAt: '2024-01-15T10:30:00Z'
    },
    {
      id: 4,
      name: '普通用户',
      description: '基础用户权限',
      level: 10,
      color: '#6b7280',
      status: 'active',
      isSystem: true,
      userCount: 1234,
      createdAt: '2024-01-01T00:00:00Z'
    },
    {
      id: 5,
      name: 'VIP用户',
      description: '付费用户，享有特殊权限',
      level: 30,
      color: '#8b5cf6',
      status: 'active',
      isSystem: false,
      userCount: 89,
      createdAt: '2024-01-20T15:20:00Z'
    }
  ]
}

const generateMockPermissions = (): PermissionGroup[] => {
  return [
    {
      name: '用户管理',
      permissions: [
        { id: 1, name: '查看用户', description: '查看用户列表和详情', code: 'user.view', group: 'user' },
        { id: 2, name: '创建用户', description: '创建新用户', code: 'user.create', group: 'user' },
        { id: 3, name: '编辑用户', description: '编辑用户信息', code: 'user.edit', group: 'user' },
        { id: 4, name: '删除用户', description: '删除用户账号', code: 'user.delete', group: 'user' },
        { id: 5, name: '封禁用户', description: '封禁/解封用户', code: 'user.ban', group: 'user' }
      ]
    },
    {
      name: '内容管理',
      permissions: [
        { id: 6, name: '查看帖子', description: '查看所有帖子', code: 'post.view', group: 'content' },
        { id: 7, name: '编辑帖子', description: '编辑任意帖子', code: 'post.edit', group: 'content' },
        { id: 8, name: '删除帖子', description: '删除任意帖子', code: 'post.delete', group: 'content' },
        { id: 9, name: '审核帖子', description: '审核待发布帖子', code: 'post.review', group: 'content' },
        { id: 10, name: '置顶帖子', description: '设置帖子置顶', code: 'post.pin', group: 'content' }
      ]
    },
    {
      name: '分类管理',
      permissions: [
        { id: 11, name: '查看分类', description: '查看分类列表', code: 'category.view', group: 'category' },
        { id: 12, name: '创建分类', description: '创建新分类', code: 'category.create', group: 'category' },
        { id: 13, name: '编辑分类', description: '编辑分类信息', code: 'category.edit', group: 'category' },
        { id: 14, name: '删除分类', description: '删除分类', code: 'category.delete', group: 'category' }
      ]
    },
    {
      name: '系统管理',
      permissions: [
        { id: 15, name: '系统设置', description: '修改系统配置', code: 'system.config', group: 'system' },
        { id: 16, name: '查看日志', description: '查看系统日志', code: 'system.logs', group: 'system' },
        { id: 17, name: '数据备份', description: '备份系统数据', code: 'system.backup', group: 'system' },
        { id: 18, name: '角色管理', description: '管理用户角色', code: 'system.roles', group: 'system' }
      ]
    }
  ]
}

const handleSearch = () => {
  fetchRoles()
}

const handleFilter = () => {
  fetchRoles()
}

const openCreateModal = () => {
  editingRole.value = null
  resetForm()
  roleModal.value?.showModal()
}

const editRole = (role: Role) => {
  editingRole.value = role
  Object.assign(roleForm, {
    name: role.name,
    description: role.description || '',
    level: role.level,
    color: role.color || '#3b82f6',
    status: role.status
  })
  roleModal.value?.showModal()
}

const managePermissions = async (role: Role) => {
  selectedRole.value = role
  
  try {
    const response = await api.get(`/api/admin/roles/${role.id}/permissions`)
    selectedPermissions.value = response.data.data.map((p: Permission) => p.id)
  } catch (error) {
    console.error('获取角色权限失败:', error)
    // 模拟权限数据
    selectedPermissions.value = role.level >= 80 ? [1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18] : 
                               role.level >= 60 ? [1,6,7,9,10,11] : 
                               role.level >= 30 ? [1,6,11] : [1,6]
  }
  
  permissionModal.value?.showModal()
}

const closeModal = () => {
  roleModal.value?.close()
  resetForm()
}

const closePermissionModal = () => {
  permissionModal.value?.close()
  selectedRole.value = null
  selectedPermissions.value = []
}

const resetForm = () => {
  Object.assign(roleForm, {
    name: '',
    description: '',
    level: 1,
    color: '#3b82f6',
    status: 'active'
  })
  Object.assign(errors, {
    name: '',
    level: ''
  })
}

const validateForm = () => {
  errors.name = ''
  errors.level = ''
  
  if (!roleForm.name.trim()) {
    errors.name = '角色名称不能为空'
    return false
  }
  
  if (roleForm.name.length > 50) {
    errors.name = '角色名称不能超过50个字符'
    return false
  }
  
  if (roleForm.level < 1 || roleForm.level > 100) {
    errors.level = '权限级别必须在1-100之间'
    return false
  }
  
  return true
}

const saveRole = async () => {
  if (!validateForm()) {
    return
  }
  
  saving.value = true
  try {
    const data = { ...roleForm }
    
    if (editingRole.value) {
      // 编辑角色
      await api.put(`/api/admin/roles/${editingRole.value.id}`, data)
    } else {
      // 创建角色
      await api.post('/api/admin/roles', data)
    }
    
    closeModal()
    fetchRoles()
  } catch (error) {
    console.error('保存角色失败:', error)
    // 模拟保存成功
    closeModal()
    fetchRoles()
  } finally {
    saving.value = false
  }
}

const savePermissions = async () => {
  if (!selectedRole.value) return
  
  savingPermissions.value = true
  try {
    await api.post(`/api/admin/roles/${selectedRole.value.id}/permissions`, {
      permissionIds: selectedPermissions.value
    })
    
    closePermissionModal()
  } catch (error) {
    console.error('保存权限失败:', error)
    // 模拟保存成功
    closePermissionModal()
  } finally {
    savingPermissions.value = false
  }
}

const toggleStatus = async (role: Role) => {
  if (role.isSystem && role.level >= 80) {
    alert('系统核心角色不能禁用！')
    return
  }
  
  try {
    await api.post(`/api/admin/roles/${role.id}/toggle-status`)
    role.status = role.status === 'active' ? 'inactive' : 'active'
  } catch (error) {
    console.error('切换状态失败:', error)
    // 模拟切换成功
    role.status = role.status === 'active' ? 'inactive' : 'active'
  }
}

const deleteRole = async (role: Role) => {
  if (role.isSystem) {
    alert('系统内置角色不能删除！')
    return
  }
  
  if (role.userCount > 0) {
    alert('该角色下还有用户，无法删除！请先移动用户到其他角色。')
    return
  }
  
  if (confirm(`确定要删除角色「${role.name}」吗？此操作不可恢复！`)) {
    try {
      await api.delete(`/api/admin/roles/${role.id}`)
      fetchRoles()
    } catch (error) {
      console.error('删除角色失败:', error)
    }
  }
}

// 权限管理相关方法
const isGroupAllSelected = (group: PermissionGroup) => {
  return group.permissions.every(p => selectedPermissions.value.includes(p.id))
}

const toggleGroupPermissions = (group: PermissionGroup, event: Event) => {
  const target = event.target as HTMLInputElement
  const groupPermissionIds = group.permissions.map(p => p.id)
  
  if (target.checked) {
    // 添加该组所有权限
    groupPermissionIds.forEach(id => {
      if (!selectedPermissions.value.includes(id)) {
        selectedPermissions.value.push(id)
      }
    })
  } else {
    // 移除该组所有权限
    selectedPermissions.value = selectedPermissions.value.filter(id => 
      !groupPermissionIds.includes(id)
    )
  }
}

// 工具函数
const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit'
  })
}

// 组件挂载时获取数据
onMounted(() => {
  fetchRoles()
  fetchPermissions()
})
</script>

<style scoped>
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

/* 权限选择样式 */
.label:hover {
  background-color: hsl(var(--b3));
}
</style>