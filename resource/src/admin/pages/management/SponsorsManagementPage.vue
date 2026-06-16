<script setup lang="ts">
import { adminText } from '@/admin/runtime/i18n-text'

import { computed, onMounted, reactive, ref } from 'vue'
import Draggable from 'vuedraggable'
import { GripVertical, Loader2, PenLine, Plus, RefreshCw, Save, Trash2, Upload } from '@lucide/vue'
import AdminActionButton from '@/admin/components/AdminActionButton.vue'
import { BasicPage } from '@/admin/components/global-layout'
import { Button } from '@/admin/components/ui/button'
import { Input } from '@/admin/components/ui/input'
import { Textarea } from '@/admin/components/ui/textarea'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/admin/components/ui/dialog'
import { getSponsors, saveSponsors } from '@/admin/runtime/api'
import { adminToast } from '@/admin/runtime/toast'
import { resolveApiMessage } from '@/runtime/api-message'
import type { AdminPayload, ManageHomeProps, SponsorItem, SponsorsConfig } from '@/admin/types'

defineProps<{
  payload: AdminPayload<ManageHomeProps>
}>()

type SponsorLevel = keyof SponsorsConfig['sponsors']

const levelNames: Record<SponsorLevel, string> = {
  level0: adminText('k001v'),
  level1: adminText('k001w'),
  level2: adminText('k001x'),
  level3: adminText('k001y'),
}

const levelMeta: Record<SponsorLevel, { level: string, tone: 'diamond' | 'gold' | 'silver' | 'bronze' }> = {
  level0: { level: 'Level 0', tone: 'diamond' },
  level1: { level: 'Level 1', tone: 'gold' },
  level2: { level: 'Level 2', tone: 'silver' },
  level3: { level: 'Level 3', tone: 'bronze' },
}

const defaultConfig: SponsorsConfig = {
  sponsors: {
    level0: [],
    level1: [],
    level2: [],
    level3: [],
  },
  content: {
    title: adminText('k001z'),
    description: adminText('k0020'),
  },
  contact: {
    title: adminText('k0021'),
    description: adminText('k0022'),
    buttonText: adminText('k0023'),
    buttonLink: 'mailto:contact@gooseforum.online',
  },
  rules: [
    { content: adminText('k0024') },
    { content: adminText('k0025') },
    { content: adminText('k0026') },
  ],
}

const loading = ref(false)
const saving = ref(false)
const error = ref('')
const config = ref<SponsorsConfig>(structuredClone(defaultConfig))
const sponsorDialog = ref<{ level: SponsorLevel, index: number | null } | null>(null)
const sponsorForm = reactive<SponsorItem>({ name: '', avatarUrl: '', message: '', link: '' })

const totalSponsors = computed(() => Object.values(config.value.sponsors).reduce((total, list) => total + list.length, 0))

function sectionGrid(level: SponsorLevel) {
  if (levelMeta[level].tone === 'diamond') return 'grid-cols-1 md:grid-cols-2'
  if (levelMeta[level].tone === 'gold') return 'grid-cols-1 sm:grid-cols-2 xl:grid-cols-3'
  return 'grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 2xl:grid-cols-5'
}

function sectionBadgeClass(level: SponsorLevel) {
  const tone = levelMeta[level].tone
  if (tone === 'diamond') return 'bg-blue-50 text-blue-700'
  if (tone === 'gold') return 'bg-amber-50 text-amber-700'
  if (tone === 'silver') return 'bg-gray-100 text-gray-600'
  return 'bg-rose-50 text-rose-700'
}

function sponsorCardClass(level: SponsorLevel) {
  const tone = levelMeta[level].tone
  if (tone === 'diamond') return 'p-4'
  if (tone === 'gold') return 'p-3'
  return 'p-2.5'
}

function normalize(value?: Partial<SponsorsConfig>): SponsorsConfig {
  return {
    sponsors: {
      level0: value?.sponsors?.level0 ?? [],
      level1: value?.sponsors?.level1 ?? [],
      level2: value?.sponsors?.level2 ?? [],
      level3: value?.sponsors?.level3 ?? [],
    },
    content: {
      title: value?.content?.title || defaultConfig.content.title,
      description: value?.content?.description || defaultConfig.content.description,
    },
    contact: {
      title: value?.contact?.title || defaultConfig.contact.title,
      description: value?.contact?.description || defaultConfig.contact.description,
      buttonText: value?.contact?.buttonText || defaultConfig.contact.buttonText,
      buttonLink: value?.contact?.buttonLink || defaultConfig.contact.buttonLink,
    },
    rules: value?.rules ?? defaultConfig.rules,
  }
}

async function loadSponsors() {
  loading.value = true
  error.value = ''
  try {
    config.value = normalize(await getSponsors())
  } catch (err) {
    error.value = err instanceof Error ? err.message : adminText('k0027')
  } finally {
    loading.value = false
  }
}

async function persist() {
  saving.value = true
  try {
    await saveSponsors(normalize(config.value))
    adminToast.success(adminText('k000e'))
  } catch (err) {
    adminToast.error(err, adminText('k001b'))
  } finally {
    saving.value = false
  }
}

function openSponsor(level: SponsorLevel, index: number | null) {
  const current = index === null ? { name: '', avatarUrl: '', message: '', link: '' } : config.value.sponsors[level][index]
  Object.assign(sponsorForm, current)
  sponsorDialog.value = { level, index }
}

function submitSponsor() {
  if (!sponsorDialog.value) return
  if (!sponsorForm.name.trim()) {
    adminToast.warning(adminText('k0028'))
    return
  }
  const item = {
    name: sponsorForm.name.trim(),
    avatarUrl: sponsorForm.avatarUrl.trim(),
    message: sponsorForm.message.trim(),
    link: sponsorForm.link.trim(),
  }
  const list = config.value.sponsors[sponsorDialog.value.level]
  if (sponsorDialog.value.index === null) list.push(item)
  else list[sponsorDialog.value.index] = item
  sponsorDialog.value = null
  adminToast.success(adminText('k0029'))
}

function removeSponsor(level: SponsorLevel, index: number) {
  config.value.sponsors[level].splice(index, 1)
}

function addRule() {
  config.value.rules.push({ content: '' })
}

function removeRule(index: number) {
  config.value.rules.splice(index, 1)
}

async function uploadAvatar(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  const body = new FormData()
  body.append('file', file)
  try {
    const response = await fetch('/file/img-upload', { method: 'POST', body })
    const data = await response.json()
    if (data.code === 0) {
      sponsorForm.avatarUrl = data.result?.url || ''
      adminToast.success(adminText('k000b'))
    } else {
      adminToast.error(new Error(resolveApiMessage(data, adminText('k000c'))), adminText('k000c'))
    }
  } catch (err) {
    adminToast.error(err, adminText('k000c'))
  } finally {
    input.value = ''
  }
}

onMounted(() => {
  void loadSponsors()
})
</script>

<template>
  <BasicPage :title="adminText('k004o')" :description="adminText('k004p')" sticky>
    <template #actions>
      <div class="flex items-center gap-2">
        <Button variant="outline" type="button" @click="loadSponsors">
          <RefreshCw class="size-4" />
          {{ adminText('k004q') }}
        </Button>
        <Button type="button" :disabled="saving" @click="persist">
          <Loader2 v-if="saving" class="size-4 animate-spin" />
          <Save v-else class="size-4" />
          {{ adminText('k004f') }}
        </Button>
      </div>
    </template>

      <div v-if="loading" class="flex h-64 items-center justify-center rounded-lg border">
        <Loader2 class="size-8 animate-spin text-primary" />
      </div>
      <div v-else-if="error" class="rounded-lg border border-destructive/30 bg-destructive/5 p-4 text-sm text-destructive">{{ error }}</div>
      <div v-else class="grid gap-4 xl:grid-cols-[minmax(0,1fr)_280px]">
        <div class="space-y-5">
          <section class="border-b pb-4">
            <div class="flex flex-wrap items-center gap-2">
              <div class="group relative min-w-[180px] flex-1 rounded-md transition-colors hover:bg-muted/50 focus-within:bg-muted/50">
                <Input v-model="config.content.title" :aria-label="adminText('k004u')" class="h-auto border-0 bg-transparent px-2 py-1 pr-9 text-2xl font-bold tracking-tight shadow-none focus-visible:ring-1" />
                <PenLine class="pointer-events-none absolute right-2 top-1/2 size-4 -translate-y-1/2 text-muted-foreground/40 opacity-0 transition-opacity group-hover:opacity-100 group-focus-within:opacity-100" />
              </div>
              <span class="rounded-full bg-muted px-2 py-0.5 text-xs font-semibold text-muted-foreground">{{ totalSponsors }}</span>
            </div>
            <div class="group relative mt-2 rounded-md transition-colors hover:bg-muted/50 focus-within:bg-muted/50">
              <Input v-model="config.content.description" :aria-label="adminText('k004v')" class="h-10 border-0 bg-transparent px-2 pr-9 text-sm text-muted-foreground shadow-none focus-visible:ring-1" />
              <PenLine class="pointer-events-none absolute right-2 top-1/2 size-4 -translate-y-1/2 text-muted-foreground/40 opacity-0 transition-opacity group-hover:opacity-100 group-focus-within:opacity-100" />
            </div>
          </section>

          <section v-for="(title, level) in levelNames" :key="level" class="space-y-3">
            <div class="flex items-center justify-between gap-3 border-b pb-2">
              <h2 class="flex min-w-0 items-center gap-2 text-base font-bold">
                <span class="rounded px-1.5 py-0.5 text-[11px] font-semibold" :class="sectionBadgeClass(level)">{{ title }}</span>
                <span class="text-xs font-semibold text-muted-foreground">{{ levelMeta[level].level }}</span>
              </h2>
              <Button variant="ghost" type="button" @click="openSponsor(level, null)">
                <Plus class="size-4" />
                {{ adminText('k0094') }}
              </Button>
            </div>
            <Draggable
              v-model="config.sponsors[level]"
              item-key="name"
              :group="{ name: 'sponsors' }"
              handle=".js-sponsor-handle"
              ghost-class="opacity-40"
              chosen-class="opacity-80"
              class="grid min-h-16 gap-3 rounded-lg border border-dashed border-transparent"
              :class="[sectionGrid(level), config.sponsors[level].length === 0 && 'border-muted-foreground/20 p-5']"
            >
              <template #item="{ element: sponsor, index }">
                <article class="group relative rounded-lg border bg-card transition-colors hover:border-primary/40" :class="sponsorCardClass(level)">
                  <GripVertical class="js-sponsor-handle absolute left-1.5 top-1.5 size-3.5 cursor-grab text-muted-foreground/30 opacity-0 transition-opacity group-hover:opacity-100 active:cursor-grabbing" />
                  <div class="flex items-start gap-3">
                    <img v-if="sponsor.avatarUrl" :src="sponsor.avatarUrl" class="size-11 shrink-0 rounded-md border object-cover" alt="" />
                    <span v-else class="grid size-11 shrink-0 place-items-center rounded-md border bg-muted text-sm font-semibold">{{ sponsor.name.slice(0, 1) }}</span>
                    <div class="min-w-0 flex-1">
                      <div class="flex min-w-0 items-center gap-2">
                        <a :href="sponsor.link || '#'" target="_blank" rel="noreferrer" class="truncate text-sm font-semibold hover:text-primary hover:underline">{{ sponsor.name }}</a>
                      </div>
                      <p class="mt-1 line-clamp-2 text-xs leading-5 text-muted-foreground">{{ sponsor.message || adminText('k004r') }}</p>
                    </div>
                  </div>
                  <div class="absolute right-1.5 top-1.5 flex shrink-0 items-center gap-1 rounded-md bg-background/90 p-0.5 opacity-0 shadow-sm ring-1 ring-border transition-opacity group-hover:opacity-100">
                    <AdminActionButton compact :title="adminText('k005j')" @click="openSponsor(level, index)">
                      <PenLine class="size-4" />
                    </AdminActionButton>
                    <AdminActionButton compact tone="danger" :title="adminText('k005i')" @click="removeSponsor(level, index)">
                      <Trash2 class="size-4" />
                    </AdminActionButton>
                  </div>
                </article>
              </template>
              <template #footer>
                <button
                  v-if="config.sponsors[level].length === 0"
                  type="button"
                  class="text-sm text-muted-foreground transition-colors hover:text-foreground"
                  @click="openSponsor(level, null)"
                >
                  {{ adminText('k009s') }}
                </button>
              </template>
            </Draggable>
          </section>
        </div>

        <aside class="space-y-3">
          <div class="rounded-lg border bg-background p-4">
            <h2 class="text-sm font-semibold text-foreground">{{ adminText('k009t') }}</h2>
            <div class="mt-4 space-y-3">
              <Input v-model="config.contact.title" class="border-0 bg-transparent px-2 text-sm font-semibold shadow-none focus-visible:ring-1" :aria-label="adminText('k004w')" />
              <Textarea v-model="config.contact.description" class="min-h-20 resize-none border-0 bg-transparent px-2 py-2 text-sm leading-6 text-muted-foreground shadow-none focus-visible:ring-1" :aria-label="adminText('k004x')" />
              <label class="grid gap-1 text-[11px] font-semibold uppercase text-muted-foreground">
                {{ adminText('k009u') }}
                <Input v-model="config.contact.buttonText" class="border-0 bg-transparent px-2 text-sm shadow-none focus-visible:ring-1" />
              </label>
              <label class="grid gap-1 text-[11px] font-semibold uppercase text-muted-foreground">
                {{ adminText('k009v') }}
                <Input v-model="config.contact.buttonLink" class="border-0 bg-transparent px-2 text-sm shadow-none focus-visible:ring-1" placeholder="mailto:contact@example.com" />
              </label>
            </div>
          </div>
          <div class="rounded-lg border bg-background p-4">
            <div class="flex items-center justify-between gap-2">
              <h2 class="text-sm font-semibold text-foreground">{{ adminText('k009w') }}</h2>
              <Button type="button" size="sm" variant="outline" @click="addRule">{{ adminText('k0094') }}</Button>
            </div>
            <div class="mt-4 space-y-2">
              <p v-if="config.rules.length === 0" class="text-sm text-muted-foreground">{{ adminText('k009x') }}</p>
              <div v-for="(rule, index) in config.rules" :key="index" class="flex gap-2">
                <Input v-model="rule.content" class="border-0 bg-transparent px-2 text-sm shadow-none focus-visible:ring-1" :aria-label="adminText('k004y', { index: index + 1 })" />
                <AdminActionButton tone="danger" @click="removeRule(index)">
                  <Trash2 class="size-3.5" />
                  {{ adminText('k005i') }}
                </AdminActionButton>
              </div>
            </div>
          </div>
        </aside>
      </div>

      <Dialog :open="sponsorDialog !== null" @update:open="(open) => !open && (sponsorDialog = null)">
        <DialogContent class="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>{{ sponsorDialog?.index === null ? adminText('k004s') : adminText('k004t') }}</DialogTitle>
            <DialogDescription>{{ adminText('k009y') }}</DialogDescription>
          </DialogHeader>
          <form class="grid gap-4" @submit.prevent="submitSponsor">
            <label class="grid gap-2 text-sm font-medium">{{ adminText('k009z') }}<Input v-model="sponsorForm.name" :placeholder="adminText('k004z')" /></label>
            <label class="grid gap-2 text-sm font-medium">
              Logo
              <div class="flex gap-2">
                <Input v-model="sponsorForm.avatarUrl" :placeholder="adminText('k0050')" />
                <label class="relative inline-flex size-9 shrink-0 cursor-pointer items-center justify-center rounded-md border bg-background shadow-xs hover:bg-accent">
                  <Upload class="size-4" />
                  <input class="absolute inset-0 cursor-pointer opacity-0" type="file" accept="image/*" @change="uploadAvatar" />
                </label>
              </div>
              <img v-if="sponsorForm.avatarUrl" :src="sponsorForm.avatarUrl" class="h-16 w-auto rounded border object-contain" :alt="adminText('k0051')" />
            </label>
            <label class="grid gap-2 text-sm font-medium">{{ adminText('k00a0') }}<Textarea v-model="sponsorForm.message" :placeholder="adminText('k0052')" /></label>
            <label class="grid gap-2 text-sm font-medium">{{ adminText('k00a1') }}<Input v-model="sponsorForm.link" :placeholder="adminText('k0053')" /></label>
            <DialogFooter>
              <Button variant="outline" type="button" @click="sponsorDialog = null">{{ adminText('k009q') }}</Button>
              <Button type="submit">{{ adminText('k005g') }}</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </BasicPage>
</template>
