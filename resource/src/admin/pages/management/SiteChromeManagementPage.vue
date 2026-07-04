<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Draggable from 'vuedraggable'
import {
  Bell,
  Eye,
  EyeOff,
  FileText,
  Flame,
  Heart,
  Inbox,
  Link as LinkIcon,
  MessageCircle,
  Pencil,
  Plus,
  Save,
  Scale,
  Trash2,
  TrendingUp,
  Upload,
} from '@lucide/vue'
import { BasicPage } from '@/admin/components/global-layout'
import { Button } from '@/admin/components/ui/button'
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
import { getSiteChrome, saveSiteChrome, uploadAdminImage } from '@/admin/runtime/api'
import { adminText } from '@/admin/runtime/i18n-text'
import { adminToast } from '@/admin/runtime/toast'
import type { AdminPayload, ManageHomeProps, SiteChromeConfig, SiteChromeGroup, SiteChromeItem } from '@/admin/types'

const props = defineProps<{
  payload: AdminPayload<ManageHomeProps>
}>()

const { t } = useI18n()

type GroupKey = 'header' | 'mainMenu' | 'resources'
type ItemDialogState = { group: GroupKey, index: number | null } | { group: 'sidebarGroup', groupIndex: number, index: number | null }
type FooterDialogState = { kind: 'link', index: number | null } | { kind: 'primary', index: number | null }

const loading = ref(false)
const loaded = ref(false)
const saving = ref(false)
const error = ref('')
const config = ref<SiteChromeConfig>({ header: [], mainMenu: [], resources: [], sidebarGroups: [] })
const dialog = ref<ItemDialogState | null>(null)
const groupDialog = ref<{ index: number | null } | null>(null)
const brandDialogOpen = ref(false)
const footerManageOpen = ref(false)
const footerDialog = ref<FooterDialogState | null>(null)
const form = reactive<SiteChromeItem>(emptyItem())
const groupForm = reactive<SiteChromeGroup>(emptyGroup())
const footerLinkForm = reactive({ name: '', url: '' })
const footerPrimaryForm = reactive({ content: '' })
const brandForm = reactive({ brandType: 'default', brandText: '', brandImage: '' })

const fixedMainItems = [
  { key: 'topics', labelKey: 'shell.nav.topics', icon: MessageCircle, active: true },
  { key: 'hot', labelKey: 'shell.nav.hot', icon: Flame },
  { key: 'popular', labelKey: 'shell.nav.popular', icon: TrendingUp },
  { key: 'messages', labelKey: 'shell.nav.messages', icon: Inbox },
  { key: 'notifications', labelKey: 'shell.nav.notifications', icon: Bell },
  { key: 'drafts', labelKey: 'shell.nav.drafts', icon: FileText },
  { key: 'moderation', labelKey: 'shell.nav.moderation', icon: Scale },
]
const fixedResourceItems = [
  { key: 'links', labelKey: 'shell.nav.links', icon: LinkIcon },
  { key: 'sponsors', labelKey: 'shell.nav.sponsors', icon: Heart },
]
const categories = computed(() => props.payload.layout.sidebar.categories || [])
const brandType = computed(() => config.value.brandType || props.payload.layout.site.brandType || 'default')
const brandText = computed(() => config.value.brandText || props.payload.layout.site.brandText || props.payload.layout.site.name || 'GooseForum')
const brandImage = computed(() => config.value.brandImage || props.payload.layout.site.brandImage || props.payload.layout.site.logo || '')
const addButtonClass = 'rounded-md border border-dashed border-line/70 bg-base-100/40 font-medium text-base-content/55 transition hover:scale-[1.01] hover:border-primary/40 hover:bg-primary/10 hover:text-primary active:scale-[0.99]'

function normalizeFooterInfo(footerInfo: SiteChromeConfig['footerInfo'] = { primary: [], list: [] }) {
  const normalized = footerInfo && typeof footerInfo === 'object'
    ? footerInfo
    : { primary: [], list: [] }
  return {
    primary: Array.isArray(normalized.primary) ? normalized.primary.map(item => ({ content: item?.content || '' })) : [],
    list: Array.isArray(normalized.list) ? normalized.list.map(item => ({ name: item?.name || '', url: item?.url || '' })) : [],
  }
}

function emptyItem(): SiteChromeItem {
  return {
    id: crypto.randomUUID(),
    enabled: true,
    type: 'link',
    label: '',
    i18nLabel: '',
    url: '',
  }
}

function emptyGroup(): SiteChromeGroup {
  return {
    id: crypto.randomUUID(),
    title: adminText('k00e4'),
    i18nLabel: '',
    items: [],
  }
}

function normalizeItem(item: Partial<SiteChromeItem> = {}): SiteChromeItem {
  return {
    ...emptyItem(),
    ...item,
    id: item.id || crypto.randomUUID(),
    enabled: item.enabled !== false,
    type: ['link', 'text'].includes(item.type || '') ? item.type || 'link' : 'link',
    label: item.label || '',
    i18nLabel: item.i18nLabel || '',
    url: item.url || '',
  }
}

function normalizeGroup(group: Partial<SiteChromeGroup> = {}): SiteChromeGroup {
  return {
    ...emptyGroup(),
    ...group,
    id: group.id || crypto.randomUUID(),
    title: group.title || adminText('k00e4'),
    i18nLabel: group.i18nLabel || '',
    items: Array.isArray(group.items) ? group.items.map(normalizeItem) : [],
  }
}

function normalize(value: Partial<SiteChromeConfig> = {}): SiteChromeConfig {
  return {
    header: normalizeHeaderItems(value.header),
    mainMenu: Array.isArray(value.mainMenu) ? value.mainMenu.map(normalizeItem) : [],
    resources: Array.isArray(value.resources) ? value.resources.map(normalizeItem) : [],
    sidebarGroups: Array.isArray(value.sidebarGroups) ? value.sidebarGroups.map(normalizeGroup) : [],
    footerInfo: normalizeFooterInfo(value.footerInfo),
    brandType: ['default', 'text', 'image'].includes(value.brandType || '') ? value.brandType || 'default' : 'default',
    brandText: value.brandText || '',
    brandImage: value.brandImage || '',
  }
}

function defaultHeaderItems(): SiteChromeItem[] {
  return [
    normalizeItem({ id: 'sponsors', enabled: true, type: 'link', label: 'Sponsors', i18nLabel: 'shell.nav.sponsors', url: '/sponsors' }),
    normalizeItem({ id: 'links', enabled: true, type: 'link', label: 'Links', i18nLabel: 'shell.nav.links', url: '/links' }),
  ]
}

function normalizeHeaderItems(items: SiteChromeConfig['header'] | undefined): SiteChromeItem[] {
  const normalized = Array.isArray(items) ? items.map(normalizeItem) : []
  const existing = new Set(normalized.map(item => item.id))
  return [
    ...defaultHeaderItems().filter(item => !existing.has(item.id)),
    ...normalized,
  ]
}

function displayChromeLabel(item: SiteChromeItem) {
  return item.i18nLabel ? t(item.i18nLabel) : item.label || adminText('k00dd')
}

function isSystemHeaderItem(item: SiteChromeItem) {
  return ['sponsors', 'links'].includes(item.id)
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const chrome = await getSiteChrome()
    config.value = normalize(chrome)
  } catch (err) {
    error.value = err instanceof Error ? err.message : adminText('k00db')
  } finally {
    loaded.value = true
    loading.value = false
  }
}

async function persist() {
  saving.value = true
  try {
    config.value = normalize(config.value)
    await saveSiteChrome(config.value)
    adminToast.success(adminText('k000e'))
  } catch (err) {
    adminToast.error(err, adminText('k000f'))
  } finally {
    saving.value = false
  }
}

function openItem(group: GroupKey, index: number | null) {
  dialog.value = { group, index }
  Object.assign(form, index === null ? emptyItem() : normalizeItem(config.value[group][index]))
}

function openSidebarGroupItem(groupIndex: number, index: number | null) {
  dialog.value = { group: 'sidebarGroup', groupIndex, index }
  const list = config.value.sidebarGroups[groupIndex]?.items || []
  Object.assign(form, index === null ? emptyItem() : normalizeItem(list[index]))
}

function saveItem() {
  if (!dialog.value) return
  const item = normalizeItem(form)
  const list = dialog.value.group === 'sidebarGroup'
    ? config.value.sidebarGroups[dialog.value.groupIndex].items
    : config.value[dialog.value.group]
  if (dialog.value.index === null) list.push(item)
  else list[dialog.value.index] = item
  dialog.value = null
}

function removeItem(group: GroupKey, index: number) {
  if (group === 'header' && isSystemHeaderItem(config.value.header[index])) {
    config.value.header[index].enabled = false
    return
  }
  config.value[group].splice(index, 1)
}

function toggleHeaderItemVisibility(index: number) {
  const item = config.value.header[index]
  if (!item) return
  item.enabled = !item.enabled
}

function removeSidebarGroupItem(groupIndex: number, index: number) {
  config.value.sidebarGroups[groupIndex]?.items.splice(index, 1)
}

function openGroup(index: number | null) {
  groupDialog.value = { index }
  Object.assign(groupForm, index === null ? emptyGroup() : normalizeGroup(config.value.sidebarGroups[index]))
}

function saveGroup() {
  if (!groupDialog.value) return
  const group = normalizeGroup(groupForm)
  if (groupDialog.value.index === null) config.value.sidebarGroups.push(group)
  else config.value.sidebarGroups[groupDialog.value.index] = group
  groupDialog.value = null
}

function removeGroup(index: number) {
  config.value.sidebarGroups.splice(index, 1)
}

function openBrandDialog() {
  Object.assign(brandForm, {
    brandType: config.value.brandType || 'default',
    brandText: config.value.brandText || '',
    brandImage: config.value.brandImage || '',
  })
  brandDialogOpen.value = true
}

function saveBrandDialog() {
  config.value.brandType = ['default', 'text', 'image'].includes(brandForm.brandType) ? brandForm.brandType : 'default'
  config.value.brandText = brandForm.brandText.trim()
  config.value.brandImage = brandForm.brandImage.trim()
  brandDialogOpen.value = false
}

async function uploadBrandImage(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  try {
    const data = await uploadAdminImage(file)
    brandForm.brandImage = data.url || ''
    adminToast.success(adminText('k000b'))
  } catch (err) {
    adminToast.error(err, adminText('k000c'))
  } finally {
    input.value = ''
  }
}

function ensureFooterInfo() {
  config.value.footerInfo ||= { primary: [], list: [] }
  return config.value.footerInfo
}

function footerLinkKey(item: { name: string, url: string }) {
  return `${ensureFooterInfo().list.indexOf(item)}-${item.name}-${item.url}`
}

function footerPrimaryKey(item: { content: string }) {
  return `${ensureFooterInfo().primary.indexOf(item)}-${item.content}`
}

function openFooterLink(index: number | null) {
  footerDialog.value = { kind: 'link', index }
  const link = index === null ? { name: '', url: '' } : ensureFooterInfo().list[index]
  Object.assign(footerLinkForm, { name: link?.name || '', url: link?.url || '' })
}

function openFooterPrimary(index: number | null) {
  footerDialog.value = { kind: 'primary', index }
  const item = index === null ? { content: '' } : ensureFooterInfo().primary[index]
  Object.assign(footerPrimaryForm, { content: item?.content || '' })
}

function saveFooterEntry() {
  if (!footerDialog.value) return
  const footer = ensureFooterInfo()
  if (footerDialog.value.kind === 'link') {
    const link = { name: footerLinkForm.name.trim(), url: footerLinkForm.url.trim() }
    if (footerDialog.value.index === null) footer.list.push(link)
    else footer.list[footerDialog.value.index] = link
  } else {
    const item = { content: footerPrimaryForm.content.trim() }
    if (footerDialog.value.index === null) footer.primary.push(item)
    else footer.primary[footerDialog.value.index] = item
  }
  footerDialog.value = null
}

function removeFooterLink(index: number) {
  ensureFooterInfo().list.splice(index, 1)
}

function removeFooterPrimary(index: number) {
  ensureFooterInfo().primary.splice(index, 1)
}

onMounted(load)
</script>

<template>
  <BasicPage :title="adminText('k00d9')" :description="adminText('k00da')">
    <template #actions>
      <Button type="button" size="sm" :disabled="saving || loading" @click="persist">
        <Save class="mr-2 h-4 w-4" />
        {{ saving ? adminText('k005f') : adminText('k0061') }}
      </Button>
    </template>

    <div v-if="error" class="rounded-md border border-destructive/30 bg-destructive/5 px-3 py-2 text-sm text-destructive">{{ error }}</div>
    <div v-else-if="loading && !loaded" class="rounded-md border bg-card px-3 py-8 text-center text-sm text-muted-foreground">{{ adminText('k0046') }}</div>

    <section v-else class="min-h-[760px] overflow-hidden rounded-xl border bg-base-200 text-base-content shadow-sm">
      <header class="border-b border-transparent bg-base-100/0 shadow-none backdrop-blur-none">
        <div class="mx-auto grid h-16 w-full max-w-[1600px] grid-cols-[auto_minmax(0,1fr)] items-center gap-8 px-5">
          <button
            type="button"
            class="group relative -ml-1 flex min-w-0 shrink-0 items-center gap-2 rounded-md border border-dashed border-transparent px-2 py-1 text-left transition hover:border-primary/25 hover:bg-primary/5"
            @click="openBrandDialog"
          >
            <img v-if="brandType === 'image' && brandImage" :src="brandImage" :alt="brandText" class="h-8 w-auto max-w-40 shrink-0 object-contain sm:h-9" />
            <span v-else-if="brandType === 'text'" class="max-w-44 truncate text-xl font-semibold tracking-tighter text-primary sm:text-2xl md:max-w-none">{{ brandText }}</span>
            <span v-else class="max-w-44 truncate text-xl font-semibold tracking-tighter sm:text-2xl md:max-w-none">
              <span style="color: oklch(54.6% 0.215 262.88)">Goose</span><span style="color: oklch(12.9% 0.042 264.695)">Forum</span>
            </span>
            <span class="ml-1 inline-flex h-5 w-5 items-center justify-center rounded-sm text-icon-muted opacity-0 transition hover:bg-primary/10 hover:text-primary group-hover:opacity-100">
              <Pencil class="h-3.5 w-3.5" />
            </span>
          </button>

          <nav class="flex min-w-0 items-center gap-1" aria-label="Header navigation preview">
            <Draggable
              v-model="config.header"
              item-key="id"
              handle=".chrome-drag-item"
              class="flex min-w-0 items-center gap-1"
            >
              <template #item="{ element, index }">
                <div
                  class="chrome-drag-item group relative inline-flex h-7 max-w-56 cursor-grab items-center rounded-md px-2 text-sm font-medium text-base-content/75 transition-colors duration-150 active:cursor-grabbing"
                  :class="element.enabled ? 'hover:bg-base-300 hover:text-base-content' : 'opacity-40'"
                >
                  <span class="min-w-0 truncate">{{ displayChromeLabel(element) }}</span>
                  <span class="pointer-events-none absolute left-1/2 top-full z-10 flex -translate-x-1/2 shrink-0 items-center gap-0.5 rounded-md bg-base-100/80 px-0.5 py-0.5 opacity-0 backdrop-blur transition-opacity duration-150 before:absolute before:inset-x-0 before:-top-2 before:h-2 group-hover:pointer-events-auto group-hover:opacity-100">
                    <button v-if="isSystemHeaderItem(element)" type="button" class="relative inline-flex h-5 w-5 items-center justify-center rounded-sm text-icon-muted transition hover:scale-110 hover:bg-primary/10 hover:text-primary" :title="adminText('k00dc')" @click.stop="toggleHeaderItemVisibility(index)">
                      <EyeOff v-if="element.enabled" class="h-3.5 w-3.5" />
                      <Eye v-else class="h-3.5 w-3.5" />
                    </button>
                    <button v-if="!isSystemHeaderItem(element)" type="button" class="relative inline-flex h-5 w-5 items-center justify-center rounded-sm text-icon-muted transition hover:scale-110 hover:bg-primary/10 hover:text-primary" @click.stop="openItem('header', index)">
                      <Pencil class="h-3.5 w-3.5" />
                    </button>
                    <button v-if="!isSystemHeaderItem(element)" type="button" class="relative inline-flex h-5 w-5 items-center justify-center rounded-sm text-icon-muted transition hover:scale-110 hover:bg-destructive/10 hover:text-destructive" @click.stop="removeItem('header', index)">
                      <Trash2 class="h-3.5 w-3.5" />
                    </button>
                  </span>
                </div>
              </template>
              <template #footer>
                <Button variant="ghost" size="sm" type="button" :class="['h-7 px-2 text-sm', addButtonClass]" @click="openItem('header', null)">
                  <Plus class="mr-1 h-3.5 w-3.5" />{{ adminText('k00de') }}
                </Button>
              </template>
            </Draggable>
          </nav>
        </div>
      </header>

      <div class="mx-auto grid w-full max-w-[1600px] grid-cols-[224px_minmax(0,1fr)] gap-3 px-5 py-3">
        <aside class="-my-3 min-w-0 self-start">
          <nav class="py-3">
            <div class="space-y-0.5 pb-2">
              <div
                v-for="item in fixedMainItems"
                :key="item.key"
                class="flex h-8 items-center gap-2 rounded-md px-2 text-[13px] font-medium transition-colors duration-150"
                :class="item.active ? 'bg-info/10 text-primary' : 'text-base-content/75'"
                :title="adminText('k00df')"
              >
                <component :is="item.icon" class="h-4 w-4 shrink-0" />
                <span class="min-w-0 flex-1 truncate">{{ t(item.labelKey) }}</span>
              </div>

              <Draggable v-model="config.mainMenu" item-key="id" handle=".chrome-drag-item" class="space-y-0.5">
                <template #item="{ element, index }">
                  <div class="chrome-drag-item group relative flex h-8 cursor-grab items-center gap-2 rounded-md px-2 text-[13px] font-medium text-base-content/75 transition-colors duration-150 hover:bg-base-300 hover:text-base-content active:cursor-grabbing" :class="{ 'opacity-40': !element.enabled }">
                    <LinkIcon class="h-4 w-4 shrink-0 opacity-80" />
                    <span class="min-w-0 flex-1 truncate">{{ element.label || adminText('k00dd') }}</span>
                    <span class="absolute right-1 top-1/2 flex -translate-y-1/2 items-center gap-0.5 opacity-0 transition-opacity duration-150 group-hover:opacity-100">
                      <button type="button" class="inline-flex h-5 w-5 items-center justify-center rounded-sm bg-base-100/80 text-icon-muted backdrop-blur transition hover:scale-110 hover:bg-primary/10 hover:text-primary" @click.stop="openItem('mainMenu', index)">
                        <Pencil class="h-4 w-4" />
                      </button>
                      <button type="button" class="inline-flex h-5 w-5 items-center justify-center rounded-sm bg-base-100/80 text-icon-muted backdrop-blur transition hover:scale-110 hover:bg-destructive/10 hover:text-destructive" @click.stop="removeItem('mainMenu', index)">
                        <Trash2 class="h-4 w-4" />
                      </button>
                    </span>
                  </div>
                </template>
                <template #footer>
                  <Button variant="ghost" size="sm" type="button" :class="['mt-1 h-7 px-2 text-[13px]', addButtonClass]" @click="openItem('mainMenu', null)">
                    <Plus class="mr-1 h-4 w-4" />{{ adminText('k00dg') }}
                  </Button>
                </template>
              </Draggable>
            </div>

            <div class="mt-2">
              <div class="mb-1 px-2 text-[10px] font-bold uppercase tracking-wide text-base-content/55">{{ t('shell.resources') }}</div>
              <div class="space-y-px">
                <div
                  v-for="item in fixedResourceItems"
                  :key="item.key"
                  class="flex h-7 items-center gap-2 rounded-md px-2 text-[13px] font-medium text-base-content/75 transition-colors duration-150"
                  :title="adminText('k00dh')"
                >
                  <component :is="item.icon" class="h-4 w-4 shrink-0" />
                  <span class="truncate">{{ t(item.labelKey) }}</span>
                </div>
                <Draggable v-model="config.resources" item-key="id" handle=".chrome-drag-item" class="space-y-px">
                  <template #item="{ element, index }">
                    <div class="chrome-drag-item group relative flex h-7 cursor-grab items-center gap-2 rounded-md px-2 text-[13px] font-medium text-base-content/75 transition-colors duration-150 hover:bg-base-300 hover:text-base-content active:cursor-grabbing" :class="{ 'opacity-40': !element.enabled }">
                      <LinkIcon class="h-4 w-4 shrink-0 opacity-80" />
                      <span class="min-w-0 flex-1 truncate">{{ element.label || adminText('k00dd') }}</span>
                      <span class="absolute right-1 top-1/2 flex -translate-y-1/2 items-center gap-0.5 opacity-0 transition-opacity duration-150 group-hover:opacity-100">
                        <button type="button" class="inline-flex h-5 w-5 items-center justify-center rounded-sm bg-base-100/80 text-icon-muted backdrop-blur transition hover:scale-110 hover:bg-primary/10 hover:text-primary" @click.stop="openItem('resources', index)">
                          <Pencil class="h-4 w-4" />
                        </button>
                        <button type="button" class="inline-flex h-5 w-5 items-center justify-center rounded-sm bg-base-100/80 text-icon-muted backdrop-blur transition hover:scale-110 hover:bg-destructive/10 hover:text-destructive" @click.stop="removeItem('resources', index)">
                          <Trash2 class="h-4 w-4" />
                        </button>
                      </span>
                    </div>
                  </template>
                  <template #footer>
                    <Button variant="ghost" size="sm" type="button" :class="['mt-1 h-7 px-2 text-[13px]', addButtonClass]" @click="openItem('resources', null)">
                      <Plus class="mr-1 h-4 w-4" />{{ adminText('k00di') }}
                    </Button>
                  </template>
                </Draggable>
              </div>
            </div>

            <div
              v-for="(group, groupIndex) in config.sidebarGroups"
              :key="group.id"
              class="group mt-2"
            >
              <div class="mb-1 flex h-5 items-center px-2">
                <div class="min-w-0 flex-1 truncate text-[10px] font-bold uppercase tracking-wide text-base-content/55">{{ group.title }}</div>
                <div class="flex items-center gap-0.5 opacity-0 transition-opacity duration-150 group-hover:opacity-100">
                  <button type="button" class="inline-flex h-5 w-5 items-center justify-center rounded-sm text-icon-muted transition hover:scale-110 hover:bg-primary/10 hover:text-primary" @click="openGroup(groupIndex)">
                    <Pencil class="h-3.5 w-3.5" />
                  </button>
                  <button type="button" class="inline-flex h-5 w-5 items-center justify-center rounded-sm text-icon-muted transition hover:scale-110 hover:bg-destructive/10 hover:text-destructive" @click="removeGroup(groupIndex)">
                    <Trash2 class="h-3.5 w-3.5" />
                  </button>
                </div>
              </div>

              <Draggable v-model="group.items" item-key="id" handle=".chrome-drag-item" class="space-y-px">
                <template #item="{ element, index }">
                  <div class="chrome-drag-item group/item relative flex h-7 cursor-grab items-center gap-2 rounded-md px-2 text-[13px] font-medium text-base-content/75 transition-colors duration-150 hover:bg-base-300 hover:text-base-content active:cursor-grabbing" :class="{ 'opacity-40': !element.enabled }">
                    <LinkIcon class="h-4 w-4 shrink-0 opacity-80" />
                    <span class="min-w-0 flex-1 truncate">{{ element.label || adminText('k00dd') }}</span>
                    <span class="absolute right-1 top-1/2 flex -translate-y-1/2 items-center gap-0.5 opacity-0 transition-opacity duration-150 group-hover/item:opacity-100">
                      <button type="button" class="inline-flex h-5 w-5 items-center justify-center rounded-sm bg-base-100/80 text-icon-muted backdrop-blur transition hover:scale-110 hover:bg-primary/10 hover:text-primary" @click.stop="openSidebarGroupItem(groupIndex, index)">
                        <Pencil class="h-4 w-4" />
                      </button>
                      <button type="button" class="inline-flex h-5 w-5 items-center justify-center rounded-sm bg-base-100/80 text-icon-muted backdrop-blur transition hover:scale-110 hover:bg-destructive/10 hover:text-destructive" @click.stop="removeSidebarGroupItem(groupIndex, index)">
                        <Trash2 class="h-4 w-4" />
                      </button>
                    </span>
                  </div>
                </template>
                <template #footer>
                  <Button variant="ghost" size="sm" type="button" :class="['mt-1 h-7 px-2 text-[13px]', addButtonClass]" @click="openSidebarGroupItem(groupIndex, null)">
                    <Plus class="mr-1 h-4 w-4" />{{ adminText('k00de') }}
                  </Button>
                </template>
              </Draggable>
            </div>

            <Button variant="ghost" size="sm" type="button" :class="['mt-2 h-7 w-full justify-start px-2 text-[13px]', addButtonClass]" @click="openGroup(null)">
              <Plus class="mr-1 h-4 w-4" />{{ adminText('k00dj') }}
            </Button>

            <div class="mt-2">
              <div class="mb-1 px-2 text-[10px] font-bold uppercase tracking-wide text-base-content/55">{{ t('shell.categories') }}</div>
              <div v-for="category in categories" :key="category.id" class="flex h-7 items-center gap-2 rounded-md px-2 text-[13px] font-medium text-base-content/75 transition-colors duration-150" :title="adminText('k00dk')">
                <span class="h-2 w-2 rounded-[3px]" :style="{ backgroundColor: category.color }" />
                <span class="truncate">{{ category.label }}</span>
              </div>
              <div v-if="!categories.length" class="px-2 py-1 text-[13px] text-base-content/55">{{ adminText('k00dl') }}</div>
            </div>

            <footer
              class="-mx-1 mt-0 cursor-pointer rounded-md border border-dashed border-transparent px-3 py-1 text-xs leading-5 text-base-content/75 transition hover:border-primary/25 hover:bg-primary/5"
              @click="footerManageOpen = true"
            >
              <div v-if="ensureFooterInfo().list.length" class="flex flex-wrap items-center gap-x-3 gap-y-0.5">
                <span
                  v-for="link in ensureFooterInfo().list"
                  :key="`${link.name}-${link.url}`"
                  class="inline-flex min-h-5 items-center rounded"
                >
                  {{ link.name || adminText('k00dd') }}
                </span>
              </div>
              <div v-if="ensureFooterInfo().primary.length" class="mt-1 flex flex-wrap items-center gap-x-3 gap-y-0.5 text-base-content/75">
                <span
                  v-for="item in ensureFooterInfo().primary"
                  :key="item.content"
                  class="inline-flex min-h-5 items-center rounded"
                >
                  {{ item.content || adminText('k00dd') }}
                </span>
              </div>
            </footer>
          </nav>
        </aside>

        <main class="min-w-0 overflow-hidden rounded-lg border border-line bg-base-100">
          <div class="flex h-14 items-center gap-2 border-b border-line px-4">
            <span class="rounded-lg bg-base-content px-4 py-2 text-sm font-semibold text-base-100">{{ t('topicList.tabs.latest') }}</span>
            <span class="px-3 py-2 text-sm font-medium text-base-content/60">{{ t('topicList.tabs.hot') }}</span>
            <span class="px-3 py-2 text-sm font-medium text-base-content/60">{{ t('topicList.tabs.popular') }}</span>
          </div>
          <div class="space-y-0 divide-y divide-line px-4">
            <div v-for="index in 7" :key="index" class="py-4">
              <div class="h-4 w-2/3 rounded bg-base-content/90" />
              <div class="mt-3 h-3.5 w-4/5 rounded bg-base-content/20" />
            </div>
          </div>
        </main>
      </div>
    </section>

    <Dialog :open="brandDialogOpen" @update:open="brandDialogOpen = $event">
      <DialogContent class="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{{ adminText('k0085') }}</DialogTitle>
          <DialogDescription>{{ adminText('k00ec') }}</DialogDescription>
        </DialogHeader>

        <form class="grid gap-4" @submit.prevent="saveBrandDialog">
          <div class="flex flex-wrap gap-4 text-sm">
            <label class="flex items-center gap-2">
              <input v-model="brandForm.brandType" type="radio" value="default" />
              {{ adminText('k0086') }}
            </label>
            <label class="flex items-center gap-2">
              <input v-model="brandForm.brandType" type="radio" value="text" />
              {{ adminText('k0087') }}
            </label>
            <label class="flex items-center gap-2">
              <input v-model="brandForm.brandType" type="radio" value="image" />
              {{ adminText('k0088') }}
            </label>
          </div>

          <label v-if="brandForm.brandType === 'default'" class="grid gap-1.5 text-sm">
            <span class="font-medium">{{ adminText('k0086') }}</span>
            <Input model-value="GooseForum" disabled />
          </label>

          <label v-if="brandForm.brandType === 'text'" class="grid gap-1.5 text-sm">
            <span class="font-medium">{{ adminText('k0089') }}</span>
            <Input v-model="brandForm.brandText" placeholder="GooseForum" />
          </label>

          <label v-if="brandForm.brandType === 'image'" class="grid gap-1.5 text-sm">
            <span class="font-medium">{{ adminText('k008a') }}</span>
            <div class="flex gap-2">
              <Input v-model="brandForm.brandImage" placeholder="Brand Image URL" />
              <Button variant="outline" type="button" class="relative shrink-0 overflow-hidden">
                <Upload class="mr-2 h-4 w-4" />{{ adminText('k0084') }}
                <input type="file" accept="image/*" class="absolute inset-0 cursor-pointer opacity-0" @change="uploadBrandImage" />
              </Button>
            </div>
          </label>

          <DialogFooter>
            <Button variant="outline" type="button" @click="brandDialogOpen = false">{{ adminText('k009q') }}</Button>
            <Button type="submit">{{ adminText('k005g') }}</Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>

    <Dialog :open="dialog !== null" @update:open="(open) => !open && (dialog = null)">
      <DialogContent class="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{{ dialog?.index === null ? adminText('k00do') : adminText('k00dp') }}</DialogTitle>
          <DialogDescription>{{ adminText('k00dq') }}</DialogDescription>
        </DialogHeader>

        <form class="grid gap-4" @submit.prevent="saveItem">
          <label class="grid gap-1.5 text-sm">
            <span class="font-medium">{{ adminText('k00dr') }}</span>
            <select v-model="form.type" class="h-9 rounded-md border bg-background px-3 text-sm">
              <option value="link">{{ adminText('k00ds') }}</option>
              <option value="text">{{ adminText('k00dt') }}</option>
            </select>
          </label>
          <label class="grid gap-1.5 text-sm">
            <span class="font-medium">{{ adminText('k00dv') }}</span>
            <Input v-model="form.label" :placeholder="adminText('k00dw')" />
          </label>
          <label class="grid gap-1.5 text-sm">
            <span class="font-medium">{{ adminText('k00dm') }}</span>
            <Input v-model="form.url" :placeholder="adminText('k00eg')" />
          </label>
          <label class="flex items-center justify-between rounded-md border px-3 py-2 text-sm">
            <span class="font-medium">{{ adminText('k00e2') }}</span>
            <Switch v-model="form.enabled" />
          </label>
          <DialogFooter>
            <Button variant="outline" type="button" @click="dialog = null">{{ adminText('k009q') }}</Button>
            <Button type="submit">{{ adminText('k00e3') }}</Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>

    <Dialog :open="groupDialog !== null" @update:open="(open) => !open && (groupDialog = null)">
      <DialogContent class="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{{ groupDialog?.index === null ? adminText('k00e4') : adminText('k00e5') }}</DialogTitle>
          <DialogDescription>{{ adminText('k00e6') }}</DialogDescription>
        </DialogHeader>

        <form class="grid gap-4" @submit.prevent="saveGroup">
          <label class="grid gap-1.5 text-sm">
            <span class="font-medium">{{ adminText('k00e7') }}</span>
            <Input v-model="groupForm.title" :placeholder="adminText('k00e8')" />
          </label>
          <DialogFooter>
            <Button variant="outline" type="button" @click="groupDialog = null">{{ adminText('k009q') }}</Button>
            <Button type="submit">{{ adminText('k00e9') }}</Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>

    <Dialog :open="footerManageOpen" @update:open="footerManageOpen = $event">
      <DialogContent class="gap-3 p-4 sm:max-w-3xl">
        <DialogHeader>
          <DialogTitle>{{ adminText('k00eb') }}</DialogTitle>
          <DialogDescription>{{ adminText('k00ec') }}</DialogDescription>
        </DialogHeader>

        <div class="rounded-md border border-dashed border-line/70 bg-base-100 px-4 py-3">
          <footer class="text-xs leading-5 text-base-content/75">
            <Draggable
              v-model="ensureFooterInfo().list"
              :item-key="footerLinkKey"
              handle=".footer-manage-drag"
              class="flex flex-wrap items-center gap-x-3 gap-y-0.5"
            >
              <template #item="{ element, index }">
                <span class="group inline-flex min-h-6 items-center gap-0.5 rounded-md transition-colors hover:bg-base-300/80">
                  <span class="footer-manage-drag cursor-grab rounded px-1 text-base-content/35 opacity-0 transition group-hover:opacity-100 active:cursor-grabbing">⋮⋮</span>
                  <button type="button" class="max-w-40 truncate rounded px-1 text-left transition hover:text-base-content" @click="openFooterLink(index)">
                    {{ element.name || adminText('k00dd') }}
                  </button>
                  <span class="flex items-center gap-0.5 opacity-0 transition-opacity group-hover:opacity-100">
                    <button type="button" class="inline-flex h-5 w-5 items-center justify-center rounded-sm text-icon-muted transition hover:scale-110 hover:bg-primary/10 hover:text-primary" @click.stop="openFooterLink(index)">
                      <Pencil class="h-3.5 w-3.5" />
                    </button>
                    <button type="button" class="inline-flex h-5 w-5 items-center justify-center rounded-sm text-icon-muted transition hover:scale-110 hover:bg-destructive/10 hover:text-destructive" @click.stop="removeFooterLink(index)">
                      <Trash2 class="h-3.5 w-3.5" />
                    </button>
                  </span>
                </span>
              </template>
              <template #footer>
                <Button variant="ghost" size="sm" type="button" :class="['h-6 px-2 text-xs', addButtonClass]" @click="openFooterLink(null)">
                  <Plus class="mr-1 h-3.5 w-3.5" />{{ adminText('k00de') }}
                </Button>
                <span v-if="!ensureFooterInfo().list.length" class="text-xs text-base-content/45">{{ adminText('k008i') }}</span>
              </template>
            </Draggable>

            <Draggable
              v-model="ensureFooterInfo().primary"
              :item-key="footerPrimaryKey"
              handle=".footer-manage-drag"
              class="mt-1 flex flex-wrap items-center gap-x-3 gap-y-0.5"
            >
              <template #item="{ element, index }">
                <span class="group inline-flex min-h-6 items-center gap-0.5 rounded-md transition-colors hover:bg-base-300/80">
                  <span class="footer-manage-drag cursor-grab rounded px-1 text-base-content/35 opacity-0 transition group-hover:opacity-100 active:cursor-grabbing">⋮⋮</span>
                  <button type="button" class="max-w-72 truncate rounded px-1 text-left transition hover:text-base-content" @click="openFooterPrimary(index)">
                    {{ element.content || adminText('k00dd') }}
                  </button>
                  <span class="flex items-center gap-0.5 opacity-0 transition-opacity group-hover:opacity-100">
                    <button type="button" class="inline-flex h-5 w-5 items-center justify-center rounded-sm text-icon-muted transition hover:scale-110 hover:bg-primary/10 hover:text-primary" @click.stop="openFooterPrimary(index)">
                      <Pencil class="h-3.5 w-3.5" />
                    </button>
                    <button type="button" class="inline-flex h-5 w-5 items-center justify-center rounded-sm text-icon-muted transition hover:scale-110 hover:bg-destructive/10 hover:text-destructive" @click.stop="removeFooterPrimary(index)">
                      <Trash2 class="h-3.5 w-3.5" />
                    </button>
                  </span>
                </span>
              </template>
              <template #footer>
                <Button variant="ghost" size="sm" type="button" :class="['h-6 px-2 text-xs', addButtonClass]" @click="openFooterPrimary(null)">
                  <Plus class="mr-1 h-3.5 w-3.5" />{{ adminText('k00dn') }}
                </Button>
                <span v-if="!ensureFooterInfo().primary.length" class="text-xs text-base-content/45">{{ adminText('k0062') }}</span>
              </template>
            </Draggable>
          </footer>
        </div>

        <DialogFooter>
          <Button type="button" @click="footerManageOpen = false">{{ adminText('k00b8') }}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <Dialog :open="footerDialog !== null" @update:open="(open) => !open && (footerDialog = null)">
      <DialogContent class="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{{ footerDialog?.kind === 'primary' ? adminText('k00dn') : adminText('k00dm') }}</DialogTitle>
          <DialogDescription>{{ adminText('k00ec') }}</DialogDescription>
        </DialogHeader>

        <form v-if="footerDialog?.kind === 'link'" class="grid gap-4" @submit.prevent="saveFooterEntry">
          <label class="grid gap-1.5 text-sm">
            <span class="font-medium">{{ adminText('k00ed') }}</span>
            <Input v-model="footerLinkForm.name" :placeholder="adminText('k00ee')" />
          </label>
          <label class="grid gap-1.5 text-sm">
            <span class="font-medium">{{ adminText('k00dm') }}</span>
            <Input v-model="footerLinkForm.url" :placeholder="adminText('k00eh')" />
          </label>
          <DialogFooter>
            <Button variant="outline" type="button" @click="footerDialog = null">{{ adminText('k009q') }}</Button>
            <Button type="submit">{{ adminText('k00ef') }}</Button>
          </DialogFooter>
        </form>

        <form v-else class="grid gap-4" @submit.prevent="saveFooterEntry">
          <label class="grid gap-1.5 text-sm">
            <span class="font-medium">{{ adminText('k00dn') }}</span>
            <Input v-model="footerPrimaryForm.content" :placeholder="adminText('k00dn')" />
          </label>
          <DialogFooter>
            <Button variant="outline" type="button" @click="footerDialog = null">{{ adminText('k009q') }}</Button>
            <Button type="submit">{{ adminText('k00ef') }}</Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>

  </BasicPage>
</template>
