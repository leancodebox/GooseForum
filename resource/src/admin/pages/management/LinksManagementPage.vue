<script setup lang="ts">import { adminText } from '@/admin/runtime/i18n-text'

import { computed, onMounted, reactive, ref } from 'vue'
import Draggable from 'vuedraggable'
import { ExternalLink, Eye, EyeOff, GripVertical, Link as LinkIcon, Pencil, Plus, RefreshCw, Send, ShieldCheck, Trash2 } from '@lucide/vue'
import AdminActionButton from '@/admin/components/AdminActionButton.vue'
import AdminConfirmDialog from '@/admin/components/AdminConfirmDialog.vue'
import { BasicPage } from '@/admin/components/global-layout'
import { Button } from '@/admin/components/ui/button'
import { Badge } from '@/admin/components/ui/badge'
import { Input } from '@/admin/components/ui/input'
import { Switch } from '@/admin/components/ui/switch'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/admin/components/ui/dialog'
import { getFriendLinks, saveFriendLinks } from '@/admin/runtime/api'
import { adminToast } from '@/admin/runtime/toast'
import type { AdminPayload, FriendLink, FriendLinkGroup, ManageHomeProps } from '@/admin/types'

defineProps<{
  payload: AdminPayload<ManageHomeProps>
}>()

const presetColors = ['#3b82f6', '#22c55e', '#a855f7', '#64748b', '#ef4444', '#f97316', '#f59e0b', '#10b981', '#06b6d4', '#6366f1', '#8b5cf6', '#ec4899']
const presetEmojis = ['🔗', '👥', '✍️', '📁', '⭐', '🔥', '💡', '🛠️', '🎨', '💬', '📢', '🏠', '🚀', '🎁']

const loading = ref(false)
const saving = ref(false)
const error = ref('')
const groups = ref<FriendLinkGroup[]>([])
const groupDialog = ref<{ mode: 'add' | 'edit', index: number | null } | null>(null)
const linkDialog = ref<{ mode: 'add' | 'edit', groupIndex: number, linkIndex: number | null } | null>(null)
const deleteDialog = ref<{ type: 'group' | 'link', groupIndex: number, linkIndex?: number } | null>(null)
const groupForm = reactive<FriendLinkGroup>({ name: '', emoji: '🔗', color: '#3b82f6', links: [] })
const linkForm = reactive<FriendLink>({ name: '', desc: '', url: '', logoUrl: '', status: 1 })

const totalLinks = computed(() => groups.value.reduce((total, group) => total + (group.links?.length || 0), 0))

function groupColor(group: FriendLinkGroup) {
  return group.color || '#64748b'
}

function groupTint(group: FriendLinkGroup, opacity = 0.1) {
  const color = groupColor(group)
  return color.startsWith('#') ? `${color}${Math.round(opacity * 255).toString(16).padStart(2, '0')}` : color
}

function normalize(groupsValue: FriendLinkGroup[] | null | undefined) {
  return Array.isArray(groupsValue) ? groupsValue.map(group => ({ ...group, links: Array.isArray(group.links) ? group.links : [] })) : []
}

async function loadLinks() {
  loading.value = true
  error.value = ''
  try {
    groups.value = normalize(await getFriendLinks())
  } catch (err) {
    error.value = err instanceof Error ? err.message : adminText('k0018')
  } finally {
    loading.value = false
  }
}

async function persist(nextGroups = groups.value) {
  saving.value = true
  try {
    const normalized = normalize(nextGroups)
    groups.value = normalized
    await saveFriendLinks(normalized)
    adminToast.success(adminText('k000e'))
  } catch (err) {
    adminToast.error(err, adminText('k0019'))
    await loadLinks()
  } finally {
    saving.value = false
  }
}

function openAddGroup() {
  Object.assign(groupForm, { name: '', emoji: '🔗', color: '#3b82f6', links: [] })
  groupDialog.value = { mode: 'add', index: null }
}

function openEditGroup(group: FriendLinkGroup, index: number) {
  Object.assign(groupForm, { name: group.name, emoji: group.emoji || '🔗', color: group.color || '#3b82f6', links: group.links || [] })
  groupDialog.value = { mode: 'edit', index }
}

async function submitGroup() {
  if (!groupForm.name.trim()) {
    adminToast.warning(adminText('k003k'))
    return
  }
  const next = normalize(groups.value)
  if (groupDialog.value?.mode === 'add') {
    next.push({ name: groupForm.name.trim(), emoji: groupForm.emoji, color: groupForm.color, links: [] })
  } else if (groupDialog.value?.index !== null && groupDialog.value?.index !== undefined) {
    next[groupDialog.value.index] = { ...next[groupDialog.value.index], name: groupForm.name.trim(), emoji: groupForm.emoji, color: groupForm.color }
  }
  groupDialog.value = null
  await persist(next)
}

function openAddLink(groupIndex: number) {
  Object.assign(linkForm, { name: '', desc: '', url: '', logoUrl: '', status: 1 })
  linkDialog.value = { mode: 'add', groupIndex, linkIndex: null }
}

function openEditLink(groupIndex: number, linkIndex: number, link: FriendLink) {
  Object.assign(linkForm, { name: link.name, desc: link.desc || '', url: link.url, logoUrl: link.logoUrl || '', status: link.status ?? 1 })
  linkDialog.value = { mode: 'edit', groupIndex, linkIndex }
}

async function submitLink() {
  if (!linkForm.name.trim() || !linkForm.url.trim()) {
    adminToast.warning(adminText('k003l'))
    return
  }
  if (!linkDialog.value) return
  const next = normalize(groups.value)
  const link = { name: linkForm.name.trim(), desc: linkForm.desc || '', url: linkForm.url.trim(), logoUrl: linkForm.logoUrl || '', status: Number(linkForm.status ?? 1) }
  if (linkDialog.value.mode === 'add') {
    next[linkDialog.value.groupIndex].links.push(link)
  } else if (linkDialog.value.linkIndex !== null) {
    next[linkDialog.value.groupIndex].links[linkDialog.value.linkIndex] = link
  }
  linkDialog.value = null
  await persist(next)
}

async function toggleLinkStatus(groupIndex: number, linkIndex: number, checked: boolean) {
  const next = normalize(groups.value)
  next[groupIndex].links[linkIndex].status = checked ? 1 : 0
  await persist(next)
}

async function persistDrag() {
  await persist(groups.value)
}

async function confirmDelete() {
  if (!deleteDialog.value) return
  const next = normalize(groups.value)
  if (deleteDialog.value.type === 'group') {
    next.splice(deleteDialog.value.groupIndex, 1)
  } else if (deleteDialog.value.linkIndex !== undefined) {
    next[deleteDialog.value.groupIndex].links.splice(deleteDialog.value.linkIndex, 1)
  }
  deleteDialog.value = null
  await persist(next)
}

onMounted(() => {
  void loadLinks()
})
</script>

<template>
  <BasicPage :title="adminText('k006u')" :description="adminText('k006v')" sticky>
    <template #actions>
      <div class="flex items-center gap-2">
        <Button variant="outline" type="button" @click="loadLinks">
          <RefreshCw class="size-4" />
          {{ adminText('k004q') }}
        </Button>
        <Button type="button" @click="openAddGroup">
          <Plus class="size-4" />
          {{ adminText('k006w') }}
        </Button>
      </div>
    </template>

      <div v-if="loading" class="flex h-64 items-center justify-center rounded-lg border text-muted-foreground">{{ adminText('k0046') }}</div>
      <div v-else-if="error" class="rounded-lg border border-destructive/30 bg-destructive/5 p-4 text-sm text-destructive">{{ error }}</div>
      <div v-else class="grid gap-4 xl:grid-cols-[minmax(0,1fr)_260px]">
        <Draggable
          v-model="groups"
          item-key="name"
          class="space-y-5 pb-8"
          handle=".js-group-handle"
          ghost-class="opacity-40"
          chosen-class="opacity-80"
          @end="persistDrag"
        >
          <template #item="{ element: group, index: groupIndex }">
            <section class="relative space-y-3">
              <header class="group/header flex cursor-grab items-center justify-between gap-3 border-b border-border/70 pb-2 active:cursor-grabbing js-group-handle">
                <div class="flex min-w-0 items-center gap-2">
                  <GripVertical class="size-4 shrink-0 text-muted-foreground/30 opacity-0 transition-opacity group-hover/header:opacity-100" />
                  <span
                    class="grid size-7 shrink-0 place-items-center rounded-md border text-sm"
                    :style="{ color: groupColor(group), borderColor: groupTint(group, 0.28), backgroundColor: groupTint(group, 0.1) }"
                  >
                    {{ group.emoji || '🔗' }}
                  </span>
                  <h2 class="truncate text-base font-semibold text-foreground">{{ group.name }}</h2>
                  <span class="rounded-full bg-muted px-2 py-0.5 text-[11px] font-semibold text-muted-foreground">{{ group.links.length }}</span>
                </div>
                <div class="flex shrink-0 items-center gap-1">
                  <Button variant="ghost" size="sm" class="h-8 gap-1.5 px-2 text-xs" type="button" @click="openAddLink(groupIndex)">
                    <Plus class="size-3.5" />
                    {{ adminText('k0094') }}
                  </Button>
                  <AdminActionButton compact :title="adminText('k005j')" @click="openEditGroup(group, groupIndex)">
                    <Pencil class="size-4" />
                  </AdminActionButton>
                  <AdminActionButton compact tone="danger" :title="adminText('k005i')" @click="deleteDialog = { type: 'group', groupIndex }">
                    <Trash2 class="size-4" />
                  </AdminActionButton>
                </div>
              </header>

              <Draggable
                v-model="group.links"
                item-key="name"
                :group="{ name: 'friend-links' }"
                handle=".js-link-handle"
                ghost-class="opacity-40"
                chosen-class="opacity-80"
                class="min-h-24 rounded-lg border border-dashed border-transparent empty:border-muted-foreground/20 empty:bg-muted/20"
                @end="persistDrag"
              >
                <template #item="{ element: link, index: linkIndex }">
                  <article
                    class="group relative mb-2 inline-flex w-full rounded-md border bg-card px-2.5 py-2 pr-3 transition hover:border-primary/30 md:mr-2 md:w-[calc(33.333%-0.5rem)] xl:w-[calc(25%-0.5rem)] 2xl:w-[calc(20%-0.5rem)]"
                    :class="(link.status ?? 1) === 0 && 'opacity-65'"
                    :style="{ borderColor: groupTint(group, 0.22) }"
                  >
                    <GripVertical class="js-link-handle absolute left-1 top-1 size-3.5 cursor-grab text-muted-foreground/30 opacity-0 transition-opacity group-hover:opacity-100 active:cursor-grabbing" />
                    <div class="flex min-w-0 flex-1 items-start gap-2 pl-1">
                      <div class="relative shrink-0">
                        <img v-if="link.logoUrl" :src="link.logoUrl" class="size-8 rounded-md border border-border/70 object-cover" alt="" :class="(link.status ?? 1) === 0 && 'grayscale'" />
                        <span v-else class="grid size-8 place-items-center rounded-md border border-border/70 bg-muted text-muted-foreground">
                          <LinkIcon class="size-4" />
                        </span>
                        <button
                          type="button"
                          class="absolute -bottom-1 -right-1 rounded-full border p-0.5 shadow-sm transition-all"
                          :class="(link.status ?? 1) === 0 ? 'bg-background text-muted-foreground hover:bg-muted' : 'bg-primary text-primary-foreground opacity-0 group-hover:opacity-100'"
                          :title="(link.status ?? 1) === 0 ? adminText('k006x') : adminText('k006y')"
                          @click="toggleLinkStatus(groupIndex, linkIndex, (link.status ?? 1) === 0)"
                        >
                          <EyeOff v-if="(link.status ?? 1) === 0" class="size-3" />
                          <Eye v-else class="size-3" />
                        </button>
                      </div>
                      <div class="min-w-0 flex-1">
                        <a
                          :href="link.url"
                          target="_blank"
                          rel="noopener noreferrer"
                          class="block truncate text-[13px] font-semibold text-foreground transition-colors hover:text-primary"
                          :class="(link.status ?? 1) === 0 && 'text-muted-foreground'"
                          :title="adminText('k0076', { name: link.name })"
                          @click.stop
                        >
                          {{ link.name }}
                        </a>
                        <p class="mt-0.5 truncate text-[11px] leading-4 text-muted-foreground">{{ link.desc || link.url }}</p>
                      </div>
                    </div>
                    <div class="absolute right-1 top-1 flex gap-1 opacity-0 transition-opacity group-hover:opacity-100">
                      <a
                        :href="link.url"
                        target="_blank"
                        rel="noopener noreferrer"
                        class="grid size-6 place-items-center rounded-md border bg-background/95 text-muted-foreground shadow-sm transition-colors hover:bg-accent hover:text-primary"
                        :title="adminText('k0077')"
                        @click.stop
                      >
                        <ExternalLink class="size-3.5" />
                      </a>
                      <AdminActionButton compact :title="adminText('k005j')" @click="openEditLink(groupIndex, linkIndex, link)">
                        <Pencil class="size-3.5" />
                      </AdminActionButton>
                      <AdminActionButton compact tone="danger" :title="adminText('k005i')" @click="deleteDialog = { type: 'link', groupIndex, linkIndex }">
                        <Trash2 class="size-3.5" />
                      </AdminActionButton>
                    </div>
                  </article>
                </template>
                <template #footer>
                  <button
                    v-if="group.links.length === 0"
                    type="button"
                    class="flex h-28 w-full flex-col items-center justify-center rounded-lg border border-dashed border-muted-foreground/20 bg-muted/20 text-sm text-muted-foreground transition-colors hover:bg-muted/50 hover:text-foreground"
                    @click="openAddLink(groupIndex)"
                  >
                    <LinkIcon class="mb-2 size-5" />
                    {{ adminText('k00a2') }}
                  </button>
                </template>
              </Draggable>
            </section>
          </template>
        </Draggable>
        <aside class="space-y-3">
          <div class="rounded-lg border bg-background p-4">
            <h2 class="text-sm font-semibold text-foreground">{{ adminText('k00a3') }}</h2>
            <p class="mt-2 text-sm leading-6 text-muted-foreground">{{ adminText('k00a4') }}</p>
            <Button class="mt-4 h-9 gap-1.5" as="a" href="/publish" target="_blank" rel="noopener noreferrer">
              <Send class="size-4" />
              {{ adminText('k00a5') }}
            </Button>
          </div>
          <div class="rounded-lg border bg-background p-4">
            <h2 class="text-sm font-semibold text-foreground">{{ adminText('k00a6') }}</h2>
            <div class="mt-3 space-y-2 text-sm text-muted-foreground">
              <div class="flex gap-2"><ShieldCheck class="mt-0.5 size-4 shrink-0 text-emerald-600" /><span>{{ adminText('k00a7') }}</span></div>
              <div class="flex gap-2"><ShieldCheck class="mt-0.5 size-4 shrink-0 text-emerald-600" /><span>{{ adminText('k00a8') }}</span></div>
              <div class="flex gap-2"><ShieldCheck class="mt-0.5 size-4 shrink-0 text-emerald-600" /><span>{{ adminText('k00a9') }}</span></div>
            </div>
          </div>
          <Badge variant="secondary" class="rounded-md">{{ adminText('k0054') }} {{ totalLinks }} {{ adminText('k00aa') }}</Badge>
        </aside>
      </div>

      <Dialog :open="groupDialog !== null" @update:open="(open) => !open && (groupDialog = null)">
        <DialogContent class="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{{ groupDialog?.mode === 'edit' ? adminText('k006z') : adminText('k006w') }}</DialogTitle>
            <DialogDescription>{{ adminText('k00ab') }}</DialogDescription>
          </DialogHeader>
          <form class="grid gap-4" @submit.prevent="submitGroup">
            <label class="grid gap-2 text-sm font-medium">
              {{ adminText('k00ac') }}
              <Input v-model="groupForm.name" :placeholder="adminText('k0078')" />
            </label>
            <div class="grid gap-4 sm:grid-cols-2">
              <label class="grid gap-2 text-sm font-medium">
                Emoji
                <Input v-model="groupForm.emoji" />
              </label>
              <label class="grid gap-2 text-sm font-medium">
                {{ adminText('k00ad') }}
                <Input v-model="groupForm.color" />
              </label>
            </div>
            <div class="flex flex-wrap gap-2">
              <button v-for="emoji in presetEmojis" :key="emoji" class="grid size-8 place-items-center rounded-md hover:bg-accent" type="button" @click="groupForm.emoji = emoji">{{ emoji }}</button>
            </div>
            <div class="flex flex-wrap gap-2">
              <button v-for="color in presetColors" :key="color" class="size-7 rounded-full border" :style="{ backgroundColor: color }" type="button" @click="groupForm.color = color" />
            </div>
            <DialogFooter>
              <Button variant="outline" type="button" @click="groupDialog = null">{{ adminText('k009q') }}</Button>
              <Button type="submit" :disabled="saving">{{ saving ? adminText('k005f') : adminText('k005g') }}</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <Dialog :open="linkDialog !== null" @update:open="(open) => !open && (linkDialog = null)">
        <DialogContent class="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{{ linkDialog?.mode === 'edit' ? adminText('k0070') : adminText('k0071') }}</DialogTitle>
            <DialogDescription>{{ adminText('k00ae') }}</DialogDescription>
          </DialogHeader>
          <form class="grid gap-4" @submit.prevent="submitLink">
            <label class="grid gap-2 text-sm font-medium">{{ adminText('k00af') }}<Input v-model="linkForm.name" :placeholder="adminText('k0079')" /></label>
            <label class="grid gap-2 text-sm font-medium">URL<Input v-model="linkForm.url" placeholder="https://..." /></label>
            <label class="grid gap-2 text-sm font-medium">{{ adminText('k00ag') }}<Input v-model="linkForm.desc" :placeholder="adminText('k007a')" /></label>
            <label class="grid gap-2 text-sm font-medium">Logo URL<Input v-model="linkForm.logoUrl" :placeholder="adminText('k007b')" /></label>
            <div class="flex items-center justify-between rounded-lg border p-3">
              <div>
                <div class="text-sm font-medium">{{ adminText('k00ah') }}</div>
                <div class="text-xs text-muted-foreground">{{ adminText('k00ai') }}</div>
              </div>
              <Switch :model-value="linkForm.status === 1" @update:model-value="(checked: boolean) => linkForm.status = checked ? 1 : 0" />
            </div>
            <DialogFooter>
              <Button variant="outline" type="button" @click="linkDialog = null">{{ adminText('k009q') }}</Button>
              <Button type="submit" :disabled="saving">{{ saving ? adminText('k005f') : adminText('k005g') }}</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <AdminConfirmDialog
        :open="deleteDialog !== null"
        :title="deleteDialog?.type === 'group' ? adminText('k0072') : adminText('k0073')"
        :description="deleteDialog?.type === 'group' ? adminText('k0074') : adminText('k0075')"
        :loading="saving"
        @update:open="(open) => !open && (deleteDialog = null)"
        @confirm="confirmDelete"
      />
    </BasicPage>
</template>
