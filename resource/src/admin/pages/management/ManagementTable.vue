<script setup lang="ts">
import { adminText } from '@/admin/runtime/i18n-text'

import { ChevronLeft, ChevronRight } from '@lucide/vue'
import AdminSection from '@/admin/components/AdminSection.vue'
import { Button } from '@/admin/components/ui/button'

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
  emptyText: adminText('k002v'),
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
  <AdminSection>
    <template v-if="$slots.header" #header>
      <slot name="header" />
    </template>
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
              {{ adminText('k0046') }}
            </td>
          </tr>
          <tr v-else-if="error">
            <td :colspan="columns.length" class="h-28 px-4 text-center">
              <div class="inline-flex items-center gap-3 rounded-md border border-destructive/30 bg-destructive/5 px-4 py-2 text-destructive">
                <span>{{ error }}</span>
                <Button variant="link" size="sm" class="h-auto px-0 text-destructive" type="button" @click="emit('retry')">{{ adminText('k002w') }}</Button>
              </div>
            </td>
          </tr>
          <slot v-else />
        </tbody>
      </table>
    </div>

    <div
      v-if="showPagination"
      class="flex flex-wrap items-center justify-between gap-3 border-t bg-muted/10 px-4 py-3 text-sm text-muted-foreground"
    >
      <div>{{ adminText('k0054') }} {{ total }} {{ adminText('k0055') }}</div>
      <div class="flex items-center gap-2">
        <select
          class="h-9 rounded-md border bg-background px-2 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
          :value="pageSize"
          @change="emit('update:pageSize', Number(($event.target as HTMLSelectElement).value))"
        >
          <option :value="10">{{ adminText('k002x') }}</option>
          <option :value="20">{{ adminText('k002y') }}</option>
          <option :value="30">{{ adminText('k002z') }}</option>
          <option :value="50">{{ adminText('k0030') }}</option>
        </select>
        <Button
          variant="outline"
          size="icon"
          type="button"
          :disabled="page <= 1"
          @click="emit('update:page', page - 1)"
        >
          <ChevronLeft class="size-4" />
        </Button>
        <span class="min-w-16 text-center">{{ adminText('k0056') }} {{ page }} / {{ Math.max(1, Math.ceil(total / pageSize)) }} {{ adminText('k0057') }}</span>
        <Button
          variant="outline"
          size="icon"
          type="button"
          :disabled="page >= Math.max(1, Math.ceil(total / pageSize))"
          @click="emit('update:page', page + 1)"
        >
          <ChevronRight class="size-4" />
        </Button>
      </div>
    </div>
  </AdminSection>
</template>
