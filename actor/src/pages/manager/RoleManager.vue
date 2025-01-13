<script setup lang="ts">
import {NFlex,NButton, NDataTable, NForm, NFormItem, NInput, NModal, NSelect, NSpace, NTag, useMessage} from 'naive-ui'
import {h, onMounted, ref} from 'vue'
import {getPermissionList, getRoleDel, getRoleList, getRoleSave} from '@/service/request'
import {Ref, UnwrapRef} from "@vue/reactivity";

const message = useMessage()
type RoleItem = {
  roleId: number | null
  roleName: string | null
  permissions: any
  effective: number | null
  createTime: string | null
}
const data: Ref<UnwrapRef<RoleItem[]>> = ref([])


let columns = [
  {
    title: 'id',
    key: 'roleId'
  },
  {
    title: '角色',
    key: 'roleName',
  },
  {
    title: '权限点',
    key: 'permissions',
    render(row: RoleItem) {
      if (!row.permissions) {
        return []
      }
      let res = []
      for (const rowKey in row.permissions) {
        console.log(rowKey, row.permissions[rowKey])
        res.push(h(
            NTag,
            {
              strong: true,
              tertiary: true,
              size: 'small',
            },
            {default: () => row.permissions[rowKey].name}
        ))
      }
      return res

    }
  },
  {
    title: '状态',
    key: 'effective',
    render: (row: RoleItem) => {
      return row.effective === 0 ? "无效" : "有效"
    }
  },
  {
    title: '创建时间',
    key: 'createTime'
  },
  {
    title: '操作',
    key: 'actions',
    render(row: RoleItem) {
      return [h(
          NButton,
          {
            strong: true,
            tertiary: true,
            size: 'small',
            onClick: () => {
              console.log(row.permissions)
              showModal.value = true
              addRole.value = {
                id: row.roleId.toString(),
                roleName: row.roleName,
                permissions: row.permissions.map(item => {
                  return item['id']
                })
              }
            }
          },
          {default: () => '编辑'}
      ),
        h(
            NButton,
            {
              strong: true,
              tertiary: true,
              size: 'small',
              onClick: () => {
                getRoleDel(row.roleId).then(r => {
                  getRoleListData()
                })
              }
            },
            {default: () => '删除'}
        )
      ]
    }
  }
]

let pagination = true
const rules = {}
let initData = {
  id: "",
  roleName: "",
  permissions: []
}
let addRole = ref(initData)
let showModal = ref(false);

function onPositiveClick() {
  getRoleSave(addRole.value.id === "" ? 0 : parseInt(addRole.value.id), addRole.value.roleName, addRole.value.permissions)
      .then(r => {
        getRoleListData()
      })
}

function onNegativeClick() {
  console.log(addRole.value)
}

let permissionOptions = ref([])

onMounted(async () => {
  let res = await getPermissionList()
  permissionOptions.value = res.result
  await getRoleListData()
})

async function getRoleListData() {
  let res = await getRoleList()
  data.value = res.result.list
}

function startAddRole() {
  showModal.value = true
  addRole.value = initData
}
</script>
<template>
  <n-modal
      v-model:show="showModal"
      style="width: 600px"
      preset="dialog"
      title="添加角色"
      content="你确认?"
      positive-text="确认"
      negative-text="算了"
      @positive-click="onPositiveClick"
      @negative-click="onNegativeClick"
  >

    <n-form
        ref="formRef"
        :model="addRole"
        :rules="rules"
        label-placement="left"
        label-width="auto"
        require-mark-placement="right-hanging"
        :style="{
          padding: '30px 0 0',
      maxWidth: '640px'
    }"
    >
      <n-form-item label="id" path="ip" v-if="addRole.id!==''">
        <n-input v-model:value="addRole.id" disabled></n-input>
      </n-form-item>
      <n-form-item label="角色名" path="ip">
        <n-input v-model:value="addRole.roleName"></n-input>
      </n-form-item>
      <n-form-item label="权限点" path="port">
        <n-select v-model:value="addRole.permissions" multiple :options="permissionOptions"></n-select>
      </n-form-item>
    </n-form>
  </n-modal>
  <n-flex vertical>
    <n-flex>
      <n-button @click="startAddRole"> 新增角色</n-button>
    </n-flex>
    <n-data-table
        :columns="columns"
        :data="data"
        :pagination="pagination"
        :bordered="false"
    />
  </n-flex>
</template>
<style scoped>
.carousel-img {
  width: 100%;
  height: 240px;
  object-fit: cover;
}
</style>
