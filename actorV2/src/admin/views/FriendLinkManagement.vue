<script setup lang="ts">
import { ref, onMounted,h } from 'vue'
import { NButton, NDataTable, NSpace, useMessage } from 'naive-ui'
import { AddOutline, TrashOutline, CreateOutline } from '@vicons/ionicons5'

const message = useMessage()
const friendLinks = ref([])
const loading = ref(false)

// 获取友情链接列表
const fetchFriendLinks = async () => {
  loading.value = true
  try {
    // 这里调用API获取数据
    // const res = await getFriendLinks()
    // friendLinks.value = res.result
    message.success('获取友情链接成功')
  } catch (error) {
    message.error('获取友情链接失败')
    console.error(error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchFriendLinks()
})

const columns = [
  {
    title: '名称',
    key: 'name'
  },
  {
    title: 'URL',
    key: 'url'
  },
  {
    title: '排序',
    key: 'sort'
  },
  {
    title: '状态',
    key: 'status'
  },
  {
    title: '操作',
    key: 'actions',
    render() {
      return h(NSpace, {}, [
        h(NButton, { size: 'small', icon: CreateOutline }, '编辑'),
        h(NButton, { size: 'small', icon: TrashOutline, type: 'error' }, '删除')
      ])
    }
  }
]
</script>

<template>
  <div>
    <n-space justify="end" style="margin-bottom: 16px">
      <n-button type="primary">
        <template #icon>
          <n-icon><AddOutline /></n-icon>
        </template>
        添加友情链接
      </n-button>
    </n-space>
    
    <n-data-table
      :columns="columns"
      :data="friendLinks"
      :loading="loading"
      striped
    />
  </div>
</template>