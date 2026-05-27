<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RefreshCw } from '@lucide/vue'
import { BasicPage } from '@/admin/components/global-layout'
import { Badge } from '@/admin/components/ui/badge'
import { Button } from '@/admin/components/ui/button'
import { getOptRecordList } from '@/admin/runtime/api'
import type { AdminOptRecord, AdminPayload, ManageHomeProps } from '@/admin/types'
import ManagementTable from './ManagementTable.vue'

const props = defineProps<{
  payload: AdminPayload<ManageHomeProps>
}>()

const rows = ref<AdminOptRecord[]>([])
const loading = ref(false)
const error = ref('')
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

const columns = ['ID', '操作人', '操作类型', '目标类型', '目标 ID', '操作内容', '时间']

const optTypeMap: Record<number, string> = {
  0: '操作用户',
  1: '编辑文章',
}

const targetTypeMap: Record<number, string> = {
  0: '系统',
  1: '用户',
  2: '文章',
  3: '文档项目',
  4: '文档版本',
  5: '文档内容',
}

function pageResultSize(result: { pageSize?: number, size?: number }) {
  return result.pageSize || result.size || pageSize.value
}

async function loadRecords() {
  loading.value = true
  error.value = ''
  try {
    const result = await getOptRecordList({ page: page.value, pageSize: pageSize.value })
    rows.value = result.list || []
    total.value = result.total || 0
    page.value = result.page + 1
    pageSize.value = pageResultSize(result)
  } catch (err) {
    error.value = err instanceof Error ? err.message : '获取操作记录失败'
  } finally {
    loading.value = false
  }
}

function updatePage(value: number) {
  page.value = value
  void loadRecords()
}

function updatePageSize(value: number) {
  pageSize.value = value
  page.value = 1
  void loadRecords()
}

function optTypeName(value: number) {
  return optTypeMap[value] || `类型 ${value}`
}

function targetTypeName(value: number) {
  return targetTypeMap[value] || `类型 ${value}`
}

function formatTime(value: string) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString('zh-CN', { hour12: false })
}

onMounted(loadRecords)
</script>

<template>
  <BasicPage title="操作记录" description="查看后台关键管理操作的审计记录。" sticky>
    <template #actions>
      <Button variant="outline" size="sm" type="button" :disabled="loading" @click="loadRecords">
        <RefreshCw class="size-4" :class="loading ? 'animate-spin' : ''" />
        刷新
      </Button>
    </template>

      <ManagementTable
        :columns="columns"
        :loading="loading"
        :error="error"
        empty-text="暂无操作记录"
        :total="total"
        :page="page"
        :page-size="pageSize"
        @retry="loadRecords"
        @update:page="updatePage"
        @update:page-size="updatePageSize"
      >
        <tr v-if="rows.length === 0">
          <td :colspan="columns.length" class="h-28 px-4 text-center text-muted-foreground">暂无操作记录</td>
        </tr>
        <tr v-for="item in rows" v-else :key="item.id" class="transition-colors hover:bg-muted/35">
          <td class="px-4 py-3 font-mono text-xs text-muted-foreground">{{ item.id }}</td>
          <td class="px-4 py-3 font-mono text-xs">{{ item.optUserId || '-' }}</td>
          <td class="px-4 py-3">
            <Badge variant="secondary">{{ optTypeName(item.optType) }}</Badge>
          </td>
          <td class="px-4 py-3 text-muted-foreground">{{ targetTypeName(item.targetType) }}</td>
          <td class="px-4 py-3 font-mono text-xs text-muted-foreground">{{ item.targetId || '-' }}</td>
          <td class="max-w-xl px-4 py-3">
            <div class="line-clamp-2 text-foreground">{{ item.optInfo || '-' }}</div>
          </td>
          <td class="whitespace-nowrap px-4 py-3 text-muted-foreground">{{ formatTime(item.createdAt) }}</td>
        </tr>
      </ManagementTable>
    </BasicPage>
</template>
