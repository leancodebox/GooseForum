<template>
  <div>
    <div class="action-bar">
      <n-space>
        <n-input v-model:value="searchText" placeholder="搜索用户" clearable>
          <template #prefix>
            <n-icon>
              <SearchOutline />
            </n-icon>
          </template>
        </n-input>
        <n-button type="primary" @click="handleAddUser">
          <template #icon>
            <n-icon>
              <AddOutline />
            </n-icon>
          </template>
          添加用户
        </n-button>
      </n-space>
    </div>

    <n-data-table
      :columns="columns"
      :data="filteredUsers"
      :pagination="pagination"
      :bordered="false"
      striped
    />

    <!-- 添加/编辑用户对话框 -->
    <n-modal v-model:show="showUserModal" :title="isEditing ? '编辑用户' : '添加用户'">
      <n-card>
        <n-form
          ref="userFormRef"
          :model="userForm"
          :rules="userRules"
          label-placement="left"
          label-width="auto"
          require-mark-placement="right-hanging"
        >
          <n-form-item path="username" label="用户名">
            <n-input v-model:value="userForm.username" placeholder="请输入用户名" />
          </n-form-item>
          <n-form-item path="email" label="邮箱">
            <n-input v-model:value="userForm.email" placeholder="请输入邮箱" />
          </n-form-item>
          <n-form-item path="role" label="角色">
            <n-select v-model:value="userForm.role" :options="roleOptions" />
          </n-form-item>
          <n-form-item path="status" label="状态">
            <n-select v-model:value="userForm.status" :options="statusOptions" />
          </n-form-item>
          <n-form-item v-if="!isEditing" path="password" label="密码">
            <n-input
              v-model:value="userForm.password"
              type="password"
              placeholder="请输入密码"
            />
          </n-form-item>
        </n-form>
        <div class="action-buttons">
          <n-space justify="end">
            <n-button @click="showUserModal = false">取消</n-button>
            <n-button type="primary" @click="handleSaveUser">保存</n-button>
          </n-space>
        </div>
      </n-card>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { h, ref, computed, reactive } from 'vue'
import { useMessage } from 'naive-ui'
import { SearchOutline, AddOutline, CreateOutline, TrashOutline } from '@vicons/ionicons5'
import type { FormInst, FormRules } from 'naive-ui'

const message = useMessage()
const searchText = ref('')
const showUserModal = ref(false)
const isEditing = ref(false)
const userFormRef = ref<FormInst | null>(null)

// 用户表单
const userForm = reactive({
  id: '',
  username: '',
  email: '',
  role: 'user',
  status: 'active',
  password: ''
})

// 表单验证规则
const userRules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' }
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入有效的邮箱地址', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' }
  ]
}

// 角色选项
const roleOptions = [
  { label: '管理员', value: 'admin' },
  { label: '普通用户', value: 'user' },
  { label: '访客', value: 'guest' }
]

// 状态选项
const statusOptions = [
  { label: '正常', value: 'active' },
  { label: '禁用', value: 'disabled' },
  { label: '待验证', value: 'pending' }
]

// 分页设置
const pagination = {
  pageSize: 10
}

// 模拟用户数据
const users = ref([
  { id: '1', username: 'admin', email: 'admin@example.com', role: 'admin', status: 'active', createdAt: '2023-01-01' },
  { id: '2', username: 'user1', email: 'user1@example.com', role: 'user', status: 'active', createdAt: '2023-01-02' },
  { id: '3', username: 'user2', email: 'user2@example.com', role: 'user', status: 'disabled', createdAt: '2023-01-03' },
  { id: '4', username: 'guest', email: 'guest@example.com', role: 'guest', status: 'pending', createdAt: '2023-01-04' }
])

// 过滤后的用户列表
const filteredUsers = computed(() => {
  if (!searchText.value) return users.value
  
  return users.value.filter(user => 
    user.username.toLowerCase().includes(searchText.value.toLowerCase()) ||
    user.email.toLowerCase().includes(searchText.value.toLowerCase())
  )
})

// 表格列定义
const columns = [
  {
    title: 'ID',
    key: 'id'
  },
  {
    title: '用户名',
    key: 'username'
  },
  {
    title: '邮箱',
    key: 'email'
  },
  {
    title: '角色',
    key: 'role',
    render(row) {
      const roleMap = {
        admin: '管理员',
        user: '普通用户',
        guest: '访客'
      }
      return roleMap[row.role] || row.role
    }
  },
  {
    title: '状态',
    key: 'status',
    render(row) {
      const statusMap = {
        active: { text: '正常', type: 'success' },
        disabled: { text: '禁用', type: 'error' },
        pending: { text: '待验证', type: 'warning' }
      }
      const status = statusMap[row.status] || { text: row.status, type: 'default' }
      return h(
        'n-tag',
        { type: status.type },
        { default: () => status.text }
      )
    }
  },
  {
    title: '注册时间',
    key: 'createdAt'
  },
  {
    title: '操作',
    key: 'actions',
    render(row) {
      return h(
        'n-space',
        {},
        {
          default: () => [
            h(
              'n-button',
              {
                size: 'small',
                quaternary: true,
                onClick: () => handleEditUser(row)
              },
              {
                default: () => '编辑',
                icon: () => h(CreateOutline)
              }
            ),
            h(
              'n-button',
              {
                size: 'small',
                quaternary: true,
                type: 'error',
                onClick: () => handleDeleteUser(row)
              },
              {
                default: () => '删除',
                icon: () => h(TrashOutline)
              }
            )
          ]
        }
      )
    }
  }
]

// 添加用户
const handleAddUser = () => {
  isEditing.value = false
  userForm.id = ''
  userForm.username = ''
  userForm.email = ''
  userForm.role = 'user'
  userForm.status = 'active'
  userForm.password = ''
  showUserModal.value = true
}

// 编辑用户
const handleEditUser = (user) => {
  isEditing.value = true
  userForm.id = user.id
  userForm.username = user.username
  userForm.email = user.email
  userForm.role = user.role
  userForm.status = user.status
  userForm.password = ''
  showUserModal.value = true
}

// 删除用户
const handleDeleteUser = (user) => {
  message.warning(`此功能尚未实现：删除用户 ${user.username}`)
}

// 保存用户
const handleSaveUser = () => {
  userFormRef.value?.validate((errors) => {
    if (!errors) {
      if (isEditing.value) {
        // 更新用户
        const index = users.value.findIndex(u => u.id === userForm.id)
        if (index !== -1) {
          users.value[index] = { ...users.value[index], ...userForm }
          message.success('用户更新成功')
        }
      } else {
        // 添加用户
        const newId = (parseInt(users.value[users.value.length - 1].id) + 1).toString()
        users.value.push({
          id: newId,
          username: userForm.username,
          email: userForm.email,
          role: userForm.role,
          status: userForm.status,
          createdAt: new Date().toISOString().split('T')[0]
        })
        message.success('用户添加成功')
      }
      showUserModal.value = false
    }
  })
}
</script>

<style scoped>
.action-bar {
  margin-bottom: 16px;
}

.action-buttons {
  margin-top: 24px;
}
</style>