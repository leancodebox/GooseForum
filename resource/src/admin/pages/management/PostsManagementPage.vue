<script setup lang="ts">
import { adminText } from '@/admin/runtime/i18n-text'

import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Ban, Eye, FileText, Heart, MessageSquare, Pin, RefreshCw, Search, Tags, Trash2, Undo2 } from '@lucide/vue'
import AdminActionButton from '@/admin/components/AdminActionButton.vue'
import AdminConfirmDialog from '@/admin/components/AdminConfirmDialog.vue'
import AdminSection from '@/admin/components/AdminSection.vue'
import AdminToolbar from '@/admin/components/AdminToolbar.vue'
import { BasicPage } from '@/admin/components/global-layout'
import { Button } from '@/admin/components/ui/button'
import { Badge } from '@/admin/components/ui/badge'
import { Input } from '@/admin/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/admin/components/ui/select'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/admin/components/ui/dialog'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/admin/components/ui/table'
import {
  reviewContent,
  getReviewPosts,
  editTopic,
  updateTopicCategories,
  updateTopicPin,
  deleteTopic,
  getTopicSource,
  getTopicsList,
  getCategoryList,
} from '@/admin/runtime/api'
import { adminToast } from '@/admin/runtime/toast'
import type { ReviewPost, AdminTopic, AdminCategory, AdminPayload, TopicSource, ManageHomeProps } from '@/admin/types'

defineProps<{
  payload: AdminPayload<ManageHomeProps>
}>()

const { t } = useI18n()
const kind = ref<'topic' | 'post'>('topic')
const moderationStatus = ref('all')
const categoryId = ref('all')
const replyRows = ref<ReviewPost[]>([])
const reviewReason = ref('')
const reviewReply = ref<ReviewPost | null>(null)
let requestVersion = 0
const loading = ref(false)
const saving = ref(false)
const sourceLoading = ref(false)
const error = ref('')
const rows = ref<AdminTopic[]>([])
const hasNext = ref(false)
const page = ref(1)
const pageSize = ref(10)
const search = ref('')
const appliedSearch = ref('')
const categories = ref<AdminCategory[]>([])
const categoryDialogRow = ref<AdminTopic | null>(null)
const selectedCategoryIds = ref<number[]>([])
const pendingCategoryIds = ref<number[] | null>(null)
const sourceDialogRow = ref<AdminTopic | null>(null)
const source = ref<TopicSource | null>(null)
const actionRow = ref<AdminTopic | null>(null)
const deleteRow = ref<AdminTopic | null>(null)
const pinDialogRow = ref<AdminTopic | null>(null)
const pinWeightInput = ref(0)

interface CategoryOption {
  id: number
  category: string
  color: string
  missing: boolean
}

const categoryMap = computed(() => new Map(categories.value.map(item => [item.id, item])))
const categoryDialogOptions = computed<CategoryOption[]>(() => {
  const options = categories.value.map(category => ({
    id: category.id,
    category: category.category,
    color: category.color || '#64748b',
    missing: false,
  }))
  const missingIds = selectedCategoryIds.value.filter(id => !categoryMap.value.has(id))
  return [
    ...options,
    ...missingIds.map(id => ({
      id,
      category: adminText('k00cc', { id }),
      color: '#94a3b8',
      missing: true,
    })),
  ]
})
const visibleRowCount = computed(() => kind.value === 'topic' ? rows.value.length : replyRows.value.length)
const rangeStart = computed(() => (visibleRowCount.value === 0 ? 0 : (page.value - 1) * pageSize.value + 1))
const rangeEnd = computed(() => visibleRowCount.value === 0 ? 0 : rangeStart.value + visibleRowCount.value - 1)
const currentMainCategoryName = computed(() => categoryName(categoryDialogRow.value?.categoryId?.[0]))
const nextMainCategoryName = computed(() => categoryName(pendingCategoryIds.value?.[0]))

function categoryName(id?: number) {
  if (!id) return ''
  return categoryMap.value.get(id)?.category || adminText('k00cc', { id })
}

function avatarText(post: AdminTopic) {
  return post.username.slice(0, 1).toUpperCase()
}

function postCategories(post: AdminTopic) {
  return (post.categoryId || []).map((id) => {
    const category = categoryMap.value.get(id)
    if (category) {
      return {
        id: category.id,
        category: category.category,
        color: category.color || '#64748b',
        missing: false,
      }
    }
    return {
      id,
      category: adminText('k00cc', { id }),
      color: '#94a3b8',
      missing: true,
    }
  })
}

function postDate(value?: string) {
  if (!value) return '-'
  return value.slice(0, 10)
}

function postTime(value?: string) {
  if (!value) return ''
  return value.slice(11, 16)
}

function reviewLabel(status: string) {
  return adminText(`reviewStatus_${status === 'denied' ? 'rejected' : status || 'none'}`)
}
function topicStatusInfo(post: AdminTopic) {
  if (post.deleted) return { label: adminText('topicDeleted'), className: 'bg-muted text-muted-foreground' }
  if (post.topicStatus !== 1 && ['rejected', 'pending', 'denied'].includes(post.moderationStatus)) return { label: reviewLabel(post.moderationStatus), className: 'bg-destructive/10 text-destructive' }
  return post.topicStatus === 1
    ? { label: adminText('k003q'), className: 'bg-primary text-primary-foreground' }
    : { label: adminText('k003r'), className: 'bg-muted text-muted-foreground' }
}

async function loadPosts() {
  const currentRequest = ++requestVersion
  loading.value = true
  error.value = ''
  try {
    if (kind.value === 'post') {
      const result = await getReviewPosts({ page: page.value, pageSize: pageSize.value, moderationStatus: moderationStatus.value === 'all' ? undefined : moderationStatus.value, search: appliedSearch.value || undefined })
      if (currentRequest !== requestVersion) return
      replyRows.value = result.list || []
      hasNext.value = Boolean(result.hasNext)
      return
    }
    const [postPage, categoryList] = await Promise.all([
      getTopicsList({ page: page.value, pageSize: pageSize.value, search: appliedSearch.value || undefined, moderationStatus: moderationStatus.value === 'all' ? undefined : moderationStatus.value, categoryId: Number(categoryId.value) || undefined }),
      categories.value.length ? Promise.resolve(categories.value) : getCategoryList(),
    ])
    if (currentRequest !== requestVersion) return
    rows.value = postPage.list || []
    hasNext.value = Boolean(postPage.hasNext)
    categories.value = categoryList || []
  } catch (err) {
    if (currentRequest === requestVersion) error.value = err instanceof Error ? err.message : adminText('k003s')
  } finally {
    if (currentRequest === requestVersion) loading.value = false
  }
}

function applySearch() {
  appliedSearch.value = search.value.trim()
  page.value = 1
  void loadPosts()
}

function changePage(nextPage: number) {
  page.value = nextPage
  void loadPosts()
}

function changePageSize(value: unknown) {
  pageSize.value = Number(value)
  page.value = 1
  void loadPosts()
}

function openCategoryDialog(post: AdminTopic) {
  categoryDialogRow.value = post
  selectedCategoryIds.value = [...(post.categoryId || [])]
}

function openPinDialog(post: AdminTopic) {
  pinDialogRow.value = post
  pinWeightInput.value = post.pinWeight || 0
}

function toggleCategory(id: number) {
  if (selectedCategoryIds.value.includes(id)) {
    selectedCategoryIds.value = selectedCategoryIds.value.filter(value => value !== id)
    return
  }
  if (selectedCategoryIds.value.length >= 3) {
    adminToast.warning(adminText('k003t'))
    return
  }
  selectedCategoryIds.value = [...selectedCategoryIds.value, id]
}

async function saveCategories() {
  if (!categoryDialogRow.value) return
  const validCategoryIds = selectedCategoryIds.value.filter(id => categoryMap.value.has(id))
  if (validCategoryIds.length === 0) {
    adminToast.warning(adminText('k003u'))
    return
  }
  if (categoryDialogRow.value.categoryId?.[0] !== validCategoryIds[0]) {
    pendingCategoryIds.value = validCategoryIds
    return
  }
  await persistCategories(validCategoryIds)
}

async function confirmCategoryChange() {
  const categoryIds = pendingCategoryIds.value
  if (!categoryIds) return
  pendingCategoryIds.value = null
  await persistCategories(categoryIds)
}

async function persistCategories(categoryIds: number[]) {
  const topic = categoryDialogRow.value
  if (!topic) return
  saving.value = true
  try {
    await updateTopicCategories({ topicId: topic.id, categoryId: categoryIds })
    categoryDialogRow.value = null
    await loadPosts()
    adminToast.success(adminText('k003v'))
  } catch (err) {
    adminToast.error(err, adminText('k0010'))
  } finally {
    saving.value = false
  }
}

async function savePinWeight() {
  if (!pinDialogRow.value) return
  const pinWeight = Math.max(0, Math.trunc(Number(pinWeightInput.value) || 0))
  saving.value = true
  try {
    await updateTopicPin({ topicId: pinDialogRow.value.id, pinWeight })
    pinDialogRow.value = null
    await loadPosts()
    adminToast.success(pinWeight > 0 ? adminText('k003w') : adminText('k003x'))
  } catch (err) {
    adminToast.error(err, adminText('k0016'))
  } finally {
    saving.value = false
  }
}

async function openSource(post: AdminTopic) {
  sourceDialogRow.value = post
  source.value = null
  reviewReason.value = ''
  sourceLoading.value = true
  try {
    const result = await getTopicSource(post.id)
    if (sourceDialogRow.value?.id === post.id) source.value = result
  } catch (err) {
    adminToast.error(err, adminText('k003y'))
  } finally {
    sourceLoading.value = false
  }
}

async function copySource() {
  if (!source.value?.content) return
  try {
    await navigator.clipboard.writeText(source.value.content)
    adminToast.success(adminText('k003z'))
  } catch (err) {
    adminToast.error(err, adminText('k0040'))
  }
}

async function toggleProcessStatus() {
  if (!actionRow.value) return
  const restoring = actionRow.value.processStatus === 1
  saving.value = true
  try {
    await editTopic({
      topicId: actionRow.value.id,
      processStatus: restoring ? 0 : 1,
    })
    actionRow.value = null
    await loadPosts()
    adminToast.success(restoring ? adminText('k0041') : adminText('k0042'))
  } catch (err) {
    adminToast.error(err, adminText('k001u'))
  } finally {
    saving.value = false
  }
}

async function confirmDeleteTopic() {
  if (!deleteRow.value) return
  saving.value = true
  try {
    await deleteTopic(deleteRow.value.id)
    deleteRow.value = null
    await loadPosts()
    adminToast.success(adminText('k00ce'))
  } catch (err) {
    adminToast.error(err, adminText('k00cd'))
  } finally {
    saving.value = false
  }
}

async function decide(action: 'approve' | 'reject' | 'recheck') {
  const item = kind.value === 'topic' ? source.value : reviewReply.value
  if (!item || !reviewReason.value.trim()) return
  saving.value = true
  try {
    await reviewContent(kind.value, { id: item.id, version: item.moderationVersion, action, reason: reviewReason.value.trim() })
    sourceDialogRow.value = null
    reviewReply.value = null
    await loadPosts()
    adminToast.success(adminText('reviewSaved'))
  } catch (err) { adminToast.error(err, adminText('reviewFailed')) }
  finally { saving.value = false }
}

onMounted(() => {
  void loadPosts()
})
</script>

<template>
  <BasicPage :title="adminText('k005u')" :description="adminText('k005v')">
      <div class="mb-4 flex flex-wrap items-center gap-2">
        <Select v-model="kind" @update:model-value="page = 1; loadPosts()">
          <SelectTrigger class="h-9 w-auto min-w-32" :aria-label="adminText('reviewTopics')"><SelectValue /></SelectTrigger>
          <SelectContent>
          <SelectItem value="topic">{{ adminText('reviewTopics') }}</SelectItem><SelectItem value="post">{{ adminText('reviewReplies') }}</SelectItem>
        </SelectContent></Select>
        <Select v-model="moderationStatus" @update:model-value="page = 1; loadPosts()">
          <SelectTrigger class="h-9 w-auto min-w-32" :aria-label="adminText('reviewAll')"><SelectValue /></SelectTrigger>
          <SelectContent>
          <SelectItem value="all">{{ adminText('reviewAll') }}</SelectItem>
          <SelectItem v-for="status in ['pending', 'approved', 'rejected', 'none']" :key="status" :value="status">{{ reviewLabel(status) }}</SelectItem>
        </SelectContent></Select>
        <Select v-if="kind === 'topic'" v-model="categoryId" @update:model-value="page = 1; loadPosts()">
          <SelectTrigger class="h-9 w-auto min-w-32" :aria-label="adminText('reviewAllCategories')"><SelectValue /></SelectTrigger>
          <SelectContent>
          <SelectItem value="all">{{ adminText('reviewAllCategories') }}</SelectItem>
          <SelectItem v-for="category in categories" :key="category.id" :value="String(category.id)">{{ category.category }}</SelectItem>
        </SelectContent></Select>
      </div>
      <AdminSection>
        <template #header>
        <AdminToolbar class="-mx-3 -my-2 border-b-0">
          <form class="flex min-w-0 flex-1 items-center gap-2" @submit.prevent="applySearch">
            <div class="relative min-w-64 max-w-xl flex-1">
              <Search class="pointer-events-none absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
              <Input v-model="search" class="h-9 pl-8 text-sm" :placeholder="adminText('k006a')" />
            </div>
            <Button type="submit" size="sm" class="h-9 px-4">{{ adminText('k00al') }}</Button>
            <Button v-if="appliedSearch" variant="ghost" size="sm" type="button" class="h-9" @click="search = ''; applySearch()">
              {{ adminText('k00at') }}
            </Button>
          </form>
          <div class="flex flex-wrap items-center gap-2 text-sm text-muted-foreground">
            <Button variant="outline" size="sm" type="button" :disabled="loading" @click="loadPosts">
              <RefreshCw class="size-4" :class="loading ? 'animate-spin' : ''" />
              {{ adminText('k004q') }}
            </Button>
            <span class="whitespace-nowrap">{{ rangeStart }}-{{ rangeEnd }}</span>
            <div class="h-4 w-px bg-border" />
            <div class="flex flex-wrap items-center gap-2">
              <Select :model-value="String(pageSize)" @update:model-value="changePageSize">
                <SelectTrigger class="h-9 w-auto"><SelectValue /></SelectTrigger>
                <SelectContent>
                <SelectItem value="10">{{ adminText('k002x') }}</SelectItem>
                <SelectItem value="20">{{ adminText('k002y') }}</SelectItem>
                <SelectItem value="50">{{ adminText('k0030') }}</SelectItem>
              </SelectContent></Select>
              <Button variant="outline" size="sm" type="button" :disabled="page <= 1 || loading" @click="changePage(page - 1)">{{ adminText('k00au') }}</Button>
              <span class="whitespace-nowrap">{{ adminText('k0056') }} {{ page }} {{ adminText('k0057') }}</span>
              <Button variant="outline" size="sm" type="button" :disabled="!hasNext || loading" @click="changePage(page + 1)">{{ adminText('k00av') }}</Button>
            </div>
          </div>
        </AdminToolbar>
        </template>

        <div v-if="kind === 'post'" class="divide-y">
          <p v-if="error" class="p-4 text-destructive">{{ error }}</p>
          <p v-else-if="!replyRows.length" class="p-4 text-muted-foreground">{{ loading ? adminText('k0046') : adminText('k00aw') }}</p>
          <article v-for="reply in replyRows" :key="reply.id" class="space-y-2 p-4">
            <div class="flex items-center justify-between gap-3">
              <span class="font-medium">{{ reply.topicTitle }} · #{{ reply.postNo }} · UID {{ reply.userId }}</span>
              <Badge>{{ reviewLabel(reply.moderationStatus) }}</Badge>
            </div>
            <p class="line-clamp-3 whitespace-pre-wrap break-words text-sm">{{ reply.content }}</p>
            <p class="break-words text-sm text-muted-foreground">{{ reply.moderationReason }}</p>
            <Button variant="outline" @click="reviewReply = reply; reviewReason = ''">{{ adminText('reviewOpen') }}</Button>
          </article>
        </div>
        <div v-if="kind === 'topic'" class="md:hidden">
          <div v-if="loading && rows.length === 0" class="px-3 py-10 text-center text-sm text-muted-foreground">{{ adminText('k0046') }}</div>
          <div v-else-if="error" class="px-3 py-10 text-center text-sm text-destructive">{{ error }}</div>
          <div v-else-if="rows.length === 0" class="px-3 py-10 text-center text-sm text-muted-foreground">{{ adminText('k00aw') }}</div>
          <div v-else class="divide-y">
            <article v-for="post in rows" :key="post.id" class="space-y-2 px-3 py-3">
              <div class="flex min-w-0 items-start justify-between gap-3">
                <div class="min-w-0 flex-1 space-y-1">
                  <div class="flex min-w-0 items-center gap-1.5">
                    <a :href="`/p/post/${post.id}`" target="_blank" rel="noreferrer" class="min-w-0 truncate text-[15px] font-semibold leading-5 text-foreground hover:text-primary hover:underline">
                      {{ post.title }}
                    </a>
                    <Badge v-if="post.moderationStatus && post.moderationStatus !== 'none'" variant="outline" class="shrink-0" :title="post.moderationReason">{{ reviewLabel(post.moderationStatus) }}</Badge>
                    <Badge v-if="post.processStatus === 1" variant="destructive" class="h-5 shrink-0 rounded-full px-1.5 text-[10px]">{{ adminText('k0069') }}</Badge>
                    <Badge v-if="post.deleted" variant="outline" class="h-5 shrink-0 rounded-full px-1.5 text-[10px]">{{ adminText('topicDeleted') }}</Badge>
                    <Badge v-if="post.pinWeight > 0" variant="secondary" class="h-5 shrink-0 rounded-full px-1.5 text-[10px]">{{ adminText('k00ax') }} {{ post.pinWeight }}</Badge>
                  </div>
                  <p class="line-clamp-2 break-words text-[12px] leading-5 text-muted-foreground">
                    {{ post.description || adminText('k005w') }}
                  </p>
                </div>
                <span class="inline-flex h-6 shrink-0 items-center rounded-md px-2 text-xs font-semibold" :class="topicStatusInfo(post).className">
                  {{ topicStatusInfo(post).label }}
                </span>
              </div>

              <div class="flex min-w-0 flex-wrap items-center gap-x-1.5 gap-y-1 text-[12px] leading-4 text-muted-foreground">
                <span v-for="category in postCategories(post)" :key="category.id" class="inline-flex h-5 max-w-32 items-center gap-1 rounded-full bg-muted px-1.5 font-medium" :class="category.missing ? 'text-destructive' : ''">
                  <span class="size-1.5 shrink-0 rounded-full" :style="{ backgroundColor: category.color || '#64748b' }" />
                  <span class="truncate">{{ category.category }}</span>
                </span>
                <span v-if="postCategories(post).length === 0" class="inline-flex h-5 items-center rounded-full bg-muted px-1.5">{{ adminText('k00ay') }}</span>
                <span class="inline-flex items-center gap-1"><Eye class="size-3.5" />{{ post.viewCount }}</span>
                <span class="inline-flex items-center gap-1"><MessageSquare class="size-3.5" />{{ post.replyCount }}</span>
                <span class="inline-flex items-center gap-1"><Heart class="size-3.5" />{{ post.likeCount }}</span>
              </div>

              <div class="flex items-center justify-between gap-3">
                <div class="flex min-w-0 items-center gap-2">
                  <img v-if="post.userAvatarUrl" :src="post.userAvatarUrl" class="size-7 shrink-0 rounded-full object-cover ring-1 ring-border" alt="" />
                  <span v-else class="flex size-7 shrink-0 items-center justify-center rounded-full bg-muted text-xs font-semibold">{{ avatarText(post) }}</span>
                  <div class="min-w-0">
                    <a :href="`/u/${post.userId}`" target="_blank" rel="noreferrer" class="block truncate text-[13px] font-semibold hover:text-primary hover:underline">
                      {{ post.username }}
                    </a>
                    <div class="truncate text-xs text-muted-foreground">{{ postDate(post.createdAt) }} {{ postTime(post.createdAt) }} · {{ post.deleted ? adminText('topicDeleted') : post.processStatus === 1 ? adminText('k005x') : adminText('k005y') }}</div>
                  </div>
                </div>
                <div class="flex shrink-0 items-center gap-0.5">
                  <AdminActionButton compact :title="adminText('reviewOpen')" @click="openSource(post)">
                    <FileText class="size-4" />
                  </AdminActionButton>
                  <AdminActionButton v-if="!post.deleted" compact :title="adminText('k006d')" @click="openCategoryDialog(post)">
                    <Tags class="size-4" />
                  </AdminActionButton>
                  <AdminActionButton v-if="!post.deleted" compact :tone="post.pinWeight > 0 ? 'primary' : 'default'" :title="adminText('k006e')" @click="openPinDialog(post)">
                    <Pin class="size-4" />
                  </AdminActionButton>
                  <AdminActionButton v-if="!post.deleted" compact :tone="post.processStatus === 1 ? 'success' : 'danger'" :title="post.processStatus === 1 ? adminText('k005z') : adminText('k0060')" @click="actionRow = post">
                    <Undo2 v-if="post.processStatus === 1" class="size-4" />
                    <Ban v-else class="size-4" />
                  </AdminActionButton>
                  <AdminActionButton v-if="!post.deleted" compact tone="danger" :title="adminText('k005i')" @click="deleteRow = post">
                    <Trash2 class="size-4" />
                  </AdminActionButton>
                </div>
              </div>
            </article>
          </div>
        </div>

          <Table v-if="kind === 'topic'" class="hidden table-fixed md:table">
            <TableHeader class="bg-muted/30">
              <TableRow>
                <TableHead class="px-3">{{ adminText('k00az') }}</TableHead>
                <TableHead class="w-[180px]">{{ adminText('k00b0') }}</TableHead>
                <TableHead class="w-[96px] text-center">{{ adminText('k007j') }}</TableHead>
                <TableHead class="w-[176px] text-right pr-3">{{ adminText('k007m') }}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-if="loading && rows.length === 0">
                <TableCell colspan="4" class="h-28 text-center text-muted-foreground">{{ adminText('k0046') }}</TableCell>
              </TableRow>
              <TableRow v-else-if="error">
                <TableCell colspan="4" class="h-28 text-center text-destructive">{{ error }}</TableCell>
              </TableRow>
              <TableRow v-else-if="rows.length === 0">
                <TableCell colspan="4" class="h-28 text-center text-muted-foreground">{{ adminText('k00aw') }}</TableCell>
              </TableRow>
              <template v-else>
                <TableRow v-for="post in rows" :key="post.id" class="group hover:bg-muted/20">
                  <TableCell class="max-w-0 whitespace-normal px-3 py-2">
                    <div class="min-w-0 space-y-1">
                      <div class="flex min-w-0 items-center gap-1.5">
                        <a :href="`/p/post/${post.id}`" target="_blank" rel="noreferrer" class="min-w-0 truncate text-[15px] font-semibold leading-5 text-foreground hover:text-primary hover:underline">
                          {{ post.title }}
                        </a>
                        <Badge v-if="post.moderationStatus && post.moderationStatus !== 'none'" variant="outline" class="shrink-0" :title="post.moderationReason">{{ reviewLabel(post.moderationStatus) }}</Badge>
                    <Badge v-if="post.processStatus === 1" variant="destructive" class="h-5 shrink-0 rounded-full px-1.5 text-[10px]">{{ adminText('k0069') }}</Badge>
                        <Badge v-if="post.pinWeight > 0" variant="secondary" class="h-5 shrink-0 rounded-full px-1.5 text-[10px]">{{ adminText('k00ax') }} {{ post.pinWeight }}</Badge>
                      </div>
                      <p class="truncate text-[12px] leading-4 text-muted-foreground">
                        {{ post.description || adminText('k005w') }}
                      </p>
                      <div class="flex min-w-0 flex-wrap items-center gap-x-1.5 gap-y-1 text-[12px] leading-4 text-muted-foreground">
                        <span class="whitespace-nowrap">{{ postDate(post.createdAt) }} {{ postTime(post.createdAt) }}</span>
                        <span v-for="category in postCategories(post)" :key="category.id" class="inline-flex h-5 max-w-32 items-center gap-1 rounded-full bg-muted px-1.5 font-medium" :class="category.missing ? 'text-destructive' : ''">
                          <span class="size-1.5 shrink-0 rounded-full" :style="{ backgroundColor: category.color || '#64748b' }" />
                          <span class="truncate">{{ category.category }}</span>
                        </span>
                        <span v-if="postCategories(post).length === 0" class="inline-flex h-5 items-center rounded-full bg-muted px-1.5">{{ adminText('k00ay') }}</span>
                        <span class="inline-flex items-center gap-1" :title="adminText('k006f')"><Eye class="size-3.5" />{{ post.viewCount }}</span>
                        <span class="inline-flex items-center gap-1" :title="adminText('k006g')"><MessageSquare class="size-3.5" />{{ post.replyCount }}</span>
                        <span class="inline-flex items-center gap-1" :title="adminText('k006h')"><Heart class="size-3.5" />{{ post.likeCount }}</span>
                      </div>
                    </div>
                  </TableCell>
                  <TableCell class="max-w-0 py-2 align-middle">
                    <div class="flex min-w-0 items-center gap-2">
                      <img v-if="post.userAvatarUrl" :src="post.userAvatarUrl" class="size-7 shrink-0 rounded-full object-cover ring-1 ring-border" alt="" />
                      <span v-else class="flex size-7 shrink-0 items-center justify-center rounded-full bg-muted text-xs font-semibold">{{ avatarText(post) }}</span>
                      <a :href="`/u/${post.userId}`" target="_blank" rel="noreferrer" class="min-w-0 truncate text-[13px] font-semibold hover:text-primary hover:underline">
                        {{ post.username }}
                      </a>
                    </div>
                  </TableCell>
                  <TableCell class="py-2 text-center align-middle">
                    <div class="inline-flex min-w-[52px] flex-col items-center gap-0.5">
                      <span class="inline-flex h-6 items-center rounded-md px-2 text-xs font-semibold" :class="topicStatusInfo(post).className">
                        {{ topicStatusInfo(post).label }}
                      </span>
                      <span class="text-[11px]" :class="post.processStatus === 1 ? 'text-destructive' : 'text-muted-foreground'">
                        {{ post.deleted ? adminText('topicDeleted') : post.processStatus === 1 ? adminText('k005x') : adminText('k005y') }}
                      </span>
                    </div>
                  </TableCell>
                  <TableCell class="pr-3">
                    <div class="flex justify-end gap-0.5">
                      <AdminActionButton compact :title="adminText('reviewOpen')" @click="openSource(post)">
                        <FileText class="size-4" />
                      </AdminActionButton>
                      <AdminActionButton v-if="!post.deleted" compact :title="adminText('k006d')" @click="openCategoryDialog(post)">
                        <Tags class="size-4" />
                      </AdminActionButton>
                      <AdminActionButton v-if="!post.deleted" compact :tone="post.pinWeight > 0 ? 'primary' : 'default'" :title="adminText('k006e')" @click="openPinDialog(post)">
                        <Pin class="size-4" />
                      </AdminActionButton>
                      <AdminActionButton v-if="!post.deleted" compact :tone="post.processStatus === 1 ? 'success' : 'danger'" :title="post.processStatus === 1 ? adminText('k005z') : adminText('k0060')" @click="actionRow = post">
                        <Undo2 v-if="post.processStatus === 1" class="size-4" />
                        <Ban v-else class="size-4" />
                      </AdminActionButton>
                      <AdminActionButton v-if="!post.deleted" compact tone="danger" :title="adminText('k005i')" @click="deleteRow = post">
                        <Trash2 class="size-4" />
                      </AdminActionButton>
                    </div>
                  </TableCell>
                </TableRow>
              </template>
            </TableBody>
          </Table>

      </AdminSection>

      <Dialog :open="categoryDialogRow !== null" @update:open="(open) => !open && (categoryDialogRow = null)">
        <DialogContent class="sm:max-w-xl">
          <DialogHeader>
            <DialogTitle>{{ adminText('k00b1') }}</DialogTitle>
            <DialogDescription class="line-clamp-2">
              {{ adminText('k00b2', { title: categoryDialogRow?.title || '' }) }}
            </DialogDescription>
          </DialogHeader>
          <div class="flex max-h-[46vh] flex-wrap gap-2 overflow-y-auto pr-1">
            <button
              v-for="category in categoryDialogOptions"
              :key="category.id"
              class="inline-flex max-w-full items-center gap-1.5 rounded-md border px-2.5 py-1.5 text-left text-sm font-medium transition-colors"
              :class="[
                selectedCategoryIds.includes(category.id) ? 'border-primary bg-primary/10 text-primary' : 'border-border text-muted-foreground hover:border-muted-foreground/30 hover:bg-muted/50',
                category.missing ? 'border-destructive/30 bg-destructive/5 text-destructive hover:bg-destructive/10' : '',
              ]"
              type="button"
              @click="toggleCategory(category.id)"
            >
              <span class="size-2 rounded-[3px]" :style="{ backgroundColor: category.color || '#64748b' }" />
              <span class="max-w-48 truncate">{{ category.category }}</span>
              <Badge v-if="selectedCategoryIds[0] === category.id" variant="secondary" class="px-1.5 py-0 text-[10px]">
                {{ t('publish.mainCategoryBadge') }}
              </Badge>
            </button>
          </div>
          <DialogFooter>
            <Button variant="outline" type="button" @click="categoryDialogRow = null">{{ adminText('k009q') }}</Button>
            <Button type="button" :disabled="saving" @click="saveCategories">{{ saving ? adminText('k005f') : adminText('k0061') }}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog :open="pendingCategoryIds !== null" @update:open="(open) => !open && (pendingCategoryIds = null)">
        <DialogContent class="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{{ t('publish.mainCategoryChangeTitle') }}</DialogTitle>
            <DialogDescription>
              {{ t('publish.mainCategoryChangeDescription', { current: currentMainCategoryName, next: nextMainCategoryName }) }}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" type="button" :disabled="saving" @click="pendingCategoryIds = null">{{ adminText('k009q') }}</Button>
            <Button type="button" :disabled="saving" @click="confirmCategoryChange">
              {{ saving ? adminText('k005f') : t('publish.confirmMainCategoryChange') }}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog :open="pinDialogRow !== null" @update:open="(open) => !open && (pinDialogRow = null)">
        <DialogContent class="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{{ adminText('k00b4') }}</DialogTitle>
            <DialogDescription class="line-clamp-2">
              {{ adminText('k00b5', { title: pinDialogRow?.title || '' }) }}
            </DialogDescription>
          </DialogHeader>
          <div class="space-y-2">
            <Input v-model.number="pinWeightInput" type="number" min="0" max="1000000" />
            <p class="text-xs text-muted-foreground">{{ adminText('k00b6') }}</p>
          </div>
          <DialogFooter>
            <Button variant="outline" type="button" @click="pinDialogRow = null">{{ adminText('k009q') }}</Button>
            <Button type="button" :disabled="saving" @click="savePinWeight">{{ saving ? adminText('k005f') : adminText('k005g') }}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog :open="sourceDialogRow !== null" @update:open="(open) => !open && (sourceDialogRow = null)">
        <DialogContent class="sm:max-w-3xl">
          <DialogHeader>
            <DialogTitle>{{ adminText('reviewOpen') }}</DialogTitle>
            <DialogDescription>{{ sourceDialogRow?.title }}</DialogDescription>
          </DialogHeader>
          <div class="max-h-[58vh] overflow-auto rounded-lg border bg-muted/20 p-4">
            <pre v-if="sourceLoading" class="text-sm text-muted-foreground">{{ adminText('k0046') }}</pre>
            <pre v-else class="whitespace-pre-wrap break-words text-sm leading-6">{{ source?.content || adminText('k0062') }}</pre>
          </div>
          <div v-if="source && !sourceLoading" class="space-y-3">
            <p class="break-words text-sm">{{ reviewLabel(source.moderationStatus) }} · {{ source.moderatedAt || '—' }}</p>
            <p class="break-words text-sm text-muted-foreground">{{ source.moderationReason }}</p>
            <template v-if="source.topicStatus === 1 || ['rejected', 'pending', 'denied'].includes(source.moderationStatus)">
              <Input v-model="reviewReason" :maxlength="300" :placeholder="adminText('reviewReason')" />
              <div class="flex flex-wrap gap-2">
                <Button :disabled="saving || !reviewReason.trim()" @click="decide('approve')">{{ adminText('reviewApprove') }}</Button>
                <Button variant="destructive" :disabled="saving || !reviewReason.trim()" @click="decide('reject')">{{ adminText('reviewReject') }}</Button>
                <Button variant="outline" :disabled="saving || !reviewReason.trim()" @click="decide('recheck')">{{ adminText('reviewRecheck') }}</Button>
              </div>
            </template>
          </div>
          <DialogFooter>
            <Button variant="outline" type="button" @click="sourceDialogRow = null">{{ adminText('k00b8') }}</Button>
            <Button type="button" :disabled="!source?.content" @click="copySource">{{ adminText('k00b9') }}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog :open="reviewReply !== null" @update:open="(open) => !open && (reviewReply = null)">
        <DialogContent class="sm:max-w-3xl">
          <DialogHeader><DialogTitle>{{ adminText('reviewReplies') }}</DialogTitle><DialogDescription>{{ reviewReply?.topicTitle }} · #{{ reviewReply?.postNo }}</DialogDescription></DialogHeader>
          <pre class="max-h-[45vh] overflow-auto whitespace-pre-wrap break-words rounded-md border p-4 text-sm">{{ reviewReply?.content }}</pre>
          <p class="break-words text-sm text-muted-foreground">{{ reviewReply?.moderationReason }}</p>
          <Input v-model="reviewReason" :maxlength="300" :placeholder="adminText('reviewReason')" />
          <DialogFooter>
            <Button :disabled="saving || !reviewReason.trim()" @click="decide('approve')">{{ adminText('reviewApprove') }}</Button>
            <Button variant="destructive" :disabled="saving || !reviewReason.trim()" @click="decide('reject')">{{ adminText('reviewReject') }}</Button>
            <Button variant="outline" :disabled="saving || !reviewReason.trim()" @click="decide('recheck')">{{ adminText('reviewRecheck') }}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog :open="actionRow !== null" @update:open="(open) => !open && (actionRow = null)">
        <DialogContent class="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{{ actionRow?.processStatus === 1 ? adminText('k0063') : adminText('k0064') }}</DialogTitle>
            <DialogDescription>
              {{ actionRow?.processStatus === 1 ? adminText('k0065') : adminText('k0066') }}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" type="button" @click="actionRow = null">{{ adminText('k009q') }}</Button>
            <Button :variant="actionRow?.processStatus === 1 ? 'default' : 'destructive'" type="button" :disabled="saving" @click="toggleProcessStatus">
              {{ saving ? adminText('k0067') : (actionRow?.processStatus === 1 ? adminText('k0068') : adminText('k0069')) }}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <AdminConfirmDialog
        :open="deleteRow !== null"
        :title="adminText('k00cf')"
        :description="adminText('k00cg', { title: deleteRow?.title || '' })"
        :loading="saving"
        @update:open="(open) => !open && (deleteRow = null)"
        @confirm="confirmDeleteTopic"
      />
    </BasicPage>
</template>
