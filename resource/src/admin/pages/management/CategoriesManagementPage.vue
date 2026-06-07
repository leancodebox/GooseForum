<script setup lang="ts">import { adminText } from '@/admin/runtime/i18n-text'

import { computed, onMounted, reactive, ref } from 'vue'
import { Pencil, Plus, RefreshCw, Search, Trash2 } from '@lucide/vue'
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
import { deleteCategory, getCategoryList, saveCategory } from '@/admin/runtime/api'
import { adminToast } from '@/admin/runtime/toast'
import type { AdminCategory, AdminPayload, ManageHomeProps } from '@/admin/types'

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

onMounted(() => {
  void loadCategories()
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

      <div class="mb-4 flex items-center gap-2">
        <div class="relative w-full max-w-md">
          <Search class="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input v-model="search" class="pl-9" :placeholder="adminText('k005q')" />
        </div>
        <Badge variant="secondary" class="h-9 rounded-md px-3">
          {{ filteredRows.length }} {{ adminText('k00c1') }}
        </Badge>
      </div>

      <div class="overflow-hidden rounded-lg border bg-card">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead class="w-20">ID</TableHead>
              <TableHead>{{ adminText('k00c2') }}</TableHead>
              <TableHead>Slug</TableHead>
              <TableHead>{{ adminText('k00ag') }}</TableHead>
              <TableHead class="w-24">{{ adminText('k00bf') }}</TableHead>
              <TableHead class="w-32 text-right">{{ adminText('k007m') }}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-if="loading">
              <TableCell colspan="6" class="h-28 text-center text-muted-foreground">{{ adminText('k0046') }}</TableCell>
            </TableRow>
            <TableRow v-else-if="error">
              <TableCell colspan="6" class="h-28 text-center text-destructive">{{ error }}</TableCell>
            </TableRow>
            <TableRow v-else-if="filteredRows.length === 0">
              <TableCell colspan="6" class="h-28 text-center text-muted-foreground">{{ adminText('k00c3') }}</TableCell>
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
                <TableCell>{{ item.sort ?? 0 }}</TableCell>
                <TableCell>
                  <div class="flex justify-end gap-2">
                    <Button variant="outline" size="sm" type="button" @click="openEdit(item)">
                      <Pencil class="size-3.5" />
                      {{ adminText('k005j') }}
                    </Button>
                    <Button variant="destructive" size="sm" type="button" @click="deletingRow = item">
                      <Trash2 class="size-3.5" />
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
            </template>
          </TableBody>
        </Table>
      </div>

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

      <Dialog :open="deletingRow !== null" @update:open="(open) => !open && (deletingRow = null)">
        <DialogContent class="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{{ adminText('k00c6') }}</DialogTitle>
            <DialogDescription>
              {{ adminText('k00c7') }}{{ deletingRow?.category }}{{ adminText('k00c8') }}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" type="button" @click="deletingRow = null">{{ adminText('k009q') }}</Button>
            <Button variant="destructive" type="button" :disabled="deleting" @click="confirmDelete">
              {{ deleting ? adminText('k005h') : adminText('k005i') }}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </BasicPage>
</template>
