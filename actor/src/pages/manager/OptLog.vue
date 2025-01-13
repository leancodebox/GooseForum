<script setup lang="ts">
import {NButton, NCard, NDataTable, NForm, NFormItem, NInput, NSelect, NSpace, useMessage} from 'naive-ui'
import {ref} from 'vue'
import {getUserList} from '@/service/request'
import {Ref, UnwrapRef} from "@vue/reactivity";

let searchInfo = ref({
  optUserId: "",
  optType: "0",
  targetType: "0",
  targetId: "",
})

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
    title: '操作人',
    key: 'username'
  },
  {
    title: '操作类型',
    key: 'email'
  },
  {
    title: '目标类型',
    key: 'targetType'
  },
  {
    title: '目标id',
    key: 'targetId'
  },
  {
    title: '操作信息',
    key: 'optInfo'
  },
  {
    title: '操作时间',
    key: 'createTime'
  },
]
getUserList().then(r => {
  data.value = r.result.list.map(item => {
    return {userId: item.userId, username: item.username, email: item.email, createTime: item.createTime, status: ""}
  })
  // console.log(r)
})

let options = [
  {
    label: '无',
    value: '0'
  },
  {
    label: '用户',
    value: '1'
  },
];
let pagination = true
</script>
<template>
  <n-space vertical>
    <n-card :bordered="false">
      <n-form
          ref="formRef"
          inline
          :label-width="80"
          :size="'small'"
      >
        <n-form-item label="操作人id" path="user.age" style="min-width:160px">
          <n-input :value="searchInfo.optUserId" placeholder="操作人id"/>
        </n-form-item>

        <n-form-item label="操作类型" path="user.age" style="min-width:160px">
          <n-select v-model:value="searchInfo.optType" :options="options"/>
        </n-form-item>

        <n-form-item label="目标类型" path="user.age" style="min-width: 120px">
          <n-select  v-model:value="searchInfo.targetType" :options="options"/>
        </n-form-item>

        <n-form-item label="目标id" path="user.age">
          <n-input v-model:value="searchInfo.targetId"  placeholder="输入编号"/>
        </n-form-item>

        <n-form-item>
          <n-button attr-type="button" type="primary">
            搜索
          </n-button>
        </n-form-item>
      </n-form>
    </n-card>
    <n-data-table
        :columns="columns"
        :data="data"
        :pagination="pagination"
        :bordered="false"
    />
  </n-space>
</template>
<style scoped>
.carousel-img {
  width: 100%;
  height: 240px;
  object-fit: cover;
}
</style>
