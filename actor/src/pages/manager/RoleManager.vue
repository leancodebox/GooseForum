<script setup lang="ts">
import {NButton, NDataTable, NSpace, useMessage} from 'naive-ui'
import {h, ref} from 'vue'
import {getUserList} from '@/service/request'
import {Ref, UnwrapRef} from "@vue/reactivity";

const message = useMessage()
type UserItem = {
  userId: number | null
  username: string | null
  email: string | null
  createTime: string | null
  status: string | null
}
const data: Ref<UnwrapRef<UserItem[]>> = ref([
  {userId: 3, username: '张三', email: '4:18', createTime: "", status: ""},
])


let columns = [
  {
    title: 'id',
    key: 'userId'
  },
  {
    title: '角色',
    key: 'role',
  },
  {
    title: '权限点',
    key: 'permission'
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
    title: 'Action',
    key: 'actions',
    render(row: UserItem) {
      return [h(
          NButton,
          {
            strong: true,
            tertiary: true,
            size: 'small',
            onClick: () => {
              message.info(`Play ${row.username}`)
            }
          },
          {default: () => '编辑'}
      ), h(
          NButton,
          {
            strong: true,
            tertiary: true,
            size: 'small',
            onClick: () => {
              message.info(`Play ${row.username}`)
            }
          },
          {default: () => '冻结'}
      ),
        h(
            NButton,
            {
              strong: true,
              tertiary: true,
              size: 'small',
              onClick: () => {
                message.info(`Play ${row.username}`)
              }
            },
            {default: () => '删除'}
        )
      ]
    }
  }
]
getUserList().then(r => {
  data.value = r.result.list.map(item => {
    return {userId: item.userId, username: item.username, email: item.email,createTime:item.createTime, status: ""}
  })
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
