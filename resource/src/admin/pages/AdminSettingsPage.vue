<script setup lang="ts">
import { adminText } from '@/admin/runtime/i18n-text'
import httpNotifyGuideZh from '@/admin/docs/http-notify-guide.zh.md?raw'
import httpNotifyGuideEn from '@/admin/docs/http-notify-guide.en.md?raw'
import httpNotifyGuideJa from '@/admin/docs/http-notify-guide.ja.md?raw'

import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import MarkdownIt from 'markdown-it'
import { Code, FileText, Globe, Loader2, MailCheck, Plus, Save, Send, Shield, Trash2, Upload, Webhook } from '@lucide/vue'
import AdminActionButton from '@/admin/components/AdminActionButton.vue'
import { BasicPage } from '@/admin/components/global-layout'
import { Button } from '@/admin/components/ui/button'
import { Badge } from '@/admin/components/ui/badge'
import { Input } from '@/admin/components/ui/input'
import { Textarea } from '@/admin/components/ui/textarea'
import { Switch } from '@/admin/components/ui/switch'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/admin/components/ui/tabs'
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
  getHttpNotifySettings,
  getMailSettings,
  getPostingSettings,
  getSecuritySettings,
  getSiteSettings,
  saveAnnouncement,
  saveHttpNotifySettings,
  saveMailSettings,
  savePostingSettings,
  saveSecuritySettings,
  saveSiteSettings,
  testMailConnection,
  uploadAdminImage,
} from '@/admin/runtime/api'
import { adminToast } from '@/admin/runtime/toast'
import { resolveApiMessage } from '@/runtime/api-message'
import type {
  AdminPayload,
  AnnouncementConfig,
  HttpNotifyEndpoint,
  HttpNotifySettings,
  MailSettings,
  ManageHomeProps,
  PostingSettings,
  SecuritySettings,
  SiteSettings,
} from '@/admin/types'

type Kind = 'site-info' | 'mail' | 'security' | 'posting' | 'http-notify' | 'announcement'

const props = defineProps<{
  payload: AdminPayload<ManageHomeProps>
  kind: Kind
}>()

const { locale } = useI18n()
const loading = ref(false)
const saving = ref(false)
const testing = ref(false)
const httpNotifyTab = ref('config')
const error = ref('')
const testEmail = ref('')
const newAllowedDomain = ref('')
const newExtension = ref('')

const siteForm = reactive<SiteSettings>({
  siteName: '',
  siteUrl: '',
  siteLogo: '',
  siteEmail: '',
  siteDescription: '',
  siteKeywords: '',
  externalLinks: '',
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

const httpNotifyEvents = computed(() => {
  locale.value
  return [
    { value: 'article.published', label: adminText('k00ck') },
    { value: 'article.updated', label: adminText('k00cl') },
    { value: 'comment.created', label: adminText('k00cm') },
    { value: 'user.signup', label: adminText('k00cn') },
    { value: 'moderation.report.created', label: adminText('k00co') },
  ]
})

const httpNotifyForm = reactive<HttpNotifySettings>({
  enabled: false,
  endpoints: [],
})

const guideMarkdown = new MarkdownIt({
  html: false,
  linkify: true,
  typographer: true,
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
    'http-notify': { title: adminText('k00cj'), description: adminText('k00cp') },
    announcement: { title: adminText('k0009'), description: adminText('k000a') },
  }
  return meta[props.kind]
})
const allowedDomains = computed(() => {
  return securityForm.allowedDomains
})

const httpNotifyGuideHtml = computed(() => {
  const guideLocale = String(locale.value).split(/[-_]/)[0]
  const guides = {
    zh: httpNotifyGuideZh,
    en: httpNotifyGuideEn,
    ja: httpNotifyGuideJa,
  }
  return guideMarkdown.render(guides[guideLocale as keyof typeof guides] || httpNotifyGuideZh)
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
  return {
    siteName: settings.siteName ?? '',
    siteUrl: settings.siteUrl ?? '',
    siteLogo: settings.siteLogo ?? '',
    siteEmail: settings.siteEmail ?? '',
    siteDescription: settings.siteDescription ?? '',
    siteKeywords: settings.siteKeywords ?? '',
    externalLinks: settings.externalLinks ?? '',
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

function normalizeHttpNotify(settings: Partial<HttpNotifySettings> = {}) {
  return {
    enabled: toBool(settings.enabled, false),
    endpoints: Array.isArray(settings.endpoints)
      ? settings.endpoints.map(endpoint => normalizeEndpoint(endpoint)).filter(endpoint => endpoint.url)
      : [],
  } satisfies HttpNotifySettings
}

function normalizeEndpoint(endpoint: Partial<HttpNotifyEndpoint> = {}) {
  const events = Array.isArray(endpoint.events)
    ? endpoint.events.map(item => String(item).trim()).filter(Boolean)
    : []
  const enabled = toBool(endpoint.enabled, true)
  return {
    id: endpoint.id || crypto.randomUUID(),
    name: endpoint.name ?? '',
    enabled,
    url: endpoint.url?.trim() ?? '',
    secret: endpoint.secret ?? '',
    events,
    timeoutSeconds: Math.min(Math.max(Number(endpoint.timeoutSeconds ?? 2), 1), 15),
    failureCount: enabled ? 0 : Number(endpoint.failureCount ?? 0),
    lastError: enabled ? '' : endpoint.lastError ?? '',
    abnormalTerminated: enabled ? false : toBool(endpoint.abnormalTerminated, false),
  } satisfies HttpNotifyEndpoint
}

function isHttpUrl(value: string) {
  try {
    const url = new URL(value)
    return url.protocol === 'http:' || url.protocol === 'https:'
  } catch {
    return false
  }
}

function validateHttpNotify(settings: HttpNotifySettings) {
  if (!settings.enabled) return true
  const enabledEndpoints = settings.endpoints.filter(endpoint => endpoint.enabled)
  if (enabledEndpoints.length === 0) {
    adminToast.warning(adminText('k00d2'))
    return false
  }
  for (const endpoint of enabledEndpoints) {
    const name = endpoint.name || endpoint.url || adminText('k00cw')
    if (!endpoint.url) {
      adminToast.warning(adminText('k00d3', { name }))
      return false
    }
    if (!isHttpUrl(endpoint.url)) {
      adminToast.warning(adminText('k00d4', { name }))
      return false
    }
    if (!endpoint.events.length) {
      adminToast.warning(adminText('k00d5', { name }))
      return false
    }
    if (!Number.isFinite(endpoint.timeoutSeconds) || endpoint.timeoutSeconds < 1 || endpoint.timeoutSeconds > 15) {
      adminToast.warning(adminText('k00d6', { name }))
      return false
    }
  }
  return true
}

function normalizeAnnouncement(settings: Partial<AnnouncementConfig> = {}) {
  return {
    enabled: toBool(settings.enabled, false),
    content: settings.content ?? '',
  } satisfies AnnouncementConfig
}

async function uploadImage(target: 'siteLogo', event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  try {
    const data = await uploadAdminImage(file)
    siteForm[target] = data.url || ''
    adminToast.success(adminText('k000b'))
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
    else if (props.kind === 'http-notify') Object.assign(httpNotifyForm, normalizeHttpNotify(await getHttpNotifySettings()))
    else Object.assign(announcementForm, normalizeAnnouncement(await getAnnouncement()))
  } catch (err) {
    error.value = err instanceof Error ? err.message : adminText('k000d')
  } finally {
    loading.value = false
  }
}

async function save() {
  const httpNotifySettings = props.kind === 'http-notify' ? normalizeHttpNotify(httpNotifyForm) : null
  if (httpNotifySettings && !validateHttpNotify(httpNotifySettings)) return

  saving.value = true
  try {
    if (props.kind === 'site-info') await saveSiteSettings(normalizeSite(siteForm))
    else if (props.kind === 'mail') await saveMailSettings(normalizeMail(mailForm))
    else if (props.kind === 'security') await saveSecuritySettings(normalizeSecurity(securityForm))
    else if (props.kind === 'posting') await savePostingSettings(normalizePosting(postingForm))
    else if (props.kind === 'http-notify') await saveHttpNotifySettings(httpNotifySettings!)
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

function addHttpEndpoint() {
  httpNotifyForm.endpoints.push({
    id: crypto.randomUUID(),
    name: '',
    enabled: true,
    url: '',
    secret: '',
    events: ['article.published'],
    timeoutSeconds: 2,
    failureCount: 0,
    lastError: '',
    abnormalTerminated: false,
  })
}

function removeHttpEndpoint(index: number) {
  httpNotifyForm.endpoints.splice(index, 1)
}

function toggleEndpointEvent(endpoint: HttpNotifyEndpoint, eventName: string, checked: boolean) {
  if (checked && !endpoint.events.includes(eventName)) {
    endpoint.events.push(eventName)
  }
  if (!checked) {
    endpoint.events = endpoint.events.filter(item => item !== eventName)
  }
}

function onEndpointEventChange(endpoint: HttpNotifyEndpoint, eventName: string, event: Event) {
  toggleEndpointEvent(endpoint, eventName, (event.target as HTMLInputElement).checked)
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
              <AdminActionButton compact tone="danger" class="-mr-1 size-5" :title="adminText('k005i')" @click="removeAllowedDomain(domain)">
                <Trash2 class="size-3.5" />
              </AdminActionButton>
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
                <AdminActionButton compact tone="danger" class="-mr-1 size-5" :title="adminText('k005i')" @click="removeExtension(ext)">
                  <Trash2 class="size-3.5" />
                </AdminActionButton>
              </Badge>
            </div>
          </div>
        </section>
      </form>

      <div v-else-if="kind === 'http-notify'" class="max-w-5xl">
        <Tabs :key="String(locale)" v-model="httpNotifyTab" class="gap-5">
          <TabsList class="w-fit">
            <TabsTrigger value="config">{{ adminText('k00d0') }}</TabsTrigger>
            <TabsTrigger value="guide">{{ adminText('k00d1') }}</TabsTrigger>
          </TabsList>

          <TabsContent value="config">
            <form class="space-y-5" @submit.prevent="save">
              <div class="flex items-center justify-between border-b pb-4">
                <div>
                  <div class="flex items-center gap-2 text-base font-medium"><Webhook class="size-4" />{{ adminText('k00cq') }}</div>
                  <p class="text-sm text-muted-foreground">{{ adminText('k00cr') }}</p>
                </div>
                <Switch v-model="httpNotifyForm.enabled" />
              </div>

              <section class="space-y-3">
                <div class="flex flex-wrap items-center justify-between gap-3">
                  <div>
                    <div class="text-lg font-medium">{{ adminText('k00cs') }}</div>
                    <p class="text-sm text-muted-foreground">{{ adminText('k00ct') }}</p>
                  </div>
                  <Button type="button" variant="secondary" @click="addHttpEndpoint"><Plus class="size-4" />{{ adminText('k00cu') }}</Button>
                </div>

                <div v-if="httpNotifyForm.endpoints.length === 0" class="rounded-lg border border-dashed p-8 text-center text-sm text-muted-foreground">
                  {{ adminText('k00cv') }}
                </div>

                <div v-for="(endpoint, index) in httpNotifyForm.endpoints" :key="endpoint.id || index" class="space-y-3 rounded-lg border bg-background p-3">
                  <div class="flex flex-wrap items-center justify-between gap-3">
                    <div class="flex min-w-0 items-center gap-3">
                      <Switch v-model="endpoint.enabled" :disabled="!httpNotifyForm.enabled" />
                      <div class="min-w-0">
                        <div class="flex flex-wrap items-center gap-2">
                          <span class="font-medium">{{ endpoint.name || adminText('k00cw') }}</span>
                          <Badge v-if="endpoint.abnormalTerminated" variant="destructive" class="px-2 py-0 text-xs">{{ adminText('k00cz') }}</Badge>
                        </div>
                        <div v-if="endpoint.lastError" class="mt-0.5 truncate text-xs text-muted-foreground">{{ endpoint.lastError }}</div>
                      </div>
                    </div>
                    <AdminActionButton compact tone="danger" :title="adminText('k005i')" @click="removeHttpEndpoint(index)">
                      <Trash2 class="size-4" />
                    </AdminActionButton>
                  </div>

                  <div class="grid gap-3 md:grid-cols-[minmax(140px,220px)_minmax(0,1fr)_120px]">
                    <label class="grid gap-2 text-sm font-medium">
                      {{ adminText('k0079') }}
                      <Input v-model="endpoint.name" :disabled="!httpNotifyForm.enabled" placeholder="Webhook" />
                    </label>
                    <label class="grid gap-2 text-sm font-medium">
                      URL
                      <Input v-model="endpoint.url" :disabled="!httpNotifyForm.enabled" placeholder="http://example.com/webhook" />
                    </label>
                    <label class="grid gap-2 text-sm font-medium">
                      {{ adminText('k00cx') }}
                      <Input v-model.number="endpoint.timeoutSeconds" :disabled="!httpNotifyForm.enabled" type="number" min="1" max="15" />
                    </label>
                  </div>

                  <label class="grid gap-2 text-sm font-medium">
                    Secret
                    <Input v-model="endpoint.secret" :disabled="!httpNotifyForm.enabled" type="password" autocomplete="new-password" />
                  </label>

                  <div class="space-y-2">
                    <div class="text-sm font-medium">{{ adminText('k00cy') }}</div>
                    <div class="flex flex-wrap gap-2">
                      <label v-for="item in httpNotifyEvents" :key="item.value" class="inline-flex items-center gap-2 rounded-md border px-2.5 py-1.5 text-sm">
                        <input
                          type="checkbox"
                          :disabled="!httpNotifyForm.enabled"
                          :checked="endpoint.events.includes(item.value)"
                          @change="onEndpointEventChange(endpoint, item.value, $event)"
                        />
                        {{ item.label }}
                      </label>
                    </div>
                  </div>
                </div>
              </section>
            </form>
          </TabsContent>

          <TabsContent value="guide">
            <div class="gf-admin-markdown rounded-lg border bg-background p-5" v-html="httpNotifyGuideHtml" />
          </TabsContent>
        </Tabs>
      </div>

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

    </BasicPage>
</template>

<style scoped>
.gf-admin-markdown {
  color: var(--foreground);
  font-size: 0.875rem;
  line-height: 1.7;
}

.gf-admin-markdown :deep(h2) {
  margin-bottom: 0.75rem;
  font-size: 1.125rem;
  font-weight: 600;
}

.gf-admin-markdown :deep(h3) {
  margin-top: 1.25rem;
  margin-bottom: 0.5rem;
  font-weight: 600;
}

.gf-admin-markdown :deep(p),
.gf-admin-markdown :deep(ul) {
  margin-bottom: 0.75rem;
}

.gf-admin-markdown :deep(ul) {
  padding-left: 1.25rem;
  list-style: disc;
}

.gf-admin-markdown :deep(code) {
  border-radius: 0.25rem;
  background: var(--muted);
  padding: 0.125rem 0.25rem;
  font-size: 0.8125rem;
}

.gf-admin-markdown :deep(pre) {
  margin: 0.75rem 0 1rem;
  overflow-x: auto;
  border-radius: 0.5rem;
  background: var(--muted);
  padding: 0.875rem;
}

.gf-admin-markdown :deep(pre code) {
  background: transparent;
  padding: 0;
}
</style>
