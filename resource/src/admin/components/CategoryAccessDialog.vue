<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { AlertTriangle, Check, LockKeyhole, RefreshCw, ShieldCheck, UsersRound } from '@lucide/vue'
import { useI18n } from 'vue-i18n'
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
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/admin/components/ui/select'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/admin/components/ui/table'
import { getAccessControlOverview, saveCategoryAccess } from '@/admin/runtime/api'
import { adminToast } from '@/admin/runtime/toast'
import type { AccessControlOverview, AccessGroup, AdminCategory } from '@/admin/types'

const props = defineProps<{
  category: AdminCategory | null
  open: boolean
}>()

const emit = defineEmits<{
  saved: [categoryId: number, isRestricted: boolean]
  'update:open': [value: boolean]
}>()

const { t } = useI18n()
const loading = ref(false)
const saving = ref(false)
const error = ref('')
const overview = ref<AccessControlOverview>({ groups: [], categories: [] })
const draft = ref<Record<number, number>>({})

const groups = computed(() => {
  const priority = (group: AccessGroup) => {
    if (group.systemKey === 'everyone') return 0
    if (group.systemKey === 'registered') return 1
    return 2
  }
  return [...overview.value.groups].sort((left, right) => {
    return priority(left) - priority(right) || left.name.localeCompare(right.name)
  })
})

const levelOptions = computed(() => {
  const keys = ['none', 'read', 'reply', 'create', 'manage'] as const
  return keys.map((key, value) => ({
    value,
    label: `${t(`accessGroups.level.${key}`)} (${t(`accessGroups.levelHint.${key}`)})`,
  }))
})

const everyoneGroup = computed(() => groups.value.find((group) => group.systemKey === 'everyone'))
const willBeRestricted = computed(() => {
  const group = everyoneGroup.value
  return !group || (draft.value[group.id] || 0) < 1
})
const isBecomingRestricted = computed(() => {
  const group = everyoneGroup.value
  return Boolean(group && currentLevel(group) >= 1 && willBeRestricted.value)
})
const restrictionConflictCount = computed(() => {
  return overview.value.categories.find((category) => category.id === props.category?.id)?.multiCategoryTopicCount || 0
})
const restrictionBlocked = computed(() => isBecomingRestricted.value && restrictionConflictCount.value > 0)
const hasReadableAudience = computed(() => {
  return groups.value.some((group) => group.status === 1 && (draft.value[group.id] || 0) >= 1)
})
const isDirty = computed(() => {
  return groups.value.some((group) => group.status === 1 && (draft.value[group.id] || 0) !== currentLevel(group))
})

function currentLevel(group: AccessGroup) {
  return group.grants.find((grant) => grant.categoryId === props.category?.id)?.level || 0
}

function updateLevel(groupId: number, value: unknown) {
  const level = Number(value)
  if (!Number.isInteger(level) || level < 0 || level > 4 || saving.value) return
  draft.value = { ...draft.value, [groupId]: level }
}

function close() {
  if (!saving.value) emit('update:open', false)
}

async function load() {
  if (!props.category || loading.value) return
  loading.value = true
  error.value = ''
  overview.value = { groups: [], categories: [] }
  draft.value = {}
  try {
    overview.value = await getAccessControlOverview()
    draft.value = Object.fromEntries(overview.value.groups.map((group) => [group.id, currentLevel(group)]))
  } catch (err) {
    error.value = err instanceof Error ? err.message : t('accessGroups.loadFailed')
  } finally {
    loading.value = false
  }
}

function enabledGrants() {
  return overview.value.groups
    .filter((group) => group.status === 1)
    .map((group) => ({ accessGroupId: group.id, level: draft.value[group.id] || 0 }))
}

// Existing multi-category topics must be resolved before the public grant can
// be removed; the backend repeats this check in the grant-write transaction.
async function save() {
  const category = props.category
  const everyone = everyoneGroup.value
  if (!category || !everyone || loading.value || saving.value || !isDirty.value || restrictionBlocked.value) return
  await persist(enabledGrants())
}

async function persist(grants: { accessGroupId: number; level: number }[]) {
  const category = props.category
  if (!category || saving.value) return
  saving.value = true
  try {
    await saveCategoryAccess({ categoryId: category.id, grants })
    const restricted = willBeRestricted.value
    adminToast.success(t('accessGroups.saved'))
    emit('saved', category.id, restricted)
    emit('update:open', false)
  } catch (err) {
    adminToast.error(err, t('accessGroups.saveFailed'))
  } finally {
    saving.value = false
  }
}

watch(
  () => [props.open, props.category?.id] as const,
  ([open]) => {
    if (open) void load()
  },
)
</script>

<template>
  <Dialog :open="open" @update:open="(value) => !value && close()">
    <DialogContent class="sm:max-w-2xl">
      <DialogHeader>
        <div class="flex items-start gap-3">
          <span class="mt-0.5 grid size-9 shrink-0 place-items-center rounded-md bg-primary/10 text-primary">
            <LockKeyhole class="size-4" />
          </span>
          <div class="min-w-0">
            <DialogTitle>{{
              t('accessGroups.categoryPermissionsTitle', { category: category?.category })
            }}</DialogTitle>
            <DialogDescription class="mt-1">{{ t('accessGroups.categoryPermissionsDescription') }}</DialogDescription>
          </div>
        </div>
      </DialogHeader>

      <div
        v-if="error"
        class="flex items-center justify-between gap-3 rounded-md border border-destructive/30 bg-destructive/5 p-3 text-sm text-destructive"
      >
        <span>{{ error }}</span>
        <Button variant="outline" size="sm" type="button" :disabled="loading" @click="load">
          {{ t('common.retry') }}
        </Button>
      </div>

      <div class="flex flex-wrap items-center justify-between gap-3 rounded-md border bg-muted/15 p-3">
        <div>
          <div class="text-sm font-medium">{{ t('accessGroups.resultingVisibility') }}</div>
          <p class="mt-0.5 text-xs text-muted-foreground">{{ t('accessGroups.visibilityDerivedHint') }}</p>
        </div>
        <Badge :variant="willBeRestricted ? 'secondary' : 'outline'">
          <LockKeyhole v-if="willBeRestricted" class="size-3" />
          <Check v-else class="size-3" />
          {{ t(willBeRestricted ? 'accessGroups.restricted' : 'accessGroups.public') }}
        </Badge>
      </div>

      <div
        v-if="!hasReadableAudience && !loading"
        class="flex gap-2 rounded-md border border-amber-500/25 bg-amber-500/5 p-3 text-sm text-muted-foreground"
      >
        <AlertTriangle class="mt-0.5 size-4 shrink-0 text-amber-600" />
        {{ t('accessGroups.noReadableAudienceWarning') }}
      </div>

      <div
        v-if="isBecomingRestricted && !loading"
        class="flex gap-2 rounded-md border border-amber-500/25 bg-amber-500/5 p-3 text-sm text-muted-foreground"
      >
        <AlertTriangle class="mt-0.5 size-4 shrink-0 text-amber-600" />
        {{ t('accessGroups.legacyImageVisibilityWarning') }}
      </div>

      <div
        v-if="restrictionBlocked && !loading"
        class="flex gap-2 rounded-md border border-destructive/30 bg-destructive/5 p-3 text-sm text-destructive"
      >
        <AlertTriangle class="mt-0.5 size-4 shrink-0" />
        {{ t('accessGroups.restrictionConflictWarning', { count: restrictionConflictCount }) }}
      </div>

      <div class="max-h-[26rem] overflow-auto rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{{ t('accessGroups.group') }}</TableHead>
              <TableHead>{{ t('accessGroups.groupType') }}</TableHead>
              <TableHead class="w-72">{{ t('accessGroups.capability') }}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-if="loading">
              <TableCell colspan="3" class="h-28 text-center text-muted-foreground">
                <RefreshCw class="mx-auto mb-2 size-4 animate-spin" />{{ t('common.loadingShort') }}
              </TableCell>
            </TableRow>
            <TableRow v-else-if="!groups.length">
              <TableCell colspan="3" class="h-28 text-center text-muted-foreground">
                {{ t('accessGroups.noGroups') }}
              </TableCell>
            </TableRow>
            <template v-else>
              <TableRow v-for="group in groups" :key="group.id" :class="group.status !== 1 ? 'opacity-55' : ''">
                <TableCell>
                  <div class="flex items-center gap-2 font-medium">
                    <ShieldCheck v-if="group.systemKey" class="size-4 text-muted-foreground" />
                    <UsersRound v-else class="size-4 text-muted-foreground" />
                    {{ group.name }}
                  </div>
                </TableCell>
                <TableCell class="text-sm text-muted-foreground">
                  {{
                    group.systemKey
                      ? t(`accessGroups.systemKey.${group.systemKey}`)
                      : t(`accessGroups.joinMode.${group.joinMode}`)
                  }}
                  <Badge v-if="group.status !== 1" variant="secondary" class="ml-1">
                    {{ t('accessGroups.disabled') }}
                  </Badge>
                </TableCell>
                <TableCell>
                  <Select
                    :model-value="String(draft[group.id] || 0)"
                    :disabled="group.status !== 1 || loading || saving"
                    @update:model-value="updateLevel(group.id, $event)"
                  >
                    <SelectTrigger class="w-full"><SelectValue /></SelectTrigger>
                    <SelectContent>
                      <SelectItem v-for="option in levelOptions" :key="option.value" :value="String(option.value)">
                        {{ option.label }}
                      </SelectItem>
                    </SelectContent>
                  </Select>
                </TableCell>
              </TableRow>
            </template>
          </TableBody>
        </Table>
      </div>

      <DialogFooter>
        <Button variant="outline" type="button" :disabled="saving" @click="close">{{ t('common.cancel') }}</Button>
        <Button type="button" :disabled="loading || saving || !everyoneGroup || !isDirty || restrictionBlocked" @click="save">
          {{ saving ? t('common.saving') : t('common.save') }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>

</template>
