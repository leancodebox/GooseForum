<script setup lang="ts">import { adminText } from '@/admin/runtime/i18n-text'

import { computed, onMounted, reactive, ref } from 'vue'
import { Pencil, Plus, RefreshCw, Search, ShieldCheck, Trash2, XCircle } from '@lucide/vue'
import { BasicPage } from '@/admin/components/global-layout'
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
        <button class="inline-flex items-center gap-2 rounded-md border bg-background px-3 py-2 text-sm font-medium shadow-sm hover:bg-muted" type="button" @click="loadRoles">
          <RefreshCw class="size-4" />
          {{ adminText('k004q') }}
        </button>
        <button class="inline-flex items-center gap-2 rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground shadow-sm hover:bg-primary/90" type="button" @click="openAdd">
          <Plus class="size-4" />
          {{ adminText('k007h') }}
        </button>
      </div>
    </template>

      <div class="mb-4 flex flex-wrap items-center gap-2">
        <form class="flex min-w-64 flex-1 gap-2 sm:flex-none" @submit.prevent="applySearch">
          <div class="relative flex-1">
            <Search class="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
            <input v-model="search" class="h-10 w-full rounded-md border bg-background pl-9 pr-3 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring" :placeholder="adminText('k007q')" />
          </div>
          <button class="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground" type="submit">{{ adminText('k00al') }}</button>
        </form>
        <select v-model="effectiveFilter" class="h-10 rounded-md border bg-background px-3 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring" @change="page = 1">
          <option value="">{{ adminText('k00am') }}</option>
          <option value="1">{{ adminText('k007n') }}</option>
          <option value="0">{{ adminText('k007o') }}</option>
        </select>
      </div>

      <ManagementTable
        :columns="['ID', adminText('k007i'), adminText('k007j'), adminText('k007k'), adminText('k007l'), adminText('k007m')]"
        :loading="loading"
        :error="error"
        :total="total"
        :page="page"
        :page-size="pageSize"
        @retry="loadRoles"
        @update:page="page = $event"
        @update:page-size="changePageSize"
      >
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
              <button class="inline-flex items-center gap-1.5 rounded-md border bg-background px-2.5 py-1.5 text-xs font-medium hover:bg-muted" type="button" @click="openEdit(role)">
                <Pencil class="size-3.5" />
                {{ adminText('k005j') }}
              </button>
              <button class="inline-flex items-center gap-1.5 rounded-md border border-destructive/30 bg-background px-2.5 py-1.5 text-xs font-medium text-destructive hover:bg-destructive/5" type="button" @click="deletingRole = role">
                <Trash2 class="size-3.5" />
                {{ adminText('k005i') }}
              </button>
            </div>
          </td>
        </tr>
      </ManagementTable>

      <div v-if="dialogMode" class="fixed inset-0 z-50 flex items-center justify-center bg-black/45 p-4" @click.self="dialogMode = null">
        <form class="w-full max-w-lg rounded-lg border bg-background p-6 shadow-xl" @submit.prevent="submitRole">
          <div class="mb-5 flex items-start justify-between gap-4">
            <div>
              <h2 class="text-lg font-semibold">{{ dialogMode === 'add' ? adminText('k007h') : adminText('k007p') }}</h2>
              <p class="mt-1 text-sm text-muted-foreground">{{ adminText('k00ao') }}</p>
            </div>
            <button class="rounded-md p-1 text-muted-foreground hover:bg-muted" type="button" @click="dialogMode = null">
              <XCircle class="size-5" />
            </button>
          </div>
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
          <div class="mt-6 flex justify-end gap-2">
            <button class="rounded-md border bg-background px-4 py-2 text-sm font-medium hover:bg-muted" type="button" @click="dialogMode = null">{{ adminText('k009q') }}</button>
            <button class="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground disabled:opacity-60" type="submit" :disabled="saving">
              {{ saving ? adminText('k005f') : adminText('k005g') }}
            </button>
          </div>
        </form>
      </div>

      <div v-if="deletingRole" class="fixed inset-0 z-50 flex items-center justify-center bg-black/45 p-4" @click.self="deletingRole = null">
        <div class="w-full max-w-md rounded-lg border bg-background p-6 shadow-xl">
          <h2 class="text-lg font-semibold">{{ adminText('k00aq') }}</h2>
          <p class="mt-2 text-sm text-muted-foreground">{{ adminText('k00ar') }} <span class="font-semibold text-foreground">{{ deletingRole.roleName }}</span>{{ adminText('k00as') }}</p>
          <div class="mt-6 flex justify-end gap-2">
            <button class="rounded-md border bg-background px-4 py-2 text-sm font-medium hover:bg-muted" type="button" @click="deletingRole = null">{{ adminText('k009q') }}</button>
            <button class="rounded-md bg-destructive px-4 py-2 text-sm font-medium text-destructive-foreground disabled:opacity-60" type="button" :disabled="deleting" @click="confirmDelete">
              {{ deleting ? adminText('k005h') : adminText('k005i') }}
            </button>
          </div>
        </div>
      </div>
    </BasicPage>
</template>
