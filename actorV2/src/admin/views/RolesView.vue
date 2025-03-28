<script setup lang="ts">
import { h, ref, reactive, onMounted } from 'vue'
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
  useMessage,
  NPopconfirm, NTag
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { getPermissionList, getRoleList, getRoleSave, getRoleDel } from "@/admin/utils/authService.ts";
import type {UserRole} from "@/admin/types/adminInterfaces.ts";

const message = useMessage()
const loading = ref(false)
const roles = ref([])
const permissions = ref([])

// 获取角色列表
const fetchRoleList = async () => {
  loading.value = true
  try {
    const res = await getRoleList()
    if (res.code === 0) {
      roles.value = res.result.list
    } else {
      message.error(res.message || '获取角色列表失败')
    }
  } catch (error) {
    message.error('获取角色列表失败')
    console.error(error)
  } finally {
    loading.value = false
  }
}

// 获取权限列表
const fetchPermissionList = async () => {
  try {
    const res = await getPermissionList()
    if (res.code === 0) {
      permissions.value = res.result
    } else {
      message.error(res.message || '获取权限列表失败')
    }
  } catch (error) {
    message.error('获取权限列表失败')
    console.error(error)
  }
}

// 初始化时获取数据
onMounted(() => {
  fetchRoleList()
  fetchPermissionList()
})

// 表格列定义
const columns = [
  {
    title: 'ID',
    key: 'roleId'
  },
  {
    title: '角色名称',
    key: 'roleName'
  },
  {
    title: '是否有效',
    key: 'effective'
  },
  {
    title: '权限',
    key: 'permissions',
    render(row:UserRole) {
      let list = []
      list = row?.permissions?.map(item => {
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
    title: '操作',
    key: 'actions',
    render(row:UserRole) {
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
              NPopconfirm,
              {
                onPositiveClick: () => handleDelete(row)
              },
              {
                default: () => '确定要删除该角色吗？',
                trigger: () => h(
                  NButton,
                  {
                    size: 'small',
                    type: 'error',
                    disabled: row.roleId === 0
                  },
                  { default: () => '删除' }
                )
              }
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
  roleName: '',
  permissions: []
})

// 编辑角色相关
const showEditModal = ref(false)
const editFormRef = ref(null)
const editFormModel = reactive({
  id: null,
  roleName: '',
  permissions: []
})

// 表单验证规则
const rules = {
  roleName: {
    required: true,
    message: '请输入角色名称',
    trigger: 'blur'
  }
}

// 处理添加角色
const handleAddRole = () => {
  formRef.value?.validate(async (errors) => {
    if (!errors) {
      try {
        const res = await getRoleSave(0, formModel.roleName, formModel.permissions)
        if (res.code === 0) {
          message.success('角色添加成功')
          showAddModal.value = false
          // 重置表单
          formModel.roleName = ''
          formModel.permissions = []
          // 重新获取角色列表
          fetchRoleList()
        } else {
          message.error(res.message || '添加角色失败')
        }
      } catch (error) {
        message.error('添加角色失败')
        console.error(error)
      }
    }
  })
}

// 处理编辑角色
const handleEdit = (row) => {
  editFormModel.id = row.id
  editFormModel.roleName = row.roleName
  editFormModel.permissions = [...row.permissions]
  showEditModal.value = true
}

const handleEditRole = () => {
  editFormRef.value?.validate(async (errors) => {
    if (!errors) {
      try {
        const res = await getRoleSave(
          editFormModel.id,
          editFormModel.roleName,
          editFormModel.permissions
        )
        if (res.code === 0) {
          message.success('角色更新成功')
          showEditModal.value = false
          // 重新获取角色列表
          fetchRoleList()
        } else {
          message.error(res.message || '更新角色失败')
        }
      } catch (error) {
        message.error('更新角色失败')
        console.error(error)
      }
    }
  })
}

// 处理删除角色
const handleDelete = async (row) => {
  if (row.id === 1) {
    message.error('超级管理员角色不能删除')
    return
  }

  try {
    const res = await getRoleDel(row.id)
    if (res.code === 0) {
      message.success('角色删除成功')
      // 重新获取角色列表
      fetchRoleList()
    } else {
      message.error(res.message || '删除角色失败')
    }
  } catch (error) {
    message.error('删除角色失败')
    console.error(error)
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
        :loading="loading"
        striped
      />
    </n-space>

    <!-- 添加角色对话框 -->
    <n-modal v-model:show="showAddModal" preset="dialog" title="添加角色" style="width: 500px;">
      <n-form
        ref="formRef"
        :model="formModel"
        :rules="rules"
        label-placement="left"
        label-width="auto"
        require-mark-placement="right-hanging"
      >
        <n-form-item label="角色名称" path="roleName">
          <n-input v-model:value="formModel.roleName" placeholder="请输入角色名称" />
        </n-form-item>

        <n-form-item label="权限点" path="permissions">
          <n-select v-model:value="formModel.permissions" multiple :options="permissions"></n-select>
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
    <n-modal v-model:show="showEditModal" preset="dialog" title="编辑角色" style="width: 500px;">
      <n-form
        ref="editFormRef"
        :model="editFormModel"
        :rules="rules"
        label-placement="left"
        label-width="auto"
        require-mark-placement="right-hanging"
      >
        <n-form-item label="角色名称" path="roleName">
          <n-input v-model:value="editFormModel.roleName" placeholder="请输入角色名称" />
        </n-form-item>
        <n-form-item label="权限点" path="permissions">
          <n-select v-model:value="formModel.permissions" multiple :options="permissions"></n-select>
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
