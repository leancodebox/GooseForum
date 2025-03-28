<script setup lang="ts">
import { h, ref, reactive } from 'vue'
import {
  NSpace,
  NPageHeader,
  NButton,
  NCard,
  NDataTable,
  NModal,
  NForm,
  NFormItem,
  NInput,
  NCheckboxGroup,
  NCheckbox,
  useMessage
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import {getPermissionList, getRoleList} from "@/admin/utils/authService.ts";

const message = useMessage()
getPermissionList()
getRoleList()
// 角色数据
const roles = ref([
  {
    id: 1,
    name: '超级管理员',
    description: '拥有所有权限的管理员角色',
    permissions: ['user_manage', 'post_manage', 'category_manage', 'role_manage', 'system_settings'],
    createdAt: '2023-01-01'
  },
  {
    id: 2,
    name: '内容管理员',
    description: '负责管理帖子和分类的角色',
    permissions: ['post_manage', 'category_manage'],
    createdAt: '2023-01-02'
  },
  {
    id: 3,
    name: '用户管理员',
    description: '负责管理用户的角色',
    permissions: ['user_manage'],
    createdAt: '2023-01-03'
  }
])

// 表格列定义
const columns: DataTableColumns = [
  {
    title: 'ID',
    key: 'id'
  },
  {
    title: '角色名称',
    key: 'name'
  },
  {
    title: '描述',
    key: 'description'
  },
  {
    title: '权限',
    key: 'permissions',
    render(row) {
      const permissionMap = {
        user_manage: '用户管理',
        post_manage: '帖子管理',
        category_manage: '分类管理',
        role_manage: '角色管理',
        system_settings: '系统设置'
      }
      return h(
          'div',
          row.permissions.map(p => permissionMap[p] || p).join(', ')
      )
    }
  },
  {
    title: '创建时间',
    key: 'createdAt'
  },
  {
    title: '操作',
    key: 'actions',
    render(row) {
      return h(
          NSpace,
          {},
          {
            default: () => [
              h(
                  NButton,
                  {
                    size: 'small',
                    onClick: () => handleEdit(row)
                  },
                  { default: () => '编辑' }
              ),
              h(
                  NButton,
                  {
                    size: 'small',
                    type: 'error',
                    onClick: () => handleDelete(row)
                  },
                  { default: () => '删除' }
              )
            ]
          }
      )
    }
  }
]

// 分页设置
const pagination = {
  pageSize: 10
}

// 添加角色相关
const showAddModal = ref(false)
const formRef = ref(null)
const formModel = reactive({
  name: '',
  description: '',
  permissions: []
})

// 编辑角色相关
const showEditModal = ref(false)
const editFormRef = ref(null)
const editFormModel = reactive({
  id: null,
  name: '',
  description: '',
  permissions: []
})
const currentEditIndex = ref(-1)

// 表单验证规则
const rules = {
  name: {
    required: true,
    message: '请输入角色名称',
    trigger: 'blur'
  }
}

// 处理添加角色
const handleAddRole = () => {
  formRef.value?.validate((errors) => {
    if (!errors) {
      const newRole = {
        id: roles.value.length + 1,
        name: formModel.name,
        description: formModel.description,
        permissions: formModel.permissions,
        createdAt: new Date().toISOString().split('T')[0]
      }
      roles.value.push(newRole)
      message.success('角色添加成功')
      showAddModal.value = false
      // 重置表单
      formModel.name = ''
      formModel.description = ''
      formModel.permissions = []
    }
  })
}

// 处理编辑角色
const handleEdit = (row) => {
  currentEditIndex.value = roles.value.findIndex(r => r.id === row.id)
  editFormModel.id = row.id
  editFormModel.name = row.name
  editFormModel.description = row.description
  editFormModel.permissions = [...row.permissions]
  showEditModal.value = true
}

const handleEditRole = () => {
  editFormRef.value?.validate((errors) => {
    if (!errors && currentEditIndex.value !== -1) {
      roles.value[currentEditIndex.value] = {
        ...roles.value[currentEditIndex.value],
        name: editFormModel.name,
        description: editFormModel.description,
        permissions: editFormModel.permissions
      }
      message.success('角色更新成功')
      showEditModal.value = false
    }
  })
}

// 处理删除角色
const handleDelete = (row) => {
  if (row.id === 1) {
    message.error('超级管理员角色不能删除')
    return
  }

  const index = roles.value.findIndex(r => r.id === row.id)
  if (index !== -1) {
    roles.value.splice(index, 1)
    message.success('角色删除成功')
  }
}
</script>
<template>
  <div>
    <n-space vertical>
      <n-page-header subtitle="管理系统中的角色和权限">
        <template #title>
          角色管理
        </template>
        <template #extra>
          <n-button type="primary" @click="showAddModal = true">
            添加角色
          </n-button>
        </template>
      </n-page-header>

        <n-data-table
          :columns="columns"
          :data="roles"
          :pagination="pagination"
          :bordered="true"
          striped
        />
    </n-space>

    <!-- 添加角色对话框 -->
    <n-modal v-model:show="showAddModal" preset="dialog" title="添加角色">
      <n-form
        ref="formRef"
        :model="formModel"
        :rules="rules"
        label-placement="left"
        label-width="auto"
        require-mark-placement="right-hanging"
      >
        <n-form-item label="角色名称" path="name">
          <n-input v-model:value="formModel.name" placeholder="请输入角色名称" />
        </n-form-item>
        <n-form-item label="角色描述" path="description">
          <n-input
            v-model:value="formModel.description"
            type="textarea"
            placeholder="请输入角色描述"
          />
        </n-form-item>
        <n-form-item label="权限" path="permissions">
          <n-checkbox-group v-model:value="formModel.permissions">
            <n-space>
              <n-checkbox value="user_manage">用户管理</n-checkbox>
              <n-checkbox value="post_manage">帖子管理</n-checkbox>
              <n-checkbox value="category_manage">分类管理</n-checkbox>
              <n-checkbox value="role_manage">角色管理</n-checkbox>
              <n-checkbox value="system_settings">系统设置</n-checkbox>
            </n-space>
          </n-checkbox-group>
        </n-form-item>
      </n-form>
      <template #action>
        <n-space>
          <n-button @click="showAddModal = false">取消</n-button>
          <n-button type="primary" @click="handleAddRole">确定</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 编辑角色对话框 -->
    <n-modal v-model:show="showEditModal" preset="dialog" title="编辑角色">
      <n-form
        ref="editFormRef"
        :model="editFormModel"
        :rules="rules"
        label-placement="left"
        label-width="auto"
        require-mark-placement="right-hanging"
      >
        <n-form-item label="角色名称" path="name">
          <n-input v-model:value="editFormModel.name" placeholder="请输入角色名称" />
        </n-form-item>
        <n-form-item label="角色描述" path="description">
          <n-input
            v-model:value="editFormModel.description"
            type="textarea"
            placeholder="请输入角色描述"
          />
        </n-form-item>
        <n-form-item label="权限" path="permissions">
          <n-checkbox-group v-model:value="editFormModel.permissions">
            <n-space>
              <n-checkbox value="user_manage">用户管理</n-checkbox>
              <n-checkbox value="post_manage">帖子管理</n-checkbox>
              <n-checkbox value="category_manage">分类管理</n-checkbox>
              <n-checkbox value="role_manage">角色管理</n-checkbox>
              <n-checkbox value="system_settings">系统设置</n-checkbox>
            </n-space>
          </n-checkbox-group>
        </n-form-item>
      </n-form>
      <template #action>
        <n-space>
          <n-button @click="showEditModal = false">取消</n-button>
          <n-button type="primary" @click="handleEditRole">确定</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>
