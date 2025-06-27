<script setup lang="ts">
import {onMounted, reactive, ref} from 'vue'
import {PlusIcon, PencilIcon, TrashIcon} from '@heroicons/vue/24/outline'
import {getPermissionList, getRoleDel, getRoleList, getRoleSave} from '../utils/adminService'
import type {UserRole, Permissions, Label} from '../utils/adminInterfaces'

// 响应式数据
const roles = ref<UserRole[]>([])
const permissions = ref<Label[]>([])
const loading = ref(false)
const saving = ref(false)
const deleting = ref(false)

const editingRole = ref<UserRole | null>(null)
const roleModal = ref<HTMLDialogElement>()
const deleteModal = ref<HTMLDialogElement>()
const modalMode = ref<'create' | 'edit'>('create')
const roleToDelete = ref<UserRole | null>(null)

// 表单数据
const formData = reactive({
  roleId: 0,
  roleName: '',
  effective: 1,
  permissions: [] as number[]
})

// 方法
const fetchRoles = async () => {
  loading.value = true
  try {
    const response = await getRoleList()
    roles.value = response.result.list
  } catch (error) {
    console.error('获取角色列表失败:', error)
  } finally {
    loading.value = false
  }
}

const fetchPermissions = async () => {
  try {
    const response = await getPermissionList()
    permissions.value = response.result || []
  } catch (error) {
    console.error('获取权限列表失败:', error)
  }
}

// 组件挂载时获取数据
onMounted(() => {
  fetchRoles()
  fetchPermissions()
})

const resetForm = () => {
  formData.roleId = 0
  formData.roleName = ''
  formData.effective = 1
  formData.permissions = []
}

const openCreateModal = () => {
  modalMode.value = 'create'
  resetForm()
  editingRole.value = null
  roleModal.value?.showModal()
}

const openEditModal = (role: UserRole) => {
  modalMode.value = 'edit'
  editingRole.value = role
  formData.roleId = role.roleId
  formData.roleName = role.roleName
  formData.effective = role.effective
  formData.permissions = role.permissions.map(p => p.id)
  roleModal.value?.showModal()
}

const closeModal = () => {
  roleModal.value?.close()
  resetForm()
}

const saveRole = async () => {
  if (!formData.roleName.trim()) {
    alert('请输入角色名称')
    return
  }
  
  saving.value = true
  try {
    const id = modalMode.value === 'edit' ? formData.roleId : 0
    await getRoleSave(id, formData.roleName, formData.permissions)
    closeModal()
    await fetchRoles()
  } catch (error) {
    console.error('保存角色失败:', error)
    alert('保存失败，请重试')
  } finally {
    saving.value = false
  }
}

const openDeleteModal = (role: UserRole) => {
  roleToDelete.value = role
  deleteModal.value?.showModal()
}

const closeDeleteModal = () => {
  deleteModal.value?.close()
  roleToDelete.value = null
}

const confirmDelete = async () => {
  if (!roleToDelete.value) return
  
  deleting.value = true
  try {
    await getRoleDel(roleToDelete.value.roleId)
    closeDeleteModal()
    await fetchRoles()
  } catch (error) {
    console.error('删除角色失败:', error)
    alert('删除失败，请重试')
  } finally {
    deleting.value = false
  }
}

const getPermissionNames = (rolePermissions: Permissions[]) => {
  return rolePermissions.map(p => p.name).join(', ') || '无权限'
}

const getStatusText = (effective: number) => {
  return effective === 1 ? '生效' : '禁用'
}

const getStatusClass = (effective: number) => {
  return effective === 1 ? 'badge-success' : 'badge-error'
}
</script>

<template>
  <div class="space-y-6">
    <!-- 页面标题和操作 -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold text-base-content">角色管理</h1>
        <p class="text-base-content/70 mt-1">管理系统角色和权限</p>
      </div>
      <button class="btn btn-primary btn-sm" @click="openCreateModal">
        <PlusIcon class="w-4 h-4"/>
        新建角色
      </button>
    </div>

    <!-- 角色列表 -->
    <div class="card bg-base-100 shadow">
      <div class="card-body p-0">
        <!-- 加载状态 -->
        <div v-if="loading" class="flex justify-center py-12">
          <span class="loading loading-spinner loading-md"></span>
        </div>
        
        <!-- 空状态 -->
        <div v-else-if="roles.length === 0" class="text-center py-12">
          <div class="text-base-content/50 mb-4">暂无角色数据</div>
          <button class="btn btn-primary btn-sm" @click="openCreateModal">
            <PlusIcon class="w-4 h-4"/>
            创建第一个角色
          </button>
        </div>
        
        <!-- 角色表格 -->
        <div v-else class="overflow-x-auto">
          <table class="table table-zebra">
            <thead>
              <tr>
                <th class="w-16">ID</th>
                <th class="w-32">角色名称</th>
                <th class="w-24">状态</th>
                <th>权限</th>
                <th class="w-32">创建时间</th>
                <th class="w-24">操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="role in roles" :key="role.roleId">
                <td class="font-mono text-sm">{{ role.roleId }}</td>
                <td>
                  <div class="font-medium">{{ role.roleName }}</div>
                </td>
                <td>
                  <div class="badge badge-sm whitespace-nowrap" :class="getStatusClass(role.effective)">
                    {{ getStatusText(role.effective) }}
                  </div>
                </td>
                <td>
                  <div class="text-sm text-base-content/70 max-w-xs truncate" :title="getPermissionNames(role.permissions)">
                    {{ getPermissionNames(role.permissions) }}
                  </div>
                </td>
                <td class="text-sm text-base-content/70 whitespace-nowrap">
                  {{ new Date(role.createTime).toLocaleDateString() }}
                </td>
                <td>
                  <div class="flex gap-1">
                    <button class="btn btn-ghost btn-xs" @click="openEditModal(role)" title="编辑">
                      <PencilIcon class="w-3 h-3"/>
                    </button>
                    <button class="btn btn-ghost btn-xs text-error" @click="openDeleteModal(role)" title="删除">
                      <TrashIcon class="w-3 h-3"/>
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- 角色编辑模态框 -->
    <dialog ref="roleModal" class="modal">
      <div class="modal-box w-11/12 max-w-2xl">
        <form method="dialog">
          <button class="btn btn-sm btn-circle btn-ghost absolute right-2 top-2" @click="closeModal">✕</button>
        </form>
        
        <h3 class="font-bold text-lg mb-4">
          {{ modalMode === 'create' ? '新建角色' : '编辑角色' }}
        </h3>
        
        <div class="space-y-4">
          <!-- 角色名称 -->
          <div class="form-control">
            <label class="label">
              <span class="label-text">角色名称 <span class="text-error">*</span></span>
            </label>
            <input 
              v-model="formData.roleName" 
              type="text" 
              placeholder="请输入角色名称" 
              class="input input-bordered w-full"
              :disabled="saving"
            />
          </div>
          
          <!-- 状态 -->
          <div class="form-control">
            <label class="label">
              <span class="label-text">状态</span>
            </label>
            <select v-model="formData.effective" class="select select-bordered w-full" :disabled="saving">
              <option :value="1">生效</option>
              <option :value="0">禁用</option>
            </select>
          </div>
          
          <!-- 权限选择 -->
          <div class="form-control">
            <label class="label">
              <span class="label-text">权限配置</span>
            </label>
            <div class="border border-base-300 rounded-lg p-4 max-h-48 overflow-y-auto">
              <div v-if="permissions.length === 0" class="text-center text-base-content/50 py-4">
                暂无权限数据
              </div>
              <div v-else class="space-y-2">
                <label v-for="permission in permissions" :key="permission.value" class="flex items-center gap-2 cursor-pointer hover:bg-base-200 p-2 rounded">
                   <input 
                     type="checkbox" 
                     :value="permission.value" 
                     v-model="formData.permissions" 
                     class="checkbox checkbox-sm"
                     :disabled="saving"
                   />
                   <span class="text-sm">{{ permission.label || permission.name }}</span>
                 </label>
              </div>
            </div>
          </div>
        </div>
        
        <div class="modal-action">
          <button class="btn btn-ghost" @click="closeModal" :disabled="saving">取消</button>
          <button class="btn btn-primary" @click="saveRole" :disabled="saving">
            <span v-if="saving" class="loading loading-spinner loading-xs"></span>
            {{ saving ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>
    </dialog>

    <!-- 删除确认模态框 -->
    <dialog ref="deleteModal" class="modal">
      <div class="modal-box">
        <h3 class="font-bold text-lg mb-4">确认删除</h3>
        <p class="mb-4">
          确定要删除角色 <span class="font-medium text-primary">{{ roleToDelete?.roleName }}</span> 吗？
        </p>
        <p class="text-sm text-base-content/70 mb-6">
          此操作不可撤销，请谨慎操作。
        </p>
        
        <div class="modal-action">
          <button class="btn btn-ghost" @click="closeDeleteModal" :disabled="deleting">取消</button>
          <button class="btn btn-error" @click="confirmDelete" :disabled="deleting">
            <span v-if="deleting" class="loading loading-spinner loading-xs"></span>
            {{ deleting ? '删除中...' : '确认删除' }}
          </button>
        </div>
      </div>
    </dialog>
  </div>
</template>



<style scoped>

</style>