<script setup lang="ts">
import { ref, onMounted, h } from 'vue'
import { NButton, NDataTable, NSpace, NTag, useMessage } from 'naive-ui'
import { EyeOutline, CheckmarkOutline, CloseOutline } from '@vicons/ionicons5'

const message = useMessage()
const tickets = ref([])
const loading = ref(false)

// 获取工单列表
const fetchTickets = async () => {
  loading.value = true
  try {
    // 这里调用API获取数据
    // const res = await getExternalTickets()
    // tickets.value = res.result
    message.success('获取工单列表成功')
  } catch (error) {
    message.error('获取工单列表失败')
    console.error(error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchTickets()
})

const columns = [
  {
    title: 'ID',
    key: 'id'
  },
  {
    title: '申请人',
    key: 'applicant'
  },
  {
    title: '类型',
    key: 'type'
  },
  {
    title: '状态',
    key: 'status',
    render(row: { status: string }) {
      return h(NTag, { type: row.status === 'pending' ? 'warning' : 'success' }, 
        { default: () => row.status === 'pending' ? '待处理' : '已处理' })
    }
  },
  {
    title: '申请时间',
    key: 'createdAt'
  },
  {
    title: '操作',
    key: 'actions',
    render() {
      return h(NSpace, {}, [
        h(NButton, { size: 'small', icon: EyeOutline }, '查看'),
        h(NButton, { size: 'small', icon: CheckmarkOutline, type: 'success' }, '通过'),
        h(NButton, { size: 'small', icon: CloseOutline, type: 'error' }, '拒绝')
      ])
    }
  }
]
</script>

<template>
  <div>
    <n-data-table
      :columns="columns"
      :data="tickets"
      :loading="loading"
      striped
      bordered
    />
  </div>
</template>