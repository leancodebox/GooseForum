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
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3 sm:gap-4">
          <div class="form-control">
            <label class="label pb-1">
              <span class="label-text text-sm">搜索角色</span>
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
              <span class="label-text text-sm">权限级别</span>
            </label>
            <select v-model="filters.level" class="select select-bordered w-full" @change="handleFilter">
              <option value="">全部级别</option>
              <option value="1">超级管理员</option>
              <option value="2">管理员</option>
              <option value="3">版主</option>
              <option value="4">普通用户</option>
            </select>
          </div>
          
          <div class="form-control">
            <label class="label pb-1">
              <span class="label-text text-sm">排序方式</span>
            </label>
            <select v-model="filters.sortBy" class="select select-bordered w-full" @change="handleFilter">
              <option value="level">权限级别</option>
              <option value="name">角色名称</option>
              <option value="created_at">创建时间</option>
              <option value="user_count">用户数量</option>
            </select>
          </div>
        </div>
      </div>
    </div>

    <!-- 角色列表 -->
    <div class="card bg-base-100 shadow">
      <div class="card-body p-0">
        <div class="space-y-0">
          <div v-for="(role, index) in roles" :key="role.roleId" 
               class="flex items-center justify-between p-4 hover:bg-base-200 transition-colors"
               :class="{ 'border-b border-base-300': index < roles.length - 1 }">
            <!-- 左侧：角色信息 -->
            <div class="flex items-center gap-4 flex-1">
              <div class="avatar placeholder">
                <div class="bg-primary text-primary-content rounded-full w-12">
                  <span class="text-lg">{{ role.roleName.charAt(0).toUpperCase() }}</span>
                </div>
              </div>
              
              <div class="flex-1">
                <div class="flex items-center gap-3 mb-1">
                  <h3 class="font-semibold text-lg">{{ role.roleName }}</h3>
                  <div class="badge badge-success badge-sm">启用</div>
                  <div class="badge badge-outline badge-sm">级别: {{ role.effective }}</div>
                  <div v-if="role.effective >= 80" class="badge badge-info badge-sm">
                    <InformationCircleIcon class="w-3 h-3 mr-1" />
                    系统角色
                  </div>
                </div>
                
                <p class="text-sm text-base-content/70 mb-2">暂无描述</p>
                
                <div class="flex items-center gap-6 text-xs text-base-content/60">
                  <span>用户数量: <span class="font-medium">0 人</span></span>
                  <span>权限数量: <span class="font-medium">0 项</span></span>
                  <span>创建于: {{ formatDate(role.createTime) }}</span>
                </div>
              </div>
            </div>
            
            <!-- 右侧：操作按钮 -->
            <div class="flex items-center gap-2">
              <button class="btn btn-ghost btn-sm" @click="editRole(role)" title="编辑角色">
                编辑
              </button>
              <button class="btn btn-ghost btn-sm" @click="managePermissions(role)" title="权限管理">
                权限
              </button>
              <button class="btn btn-ghost btn-sm" @click="toggleStatus(role)" title="切换状态">
                {{ role.effective >= 80 ? '系统' : '禁用' }}
              </button>
              <button v-if="role.effective < 80" 
                      class="btn btn-ghost btn-sm text-error" 
                      @click="deleteRole(role)" 
                      title="删除角色">
                删除
              </button>
            </div>
          </div>
          
          <!-- 空状态 -->
          <div v-if="roles.length === 0" class="text-center py-12">
            <div class="text-base-content/50 mb-2">暂无角色数据</div>
            <button class="btn btn-primary btn-sm" @click="openCreateModal">
              <PlusIcon class="w-4 h-4" />
              创建第一个角色
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 统一角色管理模态框 -->
    <dialog ref="roleModal" class="modal">
      <div class="modal-box w-11/12 max-w-5xl max-h-[90vh]">
        <form method="dialog">
          <button class="btn btn-sm btn-circle btn-ghost absolute right-2 top-2" @click="closeModal">✕</button>
        </form>
        
        <!-- 模态框标题 -->
        <div class="flex items-center justify-between mb-6">
          <h3 class="font-bold text-xl">
            {{ modalMode === 'create' ? '新建角色' : modalMode === 'edit' ? '编辑角色' : '权限管理' }}
            <span v-if="modalMode === 'permission'" class="text-base font-normal text-base-content/70">- {{ selectedRole?.roleName }}</span>
          </h3>
          
          <!-- 步骤指示器 -->
          <div v-if="modalMode !== 'permission'" class="flex items-center gap-2">
            <div class="flex items-center gap-1">
              <div class="w-8 h-8 rounded-full bg-primary text-primary-content flex items-center justify-center text-sm font-medium">
                1
              </div>
              <span class="text-sm font-medium">基本信息</span>
            </div>
            <div class="w-8 border-t border-base-300"></div>
            <div class="flex items-center gap-1">
              <div class="w-8 h-8 rounded-full border-2 border-base-300 text-base-content/50 flex items-center justify-center text-sm font-medium">
                2
              </div>
              <span class="text-sm text-base-content/50">权限配置</span>
            </div>
          </div>
        </div>
        
        <!-- 基本信息表单 -->
        <div v-if="modalMode === 'create' || modalMode === 'edit'" class="space-y-6">
          <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <!-- 左侧：基本信息 -->
            <div class="space-y-4">
              <div class="form-control">
                <label class="label">
                  <span class="label-text font-medium">角色名称 <span class="text-error">*</span></span>
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
                  <span class="label-text font-medium">角色描述</span>
                </label>
                <textarea 
                  v-model="roleForm.description" 
                  class="textarea textarea-bordered" 
                  placeholder="请输入角色描述"
                  rows="4"
                ></textarea>
              </div>
            </div>
            
            <!-- 右侧：配置选项 -->
            <div class="space-y-4">
              <div class="form-control">
                <label class="label">
                  <span class="label-text font-medium">权限级别 <span class="text-error">*</span></span>
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
                <label class="label">
                  <span class="label-text-alt text-base-content/60">级别越高权限越大，系统角色级别≥80</span>
                </label>
              </div>
              
              <div class="form-control">
                <label class="label">
                  <span class="label-text font-medium">角色颜色</span>
                </label>
                <div class="flex items-center gap-3">
                  <input 
                    v-model="roleForm.color" 
                    type="color" 
                    class="w-12 h-12 rounded-lg border border-base-300 cursor-pointer"
                  />
                  <input 
                    v-model="roleForm.color" 
                    type="text" 
                    placeholder="#3b82f6" 
                    class="input input-bordered flex-1"
                  />
                </div>
              </div>
              
              <div class="form-control">
                <label class="cursor-pointer label justify-start gap-3">
                  <input 
                    v-model="roleForm.status" 
                    type="checkbox" 
                    class="toggle toggle-primary" 
                    true-value="active"
                    false-value="inactive"
                  />
                  <div>
                    <span class="label-text font-medium">启用状态</span>
                    <div class="text-xs text-base-content/60">禁用后该角色用户将无法使用相关功能</div>
                  </div>
                </label>
              </div>
            </div>
          </div>
        </div>
        
        <!-- 权限管理界面 -->
        <div v-if="modalMode === 'permission'" class="space-y-4">
          <!-- 权限统计 -->
          <div class="stats shadow bg-base-200">
            <div class="stat">
              <div class="stat-title">已选权限</div>
              <div class="stat-value text-primary">{{ selectedPermissions.length }}</div>
              <div class="stat-desc">共 {{ getTotalPermissions() }} 项权限</div>
            </div>
            <div class="stat">
              <div class="stat-title">权限级别</div>
              <div class="stat-value text-secondary">{{ selectedRole?.effective }}</div>
              <div class="stat-desc">{{ selectedRole?.effective >= 80 ? '系统角色' : '普通角色' }}</div>
            </div>
          </div>
          
          <!-- 权限分组 -->
          <div class="max-h-96 overflow-y-auto space-y-3">
            <div v-for="group in permissionGroups" :key="group.name" class="card bg-base-100 border border-base-200">
              <div class="card-body p-4">
                <div class="flex items-center justify-between mb-4">
                  <h4 class="font-semibold text-lg flex items-center gap-2">
                    <div class="w-2 h-2 rounded-full bg-primary"></div>
                    {{ group.name }}
                  </h4>
                  <label class="cursor-pointer label gap-2">
                    <span class="label-text text-sm">全选</span>
                    <input 
                      type="checkbox" 
                      class="checkbox checkbox-primary checkbox-sm" 
                      :checked="isGroupAllSelected(group)"
                      @change="toggleGroupPermissions(group, $event)"
                    />
                  </label>
                </div>
                
                <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
                  <label 
                    v-for="permission in group.permissions" 
                    :key="permission.id"
                    class="cursor-pointer flex items-start gap-3 p-3 rounded-lg border border-base-200 hover:border-primary hover:bg-primary/5 transition-all"
                  >
                    <input 
                      type="checkbox" 
                      class="checkbox checkbox-primary checkbox-sm mt-0.5" 
                      :value="permission.id"
                      v-model="selectedPermissions"
                    />
                    <div class="flex-1">
                      <div class="font-medium text-sm">{{ permission.name }}</div>
                      <div class="text-xs text-base-content/60 mt-1">{{ permission.description }}</div>
                    </div>
                  </label>
                </div>
              </div>
            </div>
          </div>
        </div>
        
        <!-- 操作按钮 -->
        <div class="modal-action mt-8 pt-4 border-t border-base-200">
          <button type="button" class="btn btn-ghost" @click="closeModal">
            取消
          </button>
          
          <div class="flex gap-2">
            <button 
              v-if="modalMode === 'create' || modalMode === 'edit'" 
              type="button" 
              class="btn btn-outline btn-primary" 
              @click="saveRoleAndManagePermissions"
              :disabled="saving"
            >
              <span v-if="saving" class="loading loading-spinner loading-sm"></span>
              {{ saving ? '保存中...' : '保存并配置权限' }}
            </button>
            
            <button 
              type="button" 
              class="btn btn-primary" 
              @click="modalMode === 'permission' ? savePermissions() : saveRole()"
              :disabled="saving || savingPermissions"
            >
              <span v-if="saving || savingPermissions" class="loading loading-spinner loading-sm"></span>
              {{ 
                saving || savingPermissions ? '保存中...' : 
                modalMode === 'permission' ? '保存权限' : '保存角色'
              }}
            </button>
          </div>
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
import { 
  getRoleList, 
  getPermissionList, 
  getRoleSave, 
  getRoleDel 
} from '../utils/adminService'
import type { UserRole } from '../utils/adminInterfaces'



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
const roles = ref<UserRole[]>([])
const permissionGroups = ref<PermissionGroup[]>([])
const loading = ref(false)
const saving = ref(false)
const savingPermissions = ref(false)
const searchQuery = ref('')
const editingRole = ref<UserRole | null>(null)
const selectedRole = ref<UserRole | null>(null)
const selectedPermissions = ref<number[]>([])
const roleModal = ref<HTMLDialogElement>()
const modalMode = ref<'create' | 'edit' | 'permission'>('create')

// 筛选条件
const filters = reactive({
  status: '',
  level: '',
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
    const response = await getRoleList()
    console.log(response.result.list)
    roles.value = response.result.list
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
    const response = await getPermissionList()
    // 将权限列表转换为分组格式
    const permissions = response.result || []
    permissionGroups.value = groupPermissions(permissions)
  } catch (error) {
    console.error('获取权限列表失败:', error)
    // 使用模拟数据
    permissionGroups.value = generateMockPermissions()
  }
}

// 将权限列表转换为分组格式
const groupPermissions = (permissions: any[]) => {
  const groups: { [key: string]: Permission[] } = {}
  
  permissions.forEach(permission => {
    const groupName = permission.group || '其他权限'
    if (!groups[groupName]) {
      groups[groupName] = []
    }
    groups[groupName].push({
      id: permission.id,
      name: permission.name,
      description: permission.description,
      code: permission.code,
      group: groupName
    })
  })
  
  return Object.keys(groups).map(groupName => ({
    name: groupName,
    permissions: groups[groupName]
  }))
}

// 生成模拟数据
const generateMockRoles = (): UserRole[] => {
  return [
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
  modalMode.value = 'create'
  editingRole.value = null
  resetForm()
  roleModal.value?.showModal()
}

const editRole = (role: UserRole) => {
  modalMode.value = 'edit'
  editingRole.value = role
  selectedRole.value = role
  Object.assign(roleForm, {
    name: role.roleName,
    description: '', // UserRole接口中没有description属性
    level: role.effective,
    color: '#3b82f6', // UserRole接口中没有color属性
    status: 'active' // UserRole接口中没有status属性
  })
  roleModal.value?.showModal()
}

const managePermissions = async (role: UserRole) => {
  modalMode.value = 'permission'
  selectedRole.value = role
  
  // 根据角色级别模拟权限数据
  selectedPermissions.value = role.effective >= 80 ? [1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18] : 
                             role.effective >= 60 ? [1,6,7,9,10,11] : 
                             role.effective >= 30 ? [1,6,11] : [1,6]
  
  await fetchPermissions()
  roleModal.value?.showModal()
}

const closeModal = () => {
  roleModal.value?.close()
  resetForm()
  selectedRole.value = null
  selectedPermissions.value = []
  modalMode.value = 'create'
}

const saveRoleAndManagePermissions = async () => {
  if (!validateForm()) return
  
  try {
    saving.value = true
    const roleId = editingRole.value?.roleId || null
    const roleName = roleForm.name
    const permissions: number[] = []
    
    await getRoleSave(roleId, roleName, permissions)
    await fetchRoles()
    
    // 切换到权限管理模式
    modalMode.value = 'permission'
    if (!editingRole.value) {
      // 新建角色时，需要设置selectedRole
      const newRole = roles.value.find(r => r.roleName === roleName)
      if (newRole) {
        selectedRole.value = newRole
      }
    }
    selectedPermissions.value = selectedRole.value?.permissions?.map(p => p.id) || []
    await fetchPermissions()
  } catch (error) {
    console.error('保存角色失败:', error)
    // 模拟保存成功
    await fetchRoles()
    modalMode.value = 'permission'
    if (!editingRole.value) {
      const newRole = roles.value.find(r => r.roleName === roleForm.name)
      if (newRole) {
        selectedRole.value = newRole
      }
    }
    selectedPermissions.value = selectedRole.value?.permissions?.map(p => p.id) || []
    await fetchPermissions()
  } finally {
    saving.value = false
  }
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
    const roleId = editingRole.value?.roleId || null
    const roleName = roleForm.name
    const permissions: number[] = [] // 暂时为空，权限单独管理
    
    await getRoleSave(roleId, roleName, permissions)
    
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

const getTotalPermissions = () => {
  return permissionGroups.value.reduce((total, group) => total + group.permissions.length, 0)
}

const savePermissions = async () => {
  if (!selectedRole.value) return
  
  savingPermissions.value = true
  try {
    // 使用getRoleSave接口保存角色权限
    await getRoleSave(selectedRole.value.roleId, selectedRole.value.roleName, selectedPermissions.value)
    await fetchRoles()
    closeModal()
  } catch (error) {
    console.error('保存权限失败:', error)
    // 模拟保存成功
    await fetchRoles()
    closeModal()
  } finally {
    savingPermissions.value = false
  }
}

const toggleStatus = async (role: UserRole) => {
  if (role.effective >= 80) {
    alert('系统核心角色不能禁用！')
    return
  }
  
  try {
    // 暂时模拟切换成功，后续可添加对应的API接口
    console.log('切换角色状态:', role.roleName)
  } catch (error) {
    console.error('切换状态失败:', error)
  }
}

const deleteRole = async (role: UserRole) => {
  if (role.effective >= 80) {
    alert('系统内置角色不能删除！')
    return
  }
  
  if (confirm(`确定要删除角色「${role.roleName}」吗？此操作不可恢复！`)) {
    try {
      await getRoleDel(role.roleId)
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