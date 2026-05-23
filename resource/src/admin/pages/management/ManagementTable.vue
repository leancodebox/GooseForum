<script setup lang="ts">
import { ChevronLeft, ChevronRight } from '@lucide/vue'

withDefaults(defineProps<{
  loading?: boolean
  error?: string
  emptyText?: string
  columns: string[]
  total?: number
  page?: number
  pageSize?: number
  showPagination?: boolean
}>(), {
  emptyText: '暂无数据',
  total: 0,
  page: 1,
  pageSize: 10,
  showPagination: true,
})

const emit = defineEmits<{
  'update:page': [value: number]
  'update:pageSize': [value: number]
  retry: []
}>()
</script>

<template>
  <div class="overflow-hidden rounded-lg border bg-card shadow-sm">
    <div class="overflow-x-auto">
      <table class="w-full min-w-[760px] text-sm">
        <thead class="border-b bg-muted/45 text-xs font-medium text-muted-foreground">
          <tr>
            <th
              v-for="column in columns"
              :key="column"
              class="h-11 px-4 text-left align-middle"
            >
              {{ column }}
            </th>
          </tr>
        </thead>
        <tbody class="divide-y">
          <tr v-if="loading">
            <td :colspan="columns.length" class="h-28 px-4 text-center text-muted-foreground">
              加载中...
            </td>
          </tr>
          <tr v-else-if="error">
            <td :colspan="columns.length" class="h-28 px-4 text-center">
              <div class="inline-flex items-center gap-3 rounded-md border border-destructive/30 bg-destructive/5 px-4 py-2 text-destructive">
                <span>{{ error }}</span>
                <button class="text-sm font-medium underline underline-offset-4" type="button" @click="emit('retry')">重试</button>
              </div>
            </td>
          </tr>
          <slot v-else />
        </tbody>
      </table>
    </div>

    <div
      v-if="showPagination"
      class="flex flex-wrap items-center justify-between gap-3 border-t px-4 py-3 text-sm text-muted-foreground"
    >
      <div>共 {{ total }} 条</div>
      <div class="flex items-center gap-2">
        <select
          class="h-9 rounded-md border bg-background px-2 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
          :value="pageSize"
          @change="emit('update:pageSize', Number(($event.target as HTMLSelectElement).value))"
        >
          <option :value="10">10 / 页</option>
          <option :value="20">20 / 页</option>
          <option :value="30">30 / 页</option>
          <option :value="50">50 / 页</option>
        </select>
        <button
          class="inline-flex size-9 items-center justify-center rounded-md border bg-background text-foreground disabled:cursor-not-allowed disabled:opacity-50"
          type="button"
          :disabled="page <= 1"
          @click="emit('update:page', page - 1)"
        >
          <ChevronLeft class="size-4" />
        </button>
        <span class="min-w-16 text-center">第 {{ page }} / {{ Math.max(1, Math.ceil(total / pageSize)) }} 页</span>
        <button
          class="inline-flex size-9 items-center justify-center rounded-md border bg-background text-foreground disabled:cursor-not-allowed disabled:opacity-50"
          type="button"
          :disabled="page >= Math.max(1, Math.ceil(total / pageSize))"
          @click="emit('update:page', page + 1)"
        >
          <ChevronRight class="size-4" />
        </button>
      </div>
    </div>
  </div>
</template>
