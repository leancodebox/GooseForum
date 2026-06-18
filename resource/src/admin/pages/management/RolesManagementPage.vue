<script setup lang="ts">
import { adminText } from '@/admin/runtime/i18n-text'

import { computed, onMounted, reactive, ref } from 'vue'
import { Pencil, Plus, RefreshCw, Search, ShieldCheck, Trash2 } from '@lucide/vue'
import AdminActionButton from '@/admin/components/AdminActionButton.vue'
import AdminConfirmDialog from '@/admin/components/AdminConfirmDialog.vue'
import AdminToolbar from '@/admin/components/AdminToolbar.vue'
import { BasicPage } from '@/admin/components/global-layout'
import { Button } from '@/admin/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/admin/components/ui/dialog'
import { deleteRole, getPermissionList, getRoleList, saveRole } from '@/admin/runtime/api'
import { adminToast } from '@/admin/runtime/toast'
import type { AdminPayload, AdminRole, ManageHomeProps } from '@/admin/types'
import ManagementTable from './ManagementTable.vue'

defineProps<{
  payload: AdminPayload<ManageHomeProps>
}>()

const loading = ref(false)
const saving = ref(false)
const deleting = ref(false)
const error = ref('')
const rows = ref<AdminRole[]>([])
const permissionOptions = ref<{ id: number, name: string }[]>([])
const page = ref(1)
const pageSize = ref(10)
const search = ref('')
const appliedSearch = ref('')
const effectiveFilter = ref('')
const dialogMode = ref<'add' | 'edit' | null>(null)
const editingRole = ref<AdminRole | null>(null)
const deletingRole = ref<AdminRole | null>(null)
const form = reactive({ roleName: '', permissions: [] as number[] })

const filteredRows = computed(() => {
  const keyword = appliedSearch.value.toLowerCase()
  return rows.value.filter((role) => {
    const matchesName = !keyword || role.roleName.toLowerCase().includes(keyword)
    const matchesEffective = effectiveFilter.value === '' || String(role.effective) === effectiveFilter.value
    return matchesName && matchesEffective
  })
})

const total = computed(() => filteredRows.value.length)
const pagedRows = computed(() => {
  const start = (page.value - 1) * pageSize.value
  return filteredRows.value.slice(start, start + pageSize.value)
})

async function loadRoles() {
  loading.value = true
  error.value = ''
  try {
    const [data, permissions] = await Promise.all([getRoleList(), getPermissionList()])
    rows.value = data.list || []
    permissionOptions.value = permissions.map(item => ({
      id: Number(item.value),
      name: item.label || item.name,
    }))
  } catch (err) {
    error.value = err instanceof Error ? err.message : adminText('k002r')
  } finally {
    loading.value = false
  }
}

function applySearch() {
  appliedSearch.value = search.value.trim()
  page.value = 1
}

function changePageSize(value: number) {
  pageSize.value = value
  page.value = 1
}

function resetForm(role?: AdminRole) {
  editingRole.value = role || null
  form.roleName = role?.roleName || ''
  form.permissions = role?.permissions?.map(permission => permission.id) || []
}

function openAdd() {
  resetForm()
  dialogMode.value = 'add'
}

function openEdit(role: AdminRole) {
  resetForm(role)
  dialogMode.value = 'edit'
}

function togglePermission(id: number, checked: boolean) {
  if (checked && !form.permissions.includes(id)) {
    form.permissions = [...form.permissions, id]
  } else if (!checked) {
    form.permissions = form.permissions.filter(value => value !== id)
  }
}

async function submitRole() {
  if (!form.roleName.trim()) {
    adminToast.warning(adminText('k002s'))
    return
  }
  if (!form.permissions.length) {
    adminToast.warning(adminText('k002t'))
    return
  }
  saving.value = true
  try {
    await saveRole({
      id: dialogMode.value === 'edit' ? editingRole.value?.roleId || 0 : 0,
      roleName: form.roleName.trim(),
      permissions: form.permissions,
    })
    dialogMode.value = null
    await loadRoles()
    adminToast.success(adminText('k000e'))
  } catch (err) {
    adminToast.error(err, adminText('k000x'))
  } finally {
    saving.value = false
  }
}

async function confirmDelete() {
  if (!deletingRole.value) return
  deleting.value = true
  try {
    await deleteRole(deletingRole.value.roleId)
    deletingRole.value = null
    await loadRoles()
    adminToast.success(adminText('k002u'))
  } catch (err) {
    adminToast.error(err, adminText('k000y'))
  } finally {
    deleting.value = false
  }
}

onMounted(() => {
  void loadRoles()
})
</script>

<template>
  <BasicPage :title="adminText('k007f')" :description="adminText('k007g')" sticky>
    <template #actions>
      <div class="flex items-center gap-2">
        <Button variant="outline" type="button" @click="loadRoles">
          <RefreshCw class="size-4" />
          {{ adminText('k004q') }}
        </Button>
        <Button type="button" @click="openAdd">
          <Plus class="size-4" />
          {{ adminText('k007h') }}
        </Button>
      </div>
    </template>

      <ManagementTable
        :columns="['ID', adminText('k007i'), adminText('k007j'), adminText('k007k'), adminText('k007l'), adminText('k007m')]"
        :loading="loading && rows.length === 0"
        :error="error"
        :total="total"
        :page="page"
        :page-size="pageSize"
        @retry="loadRoles"
        @update:page="page = $event"
        @update:page-size="changePageSize"
      >
        <template #header>
          <AdminToolbar class="-mx-3 -my-2 border-b-0">
            <form class="flex min-w-64 flex-1 gap-2 sm:flex-none" @submit.prevent="applySearch">
              <div class="relative flex-1">
                <Search class="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                <input v-model="search" class="h-10 w-full rounded-md border bg-background pl-9 pr-3 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring" :placeholder="adminText('k007q')" />
              </div>
              <Button type="submit">{{ adminText('k00al') }}</Button>
            </form>
            <select v-model="effectiveFilter" class="h-10 rounded-md border bg-background px-3 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring" @change="page = 1">
              <option value="">{{ adminText('k00am') }}</option>
              <option value="1">{{ adminText('k007n') }}</option>
              <option value="0">{{ adminText('k007o') }}</option>
            </select>
          </AdminToolbar>
        </template>
        <tr v-if="pagedRows.length === 0">
          <td colspan="6" class="h-28 px-4 text-center text-muted-foreground">{{ adminText('k00an') }}</td>
        </tr>
        <tr v-for="role in pagedRows" :key="role.roleId" class="hover:bg-muted/35">
          <td class="px-4 py-3 font-mono text-xs text-muted-foreground">{{ role.roleId }}</td>
          <td class="px-4 py-3 font-medium">{{ role.roleName }}</td>
          <td class="px-4 py-3">
            <span class="inline-flex items-center gap-1.5 rounded-full px-2 py-1 text-xs font-medium" :class="role.effective === 1 ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground'">
              <ShieldCheck class="size-3.5" />
              {{ role.effective === 1 ? adminText('k007n') : adminText('k007o') }}
            </span>
          </td>
          <td class="px-4 py-3">
            <div class="flex max-w-lg flex-wrap gap-1.5">
              <span v-for="permission in role.permissions" :key="permission.id" class="rounded-full border px-2 py-1 text-xs">{{ permission.name }}</span>
            </div>
          </td>
          <td class="px-4 py-3 text-xs text-muted-foreground">{{ role.createTime || '-' }}</td>
          <td class="px-4 py-3">
            <div class="flex gap-2">
              <AdminActionButton @click="openEdit(role)">
                <Pencil class="size-3.5" />
                {{ adminText('k005j') }}
              </AdminActionButton>
              <AdminActionButton tone="danger" @click="deletingRole = role">
                <Trash2 class="size-3.5" />
                {{ adminText('k005i') }}
              </AdminActionButton>
            </div>
          </td>
        </tr>
      </ManagementTable>

      <Dialog :open="dialogMode !== null" @update:open="(open) => !open && (dialogMode = null)">
        <DialogContent class="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>{{ dialogMode === 'add' ? adminText('k007h') : adminText('k007p') }}</DialogTitle>
            <DialogDescription>{{ adminText('k00ao') }}</DialogDescription>
          </DialogHeader>
          <form class="grid gap-5" @submit.prevent="submitRole">
          <div class="space-y-5">
            <label class="grid gap-2 text-sm font-medium">
              {{ adminText('k007i') }}
              <input v-model="form.roleName" class="h-10 rounded-md border bg-background px-3 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring" :placeholder="adminText('k007r')" />
            </label>
            <div class="grid gap-2">
              <div class="text-sm font-medium">{{ adminText('k00ap') }}</div>
              <div class="grid max-h-56 grid-cols-2 gap-3 overflow-y-auto rounded-md border p-4">
                <label v-for="permission in permissionOptions" :key="permission.id" class="flex items-center gap-2 text-sm">
                  <input
                    class="size-4 rounded border"
                    type="checkbox"
                    :checked="form.permissions.includes(permission.id)"
                    @change="togglePermission(permission.id, ($event.target as HTMLInputElement).checked)"
                  />
                  {{ permission.name }}
                </label>
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" type="button" @click="dialogMode = null">{{ adminText('k009q') }}</Button>
            <Button type="submit" :disabled="saving">
              {{ saving ? adminText('k005f') : adminText('k005g') }}
            </Button>
          </DialogFooter>
        </form>
        </DialogContent>
      </Dialog>

      <AdminConfirmDialog
        :open="deletingRole !== null"
        :title="adminText('k00aq')"
        :description="`${adminText('k00ar')} ${deletingRole?.roleName || ''}${adminText('k00as')}`"
        :loading="deleting"
        @update:open="(open) => !open && (deletingRole = null)"
        @confirm="confirmDelete"
      />
    </BasicPage>
</template>
