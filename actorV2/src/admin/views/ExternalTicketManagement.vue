<script setup lang="ts">
import {h, onMounted, reactive, ref} from 'vue'
import {NButton, NDataTable, NSpace, NTag, useMessage} from 'naive-ui'
import {CheckmarkOutline, CloseOutline, EyeOutline} from '@vicons/ionicons5'
import {applySheetList} from "@/admin/utils/authService.ts"
import type {ApplySheet} from "@/admin/types/adminInterfaces.ts"

const message = useMessage()
const tickets = ref<ApplySheet[]>([])
const loading = ref(false)

// 分页设置
const pagination = reactive({
  page: 1,
  pageSize: 10,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 30, 50],
  onChange: (page: number) => {
    pagination.page = page
    fetchTickets()
  },
  onUpdatePageSize: (pageSize: number) => {
    pagination.pageSize = pageSize
    pagination.page = 1
    fetchTickets()
  }
})

// 获取工单列表
const fetchTickets = async () => {
  loading.value = true
  try {
    const res = await applySheetList(pagination.page, pagination.pageSize)
    if (res.code === 0) {
      tickets.value = res.result.list
      pagination.itemCount = res.result.total
    } else {
      message.error(res.message || '获取工单列表失败')
    }
  } catch (error) {
    message.error('获取工单列表失败')
    console.error(error)
  } finally {
    loading.value = false
  }
}

// 状态映射
const statusMap = {
  0: {text: '未知工单', type: 'warning'},
  1: {text: '待处理', type: 'warning'},
  2: {text: '已通过', type: 'success'},
  3: {text: '已拒绝', type: 'error'}
}

// 类型映射
const typeMap = {
  1: '友情链接申请',
  // 可以扩展其他类型
}

const columns = [
  {
    title: 'ID',
    key: 'id'
  },
  {
    title: '申请人',
    key: 'applyUserInfo'
  },
  {
    title: '类型',
    key: 'type',
    render(row: ApplySheet) {
      return typeMap[row.type] || row.type
    }
  },
  {
    title: '标题',
    key: 'title'
  },
  {
    title: '状态',
    key: 'status',
    render(row: ApplySheet) {
      const status = statusMap[row.status] || {text: row.status, type: 'default'}
      return h(NTag, {type: status.type}, {default: () => status.text})
    }
  },
  {
    title: '申请时间',
    key: 'createTime'
  },
  {
    title: '操作',
    key: 'actions',
    render(row: ApplySheet) {
      return h(NSpace, {}, [
        h(NButton, {
          size: 'small',
          icon: EyeOutline,
          onClick: () => handleView(row)
        }, '查看'),
        h(NButton, {
          size: 'small',
          icon: CheckmarkOutline,
          type: 'success',
          disabled: row.status !== 0,
          onClick: () => handleApprove(row.id)
        }, '通过'),
        h(NButton, {
          size: 'small',
          icon: CloseOutline,
          type: 'error',
          disabled: row.status !== 0,
          onClick: () => handleReject(row.id)
        }, '拒绝')
      ])
    }
  }
]

const showDetailModal = ref(false)
const currentTicket = ref<ApplySheet | null>(null)

// 查看详情
const handleView = (row: ApplySheet) => {
  currentTicket.value = row
  showDetailModal.value = true
}

// 通过工单
const handleApprove = async (id: number) => {
  // 这里调用通过API
  message.success(`已通过工单 ${id}`)
  await fetchTickets()
}

// 拒绝工单
const handleReject = async (id: number) => {
  // 这里调用拒绝API
  message.success(`已拒绝工单 ${id}`)
  await fetchTickets()
}

onMounted(() => {
  fetchTickets()
})
</script>

<template>
  <div>
    <n-data-table
        :columns="columns"
        :data="tickets"
        :loading="loading"
        :pagination="pagination"
        striped
        bordered
        remote
    />

    <!-- 详情对话框 -->
    <n-modal v-model:show="showDetailModal" preset="card" title="工单详情" style="width: 600px">
      <n-card v-if="currentTicket" title="详细信息">
        <n-space vertical>
          <n-text strong>ID:</n-text>
          <n-text>{{ currentTicket.id }}</n-text>

          <n-text strong>标题:</n-text>
          <n-text>{{ currentTicket.title }}</n-text>

          <n-text strong>申请人:</n-text>
          <n-text>{{ currentTicket.applyUserInfo }}</n-text>

          <n-text strong>类型:</n-text>
          <n-text>{{ typeMap[currentTicket.type] || currentTicket.type }}</n-text>

          <n-text strong>状态:</n-text>
          <n-tag :type="statusMap[currentTicket.status]?.type || 'default'">
            {{ statusMap[currentTicket.status]?.text || currentTicket.status }}
          </n-tag>

          <n-text strong>申请时间:</n-text>
          <n-text>{{ currentTicket.createTime }}</n-text>

          <n-text strong>内容详情:</n-text>
          <n-card content-style="white-space: pre-wrap;">
            {{ currentTicket.content }}
          </n-card>
        </n-space>
      </n-card>
    </n-modal>
  </div>
</template>
