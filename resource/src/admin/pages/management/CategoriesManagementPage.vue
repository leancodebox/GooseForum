<script setup lang="ts">import { adminText } from '@/admin/runtime/i18n-text'

import { computed, onMounted, reactive, ref, watch } from 'vue'
import { Pencil, Plus, RefreshCw, Search, ShieldCheck, Trash2, UserPlus, X } from '@lucide/vue'
import AdminActionButton from '@/admin/components/AdminActionButton.vue'
import AdminConfirmDialog from '@/admin/components/AdminConfirmDialog.vue'
import AdminSection from '@/admin/components/AdminSection.vue'
import AdminToolbar from '@/admin/components/AdminToolbar.vue'
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
import { addCategoryModerator, addGlobalModerator, deleteCategory, deleteCategoryModerator, deleteGlobalModerator, getCategoryList, getGlobalModeratorList, getUserList, saveCategory } from '@/admin/runtime/api'
import { adminToast } from '@/admin/runtime/toast'
import type { AdminCategory, AdminCategoryModerator, AdminPayload, AdminUser, ManageHomeProps } from '@/admin/types'

defineProps<{
  payload: AdminPayload<ManageHomeProps>
}>()

const presetColors = [
  '#64748b',
  '#ef4444',
  '#f97316',
  '#f59e0b',
  '#22c55e',
  '#10b981',
  '#06b6d4',
  '#3b82f6',
  '#6366f1',
  '#8b5cf6',
  '#a855f7',
  '#ec4899',
]

const loading = ref(false)
const saving = ref(false)
const deleting = ref(false)
const error = ref('')
const rows = ref<AdminCategory[]>([])
const search = ref('')
const dialogMode = ref<'create' | 'edit' | null>(null)
const deletingRow = ref<AdminCategory | null>(null)
const moderatorRow = ref<AdminCategory | null>(null)
const moderatorSaving = ref(false)
const moderatorUserInput = ref('')
const moderatorSearching = ref(false)
const moderatorCandidates = ref<AdminUser[]>([])
const selectedModeratorUser = ref<AdminUser | null>(null)
let moderatorSearchTimer: ReturnType<typeof setTimeout> | undefined
const globalModerators = ref<AdminCategoryModerator[]>([])
const globalModeratorInput = ref('')
const globalModeratorSaving = ref(false)
const globalModeratorSearching = ref(false)
const globalModeratorCandidates = ref<AdminUser[]>([])
const selectedGlobalModeratorUser = ref<AdminUser | null>(null)
let globalModeratorSearchTimer: ReturnType<typeof setTimeout> | undefined
const form = reactive<AdminCategory>({
  id: 0,
  category: '',
  desc: '',
  icon: '',
  color: '',
  slug: '',
  sort: 0,
})

const filteredRows = computed(() => {
  const keyword = search.value.trim().toLowerCase()
  if (!keyword) return rows.value
  return rows.value.filter((item) => {
    return [item.category, item.slug, item.desc].some(value => String(value || '').toLowerCase().includes(keyword))
  })
})

async function loadCategories() {
  loading.value = true
  error.value = ''
  try {
    rows.value = await getCategoryList()
  } catch (err) {
    error.value = err instanceof Error ? err.message : adminText('k0043')
  } finally {
    loading.value = false
  }
}

async function loadGlobalModerators() {
  try {
    globalModerators.value = await getGlobalModeratorList()
  } catch (err) {
    adminToast.error(err, '全站版主加载失败')
  }
}

function resetForm(row?: AdminCategory) {
  form.id = row?.id || 0
  form.category = row?.category || ''
  form.desc = row?.desc || ''
  form.icon = row?.icon || ''
  form.color = row?.color || ''
  form.slug = row?.slug || ''
  form.sort = row?.sort || 0
}

function openCreate() {
  resetForm()
  dialogMode.value = 'create'
}

function openEdit(row: AdminCategory) {
  resetForm(row)
  dialogMode.value = 'edit'
}

async function submitCategory() {
  if (!form.category.trim()) {
    adminToast.warning(adminText('k0044'))
    return
  }
  saving.value = true
  try {
    await saveCategory({
      ...form,
      id: dialogMode.value === 'edit' ? form.id : 0,
      category: form.category.trim(),
      slug: form.slug?.trim(),
      icon: form.icon?.trim(),
      color: form.color?.trim(),
      desc: form.desc?.trim(),
      sort: Number(form.sort || 0),
    })
    dialogMode.value = null
    await loadCategories()
    adminToast.success(adminText('k000e'))
  } catch (err) {
    adminToast.error(err, adminText('k0010'))
  } finally {
    saving.value = false
  }
}

async function confirmDelete() {
  if (!deletingRow.value) return
  deleting.value = true
  try {
    await deleteCategory(deletingRow.value.id)
    deletingRow.value = null
    await loadCategories()
    adminToast.success(adminText('k002u'))
  } catch (err) {
    adminToast.error(err, adminText('k0011'))
  } finally {
    deleting.value = false
  }
}

function openModerators(row: AdminCategory) {
  moderatorRow.value = row
  moderatorUserInput.value = ''
  selectedModeratorUser.value = null
  moderatorCandidates.value = []
}

function moderatorInitial(name: string) {
  return (name || '?').slice(0, 1).toUpperCase()
}

function visibleModerators(row: AdminCategory) {
  return (row.moderators || []).slice(0, 3)
}

function isAlreadyModerator(userId: number) {
  return Boolean(moderatorRow.value?.moderators?.some(item => item.userId === userId))
}

function isAlreadyGlobalModerator(userId: number) {
  return globalModerators.value.some(item => item.userId === userId)
}

async function searchModeratorUsers(keyword: string) {
  const value = keyword.trim()
  selectedModeratorUser.value = null
  if (!value) {
    moderatorCandidates.value = []
    return
  }
  moderatorSearching.value = true
  try {
    const params = /^\d+$/.test(value)
      ? { userId: Number(value), page: 1, pageSize: 8 }
      : { username: value, page: 1, pageSize: 8 }
    const result = await getUserList(params)
    moderatorCandidates.value = result.list || []
  } catch {
    moderatorCandidates.value = []
  } finally {
    moderatorSearching.value = false
  }
}

async function searchGlobalModeratorUsers(keyword: string) {
  const value = keyword.trim()
  selectedGlobalModeratorUser.value = null
  if (!value) {
    globalModeratorCandidates.value = []
    return
  }
  globalModeratorSearching.value = true
  try {
    const params = /^\d+$/.test(value)
      ? { userId: Number(value), page: 1, pageSize: 8 }
      : { username: value, page: 1, pageSize: 8 }
    const result = await getUserList(params)
    globalModeratorCandidates.value = result.list || []
  } catch {
    globalModeratorCandidates.value = []
  } finally {
    globalModeratorSearching.value = false
  }
}

function selectModeratorCandidate(user: AdminUser) {
  if (isAlreadyModerator(user.userId)) return
  selectedModeratorUser.value = user
  moderatorUserInput.value = user.username || String(user.userId)
  moderatorCandidates.value = []
}

function selectGlobalModeratorCandidate(user: AdminUser) {
  if (isAlreadyGlobalModerator(user.userId)) return
  selectedGlobalModeratorUser.value = user
  globalModeratorInput.value = user.username || String(user.userId)
  globalModeratorCandidates.value = []
}

async function addGlobalModeratorUser() {
  const value = globalModeratorInput.value.trim()
  if (!value) {
    adminToast.warning('请输入用户名或用户 ID')
    return
  }
  globalModeratorSaving.value = true
  try {
    const userId = selectedGlobalModeratorUser.value?.userId || (/^\d+$/.test(value) ? Number(value) : undefined)
    await addGlobalModerator({
      userId,
      username: userId ? undefined : value,
    })
    await loadGlobalModerators()
    globalModeratorInput.value = ''
    selectedGlobalModeratorUser.value = null
    globalModeratorCandidates.value = []
    adminToast.success('全站版主已添加')
  } catch (err) {
    adminToast.error(err, '添加全站版主失败')
  } finally {
    globalModeratorSaving.value = false
  }
}

async function addModerator() {
  if (!moderatorRow.value) return
  const value = moderatorUserInput.value.trim()
  if (!value) {
    adminToast.warning('请输入用户名或用户 ID')
    return
  }
  moderatorSaving.value = true
  try {
    const userId = selectedModeratorUser.value?.userId || (/^\d+$/.test(value) ? Number(value) : undefined)
    await addCategoryModerator({
      categoryId: moderatorRow.value.id,
      userId,
      username: userId ? undefined : value,
    })
    await loadCategories()
    const freshRow = rows.value.find(item => item.id === moderatorRow.value?.id)
    moderatorRow.value = freshRow || moderatorRow.value
    moderatorUserInput.value = ''
    selectedModeratorUser.value = null
    moderatorCandidates.value = []
    adminToast.success('版主已添加')
  } catch (err) {
    adminToast.error(err, '添加版主失败')
  } finally {
    moderatorSaving.value = false
  }
}

watch(moderatorUserInput, (value) => {
  if (!moderatorRow.value) return
  if (moderatorSearchTimer) clearTimeout(moderatorSearchTimer)
  moderatorSearchTimer = setTimeout(() => {
    void searchModeratorUsers(value)
  }, 220)
})

watch(globalModeratorInput, (value) => {
  if (globalModeratorSearchTimer) clearTimeout(globalModeratorSearchTimer)
  globalModeratorSearchTimer = setTimeout(() => {
    void searchGlobalModeratorUsers(value)
  }, 220)
})

async function removeGlobalModerator(id: number) {
  globalModeratorSaving.value = true
  try {
    await deleteGlobalModerator(id)
    await loadGlobalModerators()
    adminToast.success('全站版主已移除')
  } catch (err) {
    adminToast.error(err, '移除全站版主失败')
  } finally {
    globalModeratorSaving.value = false
  }
}

async function removeModerator(id: number) {
  if (!moderatorRow.value) return
  moderatorSaving.value = true
  try {
    await deleteCategoryModerator(id)
    await loadCategories()
    const freshRow = rows.value.find(item => item.id === moderatorRow.value?.id)
    moderatorRow.value = freshRow || moderatorRow.value
    adminToast.success('版主已移除')
  } catch (err) {
    adminToast.error(err, '移除版主失败')
  } finally {
    moderatorSaving.value = false
  }
}

onMounted(() => {
  void loadCategories()
  void loadGlobalModerators()
})
</script>

<template>
  <BasicPage :title="adminText('k005l')" :description="adminText('k0045')" sticky>
    <template #actions>
      <div class="flex items-center gap-2">
        <Button variant="outline" type="button" @click="loadCategories">
          <RefreshCw class="size-4" />
          {{ adminText('k004q') }}
        </Button>
        <Button type="button" @click="openCreate">
          <Plus class="size-4" />
          {{ adminText('k005m') }}
        </Button>
      </div>
    </template>

      <AdminSection>
        <template #header>
          <div class="-mx-3 -my-2 divide-y">
            <AdminToolbar class="border-b-0">
              <div class="relative w-full max-w-md">
                <Search class="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                <Input v-model="search" class="pl-9" :placeholder="adminText('k005q')" />
              </div>
              <Badge variant="secondary" class="h-9 rounded-md px-3">
                {{ filteredRows.length }} {{ adminText('k00c1') }}
              </Badge>
            </AdminToolbar>
            <div class="grid gap-2 px-3 py-2 lg:grid-cols-[auto_minmax(0,1fr)_20rem] lg:items-start">
              <div class="flex h-8 min-w-0 items-center gap-2 text-sm">
                <ShieldCheck class="size-4 shrink-0 text-muted-foreground" />
                <span class="shrink-0 font-medium">全站版主</span>
              </div>
              <div class="flex min-h-8 min-w-0 flex-wrap content-start gap-1">
                <span v-if="!globalModerators.length" class="inline-flex h-7 items-center text-sm text-muted-foreground">暂无</span>
                <span
                  v-for="moderator in globalModerators"
                  v-else
                  :key="moderator.id"
                  class="inline-flex h-7 max-w-40 items-center gap-1 rounded-md border bg-background px-1.5 text-xs"
                >
                  <img v-if="moderator.avatarUrl" :src="moderator.avatarUrl" class="size-4 rounded-full object-cover" alt="" />
                  <span v-else class="grid size-4 place-items-center rounded-full bg-muted text-[9px] font-semibold">{{ moderatorInitial(moderator.username) }}</span>
                  <span class="truncate">{{ moderator.username || `#${moderator.userId}` }}</span>
                  <button
                    class="-mr-1 grid size-5 place-items-center rounded text-muted-foreground hover:bg-muted hover:text-foreground"
                    type="button"
                    :disabled="globalModeratorSaving"
                    @click="removeGlobalModerator(moderator.id)"
                  >
                    <X class="size-3" />
                  </button>
                </span>
              </div>
              <form class="flex gap-1.5" @submit.prevent="addGlobalModeratorUser">
                <div class="relative min-w-0 flex-1">
                  <Input v-model="globalModeratorInput" class="h-8" placeholder="用户名或用户 ID" autocomplete="off" />
                  <div
                    v-if="globalModeratorInput.trim() && (globalModeratorCandidates.length || globalModeratorSearching)"
                    class="absolute left-0 right-0 top-[calc(100%+4px)] z-50 overflow-hidden rounded-md border bg-popover shadow-md"
                  >
                    <div v-if="globalModeratorSearching" class="px-3 py-2 text-sm text-muted-foreground">搜索中...</div>
                    <button
                      v-for="user in globalModeratorCandidates"
                      v-else
                      :key="user.userId"
                      class="flex w-full items-center gap-2 px-3 py-2 text-left text-sm transition-colors"
                      :class="isAlreadyGlobalModerator(user.userId) ? 'cursor-default opacity-55' : 'hover:bg-muted'"
                      type="button"
                      :disabled="isAlreadyGlobalModerator(user.userId)"
                      @click="selectGlobalModeratorCandidate(user)"
                    >
                      <img v-if="user.avatarUrl" :src="user.avatarUrl" class="size-7 rounded-full object-cover ring-1 ring-border" alt="" />
                      <span v-else class="flex size-7 items-center justify-center rounded-full bg-muted text-xs font-semibold">{{ moderatorInitial(user.username) }}</span>
                      <span class="min-w-0 flex-1 truncate">{{ user.username }}</span>
                      <span v-if="isAlreadyGlobalModerator(user.userId)" class="shrink-0 rounded-full bg-muted px-1.5 py-0.5 text-[11px] text-muted-foreground">已添加</span>
                      <span class="shrink-0 font-mono text-xs text-muted-foreground">#{{ user.userId }}</span>
                    </button>
                  </div>
                </div>
                <Button type="submit" variant="outline" size="sm" :disabled="globalModeratorSaving">
                  <UserPlus class="size-3.5" />
                  添加
                </Button>
              </form>
            </div>
          </div>
        </template>

        <Table>
          <TableHeader>
            <TableRow>
              <TableHead class="w-20">ID</TableHead>
              <TableHead>{{ adminText('k00c2') }}</TableHead>
              <TableHead>Slug</TableHead>
              <TableHead>{{ adminText('k00ag') }}</TableHead>
              <TableHead class="w-48">版主</TableHead>
              <TableHead class="w-24">{{ adminText('k00bf') }}</TableHead>
              <TableHead class="w-32 text-right">{{ adminText('k007m') }}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-if="loading && rows.length === 0">
              <TableCell colspan="7" class="h-28 text-center text-muted-foreground">{{ adminText('k0046') }}</TableCell>
            </TableRow>
            <TableRow v-else-if="error">
              <TableCell colspan="7" class="h-28 text-center text-destructive">{{ error }}</TableCell>
            </TableRow>
            <TableRow v-else-if="filteredRows.length === 0">
              <TableCell colspan="7" class="h-28 text-center text-muted-foreground">{{ adminText('k00c3') }}</TableCell>
            </TableRow>
            <template v-else>
              <TableRow v-for="item in filteredRows" :key="item.id">
                <TableCell class="font-mono text-xs text-muted-foreground">{{ item.id }}</TableCell>
                <TableCell>
                  <div class="flex items-center gap-2">
                    <span v-if="item.icon" class="text-base" :style="{ color: item.color || undefined }">{{ item.icon }}</span>
                    <span v-else class="size-2.5 rounded-[3px]" :style="{ backgroundColor: item.color || '#64748b' }" />
                    <span class="font-medium">{{ item.category }}</span>
                  </div>
                </TableCell>
                <TableCell class="text-muted-foreground">{{ item.slug || '-' }}</TableCell>
                <TableCell class="max-w-lg truncate text-muted-foreground">{{ item.desc || '-' }}</TableCell>
                <TableCell>
                  <button
                    class="inline-flex max-w-full items-center gap-1.5 rounded-md border bg-background px-2 py-1 text-xs text-muted-foreground transition-colors hover:border-primary/40 hover:text-primary"
                    type="button"
                    @click="openModerators(item)"
                  >
                    <ShieldCheck class="size-3.5" />
                    <span v-if="item.moderators?.length" class="flex min-w-0 items-center">
                      <span class="mr-1 flex -space-x-1">
                        <span
                          v-for="moderator in visibleModerators(item)"
                          :key="moderator.id"
                          class="grid size-5 place-items-center rounded-full border bg-muted text-[10px] font-semibold text-muted-foreground"
                          :title="moderator.username || `#${moderator.userId}`"
                        >
                          {{ moderatorInitial(moderator.username) }}
                        </span>
                      </span>
                      <span class="truncate">{{ item.moderators.length }} 人</span>
                    </span>
                    <span v-else>设置版主</span>
                  </button>
                </TableCell>
                <TableCell>{{ item.sort ?? 0 }}</TableCell>
                <TableCell>
                  <div class="flex justify-end gap-2">
                    <AdminActionButton @click="openEdit(item)">
                      <Pencil class="size-3.5" />
                      {{ adminText('k005j') }}
                    </AdminActionButton>
                    <AdminActionButton tone="danger" @click="deletingRow = item">
                      <Trash2 class="size-3.5" />
                      {{ adminText('k005i') }}
                    </AdminActionButton>
                  </div>
                </TableCell>
              </TableRow>
            </template>
          </TableBody>
        </Table>
      </AdminSection>

      <Dialog :open="dialogMode !== null" @update:open="(open) => !open && (dialogMode = null)">
        <DialogContent class="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>{{ dialogMode === 'edit' ? adminText('k005p') : adminText('k005m') }}</DialogTitle>
            <DialogDescription>
              {{ adminText('k00c4') }}
            </DialogDescription>
          </DialogHeader>
          <form class="grid gap-4" @submit.prevent="submitCategory">
            <label class="grid gap-2 text-sm font-medium">
              {{ adminText('k00c2') }}
              <Input v-model="form.category" :placeholder="adminText('k005r')" />
            </label>
            <div class="grid gap-4 sm:grid-cols-2">
              <label class="grid gap-2 text-sm font-medium">
                Slug
                <Input v-model="form.slug" placeholder="category-slug" />
              </label>
              <label class="grid gap-2 text-sm font-medium">
                {{ adminText('k00c5') }}
                <Input v-model="form.icon" :placeholder="adminText('k005s')" />
              </label>
            </div>
            <label class="grid gap-2 text-sm font-medium">
              {{ adminText('k00ag') }}
              <Input v-model="form.desc" :placeholder="adminText('k005t')" />
            </label>
            <label class="grid gap-2 text-sm font-medium">
              {{ adminText('k00bf') }}
              <Input v-model.number="form.sort" type="number" />
            </label>
            <div class="grid gap-2 text-sm font-medium">
              {{ adminText('k00ad') }}
              <div class="flex flex-wrap items-center gap-2">
                <button
                  v-for="color in presetColors"
                  :key="color"
                  class="size-7 rounded-full border transition-transform hover:scale-110"
                  :class="form.color === color ? 'ring-2 ring-primary ring-offset-2' : ''"
                  :style="{ backgroundColor: color }"
                  type="button"
                  @click="form.color = color"
                />
                <Input v-model="form.color" class="w-32 font-mono text-xs" placeholder="#64748b" />
                <Button variant="outline" size="sm" type="button" @click="form.color = ''">{{ adminText('k00at') }}</Button>
              </div>
            </div>
            <DialogFooter>
              <Button variant="outline" type="button" @click="dialogMode = null">{{ adminText('k009q') }}</Button>
              <Button type="submit" :disabled="saving">{{ saving ? adminText('k005f') : adminText('k005g') }}</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <AdminConfirmDialog
        :open="deletingRow !== null"
        :title="adminText('k00c6')"
        :description="`${adminText('k00c7')}${deletingRow?.category || ''}${adminText('k00c8')}`"
        :loading="deleting"
        @update:open="(open) => !open && (deletingRow = null)"
        @confirm="confirmDelete"
      />

      <Dialog :open="moderatorRow !== null" @update:open="(open) => !open && (moderatorRow = null)">
        <DialogContent class="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>分类版主</DialogTitle>
            <DialogDescription>
              为「{{ moderatorRow?.category }}」添加或移除前台治理版主。版主不会获得后台权限。
            </DialogDescription>
          </DialogHeader>
          <div class="space-y-4">
            <form class="flex gap-2" @submit.prevent="addModerator">
              <div class="relative min-w-0 flex-1">
                <Input v-model="moderatorUserInput" placeholder="搜索用户名或输入用户 ID" autocomplete="off" />
                <div
                  v-if="moderatorUserInput.trim() && (moderatorCandidates.length || moderatorSearching)"
                  class="absolute left-0 right-0 top-[calc(100%+4px)] z-50 overflow-hidden rounded-md border bg-popover shadow-md"
                >
                  <div v-if="moderatorSearching" class="px-3 py-2 text-sm text-muted-foreground">搜索中...</div>
                  <button
                    v-for="user in moderatorCandidates"
                    v-else
                    :key="user.userId"
                    class="flex w-full items-center gap-2 px-3 py-2 text-left text-sm transition-colors"
                    :class="isAlreadyModerator(user.userId) ? 'cursor-default opacity-55' : 'hover:bg-muted'"
                    type="button"
                    :disabled="isAlreadyModerator(user.userId)"
                    @click="selectModeratorCandidate(user)"
                  >
                    <img v-if="user.avatarUrl" :src="user.avatarUrl" class="size-7 rounded-full object-cover ring-1 ring-border" alt="" />
                    <span v-else class="flex size-7 items-center justify-center rounded-full bg-muted text-xs font-semibold">{{ moderatorInitial(user.username) }}</span>
                    <span class="min-w-0 flex-1 truncate">{{ user.username }}</span>
                    <span v-if="isAlreadyModerator(user.userId)" class="shrink-0 rounded-full bg-muted px-1.5 py-0.5 text-[11px] text-muted-foreground">已添加</span>
                    <span class="shrink-0 font-mono text-xs text-muted-foreground">#{{ user.userId }}</span>
                  </button>
                </div>
              </div>
              <Button type="submit" :disabled="moderatorSaving">
                <UserPlus class="size-4" />
                添加
              </Button>
            </form>
            <div class="overflow-hidden rounded-lg border">
              <div class="flex items-center justify-between border-b bg-muted/20 px-3 py-2 text-xs text-muted-foreground">
                <span>当前版主</span>
                <span>{{ moderatorRow?.moderators?.length || 0 }} 人</span>
              </div>
              <div v-if="!moderatorRow?.moderators?.length" class="px-4 py-8 text-center text-sm text-muted-foreground">
                暂无版主
              </div>
              <div v-else class="max-h-72 divide-y overflow-y-auto">
                <div
                  v-for="moderator in moderatorRow.moderators"
                  :key="moderator.id"
                  class="flex items-center justify-between gap-3 px-3 py-2"
                >
                  <div class="flex min-w-0 items-center gap-2">
                    <img v-if="moderator.avatarUrl" :src="moderator.avatarUrl" class="size-8 rounded-full object-cover ring-1 ring-border" alt="" />
                    <span v-else class="flex size-8 items-center justify-center rounded-full bg-muted text-xs font-semibold">{{ moderatorInitial(moderator.username) }}</span>
                    <div class="min-w-0">
                      <div class="truncate text-sm font-medium">{{ moderator.username || `#${moderator.userId}` }}</div>
                      <div class="text-xs text-muted-foreground">ID {{ moderator.userId }}</div>
                    </div>
                  </div>
                  <Button variant="ghost" size="icon-sm" type="button" :disabled="moderatorSaving" @click="removeModerator(moderator.id)">
                    <X class="size-4" />
                  </Button>
                </div>
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" type="button" @click="moderatorRow = null">{{ adminText('k009q') }}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </BasicPage>
</template>
