<script setup lang="ts">
import {computed, h, onMounted, reactive, ref} from 'vue'
import type {FormInst, FormRules} from 'naive-ui'
import {NButton, NTag, useMessage} from 'naive-ui'
import {AddOutline, CreateOutline, SearchOutline, TrashOutline} from '@vicons/ionicons5'
import {getUserList} from "@/admin/utils/authService.ts";
import type {User} from "../types/adminInterfaces.ts";

const message = useMessage()
const searchText = ref('')
const showUserModal = ref(false)
const isEditing = ref(false)
const userFormRef = ref<FormInst | null>(null)
const loading = ref(false)
const users = ref([])

// 分页设置
const pagination = reactive({
  page: 1,
  pageSize: 10,
  itemCount: 0,
  onChange: (page: number) => {
    pagination.page = page
    fetchUsers()
  }
})

// 获取用户列表
const fetchUsers = async () => {
  loading.value = true
  try {
    /**
     * {
     page: pagination.page,
     pageSize: pagination.pageSize,
     search: searchText.value
     }
     */
    const response = await getUserList()
    users.value = response.result.list || []
    pagination.itemCount = response.result.total || 0
  } catch (error) {
    message.error('获取用户列表失败')
    console.error('获取用户列表失败:', error)
  } finally {
    loading.value = false
  }
}

// 监听搜索文本变化
const handleSearch = () => {
  pagination.page = 1
  fetchUsers()
}

// 初始化加载
onMounted(() => {
  fetchUsers()
})

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
    {required: true, message: '请输入用户名', trigger: 'blur'}
  ],
  email: [
    {required: true, message: '请输入邮箱', trigger: 'blur'},
    {type: 'email', message: '请输入有效的邮箱地址', trigger: 'blur'}
  ],
  password: [
    {required: true, message: '请输入密码', trigger: 'blur'}
  ]
}

// 角色选项
const roleOptions = [
  {label: '管理员', value: 'admin'},
  {label: '普通用户', value: 'user'},
  {label: '访客', value: 'guest'}
]

// 状态选项
const statusOptions = [
  {label: '正常', value: 'active'},
  {label: '禁用', value: 'disabled'},
  {label: '待验证', value: 'pending'}
]

// 过滤后的用户列表 - 不再需要本地过滤，由后端处理
const filteredUsers = computed(() => users.value)

// 表格列定义
const columns = [
  {
    title: 'ID',
    key: 'userId'
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
    render(row: User) {
      return row.roleList
    }
  },
  {
    title: '状态',
    key: 'status',
    render(row: User) {
      const statusMap = {
        0: {text: '正常', type: 'success'},
        1: {text: '禁用', type: 'error'},
      }
      const status = statusMap[row.status] || {text: row.status, type: 'default'}
      return h(
          NTag,
          {type: status.type},
          {default: () => status.text}
      )
    }
  },
  {
    title: '验证',
    key: 'validate',
    render(row: User) {
      const statusMap = {
        0: {text: '未验证', type: 'warning'},
        1: {text: '验证', type: 'success'},
      }
      const status = statusMap[row.validate] || {text: row.validate, type: 'default'}
      return h(
          NTag,
          {type: status.type},
          {default: () => status.text}
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
    render(row: any) {
      return h(
          'n-space',
          {},
          {
            default: () => [
              h(
                  NButton,
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
                  NButton,
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
  userFormRef.value?.validate(async (errors) => {
    if (!errors) {
      try {
        if (isEditing.value) {
          // 这里应该调用更新用户的API
          message.success('用户更新成功')
        } else {
          // 这里应该调用添加用户的API
          message.success('用户添加成功')
        }
        showUserModal.value = false
        // 刷新用户列表
        fetchUsers()
      } catch (error) {
        message.error(isEditing.value ? '更新用户失败' : '添加用户失败')
        console.error(error)
      }
    }
  })
}
</script>
<template>
  <div>
    <div class="action-bar">
      <n-space>
        <n-input
            v-model:value="searchText"
            placeholder="搜索用户"
            clearable
            @keydown.enter="handleSearch"
        >
          <template #prefix>
            <n-icon>
              <SearchOutline/>
            </n-icon>
          </template>
        </n-input>
        <n-button type="primary" @click="handleSearch">
          搜索
        </n-button>
        <n-button type="primary" @click="handleAddUser">
          <template #icon>
            <n-icon>
              <AddOutline/>
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
        :loading="loading"
        striped
    />

    <!-- 添加/编辑用户对话框 保持不变 -->
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


<style scoped>
.action-bar {
  margin-bottom: 16px;
}

.action-buttons {
  margin-top: 24px;
}
</style>
