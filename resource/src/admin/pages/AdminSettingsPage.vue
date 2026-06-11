<script setup lang="ts">import { adminText } from '@/admin/runtime/i18n-text'

import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Draggable from 'vuedraggable'
import { Code, FileText, Globe, Loader2, MailCheck, Plus, Save, Send, Shield, Trash2, Upload } from '@lucide/vue'
import { BasicPage } from '@/admin/components/global-layout'
import { Button } from '@/admin/components/ui/button'
import { Badge } from '@/admin/components/ui/badge'
import { Input } from '@/admin/components/ui/input'
import { Textarea } from '@/admin/components/ui/textarea'
import { Switch } from '@/admin/components/ui/switch'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/admin/components/ui/dialog'
import {
  getAnnouncement,
  getMailSettings,
  getPostingSettings,
  getSecuritySettings,
  getSiteSettings,
  saveAnnouncement,
  saveMailSettings,
  savePostingSettings,
  saveSecuritySettings,
  saveSiteSettings,
  testMailConnection,
} from '@/admin/runtime/api'
import { adminToast } from '@/admin/runtime/toast'
import { resolveApiMessage } from '@/runtime/api-message'
import type {
  AdminPayload,
  AnnouncementConfig,
  MailSettings,
  ManageHomeProps,
  PostingSettings,
  SecuritySettings,
  SiteSettings,
} from '@/admin/types'

type Kind = 'site-info' | 'mail' | 'security' | 'posting' | 'announcement'

const props = defineProps<{
  payload: AdminPayload<ManageHomeProps>
  kind: Kind
}>()

const { locale } = useI18n()
const loading = ref(false)
const saving = ref(false)
const testing = ref(false)
const error = ref('')
const testEmail = ref('')
const newAllowedDomain = ref('')
const newExtension = ref('')
const footerDialog = ref<null | { type: 'link' | 'primary', index: number }>(null)
const footerLinkForm = reactive({ name: '', url: '' })
const footerPrimaryForm = reactive({ content: '' })

const siteForm = reactive<SiteSettings>({
  siteName: '',
  siteUrl: '',
  siteLogo: '',
  siteEmail: '',
  siteDescription: '',
  siteKeywords: '',
  externalLinks: '',
  footerInfo: { primary: [], list: [] },
  brandType: 'default',
  brandText: '',
  brandImage: '',
})

const mailForm = reactive<MailSettings>({
  enableMail: true,
  smtpHost: '',
  smtpPort: 587,
  useSSL: false,
  smtpUsername: '',
  smtpPassword: '',
  fromName: '',
  fromEmail: '',
})

const securityForm = reactive<SecuritySettings>({
  enableSignup: true,
  enableEmailVerification: false,
  allowedDomains: [],
})

const postingForm = reactive<PostingSettings>({
  textControl: {
    minPostLength: 5,
    maxPostLength: 50000,
    minTitleLength: 5,
    maxTitleLength: 100,
    newUserPostCooldownMinutes: 0,
  },
  uploadControl: {
    allowAttachments: true,
    authorizedExtensions: ['.jpg', '.jpeg', '.png', '.gif', '.webp'],
    maxAttachmentSizeKb: 5120,
    maxDailyUploadsPerUser: 10,
    newUserUploadCooldownMinutes: 0,
  },
})

const announcementForm = reactive<AnnouncementConfig>({
  enabled: false,
  content: '',
})

const pageMeta = computed(() => {
  locale.value
  const meta: Record<Kind, { title: string, description: string }> = {
    'site-info': { title: adminText('k0001'), description: adminText('k0002') },
    mail: { title: adminText('k0003'), description: adminText('k0004') },
    security: { title: adminText('k0005'), description: adminText('k0006') },
    posting: { title: adminText('k0007'), description: adminText('k0008') },
    announcement: { title: adminText('k0009'), description: adminText('k000a') },
  }
  return meta[props.kind]
})
const allowedDomains = computed(() => {
  return securityForm.allowedDomains
})

function toBool(value: unknown, fallback = false) {
  if (typeof value === 'boolean') return value
  if (typeof value === 'number') return value === 1
  if (typeof value === 'string') {
    const normalized = value.trim().toLowerCase()
    if (['true', '1', 'yes', 'on', 'enabled'].includes(normalized)) return true
    if (['false', '0', 'no', 'off', 'disabled', ''].includes(normalized)) return false
  }
  return fallback
}

function normalizeSite(settings: Partial<SiteSettings> = {}) {
  const footerInfo = settings.footerInfo && typeof settings.footerInfo === 'object'
    ? settings.footerInfo
    : { primary: [], list: [] }
  return {
    siteName: settings.siteName ?? '',
    siteUrl: settings.siteUrl ?? '',
    siteLogo: settings.siteLogo ?? '',
    siteEmail: settings.siteEmail ?? '',
    siteDescription: settings.siteDescription ?? '',
    siteKeywords: settings.siteKeywords ?? '',
    externalLinks: settings.externalLinks ?? '',
    footerInfo: {
      primary: Array.isArray(footerInfo.primary)
        ? footerInfo.primary.map(item => ({ content: item?.content ?? '' }))
        : [],
      list: Array.isArray(footerInfo.list)
        ? footerInfo.list.map(item => ({ name: item?.name ?? '', url: item?.url ?? '' }))
        : [],
    },
    brandType: ['default', 'text', 'image'].includes(settings.brandType || '') ? settings.brandType : 'default',
    brandText: settings.brandText ?? '',
    brandImage: settings.brandImage ?? '',
  } satisfies SiteSettings
}

function normalizeMail(settings: Partial<MailSettings> = {}) {
  return {
    enableMail: toBool(settings.enableMail, false),
    smtpHost: settings.smtpHost ?? '',
    smtpPort: Number(settings.smtpPort ?? 587),
    useSSL: toBool(settings.useSSL, false),
    smtpUsername: settings.smtpUsername ?? '',
    smtpPassword: settings.smtpPassword ?? '',
    fromName: settings.fromName ?? '',
    fromEmail: settings.fromEmail ?? '',
  } satisfies MailSettings
}

function normalizeSecurity(settings: Partial<SecuritySettings> = {}) {
  return {
    enableSignup: toBool(settings.enableSignup, true),
    enableEmailVerification: toBool(settings.enableEmailVerification, false),
    allowedDomains: Array.isArray(settings.allowedDomains)
      ? settings.allowedDomains.map(item => String(item).trim().toLowerCase()).filter(Boolean)
      : [],
  } satisfies SecuritySettings
}

function normalizePosting(settings: Partial<PostingSettings> = {}) {
  return {
    textControl: {
      minPostLength: Number(settings.textControl?.minPostLength ?? 5),
      maxPostLength: Number(settings.textControl?.maxPostLength ?? 50000),
      minTitleLength: Number(settings.textControl?.minTitleLength ?? 5),
      maxTitleLength: Number(settings.textControl?.maxTitleLength ?? 100),
      newUserPostCooldownMinutes: Number(settings.textControl?.newUserPostCooldownMinutes ?? 0),
    },
    uploadControl: {
      allowAttachments: toBool(settings.uploadControl?.allowAttachments, true),
      authorizedExtensions: Array.isArray(settings.uploadControl?.authorizedExtensions)
        ? settings.uploadControl.authorizedExtensions.map(item => String(item).trim().toLowerCase()).filter(Boolean)
        : ['.jpg', '.jpeg', '.png', '.gif', '.webp'],
      maxAttachmentSizeKb: Number(settings.uploadControl?.maxAttachmentSizeKb ?? 5120),
      maxDailyUploadsPerUser: Number(settings.uploadControl?.maxDailyUploadsPerUser ?? 10),
      newUserUploadCooldownMinutes: Number(settings.uploadControl?.newUserUploadCooldownMinutes ?? 1440),
    },
  } satisfies PostingSettings
}

function normalizeAnnouncement(settings: Partial<AnnouncementConfig> = {}) {
  return {
    enabled: toBool(settings.enabled, false),
    content: settings.content ?? '',
  } satisfies AnnouncementConfig
}

async function uploadImage(target: 'siteLogo' | 'brandImage', event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  const body = new FormData()
  body.append('file', file)
  try {
    const response = await fetch('/file/img-upload', { method: 'POST', body })
    const data = await response.json()
    if (data.code === 0) {
      siteForm[target] = data.result?.url || ''
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

async function load() {
  loading.value = true
  error.value = ''
  try {
    if (props.kind === 'site-info') Object.assign(siteForm, normalizeSite(await getSiteSettings()))
    else if (props.kind === 'mail') Object.assign(mailForm, normalizeMail(await getMailSettings()))
    else if (props.kind === 'security') Object.assign(securityForm, normalizeSecurity(await getSecuritySettings()))
    else if (props.kind === 'posting') Object.assign(postingForm, normalizePosting(await getPostingSettings()))
    else Object.assign(announcementForm, normalizeAnnouncement(await getAnnouncement()))
  } catch (err) {
    error.value = err instanceof Error ? err.message : adminText('k000d')
  } finally {
    loading.value = false
  }
}

async function save() {
  saving.value = true
  try {
    if (props.kind === 'site-info') await saveSiteSettings(normalizeSite(siteForm))
    else if (props.kind === 'mail') await saveMailSettings(normalizeMail(mailForm))
    else if (props.kind === 'security') await saveSecuritySettings(normalizeSecurity(securityForm))
    else if (props.kind === 'posting') await savePostingSettings(normalizePosting(postingForm))
    else await saveAnnouncement(normalizeAnnouncement(announcementForm))
    adminToast.success(adminText('k000e'))
  } catch (err) {
    adminToast.error(err, adminText('k000f'))
  } finally {
    saving.value = false
  }
}

async function sendTestMail() {
  if (!testEmail.value.trim()) {
    adminToast.warning(adminText('k000g'))
    return
  }
  testing.value = true
  try {
    const response = await testMailConnection(normalizeMail(mailForm), testEmail.value.trim())
    adminToast.success(resolveApiMessage(response.result || response, adminText('k000h')))
  } catch (err) {
    adminToast.error(err, adminText('k000i'))
  } finally {
    testing.value = false
  }
}

function addAllowedDomain() {
  const domain = newAllowedDomain.value.trim().toLowerCase()
  if (!domain || allowedDomains.value.includes(domain)) return
  const next = [...allowedDomains.value, domain]
  securityForm.allowedDomains = next
  newAllowedDomain.value = ''
}

function removeAllowedDomain(domain: string) {
  const next = allowedDomains.value.filter(item => item !== domain)
  securityForm.allowedDomains = next
}

function addExtension() {
  const ext = newExtension.value.trim().toLowerCase()
  if (!ext) return
  if (!ext.startsWith('.')) {
    adminToast.warning(adminText('k000j'))
    return
  }
  if (!postingForm.uploadControl.authorizedExtensions.includes(ext)) {
    postingForm.uploadControl.authorizedExtensions.push(ext)
  }
  newExtension.value = ''
}

function removeExtension(ext: string) {
  postingForm.uploadControl.authorizedExtensions = postingForm.uploadControl.authorizedExtensions.filter(item => item !== ext)
}

function addFooterPrimary() {
  siteForm.footerInfo ||= { primary: [], list: [] }
  siteForm.footerInfo.primary.push({ content: '' })
  openFooterPrimary(siteForm.footerInfo.primary.length - 1)
}

function addFooterLink() {
  siteForm.footerInfo ||= { primary: [], list: [] }
  siteForm.footerInfo.list.push({ name: '', url: '' })
  openFooterLink(siteForm.footerInfo.list.length - 1)
}

function openFooterLink(index: number) {
  const item = siteForm.footerInfo?.list[index]
  if (!item) return
  footerLinkForm.name = item.name
  footerLinkForm.url = item.url
  footerDialog.value = { type: 'link', index }
}

function openFooterPrimary(index: number) {
  const item = siteForm.footerInfo?.primary[index]
  if (!item) return
  footerPrimaryForm.content = item.content
  footerDialog.value = { type: 'primary', index }
}

function saveFooterItem() {
  if (!footerDialog.value) return
  const { type, index } = footerDialog.value
  if (type === 'link') {
    const item = siteForm.footerInfo?.list[index]
    if (item) {
      item.name = footerLinkForm.name.trim()
      item.url = footerLinkForm.url.trim()
    }
  } else {
    const item = siteForm.footerInfo?.primary[index]
    if (item) {
      item.content = footerPrimaryForm.content.trim()
    }
  }
  footerDialog.value = null
}

function closeFooterDialog(open: boolean) {
  if (!open) footerDialog.value = null
}

function addAnnouncementExample() {
  if (!announcementForm.content) {
    announcementForm.content = adminText('k000k')
  }
}

watch(() => props.kind, () => {
  void load()
})

onMounted(load)
</script>

<template>
  <BasicPage :title="pageMeta.title" :description="pageMeta.description" sticky>
    <template #actions>
      <Button type="button" :disabled="saving" @click="save">
        <Loader2 v-if="saving" class="size-4 animate-spin" />
        <Save v-else class="size-4" />
        {{ adminText('k004f') }}
      </Button>
    </template>

      <div v-if="loading" class="flex h-[400px] items-center justify-center">
        <Loader2 class="size-8 animate-spin text-primary" />
      </div>
      <div v-else-if="error" class="rounded-lg border border-destructive/30 bg-destructive/5 p-4 text-sm text-destructive">{{ error }}</div>

      <form v-else-if="kind === 'site-info'" class="space-y-10" @submit.prevent="save">
        <div class="grid gap-10 md:grid-cols-2">
          <section class="space-y-6">
            <div class="flex items-center gap-2 border-b pb-2 text-lg font-medium"><Globe class="size-5 text-muted-foreground" />{{ adminText('k007z') }}</div>
            <div class="space-y-4">
              <div class="flex items-center gap-2 border-b pb-2 text-base font-medium">{{ adminText('k0085') }}</div>
              <div class="flex flex-wrap gap-4 text-sm">
                <label class="flex items-center gap-2"><input v-model="siteForm.brandType" type="radio" value="default" />{{ adminText('k0086') }}</label>
                <label class="flex items-center gap-2"><input v-model="siteForm.brandType" type="radio" value="text" />{{ adminText('k0087') }}</label>
                <label class="flex items-center gap-2"><input v-model="siteForm.brandType" type="radio" value="image" />{{ adminText('k0088') }}</label>
              </div>
              <label v-if="siteForm.brandType === 'default'" class="grid gap-2 text-sm font-medium">{{ adminText('k0086') }}<Input model-value="GooseForum" disabled /></label>
              <label v-if="siteForm.brandType === 'text'" class="grid gap-2 text-sm font-medium">{{ adminText('k0089') }}<Input v-model="siteForm.brandText" placeholder="MyBrand" /></label>
              <label v-if="siteForm.brandType === 'image'" class="grid gap-2 text-sm font-medium">{{ adminText('k008a') }}
                <div class="flex gap-2">
                  <Input v-model="siteForm.brandImage" placeholder="Brand Image URL" />
                  <label class="inline-flex size-9 shrink-0 cursor-pointer items-center justify-center rounded-md border bg-background shadow-xs hover:bg-accent">
                    <Upload class="size-4" />
                    <input class="hidden" type="file" accept="image/*" @change="uploadImage('brandImage', $event)" />
                  </label>
                </div>
              </label>
            </div>
            <label class="grid gap-2 text-sm font-medium">{{ adminText('k0080') }}<Input v-model="siteForm.siteName" placeholder="GooseForum" /></label>
            <label class="grid gap-2 text-sm font-medium">{{ adminText('k0081') }}<Input v-model="siteForm.siteUrl" placeholder="https://example.com" /></label>
            <label class="grid gap-2 text-sm font-medium">{{ adminText('k0082') }}<Input v-model="siteForm.siteEmail" placeholder="contact@example.com" /></label>
            <label class="grid gap-2 text-sm font-medium">{{ adminText('k0083') }}
              <div class="flex items-start gap-4">
                <div class="grid size-20 shrink-0 place-items-center overflow-hidden rounded-lg border bg-muted">
                  <img v-if="siteForm.siteLogo" :src="siteForm.siteLogo" class="size-full object-cover" alt="Logo" />
                  <Upload v-else class="size-8 text-muted-foreground/50" />
                </div>
                <div class="grid flex-1 gap-2">
                  <Input v-model="siteForm.siteLogo" placeholder="Logo URL" />
                  <label class="inline-flex h-9 w-fit cursor-pointer items-center gap-2 rounded-md border bg-background px-3 text-sm font-medium shadow-xs hover:bg-accent">
                    {{ adminText('k0084') }}
                    <input class="hidden" type="file" accept="image/*" @change="uploadImage('siteLogo', $event)" />
                  </label>
                </div>
              </div>
            </label>
          </section>
          <section class="space-y-6">
            <div class="flex items-center gap-2 border-b pb-2 text-lg font-medium"><FileText class="size-5 text-muted-foreground" />{{ adminText('k008b') }}</div>
            <label class="grid gap-2 text-sm font-medium">{{ adminText('k008c') }}<Textarea v-model="siteForm.siteDescription" class="min-h-24" /></label>
            <label class="grid gap-2 text-sm font-medium">{{ adminText('k008d') }}<Input v-model="siteForm.siteKeywords" placeholder="forum, community" /></label>
            <label class="grid gap-2 text-sm font-medium">{{ adminText('k008e') }}<Textarea v-model="siteForm.externalLinks" class="min-h-28 font-mono text-xs" /></label>
            <div class="space-y-4 border-y py-4">
              <div class="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <div class="font-semibold">{{ adminText('k008f') }}</div>
                  <p class="text-xs text-muted-foreground">{{ adminText('k008g') }}</p>
                </div>
                <div class="flex gap-2">
                  <Button variant="ghost" size="sm" type="button" @click="addFooterLink"><Plus class="size-4" />{{ adminText('k0071') }}</Button>
                  <Button variant="ghost" size="sm" type="button" @click="addFooterPrimary"><Plus class="size-4" />{{ adminText('k008h') }}</Button>
                </div>
              </div>
              <div class="rounded-lg border bg-background px-3 py-3 text-xs leading-5 text-muted-foreground shadow-xs">
                <div class="space-y-1.5">
                  <Draggable
                    v-model="siteForm.footerInfo!.list"
                    item-key="name"
                    direction="horizontal"
                    handle=".js-footer-handle"
                    class="flex min-h-5 flex-wrap items-center gap-x-3 gap-y-1"
                    ghost-class="opacity-40"
                  >
                    <template #item="{ element: item, index }">
                      <div class="group inline-flex min-h-5 items-center rounded text-muted-foreground transition-colors hover:bg-muted/60 hover:text-primary">
                        <span class="js-footer-handle -ml-1 mr-0.5 cursor-grab text-muted-foreground/45 opacity-0 transition-opacity group-hover:opacity-100 active:cursor-grabbing">⋮⋮</span>
                        <button class="inline-flex min-h-5 items-center rounded px-0.5 text-left text-xs" type="button" @click="openFooterLink(index)">
                          {{ item.name || adminText('k004g') }}
                        </button>
                        <Button variant="ghost" size="icon" class="ml-0.5 size-5 rounded-sm opacity-0 hover:bg-destructive/10 hover:text-destructive group-hover:opacity-100" type="button" @click="siteForm.footerInfo?.list.splice(index, 1)">
                          <Trash2 class="size-3.5" />
                        </Button>
                      </div>
                    </template>
                    <template #footer>
                      <div v-if="siteForm.footerInfo?.list.length === 0" class="min-h-5 text-muted-foreground/70">{{ adminText('k008i') }}</div>
                    </template>
                  </Draggable>

                  <Draggable
                    v-model="siteForm.footerInfo!.primary"
                    item-key="content"
                    handle=".js-footer-handle"
                    class="mt-1 flex min-h-5 flex-wrap items-center gap-x-3 gap-y-1 text-muted-foreground"
                    ghost-class="opacity-40"
                  >
                    <template #item="{ element: item, index }">
                      <div class="group inline-flex min-h-5 max-w-full items-center rounded transition-colors hover:bg-muted/60">
                        <span class="js-footer-handle -ml-1 mr-0.5 cursor-grab text-muted-foreground/45 opacity-0 transition-opacity group-hover:opacity-100 active:cursor-grabbing">⋮⋮</span>
                        <button class="inline-flex min-h-5 items-center rounded px-0.5 text-left text-xs text-muted-foreground" type="button" @click="openFooterPrimary(index)">
                          {{ item.content || adminText('k004h') }}
                        </button>
                        <Button variant="ghost" size="icon" class="ml-0.5 size-5 rounded-sm opacity-0 hover:bg-destructive/10 hover:text-destructive group-hover:opacity-100" type="button" @click="siteForm.footerInfo?.primary.splice(index, 1)">
                          <Trash2 class="size-3.5" />
                        </Button>
                      </div>
                    </template>
                    <template #footer>
                      <div v-if="siteForm.footerInfo?.primary.length === 0" class="min-h-5 text-muted-foreground/70">{{ adminText('k0062') }}</div>
                    </template>
                  </Draggable>
                </div>
              </div>
            </div>
          </section>
        </div>
      </form>

      <form v-else-if="kind === 'mail'" class="grid gap-10 lg:grid-cols-[minmax(0,2fr)_minmax(260px,1fr)]" @submit.prevent="save">
        <section class="space-y-6">
          <div class="flex items-center gap-2 border-b pb-2 text-lg font-medium">{{ adminText('k008j') }}</div>
          <div class="flex items-center justify-between rounded-lg border bg-muted/20 p-4">
            <div><div class="font-medium">{{ adminText('k008k') }}</div><p class="text-sm text-muted-foreground">{{ adminText('k008l') }}</p></div>
            <Switch v-model="mailForm.enableMail" />
          </div>
          <div class="grid gap-6 md:grid-cols-2">
            <label class="grid gap-2 text-sm font-medium">{{ adminText('k008m') }}<Input v-model="mailForm.smtpHost" :disabled="!mailForm.enableMail" placeholder="smtp.example.com" /></label>
            <label class="grid gap-2 text-sm font-medium">{{ adminText('k008n') }}<Input v-model.number="mailForm.smtpPort" :disabled="!mailForm.enableMail" type="number" /></label>
            <label class="grid gap-2 text-sm font-medium">{{ adminText('k008o') }}<Input v-model="mailForm.smtpUsername" :disabled="!mailForm.enableMail" /></label>
            <label class="grid gap-2 text-sm font-medium">{{ adminText('k008p') }}<Input v-model="mailForm.smtpPassword" :disabled="!mailForm.enableMail" type="password" /></label>
          </div>
          <div class="flex items-center justify-between rounded-lg border bg-muted/20 p-4">
            <div><div class="flex items-center gap-2 font-medium"><Shield class="size-4" />{{ adminText('k008q') }}</div><p class="text-sm text-muted-foreground">{{ adminText('k008r') }}</p></div>
            <Switch v-model="mailForm.useSSL" :disabled="!mailForm.enableMail" />
          </div>
        </section>
        <aside class="space-y-8">
          <section class="space-y-4">
            <div class="flex items-center gap-2 border-b pb-2 text-lg font-medium"><Send class="size-5 text-muted-foreground" />{{ adminText('k008s') }}</div>
            <label class="grid gap-2 text-sm font-medium">{{ adminText('k008t') }}<Input v-model="mailForm.fromName" :disabled="!mailForm.enableMail" placeholder="GooseForum" /></label>
            <label class="grid gap-2 text-sm font-medium">{{ adminText('k008u') }}<Input v-model="mailForm.fromEmail" :disabled="!mailForm.enableMail" placeholder="noreply@example.com" /></label>
          </section>
          <section class="space-y-4">
            <div class="flex items-center gap-2 border-b pb-2 text-lg font-medium">{{ adminText('k008v') }}</div>
            <div class="space-y-3 rounded-lg border p-4">
              <p class="text-sm text-muted-foreground">{{ adminText('k008w') }}</p>
              <Input v-model="testEmail" :disabled="!mailForm.enableMail" :placeholder="adminText('k004k')" />
              <Button class="w-full" type="button" variant="secondary" :disabled="testing || !mailForm.enableMail" @click="sendTestMail">
                <Loader2 v-if="testing" class="size-4 animate-spin" />
                <Send v-else class="size-4" />
                {{ adminText('k008x') }}
              </Button>
            </div>
          </section>
        </aside>
      </form>

      <form v-else-if="kind === 'security'" class="max-w-2xl space-y-8" @submit.prevent="save">
        <div class="flex items-center justify-between">
          <div><div class="text-base font-medium">{{ adminText('k008y') }}</div><p class="text-sm text-muted-foreground">{{ adminText('k008z') }}</p></div>
          <Switch v-model="securityForm.enableSignup" />
        </div>
        <div class="flex items-center justify-between">
          <div><div class="flex items-center gap-2 text-base font-medium"><MailCheck class="size-4" />{{ adminText('k0090') }}</div><p class="text-sm text-muted-foreground">{{ adminText('k0091') }}</p></div>
          <Switch v-model="securityForm.enableEmailVerification" />
        </div>
        <div class="space-y-4">
          <div><div class="text-base font-medium">{{ adminText('k0092') }}</div><p class="text-sm text-muted-foreground">{{ adminText('k0093') }}</p></div>
          <div class="flex gap-2">
            <Input v-model="newAllowedDomain" class="max-w-sm" :placeholder="adminText('k004l')" @keydown.enter.prevent="addAllowedDomain" />
            <Button type="button" variant="secondary" @click="addAllowedDomain"><Plus class="size-4" />{{ adminText('k0094') }}</Button>
          </div>
          <div class="flex flex-wrap gap-2">
            <span v-if="allowedDomains.length === 0" class="text-sm italic text-muted-foreground">{{ adminText('k0095') }}</span>
            <Badge v-for="domain in allowedDomains" :key="domain" variant="secondary" class="gap-2 px-3 py-1.5 text-sm font-normal">
              {{ domain }}
              <button type="button" @click="removeAllowedDomain(domain)"><Trash2 class="size-3.5 text-muted-foreground hover:text-destructive" /></button>
            </Badge>
          </div>
        </div>
      </form>

      <form v-else-if="kind === 'posting'" class="grid gap-12 lg:grid-cols-2" @submit.prevent="save">
        <section class="space-y-6">
          <div class="flex items-center gap-2 border-b pb-2 text-lg font-medium"><FileText class="size-5 text-muted-foreground" />{{ adminText('k0096') }}</div>
          <div class="grid gap-6 sm:grid-cols-2">
            <label class="grid gap-2 text-sm font-medium">{{ adminText('k0097') }}<Input v-model.number="postingForm.textControl.minTitleLength" type="number" /></label>
            <label class="grid gap-2 text-sm font-medium">{{ adminText('k0098') }}<Input v-model.number="postingForm.textControl.maxTitleLength" type="number" /></label>
            <label class="grid gap-2 text-sm font-medium">{{ adminText('k0099') }}<Input v-model.number="postingForm.textControl.minPostLength" type="number" /></label>
            <label class="grid gap-2 text-sm font-medium">{{ adminText('k009a') }}<Input v-model.number="postingForm.textControl.maxPostLength" type="number" /></label>
            <label class="grid gap-2 text-sm font-medium">{{ adminText('k009b') }}<Input v-model.number="postingForm.textControl.newUserPostCooldownMinutes" type="number" /></label>
          </div>
        </section>
        <section class="space-y-6">
          <div class="flex items-center gap-2 border-b pb-2 text-lg font-medium"><Upload class="size-5 text-muted-foreground" />{{ adminText('k009c') }}</div>
          <div class="flex items-center justify-between rounded-lg border bg-muted/10 p-4">
            <div><div class="text-base font-medium">{{ adminText('k009d') }}</div><p class="text-sm text-muted-foreground">{{ adminText('k009e') }}</p></div>
            <Switch v-model="postingForm.uploadControl.allowAttachments" />
          </div>
          <div class="grid gap-6 sm:grid-cols-2">
            <label class="grid gap-2 text-sm font-medium">{{ adminText('k009f') }}<Input v-model.number="postingForm.uploadControl.maxDailyUploadsPerUser" :disabled="!postingForm.uploadControl.allowAttachments" type="number" /></label>
            <label class="grid gap-2 text-sm font-medium">{{ adminText('k009g') }}<Input v-model.number="postingForm.uploadControl.maxAttachmentSizeKb" :disabled="!postingForm.uploadControl.allowAttachments" type="number" /></label>
            <label class="grid gap-2 text-sm font-medium">{{ adminText('k009h') }}<Input v-model.number="postingForm.uploadControl.newUserUploadCooldownMinutes" :disabled="!postingForm.uploadControl.allowAttachments" type="number" /></label>
          </div>
          <div class="space-y-3">
            <div class="text-sm font-medium">{{ adminText('k009i') }}</div>
            <div class="flex gap-2">
              <Input v-model="newExtension" :placeholder="adminText('k004m')" @keydown.enter.prevent="addExtension" />
              <Button type="button" variant="secondary" @click="addExtension"><Plus class="size-4" />{{ adminText('k0094') }}</Button>
            </div>
            <div class="flex flex-wrap gap-2">
              <Badge v-for="ext in postingForm.uploadControl.authorizedExtensions" :key="ext" variant="secondary" class="gap-2 px-3 py-1.5 text-sm font-normal">
                {{ ext }}
                <button type="button" @click="removeExtension(ext)"><Trash2 class="size-3.5 text-muted-foreground hover:text-destructive" /></button>
              </Badge>
            </div>
          </div>
        </section>
      </form>

      <form v-else class="max-w-3xl space-y-6" @submit.prevent="save">
        <div class="flex items-center justify-between">
          <div><div class="text-base font-medium">{{ adminText('k009j') }}</div><p class="text-sm text-muted-foreground">{{ adminText('k009k') }}</p></div>
          <Switch v-model="announcementForm.enabled" />
        </div>
        <label class="grid gap-2 text-sm font-medium">
          {{ adminText('k009l') }}
          <span class="text-sm font-normal text-muted-foreground">{{ adminText('k009m') }}</span>
          <Textarea v-model="announcementForm.content" class="min-h-64 resize-y font-mono text-sm" :placeholder="adminText('k004n')" />
        </label>
        <Button variant="outline" type="button" @click="addAnnouncementExample"><Code class="size-4" />{{ adminText('k009n') }}</Button>
      </form>

      <Dialog v-if="kind === 'site-info'" :open="footerDialog !== null" @update:open="closeFooterDialog">
        <DialogContent class="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{{ footerDialog?.type === 'link' ? adminText('k004i') : adminText('k004j') }}</DialogTitle>
            <DialogDescription>
              {{ adminText('k009o') }}
            </DialogDescription>
          </DialogHeader>
          <form v-if="footerDialog?.type === 'link'" class="grid gap-4" @submit.prevent="saveFooterItem">
            <label class="grid gap-2 text-sm font-medium">
              {{ adminText('k0079') }}
              <Input v-model="footerLinkForm.name" placeholder="Github" />
            </label>
            <label class="grid gap-2 text-sm font-medium">
              {{ adminText('k009p') }}
              <Input v-model="footerLinkForm.url" placeholder="https://example.com" />
            </label>
            <DialogFooter>
              <Button variant="outline" type="button" @click="footerDialog = null">{{ adminText('k009q') }}</Button>
              <Button type="submit">{{ adminText('k005g') }}</Button>
            </DialogFooter>
          </form>
          <form v-else class="grid gap-4" @submit.prevent="saveFooterItem">
            <label class="grid gap-2 text-sm font-medium">
              {{ adminText('k009r') }}
              <Input v-model="footerPrimaryForm.content" placeholder="Powered by GooseForum" />
            </label>
            <DialogFooter>
              <Button variant="outline" type="button" @click="footerDialog = null">{{ adminText('k009q') }}</Button>
              <Button type="submit">{{ adminText('k005g') }}</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </BasicPage>
</template>
