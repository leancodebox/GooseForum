<script setup lang="ts">
import {NButton, NDataTable, NSpace, NTag, useMessage} from 'naive-ui'
import {h, ref} from 'vue'
import {getRoleList} from '@/service/request'
import {Ref, UnwrapRef} from "@vue/reactivity";

const message = useMessage()
type RoleItem = {
  roleId: number | null
  roleName: string | null
  permissions: any
  status: string | null
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
    key: 'status'
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
              message.info(`Play ${row.roleName}`)
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
                message.info(`Play ${row.roleName}`)
              }
            },
            {default: () => '删除'}
        )
      ]
    }
  }
]
getRoleList().then(r => {
  data.value = r.result.list
  // console.log(r)
})
let pagination = true
</script>
<template>
  <n-space vertical>
    <n-data-table
        :columns="columns"
        :data="data"
        :pagination="pagination"
        :bordered="false"
    />
  </n-space>
</template>
<style>
.carousel-img {
  width: 100%;
  height: 240px;
  object-fit: cover;
}
</style>
