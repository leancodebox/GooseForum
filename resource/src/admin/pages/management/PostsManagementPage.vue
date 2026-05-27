<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { Ban, ExternalLink, Eye, FileText, Heart, MessageSquare, Pin, RefreshCw, Search, Tags, Undo2 } from '@lucide/vue'
import { BasicPage } from '@/admin/components/global-layout'
import { Button } from '@/admin/components/ui/button'
import { Badge } from '@/admin/components/ui/badge'
import { Input } from '@/admin/components/ui/input'
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
  editArticle,
  editArticleCategories,
  editArticlePin,
  getArticleSource,
  getArticlesList,
  getCategoryList,
} from '@/admin/runtime/api'
import { adminToast } from '@/admin/runtime/toast'
import type { AdminArticle, AdminCategory, AdminPayload, ArticleSource, ManageHomeProps } from '@/admin/types'

defineProps<{
  payload: AdminPayload<ManageHomeProps>
}>()

const loading = ref(false)
const saving = ref(false)
const sourceLoading = ref(false)
const error = ref('')
const rows = ref<AdminArticle[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const search = ref('')
const appliedSearch = ref('')
const categories = ref<AdminCategory[]>([])
const categoryDialogRow = ref<AdminArticle | null>(null)
const selectedCategoryIds = ref<number[]>([])
const sourceDialogRow = ref<AdminArticle | null>(null)
const source = ref<ArticleSource | null>(null)
const actionRow = ref<AdminArticle | null>(null)
const pinDialogRow = ref<AdminArticle | null>(null)
const pinWeightInput = ref(0)

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))
const categoryMap = computed(() => new Map(categories.value.map(item => [item.id, item])))
const rangeStart = computed(() => (total.value === 0 ? 0 : (page.value - 1) * pageSize.value + 1))
const rangeEnd = computed(() => Math.min(page.value * pageSize.value, total.value))

const articleTypes: Record<number, { label: string, className: string }> = {
  0: { label: '博文', className: 'bg-blue-50 text-blue-700 border-blue-100' },
  1: { label: '分享', className: 'bg-emerald-50 text-emerald-700 border-emerald-100' },
  2: { label: '问答', className: 'bg-amber-50 text-amber-700 border-amber-100' },
  3: { label: '教程', className: 'bg-violet-50 text-violet-700 border-violet-100' },
}

function typeInfo(type: number) {
  return articleTypes[type] || { label: '文章', className: 'bg-slate-50 text-slate-700 border-slate-100' }
}

function avatarText(post: AdminArticle) {
  return post.username.slice(0, 1).toUpperCase()
}

function postCategories(post: AdminArticle) {
  return (post.categoryId || []).map(id => categoryMap.value.get(id)).filter(Boolean) as AdminCategory[]
}

function postDate(value?: string) {
  if (!value) return '-'
  return value.slice(0, 10)
}

function postTime(value?: string) {
  if (!value) return ''
  return value.slice(11, 16)
}

function articleStatusInfo(status: number) {
  return status === 1
    ? { label: '发布', className: 'bg-slate-950 text-white' }
    : { label: '草稿', className: 'bg-slate-100 text-slate-600' }
}

async function loadPosts() {
  loading.value = true
  error.value = ''
  try {
    const [postPage, categoryList] = await Promise.all([
      getArticlesList({ page: page.value, pageSize: pageSize.value, search: appliedSearch.value || undefined }),
      categories.value.length ? Promise.resolve(categories.value) : getCategoryList(),
    ])
    rows.value = postPage.list || []
    total.value = postPage.total || 0
    categories.value = categoryList || []
    if (page.value > totalPages.value) page.value = totalPages.value
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载帖子失败'
  } finally {
    loading.value = false
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

function changePageSize(event: Event) {
  pageSize.value = Number((event.target as HTMLSelectElement).value)
  page.value = 1
  void loadPosts()
}

function openCategoryDialog(post: AdminArticle) {
  categoryDialogRow.value = post
  selectedCategoryIds.value = [...(post.categoryId || [])]
}

function openPinDialog(post: AdminArticle) {
  pinDialogRow.value = post
  pinWeightInput.value = post.pinWeight || 0
}

function toggleCategory(id: number) {
  if (selectedCategoryIds.value.includes(id)) {
    selectedCategoryIds.value = selectedCategoryIds.value.filter(value => value !== id)
    return
  }
  if (selectedCategoryIds.value.length >= 3) {
    adminToast.warning('最多选择 3 个分类')
    return
  }
  selectedCategoryIds.value = [...selectedCategoryIds.value, id]
}

async function saveCategories() {
  if (!categoryDialogRow.value) return
  if (selectedCategoryIds.value.length === 0) {
    adminToast.warning('至少选择一个分类')
    return
  }
  saving.value = true
  try {
    await editArticleCategories({ id: categoryDialogRow.value.id, categoryId: selectedCategoryIds.value })
    categoryDialogRow.value = null
    await loadPosts()
    adminToast.success('分类已更新')
  } catch (err) {
    adminToast.error(err, '保存分类失败')
  } finally {
    saving.value = false
  }
}

async function savePinWeight() {
  if (!pinDialogRow.value) return
  const pinWeight = Math.max(0, Math.trunc(Number(pinWeightInput.value) || 0))
  saving.value = true
  try {
    await editArticlePin({ id: pinDialogRow.value.id, pinWeight })
    pinDialogRow.value = null
    await loadPosts()
    adminToast.success(pinWeight > 0 ? '置顶权重已更新' : '已取消置顶')
  } catch (err) {
    adminToast.error(err, '保存置顶权重失败')
  } finally {
    saving.value = false
  }
}

async function openSource(post: AdminArticle) {
  sourceDialogRow.value = post
  source.value = null
  sourceLoading.value = true
  try {
    source.value = await getArticleSource(post.id)
  } catch (err) {
    adminToast.error(err, '原文加载失败')
  } finally {
    sourceLoading.value = false
  }
}

function openPost(post: AdminArticle) {
  window.open(`/p/post/${post.id}`, '_blank', 'noopener,noreferrer')
}

async function copySource() {
  if (!source.value?.content) return
  try {
    await navigator.clipboard.writeText(source.value.content)
    adminToast.success('已复制原文')
  } catch (err) {
    adminToast.error(err, '复制失败')
  }
}

async function toggleProcessStatus() {
  if (!actionRow.value) return
  const restoring = actionRow.value.processStatus === 1
  saving.value = true
  try {
    await editArticle({
      id: actionRow.value.id,
      processStatus: restoring ? 0 : 1,
    })
    actionRow.value = null
    await loadPosts()
    adminToast.success(restoring ? '已恢复帖子' : '已封禁帖子')
  } catch (err) {
    adminToast.error(err, '操作失败')
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  void loadPosts()
})
</script>

<template>
  <BasicPage title="帖子管理" description="审核内容、调整分类、查看原文和处理异常帖子。" sticky>
    <template #actions>
      <Button variant="outline" type="button" @click="loadPosts">
        <RefreshCw class="size-4" />
        刷新
      </Button>
    </template>

      <div class="overflow-hidden rounded-lg border bg-card">
        <div class="flex flex-col gap-2 border-b bg-muted/10 px-3 py-2 lg:flex-row lg:items-center lg:justify-between">
          <form class="flex min-w-0 flex-1 items-center gap-2" @submit.prevent="applySearch">
            <div class="relative min-w-64 max-w-xl flex-1">
              <Search class="pointer-events-none absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
              <Input v-model="search" class="h-9 pl-8 text-sm" placeholder="搜索标题、摘要或异常内容..." />
            </div>
            <Button type="submit" size="sm" class="h-9 px-4">搜索</Button>
            <Button v-if="appliedSearch" variant="ghost" size="sm" type="button" class="h-9" @click="search = ''; applySearch()">
              清除
            </Button>
          </form>
          <div class="flex flex-wrap items-center gap-2 text-sm text-muted-foreground">
            <span class="whitespace-nowrap">{{ rangeStart }}-{{ rangeEnd }} / {{ total }}</span>
            <div class="h-4 w-px bg-border" />
            <div class="flex flex-wrap items-center gap-2">
              <select class="h-9 rounded-md border bg-background px-2 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring" :value="pageSize" @change="changePageSize">
                <option :value="10">10 / 页</option>
                <option :value="20">20 / 页</option>
                <option :value="50">50 / 页</option>
              </select>
              <Button variant="outline" size="sm" type="button" :disabled="page <= 1 || loading" @click="changePage(page - 1)">上一页</Button>
              <span class="whitespace-nowrap">{{ page }} / {{ totalPages }}</span>
              <Button variant="outline" size="sm" type="button" :disabled="page >= totalPages || loading" @click="changePage(page + 1)">下一页</Button>
            </div>
          </div>
        </div>

        <div class="md:hidden">
          <div v-if="loading" class="px-3 py-10 text-center text-sm text-muted-foreground">加载中...</div>
          <div v-else-if="error" class="px-3 py-10 text-center text-sm text-destructive">{{ error }}</div>
          <div v-else-if="rows.length === 0" class="px-3 py-10 text-center text-sm text-muted-foreground">暂无帖子</div>
          <div v-else class="divide-y">
            <article v-for="post in rows" :key="post.id" class="space-y-2 px-3 py-3">
              <div class="flex min-w-0 items-start justify-between gap-3">
                <div class="min-w-0 flex-1 space-y-1">
                  <div class="flex min-w-0 items-center gap-1.5">
                    <a :href="`/p/post/${post.id}`" target="_blank" rel="noreferrer" class="min-w-0 truncate text-[15px] font-semibold leading-5 text-foreground hover:text-primary hover:underline">
                      {{ post.title }}
                    </a>
                    <span class="inline-flex h-5 shrink-0 items-center rounded-full border px-1.5 text-[11px] font-semibold" :class="typeInfo(post.type).className">
                      {{ typeInfo(post.type).label }}
                    </span>
                    <Badge v-if="post.processStatus === 1" variant="destructive" class="h-5 shrink-0 rounded-full px-1.5 text-[10px]">封禁</Badge>
                    <Badge v-if="post.pinWeight > 0" variant="secondary" class="h-5 shrink-0 rounded-full px-1.5 text-[10px]">置顶 {{ post.pinWeight }}</Badge>
                  </div>
                  <p class="line-clamp-2 break-words text-[12px] leading-5 text-muted-foreground">
                    {{ post.description || '暂无摘要' }}
                  </p>
                </div>
                <span class="inline-flex h-6 shrink-0 items-center rounded-md px-2 text-xs font-semibold" :class="articleStatusInfo(post.articleStatus).className">
                  {{ articleStatusInfo(post.articleStatus).label }}
                </span>
              </div>

              <div class="flex min-w-0 flex-wrap items-center gap-x-1.5 gap-y-1 text-[12px] leading-4 text-muted-foreground">
                <span v-for="category in postCategories(post)" :key="category.id" class="inline-flex h-5 max-w-32 items-center gap-1 rounded-full bg-muted px-1.5 font-medium">
                  <span class="size-1.5 shrink-0 rounded-full" :style="{ backgroundColor: category.color || '#64748b' }" />
                  <span class="truncate">{{ category.category }}</span>
                </span>
                <span v-if="postCategories(post).length === 0" class="inline-flex h-5 items-center rounded-full bg-muted px-1.5">未分类</span>
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
                    <div class="truncate text-xs text-muted-foreground">{{ postDate(post.createdAt) }} {{ postTime(post.createdAt) }} · {{ post.processStatus === 1 ? '已处理' : '正常' }}</div>
                  </div>
                </div>
                <div class="flex shrink-0 items-center gap-1">
                  <Button variant="ghost" size="icon-sm" type="button" title="查看帖子" @click="openPost(post)">
                    <ExternalLink class="size-4" />
                  </Button>
                  <Button variant="ghost" size="icon-sm" type="button" title="查看原文" @click="openSource(post)">
                    <FileText class="size-4" />
                  </Button>
                  <Button variant="ghost" size="icon-sm" type="button" title="修改分类" @click="openCategoryDialog(post)">
                    <Tags class="size-4" />
                  </Button>
                  <Button variant="ghost" size="icon-sm" type="button" :class="post.pinWeight > 0 ? 'text-primary hover:text-primary' : ''" title="置顶权重" @click="openPinDialog(post)">
                    <Pin class="size-4" />
                  </Button>
                  <Button variant="ghost" size="icon-sm" type="button" :class="post.processStatus === 1 ? 'text-emerald-600 hover:text-emerald-700' : 'text-destructive hover:text-destructive'" :title="post.processStatus === 1 ? '恢复帖子' : '封禁帖子'" @click="actionRow = post">
                    <Undo2 v-if="post.processStatus === 1" class="size-4" />
                    <Ban v-else class="size-4" />
                  </Button>
                </div>
              </div>
            </article>
          </div>
        </div>

          <Table class="hidden table-fixed md:table">
            <TableHeader class="bg-muted/30">
              <TableRow>
                <TableHead class="px-3">帖子</TableHead>
                <TableHead class="w-[190px]">作者</TableHead>
                <TableHead class="w-[112px]">状态</TableHead>
                <TableHead class="w-[118px]">时间</TableHead>
                <TableHead class="w-[168px] text-right pr-3">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-if="loading">
                <TableCell colspan="5" class="h-28 text-center text-muted-foreground">加载中...</TableCell>
              </TableRow>
              <TableRow v-else-if="error">
                <TableCell colspan="5" class="h-28 text-center text-destructive">{{ error }}</TableCell>
              </TableRow>
              <TableRow v-else-if="rows.length === 0">
                <TableCell colspan="5" class="h-28 text-center text-muted-foreground">暂无帖子</TableCell>
              </TableRow>
              <template v-else>
                <TableRow v-for="post in rows" :key="post.id" class="group hover:bg-muted/20">
                  <TableCell class="max-w-0 whitespace-normal px-3 py-2">
                    <div class="min-w-0 space-y-1">
                      <div class="flex min-w-0 items-center gap-1.5">
                        <a :href="`/p/post/${post.id}`" target="_blank" rel="noreferrer" class="min-w-0 truncate text-[15px] font-semibold leading-5 text-foreground hover:text-primary hover:underline">
                          {{ post.title }}
                        </a>
                        <span class="inline-flex h-5 shrink-0 items-center rounded-full border px-1.5 text-[11px] font-semibold" :class="typeInfo(post.type).className">
                          {{ typeInfo(post.type).label }}
                        </span>
                        <Badge v-if="post.processStatus === 1" variant="destructive" class="h-5 shrink-0 rounded-full px-1.5 text-[10px]">封禁</Badge>
                        <Badge v-if="post.pinWeight > 0" variant="secondary" class="h-5 shrink-0 rounded-full px-1.5 text-[10px]">置顶 {{ post.pinWeight }}</Badge>
                      </div>
                      <p class="truncate text-[12px] leading-4 text-muted-foreground">
                        {{ post.description || '暂无摘要' }}
                      </p>
                      <div class="flex min-w-0 flex-wrap items-center gap-x-1.5 gap-y-1 text-[12px] leading-4 text-muted-foreground">
                        <span v-for="category in postCategories(post)" :key="category.id" class="inline-flex h-5 max-w-32 items-center gap-1 rounded-full bg-muted px-1.5 font-medium">
                          <span class="size-1.5 shrink-0 rounded-full" :style="{ backgroundColor: category.color || '#64748b' }" />
                          <span class="truncate">{{ category.category }}</span>
                        </span>
                        <span v-if="postCategories(post).length === 0" class="inline-flex h-5 items-center rounded-full bg-muted px-1.5">未分类</span>
                        <span class="inline-flex items-center gap-1" title="浏览量"><Eye class="size-3.5" />{{ post.viewCount }}</span>
                        <span class="inline-flex items-center gap-1" title="回复数"><MessageSquare class="size-3.5" />{{ post.replyCount }}</span>
                        <span class="inline-flex items-center gap-1" title="点赞数"><Heart class="size-3.5" />{{ post.likeCount }}</span>
                      </div>
                    </div>
                  </TableCell>
                  <TableCell class="max-w-0 py-2">
                    <div class="flex min-w-0 items-center gap-2">
                      <img v-if="post.userAvatarUrl" :src="post.userAvatarUrl" class="size-7 shrink-0 rounded-full object-cover ring-1 ring-border" alt="" />
                      <span v-else class="flex size-7 shrink-0 items-center justify-center rounded-full bg-muted text-xs font-semibold">{{ avatarText(post) }}</span>
                      <a :href="`/u/${post.userId}`" target="_blank" rel="noreferrer" class="min-w-0 truncate text-[13px] font-semibold hover:text-primary hover:underline">
                        {{ post.username }}
                      </a>
                    </div>
                  </TableCell>
                  <TableCell class="py-2">
                    <div class="flex flex-col items-start gap-0.5">
                      <span class="inline-flex h-6 items-center rounded-md px-2 text-xs font-semibold" :class="articleStatusInfo(post.articleStatus).className">
                        {{ articleStatusInfo(post.articleStatus).label }}
                      </span>
                      <span class="text-[11px]" :class="post.processStatus === 1 ? 'text-destructive' : 'text-muted-foreground'">
                        {{ post.processStatus === 1 ? '已处理' : '正常' }}
                      </span>
                    </div>
                  </TableCell>
                  <TableCell class="py-2">
                    <div class="whitespace-nowrap text-[13px] text-foreground">{{ postDate(post.createdAt) }}</div>
                    <div class="text-xs text-muted-foreground">{{ postTime(post.createdAt) }}</div>
                  </TableCell>
                  <TableCell class="pr-3">
                    <div class="flex justify-end gap-1">
                      <Button variant="ghost" size="icon-sm" type="button" title="查看帖子" @click="openPost(post)">
                        <ExternalLink class="size-4" />
                      </Button>
                      <Button variant="ghost" size="icon-sm" type="button" title="查看原文" @click="openSource(post)">
                        <FileText class="size-4" />
                      </Button>
                      <Button variant="ghost" size="icon-sm" type="button" title="修改分类" @click="openCategoryDialog(post)">
                        <Tags class="size-4" />
                      </Button>
                      <Button variant="ghost" size="icon-sm" type="button" :class="post.pinWeight > 0 ? 'text-primary hover:text-primary' : ''" title="置顶权重" @click="openPinDialog(post)">
                        <Pin class="size-4" />
                      </Button>
                      <Button variant="ghost" size="icon-sm" type="button" :class="post.processStatus === 1 ? 'text-emerald-600 hover:text-emerald-700' : 'text-destructive hover:text-destructive'" :title="post.processStatus === 1 ? '恢复帖子' : '封禁帖子'" @click="actionRow = post">
                        <Undo2 v-if="post.processStatus === 1" class="size-4" />
                        <Ban v-else class="size-4" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              </template>
            </TableBody>
          </Table>

      </div>

      <Dialog :open="categoryDialogRow !== null" @update:open="(open) => !open && (categoryDialogRow = null)">
        <DialogContent class="sm:max-w-xl">
          <DialogHeader>
            <DialogTitle>修改文章分类</DialogTitle>
            <DialogDescription class="line-clamp-2">
              为「{{ categoryDialogRow?.title }}」选择 1 到 3 个分类。
            </DialogDescription>
          </DialogHeader>
          <div class="grid gap-2 rounded-lg border bg-muted/20 p-3 sm:grid-cols-2">
            <button
              v-for="category in categories"
              :key="category.id"
              class="flex min-h-12 items-center gap-3 rounded-md border bg-background px-3 py-2 text-left text-sm transition-colors hover:border-primary/50 hover:bg-primary/5"
              :class="selectedCategoryIds.includes(category.id) ? 'border-primary bg-primary/10 text-primary' : ''"
              type="button"
              @click="toggleCategory(category.id)"
            >
              <span class="size-2.5 rounded-full" :style="{ backgroundColor: category.color || '#64748b' }" />
              <span class="min-w-0 flex-1 truncate font-medium">{{ category.category }}</span>
              <span class="grid size-5 place-items-center rounded-full border text-[11px] font-bold" :class="selectedCategoryIds.includes(category.id) ? 'border-primary bg-primary text-primary-foreground' : 'border-muted-foreground/30 text-transparent'">✓</span>
            </button>
          </div>
          <DialogFooter>
            <Button variant="outline" type="button" @click="categoryDialogRow = null">取消</Button>
            <Button type="button" :disabled="saving" @click="saveCategories">{{ saving ? '保存中...' : '保存分类' }}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog :open="pinDialogRow !== null" @update:open="(open) => !open && (pinDialogRow = null)">
        <DialogContent class="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>设置全站置顶权重</DialogTitle>
            <DialogDescription class="line-clamp-2">
              「{{ pinDialogRow?.title }}」权重为 0 表示不置顶，数字越大越靠前。
            </DialogDescription>
          </DialogHeader>
          <div class="space-y-2">
            <Input v-model.number="pinWeightInput" type="number" min="0" max="1000000" />
            <p class="text-xs text-muted-foreground">建议从 100 开始设置，后续可用更大的数字调整顺序。</p>
          </div>
          <DialogFooter>
            <Button variant="outline" type="button" @click="pinDialogRow = null">取消</Button>
            <Button type="button" :disabled="saving" @click="savePinWeight">{{ saving ? '保存中...' : '保存' }}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog :open="sourceDialogRow !== null" @update:open="(open) => !open && (sourceDialogRow = null)">
        <DialogContent class="sm:max-w-3xl">
          <DialogHeader>
            <DialogTitle>查看帖子原文</DialogTitle>
            <DialogDescription>{{ sourceDialogRow?.title }}</DialogDescription>
          </DialogHeader>
          <div class="max-h-[58vh] overflow-auto rounded-lg border bg-muted/20 p-4">
            <pre v-if="sourceLoading" class="text-sm text-muted-foreground">加载中...</pre>
            <pre v-else class="whitespace-pre-wrap break-words text-sm leading-6">{{ source?.content || '暂无内容' }}</pre>
          </div>
          <DialogFooter>
            <Button variant="outline" type="button" @click="sourceDialogRow = null">关闭</Button>
            <Button type="button" :disabled="!source?.content" @click="copySource">复制原文</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog :open="actionRow !== null" @update:open="(open) => !open && (actionRow = null)">
        <DialogContent class="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{{ actionRow?.processStatus === 1 ? '确认恢复文章？' : '确认封禁文章？' }}</DialogTitle>
            <DialogDescription>
              {{ actionRow?.processStatus === 1 ? '恢复后用户可以重新查看该文章。' : '封禁后用户将无法查看该文章。' }}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" type="button" @click="actionRow = null">取消</Button>
            <Button :variant="actionRow?.processStatus === 1 ? 'default' : 'destructive'" type="button" :disabled="saving" @click="toggleProcessStatus">
              {{ saving ? '处理中...' : (actionRow?.processStatus === 1 ? '恢复' : '封禁') }}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </BasicPage>
</template>
