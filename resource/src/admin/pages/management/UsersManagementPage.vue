<script setup lang="ts">
import { adminText } from '@/admin/runtime/i18n-text'

import { computed, onMounted, reactive, ref } from 'vue'
import { Award, ChevronLeft, ChevronRight, CheckCircle2, Loader2, RefreshCw, Search, ShieldOff, UserCog } from '@lucide/vue'
import AdminSection from '@/admin/components/AdminSection.vue'
import AdminToolbar from '@/admin/components/AdminToolbar.vue'
import { BasicPage } from '@/admin/components/global-layout'
import { Button } from '@/admin/components/ui/button'
import { Dialog, DialogContent } from '@/admin/components/ui/dialog'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/admin/components/ui/select'
import { Switch } from '@/admin/components/ui/switch'
import { editUser, getAllRoleItem, getUserBadgeOptions, getUserList, saveUserBadges } from '@/admin/runtime/api'
import { adminToast } from '@/admin/runtime/toast'
import type { AdminBadge, AdminPayload, AdminUser, ManageHomeProps, UserBadge } from '@/admin/types'

defineProps<{
  payload: AdminPayload<ManageHomeProps>
}>()

const loading = ref(false)
const saving = ref(false)
const error = ref('')
const rows = ref<AdminUser[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const search = ref('')
const appliedSearch = ref('')
const editingUser = ref<AdminUser | null>(null)
const roles = ref<{ name: string, value: number }[]>([])
const form = reactive({ status: 0, validate: 0, roleId: 0 })
const badgeLoading = ref(false)
const badgesLoaded = ref(false)
const badgeOptions = ref<AdminBadge[]>([])
const activeBadges = ref<UserBadge[]>([])
const selectedBadgeCodes = ref<string[]>([])

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))
const rangeStart = computed(() => (total.value === 0 ? 0 : (page.value - 1) * pageSize.value + 1))
const rangeEnd = computed(() => Math.min(page.value * pageSize.value, total.value))

async function loadUsers() {
  loading.value = true
  error.value = ''
  try {
    const data = await getUserList({
      page: page.value,
      pageSize: pageSize.value,
      username: appliedSearch.value || undefined,
    })
    rows.value = data.list || []
    total.value = data.total || 0
    if (page.value > totalPages.value) page.value = totalPages.value
  } catch (err) {
    error.value = err instanceof Error ? err.message : adminText('k0031')
  } finally {
    loading.value = false
  }
}

async function ensureRoles() {
  if (roles.value.length) return
  try {
    roles.value = [{ name: adminText('k0032'), value: 0 }, ...(await getAllRoleItem())]
  } catch {
    roles.value = [{ name: adminText('k0032'), value: 0 }]
  }
}

function applySearch() {
  appliedSearch.value = search.value.trim()
  page.value = 1
  void loadUsers()
}

function changePage(nextPage: number) {
  page.value = nextPage
  void loadUsers()
}

function changePageSize(nextPageSize: number) {
  pageSize.value = nextPageSize
  page.value = 1
  void loadUsers()
}

function roleNames(user: AdminUser) {
  return user.roleList?.map(role => role.name).filter(Boolean) || []
}

function badgeIconURL(badge: AdminBadge | UserBadge) {
  return badge.iconUrl || '/static/badges/contributor.svg'
}

function badgeToneClass(badge: AdminBadge | UserBadge) {
  const classes: Record<string, string> = {
    blue: 'bg-blue-100 text-blue-700 ring-blue-200',
    emerald: 'bg-emerald-100 text-emerald-700 ring-emerald-200',
    teal: 'bg-teal-100 text-teal-700 ring-teal-200',
    sky: 'bg-sky-100 text-sky-700 ring-sky-200',
    cyan: 'bg-cyan-100 text-cyan-700 ring-cyan-200',
    rose: 'bg-rose-100 text-rose-700 ring-rose-200',
    violet: 'bg-violet-100 text-violet-700 ring-violet-200',
    purple: 'bg-purple-100 text-purple-700 ring-purple-200',
    fuchsia: 'bg-fuchsia-100 text-fuchsia-700 ring-fuchsia-200',
    indigo: 'bg-indigo-100 text-indigo-700 ring-indigo-200',
    amber: 'bg-amber-100 text-amber-700 ring-amber-200',
    orange: 'bg-orange-100 text-orange-700 ring-orange-200',
    yellow: 'bg-yellow-100 text-yellow-700 ring-yellow-200',
    slate: 'bg-slate-100 text-slate-700 ring-slate-200',
  }
  if (badge.color && classes[badge.color]) return classes[badge.color]
  if (badge.level === 'gold') return classes.amber
  if (badge.level === 'special') return classes.indigo
  return classes.blue
}

function autoBadges() {
  return activeBadges.value.filter(badge => badge.source !== 'manual')
}

function toggleBadge(code: string) {
  if (selectedBadgeCodes.value.includes(code)) {
    selectedBadgeCodes.value = selectedBadgeCodes.value.filter(item => item !== code)
    return
  }
  selectedBadgeCodes.value = [...selectedBadgeCodes.value, code]
}

async function loadUserBadges(userId: number) {
  badgeLoading.value = true
  badgesLoaded.value = false
  badgeOptions.value = []
  activeBadges.value = []
  selectedBadgeCodes.value = []
  try {
    const data = await getUserBadgeOptions(userId)
    const options = data.options || []
    const active = data.active || []
    const optionCodes = new Set(options.map(item => item.code))
    badgeOptions.value = options
    activeBadges.value = active
    selectedBadgeCodes.value = active
      .filter(item => item.source === 'manual' && optionCodes.has(item.code))
      .map(item => item.code)
    badgesLoaded.value = true
  } catch (err) {
    adminToast.error(err, adminText('k0033'))
  } finally {
    badgeLoading.value = false
  }
}

async function openEdit(user: AdminUser) {
  editingUser.value = user
  form.status = user.status
  form.validate = user.validate
  form.roleId = user.roleId || 0
  await Promise.all([
    ensureRoles(),
    loadUserBadges(user.userId),
  ])
}

async function saveUser() {
  if (!editingUser.value) return
  saving.value = true
  try {
    await editUser({
      userId: editingUser.value.userId,
      status: Number(form.status),
      validate: Number(form.validate),
      roleId: Number(form.roleId),
    })
    if (badgesLoaded.value) {
      await saveUserBadges(editingUser.value.userId, selectedBadgeCodes.value)
    }
    editingUser.value = null
    await loadUsers()
    adminToast.success(adminText('k000e'))
  } catch (err) {
    adminToast.error(err, adminText('k000r'))
  } finally {
    saving.value = false
  }
}

function avatarText(user: AdminUser) {
  return user.username.slice(0, 2).toUpperCase()
}

function updateRoleId(value: unknown) {
  if (value === null || value === undefined) return
  form.roleId = Number(value)
}

onMounted(() => {
  void loadUsers()
})
</script>

<template>
  <BasicPage :title="adminText('k006i')" :description="adminText('k006j')" sticky>
    <template #actions>
      <Button variant="outline" type="button" @click="loadUsers">
        <RefreshCw class="size-4" />
        {{ adminText('k004q') }}
      </Button>
    </template>

      <AdminSection>
        <template #header>
        <AdminToolbar class="-mx-3 -my-2 border-b-0">
          <form class="flex min-w-0 flex-1 items-center gap-2" @submit.prevent="applySearch">
            <div class="relative min-w-0 flex-1 sm:max-w-md">
              <Search class="pointer-events-none absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
              <input
                v-model="search"
                class="h-9 w-full rounded-md border bg-background pl-8 pr-3 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
                :placeholder="adminText('k006t')"
              />
            </div>
            <Button size="sm" type="submit" class="h-9 px-4">{{ adminText('k00al') }}</Button>
            <Button v-if="appliedSearch" variant="ghost" size="sm" type="button" class="h-9" @click="search = ''; applySearch()">
              {{ adminText('k00at') }}
            </Button>
          </form>
          <div class="flex flex-wrap items-center gap-2 text-sm text-muted-foreground">
            <span class="whitespace-nowrap">{{ rangeStart }}-{{ rangeEnd }} / {{ total }}</span>
            <select
              class="h-9 rounded-md border bg-background px-2 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
              :value="pageSize"
              @change="changePageSize(Number(($event.target as HTMLSelectElement).value))"
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
              :disabled="page <= 1 || loading"
              @click="changePage(page - 1)"
            >
              <ChevronLeft class="size-4" />
            </Button>
            <span class="min-w-14 text-center">{{ page }} / {{ totalPages }}</span>
            <Button
              variant="outline"
              size="icon"
              type="button"
              :disabled="page >= totalPages || loading"
              @click="changePage(page + 1)"
            >
              <ChevronRight class="size-4" />
            </Button>
          </div>
        </AdminToolbar>
        </template>

        <div class="md:hidden">
          <div v-if="loading" class="px-3 py-10 text-center text-sm text-muted-foreground">{{ adminText('k0046') }}</div>
          <div v-else-if="error" class="px-3 py-10 text-center">
            <div class="text-sm text-destructive">{{ error }}</div>
            <Button variant="link" size="sm" class="mt-2 h-auto px-0 text-destructive" type="button" @click="loadUsers">{{ adminText('k002w') }}</Button>
          </div>
          <div v-else-if="rows.length === 0" class="px-3 py-10 text-center text-sm text-muted-foreground">{{ adminText('k00bm') }}</div>
          <div v-else class="divide-y">
            <article v-for="user in rows" :key="user.userId" class="px-3 py-3">
              <div class="flex min-w-0 items-start gap-3">
                <a :href="`/u/${user.userId}`" target="_blank" rel="noreferrer" class="shrink-0">
                  <img v-if="user.avatarUrl" :src="user.avatarUrl" class="size-10 rounded-full object-cover ring-1 ring-border" alt="" />
                  <span v-else class="flex size-10 items-center justify-center rounded-full bg-muted text-sm font-semibold">{{ avatarText(user) }}</span>
                </a>
                <div class="min-w-0 flex-1">
                  <div class="flex min-w-0 items-center justify-between gap-2">
                    <a :href="`/u/${user.userId}`" target="_blank" rel="noreferrer" class="truncate font-semibold hover:text-primary hover:underline">{{ user.username }}</a>
                    <Button variant="ghost" size="icon-sm" type="button" :title="adminText('k005j')" @click="openEdit(user)">
                      <UserCog class="size-4" />
                    </Button>
                  </div>
                  <div class="mt-0.5 truncate text-xs text-muted-foreground">{{ user.email || '-' }}</div>
                  <div class="mt-2 flex flex-wrap items-center gap-x-2 gap-y-1 text-xs text-muted-foreground">
                    <span class="inline-flex items-center gap-1" :class="user.status === 0 ? 'text-emerald-700' : 'text-destructive'">
                      <span class="size-1.5 rounded-full" :class="user.status === 0 ? 'bg-emerald-500' : 'bg-destructive'" />
                      {{ user.status === 0 ? adminText('k005y') : adminText('k006k') }}
                    </span>
                    <span class="inline-flex items-center gap-1" :class="user.validate === 1 ? 'text-emerald-700' : ''">
                      <span class="size-1.5 rounded-full" :class="user.validate === 1 ? 'bg-emerald-500' : 'bg-muted-foreground/40'" />
                      {{ user.validate === 1 ? adminText('k006l') : adminText('k006m') }}
                    </span>
                    <span>{{ roleNames(user).join(' / ') || adminText('k006n') }}</span>
                  </div>
                </div>
              </div>
            </article>
          </div>
        </div>

        <div class="hidden overflow-x-auto md:block">
          <table class="w-full min-w-[880px] table-fixed text-sm">
            <thead class="border-b bg-muted/25 text-xs font-medium text-muted-foreground">
              <tr>
                <th class="h-10 px-3 text-left align-middle">{{ adminText('k003f') }}</th>
                <th class="w-[220px] px-3 text-left align-middle">{{ adminText('k00bn') }}</th>
                <th class="w-[150px] px-3 text-left align-middle">{{ adminText('k007j') }}</th>
                <th class="w-[150px] px-3 text-left align-middle">{{ adminText('k00bo') }}</th>
                <th class="w-[150px] px-3 text-left align-middle">{{ adminText('k00bp') }}</th>
                <th class="w-[72px] px-3 text-right align-middle">{{ adminText('k007m') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y">
              <tr v-if="loading">
                <td colspan="6" class="h-28 px-3 text-center text-muted-foreground">{{ adminText('k0046') }}</td>
              </tr>
              <tr v-else-if="error">
                <td colspan="6" class="h-28 px-3 text-center">
                  <div class="inline-flex items-center gap-3 rounded-md border border-destructive/30 bg-destructive/5 px-4 py-2 text-destructive">
                    <span>{{ error }}</span>
                    <Button variant="link" size="sm" class="h-auto px-0 text-destructive" type="button" @click="loadUsers">{{ adminText('k002w') }}</Button>
                  </div>
                </td>
              </tr>
              <tr v-else-if="rows.length === 0">
                <td colspan="6" class="h-28 px-3 text-center text-muted-foreground">{{ adminText('k00bm') }}</td>
              </tr>
              <tr v-for="user in rows" v-else :key="user.userId" class="hover:bg-muted/25">
                <td class="max-w-0 px-3 py-2">
                  <div class="flex min-w-0 items-center gap-2.5">
                    <a :href="`/u/${user.userId}`" target="_blank" rel="noreferrer" class="shrink-0">
                      <img v-if="user.avatarUrl" :src="user.avatarUrl" class="size-9 rounded-full object-cover ring-1 ring-border" alt="" />
                      <span v-else class="flex size-9 items-center justify-center rounded-full bg-muted text-xs font-semibold">{{ avatarText(user) }}</span>
                    </a>
                    <div class="min-w-0">
                      <a :href="`/u/${user.userId}`" target="_blank" rel="noreferrer" class="block truncate font-semibold leading-5 hover:text-primary hover:underline">{{ user.username }}</a>
                      <div class="mt-0.5 flex min-w-0 items-center gap-2 text-xs text-muted-foreground">
                        <span class="truncate">{{ user.email || '-' }}</span>
                        <span class="inline-flex shrink-0 items-center gap-1" :class="user.validate === 1 ? 'text-emerald-700' : ''">
                          <span class="size-1.5 rounded-full" :class="user.validate === 1 ? 'bg-emerald-500' : 'bg-muted-foreground/40'" />
                          {{ user.validate === 1 ? adminText('k006l') : adminText('k006m') }}
                        </span>
                      </div>
                    </div>
                  </div>
                </td>
                <td class="px-3 py-2">
                  <div class="flex flex-wrap gap-1">
                    <span v-for="role in user.roleList" :key="role.value" class="rounded-md bg-muted px-1.5 py-0.5 text-xs text-muted-foreground">{{ role.name }}</span>
                    <span v-if="!user.roleList?.length" class="text-xs text-muted-foreground">{{ adminText('k0032') }}</span>
                  </div>
                </td>
                <td class="px-3 py-2">
                  <span class="inline-flex items-center gap-1.5 text-xs font-medium" :class="user.status === 0 ? 'text-emerald-700' : 'text-destructive'">
                    <CheckCircle2 v-if="user.status === 0" class="size-3.5" />
                    <ShieldOff v-else class="size-3.5" />
                    {{ user.status === 0 ? adminText('k005y') : adminText('k006k') }}
                  </span>
                </td>
                <td class="px-3 py-2 text-xs text-muted-foreground">{{ user.createTime || '-' }}</td>
                <td class="px-3 py-2 text-xs text-muted-foreground">{{ user.lastActiveTime || adminText('k006o') }}</td>
                <td class="px-3 py-2 text-right">
                  <Button variant="ghost" size="icon-sm" type="button" :title="adminText('k005j')" @click="openEdit(user)">
                    <UserCog class="size-4" />
                  </Button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </AdminSection>

      <Dialog :open="editingUser !== null" @update:open="(open) => !open && (editingUser = null)">
        <DialogContent class="flex max-h-[min(700px,calc(100vh-1.5rem))] gap-0 overflow-hidden p-0 sm:max-w-4xl">
        <form v-if="editingUser" class="flex min-h-0 w-full flex-col" @submit.prevent="saveUser">
          <div class="flex items-center justify-between gap-4 border-b px-4 py-3">
            <div class="flex min-w-0 items-center gap-3">
              <img v-if="editingUser.avatarUrl" :src="editingUser.avatarUrl" class="size-10 rounded-full object-cover ring-1 ring-border" alt="" />
              <span v-else class="flex size-10 items-center justify-center rounded-full bg-muted text-sm font-semibold">{{ avatarText(editingUser) }}</span>
              <div class="min-w-0">
                <h2 class="truncate text-base font-semibold">{{ adminText('k00bq') }}</h2>
                <p class="truncate text-xs text-muted-foreground">{{ editingUser.username }} · {{ editingUser.email || adminText('k006p') }}</p>
              </div>
            </div>
          </div>

          <div class="min-h-0 flex-1 overflow-y-auto p-4">
            <div class="grid gap-3 lg:grid-cols-[280px_minmax(0,1fr)]">
              <aside class="space-y-3">
                <section class="rounded-lg border bg-background p-3">
                  <div class="mb-2 text-sm font-semibold">{{ adminText('k00br') }}</div>
                  <div class="divide-y">
                    <label class="flex items-center justify-between gap-3 py-2.5">
                      <span class="text-sm font-medium">{{ adminText('k00br') }}</span>
                      <span class="inline-flex items-center gap-2">
                        <span class="text-xs text-muted-foreground">{{ form.status === 0 ? adminText('k005y') : adminText('k006k') }}</span>
                        <Switch :model-value="form.status === 0" @update:model-value="(checked: boolean) => form.status = checked ? 0 : 1" />
                      </span>
                    </label>
                    <label class="flex items-center justify-between gap-3 py-2.5">
                      <span class="text-sm font-medium">{{ adminText('k00bs') }}</span>
                      <span class="inline-flex items-center gap-2">
                        <span class="text-xs text-muted-foreground">{{ form.validate === 1 ? adminText('k006l') : adminText('k006m') }}</span>
                        <Switch :model-value="form.validate === 1" @update:model-value="(checked: boolean) => form.validate = checked ? 1 : 0" />
                      </span>
                    </label>
                    <div class="grid gap-1.5 py-2.5 text-sm font-medium">
                      {{ adminText('k00bt') }}
                      <Select :model-value="String(form.roleId)" @update:model-value="updateRoleId">
                        <SelectTrigger class="h-9 w-full">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem v-for="role in roles" :key="role.value" :value="String(role.value)">
                            {{ role.name }}
                          </SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                  </div>
                </section>

                <section class="rounded-lg border bg-background p-3">
                  <div class="mb-2 text-sm font-semibold">{{ adminText('k00bu') }}</div>
                  <dl class="grid gap-2 text-sm">
                    <div class="flex justify-between gap-3">
                      <dt class="text-muted-foreground">{{ adminText('k00bv') }}</dt>
                      <dd class="truncate text-right">{{ editingUser.createTime || '-' }}</dd>
                    </div>
                    <div class="flex justify-between gap-3">
                      <dt class="text-muted-foreground">{{ adminText('k00bp') }}</dt>
                      <dd class="truncate text-right">{{ editingUser.lastActiveTime || adminText('k006o') }}</dd>
                    </div>
                    <div class="flex justify-between gap-3">
                      <dt class="text-muted-foreground">{{ adminText('k00bw') }}</dt>
                      <dd class="font-medium">{{ editingUser.prestige || 0 }}</dd>
                    </div>
                  </dl>
                </section>
              </aside>

              <section class="rounded-lg border bg-background">
                <div class="flex items-center justify-between gap-3 border-b px-3 py-2.5">
                  <div>
                    <div class="flex items-center gap-2 text-sm font-semibold">
                      <Award class="size-4 text-muted-foreground" />
                      {{ adminText('k00bx') }}
                    </div>
                    <p class="mt-0.5 text-xs text-muted-foreground">{{ adminText('k00by') }}</p>
                  </div>
                  <Loader2 v-if="badgeLoading" class="size-4 animate-spin text-muted-foreground" />
                </div>

                <div class="space-y-3 p-3">
                  <div v-if="autoBadges().length" class="space-y-2">
                    <div class="text-xs font-medium text-muted-foreground">{{ adminText('k00bz') }}</div>
                    <div class="flex flex-wrap gap-1.5">
                      <span v-for="badge in autoBadges()" :key="badge.code" class="rounded-md bg-muted px-2 py-1 text-xs text-muted-foreground">
                        {{ badge.name }}
                      </span>
                    </div>
                  </div>

                  <div class="space-y-2">
                    <div class="text-xs font-medium text-muted-foreground">{{ adminText('k00c0') }}</div>
                    <div v-if="badgeOptions.length" class="grid grid-cols-3 gap-2 sm:grid-cols-4 md:grid-cols-5 xl:grid-cols-6">
                      <button
                        v-for="badge in badgeOptions"
                        :key="badge.code"
                        class="group relative flex min-h-20 min-w-0 flex-col items-center justify-center gap-1 rounded-md border p-2 text-center transition-colors hover:bg-muted/50"
                        :class="selectedBadgeCodes.includes(badge.code) ? 'border-primary bg-primary/5 shadow-xs' : 'border-transparent bg-muted/40'"
                        type="button"
                        :title="badge.description || badge.code"
                        @click="toggleBadge(badge.code)"
                      >
                        <span
                          class="flex size-10 shrink-0 items-center justify-center ring-1 ring-inset transition-transform group-hover:scale-105"
                          :class="badgeToneClass(badge)"
                          style="clip-path: polygon(25% 5%, 75% 5%, 100% 50%, 75% 95%, 25% 95%, 0 50%)"
                        >
                          <img :src="badgeIconURL(badge)" :alt="badge.name" class="size-5 object-contain" />
                        </span>
                        <span class="block max-w-full truncate text-xs font-medium leading-4">{{ badge.name }}</span>
                        <span class="absolute right-1.5 top-1.5 grid size-4 place-items-center rounded-full border bg-background text-[10px] font-bold" :class="selectedBadgeCodes.includes(badge.code) ? 'border-primary bg-primary text-primary-foreground' : 'text-transparent'">✓</span>
                      </button>
                    </div>
                    <div v-else class="rounded-md border border-dashed p-6 text-center text-sm text-muted-foreground">
                      {{ badgeLoading ? adminText('k006q') : adminText('k006r') }}
                    </div>
                  </div>
                </div>
              </section>
            </div>
          </div>

          <div class="flex justify-end gap-2 border-t bg-muted/20 px-4 py-3">
            <Button variant="outline" type="button" @click="editingUser = null">{{ adminText('k009q') }}</Button>
            <Button type="submit" :disabled="saving">
              {{ saving ? adminText('k005f') : adminText('k006s') }}
            </Button>
          </div>
        </form>
        </DialogContent>
      </Dialog>
    </BasicPage>
</template>
