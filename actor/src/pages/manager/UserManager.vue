<script setup lang="ts">
import {NButton, NDataTable, NSpace, NTag, useMessage} from 'naive-ui'
import {h, ref} from 'vue'
import {getUserList} from '@/service/request'
import {Ref, UnwrapRef} from "@vue/reactivity";

const message = useMessage()
type UserItem = {
  userId: number | null
  username: string | null
  email: string | null
  createTime: string | null
  roleList: any
  status: string | null
}
const data: Ref<UnwrapRef<UserItem[]>> = ref([])


let columns = [
  {
    title: 'id',
    key: 'userId'
  },
  {
    title: 'username',
    key: 'username'
  },
  {
    title: 'email',
    key: 'email'
  },
  {
    title: '角色',
    key: 'role',
    render(row: UserItem) {
      console.log()
      if (!row.roleList) {
        return []
      }
      let res = []
      for (const rowKey in row.roleList) {
        console.log(rowKey, row.roleList[rowKey])
        res.push(h(
            NTag,
            {
              strong: true,
              tertiary: true,
              size: 'small',
            },
            {default: () => row.roleList[rowKey].name}
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
    title: '积分',
    key: 'ponits'
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
            {default: () => '重置密码'}
        )
      ]
    }
  }
]
getUserList().then(r => {
  data.value = r.result.list
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
