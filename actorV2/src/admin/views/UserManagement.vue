<script setup lang="ts">
import {computed, h, onMounted, reactive, ref} from 'vue'
import {type FormInst, type FormRules, NButton, NSpace, NTag, useMessage} from 'naive-ui'
import {AddOutline, CreateOutline, SearchOutline, TrashOutline} from '@vicons/ionicons5'
import {editUser, getAllRoleItem, getUserList} from "@/admin/utils/authService.ts";
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

let roleOption = ref([])

// 初始化加载
onMounted(async () => {
  await fetchUsers()
  let res = await getAllRoleItem()
  roleOption.value = res.result
})

// 用户表单
const userForm = reactive({
  id: 0,
  username: '',
  email: '',
  role: 0,
  status: 0,
  validate: 0,
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
      let list = []
      list = row?.roleList?.map(item => {
        return h(
            NTag,
            {
              // size: 'small',
            },
            () => item.name
        )
      })
      return h(
          NSpace,
          {},
          {
            default: () => list
          })
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
    key: 'createTime'
  },
  {
    title: '操作',
    key: 'actions',
    render(row: User) {
      return h(
          NSpace,
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
  userForm.id = 0
  userForm.username = ''
  userForm.email = ''
  userForm.role = 0
  userForm.status = 0
  userForm.password = ''
  userForm.validate = 0
  showUserModal.value = true
}

// 编辑用户
const handleEditUser = (user: User) => {
  isEditing.value = true
  userForm.id = user.userId
  userForm.username = user.username
  userForm.email = user.email
  userForm.role = user.roleId
  userForm.status = user.status
  userForm.password = ''
  userForm.validate = user.validate
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

          await editUser(userForm.id, userForm.status, userForm.validate, userForm.role)
          // 这里应该调用更新用户的API
          message.success('用户更新成功')
        } else {
          // 这里应该调用添加用户的API
          message.success('用户添加成功')
        }
        showUserModal.value = false
        // 刷新用户列表
        await fetchUsers()
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
        :bordered="true"
        :loading="loading"
        striped
    />

    <!-- 添加/编辑用户对话框 -->
    <n-modal
        v-model:show="showUserModal"
        :title="isEditing ? '编辑用户' : '添加用户'"
        style="width: 500px; max-width: 90%;"
    >
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
            <n-input v-model:value="userForm.username" placeholder="请输入用户名"/>
          </n-form-item>

          <n-form-item path="email" label="邮箱">
            <n-input v-model:value="userForm.email" placeholder="请输入邮箱" disabled/>
          </n-form-item>

          <n-form-item label="角色" path="roleId">
            <n-select v-model:value="userForm.role"  :options="roleOption" clearable></n-select>
          </n-form-item>

          <n-form-item path="status" label="是否冻结">
            <n-switch v-model:value="userForm.status" :checked-value="1" :unchecked-value="0"/>
          </n-form-item>

          <n-form-item path="validate" label="验证通过">
            <n-switch v-model:value="userForm.validate" :checked-value="1" :unchecked-value="0"/>
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
