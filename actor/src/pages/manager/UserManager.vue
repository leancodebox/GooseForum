<script setup lang="ts">
import {
  NButton,
  NDataTable,
  NForm,
  NFormItem,
  NInput,
  NModal,
  NSelect,
  NSpace,
  NSwitch,
  NTag,
  useMessage
} from 'naive-ui'
import {h, onMounted, ref} from 'vue'
import {editUser, getAllRoleItem, getUserList} from '@/service/request'
import {Ref, UnwrapRef} from "@vue/reactivity";

const message = useMessage()
type UserItem = {
  userId: number | null
  username: string | null
  email: string | null
  createTime: string | null
  roleList: any
  status: number | null
  validate: number | null
  prestige: number | null
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
      if (!row.roleList) {
        return []
      }
      let res = []
      for (const rowKey in row.roleList) {
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
    key: 'status',
    render(row: UserItem) {
      if (row.status === 0) {
        return h(NTag, {type: 'success'}, () => "正常")
      } else {
        return h(NTag, {type: 'error'}, () => "冻结")
      }
    }
  },
  {
    title: '验证',
    key: 'validate',
    render(row: UserItem) {
      if (row.validate === 1) {
        return h(NTag, {type: 'success'}, () => "验证")
      } else {
        return h(NTag, {type: 'warning'}, () => "未验证")
      }
    }
  },
  {
    title: '声望',
    key: 'prestige'
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
              showModal.value = true
              userEntity.value = {
                userId: row.userId,
                username: row.username,
                roleId: !!row.roleList ? row.roleList.map(item => {
                  return item.value
                }) : [],
                status: row.status,
                validate: row.validate
              }
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
      )
      ]
    }
  }
]
let roleOption = ref([])
onMounted(() => {
  getAllRoleItem().then(r => {
    console.log(r.result)
    roleOption.value = r.result
  })
  showUserList()
})

function showUserList() {
  getUserList().then(r => {
    data.value = r.result.list
  })
}

let pagination = true
let showModal = ref(false)
let userEntity = ref({
  userId: 0,
  username: "",
  roleId: [],
  status: 0,
  validate: 0
})

function onPositiveClick() {
  userEdit4Role()
}

function onNegativeClick() {

}

function userEdit4Role() {
  let req = userEntity.value
  editUser(req.userId, req.status, req.validate, req.roleId)
      .then(() => {
        showUserList()
      })
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
        :model="userEntity"
        :rules="{}"
        label-placement="left"
        label-width="auto"
        require-mark-placement="right-hanging"
        :style="{
          padding: '30px 0 0',
      maxWidth: '640px'
    }"
    >
      <n-form-item label="角色名" path="username">
        <n-input v-model:value="userEntity.username" disabled></n-input>
      </n-form-item>
      <n-form-item label="是否冻结">
        <n-switch v-model:value="userEntity.status" :checked-value="1" :unchecked-value="0"/>
      </n-form-item>
      <n-form-item label="验证通过">
        <n-switch v-model:value="userEntity.validate" :checked-value="1" :unchecked-value="0"/>
      </n-form-item>
      <n-form-item label="角色" path="roleId">
        <n-select v-model:value="userEntity.roleId" multiple :options="roleOption"></n-select>
      </n-form-item>
    </n-form>
  </n-modal>
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
