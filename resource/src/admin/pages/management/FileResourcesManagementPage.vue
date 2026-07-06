<script setup lang="ts">
import { adminText } from '@/admin/runtime/i18n-text'

import { computed, onMounted, ref } from 'vue'
import { ChevronLeft, ChevronRight, Copy, File, RefreshCw } from '@lucide/vue'
import { BasicPage } from '@/admin/components/global-layout'
import AdminSection from '@/admin/components/AdminSection.vue'
import { Badge } from '@/admin/components/ui/badge'
import { Button } from '@/admin/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/admin/components/ui/dialog'
import { adminToast } from '@/admin/runtime/toast'
import { getFileResourceList } from '@/admin/runtime/api'
import type { AdminFileResource, AdminPayload, ManageHomeProps } from '@/admin/types'

defineProps<{
  payload: AdminPayload<ManageHomeProps>
}>()

const rows = ref<AdminFileResource[]>([])
const loading = ref(false)
const error = ref('')
const page = ref(1)
const pageSize = ref(20)
const maxId = ref(0)
const previewOpen = ref(false)
const previewItem = ref<AdminFileResource | null>(null)
const totalPages = computed(() => Math.max(1, Math.ceil(maxId.value / pageSize.value)))

function pageResultSize(result: { pageSize?: number, size?: number }) {
  return result.pageSize || result.size || pageSize.value
}

async function loadResources() {
  loading.value = true
  error.value = ''
  try {
    const result = await getFileResourceList({ page: page.value, pageSize: pageSize.value })
    rows.value = result.list || []
    maxId.value = result.total || 0
    page.value = result.page || page.value
    pageSize.value = pageResultSize(result)
  } catch (err) {
    error.value = err instanceof Error ? err.message : adminText('k00fb')
  } finally {
    loading.value = false
  }
}

function updatePage(value: number) {
  page.value = value
  void loadResources()
}

function updatePageSize(value: number) {
  pageSize.value = value
  page.value = 1
  void loadResources()
}

function absoluteUrl(url: string) {
  if (!url) return ''
  return new URL(url, window.location.origin).toString()
}

async function copyUrl(row: AdminFileResource) {
  try {
    await navigator.clipboard.writeText(absoluteUrl(row.url))
    adminToast.success(adminText('k00fc'))
  } catch (err) {
    adminToast.error(err, adminText('k00fd'))
  }
}

function uploaderName(row: AdminFileResource) {
  return row.uploaderUsername || (row.userId ? `#${row.userId}` : '-')
}

function isImage(row: AdminFileResource) {
  return row.type.startsWith('image/')
}

function openPreview(item: AdminFileResource) {
  previewItem.value = item
  previewOpen.value = true
}

function formatBytes(value: number) {
  if (!Number.isFinite(value) || value <= 0) return '0 B'
  if (value < 1024) return `${value} B`
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KB`
  return `${(value / 1024 / 1024).toFixed(2)} MB`
}

function formatTime(value: string) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  const pad = (part: number) => String(part).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`
}

onMounted(loadResources)
</script>

<template>
  <BasicPage :title="adminText('k00f6')" :description="adminText('k00fe')" sticky>
    <template #actions>
      <Button variant="outline" size="sm" type="button" :disabled="loading" @click="loadResources">
        <RefreshCw class="size-4" :class="loading ? 'animate-spin' : ''" />
        {{ adminText('k004q') }}
      </Button>
    </template>

    <AdminSection>
      <div v-if="loading && rows.length === 0" class="flex h-48 items-center justify-center text-sm text-muted-foreground">
        {{ adminText('k0046') }}
      </div>
      <div v-else-if="error" class="flex h-48 items-center justify-center px-4 text-center">
        <div class="inline-flex items-center gap-3 rounded-md border border-destructive/30 bg-destructive/5 px-4 py-2 text-sm text-destructive">
          <span>{{ error }}</span>
          <Button variant="link" size="sm" class="h-auto px-0 text-destructive" type="button" @click="loadResources">{{ adminText('k002w') }}</Button>
        </div>
      </div>
      <div v-else-if="rows.length === 0" class="flex h-48 items-center justify-center text-sm text-muted-foreground">
        {{ adminText('k00ff') }}
      </div>
      <div v-else class="grid gap-3 p-3 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5">
        <article
          v-for="item in rows"
          :key="item.id"
          class="group overflow-hidden rounded-lg border bg-background transition hover:border-primary/35 hover:shadow-sm"
        >
          <button
            type="button"
            class="block aspect-[4/3] w-full overflow-hidden bg-muted text-left"
            :title="item.name"
            @click="openPreview(item)"
          >
            <img v-if="isImage(item)" :src="item.url" :alt="item.name" class="h-full w-full object-cover transition duration-200 group-hover:scale-[1.02]" loading="lazy">
            <div v-else class="flex h-full w-full flex-col items-center justify-center gap-2 text-muted-foreground">
              <File class="size-8" />
              <span class="max-w-full truncate px-4 text-xs font-medium">{{ item.type || 'file' }}</span>
            </div>
          </button>
          <div class="space-y-2 p-3">
            <div class="min-w-0">
              <div class="truncate text-sm font-medium text-foreground" :title="item.name">{{ item.name }}</div>
              <div class="mt-1 truncate font-mono text-xs text-muted-foreground" :title="item.url">{{ item.url }}</div>
            </div>
            <div class="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
              <Badge variant="secondary" class="font-mono">{{ item.type }}</Badge>
              <span>{{ formatBytes(item.size) }}</span>
            </div>
            <div class="flex items-center justify-between gap-2 text-xs text-muted-foreground">
              <a
                v-if="item.userId"
                :href="`/u/${item.userId}`"
                target="_blank"
                rel="noreferrer"
                class="min-w-0 truncate hover:text-primary hover:underline"
                :title="uploaderName(item)"
              >
                {{ adminText('k00f8') }} {{ uploaderName(item) }}
              </a>
              <span v-else class="truncate">{{ adminText('k00f8') }} -</span>
              <span class="whitespace-nowrap">{{ formatTime(item.createdAt) }}</span>
            </div>
          </div>
        </article>
      </div>

      <div class="flex flex-wrap items-center justify-between gap-3 border-t bg-muted/10 px-4 py-3 text-sm text-muted-foreground">
        <div>{{ adminText('k00fj') }} {{ maxId }}</div>
        <div class="flex items-center gap-2">
          <select
            class="h-9 rounded-md border bg-background px-2 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
            :value="pageSize"
            @change="updatePageSize(Number(($event.target as HTMLSelectElement).value))"
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
            @click="updatePage(page - 1)"
          >
            <ChevronLeft class="size-4" />
          </Button>
          <span class="min-w-16 text-center">{{ adminText('k0056') }} {{ page }} / {{ totalPages }} {{ adminText('k0057') }}</span>
          <Button
            variant="outline"
            size="icon"
            type="button"
            :disabled="page >= totalPages"
            @click="updatePage(page + 1)"
          >
            <ChevronRight class="size-4" />
          </Button>
        </div>
      </div>
    </AdminSection>

    <Dialog :open="previewOpen" @update:open="previewOpen = $event">
      <DialogContent class="max-h-[calc(100vh-2rem)] gap-3 overflow-hidden p-4 sm:max-w-5xl">
        <DialogHeader v-if="previewItem" class="pr-8">
          <DialogTitle class="truncate">{{ previewItem.name }}</DialogTitle>
          <DialogDescription class="truncate font-mono text-xs">
            {{ previewItem.url }}
          </DialogDescription>
        </DialogHeader>
        <div v-if="previewItem" class="grid min-h-0 gap-3 lg:grid-cols-[minmax(0,1fr)_18rem]">
          <div class="flex max-h-[70vh] min-h-64 items-center justify-center overflow-hidden rounded-md border bg-muted/50">
            <img v-if="isImage(previewItem)" :src="previewItem.url" :alt="previewItem.name" class="max-h-full max-w-full object-contain">
            <div v-else class="flex flex-col items-center gap-3 text-muted-foreground">
              <File class="size-12" />
              <div class="font-mono text-sm">{{ previewItem.type || 'file' }}</div>
            </div>
          </div>
          <aside class="space-y-3 rounded-md border bg-muted/20 p-3 text-sm">
            <div>
              <div class="text-xs text-muted-foreground">{{ adminText('k00f8') }}</div>
              <a
                v-if="previewItem.userId"
                :href="`/u/${previewItem.userId}`"
                target="_blank"
                rel="noreferrer"
                class="font-medium hover:text-primary hover:underline"
              >
                {{ uploaderName(previewItem) }}
              </a>
              <div v-else class="font-medium">-</div>
            </div>
            <div class="grid grid-cols-2 gap-3">
              <div>
                <div class="text-xs text-muted-foreground">{{ adminText('k00f7') }}</div>
                <div class="mt-1"><Badge variant="secondary" class="font-mono">{{ previewItem.type }}</Badge></div>
              </div>
              <div>
                <div class="text-xs text-muted-foreground">{{ adminText('k00f9') }}</div>
                <div class="mt-1 text-xs">{{ formatTime(previewItem.createdAt) }}</div>
              </div>
            </div>
            <div>
              <div class="text-xs text-muted-foreground">{{ adminText('k00fi') }}</div>
              <div class="font-mono text-xs">{{ formatBytes(previewItem.size) }}</div>
            </div>
            <div>
              <div class="text-xs text-muted-foreground">{{ adminText('k00fa') }}</div>
              <div class="mt-1 break-all font-mono text-xs text-muted-foreground">{{ absoluteUrl(previewItem.url) }}</div>
            </div>
            <Button type="button" class="w-full" variant="outline" size="sm" @click="copyUrl(previewItem)">
              <Copy class="size-4" />
              {{ adminText('k00fg') }}
            </Button>
          </aside>
        </div>
      </DialogContent>
    </Dialog>
  </BasicPage>
</template>
