<script setup lang="ts">
import {onMounted, reactive, ref} from 'vue'
import {PlusIcon} from '@heroicons/vue/24/outline'
import {getPermissionList, getRoleDel, getRoleList, getRoleSave} from '../utils/adminService'
import type {UserRole} from '../utils/adminInterfaces'


interface Permission {
  id: number
  name: string
  description?: string
  code: string
  group: string
}
// 响应式数据
const roles = ref<UserRole[]>([])
const loading = ref(false)

const editingRole = ref<UserRole | null>(null)
const roleModal = ref<HTMLDialogElement>()
const modalMode = ref<'create' | 'edit' | 'permission'>('create')


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
  } catch (error) {
    console.error('获取权限列表失败:', error)
  }
}

// 组件挂载时获取数据
onMounted(() => {
  fetchRoles()
  fetchPermissions()
})

const openCreateModal = () => {
  modalMode.value = 'create'
  editingRole.value = null
  roleModal.value?.showModal()
}

// getRoleSave saveOrUpdate
// getRoleDel delRole
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
      <div class="card-body p-0 space-y-0">
        <!-- 空状态 -->
        <div v-if="roles.length === 0" class="text-center py-12">
          <div class="text-base-content/50 mb-2">暂无角色数据</div>
          <button class="btn btn-primary btn-sm" @click="openCreateModal">
            <PlusIcon class="w-4 h-4"/>
            创建第一个角色
          </button>
        </div>
        <table class="table table-sm" v-else>
          <thead>
          <tr>
            <th>id</th>
            <th>角色名</th>
            <th>生效</th>
            <th>权限</th>
            <th>操作</th>
          </tr>
          </thead>
          <tr v-for="(role, index) in roles" :key="role.roleId">
            <td>{{ role.roleId }}</td>
            <td>{{ role.roleName }}</td>
            <td>{{ role.effective }}</td>
            <td>{{ role.permissions }}</td>
          </tr>
        </table>
      </div>
    </div>

    <!-- 统一角色管理模态框 -->
    <dialog ref="roleModal" class="modal">

    </dialog>
  </div>
</template>



<style scoped>

</style>