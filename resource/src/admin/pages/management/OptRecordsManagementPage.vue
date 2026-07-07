<script setup lang="ts">import { adminText } from '@/admin/runtime/i18n-text'

import { onMounted, ref } from 'vue'
import { RefreshCw } from '@lucide/vue'
import { i18n } from '@/runtime/i18n'
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

const columns = ['ID', adminText('k0036'), adminText('k0037'), adminText('k0038'), adminText('k0039'), adminText('k003a'), adminText('k003b')]

const optTypeCodeMap: Record<number, string> = {
  0: 'editUser',
  1: 'editTopic',
  2: 'editCategory',
}

const targetTypeCodeMap: Record<number, string> = {
  0: 'system',
  1: 'user',
  2: 'topic',
  3: 'docProject',
  4: 'docVersion',
  5: 'docContent',
  6: 'category',
}

const optInfoMessageKeyMap: Record<string, string> = {
  'admin.opt.user.updated': 'adminOptLog.messages.userUpdated',
  'admin.opt.topic.statusChanged': 'adminOptLog.messages.topicStatusChanged',
  'admin.opt.topic.pinWeightChanged': 'adminOptLog.messages.topicPinWeightChanged',
  'admin.opt.topic.categoriesChanged': 'adminOptLog.messages.topicCategoriesChanged',
  'admin.opt.topic.deleted': 'adminOptLog.messages.topicDeleted',
  'moderator.opt.topic.statusChanged': 'adminOptLog.messages.moderatorTopicStatusChanged',
  'admin.opt.category.moderatorAdded': 'adminOptLog.messages.categoryModeratorAdded',
  'admin.opt.category.moderatorRemoved': 'adminOptLog.messages.categoryModeratorRemoved',
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
    error.value = err instanceof Error ? err.message : adminText('k0013')
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
  const code = optTypeCodeMap[value]
  const key = code ? `adminOptLog.optType.${code}` : ''
  return key && i18n.global.te(key) ? i18n.global.t(key) : adminText('k00aj', { value })
}

function targetTypeName(value: number) {
  const code = targetTypeCodeMap[value]
  const key = code ? `adminOptLog.targetType.${code}` : ''
  return key && i18n.global.te(key) ? i18n.global.t(key) : adminText('k00ak', { value })
}

function normalizeParam(value: unknown) {
  if (Array.isArray(value)) return value.join(', ')
  if (value === undefined || value === null) return ''
  return String(value)
}

function localizeOptParams(params: Record<string, unknown>) {
  const next: Record<string, unknown> = {}
  for (const [key, value] of Object.entries(params)) {
    next[key] = normalizeParam(value)
  }

  const status = typeof params.status === 'string' ? params.status : ''
  if (status) {
    const statusKey = `adminOptLog.status.${status}`
    next.status = i18n.global.te(statusKey) ? i18n.global.t(statusKey) : status
  }

  const changedFields = Array.isArray(params.changes)
    ? params.changes
      .map((item) => {
        const field = String(item)
        const fieldKey = `adminOptLog.userField.${field}`
        return i18n.global.te(fieldKey) ? i18n.global.t(fieldKey) : field
      })
      .join(', ')
    : ''
  if (changedFields) next.changedFields = changedFields

  return next
}

function tryParseOptInfo(value: string) {
  if (!value || value[0] !== '{') return undefined
  try {
    const parsed = JSON.parse(value) as { messageCode?: unknown, params?: unknown }
    if (typeof parsed.messageCode !== 'string') return undefined
    return {
      messageCode: parsed.messageCode,
      params: parsed.params && typeof parsed.params === 'object' && !Array.isArray(parsed.params)
        ? parsed.params as Record<string, unknown>
        : {},
    }
  } catch {
    return undefined
  }
}

function optInfoText(item: AdminOptRecord) {
  const payload = item.optInfoPayload || tryParseOptInfo(item.optInfo)
  if (!payload?.messageCode) return item.optInfo || '-'

  const key = optInfoMessageKeyMap[payload.messageCode]
  if (!i18n.global.te(key)) return item.optInfo || payload.messageCode
  return i18n.global.t(key, localizeOptParams(payload.params || {}))
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
  <BasicPage :title="adminText('k007c')" :description="adminText('k007d')" sticky>
    <template #actions>
      <Button variant="outline" size="sm" type="button" :disabled="loading" @click="loadRecords">
        <RefreshCw class="size-4" :class="loading ? 'animate-spin' : ''" />
        {{ adminText('k004q') }}
      </Button>
    </template>

      <ManagementTable
        :columns="columns"
        :loading="loading && rows.length === 0"
        :error="error"
        :empty-text="adminText('k007e')"
        :total="total"
        :page="page"
        :page-size="pageSize"
        @retry="loadRecords"
        @update:page="updatePage"
        @update:page-size="updatePageSize"
      >
        <tr v-if="rows.length === 0">
          <td :colspan="columns.length" class="h-28 px-4 text-center text-muted-foreground">{{ adminText('k007e') }}</td>
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
            <div class="line-clamp-2 text-foreground">{{ optInfoText(item) }}</div>
          </td>
          <td class="whitespace-nowrap px-4 py-3 text-muted-foreground">{{ formatTime(item.createdAt) }}</td>
        </tr>
      </ManagementTable>
    </BasicPage>
</template>
