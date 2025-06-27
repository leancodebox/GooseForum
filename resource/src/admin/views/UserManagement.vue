<template>
  <div class="space-y-6">

    <!-- 搜索和筛选 -->
    <div class="card bg-base-100 shadow">
      <div class="card-body p-3">
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3 sm:gap-4">
          <div class="form-control">
            <label class="floating-label">
              <span>用户名</span>
              <input v-model="searchQuery" type="text" placeholder="username" class="input input-sm w-full"  @input="handleSearch" />
            </label>
          </div>

          <div class="form-control">
            <label class="select w-full select-sm">
              <span class="label">封禁状态</span>
              <select v-model="filters.status" class="select select-bordered" @change="handleFilter">
                <option value=""></option>
                <option value="active">活跃</option>
                <option value="inactive">非活跃</option>
                <option value="banned">已封禁</option>
              </select>
            </label>
          </div>

          <div class="form-control">
            <label class="select select-sm w-full">
              <span class="label">验证状态</span>
              <select v-model="filters.dateRange" class="select select-bordered" @change="handleFilter">
                <option value=""></option>
                <option value="active">活跃</option>
                <option value="inactive">非活跃</option>
                <option value="banned">已封禁</option>
              </select>
            </label>
          </div>
          <div class="form-control flex justify-center">
            <button @click="fetchUsers" class="btn btn-primary btn-sm w-full" >搜索</button>
          </div>
        </div>
      </div>
    </div>

    <!-- 用户列表 -->
    <div class="card bg-base-100 shadow">
      <div class="card-body p-0">
        <div class="overflow-x-auto">
          <table class="table table-sm">
            <thead>
              <tr>
                <th>用户信息</th>
                <th>角色</th>
                <th>状态</th>
                <th>验证</th>
                <th class="w-32">注册时间</th>
                <th class="w-32">最后登录</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="user in users" :key="user.userId">
                <td>
                  <div class="flex items-center gap-3">
                    <div class="avatar">
                      <div class="mask mask-squircle w-12 h-12">
                        <img :src="user.avatarUrl || '/static/pic/default-avatar.png'" :alt="user.username" />
                      </div>
                    </div>
                    <div>
                      <div class="font-bold">{{ user.username }}</div>
                      <div class="text-sm opacity-50">{{ user.email }}</div>
                    </div>
                  </div>
                </td>
                <td>
                  <div class="badge badge-sm whitespace-nowrap badge-soft badge-primary" v-for="item in user.roleList">
                    {{ item.name }}
                  </div>
                </td>
                <td>
                  <div class="badge badge-sm badge-outline whitespace-nowrap"
                    :class="user.status === 1 ? 'badge-error' : 'badge-success'">
                    {{ user.status === 0 ? '正常' : '已封禁' }}
                  </div>
                </td>
                <td>
                  <div class="badge badge-sm badge-outline whitespace-nowrap"
                    :class="user.validate === 1 ? 'badge-success' : 'badge-warning'">
                    {{ user.validate === 1 ? '验证' : '未验证' }}
                  </div>
                </td>
                <td class="whitespace-nowrap">{{ formatDate(user.createTime) }}</td>
                <td class="whitespace-nowrap">{{ user.createTime ? formatDate(user.createTime) : '从未登录' }}</td>
                <td>
                  <div class="dropdown dropdown-end">
                    <div tabindex="0" role="button" class="btn btn-ghost btn-xs">
                      <EllipsisVerticalIcon class="w-4 h-4" />
                    </div>
                    <ul tabindex="0" class="dropdown-content z-[1] menu p-2 shadow bg-base-100 rounded-box w-52">
                      <li><a @click="editUserItem(user)">编辑</a></li>
                      <li><a @click="resetPassword(user)">重置密码</a></li>
                      <li v-if="user.status === 1">
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
            <button class="join-item btn btn-sm" :disabled="pagination.page <= 1"
              @click="changePage(pagination.page - 1)">
              上一页
            </button>
            <button v-for="page in visiblePages" :key="page" class="join-item btn btn-sm"
              :class="{ 'btn-active': page === pagination.page }" @click="changePage(page)">
              {{ page }}
            </button>
            <button class="join-item btn btn-sm" :disabled="pagination.page >= totalPages"
              @click="changePage(pagination.page + 1)">
              下一页
            </button>
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
            <label class="floating-label">
              <span class="label">用户名</span>
              <input v-model="userForm.username" placeholder="username" type="text" class="input" required disabled />
            </label>
          </div>

          <div class="form-control">
            <label class="floating-label">
              <span class="label">邮箱</span>
              <input v-model="userForm.email" placeholder="email@example.com" type="text" class="input" required
                disabled />
            </label>
          </div>

          <div class="form-control" v-if="!isEditing">
            <label class="label">
              <span class="label-text">密码 *</span>
            </label>
            <input v-model="userForm.password" type="password" class="input input-bordered" :required="!isEditing" />
          </div>

          <div class="form-control">
            <label class="label pr-4 p-2">
              <span class="label-text">封禁:</span>
            </label>
            <input v-model="userForm.status" type="checkbox" class="toggle toggle-error" 
             :true-value="1" 
            :false-value="0"/>
          </div>

          <div class="form-control">
            <label class="label pr-4 p-2">
              <span class="label-text">验证:</span>
            </label>
            <input v-model="userForm.validate" type="checkbox" class="toggle toggle-success" :true-value="1"
            :false-value="0"/>
          </div>
          <div class="form-control">
            <label class="select">
              <span class="label">角色</span>
              <select v-model="userForm.roleId">
                <option >无</option>
                
                <option :value="item.value" v-for="item in roleOption">{{item.name}}</option>
              </select>
            </label>
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
import { editUser, getAllRoleItem, getUserList } from '../utils/adminService.ts'
import type {
  Label,
  User,
} from '../utils/adminInterfaces.ts';

// 响应式数据
const users = ref<User[]>([])
const loading = ref(false)
const searchQuery = ref('')
const userModal = ref<HTMLDialogElement>()
const isEditing = ref(false)

// 筛选条件
const filters = reactive({
  status: '',
  dateRange: ''
})

// 分页
const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

// 用户表单
const userForm = reactive({
  userId: 0,
  username: '',
  email: '',
  password: '',
  roleId: 0,
  status: 0,
  validate:0,
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
const fetchUsers = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      pageSize: pagination.pageSize,
      search: searchQuery.value,
      ...filters
    }
    const response = await getUserList(pagination.page,pagination.pageSize)
    users.value = response.result.list

    pagination.total = response.result.total
  } catch (error) {
    console.error('获取用户列表失败:', error)

    pagination.total = 100
  } finally {
    loading.value = false
  }
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



const openCreateModal = () => {
  isEditing.value = false
  resetUserForm()
  userModal.value?.showModal()
}

const editUserItem = (user: User) => {
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
    await editUser(userForm.userId, userForm.status, userForm.validate, userForm.roleId)
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
      await api.post(`/api/admin/users/${user.userId}/reset-password`)
      alert('密码重置成功')
    } catch (error) {
      console.error('重置密码失败:', error)
    }
  }
}

const banUser = async (user: User) => {
  if (confirm(`确定要封禁用户 ${user.username} 吗？`)) {
    try {
      await api.post(`/api/admin/users/${user.userId}/ban`)
      fetchUsers()
    } catch (error) {
      console.error('封禁用户失败:', error)
    }
  }
}

const unbanUser = async (user: User) => {
  if (confirm(`确定要解除用户 ${user.username} 的封禁吗？`)) {
    try {
      await api.post(`/api/admin/users/${user.userId}/unban`)
      fetchUsers()
    } catch (error) {
      console.error('解除封禁失败:', error)
    }
  }
}

const deleteUser = async (user: User) => {
  if (confirm(`确定要删除用户 ${user.username} 吗？此操作不可恢复！`)) {
    try {
      await api.post('/api/admin/user-edit', { id: user.userId, action: 'delete' })
      fetchUsers()
    } catch (error) {
      console.error('删除用户失败:', error)
    }
  }
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
let roleOption = ref<Label[]>([])
// 组件挂载时获取数据
onMounted(async() => {
  fetchUsers()
  let res = await getAllRoleItem()
  roleOption.value = res.result
})
</script>

<style scoped></style>