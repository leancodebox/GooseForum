<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import {
  Check,
  Clock3,
  LockKeyhole,
  Pencil,
  Plus,
  RefreshCw,
  ShieldCheck,
  Trash2,
  UserPlus,
  UsersRound,
} from '@lucide/vue'
import { useI18n } from 'vue-i18n'
import { RouterLink } from 'vue-router'
import AdminActionButton from '@/admin/components/AdminActionButton.vue'
import AdminConfirmDialog from '@/admin/components/AdminConfirmDialog.vue'
import AdminSection from '@/admin/components/AdminSection.vue'
import { BasicPage } from '@/admin/components/global-layout'
import { Avatar, AvatarFallback, AvatarImage } from '@/admin/components/ui/avatar'
import { Badge } from '@/admin/components/ui/badge'
import { Button } from '@/admin/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/admin/components/ui/dialog'
import { Input } from '@/admin/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/admin/components/ui/select'
import { Switch } from '@/admin/components/ui/switch'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/admin/components/ui/table'
import {
  deleteAccessGroup,
  deleteAccessGroupMember,
  getAccessControlOverview,
  reviewAccessGroupApplication,
  saveAccessGroup,
  saveAccessGroupMember,
} from '@/admin/runtime/api'
import { adminToast } from '@/admin/runtime/toast'
import type {
  AccessControlOverview,
  AccessGroup,
  AccessGroupMember,
  AdminPayload,
  ManageHomeProps,
} from '@/admin/types'

defineProps<{ payload: AdminPayload<ManageHomeProps> }>()

const { t } = useI18n()
const loading = ref(false)
const error = ref('')
const overview = ref<AccessControlOverview>({ groups: [], categories: [] })
const selectedGroupId = ref(0)
const groupDialogOpen = ref(false)
const groupSaving = ref(false)
const deletingGroup = ref<AccessGroup | null>(null)
const groupDeleting = ref(false)
const memberSaving = ref(false)
const memberUsername = ref('')
const memberRole = ref<'member' | 'manager'>('member')
const deletingMember = ref<{ groupId: number; member: AccessGroupMember } | null>(null)
const memberDeleting = ref(false)
const reviewingMemberId = ref(0)
const groupForm = reactive({
  id: 0,
  name: '',
  joinMode: 'invite_only' as 'invite_only' | 'application',
  status: 1,
})

const selectedGroup = computed(() => overview.value.groups.find((group) => group.id === selectedGroupId.value))
const selectedMembers = computed(() => {
  return [...(selectedGroup.value?.members || [])].sort((a, b) => {
    if (a.status !== b.status) return b.status - a.status
    if (a.memberRole !== b.memberRole) return a.memberRole === 'manager' ? -1 : 1
    return a.userId - b.userId
  })
})

function activeMemberCount(group: AccessGroup) {
  return group.members.filter((member) => member.status === 1).length
}

function pendingMemberCount(group: AccessGroup) {
  return group.members.filter((member) => member.status === 2).length
}

const customGroupCount = computed(() => overview.value.groups.filter((group) => !group.systemKey).length)
const restrictedCategoryCount = computed(
  () => overview.value.categories.filter((category) => category.isRestricted).length,
)

async function loadOverview(preferredGroupId = selectedGroupId.value) {
  loading.value = true
  error.value = ''
  try {
    overview.value = await getAccessControlOverview()
    const available = overview.value.groups.some((group) => group.id === preferredGroupId)
    selectedGroupId.value = available ? preferredGroupId : overview.value.groups[0]?.id || 0
  } catch (err) {
    error.value = err instanceof Error ? err.message : t('accessGroups.loadFailed')
  } finally {
    loading.value = false
  }
}

function grantLevel(group: AccessGroup | undefined, categoryId: number) {
  return group?.grants.find((grant) => grant.categoryId === categoryId)?.level || 0
}

function levelLabel(level: number) {
  const keys = ['none', 'read', 'reply', 'create', 'manage'] as const
  return t(`accessGroups.level.${keys[level] || 'none'}`)
}

function updateMemberRole(value: unknown) {
  if (value === 'member' || value === 'manager') memberRole.value = value
}

function updateGroupJoinMode(value: unknown) {
  if (value === 'invite_only' || value === 'application') groupForm.joinMode = value
}

function openCreateGroup() {
  Object.assign(groupForm, { id: 0, name: '', joinMode: 'invite_only', status: 1 })
  groupDialogOpen.value = true
}

function openEditGroup(group: AccessGroup) {
  if (group.systemKey) return
  Object.assign(groupForm, {
    id: group.id,
    name: group.name,
    joinMode: group.joinMode as 'invite_only' | 'application',
    status: group.status,
  })
  groupDialogOpen.value = true
}

async function submitGroup() {
  if (groupSaving.value) return
  if (!groupForm.name.trim()) {
    adminToast.warning(t('accessGroups.groupNameRequired'))
    return
  }
  groupSaving.value = true
  try {
    const id = await saveAccessGroup({ ...groupForm, name: groupForm.name.trim() })
    groupDialogOpen.value = false
    await loadOverview(id)
    adminToast.success(t('accessGroups.saved'))
  } catch (err) {
    adminToast.error(err, t('accessGroups.saveFailed'))
  } finally {
    groupSaving.value = false
  }
}

async function confirmRemoveGroup() {
  if (!deletingGroup.value || groupDeleting.value) return
  groupDeleting.value = true
  try {
    await deleteAccessGroup(deletingGroup.value.id)
    deletingGroup.value = null
    await loadOverview(0)
    adminToast.success(t('accessGroups.deleted'))
  } catch (err) {
    adminToast.error(err, t('accessGroups.deleteFailed'))
  } finally {
    groupDeleting.value = false
  }
}

async function addMember() {
  const group = selectedGroup.value
  const username = memberUsername.value.trim()
  if (!group || group.systemKey || group.status !== 1 || !username || memberSaving.value) return
  memberSaving.value = true
  try {
    await saveAccessGroupMember({ groupId: group.id, username, memberRole: memberRole.value })
    memberUsername.value = ''
    await loadOverview(group.id)
    adminToast.success(t('accessGroups.memberSaved'))
  } catch (err) {
    adminToast.error(err, t('accessGroups.memberSaveFailed'))
  } finally {
    memberSaving.value = false
  }
}

async function confirmRemoveMember() {
  const target = deletingMember.value
  if (!target || memberDeleting.value) return
  memberDeleting.value = true
  try {
    await deleteAccessGroupMember(target.groupId, target.member.id)
    deletingMember.value = null
    await loadOverview(selectedGroupId.value)
    adminToast.success(t('accessGroups.memberDeleted'))
  } catch (err) {
    adminToast.error(err, t('accessGroups.memberDeleteFailed'))
  } finally {
    memberDeleting.value = false
  }
}

async function reviewApplication(memberId: number, approve: boolean) {
  const group = selectedGroup.value
  if (!group || reviewingMemberId.value) return
  reviewingMemberId.value = memberId
  try {
    await reviewAccessGroupApplication(group.id, memberId, approve)
    await loadOverview(group.id)
    adminToast.success(t(approve ? 'accessGroups.applicationApproved' : 'accessGroups.applicationRejected'))
  } catch (err) {
    adminToast.error(err, t('accessGroups.memberSaveFailed'))
  } finally {
    reviewingMemberId.value = 0
  }
}

function memberInitial(member: AccessGroupMember) {
  return (member.username || String(member.userId)).slice(0, 1).toUpperCase()
}

onMounted(() => void loadOverview())
</script>

<template>
  <BasicPage :title="t('accessGroups.title')" :description="t('accessGroups.description')" sticky>
    <template #actions>
      <div class="flex items-center gap-2">
        <Button variant="outline" type="button" :disabled="loading" @click="loadOverview()">
          <RefreshCw class="size-4" :class="loading ? 'animate-spin' : ''" />
          {{ t('common.refresh') }}
        </Button>
        <Button type="button" @click="openCreateGroup">
          <Plus class="size-4" />
          {{ t('accessGroups.create') }}
        </Button>
      </div>
    </template>

    <div
      v-if="error"
      class="mb-4 flex items-start justify-between gap-4 rounded-lg border border-destructive/30 bg-destructive/5 p-4 text-sm text-destructive"
    >
      <span>{{ error }}</span>
      <Button variant="outline" size="sm" type="button" @click="loadOverview()">{{ t('common.retry') }}</Button>
    </div>

    <div class="mb-3 flex flex-wrap items-center gap-2 text-sm text-muted-foreground">
      <Badge variant="secondary"
        ><UsersRound class="size-3" />{{ t('accessGroups.customGroupCount', { count: customGroupCount }) }}</Badge
      >
      <Badge variant="outline"
        ><LockKeyhole class="size-3" />{{
          t('accessGroups.restrictedCategoryCount', { count: restrictedCategoryCount })
        }}</Badge
      >
      <span v-if="loading && overview.groups.length" class="inline-flex items-center gap-1.5">
        <RefreshCw class="size-3.5 animate-spin" />{{ t('accessGroups.refreshing') }}
      </span>
    </div>

    <div class="grid items-start gap-4 xl:grid-cols-[17rem_minmax(0,1fr)]">
      <AdminSection class="xl:sticky xl:top-36">
        <template #header>
          <div class="flex items-center justify-between gap-3">
            <div>
              <div class="flex items-center gap-2 text-sm font-semibold">
                <UsersRound class="size-4" />{{ t('accessGroups.groups') }}
              </div>
              <p class="mt-0.5 text-xs text-muted-foreground">{{ t('accessGroups.systemGroupHint') }}</p>
            </div>
          </div>
        </template>
        <div v-if="loading && !overview.groups.length" class="space-y-2 p-3">
          <div v-for="index in 4" :key="index" class="h-14 animate-pulse rounded-md bg-muted" />
        </div>
        <div v-else-if="!overview.groups.length" class="p-8 text-center text-sm text-muted-foreground">
          {{ t('accessGroups.noGroups') }}
        </div>
        <div v-else class="space-y-1 p-2">
          <button
            v-for="group in overview.groups"
            :key="group.id"
            type="button"
            class="group flex w-full items-center gap-3 rounded-md px-3 py-2.5 text-left outline-none transition-colors hover:bg-muted/70 focus-visible:ring-2 focus-visible:ring-ring"
            :class="[
              selectedGroupId === group.id ? 'bg-primary/10 text-primary' : 'text-foreground',
              group.status !== 1 ? 'opacity-60' : '',
            ]"
            @click="selectedGroupId = group.id"
          >
            <span
              class="grid size-8 shrink-0 place-items-center rounded-md border bg-background"
              :class="selectedGroupId === group.id ? 'border-primary/30' : ''"
            >
              <ShieldCheck v-if="group.systemKey" class="size-4" />
              <UsersRound v-else class="size-4" />
            </span>
            <span class="min-w-0 flex-1">
              <span class="block truncate text-sm font-medium">{{ group.name }}</span>
              <span class="mt-0.5 block truncate text-xs text-muted-foreground">
                {{
                  group.systemKey
                    ? t(`accessGroups.systemKey.${group.systemKey}`)
                    : t(`accessGroups.joinMode.${group.joinMode}`)
                }}
              </span>
            </span>
            <Badge variant="outline" class="min-w-7 px-1.5">{{ activeMemberCount(group) }}</Badge>
          </button>
        </div>
      </AdminSection>

      <div v-if="selectedGroup" class="min-w-0 space-y-4">
        <AdminSection>
          <div class="flex flex-col gap-4 p-4 sm:flex-row sm:items-start sm:justify-between">
            <div class="min-w-0">
              <div class="flex flex-wrap items-center gap-2">
                <h2 class="truncate text-lg font-semibold">{{ selectedGroup.name }}</h2>
                <Badge v-if="selectedGroup.systemKey" variant="secondary"
                  ><ShieldCheck class="size-3" />{{ t('accessGroups.systemGroup') }}</Badge
                >
                <Badge :variant="selectedGroup.status === 1 ? 'outline' : 'destructive'">
                  {{ t(selectedGroup.status === 1 ? 'accessGroups.enabled' : 'accessGroups.disabled') }}
                </Badge>
              </div>
              <p class="mt-1 text-sm leading-6 text-muted-foreground">
                {{
                  selectedGroup.systemKey
                    ? t('accessGroups.systemImmutable')
                    : t(`accessGroups.joinModeDescription.${selectedGroup.joinMode}`)
                }}
              </p>
              <div class="mt-3 flex flex-wrap gap-x-5 gap-y-1 text-xs text-muted-foreground">
                <span>{{ t('accessGroups.memberCount', { count: activeMemberCount(selectedGroup) }) }}</span>
                <span v-if="pendingMemberCount(selectedGroup)">
                  {{ t('accessGroups.pendingCount', { count: pendingMemberCount(selectedGroup) }) }}
                </span>
                <span>{{
                  t('accessGroups.grantCount', {
                    count: selectedGroup.grants.filter((grant) => grant.level > 0).length,
                  })
                }}</span>
                <span class="font-mono">ID {{ selectedGroup.id }}</span>
              </div>
            </div>
            <div v-if="!selectedGroup.systemKey" class="flex shrink-0 items-center gap-2">
              <AdminActionButton @click="openEditGroup(selectedGroup)"
                ><Pencil class="size-3.5" />{{ t('common.edit') }}</AdminActionButton
              >
              <AdminActionButton tone="danger" @click="deletingGroup = selectedGroup"
                ><Trash2 class="size-3.5" />{{ t('common.delete') }}</AdminActionButton
              >
            </div>
          </div>
        </AdminSection>

        <AdminSection>
          <template #header>
            <div class="flex flex-wrap items-center justify-between gap-3">
              <div>
                <div class="flex items-center gap-2 text-sm font-semibold">
                  <LockKeyhole class="size-4" />{{ t('accessGroups.categoryPermissions') }}
                </div>
                <p class="mt-0.5 text-xs text-muted-foreground">{{ t('accessGroups.permissionSummaryHint') }}</p>
              </div>
              <Button variant="outline" size="sm" as-child>
                <RouterLink to="/admin/categories">
                  <Pencil class="size-3.5" />{{ t('accessGroups.manageCategoryPermissions') }}
                </RouterLink>
              </Button>
            </div>
          </template>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{{ t('accessGroups.category') }}</TableHead>
                <TableHead>{{ t('accessGroups.visibility') }}</TableHead>
                <TableHead class="w-52">{{ t('accessGroups.capability') }}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableEmpty v-if="!overview.categories.length" :colspan="3">{{
                t('accessGroups.noCategories')
              }}</TableEmpty>
              <TableRow v-for="category in overview.categories" :key="category.id" class="hover:bg-muted/35">
                <TableCell>
                  <div class="flex min-w-44 items-center gap-2.5">
                    <span
                      class="size-2.5 shrink-0 rounded-[3px] ring-1 ring-black/10"
                      :style="{ backgroundColor: category.color || '#64748b' }"
                    />
                    <span class="font-medium">{{ category.name }}</span>
                  </div>
                </TableCell>
                <TableCell>
                  <Badge :variant="category.isRestricted ? 'secondary' : 'outline'">
                    <LockKeyhole v-if="category.isRestricted" class="size-3" />
                    <Check v-else class="size-3" />
                    {{ t(category.isRestricted ? 'accessGroups.restricted' : 'accessGroups.public') }}
                  </Badge>
                </TableCell>
                <TableCell>
                  <Badge variant="outline">{{ levelLabel(grantLevel(selectedGroup, category.id)) }}</Badge>
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </AdminSection>

        <AdminSection v-if="!selectedGroup.systemKey">
          <template #header>
            <div>
              <div class="flex items-center gap-2 text-sm font-semibold">
                <UsersRound class="size-4" />{{ t('accessGroups.members') }}
              </div>
              <p class="mt-0.5 text-xs text-muted-foreground">{{ t('accessGroups.memberHint') }}</p>
            </div>
          </template>
          <form
            class="grid gap-3 border-b bg-muted/10 p-3 sm:grid-cols-[minmax(12rem,1fr)_10rem_auto] sm:items-end"
            @submit.prevent="addMember"
          >
            <label class="grid gap-1.5 text-xs font-medium text-muted-foreground">
              {{ t('accessGroups.username') }}
              <Input
                v-model="memberUsername"
                :disabled="memberSaving || selectedGroup.status !== 1"
                :placeholder="t('accessGroups.usernamePlaceholder')"
              />
            </label>
            <label class="grid gap-1.5 text-xs font-medium text-muted-foreground">
              {{ t('accessGroups.role') }}
              <Select
                :model-value="memberRole"
                :disabled="memberSaving || selectedGroup.status !== 1"
                @update:model-value="updateMemberRole"
              >
                <SelectTrigger class="w-full"><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="member">{{ t('accessGroups.memberRole.member') }}</SelectItem>
                  <SelectItem value="manager">{{ t('accessGroups.memberRole.manager') }}</SelectItem>
                </SelectContent>
              </Select>
            </label>
            <Button type="submit" :disabled="memberSaving || selectedGroup.status !== 1 || !memberUsername.trim()">
              <UserPlus class="size-4" />{{ memberSaving ? t('common.saving') : t('accessGroups.addMember') }}
            </Button>
          </form>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{{ t('accessGroups.member') }}</TableHead>
                <TableHead>{{ t('accessGroups.role') }}</TableHead>
                <TableHead>{{ t('accessGroups.status') }}</TableHead>
                <TableHead class="w-44 text-right">{{ t('accessGroups.actions') }}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableEmpty v-if="!selectedMembers.length" :colspan="4">{{ t('accessGroups.noMembers') }}</TableEmpty>
              <TableRow v-for="member in selectedMembers" :key="member.id" class="hover:bg-muted/35">
                <TableCell>
                  <div class="flex items-center gap-3">
                    <Avatar>
                      <AvatarImage v-if="member.avatarUrl" :src="member.avatarUrl" :alt="member.username" />
                      <AvatarFallback class="text-xs font-semibold">{{ memberInitial(member) }}</AvatarFallback>
                    </Avatar>
                    <div class="min-w-0">
                      <div class="truncate font-medium">{{ member.username || `#${member.userId}` }}</div>
                      <div class="font-mono text-xs text-muted-foreground">ID {{ member.userId }}</div>
                    </div>
                  </div>
                </TableCell>
                <TableCell
                  ><Badge variant="outline">{{ t(`accessGroups.memberRole.${member.memberRole}`) }}</Badge></TableCell
                >
                <TableCell>
                  <Badge :variant="member.status === 2 ? 'secondary' : 'outline'">
                    <Clock3 v-if="member.status === 2" class="size-3" />
                    <Check v-else class="size-3" />
                    {{ t(member.status === 2 ? 'accessGroups.pending' : 'accessGroups.active') }}
                  </Badge>
                </TableCell>
                <TableCell>
                  <div class="flex justify-end gap-2">
                    <template v-if="member.status === 2">
                      <AdminActionButton
                        :disabled="reviewingMemberId !== 0"
                        @click="reviewApplication(member.id, false)"
                        >{{ t('accessGroups.reject') }}</AdminActionButton
                      >
                      <AdminActionButton
                        tone="success"
                        :disabled="reviewingMemberId !== 0"
                        @click="reviewApplication(member.id, true)"
                        >{{ t('accessGroups.approve') }}</AdminActionButton
                      >
                    </template>
                    <AdminActionButton
                      v-else
                      compact
                      tone="danger"
                      :title="t('accessGroups.removeMember')"
                      @click="deletingMember = { groupId: selectedGroup.id, member }"
                    >
                      <Trash2 class="size-3.5" />
                    </AdminActionButton>
                  </div>
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </AdminSection>
      </div>

      <AdminSection v-else class="xl:col-start-2">
        <div class="grid min-h-64 place-items-center p-8 text-center text-sm text-muted-foreground">
          <div><UsersRound class="mx-auto mb-3 size-8 opacity-40" />{{ t('accessGroups.selectGroup') }}</div>
        </div>
      </AdminSection>
    </div>

    <Dialog :open="groupDialogOpen" @update:open="(open) => !groupSaving && (groupDialogOpen = open)">
      <DialogContent class="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{{ groupForm.id ? t('accessGroups.edit') : t('accessGroups.create') }}</DialogTitle>
          <DialogDescription>{{ t('accessGroups.groupFormHint') }}</DialogDescription>
        </DialogHeader>
        <form class="grid gap-5" @submit.prevent="submitGroup">
          <label class="grid gap-2 text-sm font-medium">
            {{ t('accessGroups.groupName') }}
            <Input
              v-model="groupForm.name"
              autofocus
              :disabled="groupSaving"
              :placeholder="t('accessGroups.groupNamePlaceholder')"
            />
          </label>
          <label class="grid gap-2 text-sm font-medium">
            {{ t('accessGroups.joinModeLabel') }}
            <Select :model-value="groupForm.joinMode" :disabled="groupSaving" @update:model-value="updateGroupJoinMode">
              <SelectTrigger class="w-full"><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="invite_only">{{ t('accessGroups.joinMode.invite_only') }}</SelectItem>
                <SelectItem value="application">{{ t('accessGroups.joinMode.application') }}</SelectItem>
              </SelectContent>
            </Select>
          </label>
          <div class="flex items-center justify-between gap-4 rounded-md border bg-muted/15 p-3">
            <div>
              <div class="text-sm font-medium">{{ t('accessGroups.enabled') }}</div>
              <p class="mt-0.5 text-xs text-muted-foreground">{{ t('accessGroups.enabledHint') }}</p>
            </div>
            <Switch
              :model-value="groupForm.status === 1"
              :disabled="groupSaving"
              @update:model-value="groupForm.status = $event ? 1 : 0"
            />
          </div>
          <DialogFooter>
            <Button variant="outline" type="button" :disabled="groupSaving" @click="groupDialogOpen = false">
              {{ t('common.cancel') }}
            </Button>
            <Button type="submit" :disabled="groupSaving">{{
              groupSaving ? t('common.saving') : t('common.save')
            }}</Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>

    <AdminConfirmDialog
      :open="deletingGroup !== null"
      :title="t('accessGroups.deleteGroupTitle')"
      :description="t('accessGroups.deleteConfirm', { name: deletingGroup?.name || '' })"
      :loading="groupDeleting"
      @update:open="(open) => !open && (deletingGroup = null)"
      @confirm="confirmRemoveGroup"
    />
    <AdminConfirmDialog
      :open="deletingMember !== null"
      :title="t('accessGroups.removeMemberTitle')"
      :description="
        t('accessGroups.removeMemberConfirm', {
          name: deletingMember?.member.username || `#${deletingMember?.member.userId || ''}`,
        })
      "
      :loading="memberDeleting"
      :confirm-text="t('accessGroups.removeMember')"
      @update:open="(open) => !open && (deletingMember = null)"
      @confirm="confirmRemoveMember"
    />
  </BasicPage>
</template>
